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

type stubGateway struct{}

func (stubGateway) Stream(_ context.Context, _ llm.Request) (*llm.StreamSession, error) {
	events := make(chan llm.StreamEvent, 3)
	errs := make(chan error, 1)
	events <- llm.StreamEvent{Type: "reasoning", Text: "Considering options..."}
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
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	AIGateway = stubGateway{}
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
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/event-stream") {
		t.Fatalf("expected event stream content type, got %s", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "event: start") {
		t.Fatalf("expected start event, got %s", body)
	}
	if !strings.Contains(body, "event: reasoning") || !strings.Contains(body, "Considering options...") {
		t.Fatalf("expected reasoning event in stream, got %s", body)
	}
	if !strings.Contains(body, "Thinking...") || !strings.Contains(body, "Done.") {
		t.Fatalf("expected chunk texts in stream, got %s", body)
	}
	if !strings.Contains(body, "event: done") {
		t.Fatalf("expected done event, got %s", body)
	}

	var conversations []domain.AIConversation
	if err := db.DB.Find(&conversations).Error; err != nil {
		t.Fatalf("failed to load conversations: %v", err)
	}
	if len(conversations) != 1 {
		t.Fatalf("expected 1 ai conversation, got %d", len(conversations))
	}

	var messages []domain.AIConversationMessage
	if err := db.DB.Where("conversation_id = ?", conversations[0].ID).Order("created_at asc").Find(&messages).Error; err != nil {
		t.Fatalf("failed to load ai conversation messages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 persisted ai messages, got %d", len(messages))
	}
	if messages[0].Role != "user" || messages[1].Role != "assistant" {
		t.Fatalf("unexpected ai message roles: %#v", messages)
	}
	if messages[1].Reasoning == "" {
		t.Fatalf("expected assistant reasoning to be persisted: %#v", messages[1])
	}
}

func TestGetAIConversationsAndDetail(t *testing.T) {
	setupTestDB(t)
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.AIConversation{
		ID:        "ai-conv-1",
		UserID:    "user-1",
		ChannelID: "ai-chat",
		Provider:  "gemini",
		Model:     "gemini-2.5-flash",
		CreatedAt: mustParseAITestTime(t, "2026-04-18T12:00:00Z"),
		UpdatedAt: mustParseAITestTime(t, "2026-04-18T12:01:00Z"),
	})
	db.DB.Create(&domain.AIConversationMessage{
		ID:             "ai-msg-1",
		ConversationID: "ai-conv-1",
		Role:           "user",
		Content:        "Summarize this channel",
		CreatedAt:      mustParseAITestTime(t, "2026-04-18T12:00:00Z"),
	})
	db.DB.Create(&domain.AIConversationMessage{
		ID:             "ai-msg-2",
		ConversationID: "ai-conv-1",
		Role:           "assistant",
		Content:        "Here is the summary.",
		Reasoning:      "Considering recent messages...",
		CreatedAt:      mustParseAITestTime(t, "2026-04-18T12:01:00Z"),
	})

	router := gin.New()
	router.GET("/api/v1/ai/conversations", GetAIConversations)
	router.GET("/api/v1/ai/conversations/:id", GetAIConversation)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/conversations", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on ai conversations list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Conversations []domain.AIConversation `json:"conversations"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode ai conversations list: %v", err)
	}
	if len(listPayload.Conversations) != 1 || listPayload.Conversations[0].ID != "ai-conv-1" {
		t.Fatalf("unexpected ai conversation list payload: %#v", listPayload.Conversations)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/ai/conversations/ai-conv-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on ai conversation detail, got %d body=%s", rec.Code, rec.Body.String())
	}

	var detailPayload struct {
		Conversation domain.AIConversation        `json:"conversation"`
		Messages     []domain.AIConversationMessage `json:"messages"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &detailPayload); err != nil {
		t.Fatalf("failed to decode ai conversation detail: %v", err)
	}
	if detailPayload.Conversation.ID != "ai-conv-1" || len(detailPayload.Messages) != 2 {
		t.Fatalf("unexpected ai conversation detail payload: %#v", detailPayload)
	}
}

func TestGetAIConfigReturnsEnabledProvidersAndModels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	AIConfig = llm.Config{
		DefaultProvider: "gemini",
		Providers: map[string]llm.ProviderConfig{
			"gemini": {
				Name:    "gemini",
				Enabled: true,
				Model:   "gemini-3-flash-preview",
			},
			"openrouter": {
				Name:    "openrouter",
				Enabled: true,
				Model:   "nvidia/nemotron-3-super-120b-a12b:free",
			},
			"openai": {
				Name:    "openai",
				Enabled: false,
				Model:   "gpt-4.1-mini",
			},
		},
	}

	router := gin.New()
	router.GET("/api/v1/ai/config", GetAIConfig)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/config", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		DefaultProvider string `json:"default_provider"`
		Providers       []struct {
			ID     string   `json:"id"`
			Models []string `json:"models"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.DefaultProvider != "gemini" {
		t.Fatalf("unexpected default provider: %s", payload.DefaultProvider)
	}
	if len(payload.Providers) != 2 {
		t.Fatalf("expected 2 enabled providers, got %d", len(payload.Providers))
	}
}

func mustParseAITestTime(t *testing.T, raw string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("failed to parse test time %q: %v", raw, err)
	}
	return parsed
}
