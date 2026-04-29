package domain

import "time"

type Organization struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Team struct {
	ID             string `gorm:"primaryKey" json:"id"`
	OrganizationID string `json:"org_id"`
	Name           string `json:"name"`
}

type User struct {
	ID                string     `gorm:"primaryKey" json:"id"`
	OrganizationID    string     `json:"org_id"`
	Name              string     `json:"name"`
	Email             string     `gorm:"unique" json:"email"`
	UserType          string     `json:"user_type"`
	Avatar            string     `json:"avatar"`
	Title             string     `json:"title"`
	Department        string     `json:"department"`
	Timezone          string     `json:"timezone"`
	WorkingHours      string     `json:"working_hours"`
	Pronouns          string     `json:"pronouns"`
	Location          string     `json:"location"`
	Phone             string     `json:"phone"`
	Bio               string     `json:"bio"`
	Status            string     `json:"status"`
	StatusText        string     `json:"status_text"`
	StatusEmoji       string     `json:"status_emoji"`
	ThemePreference   string     `json:"theme_preference"`
	MessageDensity    string     `json:"message_density"`
	Locale            string     `json:"locale"`
	StatusExpiresAt   *time.Time `json:"status_expires_at,omitempty"`
	LastSeenAt        *time.Time `json:"last_seen_at,omitempty"`
	PresenceExpiresAt *time.Time `json:"-"`
	AIProvider        string     `json:"ai_provider"`
	AIModel           string     `json:"ai_model"`
	AIMode            string     `json:"ai_mode"`
	AIInsight         string     `gorm:"-" json:"ai_insight,omitempty"`
}

