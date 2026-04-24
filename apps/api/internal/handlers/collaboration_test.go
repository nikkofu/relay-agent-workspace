package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
	"github.com/nikkofu/relay-agent-workspace/api/internal/knowledge"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
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
	if payload.Users[0].AIInsight == "" {
		t.Fatal("expected ai_insight to be populated")
	}
}

func TestGetUsersSupportsDirectoryFilters(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com", Department: "Product", Timezone: "Asia/Shanghai", Status: "online"})
	db.DB.Create(&domain.User{ID: "user-2", OrganizationID: "org-1", Name: "Jane Smith", Email: "jane@example.com", Department: "Design", Timezone: "Europe/London", Status: "away"})
	db.DB.Create(&domain.User{ID: "user-3", OrganizationID: "org-1", Name: "John Doe", Email: "john@example.com", Department: "Product", Timezone: "America/Los_Angeles", Status: "offline"})
	db.DB.Create(&domain.UserGroup{ID: "group-1", WorkspaceID: "ws-1", Name: "Product Leads", Handle: "product-leads"})
	db.DB.Create(&domain.UserGroupMember{UserGroupID: "group-1", UserID: "user-1", Role: "owner"})
	db.DB.Create(&domain.UserGroupMember{UserGroupID: "group-1", UserID: "user-3", Role: "member"})

	router := gin.New()
	router.GET("/api/v1/users", GetUsers)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?department=Product&status=online&user_group_id=group-1&q=nikko", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Users []domain.User `json:"users"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(payload.Users) != 1 || payload.Users[0].ID != "user-1" {
		t.Fatalf("expected only filtered user-1, got %#v", payload.Users)
	}
}

func TestGetUserProfileReturnsHydratedProfile(t *testing.T) {
	setupTestDB(t)

	lastSeen := time.Now().Add(-30 * time.Minute).UTC()
	db.DB.Create(&domain.User{
		ID:             "user-1",
		OrganizationID: "org-1",
		Name:           "Nikko Fu",
		Email:          "nikko@example.com",
		Avatar:         "https://example.com/avatar.png",
		Status:         "away",
		StatusText:     "Reviewing launch plans",
		LastSeenAt:     &lastSeen,
		AIProvider:     "gemini",
		AIModel:        "gemini-3-flash-preview",
		AIMode:         "planning",
	})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public", IsStarred: true})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1", Role: "owner"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Profile activity", CreatedAt: time.Now().Add(-time.Hour).UTC()})
	db.DB.Create(&domain.Artifact{ID: "artifact-1", ChannelID: "ch-1", Title: "Launch Brief", Version: 1, Type: "document", Status: "live", Content: "hello", Source: "manual", CreatedBy: "user-1", UpdatedBy: "user-1", CreatedAt: time.Now().Add(-time.Hour).UTC(), UpdatedAt: time.Now().UTC()})

	router := gin.New()
	router.GET("/api/v1/users/:id", GetUserProfile)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user-1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		User struct {
			ID         string `json:"id"`
			Status     string `json:"status"`
			StatusText string `json:"status_text"`
			AIInsight  string `json:"ai_insight"`
			LastSeenAt string `json:"last_seen_at"`
			Profile    struct {
				LocalTime       string   `json:"local_time"`
				WorkingHours    string   `json:"working_hours"`
				FocusAreas      []string `json:"focus_areas"`
				TopChannels     []any    `json:"top_channels"`
				RecentArtifacts []any    `json:"recent_artifacts"`
			} `json:"profile"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode user profile payload: %v", err)
	}
	if payload.User.ID != "user-1" || payload.User.AIInsight == "" {
		t.Fatalf("unexpected hydrated user payload: %#v", payload.User)
	}
	if payload.User.Profile.LocalTime == "" || payload.User.Profile.WorkingHours == "" {
		t.Fatalf("expected profile metadata, got %#v", payload.User.Profile)
	}
	if len(payload.User.Profile.TopChannels) != 1 || len(payload.User.Profile.RecentArtifacts) != 1 {
		t.Fatalf("expected profile aggregates, got %#v", payload.User.Profile)
	}
}

func TestPatchUserStatusUpdatesPresenceFields(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{
		ID:             "user-1",
		OrganizationID: "org-1",
		Name:           "Nikko Fu",
		Email:          "nikko@example.com",
		Status:         "offline",
	})

	router := gin.New()
	router.PATCH("/api/v1/users/:id/status", PatchUserStatus)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/user-1/status", bytes.NewBufferString(`{"status":"busy","status_text":"Heads down on API design"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var refreshed domain.User
	if err := db.DB.First(&refreshed, "id = ?", "user-1").Error; err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if refreshed.Status != "busy" || refreshed.StatusText != "Heads down on API design" {
		t.Fatalf("expected status update to persist, got %#v", refreshed)
	}
	if refreshed.LastSeenAt == nil {
		t.Fatal("expected last_seen_at to be set")
	}
}

func TestPatchUserStatusSupportsEmojiAndExpiration(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{
		ID:             "user-1",
		OrganizationID: "org-1",
		Name:           "Nikko Fu",
		Email:          "nikko@example.com",
		Status:         "online",
	})

	router := gin.New()
	router.PATCH("/api/v1/users/:id/status", PatchUserStatus)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/user-1/status", bytes.NewBufferString(`{"status":"busy","status_text":"Heads down","status_emoji":"🧠","expires_in_minutes":30}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var refreshed domain.User
	if err := db.DB.First(&refreshed, "id = ?", "user-1").Error; err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if refreshed.StatusEmoji != "🧠" {
		t.Fatalf("expected status emoji to persist, got %#v", refreshed)
	}
	if refreshed.StatusExpiresAt == nil {
		t.Fatal("expected status_expires_at to be set")
	}
}

func TestPatchUserProfileUpdatesWorkspaceIdentityFields(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{
		ID:             "user-1",
		OrganizationID: "org-1",
		Name:           "Nikko Fu",
		Email:          "nikko@example.com",
		Title:          "Founder",
		Department:     "Product",
		Timezone:       "Asia/Shanghai",
		WorkingHours:   "Mon - Fri",
	})

	router := gin.New()
	router.PATCH("/api/v1/users/:id", PatchUserProfile)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/user-1", bytes.NewBufferString(`{"title":"CEO","department":"Executive","timezone":"America/New_York","working_hours":"Mon - Thu, 09:00 - 18:00"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var refreshed domain.User
	if err := db.DB.First(&refreshed, "id = ?", "user-1").Error; err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if refreshed.Title != "CEO" || refreshed.Department != "Executive" || refreshed.Timezone != "America/New_York" || refreshed.WorkingHours != "Mon - Thu, 09:00 - 18:00" {
		t.Fatalf("expected profile fields to persist, got %#v", refreshed)
	}
}

func TestPatchUserProfileSupportsExtendedFields(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{
		ID:             "user-1",
		OrganizationID: "org-1",
		Name:           "Nikko Fu",
		Email:          "nikko@example.com",
	})

	router := gin.New()
	router.PATCH("/api/v1/users/:id", PatchUserProfile)
	router.GET("/api/v1/users/:id", GetUserProfile)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/user-1", bytes.NewBufferString(`{"pronouns":"he/him","location":"Shanghai","phone":"+86 13800000000","bio":"Building Relay with humans and agents."}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on extended profile patch, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/user-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on user detail, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		User struct {
			Pronouns string `json:"pronouns"`
			Location string `json:"location"`
			Phone    string `json:"phone"`
			Bio      string `json:"bio"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode extended user payload: %v", err)
	}
	if payload.User.Pronouns != "he/him" || payload.User.Location != "Shanghai" || payload.User.Phone == "" || payload.User.Bio == "" {
		t.Fatalf("unexpected extended user payload: %#v", payload.User)
	}
}

func TestGetHomeReturnsWorkspaceSummary(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com", Status: "online", AIProvider: "gemini", AIModel: "gemini-3-flash-preview", AIMode: "planning"})
	db.DB.Create(&domain.User{ID: "user-2", OrganizationID: "org-1", Name: "AI Assistant", Email: "ai@example.com", Status: "online"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public", IsStarred: true})
	db.DB.Create(&domain.Channel{ID: "ch-2", WorkspaceID: "ws-1", Name: "ai-lab", Type: "public"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1", Role: "owner"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-2", UserID: "user-1", Role: "member"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-2", Content: "Hello @Nikko Fu", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.SavedMessage{MessageID: "msg-1", UserID: "user-1"})
	db.DB.Create(&domain.Draft{UserID: "user-1", Scope: "channel:ch-1", Content: "Need to reply", CreatedAt: now.Add(-30 * time.Minute), UpdatedAt: now.Add(-15 * time.Minute)})
	db.DB.Create(&domain.DMConversation{ID: "dm-1", CreatedAt: now.Add(-time.Hour)})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-1"})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-2"})
	db.DB.Create(&domain.DMMessage{ID: "dm-msg-1", DMConversationID: "dm-1", UserID: "user-2", Content: "Quick sync?", CreatedAt: now.Add(-20 * time.Minute)})
	db.DB.Create(&domain.WorkflowDefinition{ID: "wf-1", Name: "Daily Standup", Category: "communication", Description: "Collect quick progress updates", Trigger: "manual", IsActive: true, CreatedAt: now.Add(-time.Hour), UpdatedAt: now})
	db.DB.Create(&domain.ToolDefinition{ID: "tool-1", Name: "Web Search", Key: "web-search", Category: "research", Description: "Search current information", Icon: "search", IsEnabled: true, CreatedAt: now.Add(-time.Hour), UpdatedAt: now})
	db.DB.Create(&domain.Artifact{ID: "artifact-1", ChannelID: "ch-1", Title: "Launch Brief", Version: 1, Type: "document", Status: "live", Content: "hello", Source: "manual", CreatedBy: "user-1", UpdatedBy: "user-1", CreatedAt: now.Add(-25 * time.Minute), UpdatedAt: now.Add(-5 * time.Minute)})
	db.DB.Create(&domain.WorkspaceList{ID: "list-1", WorkspaceID: "ws-1", ChannelID: "ch-1", Title: "Launch Checklist", Description: "Critical preflight items", CreatedBy: "user-2", CreatedAt: now.Add(-12 * time.Minute), UpdatedAt: now.Add(-10 * time.Minute)})
	db.DB.Create(&domain.WorkspaceListItem{ID: 1, ListID: "list-1", Content: "Review API contract", Position: 1, IsCompleted: true, AssignedTo: "user-1", CompletedAt: ptrTime(now.Add(-9 * time.Minute)), CreatedBy: "user-2", CreatedAt: now.Add(-12 * time.Minute), UpdatedAt: now.Add(-9 * time.Minute)})
	toolInput := `{"channel_id":"ch-1","thread_id":"msg-1"}`
	db.DB.Create(&domain.ToolRun{ID: "toolrun-1", ToolID: "tool-1", TriggeredBy: "user-1", Status: "success", Input: toolInput, Summary: "Executed Web Search", StartedAt: now.Add(-8 * time.Minute), CompletedAt: ptrTime(now.Add(-7 * time.Minute)), CreatedAt: now.Add(-8 * time.Minute), UpdatedAt: now.Add(-7 * time.Minute)})
	db.DB.Create(&domain.ToolRunLog{ToolRunID: "toolrun-1", Level: "info", Message: "Tool execution completed", CreatedAt: now.Add(-7 * time.Minute)})
	db.DB.Create(&domain.FileAsset{ID: "file-1", ChannelID: "ch-1", UploaderID: "user-2", Name: "handoff.md", StoragePath: "handoff.md", ContentType: "text/markdown", SizeBytes: 128, CreatedAt: now.Add(-6 * time.Minute), UpdatedAt: now.Add(-6 * time.Minute)})

	router := gin.New()
	router.GET("/api/v1/home", GetHome)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/home", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Home struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
			Activity struct {
				UnreadCount int `json:"unread_count"`
				DraftCount  int `json:"draft_count"`
				DMCount     int `json:"dm_count"`
			} `json:"activity"`
			StarredChannels []any `json:"starred_channels"`
			RecentDMs       []any `json:"recent_dms"`
			Drafts          []any `json:"drafts"`
			Tools           []any `json:"tools"`
			Workflows       []any `json:"workflows"`
			Stats           struct {
				PendingActions int `json:"pending_actions"`
				ActiveThreads  int `json:"active_threads"`
			} `json:"stats"`
			RecentActivity []struct {
				ID          string `json:"id"`
				ChannelID   string `json:"channel_id"`
				ChannelName string `json:"channel_name"`
				LastMessage string `json:"last_message"`
			} `json:"recent_activity"`
			RecentArtifacts []struct {
				ID string `json:"id"`
			} `json:"recent_artifacts"`
			RecentLists []struct {
				ID             string `json:"id"`
				CompletedCount int    `json:"completed_count"`
			} `json:"recent_lists"`
			RecentToolRuns []struct {
				ID        string `json:"id"`
				ChannelID string `json:"channel_id"`
			} `json:"recent_tool_runs"`
			RecentFiles []struct {
				ID        string `json:"id"`
				UserID    string `json:"userId"`
				ChannelID string `json:"channelId"`
			} `json:"recent_files"`
		} `json:"home"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode home payload: %v", err)
	}
	if payload.Home.User.ID != "user-1" {
		t.Fatalf("expected user-1 in home payload, got %#v", payload.Home.User)
	}
	if payload.Home.Activity.DraftCount != 1 || payload.Home.Activity.DMCount != 1 {
		t.Fatalf("unexpected home activity summary: %#v", payload.Home.Activity)
	}
	if payload.Home.Stats.PendingActions != payload.Home.Activity.UnreadCount || payload.Home.Stats.ActiveThreads < 0 {
		t.Fatalf("unexpected home stats payload: %#v", payload.Home.Stats)
	}
	if len(payload.Home.StarredChannels) != 1 || len(payload.Home.RecentDMs) != 1 || len(payload.Home.Tools) != 1 || len(payload.Home.Workflows) != 1 {
		t.Fatalf("expected populated home sections, got %#v", payload.Home)
	}
	if len(payload.Home.RecentActivity) == 0 || payload.Home.RecentActivity[0].ChannelID == "" || payload.Home.RecentActivity[0].ChannelName == "" || payload.Home.RecentActivity[0].LastMessage == "" {
		t.Fatalf("expected recent activity aliases for home dashboard, got %#v", payload.Home.RecentActivity)
	}
	if len(payload.Home.RecentArtifacts) == 0 || payload.Home.RecentArtifacts[0].ID == "" {
		t.Fatalf("expected top-level recent artifacts alias, got %#v", payload.Home.RecentArtifacts)
	}
	if len(payload.Home.RecentLists) != 1 || payload.Home.RecentLists[0].CompletedCount != 1 {
		t.Fatalf("expected recent list summary, got %#v", payload.Home.RecentLists)
	}
	if len(payload.Home.RecentToolRuns) != 1 || payload.Home.RecentToolRuns[0].ChannelID != "ch-1" {
		t.Fatalf("expected recent tool run summary, got %#v", payload.Home.RecentToolRuns)
	}
	if len(payload.Home.RecentFiles) != 1 || payload.Home.RecentFiles[0].UserID != "user-2" || payload.Home.RecentFiles[0].ChannelID != "ch-1" {
		t.Fatalf("expected recent files summary, got %#v", payload.Home.RecentFiles)
	}
}

func TestGetUserGroupsAndDetail(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com", Status: "online"})
	db.DB.Create(&domain.User{ID: "user-2", OrganizationID: "org-1", Name: "Jane Smith", Email: "jane@example.com", Status: "away"})
	db.DB.Create(&domain.UserGroup{ID: "group-1", WorkspaceID: "ws-1", Name: "Design Leads", Handle: "design-leads", Description: "Cross-functional design leadership", CreatedBy: "user-1", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.UserGroupMember{UserGroupID: "group-1", UserID: "user-1", Role: "owner", CreatedAt: now})
	db.DB.Create(&domain.UserGroupMember{UserGroupID: "group-1", UserID: "user-2", Role: "member", CreatedAt: now})

	router := gin.New()
	router.GET("/api/v1/user-groups", GetUserGroups)
	router.GET("/api/v1/user-groups/:id", GetUserGroup)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-groups", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on list, got %d", rec.Code)
	}

	var listPayload struct {
		Groups []struct {
			ID          string `json:"id"`
			MemberCount int    `json:"member_count"`
		} `json:"groups"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode groups list: %v", err)
	}
	if len(listPayload.Groups) != 1 || listPayload.Groups[0].MemberCount != 2 {
		t.Fatalf("unexpected groups list payload: %#v", listPayload.Groups)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/user-groups/group-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on detail, got %d", rec.Code)
	}

	var detailPayload struct {
		Group struct {
			ID      string `json:"id"`
			Members []struct {
				Role string `json:"role"`
				User struct {
					ID string `json:"id"`
				} `json:"user"`
			} `json:"members"`
		} `json:"group"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &detailPayload); err != nil {
		t.Fatalf("failed to decode group detail payload: %v", err)
	}
	if detailPayload.Group.ID != "group-1" || len(detailPayload.Group.Members) != 2 {
		t.Fatalf("unexpected group detail payload: %#v", detailPayload.Group)
	}
}

func TestUserGroupCrudLifecycle(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com", Status: "online"})
	db.DB.Create(&domain.User{ID: "user-2", OrganizationID: "org-1", Name: "Jane Smith", Email: "jane@example.com", Status: "away"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})
	db.DB.Create(&domain.UserGroup{ID: "group-1", WorkspaceID: "ws-1", Name: "Design Leads", Handle: "design-leads", Description: "Cross-functional design leadership", CreatedBy: "user-1", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.UserGroupMember{UserGroupID: "group-1", UserID: "user-1", Role: "owner", CreatedAt: now})

	router := gin.New()
	router.POST("/api/v1/user-groups", CreateUserGroup)
	router.PATCH("/api/v1/user-groups/:id", UpdateUserGroup)
	router.DELETE("/api/v1/user-groups/:id", DeleteUserGroup)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-groups", bytes.NewBufferString(`{"workspace_id":"ws-1","name":"Ops Guild","handle":"ops-guild","description":"Operations working group","member_ids":["user-1","user-2"]}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d body=%s", rec.Code, rec.Body.String())
	}

	var createPayload struct {
		Group struct {
			ID      string `json:"id"`
			Members []any  `json:"members"`
		} `json:"group"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("failed to decode create payload: %v", err)
	}
	if createPayload.Group.ID == "" || len(createPayload.Group.Members) != 2 {
		t.Fatalf("unexpected create payload: %#v", createPayload.Group)
	}
	assertPrefixedUUID(t, createPayload.Group.ID, "group")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/user-groups/group-1", bytes.NewBufferString(`{"name":"Design Leadership","description":"Updated desc","member_ids":["user-1","user-2"]}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d body=%s", rec.Code, rec.Body.String())
	}

	var updated domain.UserGroup
	if err := db.DB.First(&updated, "id = ?", "group-1").Error; err != nil {
		t.Fatalf("failed to reload group: %v", err)
	}
	if updated.Name != "Design Leadership" || updated.Description != "Updated desc" {
		t.Fatalf("expected updated group fields, got %#v", updated)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/user-groups/group-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %d body=%s", rec.Code, rec.Body.String())
	}

	var count int64
	db.DB.Model(&domain.UserGroup{}).Where("id = ?", "group-1").Count(&count)
	if count != 0 {
		t.Fatalf("expected group deletion, got count=%d", count)
	}
}

