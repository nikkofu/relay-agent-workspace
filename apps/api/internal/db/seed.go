package db

import "github.com/nikkofu/relay-agent-workspace/api/internal/domain"

func SeedData() {
	org := domain.Organization{
		ID:   "org_1",
		Name: "Relay Labs",
	}
	DB.FirstOrCreate(&org, domain.Organization{ID: org.ID})

	team := domain.Team{
		ID:             "team_1",
		OrganizationID: org.ID,
		Name:           "Product",
	}
	DB.FirstOrCreate(&team, domain.Team{ID: team.ID})

	team2 := domain.Team{
		ID:             "team_2",
		OrganizationID: org.ID,
		Name:           "Engineering",
	}
	DB.FirstOrCreate(&team2, domain.Team{ID: team2.ID})

	user := domain.User{
		ID:             "u_1",
		OrganizationID: org.ID,
		Name:           "Nikko Fu",
		Email:          "nikko@relay.ai",
		Avatar:         "https://github.com/nikkofu.png",
		Status:         "online",
	}
	DB.FirstOrCreate(&user, domain.User{ID: user.ID})

	agent := domain.Agent{
		ID:             "agent_1",
		OrganizationID: org.ID,
		Name:           "Relay Assistant",
		Type:           "general",
		OwnerID:        user.ID,
	}
	DB.FirstOrCreate(&agent, domain.Agent{ID: agent.ID})

	agent2 := domain.Agent{
		ID:             "agent_2",
		OrganizationID: org.ID,
		Name:           "Relay Researcher",
		Type:           "research",
		OwnerID:        user.ID,
	}
	DB.FirstOrCreate(&agent2, domain.Agent{ID: agent2.ID})

	workspace := domain.Workspace{
		ID:             "ws_1",
		OrganizationID: org.ID,
		Name:           "Relay Workspace",
	}
	DB.FirstOrCreate(&workspace, domain.Workspace{ID: workspace.ID})

	channel := domain.Channel{
		ID:          "ch_1",
		WorkspaceID: workspace.ID,
		Name:        "general",
		Type:        "public",
	}
	DB.FirstOrCreate(&channel, domain.Channel{ID: channel.ID})

	message := domain.Message{
		ID:        "msg_1",
		ChannelID: channel.ID,
		UserID:    user.ID,
		Content:   "Relay backend seed data is ready.",
	}
	DB.FirstOrCreate(&message, domain.Message{ID: message.ID})
}
