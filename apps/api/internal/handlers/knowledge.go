package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
	"github.com/nikkofu/relay-agent-workspace/api/internal/knowledge"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

func ListKnowledgeEntities(c *gin.Context) {
	entities, err := knowledge.ListEntities(db.DB, c.Query("workspace_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load knowledge entities"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entities": entities})
}

func CreateKnowledgeEntity(c *gin.Context) {
	var input knowledge.CreateEntityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity, err := knowledge.CreateEntity(db.DB, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := broadcastKnowledgeEvent("knowledge.entity.created", entity.WorkspaceID, "", entity.ID, gin.H{"entity": entity}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast knowledge entity event"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"entity": entity})
}

func GetKnowledgeEntity(c *gin.Context) {
	entity, err := knowledge.GetEntity(db.DB, c.Param("id"))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

func UpdateKnowledgeEntity(c *gin.Context) {
	var input knowledge.UpdateEntityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity, err := knowledge.UpdateEntity(db.DB, c.Param("id"), input)
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	if err := broadcastKnowledgeEvent("knowledge.entity.updated", entity.WorkspaceID, "", entity.ID, gin.H{"entity": entity}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast knowledge entity event"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

func ListKnowledgeEntityRefs(c *gin.Context) {
	if _, err := knowledge.GetEntity(db.DB, c.Param("id")); err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	refs, err := knowledge.ListEntityRefs(db.DB, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load knowledge refs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"refs": refs})
}

func AddKnowledgeEntityRef(c *gin.Context) {
	var input knowledge.AddEntityRefInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ref, err := knowledge.AddEntityRef(db.DB, c.Param("id"), input)
	if err != nil {
		handleKnowledgeError(c, err, "failed to add knowledge ref")
		return
	}
	if err := broadcastKnowledgeEvent("knowledge.entity.ref.created", ref.WorkspaceID, "", ref.EntityID, gin.H{"ref": ref}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast knowledge ref event"})
		return
	}
	_ = emitKnowledgeEntitySpikeAlerts(ref.EntityID, "")
	c.JSON(http.StatusCreated, gin.H{"ref": ref})
}

func ListKnowledgeEntityTimeline(c *gin.Context) {
	if _, err := knowledge.GetEntity(db.DB, c.Param("id")); err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	events, err := knowledge.ListEntityTimeline(db.DB, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load knowledge timeline"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func AddKnowledgeEntityEvent(c *gin.Context) {
	var input knowledge.AddEntityEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if currentUser, err := getCurrentUser(); err == nil && strings.TrimSpace(input.ActorUserID) == "" {
		input.ActorUserID = currentUser.ID
	}
	event, err := knowledge.AppendEntityEvent(db.DB, c.Param("id"), input)
	if err != nil {
		handleKnowledgeError(c, err, "failed to add knowledge event")
		return
	}
	if err := broadcastKnowledgeEvent("knowledge.event.created", event.WorkspaceID, "", event.EntityID, gin.H{"event": event}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast knowledge event"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"event": event})
}

func IngestKnowledgeEvent(c *gin.Context) {
	var input knowledge.IngestEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if currentUser, err := getCurrentUser(); err == nil && strings.TrimSpace(input.ActorUserID) == "" {
		input.ActorUserID = currentUser.ID
	}
	event, err := knowledge.IngestEvent(db.DB, input)
	if err != nil {
		handleKnowledgeError(c, err, "failed to ingest knowledge event")
		return
	}
	if err := broadcastKnowledgeEvent("knowledge.event.created", event.WorkspaceID, "", event.EntityID, gin.H{"event": event}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast knowledge event"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"event": event})
}

func ListKnowledgeEntityLinks(c *gin.Context) {
	if _, err := knowledge.GetEntity(db.DB, c.Param("id")); err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	links, err := knowledge.ListEntityLinks(db.DB, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load knowledge links"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"links": links})
}

func AddKnowledgeEntityLink(c *gin.Context) {
	var input knowledge.AddEntityLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	link, err := knowledge.AddEntityLink(db.DB, input)
	if err != nil {
		handleKnowledgeError(c, err, "failed to add knowledge link")
		return
	}
	if err := broadcastKnowledgeEvent("knowledge.link.created", link.WorkspaceID, "", link.FromEntityID, gin.H{"link": link}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast knowledge link event"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"link": link})
}

func GetKnowledgeEntityGraph(c *gin.Context) {
	graph, err := knowledge.BuildEntityGraph(db.DB, c.Param("id"))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"graph": graph})
}

func GetChannelKnowledgeContext(c *gin.Context) {
	context, err := knowledge.GetChannelKnowledgeContext(db.DB, c.Param("id"), parseLimit(c.Query("limit"), 20))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load channel knowledge context"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"context": context})
}

func GetChannelKnowledgeSummary(c *gin.Context) {
	summary, err := knowledge.GetChannelKnowledgeSummary(
		db.DB,
		c.Param("id"),
		parseLimit(c.Query("limit"), 5),
		parseWindowDays(c.Query("days"), 7),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load channel knowledge summary"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

func SuggestKnowledgeEntities(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}

	suggestions, err := knowledge.SuggestEntities(db.DB, knowledge.SuggestEntitiesParams{
		Query:       query,
		ChannelID:   c.Query("channel_id"),
		WorkspaceID: c.Query("workspace_id"),
		Limit:       parseLimit(c.Query("limit"), 8),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load knowledge entity suggestions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":       query,
		"suggestions": suggestions,
	})
}

func MatchKnowledgeEntitiesInText(c *gin.Context) {
	var input struct {
		WorkspaceID string `json:"workspace_id"`
		Text        string `json:"text"`
		Limit       int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(input.Text) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
		return
	}

	matches, err := knowledge.MatchEntitiesInText(db.DB, knowledge.MatchEntitiesInput{
		WorkspaceID: input.WorkspaceID,
		Text:        input.Text,
		Limit:       input.Limit,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to match knowledge entities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"matches": matches})
}

func GetMyFollowedKnowledgeEntities(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	items, err := knowledge.ListFollowedEntities(db.DB, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load followed knowledge entities"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func FollowKnowledgeEntity(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	follow, err := knowledge.FollowEntity(db.DB, c.Param("id"), currentUser.ID)
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"follow": follow, "is_following": true})
}

func PatchMyKnowledgeFollow(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		NotificationLevel string `json:"notification_level"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	follow, err := knowledge.UpdateFollowNotificationLevel(db.DB, c.Param("id"), currentUser.ID, input.NotificationLevel)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge follow not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"follow": follow, "is_following": true})
}

func UnfollowKnowledgeEntity(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err := knowledge.UnfollowEntity(db.DB, c.Param("id"), currentUser.ID); err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"entity_id": c.Param("id"), "is_following": false})
}

func SearchMessagesByEntity(c *gin.Context) {
	entityID := strings.TrimSpace(c.Query("entity_id"))
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_id is required"})
		return
	}

	matches, err := knowledge.FindMessagesByEntity(db.DB, knowledge.EntityMessageSearchParams{
		EntityID:  entityID,
		ChannelID: c.Query("channel_id"),
		Limit:     parseLimit(c.Query("limit"), 20),
	})
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}

	items := make([]gin.H, 0, len(matches))
	for _, match := range matches {
		message := match.Message
		refreshed, err := refreshMessageMetadata(message.ID)
		if err == nil && refreshed != nil {
			message = *refreshed
		}
		items = append(items, gin.H{
			"id":            message.ID,
			"channel_id":    message.ChannelID,
			"user_id":       message.UserID,
			"content":       message.Content,
			"thread_id":     message.ThreadID,
			"created_at":    message.CreatedAt,
			"metadata":      message.Metadata,
			"snippet":       buildSearchSnippet(message.Content, match.EntityTitle),
			"entity_id":     match.EntityID,
			"entity_title":  match.EntityTitle,
			"match_sources": match.MatchSources,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id":  entityID,
		"channel_id": strings.TrimSpace(c.Query("channel_id")),
		"messages":   items,
	})
}

func GetKnowledgeEntityHover(c *gin.Context) {
	entity, err := knowledge.GetEntity(db.DB, c.Param("id"))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}

	hover, err := knowledge.GetEntityHoverSummary(
		db.DB,
		c.Param("id"),
		c.Query("channel_id"),
		parseWindowDays(c.Query("days"), 7),
	)
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entity": entity,
		"hover":  hover,
	})
}

func GetChannelKnowledgeDigest(c *gin.Context) {
	digest, err := knowledge.BuildChannelKnowledgeDigest(
		db.DB,
		c.Param("id"),
		c.DefaultQuery("window", "weekly"),
		parseLimit(c.Query("limit"), 5),
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to build channel knowledge digest"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"digest": digest})
}

func PublishChannelKnowledgeDigest(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		Window string `json:"window"`
		Limit  int    `json:"limit"`
		Pin    bool   `json:"pin"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	published, err := knowledge.PublishChannelDigest(db.DB, knowledge.PublishChannelDigestInput{
		ChannelID:  c.Param("id"),
		UserID:     currentUser.ID,
		Window:     defaultString(strings.TrimSpace(input.Window), "weekly"),
		Limit:      parseLimit(strconv.Itoa(input.Limit), 5),
		Pin:        input.Pin,
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to build channel knowledge digest"})
		return
	}

	if RealtimeHub != nil {
		if err := broadcastRealtimeEvent("message.created", published.Message, published.Message); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast digest message"})
			return
		}
		if err := broadcastKnowledgeDigestPublished(published); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to broadcast digest event"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": published.Message,
		"digest":  published.Digest,
	})
}

func GetChannelKnowledgeDigestSchedule(c *gin.Context) {
	schedule, err := knowledge.GetDigestSchedule(db.DB, c.Param("id"), time.Now().UTC())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load digest schedule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

func PutChannelKnowledgeDigestSchedule(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		Window     string `json:"window"`
		Timezone   string `json:"timezone"`
		DayOfWeek  int    `json:"day_of_week"`
		DayOfMonth int    `json:"day_of_month"`
		Hour       int    `json:"hour"`
		Minute     int    `json:"minute"`
		Limit      int    `json:"limit"`
		Pin        bool   `json:"pin"`
		IsEnabled  bool   `json:"is_enabled"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schedule, err := knowledge.UpsertDigestSchedule(db.DB, knowledge.UpsertDigestScheduleInput{
		ChannelID:  c.Param("id"),
		CreatedBy:  currentUser.ID,
		Window:     input.Window,
		Timezone:   input.Timezone,
		DayOfWeek:  input.DayOfWeek,
		DayOfMonth: input.DayOfMonth,
		Hour:       input.Hour,
		Minute:     input.Minute,
		Limit:      input.Limit,
		Pin:        input.Pin,
		IsEnabled:  input.IsEnabled,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

func PreviewChannelKnowledgeDigestSchedule(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		Window     string `json:"window"`
		Timezone   string `json:"timezone"`
		DayOfWeek  int    `json:"day_of_week"`
		DayOfMonth int    `json:"day_of_month"`
		Hour       int    `json:"hour"`
		Minute     int    `json:"minute"`
		Limit      int    `json:"limit"`
		Pin        bool   `json:"pin"`
		IsEnabled  bool   `json:"is_enabled"`
		Count      int    `json:"count"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preview, err := knowledge.PreviewDigestSchedule(db.DB, knowledge.UpsertDigestScheduleInput{
		ChannelID:  c.Param("id"),
		CreatedBy:  currentUser.ID,
		Window:     input.Window,
		Timezone:   input.Timezone,
		DayOfWeek:  input.DayOfWeek,
		DayOfMonth: input.DayOfMonth,
		Hour:       input.Hour,
		Minute:     input.Minute,
		Limit:      input.Limit,
		Pin:        input.Pin,
		IsEnabled:  input.IsEnabled,
	}, input.Count, time.Now().UTC())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preview": preview})
}

func DeleteChannelKnowledgeDigestSchedule(c *gin.Context) {
	if err := knowledge.DeleteDigestSchedule(db.DB, c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to delete digest schedule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true, "channel_id": c.Param("id")})
}

func GetKnowledgeInbox(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	items, err := knowledge.ListKnowledgeInbox(db.DB, knowledge.KnowledgeInboxParams{
		UserID: currentUser.ID,
		Scope:  c.DefaultQuery("scope", "all"),
		Limit:  parseLimit(c.Query("limit"), 20),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load knowledge inbox"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func GetKnowledgeInboxItem(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	item, err := knowledge.GetKnowledgeInboxItem(db.DB, currentUser.ID, c.Param("id"), parseLimit(c.Query("sample_limit"), 3))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge inbox item not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load knowledge inbox item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": gin.H{
		"id":              item.Item.ID,
		"channel":         item.Item.Channel,
		"message":         item.Item.Message,
		"digest":          item.Item.Digest,
		"is_read":         item.Item.IsRead,
		"occurred_at":     item.Item.OccurredAt,
		"entity_contexts": item.EntityContexts,
	}})
}

func handleKnowledgeError(c *gin.Context, err error, fallback string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "knowledge entity not found"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": fallback})
}

func handleKnowledgeNotFound(c *gin.Context, err error, message string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}

func parseLimit(raw string, fallback int) int {
	limit, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || limit <= 0 {
		return fallback
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func parseWindowDays(raw string, fallback int) int {
	days, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || days <= 0 {
		return fallback
	}
	if days > 30 {
		return 30
	}
	return days
}

func buildKnowledgeDigestMessageContent(digest knowledge.ChannelKnowledgeDigest) string {
	lines := []string{
		fmt.Sprintf("Knowledge digest (%s)", digest.Window),
		digest.Headline,
		digest.Summary,
	}
	for idx, movement := range digest.TopMovements {
		if idx >= 3 {
			break
		}
		lines = append(lines, fmt.Sprintf("- %s: %d recent refs (%+d vs previous window)", movement.EntityTitle, movement.RecentRefCount, movement.Delta))
	}
	return strings.Join(lines, "\n")
}

func broadcastKnowledgeDigestPublished(published knowledge.PublishedDigest) error {
	if RealtimeHub == nil {
		return nil
	}
	return RealtimeHub.Broadcast(realtime.Event{
		ID:          ids.NewPrefixedUUID("evt"),
		Type:        "knowledge.digest.published",
		WorkspaceID: published.Channel.WorkspaceID,
		ChannelID:   published.Channel.ID,
		TS:          time.Now().UTC().Format(time.RFC3339Nano),
		Payload: gin.H{
			"channel": published.Channel,
			"message": published.Message,
			"digest":  published.Digest,
		},
	})
}

func autoLinkKnowledgeForMessage(message domain.Message) {
	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", message.ChannelID).Error; err != nil {
		return
	}

	refs, err := knowledge.AutoLinkEntitiesForMessage(db.DB, channel.WorkspaceID, message.ID, message.Content)
	if err != nil {
		return
	}
	for _, ref := range refs {
		_ = broadcastKnowledgeEvent("knowledge.entity.ref.created", ref.WorkspaceID, channel.ID, ref.EntityID, gin.H{"ref": ref})
		_ = emitKnowledgeEntitySpikeAlerts(ref.EntityID, channel.ID)
	}
}

func autoLinkKnowledgeForFile(asset domain.FileAsset, content string) {
	var channel domain.Channel
	if asset.ChannelID == "" || db.DB.First(&channel, "id = ?", asset.ChannelID).Error != nil {
		return
	}

	refs, err := knowledge.AutoLinkEntitiesForFile(db.DB, channel.WorkspaceID, asset.ID, content)
	if err != nil {
		return
	}
	for _, ref := range refs {
		_ = broadcastKnowledgeEvent("knowledge.entity.ref.created", ref.WorkspaceID, channel.ID, ref.EntityID, gin.H{"ref": ref})
		_ = emitKnowledgeEntitySpikeAlerts(ref.EntityID, channel.ID)
	}
}

func emitKnowledgeEntitySpikeAlerts(entityID, channelID string) error {
	alerts, err := knowledge.DetectEntitySpikeAlerts(db.DB, entityID, channelID, time.Now().UTC())
	if err != nil {
		return err
	}
	for _, alert := range alerts {
		if err := broadcastKnowledgeEvent("knowledge.entity.activity.spiked", alert.Entity.WorkspaceID, alert.ChannelID, alert.Entity.ID, gin.H{
			"entity":              alert.Entity,
			"user_ids":            alert.UserIDs,
			"channel_id":          alert.ChannelID,
			"recent_ref_count":    alert.RecentRefCount,
			"previous_ref_count":  alert.PreviousRefCount,
			"delta":               alert.Delta,
			"related_channel_ids": alert.RelatedChannelIDs,
			"occurred_at":         alert.OccurredAt,
		}); err != nil {
			return err
		}
	}
	return nil
}

func broadcastKnowledgeEvent(eventType, workspaceID, channelID, entityID string, payload any) error {
	if RealtimeHub == nil {
		return nil
	}

	return RealtimeHub.Broadcast(realtime.Event{
		ID:          "evt_" + time.Now().Format("20060102150405.000000"),
		Type:        eventType,
		WorkspaceID: workspaceID,
		ChannelID:   channelID,
		EntityID:    entityID,
		TS:          time.Now().UTC().Format(time.RFC3339Nano),
		Payload:     payload,
	})
}
