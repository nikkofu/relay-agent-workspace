package handlers

import (
	"log"
	"time"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/knowledge"
)

func StartKnowledgeDigestScheduler() {
	go runKnowledgeDigestScheduler()
}

func runKnowledgeDigestScheduler() {
	processKnowledgeDigestSchedules(time.Now().UTC())

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for tickAt := range ticker.C {
		processKnowledgeDigestSchedules(tickAt.UTC())
	}
}

func processKnowledgeDigestSchedules(now time.Time) {
	if db.DB == nil {
		return
	}

	published, err := knowledge.ProcessDigestSchedules(db.DB, now)
	if err != nil {
		log.Printf("knowledge digest scheduler failed: %v", err)
		return
	}

	for _, item := range published {
		if err := broadcastRealtimeEvent("message.created", item.Message, item.Message); err != nil {
			log.Printf("knowledge digest scheduler message broadcast failed: %v", err)
		}
		if err := broadcastKnowledgeDigestPublished(item); err != nil {
			log.Printf("knowledge digest scheduler knowledge broadcast failed: %v", err)
		}
	}
}
