"use client"

// ── Tool Run Detail Panel ─────────────────────────────────────────────────────
//
// Displays the real-time progress of a single tool run. Polls the backend
// detail endpoint while the run is still `running` and finalizes on the
// `tool.run.completed` WebSocket event or when polling detects termination.
//
// UX contract:
//   • Opened by clicking any run row in <ChannelToolsPanel>
//   • Shows: tool name, status badge, started-ago, duration, logs timeline
//   • Live log entries stream in via 1.5 s polling while status === "running"
//   • A "Live" pulsing indicator shows the run is still in progress
//   • Writeback target and result shown in a summary strip at the bottom

import { useCallback, useEffect, useRef, useState } from "react"
import {
  X, Loader2, CheckCircle2, XCircle, Clock, Terminal,
  MessageSquare, ListTodo, RefreshCw, AlertCircle,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { API_BASE_URL } from "@/lib/constants"
import { mapToolRun, type ToolRun, type ToolRunLog } from "@/stores/tool-store"
import { formatDistanceToNow, format } from "date-fns"

const STATUS_CONFIG = {
  success: { icon: CheckCircle2, color: "text-emerald-600", bg: "bg-emerald-500/10", label: "Completed" },
  failed:  { icon: XCircle,      color: "text-rose-600",    bg: "bg-rose-500/10",    label: "Failed"    },
  running: { icon: Loader2,      color: "text-amber-600",   bg: "bg-amber-500/10",   label: "Running"   },
  pending: { icon: Clock,        color: "text-sky-600",     bg: "bg-sky-500/10",     label: "Pending"   },
} as const

const LOG_LEVEL_STYLES: Record<string, { dot: string; label: string }> = {
  info:  { dot: "bg-sky-400",     label: "INFO"  },
  warn:  { dot: "bg-amber-400",   label: "WARN"  },
  error: { dot: "bg-rose-500",    label: "ERROR" },
}

interface ToolRunDetailPanelProps {
  runId: string
  onClose: () => void
}

export function ToolRunDetailPanel({ runId, onClose }: ToolRunDetailPanelProps) {
  const [run, setRun] = useState<ToolRun | null>(null)
  const [logs, setLogs] = useState<ToolRunLog[]>([])
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [fetchError, setFetchError] = useState(false)
  const logsEndRef = useRef<HTMLDivElement | null>(null)
  const pollRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const isMountedRef = useRef(true)

  const fetchDetail = useCallback(async (quiet = false) => {
    if (!quiet) setIsRefreshing(true)
    try {
      const res = await fetch(`${API_BASE_URL}/tools/runs/${runId}`)
      if (!res.ok) throw new Error("not found")
      const data = await res.json()
      const mapped = mapToolRun(data.run)
      if (isMountedRef.current) {
        setRun(mapped)
        setFetchError(false)
        if (mapped.structuredLogs) setLogs(mapped.structuredLogs)
        else if (mapped.logs.length > 0) {
          setLogs(mapped.logs.map((msg) => ({ level: "info", message: msg, createdAt: "" })))
        }
      }
      return mapped
    } catch {
      if (isMountedRef.current) setFetchError(true)
      return null
    } finally {
      if (isMountedRef.current && !quiet) setIsRefreshing(false)
    }
  }, [runId])

  // Initial load + polling while running
  useEffect(() => {
    isMountedRef.current = true

    const schedule = async () => {
      const mapped = await fetchDetail(false)
      if (!isMountedRef.current) return
      if (mapped?.status === "running" || mapped?.status === "pending") {
        pollRef.current = setTimeout(schedule, 1500)
      }
    }
    schedule()

    // WS-driven finalization: listen for the completed event on the window
    const handleWS = (e: Event) => {
      const detail = (e as CustomEvent<{ runId: string }>).detail
      if (detail?.runId === runId) fetchDetail(true)
    }
    window.addEventListener("tool-run-completed", handleWS)

    return () => {
      isMountedRef.current = false
      if (pollRef.current) clearTimeout(pollRef.current)
      window.removeEventListener("tool-run-completed", handleWS)
    }
  }, [runId, fetchDetail])

  // Auto-scroll logs
  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [logs.length])

  const cfg = run ? (STATUS_CONFIG[run.status] ?? STATUS_CONFIG.pending) : STATUS_CONFIG.pending
  const StatusIcon = cfg.icon
  const isLive = run?.status === "running" || run?.status === "pending"

  const startedAgo = (() => {
    if (!run?.startedAt) return null
    try { return formatDistanceToNow(new Date(run.startedAt), { addSuffix: true }) }
    catch { return null }
  })()

  const durationSec = run?.durationMs != null ? (run.durationMs / 1000).toFixed(1) : null

  return (
    <div className="flex flex-col h-full animate-in slide-in-from-right-4 duration-200">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2.5 border-b bg-muted/20 shrink-0">
        <div className="flex items-center gap-2 min-w-0">
          <div className={cn("w-6 h-6 rounded-md flex items-center justify-center shrink-0", cfg.bg)}>
            <StatusIcon className={cn("w-3.5 h-3.5", cfg.color, run?.status === "running" && "animate-spin")} />
          </div>
          <div className="min-w-0">
            <p className="text-[11px] font-bold text-foreground truncate leading-tight">
              {run?.toolName || "Tool Run"}
            </p>
            <div className="flex items-center gap-1.5 mt-0.5">
              <span className={cn(
                "text-[8px] font-black uppercase tracking-widest px-1 py-0.5 rounded",
                cfg.bg, cfg.color,
              )}>
                {cfg.label}
              </span>
              {isLive && (
                <span className="flex items-center gap-0.5 text-[8px] font-black uppercase tracking-widest text-amber-600">
                  <span className="w-1.5 h-1.5 rounded-full bg-amber-500 animate-pulse" />
                  Live
                </span>
              )}
            </div>
          </div>
        </div>
        <div className="flex items-center gap-1 shrink-0">
          <button
            type="button"
            onClick={() => fetchDetail(false)}
            className="p-1 rounded hover:bg-muted/60 transition-colors"
            title="Refresh"
          >
            <RefreshCw className={cn("w-3 h-3 text-muted-foreground", isRefreshing && "animate-spin")} />
          </button>
          <button
            type="button"
            onClick={onClose}
            className="p-1 rounded hover:bg-muted/60 transition-colors"
            title="Close"
          >
            <X className="w-3.5 h-3.5 text-muted-foreground" />
          </button>
        </div>
      </div>

      {fetchError && !run && (
        <div className="flex items-center gap-2 px-3 py-4 text-rose-600 text-xs">
          <AlertCircle className="w-4 h-4 shrink-0" />
          Could not load run details.
        </div>
      )}

      {/* Meta strip */}
      {run && (
        <div className="px-3 py-2 border-b bg-muted/10 shrink-0">
          <div className="flex items-center gap-3 flex-wrap text-[9px] text-muted-foreground uppercase tracking-widest">
            {startedAgo && <span>Started {startedAgo}</span>}
            {durationSec && <><span className="opacity-40">·</span><span>{durationSec}s</span></>}
            {run.writebackTarget === "message" && (
              <span className="inline-flex items-center gap-0.5 text-sky-600">
                <MessageSquare className="w-2.5 h-2.5" />→ message
              </span>
            )}
            {run.writebackTarget === "list_item" && (
              <span className="inline-flex items-center gap-0.5 text-violet-600">
                <ListTodo className="w-2.5 h-2.5" />→ list item
              </span>
            )}
          </div>
          {run.summary && (
            <p className="mt-1.5 text-[11px] text-foreground/80 leading-snug">{run.summary}</p>
          )}
        </div>
      )}

      {/* Log timeline */}
      <div className="flex-1 overflow-y-auto px-3 py-2 space-y-1">
        {logs.length === 0 && !isLive && (
          <p className="text-[10px] text-muted-foreground text-center py-8">No log output.</p>
        )}
        {logs.length === 0 && isLive && (
          <div className="flex items-center gap-2 py-6 justify-center text-muted-foreground text-[10px]">
            <Loader2 className="w-3 h-3 animate-spin" />
            Waiting for output…
          </div>
        )}
        {logs.map((log, i) => {
          const levelCfg = LOG_LEVEL_STYLES[log.level?.toLowerCase()] ?? LOG_LEVEL_STYLES.info
          const time = (() => {
            if (!log.createdAt) return null
            try { return format(new Date(log.createdAt), "HH:mm:ss.SSS") }
            catch { return null }
          })()
          return (
            <div key={i} className="flex items-start gap-2 group">
              <div className="flex items-center gap-1 shrink-0 mt-1">
                <span className={cn("w-1.5 h-1.5 rounded-full shrink-0", levelCfg.dot)} />
                <span className="text-[8px] font-black uppercase tracking-widest text-muted-foreground/50 w-7">
                  {levelCfg.label}
                </span>
              </div>
              <p className="text-[11px] leading-snug text-foreground/80 flex-1 font-mono break-all">
                {log.message}
              </p>
              {time && (
                <span className="text-[8px] text-muted-foreground/40 shrink-0 tabular-nums mt-1">
                  {time}
                </span>
              )}
            </div>
          )
        })}
        <div ref={logsEndRef} />
      </div>

      {/* Footer run-id chip */}
      <div className="px-3 py-1.5 border-t bg-muted/10 shrink-0">
        <div className="flex items-center gap-1.5">
          <Terminal className="w-2.5 h-2.5 text-muted-foreground/40" />
          <span className="text-[8px] font-mono text-muted-foreground/40 truncate">{runId}</span>
        </div>
      </div>
    </div>
  )
}
