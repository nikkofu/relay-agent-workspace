package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/llm"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
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

type captureGateway struct {
	lastRequest llm.Request
}

func (g *captureGateway) Stream(_ context.Context, req llm.Request) (*llm.StreamSession, error) {
	g.lastRequest = req
	return stubGateway{}.Stream(context.Background(), req)
}

type failingAfterChunkGateway struct{}

func (failingAfterChunkGateway) Stream(_ context.Context, _ llm.Request) (*llm.StreamSession, error) {
	events := make(chan llm.StreamEvent, 2)
	errs := make(chan error, 1)
	events <- llm.StreamEvent{Type: "chunk", Text: "Partial reply"}
	errs <- errors.New("upstream stream failed")
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
	assertPrefixedUUID(t, conversations[0].ID, "ai-conv")

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
	assertPrefixedUUID(t, messages[0].ID, "ai-msg")
	assertPrefixedUUID(t, messages[1].ID, "ai-msg")
	if messages[1].Reasoning == "" {
		t.Fatalf("expected assistant reasoning to be persisted: %#v", messages[1])
	}
}

func TestExecuteAIForwardsCommandField(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	gateway := &captureGateway{}
	AIGateway = gateway
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})

	router := gin.New()
	router.POST("/api/v1/ai/execute", ExecuteAI)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/execute", strings.NewReader(`{"prompt":"/canvas build release plan","channel_id":"ai-chat","command":"canvas"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gateway.lastRequest.Command != "canvas" {
		t.Fatalf("expected command to be forwarded to gateway, got %#v", gateway.lastRequest)
	}
}

func TestExecuteAIAllowsCitationPayloadField(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	gateway := &captureGateway{}
	AIGateway = gateway
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})

	router := gin.New()
	router.POST("/api/v1/ai/execute", ExecuteAI)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/execute", strings.NewReader(`{"prompt":"hello","channel_id":"ch-1","citations":[{"id":"citation-1","source_kind":"file","source_ref":"file-1"}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
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
		Conversation domain.AIConversation          `json:"conversation"`
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

func TestGenerateAndGetThreadSummary(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	AIGateway = stubGateway{}
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general"})
	db.DB.Create(&domain.Message{
		ID:        "msg-parent",
		ChannelID: "ch-1",
		UserID:    "user-1",
		Content:   "We should finalize the launch checklist today.",
		CreatedAt: mustParseAITestTime(t, "2026-04-18T12:00:00Z"),
	})
	db.DB.Create(&domain.Message{
		ID:        "msg-reply",
		ChannelID: "ch-1",
		UserID:    "user-1",
		ThreadID:  "msg-parent",
		Content:   "I can handle the API validation and release notes.",
		CreatedAt: mustParseAITestTime(t, "2026-04-18T12:02:00Z"),
	})

	router := gin.New()
	router.POST("/api/v1/messages/:id/summary", GenerateThreadSummary)
	router.GET("/api/v1/messages/:id/summary", GetThreadSummary)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/msg-parent/summary", strings.NewReader(`{"provider":"gemini"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on thread summary generate, got %d body=%s", rec.Code, rec.Body.String())
	}

	var generatePayload struct {
		Summary domain.AISummary `json:"summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &generatePayload); err != nil {
		t.Fatalf("failed to decode generate payload: %v", err)
	}
	if generatePayload.Summary.ScopeType != "thread" || generatePayload.Summary.ScopeID != "msg-parent" {
		t.Fatalf("unexpected generated thread summary: %#v", generatePayload.Summary)
	}
	if generatePayload.Summary.Content == "" || generatePayload.Summary.Reasoning == "" {
		t.Fatalf("expected generated content and reasoning: %#v", generatePayload.Summary)
	}
	if generatePayload.Summary.MessageCount != 2 {
		t.Fatalf("expected message count 2, got %#v", generatePayload.Summary)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/messages/msg-parent/summary", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on thread summary get, got %d body=%s", rec.Code, rec.Body.String())
	}

	var getPayload struct {
		Summary *domain.AISummary `json:"summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &getPayload); err != nil {
		t.Fatalf("failed to decode get payload: %v", err)
	}
	if getPayload.Summary == nil || getPayload.Summary.ScopeID != "msg-parent" {
		t.Fatalf("unexpected stored thread summary: %#v", getPayload.Summary)
	}
}

