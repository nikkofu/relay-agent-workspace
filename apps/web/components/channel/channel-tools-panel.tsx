"use client"

// ── Phase 66 T07: Channel Execution Hub — Tools Panel (wired) ────────────────
//
// T02 shell is now wired to Gemini's v0.6.39 backend contract (frozen Q3):
//   - writeback_target + writeback persisted on each ToolRun → rendered as a
//     sky/violet badge (sky for "message", violet for "list_item") next to the
//     status pill. Runs without a writeback target render unchanged.
//   - "Run Tool" CTA now enabled via <RunToolDialog> which supports both
//     writeback targets per-contract.

import { useEffect, useRef, useState } from "react"
import {
  Loader2, Terminal, CheckCircle2, XCircle, Clock, RefreshCw, Play,
  MessageSquare, ListTodo,
} from "lucide-react"
import { useToolStore } from "@/stores/tool-store"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { formatDistanceToNow } from "date-fns"
import { RunToolDialog } from "./run-tool-dialog"

interface ChannelToolsPanelProps {
  channelId: string
}

const STATUS_CONFIG = {
  success: { icon: CheckCircle2, color: "text-emerald-600", bg: "bg-emerald-500/10", label: "Success" },
  failed:  { icon: XCircle,      color: "text-rose-600",    bg: "bg-rose-500/10",    label: "Failed"  },
  running: { icon: Loader2,      color: "text-amber-600",   bg: "bg-amber-500/10",   label: "Running" },
  pending: { icon: Clock,        color: "text-sky-600",     bg: "bg-sky-500/10",     label: "Pending" },
} as const

export function ChannelToolsPanel({ channelId }: ChannelToolsPanelProps) {
  const { toolRuns, isLoading, fetchToolRuns } = useToolStore()
  const didFetch = useRef<string>("")
  const [showRunDialog, setShowRunDialog] = useState(false)

  useEffect(() => {
    if (didFetch.current !== channelId) {
      didFetch.current = channelId
      fetchToolRuns(channelId)
    }
  }, [channelId, fetchToolRuns])

  const channelRuns = toolRuns.filter(r => r.channelId === channelId)

  const handleRefresh = () => {
    didFetch.current = ""
    fetchToolRuns(channelId)
  }

  if (isLoading && channelRuns.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground gap-2">
        <Loader2 className="w-5 h-5 animate-spin" />
        <span className="text-xs">Loading tool runs…</span>
      </div>
    )
  }

  if (channelRuns.length === 0) {
    return (
      <>
        <div className="flex flex-col items-center justify-center py-12 px-4 text-center gap-2">
          <div className="w-12 h-12 rounded-full bg-amber-500/10 flex items-center justify-center">
            <Terminal className="w-6 h-6 text-amber-600" />
          </div>
          <p className="text-sm font-semibold text-foreground">No tool runs in this channel</p>
          <p className="text-[11px] text-muted-foreground max-w-[220px] leading-snug">
            Run a tool to push structured output back into this channel as a message or list item.
          </p>
          <Button
            variant="outline"
            size="sm"
            className="mt-2 gap-1.5 text-xs font-bold h-7"
            onClick={() => setShowRunDialog(true)}
          >
            <Play className="w-3 h-3" />
            Run Tool
          </Button>
        </div>
        <RunToolDialog open={showRunDialog} onOpenChange={setShowRunDialog} channelId={channelId} />
      </>
    )
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header row */}
      <div className="flex items-center justify-between px-3 py-2 border-b bg-muted/20">
        <span className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
          {channelRuns.length} Recent {channelRuns.length === 1 ? "Run" : "Runs"}
        </span>
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={() => setShowRunDialog(true)}
            className="p-1 rounded hover:bg-amber-500/15 text-amber-600 transition-colors"
            title="Run Tool"
          >
            <Play className="w-3.5 h-3.5" />
          </button>
          <button
            type="button"
            onClick={handleRefresh}
            className="p-1 rounded hover:bg-muted/60 transition-colors"
            title="Refresh"
          >
            <RefreshCw className={cn("w-3 h-3 text-muted-foreground", isLoading && "animate-spin")} />
          </button>
        </div>
      </div>
      <RunToolDialog open={showRunDialog} onOpenChange={setShowRunDialog} channelId={channelId} />

      {/* Run rows */}
      <div className="flex-1 overflow-y-auto">
        {channelRuns.map(run => {
          const cfg = STATUS_CONFIG[run.status] ?? STATUS_CONFIG.pending
          const StatusIcon = cfg.icon
          const startedAgo = (() => {
            try { return formatDistanceToNow(new Date(run.startedAt), { addSuffix: true }) }
            catch { return "" }
          })()
          const durationSec = run.durationMs ? (run.durationMs / 1000).toFixed(1) : null

          return (
            <div
              key={run.id}
              className="px-3 py-2.5 border-b hover:bg-muted/30 transition-colors cursor-pointer group"
            >
              <div className="flex items-start gap-2">
                <div className={cn(
                  "w-7 h-7 rounded-lg flex items-center justify-center shrink-0 mt-0.5",
                  cfg.bg
                )}>
                  <StatusIcon className={cn(
                    "w-3.5 h-3.5",
                    cfg.color,
                    run.status === "running" && "animate-spin"
                  )} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <p className="text-[12px] font-semibold text-foreground truncate leading-tight">
                      {run.toolName || run.toolKey}
                    </p>
                    <span className={cn(
                      "text-[9px] font-black uppercase tracking-widest px-1.5 py-0.5 rounded shrink-0",
                      cfg.bg, cfg.color
                    )}>
                      {cfg.label}
                    </span>
                  </div>
                  <div className="flex items-center gap-2 mt-1 flex-wrap">
                    {startedAgo && (
                      <span className="text-[9px] text-muted-foreground/60 uppercase tracking-widest">
                        {startedAgo}
                      </span>
                    )}
                    {durationSec && (
                      <>
                        <span className="text-muted-foreground/40 text-[9px]">•</span>
                        <span className="text-[9px] text-muted-foreground/60 uppercase tracking-widest">
                          {durationSec}s
                        </span>
                      </>
                    )}
                    {/* Phase 66 T07: writeback-target badge */}
                    {run.writebackTarget === "message" && (
                      <span className="inline-flex items-center gap-0.5 text-[9px] font-black uppercase tracking-widest px-1 py-0.5 rounded bg-sky-500/10 text-sky-700 dark:text-sky-300">
                        <MessageSquare className="w-2 h-2" />
                        → message
                      </span>
                    )}
                    {run.writebackTarget === "list_item" && (
                      <span className="inline-flex items-center gap-0.5 text-[9px] font-black uppercase tracking-widest px-1 py-0.5 rounded bg-violet-500/10 text-violet-700 dark:text-violet-300">
                        <ListTodo className="w-2 h-2" />
                        → list item
                      </span>
                    )}
                  </div>
                  {run.status === "failed" && run.logs && run.logs.length > 0 && (
                    <p className="text-[10px] text-rose-600 mt-1 line-clamp-2">
                      {run.logs[run.logs.length - 1]}
                    </p>
                  )}
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
