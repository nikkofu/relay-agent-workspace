package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
			"title":     "Test View",
			"view_type": "list",
			"source":    "manual",
			"filters":   map[string]any{"status": "open"},
			"actions":   []map[string]any{{"type": "open"}},
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
			"title":     "Invalid View",
			"view_type": "invalid_type",
			"source":    "manual",
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

	t.Run("List returns next_cursor and honors cursor pagination", func(t *testing.T) {
		base := time.Now().UTC()
		for i, id := range []string{"view-page-1", "view-page-2", "view-page-3"} {
			db.DB.Create(&domain.WorkspaceView{
				ID:        id,
				Title:     id,
				ViewType:  "list",
				Source:    "pagination-test",
				CreatedBy: user.ID,
				CreatedAt: base.Add(time.Duration(i) * time.Second),
				UpdatedAt: base.Add(time.Duration(i) * time.Second),
			})
		}

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/workspace/views?source=pagination-test&limit=2", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var firstPage struct {
			Views      []workspaceViewResponse `json:"views"`
			NextCursor string                  `json:"next_cursor"`
		}
		json.Unmarshal(rec.Body.Bytes(), &firstPage)

		if len(firstPage.Views) != 2 {
			t.Fatalf("expected 2 views, got %d", len(firstPage.Views))
		}
		if firstPage.NextCursor == "" {
			t.Fatal("expected next_cursor")
		}

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/api/v1/workspace/views?source=pagination-test&limit=2&cursor="+firstPage.NextCursor, nil)
		router.ServeHTTP(rec, req)

		var secondPage struct {
			Views      []workspaceViewResponse `json:"views"`
			NextCursor string                  `json:"next_cursor"`
		}
		json.Unmarshal(rec.Body.Bytes(), &secondPage)

		if len(secondPage.Views) != 1 {
			t.Fatalf("expected 1 view on second page, got %d", len(secondPage.Views))
		}
		if secondPage.NextCursor != "" {
			t.Fatalf("expected empty next_cursor on final page, got %q", secondPage.NextCursor)
		}
	})

	t.Run("Patch updates source and view_type with validation", func(t *testing.T) {
		db.DB.Create(&domain.WorkspaceView{
			ID:        "view-patch-source",
			Title:     "Patch Source",
			ViewType:  "list",
			Source:    "manual",
			CreatedBy: user.ID,
		})

		body := map[string]any{
			"source":    "agent",
			"view_type": "report",
		}
		jsonBody, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/workspace/views/view-patch-source", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			View workspaceViewResponse `json:"view"`
		}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		if resp.View.Source != "agent" {
			t.Fatalf("expected source agent, got %s", resp.View.Source)
		}
		if resp.View.ViewType != "report" {
			t.Fatalf("expected view_type report, got %s", resp.View.ViewType)
		}
	})

	t.Run("Reject empty title and nested filters or actions", func(t *testing.T) {
		cases := []map[string]any{
			{"title": "   ", "view_type": "list", "source": "manual"},
			{"title": "Nested Filters", "view_type": "list", "source": "manual", "filters": map[string]any{"owner": map[string]any{"id": "user-1"}}},
			{"title": "Nested Actions", "view_type": "list", "source": "manual", "actions": []map[string]any{{"type": "open", "payload": map[string]any{"x": "y"}}}},
		}

		for _, body := range cases {
			jsonBody, _ := json.Marshal(body)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/workspace/views", bytes.NewBuffer(jsonBody))
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for body %#v, got %d", body, rec.Code)
			}
		}
	})
}
