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

func StartKnowledgeEntityBriefAutomationScheduler() {
	go runKnowledgeEntityBriefAutomationScheduler()
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

func runKnowledgeEntityBriefAutomationScheduler() {
	processKnowledgeEntityBriefAutomation(time.Now().UTC())

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for tickAt := range ticker.C {
		processKnowledgeEntityBriefAutomation(tickAt.UTC())
	}
}

func processKnowledgeEntityBriefAutomation(now time.Time) {
	if db.DB == nil {
		return
	}

	if _, err := SweepKnowledgeEntityBriefAutomation(now, 15*time.Minute); err != nil {
		log.Printf("knowledge entity brief automation stale sweep failed: %v", err)
	}
	if err := ProcessKnowledgeEntityBriefAutomation(now, 10); err != nil {
		log.Printf("knowledge entity brief automation processing failed: %v", err)
	}
}
