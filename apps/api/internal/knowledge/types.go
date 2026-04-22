package knowledge

import (
	"time"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

type Citation struct {
	ID           string  `json:"id"`
	EvidenceKind string  `json:"evidence_kind"`
	SourceKind   string  `json:"source_kind"`
	SourceRef    string  `json:"source_ref"`
	RefKind      string  `json:"ref_kind"`
	Locator      string  `json:"locator,omitempty"`
	Snippet      string  `json:"snippet"`
	Title        string  `json:"title,omitempty"`
	Score        float64 `json:"score"`
	EntityID     string  `json:"entity_id,omitempty"`
	EntityTitle  string  `json:"entity_title,omitempty"`
}

type LookupParams struct {
	Query        string
	ChannelID    string
	EntityID     string
	Limit        int
	IncludeKinds map[string]bool
}

type CreateEntityInput struct {
	WorkspaceID string `json:"workspace_id"`
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Status      string `json:"status"`
	OwnerUserID string `json:"owner_user_id"`
	SourceKind  string `json:"source_kind"`
	SourceRef   string `json:"source_ref"`
	Metadata    string `json:"metadata_json"`
}

type UpdateEntityInput struct {
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Status      string `json:"status"`
	OwnerUserID string `json:"owner_user_id"`
	SourceKind  string `json:"source_kind"`
	SourceRef   string `json:"source_ref"`
	Metadata    string `json:"metadata_json"`
}

type AddEntityRefInput struct {
	RefKind  string `json:"ref_kind"`
	RefID    string `json:"ref_id"`
	Role     string `json:"role"`
	Metadata string `json:"metadata_json"`
}

type AddEntityLinkInput struct {
	WorkspaceID  string  `json:"workspace_id"`
	FromEntityID string  `json:"from_entity_id"`
	ToEntityID   string  `json:"to_entity_id"`
	Relation     string  `json:"relation"`
	Weight       float64 `json:"weight"`
	Metadata     string  `json:"metadata_json"`
}

type AddEntityEventInput struct {
	EventType   string `json:"event_type"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	ActorUserID string `json:"actor_user_id"`
	SourceKind  string `json:"source_kind"`
	SourceRef   string `json:"source_ref"`
}

type IngestEventInput struct {
	EntityID    string `json:"entity_id"`
	EventType   string `json:"event_type"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	ActorUserID string `json:"actor_user_id"`
	SourceKind  string `json:"source_kind"`
	SourceRef   string `json:"source_ref"`
}

type GraphNode struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Title      string `json:"title"`
	SourceKind string `json:"source_kind,omitempty"`
	RefKind    string `json:"ref_kind,omitempty"`
	RefID      string `json:"ref_id,omitempty"`
	Role       string `json:"role,omitempty"`
}

type GraphEdge struct {
	ID        string  `json:"id"`
	From      string  `json:"from"`
	To        string  `json:"to"`
	Relation  string  `json:"relation"`
	Weight    float64 `json:"weight"`
	Direction string  `json:"direction,omitempty"`
	Role      string  `json:"role,omitempty"`
}

type EntityGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type ChannelKnowledgeContext struct {
	ChannelID string                `json:"channel_id"`
	Refs      []ChannelKnowledgeRef `json:"refs"`
}

type ChannelKnowledgeSummary struct {
	ChannelID      string                          `json:"channel_id"`
	WindowDays     int                             `json:"window_days"`
	TotalRefs      int                             `json:"total_refs"`
	RecentRefCount int                             `json:"recent_ref_count"`
	Velocity       ChannelKnowledgeVelocity        `json:"velocity"`
	TopEntities    []ChannelKnowledgeSummaryEntity `json:"top_entities"`
}

type ChannelKnowledgeVelocity struct {
	RecentWindowDays int  `json:"recent_window_days"`
	PreviousRefCount int  `json:"previous_ref_count"`
	RecentRefCount   int  `json:"recent_ref_count"`
	Delta            int  `json:"delta"`
	IsSpiking        bool `json:"is_spiking"`
}

