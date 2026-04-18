package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

func TestGetUsersReturnsUsers(t *testing.T) {
	setupTestDB(t)
	db.DB.Create(&domain.User{
		ID:             "user-1",
		OrganizationID: "org-1",
		Name:           "Nikko Fu",
		Email:          "nikko@example.com",
		Avatar:         "https://example.com/avatar.png",
		Status:         "online",
	})

	router := gin.New()
	router.GET("/api/v1/users", GetUsers)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Users []domain.User `json:"users"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(payload.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(payload.Users))
	}
	if payload.Users[0].AIInsight == "" {
		t.Fatal("expected ai_insight to be populated")
	}
}

func TestGetMessageThreadReturnsParentAndReplies(t *testing.T) {
	setupTestDB(t)

	parent := domain.Message{
		ID:         "msg-parent",
		ChannelID:  "ch-1",
		UserID:     "user-1",
		Content:    "Parent",
		ReplyCount: 1,
		CreatedAt:  time.Now().Add(-time.Minute),
	}
	reply := domain.Message{
		ID:        "msg-reply",
		ChannelID: "ch-1",
		UserID:    "user-2",
		Content:   "Reply",
		ThreadID:  "msg-parent",
		CreatedAt: time.Now(),
	}
	db.DB.Create(&parent)
	db.DB.Create(&reply)

	router := gin.New()
	router.GET("/api/v1/messages/:id/thread", GetMessageThread)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/msg-parent/thread", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Parent  domain.Message   `json:"parent"`
		Replies []domain.Message `json:"replies"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Parent.ID != "msg-parent" {
		t.Fatalf("unexpected parent: %#v", payload.Parent)
	}
	if len(payload.Replies) != 1 || payload.Replies[0].ID != "msg-reply" {
		t.Fatalf("unexpected replies: %#v", payload.Replies)
	}
}

func TestPatchMeSettingsPersistsPreferences(t *testing.T) {
	setupTestDB(t)

	user := domain.User{
		ID:             "user-1",
		OrganizationID: "org-1",
		Name:           "Nikko Fu",
		Email:          "nikko@example.com",
	}
	db.DB.Create(&user)

	router := gin.New()
	router.PATCH("/api/v1/me/settings", PatchMeSettings)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/settings", bytes.NewBufferString(`{"provider":"gemini","model":"gemini-3-flash-preview","mode":"planning"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var refreshed domain.User
	if err := db.DB.First(&refreshed, "id = ?", "user-1").Error; err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if refreshed.AIProvider != "gemini" || refreshed.AIModel != "gemini-3-flash-preview" || refreshed.AIMode != "planning" {
		t.Fatalf("settings were not persisted: %#v", refreshed)
	}
}

