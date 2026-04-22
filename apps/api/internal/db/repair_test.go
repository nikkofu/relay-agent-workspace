package db

import (
	"testing"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepairLegacyWorkspaceIDsMovesMockWorkspaceChannels(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := database.AutoMigrate(&domain.Workspace{}, &domain.Channel{}, &domain.ChannelMember{}, &domain.ChannelPreference{}, &domain.Message{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	database.Create(&domain.Workspace{ID: "ws-1", Name: "Relay"})
	database.Create(&domain.Channel{ID: "ch-game", WorkspaceID: "ws_1", Name: "game", Type: "public"})

	if err := RepairLegacyWorkspaceIDs(database); err != nil {
		t.Fatalf("repair failed: %v", err)
	}

	var channel domain.Channel
	if err := database.First(&channel, "id = ?", "ch-game").Error; err != nil {
		t.Fatalf("failed to load repaired channel: %v", err)
	}
	if channel.WorkspaceID != "ws-1" {
		t.Fatalf("expected legacy channel to move to ws-1, got %q", channel.WorkspaceID)
	}
}

func TestRepairLegacyWorkspaceIDsRemovesDuplicateEmptyMockChannels(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := database.AutoMigrate(&domain.Workspace{}, &domain.Channel{}, &domain.ChannelMember{}, &domain.ChannelPreference{}, &domain.Message{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	database.Create(&domain.Workspace{ID: "ws-1", Name: "Relay"})
	database.Create(&domain.Channel{ID: "ch-game-a", WorkspaceID: "ws_1", Name: "game", Type: "public"})
	database.Create(&domain.Channel{ID: "ch-game-b", WorkspaceID: "ws_1", Name: "game", Type: "public"})

	if err := RepairLegacyWorkspaceIDs(database); err != nil {
		t.Fatalf("repair failed: %v", err)
	}

	var count int64
	database.Model(&domain.Channel{}).Where("workspace_id = ? AND name = ?", "ws-1", "game").Count(&count)
	if count != 1 {
		t.Fatalf("expected one recovered game channel, got %d", count)
	}
	database.Model(&domain.Channel{}).Where("workspace_id = ?", "ws_1").Count(&count)
	if count != 0 {
		t.Fatalf("expected no legacy workspace channels, got %d", count)
	}
}
