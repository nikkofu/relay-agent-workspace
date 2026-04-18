package db

import (
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

var DB *gorm.DB

func InitDB() error {
	dbPath := filepath.Join("db", "relay.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return err
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	return DB.AutoMigrate(
		&domain.Organization{},
		&domain.Team{},
		&domain.User{},
		&domain.Agent{},
		&domain.Workspace{},
		&domain.WorkspaceInvite{},
		&domain.DMConversation{},
		&domain.DMMember{},
		&domain.Channel{},
		&domain.ChannelMember{},
		&domain.Message{},
		&domain.MessageReaction{},
		&domain.SavedMessage{},
		&domain.Draft{},
		&domain.UnreadMarker{},
		&domain.NotificationRead{},
		&domain.AIFeedback{},
		&domain.AIConversation{},
		&domain.AIConversationMessage{},
		&domain.DMMessage{},
	)
}