func TestUserGroupMembersAndMentionsEndpoints(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com", Status: "online"})
	db.DB.Create(&domain.User{ID: "user-2", OrganizationID: "org-1", Name: "Jane Smith", Email: "jane@example.com", Status: "online"})
	db.DB.Create(&domain.User{ID: "user-3", OrganizationID: "org-1", Name: "John Doe", Email: "john@example.com", Status: "away"})
	db.DB.Create(&domain.UserGroup{ID: "group-1", WorkspaceID: "ws-1", Name: "Design Guild", Handle: "design-guild", Description: "Guild", CreatedBy: "user-1", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.UserGroupMember{UserGroupID: "group-1", UserID: "user-1", Role: "owner", CreatedAt: now})
	db.DB.Create(&domain.UserGroupMember{UserGroupID: "group-1", UserID: "user-2", Role: "member", CreatedAt: now})

	router := gin.New()
	router.GET("/api/v1/user-groups/:id/members", GetUserGroupMembers)
	router.POST("/api/v1/user-groups/:id/members", AddUserGroupMember)
	router.DELETE("/api/v1/user-groups/:id/members/:userId", RemoveUserGroupMember)
	router.GET("/api/v1/user-groups/mentions", SearchUserGroupMentions)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-groups/group-1/members", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on members list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Members []struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"members"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode user group members: %v", err)
	}
	if len(listPayload.Members) != 2 {
		t.Fatalf("expected 2 group members, got %d", len(listPayload.Members))
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/user-groups/group-1/members", bytes.NewBufferString(`{"user_id":"user-3","role":"member"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on group member add, got %d body=%s", rec.Code, rec.Body.String())
	}

	var memberCount int64
	db.DB.Model(&domain.UserGroupMember{}).Where("user_group_id = ?", "group-1").Count(&memberCount)
	if memberCount != 3 {
		t.Fatalf("expected 3 group members after add, got %d", memberCount)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/user-groups/mentions?q=design", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on mentions lookup, got %d body=%s", rec.Code, rec.Body.String())
	}

	var mentionsPayload struct {
		Groups []struct {
			ID          string `json:"id"`
			Handle      string `json:"handle"`
			MemberCount int    `json:"member_count"`
		} `json:"groups"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &mentionsPayload); err != nil {
		t.Fatalf("failed to decode mentions payload: %v", err)
	}
	if len(mentionsPayload.Groups) != 1 || mentionsPayload.Groups[0].MemberCount != 3 {
		t.Fatalf("unexpected mentions payload: %#v", mentionsPayload.Groups)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/user-groups/group-1/members/user-2", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on group member delete, got %d body=%s", rec.Code, rec.Body.String())
	}

	db.DB.Model(&domain.UserGroupMember{}).Where("user_group_id = ?", "group-1").Count(&memberCount)
	if memberCount != 2 {
		t.Fatalf("expected 2 group members after delete, got %d", memberCount)
	}
}

func TestGetWorkflowsAndTools(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.WorkflowDefinition{ID: "wf-1", Name: "Incident Review", Category: "operations", Description: "Run post-incident review", Trigger: "manual", IsActive: true, CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.ToolDefinition{ID: "tool-1", Name: "Knowledge Search", Key: "knowledge-search", Category: "search", Description: "Search workspace knowledge", Icon: "search", IsEnabled: true, CreatedAt: now, UpdatedAt: now})

	router := gin.New()
	router.GET("/api/v1/workflows", GetWorkflows)
	router.GET("/api/v1/tools", GetTools)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workflows", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on workflows, got %d", rec.Code)
	}

	var workflowsPayload struct {
		Workflows []domain.WorkflowDefinition `json:"workflows"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &workflowsPayload); err != nil {
		t.Fatalf("failed to decode workflows: %v", err)
	}
	if len(workflowsPayload.Workflows) != 1 || workflowsPayload.Workflows[0].ID != "wf-1" {
		t.Fatalf("unexpected workflows payload: %#v", workflowsPayload.Workflows)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tools", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on tools, got %d", rec.Code)
	}

	var toolsPayload struct {
		Tools []domain.ToolDefinition `json:"tools"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &toolsPayload); err != nil {
		t.Fatalf("failed to decode tools: %v", err)
	}
	if len(toolsPayload.Tools) != 1 || toolsPayload.Tools[0].ID != "tool-1" {
		t.Fatalf("unexpected tools payload: %#v", toolsPayload.Tools)
	}
}

func TestWorkflowRunsCanBeCreatedAndListed(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com", Status: "online"})
	db.DB.Create(&domain.WorkflowDefinition{ID: "wf-1", Name: "Incident Review", Category: "operations", Description: "Run post-incident review", Trigger: "manual", IsActive: true, CreatedAt: now, UpdatedAt: now})

	router := gin.New()
	router.POST("/api/v1/workflows/:id/runs", CreateWorkflowRun)
	router.GET("/api/v1/workflows/runs", GetWorkflowRuns)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/wf-1/runs", bytes.NewBufferString(`{"status":"running","summary":"Kicked off review","input":{"channel_id":"ch-incident"}}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on workflow run create, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/workflows/runs?workflow_id=wf-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on workflow runs list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Runs []struct {
			ID       string `json:"id"`
			Status   string `json:"status"`
			Workflow struct {
				ID string `json:"id"`
			} `json:"workflow"`
			StartedBy struct {
				ID string `json:"id"`
			} `json:"started_by"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode workflow runs payload: %v", err)
	}
	if len(payload.Runs) != 1 || payload.Runs[0].Workflow.ID != "wf-1" || payload.Runs[0].StartedBy.ID != "user-1" {
		t.Fatalf("unexpected workflow runs payload: %#v", payload.Runs)
	}
	assertPrefixedUUID(t, payload.Runs[0].ID, "run")
}

func TestWorkflowRunCreateBroadcastsRealtimeEvent(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com", Status: "online"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})
	db.DB.Create(&domain.WorkflowDefinition{ID: "wf-1", Name: "Incident Review", Category: "operations", Description: "Run post-incident review", Trigger: "manual", IsActive: true, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()})

	hub := realtime.NewHub()
	go hub.Run()
	RealtimeHub = hub
	client := realtime.NewTestClient(4)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	router := gin.New()
	router.POST("/api/v1/workflows/:id/runs", CreateWorkflowRun)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/wf-1/runs", bytes.NewBufferString(`{"status":"running","summary":"Triggered from UI"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on workflow run create, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "workflow.run.updated")
}

func TestWorkflowRunDetailCancelAndRetry(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com", Status: "online"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})
	db.DB.Create(&domain.WorkflowDefinition{ID: "wf-1", Name: "Incident Review", Category: "operations", Description: "Run post-incident review", Trigger: "manual", IsActive: true, CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.WorkflowRun{
		ID:         "run-1",
		WorkflowID: "wf-1",
		StartedBy:  "user-1",
		Status:     "running",
		Input:      `{"channel_id":"ch-incident"}`,
		Summary:    "Initial run",
		Error:      "",
		StartedAt:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	db.DB.Create(&domain.WorkflowRunStep{WorkflowRunID: "run-1", Name: "Collect input", Status: "completed", DurationMS: 120, Detail: "Loaded trigger context", CreatedAt: now})
	db.DB.Create(&domain.WorkflowRunStep{WorkflowRunID: "run-1", Name: "Call agent", Status: "running", DurationMS: 0, Detail: "Executing handoff", CreatedAt: now.Add(time.Second)})

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	router := gin.New()
	router.GET("/api/v1/workflows/runs/:id", GetWorkflowRun)
	router.POST("/api/v1/workflows/runs/:id/cancel", CancelWorkflowRun)
	router.POST("/api/v1/workflows/runs/:id/retry", RetryWorkflowRun)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workflows/runs/run-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on workflow run detail, got %d body=%s", rec.Code, rec.Body.String())
	}

	var detailPayload struct {
		Run struct {
			WorkflowName string `json:"workflow_name"`
			TriggeredBy  string `json:"triggered_by"`
			DurationMS   int    `json:"duration_ms"`
			Steps        []struct {
				Name string `json:"name"`
			} `json:"steps"`
		} `json:"run"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &detailPayload); err != nil {
		t.Fatalf("failed to decode workflow detail payload: %v", err)
	}
	if detailPayload.Run.WorkflowName != "Incident Review" || detailPayload.Run.TriggeredBy != "Nikko Fu" || len(detailPayload.Run.Steps) != 2 {
		t.Fatalf("unexpected workflow detail payload: %#v", detailPayload.Run)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/workflows/runs/run-1/cancel", bytes.NewBufferString(`{"summary":"Stopped by operator"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on workflow cancel, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "workflow.run.updated")

	var cancelled domain.WorkflowRun
	if err := db.DB.First(&cancelled, "id = ?", "run-1").Error; err != nil {
		t.Fatalf("failed to reload cancelled run: %v", err)
	}
	if cancelled.Status != "cancelled" || cancelled.CompletedAt == nil {
		t.Fatalf("expected cancelled workflow run, got %#v", cancelled)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/workflows/runs/run-1/retry", bytes.NewBufferString(`{"summary":"Retrying after fix"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on workflow retry, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "workflow.run.updated")

	var count int64
	db.DB.Model(&domain.WorkflowRun{}).Where("workflow_id = ?", "wf-1").Count(&count)
	if count != 2 {
		t.Fatalf("expected retry to create second workflow run, got %d", count)
	}
}

func TestWorkflowRunLogsAndDeleteEndpoints(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko", Email: "nikko@example.com"})
	db.DB.Create(&domain.WorkflowDefinition{ID: "wf-1", Name: "Daily Brief", Category: "automation", Trigger: "manual", IsActive: true})
	db.DB.Create(&domain.WorkflowRun{
		ID:         "run-1",
		WorkflowID: "wf-1",
		StartedBy:  "user-1",
		Status:     "running",
		Input:      `{"channel_id":"ch-1"}`,
		Summary:    "Building a daily brief",
		StartedAt:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	db.DB.Create(&domain.WorkflowRunStep{WorkflowRunID: "run-1", Name: "Collect input", Status: "completed", Detail: "Loaded trigger context", CreatedAt: now})
	db.DB.Create(&domain.WorkflowRunLog{WorkflowRunID: "run-1", Level: "info", Message: "Run accepted", Metadata: `{"source":"test"}`, CreatedAt: now})
	db.DB.Create(&domain.WorkflowRunLog{WorkflowRunID: "run-1", Level: "warning", Message: "Waiting for agent", CreatedAt: now.Add(time.Second)})

	router := gin.New()
	router.GET("/api/v1/workflows/runs/:id/logs", GetWorkflowRunLogs)
	router.DELETE("/api/v1/workflows/runs/:id", DeleteWorkflowRun)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workflows/runs/run-1/logs", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var logsPayload struct {
		Logs []struct {
			Level    string         `json:"level"`
			Message  string         `json:"message"`
			Metadata map[string]any `json:"metadata"`
		} `json:"logs"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &logsPayload); err != nil {
		t.Fatalf("failed to decode logs response: %v", err)
	}
	if len(logsPayload.Logs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(logsPayload.Logs))
	}
	if logsPayload.Logs[0].Level != "info" || logsPayload.Logs[0].Message != "Run accepted" || logsPayload.Logs[0].Metadata["source"] != "test" {
		t.Fatalf("unexpected first log payload: %+v", logsPayload.Logs[0])
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/workflows/runs/run-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %d body=%s", rec.Code, rec.Body.String())
	}

	var runCount int64
	db.DB.Model(&domain.WorkflowRun{}).Where("id = ?", "run-1").Count(&runCount)
	if runCount != 0 {
		t.Fatalf("expected workflow run to be deleted, got %d", runCount)
	}
	var stepCount int64
	db.DB.Model(&domain.WorkflowRunStep{}).Where("workflow_run_id = ?", "run-1").Count(&stepCount)
	if stepCount != 0 {
		t.Fatalf("expected workflow run steps to be deleted, got %d", stepCount)
	}
	var logCount int64
	db.DB.Model(&domain.WorkflowRunLog{}).Where("workflow_run_id = ?", "run-1").Count(&logCount)
	if logCount != 0 {
		t.Fatalf("expected workflow run logs to be deleted, got %d", logCount)
	}
}

func TestNotificationPreferencesCanBeUpdated(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com", Status: "online"})

	router := gin.New()
	router.GET("/api/v1/notifications/preferences", GetNotificationPreferences)
	router.PATCH("/api/v1/notifications/preferences", PatchNotificationPreferences)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/preferences", bytes.NewBufferString(`{"inbox_enabled":true,"mentions_enabled":true,"dm_enabled":false,"mute_all":false,"mute_rules":[{"scope_type":"channel","scope_id":"ch-1","is_muted":true}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on notification preferences patch, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/notifications/preferences", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on notification preferences get, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Preferences struct {
			DMEnabled bool `json:"dm_enabled"`
			MuteRules []struct {
				ScopeType string `json:"scope_type"`
				ScopeID   string `json:"scope_id"`
				IsMuted   bool   `json:"is_muted"`
			} `json:"mute_rules"`
		} `json:"preferences"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode notification preferences payload: %v", err)
	}
	if payload.Preferences.DMEnabled {
		t.Fatalf("expected dm_enabled=false, got %#v", payload.Preferences)
	}
	if len(payload.Preferences.MuteRules) != 1 || payload.Preferences.MuteRules[0].ScopeID != "ch-1" || !payload.Preferences.MuteRules[0].IsMuted {
		t.Fatalf("unexpected mute rules payload: %#v", payload.Preferences.MuteRules)
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

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/settings", bytes.NewBufferString(`{"provider":"gemini","model":"gemini-3-flash-preview","mode":"planning","theme":"system","message_density":"compact","locale":"zh-CN","timezone":"Asia/Shanghai"}`))
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
	if refreshed.AIProvider != "gemini" || refreshed.AIModel != "gemini-3-flash-preview" || refreshed.AIMode != "planning" || refreshed.ThemePreference != "system" || refreshed.MessageDensity != "compact" || refreshed.Locale != "zh-CN" || refreshed.Timezone != "Asia/Shanghai" {
		t.Fatalf("settings were not persisted: %#v", refreshed)
	}
}

func TestGetMeSettingsReturnsHydratedPreferences(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{
		ID:              "user-1",
		OrganizationID:  "org-1",
		Name:            "Nikko Fu",
		Email:           "nikko@example.com",
		AIProvider:      "openrouter",
		AIModel:         "nvidia/nemotron-3-super-120b-a12b:free",
		AIMode:          "focus",
		ThemePreference: "dark",
		MessageDensity:  "compact",
		Locale:          "zh-CN",
		Timezone:        "Asia/Shanghai",
	})

	router := gin.New()
	router.GET("/api/v1/me/settings", GetMeSettings)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/settings", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Settings struct {
			Provider       string `json:"provider"`
			Model          string `json:"model"`
			Mode           string `json:"mode"`
			Theme          string `json:"theme"`
			MessageDensity string `json:"message_density"`
			Locale         string `json:"locale"`
			Timezone       string `json:"timezone"`
		} `json:"settings"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Settings.Provider != "openrouter" || payload.Settings.Model != "nvidia/nemotron-3-super-120b-a12b:free" || payload.Settings.Mode != "focus" || payload.Settings.Theme != "dark" || payload.Settings.MessageDensity != "compact" || payload.Settings.Locale != "zh-CN" || payload.Settings.Timezone != "Asia/Shanghai" {
		t.Fatalf("unexpected settings payload: %#v", payload.Settings)
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

func TestCreateMessageHydratesArtifactAndFileAttachments(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.Artifact{
		ID:        "artifact-1",
		ChannelID: "ch-1",
		Title:     "Launch Plan",
		Version:   2,
		Type:      "document",
		Status:    "live",
		Content:   "Launch plan content",
		Source:    "manual",
		CreatedBy: "user-1",
		UpdatedBy: "user-1",
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now,
	})
	db.DB.Create(&domain.FileAsset{
		ID:             "file-1",
		ChannelID:      "ch-1",
		UploaderID:     "user-1",
		Name:           "brief.pdf",
		StoragePath:    "brief.pdf",
		ContentType:    "application/pdf",
		SizeBytes:      2048,
		KnowledgeState: "ready",
		SourceKind:     "wiki",
		Summary:        "Launch brief with milestone details",
		Tags:           `["launch","wiki"]`,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	db.DB.Create(&domain.FileComment{ID: "fcomment-1", FileID: "file-1", UserID: "user-1", Content: "Looks good", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.FileShare{ID: "fshare-1", FileID: "file-1", ChannelID: "ch-1", SharedBy: "user-1", Comment: "share", CreatedAt: now})
	db.DB.Create(&domain.StarredFile{FileID: "file-1", UserID: "user-1", CreatedAt: now})

	router := gin.New()
	router.POST("/api/v1/messages", CreateMessage)
	router.GET("/api/v1/messages", GetMessages)
	router.GET("/api/v1/messages/:id/files", GetMessageFiles)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBufferString(`{"channel_id":"ch-1","content":"See linked assets","user_id":"user-1","artifact_ids":["artifact-1"],"file_ids":["file-1"]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var createPayload struct {
		Message struct {
			ID       string `json:"id"`
			Metadata string `json:"metadata"`
		} `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("failed to decode create message payload: %v", err)
	}

	var createMeta struct {
		Attachments []struct {
			Kind           string         `json:"kind"`
			FileID         string         `json:"file_id"`
			PreviewKind    string         `json:"preview_kind"`
			CommentCount   int64          `json:"comment_count"`
			ShareCount     int64          `json:"share_count"`
			Starred        bool           `json:"starred"`
			KnowledgeState string         `json:"knowledge_state"`
			SourceKind     string         `json:"source_kind"`
			Summary        string         `json:"summary"`
			Tags           []string       `json:"tags"`
			File           map[string]any `json:"file"`
			Preview        map[string]any `json:"preview"`
		} `json:"attachments"`
	}
	if err := json.Unmarshal([]byte(createPayload.Message.Metadata), &createMeta); err != nil {
		t.Fatalf("failed to decode message metadata: %v", err)
	}
	if len(createMeta.Attachments) != 2 {
		t.Fatalf("expected 2 attachments, got %#v", createMeta.Attachments)
	}
	var foundRichFile bool
	for _, attachment := range createMeta.Attachments {
		if attachment.Kind != "file" {
			continue
		}
		foundRichFile = true
		if attachment.FileID != "file-1" || attachment.PreviewKind != "pdf" || attachment.CommentCount != 1 || attachment.ShareCount != 1 {
			t.Fatalf("expected rich file attachment counters and preview, got %#v", attachment)
		}
		if !attachment.Starred || attachment.KnowledgeState != "ready" || attachment.SourceKind != "wiki" || attachment.Summary == "" {
			t.Fatalf("expected rich file attachment knowledge metadata, got %#v", attachment)
		}
		if len(attachment.Tags) != 2 || attachment.File == nil || attachment.Preview == nil {
			t.Fatalf("expected nested file/preview payloads, got %#v", attachment)
		}
	}
	if !foundRichFile {
		t.Fatalf("expected metadata to include hydrated file attachment, got %s", createPayload.Message.Metadata)
	}
	assertPrefixedUUID(t, createPayload.Message.ID, "msg")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/messages?channel_id=ch-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get messages, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Messages []domain.Message `json:"messages"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode messages payload: %v", err)
	}
	if len(listPayload.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(listPayload.Messages))
	}
	if !strings.Contains(listPayload.Messages[0].Metadata, `"artifact_id":"artifact-1"`) || !strings.Contains(listPayload.Messages[0].Metadata, `"file_id":"file-1"`) {
		t.Fatalf("expected refreshed message metadata to include reference ids, got %s", listPayload.Messages[0].Metadata)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/messages/"+createPayload.Message.ID+"/files", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get message files, got %d body=%s", rec.Code, rec.Body.String())
	}

	var filesPayload struct {
		MessageID string `json:"message_id"`
		Files     []struct {
			FileID         string `json:"file_id"`
			Name           string `json:"name"`
			PreviewKind    string `json:"preview_kind"`
			DownloadURL    string `json:"download_url"`
			CommentCount   int64  `json:"comment_count"`
			ShareCount     int64  `json:"share_count"`
			Starred        bool   `json:"starred"`
			KnowledgeState string `json:"knowledge_state"`
			SourceKind     string `json:"source_kind"`
		} `json:"files"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filesPayload); err != nil {
		t.Fatalf("failed to decode message files payload: %v", err)
	}
	if filesPayload.MessageID != createPayload.Message.ID || len(filesPayload.Files) != 1 {
		t.Fatalf("expected one message file attachment, got %#v", filesPayload)
	}
	if filesPayload.Files[0].FileID != "file-1" || filesPayload.Files[0].PreviewKind != "pdf" || filesPayload.Files[0].DownloadURL == "" {
		t.Fatalf("expected hydrated file card payload, got %#v", filesPayload.Files[0])
	}
	if filesPayload.Files[0].CommentCount != 1 || filesPayload.Files[0].ShareCount != 1 || !filesPayload.Files[0].Starred {
		t.Fatalf("expected rich file counters in endpoint payload, got %#v", filesPayload.Files[0])
	}
	if filesPayload.Files[0].KnowledgeState != "ready" || filesPayload.Files[0].SourceKind != "wiki" {
		t.Fatalf("expected knowledge metadata in endpoint payload, got %#v", filesPayload.Files[0])
	}
}

func TestCreateMessageHydratesExplicitKnowledgeEntityMentions(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})

	router := gin.New()
	router.POST("/api/v1/messages", CreateMessage)
	router.GET("/api/v1/messages", GetMessages)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBufferString(`{"channel_id":"ch-1","content":"<p>@Launch Program is ready for review</p>","user_id":"user-1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var createPayload struct {
		Message struct {
			ID       string `json:"id"`
			Metadata string `json:"metadata"`
		} `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}

	var meta struct {
		EntityMentions []struct {
			EntityID    string `json:"entity_id"`
			EntityTitle string `json:"entity_title"`
			EntityKind  string `json:"entity_kind"`
			MentionText string `json:"mention_text"`
		} `json:"entity_mentions"`
	}
	if err := json.Unmarshal([]byte(createPayload.Message.Metadata), &meta); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if len(meta.EntityMentions) != 1 {
		t.Fatalf("expected one entity mention, got %#v", meta.EntityMentions)
	}
	if meta.EntityMentions[0].EntityID != "entity-1" || meta.EntityMentions[0].EntityTitle != "Launch Program" || meta.EntityMentions[0].MentionText != "@Launch Program" {
		t.Fatalf("expected hydrated entity mention payload, got %#v", meta.EntityMentions[0])
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/messages?channel_id=ch-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get messages, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Messages []domain.Message `json:"messages"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("decode list payload: %v", err)
	}
	if len(listPayload.Messages) != 1 {
		t.Fatalf("expected one message in list, got %#v", listPayload.Messages)
	}
	if err := json.Unmarshal([]byte(listPayload.Messages[0].Metadata), &meta); err != nil {
		t.Fatalf("decode list metadata: %v", err)
	}
	if len(meta.EntityMentions) != 1 || meta.EntityMentions[0].MentionText != "@Launch Program" {
		t.Fatalf("expected message list to include hydrated entity mentions, got %#v", meta.EntityMentions)
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

func TestGetAgentCollabSnapshotReturnsParsedMarkdown(t *testing.T) {
	setupTestDB(t)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "AGENT-COLLAB.md")
	content := `## 📋 Task Board

| Status | Task | Assigned To | Deadline | Description |
| :--- | :--- | :--- | :--- | :--- |
| 🟢 Done | Snapshot API | Codex | 2026-04-18 | Confirm agent-collab snapshot loads. |

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Codex** | verification | Snapshot sync | 100% |
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp collab doc: %v", err)
	}

	prevPath := CollabSnapshotPath
	CollabSnapshotPath = path
	defer func() { CollabSnapshotPath = prevPath }()

	router := gin.New()
	router.GET("/api/v1/agent-collab/snapshot", GetAgentCollabSnapshot)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agent-collab/snapshot", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		ActiveSuperpowers []map[string]any `json:"active_superpowers"`
		TaskBoard         []map[string]any `json:"task_board"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode snapshot response: %v", err)
	}
	if len(payload.ActiveSuperpowers) != 1 || len(payload.TaskBoard) != 1 {
		t.Fatalf("unexpected snapshot payload: %#v", payload)
	}
}

func TestAgentCollabMembersAndCommLogEndpoints(t *testing.T) {
	setupTestDB(t)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "AGENT-COLLAB.md")
	content := `# Relay Agent Workspace: Team Collaboration Hub

## 👥 Member Profiles

| Name | Role | Specialty | Primary Tools |
| :--- | :--- | :--- | :--- |
| **Nikko Fu** | Human Owner | Product Strategy, Design, Final Review | Brainstorming, PR Review |
| **Codex** | API/Backend Agent | Go, Gin, GORM, SQLite, WebSockets, AI Orchestration | apps/api, internal/ |
| **Windsurf** | Web/UI Agent | Component Architecture, TypeScript, UX Flows, Agent Collaboration UI | apps/web, write_file, multi_edit |

## 📋 Task Board

| Status | Task | Assigned To | Deadline | Description |
| :--- | :--- | :--- | :--- | :--- |
| 🟢 Done | Snapshot API | Codex | 2026-04-18 | Confirm agent-collab snapshot loads. |

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Codex** | verification | Snapshot sync | 100% |

## 💬 Communication Log

### 2026-04-21 - Existing Note
- **Windsurf → Codex**: "Please make the hub dynamic."
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp collab doc: %v", err)
	}

	prevPath := CollabSnapshotPath
	CollabSnapshotPath = path
	defer func() { CollabSnapshotPath = prevPath }()

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)
	client := realtime.NewTestClient(4)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	router := gin.New()
	router.GET("/api/v1/agent-collab/members", GetAgentCollabMembers)
	router.POST("/api/v1/agent-collab/comm-log", CreateAgentCollabCommLog)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/agent-collab/members", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on members, got %d body=%s", rec.Code, rec.Body.String())
	}

	var membersPayload struct {
		Members []map[string]any `json:"members"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &membersPayload); err != nil {
		t.Fatalf("failed to decode members payload: %v", err)
	}
	if len(membersPayload.Members) != 3 {
		t.Fatalf("expected 3 members, got %#v", membersPayload.Members)
	}
	if membersPayload.Members[2]["name"] != "Windsurf" || membersPayload.Members[2]["role"] != "Web/UI Agent" {
		t.Fatalf("expected Windsurf member row, got %#v", membersPayload.Members[2])
	}
	if tools, ok := membersPayload.Members[2]["primary_tools_array"].([]any); !ok || len(tools) != 3 || tools[0] != "apps/web" {
		t.Fatalf("expected normalized primary tools array, got %#v", membersPayload.Members[2])
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/agent-collab/comm-log", bytes.NewBufferString(`{"from":"Codex","to":"Windsurf","title":"Phase 40 Dynamic Agent-Collab API","content":"GET members and POST comm-log are ready for frontend integration."}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on comm log create, got %d body=%s", rec.Code, rec.Body.String())
	}

	var logPayload struct {
		Entry map[string]any `json:"entry"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &logPayload); err != nil {
		t.Fatalf("failed to decode comm log payload: %v", err)
	}
	if logPayload.Entry["from"] != "Codex" || logPayload.Entry["to"] != "Windsurf" {
		t.Fatalf("unexpected comm log entry: %#v", logPayload.Entry)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read updated collab doc: %v", err)
	}
	updatedContent := string(updated)
	if !strings.Contains(updatedContent, "Phase 40 Dynamic Agent-Collab API") || !strings.Contains(updatedContent, "**Codex → Windsurf**") {
		t.Fatalf("expected comm log to be appended, got:\n%s", updatedContent)
	}

	assertRealtimeEventType(t, client, "agent_collab.sync")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/agent-collab/snapshot", nil)
	router.GET("/api/v1/agent-collab/snapshot", GetAgentCollabSnapshot)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on snapshot, got %d body=%s", rec.Code, rec.Body.String())
	}
	var snapshotPayload struct {
		CommLog []map[string]any `json:"comm_log"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &snapshotPayload); err != nil {
		t.Fatalf("failed to decode snapshot payload: %v", err)
	}
	if len(snapshotPayload.CommLog) == 0 || snapshotPayload.CommLog[0]["to"] != "Windsurf" {
		t.Fatalf("expected snapshot comm_log to include to field, got %#v", snapshotPayload.CommLog)
	}
}

func TestDMEndpointsListCreateAndSendMessages(t *testing.T) {
	setupTestDB(t)

	users := []domain.User{
		{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"},
		{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"},
		{ID: "user-3", Name: "Jane Smith", Email: "jane@example.com"},
	}
	for _, user := range users {
		db.DB.Create(&user)
	}

	dm := domain.DMConversation{ID: "dm-1", CreatedAt: time.Now().Add(-time.Hour)}
	db.DB.Create(&dm)
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-1"})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-2"})
	db.DB.Create(&domain.DMMessage{
		ID:               "dm-msg-1",
		DMConversationID: "dm-1",
		UserID:           "user-2",
		Content:          "Hello from DM",
		CreatedAt:        time.Now().Add(-30 * time.Minute),
	})

	router := gin.New()
	router.GET("/api/v1/dms", GetDMConversations)
	router.POST("/api/v1/dms", CreateOrOpenDMConversation)
	router.GET("/api/v1/dms/:id/messages", GetDMMessages)
	router.POST("/api/v1/dms/:id/messages", CreateDMMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dms", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on dm list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Conversations []map[string]any `json:"conversations"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode dm list: %v", err)
	}
	if len(listPayload.Conversations) != 1 {
		t.Fatalf("expected 1 dm conversation, got %d", len(listPayload.Conversations))
	}
	userIDs, ok := listPayload.Conversations[0]["user_ids"].([]any)
	if !ok || len(userIDs) != 2 {
		t.Fatalf("expected user_ids in dm list response, got %#v", listPayload.Conversations[0]["user_ids"])
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/dms", bytes.NewBufferString(`{"user_id":"user-3"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on dm create, got %d body=%s", rec.Code, rec.Body.String())
	}

	var createPayload struct {
		Conversation struct {
			ID      string   `json:"id"`
			UserIDs []string `json:"user_ids"`
		} `json:"conversation"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("failed to decode dm create: %v", err)
	}
	if createPayload.Conversation.ID == "" {
		t.Fatal("expected created dm conversation id")
	}
	assertPrefixedUUID(t, createPayload.Conversation.ID, "dm")
	if len(createPayload.Conversation.UserIDs) != 2 {
		t.Fatalf("expected created conversation user_ids, got %#v", createPayload.Conversation.UserIDs)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/dms/dm-1/messages", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on dm messages, got %d body=%s", rec.Code, rec.Body.String())
	}

	var messagesPayload struct {
		Messages []domain.DMMessage `json:"messages"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &messagesPayload); err != nil {
		t.Fatalf("failed to decode dm messages: %v", err)
	}
	if len(messagesPayload.Messages) != 1 {
		t.Fatalf("expected 1 dm message, got %d", len(messagesPayload.Messages))
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/dms/dm-1/messages", bytes.NewBufferString(`{"content":"Reply in DM","user_id":"user-1"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on dm send, got %d body=%s", rec.Code, rec.Body.String())
	}

	var sendPayload struct {
		Message domain.DMMessage `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &sendPayload); err != nil {
		t.Fatalf("failed to decode dm send payload: %v", err)
	}
	assertPrefixedUUID(t, sendPayload.Message.ID, "dm-msg")

	var count int64
	db.DB.Model(&domain.DMMessage{}).Where("dm_conversation_id = ?", "dm-1").Count(&count)
	if count != 2 {
		t.Fatalf("expected 2 dm messages after send, got %d", count)
	}
}

func TestCreateDMMessageBroadcastsRealtimeEvent(t *testing.T) {
	setupTestDB(t)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(4)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"})
	db.DB.Create(&domain.DMConversation{ID: "dm-1", CreatedAt: time.Now().UTC()})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-1"})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-2"})

	router := gin.New()
	router.POST("/api/v1/dms/:id/messages", CreateDMMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dms/dm-1/messages", bytes.NewBufferString(`{"content":"Realtime DM","user_id":"user-1"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on dm send, got %d body=%s", rec.Code, rec.Body.String())
	}

	raw, err := client.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("failed to receive realtime event: %v", err)
	}

	var event realtime.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("failed to decode realtime event: %v", err)
	}
	if event.Type != "message.created" {
		t.Fatalf("expected message.created event, got %s", event.Type)
	}

	payload, ok := event.Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %#v", event.Payload)
	}
	if payload["dm_id"] != "dm-1" {
		t.Fatalf("expected dm_id in payload, got %#v", payload)
	}
}

func TestGetActivityReturnsRecentWorkspaceSignals(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	users := []domain.User{
		{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"},
		{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"},
		{ID: "user-3", Name: "Jane Smith", Email: "jane@example.com"},
	}
	for _, user := range users {
		db.DB.Create(&user)
	}
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1", Role: "member"})
	parent := domain.Message{ID: "msg-parent", ChannelID: "ch-1", UserID: "user-1", Content: "Root update", CreatedAt: now.Add(-2 * time.Hour), Metadata: "{}"}
	reply := domain.Message{ID: "msg-reply", ChannelID: "ch-1", UserID: "user-3", Content: "Replying in thread", ThreadID: "msg-parent", CreatedAt: now.Add(-time.Hour), Metadata: "{}"}
	mention := domain.Message{ID: "msg-mention", ChannelID: "ch-1", UserID: "user-2", Content: "@Nikko Fu for review", CreatedAt: now.Add(-30 * time.Minute), Metadata: `{"user_mentions":[{"user_id":"user-1","name":"Nikko Fu","mention_text":"@Nikko Fu"}]}`}
	db.DB.Create(&parent)
	db.DB.Create(&reply)
	db.DB.Create(&mention)
	db.DB.Create(&domain.MessageMention{ID: "mm-activity-1", MessageID: "msg-mention", WorkspaceID: "ws-1", ChannelID: "ch-1", MentionedUserID: "user-1", MentionedByUserID: "user-2", MentionText: "@Nikko Fu", MentionKind: "user", CreatedAt: now.Add(-30 * time.Minute)})
	db.DB.Create(&domain.MessageReaction{MessageID: "msg-parent", UserID: "user-2", Emoji: "🔥", CreatedAt: now.Add(-20 * time.Minute)})
	db.DB.Create(&domain.WorkspaceList{ID: "list-1", WorkspaceID: "ws-1", ChannelID: "ch-1", Title: "Launch Checklist", CreatedBy: "user-2", CreatedAt: now.Add(-18 * time.Minute), UpdatedAt: now.Add(-14 * time.Minute)})
	db.DB.Create(&domain.WorkspaceListItem{ID: 1, ListID: "list-1", Content: "Review launch copy", Position: 1, IsCompleted: true, AssignedTo: "user-1", CompletedAt: ptrTime(now.Add(-14 * time.Minute)), CreatedBy: "user-2", CreatedAt: now.Add(-18 * time.Minute), UpdatedAt: now.Add(-14 * time.Minute)})
	db.DB.Create(&domain.ToolDefinition{ID: "tool-1", Name: "Web Search", Key: "web-search", Category: "research", Description: "Search current information", Icon: "search", IsEnabled: true, CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Hour)})
	db.DB.Create(&domain.ToolRun{ID: "toolrun-1", ToolID: "tool-1", TriggeredBy: "user-1", Status: "success", Input: `{"channel_id":"ch-1"}`, Summary: "Executed Web Search", StartedAt: now.Add(-12 * time.Minute), CompletedAt: ptrTime(now.Add(-11 * time.Minute)), CreatedAt: now.Add(-12 * time.Minute), UpdatedAt: now.Add(-11 * time.Minute)})
	db.DB.Create(&domain.FileAsset{ID: "file-1", ChannelID: "ch-1", UploaderID: "user-2", Name: "brief.md", StoragePath: "brief.md", ContentType: "text/markdown", SizeBytes: 256, CreatedAt: now.Add(-10 * time.Minute), UpdatedAt: now.Add(-10 * time.Minute)})

	router := gin.New()
	router.GET("/api/v1/activity", GetActivity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on activity, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Activities []map[string]any `json:"activities"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode activity payload: %v", err)
	}
	if len(payload.Activities) < 6 {
		t.Fatalf("expected at least 6 activity items, got %d", len(payload.Activities))
	}
	requiredTypes := map[string]bool{
		"mention":        false,
		"thread_reply":   false,
		"reaction":       false,
		"list_completed": false,
		"tool_run":       false,
		"file_uploaded":  false,
	}
	for _, item := range payload.Activities {
		if itemType, ok := item["type"].(string); ok {
			if _, exists := requiredTypes[itemType]; exists {
				requiredTypes[itemType] = true
			}
		}
	}
	for itemType, found := range requiredTypes {
		if !found {
			t.Fatalf("expected activity type %s in payload, got %#v", itemType, payload.Activities)
		}
	}
}

func TestGetInboxReturnsAggregatedSignals(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	users := []domain.User{
		{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"},
		{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"},
		{ID: "user-3", Name: "Jane Smith", Email: "jane@example.com"},
	}
	for _, user := range users {
		db.DB.Create(&user)
	}
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1", Role: "member"})
	parent := domain.Message{ID: "msg-parent", ChannelID: "ch-1", UserID: "user-1", Content: "Root update", CreatedAt: now.Add(-2 * time.Hour), Metadata: "{}"}
	reply := domain.Message{ID: "msg-reply", ChannelID: "ch-1", UserID: "user-3", Content: "Replying in thread", ThreadID: "msg-parent", CreatedAt: now.Add(-time.Hour), Metadata: "{}"}
	mention := domain.Message{ID: "msg-mention", ChannelID: "ch-1", UserID: "user-2", Content: "@Nikko Fu for review", CreatedAt: now.Add(-30 * time.Minute), Metadata: `{"user_mentions":[{"user_id":"user-1","name":"Nikko Fu","mention_text":"@Nikko Fu"}]}`}
	db.DB.Create(&parent)
	db.DB.Create(&reply)
	db.DB.Create(&mention)
	db.DB.Create(&domain.MessageMention{ID: "mm-inbox-1", MessageID: "msg-mention", WorkspaceID: "ws-1", ChannelID: "ch-1", MentionedUserID: "user-1", MentionedByUserID: "user-2", MentionText: "@Nikko Fu", MentionKind: "user", CreatedAt: now.Add(-30 * time.Minute)})
	db.DB.Create(&domain.MessageReaction{MessageID: "msg-parent", UserID: "user-2", Emoji: "🔥", CreatedAt: now.Add(-20 * time.Minute)})
	db.DB.Create(&domain.WorkspaceList{ID: "list-1", WorkspaceID: "ws-1", ChannelID: "ch-1", Title: "Launch Checklist", CreatedBy: "user-2", CreatedAt: now.Add(-18 * time.Minute), UpdatedAt: now.Add(-14 * time.Minute)})
	db.DB.Create(&domain.WorkspaceListItem{ID: 1, ListID: "list-1", Content: "Review launch copy", Position: 1, IsCompleted: true, AssignedTo: "user-1", CompletedAt: ptrTime(now.Add(-14 * time.Minute)), CreatedBy: "user-2", CreatedAt: now.Add(-18 * time.Minute), UpdatedAt: now.Add(-14 * time.Minute)})
	db.DB.Create(&domain.ToolDefinition{ID: "tool-1", Name: "Web Search", Key: "web-search", Category: "research", Description: "Search current information", Icon: "search", IsEnabled: true, CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Hour)})
	db.DB.Create(&domain.ToolRun{ID: "toolrun-1", ToolID: "tool-1", TriggeredBy: "user-1", Status: "success", Input: `{"channel_id":"ch-1"}`, Summary: "Executed Web Search", StartedAt: now.Add(-12 * time.Minute), CompletedAt: ptrTime(now.Add(-11 * time.Minute)), CreatedAt: now.Add(-12 * time.Minute), UpdatedAt: now.Add(-11 * time.Minute)})
	db.DB.Create(&domain.FileAsset{ID: "file-1", ChannelID: "ch-1", UploaderID: "user-2", Name: "brief.md", StoragePath: "brief.md", ContentType: "text/markdown", SizeBytes: 256, CreatedAt: now.Add(-10 * time.Minute), UpdatedAt: now.Add(-10 * time.Minute)})

	router := gin.New()
	router.GET("/api/v1/inbox", GetInbox)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/inbox", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on inbox, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode inbox payload: %v", err)
	}
	if len(payload.Items) < 6 {
		t.Fatalf("expected at least 6 inbox items, got %d", len(payload.Items))
	}
}

func ptrTime(v time.Time) *time.Time {
	return &v
}

func TestGetMentionsReturnsOnlyDirectMentions(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	users := []domain.User{
		{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"},
		{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"},
	}
	for _, user := range users {
		db.DB.Create(&user)
	}
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-mention", ChannelID: "ch-1", UserID: "user-2", Content: "@Nikko Fu for review", CreatedAt: now.Add(-10 * time.Minute), Metadata: `{"user_mentions":[{"user_id":"user-1","name":"Nikko Fu","mention_text":"@Nikko Fu"}]}`})
	db.DB.Create(&domain.Message{ID: "msg-other", ChannelID: "ch-1", UserID: "user-2", Content: "No direct mention here", CreatedAt: now.Add(-5 * time.Minute), Metadata: "{}"})
	db.DB.Create(&domain.MessageMention{ID: "mm-1", MessageID: "msg-mention", WorkspaceID: "ws-1", ChannelID: "ch-1", MentionedUserID: "user-1", MentionedByUserID: "user-2", MentionText: "@Nikko Fu", MentionKind: "user", CreatedAt: now.Add(-10 * time.Minute)})

	router := gin.New()
	router.GET("/api/v1/mentions", GetMentions)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mentions", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on mentions, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode mentions payload: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected exactly 1 mention item, got %d", len(payload.Items))
	}
	if payload.Items[0]["type"] != "mention" {
		t.Fatalf("expected mention type, got %#v", payload.Items[0])
	}
}

func TestGetLaterReturnsSavedMessages(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-1",
		UserID:    "user-2",
		Content:   "Save me for later",
		CreatedAt: time.Now().UTC(),
		Metadata:  "{}",
	})
	db.DB.Create(&domain.SavedMessage{
		MessageID: "msg-1",
		UserID:    "user-1",
		CreatedAt: time.Now().UTC(),
	})

	router := gin.New()
	router.GET("/api/v1/later", GetLater)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/later", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on later, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode later payload: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 saved item, got %d", len(payload.Items))
	}
}

func TestGetDraftsReturnsCurrentUserDrafts(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com"})
	db.DB.Create(&domain.Draft{
		UserID:    "user-1",
		Scope:     "channel:ch-1",
		Content:   "Draft for general",
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	})
	db.DB.Create(&domain.Draft{
		UserID:    "user-1",
		Scope:     "dm:dm-1",
		Content:   "Draft for DM",
		CreatedAt: now.Add(-30 * time.Minute),
		UpdatedAt: now.Add(-30 * time.Minute),
	})
	db.DB.Create(&domain.Draft{
		UserID:    "user-2",
		Scope:     "channel:ch-2",
		Content:   "Other user's draft",
		CreatedAt: now,
		UpdatedAt: now,
	})

	router := gin.New()
	router.GET("/api/v1/drafts", GetDrafts)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drafts", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on drafts list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Drafts []domain.Draft `json:"drafts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode drafts payload: %v", err)
	}
	if len(payload.Drafts) != 2 {
		t.Fatalf("expected 2 drafts for current user, got %d", len(payload.Drafts))
	}
	if payload.Drafts[0].Scope != "dm:dm-1" {
		t.Fatalf("expected newest draft first, got %#v", payload.Drafts)
	}
}

