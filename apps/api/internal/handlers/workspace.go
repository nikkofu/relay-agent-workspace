package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

type channelSummaryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	MemberCount int    `json:"member_count"`
	IsStarred   bool   `json:"is_starred"`
}

type artifactSummaryResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

type userProfileSummary struct {
	LocalTime       string                    `json:"local_time"`
	WorkingHours    string                    `json:"working_hours"`
	FocusAreas      []string                  `json:"focus_areas"`
	TopChannels     []channelSummaryResponse  `json:"top_channels"`
	RecentArtifacts []artifactSummaryResponse `json:"recent_artifacts"`
}

type homeActivitySummary struct {
	UnreadCount   int `json:"unread_count"`
	DraftCount    int `json:"draft_count"`
	DMCount       int `json:"dm_count"`
	StarredCount  int `json:"starred_count"`
	GroupCount    int `json:"group_count"`
	WorkflowCount int `json:"workflow_count"`
}

type homeStatsResponse struct {
	PendingActions int `json:"pending_actions"`
	ActiveThreads  int `json:"active_threads"`
}

type homeRecentActivityResponse struct {
	ID          string    `json:"id"`
	ChannelID   string    `json:"channel_id"`
	ChannelName string    `json:"channel_name"`
	LastMessage string    `json:"last_message"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type homeDraftResponse struct {
	Scope     string    `json:"scope"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
}

type homeDMResponse struct {
	ID            string      `json:"id"`
	User          domain.User `json:"user"`
	UserIDs       []string    `json:"user_ids"`
	LastMessage   string      `json:"last_message"`
	LastMessageAt time.Time   `json:"last_message_at"`
}

type userGroupListResponse struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	Handle      string    `json:"handle"`
	Description string    `json:"description"`
	MemberCount int       `json:"member_count"`
	CreatedBy   string    `json:"created_by"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type userGroupMemberResponse struct {
	Role      string      `json:"role"`
	CreatedAt time.Time   `json:"created_at"`
	User      domain.User `json:"user"`
}

type userGroupDetailResponse struct {
	ID          string                    `json:"id"`
	WorkspaceID string                    `json:"workspace_id"`
	Name        string                    `json:"name"`
	Handle      string                    `json:"handle"`
	Description string                    `json:"description"`
	MemberCount int                       `json:"member_count"`
	CreatedBy   string                    `json:"created_by"`
	UpdatedAt   time.Time                 `json:"updated_at"`
	Members     []userGroupMemberResponse `json:"members"`
}

type notificationPreferenceResponse struct {
	InboxEnabled    bool                          `json:"inbox_enabled"`
	MentionsEnabled bool                          `json:"mentions_enabled"`
	DMEnabled       bool                          `json:"dm_enabled"`
	MuteAll         bool                          `json:"mute_all"`
	MuteRules       []domain.NotificationMuteRule `json:"mute_rules"`
}

type workflowRunResponse struct {
	ID           string                    `json:"id"`
	Status       string                    `json:"status"`
	Summary      string                    `json:"summary"`
	Input        map[string]any            `json:"input,omitempty"`
	RetryOfRunID string                    `json:"retry_of_run_id,omitempty"`
	WorkflowID   string                    `json:"workflow_id"`
	WorkflowName string                    `json:"workflow_name"`
	TriggeredBy  string                    `json:"triggered_by"`
	FinishedAt   *time.Time                `json:"finished_at,omitempty"`
	DurationMS   int                       `json:"duration_ms"`
	Error        string                    `json:"error,omitempty"`
	Steps        []workflowRunStepResponse `json:"steps,omitempty"`
	StartedAt    time.Time                 `json:"started_at"`
	CompletedAt  *time.Time                `json:"completed_at,omitempty"`
	Workflow     domain.WorkflowDefinition `json:"workflow"`
	StartedBy    domain.User               `json:"started_by"`
}

type workflowRunStepResponse struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	DurationMS int    `json:"duration_ms"`
	Detail     string `json:"detail,omitempty"`
}

type workflowRunLogResponse struct {
	ID        uint           `json:"id"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

func GetUserProfile(c *gin.Context) {
	var user domain.User
	if err := db.DB.First(&user, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	enriched := enrichUser(user)
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":                enriched.ID,
			"org_id":            enriched.OrganizationID,
			"name":              enriched.Name,
			"email":             enriched.Email,
			"avatar":            enriched.Avatar,
			"title":             enriched.Title,
			"department":        enriched.Department,
			"timezone":          defaultString(enriched.Timezone, "UTC"),
			"working_hours":     defaultString(enriched.WorkingHours, "Mon - Fri"),
			"pronouns":          enriched.Pronouns,
			"location":          enriched.Location,
			"phone":             enriched.Phone,
			"bio":               enriched.Bio,
			"status":            enriched.Status,
			"status_text":       enriched.StatusText,
			"status_emoji":      enriched.StatusEmoji,
			"status_expires_at": enriched.StatusExpiresAt,
			"last_seen_at":      enriched.LastSeenAt,
			"ai_provider":       enriched.AIProvider,
			"ai_model":          enriched.AIModel,
			"ai_mode":           enriched.AIMode,
			"ai_insight":        enriched.AIInsight,
			"profile":           buildUserProfileSummary(enriched),
		},
	})
}

