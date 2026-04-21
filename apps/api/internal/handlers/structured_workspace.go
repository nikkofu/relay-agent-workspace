package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

type workspaceListItemResponse struct {
	ID           uint         `json:"id"`
	Content      string       `json:"content"`
	Position     int          `json:"position"`
	IsCompleted  bool         `json:"is_completed"`
	AssignedTo   string       `json:"assigned_to,omitempty"`
	DueAt        *time.Time   `json:"due_at,omitempty"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
	AssignedUser *domain.User `json:"assigned_user,omitempty"`
	CreatedBy    string       `json:"created_by"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type workspaceListResponse struct {
	ID             string                      `json:"id"`
	WorkspaceID    string                      `json:"workspace_id"`
	ChannelID      string                      `json:"channel_id,omitempty"`
	Title          string                      `json:"title"`
	Description    string                      `json:"description"`
	ItemCount      int                         `json:"item_count"`
	CompletedCount int                         `json:"completed_count"`
	CreatedBy      string                      `json:"created_by"`
	CreatedByUser  *domain.User                `json:"created_by_user,omitempty"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
	Items          []workspaceListItemResponse `json:"items,omitempty"`
}

type toolRunLogResponse struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type toolRunResponse struct {
	ID              string               `json:"id"`
	ToolID          string               `json:"tool_id"`
	ToolName        string               `json:"tool_name"`
	ToolKey         string               `json:"tool_key"`
	Status          string               `json:"status"`
	Summary         string               `json:"summary"`
	Input           map[string]any       `json:"input"`
	TriggeredBy     string               `json:"triggered_by"`
	TriggeredByUser *domain.User         `json:"triggered_by_user,omitempty"`
	StartedAt       time.Time            `json:"started_at"`
	CompletedAt     *time.Time           `json:"completed_at,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	Logs            []toolRunLogResponse `json:"logs,omitempty"`
}

type artifactTemplate struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

func GetWorkspaceLists(c *gin.Context) {
	query := db.DB.Order("updated_at desc")
	if workspaceID := strings.TrimSpace(c.Query("workspace_id")); workspaceID != "" {
		query = query.Where("workspace_id = ?", workspaceID)
	}
	if channelID := strings.TrimSpace(c.Query("channel_id")); channelID != "" {
		query = query.Where("channel_id = ?", channelID)
	}

	var lists []domain.WorkspaceList
	if err := query.Find(&lists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load workspace lists"})
		return
	}

	items := make([]workspaceListResponse, 0, len(lists))
	for _, list := range lists {
		items = append(items, hydrateWorkspaceList(list, false))
	}

	c.JSON(http.StatusOK, gin.H{"lists": items})
}

func CreateWorkspaceList(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		WorkspaceID string `json:"workspace_id" binding:"required"`
		ChannelID   string `json:"channel_id"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	list := domain.WorkspaceList{
		ID:          "list-" + now.Format("20060102150405.000000"),
		WorkspaceID: input.WorkspaceID,
		ChannelID:   strings.TrimSpace(input.ChannelID),
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		CreatedBy:   currentUser.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.DB.Create(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workspace list"})
		return
	}

	broadcastStructuredRealtime("list.updated", list.ID, gin.H{"list": hydrateWorkspaceList(list, true)})
	c.JSON(http.StatusCreated, gin.H{"list": hydrateWorkspaceList(list, true)})
}

func GetWorkspaceList(c *gin.Context) {
	var list domain.WorkspaceList
	if err := db.DB.First(&list, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace list not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": hydrateWorkspaceList(list, true)})
}

func UpdateWorkspaceList(c *gin.Context) {
	var list domain.WorkspaceList
	if err := db.DB.First(&list, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace list not found"})
		return
	}

	var input struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Title != nil {
		list.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		list.Description = strings.TrimSpace(*input.Description)
	}
	list.UpdatedAt = time.Now().UTC()
	if err := db.DB.Save(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update workspace list"})
		return
	}

	broadcastStructuredRealtime("list.updated", list.ID, gin.H{"list": hydrateWorkspaceList(list, true)})
	c.JSON(http.StatusOK, gin.H{"list": hydrateWorkspaceList(list, true)})
}

func DeleteWorkspaceList(c *gin.Context) {
	var list domain.WorkspaceList
	if err := db.DB.First(&list, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace list not found"})
		return
	}

	if err := db.DB.Where("list_id = ?", list.ID).Delete(&domain.WorkspaceListItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete workspace list items"})
		return
	}
	if err := db.DB.Delete(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete workspace list"})
		return
	}

	broadcastStructuredRealtime("list.deleted", list.ID, gin.H{"list_id": list.ID})
	c.JSON(http.StatusOK, gin.H{"deleted": true, "list_id": list.ID})
}

func CreateWorkspaceListItem(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var list domain.WorkspaceList
	if err := db.DB.First(&list, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace list not found"})
		return
	}

	var input struct {
		Content    string     `json:"content" binding:"required"`
		AssignedTo string     `json:"assigned_to"`
		DueAt      *time.Time `json:"due_at"`
		Position   *int       `json:"position"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var maxPosition int
	db.DB.Model(&domain.WorkspaceListItem{}).Where("list_id = ?", list.ID).Select("coalesce(max(position), 0)").Scan(&maxPosition)

	now := time.Now().UTC()
	position := maxPosition + 1
	if input.Position != nil && *input.Position > 0 {
		position = *input.Position
	}
	item := domain.WorkspaceListItem{
		ListID:      list.ID,
		Content:     strings.TrimSpace(input.Content),
		Position:    position,
		AssignedTo:  strings.TrimSpace(input.AssignedTo),
		DueAt:       input.DueAt,
		CreatedBy:   currentUser.ID,
		IsCompleted: false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workspace list item"})
		return
	}

	broadcastStructuredRealtime("list.item.updated", list.ID, gin.H{"list_id": list.ID, "item": hydrateWorkspaceListItem(item)})
	c.JSON(http.StatusCreated, gin.H{"item": hydrateWorkspaceListItem(item)})
}

func UpdateWorkspaceListItem(c *gin.Context) {
	var item domain.WorkspaceListItem
	if err := db.DB.First(&item, "id = ? AND list_id = ?", c.Param("itemId"), c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace list item not found"})
		return
	}

	var input struct {
		Content     *string    `json:"content"`
		AssignedTo  *string    `json:"assigned_to"`
		DueAt       **time.Time `json:"due_at"`
		Position    *int       `json:"position"`
		IsCompleted *bool      `json:"is_completed"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Content != nil {
		item.Content = strings.TrimSpace(*input.Content)
	}
	if input.AssignedTo != nil {
		item.AssignedTo = strings.TrimSpace(*input.AssignedTo)
	}
	if input.DueAt != nil {
		item.DueAt = *input.DueAt
	}
	if input.Position != nil && *input.Position > 0 {
		item.Position = *input.Position
	}
	if input.IsCompleted != nil {
		item.IsCompleted = *input.IsCompleted
		if item.IsCompleted {
			now := time.Now().UTC()
			item.CompletedAt = &now
		} else {
			item.CompletedAt = nil
		}
	}
	item.UpdatedAt = time.Now().UTC()
	if err := db.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update workspace list item"})
		return
	}

	broadcastStructuredRealtime("list.item.updated", item.ListID, gin.H{"list_id": item.ListID, "item": hydrateWorkspaceListItem(item)})
	c.JSON(http.StatusOK, gin.H{"item": hydrateWorkspaceListItem(item)})
}

func DeleteWorkspaceListItem(c *gin.Context) {
	var item domain.WorkspaceListItem
	if err := db.DB.First(&item, "id = ? AND list_id = ?", c.Param("itemId"), c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace list item not found"})
		return
	}

	if err := db.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete workspace list item"})
		return
	}

	broadcastStructuredRealtime("list.item.deleted", item.ListID, gin.H{"list_id": item.ListID, "item_id": item.ID})
	c.JSON(http.StatusOK, gin.H{"deleted": true, "item_id": item.ID, "list_id": item.ListID})
}

func GetToolRuns(c *gin.Context) {
	query := db.DB.Order("started_at desc")
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	if toolID := strings.TrimSpace(c.Query("tool_id")); toolID != "" {
		query = query.Where("tool_id = ?", toolID)
	}

	var runs []domain.ToolRun
	if err := query.Find(&runs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load tool runs"})
		return
	}

	items := make([]toolRunResponse, 0, len(runs))
	for _, run := range runs {
		items = append(items, hydrateToolRun(run, false))
	}

	c.JSON(http.StatusOK, gin.H{"runs": items})
}

func GetToolRun(c *gin.Context) {
	var run domain.ToolRun
	if err := db.DB.First(&run, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool run not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"run": hydrateToolRun(run, true)})
}

func ExecuteTool(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var tool domain.ToolDefinition
	if err := db.DB.First(&tool, "id = ? AND is_enabled = ?", c.Param("id"), true).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool not found"})
		return
	}

	var input struct {
		Input map[string]any `json:"input"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rawInput := "{}"
	if len(input.Input) > 0 {
		if bytes, err := json.Marshal(input.Input); err == nil {
			rawInput = string(bytes)
		}
	}

	now := time.Now().UTC()
	run := domain.ToolRun{
		ID:          "toolrun-" + now.Format("20060102150405.000000"),
		ToolID:      tool.ID,
		TriggeredBy: currentUser.ID,
		Status:      "success",
		Input:       rawInput,
		Summary:     "Executed " + tool.Name,
		StartedAt:   now,
		CompletedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.DB.Create(&run).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to execute tool"})
		return
	}

	logs := []domain.ToolRunLog{
		{ToolRunID: run.ID, Level: "info", Message: "Tool execution requested", CreatedAt: now},
		{ToolRunID: run.ID, Level: "info", Message: "Tool execution completed", CreatedAt: now.Add(250 * time.Millisecond)},
	}
	for _, log := range logs {
		if err := db.DB.Create(&log).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store tool run log"})
			return
		}
	}

	response := hydrateToolRun(run, true)
	broadcastStructuredRealtime("tool.run.updated", run.ID, gin.H{"run": response})
	c.JSON(http.StatusCreated, gin.H{"run": response})
}

func GetArtifactTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"templates": artifactTemplateCatalog()})
}