func TestPutDraftCreatesAndUpdatesByScope(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})

	router := gin.New()
	router.PUT("/api/v1/drafts/:scope", PutDraft)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/drafts/channel:ch-1", bytes.NewBufferString(`{"content":"First draft"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on draft create, got %d body=%s", rec.Code, rec.Body.String())
	}

	var created struct {
		Draft domain.Draft `json:"draft"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode create draft response: %v", err)
	}
	if created.Draft.Scope != "channel:ch-1" || created.Draft.Content != "First draft" {
		t.Fatalf("unexpected created draft: %#v", created.Draft)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/v1/drafts/channel:ch-1", bytes.NewBufferString(`{"content":"Updated draft"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on draft update, got %d body=%s", rec.Code, rec.Body.String())
	}

	var draft domain.Draft
	if err := db.DB.First(&draft, "user_id = ? AND scope = ?", "user-1", "channel:ch-1").Error; err != nil {
		t.Fatalf("failed to reload draft: %v", err)
	}
	if draft.Content != "Updated draft" {
		t.Fatalf("expected updated content, got %#v", draft)
	}
}

func TestDeleteDraftRemovesScopedDraft(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Draft{UserID: "user-1", Scope: "channel:ch-1", Content: "Disposable draft", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()})

	router := gin.New()
	router.DELETE("/api/v1/drafts/:scope", DeleteDraft)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/drafts/channel:ch-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on draft delete, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Deleted bool   `json:"deleted"`
		Scope   string `json:"scope"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode delete draft response: %v", err)
	}
	if !payload.Deleted || payload.Scope != "channel:ch-1" {
		t.Fatalf("unexpected delete draft payload: %#v", payload)
	}

	var count int64
	db.DB.Model(&domain.Draft{}).Where("user_id = ? AND scope = ?", "user-1", "channel:ch-1").Count(&count)
	if count != 0 {
		t.Fatalf("expected draft deletion, got count=%d", count)
	}
}

func TestPresenceEndpointsListAndUpdateStatus(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com", Status: "online"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com", Status: "away"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1", Role: "owner"})

	router := gin.New()
	router.GET("/api/v1/presence", GetPresence)
	router.POST("/api/v1/presence", UpdatePresence)
	router.POST("/api/v1/presence/heartbeat", HeartbeatPresence)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/presence", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on presence list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Users []domain.User `json:"users"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode presence list: %v", err)
	}
	if len(listPayload.Users) != 2 {
		t.Fatalf("expected 2 users in presence list, got %d", len(listPayload.Users))
	}
	if listPayload.Users[0].StatusText == "" {
		t.Fatalf("expected derived status_text, got %#v", listPayload.Users[0])
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/presence?channel_id=ch-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on scoped presence list, got %d body=%s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode scoped presence list: %v", err)
	}
	if len(listPayload.Users) != 1 || listPayload.Users[0].ID != "user-1" {
		t.Fatalf("expected only channel member in scoped presence list, got %#v", listPayload.Users)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/presence", bytes.NewBufferString(`{"status":"busy","status_text":"Heads down shipping"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on presence update, got %d body=%s", rec.Code, rec.Body.String())
	}

	var refreshed domain.User
	if err := db.DB.First(&refreshed, "id = ?", "user-1").Error; err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if refreshed.Status != "busy" {
		t.Fatalf("expected busy status, got %q", refreshed.Status)
	}
	if refreshed.StatusText != "Heads down shipping" {
		t.Fatalf("expected status text to persist, got %#v", refreshed)
	}
	if refreshed.LastSeenAt == nil || refreshed.PresenceExpiresAt == nil {
		t.Fatalf("expected presence timestamps to persist, got %#v", refreshed)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/presence/heartbeat", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on presence heartbeat, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPresenceAndTypingBroadcastRealtimeEvents(t *testing.T) {
	setupTestDB(t)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com", Status: "online"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})

	router := gin.New()
	router.POST("/api/v1/presence", UpdatePresence)
	router.POST("/api/v1/presence/heartbeat", HeartbeatPresence)
	router.POST("/api/v1/typing", UpdateTyping)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/presence", bytes.NewBufferString(`{"status":"away"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on presence update, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "presence.updated")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/presence/heartbeat", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on presence heartbeat, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "presence.updated")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/typing", bytes.NewBufferString(`{"channel_id":"ch-1","is_typing":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on typing update, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "typing.updated")
}

func TestStarredChannelsEndpointsListAndToggle(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public", IsStarred: true})
	db.DB.Create(&domain.Channel{ID: "ch-2", WorkspaceID: "ws-1", Name: "design", Type: "public", IsStarred: false})

	router := gin.New()
	router.GET("/api/v1/starred", GetStarredChannels)
	router.POST("/api/v1/channels/:id/star", ToggleChannelStar)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/starred", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on starred list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Channels []domain.Channel `json:"channels"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode starred channels: %v", err)
	}
	if len(listPayload.Channels) != 1 || listPayload.Channels[0].ID != "ch-1" {
		t.Fatalf("expected only starred channel, got %#v", listPayload.Channels)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/channels/ch-2/star", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on toggle star, got %d body=%s", rec.Code, rec.Body.String())
	}

	var refreshed domain.Channel
	if err := db.DB.First(&refreshed, "id = ?", "ch-2").Error; err != nil {
		t.Fatalf("failed to reload channel: %v", err)
	}
	if !refreshed.IsStarred {
		t.Fatal("expected channel to be starred")
	}
}

func TestGetPinsReturnsPinnedMessagesWithChannelAndUser(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.Channel{ID: "ch-2", WorkspaceID: "ws-1", Name: "game", Type: "public"})
	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-1",
		UserID:    "user-2",
		Content:   "Pinned update",
		IsPinned:  true,
		CreatedAt: now,
		Metadata:  "{}",
	})
	db.DB.Create(&domain.Message{
		ID:        "msg-2",
		ChannelID: "ch-1",
		UserID:    "user-2",
		Content:   "Regular update",
		IsPinned:  false,
		CreatedAt: now.Add(time.Minute),
		Metadata:  "{}",
	})
	db.DB.Create(&domain.Message{
		ID:        "msg-3",
		ChannelID: "ch-2",
		UserID:    "user-2",
		Content:   "Pinned elsewhere",
		IsPinned:  true,
		CreatedAt: now.Add(2 * time.Minute),
		Metadata:  "{}",
	})

	router := gin.New()
	router.GET("/api/v1/pins", GetPins)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pins", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on pins list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode pins payload: %v", err)
	}
	if len(payload.Items) != 2 {
		t.Fatalf("expected 2 pinned items globally, got %d", len(payload.Items))
	}
	message, ok := payload.Items[0]["message"].(map[string]any)
	if !ok || message["id"] != "msg-3" {
		t.Fatalf("expected newest pinned message payload, got %#v", payload.Items[0]["message"])
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/pins?channel_id=ch-2", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on filtered pins list, got %d body=%s", rec.Code, rec.Body.String())
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode filtered pins payload: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 filtered pinned item, got %d", len(payload.Items))
	}
	message, ok = payload.Items[0]["message"].(map[string]any)
	if !ok || message["id"] != "msg-3" {
		t.Fatalf("expected channel-scoped pinned message payload, got %#v", payload.Items[0]["message"])
	}
}

func TestInboxIncludesReadStateAndNotificationsCanBeMarkedRead(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	users := []domain.User{
		{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"},
		{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"},
	}
	for _, user := range users {
		db.DB.Create(&user)
	}
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.Message{
		ID:        "msg-mention",
		ChannelID: "ch-1",
		UserID:    "user-2",
		Content:   "@Nikko Fu for review",
		CreatedAt: now,
		Metadata:  `{"user_mentions":[{"user_id":"user-1","name":"Nikko Fu","mention_text":"@Nikko Fu"}]}`,
	})
	db.DB.Create(&domain.MessageMention{
		ID:                "mm-read-1",
		MessageID:         "msg-mention",
		WorkspaceID:       "ws-1",
		ChannelID:         "ch-1",
		MentionedUserID:   "user-1",
		MentionedByUserID: "user-2",
		MentionText:       "@Nikko Fu",
		MentionKind:       "user",
		CreatedAt:         now,
	})

	router := gin.New()
	router.GET("/api/v1/inbox", GetInbox)
	router.POST("/api/v1/notifications/read", MarkNotificationsRead)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/inbox", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on inbox, got %d body=%s", rec.Code, rec.Body.String())
	}

	var inboxPayload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &inboxPayload); err != nil {
		t.Fatalf("failed to decode inbox payload: %v", err)
	}
	if len(inboxPayload.Items) != 1 {
		t.Fatalf("expected 1 inbox item, got %d", len(inboxPayload.Items))
	}
	if inboxPayload.Items[0]["is_read"] != false {
		t.Fatalf("expected unread inbox item, got %#v", inboxPayload.Items[0])
	}

	itemID, ok := inboxPayload.Items[0]["id"].(string)
	if !ok || itemID == "" {
		t.Fatalf("expected inbox item id, got %#v", inboxPayload.Items[0]["id"])
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/notifications/read", bytes.NewBufferString(`{"item_ids":["`+itemID+`"]}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on mark read, got %d body=%s", rec.Code, rec.Body.String())
	}

	var reads int64
	db.DB.Model(&domain.NotificationRead{}).Where("user_id = ? AND item_id = ?", "user-1", itemID).Count(&reads)
	if reads != 1 {
		t.Fatalf("expected notification read row, got %d", reads)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/inbox", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on inbox refetch, got %d body=%s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &inboxPayload); err != nil {
		t.Fatalf("failed to decode inbox payload after mark read: %v", err)
	}
	if inboxPayload.Items[0]["is_read"] != true {
		t.Fatalf("expected read inbox item after mark read, got %#v", inboxPayload.Items[0])
	}
}

func TestSearchReturnsChannelUserMessageAndDMHits(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-5", WorkspaceID: "ws-1", Name: "ai-lab", Description: "AI experiments", Type: "public"})
	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-5",
		UserID:    "user-2",
		Content:   "The AI lab launch plan is ready",
		CreatedAt: time.Now().UTC(),
		Metadata:  "{}",
	})
	db.DB.Create(&domain.DMConversation{ID: "dm-1", CreatedAt: time.Now().UTC()})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-1"})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-2"})
	db.DB.Create(&domain.DMMessage{
		ID:               "dm-msg-1",
		DMConversationID: "dm-1",
		UserID:           "user-2",
		Content:          "AI follow-up in DM",
		CreatedAt:        time.Now().UTC(),
	})
	db.DB.Create(&domain.Artifact{
		ID:        "artifact-1",
		ChannelID: "ch-5",
		Title:     "AI research canvas",
		Version:   1,
		Type:      "document",
		Status:    "live",
		Content:   "AI research notes",
		Source:    "manual",
		CreatedBy: "user-1",
		UpdatedBy: "user-1",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	db.DB.Create(&domain.FileAsset{
		ID:          "file-1",
		ChannelID:   "ch-5",
		UploaderID:  "user-1",
		Name:        "ai-brief.pdf",
		StoragePath: "ai-brief.pdf",
		ContentType: "application/pdf",
		SizeBytes:   4096,
		CreatedAt:   time.Now().UTC(),
	})

	router := gin.New()
	router.GET("/api/v1/search", SearchWorkspace)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=AI", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on search, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Query   string `json:"query"`
		Results struct {
			Channels  []map[string]any `json:"channels"`
			Users     []map[string]any `json:"users"`
			Messages  []map[string]any `json:"messages"`
			DMs       []map[string]any `json:"dms"`
			Artifacts []map[string]any `json:"artifacts"`
			Files     []map[string]any `json:"files"`
		} `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode search payload: %v", err)
	}
	if payload.Query != "AI" {
		t.Fatalf("expected query echo, got %q", payload.Query)
	}
	if len(payload.Results.Channels) == 0 || len(payload.Results.Users) == 0 || len(payload.Results.Messages) == 0 || len(payload.Results.DMs) == 0 || len(payload.Results.Artifacts) == 0 || len(payload.Results.Files) == 0 {
		t.Fatalf("expected all result groups to have hits, got %#v", payload.Results)
	}
	if payload.Results.Messages[0]["snippet"] == "" {
		t.Fatalf("expected message search results to include snippet, got %#v", payload.Results.Messages[0])
	}
	if payload.Results.Channels[0]["match_reason"] == "" {
		t.Fatalf("expected channel search results to include match_reason, got %#v", payload.Results.Channels[0])
	}
	if payload.Results.Artifacts[0]["title"] == "" || payload.Results.Files[0]["name"] == "" {
		t.Fatalf("expected artifacts and files search results to be populated, got %#v %#v", payload.Results.Artifacts, payload.Results.Files)
	}
}

func TestSearchSuggestionsReturnsRankedMatches(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-5", WorkspaceID: "ws-1", Name: "ai-lab", Description: "AI experiments", Type: "public"})
	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-5",
		UserID:    "user-2",
		Content:   "The AI lab launch plan is ready",
		CreatedAt: time.Now().UTC(),
		Metadata:  "{}",
	})
	db.DB.Create(&domain.Artifact{
		ID:        "artifact-1",
		ChannelID: "ch-5",
		Title:     "AI research canvas",
		Version:   1,
		Type:      "document",
		Status:    "live",
		Content:   "AI research notes",
		Source:    "manual",
		CreatedBy: "user-1",
		UpdatedBy: "user-1",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	db.DB.Create(&domain.FileAsset{
		ID:          "file-1",
		ChannelID:   "ch-5",
		UploaderID:  "user-1",
		Name:        "ai-brief.pdf",
		StoragePath: "ai-brief.pdf",
		ContentType: "application/pdf",
		SizeBytes:   4096,
		CreatedAt:   time.Now().UTC(),
	})

	router := gin.New()
	router.GET("/api/v1/search/suggestions", SearchSuggestions)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/suggestions?q=ai", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on search suggestions, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Query       string           `json:"query"`
		Suggestions []map[string]any `json:"suggestions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode search suggestions payload: %v", err)
	}
	if payload.Query != "ai" {
		t.Fatalf("expected query echo, got %q", payload.Query)
	}
	if len(payload.Suggestions) < 2 {
		t.Fatalf("expected at least 2 suggestions, got %#v", payload.Suggestions)
	}
	if payload.Suggestions[0]["type"] == nil || payload.Suggestions[0]["label"] == nil {
		t.Fatalf("expected typed suggestion entries, got %#v", payload.Suggestions[0])
	}
	foundArtifact := false
	foundFile := false
	for _, suggestion := range payload.Suggestions {
		switch suggestion["type"] {
		case "artifact":
			foundArtifact = true
		case "file":
			foundFile = true
		}
	}
	if !foundArtifact || !foundFile {
		t.Fatalf("expected artifact and file suggestions, got %#v", payload.Suggestions)
	}
}

func TestIntelligentSearchReturnsRankedKnowledgeResults(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-5", WorkspaceID: "ws-1", Name: "launch-war-room", Description: "Launch coordination", Type: "public"})
	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-5",
		UserID:    "user-2",
		Content:   "Launch checklist and rollout notes",
		CreatedAt: time.Now().UTC(),
		Metadata:  "{}",
	})
	db.DB.Create(&domain.Artifact{
		ID:        "artifact-1",
		ChannelID: "ch-5",
		Title:     "Launch checklist canvas",
		Version:   1,
		Type:      "document",
		Status:    "live",
		Content:   "Checklist for launch readiness",
		Source:    "manual",
		CreatedBy: "user-1",
		UpdatedBy: "user-1",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	router := gin.New()
	router.GET("/api/v1/search/intelligent", IntelligentSearch)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/intelligent?q=launch%20checklist", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on intelligent search, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Query  string `json:"query"`
		Ranked []struct {
			Type   string  `json:"type"`
			ID     string  `json:"id"`
			Label  string  `json:"label"`
			Reason string  `json:"reason"`
			Score  float64 `json:"score"`
		} `json:"ranked"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode intelligent search payload: %v", err)
	}
	if payload.Query != "launch checklist" {
		t.Fatalf("expected query echo, got %q", payload.Query)
	}
	if len(payload.Ranked) < 2 {
		t.Fatalf("expected ranked results, got %#v", payload.Ranked)
	}
	if payload.Ranked[0].Type == "" || payload.Ranked[0].Label == "" || payload.Ranked[0].Reason == "" || payload.Ranked[0].Score <= 0 {
		t.Fatalf("expected ranked result fields to be populated, got %#v", payload.Ranked[0])
	}
}

func TestMarkNotificationsReadBroadcastsRealtimeEvent(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(4)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	router := gin.New()
	router.POST("/api/v1/notifications/read", MarkNotificationsRead)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/read", bytes.NewBufferString(`{"item_ids":["activity-mention-msg-1","activity-thread-msg-2"]}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on mark notifications read, got %d body=%s", rec.Code, rec.Body.String())
	}

	raw, err := client.Receive(time.Second)
	if err != nil {
		t.Fatalf("failed to receive realtime event: %v", err)
	}

	var event realtime.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("failed to decode realtime event: %v", err)
	}
	if event.Type != "notifications.read" {
		t.Fatalf("expected notifications.read event, got %s", event.Type)
	}
}

func TestChannelMembersEndpointsListAddAndRemoveMembers(t *testing.T) {
	setupTestDB(t)

	users := []domain.User{
		{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"},
		{ID: "user-2", Name: "AI Assistant", Email: "ai@example.com"},
		{ID: "user-3", Name: "Jane Smith", Email: "jane@example.com"},
	}
	for _, user := range users {
		db.DB.Create(&user)
	}
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1", Role: "owner"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-2", Role: "member"})

	router := gin.New()
	router.GET("/api/v1/channels/:id/members", GetChannelMembers)
	router.POST("/api/v1/channels/:id/members", AddChannelMember)
	router.DELETE("/api/v1/channels/:id/members/:userId", RemoveChannelMember)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/ch-1/members", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on members list, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Members []map[string]any `json:"members"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode members list: %v", err)
	}
	if len(listPayload.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(listPayload.Members))
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/channels/ch-1/members", bytes.NewBufferString(`{"user_id":"user-3","role":"member"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on add member, got %d body=%s", rec.Code, rec.Body.String())
	}

	var memberCount int64
	db.DB.Model(&domain.ChannelMember{}).Where("channel_id = ?", "ch-1").Count(&memberCount)
	if memberCount != 3 {
		t.Fatalf("expected 3 channel members, got %d", memberCount)
	}

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", "ch-1").Error; err != nil {
		t.Fatalf("failed to reload channel: %v", err)
	}
	if channel.MemberCount != 3 {
		t.Fatalf("expected channel member_count=3, got %d", channel.MemberCount)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/channels/ch-1/members/user-2", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on remove member, got %d body=%s", rec.Code, rec.Body.String())
	}

	db.DB.Model(&domain.ChannelMember{}).Where("channel_id = ?", "ch-1").Count(&memberCount)
	if memberCount != 2 {
		t.Fatalf("expected 2 channel members after delete, got %d", memberCount)
	}
}

func TestChannelPreferencesAndLeaveEndpoint(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public", MemberCount: 1})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1", Role: "member"})

	router := gin.New()
	router.GET("/api/v1/channels/:id/preferences", GetChannelPreferences)
	router.PATCH("/api/v1/channels/:id/preferences", PatchChannelPreferences)
	router.POST("/api/v1/channels/:id/leave", LeaveChannel)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/ch-1/preferences", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on preference get, got %d body=%s", rec.Code, rec.Body.String())
	}

	var getPayload struct {
		Preferences struct {
			NotificationLevel string `json:"notification_level"`
			IsMuted           bool   `json:"is_muted"`
		} `json:"preferences"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &getPayload); err != nil {
		t.Fatalf("failed to decode preference get: %v", err)
	}
	if getPayload.Preferences.NotificationLevel != "all" || getPayload.Preferences.IsMuted {
		t.Fatalf("unexpected default preferences: %#v", getPayload.Preferences)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/channels/ch-1/preferences", bytes.NewBufferString(`{"notification_level":"mentions","is_muted":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on preference patch, got %d body=%s", rec.Code, rec.Body.String())
	}

	var pref domain.ChannelPreference
	if err := db.DB.First(&pref, "channel_id = ? AND user_id = ?", "ch-1", "user-1").Error; err != nil {
		t.Fatalf("failed to load persisted channel preference: %v", err)
	}
	if pref.NotificationLevel != "mentions" || !pref.IsMuted {
		t.Fatalf("unexpected persisted channel preference: %#v", pref)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/channels/ch-1/leave", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on channel leave, got %d body=%s", rec.Code, rec.Body.String())
	}

	var memberCount int64
	db.DB.Model(&domain.ChannelMember{}).Where("channel_id = ?", "ch-1").Count(&memberCount)
	if memberCount != 0 {
		t.Fatalf("expected channel member to be removed, got %d", memberCount)
	}
	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", "ch-1").Error; err != nil {
		t.Fatalf("failed to reload channel: %v", err)
	}
	if channel.MemberCount != 0 {
		t.Fatalf("expected channel member_count=0, got %d", channel.MemberCount)
	}
}

func TestPatchChannelUpdatesTopicPurposeAndArchiveState(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.Channel{
		ID:          "ch-1",
		WorkspaceID: "ws-1",
		Name:        "general",
		Type:        "public",
		Description: "General discussion",
	})

	router := gin.New()
	router.PATCH("/api/v1/channels/:id", UpdateChannel)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/channels/ch-1", bytes.NewBufferString(`{"topic":"Launch coordination","purpose":"Keep launch work aligned","is_archived":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on patch channel, got %d body=%s", rec.Code, rec.Body.String())
	}

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", "ch-1").Error; err != nil {
		t.Fatalf("failed to reload channel: %v", err)
	}
	if channel.Topic != "Launch coordination" || channel.Purpose != "Keep launch work aligned" || !channel.IsArchived {
		t.Fatalf("channel patch was not persisted: %#v", channel)
	}
}