func TestCreateMessageReplyUpdatesParentReplyMetadata(t *testing.T) {
	setupTestDB(t)

	parent := domain.Message{
		ID:         "msg-parent",
		ChannelID:  "ch-1",
		UserID:     "user-1",
		Content:    "Parent",
		ReplyCount: 0,
		CreatedAt:  time.Now().Add(-time.Minute),
	}
	db.DB.Create(&parent)

	router := gin.New()
	router.POST("/api/v1/messages", CreateMessage)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBufferString(`{"channel_id":"ch-1","content":"Reply","user_id":"user-2","thread_id":"msg-parent"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var refreshed domain.Message
	if err := db.DB.First(&refreshed, "id = ?", "msg-parent").Error; err != nil {
		t.Fatalf("failed to reload parent: %v", err)
	}
	if refreshed.ReplyCount != 1 {
		t.Fatalf("expected reply_count=1, got %d", refreshed.ReplyCount)
	}
	if refreshed.LastReplyAt == nil {
		t.Fatal("expected last_reply_at to be set")
	}
}

func TestToggleReactionUpdatesMetadata(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko", Email: "nikko@example.com"})
	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-1",
		UserID:    "user-2",
		Content:   "Hello",
		CreatedAt: time.Now(),
		Metadata:  "{}",
	})

	router := gin.New()
	router.POST("/api/v1/messages/:id/reactions", ToggleReaction)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/msg-1/reactions", bytes.NewBufferString(`{"emoji":"🔥"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var message domain.Message
	if err := db.DB.First(&message, "id = ?", "msg-1").Error; err != nil {
		t.Fatalf("failed to reload message: %v", err)
	}
	if !bytes.Contains([]byte(message.Metadata), []byte(`"emoji":"🔥"`)) {
		t.Fatalf("expected reaction metadata to include emoji, got %s", message.Metadata)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/messages/msg-1/reactions", bytes.NewBufferString(`{"emoji":"🔥"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on toggle off, got %d body=%s", rec.Code, rec.Body.String())
	}

	if err := db.DB.First(&message, "id = ?", "msg-1").Error; err != nil {
		t.Fatalf("failed to reload message after toggle off: %v", err)
	}
	if bytes.Contains([]byte(message.Metadata), []byte(`"emoji":"🔥"`)) {
		t.Fatalf("expected reaction metadata to remove emoji, got %s", message.Metadata)
	}
}

func TestDeleteReplyRecomputesParentThreadMetadata(t *testing.T) {
	setupTestDB(t)

	parent := domain.Message{
		ID:         "msg-parent",
		ChannelID:  "ch-1",
		UserID:     "user-1",
		Content:    "Parent",
		ReplyCount: 2,
		CreatedAt:  time.Now().Add(-3 * time.Hour),
	}
	replyOneTime := time.Now().Add(-2 * time.Hour).UTC()
	replyTwoTime := time.Now().Add(-1 * time.Hour).UTC()
	parent.LastReplyAt = &replyTwoTime

	replyOne := domain.Message{
		ID:        "msg-r1",
		ChannelID: "ch-1",
		UserID:    "user-2",
		Content:   "first",
		ThreadID:  "msg-parent",
		CreatedAt: replyOneTime,
	}
	replyTwo := domain.Message{
		ID:        "msg-r2",
		ChannelID: "ch-1",
		UserID:    "user-3",
		Content:   "second",
		ThreadID:  "msg-parent",
		CreatedAt: replyTwoTime,
	}
	db.DB.Create(&parent)
	db.DB.Create(&replyOne)
	db.DB.Create(&replyTwo)

	router := gin.New()
	router.DELETE("/api/v1/messages/:id", DeleteMessage)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/messages/msg-r2", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var refreshed domain.Message
	if err := db.DB.First(&refreshed, "id = ?", "msg-parent").Error; err != nil {
		t.Fatalf("failed to reload parent: %v", err)
	}
	if refreshed.ReplyCount != 1 {
		t.Fatalf("expected reply_count=1, got %d", refreshed.ReplyCount)
	}
	if refreshed.LastReplyAt == nil || !refreshed.LastReplyAt.Equal(replyOneTime) {
		t.Fatalf("expected last_reply_at=%s, got %#v", replyOneTime, refreshed.LastReplyAt)
	}
}

func TestPinSaveUnreadAndFeedbackEndpointsPersistState(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko", Email: "nikko@example.com"})
	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-1",
		UserID:    "user-2",
		Content:   "Hello",
		CreatedAt: time.Now(),
		Metadata:  "{}",
	})

	router := gin.New()
	router.POST("/api/v1/messages/:id/pin", TogglePinMessage)
	router.POST("/api/v1/messages/:id/later", ToggleSaveForLater)
	router.POST("/api/v1/messages/:id/unread", MarkMessageUnread)
	router.POST("/api/v1/ai/feedback", SubmitAIFeedback)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/msg-1/pin", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on pin, got %d body=%s", rec.Code, rec.Body.String())
	}
	var message domain.Message
	if err := db.DB.First(&message, "id = ?", "msg-1").Error; err != nil {
		t.Fatalf("failed to reload pinned message: %v", err)
	}
	if !message.IsPinned {
		t.Fatal("expected message to be pinned")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/messages/msg-1/later", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on save later, got %d body=%s", rec.Code, rec.Body.String())
	}
	var saved domain.SavedMessage
	if err := db.DB.First(&saved, "message_id = ? AND user_id = ?", "msg-1", "user-1").Error; err != nil {
		t.Fatalf("expected saved message row: %v", err)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/messages/msg-1/unread", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on unread, got %d body=%s", rec.Code, rec.Body.String())
	}
	var unread domain.UnreadMarker
	if err := db.DB.First(&unread, "message_id = ? AND user_id = ?", "msg-1", "user-1").Error; err != nil {
		t.Fatalf("expected unread marker row: %v", err)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/feedback", bytes.NewBufferString(`{"message_id":"msg-1","is_good":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on feedback, got %d body=%s", rec.Code, rec.Body.String())
	}
	var feedback domain.AIFeedback
	if err := db.DB.First(&feedback, "message_id = ? AND user_id = ?", "msg-1", "user-1").Error; err != nil {
		t.Fatalf("expected ai feedback row: %v", err)
	}
	if !feedback.IsGood {
		t.Fatal("expected feedback to be persisted as positive")
	}
}