type ChannelKnowledgeSummaryEntity struct {
	EntityID        string                       `json:"entity_id"`
	EntityTitle     string                       `json:"entity_title"`
	EntityKind      string                       `json:"entity_kind"`
	RefCount        int                          `json:"ref_count"`
	MessageRefCount int                          `json:"message_ref_count"`
	FileRefCount    int                          `json:"file_ref_count"`
	LastRefAt       time.Time                    `json:"last_ref_at"`
	Trend           []ChannelKnowledgeTrendPoint `json:"trend"`
}

type ChannelKnowledgeTrendPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type ChannelKnowledgeRef struct {
	ID            string    `json:"id"`
	EntityID      string    `json:"entity_id"`
	EntityTitle   string    `json:"entity_title"`
	EntityKind    string    `json:"entity_kind"`
	RefKind       string    `json:"ref_kind"`
	RefID         string    `json:"ref_id"`
	Role          string    `json:"role"`
	SourceTitle   string    `json:"source_title"`
	SourceSnippet string    `json:"source_snippet,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type SuggestEntitiesParams struct {
	Query       string
	ChannelID   string
	WorkspaceID string
	Limit       int
}

type KnowledgeEntitySuggestion struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Kind            string `json:"kind"`
	Summary         string `json:"summary"`
	SourceKind      string `json:"source_kind"`
	RefCount        int    `json:"ref_count"`
	ChannelRefCount int    `json:"channel_ref_count"`
}

type FollowedEntity struct {
	Follow      domain.KnowledgeEntityFollow `json:"follow"`
	Entity      domain.KnowledgeEntity       `json:"entity"`
	IsFollowing bool                         `json:"is_following"`
}

type EntitySpikeAlert struct {
	Entity            domain.KnowledgeEntity `json:"entity"`
	UserIDs           []string               `json:"user_ids"`
	ChannelID         string                 `json:"channel_id"`
	RecentRefCount    int                    `json:"recent_ref_count"`
	PreviousRefCount  int                    `json:"previous_ref_count"`
	Delta             int                    `json:"delta"`
	RelatedChannelIDs []string               `json:"related_channel_ids"`
	OccurredAt        time.Time              `json:"occurred_at"`
}

type MatchEntitiesInput struct {
	WorkspaceID string
	Text        string
	Limit       int
}

type EntityTextMatch struct {
	EntityID    string `json:"entity_id"`
	EntityTitle string `json:"entity_title"`
	EntityKind  string `json:"entity_kind"`
	SourceKind  string `json:"source_kind,omitempty"`
	MatchedText string `json:"matched_text"`
	Start       int    `json:"start"`
	End         int    `json:"end"`
}

type MentionedEntity struct {
	EntityID    string `json:"entity_id"`
	EntityTitle string `json:"entity_title"`
	EntityKind  string `json:"entity_kind"`
	SourceKind  string `json:"source_kind,omitempty"`
	MentionText string `json:"mention_text"`
}

type EntityHoverSummary struct {
	EntityID         string                 `json:"entity_id"`
	RefCount         int                    `json:"ref_count"`
	ChannelRefCount  int                    `json:"channel_ref_count"`
	MessageRefCount  int                    `json:"message_ref_count"`
	FileRefCount     int                    `json:"file_ref_count"`
	RecentWindowDays int                    `json:"recent_window_days"`
	RecentRefCount   int                    `json:"recent_ref_count"`
	LastActivityAt   *time.Time             `json:"last_activity_at,omitempty"`
	RelatedChannels  []EntityChannelSummary `json:"related_channels"`
}

type EntityChannelSummary struct {
	ChannelID      string     `json:"channel_id"`
	Name           string     `json:"name"`
	RefCount       int        `json:"ref_count"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
}

