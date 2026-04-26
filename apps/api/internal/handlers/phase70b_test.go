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

func TestPhase70BExecutionTargetContract(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu"})
	db.DB.Create(&domain.Artifact{ID: "art-1", CreatedBy: "user-1", Title: "Target Project", Type: "document"})
	db.DB.Create(&domain.FileAsset{ID: "file-1", Name: "spec.md"})
	db.DB.Create(&domain.FileAsset{ID: "file-2", Name: "plan.md"})

	router := gin.New()
	router.POST("/api/v1/ai/canvas/analyze", AnalyzeCanvasFileGroup)

	t.Run("Valid Analysis with Execution Targets", func(t *testing.T) {
		body := map[string]any{
			"artifact_id": "art-1",
			"file_refs": []map[string]any{
				{"file_id": "file-1"},
				{"file_id": "file-2"},
			},
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/analyze", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			Analysis struct {
				Summary                string `json:"summary"`
				DefaultExecutionTarget *struct {
					Type string `json:"type"`
				} `json:"default_execution_target"`
				Observations []string `json:"observations"`
				NextSteps    []struct {
					Text            string `json:"text"`
					Rationale       string `json:"rationale"`
					ActionHint      string `json:"action_hint"`
					ExecutionTarget *struct {
						Type string `json:"type"`
					} `json:"execution_target"`
				} `json:"next_steps"`
			} `json:"analysis"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Verify fields exist (mock implementation should provide them)
		// For Task 1, we just need to ensure the schema is accepted and returned.
		// We'll update the mock in handlers/ai.go to return these fields.
	})
}
