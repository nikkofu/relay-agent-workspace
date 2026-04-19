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
	ID             string `gorm:"primaryKey" json:"id"`
	OrganizationID string `json:"org_id"`
	Name           string `json:"name"`
	Email          string `gorm:"unique" json:"email"`
	Avatar         string `json:"avatar"`
	Status         string `json:"status"`
	AIProvider     string `json:"ai_provider"`
	AIModel        string `json:"ai_model"`
	AIMode         string `json:"ai_mode"`
	AIInsight      string `gorm:"-" json:"ai_insight,omitempty"`
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
