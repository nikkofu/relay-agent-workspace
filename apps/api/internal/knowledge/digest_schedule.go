package knowledge

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

func GetDigestSchedule(database *gorm.DB, channelID string, now time.Time) (*ChannelKnowledgeDigestSchedule, error) {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return nil, errors.New("channel_id is required")
	}

	var schedule domain.KnowledgeDigestSchedule
	if err := database.First(&schedule, "channel_id = ?", channelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	hydrated, err := hydrateDigestSchedule(schedule, now)
	if err != nil {
		return nil, err
	}
	return &hydrated, nil
}

func UpsertDigestSchedule(database *gorm.DB, input UpsertDigestScheduleInput) (ChannelKnowledgeDigestSchedule, error) {
	channelID := strings.TrimSpace(input.ChannelID)
	if channelID == "" {
		return ChannelKnowledgeDigestSchedule{}, errors.New("channel_id is required")
	}

	var channel domain.Channel
	if err := database.Select("id, workspace_id").First(&channel, "id = ?", channelID).Error; err != nil {
		return ChannelKnowledgeDigestSchedule{}, err
	}

	window, _ := normalizeDigestWindow(input.Window)
	timezone := defaultString(strings.TrimSpace(input.Timezone), "UTC")
	if _, err := time.LoadLocation(timezone); err != nil {
		return ChannelKnowledgeDigestSchedule{}, errors.New("invalid timezone")
	}

	dayOfWeek := input.DayOfWeek
	if dayOfWeek < 0 || dayOfWeek > 6 {
		dayOfWeek = 0
	}
	dayOfMonth := input.DayOfMonth
	if dayOfMonth < 1 {
		dayOfMonth = 1
	}
	if dayOfMonth > 28 {
		dayOfMonth = 28
	}
	hour := input.Hour
	if hour < 0 || hour > 23 {
		return ChannelKnowledgeDigestSchedule{}, errors.New("hour must be between 0 and 23")
	}
	minute := input.Minute
	if minute < 0 || minute > 59 {
		return ChannelKnowledgeDigestSchedule{}, errors.New("minute must be between 0 and 59")
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	now := time.Now().UTC()
	schedule := domain.KnowledgeDigestSchedule{
		ID:          newKnowledgeID("kdsch"),
		ChannelID:   channel.ID,
		WorkspaceID: channel.WorkspaceID,
		CreatedBy:   strings.TrimSpace(input.CreatedBy),
		Window:      window,
		Timezone:    timezone,
		DayOfWeek:   dayOfWeek,
		DayOfMonth:  dayOfMonth,
		Hour:        hour,
		Minute:      minute,
		Limit:       limit,
		Pin:         input.Pin,
		IsEnabled:   input.IsEnabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := database.Where("channel_id = ?", channel.ID).
		Assign(domain.KnowledgeDigestSchedule{
			WorkspaceID: channel.WorkspaceID,
			CreatedBy:   schedule.CreatedBy,
			Window:      schedule.Window,
			Timezone:    schedule.Timezone,
			DayOfWeek:   schedule.DayOfWeek,
			DayOfMonth:  schedule.DayOfMonth,
			Hour:        schedule.Hour,
			Minute:      schedule.Minute,
			Limit:       schedule.Limit,
			Pin:         schedule.Pin,
			IsEnabled:   schedule.IsEnabled,
			UpdatedAt:   now,
		}).
		FirstOrCreate(&schedule).Error; err != nil {
		return ChannelKnowledgeDigestSchedule{}, err
	}

	if err := database.First(&schedule, "channel_id = ?", channel.ID).Error; err != nil {
		return ChannelKnowledgeDigestSchedule{}, err
	}
	return hydrateDigestSchedule(schedule, now)
}

func DeleteDigestSchedule(database *gorm.DB, channelID string) error {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return errors.New("channel_id is required")
	}
	return database.Where("channel_id = ?", channelID).Delete(&domain.KnowledgeDigestSchedule{}).Error
}

func ComputeDigestScheduleNextRunAt(schedule ChannelKnowledgeDigestSchedule, now time.Time) (*time.Time, error) {
	loc, err := time.LoadLocation(defaultString(schedule.Timezone, "UTC"))
	if err != nil {
		return nil, err
	}

	localNow := now.In(loc)
	var candidate time.Time
	switch schedule.Window {
	case "daily":
		candidate = time.Date(localNow.Year(), localNow.Month(), localNow.Day(), schedule.Hour, schedule.Minute, 0, 0, loc)
		if !candidate.After(localNow) {
			candidate = candidate.AddDate(0, 0, 1)
		}
	case "monthly":
		candidate = buildMonthlyScheduleCandidate(localNow.Year(), localNow.Month(), schedule.DayOfMonth, schedule.Hour, schedule.Minute, loc)
		if !candidate.After(localNow) {
			nextMonth := localNow.AddDate(0, 1, 0)
			candidate = buildMonthlyScheduleCandidate(nextMonth.Year(), nextMonth.Month(), schedule.DayOfMonth, schedule.Hour, schedule.Minute, loc)
		}
	default:
		candidate = buildWeeklyScheduleCandidate(localNow, schedule.DayOfWeek, schedule.Hour, schedule.Minute, loc)
		if !candidate.After(localNow) {
			candidate = candidate.AddDate(0, 0, 7)
		}
	}
	return &candidate, nil
}

func PublishChannelDigest(database *gorm.DB, input PublishChannelDigestInput) (PublishedDigest, error) {
	channelID := strings.TrimSpace(input.ChannelID)
	userID := strings.TrimSpace(input.UserID)
	if channelID == "" || userID == "" {
		return PublishedDigest{}, errors.New("channel_id and user_id are required")
	}

	var channel domain.Channel
	if err := database.First(&channel, "id = ?", channelID).Error; err != nil {
		return PublishedDigest{}, err
	}

	digest, err := BuildChannelKnowledgeDigest(database, channelID, input.Window, input.Limit)
	if err != nil {
		return PublishedDigest{}, err
	}

	occurredAt := input.OccurredAt.UTC()
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	digest.GeneratedAt = occurredAt

	metadataJSON, err := json.Marshal(struct {
		KnowledgeDigest *ChannelKnowledgeDigest `json:"knowledge_digest,omitempty"`
	}{
		KnowledgeDigest: &digest,
	})
	if err != nil {
		return PublishedDigest{}, err
	}

	message := domain.Message{
		ID:        newKnowledgeID("msg"),
		ChannelID: channelID,
		UserID:    userID,
		Content:   buildDigestMessageContent(digest),
		IsPinned:  input.Pin,
		CreatedAt: occurredAt,
		Metadata:  string(metadataJSON),
	}
	if err := database.Create(&message).Error; err != nil {
		return PublishedDigest{}, err
	}

	return PublishedDigest{
		Channel: channel,
		Message: message,
		Digest:  digest,
	}, nil
}

func ProcessDigestSchedules(database *gorm.DB, now time.Time) ([]PublishedDigest, error) {
	var schedules []domain.KnowledgeDigestSchedule
	if err := database.Where("is_enabled = ?", true).Order("updated_at asc").Find(&schedules).Error; err != nil {
		return nil, err
	}

	published := make([]PublishedDigest, 0)
	for _, schedule := range schedules {
		dueAt, due, err := computeDigestScheduleDueAt(schedule, now)
		if err != nil || !due {
			if err != nil {
				return nil, err
			}
			continue
		}

		result, err := PublishChannelDigest(database, PublishChannelDigestInput{
			ChannelID:  schedule.ChannelID,
			UserID:     schedule.CreatedBy,
			Window:     schedule.Window,
			Limit:      schedule.Limit,
			Pin:        schedule.Pin,
			OccurredAt: now,
		})
		if err != nil {
			return nil, err
		}

		if err := database.Model(&domain.KnowledgeDigestSchedule{}).
			Where("id = ?", schedule.ID).
			Updates(map[string]any{
				"last_published_at": dueAt.UTC(),
				"updated_at":        now.UTC(),
			}).Error; err != nil {
			return nil, err
		}
		published = append(published, result)
	}

	return published, nil
}

func ListKnowledgeInbox(database *gorm.DB, params KnowledgeInboxParams) ([]KnowledgeInboxItem, error) {
	userID := strings.TrimSpace(params.UserID)
	if userID == "" {
		return []KnowledgeInboxItem{}, errors.New("user_id is required")
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var memberships []domain.ChannelMember
	if err := database.Where("user_id = ?", userID).Find(&memberships).Error; err != nil {
		return nil, err
	}
	if len(memberships) == 0 {
		return []KnowledgeInboxItem{}, nil
	}

	channelIDs := make([]string, 0, len(memberships))
	for _, membership := range memberships {
		channelIDs = append(channelIDs, membership.ChannelID)
	}

	var channels []domain.Channel
	if err := database.Where("id IN ?", channelIDs).Find(&channels).Error; err != nil {
		return nil, err
	}
	channelMap := make(map[string]domain.Channel, len(channels))
	filteredChannelIDs := make([]string, 0, len(channels))
	scope := strings.ToLower(strings.TrimSpace(params.Scope))
	for _, channel := range channels {
		if scope == "starred" && !channel.IsStarred {
			continue
		}
		channelMap[channel.ID] = channel
		filteredChannelIDs = append(filteredChannelIDs, channel.ID)
	}
	if len(filteredChannelIDs) == 0 {
		return []KnowledgeInboxItem{}, nil
	}

	var messages []domain.Message
	if err := database.Where("channel_id IN ? AND metadata LIKE ?", filteredChannelIDs, "%knowledge_digest%").
		Order("created_at desc").
		Limit(limit * 4).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	items := make([]KnowledgeInboxItem, 0, len(messages))
	for _, message := range messages {
		digest, ok := extractKnowledgeDigestFromMessage(message)
		if !ok {
			continue
		}
		channel, exists := channelMap[message.ChannelID]
		if !exists {
			continue
		}
		items = append(items, KnowledgeInboxItem{
			ID:         "knowledge-digest-" + message.ID,
			Channel:    channel,
			Message:    message,
			Digest:     digest,
			OccurredAt: message.CreatedAt,
		})
	}

	if len(items) == 0 {
		return items, nil
	}

	itemIDs := make([]string, 0, len(items))
	for _, item := range items {
		itemIDs = append(itemIDs, item.ID)
	}
	var reads []domain.NotificationRead
	if err := database.Where("user_id = ? AND item_id IN ?", userID, itemIDs).Find(&reads).Error; err != nil {
		return nil, err
	}
	readSet := make(map[string]struct{}, len(reads))
	for _, read := range reads {
		readSet[read.ItemID] = struct{}{}
	}
	for idx := range items {
		_, items[idx].IsRead = readSet[items[idx].ID]
	}

	sortKnowledgeInboxItems(items)
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func hydrateDigestSchedule(schedule domain.KnowledgeDigestSchedule, now time.Time) (ChannelKnowledgeDigestSchedule, error) {
	result := ChannelKnowledgeDigestSchedule{
		ID:              schedule.ID,
		ChannelID:       schedule.ChannelID,
		WorkspaceID:     schedule.WorkspaceID,
		CreatedBy:       schedule.CreatedBy,
		Window:          schedule.Window,
		Timezone:        schedule.Timezone,
		DayOfWeek:       schedule.DayOfWeek,
		DayOfMonth:      schedule.DayOfMonth,
		Hour:            schedule.Hour,
		Minute:          schedule.Minute,
		Limit:           schedule.Limit,
		Pin:             schedule.Pin,
		IsEnabled:       schedule.IsEnabled,
		LastPublishedAt: schedule.LastPublishedAt,
	}
	nextRunAt, err := ComputeDigestScheduleNextRunAt(result, now)
	if err != nil {
		return ChannelKnowledgeDigestSchedule{}, err
	}
	result.NextRunAt = nextRunAt
	return result, nil
}

func computeDigestScheduleDueAt(schedule domain.KnowledgeDigestSchedule, now time.Time) (*time.Time, bool, error) {
	loc, err := time.LoadLocation(defaultString(schedule.Timezone, "UTC"))
	if err != nil {
		return nil, false, err
	}
	localNow := now.In(loc)

	var dueAt time.Time
	switch schedule.Window {
	case "daily":
		dueAt = time.Date(localNow.Year(), localNow.Month(), localNow.Day(), schedule.Hour, schedule.Minute, 0, 0, loc)
		if dueAt.After(localNow) {
			dueAt = dueAt.AddDate(0, 0, -1)
		}
	case "monthly":
		dueAt = buildMonthlyScheduleCandidate(localNow.Year(), localNow.Month(), schedule.DayOfMonth, schedule.Hour, schedule.Minute, loc)
		if dueAt.After(localNow) {
			prevMonth := localNow.AddDate(0, -1, 0)
			dueAt = buildMonthlyScheduleCandidate(prevMonth.Year(), prevMonth.Month(), schedule.DayOfMonth, schedule.Hour, schedule.Minute, loc)
		}
	default:
		dueAt = buildWeeklyScheduleCandidate(localNow, schedule.DayOfWeek, schedule.Hour, schedule.Minute, loc)
		if dueAt.After(localNow) {
			dueAt = dueAt.AddDate(0, 0, -7)
		}
	}

	if dueAt.After(localNow) {
		return nil, false, nil
	}
	if schedule.LastPublishedAt != nil && !schedule.LastPublishedAt.Before(dueAt.UTC()) {
		return &dueAt, false, nil
	}
	return &dueAt, true, nil
}

func buildWeeklyScheduleCandidate(now time.Time, dayOfWeek int, hour int, minute int, loc *time.Location) time.Time {
	offset := dayOfWeek - int(now.Weekday())
	candidate := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
	return candidate.AddDate(0, 0, offset)
}

func buildMonthlyScheduleCandidate(year int, month time.Month, day int, hour int, minute int, loc *time.Location) time.Time {
	lastDay := daysInMonth(year, month, loc)
	if day < 1 {
		day = 1
	}
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, month, day, hour, minute, 0, 0, loc)
}

func daysInMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

func buildDigestMessageContent(digest ChannelKnowledgeDigest) string {
	lines := []string{
		"Knowledge digest (" + digest.Window + ")",
		digest.Headline,
		digest.Summary,
	}
	for idx, movement := range digest.TopMovements {
		if idx >= 3 {
			break
		}
		lines = append(lines, "- "+movement.EntityTitle+": "+itoa(movement.RecentRefCount)+" recent refs")
	}
	return strings.Join(lines, "\n")
}

func extractKnowledgeDigestFromMessage(message domain.Message) (ChannelKnowledgeDigest, bool) {
	if strings.TrimSpace(message.Metadata) == "" || !strings.Contains(message.Metadata, "knowledge_digest") {
		return ChannelKnowledgeDigest{}, false
	}

	var payload struct {
		KnowledgeDigest *ChannelKnowledgeDigest `json:"knowledge_digest"`
	}
	if err := json.Unmarshal([]byte(message.Metadata), &payload); err != nil || payload.KnowledgeDigest == nil {
		return ChannelKnowledgeDigest{}, false
	}
	return *payload.KnowledgeDigest, true
}

func sortKnowledgeInboxItems(items []KnowledgeInboxItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].OccurredAt.Equal(items[j].OccurredAt) {
			return items[i].Message.ID < items[j].Message.ID
		}
		return items[i].OccurredAt.After(items[j].OccurredAt)
	})
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