type Agent struct {
	ID             string `gorm:"primaryKey" json:"id"`
	OrganizationID string `json:"org_id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	OwnerID        string `json:"owner_id"`
}

type Workspace struct {
	ID                         string `gorm:"primaryKey" json:"id"`
	OrganizationID             string `json:"org_id"`
	Name                       string `json:"name"`
	KnowledgeSpikeThreshold    int    `json:"knowledge_spike_threshold"`
	KnowledgeSpikeCooldownMins int    `json:"knowledge_spike_cooldown_mins"`
}

type WorkspaceInvite struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	WorkspaceID string    `gorm:"index" json:"workspace_id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserGroup struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	WorkspaceID string    `gorm:"index" json:"workspace_id"`
	Name        string    `json:"name"`
	Handle      string    `gorm:"index" json:"handle"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserGroupMember struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserGroupID string    `gorm:"index;uniqueIndex:idx_user_group_member" json:"user_group_id"`
	UserID      string    `gorm:"uniqueIndex:idx_user_group_member;index" json:"user_id"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

type WorkflowDefinition struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Trigger     string    `json:"trigger"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   string    `gorm:"index" json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkflowRun struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	WorkflowID   string     `gorm:"index" json:"workflow_id"`
	StartedBy    string     `gorm:"index" json:"started_by"`
	Status       string     `json:"status"`
	Input        string     `json:"input"`
	Summary      string     `json:"summary"`
	RetryOfRunID string     `gorm:"index" json:"retry_of_run_id,omitempty"`
	Error        string     `json:"error,omitempty"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type WorkflowRunStep struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	WorkflowRunID string    `gorm:"index" json:"workflow_run_id"`
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	DurationMS    int       `json:"duration_ms"`
	Detail        string    `json:"detail"`
	CreatedAt     time.Time `json:"created_at"`
}

type WorkflowRunLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	WorkflowRunID string    `gorm:"index" json:"workflow_run_id"`
	Level         string    `json:"level"`
	Message       string    `json:"message"`
	Metadata      string    `json:"metadata"`
	CreatedAt     time.Time `json:"created_at"`
}

type ToolDefinition struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Key         string    `gorm:"uniqueIndex" json:"key"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ToolRun struct {
	ID              string     `gorm:"primaryKey" json:"id"`
	ToolID          string     `gorm:"index" json:"tool_id"`
	TriggeredBy     string     `gorm:"index" json:"triggered_by"`
	Status          string     `json:"status"`
	Input           string     `json:"input"`
	Summary         string     `json:"summary"`
	WritebackTarget string     `json:"writeback_target"`
	WritebackData   string     `json:"writeback_data"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ToolRunLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ToolRunID string    `gorm:"index" json:"tool_run_id"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type DMConversation struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type DMMember struct {
	ID               uint   `gorm:"primaryKey" json:"id"`
	DMConversationID string `gorm:"index;uniqueIndex:idx_dm_conversation_user" json:"dm_id"`
	UserID           string `gorm:"uniqueIndex:idx_dm_conversation_user" json:"user_id"`
}

type Channel struct {
	ID          string `gorm:"primaryKey" json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Topic       string `json:"topic"`
	Purpose     string `json:"purpose"`
	IsArchived  bool   `json:"is_archived"`
	MemberCount int    `json:"member_count"`
	UnreadCount int    `json:"unread_count"`
	IsStarred   bool   `json:"is_starred"`
}

type ChannelMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChannelID string    `gorm:"index;uniqueIndex:idx_channel_user" json:"channel_id"`
	UserID    string    `gorm:"uniqueIndex:idx_channel_user" json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type ChannelPreference struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	ChannelID         string    `gorm:"index;uniqueIndex:idx_channel_preference_user" json:"channel_id"`
	UserID            string    `gorm:"uniqueIndex:idx_channel_preference_user;index" json:"user_id"`
	NotificationLevel string    `json:"notification_level"`
	IsMuted           bool      `json:"is_muted"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type WorkspaceList struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	WorkspaceID string    `gorm:"index" json:"workspace_id"`
	ChannelID   string    `gorm:"index" json:"channel_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedBy   string    `gorm:"index" json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkspaceListItem struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	ListID          string     `gorm:"index" json:"list_id"`
	Content         string     `json:"content"`
	Position        int        `json:"position"`
	IsCompleted     bool       `json:"is_completed"`
	AssignedTo      string     `gorm:"index" json:"assigned_to"`
	DueAt           *time.Time `json:"due_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedBy       string     `gorm:"index" json:"created_by"`
	SourceMessageID string     `json:"source_message_id"`
	SourceChannelID string     `json:"source_channel_id"`
	SourceSnippet   string     `json:"source_snippet"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Message struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	ChannelID   string     `json:"channel_id"`
	UserID      string     `json:"user_id"`
	Content     string     `json:"content"`
	ThreadID    string     `json:"thread_id"`
	ReplyCount  int        `json:"reply_count"`
	LastReplyAt *time.Time `json:"last_reply_at"`
	IsPinned    bool       `json:"is_pinned"`
	CreatedAt   time.Time  `json:"created_at"`
	Metadata    string     `json:"metadata"` // 用于存储 Reactions/Attachments 的 JSON 字符串
}

type MessageMention struct {
	ID                string    `gorm:"primaryKey" json:"id"`
	MessageID         string    `gorm:"index;uniqueIndex:idx_message_mention_message_user_kind" json:"message_id"`
	WorkspaceID       string    `gorm:"index;index:idx_message_mention_workspace_channel;index:idx_message_mention_workspace_dm" json:"workspace_id"`
	ChannelID         string    `gorm:"index;index:idx_message_mention_workspace_channel" json:"channel_id"`
	DMID              string    `gorm:"index;index:idx_message_mention_workspace_dm" json:"dm_id"`
	MentionedUserID   string    `gorm:"index;uniqueIndex:idx_message_mention_message_user_kind;index:idx_message_mention_mentioned_user" json:"mentioned_user_id"`
	MentionedByUserID string    `gorm:"index" json:"mentioned_by_user_id"`
	MentionText       string    `json:"mention_text"`
	MentionKind       string    `gorm:"uniqueIndex:idx_message_mention_message_user_kind;index:idx_message_mention_kind" json:"mention_kind"`
	CreatedAt         time.Time `json:"created_at"`
}

type MessageReaction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID string    `gorm:"index;uniqueIndex:idx_message_user_emoji" json:"message_id"`
	UserID    string    `gorm:"uniqueIndex:idx_message_user_emoji" json:"user_id"`
	Emoji     string    `gorm:"uniqueIndex:idx_message_user_emoji" json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

type SavedMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID string    `gorm:"index;uniqueIndex:idx_saved_message_user" json:"message_id"`
	UserID    string    `gorm:"uniqueIndex:idx_saved_message_user" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Draft struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index;uniqueIndex:idx_draft_user_scope" json:"user_id"`
	Scope     string    `gorm:"uniqueIndex:idx_draft_user_scope" json:"scope"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UnreadMarker struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID string    `gorm:"index;uniqueIndex:idx_unread_message_user" json:"message_id"`
	UserID    string    `gorm:"uniqueIndex:idx_unread_message_user" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type AIFeedback struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID string    `gorm:"index;uniqueIndex:idx_ai_feedback_message_user" json:"message_id"`
	UserID    string    `gorm:"uniqueIndex:idx_ai_feedback_message_user" json:"user_id"`
	IsGood    bool      `json:"is_good"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AIComposeFeedback struct {
	ID               string    `gorm:"primaryKey" json:"id"`
	ComposeID        string    `gorm:"index;uniqueIndex:idx_ai_compose_feedback_compose_user" json:"compose_id"`
	UserID           string    `gorm:"uniqueIndex:idx_ai_compose_feedback_compose_user" json:"user_id"`
	ChannelID        string    `json:"channel_id"`
	DMConversationID string    `gorm:"index" json:"dm_id"`
	ThreadID         string    `json:"thread_id"`
	Intent           string    `json:"intent"`
	Feedback         string    `json:"feedback"`
	SuggestionText   string    `json:"suggestion_text"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AIComposeActivity struct {
	ID               string    `gorm:"primaryKey" json:"id"`
	ComposeID        string    `gorm:"uniqueIndex;index" json:"compose_id"`
	WorkspaceID      string    `gorm:"index" json:"workspace_id"`
	ChannelID        string    `gorm:"index" json:"channel_id,omitempty"`
	DMConversationID string    `gorm:"index" json:"dm_id,omitempty"`
	ThreadID         string    `gorm:"index" json:"thread_id,omitempty"`
	UserID           string    `gorm:"index:idx_ai_compose_activity_user_created" json:"user_id,omitempty"`
	Intent           string    `gorm:"index" json:"intent"`
	SuggestionCount  int       `json:"suggestion_count"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	CreatedAt        time.Time `gorm:"index;index:idx_ai_compose_activity_user_created" json:"created_at"`
}

type AIAutomationJob struct {
	ID            string     `gorm:"primaryKey" json:"id"`
	JobType       string     `gorm:"index:idx_ai_automation_scope_status" json:"job_type"`
	ScopeType     string     `gorm:"index:idx_ai_automation_scope_status;index:idx_ai_automation_scope_created" json:"scope_type"`
	ScopeID       string     `gorm:"index:idx_ai_automation_scope_status;index:idx_ai_automation_scope_created" json:"scope_id"`
	WorkspaceID   string     `gorm:"index" json:"workspace_id"`
	Status        string     `gorm:"index:idx_ai_automation_scope_status" json:"status"`
	TriggerReason string     `json:"trigger_reason"`
	DedupeKey     string     `gorm:"index" json:"dedupe_key"`
	AttemptCount  int        `json:"attempt_count"`
	LastError     string     `json:"last_error,omitempty"`
	ScheduledAt   time.Time  `json:"scheduled_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AIScheduleBooking struct {
	ID                    string    `gorm:"primaryKey" json:"id"`
	WorkspaceID           string    `gorm:"index" json:"workspace_id"`
	ChannelID             string    `gorm:"index" json:"channel_id,omitempty"`
	DMConversationID      string    `gorm:"index" json:"dm_id,omitempty"`
	RequestedBy           string    `gorm:"index" json:"requested_by"`
	IntentSourceComposeID string    `gorm:"index" json:"compose_id"`
	Title                 string    `json:"title"`
	Description           string    `json:"description"`
	StartsAt              time.Time `gorm:"index" json:"starts_at"`
	EndsAt                time.Time `json:"ends_at"`
	Timezone              string    `json:"timezone"`
	AttendeeIDsJSON       string    `json:"attendee_ids_json"`
	Provider              string    `json:"provider"`
	Status                string    `gorm:"index" json:"status"`
	ExternalRef           string    `json:"external_ref,omitempty"`
	ICSContent            string    `json:"ics_content,omitempty"`
	LastError             string    `json:"last_error,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type DMMessage struct {
	ID               string    `gorm:"primaryKey" json:"id"`
	DMConversationID string    `gorm:"index" json:"dm_id"`
	UserID           string    `json:"user_id"`
	Content          string    `json:"content"`
	Metadata         string    `json:"metadata"`
	CreatedAt        time.Time `json:"created_at"`
}

type NotificationRead struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index;uniqueIndex:idx_notification_user_item" json:"user_id"`
	ItemID    string    `gorm:"uniqueIndex:idx_notification_user_item" json:"item_id"`
	ReadAt    time.Time `json:"read_at"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificationItem struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"index" json:"user_id"`
	Type       string    `json:"type"`
	ActorID    string    `json:"actor_id"`
	ChannelID  string    `json:"channel_id,omitempty"`
	DMID       string    `json:"dm_id,omitempty"`
	MessageID  string    `json:"message_id,omitempty"`
	Summary    string    `json:"summary"`
	Metadata   string    `json:"metadata"`
	OccurredAt time.Time `json:"occurred_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type NotificationPreference struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          string    `gorm:"uniqueIndex" json:"user_id"`
	InboxEnabled    bool      `json:"inbox_enabled"`
	MentionsEnabled bool      `json:"mentions_enabled"`
	DMEnabled       bool      `json:"dm_enabled"`
	MuteAll         bool      `json:"mute_all"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type NotificationMuteRule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index:idx_notification_mute_scope" json:"user_id"`
	ScopeType string    `gorm:"index:idx_notification_mute_scope" json:"scope_type"`
	ScopeID   string    `gorm:"index:idx_notification_mute_scope" json:"scope_id"`
	IsMuted   bool      `json:"is_muted"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AIConversation struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"index" json:"user_id"`
	ChannelID  string    `gorm:"index" json:"channel_id"`
	ArtifactID string    `gorm:"index" json:"artifact_id,omitempty"`
	Provider   string    `json:"provider"`
	Model      string    `json:"model"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AIConversationMessage struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	ConversationID string    `gorm:"index" json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	Reasoning      string    `json:"reasoning,omitempty"`
	AISidecarJSON  string    `json:"ai_sidecar,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type AISummary struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	ScopeType     string     `gorm:"index;uniqueIndex:idx_ai_summary_scope" json:"scope_type"`
	ScopeID       string     `gorm:"uniqueIndex:idx_ai_summary_scope" json:"scope_id"`
	ChannelID     string     `gorm:"index" json:"channel_id"`
	Provider      string     `json:"provider"`
	Model         string     `json:"model"`
	Content       string     `json:"content"`
	Reasoning     string     `json:"reasoning,omitempty"`
	MessageCount  int        `json:"message_count"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ChannelAutoSummarySetting struct {
	ID             string     `gorm:"primaryKey" json:"id"`
	ChannelID      string     `gorm:"uniqueIndex;index" json:"channel_id"`
	WorkspaceID    string     `gorm:"index" json:"workspace_id"`
	CreatedBy      string     `gorm:"index" json:"created_by"`
	IsEnabled      bool       `gorm:"index" json:"is_enabled"`
	WindowHours    int        `json:"window_hours"`
	MessageLimit   int        `json:"message_limit"`
	MinNewMessages int        `json:"min_new_messages"`
	Provider       string     `json:"provider"`
	Model          string     `json:"model"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	LastMessageAt  *time.Time `json:"last_message_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type KnowledgeEntityAskAnswer struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	EntityID      string    `gorm:"index" json:"entity_id"`
	WorkspaceID   string    `gorm:"index" json:"workspace_id"`
	UserID        string    `gorm:"index" json:"user_id"`
	Question      string    `json:"question"`
	Answer        string    `json:"answer"`
	Reasoning     string    `json:"reasoning,omitempty"`
	Provider      string    `json:"provider"`
	Model         string    `json:"model"`
	CitationCount int       `json:"citation_count"`
	AnsweredAt    time.Time `gorm:"index" json:"answered_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Artifact struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	ChannelID  string    `gorm:"index" json:"channel_id"`
	Title      string    `json:"title"`
	Version    int       `json:"version"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Content    string    `json:"content"`
	Source     string    `json:"source"`
	TemplateID string    `json:"template_id"`
	Provider   string    `json:"provider,omitempty"`
	Model      string    `json:"model,omitempty"`
	CreatedBy  string    `json:"created_by"`
	UpdatedBy  string    `json:"updated_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ArtifactVersion struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ArtifactID string    `gorm:"index;uniqueIndex:idx_artifact_version" json:"artifact_id"`
	Version    int       `gorm:"uniqueIndex:idx_artifact_version" json:"version"`
	Title      string    `json:"title"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Content    string    `json:"content"`
	Source     string    `json:"source"`
	TemplateID string    `json:"template_id"`
	Provider   string    `json:"provider,omitempty"`
	Model      string    `json:"model,omitempty"`
	UpdatedBy  string    `json:"updated_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type FileAsset struct {
	ID               string     `gorm:"primaryKey" json:"id"`
	ChannelID        string     `gorm:"index" json:"channel_id,omitempty"`
	UploaderID       string     `gorm:"index" json:"uploader_id"`
	Name             string     `json:"name"`
	StoragePath      string     `json:"storage_path"`
	ContentType      string     `json:"content_type"`
	SizeBytes        int64      `json:"size_bytes"`
	Description      string     `json:"description"`
	SourceKind       string     `json:"source_kind"`
	KnowledgeState   string     `json:"knowledge_state"`
	Summary          string     `json:"summary"`
	Tags             string     `json:"-"`
	RetentionDays    int        `json:"retention_days"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	IsArchived       bool       `json:"is_archived"`
	ArchivedAt       *time.Time `json:"archived_at,omitempty"`
	ExtractionStatus string     `json:"extraction_status"`
	ContentSummary   string     `json:"content_summary"`
	LastIndexedAt    *time.Time `json:"last_indexed_at,omitempty"`
	NeedsOCR         bool       `json:"needs_ocr"`
	OCRProvider      string     `json:"ocr_provider"`
	OCRIsMock        bool       `json:"ocr_is_mock"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type FileAssetEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FileID    string    `gorm:"index" json:"file_id"`
	ActorID   string    `gorm:"index" json:"actor_id"`
	Action    string    `json:"action"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

type FileExtraction struct {
	ID             string     `gorm:"primaryKey" json:"id"`
	FileID         string     `gorm:"index;uniqueIndex" json:"file_id"`
	Status         string     `json:"status"`
	Extractor      string     `json:"extractor"`
	ContentText    string     `json:"content_text"`
	ContentSummary string     `json:"content_summary"`
	ErrorCode      string     `json:"error_code"`
	ErrorMessage   string     `json:"error_message"`
	NeedsOCR       bool       `json:"needs_ocr"`
	OCRProvider    string     `json:"ocr_provider"`
	OCRIsMock      bool       `json:"ocr_is_mock"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type FileExtractionChunk struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ExtractionID  string    `gorm:"index" json:"extraction_id"`
	FileID        string    `gorm:"index" json:"file_id"`
	ChunkIndex    int       `json:"chunk_index"`
	Text          string    `json:"text"`
	TokenEstimate int       `json:"token_estimate"`
	LocatorType   string    `json:"locator_type"`
	LocatorValue  string    `json:"locator_value"`
	Heading       string    `json:"heading"`
	CreatedAt     time.Time `json:"created_at"`
}

type KnowledgeEvidenceLink struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	WorkspaceID   string    `gorm:"index" json:"workspace_id"`
	EvidenceKind  string    `json:"evidence_kind"`
	EvidenceRefID string    `gorm:"index" json:"evidence_ref_id"`
	SourceKind    string    `json:"source_kind"`
	SourceRef     string    `json:"source_ref"`
	RefKind       string    `json:"ref_kind"`
	Locator       string    `json:"locator"`
	Snippet       string    `json:"snippet"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type KnowledgeEvidenceEntityRef struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	EvidenceID string    `gorm:"index" json:"evidence_id"`
	EntityID   string    `gorm:"index" json:"entity_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type KnowledgeEntity struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	WorkspaceID string    `gorm:"index" json:"workspace_id"`
	Kind        string    `gorm:"index" json:"kind"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Status      string    `gorm:"index" json:"status"`
	OwnerUserID string    `gorm:"index" json:"owner_user_id"`
	SourceKind  string    `gorm:"index" json:"source_kind"`
	SourceRef   string    `gorm:"index" json:"source_ref"`
	Metadata    string    `json:"metadata_json"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type KnowledgeEntityRef struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	WorkspaceID string    `gorm:"index" json:"workspace_id"`
	EntityID    string    `gorm:"index" json:"entity_id"`
	RefKind     string    `gorm:"index" json:"ref_kind"`
	RefID       string    `gorm:"index" json:"ref_id"`
	Role        string    `gorm:"index" json:"role"`
	Metadata    string    `json:"metadata_json"`
	CreatedAt   time.Time `json:"created_at"`
}

type KnowledgeEntityLink struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	WorkspaceID  string    `gorm:"index" json:"workspace_id"`
	FromEntityID string    `gorm:"index" json:"from_entity_id"`
	ToEntityID   string    `gorm:"index" json:"to_entity_id"`
	Relation     string    `gorm:"index" json:"relation"`
	Weight       float64   `json:"weight"`
	Metadata     string    `json:"metadata_json"`
	CreatedAt    time.Time `json:"created_at"`
}

type KnowledgeEvent struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	WorkspaceID string    `gorm:"index" json:"workspace_id"`
	EntityID    string    `gorm:"index" json:"entity_id"`
	EventType   string    `gorm:"index" json:"event_type"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	ActorUserID string    `gorm:"index" json:"actor_user_id"`
	SourceKind  string    `gorm:"index" json:"source_kind"`
	SourceRef   string    `gorm:"index" json:"source_ref"`
	OccurredAt  time.Time `json:"occurred_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type KnowledgeEntityFollow struct {
	ID                string     `gorm:"primaryKey" json:"id"`
	WorkspaceID       string     `gorm:"index" json:"workspace_id"`
	EntityID          string     `gorm:"index;uniqueIndex:idx_knowledge_entity_follow" json:"entity_id"`
	UserID            string     `gorm:"index;uniqueIndex:idx_knowledge_entity_follow" json:"user_id"`
	NotificationLevel string     `json:"notification_level"`
	LastAlertedAt     *time.Time `json:"last_alerted_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type KnowledgeDigestSchedule struct {
	ID              string     `gorm:"primaryKey" json:"id"`
	ChannelID       string     `gorm:"uniqueIndex;index" json:"channel_id"`
	WorkspaceID     string     `gorm:"index" json:"workspace_id"`
	CreatedBy       string     `gorm:"index" json:"created_by"`
	Window          string     `gorm:"index" json:"window"`
	Timezone        string     `json:"timezone"`
	DayOfWeek       int        `json:"day_of_week"`
	DayOfMonth      int        `json:"day_of_month"`
	Hour            int        `json:"hour"`
	Minute          int        `json:"minute"`
	Limit           int        `json:"limit"`
	Pin             bool       `json:"pin"`
	IsEnabled       bool       `gorm:"index" json:"is_enabled"`
	LastPublishedAt *time.Time `json:"last_published_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type MessageArtifactReference struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	MessageID  string    `gorm:"index;uniqueIndex:idx_message_artifact_ref" json:"message_id"`
	ArtifactID string    `gorm:"uniqueIndex:idx_message_artifact_ref;index" json:"artifact_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type MessageFileAttachment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID string    `gorm:"index;uniqueIndex:idx_message_file_attachment" json:"message_id"`
	FileID    string    `gorm:"uniqueIndex:idx_message_file_attachment;index" json:"file_id"`
	CreatedAt time.Time `json:"created_at"`
}

type FileComment struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	FileID    string    `gorm:"index" json:"file_id"`
	UserID    string    `gorm:"index" json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FileShare struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	FileID    string    `gorm:"index" json:"file_id"`
	ChannelID string    `gorm:"index" json:"channel_id,omitempty"`
	ThreadID  string    `gorm:"index" json:"thread_id,omitempty"`
	MessageID string    `gorm:"index" json:"message_id,omitempty"`
	SharedBy  string    `gorm:"index" json:"shared_by"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type StarredFile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FileID    string    `gorm:"index;uniqueIndex:idx_starred_file" json:"file_id"`
	UserID    string    `gorm:"index;uniqueIndex:idx_starred_file" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type AISidecar struct {
	Reasoning *AIReasoning `json:"reasoning,omitempty"`
	ToolCalls []AIToolCall `json:"tool_calls,omitempty"`
	Usage     *AIUsage     `json:"usage,omitempty"`
	Analysis  any          `json:"analysis,omitempty"` // Phase 69 structured result
}

type AIReasoning struct {
	Summary  string             `json:"summary,omitempty"`
	Segments []ReasoningSegment `json:"segments,omitempty"`
}

type ReasoningSegment struct {
	Text string `json:"text"`
	Kind string `json:"kind"` // thought, step, note
}

type AIExecutionTarget struct {
	Type          string                 `json:"type"` // list, workflow, channel_message
	WorkflowDraft *AIWorkflowDraftTarget `json:"workflow_draft,omitempty"`
	MessageDraft  *AIMessageDraftTarget  `json:"message_draft,omitempty"`
}

type AIWorkflowDraftTarget struct {
	Title string               `json:"title"`
	Goal  string               `json:"goal"`
	Steps []AIWorkflowStepTarget `json:"steps"`
}

type AIWorkflowStepTarget struct {
	Title string `json:"title"`
}

type AIMessageDraftTarget struct {
	ChannelID string `json:"channel_id"`
	Body      string `json:"body"`
}

type AnalysisListDraft struct {
	ID                 string    `gorm:"primaryKey" json:"id"`
	ArtifactID         string    `gorm:"index" json:"artifact_id"`
	ChannelID          string    `gorm:"index" json:"channel_id"`
	AnalysisSnapshotID string    `gorm:"index" json:"analysis_snapshot_id"`
	Title              string    `json:"title"`
	ItemsJSON          string    `json:"items_json"` // JSON array of strings (item titles)
	CreatedBy          string    `gorm:"index" json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
}

type AnalysisWorkflowDraft struct {
	ID                 string    `gorm:"primaryKey" json:"id"`
	ArtifactID         string    `gorm:"index" json:"artifact_id"`
	ChannelID          string    `gorm:"index" json:"channel_id"`
	AnalysisSnapshotID string    `gorm:"index" json:"analysis_snapshot_id"`
	Title              string    `json:"title"`
	Goal               string    `json:"goal"`
	StepsJSON          string    `json:"steps_json"` // JSON array of objects
	CreatedBy          string    `gorm:"index" json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
}

type AnalysisMessageDraft struct {
	ID                 string    `gorm:"primaryKey" json:"id"`
	ArtifactID         string    `gorm:"index" json:"artifact_id"`
	ChannelID          string    `gorm:"index" json:"channel_id"` // target channel
	AnalysisSnapshotID string    `gorm:"index" json:"analysis_snapshot_id"`
	Body               string    `json:"body"`
	CreatedBy          string    `gorm:"index" json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
}

type AIToolCall struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Arguments  string `json:"arguments"`
	Result     string `json:"result,omitempty"`
	DurationMS int    `json:"duration_ms,omitempty"`
}

type AIUsage struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	TotalTokens  int     `json:"total_tokens"`
	CostUSD      float64 `json:"cost_usd,omitempty"`
}

type ExecutionHistoryEvent struct {
	ID                  string    `gorm:"primaryKey" json:"id"`
	EventType           string    `gorm:"index" json:"event_type"` // draft_generated|confirmed|created|published|failed
	Status              string    `gorm:"index" json:"status"`     // success|failed
	ActorUserID         string    `gorm:"index" json:"actor_user_id"`
	AnalysisSnapshotID  string    `gorm:"index" json:"analysis_snapshot_id"`
	NextStepID          string    `gorm:"index" json:"next_step_id,omitempty"`
	StepIndex           *int      `json:"step_index,omitempty"`
	ExecutionTargetType string    `gorm:"index" json:"execution_target_type"` // list|workflow|channel_message
	DraftID             string    `gorm:"index" json:"draft_id,omitempty"`
	DraftType           string    `json:"draft_type,omitempty"`
	CreatedObjectID     string    `gorm:"index" json:"created_object_id,omitempty"`
	CreatedObjectType   string    `json:"created_object_type,omitempty"`
	FailureStage        string    `json:"failure_stage,omitempty"` // draft_generation|confirmation|creation|publish
	ErrorMessage        string    `json:"error_message,omitempty"`
	CreatedAt           time.Time `gorm:"index" json:"created_at"`
}

type WorkspaceView struct {
	ID               string    `gorm:"primaryKey" json:"id"`
	Title            string    `json:"title"`
	ViewType         string    `gorm:"index" json:"view_type"` // list|calendar|search|report|form|channel_messages
	Source           string    `gorm:"index" json:"source"`    // manual|agent|tool|system
	PrimaryChannelID string    `gorm:"index" json:"primary_channel_id,omitempty"`
	Filters          string    `json:"filters"` // JSON string
	Actions          string    `json:"actions"` // JSON string
	CreatedBy        string    `gorm:"index" json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type SalesOrder struct {
	ID                string     `gorm:"primaryKey" json:"id"`
	OrderNumber       string     `gorm:"uniqueIndex" json:"order_number"`
	CustomerName      string     `gorm:"index" json:"customer_name"`
	Amount            float64    `json:"amount"`
	Currency          string     `json:"currency"`
	Stage             string     `gorm:"index" json:"stage"` // lead, qualified, proposal, negotiation, closed_won, closed_lost
	Status            string     `gorm:"index" json:"status"` // active, archived
	ExpectedCloseDate *time.Time `gorm:"index" json:"expected_close_date"`
	OwnerUserID       string     `gorm:"index" json:"owner_user_id"`
	Summary           string     `json:"summary"`
	Tags              string     `json:"tags"` // comma-separated
	SourceChannelID   string     `gorm:"index" json:"source_channel_id,omitempty"`
	SourceMessageID   string     `gorm:"index" json:"source_message_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
