package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/automation"
	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/knowledge"
	"github.com/nikkofu/relay-agent-workspace/api/internal/llm"
)

func GetKnowledgeEntityBriefAutomation(c *gin.Context) {
	entity, err := knowledge.GetEntity(db.DB, c.Param("id"))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}

	job, err := automation.GetKnowledgeEntityBriefAutomationJob(db.DB, entity.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"job": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load entity brief automation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job":    job,
		"entity": entity,
	})
}

func RunKnowledgeEntityBriefAutomation(c *gin.Context) {
	entity, err := knowledge.GetEntity(db.DB, c.Param("id"))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}

	now := time.Now().UTC()
	job, created, err := queueKnowledgeEntityBriefAutomation(entity, "manual_run", now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue entity brief automation"})
		return
	}

	if created {
		broadcastKnowledgeEntityBriefAutomationEvent("knowledge.entity.brief.regen.queued", entity, gin.H{
			"job":    job,
			"entity": entity,
		})
		c.JSON(http.StatusAccepted, gin.H{
			"job":    job,
			"entity": entity,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job":    job,
		"entity": entity,
	})
}

func RetryKnowledgeEntityBriefAutomation(c *gin.Context) {
	entity, err := knowledge.GetEntity(db.DB, c.Param("id"))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}

	now := time.Now().UTC()
	job, created, err := automation.RetryKnowledgeEntityBriefAutomationJob(db.DB, entity.WorkspaceID, entity.ID, now)
	if err != nil {
		if errors.Is(err, automation.ErrEntityBriefAutomationIneligible) {
			c.JSON(http.StatusConflict, gin.H{"error": "latest entity brief automation job is not retryable"})
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "entity brief automation job not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retry entity brief automation"})
		return
	}

	if created {
		broadcastKnowledgeEntityBriefAutomationEvent("knowledge.entity.brief.regen.queued", entity, gin.H{
			"job":    job,
			"entity": entity,
		})
		c.JSON(http.StatusAccepted, gin.H{
			"job":    job,
			"entity": entity,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job":    job,
		"entity": entity,
	})
}

func SweepKnowledgeEntityBriefAutomation(now time.Time, staleAfter time.Duration) ([]domain.AIAutomationJob, error) {
	if db.DB == nil {
		return nil, nil
	}

	staleJobs, err := automation.SweepStaleKnowledgeEntityBriefAutomationJobs(db.DB, now, staleAfter)
	if err != nil {
		return nil, err
	}

	for _, job := range staleJobs {
		entity, entityErr := knowledge.GetEntity(db.DB, job.ScopeID)
		if entityErr != nil {
			entity = domain.KnowledgeEntity{
				ID: job.ScopeID,
			}
		}
		broadcastKnowledgeEntityBriefAutomationEvent("knowledge.entity.brief.regen.failed", entity, gin.H{
			"job":    job,
			"entity": entity,
			"reason": strings.TrimSpace(job.LastError),
		})
	}

	return staleJobs, nil
}

func QueueKnowledgeEntityBriefAutomationByID(entityID, triggerReason string, now time.Time) (domain.AIAutomationJob, bool, error) {
	entity, err := knowledge.GetEntity(db.DB, entityID)
	if err != nil {
		return domain.AIAutomationJob{}, false, err
	}
	return queueKnowledgeEntityBriefAutomation(entity, triggerReason, now)
}

func ProcessKnowledgeEntityBriefAutomation(now time.Time, limit int) error {
	if db.DB == nil || AIGateway == nil {
		return nil
	}
	if err := queueStaleKnowledgeEntityBriefs(now, limit); err != nil {
		return err
	}
	jobs, err := automation.ListPendingKnowledgeEntityBriefAutomationJobs(db.DB, limit)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		entity, err := knowledge.GetEntity(db.DB, job.ScopeID)
		if err != nil {
			continue
		}
		startedJob, err := automation.MarkKnowledgeEntityBriefAutomationStarted(db.DB, job.ID, now)
		if err != nil {
			return err
		}
		broadcastKnowledgeEntityBriefAutomationEvent("knowledge.entity.brief.regen.started", entity, gin.H{
			"job":    startedJob,
			"entity": entity,
		})

		brief, err := generateKnowledgeEntityBriefForAutomation(context.Background(), entity.ID)
		if err != nil {
			failedJob, markErr := automation.MarkKnowledgeEntityBriefAutomationFailed(db.DB, job.ID, err.Error(), time.Now().UTC())
			if markErr == nil {
				broadcastKnowledgeEntityBriefAutomationEvent("knowledge.entity.brief.regen.failed", entity, gin.H{
					"job":    failedJob,
					"entity": entity,
					"reason": strings.TrimSpace(failedJob.LastError),
				})
			}
			continue
		}
		if _, err := automation.MarkKnowledgeEntityBriefAutomationSucceeded(db.DB, job.ID, time.Now().UTC()); err != nil {
			return err
		}
		broadcastKnowledgeEntityBriefAutomationEvent("knowledge.entity.brief.generated", entity, gin.H{"brief": brief})
	}
	return nil
}

