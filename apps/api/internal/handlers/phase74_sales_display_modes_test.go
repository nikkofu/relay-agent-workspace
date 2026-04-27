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

func TestPhase74SalesDisplayModes(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	user := domain.User{ID: "user-1", Name: "Nikko Fu"}
	db.DB.Create(&user)

	router := gin.New()
	router.GET("/api/v1/apps/:id", GetApp)
	router.GET("/api/v1/apps/:id/data", GetAppData)

	t.Run("GET /api/v1/apps/sales returns new modes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/sales", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp struct {
			App struct {
				Modes []string `json:"modes"`
			} `json:"app"`
		}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		expectedModes := map[string]bool{
			"list":      true,
			"card_grid": true,
			"kanban":    true,
			"calendar":  true,
			"stat":      true,
		}

		for _, m := range resp.App.Modes {
			delete(expectedModes, m)
		}

		if len(expectedModes) > 0 {
			t.Errorf("missing expected modes: %v", expectedModes)
		}

		// Ensure 'search' and 'stats' are removed from metadata list
		for _, m := range resp.App.Modes {
			if m == "search" || m == "stats" {
				t.Errorf("deprecated mode '%s' still in metadata", m)
			}
		}
	})

	t.Run("search mode aliases to list", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/sales/data?mode=search", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 for 'search' alias, got %d", rec.Code)
		}
	})

	t.Run("stats mode aliases to stat", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/sales/data?mode=stats", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 for 'stats' alias, got %d", rec.Code)
		}
	})

	t.Run("invalid mode returns 400", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/sales/data?mode=invalid", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for invalid mode, got %d", rec.Code)
		}
	})

	t.Run("calendar mode returns projected events", func(t *testing.T) {
		// Ensure deterministic order exists
		db.DB.Create(&domain.SalesOrder{
			ID:                "so-cal-1",
			OrderNumber:       "SO-CAL-001",
			CustomerName:      "Cal Corp",
			Amount:            1000,
			ExpectedCloseDate: toPtrTime(time.Now().Add(24 * time.Hour)),
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/sales/data?mode=calendar&calendar_time_field=expected_close_date", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp struct {
			CalendarEvents []map[string]any `json:"calendar_events"`
		}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		if len(resp.CalendarEvents) == 0 {
			t.Error("expected calendar events")
		}

		event := resp.CalendarEvents[0]
		if event["id"] == nil || event["start"] == nil || event["title"] == nil {
			t.Errorf("event missing required fields: %v", event)
		}
		if event["record_id"] == nil {
			t.Error("event missing record_id")
		}
	})

	t.Run("calendar mode invalid time field returns 400", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/sales/data?mode=calendar&calendar_time_field=invalid_field", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for invalid calendar_time_field, got %d", rec.Code)
		}
	})

	router.GET("/api/v1/apps/:id/stats", GetAppStats)

	t.Run("stats returns funnel and timeline buckets", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/sales/stats?chart_style=funnel", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp struct {
			Stats struct {
				Funnel []map[string]any `json:"funnel"`
				TimelineBuckets []map[string]any `json:"timeline_buckets"`
			} `json:"stats"`
		}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		if len(resp.Stats.Funnel) == 0 {
			t.Error("expected funnel data")
		}
		if len(resp.Stats.TimelineBuckets) == 0 {
			t.Error("expected timeline buckets")
		}
	})

	t.Run("stats invalid chart style returns 400", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/sales/stats?chart_style=invalid", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for invalid chart_style, got %d", rec.Code)
		}
	})
}
