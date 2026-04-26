package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
)

type workspaceViewResponse struct {
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	ViewType         string         `json:"view_type"`
	Source           string         `json:"source"`
	PrimaryChannelID string         `json:"primary_channel_id,omitempty"`
	Filters          map[string]any `json:"filters"`
	Actions          []any          `json:"actions"`
	CreatedBy        string         `json:"created_by"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func hydrateWorkspaceView(view domain.WorkspaceView) workspaceViewResponse {
	var filters map[string]any
	json.Unmarshal([]byte(view.Filters), &filters)
	if filters == nil {
		filters = make(map[string]any)
	}

	var actions []any
	json.Unmarshal([]byte(view.Actions), &actions)
	if actions == nil {
		actions = make([]any, 0)
	}

	return workspaceViewResponse{
		ID:               view.ID,
		Title:            view.Title,
		ViewType:         view.ViewType,
		Source:           view.Source,
		PrimaryChannelID: view.PrimaryChannelID,
		Filters:          filters,
		Actions:          actions,
		CreatedBy:        view.CreatedBy,
		CreatedAt:        view.CreatedAt,
		UpdatedAt:        view.UpdatedAt,
	}
}

func ListWorkspaceViews(c *gin.Context) {
	viewType := c.Query("view_type")
	channelID := c.Query("primary_channel_id")
	source := c.Query("source")
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	query := db.DB.Model(&domain.WorkspaceView{}).Order("updated_at desc")

	if viewType != "" {
		query = query.Where("view_type = ?", viewType)
	}
	if channelID != "" {
		query = query.Where("primary_channel_id = ?", channelID)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}

	var views []domain.WorkspaceView
	if err := query.Limit(limit).Find(&views).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load workspace views"})
		return
	}

	results := make([]workspaceViewResponse, len(views))
	for i, v := range views {
		results[i] = hydrateWorkspaceView(v)
	}

	c.JSON(http.StatusOK, gin.H{
		"views": results,
	})
}

func CreateWorkspaceView(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		Title            string         `json:"title" binding:"required"`
		ViewType         string         `json:"view_type" binding:"required"`
		Source           string         `json:"source" binding:"required"`
		PrimaryChannelID string         `json:"primary_channel_id"`
		Filters          map[string]any `json:"filters"`
		Actions          []any          `json:"actions"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate view_type
	validViewTypes := map[string]bool{
		"list":             true,
		"calendar":         true,
		"search":           true,
		"report":           true,
		"form":             true,
		"channel_messages": true,
	}
	if !validViewTypes[input.ViewType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid view_type"})
		return
	}

	filtersJSON, _ := json.Marshal(input.Filters)
	if input.Filters == nil {
		filtersJSON = []byte("{}")
	}
	actionsJSON, _ := json.Marshal(input.Actions)
	if input.Actions == nil {
		actionsJSON = []byte("[]")
	}

	view := domain.WorkspaceView{
		ID:               ids.NewPrefixedUUID("view"),
		Title:            input.Title,
		ViewType:         input.ViewType,
		Source:           input.Source,
		PrimaryChannelID: input.PrimaryChannelID,
		Filters:          string(filtersJSON),
		Actions:          string(actionsJSON),
		CreatedBy:        currentUser.ID,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if err := db.DB.Create(&view).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workspace view"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"view": hydrateWorkspaceView(view)})
}

func GetWorkspaceView(c *gin.Context) {
	var view domain.WorkspaceView
	if err := db.DB.First(&view, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "view not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"view": hydrateWorkspaceView(view)})
}

func PatchWorkspaceView(c *gin.Context) {
	var view domain.WorkspaceView
	if err := db.DB.First(&view, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "view not found"})
		return
	}

	var input struct {
		Title            *string         `json:"title"`
		PrimaryChannelID *string         `json:"primary_channel_id"`
		Filters          *map[string]any `json:"filters"`
		Actions          *[]any          `json:"actions"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Title != nil {
		view.Title = *input.Title
	}
	if input.PrimaryChannelID != nil {
		view.PrimaryChannelID = *input.PrimaryChannelID
	}
	if input.Filters != nil {
		fJSON, _ := json.Marshal(*input.Filters)
		view.Filters = string(fJSON)
	}
	if input.Actions != nil {
		aJSON, _ := json.Marshal(*input.Actions)
		view.Actions = string(aJSON)
	}

	view.UpdatedAt = time.Now().UTC()
	if err := db.DB.Save(&view).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update workspace view"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"view": hydrateWorkspaceView(view)})
}
