package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

func TestPhase68FilePayloadNormalization(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", Name: "general", WorkspaceID: "ws-1"})
	db.DB.Create(&domain.FileAsset{
		ID:          "file-1",
		ChannelID:   "ch-1",
		UploaderID:  "user-1",
		Name:        "notes.md",
		ContentType: "text/markdown",
		SizeBytes:   1024,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	router := gin.New()
	router.GET("/api/v1/files", ListFiles)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Files []struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			MimeType string `json:"mime_type"`
			Size     int64  `json:"size"`
		} `json:"files"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if len(payload.Files) == 0 {
		t.Fatal("expected 1 file, got 0")
	}

	f := payload.Files[0]
	if f.ID != "file-1" || f.Title != "notes.md" || f.MimeType != "text/markdown" || f.Size != 1024 {
		t.Fatalf("normalization fields mismatch: %#v", f)
	}
}

func TestPhase68MessageFilesNormalization(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", Name: "general", WorkspaceID: "ws-1"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "here is a file", CreatedAt: time.Now()})
	db.DB.Create(&domain.FileAsset{
		ID:          "file-2",
		ChannelID:   "ch-1",
		UploaderID:  "user-1",
		Name:        "data.csv",
		ContentType: "text/csv",
		SizeBytes:   2048,
	})
	db.DB.Create(&domain.MessageFileAttachment{MessageID: "msg-1", FileID: "file-2"})

	router := gin.New()
	router.GET("/api/v1/messages/:id/files", GetMessageFiles)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/msg-1/files", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Files []struct {
			Title          string `json:"title"`
			MimeType       string `json:"mime_type"`
			LegacyMimeType string `json:"mimeType"`
		} `json:"files"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if len(payload.Files) == 0 {
		t.Fatal("expected 1 file, got 0")
	}

	f := payload.Files[0]
	if f.Title != "data.csv" || f.MimeType != "text/csv" || f.LegacyMimeType != "text/csv" {
		t.Fatalf("normalization fields mismatch: %#v", f)
	}
}

func TestPhase68FileSearchNormalization(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.FileAsset{
		ID:               "file-3",
		Name:             "report.pdf",
		ContentType:      "application/pdf",
		SizeBytes:        512,
		ExtractionStatus: "ready",
	})

	router := gin.New()
	router.GET("/api/v1/search/files", SearchFiles)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/files?q=report", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Results []struct {
			File struct {
				Title    string `json:"title"`
				MimeType string `json:"mime_type"`
			} `json:"file"`
		} `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if len(payload.Results) == 0 {
		t.Fatal("expected 1 search result, got 0")
	}

	f := payload.Results[0].File
	if f.Title != "report.pdf" || f.MimeType != "application/pdf" {
		t.Fatalf("normalization fields mismatch: %#v", f)
	}
}

func TestPhase68MessageListMetadataNormalization(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", Name: "general", WorkspaceID: "ws-1"})
	db.DB.Create(&domain.Message{
		ID:        "msg-2",
		ChannelID: "ch-1",
		UserID:    "user-1",
		Content:   "shared a file",
		CreatedAt: time.Now(),
		Metadata:  "{}", // Will be refreshed
	})
	db.DB.Create(&domain.FileAsset{
		ID:          "file-4",
		ChannelID:   "ch-1",
		UploaderID:  "user-1",
		Name:        "budget.xlsx",
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		SizeBytes:   4096,
	})
	db.DB.Create(&domain.MessageFileAttachment{MessageID: "msg-2", FileID: "file-4"})

	router := gin.New()
	router.GET("/api/v1/channels/:id/messages", GetMessages)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/ch-1/messages?channel_id=ch-1", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Messages []struct {
			Metadata string `json:"metadata"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if len(payload.Messages) == 0 {
		t.Fatal("expected 1 message, got 0")
	}

	var meta struct {
		Attachments []struct {
			Kind     string `json:"kind"`
			Title    string `json:"title"`
			MimeType string `json:"mime_type"`
		} `json:"attachments"`
	}
	if err := json.Unmarshal([]byte(payload.Messages[0].Metadata), &meta); err != nil {
		t.Fatalf("failed to decode metadata string: %v", err)
	}

	found := false
	for _, att := range meta.Attachments {
		if att.Kind == "file" && att.Title == "budget.xlsx" && att.MimeType == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("normalized file attachment not found in message metadata")
	}
}