func TestFilePreviewEndpointReturnsPreviewMetadata(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.FileAsset{
		ID:          "file-1",
		ChannelID:   "ch-1",
		UploaderID:  "user-1",
		Name:        "roadmap.png",
		StoragePath: "roadmap.png",
		ContentType: "image/png",
		SizeBytes:   2048,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	router := gin.New()
	router.GET("/api/v1/files/:id/preview", GetFilePreview)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/file-1/preview", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on file preview, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Preview struct {
			FileID        string `json:"file_id"`
			Name          string `json:"name"`
			ContentType   string `json:"content_type"`
			PreviewKind   string `json:"preview_kind"`
			PreviewURL    string `json:"preview_url"`
			DownloadURL   string `json:"download_url"`
			IsPreviewable bool   `json:"is_previewable"`
			Size          int64  `json:"size"`
		} `json:"preview"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode file preview: %v", err)
	}
	if payload.Preview.FileID != "file-1" || payload.Preview.PreviewKind != "image" || !payload.Preview.IsPreviewable {
		t.Fatalf("unexpected file preview payload: %#v", payload.Preview)
	}
	if payload.Preview.PreviewURL != "/api/v1/files/file-1/content" || payload.Preview.DownloadURL != "/api/v1/files/file-1/content" {
		t.Fatalf("unexpected preview/download urls: %#v", payload.Preview)
	}
}

func TestCreateChannelCreatesChannelAndOwnerMembership(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})

	router := gin.New()
	router.POST("/api/v1/channels", CreateChannel)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels", bytes.NewBufferString(`{"workspace_id":"ws-1","name":"launch","description":"Launch planning","type":"private"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create channel, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Channel domain.Channel `json:"channel"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode create channel payload: %v", err)
	}
	if payload.Channel.Name != "launch" || payload.Channel.Type != "private" || payload.Channel.MemberCount != 1 {
		t.Fatalf("unexpected created channel payload: %#v", payload.Channel)
	}
	assertPrefixedUUID(t, payload.Channel.ID, "ch")

	var membership domain.ChannelMember
	if err := db.DB.First(&membership, "channel_id = ? AND user_id = ?", payload.Channel.ID, "user-1").Error; err != nil {
		t.Fatalf("expected creator membership to exist: %v", err)
	}
	if membership.Role != "owner" {
		t.Fatalf("expected creator role owner, got %#v", membership)
	}
}

func TestCreateChannelRejectsUnknownWorkspace(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})

	router := gin.New()
	router.POST("/api/v1/channels", CreateChannel)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels", bytes.NewBufferString(`{"workspace_id":"ws_1","name":"game","description":"Game room","type":"public"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown workspace, got %d body=%s", rec.Code, rec.Body.String())
	}

	var count int64
	db.DB.Model(&domain.Channel{}).Where("name = ?", "game").Count(&count)
	if count != 0 {
		t.Fatalf("expected no orphan channel to be created, got %d", count)
	}
}

func TestWorkspaceInvitesEndpointsCreateAndListInvites(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})

	router := gin.New()
	router.GET("/api/v1/workspaces/:id/invites", GetWorkspaceInvites)
	router.POST("/api/v1/workspaces/:id/invites", CreateWorkspaceInvite)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/invites", bytes.NewBufferString(`{"email":"new.member@example.com","role":"member"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create invite, got %d body=%s", rec.Code, rec.Body.String())
	}

	var inviteCount int64
	db.DB.Model(&domain.WorkspaceInvite{}).Where("workspace_id = ?", "ws-1").Count(&inviteCount)
	if inviteCount != 1 {
		t.Fatalf("expected 1 workspace invite, got %d", inviteCount)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/invites", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on list invites, got %d body=%s", rec.Code, rec.Body.String())
	}

	var listPayload struct {
		Invites []map[string]any `json:"invites"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode invite list: %v", err)
	}
	if len(listPayload.Invites) != 1 {
		t.Fatalf("expected 1 invite in response, got %d", len(listPayload.Invites))
	}
	inviteID, ok := listPayload.Invites[0]["id"].(string)
	if !ok || inviteID == "" {
		t.Fatalf("expected invite id in list payload, got %#v", listPayload.Invites[0])
	}
	assertPrefixedUUID(t, inviteID, "invite")
}

func TestReactionPinAndDeleteBroadcastRealtimeEvents(t *testing.T) {
	setupTestDB(t)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
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
	router.POST("/api/v1/messages/:id/pin", TogglePinMessage)
	router.DELETE("/api/v1/messages/:id", DeleteMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/msg-1/reactions", bytes.NewBufferString(`{"emoji":"🔥"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on reaction, got %d", rec.Code)
	}
	assertRealtimeEventType(t, client, "reaction.updated")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/messages/msg-1/pin", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on pin, got %d", rec.Code)
	}
	assertRealtimeEventType(t, client, "message.updated")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/messages/msg-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %d", rec.Code)
	}
	assertRealtimeEventType(t, client, "message.deleted")
}

func assertRealtimeEventType(t *testing.T, client *realtime.TestClient, expectedType string) {
	t.Helper()

	raw, err := client.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("failed to receive realtime event: %v", err)
	}

	var event realtime.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("failed to decode realtime event: %v", err)
	}
	if event.Type != expectedType {
		t.Fatalf("expected realtime event %s, got %s", expectedType, event.Type)
	}
}

func TestFileExtractionModelsPersist(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	extraction := domain.FileExtraction{
		ID:        ids.NewPrefixedUUID("fextract"),
		FileID:    "file-1",
		Status:    "pending",
		Extractor: "plain_text",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.DB.Create(&extraction).Error; err != nil {
		t.Fatalf("failed to create extraction: %v", err)
	}

	chunk := domain.FileExtractionChunk{
		ExtractionID:  extraction.ID,
		FileID:        "file-1",
		ChunkIndex:    0,
		Text:          "hello world",
		TokenEstimate: 2,
		LocatorType:   "document",
		LocatorValue:  "root",
		CreatedAt:     now,
	}
	if err := db.DB.Create(&chunk).Error; err != nil {
		t.Fatalf("failed to create extraction chunk: %v", err)
	}

	var persistedExtraction domain.FileExtraction
	if err := db.DB.First(&persistedExtraction, "id = ?", extraction.ID).Error; err != nil {
		t.Fatalf("failed to load extraction: %v", err)
	}
	if persistedExtraction.FileID != "file-1" || persistedExtraction.Status != "pending" {
		t.Fatalf("unexpected persisted extraction: %#v", persistedExtraction)
	}

	var persistedChunk domain.FileExtractionChunk
	if err := db.DB.First(&persistedChunk, "file_id = ? AND chunk_index = ?", "file-1", 0).Error; err != nil {
		t.Fatalf("failed to load extraction chunk: %v", err)
	}
	if persistedChunk.Text != "hello world" || persistedChunk.LocatorType != "document" {
		t.Fatalf("unexpected persisted chunk: %#v", persistedChunk)
	}
}

func TestUploadFileCreatesExtractionRecord(t *testing.T) {
	setupTestDB(t)
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})

	_ = os.RemoveAll("uploads")
	t.Cleanup(func() {
		_ = os.RemoveAll("uploads")
	})

	router := gin.New()
	router.POST("/api/v1/files/upload", UploadFile)

	body, contentType := buildMultipartFileUpload(t, "channel_id", "ch-1", "notes.txt", []byte("Launch checklist\n\nShip file search"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
	req.Header.Set("Content-Type", contentType)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on file upload, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		File domain.FileAsset `json:"file"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode upload payload: %v", err)
	}

	var extraction domain.FileExtraction
	if err := db.DB.First(&extraction, "file_id = ?", payload.File.ID).Error; err != nil {
		t.Fatalf("expected extraction record for uploaded file: %v", err)
	}
	if extraction.Status != "ready" || extraction.ContentSummary == "" {
		t.Fatalf("expected ready extraction summary, got %#v", extraction)
	}

	var chunks []domain.FileExtractionChunk
	if err := db.DB.Where("file_id = ?", payload.File.ID).Find(&chunks).Error; err != nil {
		t.Fatalf("failed to load extraction chunks: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatalf("expected extraction chunks for uploaded file")
	}
}

func TestRebuildFileExtractionRegeneratesExtraction(t *testing.T) {
	setupTestDB(t)
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})

	_ = os.RemoveAll("uploads")
	t.Cleanup(func() {
		_ = os.RemoveAll("uploads")
	})

	router := gin.New()
	router.POST("/api/v1/files/upload", UploadFile)
	router.POST("/api/v1/files/:id/extraction/rebuild", RebuildFileExtraction)

	body, contentType := buildMultipartFileUpload(t, "channel_id", "ch-1", "notes.txt", []byte("First extraction body"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
	req.Header.Set("Content-Type", contentType)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on file upload, got %d body=%s", rec.Code, rec.Body.String())
	}

	var uploadPayload struct {
		File domain.FileAsset `json:"file"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &uploadPayload); err != nil {
		t.Fatalf("failed to decode upload payload: %v", err)
	}

	var asset domain.FileAsset
	if err := db.DB.First(&asset, "id = ?", uploadPayload.File.ID).Error; err != nil {
		t.Fatalf("failed to load uploaded asset: %v", err)
	}
	updatedPath := filepath.Join("uploads", asset.StoragePath)
	if err := os.WriteFile(updatedPath, []byte("Updated extraction body with different summary"), 0o644); err != nil {
		t.Fatalf("failed to update uploaded file contents: %v", err)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/files/"+asset.ID+"/extraction/rebuild", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on extraction rebuild, got %d body=%s", rec.Code, rec.Body.String())
	}

	var extraction domain.FileExtraction
	if err := db.DB.First(&extraction, "file_id = ?", asset.ID).Error; err != nil {
		t.Fatalf("failed to reload extraction: %v", err)
	}
	if !strings.Contains(extraction.ContentText, "Updated extraction body") {
		t.Fatalf("expected rebuilt extraction content, got %#v", extraction)
	}
}

func TestImageFileExtractionUsesMockOCR(t *testing.T) {
	setupTestDB(t)
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})

	_ = os.RemoveAll("uploads")
	t.Cleanup(func() {
		_ = os.RemoveAll("uploads")
	})

	router := gin.New()
	router.POST("/api/v1/files/upload", UploadFile)

	body, contentType := buildMultipartFileUpload(t, "channel_id", "ch-1", "diagram.png", []byte("fake-image"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
	req.Header.Set("Content-Type", contentType)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on image upload, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		File domain.FileAsset `json:"file"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode upload payload: %v", err)
	}

	var extraction domain.FileExtraction
	if err := db.DB.First(&extraction, "file_id = ?", payload.File.ID).Error; err != nil {
		t.Fatalf("expected extraction record for uploaded image: %v", err)
	}
	if extraction.OCRProvider != "mock" || !extraction.OCRIsMock || !extraction.NeedsOCR {
		t.Fatalf("expected mock OCR metadata, got %#v", extraction)
	}
	if !strings.Contains(extraction.ContentText, "Mock OCR text extracted from diagram.png") {
		t.Fatalf("expected mock OCR text, got %#v", extraction)
	}
}

func TestGetFileExtractionReturnsStatusAndSummary(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.FileAsset{
		ID:               "file-1",
		Name:             "notes.txt",
		StoragePath:      "notes.txt",
		UploaderID:       "user-1",
		ContentType:      "text/plain",
		ExtractionStatus: "ready",
		ContentSummary:   "Launch checklist summary",
		LastIndexedAt:    &now,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	db.DB.Create(&domain.FileExtraction{
		ID:             ids.NewPrefixedUUID("fextract"),
		FileID:         "file-1",
		Status:         "ready",
		Extractor:      "plain_text",
		ContentSummary: "Launch checklist summary",
		CreatedAt:      now,
		UpdatedAt:      now,
		CompletedAt:    &now,
	})

	router := gin.New()
	router.GET("/api/v1/files/:id/extraction", GetFileExtraction)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/file-1/extraction", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get extraction, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetFileContentReturnsExtractedText(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.FileAsset{
		ID:          "file-1",
		Name:        "notes.txt",
		StoragePath: "notes.txt",
		UploaderID:  "user-1",
		ContentType: "text/plain",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	db.DB.Create(&domain.FileExtraction{
		ID:             ids.NewPrefixedUUID("fextract"),
		FileID:         "file-1",
		Status:         "ready",
		Extractor:      "plain_text",
		ContentText:    "Launch checklist\n\nShip search",
		ContentSummary: "Launch checklist summary",
		CreatedAt:      now,
		UpdatedAt:      now,
		CompletedAt:    &now,
	})

	router := gin.New()
	router.GET("/api/v1/files/:id/content", GetFileExtractedContent)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/file-1/content", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get file content, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetFileChunksReturnsLocatorMetadata(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	extractionID := ids.NewPrefixedUUID("fextract")
	db.DB.Create(&domain.FileAsset{
		ID:          "file-1",
		Name:        "notes.txt",
		StoragePath: "notes.txt",
		UploaderID:  "user-1",
		ContentType: "text/plain",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	db.DB.Create(&domain.FileExtraction{
		ID:             extractionID,
		FileID:         "file-1",
		Status:         "ready",
		Extractor:      "plain_text",
		ContentText:    "Launch checklist",
		ContentSummary: "Launch checklist summary",
		CreatedAt:      now,
		UpdatedAt:      now,
		CompletedAt:    &now,
	})
	db.DB.Create(&domain.FileExtractionChunk{
		ExtractionID:  extractionID,
		FileID:        "file-1",
		ChunkIndex:    0,
		Text:          "Launch checklist",
		TokenEstimate: 2,
		LocatorType:   "document",
		LocatorValue:  "root",
		Heading:       "Launch",
		CreatedAt:     now,
	})

	router := gin.New()
	router.GET("/api/v1/files/:id/chunks", GetFileExtractionChunks)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/file-1/chunks", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get file chunks, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSearchFilesReturnsContentHitsAndLocators(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	extractionID := ids.NewPrefixedUUID("fextract")
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.FileAsset{
		ID:               "file-1",
		Name:             "launch-plan.docx",
		StoragePath:      "launch-plan.docx",
		UploaderID:       "user-1",
		ContentType:      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		ExtractionStatus: "ready",
		ContentSummary:   "Launch plan summary",
		LastIndexedAt:    &now,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	db.DB.Create(&domain.FileExtraction{
		ID:             extractionID,
		FileID:         "file-1",
		Status:         "ready",
		Extractor:      "docx",
		ContentText:    "Launch checklist and rollout plan",
		ContentSummary: "Launch plan summary",
		CreatedAt:      now,
		UpdatedAt:      now,
		CompletedAt:    &now,
	})
	db.DB.Create(&domain.FileExtractionChunk{
		ExtractionID:  extractionID,
		FileID:        "file-1",
		ChunkIndex:    0,
		Text:          "Launch checklist and rollout plan",
		TokenEstimate: 5,
		LocatorType:   "document",
		LocatorValue:  "root",
		Heading:       "Launch",
		CreatedAt:     now,
	})

	router := gin.New()
	router.GET("/api/v1/search/files", SearchFiles)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/files?q=launch", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on search files, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetFileCitationsReturnsChunkCandidates(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	extractionID := ids.NewPrefixedUUID("fextract")
	db.DB.Create(&domain.FileAsset{
		ID:               "file-1",
		Name:             "launch-plan.docx",
		StoragePath:      "launch-plan.docx",
		UploaderID:       "user-1",
		ContentType:      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		ExtractionStatus: "ready",
		ContentSummary:   "Launch plan summary",
		LastIndexedAt:    &now,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	db.DB.Create(&domain.FileExtraction{
		ID:             extractionID,
		FileID:         "file-1",
		Status:         "ready",
		Extractor:      "docx",
		ContentText:    "Launch checklist and rollout plan",
		ContentSummary: "Launch plan summary",
		CreatedAt:      now,
		UpdatedAt:      now,
		CompletedAt:    &now,
	})
	db.DB.Create(&domain.FileExtractionChunk{
		ExtractionID:  extractionID,
		FileID:        "file-1",
		ChunkIndex:    0,
		Text:          "Launch checklist and rollout plan",
		TokenEstimate: 5,
		LocatorType:   "document",
		LocatorValue:  "root",
		Heading:       "Launch",
		CreatedAt:     now,
	})

	router := gin.New()
	router.GET("/api/v1/files/:id/citations", GetFileCitations)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/file-1/citations", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get file citations, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestKnowledgeEvidenceModelsPersist(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	link := domain.KnowledgeEvidenceLink{
		ID:            ids.NewPrefixedUUID("evidence"),
		WorkspaceID:   "ws-1",
		EvidenceKind:  "file_chunk",
		EvidenceRefID: "chunk-1",
		SourceKind:    "file",
		SourceRef:     "file-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := db.DB.Create(&link).Error; err != nil {
		t.Fatalf("failed to create evidence link: %v", err)
	}

	ref := domain.KnowledgeEvidenceEntityRef{
		ID:         ids.NewPrefixedUUID("evidence-ref"),
		EvidenceID: link.ID,
		EntityID:   "entity-1",
		CreatedAt:  now,
	}
	if err := db.DB.Create(&ref).Error; err != nil {
		t.Fatalf("failed to create evidence entity ref: %v", err)
	}
}

func TestCitationLookupReturnsMixedEvidenceKinds(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})

	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-1",
		UserID:    "user-1",
		Content:   "Launch checklist and rollout plan",
		CreatedAt: now,
	})
	db.DB.Create(&domain.Artifact{
		ID:        "artifact-1",
		ChannelID: "ch-1",
		Title:     "Launch Canvas",
		Version:   1,
		Type:      "document",
		Status:    "draft",
		Content:   "Launch architecture and rollout plan",
		CreatedBy: "user-1",
		UpdatedBy: "user-1",
		CreatedAt: now,
		UpdatedAt: now,
	})
	extractionID := ids.NewPrefixedUUID("fextract")
	db.DB.Create(&domain.FileAsset{
		ID:               "file-1",
		ChannelID:        "ch-1",
		Name:             "launch-plan.docx",
		StoragePath:      "launch-plan.docx",
		UploaderID:       "user-1",
		ContentType:      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		ExtractionStatus: "ready",
		ContentSummary:   "Launch plan summary",
		LastIndexedAt:    &now,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	db.DB.Create(&domain.FileExtraction{
		ID:             extractionID,
		FileID:         "file-1",
		Status:         "ready",
		Extractor:      "docx",
		ContentText:    "Launch checklist and rollout plan",
		ContentSummary: "Launch plan summary",
		CreatedAt:      now,
		UpdatedAt:      now,
		CompletedAt:    &now,
	})
	db.DB.Create(&domain.FileExtractionChunk{
		ExtractionID:  extractionID,
		FileID:        "file-1",
		ChunkIndex:    0,
		Text:          "Launch checklist and rollout plan",
		TokenEstimate: 5,
		LocatorType:   "document",
		LocatorValue:  "root",
		Heading:       "Launch",
		CreatedAt:     now,
	})

	router := gin.New()
	router.GET("/api/v1/citations/lookup", LookupCitations)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/citations/lookup?q=launch", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on citation lookup, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCitationLookupIncludesOptionalEntityFields(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Message{
		ID:        "msg-1",
		ChannelID: "ch-1",
		UserID:    "user-1",
		Content:   "Launch checklist and rollout plan",
		CreatedAt: now,
	})

	router := gin.New()
	router.GET("/api/v1/citations/lookup", LookupCitations)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/citations/lookup?q=launch&entity_id=entity-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on citation lookup with entity filter, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestKnowledgeEntityModelsPersist(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	entity := domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "Q2 launch", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now}
	if err := db.DB.Create(&entity).Error; err != nil {
		t.Fatalf("failed to create entity: %v", err)
	}
	ref := domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: entity.ID, RefKind: "file", RefID: "file-1", Role: "evidence", CreatedAt: now}
	if err := db.DB.Create(&ref).Error; err != nil {
		t.Fatalf("failed to create entity ref: %v", err)
	}
	link := domain.KnowledgeEntityLink{ID: "klink-1", WorkspaceID: "ws-1", FromEntityID: entity.ID, ToEntityID: "entity-2", Relation: "depends_on", CreatedAt: now}
	if err := db.DB.Create(&link).Error; err != nil {
		t.Fatalf("failed to create entity link: %v", err)
	}
	event := domain.KnowledgeEvent{ID: "kevent-1", WorkspaceID: "ws-1", EntityID: entity.ID, EventType: "created", Title: "Created entity", SourceKind: "system", OccurredAt: now, CreatedAt: now}
	if err := db.DB.Create(&event).Error; err != nil {
		t.Fatalf("failed to create entity event: %v", err)
	}
}

func TestKnowledgeEntityCRUDEndpoints(t *testing.T) {
	setupTestDB(t)
	db.DB.Create(&domain.Workspace{ID: "ws-1", Name: "Relay"})

	router := gin.New()
	router.GET("/api/v1/knowledge/entities", ListKnowledgeEntities)
	router.POST("/api/v1/knowledge/entities", CreateKnowledgeEntity)
	router.GET("/api/v1/knowledge/entities/:id", GetKnowledgeEntity)
	router.PATCH("/api/v1/knowledge/entities/:id", UpdateKnowledgeEntity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities", strings.NewReader(`{"workspace_id":"ws-1","kind":"project","title":"Launch Program","summary":"Q2 launch"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create entity, got %d body=%s", rec.Code, rec.Body.String())
	}

	var createPayload struct {
		Entity domain.KnowledgeEntity `json:"entity"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("decode create entity: %v", err)
	}
	if createPayload.Entity.ID == "" || createPayload.Entity.Title != "Launch Program" {
		t.Fatalf("unexpected create payload: %#v", createPayload)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities", strings.NewReader(`{"kind":"file","title":"Principles of Game Design","tags":["game","design","md"]}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 when creating entity without explicit workspace, got %d body=%s", rec.Code, rec.Body.String())
	}
	var filePayload struct {
		Entity domain.KnowledgeEntity `json:"entity"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filePayload); err != nil {
		t.Fatalf("decode file entity create: %v", err)
	}
	if filePayload.Entity.WorkspaceID != "ws-1" || filePayload.Entity.Kind != "file" || !strings.Contains(filePayload.Entity.Metadata, `"tags":["game","design","md"]`) {
		t.Fatalf("expected default workspace and tags metadata, got %#v", filePayload.Entity)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/knowledge/entities/"+createPayload.Entity.ID, strings.NewReader(`{"status":"paused","summary":"Updated launch summary"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on update entity, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/entities", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on list entities, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestKnowledgeEntityRefsAndTimelineEndpoints(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})

	router := gin.New()
	router.GET("/api/v1/knowledge/entities/:id/refs", ListKnowledgeEntityRefs)
	router.POST("/api/v1/knowledge/entities/:id/refs", AddKnowledgeEntityRef)
	router.GET("/api/v1/knowledge/entities/:id/timeline", ListKnowledgeEntityTimeline)
	router.POST("/api/v1/knowledge/entities/:id/events", AddKnowledgeEntityEvent)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/refs", strings.NewReader(`{"ref_kind":"file","ref_id":"file-1","role":"evidence"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on add ref, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/entities/entity-1/timeline", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "file_linked") {
		t.Fatalf("expected timeline with file_linked event, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestKnowledgeEntityGraphEndpoint(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "service", Title: "Billing Service", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})

	router := gin.New()
	router.POST("/api/v1/knowledge/entities/:id/refs", AddKnowledgeEntityRef)
	router.POST("/api/v1/knowledge/links", AddKnowledgeEntityLink)
	router.GET("/api/v1/knowledge/entities/:id/graph", GetKnowledgeEntityGraph)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/refs", strings.NewReader(`{"ref_kind":"message","ref_id":"msg-1","role":"discussion"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on add graph ref, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/links", strings.NewReader(`{"from_entity_id":"entity-1","to_entity_id":"entity-2","relation":"depends_on","weight":2.5}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on add link, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/entities/entity-1/graph", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "depends_on") {
		t.Fatalf("expected graph with depends_on edge, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Graph struct {
			Nodes []knowledgeGraphNodePayload `json:"nodes"`
			Edges []knowledgeGraphEdgePayload `json:"edges"`
		} `json:"graph"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode graph payload: %v", err)
	}
	if !graphHasRefNode(payload.Graph.Nodes, "message", "msg-1", "discussion") {
		t.Fatalf("expected graph ref node metadata, got %#v", payload.Graph.Nodes)
	}
	if !graphHasWeightedEdge(payload.Graph.Edges, "depends_on", 2.5, "out") {
		t.Fatalf("expected weighted directional graph edge, got %#v", payload.Graph.Edges)
	}
}

func TestGetChannelKnowledgeContextReturnsRecentRefs(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program is ready", CreatedAt: now.Add(-2 * time.Minute)})
	db.DB.Create(&domain.FileAsset{ID: "file-1", ChannelID: "ch-1", UploaderID: "user-1", Name: "launch-plan.md", StoragePath: "launch-plan.md", ContentType: "text/markdown", CreatedAt: now.Add(-time.Minute), UpdatedAt: now.Add(-time.Minute)})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-msg", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-30 * time.Second)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-file", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "file", RefID: "file-1", Role: "evidence", CreatedAt: now})

	router := gin.New()
	router.GET("/api/v1/channels/:id/knowledge", GetChannelKnowledgeContext)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/ch-1/knowledge", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on channel knowledge context, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Context struct {
			ChannelID string `json:"channel_id"`
			Refs      []struct {
				RefID       string `json:"ref_id"`
				RefKind     string `json:"ref_kind"`
				EntityTitle string `json:"entity_title"`
				SourceTitle string `json:"source_title"`
			} `json:"refs"`
		} `json:"context"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode channel knowledge context: %v", err)
	}
	if payload.Context.ChannelID != "ch-1" || len(payload.Context.Refs) != 2 {
		t.Fatalf("expected two channel knowledge refs, got %#v", payload.Context)
	}
	if payload.Context.Refs[0].RefKind != "file" || payload.Context.Refs[0].EntityTitle != "Launch Program" || payload.Context.Refs[0].SourceTitle != "launch-plan.md" {
		t.Fatalf("expected newest file ref with hydrated entity/source title first, got %#v", payload.Context.Refs)
	}
}

func TestGetChannelKnowledgeSummaryReturnsTopEntitiesAndTrend(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff", CreatedAt: now.Add(-48 * time.Hour)})
	db.DB.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-1", UserID: "user-1", Content: "Launch checklist ready", CreatedAt: now.Add(-24 * time.Hour)})
	db.DB.Create(&domain.FileAsset{ID: "file-1", ChannelID: "ch-1", UploaderID: "user-1", Name: "launch-plan.md", StoragePath: "launch-plan.md", ContentType: "text/markdown", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "doc", Title: "Launch Checklist", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-msg", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-48 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-file", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "file", RefID: "file-1", Role: "evidence", CreatedAt: now})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-msg-2", WorkspaceID: "ws-1", EntityID: "entity-2", RefKind: "message", RefID: "msg-2", Role: "discussion", CreatedAt: now.Add(-24 * time.Hour)})

	router := gin.New()
	router.GET("/api/v1/channels/:id/knowledge/summary", GetChannelKnowledgeSummary)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/ch-1/knowledge/summary?limit=2&days=7", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on channel knowledge summary, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Summary struct {
			ChannelID      string `json:"channel_id"`
			WindowDays     int    `json:"window_days"`
			TotalRefs      int    `json:"total_refs"`
			RecentRefCount int    `json:"recent_ref_count"`
			Velocity       struct {
				RecentWindowDays int  `json:"recent_window_days"`
				PreviousRefCount int  `json:"previous_ref_count"`
				RecentRefCount   int  `json:"recent_ref_count"`
				Delta            int  `json:"delta"`
				IsSpiking        bool `json:"is_spiking"`
			} `json:"velocity"`
			TopEntities []struct {
				EntityID        string `json:"entity_id"`
				EntityTitle     string `json:"entity_title"`
				RefCount        int    `json:"ref_count"`
				MessageRefCount int    `json:"message_ref_count"`
				FileRefCount    int    `json:"file_ref_count"`
				Trend           []struct {
					Date  string `json:"date"`
					Count int    `json:"count"`
				} `json:"trend"`
			} `json:"top_entities"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode channel knowledge summary: %v", err)
	}
	if payload.Summary.ChannelID != "ch-1" || payload.Summary.TotalRefs != 3 || payload.Summary.WindowDays != 7 {
		t.Fatalf("unexpected summary payload: %#v", payload.Summary)
	}
	if payload.Summary.Velocity.RecentWindowDays == 0 || payload.Summary.Velocity.RecentRefCount == 0 || payload.Summary.Velocity.Delta == 0 {
		t.Fatalf("expected summary velocity payload, got %#v", payload.Summary.Velocity)
	}
	if len(payload.Summary.TopEntities) != 2 || payload.Summary.TopEntities[0].EntityID != "entity-1" {
		t.Fatalf("expected ranked entities in summary payload, got %#v", payload.Summary.TopEntities)
	}
	if payload.Summary.TopEntities[0].MessageRefCount != 1 || payload.Summary.TopEntities[0].FileRefCount != 1 || len(payload.Summary.TopEntities[0].Trend) != 7 {
		t.Fatalf("expected trend/count breakdown for top entity, got %#v", payload.Summary.TopEntities[0])
	}
}

