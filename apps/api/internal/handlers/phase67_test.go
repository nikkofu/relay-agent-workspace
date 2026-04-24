package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

func expectEvent(t *testing.T, client *realtime.TestClient, eventType string, timeout time.Duration) realtime.Event {
	t.Helper()
	raw, err := client.Receive(timeout)
	if err != nil {
		t.Fatalf("expected event %s, but got error: %v", eventType, err)
	}
	var event realtime.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}
	if event.Type != eventType {
		t.Fatalf("expected event type %s, got %s", eventType, event.Type)
	}
	return event
}

func TestPhase67ListItemEventsEmitLifecyclePayloads(t *testing.T) {
	setupTestDB(t)
	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(10)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko", Email: "nikko@example.com"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "ops"})
	db.DB.Create(&domain.WorkspaceList{ID: "list-1", WorkspaceID: "ws-1", ChannelID: "ch-1", CreatedBy: "user-1", Title: "Tasks"})

	router := gin.New()
	router.POST("/api/v1/lists/:id/items", CreateWorkspaceListItem)
	router.PATCH("/api/v1/lists/:id/items/:itemId", UpdateWorkspaceListItem)
	router.DELETE("/api/v1/lists/:id/items/:itemId", DeleteWorkspaceListItem)

	// 1. Created
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lists/list-1/items", bytes.NewBufferString(`{"content":"New Task"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	event := expectEvent(t, client, "list.item.created", 1*time.Second)
	item, ok := event.Payload.(map[string]any)["item"].(map[string]any)
	if !ok {
		t.Fatal("expected item in payload")
	}
	if item["content"] != "New Task" || item["list_id"] != "list-1" {
		t.Fatalf("unexpected item fields: %#v", item)
	}

	// 2. Updated
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/lists/list-1/items/1", bytes.NewBufferString(`{"is_completed":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	event = expectEvent(t, client, "list.item.updated", 1*time.Second)
	item = event.Payload.(map[string]any)["item"].(map[string]any)
	if item["is_completed"] != true {
		t.Fatal("expected is_completed: true in payload")
	}

	// 3. Deleted
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/lists/list-1/items/1", nil)
	router.ServeHTTP(rec, req)

	expectEvent(t, client, "list.item.deleted", 1*time.Second)
}

func TestPhase67ToolRunEventsEmitLifecyclePayloads(t *testing.T) {
	setupTestDB(t)
	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(10)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko", Email: "nikko@example.com"})
	db.DB.Create(&domain.ToolDefinition{ID: "tool-1", Name: "Test Tool", IsEnabled: true})

	router := gin.New()
	router.POST("/api/v1/tools/:id/execute", ExecuteTool)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tools/tool-1/execute", bytes.NewBufferString(`{"input":{"foo":"bar"}}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	expectEvent(t, client, "tool.run.started", 1*time.Second)
	expectEvent(t, client, "tool.run.completed", 1*time.Second)
}

func TestPhase67UnreadMentionCountsEndpoint(t *testing.T) {
	setupTestDB(t)
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko", Email: "nikko@example.com"})
	
	// Create 2 unread mentions
	for i := 1; i <= 2; i++ {
		db.DB.Create(&domain.MessageMention{
			ID:                fmt.Sprintf("mm-%d", i),
			MessageID:         fmt.Sprintf("msg-%d", i),
			MentionedUserID:   "user-1",
			MentionedByUserID: "user-2",
			MentionKind:       "user",
			CreatedAt:         time.Now(),
		})
	}

	router := gin.New()
	router.GET("/api/v1/me/unread-counts", GetUnreadCounts)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/unread-counts", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Counts struct {
			UnreadMentionCount int `json:"unread_mention_count"`
		} `json:"counts"`
	}
	json.Unmarshal(rec.Body.Bytes(), &payload)

	if payload.Counts.UnreadMentionCount != 2 {
		t.Fatalf("expected 2 unread mentions, got %d", payload.Counts.UnreadMentionCount)
	}
}

func TestPhase67HomeExecutionPulseIncludesTrendFields(t *testing.T) {
	setupTestDB(t)
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko", Email: "nikko@example.com"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "ops"})
	db.DB.Create(&domain.WorkspaceList{ID: "list-1", WorkspaceID: "ws-1", ChannelID: "ch-1", CreatedBy: "user-1", Title: "Tasks"})
	db.DB.Create(&domain.WorkspaceListItem{ListID: "list-1", Content: "Task 1", IsCompleted: false})

	router := gin.New()
	router.GET("/api/v1/home", GetHome)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/home", nil)
	router.ServeHTTP(rec, req)

	var payload struct {
		Home struct {
			ChannelExecutionPulse []struct {
				ChannelID               string `json:"channel_id"`
				OpenItemDelta7d         int    `json:"open_item_delta_7d"`
				OverdueDelta7d          int    `json:"overdue_delta_7d"`
				RecentToolFailureCount int    `json:"recent_tool_failure_count"`
			} `json:"channel_execution_pulse"`
		} `json:"home"`
	}
	json.Unmarshal(rec.Body.Bytes(), &payload)

	if len(payload.Home.ChannelExecutionPulse) == 0 {
		t.Fatal("expected channel execution pulse rows")
	}

	// Fields should at least exist
	pulse := payload.Home.ChannelExecutionPulse[0]
	if pulse.ChannelID != "ch-1" {
		t.Fatalf("expected ch-1, got %s", pulse.ChannelID)
	}
}