func PatchUserProfile(c *gin.Context) {
	var user domain.User
	if err := db.DB.First(&user, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		Title        *string `json:"title"`
		Department   *string `json:"department"`
		Timezone     *string `json:"timezone"`
		WorkingHours *string `json:"working_hours"`
		Pronouns     *string `json:"pronouns"`
		Location     *string `json:"location"`
		Phone        *string `json:"phone"`
		Bio          *string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Title != nil {
		user.Title = strings.TrimSpace(*input.Title)
	}
	if input.Department != nil {
		user.Department = strings.TrimSpace(*input.Department)
	}
	if input.Timezone != nil {
		user.Timezone = strings.TrimSpace(*input.Timezone)
	}
	if input.WorkingHours != nil {
		user.WorkingHours = strings.TrimSpace(*input.WorkingHours)
	}
	if input.Pronouns != nil {
		user.Pronouns = strings.TrimSpace(*input.Pronouns)
	}
	if input.Location != nil {
		user.Location = strings.TrimSpace(*input.Location)
	}
	if input.Phone != nil {
		user.Phone = strings.TrimSpace(*input.Phone)
	}
	if input.Bio != nil {
		user.Bio = strings.TrimSpace(*input.Bio)
	}

	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": enrichUser(user)})
}

func PatchUserStatus(c *gin.Context) {
	var user domain.User
	if err := db.DB.First(&user, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		Status           string `json:"status"`
		StatusText       string `json:"status_text"`
		StatusEmoji      string `json:"status_emoji"`
		ExpiresInMinutes int    `json:"expires_in_minutes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	now := time.Now().UTC()
	expiresAt := now.Add(5 * time.Minute)
	user.Status = input.Status
	user.StatusText = input.StatusText
	user.StatusEmoji = strings.TrimSpace(input.StatusEmoji)
	user.LastSeenAt = &now
	user.PresenceExpiresAt = &expiresAt
	if input.ExpiresInMinutes > 0 {
		statusExpiresAt := now.Add(time.Duration(input.ExpiresInMinutes) * time.Minute)
		user.StatusExpiresAt = &statusExpiresAt
	} else {
		user.StatusExpiresAt = nil
	}
	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user status"})
		return
	}

	enriched := enrichUser(user)
	if RealtimeHub != nil {
		_ = RealtimeHub.Broadcast(realtime.Event{
			ID:          "evt_" + time.Now().Format("20060102150405.000000"),
			Type:        "presence.updated",
			WorkspaceID: primaryWorkspaceID(),
			EntityID:    enriched.ID,
			TS:          time.Now().UTC().Format(time.RFC3339Nano),
			Payload: gin.H{
				"user_id":           enriched.ID,
				"status":            enriched.Status,
				"status_text":       enriched.StatusText,
				"status_emoji":      enriched.StatusEmoji,
				"status_expires_at": enriched.StatusExpiresAt,
				"last_seen_at":      enriched.LastSeenAt,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{"user": enriched})
}

func GetUserGroupMembers(c *gin.Context) {
	var group domain.UserGroup
	if err := db.DB.First(&group, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user group not found"})
		return
	}

	var memberships []domain.UserGroupMember
	if err := db.DB.Where("user_group_id = ?", group.ID).Order("role asc, created_at asc").Find(&memberships).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user group members"})
		return
	}

	members := make([]userGroupMemberResponse, 0, len(memberships))
	for _, membership := range memberships {
		var user domain.User
		if err := db.DB.First(&user, "id = ?", membership.UserID).Error; err != nil {
			continue
		}
		members = append(members, userGroupMemberResponse{
			Role:      membership.Role,
			CreatedAt: membership.CreatedAt,
			User:      enrichUser(user),
		})
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

func AddUserGroupMember(c *gin.Context) {
	var group domain.UserGroup
	if err := db.DB.First(&group, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user group not found"})
		return
	}

	var input struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(input.UserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	role := defaultString(strings.TrimSpace(input.Role), "member")
	member := domain.UserGroupMember{
		UserGroupID: group.ID,
		UserID:      strings.TrimSpace(input.UserID),
		Role:        role,
		CreatedAt:   time.Now().UTC(),
	}
	if err := db.DB.Where("user_group_id = ? AND user_id = ?", group.ID, member.UserID).Delete(&domain.UserGroupMember{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to replace user group member"})
		return
	}
	if err := db.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add user group member"})
		return
	}

	var user domain.User
	if err := db.DB.First(&user, "id = ?", member.UserID).Error; err != nil {
		c.JSON(http.StatusCreated, gin.H{"member": gin.H{"role": member.Role, "created_at": member.CreatedAt}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"member": userGroupMemberResponse{
		Role:      member.Role,
		CreatedAt: member.CreatedAt,
		User:      enrichUser(user),
	}})
}

func RemoveUserGroupMember(c *gin.Context) {
	groupID := c.Param("id")
	userID := c.Param("userId")
	if err := db.DB.Where("user_group_id = ? AND user_id = ?", groupID, userID).Delete(&domain.UserGroupMember{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove user group member"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true, "group_id": groupID, "user_id": userID})
}

func SearchUserGroupMentions(c *gin.Context) {
	queryText := strings.TrimSpace(strings.ToLower(c.Query("q")))
	query := db.DB.Model(&domain.UserGroup{}).Order("updated_at desc, handle asc")
	if queryText != "" {
		like := "%" + queryText + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(handle) LIKE ? OR LOWER(description) LIKE ?", like, like, like)
	}

	var groups []domain.UserGroup
	if err := query.Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search user groups"})
		return
	}

	items := make([]userGroupListResponse, 0, len(groups))
	for _, group := range groups {
		var memberCount int64
		db.DB.Model(&domain.UserGroupMember{}).Where("user_group_id = ?", group.ID).Count(&memberCount)
		items = append(items, userGroupListResponse{
			ID:          group.ID,
			WorkspaceID: group.WorkspaceID,
			Name:        group.Name,
			Handle:      group.Handle,
			Description: group.Description,
			MemberCount: int(memberCount),
			CreatedBy:   group.CreatedBy,
			UpdatedAt:   group.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"groups": items})
}

func GetHome(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	currentUser = enrichUser(currentUser)

	starredChannels := listStarredChannels()
	recentDMs := listRecentDMs(currentUser.ID, 5)
	drafts := listRecentDrafts(currentUser.ID, 5)
	tools := listEnabledTools()
	workflows := listActiveWorkflows()

	var groupCount int64
	db.DB.Model(&domain.UserGroupMember{}).Where("user_id = ?", currentUser.ID).Count(&groupCount)

	activity := homeActivitySummary{
		UnreadCount:   sumUnreadChannels(),
		DraftCount:    len(drafts),
		DMCount:       len(recentDMs),
		StarredCount:  len(starredChannels),
		GroupCount:    int(groupCount),
		WorkflowCount: len(workflows),
	}
	stats := homeStatsResponse{
		PendingActions: activity.UnreadCount,
		ActiveThreads:  countActiveThreadsForHome(currentUser.ID),
	}
	recentArtifacts := buildUserProfileSummary(currentUser).RecentArtifacts

	c.JSON(http.StatusOK, gin.H{
		"home": gin.H{
			"user":             currentUser,
			"profile":          buildUserProfileSummary(currentUser),
			"activity":         activity,
			"stats":            stats,
			"starred_channels": starredChannels,
			"recent_dms":       recentDMs,
			"drafts":           drafts,
			"tools":            tools,
			"workflows":        workflows,
			"recent_activity":  listRecentChannelActivity(currentUser.ID, 6),
			"recent_artifacts": recentArtifacts,
			"recent_lists":     listRecentWorkspaceLists(currentUser.ID, 5),
			"recent_tool_runs": listRecentToolRunsForHome(currentUser.ID, 5),
			"recent_files":     listRecentFilesForHome(currentUser.ID, 5),
		},
	})
}

func GetUserGroups(c *gin.Context) {
	query := db.DB.Model(&domain.UserGroup{})
	if workspaceID := c.Query("workspace_id"); workspaceID != "" {
		query = query.Where("workspace_id = ?", workspaceID)
	}

	var groups []domain.UserGroup
	if err := query.Order("updated_at desc, name asc").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user groups"})
		return
	}

	items := make([]userGroupListResponse, 0, len(groups))
	for _, group := range groups {
		var memberCount int64
		db.DB.Model(&domain.UserGroupMember{}).Where("user_group_id = ?", group.ID).Count(&memberCount)
		items = append(items, userGroupListResponse{
			ID:          group.ID,
			WorkspaceID: group.WorkspaceID,
			Name:        group.Name,
			Handle:      group.Handle,
			Description: group.Description,
			MemberCount: int(memberCount),
			CreatedBy:   group.CreatedBy,
			UpdatedAt:   group.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"groups": items})
}

func CreateUserGroup(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		WorkspaceID string   `json:"workspace_id"`
		Name        string   `json:"name"`
		Handle      string   `json:"handle"`
		Description string   `json:"description"`
		MemberIDs   []string `json:"member_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(input.WorkspaceID) == "" || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Handle) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id, name, and handle are required"})
		return
	}

	now := time.Now().UTC()
	group := domain.UserGroup{
		ID:          ids.NewPrefixedUUID("group"),
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Handle:      input.Handle,
		Description: input.Description,
		CreatedBy:   currentUser.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user group"})
		return
	}
	memberIDs := normalizeMemberIDs(input.MemberIDs, currentUser.ID)
	if err := replaceUserGroupMembers(group.ID, memberIDs, currentUser.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user group members"})
		return
	}
	respondWithUserGroupDetail(c, http.StatusCreated, group)
}

