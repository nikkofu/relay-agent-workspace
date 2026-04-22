package db

import (
	"encoding/json"
	"time"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

func SeedData() {
	seedNow := time.Now().UTC()
	at := func(offset time.Duration) time.Time {
		return seedNow.Add(offset)
	}
	atPtr := func(offset time.Duration) *time.Time {
		value := at(offset)
		return &value
	}

	// 1. Organizations
	org := domain.Organization{ID: "ws-1", Name: "Relay"}
	DB.FirstOrCreate(&org, domain.Organization{ID: org.ID})

	org2 := domain.Organization{ID: "ws-2", Name: "Side Project"}
	DB.FirstOrCreate(&org2, domain.Organization{ID: org2.ID})

	// 2. Users (from mock-data.ts)
	users := []domain.User{
		{
			ID:             "user-1",
			OrganizationID: org.ID,
			Name:           "Nikko Fu",
			Email:          "nikko@example.com",
			Avatar:         "https://github.com/nikkofu.png",
			Title:          "Founder",
			Department:     "Product",
			Timezone:       "Asia/Shanghai",
			WorkingHours:   "Mon - Fri, 10:00 - 19:00",
			Pronouns:       "he/him",
			Location:       "Shanghai",
			Phone:          "+86 13800000000",
			Bio:            "Designing Relay as an AI-native workspace for humans and agents.",
			Status:         "online",
			AIProvider:     "gemini",
			AIModel:        "gemini-3-flash-preview",
			AIMode:         "fast",
		},
		{
			ID:             "user-2",
			OrganizationID: org.ID,
			Name:           "AI Assistant",
			Email:          "ai@acim.ai",
			Avatar:         "/ai-wand-avatar.svg",
			Title:          "Workspace AI",
			Department:     "Automation",
			Timezone:       "UTC",
			WorkingHours:   "Always on",
			Status:         "online",
		},
		{
			ID:             "user-3",
			OrganizationID: org.ID,
			Name:           "John Doe",
			Email:          "john@example.com",
			Avatar:         "https://i.pravatar.cc/150?u=john",
			Title:          "Engineering Manager",
			Department:     "Engineering",
			Timezone:       "America/Los_Angeles",
			WorkingHours:   "Mon - Fri, 09:00 - 18:00",
			Status:         "away",
		},
		{
			ID:             "user-4",
			OrganizationID: org.ID,
			Name:           "Jane Smith",
			Email:          "jane@example.com",
			Avatar:         "https://i.pravatar.cc/150?u=jane",
			Title:          "Design Lead",
			Department:     "Design",
			Timezone:       "Europe/London",
			WorkingHours:   "Mon - Fri, 09:00 - 17:30",
			Status:         "offline",
		},
	}

	for _, u := range users {
		DB.FirstOrCreate(&u, domain.User{ID: u.ID})
	}

	// 3. Workspaces (from mock-data.ts)
	workspaces := []domain.Workspace{
		{ID: "ws-1", OrganizationID: org.ID, Name: "Relay", KnowledgeSpikeThreshold: 3, KnowledgeSpikeCooldownMins: 360},
		{ID: "ws-2", OrganizationID: org2.ID, Name: "Side Project", KnowledgeSpikeThreshold: 3, KnowledgeSpikeCooldownMins: 360},
	}

	for _, w := range workspaces {
		DB.FirstOrCreate(&w, domain.Workspace{ID: w.ID})
	}

	// 4. Channels (from mock-data.ts)
	channels := []domain.Channel{
		{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public", Description: "Company-wide announcements and work-based matters", Topic: "Announcements and team alignment", Purpose: "Keep everyone aligned on company progress", MemberCount: 15, IsStarred: true},
		{ID: "ch-2", WorkspaceID: "ws-1", Name: "random", Type: "public", Description: "Non-work-related banter and water cooler talk", Topic: "Watercooler", Purpose: "Casual team conversation", MemberCount: 12},
		{ID: "ch-3", WorkspaceID: "ws-1", Name: "engineering", Type: "public", Description: "Technical discussions and code reviews", Topic: "Build and ship", Purpose: "Coordinate engineering decisions", MemberCount: 8, UnreadCount: 3},
		{ID: "ch-4", WorkspaceID: "ws-1", Name: "design", Type: "public", Description: "UI/UX design critiques and inspiration", Topic: "Design reviews", Purpose: "Centralize design critique and exploration", MemberCount: 5},
		{ID: "ch-5", WorkspaceID: "ws-1", Name: "ai-lab", Type: "public", Description: "Exploring the future of AI-native applications", Topic: "AI-native collaboration", Purpose: "Prototype the Relay AI surface", MemberCount: 10, IsStarred: true, UnreadCount: 12},
		{ID: "ch-6", WorkspaceID: "ws-1", Name: "private-deal", Type: "private", Topic: "Sensitive planning", Purpose: "Private coordination for leadership topics", MemberCount: 3},
		{ID: "ch-collab", WorkspaceID: "ws-1", Name: "agent-collab", Type: "public", Description: "Real-time Agent collaboration and synchronization hub", Topic: "Agent command center", Purpose: "Track humans and agents working together", MemberCount: 3, IsStarred: true},
	}

	for _, ch := range channels {
		DB.FirstOrCreate(&ch, domain.Channel{ID: ch.ID})
	}

	channelMembers := []domain.ChannelMember{
		{ChannelID: "ch-1", UserID: "user-1", Role: "owner"},
		{ChannelID: "ch-1", UserID: "user-2", Role: "member"},
		{ChannelID: "ch-1", UserID: "user-3", Role: "member"},
		{ChannelID: "ch-1", UserID: "user-4", Role: "member"},
		{ChannelID: "ch-3", UserID: "user-1", Role: "owner"},
		{ChannelID: "ch-3", UserID: "user-2", Role: "member"},
		{ChannelID: "ch-3", UserID: "user-3", Role: "member"},
		{ChannelID: "ch-5", UserID: "user-1", Role: "owner"},
		{ChannelID: "ch-5", UserID: "user-2", Role: "member"},
		{ChannelID: "ch-5", UserID: "user-4", Role: "member"},
		{ChannelID: "ch-collab", UserID: "user-1", Role: "owner"},
		{ChannelID: "ch-collab", UserID: "user-2", Role: "member"},
	}
	for _, member := range channelMembers {
		DB.FirstOrCreate(&member, domain.ChannelMember{
			ChannelID: member.ChannelID,
			UserID:    member.UserID,
		})
	}

	channelPreferences := []domain.ChannelPreference{
		{ChannelID: "ch-1", UserID: "user-1", NotificationLevel: "all", IsMuted: false, CreatedAt: at(-6 * time.Hour), UpdatedAt: at(-6 * time.Hour)},
		{ChannelID: "ch-collab", UserID: "user-1", NotificationLevel: "mentions", IsMuted: false, CreatedAt: at(-5 * time.Hour), UpdatedAt: at(-5 * time.Hour)},
	}
	for _, preference := range channelPreferences {
		DB.FirstOrCreate(&preference, domain.ChannelPreference{
			ChannelID: preference.ChannelID,
			UserID:    preference.UserID,
		})
	}

	workspaceInvites := []domain.WorkspaceInvite{
		{
			ID:          "invite-1",
			WorkspaceID: "ws-1",
			Email:       "new.designer@example.com",
			Role:        "member",
			Status:      "pending",
			CreatedAt:   at(-96 * time.Hour),
		},
	}
	for _, invite := range workspaceInvites {
		DB.FirstOrCreate(&invite, domain.WorkspaceInvite{ID: invite.ID})
	}

	userGroups := []domain.UserGroup{
		{
			ID:          "group-1",
			WorkspaceID: "ws-1",
			Name:        "Leadership",
			Handle:      "leadership",
			Description: "Cross-functional leadership group for company planning.",
			CreatedBy:   "user-1",
			CreatedAt:   at(-72 * time.Hour),
			UpdatedAt:   at(-24 * time.Hour),
		},
		{
			ID:          "group-2",
			WorkspaceID: "ws-1",
			Name:        "Design Guild",
			Handle:      "design-guild",
			Description: "Design systems and UX critique circle.",
			CreatedBy:   "user-4",
			CreatedAt:   at(-70 * time.Hour),
			UpdatedAt:   at(-23 * time.Hour),
		},
	}
	for _, group := range userGroups {
		DB.FirstOrCreate(&group, domain.UserGroup{ID: group.ID})
	}

	groupMembers := []domain.UserGroupMember{
		{UserGroupID: "group-1", UserID: "user-1", Role: "owner", CreatedAt: at(-72 * time.Hour)},
		{UserGroupID: "group-1", UserID: "user-3", Role: "member", CreatedAt: at(-71 * time.Hour)},
		{UserGroupID: "group-2", UserID: "user-4", Role: "owner", CreatedAt: at(-70 * time.Hour)},
		{UserGroupID: "group-2", UserID: "user-1", Role: "member", CreatedAt: at(-69 * time.Hour)},
	}
	for _, member := range groupMembers {
		DB.FirstOrCreate(&member, domain.UserGroupMember{
			UserGroupID: member.UserGroupID,
			UserID:      member.UserID,
		})
	}

	workflows := []domain.WorkflowDefinition{
		{
			ID:          "workflow-1",
			Name:        "Daily Standup",
			Category:    "communication",
			Description: "Collect daily progress updates and blockers from the team.",
			Trigger:     "manual",
			IsActive:    true,
			CreatedAt:   at(-48 * time.Hour),
			UpdatedAt:   at(-20 * time.Hour),
		},
		{
			ID:          "workflow-2",
			Name:        "Launch Review",
			Category:    "delivery",
			Description: "Review launch readiness across product, design, and engineering.",
			Trigger:     "scheduled",
			IsActive:    true,
			CreatedAt:   at(-47 * time.Hour),
			UpdatedAt:   at(-19 * time.Hour),
		},
	}
	for _, workflow := range workflows {
		DB.FirstOrCreate(&workflow, domain.WorkflowDefinition{ID: workflow.ID})
	}

	workflowRuns := []domain.WorkflowRun{
		{
			ID:          "run-1",
			WorkflowID:  "workflow-1",
			StartedBy:   "user-1",
			Status:      "completed",
			Input:       `{"channel_id":"ch-1"}`,
			Summary:     "Collected standup updates from the product channel.",
			StartedAt:   at(-2 * time.Hour),
			CompletedAt: atPtr(-2*time.Hour + 5*time.Minute),
			CreatedAt:   at(-2 * time.Hour),
			UpdatedAt:   at(-2*time.Hour + 5*time.Minute),
		},
	}
	for _, run := range workflowRuns {
		DB.FirstOrCreate(&run, domain.WorkflowRun{ID: run.ID})
	}

	workflowRunSteps := []domain.WorkflowRunStep{
		{
			WorkflowRunID: "run-1",
			Name:          "Collect updates",
			Status:        "completed",
			DurationMS:    210,
			Detail:        "Read the latest standup messages from #general.",
			CreatedAt:     at(-2*time.Hour + time.Minute),
		},
		{
			WorkflowRunID: "run-1",
			Name:          "Summarize blockers",
			Status:        "completed",
			DurationMS:    340,
			Detail:        "Grouped progress and blockers into a concise update.",
			CreatedAt:     at(-2*time.Hour + 2*time.Minute),
		},
	}
	for _, step := range workflowRunSteps {
		DB.FirstOrCreate(&step, domain.WorkflowRunStep{
			WorkflowRunID: step.WorkflowRunID,
			Name:          step.Name,
		})
	}

	workflowRunLogs := []domain.WorkflowRunLog{
		{
			WorkflowRunID: "run-1",
			Level:         "info",
			Message:       "Workflow run accepted by Relay automation engine.",
			Metadata:      `{"source":"seed","stage":"queued"}`,
			CreatedAt:     at(-2*time.Hour + 5*time.Second),
		},
		{
			WorkflowRunID: "run-1",
			Level:         "info",
			Message:       "Generated summary and closed run successfully.",
			Metadata:      `{"source":"seed","stage":"completed"}`,
			CreatedAt:     at(-2*time.Hour + 5*time.Minute),
		},
	}
	for _, log := range workflowRunLogs {
		DB.FirstOrCreate(&log, domain.WorkflowRunLog{
			WorkflowRunID: log.WorkflowRunID,
			Message:       log.Message,
		})
	}

	tools := []domain.ToolDefinition{
		{
			ID:          "tool-1",
			Name:        "Web Search",
			Key:         "web-search",
			Category:    "research",
			Description: "Search the web for current information and citations.",
			Icon:        "search",
			IsEnabled:   true,
			CreatedAt:   at(-46 * time.Hour),
			UpdatedAt:   at(-18 * time.Hour),
		},
		{
			ID:          "tool-2",
			Name:        "Canvas Generator",
			Key:         "canvas-generate",
			Category:    "creation",
			Description: "Turn prompts into artifacts and editable canvas documents.",
			Icon:        "file-text",
			IsEnabled:   true,
			CreatedAt:   at(-45 * time.Hour),
			UpdatedAt:   at(-17 * time.Hour),
		},
	}
	for _, tool := range tools {
		DB.FirstOrCreate(&tool, domain.ToolDefinition{ID: tool.ID})
	}

	toolRuns := []domain.ToolRun{
		{
			ID:          "toolrun-1",
			ToolID:      "tool-2",
			TriggeredBy: "user-1",
			Status:      "success",
			Input:       `{"channel_id":"ch-5","prompt":"Summarize launch blockers"}`,
			Summary:     "Generated a draft launch readiness canvas from the latest discussion.",
			StartedAt:   at(-110 * time.Minute),
			CompletedAt: atPtr(-109 * time.Minute),
			CreatedAt:   at(-110 * time.Minute),
			UpdatedAt:   at(-109 * time.Minute),
		},
	}
	for _, run := range toolRuns {
		DB.FirstOrCreate(&run, domain.ToolRun{ID: run.ID})
	}

	toolRunLogs := []domain.ToolRunLog{
		{
			ToolRunID: "toolrun-1",
			Level:     "info",
			Message:   "Canvas Generator accepted the request from Home.",
			CreatedAt: at(-110 * time.Minute),
		},
		{
			ToolRunID: "toolrun-1",
			Level:     "info",
			Message:   "Canvas draft rendered and attached to #ai-lab.",
			CreatedAt: at(-109 * time.Minute),
		},
	}
	for _, log := range toolRunLogs {
		DB.FirstOrCreate(&log, domain.ToolRunLog{
			ToolRunID: log.ToolRunID,
			Message:   log.Message,
		})
	}

	workspaceLists := []domain.WorkspaceList{
		{
			ID:          "list-1",
			WorkspaceID: "ws-1",
			ChannelID:   "ch-5",
			Title:       "Launch Checklist",
			Description: "Shared readiness checks before the next Relay release.",
			CreatedBy:   "user-1",
			CreatedAt:   at(-150 * time.Minute),
			UpdatedAt:   at(-108 * time.Minute),
		},
	}
	for _, list := range workspaceLists {
		DB.FirstOrCreate(&list, domain.WorkspaceList{ID: list.ID})
	}

	workspaceListItems := []domain.WorkspaceListItem{
		{
			ListID:      "list-1",
			Content:     "Finalize release notes",
			Position:    1,
			IsCompleted: true,
			AssignedTo:  "user-1",
			CompletedAt: atPtr(-120 * time.Minute),
			CreatedBy:   "user-1",
			CreatedAt:   at(-149 * time.Minute),
			UpdatedAt:   at(-120 * time.Minute),
		},
		{
			ListID:      "list-1",
			Content:     "QA final smoke run",
			Position:    2,
			IsCompleted: false,
			AssignedTo:  "user-3",
			DueAt:       atPtr(6 * time.Hour),
			CreatedBy:   "user-1",
			CreatedAt:   at(-148 * time.Minute),
			UpdatedAt:   at(-108 * time.Minute),
		},
	}
	for _, item := range workspaceListItems {
		DB.FirstOrCreate(&item, domain.WorkspaceListItem{
			ListID:   item.ListID,
			Position: item.Position,
		})
	}

	notificationPreference := domain.NotificationPreference{
		UserID:          "user-1",
		InboxEnabled:    true,
		MentionsEnabled: true,
		DMEnabled:       true,
		MuteAll:         false,
	}
	DB.FirstOrCreate(&notificationPreference, domain.NotificationPreference{UserID: notificationPreference.UserID})

	muteRules := []domain.NotificationMuteRule{
		{
			UserID:    "user-1",
			ScopeType: "channel",
			ScopeID:   "ch-2",
			IsMuted:   true,
		},
	}
	for _, rule := range muteRules {
		DB.FirstOrCreate(&rule, domain.NotificationMuteRule{
			UserID:    rule.UserID,
			ScopeType: rule.ScopeType,
			ScopeID:   rule.ScopeID,
		})
	}

	// 5. Messages (from mock-data.ts)
	type MockMessage struct {
		ID          string
		Content     string
		SenderID    string
		ChannelID   string
		CreatedAt   time.Time
		Reactions   interface{}
		Attachments interface{}
	}

	mockMessages := []MockMessage{
		{
			ID:        "msg-1",
			Content:   "Welcome to the team everyone! Glad to have you here.",
			SenderID:  "user-1",
			ChannelID: "ch-1",
			CreatedAt: at(-7 * time.Hour),
			Reactions: []map[string]interface{}{{"emoji": "👋", "count": 3, "userIds": []string{"user-2", "user-3", "user-4"}}},
		},
		{
			ID:        "msg-2",
			Content:   "Thanks @Nikko Fu! Excited to get started on the AI integration.",
			SenderID:  "user-3",
			ChannelID: "ch-1",
			CreatedAt: at(-6*time.Hour - 30*time.Minute),
		},
		{
			ID:          "msg-3",
			Content:     "I've drafted the implementation plan for Relay Agent Workspace. Take a look: https://github.com/nikkofu/relay-agent-workspace",
			SenderID:    "user-1",
			ChannelID:   "ch-5",
			CreatedAt:   at(-5 * time.Hour),
			Attachments: []map[string]interface{}{{"id": "att-1", "type": "link", "url": "https://github.com/nikkofu/relay-agent-workspace", "name": "Relay Agent Workspace Repository"}},
		},
		{
			ID:        "msg-4",
			Content:   "The AI components look promising. Can we add support for streaming responses?",
			SenderID:  "user-4",
			ChannelID: "ch-5",
			CreatedAt: at(-4*time.Hour - 30*time.Minute),
		},
		{
			ID:        "msg-5",
			Content:   "Absolutely. Vercel AI SDK provides great primitives for that. I'll include it in Phase 3.",
			SenderID:  "user-1",
			ChannelID: "ch-5",
			CreatedAt: at(-4 * time.Hour),
			Reactions: []map[string]interface{}{{"emoji": "✅", "count": 2, "userIds": []string{"user-4", "user-2"}}},
		},
		{
			ID:        "msg-6",
			Content:   "Here is a quick summary of today's progress:\n- Setup project with Next.js 15\n- Initialized shadcn/ui components\n- Defined core types and mock data",
			SenderID:  "user-2",
			ChannelID: "ch-5",
			CreatedAt: at(-3*time.Hour - 30*time.Minute),
		},
	}

	for _, m := range mockMessages {
		metadata := make(map[string]interface{})
		if m.Reactions != nil {
			metadata["reactions"] = m.Reactions
		}
		if m.Attachments != nil {
			metadata["attachments"] = m.Attachments
		}

		metadataJSON, _ := json.Marshal(metadata)

		msg := domain.Message{
			ID:         m.ID,
			ChannelID:  m.ChannelID,
			UserID:     m.SenderID,
			Content:    m.Content,
			CreatedAt:  m.CreatedAt,
			Metadata:   string(metadataJSON),
			ThreadID:   "",
			ReplyCount: 0,
			IsPinned:   m.ID == "msg-3",
		}
		DB.FirstOrCreate(&msg, domain.Message{ID: msg.ID})

		if reactions, ok := m.Reactions.([]map[string]interface{}); ok {
			for _, reaction := range reactions {
				emoji, _ := reaction["emoji"].(string)
				userIDs, _ := reaction["userIds"].([]string)
				if userIDs == nil {
					if rawUserIDs, ok := reaction["userIds"].([]interface{}); ok {
						for _, raw := range rawUserIDs {
							if userID, ok := raw.(string); ok {
								userIDs = append(userIDs, userID)
							}
						}
					}
				}
				for _, userID := range userIDs {
					DB.FirstOrCreate(&domain.MessageReaction{}, domain.MessageReaction{
						MessageID: msg.ID,
						UserID:    userID,
						Emoji:     emoji,
					})
				}
			}
		}
	}

	artifacts := []domain.Artifact{
		{
			ID:        "artifact-1",
			ChannelID: "ch-5",
			Title:     "Relay Launch Roadmap",
			Version:   1,
			Type:      "document",
			Status:    "live",
			Content:   "This draft captures the launch sequencing for Relay Agent Workspace, including API validation, UI integration, and release readiness.",
			Source:    "ai",
			Provider:  "gemini",
			Model:     "gemini-3-flash-preview",
			CreatedBy: "user-2",
			UpdatedBy: "user-2",
			CreatedAt: at(-3 * time.Hour),
			UpdatedAt: at(-3*time.Hour + 5*time.Minute),
		},
	}
	for _, artifact := range artifacts {
		DB.FirstOrCreate(&artifact, domain.Artifact{ID: artifact.ID})
		DB.Model(&domain.Artifact{}).Where("id = ? AND version = 0", artifact.ID).Update("version", 1)
		DB.FirstOrCreate(&domain.ArtifactVersion{}, domain.ArtifactVersion{
			ArtifactID: artifact.ID,
			Version:    1,
		})
		DB.Model(&domain.ArtifactVersion{}).Where("artifact_id = ? AND version = 1", artifact.ID).Updates(map[string]any{
			"title":      artifact.Title,
			"type":       artifact.Type,
			"status":     artifact.Status,
			"content":    artifact.Content,
			"source":     artifact.Source,
			"provider":   artifact.Provider,
			"model":      artifact.Model,
			"updated_by": artifact.UpdatedBy,
			"created_at": artifact.UpdatedAt,
		})
	}

	threadReply := domain.Message{
		ID:        "msg-7",
		ChannelID: "ch-5",
		UserID:    "user-4",
		Content:   "Yes, thread support will make the channel list much cleaner.",
		ThreadID:  "msg-3",
		CreatedAt: at(-3*time.Hour - 45*time.Minute),
		Metadata:  "{}",
	}
	DB.FirstOrCreate(&threadReply, domain.Message{ID: threadReply.ID})

	lastReplyAt := threadReply.CreatedAt
	DB.Model(&domain.Message{}).Where("id = ?", "msg-3").Updates(map[string]any{
		"reply_count":   1,
		"last_reply_at": &lastReplyAt,
	})

	drafts := []domain.Draft{
		{
			UserID:    "user-1",
			Scope:     "channel:ch-5",
			Content:   "Need to summarize the AI provider matrix before posting the next update.",
			CreatedAt: at(-90 * time.Minute),
			UpdatedAt: at(-90 * time.Minute),
		},
	}
	for _, draft := range drafts {
		DB.FirstOrCreate(&draft, domain.Draft{
			UserID: draft.UserID,
			Scope:  draft.Scope,
		})
	}

	// 6. DM conversations
	dmConversation := domain.DMConversation{
		ID:        "dm-1",
		CreatedAt: at(-2 * time.Hour),
	}
	DB.FirstOrCreate(&dmConversation, domain.DMConversation{ID: dmConversation.ID})

	dmMembers := []domain.DMMember{
		{DMConversationID: dmConversation.ID, UserID: "user-1"},
		{DMConversationID: dmConversation.ID, UserID: "user-2"},
	}
	for _, member := range dmMembers {
		DB.FirstOrCreate(&member, domain.DMMember{
			DMConversationID: member.DMConversationID,
			UserID:           member.UserID,
		})
	}

	dmMessages := []domain.DMMessage{
		{
			ID:               "dm-msg-1",
			DMConversationID: dmConversation.ID,
			UserID:           "user-2",
			Content:          "I can help summarize the latest delivery status when you are ready.",
			CreatedAt:        at(-100 * time.Minute),
		},
		{
			ID:               "dm-msg-2",
			DMConversationID: dmConversation.ID,
			UserID:           "user-1",
			Content:          "Perfect. Let's keep the launch checklist tight today.",
			CreatedAt:        at(-95 * time.Minute),
		},
	}
	for _, message := range dmMessages {
		DB.FirstOrCreate(&message, domain.DMMessage{ID: message.ID})
	}
}
