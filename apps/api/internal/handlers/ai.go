package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/llm"
)

type AIStreamer interface {
	Stream(ctx context.Context, req llm.Request) (*llm.StreamSession, error)
}

var AIGateway AIStreamer

func SetAIGateway(gateway AIStreamer) {
	AIGateway = gateway
}

func ExecuteAI(c *gin.Context) {
	if AIGateway == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ai gateway is not configured"})
		return
	}

	var input llm.Request
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Prompt == "" || input.ChannelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt and channel_id are required"})
		return
	}

	session, err := AIGateway.Stream(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	writer := c.Writer
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	flusher, ok := writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	writeSSE(writer, "start", map[string]any{
		"provider": session.Provider,
		"model":    session.Model,
	})
	flusher.Flush()

	for {
		select {
		case event, ok := <-session.Events:
			if !ok {
				writeSSE(writer, "done", map[string]any{
					"provider": session.Provider,
					"model":    session.Model,
				})
				flusher.Flush()
				return
			}
			writeSSE(writer, event.Type, map[string]any{"text": event.Text})
			flusher.Flush()
		case err, ok := <-session.Errors:
			if !ok {
				session.Errors = nil
				continue
			}
			if err != nil {
				writeSSE(writer, "error", map[string]any{"message": err.Error()})
				flusher.Flush()
			}
			return
		case <-c.Request.Context().Done():
			return
		}
	}
}

func writeSSE(writer http.ResponseWriter, event string, payload any) {
	data, _ := json.Marshal(payload)
	_, _ = writer.Write([]byte("event: " + event + "\n"))
	_, _ = writer.Write([]byte("data: " + string(data) + "\n\n"))
}