func TestSearchMessagesByEntityReturnsKnowledgeBackedAndExplicitMatches(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "<p>@Launch Program update</p>", CreatedAt: now.Add(-time.Minute)})
	db.DB.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program doc synced", CreatedAt: now})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-msg-2", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-2", Role: "discussion", CreatedAt: now})

	router := gin.New()
	router.GET("/api/v1/search/messages/by-entity", SearchMessagesByEntity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/messages/by-entity?entity_id=entity-1&limit=10", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on entity message search, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		EntityID string `json:"entity_id"`
		Messages []struct {
			ID           string   `json:"id"`
			EntityTitle  string   `json:"entity_title"`
			MatchSources []string `json:"match_sources"`
			Metadata     string   `json:"metadata"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode entity message search payload: %v", err)
	}
	if payload.EntityID != "entity-1" || len(payload.Messages) != 2 {
		t.Fatalf("expected two entity message matches, got %#v", payload)
	}
	if payload.Messages[0].ID != "msg-2" || payload.Messages[0].EntityTitle != "Launch Program" {
		t.Fatalf("expected newest result first with hydrated entity title, got %#v", payload.Messages)
	}
	if len(payload.Messages[0].MatchSources) == 0 || len(payload.Messages[1].MatchSources) == 0 {
		t.Fatalf("expected match_sources for every result, got %#v", payload.Messages)
	}
	if !strings.Contains(payload.Messages[1].Metadata, "\"entity_mentions\"") {
		t.Fatalf("expected explicit mention metadata to be refreshed, got %s", payload.Messages[1].Metadata)
	}
}

func TestGetKnowledgeEntityHoverReturnsLiveActivitySummary(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.FileAsset{ID: "file-1", ChannelID: "ch-1", UploaderID: "user-1", Name: "launch-plan.md", StoragePath: "launch-plan.md", ContentType: "text/markdown", CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Hour)})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "Main initiative", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-msg", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-file", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "file", RefID: "file-1", Role: "evidence", CreatedAt: now.Add(-time.Hour)})

	router := gin.New()
	router.GET("/api/v1/knowledge/entities/:id/hover", GetKnowledgeEntityHover)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/entities/entity-1/hover?channel_id=ch-1&days=7", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on entity hover summary, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Entity struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"entity"`
		Hover struct {
			EntityID        string `json:"entity_id"`
			RefCount        int    `json:"ref_count"`
			ChannelRefCount int    `json:"channel_ref_count"`
			MessageRefCount int    `json:"message_ref_count"`
			FileRefCount    int    `json:"file_ref_count"`
			RecentRefCount  int    `json:"recent_ref_count"`
			RelatedChannels []struct {
				ChannelID string `json:"channel_id"`
				Name      string `json:"name"`
				RefCount  int    `json:"ref_count"`
			} `json:"related_channels"`
		} `json:"hover"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode entity hover payload: %v", err)
	}
	if payload.Entity.ID != "entity-1" || payload.Hover.RefCount != 2 || payload.Hover.ChannelRefCount != 2 {
		t.Fatalf("unexpected hover envelope: %#v", payload)
	}
	if payload.Hover.MessageRefCount != 1 || payload.Hover.FileRefCount != 1 || payload.Hover.RecentRefCount != 2 {
		t.Fatalf("expected hover to split message/file and recent counts, got %#v", payload.Hover)
	}
	if len(payload.Hover.RelatedChannels) != 1 || payload.Hover.RelatedChannels[0].ChannelID != "ch-1" {
		t.Fatalf("expected channel activity breakdown, got %#v", payload.Hover.RelatedChannels)
	}
}

func TestGetChannelKnowledgeDigestReturnsTopMovements(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff", CreatedAt: now.Add(-48 * time.Hour)})
	db.DB.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-1", UserID: "user-1", Content: "Billing Service dependency", CreatedAt: now.Add(-12 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "service", Title: "Billing Service", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-48 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-2", WorkspaceID: "ws-1", EntityID: "entity-2", RefKind: "message", RefID: "msg-2", Role: "discussion", CreatedAt: now.Add(-12 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-3", WorkspaceID: "ws-1", EntityID: "entity-2", RefKind: "message", RefID: "msg-2", Role: "discussion", CreatedAt: now.Add(-6 * time.Hour)})

	router := gin.New()
	router.GET("/api/v1/channels/:id/knowledge/digest", GetChannelKnowledgeDigest)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/ch-1/knowledge/digest?window=weekly&limit=3", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on knowledge digest, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Digest struct {
			ChannelID    string `json:"channel_id"`
			Window       string `json:"window"`
			WindowDays   int    `json:"window_days"`
			Headline     string `json:"headline"`
			TotalRefs    int    `json:"total_refs"`
			TopMovements []struct {
				EntityID       string `json:"entity_id"`
				EntityTitle    string `json:"entity_title"`
				RecentRefCount int    `json:"recent_ref_count"`
				Delta          int    `json:"delta"`
			} `json:"top_movements"`
		} `json:"digest"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode knowledge digest payload: %v", err)
	}
	if payload.Digest.ChannelID != "ch-1" || payload.Digest.Window != "weekly" || payload.Digest.WindowDays != 7 {
		t.Fatalf("unexpected digest envelope: %#v", payload.Digest)
	}
	if payload.Digest.TotalRefs != 3 || len(payload.Digest.TopMovements) == 0 {
		t.Fatalf("expected digest to contain movements, got %#v", payload.Digest)
	}
	if payload.Digest.TopMovements[0].EntityID != "entity-2" || payload.Digest.TopMovements[0].RecentRefCount == 0 {
		t.Fatalf("expected recent hottest entity first, got %#v", payload.Digest.TopMovements)
	}
	if strings.TrimSpace(payload.Digest.Headline) == "" {
		t.Fatalf("expected digest headline, got %#v", payload.Digest)
	}
}

func TestPublishChannelKnowledgeDigestCreatesPinnedMessage(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff", CreatedAt: now.Add(-12 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-2 * time.Hour)})

	router := gin.New()
	router.POST("/api/v1/channels/:id/knowledge/digest/publish", PublishChannelKnowledgeDigest)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels/ch-1/knowledge/digest/publish", strings.NewReader(`{"window":"daily","limit":3,"pin":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on digest publish, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Message struct {
			ID        string `json:"id"`
			ChannelID string `json:"channel_id"`
			IsPinned  bool   `json:"is_pinned"`
			Metadata  string `json:"metadata"`
		} `json:"message"`
		Digest struct {
			Window string `json:"window"`
		} `json:"digest"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode publish digest payload: %v", err)
	}
	if payload.Message.ChannelID != "ch-1" || !payload.Message.IsPinned || payload.Digest.Window != "daily" {
		t.Fatalf("expected pinned digest message in channel, got %#v", payload)
	}
	if !strings.Contains(payload.Message.Metadata, "\"knowledge_digest\"") {
		t.Fatalf("expected published message metadata to retain knowledge digest payload, got %s", payload.Message.Metadata)
	}
}

func TestDigestScheduleEndpointsSupportUpsertGetAndDelete(t *testing.T) {
	setupTestDB(t)
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})

	router := gin.New()
	router.GET("/api/v1/channels/:id/knowledge/digest/schedule", GetChannelKnowledgeDigestSchedule)
	router.PUT("/api/v1/channels/:id/knowledge/digest/schedule", PutChannelKnowledgeDigestSchedule)
	router.DELETE("/api/v1/channels/:id/knowledge/digest/schedule", DeleteChannelKnowledgeDigestSchedule)

	putReq := httptest.NewRequest(http.MethodPut, "/api/v1/channels/ch-1/knowledge/digest/schedule", strings.NewReader(`{"window":"weekly","timezone":"Asia/Shanghai","day_of_week":0,"hour":9,"minute":0,"limit":5,"pin":true,"is_enabled":true}`))
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()
	router.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on put digest schedule, got %d body=%s", putRec.Code, putRec.Body.String())
	}

	getRec := httptest.NewRecorder()
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/channels/ch-1/knowledge/digest/schedule", nil)
	router.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get digest schedule, got %d body=%s", getRec.Code, getRec.Body.String())
	}

	var payload struct {
		Schedule struct {
			ChannelID string `json:"channel_id"`
			Window    string `json:"window"`
			Timezone  string `json:"timezone"`
			Pin       bool   `json:"pin"`
		} `json:"schedule"`
	}
	if err := json.Unmarshal(getRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode digest schedule payload: %v", err)
	}
	if payload.Schedule.ChannelID != "ch-1" || payload.Schedule.Window != "weekly" || payload.Schedule.Timezone != "Asia/Shanghai" || !payload.Schedule.Pin {
		t.Fatalf("unexpected schedule payload: %#v", payload.Schedule)
	}

	delRec := httptest.NewRecorder()
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/channels/ch-1/knowledge/digest/schedule", nil)
	router.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete digest schedule, got %d body=%s", delRec.Code, delRec.Body.String())
	}
}

func TestPreviewChannelKnowledgeDigestScheduleReturnsUpcomingRuns(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "Main launch initiative", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program is accelerating", CreatedAt: now})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now})

	router := gin.New()
	router.POST("/api/v1/channels/:id/knowledge/digest/preview-schedule", PreviewChannelKnowledgeDigestSchedule)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels/ch-1/knowledge/digest/preview-schedule", strings.NewReader(`{"window":"weekly","timezone":"Asia/Shanghai","day_of_week":1,"hour":9,"minute":30,"limit":3,"pin":true,"is_enabled":true,"count":3}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on digest schedule preview, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Preview struct {
			UpcomingRuns []struct {
				RunAt string `json:"run_at"`
			} `json:"upcoming_runs"`
			Digest struct {
				ChannelID string `json:"channel_id"`
				Window    string `json:"window"`
			} `json:"digest"`
		} `json:"preview"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode digest preview payload: %v", err)
	}
	if len(payload.Preview.UpcomingRuns) != 3 || payload.Preview.Digest.ChannelID != "ch-1" || payload.Preview.Digest.Window != "weekly" {
		t.Fatalf("unexpected preview payload: %#v", payload.Preview)
	}
}

func TestGetKnowledgeInboxReturnsDigestMessages(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public", IsStarred: true})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1", Role: "member"})
	db.DB.Create(&domain.Message{
		ID:        "msg-digest-1",
		ChannelID: "ch-1",
		UserID:    "user-1",
		Content:   "Knowledge digest",
		IsPinned:  true,
		CreatedAt: now,
		Metadata:  `{"knowledge_digest":{"channel_id":"ch-1","window":"weekly","window_days":7,"generated_at":"` + now.Format(time.RFC3339Nano) + `","total_refs":3,"recent_ref_count":2,"headline":"Digest","summary":"Summary","top_movements":[]}}`,
	})

	router := gin.New()
	router.GET("/api/v1/knowledge/inbox", GetKnowledgeInbox)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/inbox?scope=all&limit=10", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on knowledge inbox, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []struct {
			ID      string `json:"id"`
			IsRead  bool   `json:"is_read"`
			Channel struct {
				ID string `json:"id"`
			} `json:"channel"`
			Message struct {
				ID string `json:"id"`
			} `json:"message"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode knowledge inbox payload: %v", err)
	}
	if len(payload.Items) != 1 || payload.Items[0].Channel.ID != "ch-1" || payload.Items[0].Message.ID != "msg-digest-1" {
		t.Fatalf("unexpected knowledge inbox payload: %#v", payload.Items)
	}
}

func TestGetKnowledgeInboxItemReturnsDigestContext(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public", IsStarred: true})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1", Role: "member"})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "Main launch initiative", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.Message{ID: "msg-source-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program is now blocked on QA sign-off", CreatedAt: now.Add(-10 * time.Minute)})
	db.DB.Create(&domain.Message{ID: "msg-source-2", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program rollout window moved to Friday", CreatedAt: now.Add(-5 * time.Minute)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-source-1", Role: "discussion", CreatedAt: now.Add(-10 * time.Minute)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-2", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-source-2", Role: "decision", CreatedAt: now.Add(-5 * time.Minute)})
	db.DB.Create(&domain.Message{
		ID:        "msg-digest-1",
		ChannelID: "ch-1",
		UserID:    "user-1",
		Content:   "Knowledge digest",
		IsPinned:  true,
		CreatedAt: now,
		Metadata:  `{"knowledge_digest":{"channel_id":"ch-1","window":"weekly","window_days":7,"generated_at":"` + now.Format(time.RFC3339Nano) + `","total_refs":3,"recent_ref_count":2,"headline":"Digest","summary":"Summary","top_movements":[{"entity_id":"entity-1","entity_title":"Launch Program","entity_kind":"project","ref_count":2,"recent_ref_count":2,"previous_ref_count":0,"delta":2}]}}`,
	})

	router := gin.New()
	router.GET("/api/v1/knowledge/inbox/:id", GetKnowledgeInboxItem)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/inbox/knowledge-digest-msg-digest-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on knowledge inbox item detail, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Item struct {
			ID      string `json:"id"`
			Message struct {
				ID string `json:"id"`
			} `json:"message"`
			EntityContexts []struct {
				EntityID string `json:"entity_id"`
				Messages []struct {
					ID string `json:"id"`
				} `json:"messages"`
			} `json:"entity_contexts"`
		} `json:"item"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode knowledge inbox item payload: %v", err)
	}
	if payload.Item.ID != "knowledge-digest-msg-digest-1" || payload.Item.Message.ID != "msg-digest-1" || len(payload.Item.EntityContexts) != 1 || len(payload.Item.EntityContexts[0].Messages) < 2 {
		t.Fatalf("unexpected knowledge inbox detail payload: %#v", payload.Item)
	}
}

func TestGetHomeIncludesKnowledgeDigestSummary(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com", Status: "online"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public", IsStarred: true})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1", Role: "member"})
	db.DB.Create(&domain.Message{
		ID:        "msg-digest-1",
		ChannelID: "ch-1",
		UserID:    "user-1",
		Content:   "Knowledge digest",
		IsPinned:  true,
		CreatedAt: now,
		Metadata:  `{"knowledge_digest":{"channel_id":"ch-1","window":"weekly","window_days":7,"generated_at":"` + now.Format(time.RFC3339Nano) + `","total_refs":3,"recent_ref_count":2,"headline":"Digest","summary":"Summary","top_movements":[]}}`,
	})

	router := gin.New()
	router.GET("/api/v1/home", GetHome)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/home", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on home, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Home struct {
			KnowledgeInboxCount    int   `json:"knowledge_inbox_count"`
			RecentKnowledgeDigests []any `json:"recent_knowledge_digests"`
		} `json:"home"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode home knowledge payload: %v", err)
	}
	if payload.Home.KnowledgeInboxCount != 1 || len(payload.Home.RecentKnowledgeDigests) != 1 {
		t.Fatalf("expected home knowledge digest aggregation, got %#v", payload.Home)
	}
}

func TestSuggestKnowledgeEntitiesReturnsScopedAutocompleteResults(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Channel{ID: "ch-2", WorkspaceID: "ws-2", Name: "external", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff", CreatedAt: now})
	db.DB.Create(&domain.FileAsset{ID: "file-1", ChannelID: "ch-1", UploaderID: "user-1", Name: "launch-plan.md", StoragePath: "launch-plan.md", ContentType: "text/markdown", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "Main launch initiative", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "doc", Title: "Launch Checklist", Summary: "Pre-flight list", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-3", WorkspaceID: "ws-2", Kind: "project", Title: "Launch External", Summary: "Other workspace", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-msg", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-file", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "file", RefID: "file-1", Role: "evidence", CreatedAt: now})

	router := gin.New()
	router.GET("/api/v1/knowledge/entities/suggest", SuggestKnowledgeEntities)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/entities/suggest?q=launch&channel_id=ch-1&limit=5", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on knowledge entity suggestions, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Query       string `json:"query"`
		Suggestions []struct {
			ID              string `json:"id"`
			Title           string `json:"title"`
			ChannelRefCount int    `json:"channel_ref_count"`
		} `json:"suggestions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode knowledge suggestions payload: %v", err)
	}
	if payload.Query != "launch" || len(payload.Suggestions) != 2 {
		t.Fatalf("expected two scoped suggestions, got %#v", payload)
	}
	if payload.Suggestions[0].ID != "entity-1" || payload.Suggestions[0].ChannelRefCount != 2 {
		t.Fatalf("expected entity-1 ranked first by channel refs, got %#v", payload.Suggestions)
	}
	for _, suggestion := range payload.Suggestions {
		if suggestion.ID == "entity-3" {
			t.Fatalf("expected cross-workspace entity to be excluded, got %#v", payload.Suggestions)
		}
	}
}

func TestKnowledgeEntityFollowEndpoints(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})

	router := gin.New()
	router.GET("/api/v1/users/me/knowledge/followed", GetMyFollowedKnowledgeEntities)
	router.POST("/api/v1/knowledge/entities/:id/follow", FollowKnowledgeEntity)
	router.PATCH("/api/v1/users/me/knowledge/followed/:id", PatchMyKnowledgeFollow)
	router.DELETE("/api/v1/knowledge/entities/:id/follow", UnfollowKnowledgeEntity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/follow", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 follow, got %d body=%s", rec.Code, rec.Body.String())
	}
	var followPayload struct {
		Follow      domain.KnowledgeEntityFollow `json:"follow"`
		IsFollowing bool                         `json:"is_following"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &followPayload); err != nil {
		t.Fatalf("failed to decode follow payload: %v", err)
	}
	if followPayload.Follow.EntityID != "entity-1" || !followPayload.IsFollowing {
		t.Fatalf("unexpected follow payload: %#v", followPayload)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/me/knowledge/followed", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 followed list, got %d body=%s", rec.Code, rec.Body.String())
	}
	var listPayload struct {
		Items []knowledge.FollowedEntity `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode followed list: %v", err)
	}
	if len(listPayload.Items) != 1 || listPayload.Items[0].Entity.ID != "entity-1" {
		t.Fatalf("unexpected followed list: %#v", listPayload.Items)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/knowledge/followed/entity-1", strings.NewReader(`{"notification_level":"digest_only"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 follow patch, got %d body=%s", rec.Code, rec.Body.String())
	}
	if err := db.DB.First(&followPayload.Follow, "entity_id = ? AND user_id = ?", "entity-1", "user-1").Error; err != nil {
		t.Fatalf("reload follow after patch: %v", err)
	}
	if followPayload.Follow.NotificationLevel != "digest_only" {
		t.Fatalf("expected notification level to persist, got %#v", followPayload.Follow)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/knowledge/entities/entity-1/follow", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 unfollow, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPhase59KnowledgeOpsEndpoints(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay", KnowledgeSpikeThreshold: 3, KnowledgeSpikeCooldownMins: 360})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "concept", Title: "Agent Mesh", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Channel{ID: "ch-2", WorkspaceID: "ws-1", Name: "ai-lab", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff", CreatedAt: now.Add(-40 * time.Hour)})
	db.DB.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program sprint", CreatedAt: now.Add(-20 * time.Hour)})
	db.DB.Create(&domain.Message{ID: "msg-3", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program QA", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.Message{ID: "msg-4", ChannelID: "ch-2", UserID: "user-1", Content: "Agent Mesh design", CreatedAt: now.Add(-3 * time.Hour)})
	db.DB.Create(&domain.Message{ID: "msg-5", ChannelID: "ch-2", UserID: "user-1", Content: "Agent Mesh rollout", CreatedAt: now.Add(-90 * time.Minute)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-40 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-2", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-2", Role: "discussion", CreatedAt: now.Add(-20 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-3", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-3", Role: "decision", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-4", WorkspaceID: "ws-1", EntityID: "entity-2", RefKind: "message", RefID: "msg-4", Role: "discussion", CreatedAt: now.Add(-3 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-5", WorkspaceID: "ws-1", EntityID: "entity-2", RefKind: "message", RefID: "msg-5", Role: "decision", CreatedAt: now.Add(-90 * time.Minute)})

	router := gin.New()
	router.POST("/api/v1/knowledge/entities/:id/follow", FollowKnowledgeEntity)
	router.PATCH("/api/v1/users/me/knowledge/followed/bulk", PatchMyKnowledgeFollowsBulk)
	router.GET("/api/v1/workspace/settings", GetWorkspaceSettings)
	router.PATCH("/api/v1/workspace/settings", PatchWorkspaceSettings)
	router.GET("/api/v1/knowledge/entities/:id/activity", GetKnowledgeEntityActivity)
	router.GET("/api/v1/knowledge/trending", GetKnowledgeTrending)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/follow", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 follow entity-1, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-2/follow", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 follow entity-2, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/knowledge/followed/bulk", strings.NewReader(`{"entity_ids":["entity-1","entity-2"],"notification_level":"silent"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 bulk follow patch, got %d body=%s", rec.Code, rec.Body.String())
	}
	var bulkPayload struct {
		Items []domain.KnowledgeEntityFollow `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &bulkPayload); err != nil {
		t.Fatalf("decode bulk follow payload: %v", err)
	}
	if len(bulkPayload.Items) != 2 || bulkPayload.Items[0].NotificationLevel != "silent" {
		t.Fatalf("unexpected bulk follow payload: %#v", bulkPayload.Items)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/workspace/settings?workspace_id=ws-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 get workspace settings, got %d body=%s", rec.Code, rec.Body.String())
	}
	var settingsPayload struct {
		Settings struct {
			WorkspaceID          string `json:"workspace_id"`
			SpikeThreshold       int    `json:"spike_threshold"`
			SpikeCooldownMinutes int    `json:"spike_cooldown_minutes"`
		} `json:"settings"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &settingsPayload); err != nil {
		t.Fatalf("decode workspace settings: %v", err)
	}
	if settingsPayload.Settings.SpikeThreshold != 3 || settingsPayload.Settings.SpikeCooldownMinutes != 360 {
		t.Fatalf("unexpected workspace settings payload: %#v", settingsPayload.Settings)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/workspace/settings", strings.NewReader(`{"workspace_id":"ws-1","spike_threshold":5,"spike_cooldown_minutes":90}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 patch workspace settings, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/entities/entity-1/activity?days=7", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 get entity activity, got %d body=%s", rec.Code, rec.Body.String())
	}
	var activityPayload struct {
		Activity struct {
			EntityID string `json:"entity_id"`
			Buckets  []struct {
				Date     string `json:"date"`
				RefCount int    `json:"ref_count"`
			} `json:"buckets"`
		} `json:"activity"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &activityPayload); err != nil {
		t.Fatalf("decode entity activity payload: %v", err)
	}
	if activityPayload.Activity.EntityID != "entity-1" || len(activityPayload.Activity.Buckets) != 7 {
		t.Fatalf("unexpected entity activity payload: %#v", activityPayload.Activity)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/trending?workspace_id=ws-1&days=7&limit=5", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 get knowledge trending, got %d body=%s", rec.Code, rec.Body.String())
	}
	var trendingPayload struct {
		Items []struct {
			Entity struct {
				ID string `json:"id"`
			} `json:"entity"`
			VelocityDelta int `json:"velocity_delta"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &trendingPayload); err != nil {
		t.Fatalf("decode trending payload: %v", err)
	}
	if len(trendingPayload.Items) < 2 || trendingPayload.Items[0].Entity.ID != "entity-1" {
		t.Fatalf("unexpected trending payload: %#v", trendingPayload.Items)
	}
}

func TestPhase60KnowledgeShareStatsAndTrendingRealtime(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay", KnowledgeSpikeThreshold: 3, KnowledgeSpikeCooldownMins: 360})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "concept", Title: "Agent Mesh", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program QA", CreatedAt: now.Add(-90 * time.Minute)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-2", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-2", Role: "decision", CreatedAt: now.Add(-90 * time.Minute)})

	router := gin.New()
	router.POST("/api/v1/knowledge/entities/:id/follow", FollowKnowledgeEntity)
	router.GET("/api/v1/users/me/knowledge/followed/stats", GetMyFollowedKnowledgeStats)
	router.POST("/api/v1/knowledge/entities/:id/share", ShareKnowledgeEntity)
	router.POST("/api/v1/knowledge/entities/:id/refs", AddKnowledgeEntityRef)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/follow", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 follow entity-1, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-2/follow", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 follow entity-2, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.followed.stats.changed")
	assertRealtimeEventType(t, client, "knowledge.followed.stats.changed")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/me/knowledge/followed/stats", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 followed stats, got %d body=%s", rec.Code, rec.Body.String())
	}
	var statsPayload struct {
		Stats struct {
			TotalCount   int `json:"total_count"`
			SpikingCount int `json:"spiking_count"`
			MutedCount   int `json:"muted_count"`
		} `json:"stats"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &statsPayload); err != nil {
		t.Fatalf("decode followed stats payload: %v", err)
	}
	if statsPayload.Stats.TotalCount != 2 {
		t.Fatalf("unexpected followed stats payload: %#v", statsPayload.Stats)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/share", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 entity share, got %d body=%s", rec.Code, rec.Body.String())
	}
	var sharePayload struct {
		Share struct {
			EntityID string `json:"entity_id"`
			URL      string `json:"url"`
			ShortURL string `json:"short_url"`
		} `json:"share"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &sharePayload); err != nil {
		t.Fatalf("decode share payload: %v", err)
	}
	if sharePayload.Share.EntityID != "entity-1" || !strings.Contains(sharePayload.Share.URL, "/workspace/knowledge/entity-1") {
		t.Fatalf("unexpected share payload: %#v", sharePayload.Share)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/refs", strings.NewReader(`{"ref_kind":"message","ref_id":"msg-3","role":"decision"}`))
	req.Header.Set("Content-Type", "application/json")
	db.DB.Create(&domain.Message{ID: "msg-3", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program launch", CreatedAt: now.Add(-30 * time.Minute)})
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 entity ref create, got %d body=%s", rec.Code, rec.Body.String())
	}

	assertRealtimeEventType(t, client, "knowledge.entity.ref.created")
	assertRealtimeEventType(t, client, "knowledge.entity.activity.spiked")
	raw, err := client.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("failed to receive trending event: %v", err)
	}
	var event realtime.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("decode trending event: %v", err)
	}
	if event.Type != "knowledge.trending.changed" {
		t.Fatalf("expected knowledge.trending.changed, got %s", event.Type)
	}
	payload, ok := event.Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %#v", event.Payload)
	}
	items, ok := payload["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected trending items payload, got %#v", payload)
	}
}

func TestPhase61KnowledgeBriefBackfillStatsRealtimeAndPresenceBulk(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)
	AIGateway = stubGateway{}
	defer SetAIGateway(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	expiresAt := now.Add(2 * time.Minute)
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com", Status: "online", LastSeenAt: &now, PresenceExpiresAt: &expiresAt})
	db.DB.Create(&domain.User{ID: "user-2", OrganizationID: "org-1", Name: "Windsurf", Email: "windsurf@example.com", Status: "away", LastSeenAt: &now})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay", KnowledgeSpikeThreshold: 3, KnowledgeSpikeCooldownMins: 360})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-2"})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "AI-native launch workspace", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "concept", Title: "Agent Mesh", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff has clear owners", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-1", UserID: "user-2", Content: "Agent Mesh depends on Launch Program docs", CreatedAt: now.Add(-90 * time.Minute)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityFollow{ID: "follow-1", WorkspaceID: "ws-1", EntityID: "entity-1", UserID: "user-1", NotificationLevel: "all", CreatedAt: now.Add(-2 * time.Hour)})

	router := gin.New()
	router.GET("/api/v1/presence/bulk", GetPresenceBulk)
	router.POST("/api/v1/knowledge/entities/:id/brief", GenerateKnowledgeEntityBrief)
	router.POST("/api/v1/knowledge/weekly-brief", GenerateMyKnowledgeWeeklyBrief)
	router.GET("/api/v1/knowledge/entities/:id/activity/backfill-status", GetKnowledgeEntityActivityBackfillStatus)
	router.POST("/api/v1/knowledge/entities/:id/activity/backfill", BackfillKnowledgeEntityActivity)
	router.POST("/api/v1/knowledge/entities/:id/follow", FollowKnowledgeEntity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/presence/bulk?channel_id=ch-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 presence bulk, got %d body=%s", rec.Code, rec.Body.String())
	}
	var presencePayload struct {
		Users []domain.User `json:"users"`
		Bulk  struct {
			OnlineCount  int `json:"online_count"`
			OfflineCount int `json:"offline_count"`
		} `json:"bulk"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &presencePayload); err != nil {
		t.Fatalf("decode presence bulk payload: %v", err)
	}
	if len(presencePayload.Users) != 2 || presencePayload.Bulk.OnlineCount != 1 || presencePayload.Bulk.OfflineCount != 1 {
		t.Fatalf("unexpected presence bulk payload: %#v", presencePayload)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/brief", strings.NewReader(`{"force":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 entity brief, got %d body=%s", rec.Code, rec.Body.String())
	}
	var briefPayload struct {
		Brief struct {
			EntityID string `json:"entity_id"`
			Content  string `json:"content"`
			Provider string `json:"provider"`
			Model    string `json:"model"`
		} `json:"brief"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &briefPayload); err != nil {
		t.Fatalf("decode entity brief payload: %v", err)
	}
	if briefPayload.Brief.EntityID != "entity-1" || !strings.Contains(briefPayload.Brief.Content, "Done.") || briefPayload.Brief.Provider != "stub" {
		t.Fatalf("unexpected entity brief payload: %#v", briefPayload.Brief)
	}
	assertRealtimeEventType(t, client, "knowledge.entity.brief.generated")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/weekly-brief", strings.NewReader(`{"workspace_id":"ws-1","force":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 weekly brief, got %d body=%s", rec.Code, rec.Body.String())
	}
	var weeklyPayload struct {
		Brief struct {
			UserID  string `json:"user_id"`
			Content string `json:"content"`
		} `json:"brief"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &weeklyPayload); err != nil {
		t.Fatalf("decode weekly brief payload: %v", err)
	}
	if weeklyPayload.Brief.UserID != "user-1" || !strings.Contains(weeklyPayload.Brief.Content, "Done.") {
		t.Fatalf("unexpected weekly brief payload: %#v", weeklyPayload.Brief)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/entities/entity-1/activity/backfill-status", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 backfill status, got %d body=%s", rec.Code, rec.Body.String())
	}
	var statusPayload struct {
		Status struct {
			EntityID        string `json:"entity_id"`
			MissingRefCount int    `json:"missing_ref_count"`
			IsBackfilled    bool   `json:"is_backfilled"`
		} `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &statusPayload); err != nil {
		t.Fatalf("decode backfill status payload: %v", err)
	}
	if statusPayload.Status.EntityID != "entity-1" || statusPayload.Status.MissingRefCount != 1 || statusPayload.Status.IsBackfilled {
		t.Fatalf("unexpected backfill status before run: %#v", statusPayload.Status)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/activity/backfill", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 backfill, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.entity.ref.created")
	assertRealtimeEventType(t, client, "knowledge.trending.changed")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-2/follow", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 follow entity-2, got %d body=%s", rec.Code, rec.Body.String())
	}
	raw, err := client.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("failed to receive followed stats event: %v", err)
	}
	var event realtime.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("decode followed stats event: %v", err)
	}
	if event.Type != "knowledge.followed.stats.changed" {
		t.Fatalf("expected knowledge.followed.stats.changed, got %s", event.Type)
	}
}

