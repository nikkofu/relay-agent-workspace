package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"github.com/nikkofu/relay-agent-workspace/api/internal/llm"
)

type AIStreamer interface {
	Stream(ctx context.Context, req llm.Request) (*llm.StreamSession, error)
}

var AIGateway AIStreamer
var AIConfig llm.Config

func SetAIGateway(gateway AIStreamer) {
	AIGateway = gateway
}

func SetAIConfig(config llm.Config) {
	AIConfig = config
}

func GetAIConfig(c *gin.Context) {
	providers := make([]gin.H, 0, len(AIConfig.Providers))
	for id, provider := range AIConfig.Providers {
		if !provider.Enabled {
			continue
		}
		providers = append(providers, gin.H{
			"id":     id,
			"models": []string{provider.Model},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"default_provider": AIConfig.DefaultProvider,
		"providers":        providers,
	})
}

func GetAIConversations(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var conversations []domain.AIConversation
	if err := db.DB.Where("user_id = ?", currentUser.ID).Order("updated_at desc").Find(&conversations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load ai conversations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

func GetAIConversation(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	conversationID := c.Param("id")
	var conversation domain.AIConversation
	if err := db.DB.Where("id = ? AND user_id = ?", conversationID, currentUser.ID).First(&conversation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ai conversation not found"})
		return
	}

	var messages []domain.AIConversationMessage
	if err := db.DB.Where("conversation_id = ?", conversationID).Order("created_at asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load ai conversation messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversation": conversation, "messages": messages})
}

func ComposeAI(c *gin.Context) {
	input, err := bindComposeInput(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, _, err := runComposeAI(c, input)
	if err != nil {
		handleComposeError(c, err)
		return
	}

	_ = recordAndBroadcastComposeSuggestionGenerated(response)
	c.JSON(http.StatusOK, gin.H{"compose": response})
}

func ComposeAIStream(c *gin.Context) {
	input, err := bindComposeInput(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if AIGateway == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ai gateway is not configured"})
		return
	}

	prepared, err := prepareComposeAI(input)
	if err != nil {
		handleComposeError(c, err)
		return
	}
	session, err := AIGateway.Stream(c.Request.Context(), llm.Request{
		Prompt:    prepared.Prompt,
		ChannelID: input.ComposeScopeID(),
	})
	if err != nil {
		handleComposeError(c, err)
		return
	}

	writer := c.Writer
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	flusher, ok := writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	requestID := ids.NewPrefixedUUID("compose-request")
	provisionalSuggestionID := ids.NewPrefixedUUID("compose")
	fullContent := ""

	writeSSE(writer, "start", gin.H{
		"request_id": requestID,
		"channel_id": input.ChannelID,
		"dm_id":      input.DMID,
		"thread_id":  input.ThreadID,
		"intent":     input.Intent,
		"limit":      input.Limit,
		"provider":   session.Provider,
		"model":      session.Model,
	})
	flusher.Flush()

	for {
		select {
		case event, ok := <-session.Events:
			if !ok {
				select {
				case err, ok := <-session.Errors:
					if ok && err != nil {
						writeSSE(writer, "error", gin.H{"message": err.Error()})
						flusher.Flush()
						return
					}
				default:
				}
				response := buildComposeResponse(input.ChannelID, input.DMID, input.ThreadID, input.Intent, input.Limit, input.Draft, prepared.ThreadParent, prepared.RecentMessages, prepared.RecentDMMessages, prepared.KnowledgeContext, prepared.EntityMatches, fullContent, session.Provider, session.Model)
				if len(response.Suggestions) > 0 {
					response.Suggestions[0].ID = provisionalSuggestionID
				}
				_ = recordAndBroadcastComposeSuggestionGenerated(response)
				for index, suggestion := range response.Suggestions {
					writeSSE(writer, "suggestion.done", gin.H{
						"index":            index,
						"suggestion":       suggestion,
						"citations":        response.Citations,
						"context_entities": response.ContextEntities,
						"channel_id":       response.ChannelID,
						"dm_id":            response.DMID,
						"thread_id":        response.ThreadID,
					})
					flusher.Flush()
				}
				writeSSE(writer, "done", gin.H{
					"request_id":       requestID,
					"suggestion_count": len(response.Suggestions),
					"provider":         session.Provider,
					"model":            session.Model,
				})
				flusher.Flush()
				return
			}
			if event.Type != "chunk" {
				continue
			}
			fullContent += event.Text
			writeSSE(writer, "suggestion.delta", gin.H{
				"suggestion_id": provisionalSuggestionID,
				"index":         0,
				"text_delta":    event.Text,
			})
			flusher.Flush()
		case err, ok := <-session.Errors:
			if !ok {
				session.Errors = nil
				continue
			}
			if err != nil {
				writeSSE(writer, "error", gin.H{"message": err.Error()})
				flusher.Flush()
			}
			return
		case <-c.Request.Context().Done():
			return
		}
	}
}

func SubmitAIComposeFeedback(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		ChannelID      string `json:"channel_id"`
		DMID           string `json:"dm_id"`
		ThreadID       string `json:"thread_id"`
		Intent         string `json:"intent"`
		Feedback       string `json:"feedback"`
		SuggestionText string `json:"suggestion_text"`
		Provider       string `json:"provider"`
		Model          string `json:"model"`
	}
	if err := bindJSONBody(c, &input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.ChannelID = strings.TrimSpace(input.ChannelID)
	input.DMID = strings.TrimSpace(input.DMID)
	input.ThreadID = strings.TrimSpace(input.ThreadID)
	input.Intent = strings.ToLower(strings.TrimSpace(input.Intent))
	input.Feedback = strings.ToLower(strings.TrimSpace(input.Feedback))
	input.SuggestionText = strings.TrimSpace(input.SuggestionText)
	input.Provider = strings.TrimSpace(input.Provider)
	input.Model = strings.TrimSpace(input.Model)
	if (input.ChannelID == "") == (input.DMID == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exactly one of channel_id or dm_id is required"})
		return
	}
	if input.Intent == "" {
		input.Intent = "reply"
	}
	if !isSupportedComposeIntent(input.Intent) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "intent must be one of reply, summarize, followup, or schedule"})
		return
	}
	switch input.Feedback {
	case "up", "down", "edited":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "feedback must be up, down, or edited"})
		return
	}

	if input.ChannelID != "" {
		var channel domain.Channel
		if err := db.DB.First(&channel, "id = ?", input.ChannelID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load channel"})
			return
		}
	} else {
		var conversation domain.DMConversation
		if err := db.DB.First(&conversation, "id = ?", input.DMID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "dm conversation not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dm conversation"})
			return
		}
	}

	now := time.Now().UTC()
	composeID := c.Param("id")
	var feedback domain.AIComposeFeedback
	err = db.DB.Where("compose_id = ? AND user_id = ?", composeID, currentUser.ID).First(&feedback).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		feedback = domain.AIComposeFeedback{
			ID:               ids.NewPrefixedUUID("compose-feedback"),
			ComposeID:        composeID,
			UserID:           currentUser.ID,
			ChannelID:        input.ChannelID,
			DMConversationID: input.DMID,
			ThreadID:         input.ThreadID,
			Intent:           input.Intent,
			Feedback:         input.Feedback,
			SuggestionText:   input.SuggestionText,
			Provider:         input.Provider,
			Model:            input.Model,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := db.DB.Create(&feedback).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist compose feedback"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"feedback": feedback})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load compose feedback"})
		return
	}

	feedback.ChannelID = input.ChannelID
	feedback.DMConversationID = input.DMID
	feedback.ThreadID = input.ThreadID
	feedback.Intent = input.Intent
	feedback.Feedback = input.Feedback
	feedback.SuggestionText = input.SuggestionText
	feedback.Provider = input.Provider
	feedback.Model = input.Model
	feedback.UpdatedAt = now
	if err := db.DB.Save(&feedback).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist compose feedback"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"feedback": feedback})
}

