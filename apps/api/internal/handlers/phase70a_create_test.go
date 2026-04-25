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
)

func TestPhase70ACreate(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	user := domain.User{ID: "user-1", Name: "Nikko Fu"}
	db.DB.Create(&user)
	workspace := domain.Workspace{ID: "ws-1", Name: "Relay"}
	db.DB.Create(&workspace)
	channel := domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "ops"}
	db.DB.Create(&channel)

	// Create a mock draft
	itemTitles := []string{"Task 1", "Task 2"}
	itemsJSON, _ := json.Marshal(itemTitles)
	draft := domain.AnalysisListDraft{
		ID:                 "draft-1",
		ArtifactID:         "art-1",
		ChannelID:          "ch-1",
		AnalysisSnapshotID: "msg-1",
		Title:              "Launch Plan",
		ItemsJSON:          string(itemsJSON),
		CreatedBy:          "user-1",
	}
	db.DB.Create(&draft)

	router := gin.New()
	router.POST("/api/v1/ai/canvas/confirm-create-list", ConfirmCreateListFromDraft)

	t.Run("Valid Draft ID -> List Created", func(t *testing.T) {
		body := map[string]any{
			"draft_id": "draft-1",
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/confirm-create-list", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			ListID string `json:"list_id"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.ListID == "" {
			t.Error("expected non-empty list_id")
		}

		// Verify list and items exist in DB
		var list domain.WorkspaceList
		if err := db.DB.First(&list, "id = ?", resp.ListID).Error; err != nil {
			t.Fatalf("list not found in DB: %v", err)
		}
		if list.Title != "Launch Plan" {
			t.Errorf("expected title 'Launch Plan', got '%s'", list.Title)
		}

		var itemCount int64
		db.DB.Model(&domain.WorkspaceListItem{}).Where("list_id = ?", resp.ListID).Count(&itemCount)
		if itemCount != 2 {
			t.Errorf("expected 2 items, got %d", itemCount)
		}
	})

	t.Run("Missing Draft ID -> 404", func(t *testing.T) {
		body := map[string]any{
			"draft_id": "draft-missing",
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/confirm-create-list", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("Empty Draft ID -> 400", func(t *testing.T) {
		body := map[string]any{
			"draft_id": "",
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/canvas/confirm-create-list", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}
