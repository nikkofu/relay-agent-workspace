package db

import (
	"encoding/json"
	"time"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

func SeedData() {
	// 1. Organizations
	org := domain.Organization{ID: "ws-1", Name: "Acme Corp"}
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
			Avatar:         "https://api.dicebear.com/7.x/bottts/svg?seed=ai",
			Status:         "online",
		},
		{
			ID:             "user-3",
			OrganizationID: org.ID,
			Name:           "John Doe",
			Email:          "john@example.com",
			Avatar:         "https://i.pravatar.cc/150?u=john",
			Status:         "away",
		},
		{
			ID:             "user-4",
			OrganizationID: org.ID,
			Name:           "Jane Smith",
			Email:          "jane@example.com",
			Avatar:         "https://i.pravatar.cc/150?u=jane",
			Status:         "offline",
		},
	}

	for _, u := range users {
		DB.FirstOrCreate(&u, domain.User{ID: u.ID})
	}

	// 3. Workspaces (from mock-data.ts)
	workspaces := []domain.Workspace{
		{ID: "ws-1", OrganizationID: org.ID, Name: "Acme Corp"},
		{ID: "ws-2", OrganizationID: org2.ID, Name: "Side Project"},
	}

	for _, w := range workspaces {
		DB.FirstOrCreate(&w, domain.Workspace{ID: w.ID})
	}

	// 4. Channels (from mock-data.ts)
	channels := []domain.Channel{
		{ID: "ch-1", WorkspaceID: "ws-1", Name: "general", Type: "public", Description: "Company-wide announcements and work-based matters", MemberCount: 15, IsStarred: true},
		{ID: "ch-2", WorkspaceID: "ws-1", Name: "random", Type: "public", Description: "Non-work-related banter and water cooler talk", MemberCount: 12},
		{ID: "ch-3", WorkspaceID: "ws-1", Name: "engineering", Type: "public", Description: "Technical discussions and code reviews", MemberCount: 8, UnreadCount: 3},
		{ID: "ch-4", WorkspaceID: "ws-1", Name: "design", Type: "public", Description: "UI/UX design critiques and inspiration", MemberCount: 5},
		{ID: "ch-5", WorkspaceID: "ws-1", Name: "ai-lab", Type: "public", Description: "Exploring the future of AI-native applications", MemberCount: 10, IsStarred: true, UnreadCount: 12},
		{ID: "ch-6", WorkspaceID: "ws-1", Name: "private-deal", Type: "private", MemberCount: 3},
		{ID: "ch-collab", WorkspaceID: "ws-1", Name: "agent-collab", Type: "public", Description: "Real-time Agent collaboration and synchronization hub", MemberCount: 3, IsStarred: true},
	}

	for _, ch := range channels {
		DB.FirstOrCreate(&ch, domain.Channel{ID: ch.ID})
	}

	// 5. Messages (from mock-data.ts)
	type MockMessage struct {
		ID          string
		Content     string
		SenderID    string
		ChannelID   string
		CreatedAt   string
		Reactions   interface{}
		Attachments interface{}
	}

	mockMessages := []MockMessage{
		{
			ID:        "msg-1",
			Content:   "Welcome to the team everyone! Glad to have you here.",
			SenderID:  "user-1",
			ChannelID: "ch-1",
			CreatedAt: "2026-04-14T10:00:00Z",
			Reactions: []map[string]interface{}{{"emoji": "👋", "count": 3, "userIds": []string{"user-2", "user-3", "user-4"}}},
		},
		{
			ID:        "msg-2",
			Content:   "Thanks @Nikko Fu! Excited to get started on the AI integration.",
			SenderID:  "user-3",
			ChannelID: "ch-1",
			CreatedAt: "2026-04-14T11:30:00Z",
		},
		{
			ID:          "msg-3",
			Content:     "I've drafted the implementation plan for Relay Agent Workspace. Take a look: https://github.com/nikkofu/relay-agent-workspace",
			SenderID:    "user-1",
			ChannelID:   "ch-5",
			CreatedAt:   "2026-04-15T08:00:00Z",
			Attachments: []map[string]interface{}{{"id": "att-1", "type": "link", "url": "https://github.com/nikkofu/relay-agent-workspace", "name": "Relay Agent Workspace Repository"}},
		},
		{
			ID:        "msg-4",
			Content:   "The AI components look promising. Can we add support for streaming responses?",
			SenderID:  "user-4",
			ChannelID: "ch-5",
			CreatedAt: "2026-04-15T08:30:00Z",
		},
		{
			ID:        "msg-5",
			Content:   "Absolutely. Vercel AI SDK provides great primitives for that. I'll include it in Phase 3.",
			SenderID:  "user-1",
			ChannelID: "ch-5",
			CreatedAt: "2026-04-15T09:00:00Z",
			Reactions: []map[string]interface{}{{"emoji": "✅", "count": 2, "userIds": []string{"user-4", "user-2"}}},
		},
		{
			ID:        "msg-6",
			Content:   "Here is a quick summary of today's progress:\n- Setup project with Next.js 15\n- Initialized shadcn/ui components\n- Defined core types and mock data",
			SenderID:  "user-2",
			ChannelID: "ch-5",
			CreatedAt: "2026-04-15T09:30:00Z",
		},
	}

	for _, m := range mockMessages {
		t, _ := time.Parse(time.RFC3339, m.CreatedAt)

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
			CreatedAt:  t,
			Metadata:   string(metadataJSON),
			ThreadID:   "",
			ReplyCount: 0,
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

	threadReply := domain.Message{
		ID:        "msg-7",
		ChannelID: "ch-5",
		UserID:    "user-4",
		Content:   "Yes, thread support will make the channel list much cleaner.",
		ThreadID:  "msg-3",
		CreatedAt: time.Date(2026, 4, 15, 9, 15, 0, 0, time.UTC),
		Metadata:  "{}",
	}
	DB.FirstOrCreate(&threadReply, domain.Message{ID: threadReply.ID})

	lastReplyAt := threadReply.CreatedAt
	DB.Model(&domain.Message{}).Where("id = ?", "msg-3").Updates(map[string]any{
		"reply_count":   1,
		"last_reply_at": &lastReplyAt,
	})

	// 6. DM conversations
	dmConversation := domain.DMConversation{
		ID:        "dm-1",
		CreatedAt: time.Date(2026, 4, 15, 7, 45, 0, 0, time.UTC),
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
			CreatedAt:        time.Date(2026, 4, 15, 7, 50, 0, 0, time.UTC),
		},
		{
			ID:               "dm-msg-2",
			DMConversationID: dmConversation.ID,
			UserID:           "user-1",
			Content:          "Perfect. Let's keep the launch checklist tight today.",
			CreatedAt:        time.Date(2026, 4, 15, 7, 52, 0, 0, time.UTC),
		},
	}
	for _, message := range dmMessages {
		DB.FirstOrCreate(&message, domain.DMMessage{ID: message.ID})
	}
}