func TestGenerateChannelSummary(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	AIGateway = stubGateway{}
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general"})
	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-1",
		UserID:    "user-1",
		Content:   "Launch blockers are mostly resolved.",
		CreatedAt: mustParseAITestTime(t, "2026-04-18T12:00:00Z"),
	})
	db.DB.Create(&domain.Message{
		ID:        "msg-2",
		ChannelID: "ch-1",
		UserID:    "user-1",
		Content:   "Still waiting on final QA notes.",
		CreatedAt: mustParseAITestTime(t, "2026-04-18T12:03:00Z"),
	})

	router := gin.New()
	router.POST("/api/v1/channels/:id/summary", GenerateChannelSummary)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels/ch-1/summary", strings.NewReader(`{"provider":"openrouter","model":"test-model"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on channel summary generate, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Summary domain.AISummary `json:"summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode channel summary payload: %v", err)
	}
	if payload.Summary.ScopeType != "channel" || payload.Summary.ScopeID != "ch-1" {
		t.Fatalf("unexpected channel summary payload: %#v", payload.Summary)
	}
	if payload.Summary.MessageCount != 2 {
		t.Fatalf("expected message count 2, got %#v", payload.Summary)
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

func TestChannelExecutionCreateListItemWithAIAssistFallsBackWithoutAI(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Should we add this to list?"})

	// Force gateway failure
	AIGateway = &errorGateway{err: errors.New("upstream unreachable")}

	router := gin.New()
	router.POST("/api/v1/ai/lists/draft", CreateListItemDraft)

	rec := httptest.NewRecorder()
	reqBody := `{"message_id":"msg-1","list_id":"list-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/lists/draft", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	// Acceptable failure mode: 200 with ok: false OR graceful fallback fields
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 even on AI failure (soft fallback), got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		OK       bool   `json:"ok"`
		Fallback string `json:"fallback"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.OK != false || payload.Fallback != "manual_entry" {
		t.Fatalf("expected graceful fallback, got %#+v", payload)
	}
}

type errorGateway struct {
	err error
}

func (g *errorGateway) Stream(_ context.Context, _ llm.Request) (*llm.StreamSession, error) {
	return nil, g.err
}

func TestPhase65CAISlashCommandAsk(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko", OrganizationID: "org-1"})
	db.DB.Create(&domain.Channel{ID: "ch-1", Name: "ops", WorkspaceID: "ws-1"})
	
	AIGateway = &stubGateway{}

	router := gin.New()
	router.POST("/api/v1/channels/:id/messages/ask", AISlashCommandAsk)

	for _, content := range []string{"/ask how to deploy?", "/ask    how to deploy?", "how to deploy?"} {
		t.Run(content, func(t *testing.T) {
			rec := httptest.NewRecorder()
			reqBody := fmt.Sprintf(`{"content":%q}`, content)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/channels/ch-1/messages/ask", strings.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
				t.Fatalf("expected 200/201, got %d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestPhase75ChannelAIStreamFinalIncludesMessageID(t *testing.T) {
	setupTestDB(t)
	db.DB.Create(&domain.Workspace{ID: "ws-1", Name: "Relay"})
	db.DB.Create(&domain.Channel{ID: "ch-1", Name: "ops", WorkspaceID: "ws-1"})

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)
	client := realtime.NewTestClient(2)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 1 })
	broadcastChannelAIStreamChunkWithMessageID("ch-1", "stream-1", "msg-trigger", "user-2", "answer", "", "msg-final", true)

	raw, err := client.Receive(time.Second)
	if err != nil {
		t.Fatalf("failed to receive final stream event: %v", err)
	}
	var event realtime.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("failed to decode final stream event: %v", err)
	}
	if event.Type != "channel.ai.stream.chunk" {
		t.Fatalf("expected channel.ai.stream.chunk, got %s", event.Type)
	}
	payload, ok := event.Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %#v", event.Payload)
	}
	if payload["message_id"] != "msg-final" || payload["is_final"] != true {
		t.Fatalf("expected final payload with message_id, got %#v", payload)
	}
}
