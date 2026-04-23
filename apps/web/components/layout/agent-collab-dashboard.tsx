"use client"

import { useEffect } from "react"
import { AgentStateCard } from "./agent-state-card"
import { useCollabStore } from "@/stores/collab-store"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { ComposeActivityPane } from "@/components/knowledge/compose-activity-pane"
import { ComposeActivityDigestStrip } from "@/components/knowledge/compose-activity-digest-strip"
import { KnowledgeAskFeedPane } from "@/components/knowledge/knowledge-ask-feed-pane"
import { AutomationAuditPanel } from "@/components/knowledge/automation-audit-panel"

export function AgentCollabDashboard() {
  const { agents, tasks, fetchSnapshot } = useCollabStore()
  // Phase 63G: workspace id for the co-drafting firehose scope.
  const currentWorkspaceId = useWorkspaceStore(s => s.currentWorkspace?.id)

  useEffect(() => {
    fetchSnapshot()
  }, [fetchSnapshot])

  if (agents.length === 0 && tasks.length === 0) {
    return (
      <div className="p-4 bg-muted/20 border-b flex gap-3 overflow-x-auto no-scrollbar scroll-smooth shrink-0 items-center justify-center text-xs text-muted-foreground italic">
        Waiting for Agent synchronization...
      </div>
    )
  }

  return (
    <div className="p-4 bg-muted/20 border-b shrink-0 flex flex-col gap-4">
      {agents.length > 0 && (
        <div className="flex gap-3 overflow-x-auto no-scrollbar scroll-smooth">
          {agents.map(agent => (
            <AgentStateCard key={agent.name} agent={agent} />
          ))}
        </div>
      )}

      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(280px,360px)]">
        {tasks.length > 0 ? (
          <div className="grid gap-2">
            <div className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">Task Board</div>
            {tasks.slice(0, 5).map(task => (
              <div key={`${task.task}-${task.assignedTo}`} className="rounded-lg border bg-background/80 px-3 py-2">
                <div className="flex items-center justify-between gap-3">
                  <span className="text-xs font-semibold">{task.task}</span>
                  <span className="text-[10px] text-muted-foreground">{task.status}</span>
                </div>
                <div className="mt-1 text-[11px] text-muted-foreground">
                  {task.assignedTo} · {task.deadline}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div />
        )}

        {/* Phase 63G + 63H + 63I: workspace-wide AI activity column */}
        <div className="space-y-3">
          {/* Phase 63H: 24h rolling digest (counts by intent) */}
          {currentWorkspaceId && (
            <ComposeActivityDigestStrip
              workspaceId={currentWorkspaceId}
              window="24h"
              groupBy="intent"
              topN={4}
            />
          )}
          {/* Phase 63G: live activity firehose */}
          <ComposeActivityPane
            workspaceId={currentWorkspaceId}
            limit={15}
            title="Team co-drafting"
            showChannelLabel
            emptyLabel="No AI Suggest activity in this workspace yet."
          />
          {/* Phase 63I: shared Ask AI feed */}
          <KnowledgeAskFeedPane
            workspaceId={currentWorkspaceId}
            limit={10}
            title="Team Ask AI feed"
            compact
            emptyLabel="No Ask AI activity in this workspace yet."
          />
        </div>
      </div>

      {/* Phase 63I: workspace automation audit — full-width below the grid */}
      <AutomationAuditPanel
        workspaceId={currentWorkspaceId}
        limit={15}
      />
    </div>
  )
}
