package handlers

import (
	"encoding/json"
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

func setupTestDB(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	testDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite test db: %v", err)
	}
	db.DB = testDB
	if err := db.DB.AutoMigrate(&domain.Organization{}, &domain.Team{}, &domain.User{}, &domain.Agent{}, &domain.Workspace{}, &domain.Channel{}, &domain.Message{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
}
