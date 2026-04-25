package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/llm"
)

type sidecarMockGateway struct{}

func (sidecarMockGateway) Stream(_ context.Context, _ llm.Request) (*llm.StreamSession, error) {
	events := make(chan llm.StreamEvent, 5)
	errs := make(chan error, 1)
	events <- llm.StreamEvent{Type: "reasoning", Text: "Analyzing channel context..."}
	events <- llm.StreamEvent{Type: "tool_call", Text: "searching knowledge..."}
	events <- llm.StreamEvent{Type: "chunk", Text: "Here is the "}
	events <- llm.StreamEvent{Type: "chunk", Text: "answer."}
	events <- llm.StreamEvent{Type: "usage", Text: "{\"input_tokens\":10, \"output_tokens\":20}"}
	close(events)
	close(errs)

	return &llm.StreamSession{
		Provider: "stub",
		Model:    "stub-model",
		Events:   events,
		Errors:   errs,
	}, nil
}

func TestUnifiedAISidecarExecuteAIStreamsNormativeKinds(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	AIGateway = sidecarMockGateway{}
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})

	router := gin.New()
	router.POST("/api/v1/ai/execute", ExecuteAI)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/execute", strings.NewReader(`{"prompt":"hello","channel_id":"ch-1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()

	if !strings.Contains(body, "\"kind\":\"reasoning\"") {
		t.Errorf("expected normative reasoning kind in stream, got %s", body)
	}
	if !strings.Contains(body, "\"kind\":\"tool_call\"") {
		t.Errorf("expected normative tool_call kind in stream, got %s", body)
	}
	if !strings.Contains(body, "\"kind\":\"usage\"") {
		t.Errorf("expected normative usage kind in stream, got %s", body)
	}
	if !strings.Contains(body, "\"kind\":\"answer\"") {
		t.Errorf("expected normative answer kind in stream, got %s", body)
	}
}

func TestUnifiedAISidecarPersistsCanonicalMetadata(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	AIGateway = sidecarMockGateway{}
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", Name: "general", WorkspaceID: "ws-1"})
	
	router := gin.New()
	router.POST("/api/v1/ai/execute", ExecuteAI)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/execute", strings.NewReader(`{"prompt":"hello","channel_id":"ch-1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var msg domain.AIConversationMessage
	if err := db.DB.Where("role = ?", "assistant").Order("created_at desc").First(&msg).Error; err != nil {
		t.Fatalf("failed to find persisted AI message: %v", err)
	}

	if msg.AISidecarJSON == "" {
		t.Fatal("expected AISidecarJSON in conversation message")
	}
}

func TestUnifiedAISidecarDualReadsLegacyFlatFields(t *testing.T) {
	setupTestDB(t)
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", Name: "general", WorkspaceID: "ws-1"})
	
	// Create a message with legacy flat fields
	legacyMetadata := `{"reasoning":"Legacy thought","tool_calls":[{"id":"t1","name":"tool","arguments":"{}"}],"usage":{"total_tokens":50}}`
	msg := domain.Message{
		ID:        "msg-legacy",
		ChannelID: "ch-1",
		UserID:    "ai-assistant",
		Content:   "Old answer",
		Metadata:  legacyMetadata,
		CreatedAt: time.Now(),
	}
	if err := db.DB.Create(&msg).Error; err != nil {
		t.Fatalf("failed to create legacy message: %v", err)
	}

	// Verify it's there
	var count int64
	db.DB.Model(&domain.Message{}).Where("channel_id = ?", "ch-1").Count(&count)
	if count == 0 {
		t.Fatal("message not found in DB after create")
	}

	router := gin.New()
	router.GET("/api/v1/channels/:id/messages", GetMessages)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/ch-1/messages?channel_id=ch-1", nil)
	router.ServeHTTP(rec, req)

	var payload struct {
		Messages []struct {
			Metadata string `json:"metadata"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(payload.Messages) == 0 {
		t.Fatal("no messages returned")
	}
	
	var meta struct {
		AISidecar map[string]any `json:"ai_sidecar"`
	}
	if err := json.Unmarshal([]byte(payload.Messages[0].Metadata), &meta); err != nil {
		t.Fatalf("failed to decode message metadata: %v", err)
	}

	sidecar := meta.AISidecar
	if sidecar == nil {
		t.Fatal("expected synthesized ai_sidecar from legacy fields")
	}
	if sidecar["reasoning"] != "Legacy thought" {
		t.Errorf("expected reasoning from legacy, got %v", sidecar["reasoning"])
	}
}
