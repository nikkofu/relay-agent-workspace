package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

func TestPhase73BusinessApps(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	user := domain.User{ID: "user-1", Name: "Nikko Fu"}
	db.DB.Create(&user)

	router := gin.New()
	router.GET("/api/v1/apps", ListApps)
	router.GET("/api/v1/apps/:id", GetApp)
	router.GET("/api/v1/apps/:id/data", GetAppData)
	router.GET("/api/v1/apps/:id/stats", GetAppStats)

	t.Run("GET /api/v1/apps includes sales", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/apps", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp struct {
			Apps []map[string]any `json:"apps"`
		}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		found := false
		for _, app := range resp.Apps {
			if app["id"] == "sales" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'sales' app in list")
		}
	})

	t.Run("GET /api/v1/apps/sales returns metadata", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/sales", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp struct {
			App map[string]any `json:"app"`
		}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		if resp.App["id"] != "sales" {
			t.Errorf("expected app id 'sales', got %v", resp.App["id"])
		}
		if resp.App["modes"] == nil {
			t.Error("expected 'modes' in app metadata")
		}
	})

	t.Run("GET /api/v1/apps/sales/data mode=list", func(t *testing.T) {
		// Seed a sales order
		db.DB.Create(&domain.SalesOrder{
			ID:           "so-1",
			OrderNumber:  "SO-001",
			CustomerName: "Acme Corp",
			Amount:       5000,
			Currency:     "USD",
			Stage:        "proposal",
			Status:       "active",
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/sales/data?mode=list", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp struct {
			Data   []domain.SalesOrder `json:"data"`
			Schema map[string]any       `json:"schema"`
		}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		if len(resp.Data) == 0 {
			t.Error("expected sales order records")
		}
		if resp.Schema == nil {
			t.Error("expected schema in response")
		}
	})

	t.Run("GET /api/v1/apps/sales/stats", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/sales/stats", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp struct {
			Stats map[string]any `json:"stats"`
		}
		json.Unmarshal(rec.Body.Bytes(), &resp)

		if resp.Stats["total_amount"] == nil {
			t.Error("expected 'total_amount' in stats")
		}
	})

	t.Run("GET /api/v1/apps/invalid returns 404", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/unknown", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}
