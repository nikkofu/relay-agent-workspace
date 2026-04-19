package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

func TestArtifactCRUDAndAI_generate(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	AIGateway = stubGateway{}

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-5", WorkspaceID: "ws-1", Name: "ai-lab", Type: "public"})

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	router := gin.New()
	router.GET("/api/v1/artifacts", GetArtifacts)
	router.POST("/api/v1/artifacts", CreateArtifact)
	router.GET("/api/v1/artifacts/:id", GetArtifact)
	router.PATCH("/api/v1/artifacts/:id", UpdateArtifact)
	router.POST("/api/v1/ai/canvas/generate", GenerateCanvasArtifact)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/artifacts", bytes.NewBufferString(`{"channel_id":"ch-5","title":"Launch Notes","content":"Initial outline"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on artifact create, got %d body=%s", rec.Code, rec.Body.String())
	}

	var createPayload struct {
		Artifact struct {
			domain.Artifact
			CreatedByUser *domain.User `json:"created_by_user"`
			UpdatedByUser *domain.User `json:"updated_by_user"`
		} `json:"artifact"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("failed to decode create artifact payload: %v", err)
	}
	if createPayload.Artifact.Source != "manual" || createPayload.Artifact.Type != "document" {
		t.Fatalf("unexpected manual artifact payload: %#v", createPayload.Artifact)
	}
	if createPayload.Artifact.CreatedByUser == nil || createPayload.Artifact.CreatedByUser.ID != "user-1" {
		t.Fatalf("expected created_by_user hydration, got %#v", createPayload.Artifact.CreatedByUser)
	}
	assertRealtimeEventType(t, client, "artifact.updated")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/artifacts/"+createPayload.Artifact.ID, bytes.NewBufferString(`{"status":"live","content":"Revised outline"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on artifact update, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/artifacts/"+createPayload.Artifact.ID, nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on artifact get, got %d body=%s", rec.Code, rec.Body.String())
	}

	var detailPayload struct {
		Artifact struct {
			domain.Artifact
			UpdatedByUser *domain.User `json:"updated_by_user"`
		} `json:"artifact"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &detailPayload); err != nil {
		t.Fatalf("failed to decode artifact detail: %v", err)
	}
	if detailPayload.Artifact.Status != "live" || detailPayload.Artifact.Content != "Revised outline" {
		t.Fatalf("unexpected artifact detail payload: %#v", detailPayload.Artifact)
	}
	if detailPayload.Artifact.UpdatedByUser == nil || detailPayload.Artifact.UpdatedByUser.ID != "user-1" {
		t.Fatalf("expected updated_by_user hydration, got %#v", detailPayload.Artifact.UpdatedByUser)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/artifacts?channel_id=ch-5", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on artifact list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Artifacts []struct {
			domain.Artifact
			UpdatedByUser *domain.User `json:"updated_by_user"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode artifact list: %v", err)
	}
	if len(listPayload.Artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(listPayload.Artifacts))
	}
	if listPayload.Artifacts[0].UpdatedByUser == nil {
		t.Fatalf("expected list artifact to include updated_by_user")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/generate", bytes.NewBufferString(`{"channel_id":"ch-5","prompt":"Create a launch checklist","title":"AI Launch Checklist","provider":"gemini"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on ai canvas generate, got %d body=%s", rec.Code, rec.Body.String())
	}

	var generatePayload struct {
		Artifact struct {
			domain.Artifact
			UpdatedByUser *domain.User `json:"updated_by_user"`
		} `json:"artifact"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &generatePayload); err != nil {
		t.Fatalf("failed to decode generated artifact payload: %v", err)
	}
	if generatePayload.Artifact.Source != "ai" || generatePayload.Artifact.Provider != "stub" {
		t.Fatalf("unexpected generated artifact payload: %#v", generatePayload.Artifact)
	}
	if !strings.Contains(generatePayload.Artifact.Content, "Thinking...") || !strings.Contains(generatePayload.Artifact.Content, "Done.") {
		t.Fatalf("expected generated artifact content from stream, got %#v", generatePayload.Artifact.Content)
	}
	if generatePayload.Artifact.UpdatedByUser == nil || generatePayload.Artifact.UpdatedByUser.ID != "user-1" {
		t.Fatalf("expected generated artifact updated_by_user hydration, got %#v", generatePayload.Artifact.UpdatedByUser)
	}
	assertRealtimeEventType(t, client, "artifact.updated")

	var count int64
	db.DB.Model(&domain.Artifact{}).Count(&count)
	if count != 2 {
		t.Fatalf("expected 2 persisted artifacts, got %d", count)
	}
}

func TestActivityFeedUsesStableUniqueReactionIDs(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"})
	db.DB.Create(&domain.User{ID: "user-3", Name: "Jane Smith", Email: "jane@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-1",
		UserID:    "user-1",
		Content:   "Launch update",
		CreatedAt: now,
		Metadata:  "{}",
	})
	db.DB.Create(&domain.MessageReaction{MessageID: "msg-1", UserID: "user-2", Emoji: "🔥", CreatedAt: now.Add(time.Minute)})
	db.DB.Create(&domain.MessageReaction{MessageID: "msg-1", UserID: "user-3", Emoji: "🔥", CreatedAt: now.Add(2 * time.Minute)})

	feed := buildActivityFeed(domain.User{ID: "user-1", Name: "Nikko Fu"})
	ids := map[string]struct{}{}
	reactionCount := 0
	for _, item := range feed {
		if item.Type != "reaction" {
			continue
		}
		reactionCount++
		if _, exists := ids[item.ID]; exists {
			t.Fatalf("expected unique reaction activity ids, got duplicate %q", item.ID)
		}
		ids[item.ID] = struct{}{}
	}
	if reactionCount != 2 {
		t.Fatalf("expected 2 reaction activity items, got %d", reactionCount)
	}
}

func TestFileUploadListAndDetail(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})

	_ = os.RemoveAll("uploads")
	t.Cleanup(func() {
		_ = os.RemoveAll("uploads")
	})

	router := gin.New()
	router.POST("/api/v1/files/upload", UploadFile)
	router.GET("/api/v1/files", ListFiles)
	router.GET("/api/v1/files/:id", GetFile)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("channel_id", "ch-5")
	fileWriter, err := writer.CreateFormFile("file", "notes.txt")
	if err != nil {
		t.Fatalf("failed to create multipart file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("hello relay files")); err != nil {
		t.Fatalf("failed to write multipart contents: %v", err)
	}
	_ = writer.Close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on file upload, got %d body=%s", rec.Code, rec.Body.String())
	}

	var uploadPayload struct {
		File struct {
			domain.FileAsset
			Uploader *domain.User `json:"uploader"`
			URL      string       `json:"url"`
		} `json:"file"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &uploadPayload); err != nil {
		t.Fatalf("failed to decode upload payload: %v", err)
	}
	if uploadPayload.File.Uploader == nil || uploadPayload.File.Uploader.ID != "user-1" {
		t.Fatalf("expected uploader hydration, got %#v", uploadPayload.File.Uploader)
	}
	if uploadPayload.File.URL == "" {
		t.Fatalf("expected file url in upload payload")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/files?channel_id=ch-5", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on file list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Files []map[string]any `json:"files"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode file list payload: %v", err)
	}
	if len(listPayload.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(listPayload.Files))
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/files/"+uploadPayload.File.ID, nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on file detail, got %d body=%s", rec.Code, rec.Body.String())
	}
}
