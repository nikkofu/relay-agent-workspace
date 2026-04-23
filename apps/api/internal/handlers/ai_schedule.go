package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

type aiScheduleBookingSlotInput struct {
	StartsAt    string   `json:"starts_at"`
	EndsAt      string   `json:"ends_at"`
	Timezone    string   `json:"timezone"`
	AttendeeIDs []string `json:"attendee_ids"`
}

type aiScheduleBookingRequest struct {
	ComposeID   string                     `json:"compose_id"`
	ChannelID   string                     `json:"channel_id"`
	DMID        string                     `json:"dm_id"`
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	Provider    string                     `json:"provider"`
	Slot        aiScheduleBookingSlotInput `json:"slot"`
}

type aiScheduleBookingResponse struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	ChannelID   string    `json:"channel_id,omitempty"`
	DMID        string    `json:"dm_id,omitempty"`
	RequestedBy string    `json:"requested_by"`
	ComposeID   string    `json:"compose_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	Timezone    string    `json:"timezone"`
	AttendeeIDs []string  `json:"attendee_ids"`
	Provider    string    `json:"provider"`
	Status      string    `json:"status"`
	ExternalRef string    `json:"external_ref,omitempty"`
	ICSContent  string    `json:"ics_content,omitempty"`
	LastError   string    `json:"last_error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func BookAISchedule(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input aiScheduleBookingRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateAIScheduleBookingRequest(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startsAt, endsAt, err := parseScheduleBookingSlot(input.Slot)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspaceID, err := resolveScheduleWorkspaceID(input.ChannelID, input.DMID, currentUser.OrganizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	booking, err := createAIScheduleBooking(db.DB, currentUser.ID, workspaceID, input, startsAt, endsAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create schedule booking"})
		return
	}

	_ = broadcastScheduleBookingEvent("schedule.event.booked", booking)
	c.JSON(http.StatusCreated, gin.H{"booking": hydrateAIScheduleBooking(booking)})
}

func ListAIScheduleBookings(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	channelID := strings.TrimSpace(c.Query("channel_id"))
	dmID := strings.TrimSpace(c.Query("dm_id"))
	if channelID != "" && dmID != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "specify only one of channel_id or dm_id"})
		return
	}

	bookings, err := listAIScheduleBookings(db.DB, currentUser.ID, channelID, dmID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list schedule bookings"})
		return
	}

	payload := make([]aiScheduleBookingResponse, 0, len(bookings))
	for _, booking := range bookings {
		payload = append(payload, hydrateAIScheduleBooking(booking))
	}

	c.JSON(http.StatusOK, gin.H{"bookings": payload})
}

func GetAIScheduleBooking(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	booking, err := getAIScheduleBooking(db.DB, c.Param("id"), currentUser.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schedule booking not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"booking": hydrateAIScheduleBooking(booking)})
}

func CancelAIScheduleBooking(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	booking, changed, err := cancelAIScheduleBooking(db.DB, c.Param("id"), currentUser.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schedule booking not found"})
		return
	}

	if changed {
		_ = broadcastScheduleBookingEvent("schedule.event.cancelled", booking)
	}

	c.JSON(http.StatusOK, gin.H{"booking": hydrateAIScheduleBooking(booking)})
}

func validateAIScheduleBookingRequest(input aiScheduleBookingRequest) error {
	scopeCount := 0
	if strings.TrimSpace(input.ChannelID) != "" {
		scopeCount++
	}
	if strings.TrimSpace(input.DMID) != "" {
		scopeCount++
	}
	if scopeCount != 1 {
		return errors.New("specify exactly one of channel_id or dm_id")
	}
	if strings.TrimSpace(input.ComposeID) == "" {
		return errors.New("compose_id is required")
	}
	if strings.TrimSpace(input.Slot.StartsAt) == "" || strings.TrimSpace(input.Slot.EndsAt) == "" || strings.TrimSpace(input.Slot.Timezone) == "" {
		return errors.New("slot requires starts_at, ends_at, and timezone")
	}
	if len(input.Slot.AttendeeIDs) == 0 {
		return errors.New("slot.attendee_ids is required")
	}
	return nil
}

func parseScheduleBookingSlot(slot aiScheduleBookingSlotInput) (time.Time, time.Time, error) {
	startsAt, err := time.Parse(time.RFC3339, strings.TrimSpace(slot.StartsAt))
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("slot.starts_at must be an RFC3339 timestamp")
	}
	endsAt, err := time.Parse(time.RFC3339, strings.TrimSpace(slot.EndsAt))
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("slot.ends_at must be an RFC3339 timestamp")
	}
	if !endsAt.After(startsAt) {
		return time.Time{}, time.Time{}, errors.New("slot.ends_at must be after slot.starts_at")
	}
	if _, err := time.LoadLocation(strings.TrimSpace(slot.Timezone)); err != nil {
		return time.Time{}, time.Time{}, errors.New("slot.timezone is invalid")
	}
	return startsAt.UTC(), endsAt.UTC(), nil
}

