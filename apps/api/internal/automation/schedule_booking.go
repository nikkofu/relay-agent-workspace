package automation

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
)

type ScheduleSlotInput struct {
	StartsAt    string   `json:"starts_at"`
	EndsAt      string   `json:"ends_at"`
	Timezone    string   `json:"timezone"`
	AttendeeIDs []string `json:"attendee_ids"`
}

type ScheduleBookingInput struct {
	ComposeID   string
	ChannelID   string
	DMID        string
	Title       string
	Description string
	Provider    string
	Slot        ScheduleSlotInput
}

type ScheduleBookingFilter struct {
	RequestedBy string
	ChannelID   string
	DMID        string
}

func CreateScheduleBooking(db *gorm.DB, input ScheduleBookingInput, requestedBy string, workspaceID string, startsAt, endsAt time.Time) (domain.AIScheduleBooking, error) {
	now := time.Now().UTC()
	provider := normalizeScheduleProvider(input.Provider)
	attendeeIDsJSON, err := json.Marshal(normalizeIDs(input.Slot.AttendeeIDs))
	if err != nil {
		return domain.AIScheduleBooking{}, err
	}

	booking := domain.AIScheduleBooking{
		ID:                    ids.NewPrefixedUUID("ai-sched-booking"),
		WorkspaceID:           workspaceID,
		ChannelID:             strings.TrimSpace(input.ChannelID),
		DMConversationID:      strings.TrimSpace(input.DMID),
		RequestedBy:           strings.TrimSpace(requestedBy),
		IntentSourceComposeID: strings.TrimSpace(input.ComposeID),
		Title:                 defaultScheduleTitle(input.Title),
		Description:           strings.TrimSpace(input.Description),
		StartsAt:              startsAt.UTC(),
		EndsAt:                endsAt.UTC(),
		Timezone:              defaultScheduleTimezone(input.Slot.Timezone, startsAt),
		AttendeeIDsJSON:       string(attendeeIDsJSON),
		Provider:              provider,
		Status:                "booked",
		ICSContent:            buildScheduleICSContent(input, requestedBy, startsAt.UTC(), endsAt.UTC()),
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := db.Create(&booking).Error; err != nil {
		return domain.AIScheduleBooking{}, err
	}

	return booking, nil
}

func ListScheduleBookings(db *gorm.DB, filter ScheduleBookingFilter) ([]domain.AIScheduleBooking, error) {
	query := db.Model(&domain.AIScheduleBooking{})
	if strings.TrimSpace(filter.RequestedBy) != "" {
		query = query.Where("requested_by = ?", strings.TrimSpace(filter.RequestedBy))
	}
	if strings.TrimSpace(filter.ChannelID) != "" {
		query = query.Where("channel_id = ?", strings.TrimSpace(filter.ChannelID))
	}
	if strings.TrimSpace(filter.DMID) != "" {
		query = query.Where("dm_conversation_id = ?", strings.TrimSpace(filter.DMID))
	}

	var bookings []domain.AIScheduleBooking
	if err := query.Order("created_at desc").Find(&bookings).Error; err != nil {
		return nil, err
	}
	return bookings, nil
}

func GetScheduleBooking(db *gorm.DB, bookingID, requestedBy string) (domain.AIScheduleBooking, error) {
	var booking domain.AIScheduleBooking
	err := db.Where("id = ? AND requested_by = ?", strings.TrimSpace(bookingID), strings.TrimSpace(requestedBy)).First(&booking).Error
	return booking, err
}

func CancelScheduleBooking(db *gorm.DB, bookingID, requestedBy string) (domain.AIScheduleBooking, bool, error) {
	var booking domain.AIScheduleBooking
	if err := db.Where("id = ? AND requested_by = ?", strings.TrimSpace(bookingID), strings.TrimSpace(requestedBy)).First(&booking).Error; err != nil {
		return domain.AIScheduleBooking{}, false, err
	}

	if booking.Status == "cancelled" {
		return booking, false, nil
	}

	now := time.Now().UTC()
	booking.Status = "cancelled"
	booking.UpdatedAt = now
	if err := db.Save(&booking).Error; err != nil {
		return domain.AIScheduleBooking{}, false, err
	}

	return booking, true, nil
}

func normalizeScheduleProvider(provider string) string {
	value := strings.ToLower(strings.TrimSpace(provider))
	if value == "" {
		return "internal"
	}
	return value
}

func normalizeIDs(ids []string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func defaultScheduleTitle(title string) string {
	title = strings.TrimSpace(title)
	if title != "" {
		return title
	}
	return "AI schedule booking"
}

func defaultScheduleTimezone(timezone string, startsAt time.Time) string {
	timezone = strings.TrimSpace(timezone)
	if timezone != "" {
		return timezone
	}
	return startsAt.Location().String()
}

func buildScheduleICSContent(input ScheduleBookingInput, requestedBy string, startsAt, endsAt time.Time) string {
	title := defaultScheduleTitle(input.Title)
	description := strings.TrimSpace(input.Description)
	if description == "" {
		description = title
	}
	attendees := normalizeIDs(input.Slot.AttendeeIDs)

	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//Relay Agent Workspace//AI Schedule Booking//EN",
		"CALSCALE:GREGORIAN",
		"METHOD:REQUEST",
		"BEGIN:VEVENT",
		fmt.Sprintf("UID:%s", ids.NewPrefixedUUID("ai-sched-event")),
		fmt.Sprintf("DTSTAMP:%s", startsAt.UTC().Format("20060102T150405Z")),
		fmt.Sprintf("DTSTART:%s", startsAt.UTC().Format("20060102T150405Z")),
		fmt.Sprintf("DTEND:%s", endsAt.UTC().Format("20060102T150405Z")),
		fmt.Sprintf("SUMMARY:%s", escapeICSValue(title)),
		fmt.Sprintf("DESCRIPTION:%s", escapeICSValue(description)),
		fmt.Sprintf("ORGANIZER:mailto:%s", escapeICSValue(requestedBy)),
	}
	for _, attendee := range attendees {
		lines = append(lines, fmt.Sprintf("ATTENDEE:mailto:%s", escapeICSValue(attendee)))
	}
	lines = append(lines,
		"STATUS:CONFIRMED",
		"END:VEVENT",
		"END:VCALENDAR",
	)
	return strings.Join(lines, "\r\n")
}

func escapeICSValue(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "\r\n", `\n`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	value = strings.ReplaceAll(value, ",", `\,`)
	value = strings.ReplaceAll(value, ";", `\;`)
	return value
}
