package automation

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
)

const (
	EntityBriefAutomationJobType   = "entity_brief_regen"
	EntityBriefAutomationScopeType = "knowledge_entity"
)

var (
	ErrEntityBriefAutomationUnavailable = errors.New("entity brief automation job not available")
	ErrEntityBriefAutomationIneligible  = errors.New("entity brief automation job is not eligible for retry")
)

func GetKnowledgeEntityBriefAutomationJob(database *gorm.DB, entityID string) (domain.AIAutomationJob, error) {
	entityID = strings.TrimSpace(entityID)
	if entityID == "" {
		return domain.AIAutomationJob{}, gorm.ErrRecordNotFound
	}

	job, ok, err := loadActiveKnowledgeEntityBriefAutomationJob(database, entityID)
	if err != nil {
		return domain.AIAutomationJob{}, err
	}
	if ok {
		return job, nil
	}

	return loadLatestKnowledgeEntityBriefAutomationJob(database, entityID)
}

func EnqueueKnowledgeEntityBriefAutomationJob(database *gorm.DB, workspaceID, entityID, triggerReason string, now time.Time) (domain.AIAutomationJob, bool, error) {
	entityID = strings.TrimSpace(entityID)
	if entityID == "" {
		return domain.AIAutomationJob{}, false, gorm.ErrRecordNotFound
	}

	if job, ok, err := loadActiveKnowledgeEntityBriefAutomationJob(database, entityID); err != nil {
		return domain.AIAutomationJob{}, false, err
	} else if ok {
		return job, false, nil
	}

	job := newKnowledgeEntityBriefAutomationJob(workspaceID, entityID, triggerReason, now)
	if existing, err := loadKnowledgeEntityBriefAutomationJobByDedupeKey(database, entityID, job.DedupeKey); err == nil {
		return existing, false, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.AIAutomationJob{}, false, err
	}
	if err := database.Create(&job).Error; err != nil {
		return domain.AIAutomationJob{}, false, err
	}
	return job, true, nil
}

func RetryKnowledgeEntityBriefAutomationJob(database *gorm.DB, workspaceID, entityID string, now time.Time) (domain.AIAutomationJob, bool, error) {
	entityID = strings.TrimSpace(entityID)
	if entityID == "" {
		return domain.AIAutomationJob{}, false, gorm.ErrRecordNotFound
	}

	if job, ok, err := loadActiveKnowledgeEntityBriefAutomationJob(database, entityID); err != nil {
		return domain.AIAutomationJob{}, false, err
	} else if ok {
		return job, false, nil
	}

	latest, err := loadLatestKnowledgeEntityBriefAutomationJob(database, entityID)
	if err != nil {
		return domain.AIAutomationJob{}, false, err
	}
	if !isRetryableKnowledgeEntityBriefAutomationStatus(latest.Status) {
		return domain.AIAutomationJob{}, false, ErrEntityBriefAutomationIneligible
	}

	job := newKnowledgeEntityBriefAutomationJob(workspaceID, entityID, "manual_retry", now)
	if err := database.Create(&job).Error; err != nil {
		return domain.AIAutomationJob{}, false, err
	}
	return job, true, nil
}

func MarkKnowledgeEntityBriefAutomationStarted(database *gorm.DB, jobID string, now time.Time) (domain.AIAutomationJob, error) {
	return updateKnowledgeEntityBriefAutomationStatus(database, jobID, "running", now, "")
}

func MarkKnowledgeEntityBriefAutomationFailed(database *gorm.DB, jobID, lastError string, now time.Time) (domain.AIAutomationJob, error) {
	return updateKnowledgeEntityBriefAutomationStatus(database, jobID, "failed", now, lastError)
}

func MarkKnowledgeEntityBriefAutomationSucceeded(database *gorm.DB, jobID string, now time.Time) (domain.AIAutomationJob, error) {
	return updateKnowledgeEntityBriefAutomationStatus(database, jobID, "succeeded", now, "")
}

func SweepStaleKnowledgeEntityBriefAutomationJobs(database *gorm.DB, now time.Time, staleAfter time.Duration) ([]domain.AIAutomationJob, error) {
	if staleAfter <= 0 {
		staleAfter = 15 * time.Minute
	}

	cutoff := now.Add(-staleAfter)
	var jobs []domain.AIAutomationJob
	if err := database.
		Where("job_type = ? AND scope_type = ? AND status IN ? AND updated_at < ?", EntityBriefAutomationJobType, EntityBriefAutomationScopeType, []string{"pending", "running"}, cutoff).
		Order("updated_at asc, created_at asc").
		Find(&jobs).Error; err != nil {
		return nil, err
	}

	for i := range jobs {
		if _, err := updateKnowledgeEntityBriefAutomationStatus(database, jobs[i].ID, "failed", now, "stale entity brief automation job"); err != nil {
			return nil, err
		}
		jobs[i].Status = "failed"
		jobs[i].LastError = "stale entity brief automation job"
		jobs[i].FinishedAt = &now
		jobs[i].UpdatedAt = now
	}

	return jobs, nil
}

