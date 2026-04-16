package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/handlers"
)

func main() {
	if err := db.InitDB(); err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	db.SeedData()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/me", handlers.GetMe)
		v1.GET("/orgs", handlers.GetOrganizations)
		v1.GET("/orgs/:id/teams", handlers.GetTeams)
		v1.POST("/orgs/:id/agents", handlers.CreateAgent)
		v1.GET("/workspaces", handlers.GetWorkspaces)
		v1.GET("/channels", handlers.GetChannels)
		v1.GET("/messages", handlers.GetMessages)
		v1.POST("/messages", handlers.CreateMessage)
	}

	r.Run(":8080")
}
