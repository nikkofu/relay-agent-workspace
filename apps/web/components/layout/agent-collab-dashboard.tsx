"use client"

import { AgentStateCard } from "./agent-state-card"
import { useCollabStore } from "@/stores/collab-store"

export function AgentCollabDashboard() {
  const { agents } = useCollabStore()

  if (agents.length === 0) {
    return (
      <div className="p-4 bg-muted/20 border-b flex gap-3 overflow-x-auto no-scrollbar scroll-smooth shrink-0 items-center justify-center text-xs text-muted-foreground italic">
        Waiting for Agent synchronization...
      </div>
    )
  }

  return (
    <div className="p-4 bg-muted/20 border-b flex gap-3 overflow-x-auto no-scrollbar scroll-smooth shrink-0">
      {agents.map(agent => (
        <AgentStateCard key={agent.name} agent={agent} />
      ))}
    </div>
  )
}
