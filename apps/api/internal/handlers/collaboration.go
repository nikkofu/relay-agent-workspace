package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/agentcollab"
	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

type reactionSummary struct {
	Emoji   string   `json:"emoji"`
	Count   int      `json:"count"`
	UserIDs []string `json:"userIds"`
}

type messageAttachment struct {
	ID             string               `json:"id"`
	Kind           string               `json:"kind"`
	Type           string               `json:"type"`
	URL            string               `json:"url"`
	Name           string               `json:"name"`
	Size           int64                `json:"size,omitempty"`
	MimeType       string               `json:"mimeType,omitempty"`
	ArtifactID     string               `json:"artifact_id,omitempty"`
	FileID         string               `json:"file_id,omitempty"`
	Version        int                  `json:"version,omitempty"`
	Status         string               `json:"status,omitempty"`
	PreviewKind    string               `json:"preview_kind,omitempty"`
	PreviewURL     string               `json:"preview_url,omitempty"`
	DownloadURL    string               `json:"download_url,omitempty"`
	ChannelID      string               `json:"channel_id,omitempty"`
	CreatedAt      *time.Time           `json:"created_at,omitempty"`
	Uploader       *domain.User         `json:"uploader,omitempty"`
	CommentCount   int64                `json:"comment_count,omitempty"`
	ShareCount     int64                `json:"share_count,omitempty"`
	Starred        bool                 `json:"starred,omitempty"`
	Tags           []string             `json:"tags,omitempty"`
	KnowledgeState string               `json:"knowledge_state,omitempty"`
	SourceKind     string               `json:"source_kind,omitempty"`
	Summary        string               `json:"summary,omitempty"`
	IsArchived     bool                 `json:"is_archived,omitempty"`
	RetentionDays  int                  `json:"retention_days,omitempty"`
	ExpiresAt      *time.Time           `json:"expires_at,omitempty"`
	File           *fileAssetResponse   `json:"file,omitempty"`
	Preview        *filePreviewResponse `json:"preview,omitempty"`
}

var CollabSnapshotPath = agentcollab.DefaultPath()

type messageMetadata struct {
	Reactions   []reactionSummary   `json:"reactions,omitempty"`
	Attachments []messageAttachment `json:"attachments,omitempty"`
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
	user = derivePresence(user)
	user.AIInsight = buildUserInsight(user)
	return user
}

func derivePresence(user domain.User) domain.User {
	now := time.Now().UTC()

	if user.StatusExpiresAt != nil && now.After(*user.StatusExpiresAt) {
		user.StatusText = ""
		user.StatusEmoji = ""
		user.StatusExpiresAt = nil
	}

	if user.PresenceExpiresAt != nil && now.After(*user.PresenceExpiresAt) {
		if user.LastSeenAt != nil && now.Sub(*user.LastSeenAt) > 10*time.Minute {
			user.Status = "offline"
		} else if user.Status != "offline" {
			user.Status = "away"
		}
	}

	if strings.TrimSpace(user.StatusText) == "" {
		switch user.Status {
		case "online":
			user.StatusText = "Active now"
		case "busy":
			user.StatusText = "Busy"
		case "away":
			if user.LastSeenAt != nil {
				user.StatusText = "Last active " + humanizeLastSeen(*user.LastSeenAt, now)
			} else {
				user.StatusText = "Away"
			}
		default:
			if user.LastSeenAt != nil {
				user.StatusText = "Last active " + humanizeLastSeen(*user.LastSeenAt, now)
			} else {
				user.StatusText = "Offline"
			}
		}
	}

	return user
}

func humanizeLastSeen(lastSeen, now time.Time) string {
	diff := now.Sub(lastSeen)
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return strconv.Itoa(int(diff.Minutes())) + "m ago"
	case diff < 24*time.Hour:
		return strconv.Itoa(int(diff.Hours())) + "h ago"
	default:
		return strconv.Itoa(int(diff.Hours()/24)) + "d ago"
	}
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

func loadChannelPreference(channelID, userID string) domain.ChannelPreference {
	prefs := domain.ChannelPreference{
		ChannelID:         channelID,
		UserID:            userID,
		NotificationLevel: "all",
		IsMuted:           false,
	}
	_ = db.DB.Where("channel_id = ? AND user_id = ?", channelID, userID).First(&prefs).Error
	if strings.TrimSpace(prefs.NotificationLevel) == "" {
		prefs.NotificationLevel = "all"
	}
	return prefs
}

