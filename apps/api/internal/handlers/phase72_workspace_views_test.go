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

func TestPhase72WorkspaceViews(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	user := domain.User{ID: "user-1", Name: "Nikko Fu"}
	db.DB.Create(&user)

	router := gin.New()
	router.GET("/api/v1/workspace/views", ListWorkspaceViews)
	router.POST("/api/v1/workspace/views", CreateWorkspaceView)
	router.GET("/api/v1/workspace/views/:id", GetWorkspaceView)
	router.PATCH("/api/v1/workspace/views/:id", PatchWorkspaceView)

	t.Run("Create valid WorkspaceView", func(t *testing.T) {
		body := map[string]any{
			"title":    "Test View",
			"view_type": "list",
			"source":   "manual",
			"filters":  map[string]any{"status": "open"},
			"actions":  []map[string]any{{"type": "open"}},
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/workspace/views", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			View workspaceViewResponse `json:"view"`
		}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		if resp.View.Title != "Test View" {
			t.Errorf("expected title 'Test View', got %s", resp.View.Title)
		}
		if resp.View.ViewType != "list" {
			t.Errorf("expected view_type 'list', got %s", resp.View.ViewType)
		}
	})

	t.Run("Reject invalid view_type", func(t *testing.T) {
		body := map[string]any{
			"title":    "Invalid View",
			"view_type": "invalid_type",
			"source":   "manual",
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/workspace/views", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("List views with filters", func(t *testing.T) {
		// Create another view
		db.DB.Create(&domain.WorkspaceView{
			ID:       "view-calendar-1",
			Title:    "Calendar View",
			ViewType: "calendar",
			Source:   "agent",
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/workspace/views?view_type=calendar", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp struct {
			Views []workspaceViewResponse `json:"views"`
		}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		if len(resp.Views) == 0 {
			t.Fatal("expected at least one view")
		}
		for _, v := range resp.Views {
			if v.ViewType != "calendar" {
				t.Errorf("expected only calendar views, got %s", v.ViewType)
			}
		}
	})
}
