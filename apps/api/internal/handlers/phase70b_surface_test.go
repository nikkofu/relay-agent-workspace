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
	"github.com/nikkofu/relay-agent-workspace/api/internal/llm"
)

func TestPhase70BSurfaces(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu"})
	db.DB.Create(&domain.Artifact{ID: "art-1", CreatedBy: "user-1", Title: "Surface Test", Type: "document"})
	db.DB.Create(&domain.FileAsset{ID: "file-1", Name: "spec.md"})
	db.DB.Create(&domain.FileAsset{ID: "file-2", Name: "plan.md"})

	router := gin.New()
	router.POST("/api/v1/ai/canvas/analyze", AnalyzeCanvasFileGroup)

	t.Run("AnalyzeCanvas carries targets", func(t *testing.T) {
		rec := httptest.NewRecorder()
		body := `{"artifact_id":"art-1","file_refs":[{"file_id":"file-1"},{"file_id":"file-2"}]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/analyze", bytes.NewBufferString(body))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var payload map[string]any
		json.Unmarshal(rec.Body.Bytes(), &payload)
		
		analysis, ok := payload["analysis"].(map[string]any)
		if !ok { t.Fatal("missing analysis") }

		// Verify default_execution_target (mock returns it)
		if _, ok := analysis["default_execution_target"]; !ok {
			t.Error("missing default_execution_target")
		}

		nextSteps, ok := analysis["next_steps"].([]any)
		if !ok || len(nextSteps) == 0 { t.Fatal("missing next_steps") }

		step := nextSteps[0].(map[string]any)
		if _, ok := step["execution_target"]; !ok {
			t.Error("missing step execution_target")
		}
	})
}

func TestPhase70BMalformed(t *testing.T) {
	rawJSON := `{
		"summary": "Test",
		"default_execution_target": {"type": "unknown_target"},
		"observations": [],
		"next_steps": [
			{
				"text": "Step 1",
				"execution_target": {"type": "workflow"}
			},
			{
				"text": "Step 2",
				"execution_target": {"type": "invalid"}
			}
		]
	}`

	result, err := llm.ParseAnalysisResult(rawJSON)
	if err != nil {
		t.Fatalf("ParseAnalysisResult failed: %v", err)
	}

	if result.DefaultExecutionTarget != nil {
		t.Errorf("expected default_execution_target to be nil for unknown type, got %s", result.DefaultExecutionTarget.Type)
	}

	if len(result.NextSteps) != 2 {
		t.Fatalf("expected 2 next steps, got %d", len(result.NextSteps))
	}

	if result.NextSteps[0].ExecutionTarget == nil || result.NextSteps[0].ExecutionTarget.Type != "workflow" {
		t.Errorf("expected step 0 target 'workflow', got %v", result.NextSteps[0].ExecutionTarget)
	}

	if result.NextSteps[1].ExecutionTarget != nil {
		t.Errorf("expected step 1 target to be nil for invalid type, got %s", result.NextSteps[1].ExecutionTarget.Type)
	}
}