func TestGetAgentCollabSnapshotReturnsParsedMarkdown(t *testing.T) {
	setupTestDB(t)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "AGENT-COLLAB.md")
	content := `## 📋 Task Board

| Status | Task | Assigned To | Deadline | Description |
| :--- | :--- | :--- | :--- | :--- |
| 🟢 Done | Snapshot API | Codex | 2026-04-18 | Confirm agent-collab snapshot loads. |

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Codex** | verification | Snapshot sync | 100% |
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp collab doc: %v", err)
	}

	prevPath := CollabSnapshotPath
	CollabSnapshotPath = path
	defer func() { CollabSnapshotPath = prevPath }()

	router := gin.New()
	router.GET("/api/v1/agent-collab/snapshot", GetAgentCollabSnapshot)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agent-collab/snapshot", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		ActiveSuperpowers []map[string]any `json:"active_superpowers"`
		TaskBoard         []map[string]any `json:"task_board"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode snapshot response: %v", err)
	}
	if len(payload.ActiveSuperpowers) != 1 || len(payload.TaskBoard) != 1 {
		t.Fatalf("unexpected snapshot payload: %#v", payload)
	}
}

func TestDMEndpointsListCreateAndSendMessages(t *testing.T) {
	setupTestDB(t)

	users := []domain.User{
		{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"},
		{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"},
		{ID: "user-3", Name: "Jane Smith", Email: "jane@example.com"},
	}
	for _, user := range users {
		db.DB.Create(&user)
	}

	dm := domain.DMConversation{ID: "dm-1", CreatedAt: time.Now().Add(-time.Hour)}
	db.DB.Create(&dm)
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-1"})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-2"})
	db.DB.Create(&domain.DMMessage{
		ID:               "dm-msg-1",
		DMConversationID: "dm-1",
		UserID:           "user-2",
		Content:          "Hello from DM",
		CreatedAt:        time.Now().Add(-30 * time.Minute),
	})

	router := gin.New()
	router.GET("/api/v1/dms", GetDMConversations)
	router.POST("/api/v1/dms", CreateOrOpenDMConversation)
	router.GET("/api/v1/dms/:id/messages", GetDMMessages)
	router.POST("/api/v1/dms/:id/messages", CreateDMMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dms", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on dm list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Conversations []map[string]any `json:"conversations"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode dm list: %v", err)
	}
	if len(listPayload.Conversations) != 1 {
		t.Fatalf("expected 1 dm conversation, got %d", len(listPayload.Conversations))
	}
	userIDs, ok := listPayload.Conversations[0]["user_ids"].([]any)
	if !ok || len(userIDs) != 2 {
		t.Fatalf("expected user_ids in dm list response, got %#v", listPayload.Conversations[0]["user_ids"])
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/dms", bytes.NewBufferString(`{"user_id":"user-3"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on dm create, got %d body=%s", rec.Code, rec.Body.String())
	}

	var createPayload struct {
		Conversation struct {
			ID      string   `json:"id"`
			UserIDs []string `json:"user_ids"`
		} `json:"conversation"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("failed to decode dm create: %v", err)
	}
	if createPayload.Conversation.ID == "" {
		t.Fatal("expected created dm conversation id")
	}
	if len(createPayload.Conversation.UserIDs) != 2 {
		t.Fatalf("expected created conversation user_ids, got %#v", createPayload.Conversation.UserIDs)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/dms/dm-1/messages", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on dm messages, got %d body=%s", rec.Code, rec.Body.String())
	}

	var messagesPayload struct {
		Messages []domain.DMMessage `json:"messages"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &messagesPayload); err != nil {
		t.Fatalf("failed to decode dm messages: %v", err)
	}
	if len(messagesPayload.Messages) != 1 {
		t.Fatalf("expected 1 dm message, got %d", len(messagesPayload.Messages))
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/dms/dm-1/messages", bytes.NewBufferString(`{"content":"Reply in DM","user_id":"user-1"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on dm send, got %d body=%s", rec.Code, rec.Body.String())
	}

	var count int64
	db.DB.Model(&domain.DMMessage{}).Where("dm_conversation_id = ?", "dm-1").Count(&count)
	if count != 2 {
		t.Fatalf("expected 2 dm messages after send, got %d", count)
	}
}