func GetAIComposeFeedbackSummary(c *gin.Context) {
	composeID := strings.TrimSpace(c.Param("id"))
	if composeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "compose id is required"})
		return
	}

	var feedbackRows []domain.AIComposeFeedback
	if err := db.DB.Where("compose_id = ?", composeID).Order("updated_at desc").Find(&feedbackRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load compose feedback"})
		return
	}

	counts := gin.H{"up": int64(0), "down": int64(0), "edited": int64(0)}
	var total int64
	for _, row := range feedbackRows {
		total++
		switch row.Feedback {
		case "up":
			counts["up"] = counts["up"].(int64) + 1
		case "down":
			counts["down"] = counts["down"].(int64) + 1
		case "edited":
			counts["edited"] = counts["edited"].(int64) + 1
		}
	}
	recentRows := feedbackRows
	if len(recentRows) > 20 {
		recentRows = recentRows[:20]
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": gin.H{
			"compose_id": composeID,
			"total":      total,
			"counts":     counts,
			"recent":     recentRows,
		},
	})
}

func GetAIComposeActivity(c *gin.Context) {
	limit := parsePositiveInt(c.DefaultQuery("limit", "50"), 50)
	if limit > 100 {
		limit = 100
	}
	channelID := strings.TrimSpace(c.Query("channel_id"))
	dmID := strings.TrimSpace(c.Query("dm_id"))
	if channelID != "" && dmID != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id and dm_id cannot both be provided"})
		return
	}

	query := db.DB.Model(&domain.AIComposeActivity{}).Order("created_at desc").Limit(limit)
	if channelID != "" {
		query = query.Where("channel_id = ?", channelID)
	}
	if dmID != "" {
		query = query.Where("dm_conversation_id = ?", dmID)
	}
	if workspaceID := strings.TrimSpace(c.Query("workspace_id")); workspaceID != "" {
		query = query.Where("workspace_id = ?", workspaceID)
	}
	if intent := strings.ToLower(strings.TrimSpace(c.Query("intent"))); intent != "" {
		query = query.Where("intent = ?", intent)
	}

	var items []domain.AIComposeActivity
	if err := query.Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load compose activity"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func GetThreadSummary(c *gin.Context) {
	summary, err := getStoredSummary("thread", c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"summary": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load thread summary"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

func GenerateThreadSummary(c *gin.Context) {
	parentID := c.Param("id")

	var parent domain.Message
	if err := db.DB.First(&parent, "id = ?", parentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "thread parent not found"})
		return
	}

	var replies []domain.Message
	if err := db.DB.Where("thread_id = ?", parentID).Order("created_at asc").Find(&replies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load thread replies"})
		return
	}

	messages := append([]domain.Message{parent}, replies...)
	summary, err := generateSummaryFromMessages(c, "thread", parentID, parent.ChannelID, messages)
	if err != nil {
		handleSummaryGenerationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

func GetChannelSummary(c *gin.Context) {
	summary, err := getStoredSummary("channel", c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"summary": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load channel summary"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

func GenerateChannelSummary(c *gin.Context) {
	channelID := c.Param("id")

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", channelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	var messages []domain.Message
	if err := db.DB.Where("channel_id = ?", channelID).Order("created_at desc").Limit(50).Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load channel messages"})
		return
	}
	if len(messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel has no messages to summarize"})
		return
	}
	reverseMessages(messages)

	summary, err := generateSummaryFromMessages(c, "channel", channelID, channelID, messages)
	if err != nil {
		handleSummaryGenerationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

func GetChannelKnowledgeAutoSummarize(c *gin.Context) {
	channel, ok := loadAutoSummaryChannel(c)
	if !ok {
		return
	}
	setting, err := getOrDefaultChannelAutoSummarySetting(channel, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load auto-summarize setting"})
		return
	}
	summary, err := getStoredSummary("channel", channel.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load channel summary"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"setting": setting, "summary": summary})
}

func PutChannelKnowledgeAutoSummarize(c *gin.Context) {
	channel, ok := loadAutoSummaryChannel(c)
	if !ok {
		return
	}
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	var input channelAutoSummaryInput
	if err := bindJSONBody(c, &input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	setting, err := upsertChannelAutoSummarySetting(channel, currentUser.ID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist auto-summarize setting"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"setting": setting})
}

func RunChannelKnowledgeAutoSummarize(c *gin.Context) {
	if AIGateway == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ai gateway is not configured"})
		return
	}
	channel, ok := loadAutoSummaryChannel(c)
	if !ok {
		return
	}
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	var input channelAutoSummaryInput
	if err := bindJSONBody(c, &input); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	setting, err := getOrDefaultChannelAutoSummarySetting(channel, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load auto-summarize setting"})
		return
	}
	setting.Provider = defaultString(strings.TrimSpace(input.Provider), setting.Provider)
	setting.Model = defaultString(strings.TrimSpace(input.Model), setting.Model)
	setting.MessageLimit = normalizeAutoSummaryMessageLimit(firstPositive(input.MessageLimit, setting.MessageLimit))

	summary, lastMessageAt, err := generateChannelSummaryWithOptions(c, channel, setting.Provider, setting.Model, setting.MessageLimit)
	if err != nil {
		handleSummaryGenerationError(c, err)
		return
	}
	now := time.Now().UTC()
	setting.LastRunAt = &now
	setting.LastMessageAt = lastMessageAt
	setting.UpdatedAt = now
	if setting.ID == "" {
		setting.ID = ids.NewPrefixedUUID("auto-summary")
		setting.ChannelID = channel.ID
		setting.WorkspaceID = channel.WorkspaceID
		setting.CreatedBy = currentUser.ID
		setting.CreatedAt = now
	}
	if err := db.DB.Where("channel_id = ?", channel.ID).Assign(setting).FirstOrCreate(&setting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist auto-summarize setting"})
		return
	}
	_ = broadcastChannelSummaryUpdated(channel, summary, setting, "manual_run")
	c.JSON(http.StatusOK, gin.H{"setting": setting, "summary": summary})
}

type channelAutoSummaryInput struct {
	IsEnabled      bool   `json:"is_enabled"`
	WindowHours    int    `json:"window_hours"`
	MessageLimit   int    `json:"message_limit"`
	MinNewMessages int    `json:"min_new_messages"`
	Provider       string `json:"provider"`
	Model          string `json:"model"`
	Force          bool   `json:"force"`
}

func loadAutoSummaryChannel(c *gin.Context) (domain.Channel, bool) {
	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return domain.Channel{}, false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load channel"})
		return domain.Channel{}, false
	}
	return channel, true
}

func getOrDefaultChannelAutoSummarySetting(channel domain.Channel, userID string) (domain.ChannelAutoSummarySetting, error) {
	var setting domain.ChannelAutoSummarySetting
	if err := db.DB.Where("channel_id = ?", channel.ID).First(&setting).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ChannelAutoSummarySetting{}, err
		}
		now := time.Now().UTC()
		return domain.ChannelAutoSummarySetting{
			ID:             "",
			ChannelID:      channel.ID,
			WorkspaceID:    channel.WorkspaceID,
			CreatedBy:      userID,
			IsEnabled:      false,
			WindowHours:    24,
			MessageLimit:   50,
			MinNewMessages: 5,
			CreatedAt:      now,
			UpdatedAt:      now,
		}, nil
	}
	return setting, nil
}

func upsertChannelAutoSummarySetting(channel domain.Channel, userID string, input channelAutoSummaryInput) (domain.ChannelAutoSummarySetting, error) {
	current, err := getOrDefaultChannelAutoSummarySetting(channel, userID)
	if err != nil {
		return domain.ChannelAutoSummarySetting{}, err
	}
	now := time.Now().UTC()
	if current.ID == "" {
		current.ID = ids.NewPrefixedUUID("auto-summary")
		current.CreatedAt = now
	}
	current.ChannelID = channel.ID
	current.WorkspaceID = channel.WorkspaceID
	current.CreatedBy = defaultString(current.CreatedBy, userID)
	current.IsEnabled = input.IsEnabled
	current.WindowHours = normalizeAutoSummaryWindowHours(input.WindowHours)
	current.MessageLimit = normalizeAutoSummaryMessageLimit(input.MessageLimit)
	current.MinNewMessages = normalizeAutoSummaryMinNewMessages(input.MinNewMessages)
	current.Provider = strings.TrimSpace(input.Provider)
	current.Model = strings.TrimSpace(input.Model)
	current.UpdatedAt = now
	if err := db.DB.Where("channel_id = ?", channel.ID).Assign(current).FirstOrCreate(&current).Error; err != nil {
		return domain.ChannelAutoSummarySetting{}, err
	}
	return current, nil
}

func normalizeAutoSummaryWindowHours(value int) int {
	if value <= 0 {
		return 24
	}
	if value > 24*30 {
		return 24 * 30
	}
	return value
}

func normalizeAutoSummaryMessageLimit(value int) int {
	if value <= 0 {
		return 50
	}
	if value > 200 {
		return 200
	}
	return value
}

func normalizeAutoSummaryMinNewMessages(value int) int {
	if value <= 0 {
		return 5
	}
	if value > 100 {
		return 100
	}
	return value
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func broadcastChannelSummaryUpdated(channel domain.Channel, summary *domain.AISummary, setting domain.ChannelAutoSummarySetting, reason string) error {
	if summary == nil {
		return nil
	}
	return broadcastKnowledgeEvent("channel.summary.updated", channel.WorkspaceID, channel.ID, "", gin.H{
		"channel_id":   channel.ID,
		"workspace_id": channel.WorkspaceID,
		"reason":       reason,
		"summary":      summary,
		"setting":      setting,
	})
}

func ExecuteAI(c *gin.Context) {
	if AIGateway == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ai gateway is not configured"})
		return
	}

	var input llm.Request
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Prompt == "" || input.ChannelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt and channel_id are required"})
		return
	}

	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	session, err := AIGateway.Stream(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	writer := c.Writer
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	flusher, ok := writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	fullContent := ""
	reasoningContent := ""
	conversationID := input.ConversationID
	if conversationID == "" {
		conversationID = ids.NewPrefixedUUID("ai-conv")
	}

	writeSSE(writer, "start", map[string]any{
		"provider":        session.Provider,
		"model":           session.Model,
		"conversation_id": conversationID,
	})
	flusher.Flush()
	if input.ConversationID == "" {
		writeSSE(writer, "conversation", map[string]any{"conversation_id": conversationID})
		flusher.Flush()
	}

	for {
		select {
		case event, ok := <-session.Events:
			if !ok {
				if err := persistAIConversation(currentUser, input, session, conversationID, fullContent, reasoningContent); err != nil {
					writeSSE(writer, "error", map[string]any{"message": "failed to persist ai conversation"})
					flusher.Flush()
					return
				}
				writeSSE(writer, "done", map[string]any{
					"provider":        session.Provider,
					"model":           session.Model,
					"conversation_id": conversationID,
				})
				flusher.Flush()
				return
			}
			if event.Type == "chunk" {
				fullContent += event.Text
			} else if event.Type == "reasoning" {
				reasoningContent += event.Text
			}
			writeSSE(writer, event.Type, map[string]any{"text": event.Text})
			flusher.Flush()
		case err, ok := <-session.Errors:
			if !ok {
				session.Errors = nil
				continue
			}
			if err != nil {
				writeSSE(writer, "error", map[string]any{"message": err.Error()})
				flusher.Flush()
			}
			return
		case <-c.Request.Context().Done():
			return
		}
	}
}

func getStoredSummary(scopeType, scopeID string) (*domain.AISummary, error) {
	var summary domain.AISummary
	if err := db.DB.Where("scope_type = ? AND scope_id = ?", scopeType, scopeID).First(&summary).Error; err != nil {
		return nil, err
	}
	return &summary, nil
}

func generateSummaryFromMessages(c *gin.Context, scopeType, scopeID, channelID string, messages []domain.Message) (*domain.AISummary, error) {
	var input struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	}
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	summary, _, err := generateSummaryWithOptions(c, scopeType, scopeID, channelID, messages, strings.TrimSpace(input.Provider), strings.TrimSpace(input.Model))
	return summary, err
}

func generateChannelSummaryWithOptions(c *gin.Context, channel domain.Channel, provider, model string, limit int) (*domain.AISummary, *time.Time, error) {
	limit = normalizeAutoSummaryMessageLimit(limit)
	var messages []domain.Message
	if err := db.DB.Where("channel_id = ?", channel.ID).Order("created_at desc").Limit(limit).Find(&messages).Error; err != nil {
		return nil, nil, err
	}
	if len(messages) == 0 {
		return nil, nil, errors.New("channel has no messages to summarize")
	}
	reverseMessages(messages)
	return generateSummaryWithOptions(c, "channel", channel.ID, channel.ID, messages, provider, model)
}

func generateSummaryWithOptions(c *gin.Context, scopeType, scopeID, channelID string, messages []domain.Message, provider, model string) (*domain.AISummary, *time.Time, error) {
	if AIGateway == nil {
		return nil, nil, errors.New("ai gateway is not configured")
	}
	if len(messages) == 0 {
		return nil, nil, errors.New("channel has no messages to summarize")
	}

	req := llm.Request{
		Prompt:    buildSummaryPrompt(scopeType, messages),
		ChannelID: channelID,
		Provider:  provider,
		Model:     model,
	}

	session, err := AIGateway.Stream(c.Request.Context(), req)
	if err != nil {
		return nil, nil, err
	}

	content, reasoning, err := collectStreamOutput(c.Request.Context(), session)
	if err != nil {
		return nil, nil, err
	}

	lastMessageAt := messages[len(messages)-1].CreatedAt
	summary := domain.AISummary{
		ScopeType:     scopeType,
		ScopeID:       scopeID,
		ChannelID:     channelID,
		Provider:      session.Provider,
		Model:         session.Model,
		Content:       strings.TrimSpace(content),
		Reasoning:     strings.TrimSpace(reasoning),
		MessageCount:  len(messages),
		LastMessageAt: &lastMessageAt,
		UpdatedAt:     time.Now().UTC(),
	}

	if err := db.DB.Where("scope_type = ? AND scope_id = ?", scopeType, scopeID).
		Assign(summary).
		FirstOrCreate(&summary).Error; err != nil {
		return nil, nil, err
	}

	return &summary, &lastMessageAt, nil
}

func buildSummaryPrompt(scopeType string, messages []domain.Message) string {
	var builder strings.Builder
	if scopeType == "thread" {
		builder.WriteString("Summarize the following Slack-style thread for a busy teammate. Focus on decisions, owners, risks, and next steps.\n\n")
	} else {
		builder.WriteString("Summarize the following Slack-style channel activity for a busy teammate. Focus on themes, decisions, blockers, owners, and next steps.\n\n")
	}

	for _, message := range messages {
		builder.WriteString("- [")
		builder.WriteString(message.UserID)
		builder.WriteString("] ")
		builder.WriteString(message.Content)
		builder.WriteString("\n")
	}

	builder.WriteString("\nReturn a concise paragraph summary.")
	return builder.String()
}

func collectStreamOutput(ctx context.Context, session *llm.StreamSession) (string, string, error) {
	content := ""
	reasoning := ""

	for {
		select {
		case event, ok := <-session.Events:
			if !ok {
				return content, reasoning, nil
			}
			switch event.Type {
			case "chunk":
				content += event.Text
			case "reasoning":
				reasoning += event.Text
			}
		case err, ok := <-session.Errors:
			if !ok {
				session.Errors = nil
				continue
			}
			if err != nil {
				return "", "", err
			}
		case <-ctx.Done():
			return "", "", ctx.Err()
		}
	}
}

func reverseMessages(messages []domain.Message) {
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}
}

func handleSummaryGenerationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, context.Canceled):
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "summary generation canceled"})
	case strings.Contains(err.Error(), "ai gateway is not configured"):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}

