package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/llm"
)

type stubGateway struct{}

func (stubGateway) Stream(_ context.Context, _ llm.Request) (*llm.StreamSession, error) {
	events := make(chan llm.StreamEvent, 2)
	errs := make(chan error, 1)
	events <- llm.StreamEvent{Type: "chunk", Text: "Thinking..."}
	events <- llm.StreamEvent{Type: "chunk", Text: "Done."}
	close(events)
	close(errs)

	return &llm.StreamSession{
		Provider: "stub",
		Model:    "stub-model",
		Events:   events,
		Errors:   errs,
	}, nil
}

func TestExecuteAIStreamsSSE(t *testing.T) {
	gin.SetMode(gin.TestMode)
	AIGateway = stubGateway{}

	router := gin.New()
	router.POST("/api/v1/ai/execute", ExecuteAI)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/execute", strings.NewReader(`{"prompt":"hello","channel_id":"ch-1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/event-stream") {
		t.Fatalf("expected event stream content type, got %s", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "event: start") {
		t.Fatalf("expected start event, got %s", body)
	}
	if !strings.Contains(body, "Thinking...") || !strings.Contains(body, "Done.") {
		t.Fatalf("expected chunk texts in stream, got %s", body)
	}
	if !strings.Contains(body, "event: done") {
		t.Fatalf("expected done event, got %s", body)
	}
}
