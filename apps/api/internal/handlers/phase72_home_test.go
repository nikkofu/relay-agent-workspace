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

func TestPhase72HomeWorkbenchAggregation(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	// Create test data
	user := domain.User{ID: "user-1", Name: "Nikko Fu"}
	db.DB.Create(&user)

	router := gin.New()
	router.GET("/api/v1/home", GetHome)

	t.Run("Home response includes new workbench sections", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/home", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var payload struct {
			Home struct {
				Today struct {
					Items []any `json:"items"`
				} `json:"today"`
				MyWork struct {
					Items []any `json:"items"`
				} `json:"my_work"`
				RecentChannels struct {
					Items []any `json:"items"`
				} `json:"recent_channels"`
				AISuggestions struct {
					Items []any `json:"items"`
				} `json:"ai_suggestions"`
				AppsTools struct {
					Items []any `json:"items"`
				} `json:"apps_tools"`
				Activity struct {
					Items []any `json:"items"`
				} `json:"activity"`
			} `json:"home"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Verify sections exist (even if empty)
		// Today
		if payload.Home.Today.Items == nil {
			t.Error("expected today.items section")
		}
		// MyWork
		if payload.Home.MyWork.Items == nil {
			t.Error("expected my_work.items section")
		}
		// RecentChannels
		if payload.Home.RecentChannels.Items == nil {
			t.Error("expected recent_channels.items section")
		}
		// AISuggestions
		if payload.Home.AISuggestions.Items == nil {
			t.Error("expected ai_suggestions.items section")
		}
		// AppsTools
		if payload.Home.AppsTools.Items == nil {
			t.Error("expected apps_tools.items section")
		}
		// Activity (new events list)
		if payload.Home.Activity.Items == nil {
			t.Error("expected activity.items section")
		}
	})

	t.Run("Home response maintains backwards compatibility", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/home", nil)
		router.ServeHTTP(rec, req)

		var payload map[string]any
		json.Unmarshal(rec.Body.Bytes(), &payload)
		
		home := payload["home"].(map[string]any)
		
		// Check for existing fields
		fields := []string{
			"user", "profile", "stats", "starred_channels", "recent_dms",
			"drafts", "tools", "workflows", "recent_activity", "recent_artifacts",
			"recent_lists", "recent_tool_runs", "recent_files", "open_list_work",
			"tool_runs_needing_attention", "channel_execution_pulse", "recent_ai_executions",
			"knowledge_inbox_count", "recent_knowledge_digests",
		}
		
		for _, f := range fields {
			if _, ok := home[f]; !ok {
				t.Errorf("missing backwards-compatible field: %s", f)
			}
		}
		
		// Check that activity still has its summary fields
		activity := home["activity"].(map[string]any)
		if _, ok := activity["unread_count"]; !ok {
			t.Error("missing unread_count in activity summary")
		}
	})
}
