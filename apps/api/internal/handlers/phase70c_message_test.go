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

func TestPhase70CMessageFlow(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	user := domain.User{ID: "user-1", Name: "Nikko Fu"}
	db.DB.Create(&user)
	artifact := domain.Artifact{ID: "art-1", CreatedBy: "user-1", Title: "Project Alpha", Type: "document"}
	db.DB.Create(&artifact)

	// Create a mock analysis snapshot in a message
	analysisData := map[string]any{
		"summary":      "We should build a new feature.",
		"observations": []string{"User needs X", "Tech debt in Y"},
		"next_steps": []map[string]any{
			{
				"text": "Announce to team",
				"rationale": "Transparency",
				"action_hint": "share",
				"execution_target": map[string]any{
					"type": "channel_message",
					"message_draft": map[string]any{
						"channel_id": "ch-updates",
						"body":       "Hi team, we're building X!",
					},
				},
			},
		},
	}
	sidecar := domain.AISidecar{
		Analysis: analysisData,
	}
	sidecarJSON, _ := json.Marshal(sidecar)
	msg := domain.AIConversationMessage{
		ID:             "msg-msg-ana-1",
		ConversationID: "conv-1",
		Role:           "assistant",
		Content:        "Analysis complete.",
		AISidecarJSON:  string(sidecarJSON),
	}
	db.DB.Create(&msg)

	router := gin.New()
	router.POST("/api/v1/ai/canvas/generate-message-draft", GenerateMessageDraftFromAnalysis)
	router.POST("/api/v1/ai/canvas/confirm-publish-message", ConfirmPublishMessageFromDraft)

	var draftID string

	t.Run("Generate Message Draft", func(t *testing.T) {
		body := map[string]any{
			"artifact_id":          "art-1",
			"channel_id":           "ch-1",
			"analysis_snapshot_id": "msg-msg-ana-1",
			"step_index":           0,
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/generate-message-draft", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			Draft struct {
				DraftID   string `json:"draft_id"`
				ChannelID string `json:"channel_id"`
				Body      string `json:"body"`
			} `json:"draft"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Draft.DraftID == "" {
			t.Error("expected non-empty draft_id")
		}
		if resp.Draft.ChannelID != "ch-updates" {
			t.Errorf("expected channel_id 'ch-updates', got '%s'", resp.Draft.ChannelID)
		}
		if resp.Draft.Body != "Hi team, we're building X!" {
			t.Errorf("expected body 'Hi team, we're building X!', got '%s'", resp.Draft.Body)
		}
		draftID = resp.Draft.DraftID
	})

	t.Run("Confirm Publish Message", func(t *testing.T) {
		if draftID == "" {
			t.Skip("skipping confirm because draft generation failed")
		}

		body := map[string]any{
			"draft_id": draftID,
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/confirm-publish-message", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			MessageID string `json:"message_id"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.MessageID == "" {
			t.Error("expected non-empty message_id")
		}

		// Verify message exists in DB
		var m domain.Message
		if err := db.DB.First(&m, "id = ?", resp.MessageID).Error; err != nil {
			t.Fatalf("message not found in DB: %v", err)
		}
		if m.ChannelID != "ch-updates" {
			t.Errorf("expected channel 'ch-updates', got '%s'", m.ChannelID)
		}
		if m.Content != "Hi team, we're building X!" {
			t.Errorf("expected content 'Hi team, we're building X!', got '%s'", m.Content)
		}
	})
}