func GetUserGroup(c *gin.Context) {
	var group domain.UserGroup
	if err := db.DB.First(&group, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user group not found"})
		return
	}

	var memberships []domain.UserGroupMember
	if err := db.DB.Where("user_group_id = ?", group.ID).Order("role asc, created_at asc").Find(&memberships).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user group members"})
		return
	}

	members := make([]userGroupMemberResponse, 0, len(memberships))
	for _, membership := range memberships {
		var user domain.User
		if err := db.DB.First(&user, "id = ?", membership.UserID).Error; err != nil {
			continue
		}
		members = append(members, userGroupMemberResponse{
			Role:      membership.Role,
			CreatedAt: membership.CreatedAt,
			User:      enrichUser(user),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"group": userGroupDetailResponse{
			ID:          group.ID,
			WorkspaceID: group.WorkspaceID,
			Name:        group.Name,
			Handle:      group.Handle,
			Description: group.Description,
			MemberCount: len(members),
			CreatedBy:   group.CreatedBy,
			UpdatedAt:   group.UpdatedAt,
			Members:     members,
		},
	})
}

func UpdateUserGroup(c *gin.Context) {
	var group domain.UserGroup
	if err := db.DB.First(&group, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user group not found"})
		return
	}

	var input struct {
		Name        *string  `json:"name"`
		Handle      *string  `json:"handle"`
		Description *string  `json:"description"`
		MemberIDs   []string `json:"member_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Name != nil {
		group.Name = strings.TrimSpace(*input.Name)
	}
	if input.Handle != nil {
		group.Handle = strings.TrimSpace(*input.Handle)
	}
	if input.Description != nil {
		group.Description = strings.TrimSpace(*input.Description)
	}
	group.UpdatedAt = time.Now().UTC()
	if err := db.DB.Save(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user group"})
		return
	}
	if input.MemberIDs != nil {
		if err := replaceUserGroupMembers(group.ID, normalizeMemberIDs(input.MemberIDs, group.CreatedBy), group.CreatedBy); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to replace user group members"})
			return
		}
	}
	respondWithUserGroupDetail(c, http.StatusOK, group)
}

func DeleteUserGroup(c *gin.Context) {
	groupID := c.Param("id")
	if err := db.DB.Where("user_group_id = ?", groupID).Delete(&domain.UserGroupMember{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user group members"})
		return
	}
	if err := db.DB.Delete(&domain.UserGroup{}, "id = ?", groupID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user group"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true, "group_id": groupID})
}

func GetWorkflows(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"workflows": listActiveWorkflows()})
}

func GetTools(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"tools": listEnabledTools()})
}

func CreateWorkflowRun(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var workflow domain.WorkflowDefinition
	if err := db.DB.First(&workflow, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		return
	}

	var input struct {
		Status      string         `json:"status"`
		Summary     string         `json:"summary"`
		Input       map[string]any `json:"input"`
		CompletedAt *time.Time     `json:"completed_at"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rawInput := "{}"
	if len(input.Input) > 0 {
		if bytes, err := json.Marshal(input.Input); err == nil {
			rawInput = string(bytes)
		}
	}
	now := time.Now().UTC()
	run := domain.WorkflowRun{
		ID:          ids.NewPrefixedUUID("run"),
		WorkflowID:  workflow.ID,
		StartedBy:   currentUser.ID,
		Status:      defaultString(strings.TrimSpace(input.Status), "queued"),
		Input:       rawInput,
		Summary:     input.Summary,
		StartedAt:   now,
		CompletedAt: input.CompletedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.DB.Create(&run).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workflow run"})
		return
	}
	createWorkflowRunSteps(run.ID, []workflowRunStepResponse{
		{Name: "Queued", Status: "queued", DurationMS: 0, Detail: "Run accepted by workflow engine"},
	})
	appendWorkflowRunLog(run.ID, "info", "Run accepted by workflow engine", gin.H{"workflow_id": workflow.ID})
	response := hydrateWorkflowRun(run, workflow, enrichUser(currentUser))
	if RealtimeHub != nil {
		_ = RealtimeHub.Broadcast(realtime.Event{
			ID:          "evt_" + now.Format("20060102150405.000000"),
			Type:        "workflow.run.updated",
			WorkspaceID: primaryWorkspaceID(),
			EntityID:    run.ID,
			TS:          now.Format(time.RFC3339Nano),
			Payload:     gin.H{"run": response},
		})
	}

	c.JSON(http.StatusCreated, gin.H{"run": response})
}

func GetWorkflowRun(c *gin.Context) {
	var run domain.WorkflowRun
	if err := db.DB.First(&run, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow run not found"})
		return
	}

	workflow, starter, err := loadWorkflowRunContext(run)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hydrate workflow run"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"run": hydrateWorkflowRun(run, workflow, starter)})
}

func GetWorkflowRunLogs(c *gin.Context) {
	var run domain.WorkflowRun
	if err := db.DB.First(&run, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow run not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": loadWorkflowRunLogs(run.ID)})
}

func DeleteWorkflowRun(c *gin.Context) {
	var run domain.WorkflowRun
	if err := db.DB.First(&run, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow run not found"})
		return
	}

	if err := db.DB.Where("workflow_run_id = ?", run.ID).Delete(&domain.WorkflowRunStep{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete workflow run steps"})
		return
	}
	if err := db.DB.Where("workflow_run_id = ?", run.ID).Delete(&domain.WorkflowRunLog{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete workflow run logs"})
		return
	}
	if err := db.DB.Delete(&run).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete workflow run"})
		return
	}

	now := time.Now().UTC()
	if RealtimeHub != nil {
		_ = RealtimeHub.Broadcast(realtime.Event{
			ID:          "evt_" + now.Format("20060102150405.000000"),
			Type:        "workflow.run.deleted",
			WorkspaceID: primaryWorkspaceID(),
			EntityID:    run.ID,
			TS:          now.Format(time.RFC3339Nano),
			Payload:     gin.H{"run_id": run.ID, "workflow_id": run.WorkflowID},
		})
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true, "run_id": run.ID})
}

func CancelWorkflowRun(c *gin.Context) {
	var run domain.WorkflowRun
	if err := db.DB.First(&run, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow run not found"})
		return
	}

	var input struct {
		Summary string `json:"summary"`
	}
	if err := c.ShouldBindJSON(&input); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	run.Status = "cancelled"
	if strings.TrimSpace(input.Summary) != "" {
		run.Summary = strings.TrimSpace(input.Summary)
	}
	run.CompletedAt = &now
	run.UpdatedAt = now
	if err := db.DB.Save(&run).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel workflow run"})
		return
	}
	appendWorkflowRunStep(run.ID, workflowRunStepResponse{
		Name:       "Cancelled",
		Status:     "cancelled",
		DurationMS: 0,
		Detail:     defaultString(strings.TrimSpace(input.Summary), "Run cancelled by operator"),
	})
	appendWorkflowRunLog(run.ID, "warning", "Run cancelled", gin.H{"summary": defaultString(strings.TrimSpace(input.Summary), "Run cancelled by operator")})

	workflow, starter, err := loadWorkflowRunContext(run)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hydrate workflow run"})
		return
	}
	response := hydrateWorkflowRun(run, workflow, starter)
	broadcastWorkflowRun(response, now)
	c.JSON(http.StatusOK, gin.H{"run": response})
}

