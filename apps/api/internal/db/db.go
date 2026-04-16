package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

var DB *gorm.DB

func InitDB() error {
	var err error
	DB, err = gorm.Open(sqlite.Open("relay.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	return DB.AutoMigrate(
		&domain.Organization{},
		&domain.Team{},
		&domain.User{},
		&domain.Agent{},
		&domain.Workspace{},
		&domain.Channel{},
		&domain.Message{},
	)
}
