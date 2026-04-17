package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
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
