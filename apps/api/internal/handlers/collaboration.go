package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/agentcollab"
	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

type reactionSummary struct {
	Emoji   string   `json:"emoji"`
	Count   int      `json:"count"`
	UserIDs []string `json:"userIds"`
}

var CollabSnapshotPath = agentcollab.DefaultPath()

type messageMetadata struct {
	Reactions   []reactionSummary `json:"reactions,omitempty"`
	Attachments any               `json:"attachments,omitempty"`
}

func getCurrentUser() (domain.User, error) {
	var user domain.User
	if err := db.DB.First(&user).Error; err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func buildUserInsight(user domain.User) string {
	var messageCount int64
	db.DB.Model(&domain.Message{}).Where("user_id = ?", user.ID).Count(&messageCount)

	var threadReplyCount int64
	db.DB.Model(&domain.Message{}).Where("user_id = ? AND thread_id <> ''", user.ID).Count(&threadReplyCount)

	var reactionCount int64
	db.DB.Table("message_reactions").
		Joins("JOIN messages ON messages.id = message_reactions.message_id").
		Where("messages.user_id = ?", user.ID).
		Count(&reactionCount)

	switch {
	case messageCount >= 5 && reactionCount >= 2:
		return user.Name + " is driving visible collaboration with frequent updates and strong team response."
	case threadReplyCount >= 2:
		return user.Name + " is active in threads and helps move discussions toward concrete next steps."
	case messageCount > 0:
		return user.Name + " is participating in the workspace and building recent collaboration history."
	default:
		return user.Name + " is present in the workspace, but there is not enough interaction history yet for a stronger insight."
	}
}

func enrichUser(user domain.User) domain.User {
	user.AIInsight = buildUserInsight(user)
	return user
}

func broadcastRealtimeEvent(eventType string, message domain.Message, payload any) error {
	if RealtimeHub == nil {
		return nil
	}

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", message.ChannelID).Error; err != nil {
		return err
	}

	return RealtimeHub.Broadcast(realtime.Event{
		ID:          "evt_" + time.Now().Format("20060102150405.000000"),
		Type:        eventType,
		WorkspaceID: channel.WorkspaceID,
		ChannelID:   message.ChannelID,
		EntityID:    message.ID,
		TS:          time.Now().UTC().Format(time.RFC3339Nano),
		Payload:     payload,
	})
}

func broadcastDMRealtimeEvent(dmID, messageID string, payload any) error {
	if RealtimeHub == nil {
		return nil
	}

	workspaceID := ""
	var workspace domain.Workspace
	if err := db.DB.Order("id asc").First(&workspace).Error; err == nil {
		workspaceID = workspace.ID
	}

	return RealtimeHub.Broadcast(realtime.Event{
		ID:          "evt_" + time.Now().Format("20060102150405.000000"),
		Type:        "message.created",
		WorkspaceID: workspaceID,
		EntityID:    messageID,
		TS:          time.Now().UTC().Format(time.RFC3339Nano),
		Payload:     payload,
	})
}

func decodeMessageMetadata(message domain.Message) messageMetadata {
	if message.Metadata == "" {
		return messageMetadata{}
	}

	var meta messageMetadata
	if err := json.Unmarshal([]byte(message.Metadata), &meta); err != nil {
		return messageMetadata{}
	}
	return meta
}

