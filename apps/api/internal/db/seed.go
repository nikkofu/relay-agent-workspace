package db

import "github.com/nikkofu/relay-agent-workspace/api/internal/domain"

func SeedData() {
	var count int64
	DB.Model(&domain.User{}).Count(&count)
	if count > 0 {
		return
	}

	org := domain.Organization{
		ID:   "org_1",
		Name: "Relay Labs",
	}
	DB.Create(&org)

	team := domain.Team{
		ID:             "team_1",
		OrganizationID: org.ID,
		Name:           "Product",
	}
	DB.Create(&team)

	user := domain.User{
		ID:             "u_1",
		OrganizationID: org.ID,
		Name:           "Nikko Fu",
		Email:          "nikko@relay.ai",
		Avatar:         "https://github.com/nikkofu.png",
		Status:         "online",
	}
	DB.Create(&user)

	agent := domain.Agent{
		ID:             "agent_1",
		OrganizationID: org.ID,
		Name:           "Relay Assistant",
		Type:           "general",
		OwnerID:        user.ID,
	}
	DB.Create(&agent)

	workspace := domain.Workspace{
		ID:             "ws_1",
		OrganizationID: org.ID,
		Name:           "Relay Workspace",
	}
	DB.Create(&workspace)

	channel := domain.Channel{
		ID:          "ch_1",
		WorkspaceID: workspace.ID,
		Name:        "general",
		Type:        "public",
	}
	DB.Create(&channel)

	message := domain.Message{
		ID:        "msg_1",
		ChannelID: channel.ID,
		UserID:    user.ID,
		Content:   "Relay backend seed data is ready.",
	}
	DB.Create(&message)
}