func persistAIConversation(user domain.User, input llm.Request, session *llm.StreamSession, conversationID, assistantContent, reasoning string) error {
	now := time.Now().UTC()
	conversation := domain.AIConversation{
		ID:        conversationID,
		UserID:    user.ID,
		ChannelID: input.ChannelID,
		Provider:  session.Provider,
		Model:     session.Model,
		UpdatedAt: now,
	}
	if err := db.DB.Where("id = ?", conversationID).
		Assign(domain.AIConversation{
			UserID:    user.ID,
			ChannelID: input.ChannelID,
			Provider:  session.Provider,
			Model:     session.Model,
			UpdatedAt: now,
		}).
		FirstOrCreate(&conversation).Error; err != nil {
		return err
	}

	userMessage := domain.AIConversationMessage{
		ID:             ids.NewPrefixedUUID("ai-msg"),
		ConversationID: conversationID,
		Role:           "user",
		Content:        input.Prompt,
		CreatedAt:      now,
	}
	if err := db.DB.Create(&userMessage).Error; err != nil {
		return err
	}

	assistantMessage := domain.AIConversationMessage{
		ID:             ids.NewPrefixedUUID("ai-msg"),
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        assistantContent,
		Reasoning:      reasoning,
		CreatedAt:      now,
	}
	if err := db.DB.Create(&assistantMessage).Error; err != nil {
		return err
	}

	return db.DB.Model(&domain.AIConversation{}).Where("id = ?", conversationID).Update("updated_at", now).Error
}

