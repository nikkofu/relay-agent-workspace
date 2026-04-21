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

	if err := DB.AutoMigrate(
		&domain.Organization{},
		&domain.Team{},
		&domain.User{},
		&domain.Agent{},
		&domain.Workspace{},
		&domain.WorkspaceInvite{},
		&domain.UserGroup{},
		&domain.UserGroupMember{},
		&domain.WorkflowDefinition{},
		&domain.WorkflowRun{},
		&domain.WorkflowRunStep{},
		&domain.WorkflowRunLog{},
		&domain.ToolDefinition{},
		&domain.ToolRun{},
		&domain.ToolRunLog{},
		&domain.DMConversation{},
		&domain.DMMember{},
		&domain.Channel{},
		&domain.ChannelMember{},
		&domain.ChannelPreference{},
		&domain.WorkspaceList{},
		&domain.WorkspaceListItem{},
		&domain.Message{},
		&domain.MessageReaction{},
		&domain.SavedMessage{},
		&domain.Draft{},
		&domain.UnreadMarker{},
		&domain.NotificationRead{},
		&domain.NotificationPreference{},
		&domain.NotificationMuteRule{},
		&domain.AIFeedback{},
		&domain.AIConversation{},
		&domain.AIConversationMessage{},
		&domain.AISummary{},
		&domain.Artifact{},
		&domain.ArtifactVersion{},
		&domain.FileAsset{},
		&domain.FileAssetEvent{},
		&domain.FileExtraction{},
		&domain.FileExtractionChunk{},
		&domain.FileComment{},
		&domain.FileShare{},
		&domain.StarredFile{},
		&domain.MessageArtifactReference{},
		&domain.MessageFileAttachment{},
		&domain.DMMessage{},
	); err != nil {
		return err
	}

	// Fixup: Rename Acme Corp to Relay
	DB.Model(&domain.Organization{}).Where("name = ?", "Acme Corp").Update("name", "Relay")
	DB.Model(&domain.Workspace{}).Where("name = ?", "Acme Corp").Update("name", "Relay")
	DB.Model(&domain.Artifact{}).Where("version = 0").Update("version", 1)

	var artifacts []domain.Artifact
	if err := DB.Find(&artifacts).Error; err == nil {
		for _, artifact := range artifacts {
			var count int64
			DB.Model(&domain.ArtifactVersion{}).
				Where("artifact_id = ? AND version = ?", artifact.ID, artifact.Version).
				Count(&count)
			if count > 0 {
				continue
			}
			DB.Create(&domain.ArtifactVersion{
				ArtifactID: artifact.ID,
				Version:    artifact.Version,
				Title:      artifact.Title,
				Type:       artifact.Type,
				Status:     artifact.Status,
				Content:    artifact.Content,
				Source:     artifact.Source,
				Provider:   artifact.Provider,
				Model:      artifact.Model,
				UpdatedBy:  artifact.UpdatedBy,
				CreatedAt:  artifact.UpdatedAt,
			})
		}
	}

	return nil
}