type ChannelKnowledgeDigest struct {
	ChannelID      string                           `json:"channel_id"`
	Window         string                           `json:"window"`
	WindowDays     int                              `json:"window_days"`
	GeneratedAt    time.Time                        `json:"generated_at"`
	TotalRefs      int                              `json:"total_refs"`
	RecentRefCount int                              `json:"recent_ref_count"`
	Headline       string                           `json:"headline"`
	Summary        string                           `json:"summary"`
	TopMovements   []ChannelKnowledgeDigestMovement `json:"top_movements"`
}

type ChannelKnowledgeDigestMovement struct {
	EntityID         string     `json:"entity_id"`
	EntityTitle      string     `json:"entity_title"`
	EntityKind       string     `json:"entity_kind"`
	RefCount         int        `json:"ref_count"`
	RecentRefCount   int        `json:"recent_ref_count"`
	PreviousRefCount int        `json:"previous_ref_count"`
	Delta            int        `json:"delta"`
	LastActivityAt   *time.Time `json:"last_activity_at,omitempty"`
}

type UpsertDigestScheduleInput struct {
	ChannelID  string
	CreatedBy  string
	Window     string `json:"window"`
	Timezone   string `json:"timezone"`
	DayOfWeek  int    `json:"day_of_week"`
	DayOfMonth int    `json:"day_of_month"`
	Hour       int    `json:"hour"`
	Minute     int    `json:"minute"`
	Limit      int    `json:"limit"`
	Pin        bool   `json:"pin"`
	IsEnabled  bool   `json:"is_enabled"`
}

type ChannelKnowledgeDigestSchedule struct {
	ID              string     `json:"id"`
	ChannelID       string     `json:"channel_id"`
	WorkspaceID     string     `json:"workspace_id"`
	CreatedBy       string     `json:"created_by"`
	Window          string     `json:"window"`
	Timezone        string     `json:"timezone"`
	DayOfWeek       int        `json:"day_of_week"`
	DayOfMonth      int        `json:"day_of_month"`
	Hour            int        `json:"hour"`
	Minute          int        `json:"minute"`
	Limit           int        `json:"limit"`
	Pin             bool       `json:"pin"`
	IsEnabled       bool       `json:"is_enabled"`
	LastPublishedAt *time.Time `json:"last_published_at,omitempty"`
	NextRunAt       *time.Time `json:"next_run_at,omitempty"`
}

type PublishChannelDigestInput struct {
	ChannelID  string
	UserID     string
	Window     string
	Limit      int
	Pin        bool
	OccurredAt time.Time
}

type PublishedDigest struct {
	Channel domain.Channel         `json:"channel"`
	Message domain.Message         `json:"message"`
	Digest  ChannelKnowledgeDigest `json:"digest"`
}

type KnowledgeInboxParams struct {
	UserID string
	Scope  string
	Limit  int
}

type KnowledgeInboxItem struct {
	ID         string                 `json:"id"`
	Channel    domain.Channel         `json:"channel"`
	Message    domain.Message         `json:"message"`
	Digest     ChannelKnowledgeDigest `json:"digest"`
	IsRead     bool                   `json:"is_read"`
	OccurredAt time.Time              `json:"occurred_at"`
}

type KnowledgeInboxEntityContext struct {
	EntityID    string           `json:"entity_id"`
	EntityTitle string           `json:"entity_title"`
	EntityKind  string           `json:"entity_kind"`
	Delta       int              `json:"delta"`
	Messages    []domain.Message `json:"messages"`
}

type KnowledgeInboxDetail struct {
	Item           KnowledgeInboxItem            `json:"item"`
	EntityContexts []KnowledgeInboxEntityContext `json:"entity_contexts"`
}

type DigestSchedulePreview struct {
	Schedule     ChannelKnowledgeDigestSchedule `json:"schedule"`
	UpcomingRuns []DigestScheduleUpcomingRun    `json:"upcoming_runs"`
	Digest       ChannelKnowledgeDigest         `json:"digest"`
}

type DigestScheduleUpcomingRun struct {
	RunAt time.Time `json:"run_at"`
}
