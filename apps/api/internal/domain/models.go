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

type Channel struct {
	ID          string `gorm:"primaryKey" json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	MemberCount int    `json:"member_count"`
	UnreadCount int    `json:"unread_count"`
	IsStarred   bool   `json:"is_starred"`
}

type Message struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	ChannelID   string     `json:"channel_id"`
	UserID      string     `json:"user_id"`
	Content     string     `json:"content"`
	ThreadID    string     `json:"thread_id"`
	ReplyCount  int        `json:"reply_count"`
	LastReplyAt *time.Time `json:"last_reply_at"`
	CreatedAt   time.Time  `json:"created_at"`
	Metadata    string     `json:"metadata"` // 用于存储 Reactions/Attachments 的 JSON 字符串
}