func CreateArtifactFromTemplate(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		ChannelID  string `json:"channel_id" binding:"required"`
		TemplateID string `json:"template_id" binding:"required"`
		Title      string `json:"title"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, ok := findArtifactTemplate(input.TemplateID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "artifact template not found"})
		return
	}

	now := time.Now().UTC()
	artifact := domain.Artifact{
		ID:         "artifact-" + now.Format("20060102150405.000000"),
		ChannelID:  input.ChannelID,
		Title:      defaultString(strings.TrimSpace(input.Title), template.Title),
		Version:    1,
		Type:       template.Type,
		Status:     "draft",
		Content:    template.Content,
		Source:     "template",
		TemplateID: template.ID,
		CreatedBy:  currentUser.ID,
		UpdatedBy:  currentUser.ID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := db.DB.Create(&artifact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create artifact from template"})
		return
	}
	if err := createArtifactVersionSnapshot(artifact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to snapshot artifact version"})
		return
	}

	response := hydrateArtifactResponse(artifact)
	broadcastStructuredRealtime("artifact.updated", artifact.ID, gin.H{"artifact": response})
	c.JSON(http.StatusCreated, gin.H{"artifact": response})
}

func buildVirtualArtifactResponse(channelID string) artifactResponse {
	currentUser, _ := getCurrentUser()
	now := time.Now().UTC()
	return artifactResponse{
		Artifact: domain.Artifact{
			ID:         "new-doc",
			ChannelID:  channelID,
			Title:      "Untitled document",
			Version:    0,
			Type:       "document",
			Status:     "draft",
			Content:    "",
			Source:     "template",
			TemplateID: "blank-document",
			CreatedBy:  currentUser.ID,
			UpdatedBy:  currentUser.ID,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		CreatedByUser: optionalEnrichedUser(currentUser.ID),
		UpdatedByUser: optionalEnrichedUser(currentUser.ID),
		IsVirtual:     true,
		TemplateID:    "blank-document",
	}
}

func hydrateWorkspaceList(list domain.WorkspaceList, includeItems bool) workspaceListResponse {
	items := loadWorkspaceListItems(list.ID)
	response := workspaceListResponse{
		ID:             list.ID,
		WorkspaceID:    list.WorkspaceID,
		ChannelID:      list.ChannelID,
		Title:          list.Title,
		Description:    list.Description,
		ItemCount:      len(items),
		CompletedCount: countCompletedListItems(items),
		CreatedBy:      list.CreatedBy,
		CreatedByUser:  optionalEnrichedUser(list.CreatedBy),
		CreatedAt:      list.CreatedAt,
		UpdatedAt:      list.UpdatedAt,
	}
	if includeItems {
		response.Items = items
	}
	return response
}

func loadWorkspaceListItems(listID string) []workspaceListItemResponse {
	var rows []domain.WorkspaceListItem
	_ = db.DB.Where("list_id = ?", listID).Order("position asc, id asc").Find(&rows).Error

	items := make([]workspaceListItemResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, hydrateWorkspaceListItem(row))
	}
	return items
}

func hydrateWorkspaceListItem(item domain.WorkspaceListItem) workspaceListItemResponse {
	return workspaceListItemResponse{
		ID:           item.ID,
		Content:      item.Content,
		Position:     item.Position,
		IsCompleted:  item.IsCompleted,
		AssignedTo:   item.AssignedTo,
		DueAt:        item.DueAt,
		CompletedAt:  item.CompletedAt,
		AssignedUser: optionalEnrichedUser(item.AssignedTo),
		CreatedBy:    item.CreatedBy,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

func countCompletedListItems(items []workspaceListItemResponse) int {
	count := 0
	for _, item := range items {
		if item.IsCompleted {
			count++
		}
	}
	return count
}

func hydrateToolRun(run domain.ToolRun, includeLogs bool) toolRunResponse {
	var tool domain.ToolDefinition
	_ = db.DB.First(&tool, "id = ?", run.ToolID).Error

	input := map[string]any{}
	if strings.TrimSpace(run.Input) != "" {
		_ = json.Unmarshal([]byte(run.Input), &input)
	}

	response := toolRunResponse{
		ID:              run.ID,
		ToolID:          run.ToolID,
		ToolName:        tool.Name,
		ToolKey:         tool.Key,
		Status:          run.Status,
		Summary:         run.Summary,
		Input:           input,
		TriggeredBy:     run.TriggeredBy,
		TriggeredByUser: optionalEnrichedUser(run.TriggeredBy),
		StartedAt:       run.StartedAt,
		CompletedAt:     run.CompletedAt,
		CreatedAt:       run.CreatedAt,
		UpdatedAt:       run.UpdatedAt,
	}
	if includeLogs {
		response.Logs = loadToolRunLogs(run.ID)
	}
	return response
}

func loadToolRunLogs(runID string) []toolRunLogResponse {
	var logs []domain.ToolRunLog
	_ = db.DB.Where("tool_run_id = ?", runID).Order("created_at asc, id asc").Find(&logs).Error

	items := make([]toolRunLogResponse, 0, len(logs))
	for _, log := range logs {
		items = append(items, toolRunLogResponse{
			Level:     log.Level,
			Message:   log.Message,
			CreatedAt: log.CreatedAt,
		})
	}
	return items
}

func artifactTemplateCatalog() []artifactTemplate {
	templates := []artifactTemplate{
		{
			ID:          "blank-document",
			Title:       "Untitled document",
			Type:        "document",
			Description: "Start from a clean collaborative document.",
			Content:     "",
		},
		{
			ID:          "launch-checklist",
			Title:       "Launch Checklist",
			Type:        "document",
			Description: "A simple release readiness checklist for cross-functional teams.",
			Content:     "- [ ] Final QA sign-off\n- [ ] Release notes approved\n- [ ] Rollout owner assigned\n- [ ] Monitoring links attached",
		},
		{
			ID:          "incident-retro",
			Title:       "Incident Retro",
			Type:        "document",
			Description: "Capture timeline, impact, and follow-up actions after an incident.",
			Content:     "## Summary\n\n## Timeline\n\n## Impact\n\n## Action Items",
		},
		{
			ID:          "code-scratchpad",
			Title:       "Code Scratchpad",
			Type:        "code",
			Description: "A lightweight code-oriented canvas for experimenting with snippets.",
			Content:     "// Sketch here\n",
		},
	}
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Title < templates[j].Title
	})
	return templates
}

func findArtifactTemplate(id string) (artifactTemplate, bool) {
	for _, template := range artifactTemplateCatalog() {
		if template.ID == id {
			return template, true
		}
	}
	return artifactTemplate{}, false
}

func optionalEnrichedUser(userID string) *domain.User {
	if strings.TrimSpace(userID) == "" {
		return nil
	}

	var user domain.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		return nil
	}
	enriched := enrichUser(user)
	return &enriched
}

func broadcastStructuredRealtime(eventType, entityID string, payload any) {
	if RealtimeHub == nil {
		return
	}

	now := time.Now().UTC()
	_ = RealtimeHub.Broadcast(realtime.Event{
		ID:          "evt_" + now.Format("20060102150405.000000"),
		Type:        eventType,
		WorkspaceID: primaryWorkspaceID(),
		EntityID:    entityID,
		TS:          now.Format(time.RFC3339Nano),
		Payload:     payload,
	})
}