func broadcastKnowledgeEntityBriefAutomationEvent(eventType string, entity domain.KnowledgeEntity, payload any) {
	if strings.TrimSpace(entity.ID) == "" {
		return
	}
	_ = broadcastKnowledgeEvent(eventType, entity.WorkspaceID, "", entity.ID, payload)
}

func queueKnowledgeEntityBriefAutomation(entity domain.KnowledgeEntity, triggerReason string, now time.Time) (domain.AIAutomationJob, bool, error) {
	return automation.EnqueueKnowledgeEntityBriefAutomationJob(db.DB, entity.WorkspaceID, entity.ID, triggerReason, now)
}

func queueStaleKnowledgeEntityBriefs(now time.Time, limit int) error {
	if limit <= 0 {
		limit = 10
	}
	var entities []domain.KnowledgeEntity
	if err := db.DB.Order("updated_at desc").Find(&entities).Error; err != nil {
		return err
	}
	queued := 0
	for _, entity := range entities {
		if queued >= limit {
			break
		}
		stale, err := isKnowledgeEntityBriefStale(entity.ID)
		if err != nil || !stale {
			continue
		}
		if _, created, err := queueKnowledgeEntityBriefAutomation(entity, "periodic_stale_sweep", now); err == nil && created {
			queued++
		}
	}
	return nil
}

func isKnowledgeEntityBriefStale(entityID string) (bool, error) {
	var summary domain.AISummary
	summaryErr := db.DB.Where("scope_type = ? AND scope_id = ?", "knowledge_entity", entityID).First(&summary).Error
	if summaryErr != nil && !errors.Is(summaryErr, gorm.ErrRecordNotFound) {
		return false, summaryErr
	}
	var entity domain.KnowledgeEntity
	if err := db.DB.First(&entity, "id = ?", entityID).Error; err != nil {
		return false, err
	}
	if errors.Is(summaryErr, gorm.ErrRecordNotFound) {
		return true, nil
	}
	if entity.UpdatedAt.After(summary.UpdatedAt) {
		return true, nil
	}

	var latestRef domain.KnowledgeEntityRef
	if err := db.DB.Where("entity_id = ?", entityID).Order("created_at desc").First(&latestRef).Error; err == nil && latestRef.CreatedAt.After(summary.UpdatedAt) {
		return true, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	var latestEvent domain.KnowledgeEvent
	if err := db.DB.Where("entity_id = ?", entityID).Order("created_at desc").First(&latestEvent).Error; err == nil && latestEvent.CreatedAt.After(summary.UpdatedAt) {
		return true, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	return false, nil
}

func generateKnowledgeEntityBriefForAutomation(ctx context.Context, scopeID string) (knowledge.EntityBrief, error) {
	entity, refs, events, prompt, err := knowledge.BuildEntityBriefPrompt(db.DB, scopeID)
	if err != nil {
		return knowledge.EntityBrief{}, err
	}
	session, err := AIGateway.Stream(ctx, llm.Request{
		Prompt:    prompt,
		ChannelID: entity.WorkspaceID,
	})
	if err != nil {
		return knowledge.EntityBrief{}, err
	}
	content, reasoning, err := collectStreamOutput(ctx, session)
	if err != nil {
		return knowledge.EntityBrief{}, err
	}

	now := time.Now().UTC()
	var lastRefAt *time.Time
	for _, ref := range refs {
		if lastRefAt == nil || ref.CreatedAt.After(*lastRefAt) {
			ts := ref.CreatedAt
			lastRefAt = &ts
		}
	}
	summary := domain.AISummary{
		ScopeType:     "knowledge_entity",
		ScopeID:       entity.ID,
		ChannelID:     entity.WorkspaceID,
		Provider:      session.Provider,
		Model:         session.Model,
		Content:       strings.TrimSpace(content),
		Reasoning:     strings.TrimSpace(reasoning),
		MessageCount:  len(refs),
		LastMessageAt: lastRefAt,
		UpdatedAt:     now,
	}
	if err := db.DB.Where("scope_type = ? AND scope_id = ?", summary.ScopeType, summary.ScopeID).Assign(summary).FirstOrCreate(&summary).Error; err != nil {
		return knowledge.EntityBrief{}, err
	}
	return knowledge.EntityBrief{
		EntityID:    entity.ID,
		WorkspaceID: entity.WorkspaceID,
		Title:       entity.Title,
		Content:     summary.Content,
		Reasoning:   summary.Reasoning,
		Provider:    summary.Provider,
		Model:       summary.Model,
		GeneratedAt: now,
		RefCount:    len(refs),
		EventCount:  len(events),
		LastRefAt:   lastRefAt,
		Cached:      false,
	}, nil
}
