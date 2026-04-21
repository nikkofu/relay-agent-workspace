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
	ID             string `gorm:"primaryKey" json:"id"`
	OrganizationID string `json:"org_id"`
	Name           string `json:"name"`
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

type DMMessage struct {
	ID               string    `gorm:"primaryKey" json:"id"`
	DMConversationID string    `gorm:"index" json:"dm_id"`
	UserID           string    `json:"user_id"`
	Content          string    `json:"content"`
	CreatedAt        time.Time `json:"created_at"`
}

type NotificationRead struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index;uniqueIndex:idx_notification_user_item" json:"user_id"`
	ItemID    string    `gorm:"uniqueIndex:idx_notification_user_item" json:"item_id"`
	ReadAt    time.Time `json:"read_at"`
	CreatedAt time.Time `json:"created_at"`
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
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"user_id"`
	ChannelID string    `gorm:"index" json:"channel_id"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AIConversationMessage struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	ConversationID string    `gorm:"index" json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	Reasoning      string    `json:"reasoning,omitempty"`
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

type Artifact struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	ChannelID string    `gorm:"index" json:"channel_id"`
	Title     string    `json:"title"`
	Version   int       `json:"version"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Content   string    `json:"content"`
	Source    string    `json:"source"`
	Provider  string    `json:"provider,omitempty"`
	Model     string    `json:"model,omitempty"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	Provider   string    `json:"provider,omitempty"`
	Model      string    `json:"model,omitempty"`
	UpdatedBy  string    `json:"updated_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type FileAsset struct {
	ID            string     `gorm:"primaryKey" json:"id"`
	ChannelID     string     `gorm:"index" json:"channel_id,omitempty"`
	UploaderID    string     `gorm:"index" json:"uploader_id"`
	Name          string     `json:"name"`
	StoragePath   string     `json:"storage_path"`
	ContentType   string     `json:"content_type"`
	SizeBytes     int64      `json:"size_bytes"`
	Description   string     `json:"description"`
	RetentionDays int        `json:"retention_days"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	IsArchived    bool       `json:"is_archived"`
	ArchivedAt    *time.Time `json:"archived_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type FileAssetEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FileID    string    `gorm:"index" json:"file_id"`
	ActorID   string    `gorm:"index" json:"actor_id"`
	Action    string    `json:"action"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
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
