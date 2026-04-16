"use client"

import { AgentStateCard } from "./agent-state-card"

const MOCK_AGENTS = [
  { name: "Gemini", skill: "executing-plans", task: "WebSocket Integration", progress: 60, status: 'active' as const },
  { name: "Codex", skill: "watching-file", task: "AGENT-COLLAB.md Monitoring", progress: 100, status: 'thinking' as const },
  { name: "Claude", skill: "idle", task: "-", progress: 0, status: 'idle' as const },
]

export function AgentCollabDashboard() {
  return (
    <div className="p-4 bg-muted/20 border-b flex gap-3 overflow-x-auto no-scrollbar scroll-smooth shrink-0">
      {MOCK_AGENTS.map(agent => (
        <AgentStateCard key={agent.name} agent={agent} />
      ))}
    </div>
  )
}
