package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/llm"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

type artifactResponse struct {
	domain.Artifact
	CreatedByUser *domain.User `json:"created_by_user,omitempty"`
	UpdatedByUser *domain.User `json:"updated_by_user,omitempty"`
}

type artifactVersionResponse struct {
	domain.ArtifactVersion
	UpdatedByUser *domain.User `json:"updated_by_user,omitempty"`
}

func GetArtifacts(c *gin.Context) {
	var artifacts []domain.Artifact

	query := db.DB.Order("updated_at desc")
	if channelID := c.Query("channel_id"); channelID != "" {
		query = query.Where("channel_id = ?", channelID)
	}
	if err := query.Find(&artifacts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load artifacts"})
		return
	}

	items := make([]artifactResponse, 0, len(artifacts))
	for _, artifact := range artifacts {
		items = append(items, hydrateArtifactResponse(artifact))
	}

	c.JSON(http.StatusOK, gin.H{"artifacts": items})
}

func GetArtifact(c *gin.Context) {
	var artifact domain.Artifact
	if err := db.DB.First(&artifact, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "artifact not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"artifact": hydrateArtifactResponse(artifact)})
}

func GetArtifactVersions(c *gin.Context) {
	var artifact domain.Artifact
	if err := db.DB.First(&artifact, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "artifact not found"})
		return
	}

	var versions []domain.ArtifactVersion
	if err := db.DB.Where("artifact_id = ?", artifact.ID).Order("version desc").Find(&versions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load artifact versions"})
		return
	}

	items := make([]artifactVersionResponse, 0, len(versions))
	for _, version := range versions {
		items = append(items, hydrateArtifactVersionResponse(version))
	}

	c.JSON(http.StatusOK, gin.H{"versions": items})
}

func GetArtifactVersion(c *gin.Context) {
	var version domain.ArtifactVersion
	if err := db.DB.Where("artifact_id = ? AND version = ?", c.Param("id"), c.Param("version")).First(&version).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "artifact version not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"version": hydrateArtifactVersionResponse(version)})
}

func CreateArtifact(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		ChannelID string `json:"channel_id" binding:"required"`
		Title     string `json:"title" binding:"required"`
		Type      string `json:"type"`
		Status    string `json:"status"`
		Content   string `json:"content"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	artifact := domain.Artifact{
		ID:        "artifact-" + time.Now().UTC().Format("20060102150405.000000"),
		ChannelID: input.ChannelID,
		Title:     input.Title,
		Version:   1,
		Type:      defaultString(input.Type, "document"),
		Status:    defaultString(input.Status, "draft"),
		Content:   input.Content,
		Source:    "manual",
		CreatedBy: currentUser.ID,
		UpdatedBy: currentUser.ID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := db.DB.Create(&artifact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create artifact"})
		return
	}
	if err := createArtifactVersionSnapshot(artifact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to snapshot artifact version"})
		return
	}

	_ = broadcastArtifactRealtimeEvent("artifact.updated", artifact)

	c.JSON(http.StatusCreated, gin.H{"artifact": hydrateArtifactResponse(artifact)})
}

func UpdateArtifact(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var artifact domain.Artifact
	if err := db.DB.First(&artifact, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "artifact not found"})
		return
	}

	var input struct {
		Title   *string `json:"title"`
		Type    *string `json:"type"`
		Status  *string `json:"status"`
		Content *string `json:"content"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Title != nil {
		artifact.Title = *input.Title
	}
	if input.Type != nil {
		artifact.Type = *input.Type
	}
	if input.Status != nil {
		artifact.Status = *input.Status
	}
	if input.Content != nil {
		artifact.Content = *input.Content
	}
	artifact.Version++
	artifact.UpdatedBy = currentUser.ID
	artifact.UpdatedAt = time.Now().UTC()

	if err := db.DB.Save(&artifact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update artifact"})
		return
	}
	if err := createArtifactVersionSnapshot(artifact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to snapshot artifact version"})
		return
	}

	_ = broadcastArtifactRealtimeEvent("artifact.updated", artifact)

	c.JSON(http.StatusOK, gin.H{"artifact": hydrateArtifactResponse(artifact)})
}

