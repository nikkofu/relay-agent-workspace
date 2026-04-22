package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
	"github.com/nikkofu/relay-agent-workspace/api/internal/knowledge"
	"github.com/nikkofu/relay-agent-workspace/api/internal/llm"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

func ListKnowledgeEntities(c *gin.Context) {
	entities, err := knowledge.ListEntities(db.DB, c.Query("workspace_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load knowledge entities"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entities": entities})
}

func CreateKnowledgeEntity(c *gin.Context) {
	var input knowledge.CreateEntityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity, err := knowledge.CreateEntity(db.DB, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := broadcastKnowledgeEvent("knowledge.entity.created", entity.WorkspaceID, "", entity.ID, gin.H{"entity": entity}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast knowledge entity event"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"entity": entity})
}

func GetKnowledgeEntity(c *gin.Context) {
	entity, err := knowledge.GetEntity(db.DB, c.Param("id"))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

func UpdateKnowledgeEntity(c *gin.Context) {
	var input knowledge.UpdateEntityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity, err := knowledge.UpdateEntity(db.DB, c.Param("id"), input)
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	if err := broadcastKnowledgeEvent("knowledge.entity.updated", entity.WorkspaceID, "", entity.ID, gin.H{"entity": entity}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast knowledge entity event"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

func ListKnowledgeEntityRefs(c *gin.Context) {
	if _, err := knowledge.GetEntity(db.DB, c.Param("id")); err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	refs, err := knowledge.ListEntityRefs(db.DB, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load knowledge refs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"refs": refs})
}

func AddKnowledgeEntityRef(c *gin.Context) {
	var input knowledge.AddEntityRefInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ref, err := knowledge.AddEntityRef(db.DB, c.Param("id"), input)
	if err != nil {
		handleKnowledgeError(c, err, "failed to add knowledge ref")
		return
	}
	if err := broadcastKnowledgeEvent("knowledge.entity.ref.created", ref.WorkspaceID, "", ref.EntityID, gin.H{"ref": ref}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast knowledge ref event"})
		return
	}
	_ = emitKnowledgeEntitySpikeAlerts(ref.EntityID, "")
	_ = emitKnowledgeTrendingChanged(ref.WorkspaceID)
	c.JSON(http.StatusCreated, gin.H{"ref": ref})
}

func ListKnowledgeEntityTimeline(c *gin.Context) {
	if _, err := knowledge.GetEntity(db.DB, c.Param("id")); err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	events, err := knowledge.ListEntityTimeline(db.DB, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load knowledge timeline"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func AddKnowledgeEntityEvent(c *gin.Context) {
	var input knowledge.AddEntityEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if currentUser, err := getCurrentUser(); err == nil && strings.TrimSpace(input.ActorUserID) == "" {
		input.ActorUserID = currentUser.ID
	}
	event, err := knowledge.AppendEntityEvent(db.DB, c.Param("id"), input)
	if err != nil {
		handleKnowledgeError(c, err, "failed to add knowledge event")
		return
	}
	if err := broadcastKnowledgeEvent("knowledge.event.created", event.WorkspaceID, "", event.EntityID, gin.H{"event": event}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast knowledge event"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"event": event})
}

func IngestKnowledgeEvent(c *gin.Context) {
	var input knowledge.IngestEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if currentUser, err := getCurrentUser(); err == nil && strings.TrimSpace(input.ActorUserID) == "" {
		input.ActorUserID = currentUser.ID
	}
	event, err := knowledge.IngestEvent(db.DB, input)
	if err != nil {
		handleKnowledgeError(c, err, "failed to ingest knowledge event")
		return
	}
	if err := broadcastKnowledgeEvent("knowledge.event.created", event.WorkspaceID, "", event.EntityID, gin.H{"event": event}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast knowledge event"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"event": event})
}

func ListKnowledgeEntityLinks(c *gin.Context) {
	if _, err := knowledge.GetEntity(db.DB, c.Param("id")); err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	links, err := knowledge.ListEntityLinks(db.DB, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load knowledge links"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"links": links})
}

func AddKnowledgeEntityLink(c *gin.Context) {
	var input knowledge.AddEntityLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	link, err := knowledge.AddEntityLink(db.DB, input)
	if err != nil {
		handleKnowledgeError(c, err, "failed to add knowledge link")
		return
	}
	if err := broadcastKnowledgeEvent("knowledge.link.created", link.WorkspaceID, "", link.FromEntityID, gin.H{"link": link}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast knowledge link event"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"link": link})
}

func GetKnowledgeEntityGraph(c *gin.Context) {
	graph, err := knowledge.BuildEntityGraph(db.DB, c.Param("id"))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"graph": graph})
}

func GetChannelKnowledgeContext(c *gin.Context) {
	context, err := knowledge.GetChannelKnowledgeContext(db.DB, c.Param("id"), parseLimit(c.Query("limit"), 20))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load channel knowledge context"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"context": context})
}

func GetChannelKnowledgeSummary(c *gin.Context) {
	summary, err := knowledge.GetChannelKnowledgeSummary(
		db.DB,
		c.Param("id"),
		parseLimit(c.Query("limit"), 5),
		parseWindowDays(c.Query("days"), 7),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load channel knowledge summary"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

func SuggestKnowledgeEntities(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}

	suggestions, err := knowledge.SuggestEntities(db.DB, knowledge.SuggestEntitiesParams{
		Query:       query,
		ChannelID:   c.Query("channel_id"),
		WorkspaceID: c.Query("workspace_id"),
		Limit:       parseLimit(c.Query("limit"), 8),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load knowledge entity suggestions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":       query,
		"suggestions": suggestions,
	})
}

func MatchKnowledgeEntitiesInText(c *gin.Context) {
	var input struct {
		WorkspaceID string `json:"workspace_id"`
		Text        string `json:"text"`
		Limit       int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(input.Text) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
		return
	}

	matches, err := knowledge.MatchEntitiesInText(db.DB, knowledge.MatchEntitiesInput{
		WorkspaceID: input.WorkspaceID,
		Text:        input.Text,
		Limit:       input.Limit,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to match knowledge entities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"matches": matches})
}

func GetMyFollowedKnowledgeEntities(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	items, err := knowledge.ListFollowedEntities(db.DB, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load followed knowledge entities"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func GetMyFollowedKnowledgeStats(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	stats, err := knowledge.GetFollowedEntityStats(db.DB, currentUser.ID, time.Now().UTC())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load followed knowledge stats"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func FollowKnowledgeEntity(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	follow, err := knowledge.FollowEntity(db.DB, c.Param("id"), currentUser.ID)
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	_ = emitFollowedStatsChanged(currentUser.ID, follow.WorkspaceID)
	c.JSON(http.StatusOK, gin.H{"follow": follow, "is_following": true})
}

func PatchMyKnowledgeFollow(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		NotificationLevel string `json:"notification_level"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	follow, err := knowledge.UpdateFollowNotificationLevel(db.DB, c.Param("id"), currentUser.ID, input.NotificationLevel)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge follow not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = emitFollowedStatsChanged(currentUser.ID, follow.WorkspaceID)
	c.JSON(http.StatusOK, gin.H{"follow": follow, "is_following": true})
}

func PatchMyKnowledgeFollowsBulk(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		EntityIDs         []string `json:"entity_ids"`
		NotificationLevel string   `json:"notification_level"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, err := knowledge.BulkUpdateFollowNotificationLevels(db.DB, currentUser.ID, input.EntityIDs, input.NotificationLevel)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	workspaceIDs := map[string]struct{}{}
	for _, follow := range items {
		workspaceIDs[follow.WorkspaceID] = struct{}{}
	}
	for workspaceID := range workspaceIDs {
		_ = emitFollowedStatsChanged(currentUser.ID, workspaceID)
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func UnfollowKnowledgeEntity(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err := knowledge.UnfollowEntity(db.DB, c.Param("id"), currentUser.ID); err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	_ = emitFollowedStatsChanged(currentUser.ID, "")
	c.JSON(http.StatusOK, gin.H{"entity_id": c.Param("id"), "is_following": false})
}

func ShareKnowledgeEntity(c *gin.Context) {
	share, err := knowledge.BuildSharedEntityLink(db.DB, c.Param("id"), strings.TrimSpace(os.Getenv("RELAY_APP_BASE_URL")))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"share": share})
}

func GenerateKnowledgeEntityBrief(c *gin.Context) {
	if AIGateway == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ai gateway is not configured"})
		return
	}
	var input struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
		Force    bool   `json:"force"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scopeID := c.Param("id")
	if !input.Force {
		if summary, err := getStoredSummary("knowledge_entity", scopeID); err == nil {
			c.JSON(http.StatusOK, gin.H{"brief": knowledge.EntityBrief{
				EntityID:    scopeID,
				Content:     summary.Content,
				Reasoning:   summary.Reasoning,
				Provider:    summary.Provider,
				Model:       summary.Model,
				GeneratedAt: summary.UpdatedAt,
				Cached:      true,
			}})
			return
		}
	}

	entity, refs, events, prompt, err := knowledge.BuildEntityBriefPrompt(db.DB, scopeID)
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	session, err := AIGateway.Stream(c.Request.Context(), llm.Request{
		Prompt:    prompt,
		ChannelID: entity.WorkspaceID,
		Provider:  input.Provider,
		Model:     input.Model,
	})
	if err != nil {
		handleSummaryGenerationError(c, err)
		return
	}
	content, reasoning, err := collectStreamOutput(c.Request.Context(), session)
	if err != nil {
		handleSummaryGenerationError(c, err)
		return
	}

	now := time.Now().UTC()
	var lastRefAt *time.Time
	for _, ref := range refs {
		if lastRefAt == nil || ref.CreatedAt.After(*lastRefAt) {
			ts := ref.CreatedAt
			lastRefAt = &ts
		}
	}
	summary := domain.AISummary{
		ScopeType:     "knowledge_entity",
		ScopeID:       entity.ID,
		ChannelID:     entity.WorkspaceID,
		Provider:      session.Provider,
		Model:         session.Model,
		Content:       strings.TrimSpace(content),
		Reasoning:     strings.TrimSpace(reasoning),
		MessageCount:  len(refs),
		LastMessageAt: lastRefAt,
		UpdatedAt:     now,
	}
	if err := db.DB.Where("scope_type = ? AND scope_id = ?", summary.ScopeType, summary.ScopeID).Assign(summary).FirstOrCreate(&summary).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist entity brief"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"brief": knowledge.EntityBrief{
		EntityID:    entity.ID,
		WorkspaceID: entity.WorkspaceID,
		Title:       entity.Title,
		Content:     summary.Content,
		Reasoning:   summary.Reasoning,
		Provider:    summary.Provider,
		Model:       summary.Model,
		GeneratedAt: now,
		RefCount:    len(refs),
		EventCount:  len(events),
		LastRefAt:   lastRefAt,
		Cached:      false,
	}})
}

func GenerateMyKnowledgeWeeklyBrief(c *gin.Context) {
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
		WorkspaceID string `json:"workspace_id"`
		Provider    string `json:"provider"`
		Model       string `json:"model"`
		Force       bool   `json:"force"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	workspaceID := strings.TrimSpace(input.WorkspaceID)
	if workspaceID == "" {
		workspaceID = "ws-1"
	}
	scopeID := currentUser.ID + ":" + workspaceID + ":weekly"
	if !input.Force {
		if summary, err := getStoredSummary("knowledge_weekly", scopeID); err == nil {
			c.JSON(http.StatusOK, gin.H{"brief": knowledge.WeeklyBrief{
				UserID:      currentUser.ID,
				WorkspaceID: workspaceID,
				Content:     summary.Content,
				Reasoning:   summary.Reasoning,
				Provider:    summary.Provider,
				Model:       summary.Model,
				GeneratedAt: summary.UpdatedAt,
				Cached:      true,
			}})
			return
		}
	}
	now := time.Now().UTC()
	stats, followed, trending, prompt, err := knowledge.BuildWeeklyBriefPrompt(db.DB, currentUser.ID, workspaceID, now)
	if err != nil {
		handleKnowledgeError(c, err, "failed to build weekly brief")
		return
	}
	session, err := AIGateway.Stream(c.Request.Context(), llm.Request{
		Prompt:    prompt,
		ChannelID: workspaceID,
		Provider:  input.Provider,
		Model:     input.Model,
	})
	if err != nil {
		handleSummaryGenerationError(c, err)
		return
	}
	content, reasoning, err := collectStreamOutput(c.Request.Context(), session)
	if err != nil {
		handleSummaryGenerationError(c, err)
		return
	}
	summary := domain.AISummary{
		ScopeType:    "knowledge_weekly",
		ScopeID:      scopeID,
		ChannelID:    workspaceID,
		Provider:     session.Provider,
		Model:        session.Model,
		Content:      strings.TrimSpace(content),
		Reasoning:    strings.TrimSpace(reasoning),
		MessageCount: len(followed),
		UpdatedAt:    now,
	}
	if err := db.DB.Where("scope_type = ? AND scope_id = ?", summary.ScopeType, summary.ScopeID).Assign(summary).FirstOrCreate(&summary).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist weekly brief"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"brief": knowledge.WeeklyBrief{
		UserID:      currentUser.ID,
		WorkspaceID: workspaceID,
		Content:     summary.Content,
		Reasoning:   summary.Reasoning,
		Provider:    summary.Provider,
		Model:       summary.Model,
		GeneratedAt: now,
		Stats:       stats,
		Trending:    trending,
		Followed:    followed,
		Cached:      false,
	}})
}

func GetKnowledgeEntityActivityBackfillStatus(c *gin.Context) {
	status, err := knowledge.GetEntityActivityBackfillStatus(db.DB, c.Param("id"))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}

func BackfillKnowledgeEntityActivity(c *gin.Context) {
	status, refs, err := knowledge.BackfillEntityActivity(db.DB, c.Param("id"))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	for _, ref := range refs {
		_ = broadcastKnowledgeEvent("knowledge.entity.ref.created", ref.WorkspaceID, "", ref.EntityID, gin.H{"ref": ref, "backfilled": true})
	}
	if status.CreatedRefCount > 0 {
		_ = emitKnowledgeTrendingChanged(status.WorkspaceID)
	}
	c.JSON(http.StatusOK, gin.H{"status": status, "created_refs": refs})
}

func GetWorkspaceSettings(c *gin.Context) {
	workspaceID, err := resolveWorkspaceID(c.Query("workspace_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	settings, err := knowledge.GetWorkspaceKnowledgeSettings(db.DB, workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load workspace settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func PatchWorkspaceSettings(c *gin.Context) {
	var input struct {
		WorkspaceID          string `json:"workspace_id"`
		SpikeThreshold       int    `json:"spike_threshold"`
		SpikeCooldownMinutes int    `json:"spike_cooldown_minutes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspaceID, err := resolveWorkspaceID(input.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	settings, err := knowledge.UpdateWorkspaceKnowledgeSettings(db.DB, workspaceID, input.SpikeThreshold, input.SpikeCooldownMinutes)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func GetKnowledgeEntityActivity(c *gin.Context) {
	activity, err := knowledge.GetEntityActivity(db.DB, c.Param("id"), parseWindowDays(c.Query("days"), 30), time.Now().UTC())
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"activity": activity})
}

func GetKnowledgeTrending(c *gin.Context) {
	workspaceID, err := resolveWorkspaceID(c.Query("workspace_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	items, err := knowledge.GetTrendingEntities(db.DB, knowledge.TrendingEntitiesParams{
		WorkspaceID: workspaceID,
		Days:        parseWindowDays(c.Query("days"), 7),
		Limit:       parseLimit(c.Query("limit"), 5),
		Now:         time.Now().UTC(),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load knowledge trending"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func SearchMessagesByEntity(c *gin.Context) {
	entityID := strings.TrimSpace(c.Query("entity_id"))
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_id is required"})
		return
	}

	matches, err := knowledge.FindMessagesByEntity(db.DB, knowledge.EntityMessageSearchParams{
		EntityID:  entityID,
		ChannelID: c.Query("channel_id"),
		Limit:     parseLimit(c.Query("limit"), 20),
	})
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}

	items := make([]gin.H, 0, len(matches))
	for _, match := range matches {
		message := match.Message
		refreshed, err := refreshMessageMetadata(message.ID)
		if err == nil && refreshed != nil {
			message = *refreshed
		}
		items = append(items, gin.H{
			"id":            message.ID,
			"channel_id":    message.ChannelID,
			"user_id":       message.UserID,
			"content":       message.Content,
			"thread_id":     message.ThreadID,
			"created_at":    message.CreatedAt,
			"metadata":      message.Metadata,
			"snippet":       buildSearchSnippet(message.Content, match.EntityTitle),
			"entity_id":     match.EntityID,
			"entity_title":  match.EntityTitle,
			"match_sources": match.MatchSources,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id":  entityID,
		"channel_id": strings.TrimSpace(c.Query("channel_id")),
		"messages":   items,
	})
}

func GetKnowledgeEntityHover(c *gin.Context) {
	entity, err := knowledge.GetEntity(db.DB, c.Param("id"))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}

	hover, err := knowledge.GetEntityHoverSummary(
		db.DB,
		c.Param("id"),
		c.Query("channel_id"),
		parseWindowDays(c.Query("days"), 7),
	)
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entity": entity,
		"hover":  hover,
	})
}

func GetChannelKnowledgeDigest(c *gin.Context) {
	digest, err := knowledge.BuildChannelKnowledgeDigest(
		db.DB,
		c.Param("id"),
		c.DefaultQuery("window", "weekly"),
		parseLimit(c.Query("limit"), 5),
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to build channel knowledge digest"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"digest": digest})
}

func PublishChannelKnowledgeDigest(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		Window string `json:"window"`
		Limit  int    `json:"limit"`
		Pin    bool   `json:"pin"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	published, err := knowledge.PublishChannelDigest(db.DB, knowledge.PublishChannelDigestInput{
		ChannelID:  c.Param("id"),
		UserID:     currentUser.ID,
		Window:     defaultString(strings.TrimSpace(input.Window), "weekly"),
		Limit:      parseLimit(strconv.Itoa(input.Limit), 5),
		Pin:        input.Pin,
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to build channel knowledge digest"})
		return
	}

	if RealtimeHub != nil {
		if err := broadcastRealtimeEvent("message.created", published.Message, published.Message); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast digest message"})
			return
		}
		if err := broadcastKnowledgeDigestPublished(published); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast digest event"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": published.Message,
		"digest":  published.Digest,
	})
}

func GetChannelKnowledgeDigestSchedule(c *gin.Context) {
	schedule, err := knowledge.GetDigestSchedule(db.DB, c.Param("id"), time.Now().UTC())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load digest schedule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

func PutChannelKnowledgeDigestSchedule(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		Window     string `json:"window"`
		Timezone   string `json:"timezone"`
		DayOfWeek  int    `json:"day_of_week"`
		DayOfMonth int    `json:"day_of_month"`
		Hour       int    `json:"hour"`
		Minute     int    `json:"minute"`
		Limit      int    `json:"limit"`
		Pin        bool   `json:"pin"`
		IsEnabled  bool   `json:"is_enabled"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schedule, err := knowledge.UpsertDigestSchedule(db.DB, knowledge.UpsertDigestScheduleInput{
		ChannelID:  c.Param("id"),
		CreatedBy:  currentUser.ID,
		Window:     input.Window,
		Timezone:   input.Timezone,
		DayOfWeek:  input.DayOfWeek,
		DayOfMonth: input.DayOfMonth,
		Hour:       input.Hour,
		Minute:     input.Minute,
		Limit:      input.Limit,
		Pin:        input.Pin,
		IsEnabled:  input.IsEnabled,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

func PreviewChannelKnowledgeDigestSchedule(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		Window     string `json:"window"`
		Timezone   string `json:"timezone"`
		DayOfWeek  int    `json:"day_of_week"`
		DayOfMonth int    `json:"day_of_month"`
		Hour       int    `json:"hour"`
		Minute     int    `json:"minute"`
		Limit      int    `json:"limit"`
		Pin        bool   `json:"pin"`
		IsEnabled  bool   `json:"is_enabled"`
		Count      int    `json:"count"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preview, err := knowledge.PreviewDigestSchedule(db.DB, knowledge.UpsertDigestScheduleInput{
		ChannelID:  c.Param("id"),
		CreatedBy:  currentUser.ID,
		Window:     input.Window,
		Timezone:   input.Timezone,
		DayOfWeek:  input.DayOfWeek,
		DayOfMonth: input.DayOfMonth,
		Hour:       input.Hour,
		Minute:     input.Minute,
		Limit:      input.Limit,
		Pin:        input.Pin,
		IsEnabled:  input.IsEnabled,
	}, input.Count, time.Now().UTC())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preview": preview})
}

func DeleteChannelKnowledgeDigestSchedule(c *gin.Context) {
	if err := knowledge.DeleteDigestSchedule(db.DB, c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to delete digest schedule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true, "channel_id": c.Param("id")})
}

func GetKnowledgeInbox(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	items, err := knowledge.ListKnowledgeInbox(db.DB, knowledge.KnowledgeInboxParams{
		UserID: currentUser.ID,
		Scope:  c.DefaultQuery("scope", "all"),
		Limit:  parseLimit(c.Query("limit"), 20),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load knowledge inbox"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func GetKnowledgeInboxItem(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	item, err := knowledge.GetKnowledgeInboxItem(db.DB, currentUser.ID, c.Param("id"), parseLimit(c.Query("sample_limit"), 3))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge inbox item not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load knowledge inbox item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": gin.H{
		"id":              item.Item.ID,
		"channel":         item.Item.Channel,
		"message":         item.Item.Message,
		"digest":          item.Item.Digest,
		"is_read":         item.Item.IsRead,
		"occurred_at":     item.Item.OccurredAt,
		"entity_contexts": item.EntityContexts,
	}})
}

func handleKnowledgeError(c *gin.Context, err error, fallback string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "knowledge entity not found"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": fallback})
}

func handleKnowledgeNotFound(c *gin.Context, err error, message string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}

func resolveWorkspaceID(raw string) (string, error) {
	if workspaceID := strings.TrimSpace(raw); workspaceID != "" {
		return workspaceID, nil
	}

	currentUser, err := getCurrentUser()
	if err != nil {
		return "", err
	}

	var workspace domain.Workspace
	if err := db.DB.Where("organization_id = ?", currentUser.OrganizationID).Order("id asc").First(&workspace).Error; err == nil {
		return workspace.ID, nil
	}
	if err := db.DB.Order("id asc").First(&workspace).Error; err != nil {
		return "", err
	}
	return workspace.ID, nil
}

func parseLimit(raw string, fallback int) int {
	limit, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || limit <= 0 {
		return fallback
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func parseWindowDays(raw string, fallback int) int {
	days, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || days <= 0 {
		return fallback
	}
	if days > 30 {
		return 30
	}
	return days
}

func buildKnowledgeDigestMessageContent(digest knowledge.ChannelKnowledgeDigest) string {
	lines := []string{
		fmt.Sprintf("Knowledge digest (%s)", digest.Window),
		digest.Headline,
		digest.Summary,
	}
	for idx, movement := range digest.TopMovements {
		if idx >= 3 {
			break
		}
		lines = append(lines, fmt.Sprintf("- %s: %d recent refs (%+d vs previous window)", movement.EntityTitle, movement.RecentRefCount, movement.Delta))
	}
	return strings.Join(lines, "\n")
}

func broadcastKnowledgeDigestPublished(published knowledge.PublishedDigest) error {
	if RealtimeHub == nil {
		return nil
	}
	return RealtimeHub.Broadcast(realtime.Event{
		ID:          ids.NewPrefixedUUID("evt"),
		Type:        "knowledge.digest.published",
		WorkspaceID: published.Channel.WorkspaceID,
		ChannelID:   published.Channel.ID,
		TS:          time.Now().UTC().Format(time.RFC3339Nano),
		Payload: gin.H{
			"channel": published.Channel,
			"message": published.Message,
			"digest":  published.Digest,
		},
	})
}

func autoLinkKnowledgeForMessage(message domain.Message) {
	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", message.ChannelID).Error; err != nil {
		return
	}

	refs, err := knowledge.AutoLinkEntitiesForMessage(db.DB, channel.WorkspaceID, message.ID, message.Content)
	if err != nil {
		return
	}
	for _, ref := range refs {
		_ = broadcastKnowledgeEvent("knowledge.entity.ref.created", ref.WorkspaceID, channel.ID, ref.EntityID, gin.H{"ref": ref})
		_ = emitKnowledgeEntitySpikeAlerts(ref.EntityID, channel.ID)
	}
	if len(refs) > 0 {
		_ = emitKnowledgeTrendingChanged(channel.WorkspaceID)
	}
}

func autoLinkKnowledgeForFile(asset domain.FileAsset, content string) {
	var channel domain.Channel
	if asset.ChannelID == "" || db.DB.First(&channel, "id = ?", asset.ChannelID).Error != nil {
		return
	}

	refs, err := knowledge.AutoLinkEntitiesForFile(db.DB, channel.WorkspaceID, asset.ID, content)
	if err != nil {
		return
	}
	for _, ref := range refs {
		_ = broadcastKnowledgeEvent("knowledge.entity.ref.created", ref.WorkspaceID, channel.ID, ref.EntityID, gin.H{"ref": ref})
		_ = emitKnowledgeEntitySpikeAlerts(ref.EntityID, channel.ID)
	}
	if len(refs) > 0 {
		_ = emitKnowledgeTrendingChanged(channel.WorkspaceID)
	}
}

func emitKnowledgeEntitySpikeAlerts(entityID, channelID string) error {
	alerts, err := knowledge.DetectEntitySpikeAlerts(db.DB, entityID, channelID, time.Now().UTC())
	if err != nil {
		return err
	}
	for _, alert := range alerts {
		if err := broadcastKnowledgeEvent("knowledge.entity.activity.spiked", alert.Entity.WorkspaceID, alert.ChannelID, alert.Entity.ID, gin.H{
			"entity":              alert.Entity,
			"user_ids":            alert.UserIDs,
			"channel_id":          alert.ChannelID,
			"recent_ref_count":    alert.RecentRefCount,
			"previous_ref_count":  alert.PreviousRefCount,
			"delta":               alert.Delta,
			"related_channel_ids": alert.RelatedChannelIDs,
			"occurred_at":         alert.OccurredAt,
		}); err != nil {
			return err
		}
	}
	return nil
}

func emitKnowledgeTrendingChanged(workspaceID string) error {
	items, err := knowledge.GetTrendingEntities(db.DB, knowledge.TrendingEntitiesParams{
		WorkspaceID: workspaceID,
		Days:        7,
		Limit:       5,
		Now:         time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	return broadcastKnowledgeEvent("knowledge.trending.changed", workspaceID, "", "", gin.H{
		"workspace_id": workspaceID,
		"days":         7,
		"items":        items,
	})
}

func emitFollowedStatsChanged(userID, workspaceID string) error {
	if RealtimeHub == nil || strings.TrimSpace(userID) == "" {
		return nil
	}
	stats, err := knowledge.GetFollowedEntityStats(db.DB, userID, time.Now().UTC())
	if err != nil {
		return err
	}
	return RealtimeHub.Broadcast(realtime.Event{
		ID:          ids.NewPrefixedUUID("evt"),
		Type:        "knowledge.followed.stats.changed",
		WorkspaceID: workspaceID,
		EntityID:    userID,
		TS:          time.Now().UTC().Format(time.RFC3339Nano),
		Payload: gin.H{
			"user_id":      userID,
			"workspace_id": workspaceID,
			"stats":        stats,
		},
	})
}

func broadcastKnowledgeEvent(eventType, workspaceID, channelID, entityID string, payload any) error {
	if RealtimeHub == nil {
		return nil
	}

	return RealtimeHub.Broadcast(realtime.Event{
		ID:          "evt_" + time.Now().Format("20060102150405.000000"),
		Type:        eventType,
		WorkspaceID: workspaceID,
		ChannelID:   channelID,
		EntityID:    entityID,
		TS:          time.Now().UTC().Format(time.RFC3339Nano),
		Payload:     payload,
	})
}
