package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

type aiComposeActivityDigestSummary struct {
	TotalRequests int64 `json:"total_requests"`
	UniqueUsers   int64 `json:"unique_users"`
}

type aiComposeActivityDigestRow struct {
	Key           string `json:"key"`
	Count         int64  `json:"count"`
	UniqueUsers   int64  `json:"unique_users,omitempty"`
	SuggestionSum int64  `json:"suggestion_sum,omitempty"`
}

func GetAIComposeActivityDigest(c *gin.Context) {
	scopeType, scopeValue, err := resolveAIComposeActivityDigestScope(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	window, startAt, endAt, err := parseAIComposeActivityDigestWindow(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	groupBy := strings.ToLower(strings.TrimSpace(c.DefaultQuery("group_by", "intent")))
	groupExpr, err := aiComposeActivityDigestGroupExpr(groupBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := parsePositiveInt(c.DefaultQuery("limit", "10"), 10)
	if limit > 100 {
		limit = 100
	}

	filtered := db.DB.Model(&domain.AIComposeActivity{})
	filtered = applyAIComposeActivityDigestFilters(filtered, scopeType, scopeValue, c.Query("intent"), startAt, endAt)

	var summary aiComposeActivityDigestSummary
	if err := filtered.Session(&gorm.Session{}).Count(&summary.TotalRequests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load compose activity digest"})
		return
	}
	if err := filtered.Session(&gorm.Session{}).
		Select("COUNT(DISTINCT NULLIF(TRIM(user_id), ''))").
		Scan(&summary.UniqueUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load compose activity digest"})
		return
	}

	var breakdown []aiComposeActivityDigestRow
	if err := filtered.Session(&gorm.Session{}).
		Select(fmt.Sprintf("%s AS key, COUNT(*) AS count, COUNT(DISTINCT NULLIF(TRIM(user_id), '')) AS unique_users, COALESCE(SUM(suggestion_count), 0) AS suggestion_sum", groupExpr)).
		Group(groupExpr).
		Order("count desc, key asc").
		Limit(limit).
		Scan(&breakdown).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load compose activity digest"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"window":    window,
		"scope":     gin.H{"type": scopeType, "value": scopeValue},
		"group_by":  groupBy,
		"summary":   summary,
		"breakdown": breakdown,
	})
}

func resolveAIComposeActivityDigestScope(c *gin.Context) (string, string, error) {
	channelID := strings.TrimSpace(c.Query("channel_id"))
	if channelID != "" {
		return "channel", channelID, nil
	}

	dmID := strings.TrimSpace(c.Query("dm_id"))
	if dmID != "" {
		return "dm", dmID, nil
	}

	workspaceID := strings.TrimSpace(c.Query("workspace_id"))
	if workspaceID != "" {
		return "workspace", workspaceID, nil
	}

	return "", "", fmt.Errorf("workspace_id is required")
}

func parseAIComposeActivityDigestWindow(c *gin.Context) (string, *time.Time, *time.Time, error) {
	window := strings.ToLower(strings.TrimSpace(c.DefaultQuery("window", "24h")))
	now := time.Now().UTC()

	switch window {
	case "1h":
		start := now.Add(-1 * time.Hour)
		return window, &start, &now, nil
	case "24h":
		start := now.Add(-24 * time.Hour)
		return window, &start, &now, nil
	case "7d":
		start := now.Add(-7 * 24 * time.Hour)
		return window, &start, &now, nil
	case "custom":
		startAtRaw := strings.TrimSpace(c.Query("start_at"))
		endAtRaw := strings.TrimSpace(c.Query("end_at"))
		if startAtRaw == "" || endAtRaw == "" {
			return "", nil, nil, fmt.Errorf("start_at and end_at are required when window=custom")
		}
		startAt, err := parseAIComposeActivityDigestTime(startAtRaw)
		if err != nil {
			return "", nil, nil, fmt.Errorf("invalid start_at: %w", err)
		}
		endAt, err := parseAIComposeActivityDigestTime(endAtRaw)
		if err != nil {
			return "", nil, nil, fmt.Errorf("invalid end_at: %w", err)
		}
		if !endAt.After(startAt) {
			return "", nil, nil, fmt.Errorf("end_at must be after start_at")
		}
		return window, &startAt, &endAt, nil
	default:
		return "", nil, nil, fmt.Errorf("unsupported window")
	}
}

func parseAIComposeActivityDigestTime(raw string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed.UTC(), nil
		}
	}

	if unixSeconds, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return time.Unix(unixSeconds, 0).UTC(), nil
	}

	return time.Time{}, fmt.Errorf("expected RFC3339 timestamp or unix seconds")
}

func aiComposeActivityDigestGroupExpr(groupBy string) (string, error) {
	switch groupBy {
	case "", "intent":
		return "COALESCE(NULLIF(TRIM(intent), ''), 'unknown')", nil
	case "user":
		return "COALESCE(NULLIF(TRIM(user_id), ''), 'unknown')", nil
	case "channel":
		return "COALESCE(NULLIF(TRIM(channel_id), ''), 'unknown')", nil
	case "dm":
		return "COALESCE(NULLIF(TRIM(dm_conversation_id), ''), 'unknown')", nil
	case "workspace":
		return "COALESCE(NULLIF(TRIM(workspace_id), ''), 'unknown')", nil
	case "provider":
		return "COALESCE(NULLIF(TRIM(provider), ''), 'unknown')", nil
	case "model":
		return "COALESCE(NULLIF(TRIM(model), ''), 'unknown')", nil
	default:
		return "", fmt.Errorf("unsupported group_by")
	}
}

func applyAIComposeActivityDigestFilters(query *gorm.DB, scopeType, scopeValue, intent string, startAt, endAt *time.Time) *gorm.DB {
	switch scopeType {
	case "channel":
		query = query.Where("channel_id = ?", scopeValue)
	case "dm":
		query = query.Where("dm_conversation_id = ?", scopeValue)
	case "workspace":
		query = query.Where("workspace_id = ?", scopeValue)
	}

	if trimmedIntent := strings.TrimSpace(intent); trimmedIntent != "" {
		query = query.Where("LOWER(intent) = ?", strings.ToLower(trimmedIntent))
	}
	if startAt != nil {
		query = query.Where("created_at >= ?", *startAt)
	}
	if endAt != nil {
		query = query.Where("created_at < ?", *endAt)
	}

	return query
}