func TestPhase62CachedBriefsRealtimeAndBulkRead(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)
	AIGateway = stubGateway{}
	defer SetAIGateway(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay", KnowledgeSpikeThreshold: 3, KnowledgeSpikeCooldownMins: 360})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "AI-native launch workspace", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff has clear owners", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityFollow{ID: "follow-1", WorkspaceID: "ws-1", EntityID: "entity-1", UserID: "user-1", NotificationLevel: "all", CreatedAt: now.Add(-2 * time.Hour)})

	router := gin.New()
	router.GET("/api/v1/knowledge/entities/:id/brief", GetKnowledgeEntityBrief)
	router.POST("/api/v1/knowledge/entities/:id/brief", GenerateKnowledgeEntityBrief)
	router.GET("/api/v1/knowledge/weekly-brief", GetMyKnowledgeWeeklyBrief)
	router.POST("/api/v1/knowledge/weekly-brief", GenerateMyKnowledgeWeeklyBrief)
	router.POST("/api/v1/notifications/bulk-read", MarkNotificationsBulkRead)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/entities/entity-1/brief", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 empty cached entity brief, got %d body=%s", rec.Code, rec.Body.String())
	}
	var emptyBriefPayload struct {
		Brief *knowledge.EntityBrief `json:"brief"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &emptyBriefPayload); err != nil {
		t.Fatalf("decode empty entity brief payload: %v", err)
	}
	if emptyBriefPayload.Brief != nil {
		t.Fatalf("expected nil cached entity brief before generation, got %#v", emptyBriefPayload.Brief)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/brief", strings.NewReader(`{"force":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 generate entity brief, got %d body=%s", rec.Code, rec.Body.String())
	}
	raw, err := client.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("failed to receive brief generated event: %v", err)
	}
	var event realtime.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("decode brief event: %v", err)
	}
	if event.Type != "knowledge.entity.brief.generated" || event.EntityID != "entity-1" {
		t.Fatalf("expected knowledge.entity.brief.generated for entity-1, got %#v", event)
	}

	SetAIGateway(nil)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/entities/entity-1/brief", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 cached entity brief without llm gateway, got %d body=%s", rec.Code, rec.Body.String())
	}
	var cachedBriefPayload struct {
		Brief knowledge.EntityBrief `json:"brief"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &cachedBriefPayload); err != nil {
		t.Fatalf("decode cached entity brief payload: %v", err)
	}
	if !cachedBriefPayload.Brief.Cached || cachedBriefPayload.Brief.EntityID != "entity-1" || cachedBriefPayload.Brief.Title != "Launch Program" {
		t.Fatalf("unexpected cached entity brief payload: %#v", cachedBriefPayload.Brief)
	}

	AIGateway = stubGateway{}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/weekly-brief", strings.NewReader(`{"workspace_id":"ws-1","force":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 generate weekly brief, got %d body=%s", rec.Code, rec.Body.String())
	}

	SetAIGateway(nil)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/weekly-brief?workspace_id=ws-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 cached weekly brief without llm gateway, got %d body=%s", rec.Code, rec.Body.String())
	}
	var cachedWeeklyPayload struct {
		Brief knowledge.WeeklyBrief `json:"brief"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &cachedWeeklyPayload); err != nil {
		t.Fatalf("decode cached weekly brief payload: %v", err)
	}
	if !cachedWeeklyPayload.Brief.Cached || cachedWeeklyPayload.Brief.UserID != "user-1" || cachedWeeklyPayload.Brief.WorkspaceID != "ws-1" {
		t.Fatalf("unexpected cached weekly brief payload: %#v", cachedWeeklyPayload.Brief)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/notifications/bulk-read", strings.NewReader(`{"item_ids":["activity-mention-msg-1","knowledge-digest-msg-2","activity-thread-msg-3"]}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 bulk read, got %d body=%s", rec.Code, rec.Body.String())
	}
	var readCount int64
	db.DB.Model(&domain.NotificationRead{}).Where("user_id = ?", "user-1").Count(&readCount)
	if readCount != 3 {
		t.Fatalf("expected 3 notification read rows, got %d", readCount)
	}
	raw, err = client.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("failed to receive bulk read event: %v", err)
	}
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("decode bulk read event: %v", err)
	}
	if event.Type != "notifications.bulk_read" {
		t.Fatalf("expected notifications.bulk_read event, got %s", event.Type)
	}
}

func TestPhase63EntityAskWeeklyShareAndBriefInvalidation(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)
	AIGateway = stubGateway{}
	defer SetAIGateway(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay", KnowledgeSpikeThreshold: 3, KnowledgeSpikeCooldownMins: 360})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "AI-native launch workspace", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff has clear owners", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.Message{ID: "msg-seed-2", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program moved into release review", CreatedAt: now.Add(-100 * time.Minute)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-seed-2", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-seed-2", Role: "decision", CreatedAt: now.Add(-100 * time.Minute)})
	db.DB.Create(&domain.KnowledgeEvent{ID: "kevent-1", WorkspaceID: "ws-1", EntityID: "entity-1", EventType: "decision", Title: "Ship v1", Body: "Decision confirmed", ActorUserID: "user-1", CreatedAt: now.Add(-90 * time.Minute), OccurredAt: now.Add(-90 * time.Minute)})
	db.DB.Create(&domain.KnowledgeEntityFollow{ID: "follow-1", WorkspaceID: "ws-1", EntityID: "entity-1", UserID: "user-1", NotificationLevel: "all", CreatedAt: now.Add(-2 * time.Hour)})

	router := gin.New()
	router.POST("/api/v1/knowledge/entities/:id/brief", GenerateKnowledgeEntityBrief)
	router.POST("/api/v1/knowledge/entities/:id/ask", AskKnowledgeEntity)
	router.POST("/api/v1/knowledge/weekly-brief", GenerateMyKnowledgeWeeklyBrief)
	router.POST("/api/v1/knowledge/weekly-brief/:id/share", ShareMyKnowledgeWeeklyBrief)
	router.POST("/api/v1/knowledge/entities/:id/refs", AddKnowledgeEntityRef)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/brief", strings.NewReader(`{"force":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 generate entity brief, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.entity.brief.generated")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/ask", strings.NewReader(`{"question":"What decisions were made last week?"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 entity ask, got %d body=%s", rec.Code, rec.Body.String())
	}
	var askPayload struct {
		Answer struct {
			Entity struct {
				ID string `json:"id"`
			} `json:"entity"`
			Question  string               `json:"question"`
			Answer    string               `json:"answer"`
			Citations []knowledge.Citation `json:"citations"`
			Provider  string               `json:"provider"`
			Model     string               `json:"model"`
		} `json:"answer"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &askPayload); err != nil {
		t.Fatalf("decode ask payload: %v", err)
	}
	if askPayload.Answer.Entity.ID != "entity-1" || askPayload.Answer.Question == "" || !strings.Contains(askPayload.Answer.Answer, "Done.") || len(askPayload.Answer.Citations) == 0 {
		t.Fatalf("unexpected ask payload: %#v", askPayload.Answer)
	}
	assertRealtimeEventType(t, client, "knowledge.entity.ask.answered")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/weekly-brief", strings.NewReader(`{"workspace_id":"ws-1","force":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 generate weekly brief, got %d body=%s", rec.Code, rec.Body.String())
	}
	var weeklyPayload struct {
		Brief struct {
			ID          string `json:"id"`
			WorkspaceID string `json:"workspace_id"`
		} `json:"brief"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &weeklyPayload); err != nil {
		t.Fatalf("decode weekly brief payload: %v", err)
	}
	if weeklyPayload.Brief.ID == "" || weeklyPayload.Brief.WorkspaceID != "ws-1" {
		t.Fatalf("unexpected weekly brief payload: %#v", weeklyPayload.Brief)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/weekly-brief/"+weeklyPayload.Brief.ID+"/share", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 weekly brief share, got %d body=%s", rec.Code, rec.Body.String())
	}
	var sharePayload struct {
		Share struct {
			ID           string `json:"id"`
			WorkspaceID  string `json:"workspace_id"`
			URL          string `json:"url"`
			ShortURL     string `json:"short_url"`
			RelativePath string `json:"relative_path"`
		} `json:"share"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &sharePayload); err != nil {
		t.Fatalf("decode weekly brief share payload: %v", err)
	}
	if sharePayload.Share.ID != weeklyPayload.Brief.ID || sharePayload.Share.WorkspaceID != "ws-1" || !strings.Contains(sharePayload.Share.URL, "/workspace/knowledge/following") {
		t.Fatalf("unexpected weekly brief share payload: %#v", sharePayload.Share)
	}

	db.DB.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program moved to release prep", CreatedAt: now.Add(-30 * time.Minute)})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/refs", strings.NewReader(`{"ref_kind":"message","ref_id":"msg-2","role":"decision"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 entity ref create, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.entity.ref.created")
	assertRealtimeEventType(t, client, "knowledge.entity.activity.spiked")
	raw, err := client.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("failed to receive invalidation event: %v", err)
	}
	var event realtime.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("decode invalidation event: %v", err)
	}
	if event.Type != "knowledge.trending.changed" {
		t.Fatalf("expected knowledge.trending.changed before brief invalidation, got %s", event.Type)
	}
	raw, err = client.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("failed to receive brief.changed event: %v", err)
	}
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("decode brief.changed event: %v", err)
	}
	if event.Type != "knowledge.entity.brief.changed" || event.EntityID != "entity-1" {
		t.Fatalf("expected knowledge.entity.brief.changed for entity-1, got %#v", event)
	}
}

func TestPhase63BAIComposeContract(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	AIGateway = stubGateway{}
	defer SetAIGateway(nil)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1"})
	db.DB.Create(&domain.Message{ID: "msg-parent", ChannelID: "ch-1", UserID: "user-1", Content: "Can we confirm the launch owner and timeline?", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.Message{ID: "msg-thread-1", ChannelID: "ch-1", ThreadID: "msg-parent", UserID: "user-1", Content: "The Friday review is still the current target.", CreatedAt: now.Add(-90 * time.Minute)})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "AI-native launch workspace", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-parent", Role: "discussion", CreatedAt: now.Add(-2 * time.Hour)})

	router := gin.New()
	router.POST("/api/v1/ai/compose", ComposeAI)

	assertCompose := func(t *testing.T, rec *httptest.ResponseRecorder, wantChannelID, wantThreadID string) {
		t.Helper()

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}

		var payload struct {
			Compose struct {
				ChannelID   string `json:"channel_id"`
				ThreadID    string `json:"thread_id"`
				Intent      string `json:"intent"`
				Suggestions []struct {
					ID   string `json:"id"`
					Text string `json:"text"`
					Tone string `json:"tone"`
					Kind string `json:"kind"`
				} `json:"suggestions"`
				Citations []struct {
					ID           string `json:"id"`
					EvidenceKind string `json:"evidence_kind"`
					SourceKind   string `json:"source_kind"`
					SourceRef    string `json:"source_ref"`
					RefKind      string `json:"ref_kind"`
					Snippet      string `json:"snippet"`
					Title        string `json:"title"`
					Score        int    `json:"score"`
					EntityID     string `json:"entity_id"`
					EntityTitle  string `json:"entity_title"`
				} `json:"citations"`
				ContextEntities []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
					Kind  string `json:"kind"`
				} `json:"context_entities"`
			} `json:"compose"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode compose payload: %v", err)
		}

		if payload.Compose.ChannelID != wantChannelID || payload.Compose.ThreadID != wantThreadID || payload.Compose.Intent != "reply" {
			t.Fatalf("unexpected compose envelope: %#v", payload.Compose)
		}
		if len(payload.Compose.Suggestions) == 0 {
			t.Fatalf("expected 1 suggestion, got %#v", payload.Compose.Suggestions)
		}
		if got := payload.Compose.Suggestions[0]; got.ID == "" || got.Text == "" || got.Tone == "" || got.Kind != "reply" {
			t.Fatalf("unexpected suggestion: %#v", got)
		}
		if len(payload.Compose.Citations) == 0 {
			t.Fatalf("expected at least 1 citation, got %#v", payload.Compose.Citations)
		}
		if got := payload.Compose.Citations[0]; got.ID == "" || got.EvidenceKind == "" || got.SourceRef == "" || got.EntityID != "entity-1" || got.EntityTitle != "Launch Program" {
			t.Fatalf("unexpected citation: %#v", got)
		}
		if len(payload.Compose.ContextEntities) != 1 {
			t.Fatalf("expected 1 context entity, got %#v", payload.Compose.ContextEntities)
		}
		if got := payload.Compose.ContextEntities[0]; got.ID != "entity-1" || got.Title != "Launch Program" || got.Kind != "project" {
			t.Fatalf("unexpected context entity: %#v", got)
		}
	}

	t.Run("channel compose", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose", strings.NewReader(`{"channel_id":"ch-1","draft":"Can we confirm the launch owner and timeline?","intent":"reply","limit":3}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assertCompose(t, rec, "ch-1", "")
	})

	t.Run("thread compose", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose", strings.NewReader(`{"channel_id":"ch-1","thread_id":"msg-parent","draft":"The Friday review is still the current target.","intent":"reply","limit":3}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assertCompose(t, rec, "ch-1", "msg-parent")
	})

	t.Run("missing ai gateway", func(t *testing.T) {
		SetAIGateway(nil)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose", strings.NewReader(`{"channel_id":"ch-1","draft":"Can we confirm the launch owner and timeline?","intent":"reply","limit":3}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestPhase63CComposeStreamContract(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1"})
	db.DB.Create(&domain.Message{ID: "msg-parent", ChannelID: "ch-1", UserID: "user-1", Content: "Can we confirm the launch owner and timeline?", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.Message{ID: "msg-thread-1", ChannelID: "ch-1", ThreadID: "msg-parent", UserID: "user-1", Content: "The Friday review is still the current target.", CreatedAt: now.Add(-90 * time.Minute)})

	router := gin.New()
	router.POST("/api/v1/ai/compose/stream", ComposeAIStream)

	t.Run("channel stream", func(t *testing.T) {
		AIGateway = stubGateway{}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose/stream", strings.NewReader(`{"channel_id":"ch-1","draft":"Can we confirm the launch owner and timeline?","intent":"reply","limit":3}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/event-stream") {
			t.Fatalf("expected event-stream content type, got %s", got)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "event: start") {
			t.Fatalf("expected start event, got %s", body)
		}
		if !strings.Contains(body, "event: suggestion.delta") {
			t.Fatalf("expected suggestion.delta event, got %s", body)
		}
		if !strings.Contains(body, "event: suggestion.done") {
			t.Fatalf("expected suggestion.done event, got %s", body)
		}
		if !strings.Contains(body, "event: done") {
			t.Fatalf("expected done event, got %s", body)
		}
		if !strings.Contains(body, `"request_id":"compose-request-`) {
			t.Fatalf("expected request_id in start/done payload, got %s", body)
		}
		if !strings.Contains(body, `"suggestion_count":1`) {
			t.Fatalf("expected suggestion_count in done payload, got %s", body)
		}
	})

	t.Run("thread stream", func(t *testing.T) {
		AIGateway = stubGateway{}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose/stream", strings.NewReader(`{"channel_id":"ch-1","thread_id":"msg-parent","draft":"The Friday review is still the current target.","intent":"reply","limit":3}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		if !strings.Contains(body, `"thread_id":"msg-parent"`) {
			t.Fatalf("expected thread_id in stream payload, got %s", body)
		}
		if !strings.Contains(body, "event: suggestion.done") {
			t.Fatalf("expected suggestion.done event, got %s", body)
		}
	})

	t.Run("stream error after headers", func(t *testing.T) {
		AIGateway = failingAfterChunkGateway{}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose/stream", strings.NewReader(`{"channel_id":"ch-1","draft":"Can we confirm the launch owner and timeline?","intent":"reply","limit":3}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		if !strings.Contains(body, "event: error") || !strings.Contains(body, "upstream stream failed") {
			t.Fatalf("expected stream error event, got %s", body)
		}
	})

	t.Run("missing ai gateway", func(t *testing.T) {
		AIGateway = nil
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose/stream", strings.NewReader(`{"channel_id":"ch-1","draft":"Can we confirm the launch owner and timeline?","intent":"reply","limit":3}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestPhase63CComposeFeedbackContract(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1"})

	router := gin.New()
	router.POST("/api/v1/ai/compose/:id/feedback", SubmitAIComposeFeedback)

	postFeedback := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose/compose-1/feedback", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	t.Run("creates row", func(t *testing.T) {
		rec := postFeedback(`{"channel_id":"ch-1","thread_id":"msg-parent","intent":"reply","feedback":"up","suggestion_text":"Let's confirm the owner and keep Friday as the target."}`)
		if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
			t.Fatalf("expected success, got %d body=%s", rec.Code, rec.Body.String())
		}

		var rows []struct {
			ID             string    `gorm:"column:id"`
			ComposeID      string    `gorm:"column:compose_id"`
			UserID         string    `gorm:"column:user_id"`
			ChannelID      string    `gorm:"column:channel_id"`
			ThreadID       string    `gorm:"column:thread_id"`
			Intent         string    `gorm:"column:intent"`
			Feedback       string    `gorm:"column:feedback"`
			SuggestionText string    `gorm:"column:suggestion_text"`
			CreatedAt      time.Time `gorm:"column:created_at"`
			UpdatedAt      time.Time `gorm:"column:updated_at"`
		}
		if err := db.DB.Table("ai_compose_feedbacks").Where("compose_id = ? AND user_id = ?", "compose-1", "user-1").Find(&rows).Error; err != nil {
			t.Fatalf("failed to load compose feedback rows: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected one compose feedback row, got %#v", rows)
		}
		if rows[0].ComposeID != "compose-1" || rows[0].UserID != "user-1" || rows[0].ChannelID != "ch-1" || rows[0].ThreadID != "msg-parent" || rows[0].Intent != "reply" || rows[0].Feedback != "up" {
			t.Fatalf("unexpected compose feedback row: %#v", rows[0])
		}
		if rows[0].SuggestionText != "Let's confirm the owner and keep Friday as the target." {
			t.Fatalf("unexpected suggestion text: %#v", rows[0])
		}
	})

	t.Run("updates same row", func(t *testing.T) {
		rec := postFeedback(`{"channel_id":"ch-1","thread_id":"msg-parent","intent":"reply","feedback":"edited","suggestion_text":"Let's confirm the owner and keep Friday as the target."}`)
		if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
			t.Fatalf("expected success, got %d body=%s", rec.Code, rec.Body.String())
		}

		var rows []struct {
			ID       string `gorm:"column:id"`
			Feedback string `gorm:"column:feedback"`
		}
		if err := db.DB.Table("ai_compose_feedbacks").Where("compose_id = ? AND user_id = ?", "compose-1", "user-1").Find(&rows).Error; err != nil {
			t.Fatalf("failed to load compose feedback rows: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected one compose feedback row after update, got %#v", rows)
		}
		if rows[0].Feedback != "edited" {
			t.Fatalf("expected feedback to update in place, got %#v", rows[0])
		}
	})

	t.Run("invalid feedback", func(t *testing.T) {
		rec := postFeedback(`{"channel_id":"ch-1","intent":"reply","feedback":"maybe"}`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestPhase63DComposeDMAndIntentContracts(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	AIGateway = stubGateway{}
	defer SetAIGateway(nil)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", OrganizationID: "org-1", Name: "Windsurf", Email: "windsurf@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Can we confirm the launch owner and timeline?", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.DMConversation{ID: "dm-1", CreatedAt: now.Add(-3 * time.Hour)})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-1"})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-2"})
	db.DB.Create(&domain.DMMessage{ID: "dm-msg-1", DMConversationID: "dm-1", UserID: "user-2", Content: "Can Codex unblock the composer DM parity?", CreatedAt: now.Add(-45 * time.Minute)})

	router := gin.New()
	router.POST("/api/v1/ai/compose", ComposeAI)
	router.POST("/api/v1/ai/compose/stream", ComposeAIStream)

	t.Run("dm compose uses dm scope", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose", strings.NewReader(`{"dm_id":"dm-1","draft":"What should I reply?","intent":"followup","limit":2}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		var payload struct {
			Compose struct {
				DMID        string `json:"dm_id"`
				ChannelID   string `json:"channel_id"`
				Intent      string `json:"intent"`
				Suggestions []struct {
					ID   string `json:"id"`
					Text string `json:"text"`
					Kind string `json:"kind"`
				} `json:"suggestions"`
				Provider string `json:"provider"`
				Model    string `json:"model"`
			} `json:"compose"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode compose payload: %v", err)
		}
		if payload.Compose.DMID != "dm-1" || payload.Compose.ChannelID != "" || payload.Compose.Intent != "followup" {
			t.Fatalf("unexpected dm compose envelope: %#v", payload.Compose)
		}
		if len(payload.Compose.Suggestions) == 0 || payload.Compose.Suggestions[0].ID == "" || payload.Compose.Suggestions[0].Text == "" {
			t.Fatalf("expected dm compose suggestions, got %#v", payload.Compose.Suggestions)
		}
	})

	t.Run("dm stream includes dm scope", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose/stream", strings.NewReader(`{"dm_id":"dm-1","draft":"Summarize this private sync","intent":"summarize","limit":1}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		if !strings.Contains(body, `"dm_id":"dm-1"`) || !strings.Contains(body, `"intent":"summarize"`) {
			t.Fatalf("expected dm_id and summarize intent in stream payload, got %s", body)
		}
		if !strings.Contains(body, "event: suggestion.done") || !strings.Contains(body, "event: done") {
			t.Fatalf("expected completed stream events, got %s", body)
		}
	})

	t.Run("supported intents", func(t *testing.T) {
		for _, intent := range []string{"reply", "summarize", "followup", "schedule"} {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose", strings.NewReader(`{"channel_id":"ch-1","draft":"Need next step","intent":"`+intent+`","limit":1}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 for intent %s, got %d body=%s", intent, rec.Code, rec.Body.String())
			}
		}
	})

	t.Run("invalid scope and intent", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose", strings.NewReader(`{"channel_id":"ch-1","dm_id":"dm-1","draft":"Need next step","intent":"reply"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for both scopes, got %d body=%s", rec.Code, rec.Body.String())
		}

		req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose", strings.NewReader(`{"channel_id":"ch-1","draft":"Need next step","intent":"unknown"}`))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for unknown intent, got %d body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestPhase63DComposeFeedbackSummaryContract(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", OrganizationID: "org-1", Name: "Windsurf", Email: "windsurf@example.com"})
	db.DB.Create(&domain.DMConversation{ID: "dm-1", CreatedAt: now.Add(-3 * time.Hour)})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-1"})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-2"})

	router := gin.New()
	router.POST("/api/v1/ai/compose/:id/feedback", SubmitAIComposeFeedback)
	router.GET("/api/v1/ai/compose/:id/feedback/summary", GetAIComposeFeedbackSummary)

	postFeedback := func(path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	rec := postFeedback("/api/v1/ai/compose/compose-dm-1/feedback", `{"dm_id":"dm-1","intent":"followup","feedback":"up","suggestion_text":"Can you share the blocker?"}`)
	if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
		t.Fatalf("expected success for dm feedback, got %d body=%s", rec.Code, rec.Body.String())
	}

	var feedback domain.AIComposeFeedback
	if err := db.DB.First(&feedback, "compose_id = ? AND user_id = ?", "compose-dm-1", "user-1").Error; err != nil {
		t.Fatalf("expected dm compose feedback row: %v", err)
	}
	if feedback.DMConversationID != "dm-1" || feedback.ChannelID != "" || feedback.Intent != "followup" || feedback.Feedback != "up" {
		t.Fatalf("unexpected dm compose feedback row: %#v", feedback)
	}

	db.DB.Create(&domain.AIComposeFeedback{ID: "compose-feedback-2", ComposeID: "compose-dm-1", UserID: "user-2", DMConversationID: "dm-1", Intent: "followup", Feedback: "down", SuggestionText: "Needs more context.", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.AIComposeFeedback{ID: "compose-feedback-3", ComposeID: "compose-dm-1", UserID: "user-3", DMConversationID: "dm-1", Intent: "followup", Feedback: "edited", SuggestionText: "Edited before send.", CreatedAt: now, UpdatedAt: now})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/compose/compose-dm-1/feedback/summary", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on feedback summary, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Summary struct {
			ComposeID string `json:"compose_id"`
			Total     int64  `json:"total"`
			Counts    struct {
				Up     int64 `json:"up"`
				Down   int64 `json:"down"`
				Edited int64 `json:"edited"`
			} `json:"counts"`
			Recent []domain.AIComposeFeedback `json:"recent"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode feedback summary: %v", err)
	}
	if payload.Summary.ComposeID != "compose-dm-1" || payload.Summary.Total != 3 || payload.Summary.Counts.Up != 1 || payload.Summary.Counts.Down != 1 || payload.Summary.Counts.Edited != 1 {
		t.Fatalf("unexpected feedback summary: %#v", payload.Summary)
	}
	if len(payload.Summary.Recent) != 3 {
		t.Fatalf("expected 3 recent feedback rows, got %#v", payload.Summary.Recent)
	}
}

func TestPhase63EEntityAskStreamAndHistoryContracts(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	AIGateway = stubGateway{}
	defer SetAIGateway(nil)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", OrganizationID: "org-1", Name: "Windsurf", Email: "windsurf@example.com"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "AI-native launch workspace", Status: "active", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program has a release review decision.", CreatedAt: now.Add(-2 * time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "decision", CreatedAt: now.Add(-2 * time.Hour)})

	router := gin.New()
	router.POST("/api/v1/knowledge/entities/:id/ask", AskKnowledgeEntity)
	router.GET("/api/v1/knowledge/entities/:id/ask/history", GetKnowledgeEntityAskHistory)
	router.POST("/api/v1/knowledge/entities/:id/ask/stream", AskKnowledgeEntityStream)

	t.Run("sync ask persists history", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/ask", strings.NewReader(`{"question":"What decision exists?"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 sync ask, got %d body=%s", rec.Code, rec.Body.String())
		}

		var rows []domain.KnowledgeEntityAskAnswer
		if err := db.DB.Where("entity_id = ? AND user_id = ?", "entity-1", "user-1").Order("answered_at desc").Find(&rows).Error; err != nil {
			t.Fatalf("failed to load ask history rows: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected 1 ask history row, got %#v", rows)
		}
		if rows[0].Question != "What decision exists?" || !strings.Contains(rows[0].Answer, "Done.") || rows[0].CitationCount == 0 {
			t.Fatalf("unexpected ask history row: %#v", rows[0])
		}
	})

	t.Run("history is current user and newest first", func(t *testing.T) {
		db.DB.Create(&domain.KnowledgeEntityAskAnswer{
			ID:            "entity-ask-old",
			EntityID:      "entity-1",
			WorkspaceID:   "ws-1",
			UserID:        "user-1",
			Question:      "Older question",
			Answer:        "Older answer",
			Provider:      "stub",
			Model:         "stub-model",
			CitationCount: 1,
			AnsweredAt:    now.Add(-30 * time.Minute),
			CreatedAt:     now.Add(-30 * time.Minute),
			UpdatedAt:     now.Add(-30 * time.Minute),
		})
		db.DB.Create(&domain.KnowledgeEntityAskAnswer{
			ID:            "entity-ask-other-user",
			EntityID:      "entity-1",
			WorkspaceID:   "ws-1",
			UserID:        "user-2",
			Question:      "Other user question",
			Answer:        "Other answer",
			Provider:      "stub",
			Model:         "stub-model",
			CitationCount: 1,
			AnsweredAt:    now,
			CreatedAt:     now,
			UpdatedAt:     now,
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/entities/entity-1/ask/history?limit=10", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 history, got %d body=%s", rec.Code, rec.Body.String())
		}

		var payload struct {
			Entity domain.KnowledgeEntity            `json:"entity"`
			Items  []domain.KnowledgeEntityAskAnswer `json:"items"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode history payload: %v", err)
		}
		if payload.Entity.ID != "entity-1" {
			t.Fatalf("expected entity context, got %#v", payload.Entity)
		}
		if len(payload.Items) != 2 {
			t.Fatalf("expected 2 current-user history rows, got %#v", payload.Items)
		}
		if payload.Items[0].AnsweredAt.Before(payload.Items[1].AnsweredAt) {
			t.Fatalf("expected newest-first history, got %#v", payload.Items)
		}
		for _, item := range payload.Items {
			if item.UserID != "user-1" {
				t.Fatalf("history leaked another user's row: %#v", item)
			}
		}
	})

	t.Run("stream ask emits answer events and persists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/ask/stream", strings.NewReader(`{"question":"Stream the answer"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 stream ask, got %d body=%s", rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/event-stream") {
			t.Fatalf("expected event-stream content type, got %s", got)
		}
		body := rec.Body.String()
		for _, want := range []string{"event: start", "event: answer.delta", "event: answer.done", "event: done", `"entity_id":"entity-1"`} {
			if !strings.Contains(body, want) {
				t.Fatalf("expected %q in stream body, got %s", want, body)
			}
		}

		var count int64
		if err := db.DB.Model(&domain.KnowledgeEntityAskAnswer{}).Where("entity_id = ? AND user_id = ? AND question = ?", "entity-1", "user-1", "Stream the answer").Count(&count).Error; err != nil {
			t.Fatalf("failed to count stream history rows: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected stream ask to persist one row, got %d", count)
		}
	})
}

