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

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/settings", bytes.NewBufferString(`{"provider":"gemini","model":"gemini-3-flash-preview","mode":"planning"}`))
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
	if refreshed.AIProvider != "gemini" || refreshed.AIModel != "gemini-3-flash-preview" || refreshed.AIMode != "planning" {
		t.Fatalf("settings were not persisted: %#v", refreshed)
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
	mention := domain.Message{ID: "msg-mention", ChannelID: "ch-1", UserID: "user-2", Content: "Looping in Nikko Fu for review", CreatedAt: now.Add(-30 * time.Minute), Metadata: "{}"}
	db.DB.Create(&parent)
	db.DB.Create(&reply)
	db.DB.Create(&mention)
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
	mention := domain.Message{ID: "msg-mention", ChannelID: "ch-1", UserID: "user-2", Content: "Looping in Nikko Fu for review", CreatedAt: now.Add(-30 * time.Minute), Metadata: "{}"}
	db.DB.Create(&parent)
	db.DB.Create(&reply)
	db.DB.Create(&mention)
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
	db.DB.Create(&domain.Message{ID: "msg-mention", ChannelID: "ch-1", UserID: "user-2", Content: "Looping in Nikko Fu for review", CreatedAt: now.Add(-10 * time.Minute), Metadata: "{}"})
	db.DB.Create(&domain.Message{ID: "msg-other", ChannelID: "ch-1", UserID: "user-2", Content: "No direct mention here", CreatedAt: now.Add(-5 * time.Minute), Metadata: "{}"})

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
		Content:   "Looping in Nikko Fu for review",
		CreatedAt: now,
		Metadata:  "{}",
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
	if err := db.DB.AutoMigrate(&domain.Organization{}, &domain.Team{}, &domain.User{}, &domain.Agent{}, &domain.Workspace{}, &domain.WorkspaceInvite{}, &domain.UserGroup{}, &domain.UserGroupMember{}, &domain.WorkflowDefinition{}, &domain.WorkflowRunStep{}, &domain.WorkflowRunLog{}, &domain.ToolDefinition{}, &domain.ToolRun{}, &domain.ToolRunLog{}, &domain.Channel{}, &domain.ChannelMember{}, &domain.ChannelPreference{}, &domain.WorkspaceList{}, &domain.WorkspaceListItem{}, &domain.Message{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	if err := db.DB.AutoMigrate(&domain.MessageReaction{}, &domain.SavedMessage{}, &domain.Draft{}, &domain.UnreadMarker{}, &domain.NotificationRead{}, &domain.NotificationPreference{}, &domain.NotificationMuteRule{}, &domain.AIFeedback{}, &domain.AIConversation{}, &domain.AIConversationMessage{}, &domain.AISummary{}, &domain.Artifact{}, &domain.ArtifactVersion{}, &domain.FileAsset{}, &domain.FileAssetEvent{}, &domain.FileExtraction{}, &domain.FileExtractionChunk{}, &domain.FileComment{}, &domain.FileShare{}, &domain.StarredFile{}, &domain.MessageArtifactReference{}, &domain.MessageFileAttachment{}, &domain.KnowledgeEvidenceLink{}, &domain.KnowledgeEvidenceEntityRef{}, &domain.KnowledgeEntity{}, &domain.KnowledgeEntityRef{}, &domain.KnowledgeEntityLink{}, &domain.KnowledgeEvent{}, &domain.DMConversation{}, &domain.DMMember{}, &domain.DMMessage{}, &domain.WorkflowRun{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
}