func RetryWorkflowRun(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var previous domain.WorkflowRun
	if err := db.DB.First(&previous, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow run not found"})
		return
	}

	var workflow domain.WorkflowDefinition
	if err := db.DB.First(&workflow, "id = ?", previous.WorkflowID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		return
	}

	var input struct {
		Summary string `json:"summary"`
	}
	if err := c.ShouldBindJSON(&input); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	run := domain.WorkflowRun{
		ID:           ids.NewPrefixedUUID("run"),
		WorkflowID:   previous.WorkflowID,
		StartedBy:    currentUser.ID,
		Status:       "queued",
		Input:        previous.Input,
		Summary:      defaultString(strings.TrimSpace(input.Summary), "Retry of "+previous.ID),
		RetryOfRunID: previous.ID,
		StartedAt:    now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := db.DB.Create(&run).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retry workflow run"})
		return
	}
	createWorkflowRunSteps(run.ID, []workflowRunStepResponse{
		{Name: "Retry queued", Status: "queued", DurationMS: 0, Detail: "Retry created from " + previous.ID},
	})
	appendWorkflowRunLog(run.ID, "info", "Retry queued", gin.H{"retry_of_run_id": previous.ID})

	response := hydrateWorkflowRun(run, workflow, enrichUser(currentUser))
	broadcastWorkflowRun(response, now)
	c.JSON(http.StatusCreated, gin.H{"run": response})
}