func TestCreateDMMessageBroadcastsRealtimeEvent(t *testing.T) {
	setupTestDB(t)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(4)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"})
	db.DB.Create(&domain.DMConversation{ID: "dm-1", CreatedAt: time.Now().UTC()})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-1"})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-2"})

	router := gin.New()
	router.POST("/api/v1/dms/:id/messages", CreateDMMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dms/dm-1/messages", bytes.NewBufferString(`{"content":"Realtime DM","user_id":"user-1"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on dm send, got %d body=%s", rec.Code, rec.Body.String())
	}

	raw, err := client.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("failed to receive realtime event: %v", err)
	}

	var event realtime.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("failed to decode realtime event: %v", err)
	}
	if event.Type != "message.created" {
		t.Fatalf("expected message.created event, got %s", event.Type)
	}

	payload, ok := event.Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %#v", event.Payload)
	}
	if payload["dm_id"] != "dm-1" {
		t.Fatalf("expected dm_id in payload, got %#v", payload)
	}
}

func TestGetActivityReturnsRecentWorkspaceSignals(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	users := []domain.User{
		{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"},
		{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"},
		{ID: "user-3", Name: "Jane Smith", Email: "jane@example.com"},
	}
	for _, user := range users {
		db.DB.Create(&user)
	}
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	parent := domain.Message{ID: "msg-parent", ChannelID: "ch-1", UserID: "user-1", Content: "Root update", CreatedAt: now.Add(-2 * time.Hour), Metadata: "{}"}
	reply := domain.Message{ID: "msg-reply", ChannelID: "ch-1", UserID: "user-3", Content: "Replying in thread", ThreadID: "msg-parent", CreatedAt: now.Add(-time.Hour), Metadata: "{}"}
	mention := domain.Message{ID: "msg-mention", ChannelID: "ch-1", UserID: "user-2", Content: "Looping in Nikko Fu for review", CreatedAt: now.Add(-30 * time.Minute), Metadata: "{}"}
	db.DB.Create(&parent)
	db.DB.Create(&reply)
	db.DB.Create(&mention)
	db.DB.Create(&domain.MessageReaction{MessageID: "msg-parent", UserID: "user-2", Emoji: "🔥", CreatedAt: now.Add(-20 * time.Minute)})

	router := gin.New()
	router.GET("/api/v1/activity", GetActivity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on activity, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Activities []map[string]any `json:"activities"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode activity payload: %v", err)
	}
	if len(payload.Activities) < 3 {
		t.Fatalf("expected at least 3 activity items, got %d", len(payload.Activities))
	}
}

func TestGetInboxReturnsAggregatedSignals(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	users := []domain.User{
		{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"},
		{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"},
		{ID: "user-3", Name: "Jane Smith", Email: "jane@example.com"},
	}
	for _, user := range users {
		db.DB.Create(&user)
	}
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	parent := domain.Message{ID: "msg-parent", ChannelID: "ch-1", UserID: "user-1", Content: "Root update", CreatedAt: now.Add(-2 * time.Hour), Metadata: "{}"}
	reply := domain.Message{ID: "msg-reply", ChannelID: "ch-1", UserID: "user-3", Content: "Replying in thread", ThreadID: "msg-parent", CreatedAt: now.Add(-time.Hour), Metadata: "{}"}
	mention := domain.Message{ID: "msg-mention", ChannelID: "ch-1", UserID: "user-2", Content: "Looping in Nikko Fu for review", CreatedAt: now.Add(-30 * time.Minute), Metadata: "{}"}
	db.DB.Create(&parent)
	db.DB.Create(&reply)
	db.DB.Create(&mention)
	db.DB.Create(&domain.MessageReaction{MessageID: "msg-parent", UserID: "user-2", Emoji: "🔥", CreatedAt: now.Add(-20 * time.Minute)})

	router := gin.New()
	router.GET("/api/v1/inbox", GetInbox)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/inbox", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on inbox, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode inbox payload: %v", err)
	}
	if len(payload.Items) < 3 {
		t.Fatalf("expected at least 3 inbox items, got %d", len(payload.Items))
	}
}

func TestGetMentionsReturnsOnlyDirectMentions(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	users := []domain.User{
		{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"},
		{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"},
	}
	for _, user := range users {
		db.DB.Create(&user)
	}
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-mention", ChannelID: "ch-1", UserID: "user-2", Content: "Looping in Nikko Fu for review", CreatedAt: now.Add(-10 * time.Minute), Metadata: "{}"})
	db.DB.Create(&domain.Message{ID: "msg-other", ChannelID: "ch-1", UserID: "user-2", Content: "No direct mention here", CreatedAt: now.Add(-5 * time.Minute), Metadata: "{}"})

	router := gin.New()
	router.GET("/api/v1/mentions", GetMentions)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mentions", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on mentions, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode mentions payload: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected exactly 1 mention item, got %d", len(payload.Items))
	}
	if payload.Items[0]["type"] != "mention" {
		t.Fatalf("expected mention type, got %#v", payload.Items[0])
	}
}

func TestGetLaterReturnsSavedMessages(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-1",
		UserID:    "user-2",
		Content:   "Save me for later",
		CreatedAt: time.Now().UTC(),
		Metadata:  "{}",
	})
	db.DB.Create(&domain.SavedMessage{
		MessageID: "msg-1",
		UserID:    "user-1",
		CreatedAt: time.Now().UTC(),
	})

	router := gin.New()
	router.GET("/api/v1/later", GetLater)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/later", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on later, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode later payload: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 saved item, got %d", len(payload.Items))
	}
}

func TestGetDraftsReturnsCurrentUserDrafts(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com"})
	db.DB.Create(&domain.Draft{
		UserID:    "user-1",
		Scope:     "channel:ch-1",
		Content:   "Draft for general",
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	})
	db.DB.Create(&domain.Draft{
		UserID:    "user-1",
		Scope:     "dm:dm-1",
		Content:   "Draft for DM",
		CreatedAt: now.Add(-30 * time.Minute),
		UpdatedAt: now.Add(-30 * time.Minute),
	})
	db.DB.Create(&domain.Draft{
		UserID:    "user-2",
		Scope:     "channel:ch-2",
		Content:   "Other user's draft",
		CreatedAt: now,
		UpdatedAt: now,
	})

	router := gin.New()
	router.GET("/api/v1/drafts", GetDrafts)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drafts", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on drafts list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Drafts []domain.Draft `json:"drafts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode drafts payload: %v", err)
	}
	if len(payload.Drafts) != 2 {
		t.Fatalf("expected 2 drafts for current user, got %d", len(payload.Drafts))
	}
	if payload.Drafts[0].Scope != "dm:dm-1" {
		t.Fatalf("expected newest draft first, got %#v", payload.Drafts)
	}
}

func TestPutDraftCreatesAndUpdatesByScope(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})

	router := gin.New()
	router.PUT("/api/v1/drafts/:scope", PutDraft)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/drafts/channel:ch-1", bytes.NewBufferString(`{"content":"First draft"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on draft create, got %d body=%s", rec.Code, rec.Body.String())
	}

	var created struct {
		Draft domain.Draft `json:"draft"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode create draft response: %v", err)
	}
	if created.Draft.Scope != "channel:ch-1" || created.Draft.Content != "First draft" {
		t.Fatalf("unexpected created draft: %#v", created.Draft)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/v1/drafts/channel:ch-1", bytes.NewBufferString(`{"content":"Updated draft"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on draft update, got %d body=%s", rec.Code, rec.Body.String())
	}

	var draft domain.Draft
	if err := db.DB.First(&draft, "user_id = ? AND scope = ?", "user-1", "channel:ch-1").Error; err != nil {
		t.Fatalf("failed to reload draft: %v", err)
	}
	if draft.Content != "Updated draft" {
		t.Fatalf("expected updated content, got %#v", draft)
	}
}