func isValidChannelNotificationLevel(level string) bool {
	switch level {
	case "all", "mentions", "none":
		return true
	default:
		return false
	}
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

	attachments, err := buildMessageAttachments(message.ID)
	if err != nil {
		return nil, err
	}
	meta.Attachments = attachments

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

func buildMessageAttachments(messageID string) ([]messageAttachment, error) {
	attachments := make([]messageAttachment, 0)

	var artifactRefs []domain.MessageArtifactReference
	if err := db.DB.Where("message_id = ?", messageID).Order("artifact_id asc").Find(&artifactRefs).Error; err != nil {
		return nil, err
	}
	for _, ref := range artifactRefs {
		var artifact domain.Artifact
		if err := db.DB.First(&artifact, "id = ?", ref.ArtifactID).Error; err != nil {
			continue
		}
		attachments = append(attachments, messageAttachment{
			ID:         "artifact:" + artifact.ID,
			Kind:       "artifact",
			Type:       "link",
			URL:        "/workspace?artifactId=" + artifact.ID,
			Name:       artifact.Title,
			ArtifactID: artifact.ID,
			Version:    artifact.Version,
			Status:     artifact.Status,
		})
	}

	var fileRefs []domain.MessageFileAttachment
	if err := db.DB.Where("message_id = ?", messageID).Order("file_id asc").Find(&fileRefs).Error; err != nil {
		return nil, err
	}
	for _, ref := range fileRefs {
		var asset domain.FileAsset
		if err := db.DB.First(&asset, "id = ?", ref.FileID).Error; err != nil {
			continue
		}
		attachments = append(attachments, hydrateMessageFileAttachment(asset))
	}

	if len(attachments) == 0 {
		return nil, nil
	}

	return attachments, nil
}

func hydrateMessageFileAttachment(asset domain.FileAsset) messageAttachment {
	file := hydrateFileAssetResponse(asset)
	preview := buildFilePreview(asset)
	createdAt := asset.CreatedAt

	attachmentType := "file"
	if strings.HasPrefix(strings.ToLower(file.ContentType), "image/") {
		attachmentType = "image"
	}

	return messageAttachment{
		ID:             "file:" + asset.ID,
		Kind:           "file",
		Type:           attachmentType,
		URL:            file.URL,
		Name:           asset.Name,
		Size:           asset.SizeBytes,
		MimeType:       file.ContentType,
		FileID:         asset.ID,
		PreviewKind:    file.PreviewKind,
		PreviewURL:     file.PreviewURL,
		DownloadURL:    file.URL,
		ChannelID:      asset.ChannelID,
		CreatedAt:      &createdAt,
		Uploader:       file.Uploader,
		CommentCount:   file.CommentCount,
		ShareCount:     file.ShareCount,
		Starred:        file.Starred,
		Tags:           file.Tags,
		KnowledgeState: asset.KnowledgeState,
		SourceKind:     asset.SourceKind,
		Summary:        asset.Summary,
		IsArchived:     asset.IsArchived,
		RetentionDays:  asset.RetentionDays,
		ExpiresAt:      asset.ExpiresAt,
		File:           &file,
		Preview:        &preview,
	}
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
	IsRead     bool        `json:"is_read"`
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
				ID:         "activity-reaction-" + strconv.FormatUint(uint64(reaction.ID), 10),
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

	activities = append(activities, buildStructuredListActivities(currentUser)...)
	activities = append(activities, buildStructuredToolRunActivities(currentUser)...)
	activities = append(activities, buildStructuredFileActivities(currentUser)...)

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].OccurredAt.After(activities[j].OccurredAt)
	})

	if len(activities) > 50 {
		activities = activities[:50]
	}

	if len(activities) == 0 {
		return activities
	}

	itemIDs := make([]string, 0, len(activities))
	for _, item := range activities {
		itemIDs = append(itemIDs, item.ID)
	}

	var reads []domain.NotificationRead
	db.DB.Where("user_id = ? AND item_id IN ?", currentUser.ID, itemIDs).Find(&reads)
	readSet := make(map[string]struct{}, len(reads))
	for _, read := range reads {
		readSet[read.ItemID] = struct{}{}
	}

	for idx := range activities {
		_, ok := readSet[activities[idx].ID]
		activities[idx].IsRead = ok
	}

	return activities
}

func buildStructuredListActivities(currentUser domain.User) []activityItem {
	var rows []domain.WorkspaceListItem
	_ = db.DB.Where("is_completed = ? AND assigned_to = ?", true, currentUser.ID).Order("coalesce(completed_at, updated_at) desc, id desc").Find(&rows).Error

	items := make([]activityItem, 0, len(rows))
	for _, row := range rows {
		var list domain.WorkspaceList
		if err := db.DB.First(&list, "id = ?", row.ListID).Error; err != nil {
			continue
		}
		var actor domain.User
		if err := db.DB.First(&actor, "id = ?", row.CreatedBy).Error; err != nil {
			continue
		}
		var channel domain.Channel
		_ = db.DB.First(&channel, "id = ?", list.ChannelID).Error

		occurredAt := row.UpdatedAt
		if row.CompletedAt != nil {
			occurredAt = *row.CompletedAt
		}

		items = append(items, activityItem{
			ID:      "activity-list-completed-" + strconv.FormatUint(uint64(row.ID), 10),
			Type:    "list_completed",
			User:    enrichUser(actor),
			Channel: channel,
			Message: gin.H{
				"list_id":      list.ID,
				"list_title":   list.Title,
				"item_id":      row.ID,
				"content":      row.Content,
				"is_completed": row.IsCompleted,
			},
			Target:     list.Title,
			Summary:    actor.Name + " completed a checklist item in " + list.Title,
			OccurredAt: occurredAt,
		})
	}
	return items
}