func GenerateCanvasArtifact(c *gin.Context) {
	if AIGateway == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ai gateway is not configured"})
		return
	}

	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		ChannelID string `json:"channel_id" binding:"required"`
		Prompt    string `json:"prompt" binding:"required"`
		Title     string `json:"title"`
		Type      string `json:"type"`
		Provider  string `json:"provider"`
		Model     string `json:"model"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := llm.Request{
		Prompt:    "Generate a collaborative canvas artifact for the following request. Return clean, directly usable content.\n\n" + input.Prompt,
		ChannelID: input.ChannelID,
		Provider:  input.Provider,
		Model:     input.Model,
	}
	session, err := AIGateway.Stream(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	content, _, err := collectStreamOutput(c.Request.Context(), session)
	if err != nil {
		handleArtifactGenerationError(c, err)
		return
	}

	now := time.Now().UTC()
	artifact := domain.Artifact{
		ID:        "artifact-" + now.Format("20060102150405.000000"),
		ChannelID: input.ChannelID,
		Title:     defaultString(strings.TrimSpace(input.Title), "AI Canvas"),
		Version:   1,
		Type:      defaultString(input.Type, "document"),
		Status:    "live",
		Content:   strings.TrimSpace(content),
		Source:    "ai",
		Provider:  session.Provider,
		Model:     session.Model,
		CreatedBy: currentUser.ID,
		UpdatedBy: currentUser.ID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.DB.Create(&artifact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save artifact"})
		return
	}
	if err := createArtifactVersionSnapshot(artifact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to snapshot artifact version"})
		return
	}

	_ = broadcastArtifactRealtimeEvent("artifact.updated", artifact)

	c.JSON(http.StatusCreated, gin.H{"artifact": hydrateArtifactResponse(artifact)})
}

func broadcastArtifactRealtimeEvent(eventType string, artifact domain.Artifact) error {
	if RealtimeHub == nil {
		return nil
	}

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", artifact.ChannelID).Error; err != nil {
		return err
	}

	return RealtimeHub.Broadcast(realtime.Event{
		ID:          "evt_" + time.Now().Format("20060102150405.000000"),
		Type:        eventType,
		WorkspaceID: channel.WorkspaceID,
		ChannelID:   artifact.ChannelID,
		EntityID:    artifact.ID,
		TS:          time.Now().UTC().Format(time.RFC3339Nano),
		Payload:     hydrateArtifactResponse(artifact),
	})
}

func hydrateArtifactResponse(artifact domain.Artifact) artifactResponse {
	response := artifactResponse{Artifact: artifact}

	if artifact.CreatedBy != "" {
		var createdBy domain.User
		if err := db.DB.First(&createdBy, "id = ?", artifact.CreatedBy).Error; err == nil {
			enriched := enrichUser(createdBy)
			response.CreatedByUser = &enriched
		}
	}
	if artifact.UpdatedBy != "" {
		var updatedBy domain.User
		if err := db.DB.First(&updatedBy, "id = ?", artifact.UpdatedBy).Error; err == nil {
			enriched := enrichUser(updatedBy)
			response.UpdatedByUser = &enriched
		}
	}

	return response
}

func hydrateArtifactVersionResponse(version domain.ArtifactVersion) artifactVersionResponse {
	response := artifactVersionResponse{ArtifactVersion: version}

	if version.UpdatedBy != "" {
		var updatedBy domain.User
		if err := db.DB.First(&updatedBy, "id = ?", version.UpdatedBy).Error; err == nil {
			enriched := enrichUser(updatedBy)
			response.UpdatedByUser = &enriched
		}
	}

	return response
}

func createArtifactVersionSnapshot(artifact domain.Artifact) error {
	version := domain.ArtifactVersion{
		ArtifactID: artifact.ID,
		Version:    artifact.Version,
		Title:      artifact.Title,
		Type:       artifact.Type,
		Status:     artifact.Status,
		Content:    artifact.Content,
		Source:     artifact.Source,
		Provider:   artifact.Provider,
		Model:      artifact.Model,
		UpdatedBy:  artifact.UpdatedBy,
		CreatedAt:  artifact.UpdatedAt,
	}

	return db.DB.Create(&version).Error
}

func handleArtifactGenerationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, context.Canceled):
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "artifact generation canceled"})
	default:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
