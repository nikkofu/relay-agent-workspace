package main

import (
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/agentcollab"
	"github.com/nikkofu/relay-agent-workspace/api/internal/config"
	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/handlers"
	"github.com/nikkofu/relay-agent-workspace/api/internal/llm"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

func main() {
	if err := db.InitDB(); err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	db.SeedData()

	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	hub := realtime.NewHub()
	go hub.Run()
	handlers.SetRealtimeHub(hub)

	llmConfig, err := config.LoadLLMConfig(filepath.Join("config"))
	if err != nil {
		log.Printf("llm config load failed: %v", err)
	} else {
		handlers.SetAIConfig(llmConfig)
		handlers.SetAIGateway(llm.NewGateway(llmConfig, map[string]llm.Provider{
			"openai":            llm.NewOpenAICompatibleProvider(),
			"openai-compatible": llm.NewOpenAICompatibleProvider(),
			"openrouter":        llm.NewOpenRouterProvider(),
			"gemini":            llm.NewGeminiProvider(),
		}))
	}

	collabService := agentcollab.NewService(agentcollab.DefaultPath(), hub)
	if err := collabService.Start(); err != nil {
		log.Printf("agent collab watcher disabled: %v", err)
	} else {
		defer func() {
			if err := collabService.Close(); err != nil {
				log.Printf("failed to close agent collab watcher: %v", err)
			}
		}()
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/me", handlers.GetMe)
		v1.PATCH("/me/settings", handlers.PatchMeSettings)
		v1.GET("/users", handlers.GetUsers)
		v1.GET("/agent-collab/snapshot", handlers.GetAgentCollabSnapshot)
		v1.GET("/orgs", handlers.GetOrganizations)
		v1.GET("/orgs/:id/teams", handlers.GetTeams)
		v1.POST("/orgs/:id/agents", handlers.CreateAgent)
		v1.GET("/workspaces", handlers.GetWorkspaces)
		v1.GET("/channels", handlers.GetChannels)
		v1.GET("/dms", handlers.GetDMConversations)
		v1.POST("/dms", handlers.CreateOrOpenDMConversation)
		v1.GET("/dms/:id/messages", handlers.GetDMMessages)
		v1.POST("/dms/:id/messages", handlers.CreateDMMessage)
		v1.GET("/messages", handlers.GetMessages)
		v1.GET("/messages/:id/thread", handlers.GetMessageThread)
		v1.POST("/messages", handlers.CreateMessage)
		v1.POST("/messages/:id/reactions", handlers.ToggleReaction)
		v1.DELETE("/messages/:id", handlers.DeleteMessage)
		v1.POST("/messages/:id/pin", handlers.TogglePinMessage)
		v1.POST("/messages/:id/later", handlers.ToggleSaveForLater)
		v1.POST("/messages/:id/unread", handlers.MarkMessageUnread)
		v1.GET("/realtime", handlers.HandleRealtime)
		v1.GET("/ai/config", handlers.GetAIConfig)
		v1.POST("/ai/execute", handlers.ExecuteAI)
		v1.POST("/ai/feedback", handlers.SubmitAIFeedback)
	}

	r.Run(":8080")
}