func TestPhase63FAutoSummarizeAndComposeRealtimeContracts(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	AIGateway = stubGateway{}
	defer SetAIGateway(nil)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program owner is Nikko", CreatedAt: now.Add(-10 * time.Minute)})
	db.DB.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-1", UserID: "user-1", Content: "Timeline is still aligned for Friday review", CreatedAt: now.Add(-5 * time.Minute)})

	router := gin.New()
	router.GET("/api/v1/channels/:id/knowledge/auto-summarize", GetChannelKnowledgeAutoSummarize)
	router.PUT("/api/v1/channels/:id/knowledge/auto-summarize", PutChannelKnowledgeAutoSummarize)
	router.POST("/api/v1/channels/:id/knowledge/auto-summarize", RunChannelKnowledgeAutoSummarize)
	router.POST("/api/v1/ai/compose", ComposeAI)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/channels/ch-1/knowledge/auto-summarize", strings.NewReader(`{"is_enabled":true,"window_hours":12,"message_limit":25,"min_new_messages":2,"provider":"stub","model":"stub-model"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on auto summarize setting, got %d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/channels/ch-1/knowledge/auto-summarize", strings.NewReader(`{"force":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on auto summarize run, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "channel.summary.updated")

	var runPayload struct {
		Setting domain.ChannelAutoSummarySetting `json:"setting"`
		Summary domain.AISummary                 `json:"summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &runPayload); err != nil {
		t.Fatalf("decode auto summarize run payload: %v", err)
	}
	if runPayload.Setting.ChannelID != "ch-1" || !runPayload.Setting.IsEnabled || runPayload.Summary.ScopeType != "channel" || runPayload.Summary.ScopeID != "ch-1" {
		t.Fatalf("unexpected auto summarize payload: %#v", runPayload)
	}
	if runPayload.Setting.LastRunAt == nil || runPayload.Setting.LastMessageAt == nil {
		t.Fatalf("expected setting run metadata, got %#v", runPayload.Setting)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose", strings.NewReader(`{"channel_id":"ch-1","draft":"Can we schedule the Launch Program review?","intent":"schedule","limit":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on schedule compose, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.compose.suggestion.generated")

	var composePayload struct {
		Compose struct {
			ProposedSlots []struct {
				StartsAt string `json:"starts_at"`
				Duration int    `json:"duration_minutes"`
				Timezone string `json:"timezone"`
			} `json:"proposed_slots"`
		} `json:"compose"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &composePayload); err != nil {
		t.Fatalf("decode compose payload: %v", err)
	}
	if len(composePayload.Compose.ProposedSlots) == 0 || composePayload.Compose.ProposedSlots[0].Duration == 0 {
		t.Fatalf("expected schedule proposed slots, got %#v", composePayload.Compose.ProposedSlots)
	}
}

func TestPhase63GComposeActivityPersistsAndLists(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	AIGateway = stubGateway{}
	defer SetAIGateway(nil)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Channel{ID: "ch-2", WorkspaceID: "ws-1", Name: "random", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program owner is Nikko", CreatedAt: now.Add(-10 * time.Minute)})
	db.DB.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-2", UserID: "user-1", Content: "Other channel context", CreatedAt: now.Add(-5 * time.Minute)})

	router := gin.New()
	router.POST("/api/v1/ai/compose", ComposeAI)
	router.GET("/api/v1/ai/compose/activity", GetAIComposeActivity)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose", strings.NewReader(`{"channel_id":"ch-1","draft":"Please draft the launch owner update","intent":"reply","limit":2}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on compose, got %d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/compose", strings.NewReader(`{"channel_id":"ch-2","draft":"Other draft","intent":"followup","limit":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on second compose, got %d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/ai/compose/activity?channel_id=ch-1&limit=10", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on compose activity, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []domain.AIComposeActivity `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode compose activity payload: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected one channel-scoped compose activity item, got %#v", payload.Items)
	}
	item := payload.Items[0]
	if item.ChannelID != "ch-1" || item.WorkspaceID != "ws-1" || item.Intent != "reply" || item.SuggestionCount == 0 || item.ComposeID == "" {
		t.Fatalf("unexpected compose activity item: %#v", item)
	}
}

func TestPhase63HComposeActivityDigest(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})

	db.DB.Create(&domain.AIComposeActivity{
		ID:              "compose-activity-1",
		ComposeID:       "compose-1",
		WorkspaceID:     "ws-1",
		ChannelID:       "ch-1",
		UserID:          "user-1",
		Intent:          "reply",
		SuggestionCount: 2,
		Provider:        "openrouter",
		Model:           "nemotron",
		CreatedAt:       now.Add(-20 * time.Minute),
	})
	db.DB.Create(&domain.AIComposeActivity{
		ID:              "compose-activity-2",
		ComposeID:       "compose-2",
		WorkspaceID:     "ws-1",
		ChannelID:       "ch-1",
		UserID:          "",
		Intent:          "reply",
		SuggestionCount: 1,
		Provider:        "openrouter",
		Model:           "nemotron",
		CreatedAt:       now.Add(-10 * time.Minute),
	})
	db.DB.Create(&domain.AIComposeActivity{
		ID:              "compose-activity-3",
		ComposeID:       "compose-3",
		WorkspaceID:     "ws-1",
		ChannelID:       "ch-1",
		UserID:          "user-2",
		Intent:          "schedule",
		SuggestionCount: 3,
		Provider:        "gemini",
		Model:           "gemini-2.5-pro",
		CreatedAt:       now.Add(-5 * time.Minute),
	})

	router := gin.New()
	router.GET("/api/v1/ai/compose/activity/digest", GetAIComposeActivityDigest)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/compose/activity/digest?workspace_id=ws-1&channel_id=ch-1&window=24h&group_by=user&limit=10", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on compose activity digest, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Summary struct {
			TotalRequests int `json:"total_requests"`
			UniqueUsers   int `json:"unique_users"`
		} `json:"summary"`
		Breakdown []struct {
			Key   string `json:"key"`
			Count int    `json:"count"`
		} `json:"breakdown"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode compose activity digest payload: %v", err)
	}
	if payload.Summary.TotalRequests != 3 {
		t.Fatalf("expected 3 compose requests, got %#v", payload.Summary)
	}
	if payload.Summary.UniqueUsers != 2 {
		t.Fatalf("expected null user rows to be excluded from unique_users, got %#v", payload.Summary)
	}
	if len(payload.Breakdown) < 2 {
		t.Fatalf("expected grouped digest rows, got %#v", payload.Breakdown)
	}
	foundUnknown := false
	for _, item := range payload.Breakdown {
		if item.Key == "unknown" {
			foundUnknown = true
		}
	}
	if !foundUnknown {
		t.Fatalf("expected null user rows to map to unknown in user grouping, got %#v", payload.Breakdown)
	}
}

func TestPhase63IAutoSummarizeSchedulerRunsEnabledChannels(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	AIGateway = stubGateway{}
	defer SetAIGateway(nil)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	now := time.Now().UTC()
	lastRun := now.Add(-2 * time.Hour)
	lastMessageAt := now.Add(-90 * time.Minute)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch update one", CreatedAt: now.Add(-30 * time.Minute)})
	db.DB.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-1", UserID: "user-1", Content: "Launch update two", CreatedAt: now.Add(-20 * time.Minute)})
	db.DB.Create(&domain.ChannelAutoSummarySetting{
		ID:             "auto-summary-1",
		ChannelID:      "ch-1",
		WorkspaceID:    "ws-1",
		CreatedBy:      "user-1",
		IsEnabled:      true,
		WindowHours:    24,
		MessageLimit:   25,
		MinNewMessages: 2,
		Provider:       "stub",
		Model:          "stub-model",
		LastRunAt:      &lastRun,
		LastMessageAt:  &lastMessageAt,
		CreatedAt:      now.Add(-3 * time.Hour),
		UpdatedAt:      now.Add(-2 * time.Hour),
	})

	if err := processChannelAutoSummaries(now); err != nil {
		t.Fatalf("process channel auto summaries: %v", err)
	}

	var summary domain.AISummary
	if err := db.DB.Where("scope_type = ? AND scope_id = ?", "channel", "ch-1").First(&summary).Error; err != nil {
		t.Fatalf("expected scheduler to persist channel summary: %v", err)
	}
	if summary.Content == "" {
		t.Fatalf("expected non-empty summary content, got %#v", summary)
	}

	var refreshed domain.ChannelAutoSummarySetting
	if err := db.DB.First(&refreshed, "channel_id = ?", "ch-1").Error; err != nil {
		t.Fatalf("reload auto summary setting: %v", err)
	}
	if refreshed.LastRunAt == nil || !refreshed.LastRunAt.After(lastRun) {
		t.Fatalf("expected last_run_at to advance, got %#v", refreshed)
	}
	assertRealtimeEventType(t, client, "channel.summary.updated")
}

func TestPhase63IEntityAskRecentFeedAndRealtime(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)
	AIGateway = stubGateway{}
	defer SetAIGateway(nil)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com"})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", SourceKind: "manual", CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-time.Hour)})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "service", Title: "Billing Service", Status: "active", SourceKind: "manual", CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-time.Hour)})
	db.DB.Create(&domain.KnowledgeEntityAskAnswer{
		ID:            "entity-ask-1",
		EntityID:      "entity-1",
		WorkspaceID:   "ws-1",
		UserID:        "user-1",
		Question:      "What changed?",
		Answer:        "Launch moved ahead.",
		Provider:      "stub",
		Model:         "stub-model",
		CitationCount: 1,
		AnsweredAt:    now.Add(-30 * time.Minute),
		CreatedAt:     now.Add(-30 * time.Minute),
		UpdatedAt:     now.Add(-30 * time.Minute),
	})

	router := gin.New()
	router.GET("/api/v1/knowledge/ask/recent", GetRecentKnowledgeEntityAsks)
	router.POST("/api/v1/knowledge/entities/:id/ask", AskKnowledgeEntity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-2/ask", strings.NewReader(`{"question":"What is the latest status?","provider":"stub","model":"stub-model"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on entity ask, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.entity.ask.answered")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/ask/recent?workspace_id=ws-1&limit=10", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on recent entity asks, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []struct {
			ID          string `json:"id"`
			EntityID    string `json:"entity_id"`
			WorkspaceID string `json:"workspace_id"`
			UserID      string `json:"user_id"`
			Question    string `json:"question"`
			EntityTitle string `json:"entity_title"`
			EntityKind  string `json:"entity_kind"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode recent entity asks payload: %v", err)
	}
	if len(payload.Items) < 2 {
		t.Fatalf("expected at least 2 recent asks, got %#v", payload.Items)
	}
	if payload.Items[0].EntityID != "entity-2" || payload.Items[0].WorkspaceID != "ws-1" || payload.Items[0].UserID != "user-1" {
		t.Fatalf("expected newest item to be freshly asked entity-2 row, got %#v", payload.Items[0])
	}
	if payload.Items[0].EntityTitle != "Billing Service" || payload.Items[0].EntityKind != "service" {
		t.Fatalf("expected denormalized entity metadata on recent ask row, got %#v", payload.Items[0])
	}
}

func TestPhase63IAutomationJobsAuditView(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	now := time.Now().UTC()
	db.DB.Create(&domain.AIAutomationJob{
		ID:            "job-1",
		JobType:       "entity_brief_regen",
		ScopeType:     "knowledge_entity",
		ScopeID:       "entity-1",
		WorkspaceID:   "ws-1",
		Status:        "running",
		TriggerReason: "brief_changed",
		DedupeKey:     "entity-1:2026042311",
		AttemptCount:  1,
		ScheduledAt:   now.Add(-10 * time.Minute),
		StartedAt:     ptrTime(now.Add(-9 * time.Minute)),
		CreatedAt:     now.Add(-10 * time.Minute),
		UpdatedAt:     now.Add(-8 * time.Minute),
	})
	db.DB.Create(&domain.AIAutomationJob{
		ID:            "job-2",
		JobType:       "entity_brief_regen",
		ScopeType:     "knowledge_entity",
		ScopeID:       "entity-2",
		WorkspaceID:   "ws-1",
		Status:        "failed",
		TriggerReason: "manual_retry",
		DedupeKey:     "entity-2:2026042310",
		AttemptCount:  2,
		LastError:     "provider timeout",
		ScheduledAt:   now.Add(-30 * time.Minute),
		FinishedAt:    ptrTime(now.Add(-25 * time.Minute)),
		CreatedAt:     now.Add(-30 * time.Minute),
		UpdatedAt:     now.Add(-25 * time.Minute),
	})
	db.DB.Create(&domain.AIAutomationJob{
		ID:            "job-3",
		JobType:       "entity_brief_regen",
		ScopeType:     "knowledge_entity",
		ScopeID:       "entity-3",
		WorkspaceID:   "ws-2",
		Status:        "succeeded",
		TriggerReason: "brief_changed",
		DedupeKey:     "entity-3:2026042309",
		AttemptCount:  1,
		ScheduledAt:   now.Add(-40 * time.Minute),
		FinishedAt:    ptrTime(now.Add(-35 * time.Minute)),
		CreatedAt:     now.Add(-40 * time.Minute),
		UpdatedAt:     now.Add(-35 * time.Minute),
	})

	router := gin.New()
	router.GET("/api/v1/ai/automation/jobs", ListAIAutomationJobs)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/automation/jobs?workspace_id=ws-1&status=failed&limit=10", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when listing automation jobs, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []domain.AIAutomationJob `json:"items"`
		Total int64                    `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode automation jobs payload: %v", err)
	}
	if payload.Total != 1 || len(payload.Items) != 1 {
		t.Fatalf("expected one filtered automation job, got total=%d items=%#v", payload.Total, payload.Items)
	}
	if payload.Items[0].ID != "job-2" || payload.Items[0].WorkspaceID != "ws-1" || payload.Items[0].Status != "failed" {
		t.Fatalf("unexpected automation job payload: %#v", payload.Items[0])
	}
}

func TestPhase64BUnifiedActivityFeedContract(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	now := time.Now().UTC()
	db.DB.Create(&domain.Workspace{ID: "ws-1", Name: "Relay"})
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", SourceKind: "manual", CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-time.Hour)})

	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch kickoff is live", CreatedAt: now.Add(-5 * time.Minute)})
	db.DB.Create(&domain.FileAsset{ID: "file-1", ChannelID: "ch-1", UploaderID: "user-1", Name: "launch-plan.md", ContentType: "text/markdown", SourceKind: "upload", CreatedAt: now.Add(-4 * time.Minute), UpdatedAt: now.Add(-4 * time.Minute)})
	db.DB.Create(&domain.AIScheduleBooking{ID: "booking-1", WorkspaceID: "ws-1", ChannelID: "ch-1", RequestedBy: "user-1", Title: "Launch sync", Description: "Review blockers", StartsAt: now.Add(time.Hour), EndsAt: now.Add(90 * time.Minute), Timezone: "UTC", Provider: "internal", Status: "booked", CreatedAt: now.Add(-3 * time.Minute), UpdatedAt: now.Add(-3 * time.Minute)})
	db.DB.Create(&domain.AIComposeActivity{ID: "compose-activity-1", ComposeID: "compose-1", WorkspaceID: "ws-1", ChannelID: "ch-1", UserID: "user-1", Intent: "reply", SuggestionCount: 2, Provider: "stub", Model: "stub-model", CreatedAt: now.Add(-2 * time.Minute)})
	db.DB.Create(&domain.KnowledgeEntityAskAnswer{ID: "entity-ask-1", EntityID: "entity-1", WorkspaceID: "ws-1", UserID: "user-1", Question: "What changed?", Answer: "Launch moved ahead.", Provider: "stub", Model: "stub-model", CitationCount: 1, AnsweredAt: now.Add(-90 * time.Second), CreatedAt: now.Add(-90 * time.Second), UpdatedAt: now.Add(-90 * time.Second)})
	db.DB.Create(&domain.AIAutomationJob{ID: "job-1", JobType: "entity_brief_regen", ScopeType: "knowledge_entity", ScopeID: "entity-1", WorkspaceID: "ws-1", Status: "running", TriggerReason: "brief_changed", DedupeKey: "entity-1:1", AttemptCount: 1, ScheduledAt: now.Add(-60 * time.Second), StartedAt: ptrTime(now.Add(-55 * time.Second)), CreatedAt: now.Add(-60 * time.Second), UpdatedAt: now.Add(-55 * time.Second)})

	router := gin.New()
	router.GET("/api/v1/activity/feed", GetActivityFeed)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity/feed?workspace_id=ws-1&limit=10", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when loading activity feed, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []struct {
			ID          string         `json:"id"`
			EventType   string         `json:"event_type"`
			WorkspaceID string         `json:"workspace_id"`
			ActorID     string         `json:"actor_id"`
			ActorName   string         `json:"actor_name"`
			ChannelID   string         `json:"channel_id"`
			ChannelName string         `json:"channel_name"`
			EntityID    string         `json:"entity_id"`
			EntityTitle string         `json:"entity_title"`
			EntityKind  string         `json:"entity_kind"`
			Title       string         `json:"title"`
			Body        string         `json:"body"`
			Link        string         `json:"link"`
			OccurredAt  string         `json:"occurred_at"`
			Meta        map[string]any `json:"meta"`
		} `json:"items"`
		NextCursor string `json:"next_cursor"`
		Total      int64  `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode activity feed payload: %v", err)
	}
	if len(payload.Items) < 6 {
		t.Fatalf("expected at least 6 unified feed items, got %#v", payload.Items)
	}

	eventTypes := map[string]bool{}
	for _, item := range payload.Items {
		eventTypes[item.EventType] = true
		if item.EventType == "" || item.Title == "" || item.OccurredAt == "" {
			t.Fatalf("expected unified feed contract fields on item, got %#v", item)
		}
	}
	for _, required := range []string{"message", "file_uploaded", "schedule_booking", "compose_activity", "knowledge_ask", "automation_job"} {
		if !eventTypes[required] {
			t.Fatalf("expected event type %s in unified feed, got %#v", required, payload.Items)
		}
	}

	var askItem struct {
		EntityTitle string
		EntityKind  string
		Link        string
	}
	for _, item := range payload.Items {
		if item.EventType == "knowledge_ask" {
			askItem.EntityTitle = item.EntityTitle
			askItem.EntityKind = item.EntityKind
			askItem.Link = item.Link
			break
		}
	}
	if askItem.EntityTitle != "Launch Program" || askItem.EntityKind != "project" || askItem.Link != "/workspace/knowledge/entity-1" {
		t.Fatalf("expected denormalized entity fields on knowledge_ask item, got %#v", askItem)
	}
}

func TestPhase64CArtifactAndToolRunFeedItems(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	now := time.Now().UTC()
	db.DB.Create(&domain.Workspace{ID: "ws-1", Name: "Relay"})
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Artifact{ID: "artifact-1", ChannelID: "ch-1", Title: "Launch Doc", Type: "document", Status: "live", Content: "v2", Source: "manual", CreatedBy: "user-1", UpdatedBy: "user-1", CreatedAt: now.Add(-20 * time.Minute), UpdatedAt: now.Add(-7 * time.Minute)})
	db.DB.Create(&domain.ToolDefinition{ID: "tool-1", Name: "Summarize Thread", Key: "summarize-thread", Category: "ai", IsEnabled: true, CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Hour)})
	toolInput := `{"channel_id":"ch-1"}`
	db.DB.Create(&domain.ToolRun{ID: "tool-run-1", ToolID: "tool-1", TriggeredBy: "user-1", Status: "succeeded", Input: toolInput, Summary: "Summarized launch blockers", StartedAt: now.Add(-6 * time.Minute), CompletedAt: ptrTime(now.Add(-5 * time.Minute)), CreatedAt: now.Add(-6 * time.Minute), UpdatedAt: now.Add(-5 * time.Minute)})

	router := gin.New()
	router.GET("/api/v1/activity/feed", GetActivityFeed)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity/feed?workspace_id=ws-1&limit=10", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when loading activity feed, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []struct {
			EventType  string         `json:"event_type"`
			ActorID    string         `json:"actor_id"`
			ActorName  string         `json:"actor_name"`
			ChannelID  string         `json:"channel_id"`
			Title      string         `json:"title"`
			Link       string         `json:"link"`
			OccurredAt string         `json:"occurred_at"`
			Meta       map[string]any `json:"meta"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode activity feed payload: %v", err)
	}

	var seenArtifact, seenTool bool
	for _, item := range payload.Items {
		switch item.EventType {
		case "artifact_updated":
			seenArtifact = true
			if item.ActorID != "user-1" || item.ActorName != "Nikko Fu" || item.ChannelID != "ch-1" || item.Title == "" || item.OccurredAt == "" {
				t.Fatalf("unexpected artifact feed item: %#v", item)
			}
			if item.Link == "" || item.Meta["artifact_id"] != "artifact-1" || item.Meta["artifact_type"] != "document" {
				t.Fatalf("expected artifact metadata on unified feed row, got %#v", item)
			}
		case "tool_run":
			seenTool = true
			if item.ActorID != "user-1" || item.ActorName != "Nikko Fu" || item.ChannelID != "ch-1" || item.Title == "" || item.OccurredAt == "" {
				t.Fatalf("unexpected tool run feed item: %#v", item)
			}
			if item.Link != "/workspace/workflows" || item.Meta["run_id"] != "tool-run-1" || item.Meta["tool_name"] != "Summarize Thread" || item.Meta["status"] != "succeeded" {
				t.Fatalf("expected tool_run metadata on unified feed row, got %#v", item)
			}
		}
	}

	if !seenArtifact || !seenTool {
		t.Fatalf("expected artifact_updated and tool_run in feed, got %#v", payload.Items)
	}

	filtered := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/activity/feed?workspace_id=ws-1&event_type=tool_run&limit=10", nil)
	router.ServeHTTP(filtered, req)
	if filtered.Code != http.StatusOK {
		t.Fatalf("expected 200 when filtering tool_run feed, got %d body=%s", filtered.Code, filtered.Body.String())
	}
	if !strings.Contains(filtered.Body.String(), `"event_type":"tool_run"`) || strings.Contains(filtered.Body.String(), `"event_type":"artifact_updated"`) {
		t.Fatalf("expected tool_run filter to exclude artifact_updated, got %s", filtered.Body.String())
	}
}

func TestPhase64CReplyAndDMFeedItems(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	now := time.Now().UTC()
	db.DB.Create(&domain.Workspace{ID: "ws-1", Name: "Relay"})
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Windsurf", Email: "windsurf@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-parent", ChannelID: "ch-1", UserID: "user-1", Content: "Can we confirm the launch owner?", CreatedAt: now.Add(-20 * time.Minute)})
	db.DB.Create(&domain.Message{ID: "msg-reply", ChannelID: "ch-1", ThreadID: "msg-parent", UserID: "user-2", Content: "Yes, Nikko owns the launch.", CreatedAt: now.Add(-8 * time.Minute)})
	db.DB.Create(&domain.DMConversation{ID: "dm-1", CreatedAt: now.Add(-30 * time.Minute)})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-1"})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-2"})
	db.DB.Create(&domain.DMMessage{ID: "dm-msg-1", DMConversationID: "dm-1", UserID: "user-2", Content: "Shipping the build now.", CreatedAt: now.Add(-5 * time.Minute)})

	router := gin.New()
	router.GET("/api/v1/activity/feed", GetActivityFeed)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity/feed?workspace_id=ws-1&limit=10", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when loading activity feed, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []struct {
			EventType  string         `json:"event_type"`
			ActorID    string         `json:"actor_id"`
			ActorName  string         `json:"actor_name"`
			ChannelID  string         `json:"channel_id"`
			DMID       string         `json:"dm_id"`
			Link       string         `json:"link"`
			OccurredAt string         `json:"occurred_at"`
			Meta       map[string]any `json:"meta"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode activity feed payload: %v", err)
	}

	var replyOK, dmOK bool
	for _, item := range payload.Items {
		switch item.EventType {
		case "reply":
			replyOK = true
			if item.ActorID != "user-2" || item.ActorName != "Windsurf" || item.ChannelID != "ch-1" || item.Link != "/workspace?c=ch-1&m=msg-reply" || item.OccurredAt == "" || item.Meta["thread_id"] != "msg-parent" {
				t.Fatalf("unexpected reply feed item: %#v", item)
			}
		case "dm_message":
			dmOK = true
			if item.ActorID != "user-2" || item.ActorName != "Windsurf" || item.DMID != "dm-1" || item.Link != "/workspace/dms?id=dm-1" || item.OccurredAt == "" || item.Meta["message_id"] != "dm-msg-1" {
				t.Fatalf("unexpected dm_message feed item: %#v", item)
			}
		}
	}
	if !replyOK || !dmOK {
		t.Fatalf("expected reply and dm_message in feed, got %#v", payload.Items)
	}

	filtered := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/activity/feed?workspace_id=ws-1&dm_id=dm-1&event_type=dm_message&limit=10", nil)
	router.ServeHTTP(filtered, req)
	if filtered.Code != http.StatusOK {
		t.Fatalf("expected 200 when filtering dm activity feed, got %d body=%s", filtered.Code, filtered.Body.String())
	}
	if !strings.Contains(filtered.Body.String(), `"event_type":"dm_message"`) || strings.Contains(filtered.Body.String(), `"event_type":"reply"`) {
		t.Fatalf("expected dm_message filter to exclude reply rows, got %s", filtered.Body.String())
	}
}

func TestPhase64CMentionAndReactionFeedItems(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	now := time.Now().UTC()
	db.DB.Create(&domain.Workspace{ID: "ws-1", Name: "Relay"})
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Windsurf", Email: "windsurf@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	mentionMetadata := `{"entity_mentions":[{"entity_id":"entity-1","entity_title":"Launch Program","entity_kind":"project","source_kind":"explicit","mention_text":"@Launch Program"}]}`
	db.DB.Create(&domain.Message{ID: "msg-mention", ChannelID: "ch-1", UserID: "user-2", Content: "@Launch Program is ready for review.", CreatedAt: now.Add(-10 * time.Minute), Metadata: mentionMetadata})
	db.DB.Create(&domain.Message{ID: "msg-react-target", ChannelID: "ch-1", UserID: "user-1", Content: "We are good to launch.", CreatedAt: now.Add(-9 * time.Minute)})
	db.DB.Create(&domain.MessageReaction{MessageID: "msg-react-target", UserID: "user-2", Emoji: "🔥", CreatedAt: now.Add(-4 * time.Minute)})

	router := gin.New()
	router.GET("/api/v1/activity/feed", GetActivityFeed)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity/feed?workspace_id=ws-1&limit=10", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when loading activity feed, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []struct {
			EventType  string         `json:"event_type"`
			ActorID    string         `json:"actor_id"`
			ActorName  string         `json:"actor_name"`
			ChannelID  string         `json:"channel_id"`
			Link       string         `json:"link"`
			OccurredAt string         `json:"occurred_at"`
			Meta       map[string]any `json:"meta"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode activity feed payload: %v", err)
	}

	var mentionOK, reactionOK bool
	for _, item := range payload.Items {
		switch item.EventType {
		case "mention":
			mentionOK = true
			if item.ActorID != "user-2" || item.ActorName != "Windsurf" || item.ChannelID != "ch-1" || item.Link != "/workspace?c=ch-1&m=msg-mention" || item.OccurredAt == "" || item.Meta["message_id"] != "msg-mention" || item.Meta["mention_kind"] != "entity" {
				t.Fatalf("unexpected mention feed item: %#v", item)
			}
		case "reaction":
			reactionOK = true
			if item.ActorID != "user-2" || item.ActorName != "Windsurf" || item.ChannelID != "ch-1" || item.Link != "/workspace?c=ch-1&m=msg-react-target" || item.OccurredAt == "" || item.Meta["message_id"] != "msg-react-target" || item.Meta["emoji"] != "🔥" {
				t.Fatalf("unexpected reaction feed item: %#v", item)
			}
		}
	}
	if !mentionOK || !reactionOK {
		t.Fatalf("expected mention and reaction in feed, got %#v", payload.Items)
	}
}

