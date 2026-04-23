package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

func ListAIAutomationJobs(c *gin.Context) {
	workspaceID := strings.TrimSpace(c.Query("workspace_id"))
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	limit := parseLimit(c.Query("limit"), 20)
	if limit > 100 {
		limit = 100
	}

	query := db.DB.Model(&domain.AIAutomationJob{}).Where("workspace_id = ?", workspaceID)
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	if jobType := strings.TrimSpace(c.Query("job_type")); jobType != "" {
		query = query.Where("job_type = ?", jobType)
	}
	if scopeType := strings.TrimSpace(c.Query("scope_type")); scopeType != "" {
		query = query.Where("scope_type = ?", scopeType)
	}
	if scopeID := strings.TrimSpace(c.Query("scope_id")); scopeID != "" {
		query = query.Where("scope_id = ?", scopeID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count automation jobs"})
		return
	}

	var items []domain.AIAutomationJob
	if err := query.Order("created_at desc, id desc").Limit(limit).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load automation jobs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
	})
}
