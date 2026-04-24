package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
	"github.com/nikkofu/relay-agent-workspace/api/internal/knowledge"
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
	UnreadCount        int `json:"unread_count"`
	UnreadMentionCount int `json:"unread_mention_count"`
	DraftCount         int `json:"draft_count"`
	DMCount            int `json:"dm_count"`
	StarredCount       int `json:"starred_count"`
	GroupCount         int `json:"group_count"`
	WorkflowCount      int `json:"workflow_count"`
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
	knowledgeDigests, _ := knowledge.ListKnowledgeInbox(db.DB, knowledge.KnowledgeInboxParams{
		UserID: currentUser.ID,
		Scope:  "all",
		Limit:  5,
	})

	var groupCount int64
	db.DB.Model(&domain.UserGroupMember{}).Where("user_id = ?", currentUser.ID).Count(&groupCount)

	activity := homeActivitySummary{
		UnreadCount:        sumUnreadChannels(),
		UnreadMentionCount: countUnreadMentions(currentUser.ID),
		DraftCount:         len(drafts),
		DMCount:            len(recentDMs),
		StarredCount:       len(starredChannels),
		GroupCount:         int(groupCount),
		WorkflowCount:      len(workflows),
	}
	stats := homeStatsResponse{
		PendingActions: activity.UnreadCount,
		ActiveThreads:  countActiveThreadsForHome(currentUser.ID),
	}
	recentArtifacts := buildUserProfileSummary(currentUser).RecentArtifacts
	knowledgeInboxCount := 0
	for _, item := range knowledgeDigests {
		if !item.IsRead {
			knowledgeInboxCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"home": gin.H{
			"user":                     currentUser,
			"profile":                  buildUserProfileSummary(currentUser),
			"activity":                 activity,
			"stats":                    stats,
			"starred_channels":         starredChannels,
			"recent_dms":               recentDMs,
			"drafts":                   drafts,
			"tools":                    tools,
			"workflows":                workflows,
			"recent_activity":          listRecentChannelActivity(currentUser.ID, 6),
			"recent_artifacts":         recentArtifacts,
			"recent_lists":             listRecentWorkspaceLists(currentUser.ID, 5),
			"recent_tool_runs":         listRecentToolRunsForHome(currentUser.ID, 5),
			"recent_files":             listRecentFilesForHome(currentUser.ID, 5),
			"open_list_work":              listOpenWorkForHome(currentUser.ID, 5),
			"tool_runs_needing_attention": listToolRunsNeedingAttentionForHome(currentUser.ID, 5),
			"channel_execution_pulse":     listChannelExecutionPulseForHome(currentUser.ID, 5),
			"knowledge_inbox_count":    knowledgeInboxCount,
			"recent_knowledge_digests": knowledgeDigests,
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

func CreateWorkflowDefinition(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Trigger     string `json:"trigger"`
		Category    string `json:"category"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	now := time.Now().UTC()
	wf := domain.WorkflowDefinition{
		ID:          ids.NewPrefixedUUID("wf"),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Trigger:     defaultString(strings.TrimSpace(input.Trigger), "manual"),
		Category:    defaultString(strings.TrimSpace(input.Category), "custom"),
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.DB.Create(&wf).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workflow"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"workflow": wf})
}

func GetActivityFeed(c *gin.Context) {
	workspaceID := strings.TrimSpace(c.Query("workspace_id"))
	channelID := strings.TrimSpace(c.Query("channel_id"))
	dmID := strings.TrimSpace(c.Query("dm_id"))
	actorID := strings.TrimSpace(c.Query("actor_id"))
	eventType := strings.TrimSpace(c.Query("event_type"))
	cursor := strings.TrimSpace(c.Query("cursor"))
	limit := 20
	if n, err := strconv.Atoi(strings.TrimSpace(c.Query("limit"))); err == nil && n > 0 && n <= 100 {
		limit = n
	}

	type feedItem struct {
		ID          string         `json:"id"`
		EventType   string         `json:"event_type"`
		WorkspaceID string         `json:"workspace_id,omitempty"`
		ActorID     string         `json:"actor_id,omitempty"`
		ActorName   string         `json:"actor_name,omitempty"`
		ChannelID   string         `json:"channel_id,omitempty"`
		ChannelName string         `json:"channel_name,omitempty"`
		DMID        string         `json:"dm_id,omitempty"`
		EntityID    string         `json:"entity_id,omitempty"`
		EntityTitle string         `json:"entity_title,omitempty"`
		EntityKind  string         `json:"entity_kind,omitempty"`
		Title       string         `json:"title"`
		Body        string         `json:"body,omitempty"`
		Link        string         `json:"link,omitempty"`
		OccurredAt  string         `json:"occurred_at"`
		Meta        map[string]any `json:"meta,omitempty"`
	}

	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	items := make([]feedItem, 0, limit*12)
	var cursorTime time.Time
	if cursor != "" {
		cursorTime, _ = time.Parse(time.RFC3339Nano, cursor)
	}

	var workspaceChannelIDs []string
	_ = db.DB.Model(&domain.Channel{}).Where("workspace_id = ?", workspaceID).Pluck("id", &workspaceChannelIDs).Error
	workspaceChannelSet := make(map[string]struct{}, len(workspaceChannelIDs))
	for _, id := range workspaceChannelIDs {
		workspaceChannelSet[id] = struct{}{}
	}

	channelNames := map[string]string{}
	entityMeta := map[string]domain.KnowledgeEntity{}
	userNames := map[string]string{}

	loadChannelName := func(id string) string {
		if id == "" {
			return ""
		}
		if name, ok := channelNames[id]; ok {
			return name
		}
		var channel domain.Channel
		if err := db.DB.Select("id", "name").First(&channel, "id = ?", id).Error; err == nil {
			channelNames[id] = channel.Name
			return channel.Name
		}
		channelNames[id] = ""
		return ""
	}
	loadEntity := func(id string) domain.KnowledgeEntity {
		if id == "" {
			return domain.KnowledgeEntity{}
		}
		if entity, ok := entityMeta[id]; ok {
			return entity
		}
		var entity domain.KnowledgeEntity
		if err := db.DB.First(&entity, "id = ?", id).Error; err == nil {
			entityMeta[id] = entity
			return entity
		}
		entityMeta[id] = domain.KnowledgeEntity{}
		return domain.KnowledgeEntity{}
	}
	loadUserName := func(id string) string {
		if id == "" {
			return ""
		}
		if name, ok := userNames[id]; ok {
			return name
		}
		var user domain.User
		if err := db.DB.First(&user, "id = ?", id).Error; err == nil {
			userNames[id] = enrichUser(user).Name
			return userNames[id]
		}
		userNames[id] = ""
		return ""
	}
	includeEvent := func(t string) bool {
		return eventType == "" || eventType == t
	}
	appendItem := func(item feedItem) {
		if actorID != "" && item.ActorID != actorID {
			return
		}
		if channelID != "" && item.ChannelID != channelID {
			return
		}
		if dmID != "" && item.DMID != dmID {
			return
		}
		items = append(items, item)
	}

	// 1. Channel messages
	if includeEvent("message") {
		var messages []domain.Message
		msgQ := db.DB.Order("created_at desc").Limit(limit).Where("thread_id = ''")
		if channelID != "" {
			msgQ = msgQ.Where("channel_id = ?", channelID)
		} else {
			if len(workspaceChannelIDs) == 0 {
				msgQ = msgQ.Where("1=0")
			} else {
				msgQ = msgQ.Where("channel_id IN ?", workspaceChannelIDs)
			}
		}
		if !cursorTime.IsZero() {
			msgQ = msgQ.Where("created_at < ?", cursorTime)
		}
		_ = msgQ.Find(&messages).Error
		for _, m := range messages {
			appendItem(feedItem{
				ID:          "message:" + m.ID,
				EventType:   "message",
				WorkspaceID: workspaceID,
				ActorID:     m.UserID,
				ActorName:   loadUserName(m.UserID),
				ChannelID:   m.ChannelID,
				ChannelName: loadChannelName(m.ChannelID),
				Title:       "Message in #" + defaultString(loadChannelName(m.ChannelID), "channel"),
				Body:        m.Content,
				Link:        "/workspace?c=" + m.ChannelID,
				OccurredAt:  m.CreatedAt.Format(time.RFC3339Nano),
				Meta:        map[string]any{"message_id": m.ID},
			})
		}
	}

	// 1b. Thread replies
	if includeEvent("reply") {
		var replies []domain.Message
		replyQ := db.DB.Order("created_at desc").Limit(limit).Where("thread_id <> ''")
		if channelID != "" {
			replyQ = replyQ.Where("channel_id = ?", channelID)
		} else {
			if len(workspaceChannelIDs) == 0 {
				replyQ = replyQ.Where("1=0")
			} else {
				replyQ = replyQ.Where("channel_id IN ?", workspaceChannelIDs)
			}
		}
		if !cursorTime.IsZero() {
			replyQ = replyQ.Where("created_at < ?", cursorTime)
		}
		_ = replyQ.Find(&replies).Error
		for _, m := range replies {
			appendItem(feedItem{
				ID:          "reply:" + m.ID,
				EventType:   "reply",
				WorkspaceID: workspaceID,
				ActorID:     m.UserID,
				ActorName:   loadUserName(m.UserID),
				ChannelID:   m.ChannelID,
				ChannelName: loadChannelName(m.ChannelID),
				Title:       defaultString(loadUserName(m.UserID), "Someone") + " replied in #" + defaultString(loadChannelName(m.ChannelID), "channel"),
				Body:        m.Content,
				Link:        activityFeedMessageLink(m.ChannelID, m.ID),
				OccurredAt:  m.CreatedAt.Format(time.RFC3339Nano),
				Meta:        map[string]any{"message_id": m.ID, "thread_id": m.ThreadID},
			})
		}
	}

	// 2. Compose activities
	if includeEvent("compose_activity") {
		var composeActivities []domain.AIComposeActivity
		compQ := db.DB.Order("created_at desc").Limit(limit)
		compQ = compQ.Where("workspace_id = ?", workspaceID)
		if channelID != "" {
			compQ = compQ.Where("channel_id = ?", channelID)
		}
		if dmID != "" {
			compQ = compQ.Where("dm_conversation_id = ?", dmID)
		}
		if !cursorTime.IsZero() {
			compQ = compQ.Where("created_at < ?", cursorTime)
		}
		_ = compQ.Find(&composeActivities).Error
		for _, a := range composeActivities {
			appendItem(feedItem{
				ID:          "compose_activity:" + a.ID,
				EventType:   "compose_activity",
				WorkspaceID: a.WorkspaceID,
				ActorID:     a.UserID,
				ActorName:   loadUserName(a.UserID),
				ChannelID:   a.ChannelID,
				ChannelName: loadChannelName(a.ChannelID),
				DMID:        a.DMConversationID,
				Title:       "AI Compose · " + defaultString(a.Intent, "reply"),
				Body:        a.Provider + " / " + a.Model,
				Link:        activityFeedLink(a.ChannelID, a.DMConversationID, ""),
				OccurredAt:  a.CreatedAt.Format(time.RFC3339Nano),
				Meta:        map[string]any{"compose_id": a.ComposeID, "intent": a.Intent, "suggestion_count": a.SuggestionCount, "provider": a.Provider, "model": a.Model},
			})
		}
	}

	// 3. Knowledge ask answers
	if includeEvent("knowledge_ask") {
		var askAnswers []domain.KnowledgeEntityAskAnswer
		askQ := db.DB.Order("created_at desc").Limit(limit)
		askQ = askQ.Where("workspace_id = ?", workspaceID)
		if !cursorTime.IsZero() {
			askQ = askQ.Where("created_at < ?", cursorTime)
		}
		_ = askQ.Find(&askAnswers).Error
		for _, a := range askAnswers {
			entity := loadEntity(a.EntityID)
			appendItem(feedItem{
				ID:          "knowledge_ask:" + a.ID,
				EventType:   "knowledge_ask",
				WorkspaceID: a.WorkspaceID,
				ActorID:     a.UserID,
				ActorName:   loadUserName(a.UserID),
				EntityID:    a.EntityID,
				EntityTitle: entity.Title,
				EntityKind:  entity.Kind,
				Title:       "Ask AI · " + defaultString(entity.Title, "entity"),
				Body:        a.Question,
				Link:        "/workspace/knowledge/" + a.EntityID,
				OccurredAt:  a.AnsweredAt.Format(time.RFC3339Nano),
				Meta:        map[string]any{"provider": a.Provider, "model": a.Model, "citation_count": a.CitationCount},
			})
		}
	}

	// 3b. User mentions from persisted mention rows
	if includeEvent("mention") {
		var mentions []domain.MessageMention
		mentionQ := db.DB.Order("created_at desc").Limit(limit).Where("mention_kind = ?", "user")
		if channelID != "" {
			mentionQ = mentionQ.Where("channel_id = ?", channelID)
		} else if dmID != "" {
			mentionQ = mentionQ.Where("dm_id = ?", dmID)
		} else {
			if len(workspaceChannelIDs) == 0 {
				mentionQ = mentionQ.Where("workspace_id = ?", workspaceID)
			} else {
				mentionQ = mentionQ.Where("workspace_id = ?", workspaceID)
			}
		}
		if !cursorTime.IsZero() {
			mentionQ = mentionQ.Where("created_at < ?", cursorTime)
		}
		_ = mentionQ.Find(&mentions).Error
		for _, mention := range mentions {
			var actor domain.User
			if err := db.DB.First(&actor, "id = ?", mention.MentionedByUserID).Error; err != nil {
				continue
			}
			body := ""
			link := activityFeedScopedMessageLink(mention.ChannelID, mention.DMID, mention.MessageID)
			if mention.ChannelID != "" {
				var message domain.Message
				if err := db.DB.Select("id", "content").First(&message, "id = ?", mention.MessageID).Error; err == nil {
					body = message.Content
				}
			} else if mention.DMID != "" {
				var dmMessage domain.DMMessage
				if err := db.DB.Select("id", "content").First(&dmMessage, "id = ?", mention.MessageID).Error; err == nil {
					body = dmMessage.Content
				}
			}

			item := feedItem{
				ID:          "mention:" + mention.ID,
				EventType:   "mention",
				WorkspaceID: workspaceID,
				ActorID:     mention.MentionedByUserID,
				ActorName:   loadUserName(mention.MentionedByUserID),
				ChannelID:   mention.ChannelID,
				DMID:        mention.DMID,
				Title:       defaultString(loadUserName(mention.MentionedByUserID), "Someone") + " mentioned you",
				Body:        body,
				Link:        link,
				OccurredAt:  mention.CreatedAt.Format(time.RFC3339Nano),
				Meta: map[string]any{
					"message_id":           mention.MessageID,
					"mentioned_user_id":    mention.MentionedUserID,
					"mentioned_by_user_id": mention.MentionedByUserID,
					"mention_kind":         mention.MentionKind,
				},
			}

			if mention.ChannelID != "" {
				item.ChannelName = loadChannelName(mention.ChannelID)
				item.Title = defaultString(loadUserName(mention.MentionedByUserID), "Someone") + " mentioned you in #" + defaultString(item.ChannelName, "channel")
			} else if mention.DMID != "" {
				item.Title = defaultString(loadUserName(mention.MentionedByUserID), "Someone") + " mentioned you in DM"
			}

			if item.ActorID != actor.ID {
				item.ActorID = actor.ID
			}
			appendItem(item)
		}

		var messages []domain.Message
		mentionQ = db.DB.Order("created_at desc").Limit(limit)
		if channelID != "" {
			mentionQ = mentionQ.Where("channel_id = ?", channelID)
		} else {
			if len(workspaceChannelIDs) == 0 {
				mentionQ = mentionQ.Where("1=0")
			} else {
				mentionQ = mentionQ.Where("channel_id IN ?", workspaceChannelIDs)
			}
		}
		if !cursorTime.IsZero() {
			mentionQ = mentionQ.Where("created_at < ?", cursorTime)
		}
		_ = mentionQ.Find(&messages).Error
		for _, m := range messages {
			if strings.TrimSpace(m.Metadata) == "" {
				continue
			}
			var meta messageMetadata
			if err := json.Unmarshal([]byte(m.Metadata), &meta); err != nil || len(meta.EntityMentions) == 0 {
				continue
			}
			appendItem(feedItem{
				ID:          "mention:" + m.ID,
				EventType:   "mention",
				WorkspaceID: workspaceID,
				ActorID:     m.UserID,
				ActorName:   loadUserName(m.UserID),
				ChannelID:   m.ChannelID,
				ChannelName: loadChannelName(m.ChannelID),
				EntityID:    meta.EntityMentions[0].EntityID,
				EntityTitle: meta.EntityMentions[0].EntityTitle,
				EntityKind:  meta.EntityMentions[0].EntityKind,
				Title:       defaultString(loadUserName(m.UserID), "Someone") + " mentioned " + defaultString(meta.EntityMentions[0].EntityTitle, "an entity"),
				Body:        m.Content,
				Link:        activityFeedMessageLink(m.ChannelID, m.ID),
				OccurredAt:  m.CreatedAt.Format(time.RFC3339Nano),
				Meta:        map[string]any{"message_id": m.ID, "mention_kind": "entity"},
			})
		}
	}

	// 4. Automation jobs
	if includeEvent("automation_job") {
		var jobs []domain.AIAutomationJob
		jobQ := db.DB.Order("created_at desc").Limit(limit).Where("workspace_id = ?", workspaceID)
		if !cursorTime.IsZero() {
			jobQ = jobQ.Where("created_at < ?", cursorTime)
		}
		_ = jobQ.Find(&jobs).Error
		for _, j := range jobs {
			entity := domain.KnowledgeEntity{}
			if j.ScopeType == "knowledge_entity" {
				entity = loadEntity(j.ScopeID)
			}
			occurredAt := j.CreatedAt
			if j.FinishedAt != nil {
				occurredAt = *j.FinishedAt
			} else if j.StartedAt != nil {
				occurredAt = *j.StartedAt
			} else if !j.ScheduledAt.IsZero() {
				occurredAt = j.ScheduledAt
			}
			appendItem(feedItem{
				ID:          "automation_job:" + j.ID,
				EventType:   "automation_job",
				WorkspaceID: j.WorkspaceID,
				EntityID:    j.ScopeID,
				EntityTitle: entity.Title,
				EntityKind:  entity.Kind,
				Title:       "Automation · " + strings.ReplaceAll(j.JobType, "_", " ") + " · " + j.Status,
				Body:        strings.ReplaceAll(defaultString(j.LastError, j.TriggerReason), "_", " "),
				Link:        activityFeedLink("", "", j.ScopeID),
				OccurredAt:  occurredAt.Format(time.RFC3339Nano),
				Meta:        map[string]any{"status": j.Status, "attempt_count": j.AttemptCount, "job_type": j.JobType},
			})
		}
	}

	// 5. File uploads
	if includeEvent("file_uploaded") {
		var files []domain.FileAsset
		fileQ := db.DB.Order("created_at desc").Limit(limit)
		if channelID != "" {
			fileQ = fileQ.Where("channel_id = ?", channelID)
		} else {
			if len(workspaceChannelIDs) == 0 {
				fileQ = fileQ.Where("1=0")
			} else {
				fileQ = fileQ.Where("channel_id IN ?", workspaceChannelIDs)
			}
		}
		if !cursorTime.IsZero() {
			fileQ = fileQ.Where("created_at < ?", cursorTime)
		}
		_ = fileQ.Find(&files).Error
		for _, file := range files {
			appendItem(feedItem{
				ID:          "file_uploaded:" + file.ID,
				EventType:   "file_uploaded",
				WorkspaceID: workspaceID,
				ActorID:     file.UploaderID,
				ActorName:   loadUserName(file.UploaderID),
				ChannelID:   file.ChannelID,
				ChannelName: loadChannelName(file.ChannelID),
				Title:       "Uploaded file · " + file.Name,
				Body:        defaultString(file.ContentSummary, file.ContentType),
				Link:        "/workspace?c=" + file.ChannelID,
				OccurredAt:  file.CreatedAt.Format(time.RFC3339Nano),
				Meta:        map[string]any{"file_id": file.ID, "content_type": file.ContentType, "source_kind": file.SourceKind},
			})
		}
	}

	// 5b. Artifact updates
	if includeEvent("artifact_updated") {
		var artifacts []domain.Artifact
		artifactQ := db.DB.Joins("JOIN channels ON channels.id = artifacts.channel_id").
			Where("channels.workspace_id = ?", workspaceID).
			Order("artifacts.updated_at desc").
			Limit(limit)
		if channelID != "" {
			artifactQ = artifactQ.Where("artifacts.channel_id = ?", channelID)
		}
		if !cursorTime.IsZero() {
			artifactQ = artifactQ.Where("artifacts.updated_at < ?", cursorTime)
		}
		_ = artifactQ.Find(&artifacts).Error
		for _, artifact := range artifacts {
			appendItem(feedItem{
				ID:          "artifact_updated:" + artifact.ID,
				EventType:   "artifact_updated",
				WorkspaceID: workspaceID,
				ActorID:     artifact.UpdatedBy,
				ActorName:   loadUserName(artifact.UpdatedBy),
				ChannelID:   artifact.ChannelID,
				ChannelName: loadChannelName(artifact.ChannelID),
				Title:       "Artifact updated · " + artifact.Title,
				Body:        strings.TrimSpace(strings.Join([]string{artifact.Type, artifact.Status}, " · ")),
				Link:        activityFeedLink(artifact.ChannelID, "", ""),
				OccurredAt:  artifact.UpdatedAt.Format(time.RFC3339Nano),
				Meta:        map[string]any{"artifact_id": artifact.ID, "artifact_type": artifact.Type, "version_id": artifact.Version},
			})
		}
	}

	// 5c. Reactions on persisted channel messages
	if includeEvent("reaction") {
		var reactions []domain.MessageReaction
		reactionQ := db.DB.Table("message_reactions").
			Select("message_reactions.*").
			Joins("JOIN messages ON messages.id = message_reactions.message_id").
			Joins("JOIN channels ON channels.id = messages.channel_id").
			Where("channels.workspace_id = ?", workspaceID).
			Order("message_reactions.created_at desc").
			Limit(limit)
		if channelID != "" {
			reactionQ = reactionQ.Where("messages.channel_id = ?", channelID)
		}
		if !cursorTime.IsZero() {
			reactionQ = reactionQ.Where("message_reactions.created_at < ?", cursorTime)
		}
		_ = reactionQ.Find(&reactions).Error
		for _, reaction := range reactions {
			var message domain.Message
			if err := db.DB.First(&message, "id = ?", reaction.MessageID).Error; err != nil {
				continue
			}
			appendItem(feedItem{
				ID:          "reaction:" + strconv.FormatUint(uint64(reaction.ID), 10),
				EventType:   "reaction",
				WorkspaceID: workspaceID,
				ActorID:     reaction.UserID,
				ActorName:   loadUserName(reaction.UserID),
				ChannelID:   message.ChannelID,
				ChannelName: loadChannelName(message.ChannelID),
				Title:       defaultString(loadUserName(reaction.UserID), "Someone") + " reacted " + reaction.Emoji,
				Body:        message.Content,
				Link:        activityFeedMessageLink(message.ChannelID, message.ID),
				OccurredAt:  reaction.CreatedAt.Format(time.RFC3339Nano),
				Meta:        map[string]any{"message_id": message.ID, "emoji": reaction.Emoji},
			})
		}
	}

	// 6. Schedule bookings
	if includeEvent("schedule_booking") {
		var bookings []domain.AIScheduleBooking
		bookingQ := db.DB.Order("created_at desc").Limit(limit).Where("workspace_id = ?", workspaceID)
		if channelID != "" {
			bookingQ = bookingQ.Where("channel_id = ?", channelID)
		}
		if dmID != "" {
			bookingQ = bookingQ.Where("dm_conversation_id = ?", dmID)
		}
		if !cursorTime.IsZero() {
			bookingQ = bookingQ.Where("created_at < ?", cursorTime)
		}
		_ = bookingQ.Find(&bookings).Error
		for _, booking := range bookings {
			appendItem(feedItem{
				ID:          "schedule_booking:" + booking.ID,
				EventType:   "schedule_booking",
				WorkspaceID: booking.WorkspaceID,
				ActorID:     booking.RequestedBy,
				ActorName:   loadUserName(booking.RequestedBy),
				ChannelID:   booking.ChannelID,
				ChannelName: loadChannelName(booking.ChannelID),
				DMID:        booking.DMConversationID,
				Title:       booking.Title,
				Body:        booking.Description,
				Link:        activityFeedLink(booking.ChannelID, booking.DMConversationID, ""),
				OccurredAt:  booking.CreatedAt.Format(time.RFC3339Nano),
				Meta:        map[string]any{"booking_id": booking.ID, "status": booking.Status, "provider": booking.Provider, "starts_at": booking.StartsAt.Format(time.RFC3339Nano)},
			})
		}
	}

	// 6b. Tool runs
	if includeEvent("tool_run") {
		var runs []domain.ToolRun
		runQ := db.DB.Order("updated_at desc").Limit(limit * 3)
		if !cursorTime.IsZero() {
			runQ = runQ.Where("updated_at < ?", cursorTime)
		}
		_ = runQ.Find(&runs).Error
		for _, run := range runs {
			response := hydrateToolRun(run, false)
			if response.ChannelID == "" {
				continue
			}
			if channelID != "" && response.ChannelID != channelID {
				continue
			}
			if _, ok := workspaceChannelSet[response.ChannelID]; !ok {
				continue
			}
			occurredAt := run.StartedAt
			if run.CompletedAt != nil {
				occurredAt = *run.CompletedAt
			}
			appendItem(feedItem{
				ID:          "tool_run:" + run.ID,
				EventType:   "tool_run",
				WorkspaceID: workspaceID,
				ActorID:     run.TriggeredBy,
				ActorName:   loadUserName(run.TriggeredBy),
				ChannelID:   response.ChannelID,
				ChannelName: loadChannelName(response.ChannelID),
				Title:       defaultString(response.ToolName, "Tool") + " · " + defaultString(run.Status, "unknown"),
				Body:        defaultString(strings.TrimSpace(run.Summary), defaultString(response.ToolKey, run.ToolID)),
				Link:        "/workspace/workflows",
				OccurredAt:  occurredAt.Format(time.RFC3339Nano),
				Meta:        map[string]any{"run_id": run.ID, "tool_name": response.ToolName, "status": defaultString(run.Status, "unknown")},
			})
		}
	}

	// 6c. DM messages
	if includeEvent("dm_message") {
		var messages []domain.DMMessage
		dmQ := db.DB.Order("created_at desc").Limit(limit)
		if dmID != "" {
			dmQ = dmQ.Where("dm_conversation_id = ?", dmID)
		}
		if !cursorTime.IsZero() {
			dmQ = dmQ.Where("created_at < ?", cursorTime)
		}
		_ = dmQ.Find(&messages).Error
		for _, m := range messages {
			appendItem(feedItem{
				ID:          "dm_message:" + m.ID,
				EventType:   "dm_message",
				WorkspaceID: workspaceID,
				ActorID:     m.UserID,
				ActorName:   loadUserName(m.UserID),
				DMID:        m.DMConversationID,
				Title:       defaultString(loadUserName(m.UserID), "Someone") + " sent a DM",
				Body:        m.Content,
				Link:        activityFeedDMMessageLink(m.DMConversationID),
				OccurredAt:  m.CreatedAt.Format(time.RFC3339Nano),
				Meta:        map[string]any{"message_id": m.ID},
			})
		}
	}

	sort.Slice(items, func(i, j int) bool { return items[i].OccurredAt > items[j].OccurredAt })

	if len(items) > limit {
		items = items[:limit]
	}

	var nextCursor string
	if len(items) == limit {
		nextCursor = items[len(items)-1].OccurredAt
	}

	var total int64
	for _, tableCount := range []func() int64{
		func() int64 {
			var n int64
			_ = db.DB.Model(&domain.Message{}).Joins("JOIN channels ON channels.id = messages.channel_id").Where("channels.workspace_id = ? AND messages.thread_id = ''", workspaceID).Count(&n).Error
			return n
		},
		func() int64 {
			var n int64
			_ = db.DB.Model(&domain.Message{}).Joins("JOIN channels ON channels.id = messages.channel_id").Where("channels.workspace_id = ? AND messages.thread_id <> ''", workspaceID).Count(&n).Error
			return n
		},
		func() int64 {
			var n int64
			_ = db.DB.Model(&domain.FileAsset{}).Joins("JOIN channels ON channels.id = file_assets.channel_id").Where("channels.workspace_id = ?", workspaceID).Count(&n).Error
			return n
		},
		func() int64 {
			var n int64
			_ = db.DB.Model(&domain.Artifact{}).Joins("JOIN channels ON channels.id = artifacts.channel_id").Where("channels.workspace_id = ?", workspaceID).Count(&n).Error
			return n
		},
		func() int64 {
			var n int64
			_ = db.DB.Model(&domain.AIScheduleBooking{}).Where("workspace_id = ?", workspaceID).Count(&n).Error
			return n
		},
		func() int64 {
			var n int64
			_ = db.DB.Model(&domain.AIComposeActivity{}).Where("workspace_id = ?", workspaceID).Count(&n).Error
			return n
		},
		func() int64 {
			var n int64
			_ = db.DB.Model(&domain.KnowledgeEntityAskAnswer{}).Where("workspace_id = ?", workspaceID).Count(&n).Error
			return n
		},
		func() int64 {
			var n int64
			_ = db.DB.Model(&domain.AIAutomationJob{}).Where("workspace_id = ?", workspaceID).Count(&n).Error
			return n
		},
		func() int64 {
			var n int64
			_ = db.DB.Model(&domain.MessageReaction{}).
				Joins("JOIN messages ON messages.id = message_reactions.message_id").
				Joins("JOIN channels ON channels.id = messages.channel_id").
				Where("channels.workspace_id = ?", workspaceID).
				Count(&n).Error
			return n
		},
		func() int64 {
			var n int64
			var messages []domain.Message
			_ = db.DB.Joins("JOIN channels ON channels.id = messages.channel_id").Where("channels.workspace_id = ?", workspaceID).Find(&messages).Error
			for _, m := range messages {
				if strings.TrimSpace(m.Metadata) == "" {
					continue
				}
				var meta messageMetadata
				if err := json.Unmarshal([]byte(m.Metadata), &meta); err == nil && len(meta.EntityMentions) > 0 {
					n++
				}
			}
			return n
		},
		func() int64 {
			var n int64
			_ = db.DB.Model(&domain.ToolRun{}).Count(&n).Error
			return n
		},
		func() int64 {
			var n int64
			_ = db.DB.Model(&domain.DMMessage{}).Count(&n).Error
			return n
		},
	} {
		total += tableCount()
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "next_cursor": nextCursor, "total": total})
}

func activityFeedLink(channelID, dmID, entityID string) string {
	if entityID != "" {
		return "/workspace/knowledge/" + entityID
	}
	if channelID != "" {
		return "/workspace?c=" + channelID
	}
	if dmID != "" {
		return "/workspace/dms"
	}
	return "/workspace/activity"
}

func activityFeedMessageLink(channelID, messageID string) string {
	if channelID == "" {
		return "/workspace/activity"
	}
	if messageID == "" {
		return "/workspace?c=" + channelID
	}
	return "/workspace?c=" + channelID + "&m=" + messageID
}

func activityFeedScopedMessageLink(channelID, dmID, messageID string) string {
	if channelID != "" {
		return activityFeedMessageLink(channelID, messageID)
	}
	if dmID != "" {
		return activityFeedDMMessageLink(dmID)
	}
	return "/workspace/activity"
}

func activityFeedDMMessageLink(dmID string) string {
	if dmID == "" {
		return "/workspace/dms"
	}
	return "/workspace/dms?id=" + dmID
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

func listOpenWorkForHome(userID string, limit int) []any {
	var items []domain.WorkspaceListItem
	// Find items assigned to user or created by user that are not completed
	db.DB.Where("(assigned_to = ? OR created_by = ?) AND is_completed = ?", userID, userID, false).
		Order("due_at asc, updated_at desc").
		Limit(limit).
		Find(&items)
	
	results := make([]any, 0, len(items))
	for _, item := range items {
		var list domain.WorkspaceList
		db.DB.First(&list, "id = ?", item.ListID)
		results = append(results, gin.H{
			"item":    hydrateWorkspaceListItem(item),
			"channel_id": list.ChannelID,
			"list_title": list.Title,
		})
	}
	return results
}

func listToolRunsNeedingAttentionForHome(userID string, limit int) []any {
	var runs []domain.ToolRun
	// Find failed runs or running runs triggered by user
	db.DB.Where("triggered_by = ? AND status IN ?", userID, []string{"failed", "running"}).
		Order("updated_at desc").
		Limit(limit).
		Find(&runs)
	
	results := make([]any, 0, len(runs))
	for _, run := range runs {
		results = append(results, hydrateToolRun(run, false))
	}
	return results
}

func listChannelExecutionPulseForHome(userID string, limit int) []any {
	// Find channels with recent execution activity
	var channels []domain.Channel
	db.DB.Limit(limit).Find(&channels) // Simplified for now
	
	results := make([]any, 0, len(channels))
	for _, channel := range channels {
		var openItemCount int64
		db.DB.Model(&domain.WorkspaceListItem{}).
			Joins("JOIN workspace_lists ON workspace_lists.id = workspace_list_items.list_id").
			Where("workspace_lists.channel_id = ? AND workspace_list_items.is_completed = ?", channel.ID, false).
			Count(&openItemCount)
		
		var overdueCount int64
		db.DB.Model(&domain.WorkspaceListItem{}).
			Joins("JOIN workspace_lists ON workspace_lists.id = workspace_list_items.list_id").
			Where("workspace_lists.channel_id = ? AND workspace_list_items.is_completed = ? AND due_at < ?", channel.ID, false, time.Now().UTC()).
			Count(&overdueCount)
		
		sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour).UTC()
		var newItems7d int64
		db.DB.Model(&domain.WorkspaceListItem{}).
			Joins("JOIN workspace_lists ON workspace_lists.id = workspace_list_items.list_id").
			Where("workspace_lists.channel_id = ? AND workspace_list_items.created_at > ?", channel.ID, sevenDaysAgo).
			Count(&newItems7d)
		
		var completedItems7d int64
		db.DB.Model(&domain.WorkspaceListItem{}).
			Joins("JOIN workspace_lists ON workspace_lists.id = workspace_list_items.list_id").
			Where("workspace_lists.channel_id = ? AND workspace_list_items.is_completed = ? AND workspace_list_items.completed_at > ?", channel.ID, true, sevenDaysAgo).
			Count(&completedItems7d)
		
		var toolFailures int64
		db.DB.Model(&domain.ToolRun{}).
			// Finding tool failures roughly related to channel via triggered context if any
			// For now, workspace-wide failed runs since toolrun doesn't have channel_id yet
			Where("status = ? AND created_at > ?", "failed", sevenDaysAgo).
			Count(&toolFailures)
		
		if openItemCount > 0 {
			summary := fmt.Sprintf("%d open items in #%s", openItemCount, channel.Name)
			if overdueCount > 0 {
				summary += fmt.Sprintf(" (%d overdue)", overdueCount)
			}

			results = append(results, gin.H{
				"channel_id":                channel.ID,
				"channel_name":              channel.Name,
				"open_item_count":           openItemCount,
				"overdue_count":             overdueCount,
				"open_item_delta_7d":        int(newItems7d - completedItems7d),
				"overdue_delta_7d":          0, // Stable for now
				"recent_tool_failure_count": int(toolFailures),
				"summary":                   summary,
			})
		}
	}
	return results
}
