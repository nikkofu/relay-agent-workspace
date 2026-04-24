"use client"

// ── Phase 66 T02: Channel Execution Hub — Container Panel Shell ──────────────
//
// Turns a channel conversation into a structured execution surface by
// surfacing its Lists and recent Tool Runs side-by-side with the message area.
// This is the T02 shell — it intentionally does NOT consume backend fields
// that Gemini has not yet frozen (source_message_id on list items,
// writeback_target on tool runs). T07 will extend this panel with:
//   - "Create List" action
//   - "Run Tool" action
//   - Message-linked list-item rendering
//   - Writeback-target badges on tool runs

import { useState } from "react"
import { ListTodo, Terminal, Zap, X } from "lucide-react"
import { cn } from "@/lib/utils"
import { ChannelListsPanel } from "./channel-lists-panel"
import { ChannelToolsPanel } from "./channel-tools-panel"
import { useListStore } from "@/stores/list-store"
import { useToolStore } from "@/stores/tool-store"

type ExecutionTab = "lists" | "tools"

interface ChannelExecutionPanelProps {
  channelId: string
  onClose?: () => void
}

const TABS: { id: ExecutionTab; label: string; icon: React.ElementType }[] = [
  { id: "lists", label: "Lists", icon: ListTodo },
  { id: "tools", label: "Tools", icon: Terminal },
]

export function ChannelExecutionPanel({ channelId, onClose }: ChannelExecutionPanelProps) {
  const [activeTab, setActiveTab] = useState<ExecutionTab>("lists")

  // Counts for tab badges — read from existing stores (already filtered on fetch)
  const listCount = useListStore(s => s.lists.filter(l => l.channelId === channelId).length)
  const toolRunCount = useToolStore(s => s.toolRuns.filter(r => r.channelId === channelId).length)

  return (
    <div className="flex flex-col h-full bg-white dark:bg-[#1a1d21] border-l w-72 shrink-0">
      {/* Header */}
      <div className="h-14 px-4 flex items-center justify-between border-b shrink-0">
        <div className="flex items-center gap-2">
          <Zap className="w-4 h-4 text-violet-600" />
          <span className="text-sm font-black uppercase tracking-tight">Execution</span>
        </div>
        {onClose && (
          <button
            type="button"
            onClick={onClose}
            className="p-1 rounded hover:bg-muted/60 transition-colors"
            title="Close"
          >
            <X className="w-4 h-4 text-muted-foreground" />
          </button>
        )}
      </div>

      {/* Tab bar */}
      <div className="flex items-center gap-1 px-2 py-1.5 border-b bg-muted/10 shrink-0">
        {TABS.map(tab => {
          const Icon = tab.icon
          const count = tab.id === "lists" ? listCount : toolRunCount
          const isActive = activeTab === tab.id
          return (
            <button
              key={tab.id}
              type="button"
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                "flex-1 flex items-center justify-center gap-1.5 px-2 py-1.5 rounded-md text-[11px] font-bold uppercase tracking-wider transition-colors",
                isActive
                  ? "bg-violet-500/10 text-violet-700 dark:text-violet-300"
                  : "text-muted-foreground hover:bg-muted/40 hover:text-foreground"
              )}
            >
              <Icon className="w-3 h-3" />
              {tab.label}
              {count > 0 && (
                <span className={cn(
                  "text-[9px] font-black px-1 rounded-full",
                  isActive
                    ? "bg-violet-500/20 text-violet-700 dark:text-violet-300"
                    : "bg-muted text-muted-foreground"
                )}>
                  {count}
                </span>
              )}
            </button>
          )
        })}
      </div>

      {/* Panel body */}
      <div className="flex-1 min-h-0 overflow-hidden">
        {activeTab === "lists" && <ChannelListsPanel channelId={channelId} />}
        {activeTab === "tools" && <ChannelToolsPanel channelId={channelId} />}
      </div>
    </div>
  )
}
