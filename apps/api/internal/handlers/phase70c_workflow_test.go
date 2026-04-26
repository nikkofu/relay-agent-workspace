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

func TestPhase70CWorkflowFlow(t *testing.T) {
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
				"text": "Build feature workflow",
				"rationale": "Direct solution",
				"action_hint": "plan",
				"execution_target": map[string]any{
					"type": "workflow",
					"workflow_draft": map[string]any{
						"title": "New Feature Workflow",
						"goal": "Deliver X by Monday",
						"steps": []map[string]any{
							{"title": "Design API"},
							{"title": "Implement backend"},
							{"title": "Test end-to-end"},
						},
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
		ID:             "msg-workflow-ana-1",
		ConversationID: "conv-1",
		Role:           "assistant",
		Content:        "Analysis complete.",
		AISidecarJSON:  string(sidecarJSON),
	}
	db.DB.Create(&msg)

	router := gin.New()
	router.POST("/api/v1/ai/canvas/generate-workflow-draft", GenerateWorkflowDraftFromAnalysis)
	router.POST("/api/v1/ai/canvas/confirm-create-workflow", ConfirmCreateWorkflowFromDraft)

	var draftID string

	t.Run("Generate Workflow Draft", func(t *testing.T) {
		body := map[string]any{
			"artifact_id":          "art-1",
			"channel_id":           "ch-1",
			"analysis_snapshot_id": "msg-workflow-ana-1",
			"step_index":           0, // Use the first step which has the workflow target
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/generate-workflow-draft", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			Draft struct {
				DraftID   string `json:"draft_id"`
				Title     string `json:"title"`
				Goal      string `json:"goal"`
				Steps     []struct {
					Title string `json:"title"`
				} `json:"steps"`
			} `json:"draft"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Draft.DraftID == "" {
			t.Error("expected non-empty draft_id")
		}
		if resp.Draft.Title != "New Feature Workflow" {
			t.Errorf("expected title 'New Feature Workflow', got '%s'", resp.Draft.Title)
		}
		if len(resp.Draft.Steps) != 3 {
			t.Errorf("expected 3 steps, got %d", len(resp.Draft.Steps))
		}
		draftID = resp.Draft.DraftID
	})

	t.Run("Confirm Create Workflow", func(t *testing.T) {
		if draftID == "" {
			t.Skip("skipping confirm because draft generation failed")
		}

		body := map[string]any{
			"draft_id": draftID,
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/confirm-create-workflow", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			WorkflowID string `json:"workflow_id"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.WorkflowID == "" {
			t.Error("expected non-empty workflow_id")
		}

		// Verify workflow exists in DB
		var wf domain.WorkflowDefinition
		if err := db.DB.First(&wf, "id = ?", resp.WorkflowID).Error; err != nil {
			t.Fatalf("workflow not found in DB: %v", err)
		}
		if wf.Name != "New Feature Workflow" {
			t.Errorf("expected name 'New Feature Workflow', got '%s'", wf.Name)
		}
	})
}
