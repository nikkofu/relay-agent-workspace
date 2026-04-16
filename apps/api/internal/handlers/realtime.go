package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

var RealtimeHub *realtime.Hub

func SetRealtimeHub(hub *realtime.Hub) {
	RealtimeHub = hub
}

func HandleRealtime(c *gin.Context) {
	if RealtimeHub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "realtime hub is not configured"})
		return
	}

	if err := RealtimeHub.ServeWS(c.Writer, c.Request); err != nil {
		if c.Writer.Written() {
			log.Printf("websocket upgrade failed after response write: %v", err)
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade websocket"})
	}
}