func SubmitAIFeedback(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input struct {
		MessageID string `json:"message_id" binding:"required"`
		IsGood    bool   `json:"is_good"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var message domain.Message
	if err := db.DB.First(&message, "id = ?", input.MessageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	feedback := domain.AIFeedback{
		MessageID: input.MessageID,
		UserID:    currentUser.ID,
		IsGood:    input.IsGood,
		UpdatedAt: time.Now().UTC(),
	}
	if err := db.DB.Where("message_id = ? AND user_id = ?", input.MessageID, currentUser.ID).Assign(feedback).FirstOrCreate(&feedback).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist feedback"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"feedback": feedback})
}

func writeSSE(writer http.ResponseWriter, event string, payload any) {
	data, _ := json.Marshal(payload)
	_, _ = writer.Write([]byte("event: " + event + "\n"))
	_, _ = writer.Write([]byte("data: " + string(data) + "\n\n"))
}

func normalizeComposeLimit(limit int) int {
	switch {
	case limit < 1:
		return 3
	case limit > 5:
		return 5
	default:
		return limit
	}
}

func parsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func isSupportedComposeIntent(intent string) bool {
	switch intent {
	case "reply", "summarize", "followup", "schedule":
		return true
	default:
		return false
	}
}

func loadComposeRecentMessages(channelID, threadID string) ([]domain.Message, error) {
	query := db.DB.Model(&domain.Message{}).Where("channel_id = ?", channelID).Order("created_at desc, id desc").Limit(12)
	if threadID != "" {
		query = query.Where("(thread_id = ? OR id = ?)", threadID, threadID)
	}

	var messages []domain.Message
	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func loadComposeRecentDMMessages(dmID string) ([]domain.DMMessage, error) {
	var messages []domain.DMMessage
	if err := db.DB.Where("dm_conversation_id = ?", dmID).Order("created_at desc, id desc").Limit(12).Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func composeMatchText(draft string, threadParent *domain.Message, recentMessages []domain.Message) string {
	var builder strings.Builder
	if strings.TrimSpace(draft) != "" {
		builder.WriteString(strings.TrimSpace(draft))
		builder.WriteString("\n")
	}
	if threadParent != nil && strings.TrimSpace(threadParent.Content) != "" {
		builder.WriteString(threadParent.Content)
		builder.WriteString("\n")
	}
	for _, message := range recentMessages {
		if strings.TrimSpace(message.Content) == "" {
			continue
		}
		builder.WriteString(message.Content)
		builder.WriteString("\n")
	}
	return builder.String()
}

func composeDMMatchText(draft string, recentMessages []domain.DMMessage) string {
	var builder strings.Builder
	if strings.TrimSpace(draft) != "" {
		builder.WriteString(strings.TrimSpace(draft))
		builder.WriteString("\n")
	}
	for _, message := range recentMessages {
		if strings.TrimSpace(message.Content) == "" {
			continue
		}
		builder.WriteString(message.Content)
		builder.WriteString("\n")
	}
	return builder.String()
}

func composeScopeMatchText(draft string, threadParent *domain.Message, recentMessages []domain.Message, recentDMMessages []domain.DMMessage) string {
	if len(recentDMMessages) > 0 {
		return composeDMMatchText(draft, recentDMMessages)
	}
	return composeMatchText(draft, threadParent, recentMessages)
}

func composeIntentInstruction(intent string) string {
	switch intent {
	case "summarize":
		return "Summarize the unread context into a short draft the user can send."
	case "followup":
		return "Draft clarifying follow-up questions rooted in missing facts or blockers."
	case "schedule":
		return "Draft scheduling language that proposes next coordination steps without inventing availability."
	default:
		return "Draft concise reply suggestions."
	}
}

func buildComposePrompt(channelID, dmID, threadID, intent, draft string, limit int, threadParent *domain.Message, recentMessages []domain.Message, recentDMMessages []domain.DMMessage, knowledgeContext knowledge.ChannelKnowledgeContext, matches []knowledge.EntityTextMatch) string {
	var builder strings.Builder
	builder.WriteString("You are a Slack-like AI writing assistant for grounded composer suggestions.\n")
	builder.WriteString("Stay grounded in the provided context and only return concise sendable suggestions.\n")
	builder.WriteString("Do not invent owners, dates, or decisions.\n")
	builder.WriteString("Return JSON with suggestions, citations, and context_entities.\n\n")
	builder.WriteString("scope: ")
	if dmID != "" {
		builder.WriteString("dm")
		builder.WriteString("\ndm_id: ")
		builder.WriteString(dmID)
	} else {
		builder.WriteString("channel")
		builder.WriteString("\nchannel_id: ")
		builder.WriteString(channelID)
	}
	builder.WriteString("\nintent: ")
	builder.WriteString(intent)
	builder.WriteString("\nintent_instruction: ")
	builder.WriteString(composeIntentInstruction(intent))
	builder.WriteString("\nlimit: ")
	builder.WriteString(fmt.Sprintf("%d", limit))
	builder.WriteString("\n")
	if threadID != "" {
		builder.WriteString("thread_id: ")
		builder.WriteString(threadID)
		builder.WriteString("\n")
	}
	if draft != "" {
		builder.WriteString("draft: ")
		builder.WriteString(draft)
		builder.WriteString("\n")
	}
	if threadParent != nil {
		builder.WriteString("\nthread_parent:\n- ")
		builder.WriteString(threadParent.Content)
		builder.WriteString("\n")
	}
	if len(recentMessages) > 0 {
		builder.WriteString("\nrecent_messages:\n")
		for _, message := range recentMessages {
			builder.WriteString("- [")
			builder.WriteString(message.UserID)
			builder.WriteString("] ")
			builder.WriteString(message.Content)
			builder.WriteString("\n")
		}
	}
	if len(recentDMMessages) > 0 {
		builder.WriteString("\nrecent_dm_messages:\n")
		for _, message := range recentDMMessages {
			builder.WriteString("- [")
			builder.WriteString(message.UserID)
			builder.WriteString("] ")
			builder.WriteString(message.Content)
			builder.WriteString("\n")
		}
	}
	if len(knowledgeContext.Refs) > 0 {
		builder.WriteString("\nchannel_knowledge:\n")
		for _, ref := range knowledgeContext.Refs {
			builder.WriteString("- ")
			builder.WriteString(ref.EntityTitle)
			builder.WriteString(" / ")
			builder.WriteString(ref.RefKind)
			builder.WriteString(": ")
			builder.WriteString(ref.SourceSnippet)
			builder.WriteString("\n")
		}
	}
	if len(matches) > 0 {
		builder.WriteString("\nmatches:\n")
		for _, match := range matches {
			builder.WriteString("- ")
			builder.WriteString(match.EntityTitle)
			builder.WriteString(" (")
			builder.WriteString(match.EntityKind)
			builder.WriteString(")\n")
		}
	}
	builder.WriteString("\nReturn up to ")
	builder.WriteString(fmt.Sprintf("%d", limit))
	builder.WriteString(" suggestions.")
	return builder.String()
}

func buildComposeResponse(channelID, dmID, threadID, intent string, limit int, draft string, threadParent *domain.Message, recentMessages []domain.Message, recentDMMessages []domain.DMMessage, knowledgeContext knowledge.ChannelKnowledgeContext, matches []knowledge.EntityTextMatch, llmOutput, provider, model string) composePayload {
	structured := parseComposeOutput(llmOutput)
	citations, contextEntities := buildComposeGrounding(knowledgeContext.Refs, matches, draft, threadParent, recentMessages, recentDMMessages)

	suggestions := structured.Suggestions
	if len(suggestions) == 0 {
		suggestions = []knowledge.ComposeSuggestion{{
			ID:   ids.NewPrefixedUUID("compose"),
			Text: buildFallbackComposeSuggestion(intent, draft, threadParent, recentMessages, recentDMMessages),
			Tone: "clear",
			Kind: intent,
		}}
	}
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}
	for index := range suggestions {
		if strings.TrimSpace(suggestions[index].ID) == "" || suggestions[index].ID == "cmp-1" {
			suggestions[index].ID = ids.NewPrefixedUUID("compose")
		}
		if strings.TrimSpace(suggestions[index].Tone) == "" {
			suggestions[index].Tone = "clear"
		}
		if strings.TrimSpace(suggestions[index].Kind) == "" {
			suggestions[index].Kind = intent
		}
	}
	if len(structured.Citations) > 0 {
		citations = structured.Citations
	}
	if len(structured.ContextEntities) > 0 {
		contextEntities = structured.ContextEntities
	}
	proposedSlots := structured.ProposedSlots
	if intent == "schedule" && len(proposedSlots) == 0 {
		proposedSlots = buildFallbackProposedSlots(recentMessages, recentDMMessages)
	}

	return composePayload{
		ChannelID:       channelID,
		DMID:            dmID,
		ThreadID:        threadID,
		Intent:          intent,
		Suggestions:     suggestions,
		Citations:       citations,
		ContextEntities: contextEntities,
		ProposedSlots:   proposedSlots,
		Provider:        provider,
		Model:           model,
	}
}

type composePayload struct {
	ChannelID       string                           `json:"channel_id"`
	DMID            string                           `json:"dm_id,omitempty"`
	ThreadID        string                           `json:"thread_id,omitempty"`
	Intent          string                           `json:"intent"`
	Suggestions     []knowledge.ComposeSuggestion    `json:"suggestions"`
	Citations       []knowledge.Citation             `json:"citations"`
	ContextEntities []knowledge.ComposeContextEntity `json:"context_entities"`
	ProposedSlots   []knowledge.ComposeProposedSlot  `json:"proposed_slots,omitempty"`
	Provider        string                           `json:"provider"`
	Model           string                           `json:"model"`
}

type composeStructuredOutput struct {
	Suggestions     []knowledge.ComposeSuggestion    `json:"suggestions"`
	Citations       []knowledge.Citation             `json:"citations"`
	ContextEntities []knowledge.ComposeContextEntity `json:"context_entities"`
	ProposedSlots   []knowledge.ComposeProposedSlot  `json:"proposed_slots"`
}

func parseComposeOutput(raw string) composeStructuredOutput {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(raw, "{") {
		return composeStructuredOutput{}
	}

	var parsed composeStructuredOutput
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return composeStructuredOutput{}
	}
	return parsed
}

func buildFallbackProposedSlots(recentMessages []domain.Message, recentDMMessages []domain.DMMessage) []knowledge.ComposeProposedSlot {
	start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Hour)
	attendees := make([]string, 0, 4)
	seen := map[string]struct{}{}
	for _, message := range recentMessages {
		if message.UserID == "" {
			continue
		}
		if _, ok := seen[message.UserID]; ok {
			continue
		}
		seen[message.UserID] = struct{}{}
		attendees = append(attendees, message.UserID)
	}
	for _, message := range recentDMMessages {
		if message.UserID == "" {
			continue
		}
		if _, ok := seen[message.UserID]; ok {
			continue
		}
		seen[message.UserID] = struct{}{}
		attendees = append(attendees, message.UserID)
	}
	if len(attendees) > 4 {
		attendees = attendees[:4]
	}
	return []knowledge.ComposeProposedSlot{{
		StartsAt:        start.Format(time.RFC3339),
		EndsAt:          start.Add(30 * time.Minute).Format(time.RFC3339),
		DurationMinutes: 30,
		Timezone:        "UTC",
		AttendeeIDs:     attendees,
		Reason:          "Fallback schedule slot from recent conversation participants.",
	}}
}

func recordAndBroadcastComposeSuggestionGenerated(response composePayload) error {
	activity, err := persistComposeActivity(response)
	if err != nil {
		return err
	}
	return broadcastComposeSuggestionGenerated(response, activity)
}

func persistComposeActivity(response composePayload) (domain.AIComposeActivity, error) {
	composeID := firstComposeSuggestionID(response)
	if composeID == "" {
		composeID = ids.NewPrefixedUUID("compose")
	}
	workspaceID := composeWorkspaceID(response)
	currentUser, _ := getCurrentUser()
	now := time.Now().UTC()
	activity := domain.AIComposeActivity{
		ID:               ids.NewPrefixedUUID("compose-activity"),
		ComposeID:        composeID,
		WorkspaceID:      workspaceID,
		ChannelID:        response.ChannelID,
		DMConversationID: response.DMID,
		ThreadID:         response.ThreadID,
		UserID:           currentUser.ID,
		Intent:           response.Intent,
		SuggestionCount:  len(response.Suggestions),
		Provider:         response.Provider,
		Model:            response.Model,
		CreatedAt:        now,
	}
	if err := db.DB.Where("compose_id = ?", composeID).FirstOrCreate(&activity).Error; err != nil {
		return domain.AIComposeActivity{}, err
	}
	return activity, nil
}

func firstComposeSuggestionID(response composePayload) string {
	for _, suggestion := range response.Suggestions {
		if strings.TrimSpace(suggestion.ID) != "" {
			return strings.TrimSpace(suggestion.ID)
		}
	}
	return ""
}

func composeWorkspaceID(response composePayload) string {
	if response.ChannelID != "" {
		var channel domain.Channel
		if err := db.DB.First(&channel, "id = ?", response.ChannelID).Error; err == nil {
			return channel.WorkspaceID
		}
	}
	return ""
}

func broadcastComposeSuggestionGenerated(response composePayload, activity domain.AIComposeActivity) error {
	workspaceID := activity.WorkspaceID
	return broadcastKnowledgeEvent("knowledge.compose.suggestion.generated", workspaceID, response.ChannelID, "", gin.H{
		"compose":  response,
		"activity": activity,
	})
}

func buildComposeGrounding(refs []knowledge.ChannelKnowledgeRef, matches []knowledge.EntityTextMatch, draft string, threadParent *domain.Message, recentMessages []domain.Message, recentDMMessages []domain.DMMessage) ([]knowledge.Citation, []knowledge.ComposeContextEntity) {
	matchText := strings.ToLower(composeScopeMatchText(draft, threadParent, recentMessages, recentDMMessages))
	contextEntities := make([]knowledge.ComposeContextEntity, 0, len(refs)+len(matches))
	seenEntities := map[string]struct{}{}
	for _, ref := range refs {
		if ref.EntityID == "" {
			continue
		}
		if _, ok := seenEntities[ref.EntityID]; !ok {
			seenEntities[ref.EntityID] = struct{}{}
			contextEntities = append(contextEntities, knowledge.ComposeContextEntity{
				ID:    ref.EntityID,
				Title: ref.EntityTitle,
				Kind:  ref.EntityKind,
			})
		}
	}
	for _, match := range matches {
		if match.EntityID == "" {
			continue
		}
		if _, ok := seenEntities[match.EntityID]; ok {
			continue
		}
		seenEntities[match.EntityID] = struct{}{}
		contextEntities = append(contextEntities, knowledge.ComposeContextEntity{
			ID:    match.EntityID,
			Title: match.EntityTitle,
			Kind:  match.EntityKind,
		})
	}

	citations := make([]knowledge.Citation, 0, len(refs))
	for _, ref := range refs {
		citation := knowledge.Citation{
			ID:           ref.ID,
			EvidenceKind: ref.RefKind,
			SourceKind:   ref.RefKind,
			SourceRef:    ref.RefID,
			RefKind:      ref.RefKind,
			Snippet:      ref.SourceSnippet,
			Title:        ref.SourceTitle,
			Score:        1,
			EntityID:     ref.EntityID,
			EntityTitle:  ref.EntityTitle,
		}
		if citation.Snippet == "" {
			citation.Snippet = citation.Title
		}
		if citation.Title == "" {
			citation.Title = citation.Snippet
		}
		if strings.Contains(matchText, "owner") || strings.Contains(matchText, "timeline") {
			citation.Score = 2
		}
		citations = append(citations, citation)
		if len(citations) >= 1 {
			break
		}
	}

	if len(citations) == 0 && len(contextEntities) > 0 {
		citations = []knowledge.Citation{{
			ID:          "compose-grounding",
			Title:       contextEntities[0].Title,
			Snippet:     contextEntities[0].Title,
			Score:       1,
			EntityID:    contextEntities[0].ID,
			EntityTitle: contextEntities[0].Title,
		}}
	}

	return citations, contextEntities
}

func buildFallbackComposeSuggestion(intent, draft string, threadParent *domain.Message, recentMessages []domain.Message, recentDMMessages []domain.DMMessage) string {
	context := strings.ToLower(composeScopeMatchText(draft, threadParent, recentMessages, recentDMMessages))
	switch intent {
	case "summarize":
		if len(recentDMMessages) > 0 {
			return "Here is the current private-sync summary and the next open point."
		}
		return "Here is the current thread summary and the next open point."
	case "followup":
		return "Can you share the missing context or blocker so we can move this forward?"
	case "schedule":
		return "Let's align on a time window and confirm who needs to join before we schedule it."
	}
	base := "I can help draft a reply."
	if strings.Contains(context, "owner") || strings.Contains(context, "timeline") || strings.Contains(context, "confirm") {
		base = "I can take the owner update."
	}

	if strings.Contains(context, "friday review") || strings.Contains(context, "current target") {
		return base + " Current launch target still looks aligned with the Friday review."
	}
	if strings.Contains(context, "timeline") {
		return base + " The timeline still looks consistent with the current plan."
	}
	return base
}

func bindComposeInput(c *gin.Context) (composeInput, error) {
	var input composeInput
	if err := bindJSONBody(c, &input); err != nil {
		return composeInput{}, err
	}

	input.ChannelID = strings.TrimSpace(input.ChannelID)
	input.DMID = strings.TrimSpace(input.DMID)
	input.ThreadID = strings.TrimSpace(input.ThreadID)
	input.Draft = strings.TrimSpace(input.Draft)
	input.Intent = strings.ToLower(strings.TrimSpace(input.Intent))
	if (input.ChannelID == "") == (input.DMID == "") {
		return composeInput{}, errors.New("exactly one of channel_id or dm_id is required")
	}
	if input.Intent == "" {
		input.Intent = "reply"
	}
	if !isSupportedComposeIntent(input.Intent) {
		return composeInput{}, errors.New("intent must be one of reply, summarize, followup, or schedule")
	}
	input.Limit = normalizeComposeLimit(input.Limit)
	return input, nil
}

func runComposeAI(c *gin.Context, input composeInput) (composePayload, *llm.StreamSession, error) {
	if AIGateway == nil {
		return composePayload{}, nil, errors.New("ai gateway is not configured")
	}
	prepared, err := prepareComposeAI(input)
	if err != nil {
		return composePayload{}, nil, err
	}
	req := llm.Request{
		Prompt:    prepared.Prompt,
		ChannelID: input.ComposeScopeID(),
	}
	session, err := AIGateway.Stream(c.Request.Context(), req)
	if err != nil {
		return composePayload{}, nil, err
	}

	content, _, err := collectStreamOutput(c.Request.Context(), session)
	if err != nil {
		return composePayload{}, nil, err
	}

	return buildComposeResponse(input.ChannelID, input.DMID, input.ThreadID, input.Intent, input.Limit, input.Draft, prepared.ThreadParent, prepared.RecentMessages, prepared.RecentDMMessages, prepared.KnowledgeContext, prepared.EntityMatches, content, session.Provider, session.Model), session, nil
}

func handleComposeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
	case strings.Contains(err.Error(), "thread parent not found"):
		c.JSON(http.StatusNotFound, gin.H{"error": "thread parent not found"})
	case strings.Contains(err.Error(), "dm conversation not found"):
		c.JSON(http.StatusNotFound, gin.H{"error": "dm conversation not found"})
	case strings.Contains(err.Error(), "ai gateway is not configured"):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	case strings.Contains(err.Error(), "exactly one of channel_id or dm_id"), strings.Contains(err.Error(), "intent must be"):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}

type composeInput struct {
	ChannelID string `json:"channel_id"`
	DMID      string `json:"dm_id"`
	ThreadID  string `json:"thread_id"`
	Draft     string `json:"draft"`
	Intent    string `json:"intent"`
	Limit     int    `json:"limit"`
}

func (input composeInput) ComposeScopeID() string {
	if input.DMID != "" {
		return input.DMID
	}
	return input.ChannelID
}

type preparedComposeAI struct {
	Prompt           string
	ThreadParent     *domain.Message
	RecentMessages   []domain.Message
	RecentDMMessages []domain.DMMessage
	KnowledgeContext knowledge.ChannelKnowledgeContext
	EntityMatches    []knowledge.EntityTextMatch
}

func prepareComposeAI(input composeInput) (preparedComposeAI, error) {
	if input.DMID != "" {
		var conversation domain.DMConversation
		if err := db.DB.First(&conversation, "id = ?", input.DMID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return preparedComposeAI{}, fmt.Errorf("dm conversation not found")
			}
			return preparedComposeAI{}, err
		}
		recentDMMessages, err := loadComposeRecentDMMessages(input.DMID)
		if err != nil {
			return preparedComposeAI{}, err
		}
		return preparedComposeAI{
			Prompt:           buildComposePrompt("", input.DMID, "", input.Intent, input.Draft, input.Limit, nil, nil, recentDMMessages, knowledge.ChannelKnowledgeContext{}, nil),
			RecentDMMessages: recentDMMessages,
		}, nil
	}

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", input.ChannelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return preparedComposeAI{}, gorm.ErrRecordNotFound
		}
		return preparedComposeAI{}, err
	}

	var threadParent *domain.Message
	if input.ThreadID != "" {
		var parent domain.Message
		if err := db.DB.First(&parent, "id = ? AND channel_id = ?", input.ThreadID, input.ChannelID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return preparedComposeAI{}, fmt.Errorf("thread parent not found")
			}
			return preparedComposeAI{}, err
		}
		threadParent = &parent
	}

	recentMessages, err := loadComposeRecentMessages(input.ChannelID, input.ThreadID)
	if err != nil {
		return preparedComposeAI{}, err
	}
	knowledgeContext, err := knowledge.GetChannelKnowledgeContext(db.DB, input.ChannelID, input.Limit)
	if err != nil {
		return preparedComposeAI{}, err
	}
	entityMatches, err := knowledge.MatchEntitiesInText(db.DB, knowledge.MatchEntitiesInput{
		WorkspaceID: channel.WorkspaceID,
		Text:        composeScopeMatchText(input.Draft, threadParent, recentMessages, nil),
		Limit:       input.Limit * 2,
	})
	if err != nil {
		return preparedComposeAI{}, err
	}

	return preparedComposeAI{
		Prompt:           buildComposePrompt(input.ChannelID, "", input.ThreadID, input.Intent, input.Draft, input.Limit, threadParent, recentMessages, nil, knowledgeContext, entityMatches),
		ThreadParent:     threadParent,
		RecentMessages:   recentMessages,
		KnowledgeContext: knowledgeContext,
		EntityMatches:    entityMatches,
	}, nil
}

func bindJSONBody(c *gin.Context, target any) error {
	raw, err := c.GetRawData()
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}
