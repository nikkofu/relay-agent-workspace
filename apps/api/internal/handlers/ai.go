package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
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
	if AIGateway == nil {
		return nil, errors.New("ai gateway is not configured")
	}

	var input struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	}
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	req := llm.Request{
		Prompt:    buildSummaryPrompt(scopeType, messages),
		ChannelID: channelID,
		Provider:  input.Provider,
		Model:     input.Model,
	}

	session, err := AIGateway.Stream(c.Request.Context(), req)
	if err != nil {
		return nil, err
	}

	content, reasoning, err := collectStreamOutput(c.Request.Context(), session)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return &summary, nil
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