func GetWorkflowRuns(c *gin.Context) {
	query := db.DB.Order("started_at desc")
	if workflowID := strings.TrimSpace(c.Query("workflow_id")); workflowID != "" {
		query = query.Where("workflow_id = ?", workflowID)
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	}

	var runs []domain.WorkflowRun
	if err := query.Find(&runs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load workflow runs"})
		return
	}

	items := make([]workflowRunResponse, 0, len(runs))
	for _, run := range runs {
		var workflow domain.WorkflowDefinition
		if err := db.DB.First(&workflow, "id = ?", run.WorkflowID).Error; err != nil {
			continue
		}
		var starter domain.User
		if err := db.DB.First(&starter, "id = ?", run.StartedBy).Error; err != nil {
			continue
		}
		items = append(items, hydrateWorkflowRun(run, workflow, enrichUser(starter)))
	}

	c.JSON(http.StatusOK, gin.H{"runs": items})
}

func GetNotificationPreferences(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	prefs, rules := loadNotificationPreferences(currentUser.ID)
	c.JSON(http.StatusOK, gin.H{"preferences": notificationPreferenceResponse{
		InboxEnabled:    prefs.InboxEnabled,
		MentionsEnabled: prefs.MentionsEnabled,
		DMEnabled:       prefs.DMEnabled,
		MuteAll:         prefs.MuteAll,
		MuteRules:       rules,
	}})
}

