package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
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

func setupTestDB(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	testDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite test db: %v", err)
	}
	db.DB = testDB
	if err := db.DB.AutoMigrate(&domain.Organization{}, &domain.Team{}, &domain.User{}, &domain.Agent{}, &domain.Workspace{}, &domain.Channel{}, &domain.Message{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	if err := db.DB.AutoMigrate(&domain.MessageReaction{}, &domain.SavedMessage{}, &domain.UnreadMarker{}, &domain.AIFeedback{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
}
