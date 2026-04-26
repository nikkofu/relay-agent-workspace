package db

import (
	"os"
	"path/filepath"
	"strings"

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
		&domain.MessageMention{},
		&domain.MessageReaction{},
		&domain.SavedMessage{},
		&domain.Draft{},
		&domain.UnreadMarker{},
		&domain.NotificationRead{},
		&domain.NotificationItem{},
		&domain.NotificationPreference{},
		&domain.NotificationMuteRule{},
		&domain.AIFeedback{},
		&domain.AIComposeFeedback{},
		&domain.AIComposeActivity{},
		&domain.AIAutomationJob{},
		&domain.AIScheduleBooking{},
		&domain.AIConversation{},
		&domain.AIConversationMessage{},
		&domain.AISummary{},
		&domain.ChannelAutoSummarySetting{},
		&domain.KnowledgeEntityAskAnswer{},
		&domain.Artifact{},
		&domain.ArtifactVersion{},
		&domain.FileAsset{},
		&domain.FileAssetEvent{},
		&domain.FileExtraction{},
		&domain.FileExtractionChunk{},
		&domain.KnowledgeEvidenceLink{},
		&domain.KnowledgeEvidenceEntityRef{},
		&domain.KnowledgeEntity{},
		&domain.KnowledgeEntityRef{},
		&domain.KnowledgeEntityLink{},
		&domain.KnowledgeEvent{},
		&domain.KnowledgeEntityFollow{},
		&domain.KnowledgeDigestSchedule{},
		&domain.FileComment{},
		&domain.FileShare{},
		&domain.StarredFile{},
		&domain.MessageArtifactReference{},
		&domain.MessageFileAttachment{},
		&domain.DMMessage{},
		&domain.AnalysisListDraft{},
		&domain.AnalysisWorkflowDraft{},
		&domain.AnalysisMessageDraft{},
		&domain.ExecutionHistoryEvent{},
		&domain.WorkspaceView{},
		); err != nil {
		return err
	}

	// Fixup: Rename Acme Corp to Relay
	DB.Model(&domain.Organization{}).Where("name = ?", "Acme Corp").Update("name", "Relay")
	DB.Model(&domain.Workspace{}).Where("name = ?", "Acme Corp").Update("name", "Relay")
	DB.Model(&domain.Artifact{}).Where("version = 0").Update("version", 1)
	if err := RepairLegacyWorkspaceIDs(DB); err != nil {
		return err
	}

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

func RepairLegacyWorkspaceIDs(database *gorm.DB) error {
	var relayWorkspace domain.Workspace
	if err := database.First(&relayWorkspace, "id = ?", "ws-1").Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	var existing []domain.Channel
	if err := database.Where("workspace_id = ?", relayWorkspace.ID).Find(&existing).Error; err != nil {
		return err
	}
	recoveredNames := make(map[string]bool, len(existing))
	for _, channel := range existing {
		recoveredNames[strings.ToLower(channel.Name)] = true
	}

	var legacyChannels []domain.Channel
	if err := database.Where("workspace_id = ?", "ws_1").Order("id asc").Find(&legacyChannels).Error; err != nil {
		return err
	}
	for _, channel := range legacyChannels {
		nameKey := strings.ToLower(channel.Name)
		var messageCount int64
		if err := database.Model(&domain.Message{}).Where("channel_id = ?", channel.ID).Count(&messageCount).Error; err != nil {
			return err
		}

		if recoveredNames[nameKey] && messageCount == 0 {
			if err := database.Where("channel_id = ?", channel.ID).Delete(&domain.ChannelMember{}).Error; err != nil {
				return err
			}
			if err := database.Where("channel_id = ?", channel.ID).Delete(&domain.ChannelPreference{}).Error; err != nil {
				return err
			}
			if err := database.Delete(&domain.Channel{}, "id = ?", channel.ID).Error; err != nil {
				return err
			}
			continue
		}

		if err := database.Model(&domain.Channel{}).
			Where("id = ?", channel.ID).
			Update("workspace_id", relayWorkspace.ID).Error; err != nil {
			return err
		}
		recoveredNames[nameKey] = true
	}

	return nil
}