func refreshMessageMetadata(messageID string) (*domain.Message, error) {
	var message domain.Message
	if err := db.DB.First(&message, "id = ?", messageID).Error; err != nil {
		return nil, err
	}

	meta := decodeMessageMetadata(message)

	var reactions []domain.MessageReaction
	if err := db.DB.Where("message_id = ?", messageID).Order("emoji asc, created_at asc").Find(&reactions).Error; err != nil {
		return nil, err
	}

	if len(reactions) == 0 {
		meta.Reactions = nil
	} else {
		byEmoji := map[string]*reactionSummary{}
		order := make([]string, 0)
		for _, reaction := range reactions {
			summary, ok := byEmoji[reaction.Emoji]
			if !ok {
				summary = &reactionSummary{Emoji: reaction.Emoji}
				byEmoji[reaction.Emoji] = summary
				order = append(order, reaction.Emoji)
			}
			summary.Count++
			summary.UserIDs = append(summary.UserIDs, reaction.UserID)
		}

		meta.Reactions = make([]reactionSummary, 0, len(order))
		sort.Strings(order)
		for _, emoji := range order {
			meta.Reactions = append(meta.Reactions, *byEmoji[emoji])
		}
	}

	metadataJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}

	message.Metadata = string(metadataJSON)
	if err := db.DB.Model(&message).Update("metadata", message.Metadata).Error; err != nil {
		return nil, err
	}

	return &message, nil
}

func recomputeThreadParentWithDB(conn *gorm.DB, parentID string) error {
	if parentID == "" {
		return nil
	}

	var replies []domain.Message
	if err := conn.Where("thread_id = ?", parentID).Order("created_at asc").Find(&replies).Error; err != nil {
		return err
	}

	updates := map[string]any{"reply_count": len(replies)}
	if len(replies) == 0 {
		updates["last_reply_at"] = nil
	} else {
		lastReplyAt := replies[len(replies)-1].CreatedAt
		updates["last_reply_at"] = &lastReplyAt
	}

	return conn.Model(&domain.Message{}).Where("id = ?", parentID).Updates(updates).Error
}

type activityItem struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	User       domain.User `json:"user"`
	Channel    any         `json:"channel,omitempty"`
	Message    any         `json:"message,omitempty"`
	Target     string      `json:"target,omitempty"`
	Summary    string      `json:"summary"`
	OccurredAt time.Time   `json:"occurred_at"`
}

func buildActivityFeed(currentUser domain.User) []activityItem {
	activities := make([]activityItem, 0)
	nameNeedle := strings.ToLower(currentUser.Name)

	var mentionMessages []domain.Message
	db.DB.Where("user_id <> ? AND LOWER(content) LIKE ?", currentUser.ID, "%"+nameNeedle+"%").Order("created_at desc").Find(&mentionMessages)
	for _, message := range mentionMessages {
		var actor domain.User
		var channel domain.Channel
		if err := db.DB.First(&actor, "id = ?", message.UserID).Error; err != nil {
			continue
		}
		_ = db.DB.First(&channel, "id = ?", message.ChannelID).Error
		activities = append(activities, activityItem{
			ID:         "activity-mention-" + message.ID,
			Type:       "mention",
			User:       enrichUser(actor),
			Channel:    channel,
			Message:    message,
			Target:     "#" + channel.Name,
			Summary:    actor.Name + " mentioned you in #" + channel.Name,
			OccurredAt: message.CreatedAt,
		})
	}

	var replies []domain.Message
	db.DB.Where("thread_id <> ''").Order("created_at desc").Find(&replies)
	for _, reply := range replies {
		var parent domain.Message
		if err := db.DB.First(&parent, "id = ?", reply.ThreadID).Error; err != nil || parent.UserID != currentUser.ID || reply.UserID == currentUser.ID {
			continue
		}
		var actor domain.User
		var channel domain.Channel
		if err := db.DB.First(&actor, "id = ?", reply.UserID).Error; err != nil {
			continue
		}
		_ = db.DB.First(&channel, "id = ?", reply.ChannelID).Error
		activities = append(activities, activityItem{
			ID:         "activity-thread-" + reply.ID,
			Type:       "thread_reply",
			User:       enrichUser(actor),
			Channel:    channel,
			Message:    reply,
			Target:     "#" + channel.Name,
			Summary:    actor.Name + " replied to your thread in #" + channel.Name,
			OccurredAt: reply.CreatedAt,
		})
	}

	var reactions []domain.MessageReaction
	if err := db.DB.Table("message_reactions").
		Select("message_reactions.*").
		Joins("JOIN messages ON messages.id = message_reactions.message_id").
		Where("messages.user_id = ?", currentUser.ID).
		Order("message_reactions.created_at desc").
		Find(&reactions).Error; err == nil {
		for _, reaction := range reactions {
			var actor domain.User
			var message domain.Message
			var channel domain.Channel
			if err := db.DB.First(&actor, "id = ?", reaction.UserID).Error; err != nil {
				continue
			}
			if err := db.DB.First(&message, "id = ?", reaction.MessageID).Error; err != nil {
				continue
			}
			_ = db.DB.First(&channel, "id = ?", message.ChannelID).Error
			activities = append(activities, activityItem{
				ID:         "activity-reaction-" + reaction.MessageID + "-" + reaction.Emoji,
				Type:       "reaction",
				User:       enrichUser(actor),
				Channel:    channel,
				Message:    message,
				Target:     reaction.Emoji,
				Summary:    actor.Name + " reacted " + reaction.Emoji + " to your message in #" + channel.Name,
				OccurredAt: reaction.CreatedAt,
			})
		}
	}

	var dmMemberships []domain.DMMember
	db.DB.Where("user_id = ?", currentUser.ID).Find(&dmMemberships)
	for _, membership := range dmMemberships {
		var dmMessages []domain.DMMessage
		db.DB.Where("dm_conversation_id = ? AND user_id <> ?", membership.DMConversationID, currentUser.ID).Order("created_at desc").Find(&dmMessages)
		for _, message := range dmMessages {
			var actor domain.User
			if err := db.DB.First(&actor, "id = ?", message.UserID).Error; err != nil {
				continue
			}
			activities = append(activities, activityItem{
				ID:         "activity-dm-" + message.ID,
				Type:       "dm_message",
				User:       enrichUser(actor),
				Message:    message,
				Target:     "Direct messages",
				Summary:    actor.Name + " sent you a DM",
				OccurredAt: message.CreatedAt,
			})
		}
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].OccurredAt.After(activities[j].OccurredAt)
	})

	if len(activities) > 50 {
		activities = activities[:50]
	}

	return activities
}