func createAIScheduleBooking(database *gorm.DB, requestedBy, workspaceID string, input aiScheduleBookingRequest, startsAt, endsAt time.Time) (domain.AIScheduleBooking, error) {
	now := time.Now().UTC()
	attendeeIDs := normalizeScheduleAttendeeIDs(input.Slot.AttendeeIDs)
	attendeeIDsJSON, err := json.Marshal(attendeeIDs)
	if err != nil {
		return domain.AIScheduleBooking{}, err
	}
	organizerEmail, attendeeEmails := resolveAIScheduleParticipantEmails(requestedBy, attendeeIDs)

	booking := domain.AIScheduleBooking{
		ID:                    ids.NewPrefixedUUID("ai-sched-booking"),
		WorkspaceID:           strings.TrimSpace(workspaceID),
		ChannelID:             strings.TrimSpace(input.ChannelID),
		DMConversationID:      strings.TrimSpace(input.DMID),
		RequestedBy:           strings.TrimSpace(requestedBy),
		IntentSourceComposeID: strings.TrimSpace(input.ComposeID),
		Title:                 defaultAIScheduleTitle(input.Title),
		Description:           strings.TrimSpace(input.Description),
		StartsAt:              startsAt.UTC(),
		EndsAt:                endsAt.UTC(),
		Timezone:              defaultAIScheduleTimezone(input.Slot.Timezone, startsAt),
		AttendeeIDsJSON:       string(attendeeIDsJSON),
		Provider:              normalizeAIScheduleProvider(input.Provider),
		Status:                "booked",
		ICSContent:            buildAIScheduleICSContent(input, organizerEmail, attendeeEmails, startsAt.UTC(), endsAt.UTC()),
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := database.Create(&booking).Error; err != nil {
		return domain.AIScheduleBooking{}, err
	}

	return booking, nil
}

func listAIScheduleBookings(database *gorm.DB, requestedBy, channelID, dmID string) ([]domain.AIScheduleBooking, error) {
	query := database.Model(&domain.AIScheduleBooking{}).Where("requested_by = ?", strings.TrimSpace(requestedBy))
	if strings.TrimSpace(channelID) != "" {
		query = query.Where("channel_id = ?", strings.TrimSpace(channelID))
	}
	if strings.TrimSpace(dmID) != "" {
		query = query.Where("dm_conversation_id = ?", strings.TrimSpace(dmID))
	}

	var bookings []domain.AIScheduleBooking
	if err := query.Order("created_at desc").Find(&bookings).Error; err != nil {
		return nil, err
	}
	return bookings, nil
}

func getAIScheduleBooking(database *gorm.DB, bookingID, requestedBy string) (domain.AIScheduleBooking, error) {
	var booking domain.AIScheduleBooking
	err := database.Where("id = ? AND requested_by = ?", strings.TrimSpace(bookingID), strings.TrimSpace(requestedBy)).First(&booking).Error
	return booking, err
}

func cancelAIScheduleBooking(database *gorm.DB, bookingID, requestedBy string) (domain.AIScheduleBooking, bool, error) {
	var booking domain.AIScheduleBooking
	if err := database.Where("id = ? AND requested_by = ?", strings.TrimSpace(bookingID), strings.TrimSpace(requestedBy)).First(&booking).Error; err != nil {
		return domain.AIScheduleBooking{}, false, err
	}
	if booking.Status == "cancelled" {
		return booking, false, nil
	}

	booking.Status = "cancelled"
	booking.UpdatedAt = time.Now().UTC()
	if err := database.Save(&booking).Error; err != nil {
		return domain.AIScheduleBooking{}, false, err
	}
	return booking, true, nil
}

func resolveScheduleWorkspaceID(channelID, dmID, fallbackWorkspaceID string) (string, error) {
	if strings.TrimSpace(channelID) != "" {
		var channel domain.Channel
		if err := db.DB.First(&channel, "id = ?", strings.TrimSpace(channelID)).Error; err != nil {
			return "", err
		}
		return channel.WorkspaceID, nil
	}

	var dm domain.DMConversation
	if err := db.DB.First(&dm, "id = ?", strings.TrimSpace(dmID)).Error; err != nil {
		return "", err
	}
	return strings.TrimSpace(fallbackWorkspaceID), nil
}

func hydrateAIScheduleBooking(booking domain.AIScheduleBooking) aiScheduleBookingResponse {
	return aiScheduleBookingResponse{
		ID:          booking.ID,
		WorkspaceID: booking.WorkspaceID,
		ChannelID:   booking.ChannelID,
		DMID:        booking.DMConversationID,
		RequestedBy: booking.RequestedBy,
		ComposeID:   booking.IntentSourceComposeID,
		Title:       booking.Title,
		Description: booking.Description,
		StartsAt:    booking.StartsAt,
		EndsAt:      booking.EndsAt,
		Timezone:    booking.Timezone,
		AttendeeIDs: parseScheduleAttendeeIDs(booking.AttendeeIDsJSON),
		Provider:    booking.Provider,
		Status:      booking.Status,
		ExternalRef: booking.ExternalRef,
		ICSContent:  booking.ICSContent,
		LastError:   booking.LastError,
		CreatedAt:   booking.CreatedAt,
		UpdatedAt:   booking.UpdatedAt,
	}
}

func parseScheduleAttendeeIDs(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	var attendeeIDs []string
	if err := json.Unmarshal([]byte(raw), &attendeeIDs); err != nil {
		return nil
	}
	return attendeeIDs
}

func normalizeScheduleAttendeeIDs(attendeeIDs []string) []string {
	out := make([]string, 0, len(attendeeIDs))
	for _, attendeeID := range attendeeIDs {
		if trimmed := strings.TrimSpace(attendeeID); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func normalizeAIScheduleProvider(provider string) string {
	value := strings.ToLower(strings.TrimSpace(provider))
	if value == "" {
		return "internal"
	}
	return value
}

func defaultAIScheduleTitle(title string) string {
	title = strings.TrimSpace(title)
	if title != "" {
		return title
	}
	return "AI schedule booking"
}

func defaultAIScheduleTimezone(timezone string, startsAt time.Time) string {
	timezone = strings.TrimSpace(timezone)
	if timezone != "" {
		return timezone
	}
	return startsAt.Location().String()
}

func buildAIScheduleICSContent(input aiScheduleBookingRequest, organizerEmail string, attendeeEmails []string, startsAt, endsAt time.Time) string {
	title := defaultAIScheduleTitle(input.Title)
	description := strings.TrimSpace(input.Description)
	if description == "" {
		description = title
	}
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
	}
	if organizerEmail != "" {
		lines = append(lines, fmt.Sprintf("ORGANIZER:mailto:%s", escapeICSValue(organizerEmail)))
	}
	for _, attendeeEmail := range attendeeEmails {
		lines = append(lines, fmt.Sprintf("ATTENDEE:mailto:%s", escapeICSValue(attendeeEmail)))
	}
	lines = append(lines,
		"STATUS:CONFIRMED",
		"END:VEVENT",
		"END:VCALENDAR",
	)
	return strings.Join(lines, "\r\n")
}

func resolveAIScheduleParticipantEmails(requestedBy string, attendeeIDs []string) (string, []string) {
	lookupIDs := make([]string, 0, len(attendeeIDs)+1)
	seen := map[string]struct{}{}
	for _, id := range append([]string{requestedBy}, attendeeIDs...) {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		lookupIDs = append(lookupIDs, id)
	}
	if len(lookupIDs) == 0 {
		return "", nil
	}

	var users []domain.User
	emailByID := map[string]string{}
	if err := db.DB.Where("id IN ?", lookupIDs).Find(&users).Error; err == nil {
		for _, user := range users {
			if email := strings.TrimSpace(user.Email); email != "" {
				emailByID[user.ID] = email
			}
		}
	}

	organizerEmail := emailByID[strings.TrimSpace(requestedBy)]
	attendeeEmails := make([]string, 0, len(attendeeIDs))
	seenEmails := map[string]struct{}{}
	for _, attendeeID := range attendeeIDs {
		email := strings.TrimSpace(emailByID[strings.TrimSpace(attendeeID)])
		if email == "" {
			continue
		}
		if _, ok := seenEmails[email]; ok {
			continue
		}
		seenEmails[email] = struct{}{}
		attendeeEmails = append(attendeeEmails, email)
	}
	return organizerEmail, attendeeEmails
}

func escapeICSValue(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "\r\n", `\n`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	value = strings.ReplaceAll(value, ",", `\,`)
	value = strings.ReplaceAll(value, ";", `\;`)
	return value
}

func broadcastScheduleBookingEvent(eventType string, booking domain.AIScheduleBooking) error {
	if RealtimeHub == nil {
		return nil
	}

	return RealtimeHub.Broadcast(realtime.Event{
		ID:          ids.NewPrefixedUUID("evt"),
		Type:        eventType,
		WorkspaceID: booking.WorkspaceID,
		ChannelID:   booking.ChannelID,
		EntityID:    booking.ID,
		TS:          time.Now().UTC().Format(time.RFC3339Nano),
		Payload: gin.H{
			"booking": hydrateAIScheduleBooking(booking),
		},
	})
}
