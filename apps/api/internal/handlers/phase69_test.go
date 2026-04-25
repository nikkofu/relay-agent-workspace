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

func TestPhase69MultiFileAnalysisContract(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu"})
	db.DB.Create(&domain.Artifact{ID: "art-1", CreatedBy: "user-1", Title: "Multi-file Project", Type: "document"})
	db.DB.Create(&domain.FileAsset{ID: "file-1", Name: "notes.md"})
	db.DB.Create(&domain.FileAsset{ID: "file-2", Name: "data.csv"})

	router := gin.New()
	router.POST("/api/v1/ai/canvas/analyze", AnalyzeCanvasFileGroup)

	t.Run("Valid Multi-file Request", func(t *testing.T) {
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
				Summary      string   `json:"summary"`
				Observations []string `json:"observations"`
				NextSteps    []struct {
					Text       string `json:"text"`
					Rationale  string `json:"rationale"`
					ActionHint string `json:"action_hint"`
				} `json:"next_steps"`
			} `json:"analysis"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Analysis.Summary == "" {
			t.Error("expected non-empty summary")
		}
	})

	t.Run("Reject Raw Drag Payload", func(t *testing.T) {
		// Shaped like Phase 68 drag payload, not Phase 69 snapshotted group
		body := map[string]any{
			"kind": "file-to-canvas",
			"file": map[string]any{
				"id": "file-1",
				"title": "notes.md",
			},
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/analyze", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code == http.StatusOK {
			t.Error("expected error for raw drag payload, got 200")
		}
	})

	t.Run("Invalid Request Missing File Refs", func(t *testing.T) {
		body := map[string]any{
			"artifact_id": "art-1",
			"file_refs":    []any{},
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/analyze", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code == http.StatusOK {
			t.Error("expected error for empty file_refs, got 200")
		}
	})
}