func TestSearchReturnsChannelUserMessageAndDMHits(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-5", WorkspaceID: "ws-1", Name: "ai-lab", Description: "AI experiments", Type: "public"})
	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-5",
		UserID:    "user-2",
		Content:   "The AI lab launch plan is ready",
		CreatedAt: time.Now().UTC(),
		Metadata:  "{}",
	})
	db.DB.Create(&domain.DMConversation{ID: "dm-1", CreatedAt: time.Now().UTC()})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-1"})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-2"})
	db.DB.Create(&domain.DMMessage{
		ID:               "dm-msg-1",
		DMConversationID: "dm-1",
		UserID:           "user-2",
		Content:          "AI follow-up in DM",
		CreatedAt:        time.Now().UTC(),
	})

	router := gin.New()
	router.GET("/api/v1/search", SearchWorkspace)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=AI", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on search, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Query   string `json:"query"`
		Results struct {
			Channels []map[string]any `json:"channels"`
			Users    []map[string]any `json:"users"`
			Messages []map[string]any `json:"messages"`
			DMs      []map[string]any `json:"dms"`
		} `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode search payload: %v", err)
	}
	if payload.Query != "AI" {
		t.Fatalf("expected query echo, got %q", payload.Query)
	}
	if len(payload.Results.Channels) == 0 || len(payload.Results.Users) == 0 || len(payload.Results.Messages) == 0 || len(payload.Results.DMs) == 0 {
		t.Fatalf("expected all result groups to have hits, got %#v", payload.Results)
	}
}

func TestChannelMembersEndpointsListAddAndRemoveMembers(t *testing.T) {
	setupTestDB(t)

	users := []domain.User{
		{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"},
		{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"},
		{ID: "user-3", Name: "Jane Smith", Email: "jane@example.com"},
	}
	for _, user := range users {
		db.DB.Create(&user)
	}
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1", Role: "owner"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-2", Role: "member"})

	router := gin.New()
	router.GET("/api/v1/channels/:id/members", GetChannelMembers)
	router.POST("/api/v1/channels/:id/members", AddChannelMember)
	router.DELETE("/api/v1/channels/:id/members/:userId", RemoveChannelMember)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/ch-1/members", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on members list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Members []map[string]any `json:"members"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode members list: %v", err)
	}
	if len(listPayload.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(listPayload.Members))
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/channels/ch-1/members", bytes.NewBufferString(`{"user_id":"user-3","role":"member"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on add member, got %d body=%s", rec.Code, rec.Body.String())
	}

	var memberCount int64
	db.DB.Model(&domain.ChannelMember{}).Where("channel_id = ?", "ch-1").Count(&memberCount)
	if memberCount != 3 {
		t.Fatalf("expected 3 channel members, got %d", memberCount)
	}

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", "ch-1").Error; err != nil {
		t.Fatalf("failed to reload channel: %v", err)
	}
	if channel.MemberCount != 3 {
		t.Fatalf("expected channel member_count=3, got %d", channel.MemberCount)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/channels/ch-1/members/user-2", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on remove member, got %d body=%s", rec.Code, rec.Body.String())
	}

	db.DB.Model(&domain.ChannelMember{}).Where("channel_id = ?", "ch-1").Count(&memberCount)
	if memberCount != 2 {
		t.Fatalf("expected 2 channel members after delete, got %d", memberCount)
	}
}

