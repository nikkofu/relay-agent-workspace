package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

func TestWorkspaceListLifecycleEndpoints(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "ops", Type: "public"})

	router := gin.New()
	router.GET("/api/v1/lists", GetWorkspaceLists)
	router.POST("/api/v1/lists", CreateWorkspaceList)
	router.GET("/api/v1/lists/:id", GetWorkspaceList)
	router.PATCH("/api/v1/lists/:id", UpdateWorkspaceList)
	router.DELETE("/api/v1/lists/:id", DeleteWorkspaceList)
	router.POST("/api/v1/lists/:id/items", CreateWorkspaceListItem)
	router.PATCH("/api/v1/lists/:id/items/:itemId", UpdateWorkspaceListItem)
	router.DELETE("/api/v1/lists/:id/items/:itemId", DeleteWorkspaceListItem)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lists", bytes.NewBufferString(`{"channel_id":"ch-1","title":"Launch Checklist","description":"Critical preflight items","user_id":"user-1"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on list create, got %d body=%s", rec.Code, rec.Body.String())
	}

	var createPayload struct {
		List struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			ItemCount   int    `json:"item_count"`
			WorkspaceID string `json:"workspace_id"`
			UserID      string `json:"user_id"`
		} `json:"list"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("failed to decode list create payload: %v", err)
	}
	if createPayload.List.ID == "" || createPayload.List.Title != "Launch Checklist" || createPayload.List.ItemCount != 0 || createPayload.List.WorkspaceID != "ws-1" || createPayload.List.UserID != "user-1" {
		t.Fatalf("unexpected list create payload: %#v", createPayload.List)
	}
	assertPrefixedUUID(t, createPayload.List.ID, "list")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/lists/"+createPayload.List.ID+"/items", bytes.NewBufferString(`{"content":"Freeze release notes","assigned_to":"user-1","due_at":"2026-04-22T10:00:00Z"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on item create, got %d body=%s", rec.Code, rec.Body.String())
	}

	var createItemPayload struct {
		Item struct {
			ID           uint   `json:"id"`
			ListID       string `json:"list_id"`
			Content      string `json:"content"`
			IsCompleted  bool   `json:"is_completed"`
			AssignedTo   string `json:"assigned_to"`
			UserID       string `json:"user_id"`
			AssignedUser *struct {
				ID string `json:"id"`
			} `json:"assigned_user"`
		} `json:"item"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createItemPayload); err != nil {
		t.Fatalf("failed to decode list item payload: %v", err)
	}
	if createItemPayload.Item.ID == 0 || createItemPayload.Item.Content != "Freeze release notes" || createItemPayload.Item.AssignedUser == nil || createItemPayload.Item.AssignedUser.ID != "user-1" || createItemPayload.Item.ListID != createPayload.List.ID || createItemPayload.Item.UserID != "user-1" {
		t.Fatalf("unexpected list item payload: %#v", createItemPayload.Item)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/lists/"+createPayload.List.ID+"/items/1", bytes.NewBufferString(`{"is_completed":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on item update, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/lists?channel_id=ch-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on list fetch, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Lists []struct {
			ID             string `json:"id"`
			ItemCount      int    `json:"item_count"`
			CompletedCount int    `json:"completed_count"`
			UserID         string `json:"user_id"`
		} `json:"lists"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode list payload: %v", err)
	}
	if len(listPayload.Lists) != 1 || listPayload.Lists[0].CompletedCount != 1 || listPayload.Lists[0].ItemCount != 1 || listPayload.Lists[0].UserID != "user-1" {
		t.Fatalf("unexpected lists payload: %#v", listPayload.Lists)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/lists/"+createPayload.List.ID, nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on list detail, got %d body=%s", rec.Code, rec.Body.String())
	}

	var detailPayload struct {
		List struct {
			ID    string `json:"id"`
			Items []struct {
				ID          uint   `json:"id"`
				ListID      string `json:"list_id"`
				Content     string `json:"content"`
				IsCompleted bool   `json:"is_completed"`
				UserID      string `json:"user_id"`
			} `json:"items"`
		} `json:"list"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &detailPayload); err != nil {
		t.Fatalf("failed to decode list detail payload: %v", err)
	}
	if len(detailPayload.List.Items) != 1 || !detailPayload.List.Items[0].IsCompleted || detailPayload.List.Items[0].ListID != createPayload.List.ID || detailPayload.List.Items[0].UserID != "user-1" {
		t.Fatalf("unexpected list detail payload: %#v", detailPayload.List)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/lists/"+createPayload.List.ID+"/items/1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on item delete, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/lists/"+createPayload.List.ID, nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on list delete, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestToolExecutionEndpoints(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.ToolDefinition{
		ID:          "tool-1",
		Name:        "Summarize Thread",
		Key:         "summarize-thread",
		Category:    "ai",
		Description: "Summarize a busy thread",
		Icon:        "sparkles",
		IsEnabled:   true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})

	router := gin.New()
	router.GET("/api/v1/tools", GetTools)
	router.GET("/api/v1/tools/runs", GetToolRuns)
	router.GET("/api/v1/tools/runs/:id", GetToolRun)
	router.POST("/api/v1/tools/:id/execute", ExecuteTool)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tools/tool-1/execute", bytes.NewBufferString(`{"channel_id":"ch-1","input":{"channel_id":"ch-1","thread_id":"msg-9"}}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on tool execute, got %d body=%s", rec.Code, rec.Body.String())
	}

	var createPayload struct {
		Run struct {
			ID         string `json:"id"`
			ToolID     string `json:"tool_id"`
			Status     string `json:"status"`
			Summary    string `json:"summary"`
			ToolName   string `json:"tool_name"`
			UserID     string `json:"user_id"`
			ChannelID  string `json:"channel_id"`
			FinishedAt string `json:"finished_at"`
			DurationMS int    `json:"duration_ms"`
			Logs       []struct {
				Level   string `json:"level"`
				Message string `json:"message"`
			} `json:"logs"`
		} `json:"run"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("failed to decode tool execute payload: %v", err)
	}
	if createPayload.Run.ID == "" || createPayload.Run.ToolID != "tool-1" || createPayload.Run.Status == "" || len(createPayload.Run.Logs) == 0 || createPayload.Run.UserID != "user-1" || createPayload.Run.ChannelID != "ch-1" || createPayload.Run.FinishedAt == "" || createPayload.Run.DurationMS < 0 {
		t.Fatalf("unexpected tool run create payload: %#v", createPayload.Run)
	}
	assertPrefixedUUID(t, createPayload.Run.ID, "toolrun")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tools/runs?channel_id=ch-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on tool runs list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Runs []struct {
			ID        string `json:"id"`
			ToolName  string `json:"tool_name"`
			Status    string `json:"status"`
			UserID    string `json:"user_id"`
			ChannelID string `json:"channel_id"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode tool runs list: %v", err)
	}
	if len(listPayload.Runs) != 1 || listPayload.Runs[0].ToolName != "Summarize Thread" || listPayload.Runs[0].UserID != "user-1" || listPayload.Runs[0].ChannelID != "ch-1" {
		t.Fatalf("unexpected tool runs list payload: %#v", listPayload.Runs)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tools/runs/"+createPayload.Run.ID, nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on tool run detail, got %d body=%s", rec.Code, rec.Body.String())
	}

	var detailPayload struct {
		Run struct {
			ID         string `json:"id"`
			ToolName   string `json:"tool_name"`
			Input      any    `json:"input"`
			UserID     string `json:"user_id"`
			ChannelID  string `json:"channel_id"`
			FinishedAt string `json:"finished_at"`
			DurationMS int    `json:"duration_ms"`
			Logs       []struct {
				Level string `json:"level"`
			} `json:"logs"`
		} `json:"run"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &detailPayload); err != nil {
		t.Fatalf("failed to decode tool run detail payload: %v", err)
	}
	if detailPayload.Run.ID != createPayload.Run.ID || len(detailPayload.Run.Logs) == 0 || detailPayload.Run.ToolName != "Summarize Thread" || detailPayload.Run.UserID != "user-1" || detailPayload.Run.ChannelID != "ch-1" || detailPayload.Run.FinishedAt == "" || detailPayload.Run.DurationMS < 0 {
		t.Fatalf("unexpected tool run detail payload: %#v", detailPayload.Run)
	}
}

func TestArtifactTemplatesAndVirtualNewDoc(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})

	router := gin.New()
	router.GET("/api/v1/artifacts/:id", GetArtifact)
	router.GET("/api/v1/artifacts/templates", GetArtifactTemplates)
	router.POST("/api/v1/artifacts/from-template", CreateArtifactFromTemplate)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/artifacts/new-doc?channel_id=ch-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on new-doc virtual artifact, got %d body=%s", rec.Code, rec.Body.String())
	}

	var virtualPayload struct {
		Artifact struct {
			ID         string `json:"id"`
			ChannelID  string `json:"channel_id"`
			UserID     string `json:"user_id"`
			Type       string `json:"type"`
			Title      string `json:"title"`
			IsVirtual  bool   `json:"is_virtual"`
			TemplateID string `json:"template_id"`
		} `json:"artifact"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &virtualPayload); err != nil {
		t.Fatalf("failed to decode virtual artifact payload: %v", err)
	}
	if virtualPayload.Artifact.ID != "new-doc" || !virtualPayload.Artifact.IsVirtual || virtualPayload.Artifact.TemplateID == "" || virtualPayload.Artifact.UserID != "user-1" {
		t.Fatalf("unexpected virtual artifact payload: %#v", virtualPayload.Artifact)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/artifacts/templates", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on template list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var templatesPayload struct {
		Templates []struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			Type    string `json:"type"`
			Content string `json:"content"`
		} `json:"templates"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &templatesPayload); err != nil {
		t.Fatalf("failed to decode templates payload: %v", err)
	}
	if len(templatesPayload.Templates) < 2 {
		t.Fatalf("expected multiple artifact templates, got %#v", templatesPayload.Templates)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/artifacts/from-template", bytes.NewBufferString(`{"channel_id":"ch-1","template_id":"launch-checklist","title":"Release Readiness"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create from template, got %d body=%s", rec.Code, rec.Body.String())
	}

	var createPayload struct {
		Artifact struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			ChannelID  string `json:"channel_id"`
			UserID     string `json:"user_id"`
			Source     string `json:"source"`
			TemplateID string `json:"template_id"`
			Content    string `json:"content"`
		} `json:"artifact"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("failed to decode create-from-template payload: %v", err)
	}
	if createPayload.Artifact.ID == "" || createPayload.Artifact.ChannelID != "ch-1" || createPayload.Artifact.UserID != "user-1" || createPayload.Artifact.Source != "template" || createPayload.Artifact.TemplateID != "launch-checklist" || createPayload.Artifact.Content == "" {
		t.Fatalf("unexpected create-from-template payload: %#v", createPayload.Artifact)
	}
	assertPrefixedUUID(t, createPayload.Artifact.ID, "artifact")
}

func TestChannelExecutionListItemPersistsMessageReference(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "ops", Type: "public"})
	db.DB.Create(&domain.WorkspaceList{ID: "list-1", WorkspaceID: "ws-1", ChannelID: "ch-1", CreatedBy: "user-1", Title: "Tasks"})

	router := gin.New()
	router.POST("/api/v1/lists/:id/items", CreateWorkspaceListItem)

	rec := httptest.NewRecorder()
	reqBody := `{"content":"Review release notes","source_message_id":"msg-123","source_channel_id":"ch-1","source_snippet":"The release notes are ready..."}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lists/list-1/items", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Item struct {
			ID              uint   `json:"id"`
			SourceMessageID string `json:"source_message_id"`
			SourceChannelID string `json:"source_channel_id"`
			SourceSnippet   string `json:"source_snippet"`
		} `json:"item"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if payload.Item.SourceMessageID != "msg-123" || payload.Item.SourceChannelID != "ch-1" || payload.Item.SourceSnippet != "The release notes are ready..." {
		t.Fatalf("unexpected source message fields: %#+v", payload.Item)
	}
}

func TestChannelExecutionToolExecuteSupportsMessageWriteback(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.ToolDefinition{ID: "tool-1", Name: "Summary Generator", IsEnabled: true})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "ops", Type: "public"})

	router := gin.New()
	router.POST("/api/v1/tools/:id/execute", ExecuteTool)

	rec := httptest.NewRecorder()
	reqBody := `{"channel_id":"ch-1","input":{"topic":"Phase 66"},"writeback_target":"message","writeback":{"channel_id":"ch-1"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tools/tool-1/execute", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Run struct {
			ID              string `json:"id"`
			WritebackTarget string `json:"writeback_target"`
		} `json:"run"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if payload.Run.WritebackTarget != "message" {
		t.Fatalf("expected writeback_target to be message, got %s", payload.Run.WritebackTarget)
	}

	// Verify message was created
	var count int64
	db.DB.Model(&domain.Message{}).Where("channel_id = ? AND user_id = ?", "ch-1", "user-1").Count(&count)
	if count == 0 {
		t.Fatal("expected a writeback message to be created")
	}
}

func TestChannelExecutionToolExecuteSupportsListItemWriteback(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.ToolDefinition{ID: "tool-1", Name: "Task Generator", IsEnabled: true})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "ops", Type: "public"})
	db.DB.Create(&domain.WorkspaceList{ID: "list-1", WorkspaceID: "ws-1", ChannelID: "ch-1", CreatedBy: "user-1", Title: "Tasks"})

	router := gin.New()
	router.POST("/api/v1/tools/:id/execute", ExecuteTool)

	rec := httptest.NewRecorder()
	reqBody := `{"input":{"task":"New Task"},"writeback_target":"list_item","writeback":{"list_id":"list-1"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tools/tool-1/execute", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var count int64
	db.DB.Model(&domain.WorkspaceListItem{}).Where("list_id = ?", "list-1").Count(&count)
	if count == 0 {
		t.Fatal("expected a writeback list item to be created")
	}
}