func PatchNotificationPreferences(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		InboxEnabled    *bool `json:"inbox_enabled"`
		MentionsEnabled *bool `json:"mentions_enabled"`
		DMEnabled       *bool `json:"dm_enabled"`
		MuteAll         *bool `json:"mute_all"`
		MuteRules       []struct {
			ScopeType string `json:"scope_type"`
			ScopeID   string `json:"scope_id"`
			IsMuted   bool   `json:"is_muted"`
		} `json:"mute_rules"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prefs, _ := loadNotificationPreferences(currentUser.ID)
	if input.InboxEnabled != nil {
		prefs.InboxEnabled = *input.InboxEnabled
	}
	if input.MentionsEnabled != nil {
		prefs.MentionsEnabled = *input.MentionsEnabled
	}
	if input.DMEnabled != nil {
		prefs.DMEnabled = *input.DMEnabled
	}
	if input.MuteAll != nil {
		prefs.MuteAll = *input.MuteAll
	}
	if prefs.UserID == "" {
		prefs.UserID = currentUser.ID
	}
	if err := db.DB.Save(&prefs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save notification preferences"})
		return
	}

	if input.MuteRules != nil {
		if err := db.DB.Where("user_id = ?", currentUser.ID).Delete(&domain.NotificationMuteRule{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to replace mute rules"})
			return
		}
		for _, rule := range input.MuteRules {
			entry := domain.NotificationMuteRule{
				UserID:    currentUser.ID,
				ScopeType: rule.ScopeType,
				ScopeID:   rule.ScopeID,
				IsMuted:   rule.IsMuted,
			}
			if err := db.DB.Create(&entry).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save mute rules"})
				return
			}
		}
	}

	updatedPrefs, rules := loadNotificationPreferences(currentUser.ID)
	c.JSON(http.StatusOK, gin.H{"preferences": notificationPreferenceResponse{
		InboxEnabled:    updatedPrefs.InboxEnabled,
		MentionsEnabled: updatedPrefs.MentionsEnabled,
		DMEnabled:       updatedPrefs.DMEnabled,
		MuteAll:         updatedPrefs.MuteAll,
		MuteRules:       rules,
	}})
}

func buildUserProfileSummary(user domain.User) userProfileSummary {
	location := time.Now().UTC()
	if strings.TrimSpace(user.Timezone) != "" {
		if loc, err := time.LoadLocation(user.Timezone); err == nil {
			location = time.Now().In(loc)
		}
	}

	var memberships []domain.ChannelMember
	_ = db.DB.Where("user_id = ?", user.ID).Find(&memberships).Error
	channelIDs := make([]string, 0, len(memberships))
	for _, membership := range memberships {
		channelIDs = append(channelIDs, membership.ChannelID)
	}

	channels := make([]channelSummaryResponse, 0, len(channelIDs))
	if len(channelIDs) > 0 {
		var rows []domain.Channel
		_ = db.DB.Where("id IN ?", channelIDs).Order("is_starred desc, member_count desc, name asc").Limit(3).Find(&rows).Error
		for _, channel := range rows {
			channels = append(channels, channelSummaryResponse{
				ID:          channel.ID,
				Name:        channel.Name,
				Type:        channel.Type,
				MemberCount: channel.MemberCount,
				IsStarred:   channel.IsStarred,
			})
		}
	}

	var artifacts []domain.Artifact
	_ = db.DB.Where("created_by = ? OR updated_by = ?", user.ID, user.ID).Order("updated_at desc").Limit(3).Find(&artifacts).Error
	recentArtifacts := make([]artifactSummaryResponse, 0, len(artifacts))
	for _, artifact := range artifacts {
		recentArtifacts = append(recentArtifacts, artifactSummaryResponse{
			ID:        artifact.ID,
			Title:     artifact.Title,
			Type:      artifact.Type,
			Status:    artifact.Status,
			Version:   artifact.Version,
			UpdatedAt: artifact.UpdatedAt,
		})
	}

	focusAreas := make([]string, 0, 4)
	if user.Department != "" {
		focusAreas = append(focusAreas, user.Department)
	}
	if user.Title != "" {
		focusAreas = append(focusAreas, user.Title)
	}
	for _, channel := range channels {
		focusAreas = append(focusAreas, "#"+channel.Name)
		if len(focusAreas) >= 4 {
			break
		}
	}
	if len(focusAreas) == 0 {
		focusAreas = append(focusAreas, "Collaboration")
	}

	return userProfileSummary{
		LocalTime:       location.Format("3:04 PM"),
		WorkingHours:    defaultString(user.WorkingHours, "Mon - Fri"),
		FocusAreas:      focusAreas,
		TopChannels:     channels,
		RecentArtifacts: recentArtifacts,
	}
}

func listStarredChannels() []channelSummaryResponse {
	var channels []domain.Channel
	_ = db.DB.Where("is_starred = ?", true).Order("name asc").Find(&channels).Error
	items := make([]channelSummaryResponse, 0, len(channels))
	for _, channel := range channels {
		items = append(items, channelSummaryResponse{
			ID:          channel.ID,
			Name:        channel.Name,
			Type:        channel.Type,
			MemberCount: channel.MemberCount,
			IsStarred:   channel.IsStarred,
		})
	}
	return items
}

func listRecentDMs(userID string, limit int) []homeDMResponse {
	var memberships []domain.DMMember
	_ = db.DB.Where("user_id = ?", userID).Find(&memberships).Error
	items := make([]homeDMResponse, 0, len(memberships))
	for _, membership := range memberships {
		var otherMember domain.DMMember
		if err := db.DB.Where("dm_conversation_id = ? AND user_id <> ?", membership.DMConversationID, userID).First(&otherMember).Error; err != nil {
			continue
		}
		var otherUser domain.User
		if err := db.DB.First(&otherUser, "id = ?", otherMember.UserID).Error; err != nil {
			continue
		}
		var lastMessage domain.DMMessage
		_ = db.DB.Where("dm_conversation_id = ?", membership.DMConversationID).Order("created_at desc").First(&lastMessage).Error

		items = append(items, homeDMResponse{
			ID:            membership.DMConversationID,
			User:          enrichUser(otherUser),
			UserIDs:       []string{userID, otherUser.ID},
			LastMessage:   lastMessage.Content,
			LastMessageAt: lastMessage.CreatedAt,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].LastMessageAt.After(items[j].LastMessageAt)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

func listRecentDrafts(userID string, limit int) []homeDraftResponse {
	var drafts []domain.Draft
	_ = db.DB.Where("user_id = ?", userID).Order("updated_at desc").Limit(limit).Find(&drafts).Error
	items := make([]homeDraftResponse, 0, len(drafts))
	for _, draft := range drafts {
		items = append(items, homeDraftResponse{
			Scope:     draft.Scope,
			Content:   draft.Content,
			UpdatedAt: draft.UpdatedAt,
		})
	}
	return items
}

func listRecentChannelActivity(userID string, limit int) []homeRecentActivityResponse {
	channelIDs := loadCurrentUserChannelIDs(userID)
	if len(channelIDs) == 0 {
		return nil
	}

	var messages []domain.Message
	_ = db.DB.Where("channel_id IN ? AND thread_id = ''", channelIDs).Order("created_at desc").Find(&messages).Error

	items := make([]homeRecentActivityResponse, 0, len(messages))
	seen := make(map[string]struct{}, len(channelIDs))
	for _, message := range messages {
		if _, ok := seen[message.ChannelID]; ok {
			continue
		}

		var channel domain.Channel
		if err := db.DB.First(&channel, "id = ?", message.ChannelID).Error; err != nil {
			continue
		}

		seen[message.ChannelID] = struct{}{}
		items = append(items, homeRecentActivityResponse{
			ID:          "home-activity-" + message.ID,
			ChannelID:   channel.ID,
			ChannelName: channel.Name,
			LastMessage: message.Content,
			OccurredAt:  message.CreatedAt,
		})
		if len(items) >= limit {
			break
		}
	}
	return items
}

func countActiveThreadsForHome(userID string) int {
	channelIDs := loadCurrentUserChannelIDs(userID)
	if len(channelIDs) == 0 {
		return 0
	}

	var count int64
	_ = db.DB.Model(&domain.Message{}).Where("channel_id IN ? AND thread_id <> ''", channelIDs).Count(&count).Error
	return int(count)
}

func listActiveWorkflows() []domain.WorkflowDefinition {
	var workflows []domain.WorkflowDefinition
	_ = db.DB.Where("is_active = ?", true).Order("category asc, name asc").Find(&workflows).Error
	return workflows
}

func listEnabledTools() []domain.ToolDefinition {
	var tools []domain.ToolDefinition
	_ = db.DB.Where("is_enabled = ?", true).Order("category asc, name asc").Find(&tools).Error
	return tools
}

func listRecentWorkspaceLists(userID string, limit int) []workspaceListResponse {
	query := db.DB.Model(&domain.WorkspaceList{}).Order("updated_at desc")
	query = query.Where(
		"created_by = ? OR id IN (?)",
		userID,
		db.DB.Model(&domain.WorkspaceListItem{}).Select("list_id").Where("assigned_to = ?", userID),
	)

	var lists []domain.WorkspaceList
	_ = query.Limit(limit).Find(&lists).Error

	items := make([]workspaceListResponse, 0, len(lists))
	for _, list := range lists {
		items = append(items, hydrateWorkspaceList(list, false))
	}
	return items
}

func listRecentToolRunsForHome(userID string, limit int) []toolRunResponse {
	channelIDs := loadCurrentUserChannelIDs(userID)
	query := db.DB.Model(&domain.ToolRun{}).Order("started_at desc")
	if len(channelIDs) == 0 {
		query = query.Where("triggered_by = ?", userID)
	}

	var runs []domain.ToolRun
	_ = query.Find(&runs).Error

	items := make([]toolRunResponse, 0, len(runs))
	for _, run := range runs {
		response := hydrateToolRun(run, false)
		if run.TriggeredBy != userID && (response.ChannelID == "" || !containsString(channelIDs, response.ChannelID)) {
			continue
		}
		items = append(items, response)
		if len(items) >= limit {
			break
		}
	}
	return items
}

func listRecentFilesForHome(userID string, limit int) []fileAssetResponse {
	channelIDs := loadCurrentUserChannelIDs(userID)
	query := db.DB.Model(&domain.FileAsset{}).Where("uploader_id = ?", userID)
	if len(channelIDs) > 0 {
		query = query.Or("channel_id IN ?", channelIDs)
	}

	var files []domain.FileAsset
	_ = query.Order("created_at desc").Limit(limit).Find(&files).Error

	items := make([]fileAssetResponse, 0, len(files))
	for _, file := range files {
		items = append(items, hydrateFileAssetResponse(file))
	}
	return items
}

func loadCurrentUserChannelIDs(userID string) []string {
	var memberships []domain.ChannelMember
	_ = db.DB.Where("user_id = ?", userID).Find(&memberships).Error

	ids := make([]string, 0, len(memberships))
	seen := make(map[string]struct{}, len(memberships))
	for _, membership := range memberships {
		if _, ok := seen[membership.ChannelID]; ok {
			continue
		}
		seen[membership.ChannelID] = struct{}{}
		ids = append(ids, membership.ChannelID)
	}
	return ids
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func loadNotificationPreferences(userID string) (domain.NotificationPreference, []domain.NotificationMuteRule) {
	prefs := domain.NotificationPreference{
		UserID:          userID,
		InboxEnabled:    true,
		MentionsEnabled: true,
		DMEnabled:       true,
		MuteAll:         false,
	}
	_ = db.DB.Where("user_id = ?", userID).First(&prefs).Error

	var rules []domain.NotificationMuteRule
	_ = db.DB.Where("user_id = ?", userID).Order("scope_type asc, scope_id asc").Find(&rules).Error
	return prefs, rules
}

func hydrateWorkflowRun(run domain.WorkflowRun, workflow domain.WorkflowDefinition, starter domain.User) workflowRunResponse {
	input := map[string]any{}
	if strings.TrimSpace(run.Input) != "" {
		_ = json.Unmarshal([]byte(run.Input), &input)
	}
	steps := loadWorkflowRunSteps(run.ID)
	durationMS := 0
	if run.CompletedAt != nil {
		durationMS = int(run.CompletedAt.Sub(run.StartedAt).Milliseconds())
	} else {
		for _, step := range steps {
			durationMS += step.DurationMS
		}
	}
	return workflowRunResponse{
		ID:           run.ID,
		Status:       run.Status,
		Summary:      run.Summary,
		Input:        input,
		RetryOfRunID: run.RetryOfRunID,
		WorkflowID:   workflow.ID,
		WorkflowName: workflow.Name,
		TriggeredBy:  starter.Name,
		FinishedAt:   run.CompletedAt,
		DurationMS:   durationMS,
		Error:        run.Error,
		Steps:        steps,
		StartedAt:    run.StartedAt,
		CompletedAt:  run.CompletedAt,
		Workflow:     workflow,
		StartedBy:    starter,
	}
}

func loadWorkflowRunSteps(runID string) []workflowRunStepResponse {
	var rows []domain.WorkflowRunStep
	_ = db.DB.Where("workflow_run_id = ?", runID).Order("created_at asc, id asc").Find(&rows).Error
	items := make([]workflowRunStepResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, workflowRunStepResponse{
			Name:       row.Name,
			Status:     row.Status,
			DurationMS: row.DurationMS,
			Detail:     row.Detail,
		})
	}
	return items
}

func loadWorkflowRunLogs(runID string) []workflowRunLogResponse {
	var rows []domain.WorkflowRunLog
	_ = db.DB.Where("workflow_run_id = ?", runID).Order("created_at asc, id asc").Find(&rows).Error
	items := make([]workflowRunLogResponse, 0, len(rows))
	for _, row := range rows {
		metadata := map[string]any{}
		if strings.TrimSpace(row.Metadata) != "" {
			_ = json.Unmarshal([]byte(row.Metadata), &metadata)
		}
		items = append(items, workflowRunLogResponse{
			ID:        row.ID,
			Level:     defaultString(strings.TrimSpace(row.Level), "info"),
			Message:   row.Message,
			Metadata:  metadata,
			CreatedAt: row.CreatedAt,
		})
	}
	return items
}

func createWorkflowRunSteps(runID string, steps []workflowRunStepResponse) {
	for _, step := range steps {
		appendWorkflowRunStep(runID, step)
	}
}

func appendWorkflowRunStep(runID string, step workflowRunStepResponse) {
	if runID == "" || strings.TrimSpace(step.Name) == "" {
		return
	}
	_ = db.DB.Create(&domain.WorkflowRunStep{
		WorkflowRunID: runID,
		Name:          step.Name,
		Status:        step.Status,
		DurationMS:    step.DurationMS,
		Detail:        step.Detail,
		CreatedAt:     time.Now().UTC(),
	}).Error
}

func appendWorkflowRunLog(runID, level, message string, metadata map[string]any) {
	if runID == "" || strings.TrimSpace(message) == "" {
		return
	}
	rawMetadata := "{}"
	if len(metadata) > 0 {
		if bytes, err := json.Marshal(metadata); err == nil {
			rawMetadata = string(bytes)
		}
	}
	_ = db.DB.Create(&domain.WorkflowRunLog{
		WorkflowRunID: runID,
		Level:         defaultString(strings.TrimSpace(level), "info"),
		Message:       strings.TrimSpace(message),
		Metadata:      rawMetadata,
		CreatedAt:     time.Now().UTC(),
	}).Error
}

func loadWorkflowRunContext(run domain.WorkflowRun) (domain.WorkflowDefinition, domain.User, error) {
	var workflow domain.WorkflowDefinition
	if err := db.DB.First(&workflow, "id = ?", run.WorkflowID).Error; err != nil {
		return domain.WorkflowDefinition{}, domain.User{}, err
	}

	var starter domain.User
	if err := db.DB.First(&starter, "id = ?", run.StartedBy).Error; err != nil {
		return domain.WorkflowDefinition{}, domain.User{}, err
	}

	return workflow, enrichUser(starter), nil
}

func broadcastWorkflowRun(run workflowRunResponse, now time.Time) {
	if RealtimeHub == nil {
		return
	}
	_ = RealtimeHub.Broadcast(realtime.Event{
		ID:          "evt_" + now.Format("20060102150405.000000"),
		Type:        "workflow.run.updated",
		WorkspaceID: primaryWorkspaceID(),
		EntityID:    run.ID,
		TS:          now.Format(time.RFC3339Nano),
		Payload:     gin.H{"run": run},
	})
}

func normalizeMemberIDs(memberIDs []string, ownerID string) []string {
	seen := map[string]struct{}{}
	items := make([]string, 0, len(memberIDs)+1)
	if ownerID != "" {
		seen[ownerID] = struct{}{}
		items = append(items, ownerID)
	}
	for _, memberID := range memberIDs {
		memberID = strings.TrimSpace(memberID)
		if memberID == "" {
			continue
		}
		if _, ok := seen[memberID]; ok {
			continue
		}
		seen[memberID] = struct{}{}
		items = append(items, memberID)
	}
	return items
}

func replaceUserGroupMembers(groupID string, memberIDs []string, ownerID string) error {
	if err := db.DB.Where("user_group_id = ?", groupID).Delete(&domain.UserGroupMember{}).Error; err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, memberID := range memberIDs {
		role := "member"
		if memberID == ownerID {
			role = "owner"
		}
		entry := domain.UserGroupMember{
			UserGroupID: groupID,
			UserID:      memberID,
			Role:        role,
			CreatedAt:   now,
		}
		if err := db.DB.Create(&entry).Error; err != nil {
			return err
		}
	}
	return nil
}

func respondWithUserGroupDetail(c *gin.Context, statusCode int, group domain.UserGroup) {
	var memberships []domain.UserGroupMember
	_ = db.DB.Where("user_group_id = ?", group.ID).Order("role asc, created_at asc").Find(&memberships).Error
	members := make([]userGroupMemberResponse, 0, len(memberships))
	for _, membership := range memberships {
		var user domain.User
		if err := db.DB.First(&user, "id = ?", membership.UserID).Error; err != nil {
			continue
		}
		members = append(members, userGroupMemberResponse{
			Role:      membership.Role,
			CreatedAt: membership.CreatedAt,
			User:      enrichUser(user),
		})
	}
	c.JSON(statusCode, gin.H{"group": userGroupDetailResponse{
		ID:          group.ID,
		WorkspaceID: group.WorkspaceID,
		Name:        group.Name,
		Handle:      group.Handle,
		Description: group.Description,
		MemberCount: len(members),
		CreatedBy:   group.CreatedBy,
		UpdatedAt:   group.UpdatedAt,
		Members:     members,
	}})
}

func sumUnreadChannels() int {
	var channels []domain.Channel
	_ = db.DB.Where("unread_count > 0").Find(&channels).Error
	total := 0
	for _, channel := range channels {
		total += channel.UnreadCount
	}
	return total
}

func primaryWorkspaceID() string {
	var workspace domain.Workspace
	if err := db.DB.Order("id asc").First(&workspace).Error; err != nil {
		return ""
	}
	return workspace.ID
}
