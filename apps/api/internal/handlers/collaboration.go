package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

func GetMe(c *gin.Context) {
	var user domain.User
	if err := db.DB.First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
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
	db.DB.Where("channel_id = ?", channelID).Order("created_at asc").Find(&messages)
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func CreateMessage(c *gin.Context) {
	var input struct {
		ChannelID string `json:"channel_id" binding:"required"`
		Content   string `json:"content" binding:"required"`
		UserID    string `json:"user_id" binding:"required"`
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
		CreatedAt: time.Now(),
	}

	if err := db.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create message"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": msg})
}
