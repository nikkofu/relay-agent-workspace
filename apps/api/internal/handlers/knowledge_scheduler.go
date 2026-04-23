package handlers

import (
	"context"
	"log"
	"time"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/knowledge"
)

func StartKnowledgeDigestScheduler() {
	go runKnowledgeDigestScheduler()
}

func StartKnowledgeEntityBriefAutomationScheduler() {
	go runKnowledgeEntityBriefAutomationScheduler()
}

func StartChannelAutoSummaryScheduler() {
	go runChannelAutoSummaryScheduler()
}

func runKnowledgeDigestScheduler() {
	processKnowledgeDigestSchedules(time.Now().UTC())

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for tickAt := range ticker.C {
		processKnowledgeDigestSchedules(tickAt.UTC())
	}
}

func processKnowledgeDigestSchedules(now time.Time) {
	if db.DB == nil {
		return
	}

	published, err := knowledge.ProcessDigestSchedules(db.DB, now)
	if err != nil {
		log.Printf("knowledge digest scheduler failed: %v", err)
		return
	}

	for _, item := range published {
		if err := broadcastRealtimeEvent("message.created", item.Message, item.Message); err != nil {
			log.Printf("knowledge digest scheduler message broadcast failed: %v", err)
		}
		if err := broadcastKnowledgeDigestPublished(item); err != nil {
			log.Printf("knowledge digest scheduler knowledge broadcast failed: %v", err)
		}
	}
}

func runKnowledgeEntityBriefAutomationScheduler() {
	processKnowledgeEntityBriefAutomation(time.Now().UTC())

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for tickAt := range ticker.C {
		processKnowledgeEntityBriefAutomation(tickAt.UTC())
	}
}

func runChannelAutoSummaryScheduler() {
	if err := processChannelAutoSummaries(time.Now().UTC()); err != nil {
		log.Printf("channel auto summary scheduler failed: %v", err)
	}

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for tickAt := range ticker.C {
		if err := processChannelAutoSummaries(tickAt.UTC()); err != nil {
			log.Printf("channel auto summary scheduler failed: %v", err)
		}
	}
}

func processKnowledgeEntityBriefAutomation(now time.Time) {
	if db.DB == nil {
		return
	}

	if _, err := SweepKnowledgeEntityBriefAutomation(now, 15*time.Minute); err != nil {
		log.Printf("knowledge entity brief automation stale sweep failed: %v", err)
	}
	if err := ProcessKnowledgeEntityBriefAutomation(now, 10); err != nil {
		log.Printf("knowledge entity brief automation processing failed: %v", err)
	}
}

func processChannelAutoSummaries(now time.Time) error {
	if db.DB == nil || AIGateway == nil {
		return nil
	}

	var settings []domain.ChannelAutoSummarySetting
	if err := db.DB.Where("is_enabled = ?", true).Order("updated_at asc, id asc").Find(&settings).Error; err != nil {
		return err
	}

	for _, setting := range settings {
		if err := processSingleChannelAutoSummary(now, setting); err != nil {
			log.Printf("channel auto summary failed for channel %s: %v", setting.ChannelID, err)
		}
	}
	return nil
}

func processSingleChannelAutoSummary(now time.Time, setting domain.ChannelAutoSummarySetting) error {
	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", setting.ChannelID).Error; err != nil {
		return err
	}

	cursor := time.Time{}
	if setting.LastMessageAt != nil {
		cursor = setting.LastMessageAt.UTC()
	} else if setting.LastRunAt != nil {
		cursor = setting.LastRunAt.UTC()
	}

	var newMessages []domain.Message
	query := db.DB.Where("channel_id = ?", channel.ID)
	if !cursor.IsZero() {
		query = query.Where("created_at > ?", cursor)
	}
	minNewMessages := maxInt(setting.MinNewMessages, 1)
	if err := query.Order("created_at asc").Limit(minNewMessages).Find(&newMessages).Error; err != nil {
		return err
	}
	if len(newMessages) < minNewMessages {
		return nil
	}

	summary, lastMessageAt, err := generateChannelSummaryWithContext(
		context.Background(),
		channel,
		setting.Provider,
		setting.Model,
		setting.MessageLimit,
	)
	if err != nil {
		return err
	}

	setting.LastRunAt = &now
	if lastMessageAt != nil {
		setting.LastMessageAt = lastMessageAt
	}
	setting.UpdatedAt = now
	if err := db.DB.Save(&setting).Error; err != nil {
		return err
	}

	return broadcastChannelSummaryUpdated(channel, summary, setting, "auto_run")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
