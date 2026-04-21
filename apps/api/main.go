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
	r.Static("/uploads", "./uploads")

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
		v1.GET("/home", handlers.GetHome)
		v1.GET("/me", handlers.GetMe)
		v1.PATCH("/me/settings", handlers.PatchMeSettings)
		v1.GET("/users", handlers.GetUsers)
		v1.GET("/users/:id", handlers.GetUserProfile)
		v1.PATCH("/users/:id", handlers.PatchUserProfile)
		v1.PATCH("/users/:id/status", handlers.PatchUserStatus)
		v1.GET("/user-groups/mentions", handlers.SearchUserGroupMentions)
		v1.GET("/user-groups", handlers.GetUserGroups)
		v1.POST("/user-groups", handlers.CreateUserGroup)
		v1.GET("/user-groups/:id/members", handlers.GetUserGroupMembers)
		v1.POST("/user-groups/:id/members", handlers.AddUserGroupMember)
		v1.DELETE("/user-groups/:id/members/:userId", handlers.RemoveUserGroupMember)
		v1.GET("/user-groups/:id", handlers.GetUserGroup)
		v1.PATCH("/user-groups/:id", handlers.UpdateUserGroup)
		v1.DELETE("/user-groups/:id", handlers.DeleteUserGroup)
		v1.GET("/workflows", handlers.GetWorkflows)
		v1.GET("/workflows/runs", handlers.GetWorkflowRuns)
		v1.GET("/workflows/runs/:id/logs", handlers.GetWorkflowRunLogs)
		v1.GET("/workflows/runs/:id", handlers.GetWorkflowRun)
		v1.DELETE("/workflows/runs/:id", handlers.DeleteWorkflowRun)
		v1.POST("/workflows/runs/:id/cancel", handlers.CancelWorkflowRun)
		v1.POST("/workflows/runs/:id/retry", handlers.RetryWorkflowRun)
		v1.POST("/workflows/:id/runs", handlers.CreateWorkflowRun)
		v1.GET("/tools", handlers.GetTools)
		v1.GET("/tools/runs", handlers.GetToolRuns)
		v1.GET("/tools/runs/:id", handlers.GetToolRun)
		v1.POST("/tools/:id/execute", handlers.ExecuteTool)
		v1.GET("/agent-collab/snapshot", handlers.GetAgentCollabSnapshot)
		v1.GET("/agent-collab/members", handlers.GetAgentCollabMembers)
		v1.POST("/agent-collab/comm-log", handlers.CreateAgentCollabCommLog)
		v1.GET("/orgs", handlers.GetOrganizations)
		v1.GET("/orgs/:id/teams", handlers.GetTeams)
		v1.POST("/orgs/:id/agents", handlers.CreateAgent)
		v1.GET("/workspaces", handlers.GetWorkspaces)
		v1.GET("/workspaces/:id/invites", handlers.GetWorkspaceInvites)
		v1.POST("/workspaces/:id/invites", handlers.CreateWorkspaceInvite)
		v1.GET("/channels", handlers.GetChannels)
		v1.POST("/channels", handlers.CreateChannel)
		v1.GET("/channels/:id/members", handlers.GetChannelMembers)
		v1.POST("/channels/:id/members", handlers.AddChannelMember)
		v1.DELETE("/channels/:id/members/:userId", handlers.RemoveChannelMember)
		v1.GET("/channels/:id/preferences", handlers.GetChannelPreferences)
		v1.PATCH("/channels/:id/preferences", handlers.PatchChannelPreferences)
		v1.POST("/channels/:id/leave", handlers.LeaveChannel)
		v1.PATCH("/channels/:id", handlers.UpdateChannel)
		v1.POST("/channels/:id/star", handlers.ToggleChannelStar)
		v1.GET("/channels/:id/summary", handlers.GetChannelSummary)
		v1.POST("/channels/:id/summary", handlers.GenerateChannelSummary)
		v1.GET("/dms", handlers.GetDMConversations)
		v1.POST("/dms", handlers.CreateOrOpenDMConversation)
		v1.GET("/dms/:id/messages", handlers.GetDMMessages)
		v1.POST("/dms/:id/messages", handlers.CreateDMMessage)
		v1.GET("/activity", handlers.GetActivity)
		v1.GET("/inbox", handlers.GetInbox)
		v1.GET("/mentions", handlers.GetMentions)
		v1.POST("/notifications/read", handlers.MarkNotificationsRead)
		v1.GET("/notifications/preferences", handlers.GetNotificationPreferences)
		v1.PATCH("/notifications/preferences", handlers.PatchNotificationPreferences)
		v1.GET("/later", handlers.GetLater)
		v1.GET("/starred", handlers.GetStarredChannels)
		v1.GET("/pins", handlers.GetPins)
		v1.GET("/presence", handlers.GetPresence)
		v1.POST("/presence", handlers.UpdatePresence)
		v1.POST("/presence/heartbeat", handlers.HeartbeatPresence)
		v1.POST("/typing", handlers.UpdateTyping)
		v1.GET("/drafts", handlers.GetDrafts)
		v1.PUT("/drafts/:scope", handlers.PutDraft)
		v1.DELETE("/drafts/:scope", handlers.DeleteDraft)
		v1.GET("/search", handlers.SearchWorkspace)
		v1.GET("/search/files", handlers.SearchFiles)
		v1.GET("/search/intelligent", handlers.IntelligentSearch)
		v1.GET("/search/suggestions", handlers.SearchSuggestions)
		v1.GET("/lists", handlers.GetWorkspaceLists)
		v1.POST("/lists", handlers.CreateWorkspaceList)
		v1.GET("/lists/:id", handlers.GetWorkspaceList)
		v1.PATCH("/lists/:id", handlers.UpdateWorkspaceList)
		v1.DELETE("/lists/:id", handlers.DeleteWorkspaceList)
		v1.POST("/lists/:id/items", handlers.CreateWorkspaceListItem)
		v1.PATCH("/lists/:id/items/:itemId", handlers.UpdateWorkspaceListItem)
		v1.DELETE("/lists/:id/items/:itemId", handlers.DeleteWorkspaceListItem)
		v1.GET("/artifacts", handlers.GetArtifacts)
		v1.GET("/artifacts/templates", handlers.GetArtifactTemplates)
		v1.POST("/artifacts/from-template", handlers.CreateArtifactFromTemplate)
		v1.POST("/artifacts", handlers.CreateArtifact)
		v1.GET("/artifacts/:id", handlers.GetArtifact)
		v1.GET("/artifacts/:id/versions", handlers.GetArtifactVersions)
		v1.GET("/artifacts/:id/versions/:version", handlers.GetArtifactVersion)
		v1.GET("/artifacts/:id/diff/:from/:to", handlers.GetArtifactDiff)
		v1.GET("/artifacts/:id/references", handlers.GetArtifactReferences)
		v1.POST("/artifacts/:id/duplicate", handlers.DuplicateArtifact)
		v1.POST("/artifacts/:id/restore/:version", handlers.RestoreArtifactVersion)
		v1.PATCH("/artifacts/:id", handlers.UpdateArtifact)
		v1.GET("/files", handlers.ListFiles)
		v1.GET("/files/starred", handlers.GetStarredFiles)
		v1.GET("/files/archive", handlers.GetArchivedFiles)
		v1.POST("/files/upload", handlers.UploadFile)
		v1.PATCH("/files/:id/archive", handlers.ToggleFileArchive)
		v1.POST("/files/:id/star", handlers.ToggleFileStar)
		v1.PATCH("/files/:id/retention", handlers.UpdateFileRetention)
		v1.PATCH("/files/:id/knowledge", handlers.UpdateFileKnowledge)
		v1.GET("/files/:id/audit", handlers.GetFileAudit)
		v1.GET("/files/:id/comments", handlers.GetFileComments)
		v1.POST("/files/:id/comments", handlers.CreateFileComment)
		v1.GET("/files/:id/shares", handlers.GetFileShares)
		v1.POST("/files/:id/share", handlers.ShareFile)
		v1.GET("/files/:id/extraction", handlers.GetFileExtraction)
		v1.POST("/files/:id/extraction/rebuild", handlers.RebuildFileExtraction)
		v1.GET("/files/:id/preview", handlers.GetFilePreview)
		v1.GET("/files/:id/chunks", handlers.GetFileExtractionChunks)
		v1.GET("/files/:id/citations", handlers.GetFileCitations)
		v1.DELETE("/files/:id", handlers.DeleteFile)
		v1.GET("/files/:id", handlers.GetFile)
		v1.GET("/files/:id/extracted-content", handlers.GetFileExtractedContent)
		v1.GET("/files/:id/content", handlers.GetFileContent)
		v1.GET("/messages", handlers.GetMessages)
		v1.GET("/messages/:id/thread", handlers.GetMessageThread)
		v1.GET("/messages/:id/files", handlers.GetMessageFiles)
		v1.GET("/messages/:id/summary", handlers.GetThreadSummary)
		v1.POST("/messages/:id/summary", handlers.GenerateThreadSummary)
		v1.POST("/messages", handlers.CreateMessage)
		v1.POST("/messages/:id/reactions", handlers.ToggleReaction)
		v1.DELETE("/messages/:id", handlers.DeleteMessage)
		v1.POST("/messages/:id/pin", handlers.TogglePinMessage)
		v1.POST("/messages/:id/later", handlers.ToggleSaveForLater)
		v1.POST("/messages/:id/unread", handlers.MarkMessageUnread)
		v1.GET("/realtime", handlers.HandleRealtime)
		v1.GET("/ai/config", handlers.GetAIConfig)
		v1.GET("/ai/conversations", handlers.GetAIConversations)
		v1.GET("/ai/conversations/:id", handlers.GetAIConversation)
		v1.POST("/ai/execute", handlers.ExecuteAI)
		v1.POST("/ai/canvas/generate", handlers.GenerateCanvasArtifact)
		v1.POST("/ai/feedback", handlers.SubmitAIFeedback)
	}

	r.Run(":8080")
}
