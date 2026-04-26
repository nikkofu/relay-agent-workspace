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

func TestPhase71ExecutionHistoryModel(t *testing.T) {
	setupTestDB(t)

	t.Run("Create valid event", func(t *testing.T) {
		event := domain.ExecutionHistoryEvent{
			ID:                  "exec-hist-1",
			EventType:           "draft_generated",
			Status:              "success",
			ActorUserID:         "user-1",
			AnalysisSnapshotID:  "analysis-1",
			ExecutionTargetType: "list",
			DraftID:             "draft-1",
			DraftType:           "list",
			CreatedAt:           time.Now().UTC(),
		}

		if err := db.DB.Create(&event).Error; err != nil {
			t.Fatalf("failed to create execution history event: %v", err)
		}

		var saved domain.ExecutionHistoryEvent
		if err := db.DB.First(&saved, "id = ?", "exec-hist-1").Error; err != nil {
			t.Fatalf("failed to find saved event: %v", err)
		}

		if saved.EventType != "draft_generated" {
			t.Errorf("expected EventType 'draft_generated', got %s", saved.EventType)
		}
	})

	t.Run("Append-only behavior", func(t *testing.T) {
		// Create first event
		e1 := domain.ExecutionHistoryEvent{
			ID:                  "exec-hist-2",
			EventType:           "draft_generated",
			Status:              "success",
			ActorUserID:         "user-1",
			AnalysisSnapshotID:  "analysis-2",
			ExecutionTargetType: "workflow",
			DraftID:             "draft-2",
			DraftType:           "workflow",
			CreatedAt:           time.Now().UTC(),
		}
		db.DB.Create(&e1)

		// Append second event for same group
		e2 := domain.ExecutionHistoryEvent{
			ID:                  "exec-hist-3",
			EventType:           "confirmed",
			Status:              "success",
			ActorUserID:         "user-1",
			AnalysisSnapshotID:  "analysis-2",
			ExecutionTargetType: "workflow",
			DraftID:             "draft-2",
			DraftType:           "workflow",
			CreatedAt:           time.Now().UTC().Add(time.Second),
		}
		if err := db.DB.Create(&e2).Error; err != nil {
			t.Fatalf("failed to append event: %v", err)
		}

		var count int64
		db.DB.Model(&domain.ExecutionHistoryEvent{}).Where("analysis_snapshot_id = ?", "analysis-2").Count(&count)
		if count != 2 {
			t.Errorf("expected 2 events for analysis-2, got %d", count)
		}
	})
}

func TestPhase71ExecutionHistoryWrites(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu"})
	db.DB.Create(&domain.Artifact{ID: "art-1", CreatedBy: "user-1", Title: "Project Alpha", Type: "document"})

	// Mock analysis snapshot
	analysisData := map[string]any{
		"summary": "Build things.",
		"next_steps": []map[string]any{
			{"text": "Task 1", "action_hint": "plan"},
		},
	}
	sidecar := domain.AISidecar{Analysis: analysisData}
	sidecarJSON, _ := json.Marshal(sidecar)
	msg := domain.AIConversationMessage{
		ID:             "msg-1",
		ConversationID: "conv-1",
		Role:           "assistant",
		Content:        "Analysis complete.",
		AISidecarJSON:  string(sidecarJSON),
	}
	db.DB.Create(&msg)

	router := gin.New()
	router.POST("/api/v1/ai/canvas/generate-list-draft", GenerateListDraftFromAnalysis)
	router.POST("/api/v1/ai/canvas/confirm-create-list", ConfirmCreateListFromDraft)

	t.Run("List draft generation writes event", func(t *testing.T) {
		body := map[string]any{
			"artifact_id":          "art-1",
			"channel_id":           "ch-1",
			"analysis_snapshot_id": "msg-1",
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/ai/canvas/generate-list-draft", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var event domain.ExecutionHistoryEvent
		err := db.DB.Where("analysis_snapshot_id = ? AND event_type = ?", "msg-1", "draft_generated").First(&event).Error
		if err != nil {
			t.Fatalf("failed to find draft_generated event: %v", err)
		}
		if event.ExecutionTargetType != "list" {
			t.Errorf("expected target 'list', got %s", event.ExecutionTargetType)
		}
	})

	t.Run("List confirmation writes event", func(t *testing.T) {
		// First generate draft to get draft_id
		var draft domain.AnalysisListDraft
		db.DB.Where("analysis_snapshot_id = ?", "msg-1").First(&draft)

		db.DB.Create(&domain.Workspace{ID: "ws-1", Name: "Relay"})
		db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "ops"})

		body := map[string]any{
			"draft_id": draft.ID,
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/ai/canvas/confirm-create-list", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var event domain.ExecutionHistoryEvent
		err := db.DB.Where("analysis_snapshot_id = ? AND event_type = ?", "msg-1", "created").First(&event).Error
		if err != nil {
			t.Fatalf("failed to find created event: %v", err)
		}
		
		// Check confirmed event too
		var confirmed domain.ExecutionHistoryEvent
		err = db.DB.Where("analysis_snapshot_id = ? AND event_type = ?", "msg-1", "confirmed").First(&confirmed).Error
		if err != nil {
			t.Fatalf("failed to find confirmed event: %v", err)
		}
	})
}

