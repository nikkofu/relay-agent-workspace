package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

func TestPhase70ADraftGeneration(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	user := domain.User{ID: "user-1", Name: "Nikko Fu"}
	db.DB.Create(&user)
	artifact := domain.Artifact{ID: "art-1", CreatedBy: "user-1", Title: "Project Alpha", Type: "document"}
	db.DB.Create(&artifact)

	// Create a mock analysis snapshot in a message
	analysisData := map[string]any{
		"summary":      "We should do A, B, and C.",
		"observations": []string{"A is good", "B is missing"},
		"next_steps": []map[string]any{
			{"text": "Implement A", "rationale": "High priority", "action_hint": "plan"},
			{"text": "Fix B", "rationale": "Blocker", "action_hint": "decide"},
		},
	}
	sidecar := domain.AISidecar{
		Analysis: analysisData,
	}
	sidecarJSON, _ := json.Marshal(sidecar)
	msg := domain.AIConversationMessage{
		ID:             "msg-analysis-1",
		ConversationID: "conv-1",
		Role:           "assistant",
		Content:        "Analysis complete.",
		AISidecarJSON:  string(sidecarJSON),
	}
	db.DB.Create(&msg)

	router := gin.New()
	// We'll need a way to mock getCurrentUser or ensure it works in test
	router.POST("/api/v1/ai/canvas/generate-list-draft", GenerateListDraftFromAnalysis)

	t.Run("Valid Analysis Snapshot -> Draft", func(t *testing.T) {
		body := map[string]any{
			"artifact_id":          "art-1",
			"channel_id":           "ch-1",
			"analysis_snapshot_id": "msg-analysis-1",
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/generate-list-draft", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			Draft struct {
				DraftID   string `json:"draft_id"`
				ChannelID string `json:"channel_id"`
				Title     string `json:"title"`
				Items     []struct {
					Title string `json:"title"`
				} `json:"items"`
			} `json:"draft"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Draft.DraftID == "" {
			t.Error("expected non-empty draft_id")
		}
		if resp.Draft.Title == "" {
			t.Error("expected default title")
		}
		if len(resp.Draft.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(resp.Draft.Items))
		}
		if resp.Draft.Items[0].Title != "Implement A" {
			t.Errorf("expected item 0 title 'Implement A', got '%s'", resp.Draft.Items[0].Title)
		}
	})

	t.Run("Missing Analysis Snapshot -> 404", func(t *testing.T) {
		body := map[string]any{
			"artifact_id":          "art-1",
			"channel_id":           "ch-1",
			"analysis_snapshot_id": "msg-missing",
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/generate-list-draft", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("Invalid Request -> 400", func(t *testing.T) {
		body := map[string]any{
			"artifact_id": "art-1",
			// missing analysis_snapshot_id
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/generate-list-draft", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}