func buildStructuredToolRunActivities(currentUser domain.User) []activityItem {
	channelIDs := loadCurrentUserChannelIDs(currentUser.ID)

	var runs []domain.ToolRun
	_ = db.DB.Order("started_at desc").Find(&runs).Error

	items := make([]activityItem, 0, len(runs))
	for _, run := range runs {
		response := hydrateToolRun(run, false)
		if run.TriggeredBy != currentUser.ID && (response.ChannelID == "" || !containsString(channelIDs, response.ChannelID)) {
			continue
		}

		var actor domain.User
		if err := db.DB.First(&actor, "id = ?", run.TriggeredBy).Error; err != nil {
			continue
		}

		var channel domain.Channel
		if response.ChannelID != "" {
			_ = db.DB.First(&channel, "id = ?", response.ChannelID).Error
		}

		occurredAt := run.StartedAt
		if run.CompletedAt != nil {
			occurredAt = *run.CompletedAt
		}

		summary := actor.Name + " ran " + response.ToolName
		if response.Status != "" {
			summary = actor.Name + " completed " + response.ToolName
		}

		items = append(items, activityItem{
			ID:      "activity-tool-run-" + run.ID,
			Type:    "tool_run",
			User:    enrichUser(actor),
			Channel: channel,
			Message: gin.H{
				"id":          run.ID,
				"tool_id":     run.ToolID,
				"tool_name":   response.ToolName,
				"channel_id":  response.ChannelID,
				"status":      response.Status,
				"duration_ms": response.DurationMS,
			},
			Target:     response.ToolName,
			Summary:    summary,
			OccurredAt: occurredAt,
		})
	}
	return items
}

func buildStructuredFileActivities(currentUser domain.User) []activityItem {
	channelIDs := loadCurrentUserChannelIDs(currentUser.ID)
	if len(channelIDs) == 0 {
		return nil
	}

	var files []domain.FileAsset
	_ = db.DB.Where("channel_id IN ? AND uploader_id <> ?", channelIDs, currentUser.ID).Order("created_at desc").Find(&files).Error

	items := make([]activityItem, 0, len(files))
	for _, file := range files {
		var actor domain.User
		if err := db.DB.First(&actor, "id = ?", file.UploaderID).Error; err != nil {
			continue
		}
		var channel domain.Channel
		_ = db.DB.First(&channel, "id = ?", file.ChannelID).Error

		items = append(items, activityItem{
			ID:      "activity-file-uploaded-" + file.ID,
			Type:    "file_uploaded",
			User:    enrichUser(actor),
			Channel: channel,
			Message: gin.H{
				"id":         file.ID,
				"channel_id": file.ChannelID,
				"name":       file.Name,
				"type":       file.ContentType,
				"size":       file.SizeBytes,
			},
			Target:     file.Name,
			Summary:    actor.Name + " uploaded a file in #" + channel.Name,
			OccurredAt: file.CreatedAt,
		})
	}
	return items
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
	if department := strings.TrimSpace(c.Query("department")); department != "" {
		query = query.Where("department = ?", department)
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	if timezone := strings.TrimSpace(c.Query("timezone")); timezone != "" {
		query = query.Where("timezone = ?", timezone)
	}
	if q := strings.TrimSpace(strings.ToLower(c.Query("q"))); q != "" {
		like := "%" + q + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(title) LIKE ? OR LOWER(department) LIKE ?", like, like, like, like)
	}
	if userGroupID := strings.TrimSpace(c.Query("user_group_id")); userGroupID != "" {
		var memberIDs []string
		if err := db.DB.Model(&domain.UserGroupMember{}).Where("user_group_id = ?", userGroupID).Pluck("user_id", &memberIDs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user group members"})
			return
		}
		if len(memberIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{"users": []domain.User{}})
			return
		}
		query = query.Where("id IN ?", memberIDs)
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
		"comm_log":           snapshot.CommLog,
		"members":            snapshot.Members,
		"task_board":         snapshot.TaskBoard,
	})
}