func TestPhase71ExecutionHistoryQuery(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu"})
	
	// Create events for snapshot-1
	db.DB.Create(&domain.ExecutionHistoryEvent{
		ID: "e1", AnalysisSnapshotID: "snapshot-1", EventType: "draft_generated", Status: "success", CreatedAt: time.Now().Add(-time.Minute),
	})
	db.DB.Create(&domain.ExecutionHistoryEvent{
		ID: "e2", AnalysisSnapshotID: "snapshot-1", EventType: "confirmed", Status: "success", CreatedAt: time.Now(),
	})
	
	// Create event for snapshot-2
	db.DB.Create(&domain.ExecutionHistoryEvent{
		ID: "e3", AnalysisSnapshotID: "snapshot-2", EventType: "draft_generated", Status: "success",
	})

	router := gin.New()
	router.GET("/api/v1/ai/canvas/analysis-execution-history", GetAnalysisExecutionHistory)

	t.Run("Query by analysis_snapshot_id", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/ai/canvas/analysis-execution-history?analysis_snapshot_id=snapshot-1", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp struct {
			Events []domain.ExecutionHistoryEvent `json:"events"`
		}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		if len(resp.Events) != 2 {
			t.Errorf("expected 2 events, got %d", len(resp.Events))
		}
		if resp.Events[0].AnalysisSnapshotID != "snapshot-1" {
			t.Errorf("expected snapshot-1, got %s", resp.Events[0].AnalysisSnapshotID)
		}
	})

	t.Run("Missing analysis_snapshot_id -> 400", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/ai/canvas/analysis-execution-history", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}

func TestPhase71ActivityProjection(t *testing.T) {
	setupTestDB(t)

	user := domain.User{ID: "user-1", Name: "Nikko Fu"}
	db.DB.Create(&user)
	
	db.DB.Create(&domain.ExecutionHistoryEvent{
		ID: "e-created", ActorUserID: "user-1", EventType: "created", Status: "success", ExecutionTargetType: "list", CreatedObjectID: "list-1", CreatedAt: time.Now(),
	})

	router := gin.New()
	router.GET("/api/v1/activity", GetActivity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/activity", nil)
	// We need to mock getCurrentUser. In setupTestDB it might be using a header or something.
	// Let's check how setupTestDB handles it.
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Activities []activityItem `json:"activities"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	found := false
	for _, a := range resp.Activities {
		if a.Type == "ai_execution" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ai_execution activity item")
	}
}

func TestPhase71HomeProjection(t *testing.T) {
	setupTestDB(t)

	user := domain.User{ID: "user-1", Name: "Nikko Fu"}
	db.DB.Create(&user)
	
	db.DB.Create(&domain.ExecutionHistoryEvent{
		ID: "e-created-home", ActorUserID: "user-1", EventType: "created", Status: "success", ExecutionTargetType: "list", CreatedObjectID: "list-1", CreatedAt: time.Now(),
	})

	router := gin.New()
	router.GET("/api/v1/home", GetHome)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/home", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Home struct {
			RecentAIExecutions []domain.ExecutionHistoryEvent `json:"recent_ai_executions"`
		} `json:"home"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if len(resp.Home.RecentAIExecutions) == 0 {
		t.Error("expected recent_ai_executions in home response")
	}
}