func TestPatchChannelUpdatesTopicPurposeAndArchiveState(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.Channel{
		ID:          "ch-1",
		WorkspaceID: "ws-1",
		Name:        "general",
		Type:        "public",
		Description: "General discussion",
	})

	router := gin.New()
	router.PATCH("/api/v1/channels/:id", UpdateChannel)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/channels/ch-1", bytes.NewBufferString(`{"topic":"Launch coordination","purpose":"Keep launch work aligned","is_archived":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on patch channel, got %d body=%s", rec.Code, rec.Body.String())
	}

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", "ch-1").Error; err != nil {
		t.Fatalf("failed to reload channel: %v", err)
	}
	if channel.Topic != "Launch coordination" || channel.Purpose != "Keep launch work aligned" || !channel.IsArchived {
		t.Fatalf("channel patch was not persisted: %#v", channel)
	}
}

func TestWorkspaceInvitesEndpointsCreateAndListInvites(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Acme Corp"})

	router := gin.New()
	router.GET("/api/v1/workspaces/:id/invites", GetWorkspaceInvites)
	router.POST("/api/v1/workspaces/:id/invites", CreateWorkspaceInvite)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/invites", bytes.NewBufferString(`{"email":"new.member@example.com","role":"member"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create invite, got %d body=%s", rec.Code, rec.Body.String())
	}

	var inviteCount int64
	db.DB.Model(&domain.WorkspaceInvite{}).Where("workspace_id = ?", "ws-1").Count(&inviteCount)
	if inviteCount != 1 {
		t.Fatalf("expected 1 workspace invite, got %d", inviteCount)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/invites", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on list invites, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Invites []map[string]any `json:"invites"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode invite list: %v", err)
	}
	if len(listPayload.Invites) != 1 {
		t.Fatalf("expected 1 invite in response, got %d", len(listPayload.Invites))
	}
}