func TestPhase65AChannelMessagePersistsUserMentions(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com"})
	db.DB.Create(&domain.User{ID: "user-3", Name: "Ann Lee", Email: "ann@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})

	router := gin.New()
	router.POST("/api/v1/messages", CreateMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBufferString(`{"channel_id":"ch-1","content":"@Jane Smith please sync with @Ann Lee and @Jane Smith again","user_id":"user-1"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on message create, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Message struct {
			ID       string `json:"id"`
			Metadata string `json:"metadata"`
		} `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode message create payload: %v", err)
	}

	var meta struct {
		UserMentions []struct {
			UserID      string `json:"user_id"`
			Name        string `json:"name"`
			MentionText string `json:"mention_text"`
		} `json:"user_mentions"`
	}
	if err := json.Unmarshal([]byte(payload.Message.Metadata), &meta); err != nil {
		t.Fatalf("failed to decode message metadata: %v", err)
	}
	if len(meta.UserMentions) != 2 {
		t.Fatalf("expected 2 deduped user mentions in metadata, got %#v", meta.UserMentions)
	}

	var mentions []domain.MessageMention
	if err := db.DB.Order("mentioned_user_id asc").Find(&mentions, "message_id = ?", payload.Message.ID).Error; err != nil {
		t.Fatalf("failed to load message mentions: %v", err)
	}
	if len(mentions) != 2 {
		t.Fatalf("expected 2 persisted message mentions, got %#v", mentions)
	}
	if mentions[0].WorkspaceID != "ws-1" || mentions[0].ChannelID != "ch-1" || mentions[0].MentionedByUserID != "user-1" || mentions[0].DMID != "" {
		t.Fatalf("unexpected first message mention row: %#v", mentions[0])
	}
}

func TestPhase65ADMMessagePersistsUserMentions(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.Workspace{ID: "ws-1", Name: "Relay"})
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com"})
	db.DB.Create(&domain.DMConversation{ID: "dm-1", CreatedAt: time.Now().UTC()})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-1"})
	db.DB.Create(&domain.DMMember{DMConversationID: "dm-1", UserID: "user-2"})

	router := gin.New()
	router.POST("/api/v1/dms/:id/messages", CreateDMMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dms/dm-1/messages", bytes.NewBufferString(`{"content":"@Jane Smith please review","user_id":"user-1"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on dm message create, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Message struct {
			ID string `json:"id"`
		} `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode dm message create payload: %v", err)
	}

	var mentions []domain.MessageMention
	if err := db.DB.Find(&mentions, "message_id = ?", payload.Message.ID).Error; err != nil {
		t.Fatalf("failed to load dm message mentions: %v", err)
	}
	if len(mentions) != 1 {
		t.Fatalf("expected 1 persisted dm message mention, got %#v", mentions)
	}
	if mentions[0].WorkspaceID != "ws-1" || mentions[0].DMID != "dm-1" || mentions[0].ChannelID != "" || mentions[0].MentionedUserID != "user-2" || mentions[0].MentionedByUserID != "user-1" {
		t.Fatalf("unexpected dm message mention row: %#v", mentions[0])
	}
}

func TestPhase65AGetMentionsUsesMessageMention(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com"})
	db.DB.Create(&domain.User{ID: "user-3", Name: "Ann Lee", Email: "ann@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-2", Content: "@Nikko Fu please review", CreatedAt: now.Add(-10 * time.Minute), Metadata: `{"user_mentions":[{"user_id":"user-1","name":"Nikko Fu","mention_text":"@Nikko Fu"}]}`})
	db.DB.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-1", UserID: "user-1", Content: "@Nikko Fu self note", CreatedAt: now.Add(-5 * time.Minute), Metadata: `{"user_mentions":[{"user_id":"user-1","name":"Nikko Fu","mention_text":"@Nikko Fu"}]}`})
	db.DB.Create(&domain.MessageMention{ID: "mm-1", MessageID: "msg-1", WorkspaceID: "ws-1", ChannelID: "ch-1", MentionedUserID: "user-1", MentionedByUserID: "user-2", MentionText: "@Nikko Fu", MentionKind: "user", CreatedAt: now.Add(-10 * time.Minute)})
	db.DB.Create(&domain.MessageMention{ID: "mm-2", MessageID: "msg-2", WorkspaceID: "ws-1", ChannelID: "ch-1", MentionedUserID: "user-1", MentionedByUserID: "user-1", MentionText: "@Nikko Fu", MentionKind: "user", CreatedAt: now.Add(-5 * time.Minute)})

	router := gin.New()
	router.GET("/api/v1/mentions", GetMentions)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mentions", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on mentions, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode mentions payload: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected exactly 1 mention item from persisted mentions, got %#v", payload.Items)
	}
	if payload.Items[0]["type"] != "mention" {
		t.Fatalf("expected mention type, got %#v", payload.Items[0])
	}
}

func TestPhase65AUnifiedFeedReturnsUserMentionRows(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	now := time.Now().UTC()
	db.DB.Create(&domain.Workspace{ID: "ws-1", Name: "Relay"})
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-user-mention", ChannelID: "ch-1", UserID: "user-2", Content: "@Nikko Fu please review", CreatedAt: now.Add(-10 * time.Minute), Metadata: `{"user_mentions":[{"user_id":"user-1","name":"Nikko Fu","mention_text":"@Nikko Fu"}]}`})
	db.DB.Create(&domain.MessageMention{ID: "mm-1", MessageID: "msg-user-mention", WorkspaceID: "ws-1", ChannelID: "ch-1", MentionedUserID: "user-1", MentionedByUserID: "user-2", MentionText: "@Nikko Fu", MentionKind: "user", CreatedAt: now.Add(-10 * time.Minute)})
	db.DB.Create(&domain.Message{ID: "msg-entity-mention", ChannelID: "ch-1", UserID: "user-2", Content: "@Launch Program is ready", CreatedAt: now.Add(-9 * time.Minute), Metadata: `{"entity_mentions":[{"entity_id":"entity-1","entity_title":"Launch Program","entity_kind":"project","source_kind":"explicit","mention_text":"@Launch Program"}]}`})

	router := gin.New()
	router.GET("/api/v1/activity/feed", GetActivityFeed)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity/feed?workspace_id=ws-1&event_type=mention&limit=10", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when loading mention feed, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []struct {
			EventType string         `json:"event_type"`
			ActorID   string         `json:"actor_id"`
			ChannelID string         `json:"channel_id"`
			Meta      map[string]any `json:"meta"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode activity feed payload: %v", err)
	}
	if len(payload.Items) < 2 {
		t.Fatalf("expected both user and entity mention rows, got %#v", payload.Items)
	}

	var sawUser, sawEntity bool
	for _, item := range payload.Items {
		if item.EventType != "mention" {
			t.Fatalf("expected only mention rows, got %#v", item)
		}
		switch item.Meta["mention_kind"] {
		case "user":
			sawUser = true
			if item.ActorID != "user-2" || item.ChannelID != "ch-1" || item.Meta["mentioned_user_id"] != "user-1" || item.Meta["message_id"] != "msg-user-mention" {
				t.Fatalf("unexpected user mention feed item: %#v", item)
			}
		case "entity":
			sawEntity = true
		}
	}
	if !sawUser || !sawEntity {
		t.Fatalf("expected both mention kinds in feed, got %#v", payload.Items)
	}
}

func TestPhase65AInboxMentionBranchUsesMessageMention(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})
	db.DB.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-2", Content: "@Nikko Fu please review", CreatedAt: now.Add(-10 * time.Minute), Metadata: `{"user_mentions":[{"user_id":"user-1","name":"Nikko Fu","mention_text":"@Nikko Fu"}]}`})
	db.DB.Create(&domain.MessageMention{ID: "mm-1", MessageID: "msg-1", WorkspaceID: "ws-1", ChannelID: "ch-1", MentionedUserID: "user-1", MentionedByUserID: "user-2", MentionText: "@Nikko Fu", MentionKind: "user", CreatedAt: now.Add(-10 * time.Minute)})

	router := gin.New()
	router.GET("/api/v1/inbox", GetInbox)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/inbox", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on inbox, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode inbox payload: %v", err)
	}
	if len(payload.Items) == 0 {
		t.Fatalf("expected inbox mention item, got %#v", payload.Items)
	}
	if payload.Items[0]["type"] != "mention" {
		t.Fatalf("expected mention inbox item, got %#v", payload.Items[0])
	}
}

func TestPhase65AMentionCreatedBroadcast(t *testing.T) {
	setupTestDB(t)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})

	router := gin.New()
	router.POST("/api/v1/messages", CreateMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBufferString(`{"channel_id":"ch-1","content":"@Jane Smith please review","user_id":"user-1"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on message create, got %d body=%s", rec.Code, rec.Body.String())
	}

	var sawMention bool
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		raw, err := client.Receive(200 * time.Millisecond)
		if err != nil {
			continue
		}
		var event realtime.Event
		if err := json.Unmarshal(raw, &event); err != nil {
			t.Fatalf("failed to decode realtime event: %v", err)
		}
		if event.Type == "mention.created" {
			sawMention = true
			payload, ok := event.Payload.(map[string]any)
			if !ok {
				t.Fatalf("expected map payload, got %#v", event.Payload)
			}
			if payload["mentioned_user_id"] != "user-2" {
				t.Fatalf("expected mentioned_user_id in payload, got %#v", payload)
			}
			break
		}
	}
	if !sawMention {
		t.Fatal("expected mention.created realtime event")
	}
}

func TestPhase63HEntityBriefAutomationLifecycle(t *testing.T) {
	setupTestDB(t)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	now := time.Now().UTC()
	db.DB.Create(&domain.Workspace{ID: "ws-1", Name: "Relay"})
	db.DB.Create(&domain.KnowledgeEntity{
		ID:          "entity-1",
		WorkspaceID: "ws-1",
		Kind:        "project",
		Title:       "Launch Program",
		Status:      "active",
		SourceKind:  "manual",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	db.DB.Create(&domain.AISummary{
		ScopeType: "knowledge_entity",
		ScopeID:   "entity-1",
		ChannelID: "",
		Provider:  "openrouter",
		Model:     "nemotron",
		Content:   "Existing cached brief",
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	})

	router := gin.New()
	router.GET("/api/v1/knowledge/entities/:id/brief/automation", GetKnowledgeEntityBriefAutomation)
	router.POST("/api/v1/knowledge/entities/:id/brief/automation/run", RunKnowledgeEntityBriefAutomation)
	router.POST("/api/v1/knowledge/entities/:id/brief/automation/retry", RetryKnowledgeEntityBriefAutomation)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/brief/automation/run", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 when queueing entity brief automation, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.entity.brief.regen.queued")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/entities/entity-1/brief/automation", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when loading entity brief automation state, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Job domain.AIAutomationJob `json:"job"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode entity brief automation payload: %v", err)
	}
	if payload.Job.ScopeID != "entity-1" || payload.Job.Status != "pending" || payload.Job.JobType != "entity_brief_regen" {
		t.Fatalf("unexpected entity brief automation job: %#v", payload.Job)
	}

	if err := db.DB.Model(&domain.AIAutomationJob{}).Where("id = ?", payload.Job.ID).Updates(map[string]any{
		"status":      "failed",
		"last_error":  "provider timeout",
		"finished_at": now,
	}).Error; err != nil {
		t.Fatalf("mark automation job failed: %v", err)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/entity-1/brief/automation/retry", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 when retrying entity brief automation, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.entity.brief.regen.queued")
}

func TestPhase63HScheduleBookingLifecycle(t *testing.T) {
	setupTestDB(t)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.User{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.AIComposeActivity{
		ID:              "compose-activity-1",
		ComposeID:       "compose-1",
		WorkspaceID:     "ws-1",
		ChannelID:       "ch-1",
		UserID:          "user-1",
		Intent:          "schedule",
		SuggestionCount: 1,
		Provider:        "openrouter",
		Model:           "nemotron",
		CreatedAt:       now.Add(-15 * time.Minute),
	})

	router := gin.New()
	router.POST("/api/v1/ai/schedule/book", BookAISchedule)
	router.GET("/api/v1/ai/schedule/bookings", ListAIScheduleBookings)
	router.GET("/api/v1/ai/schedule/bookings/:id", GetAIScheduleBooking)
	router.POST("/api/v1/ai/schedule/bookings/:id/cancel", CancelAIScheduleBooking)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/schedule/book", strings.NewReader(`{"compose_id":"compose-1","channel_id":"ch-1","title":"Launch standup","description":"Review release blockers","provider":"internal","slot":{"starts_at":"2026-04-23T09:00:00Z","ends_at":"2026-04-23T09:30:00Z","timezone":"UTC","attendee_ids":["user-1","user-2"]}}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 when booking schedule slot, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "schedule.event.booked")

	var createPayload struct {
		Booking struct {
			ID          string `json:"id"`
			RequestedBy string `json:"requested_by"`
			Status      string `json:"status"`
			Provider    string `json:"provider"`
			ICSContent  string `json:"ics_content"`
		} `json:"booking"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("decode schedule booking payload: %v", err)
	}
	if createPayload.Booking.ID == "" || createPayload.Booking.Status != "booked" || createPayload.Booking.RequestedBy != "user-1" || createPayload.Booking.ICSContent == "" {
		t.Fatalf("unexpected schedule booking payload: %#v", createPayload.Booking)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/ai/schedule/bookings?channel_id=ch-1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when listing schedule bookings, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/ai/schedule/bookings/"+createPayload.Booking.ID, nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when fetching schedule booking detail, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/schedule/bookings/"+createPayload.Booking.ID+"/cancel", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when cancelling schedule booking, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "schedule.event.cancelled")
}

func TestMatchKnowledgeEntitiesInTextEndpoint(t *testing.T) {
	setupTestDB(t)
	now := time.Now().UTC()

	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "doc", Title: "Launch", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})

	router := gin.New()
	router.POST("/api/v1/knowledge/entities/match-text", MatchKnowledgeEntitiesInText)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/match-text", strings.NewReader(`{"workspace_id":"ws-1","text":"Launch Program needs a decision","limit":5}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 match text, got %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Matches []knowledge.EntityTextMatch `json:"matches"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode match payload: %v", err)
	}
	if len(payload.Matches) != 1 || payload.Matches[0].EntityID != "entity-1" || payload.Matches[0].MatchedText != "Launch Program" {
		t.Fatalf("expected longest entity match, got %#v", payload.Matches)
	}
}

type knowledgeGraphNodePayload struct {
	Kind    string `json:"kind"`
	RefKind string `json:"ref_kind"`
	RefID   string `json:"ref_id"`
	Role    string `json:"role"`
}

type knowledgeGraphEdgePayload struct {
	Relation  string  `json:"relation"`
	Weight    float64 `json:"weight"`
	Direction string  `json:"direction"`
}

func graphHasRefNode(nodes []knowledgeGraphNodePayload, kind, refID, role string) bool {
	for _, node := range nodes {
		if node.Kind == kind && node.RefKind == kind && node.RefID == refID && node.Role == role {
			return true
		}
	}
	return false
}

func graphHasWeightedEdge(edges []knowledgeGraphEdgePayload, relation string, weight float64, direction string) bool {
	for _, edge := range edges {
		if edge.Relation == relation && edge.Weight == weight && edge.Direction == direction {
			return true
		}
	}
	return false
}

func TestKnowledgeEntityEndpointsBroadcastRealtimeEvents(t *testing.T) {
	setupTestDB(t)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	router := gin.New()
	router.POST("/api/v1/knowledge/entities", CreateKnowledgeEntity)
	router.PATCH("/api/v1/knowledge/entities/:id", UpdateKnowledgeEntity)
	router.POST("/api/v1/knowledge/entities/:id/refs", AddKnowledgeEntityRef)
	router.POST("/api/v1/knowledge/entities/:id/events", AddKnowledgeEntityEvent)
	router.POST("/api/v1/knowledge/links", AddKnowledgeEntityLink)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities", strings.NewReader(`{"workspace_id":"ws-1","kind":"project","title":"Launch Program"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on entity create, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.entity.created")

	var createPayload struct {
		Entity domain.KnowledgeEntity `json:"entity"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("decode entity create: %v", err)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/knowledge/entities/"+createPayload.Entity.ID, strings.NewReader(`{"summary":"Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on entity update, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.entity.updated")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/"+createPayload.Entity.ID+"/refs", strings.NewReader(`{"ref_kind":"message","ref_id":"msg-1","role":"discussion"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on entity ref, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.entity.ref.created")
	assertRealtimeEventType(t, client, "knowledge.trending.changed")

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities/"+createPayload.Entity.ID+"/events", strings.NewReader(`{"event_type":"live_update","title":"Live update"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on entity event, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.event.created")

	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "service", Title: "Billing Service", Status: "active", SourceKind: "manual", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/links", strings.NewReader(`{"from_entity_id":"`+createPayload.Entity.ID+`","to_entity_id":"entity-2","relation":"depends_on","weight":2}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on entity link, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.link.created")
}

func TestKnowledgeEventIngestCreatesTimelineEventAndBroadcasts(t *testing.T) {
	setupTestDB(t)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(4)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	now := time.Now().UTC()
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "customer", Title: "ACME", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})

	router := gin.New()
	router.POST("/api/v1/knowledge/events/ingest", IngestKnowledgeEvent)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/events/ingest", strings.NewReader(`{"entity_id":"entity-1","event_type":"live_update","title":"CRM account updated","body":"ACME moved to renewal risk","source_kind":"live","source_ref":"crm:account:acme"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on ingest, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "knowledge.event.created")
}

func TestCreateMessageAutoLinksMentionedKnowledgeEntity(t *testing.T) {
	setupTestDB(t)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})

	router := gin.New()
	router.POST("/api/v1/messages", CreateMessage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", strings.NewReader(`{"channel_id":"ch-1","user_id":"user-1","content":"Launch Program is ready for review"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on message create, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "message.created")
	assertRealtimeEventType(t, client, "knowledge.entity.ref.created")

	var count int64
	db.DB.Model(&domain.KnowledgeEntityRef{}).Where("entity_id = ? AND ref_kind = ? AND role = ?", "entity-1", "message", "discussion").Count(&count)
	if count != 1 {
		t.Fatalf("expected one auto-linked message ref, got %d", count)
	}
}

func TestUploadFileAutoLinksMentionedKnowledgeEntity(t *testing.T) {
	setupTestDB(t)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	now := time.Now().UTC()
	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	db.DB.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})

	_ = os.RemoveAll("uploads")
	t.Cleanup(func() {
		_ = os.RemoveAll("uploads")
	})

	router := gin.New()
	router.POST("/api/v1/files/upload", UploadFile)

	body, contentType := buildMultipartFileUpload(t, "channel_id", "ch-1", "launch-program.txt", []byte("Launch Program file notes"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
	req.Header.Set("Content-Type", contentType)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on file upload, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertRealtimeEventType(t, client, "file.extraction.updated")
	assertRealtimeEventType(t, client, "knowledge.entity.ref.created")

	var count int64
	db.DB.Model(&domain.KnowledgeEntityRef{}).Where("entity_id = ? AND ref_kind = ? AND role = ?", "entity-1", "file", "evidence").Count(&count)
	if count != 1 {
		t.Fatalf("expected one auto-linked file ref, got %d", count)
	}
}

func TestFileExtractionBroadcastsRealtimeUpdate(t *testing.T) {
	setupTestDB(t)

	hub := realtime.NewHub()
	go hub.Run()
	SetRealtimeHub(hub)
	defer SetRealtimeHub(nil)

	client := realtime.NewTestClient(8)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	db.DB.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public"})

	_ = os.RemoveAll("uploads")
	t.Cleanup(func() {
		_ = os.RemoveAll("uploads")
	})

	router := gin.New()
	router.POST("/api/v1/files/upload", UploadFile)

	body, contentType := buildMultipartFileUpload(t, "channel_id", "ch-1", "notes.txt", []byte("Launch checklist\n\nShip file search"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
	req.Header.Set("Content-Type", contentType)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on upload, got %d body=%s", rec.Code, rec.Body.String())
	}

	assertRealtimeEventType(t, client, "file.extraction.updated")
}

func buildMultipartFileUpload(t *testing.T, fieldName, fieldValue, filename string, contents []byte) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if fieldName != "" {
		_ = writer.WriteField(fieldName, fieldValue)
	}
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("failed to create multipart file: %v", err)
	}
	if _, err := fileWriter.Write(contents); err != nil {
		t.Fatalf("failed to write multipart file contents: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}
	return body, writer.FormDataContentType()
}

func setupTestDB(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	RealtimeHub = nil
	dsn := fmt.Sprintf("file:test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	testDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite test db: %v", err)
	}
	db.DB = testDB
	if err := db.DB.AutoMigrate(&domain.Organization{}, &domain.Team{}, &domain.User{}, &domain.Agent{}, &domain.Workspace{}, &domain.WorkspaceInvite{}, &domain.UserGroup{}, &domain.UserGroupMember{}, &domain.WorkflowDefinition{}, &domain.WorkflowRunStep{}, &domain.WorkflowRunLog{}, &domain.ToolDefinition{}, &domain.ToolRun{}, &domain.ToolRunLog{}, &domain.Channel{}, &domain.ChannelMember{}, &domain.ChannelPreference{}, &domain.WorkspaceList{}, &domain.WorkspaceListItem{}, &domain.Message{}, &domain.MessageMention{}, &domain.NotificationItem{}, &domain.NotificationRead{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	if err := db.DB.AutoMigrate(&domain.MessageReaction{}, &domain.SavedMessage{}, &domain.Draft{}, &domain.UnreadMarker{}, &domain.NotificationRead{}, &domain.NotificationPreference{}, &domain.NotificationMuteRule{}, &domain.AIFeedback{}, &domain.AIComposeFeedback{}, &domain.AIComposeActivity{}, &domain.AIAutomationJob{}, &domain.AIScheduleBooking{}, &domain.AIConversation{}, &domain.AIConversationMessage{}, &domain.AISummary{}, &domain.ChannelAutoSummarySetting{}, &domain.KnowledgeEntityAskAnswer{}, &domain.Artifact{}, &domain.ArtifactVersion{}, &domain.FileAsset{}, &domain.FileAssetEvent{}, &domain.FileExtraction{}, &domain.FileExtractionChunk{}, &domain.FileComment{}, &domain.FileShare{}, &domain.StarredFile{}, &domain.MessageArtifactReference{}, &domain.MessageFileAttachment{}, &domain.KnowledgeEvidenceLink{}, &domain.KnowledgeEvidenceEntityRef{}, &domain.KnowledgeEntity{}, &domain.KnowledgeEntityRef{}, &domain.KnowledgeEntityLink{}, &domain.KnowledgeEvent{}, &domain.KnowledgeEntityFollow{}, &domain.KnowledgeDigestSchedule{}, &domain.DMConversation{}, &domain.DMMember{}, &domain.DMMessage{}, &domain.WorkflowRun{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
}

func TestChannelExecutionHomeIncludesExecutionBlocks(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "ops", Type: "public"})
	
	// Add some data to be aggregated
	db.DB.Create(&domain.WorkspaceList{ID: "list-1", WorkspaceID: "ws-1", ChannelID: "ch-1", CreatedBy: "user-1", Title: "Tasks"})
	db.DB.Create(&domain.WorkspaceListItem{ListID: "list-1", Content: "Open Task", AssignedTo: "user-1", IsCompleted: false})
	db.DB.Create(&domain.ToolDefinition{ID: "tool-1", Name: "Tester", IsEnabled: true})
	db.DB.Create(&domain.ToolRun{ID: "run-1", ToolID: "tool-1", Status: "failed", TriggeredBy: "user-1", Summary: "Failed Run", CreatedAt: time.Now()})

	router := gin.New()
	router.GET("/api/v1/home", GetHome)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/home", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Home struct {
			OpenListWork             []any `json:"open_list_work"`
			ToolRunsNeedingAttention []any `json:"tool_runs_needing_attention"`
			ChannelExecutionPulse    []any `json:"channel_execution_pulse"`
		} `json:"home"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.Home.OpenListWork == nil {
		t.Fatal("expected open_list_work block in home response")
	}
	if payload.Home.ToolRunsNeedingAttention == nil {
		t.Fatal("expected tool_runs_needing_attention block in home response")
	}
	if payload.Home.ChannelExecutionPulse == nil {
		t.Fatal("expected channel_execution_pulse block in home response")
	}
}

func TestChannelExecutionHomeSummaryIncludesOverdue(t *testing.T) {
	setupTestDB(t)

	db.DB.Create(&domain.User{ID: "user-1", OrganizationID: "org-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	db.DB.Create(&domain.Workspace{ID: "ws-1", OrganizationID: "org-1", Name: "Relay"})
	db.DB.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "ops", Type: "public"})
	
	db.DB.Create(&domain.WorkspaceList{ID: "list-1", WorkspaceID: "ws-1", ChannelID: "ch-1", CreatedBy: "user-1", Title: "Tasks"})
	
	overdue := time.Now().Add(-24 * time.Hour)
	db.DB.Create(&domain.WorkspaceListItem{ListID: "list-1", Content: "Overdue Task", AssignedTo: "user-1", IsCompleted: false, DueAt: &overdue})

	router := gin.New()
	router.GET("/api/v1/home", GetHome)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/home", nil)
	router.ServeHTTP(rec, req)

	var payload struct {
		Home struct {
			ChannelExecutionPulse []struct {
				Summary string `json:"summary"`
			} `json:"channel_execution_pulse"`
		} `json:"home"`
	}
	json.Unmarshal(rec.Body.Bytes(), &payload)

	found := false
	for _, pulse := range payload.Home.ChannelExecutionPulse {
		if strings.Contains(pulse.Summary, "overdue") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected summary to contain 'overdue'")
	}
}

func TestPhase65CMentionPagination(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	user := domain.User{ID: "user-1", Name: "Nikko", OrganizationID: "org-1", Email: "nikko@example.com"}
	db.DB.Create(&user)
	db.DB.Create(&domain.User{ID: "user-2", Name: "Actor", OrganizationID: "org-1", Email: "actor@example.com"})
	
	// Create 5 mentions
	for i := 1; i <= 5; i++ {
		db.DB.Create(&domain.MessageMention{
			ID:                 fmt.Sprintf("mention-%d", i),
			MessageID:          fmt.Sprintf("msg-%d", i),
			MentionedUserID:    "user-1",
			MentionedByUserID:  "user-2",
			MentionKind:        "user",
			CreatedAt:          time.Now().Add(time.Duration(i) * time.Minute),
		})
	}

	router := gin.New()
	router.GET("/api/v1/mentions", GetMentions)

	// Fetch with limit 2
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mentions?limit=2", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Items []any  `json:"items"`
		Next  string `json:"next_cursor"`
	}
	json.Unmarshal(rec.Body.Bytes(), &payload)

	if len(payload.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(payload.Items))
	}
	if payload.Next == "" {
		t.Fatal("expected next_cursor")
	}
}

func TestPhase65CInboxUsesNotificationItem(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	user := domain.User{ID: "user-1", Name: "Nikko", OrganizationID: "org-1", Email: "nikko@example.com"}
	db.DB.Create(&user)
	db.DB.Create(&domain.User{ID: "user-2", Name: "Actor", OrganizationID: "org-1", Email: "actor@example.com"})
	
	// Pre-populate a NotificationItem
	db.DB.Create(&domain.NotificationItem{
		ID:         "notify-1",
		UserID:     "user-1",
		Type:       "mention",
		ActorID:    "user-2",
		Summary:    "Someone mentioned you",
		OccurredAt: time.Now(),
	})

	router := gin.New()
	router.GET("/api/v1/inbox", GetInbox)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/inbox", nil)
	router.ServeHTTP(rec, req)

	var payload struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	json.Unmarshal(rec.Body.Bytes(), &payload)

	found := false
	for _, item := range payload.Items {
		if item.ID == "notify-1" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected notify-1 in inbox")
	}
}

func TestPhase65CHomeIncludesUnreadMentionCount(t *testing.T) {
	setupTestDB(t)
	gin.SetMode(gin.TestMode)

	user := domain.User{ID: "user-1", Name: "Nikko", OrganizationID: "org-1", Email: "nikko@example.com"}
	db.DB.Create(&user)
	db.DB.Create(&domain.User{ID: "user-2", Name: "Actor", OrganizationID: "org-1", Email: "actor@example.com"})
	
	// Create 3 unread mentions
	for i := 1; i <= 3; i++ {
		db.DB.Create(&domain.MessageMention{
			ID:                 fmt.Sprintf("mention-%d", i),
			MessageID:          fmt.Sprintf("msg-%d", i),
			MentionedUserID:    "user-1",
			MentionedByUserID:  "user-2",
			MentionKind:        "user",
			CreatedAt:          time.Now(),
		})
	}

	router := gin.New()
	router.GET("/api/v1/home", GetHome)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/home", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Home struct {
			Activity struct {
				UnreadMentionCount int `json:"unread_mention_count"`
			} `json:"activity"`
		} `json:"home"`
	}
	json.Unmarshal(rec.Body.Bytes(), &payload)

	if payload.Home.Activity.UnreadMentionCount != 3 {
		t.Fatalf("expected 3 unread mentions, got %d", payload.Home.Activity.UnreadMentionCount)
	}
}