func GetMe(c *gin.Context) {
	var user domain.User
	if err := db.DB.First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": enrichUser(user)})
}

func PatchMeSettings(c *gin.Context) {
	var user domain.User
	if err := db.DB.First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
		Mode     string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.AIProvider = input.Provider
	user.AIModel = input.Model
	user.AIMode = input.Mode

	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func GetOrganizations(c *gin.Context) {
	var orgs []domain.Organization
	db.DB.Order("created_at asc").Find(&orgs)
	c.JSON(http.StatusOK, gin.H{"organizations": orgs})
}

func GetUsers(c *gin.Context) {
	userID := c.Query("id")
	var users []domain.User

	query := db.DB.Order("name asc")
	if userID != "" {
		query = query.Where("id = ?", userID)
	}
	query.Find(&users)
	for idx := range users {
		users[idx] = enrichUser(users[idx])
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func GetAgentCollabSnapshot(c *gin.Context) {
	snapshot, err := agentcollab.ReadSnapshot(CollabSnapshotPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read agent collab snapshot"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"active_superpowers": snapshot.ActiveSuperpowers,
		"task_board":         snapshot.TaskBoard,
	})
}

func GetTeams(c *gin.Context) {
	orgID := c.Param("id")
	var teams []domain.Team

	db.DB.Where("organization_id = ?", orgID).Find(&teams)
	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

func CreateAgent(c *gin.Context) {
	orgID := c.Param("id")
	var input struct {
		Name    string `json:"name" binding:"required"`
		Type    string `json:"type" binding:"required"`
		OwnerID string `json:"owner_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agent := domain.Agent{
		ID:             "agent_" + time.Now().Format("20060102150405"),
		OrganizationID: orgID,
		Name:           input.Name,
		Type:           input.Type,
		OwnerID:        input.OwnerID,
	}

	if err := db.DB.Create(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create agent"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"agent": agent})
}

func GetWorkspaces(c *gin.Context) {
	var workspaces []domain.Workspace
	db.DB.Find(&workspaces)
	c.JSON(http.StatusOK, gin.H{"workspaces": workspaces})
}

func GetChannels(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	var channels []domain.Channel

	query := db.DB
	if workspaceID != "" {
		query = query.Where("workspace_id = ?", workspaceID)
	}
	query.Find(&channels)

	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

func GetChannelMembers(c *gin.Context) {
	channelID := c.Param("id")

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", channelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	type channelMemberResponse struct {
		User domain.User `json:"user"`
		Role string      `json:"role"`
	}

	var memberships []domain.ChannelMember
	if err := db.DB.Where("channel_id = ?", channelID).Order("created_at asc").Find(&memberships).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load channel members"})
		return
	}

	members := make([]channelMemberResponse, 0, len(memberships))
	for _, membership := range memberships {
		var user domain.User
		if err := db.DB.First(&user, "id = ?", membership.UserID).Error; err != nil {
			continue
		}
		members = append(members, channelMemberResponse{
			User: enrichUser(user),
			Role: membership.Role,
		})
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

func AddChannelMember(c *gin.Context) {
	channelID := c.Param("id")

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", channelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	var input struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user domain.User
	if err := db.DB.First(&user, "id = ?", input.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	role := input.Role
	if role == "" {
		role = "member"
	}

	membership := domain.ChannelMember{
		ChannelID: channelID,
		UserID:    input.UserID,
		Role:      role,
		CreatedAt: time.Now().UTC(),
	}
	if err := db.DB.FirstOrCreate(&membership, domain.ChannelMember{
		ChannelID: channelID,
		UserID:    input.UserID,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add channel member"})
		return
	}

	var memberCount int64
	db.DB.Model(&domain.ChannelMember{}).Where("channel_id = ?", channelID).Count(&memberCount)
	db.DB.Model(&channel).Update("member_count", int(memberCount))

	c.JSON(http.StatusCreated, gin.H{
		"member": gin.H{
			"user": enrichUser(user),
			"role": role,
		},
	})
}

func RemoveChannelMember(c *gin.Context) {
	channelID := c.Param("id")
	userID := c.Param("userId")

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", channelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	if err := db.DB.Where("channel_id = ? AND user_id = ?", channelID, userID).Delete(&domain.ChannelMember{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove channel member"})
		return
	}

	var memberCount int64
	db.DB.Model(&domain.ChannelMember{}).Where("channel_id = ?", channelID).Count(&memberCount)
	db.DB.Model(&channel).Update("member_count", int(memberCount))

	c.JSON(http.StatusOK, gin.H{"removed": true, "user_id": userID})
}

func UpdateChannel(c *gin.Context) {
	channelID := c.Param("id")

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", channelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	var input struct {
		Topic      *string `json:"topic"`
		Purpose    *string `json:"purpose"`
		IsArchived *bool   `json:"is_archived"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]any{}
	if input.Topic != nil {
		updates["topic"] = *input.Topic
	}
	if input.Purpose != nil {
		updates["purpose"] = *input.Purpose
	}
	if input.IsArchived != nil {
		updates["is_archived"] = *input.IsArchived
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no channel updates provided"})
		return
	}

	if err := db.DB.Model(&channel).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update channel"})
		return
	}
	db.DB.First(&channel, "id = ?", channelID)

	c.JSON(http.StatusOK, gin.H{"channel": channel})
}

func GetWorkspaceInvites(c *gin.Context) {
	workspaceID := c.Param("id")

	var invites []domain.WorkspaceInvite
	if err := db.DB.Where("workspace_id = ?", workspaceID).Order("created_at desc").Find(&invites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load invites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invites": invites})
}

func CreateWorkspaceInvite(c *gin.Context) {
	workspaceID := c.Param("id")

	var workspace domain.Workspace
	if err := db.DB.First(&workspace, "id = ?", workspaceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	var input struct {
		Email string `json:"email" binding:"required,email"`
		Role  string `json:"role"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := input.Role
	if role == "" {
		role = "member"
	}

	invite := domain.WorkspaceInvite{
		ID:          "invite_" + time.Now().Format("20060102150405"),
		WorkspaceID: workspaceID,
		Email:       input.Email,
		Role:        role,
		Status:      "pending",
		CreatedAt:   time.Now().UTC(),
	}
	if err := db.DB.Create(&invite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workspace invite"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"invite": invite})
}

func GetDMConversations(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var memberships []domain.DMMember
	if err := db.DB.Where("user_id = ?", currentUser.ID).Find(&memberships).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dm memberships"})
		return
	}

	type dmConversationResponse struct {
		ID            string      `json:"id"`
		User          domain.User `json:"user"`
		UserIDs       []string    `json:"user_ids"`
		LastMessage   string      `json:"last_message"`
		LastMessageAt time.Time   `json:"last_message_at"`
	}

	conversations := make([]dmConversationResponse, 0, len(memberships))
	for _, membership := range memberships {
		var otherMember domain.DMMember
		if err := db.DB.Where("dm_conversation_id = ? AND user_id <> ?", membership.DMConversationID, currentUser.ID).First(&otherMember).Error; err != nil {
			continue
		}

		var otherUser domain.User
		if err := db.DB.First(&otherUser, "id = ?", otherMember.UserID).Error; err != nil {
			continue
		}

		var lastMessage domain.DMMessage
		db.DB.Where("dm_conversation_id = ?", membership.DMConversationID).Order("created_at desc").First(&lastMessage)

		conversations = append(conversations, dmConversationResponse{
			ID:            membership.DMConversationID,
			User:          enrichUser(otherUser),
			UserIDs:       []string{currentUser.ID, otherUser.ID},
			LastMessage:   lastMessage.Content,
			LastMessageAt: lastMessage.CreatedAt,
		})
	}

	sort.Slice(conversations, func(i, j int) bool {
		return conversations[i].LastMessageAt.After(conversations[j].LastMessageAt)
	})

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

func CreateOrOpenDMConversation(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		UserID  string   `json:"user_id"`
		UserIDs []string `json:"user_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	targetUserID := input.UserID
	if targetUserID == "" && len(input.UserIDs) > 0 {
		for _, candidate := range input.UserIDs {
			if candidate != "" && candidate != currentUser.ID {
				targetUserID = candidate
				break
			}
		}
	}
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	if targetUserID == currentUser.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot create dm with current user"})
		return
	}

	var currentMemberships []domain.DMMember
	db.DB.Where("user_id = ?", currentUser.ID).Find(&currentMemberships)
	for _, membership := range currentMemberships {
		var members []domain.DMMember
		db.DB.Where("dm_conversation_id = ?", membership.DMConversationID).Order("user_id asc").Find(&members)
		if len(members) == 2 {
			memberIDs := []string{members[0].UserID, members[1].UserID}
			sort.Strings(memberIDs)
			expected := []string{currentUser.ID, targetUserID}
			sort.Strings(expected)
			if memberIDs[0] == expected[0] && memberIDs[1] == expected[1] {
				var conversation domain.DMConversation
				db.DB.First(&conversation, "id = ?", membership.DMConversationID)
				c.JSON(http.StatusOK, gin.H{"conversation": gin.H{
					"id":         conversation.ID,
					"user_ids":   []string{currentUser.ID, targetUserID},
					"created_at": conversation.CreatedAt,
				}})
				return
			}
		}
	}

	conversation := domain.DMConversation{
		ID:        "dm_" + time.Now().Format("20060102150405"),
		CreatedAt: time.Now().UTC(),
	}
	if err := db.DB.Create(&conversation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create dm conversation"})
		return
	}

	members := []domain.DMMember{
		{DMConversationID: conversation.ID, UserID: currentUser.ID},
		{DMConversationID: conversation.ID, UserID: targetUserID},
	}
	for _, member := range members {
		if err := db.DB.Create(&member).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create dm membership"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"conversation": gin.H{
		"id":         conversation.ID,
		"user_ids":   []string{currentUser.ID, targetUserID},
		"created_at": conversation.CreatedAt,
	}})
}

func GetDMMessages(c *gin.Context) {
	dmID := c.Param("id")

	var conversation domain.DMConversation
	if err := db.DB.First(&conversation, "id = ?", dmID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dm conversation not found"})
		return
	}

	var messages []domain.DMMessage
	db.DB.Where("dm_conversation_id = ?", dmID).Order("created_at asc").Find(&messages)
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func CreateDMMessage(c *gin.Context) {
	dmID := c.Param("id")

	var conversation domain.DMConversation
	if err := db.DB.First(&conversation, "id = ?", dmID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dm conversation not found"})
		return
	}

	var input struct {
		Content string `json:"content" binding:"required"`
		UserID  string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := domain.DMMessage{
		ID:               "dm_msg_" + time.Now().Format("20060102150405"),
		DMConversationID: dmID,
		UserID:           input.UserID,
		Content:          input.Content,
		CreatedAt:        time.Now().UTC(),
	}
	if err := db.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create dm message"})
		return
	}

	if err := broadcastDMRealtimeEvent(dmID, message.ID, gin.H{
		"id":         message.ID,
		"dm_id":      message.DMConversationID,
		"user_id":    message.UserID,
		"content":    message.Content,
		"created_at": message.CreatedAt,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast dm message event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": message})
}

func GetActivity(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activities": buildActivityFeed(currentUser)})
}

func GetInbox(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": buildActivityFeed(currentUser)})
}

func GetMentions(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	items := make([]activityItem, 0)
	for _, item := range buildActivityFeed(currentUser) {
		if item.Type == "mention" {
			items = append(items, item)
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func GetLater(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	type laterItem struct {
		Message any         `json:"message"`
		Channel any         `json:"channel,omitempty"`
		User    domain.User `json:"user"`
		SavedAt time.Time   `json:"saved_at"`
	}

	var savedRows []domain.SavedMessage
	db.DB.Where("user_id = ?", currentUser.ID).Order("created_at desc").Find(&savedRows)

	items := make([]laterItem, 0, len(savedRows))
	for _, saved := range savedRows {
		message, err := refreshMessageMetadata(saved.MessageID)
		if err != nil || message == nil {
			continue
		}

		var actor domain.User
		var channel domain.Channel
		if err := db.DB.First(&actor, "id = ?", message.UserID).Error; err != nil {
			continue
		}
		_ = db.DB.First(&channel, "id = ?", message.ChannelID).Error
		items = append(items, laterItem{
			Message: message,
			Channel: channel,
			User:    enrichUser(actor),
			SavedAt: saved.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func SearchWorkspace(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}

	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	needle := "%" + strings.ToLower(query) + "%"

	type searchResults struct {
		Channels []domain.Channel `json:"channels"`
		Users    []domain.User    `json:"users"`
		Messages []domain.Message `json:"messages"`
		DMs      []gin.H          `json:"dms"`
	}

	results := searchResults{
		Channels: []domain.Channel{},
		Users:    []domain.User{},
		Messages: []domain.Message{},
		DMs:      []gin.H{},
	}

	db.DB.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", needle, needle).Order("name asc").Limit(10).Find(&results.Channels)
	db.DB.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", needle, needle).Order("name asc").Limit(10).Find(&results.Users)
	for idx := range results.Users {
		results.Users[idx] = enrichUser(results.Users[idx])
	}
	db.DB.Where("LOWER(content) LIKE ?", needle).Order("created_at desc").Limit(10).Find(&results.Messages)
	for idx := range results.Messages {
		refreshed, err := refreshMessageMetadata(results.Messages[idx].ID)
		if err == nil && refreshed != nil {
			results.Messages[idx] = *refreshed
		}
	}

	var memberships []domain.DMMember
	db.DB.Where("user_id = ?", currentUser.ID).Find(&memberships)
	for _, membership := range memberships {
		var otherMember domain.DMMember
		if err := db.DB.Where("dm_conversation_id = ? AND user_id <> ?", membership.DMConversationID, currentUser.ID).First(&otherMember).Error; err != nil {
			continue
		}

		var otherUser domain.User
		if err := db.DB.First(&otherUser, "id = ?", otherMember.UserID).Error; err != nil {
			continue
		}

		var hitCount int64
		db.DB.Model(&domain.DMMessage{}).Where("dm_conversation_id = ? AND LOWER(content) LIKE ?", membership.DMConversationID, needle).Count(&hitCount)
		if !strings.Contains(strings.ToLower(otherUser.Name), strings.ToLower(query)) && !strings.Contains(strings.ToLower(otherUser.Email), strings.ToLower(query)) && hitCount == 0 {
			continue
		}

		var lastMessage domain.DMMessage
		db.DB.Where("dm_conversation_id = ?", membership.DMConversationID).Order("created_at desc").First(&lastMessage)
		results.DMs = append(results.DMs, gin.H{
			"id":              membership.DMConversationID,
			"user":            enrichUser(otherUser),
			"last_message":    lastMessage.Content,
			"last_message_at": lastMessage.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": results,
	})
}

func GetMessages(c *gin.Context) {
	channelID := c.Query("channel_id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id is required"})
		return
	}

	var messages []domain.Message
	db.DB.Where("channel_id = ? AND (thread_id = '' OR thread_id IS NULL)", channelID).Order("created_at asc").Find(&messages)
	for idx := range messages {
		refreshed, err := refreshMessageMetadata(messages[idx].ID)
		if err == nil && refreshed != nil {
			messages[idx] = *refreshed
		}
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func GetMessageThread(c *gin.Context) {
	messageID := c.Param("id")

	var parent domain.Message
	if err := db.DB.First(&parent, "id = ?", messageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	var replies []domain.Message
	db.DB.Where("thread_id = ?", messageID).Order("created_at asc").Find(&replies)
	refreshedParent, err := refreshMessageMetadata(parent.ID)
	if err == nil && refreshedParent != nil {
		parent = *refreshedParent
	}
	for idx := range replies {
		refreshed, err := refreshMessageMetadata(replies[idx].ID)
		if err == nil && refreshed != nil {
			replies[idx] = *refreshed
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"parent":  parent,
		"replies": replies,
	})
}

func CreateMessage(c *gin.Context) {
	var input struct {
		ChannelID string `json:"channel_id" binding:"required"`
		Content   string `json:"content" binding:"required"`
		UserID    string `json:"user_id" binding:"required"`
		ThreadID  string `json:"thread_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := domain.Message{
		ID:        "msg_" + time.Now().Format("20060102150405"),
		ChannelID: input.ChannelID,
		UserID:    input.UserID,
		Content:   input.Content,
		ThreadID:  input.ThreadID,
		CreatedAt: time.Now(),
	}

	if err := db.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create message"})
		return
	}

	if input.ThreadID != "" {
		lastReplyAt := msg.CreatedAt
		db.DB.Model(&domain.Message{}).
			Where("id = ?", input.ThreadID).
			Updates(map[string]any{
				"reply_count":   gorm.Expr("reply_count + ?", 1),
				"last_reply_at": &lastReplyAt,
			})
	}

	if RealtimeHub != nil {
		if err := broadcastRealtimeEvent("message.created", msg, msg); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast message event"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": msg})
}

func ToggleReaction(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	messageID := c.Param("id")
	var message domain.Message
	if err := db.DB.First(&message, "id = ?", messageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	var input struct {
		Emoji string `json:"emoji" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var reaction domain.MessageReaction
	result := db.DB.Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, currentUser.ID, input.Emoji).First(&reaction)
	added := false
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		reaction = domain.MessageReaction{
			MessageID: messageID,
			UserID:    currentUser.ID,
			Emoji:     input.Emoji,
			CreatedAt: time.Now().UTC(),
		}
		if err := db.DB.Create(&reaction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist reaction"})
			return
		}
		added = true
	case result.Error != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load reaction"})
		return
	default:
		if err := db.DB.Delete(&reaction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle reaction"})
			return
		}
	}

	refreshed, err := refreshMessageMetadata(messageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh message metadata"})
		return
	}

	if err := broadcastRealtimeEvent("reaction.updated", *refreshed, gin.H{
		"message": refreshed,
		"added":   added,
		"emoji":   input.Emoji,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast reaction event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": refreshed,
		"added":   added,
	})
}

func DeleteMessage(c *gin.Context) {
	messageID := c.Param("id")

	var message domain.Message
	if err := db.DB.First(&message, "id = ?", messageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		idsToDelete := []string{messageID}
		if message.ThreadID == "" {
			var replies []domain.Message
			if err := tx.Where("thread_id = ?", messageID).Find(&replies).Error; err != nil {
				return err
			}
			for _, reply := range replies {
				idsToDelete = append(idsToDelete, reply.ID)
			}
		}

		if err := tx.Where("message_id IN ?", idsToDelete).Delete(&domain.MessageReaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("message_id IN ?", idsToDelete).Delete(&domain.SavedMessage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("message_id IN ?", idsToDelete).Delete(&domain.UnreadMarker{}).Error; err != nil {
			return err
		}
		if err := tx.Where("message_id IN ?", idsToDelete).Delete(&domain.AIFeedback{}).Error; err != nil {
			return err
		}

		if message.ThreadID == "" {
			if err := tx.Where("thread_id = ?", messageID).Delete(&domain.Message{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Delete(&message).Error; err != nil {
			return err
		}

		if message.ThreadID != "" {
			if err := recomputeThreadParentWithDB(tx, message.ThreadID); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete message"})
		return
	}

	if err := broadcastRealtimeEvent("message.deleted", message, gin.H{"message_id": messageID}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true, "message_id": messageID})
}

func TogglePinMessage(c *gin.Context) {
	messageID := c.Param("id")

	var message domain.Message
	if err := db.DB.First(&message, "id = ?", messageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	message.IsPinned = !message.IsPinned
	if err := db.DB.Save(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update pin state"})
		return
	}

	if err := broadcastRealtimeEvent("message.updated", message, gin.H{
		"message":   message,
		"is_pinned": message.IsPinned,
		"field":     "is_pinned",
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast pin event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": message, "is_pinned": message.IsPinned})
}

func ToggleSaveForLater(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	messageID := c.Param("id")
	var message domain.Message
	if err := db.DB.First(&message, "id = ?", messageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	var saved domain.SavedMessage
	result := db.DB.Where("message_id = ? AND user_id = ?", messageID, currentUser.ID).First(&saved)
	isSaved := false
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		saved = domain.SavedMessage{
			MessageID: messageID,
			UserID:    currentUser.ID,
			CreatedAt: time.Now().UTC(),
		}
		if err := db.DB.Create(&saved).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save message"})
			return
		}
		isSaved = true
	case result.Error != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load saved state"})
		return
	default:
		if err := db.DB.Delete(&saved).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle saved state"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message_id": messageID, "saved": isSaved})
}

func MarkMessageUnread(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	messageID := c.Param("id")
	var message domain.Message
	if err := db.DB.First(&message, "id = ?", messageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	marker := domain.UnreadMarker{
		MessageID: messageID,
		UserID:    currentUser.ID,
		CreatedAt: time.Now().UTC(),
	}
	if err := db.DB.Where("message_id = ? AND user_id = ?", messageID, currentUser.ID).Assign(marker).FirstOrCreate(&marker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark unread"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message_id": messageID, "unread": true})
}
