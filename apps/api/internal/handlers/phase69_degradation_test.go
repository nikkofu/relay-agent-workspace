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

func TestPhase69Degradation(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu"})
	db.DB.Create(&domain.Artifact{ID: "art-1", CreatedBy: "user-1", Title: "Degradation Test", Type: "document"})
	
	// file-1: fully indexed
	db.DB.Create(&domain.FileAsset{
		ID:               "file-1",
		Name:             "report.txt",
		ExtractionStatus: "ready",
		// We'll mock the content retrieval in the handler
	})
	
	// file-2: indexing failed or missing extraction
	db.DB.Create(&domain.FileAsset{
		ID:               "file-2",
		Name:             "unknown.dat",
		ExtractionStatus: "failed",
	})

	router := gin.New()
	router.POST("/api/v1/ai/canvas/analyze", AnalyzeCanvasFileGroup)

	t.Run("Graceful Degradation with Mixed Quality", func(t *testing.T) {
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
			t.Errorf("expected 200 even with partial failure, got %d", rec.Code)
		}

		var resp AnalyzeCanvasResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Analysis.Summary == "" {
			t.Error("expected summary even with degraded files")
		}
	})
}