func ListPendingKnowledgeEntityBriefAutomationJobs(database *gorm.DB, limit int) ([]domain.AIAutomationJob, error) {
	if limit <= 0 {
		limit = 10
	}
	var jobs []domain.AIAutomationJob
	if err := database.
		Where("job_type = ? AND scope_type = ? AND status = ?", EntityBriefAutomationJobType, EntityBriefAutomationScopeType, "pending").
		Order("scheduled_at asc, created_at asc").
		Limit(limit).
		Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func loadActiveKnowledgeEntityBriefAutomationJob(database *gorm.DB, entityID string) (domain.AIAutomationJob, bool, error) {
	var job domain.AIAutomationJob
	err := database.
		Where("job_type = ? AND scope_type = ? AND scope_id = ? AND status IN ?", EntityBriefAutomationJobType, EntityBriefAutomationScopeType, entityID, []string{"pending", "running"}).
		Order("updated_at desc, created_at desc").
		First(&job).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AIAutomationJob{}, false, nil
		}
		return domain.AIAutomationJob{}, false, err
	}
	return job, true, nil
}

func loadLatestKnowledgeEntityBriefAutomationJob(database *gorm.DB, entityID string) (domain.AIAutomationJob, error) {
	var job domain.AIAutomationJob
	if err := database.
		Where("job_type = ? AND scope_type = ? AND scope_id = ?", EntityBriefAutomationJobType, EntityBriefAutomationScopeType, entityID).
		Order("updated_at desc, created_at desc").
		First(&job).Error; err != nil {
		return domain.AIAutomationJob{}, err
	}
	return job, nil
}

func loadKnowledgeEntityBriefAutomationJobByDedupeKey(database *gorm.DB, entityID, dedupeKey string) (domain.AIAutomationJob, error) {
	var job domain.AIAutomationJob
	if err := database.
		Where("job_type = ? AND scope_type = ? AND scope_id = ? AND dedupe_key = ?", EntityBriefAutomationJobType, EntityBriefAutomationScopeType, entityID, strings.TrimSpace(dedupeKey)).
		Order("created_at desc").
		First(&job).Error; err != nil {
		return domain.AIAutomationJob{}, err
	}
	return job, nil
}

func newKnowledgeEntityBriefAutomationJob(workspaceID, entityID, triggerReason string, now time.Time) domain.AIAutomationJob {
	bucket := now.UTC().Unix() / 120
	return domain.AIAutomationJob{
		ID:            ids.NewPrefixedUUID("aijob"),
		JobType:       EntityBriefAutomationJobType,
		ScopeType:     EntityBriefAutomationScopeType,
		ScopeID:       entityID,
		WorkspaceID:   strings.TrimSpace(workspaceID),
		Status:        "pending",
		TriggerReason: strings.TrimSpace(triggerReason),
		DedupeKey:     "entity_brief_regen:" + entityID + ":" + strconv.FormatInt(bucket, 10),
		AttemptCount:  0,
		ScheduledAt:   now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func updateKnowledgeEntityBriefAutomationStatus(database *gorm.DB, jobID, status string, now time.Time, lastError string) (domain.AIAutomationJob, error) {
	var job domain.AIAutomationJob
	if err := database.First(&job, "id = ?", strings.TrimSpace(jobID)).Error; err != nil {
		return domain.AIAutomationJob{}, err
	}

	updates := map[string]any{
		"status":     status,
		"updated_at": now,
	}
	switch status {
	case "running":
		if job.StartedAt == nil {
			startedAt := now
			updates["started_at"] = &startedAt
		}
		updates["finished_at"] = nil
		updates["last_error"] = ""
		updates["attempt_count"] = job.AttemptCount + 1
	case "failed":
		finishedAt := now
		updates["finished_at"] = &finishedAt
		updates["last_error"] = strings.TrimSpace(lastError)
	case "succeeded":
		finishedAt := now
		updates["finished_at"] = &finishedAt
		updates["last_error"] = ""
	}

	if err := database.Model(&domain.AIAutomationJob{}).Where("id = ?", job.ID).Updates(updates).Error; err != nil {
		return domain.AIAutomationJob{}, err
	}
	job.Status = status
	job.UpdatedAt = now
	if status == "running" {
		startedAt := now
		job.StartedAt = &startedAt
		job.FinishedAt = nil
		job.LastError = ""
		job.AttemptCount++
	} else {
		finishedAt := now
		job.FinishedAt = &finishedAt
		job.LastError = strings.TrimSpace(lastError)
	}
	return job, nil
}

func isRetryableKnowledgeEntityBriefAutomationStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "failed", "succeeded":
		return true
	default:
		return false
	}
}