func GetAgentCollabMembers(c *gin.Context) {
	snapshot, err := agentcollab.ReadSnapshot(CollabSnapshotPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read agent collab members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": snapshot.Members})
}

func CreateAgentCollabCommLog(c *gin.Context) {
	var input struct {
		From    string `json:"from" binding:"required"`
		To      string `json:"to"`
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := agentcollab.AppendCommLogEntry(CollabSnapshotPath, input.From, input.To, input.Title, input.Content, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to append agent collab comm log"})
		return
	}

	if RealtimeHub != nil {
		snapshot, err := agentcollab.ReadSnapshot(CollabSnapshotPath)
		if err == nil {
			_ = RealtimeHub.Broadcast(realtime.Event{
				ID:        "evt_" + time.Now().Format("20060102150405.000000"),
				Type:      "agent_collab.sync",
				ChannelID: "ch-collab",
				EntityID:  entry.ID,
				TS:        time.Now().UTC().Format(time.RFC3339Nano),
				Payload: gin.H{
					"active_superpowers": snapshot.ActiveSuperpowers,
					"comm_log":           snapshot.CommLog,
					"entry":              entry,
					"members":            snapshot.Members,
					"task_board":         snapshot.TaskBoard,
				},
			})
		}
	}

	c.JSON(http.StatusCreated, gin.H{"entry": entry})
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
		ID:             ids.NewPrefixedUUID("agent"),
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

func CreateChannel(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		WorkspaceID string `json:"workspace_id" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channelType := input.Type
	if channelType == "" {
		channelType = "public"
	}

	channel := domain.Channel{
		ID:          ids.NewPrefixedUUID("ch"),
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Type:        channelType,
		Description: input.Description,
		MemberCount: 1,
	}
	if err := db.DB.Create(&channel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create channel"})
		return
	}

	member := domain.ChannelMember{
		ChannelID: channel.ID,
		UserID:    currentUser.ID,
		Role:      "owner",
		CreatedAt: time.Now().UTC(),
	}
	if err := db.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add creator to channel"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"channel": channel})
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

func GetChannelPreferences(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	prefs := loadChannelPreference(channel.ID, currentUser.ID)
	c.JSON(http.StatusOK, gin.H{"preferences": prefs})
}

func PatchChannelPreferences(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	var input struct {
		NotificationLevel *string `json:"notification_level"`
		IsMuted           *bool   `json:"is_muted"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prefs := loadChannelPreference(channel.ID, currentUser.ID)
	if input.NotificationLevel != nil {
		level := strings.TrimSpace(*input.NotificationLevel)
		if !isValidChannelNotificationLevel(level) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "notification_level must be all, mentions, or none"})
			return
		}
		prefs.NotificationLevel = level
	}
	if input.IsMuted != nil {
		prefs.IsMuted = *input.IsMuted
	}
	prefs.UpdatedAt = time.Now().UTC()

	if prefs.ID == 0 {
		prefs.CreatedAt = prefs.UpdatedAt
		if err := db.DB.Create(&prefs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create channel preferences"})
			return
		}
	} else if err := db.DB.Save(&prefs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update channel preferences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preferences": prefs})
}

func LeaveChannel(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	if err := db.DB.Where("channel_id = ? AND user_id = ?", channel.ID, currentUser.ID).Delete(&domain.ChannelMember{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to leave channel"})
		return
	}
	var memberCount int64
	db.DB.Model(&domain.ChannelMember{}).Where("channel_id = ?", channel.ID).Count(&memberCount)
	db.DB.Model(&channel).Update("member_count", int(memberCount))
	channel.MemberCount = int(memberCount)

	c.JSON(http.StatusOK, gin.H{"left": true, "channel": channel})
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
		ID:          ids.NewPrefixedUUID("invite"),
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
		ID:        ids.NewPrefixedUUID("dm"),
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
		ID:               ids.NewPrefixedUUID("dm-msg"),
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

func MarkNotificationsRead(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		ItemIDs []string `json:"item_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	for _, itemID := range input.ItemIDs {
		if itemID == "" {
			continue
		}
		read := domain.NotificationRead{
			UserID: currentUser.ID,
			ItemID: itemID,
			ReadAt: now,
		}
		if err := db.DB.Where("user_id = ? AND item_id = ?", currentUser.ID, itemID).
			Assign(domain.NotificationRead{ReadAt: now}).
			FirstOrCreate(&read).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist notification read state"})
			return
		}
	}

	if RealtimeHub != nil {
		_ = RealtimeHub.Broadcast(realtime.Event{
			ID:       "evt_" + time.Now().Format("20060102150405.000000"),
			Type:     "notifications.read",
			EntityID: currentUser.ID,
			TS:       time.Now().UTC().Format(time.RFC3339Nano),
			Payload:  gin.H{"user_id": currentUser.ID, "item_ids": input.ItemIDs, "read": true},
		})
	}

	c.JSON(http.StatusOK, gin.H{"read": true, "item_ids": input.ItemIDs})
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

func GetPresence(c *gin.Context) {
	var users []domain.User
	query := db.DB.Order("name asc")
	if channelID := c.Query("channel_id"); channelID != "" {
		var memberIDs []string
		if err := db.DB.Model(&domain.ChannelMember{}).Where("channel_id = ?", channelID).Pluck("user_id", &memberIDs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load channel presence members"})
			return
		}
		if len(memberIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{"users": []domain.User{}})
			return
		}
		query = query.Where("id IN ?", memberIDs)
	}
	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load presence"})
		return
	}
	for idx := range users {
		users[idx] = enrichUser(users[idx])
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func GetStarredChannels(c *gin.Context) {
	var channels []domain.Channel
	if err := db.DB.Where("is_starred = ?", true).Order("name asc").Find(&channels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load starred channels"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

func UpdatePresence(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		Status     string `json:"status" binding:"required,oneof=online away busy offline"`
		StatusText string `json:"status_text"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	currentUser.Status = input.Status
	currentUser.StatusText = input.StatusText
	currentUser.LastSeenAt = &now
	if input.Status == "offline" {
		currentUser.PresenceExpiresAt = nil
	} else {
		expiresAt := now.Add(2 * time.Minute)
		currentUser.PresenceExpiresAt = &expiresAt
	}
	if err := db.DB.Save(&currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update presence"})
		return
	}

	user := enrichUser(currentUser)
	if RealtimeHub != nil {
		_ = RealtimeHub.Broadcast(realtime.Event{
			ID:          "evt_" + time.Now().Format("20060102150405.000000"),
			Type:        "presence.updated",
			WorkspaceID: currentUser.OrganizationID,
			EntityID:    currentUser.ID,
			TS:          time.Now().UTC().Format(time.RFC3339Nano),
			Payload:     gin.H{"user": user},
		})
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func HeartbeatPresence(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	if input.Status != "" {
		currentUser.Status = input.Status
	}
	if currentUser.Status == "" {
		currentUser.Status = "online"
	}
	currentUser.LastSeenAt = &now
	expiresAt := now.Add(2 * time.Minute)
	currentUser.PresenceExpiresAt = &expiresAt
	if strings.TrimSpace(currentUser.StatusText) == "" || strings.HasPrefix(currentUser.StatusText, "Last active") {
		currentUser.StatusText = ""
	}

	if err := db.DB.Save(&currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh presence"})
		return
	}

	user := enrichUser(currentUser)
	if RealtimeHub != nil {
		_ = RealtimeHub.Broadcast(realtime.Event{
			ID:          "evt_" + time.Now().Format("20060102150405.000000"),
			Type:        "presence.updated",
			WorkspaceID: currentUser.OrganizationID,
			EntityID:    currentUser.ID,
			TS:          time.Now().UTC().Format(time.RFC3339Nano),
			Payload:     gin.H{"user": user},
		})
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func ToggleChannelStar(c *gin.Context) {
	channelID := c.Param("id")

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", channelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	channel.IsStarred = !channel.IsStarred
	if err := db.DB.Save(&channel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update channel star state"})
		return
	}

	if RealtimeHub != nil {
		_ = RealtimeHub.Broadcast(realtime.Event{
			ID:          "evt_" + time.Now().Format("20060102150405.000000"),
			Type:        "channel.updated",
			WorkspaceID: channel.WorkspaceID,
			ChannelID:   channel.ID,
			EntityID:    channel.ID,
			TS:          time.Now().UTC().Format(time.RFC3339Nano),
			Payload:     gin.H{"channel": channel, "is_starred": channel.IsStarred, "field": "is_starred"},
		})
	}

	c.JSON(http.StatusOK, gin.H{"channel": channel, "is_starred": channel.IsStarred})
}

func UpdateTyping(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		ChannelID string `json:"channel_id"`
		ThreadID  string `json:"thread_id"`
		DMID      string `json:"dm_id"`
		IsTyping  bool   `json:"is_typing"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.ChannelID == "" && input.DMID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id or dm_id is required"})
		return
	}

	workspaceID := currentUser.OrganizationID
	if input.ChannelID != "" {
		var channel domain.Channel
		if err := db.DB.First(&channel, "id = ?", input.ChannelID).Error; err == nil {
			workspaceID = channel.WorkspaceID
		}
	}

	payload := gin.H{
		"user_id":    currentUser.ID,
		"channel_id": input.ChannelID,
		"thread_id":  input.ThreadID,
		"dm_id":      input.DMID,
		"is_typing":  input.IsTyping,
	}

	if RealtimeHub != nil {
		_ = RealtimeHub.Broadcast(realtime.Event{
			ID:          "evt_" + time.Now().Format("20060102150405.000000"),
			Type:        "typing.updated",
			WorkspaceID: workspaceID,
			ChannelID:   input.ChannelID,
			EntityID:    currentUser.ID,
			TS:          time.Now().UTC().Format(time.RFC3339Nano),
			Payload:     payload,
		})
	}

	c.JSON(http.StatusOK, gin.H{"typing": payload})
}

func GetPins(c *gin.Context) {
	type pinItem struct {
		Message domain.Message `json:"message"`
		Channel domain.Channel `json:"channel"`
		User    domain.User    `json:"user"`
	}

	query := db.DB.Where("is_pinned = ?", true)
	if channelID := strings.TrimSpace(c.Query("channel_id")); channelID != "" {
		query = query.Where("channel_id = ?", channelID)
	}

	var messages []domain.Message
	if err := query.Order("created_at desc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load pins"})
		return
	}

	items := make([]pinItem, 0, len(messages))
	for _, message := range messages {
		refreshed, err := refreshMessageMetadata(message.ID)
		if err == nil && refreshed != nil {
			message = *refreshed
		}

		var channel domain.Channel
		if err := db.DB.First(&channel, "id = ?", message.ChannelID).Error; err != nil {
			continue
		}

		var user domain.User
		if err := db.DB.First(&user, "id = ?", message.UserID).Error; err != nil {
			continue
		}

		items = append(items, pinItem{
			Message: message,
			Channel: channel,
			User:    enrichUser(user),
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func GetDrafts(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	scope := c.Query("scope")
	var drafts []domain.Draft
	query := db.DB.Where("user_id = ?", currentUser.ID).Order("updated_at desc")
	if scope != "" {
		query = query.Where("scope = ?", scope)
	}
	if err := query.Find(&drafts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load drafts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"drafts": drafts})
}

func PutDraft(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	scope := c.Param("scope")
	if scope == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scope is required"})
		return
	}

	var input struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	draft := domain.Draft{
		UserID:    currentUser.ID,
		Scope:     scope,
		Content:   input.Content,
		UpdatedAt: now,
	}

	if err := db.DB.Where("user_id = ? AND scope = ?", currentUser.ID, scope).
		Assign(domain.Draft{Content: input.Content, UpdatedAt: now}).
		FirstOrCreate(&draft).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save draft"})
		return
	}
	if err := db.DB.First(&draft, "user_id = ? AND scope = ?", currentUser.ID, scope).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load saved draft"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"draft": draft})
}

func DeleteDraft(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	scope := c.Param("scope")
	if scope == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scope is required"})
		return
	}

	if err := db.DB.Where("user_id = ? AND scope = ?", currentUser.ID, scope).Delete(&domain.Draft{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete draft"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true, "scope": scope})
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
		Channels  []gin.H `json:"channels"`
		Users     []gin.H `json:"users"`
		Messages  []gin.H `json:"messages"`
		DMs       []gin.H `json:"dms"`
		Artifacts []gin.H `json:"artifacts"`
		Files     []gin.H `json:"files"`
	}

	results := searchResults{
		Channels:  []gin.H{},
		Users:     []gin.H{},
		Messages:  []gin.H{},
		DMs:       []gin.H{},
		Artifacts: []gin.H{},
		Files:     []gin.H{},
	}

	var channels []domain.Channel
	db.DB.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", needle, needle).Order("name asc").Limit(10).Find(&channels)
	for _, channel := range channels {
		matchReason := "name"
		if !strings.Contains(strings.ToLower(channel.Name), strings.ToLower(query)) {
			matchReason = "description"
		}
		results.Channels = append(results.Channels, gin.H{
			"id":           channel.ID,
			"workspace_id": channel.WorkspaceID,
			"name":         channel.Name,
			"type":         channel.Type,
			"description":  channel.Description,
			"match_reason": matchReason,
		})
	}

	var users []domain.User
	db.DB.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", needle, needle).Order("name asc").Limit(10).Find(&users)
	for _, user := range users {
		enriched := enrichUser(user)
		matchReason := "name"
		if !strings.Contains(strings.ToLower(user.Name), strings.ToLower(query)) {
			matchReason = "email"
		}
		results.Users = append(results.Users, gin.H{
			"id":           enriched.ID,
			"name":         enriched.Name,
			"email":        enriched.Email,
			"avatar":       enriched.Avatar,
			"status":       enriched.Status,
			"status_text":  enriched.StatusText,
			"last_seen_at": enriched.LastSeenAt,
			"ai_insight":   enriched.AIInsight,
			"match_reason": matchReason,
		})
	}

	var messages []domain.Message
	db.DB.Where("LOWER(content) LIKE ?", needle).Order("created_at desc").Limit(10).Find(&messages)
	for idx := range messages {
		refreshed, err := refreshMessageMetadata(messages[idx].ID)
		if err == nil && refreshed != nil {
			messages[idx] = *refreshed
		}
		message := messages[idx]
		results.Messages = append(results.Messages, gin.H{
			"id":         message.ID,
			"channel_id": message.ChannelID,
			"user_id":    message.UserID,
			"content":    message.Content,
			"thread_id":  message.ThreadID,
			"created_at": message.CreatedAt,
			"metadata":   message.Metadata,
			"snippet":    buildSearchSnippet(message.Content, query),
		})
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

	var artifacts []domain.Artifact
	db.DB.Where("LOWER(title) LIKE ? OR LOWER(content) LIKE ?", needle, needle).Order("updated_at desc").Limit(10).Find(&artifacts)
	for _, artifact := range artifacts {
		matchReason := "title"
		if !strings.Contains(strings.ToLower(artifact.Title), strings.ToLower(query)) {
			matchReason = "content"
		}
		results.Artifacts = append(results.Artifacts, gin.H{
			"id":           artifact.ID,
			"channel_id":   artifact.ChannelID,
			"title":        artifact.Title,
			"type":         artifact.Type,
			"status":       artifact.Status,
			"version":      artifact.Version,
			"match_reason": matchReason,
			"snippet":      buildSearchSnippet(artifact.Content, query),
		})
	}

	var files []domain.FileAsset
	db.DB.Where("LOWER(name) LIKE ? OR LOWER(content_type) LIKE ?", needle, needle).Order("created_at desc").Limit(10).Find(&files)
	for _, file := range files {
		matchReason := "name"
		if !strings.Contains(strings.ToLower(file.Name), strings.ToLower(query)) {
			matchReason = "content_type"
		}
		results.Files = append(results.Files, gin.H{
			"id":           file.ID,
			"channel_id":   file.ChannelID,
			"name":         file.Name,
			"content_type": file.ContentType,
			"size_bytes":   file.SizeBytes,
			"url":          "/api/v1/files/" + file.ID + "/content",
			"match_reason": matchReason,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": results,
	})
}

func SearchSuggestions(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}

	needle := "%" + strings.ToLower(query) + "%"
	suggestions := make([]gin.H, 0, 12)
	seen := map[string]struct{}{}

	var channels []domain.Channel
	db.DB.Where("LOWER(name) LIKE ?", needle).Order("name asc").Limit(4).Find(&channels)
	for _, channel := range channels {
		key := "channel:" + channel.ID
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		suggestions = append(suggestions, gin.H{
			"type":  "channel",
			"id":    channel.ID,
			"label": "#" + channel.Name,
			"hint":  defaultString(channel.Description, "Channel"),
		})
	}

	var users []domain.User
	db.DB.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", needle, needle).Order("name asc").Limit(4).Find(&users)
	for _, user := range users {
		key := "user:" + user.ID
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		suggestions = append(suggestions, gin.H{
			"type":  "user",
			"id":    user.ID,
			"label": user.Name,
			"hint":  user.Email,
		})
	}

	var messages []domain.Message
	db.DB.Where("LOWER(content) LIKE ?", needle).Order("created_at desc").Limit(4).Find(&messages)
	for _, message := range messages {
		key := "message:" + message.ID
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		suggestions = append(suggestions, gin.H{
			"type":  "message",
			"id":    message.ID,
			"label": buildSearchSnippet(message.Content, query),
			"hint":  "Message result",
		})
	}

	var artifacts []domain.Artifact
	db.DB.Where("LOWER(title) LIKE ? OR LOWER(content) LIKE ?", needle, needle).Order("updated_at desc").Limit(4).Find(&artifacts)
	for _, artifact := range artifacts {
		key := "artifact:" + artifact.ID
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		suggestions = append(suggestions, gin.H{
			"type":  "artifact",
			"id":    artifact.ID,
			"label": artifact.Title,
			"hint":  "Canvas artifact",
		})
	}

	var files []domain.FileAsset
	db.DB.Where("LOWER(name) LIKE ? OR LOWER(content_type) LIKE ?", needle, needle).Order("created_at desc").Limit(4).Find(&files)
	for _, file := range files {
		key := "file:" + file.ID
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		suggestions = append(suggestions, gin.H{
			"type":  "file",
			"id":    file.ID,
			"label": file.Name,
			"hint":  defaultString(file.ContentType, "File"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"query":       query,
		"suggestions": suggestions,
	})
}

func IntelligentSearch(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}

	needle := "%" + strings.ToLower(query) + "%"
	tokens := strings.Fields(strings.ToLower(query))

	type rankedResult struct {
		Type   string  `json:"type"`
		ID     string  `json:"id"`
		Label  string  `json:"label"`
		Reason string  `json:"reason"`
		Score  float64 `json:"score"`
	}

	ranked := make([]rankedResult, 0, 20)

	var channels []domain.Channel
	db.DB.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", needle, needle).Order("name asc").Limit(8).Find(&channels)
	for _, channel := range channels {
		text := strings.ToLower(channel.Name + " " + channel.Description)
		ranked = append(ranked, rankedResult{
			Type:   "channel",
			ID:     channel.ID,
			Label:  "#" + channel.Name,
			Reason: "Matched channel name or description",
			Score:  computeIntelligentSearchScore(text, tokens) + 0.5,
		})
	}

	var messages []domain.Message
	db.DB.Where("LOWER(content) LIKE ?", needle).Order("created_at desc").Limit(8).Find(&messages)
	for idx := range messages {
		refreshed, err := refreshMessageMetadata(messages[idx].ID)
		if err == nil && refreshed != nil {
			messages[idx] = *refreshed
		}
		ranked = append(ranked, rankedResult{
			Type:   "message",
			ID:     messages[idx].ID,
			Label:  buildSearchSnippet(messages[idx].Content, query),
			Reason: "Matched message content",
			Score:  computeIntelligentSearchScore(strings.ToLower(messages[idx].Content), tokens) + 0.7,
		})
	}

	var artifacts []domain.Artifact
	db.DB.Where("LOWER(title) LIKE ? OR LOWER(content) LIKE ?", needle, needle).Order("updated_at desc").Limit(8).Find(&artifacts)
	for _, artifact := range artifacts {
		ranked = append(ranked, rankedResult{
			Type:   "artifact",
			ID:     artifact.ID,
			Label:  artifact.Title,
			Reason: "Matched artifact title or content",
			Score:  computeIntelligentSearchScore(strings.ToLower(artifact.Title+" "+artifact.Content), tokens) + 1.0,
		})
	}

	var files []domain.FileAsset
	db.DB.Where("LOWER(name) LIKE ? OR LOWER(content_type) LIKE ?", needle, needle).Order("created_at desc").Limit(8).Find(&files)
	for _, file := range files {
		ranked = append(ranked, rankedResult{
			Type:   "file",
			ID:     file.ID,
			Label:  file.Name,
			Reason: "Matched file name or MIME type",
			Score:  computeIntelligentSearchScore(strings.ToLower(file.Name+" "+file.ContentType), tokens) + 0.4,
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return ranked[i].Label < ranked[j].Label
		}
		return ranked[i].Score > ranked[j].Score
	})

	if len(ranked) > 12 {
		ranked = ranked[:12]
	}

	c.JSON(http.StatusOK, gin.H{
		"query":  query,
		"ranked": ranked,
	})
}

func computeIntelligentSearchScore(text string, tokens []string) float64 {
	score := 0.0
	for _, token := range tokens {
		if token == "" {
			continue
		}
		if strings.Contains(text, token) {
			score += 1.0
			if strings.HasPrefix(text, token) {
				score += 0.25
			}
		}
	}
	return score
}

func buildSearchSnippet(content, query string) string {
	plain := strings.TrimSpace(content)
	if plain == "" {
		return ""
	}
	lowerContent := strings.ToLower(plain)
	lowerQuery := strings.ToLower(strings.TrimSpace(query))
	idx := strings.Index(lowerContent, lowerQuery)
	if idx == -1 {
		if len(plain) <= 120 {
			return plain
		}
		return plain[:117] + "..."
	}

	start := idx - 24
	if start < 0 {
		start = 0
	}
	end := idx + len(lowerQuery) + 48
	if end > len(plain) {
		end = len(plain)
	}

	snippet := plain[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(plain) {
		snippet += "..."
	}
	return snippet
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

func GetMessageFiles(c *gin.Context) {
	messageID := c.Param("id")

	var message domain.Message
	if err := db.DB.First(&message, "id = ?", messageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	attachments, err := buildMessageAttachments(messageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load message attachments"})
		return
	}

	files := make([]messageAttachment, 0)
	for _, attachment := range attachments {
		if attachment.Kind == "file" {
			files = append(files, attachment)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message_id": message.ID,
		"files":      files,
	})
}

func CreateMessage(c *gin.Context) {
	var input struct {
		ChannelID   string   `json:"channel_id" binding:"required"`
		Content     string   `json:"content" binding:"required"`
		UserID      string   `json:"user_id" binding:"required"`
		ThreadID    string   `json:"thread_id"`
		ArtifactIDs []string `json:"artifact_ids"`
		FileIDs     []string `json:"file_ids"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := domain.Message{
		ID:        ids.NewPrefixedUUID("msg"),
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

	for _, artifactID := range input.ArtifactIDs {
		if strings.TrimSpace(artifactID) == "" {
			continue
		}
		ref := domain.MessageArtifactReference{
			MessageID:  msg.ID,
			ArtifactID: artifactID,
			CreatedAt:  time.Now().UTC(),
		}
		if err := db.DB.FirstOrCreate(&ref, domain.MessageArtifactReference{MessageID: msg.ID, ArtifactID: artifactID}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to attach artifact reference"})
			return
		}
	}

	for _, fileID := range input.FileIDs {
		if strings.TrimSpace(fileID) == "" {
			continue
		}
		ref := domain.MessageFileAttachment{
			MessageID: msg.ID,
			FileID:    fileID,
			CreatedAt: time.Now().UTC(),
		}
		if err := db.DB.FirstOrCreate(&ref, domain.MessageFileAttachment{MessageID: msg.ID, FileID: fileID}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to attach file reference"})
			return
		}
	}

	refreshed, err := refreshMessageMetadata(msg.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh message metadata"})
		return
	}
	msg = *refreshed

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