func TestReactionPinAndDeleteBroadcastRealtimeEvents(t *testing.T) {
	setupTestDB(t)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-1",
		UserID:    "user-2",
		Content:   "Hello",
		CreatedAt: time.Now(),
		Metadata:  "{}",
	})

	router := gin.New()
	router.POST("/api/v1/messages/:id/reactions", ToggleReaction)
	router.POST("/api/v1/messages/:id/pin", TogglePinMessage)
	router.DELETE("/api/v1/messages/:id", DeleteMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/msg-1/reactions", bytes.NewBufferString(`{"emoji":"🔥"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on reaction, got %d", rec.Code)
	}
	assertRealtimeEventType(t, client, "reaction.updated")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/messages/msg-1/pin", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on pin, got %d", rec.Code)
	}
	assertRealtimeEventType(t, client, "message.updated")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/messages/msg-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %d", rec.Code)
	}
	assertRealtimeEventType(t, client, "message.deleted")
}

func assertRealtimeEventType(t *testing.T, client *realtime.TestClient, expectedType string) {
	t.Helper()

	raw, err := client.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("failed to receive realtime event: %v", err)
	}

	var event realtime.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("failed to decode realtime event: %v", err)
	}
	if event.Type != expectedType {
		t.Fatalf("expected realtime event %s, got %s", expectedType, event.Type)
	}
}

func setupTestDB(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	testDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite test db: %v", err)
	}
	db.DB = testDB
	if err := db.DB.AutoMigrate(&domain.Organization{}, &domain.Team{}, &domain.User{}, &domain.Agent{}, &domain.Workspace{}, &domain.WorkspaceInvite{}, &domain.Channel{}, &domain.ChannelMember{}, &domain.Message{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	if err := db.DB.AutoMigrate(&domain.MessageReaction{}, &domain.SavedMessage{}, &domain.Draft{}, &domain.UnreadMarker{}, &domain.AIFeedback{}, &domain.DMConversation{}, &domain.DMMember{}, &domain.DMMessage{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
}
