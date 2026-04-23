"use client"

import { useEffect, useState } from "react"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { cn } from "@/lib/utils"
import {
  Cpu, Loader2, RefreshCw, CheckCircle2, XCircle,
  AlertCircle, Clock, Play, ExternalLink,
} from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import Link from "next/link"
import type { AIAutomationJobStatus } from "@/types"

// ── Phase 63I: Workspace Automation Audit Panel ──────────────────────────────
//
// Consumes GET /api/v1/ai/automation/jobs and renders a filterable list of
// AIAutomationJob rows with status badges, scope links back to entity detail
// pages, and attempt-count indicators.

interface AutomationAuditPanelProps {
  workspaceId?: string
  limit?: number
  compact?: boolean
}

type StatusFilter = 'all' | AIAutomationJobStatus

const STATUS_CONFIG: Record<string, { label: string; icon: React.ElementType; pill: string }> = {
  pending:   { label: 'Pending',   icon: Clock,         pill: 'bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-400/40' },
  running:   { label: 'Running',   icon: Play,          pill: 'bg-sky-500/10 text-sky-700 dark:text-sky-300 border-sky-400/40' },
  succeeded: { label: 'Succeeded', icon: CheckCircle2,  pill: 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-400/40' },
  failed:    { label: 'Failed',    icon: XCircle,       pill: 'bg-rose-500/10 text-rose-700 dark:text-rose-300 border-rose-400/40' },
  cancelled: { label: 'Cancelled', icon: AlertCircle,   pill: 'bg-muted/60 text-muted-foreground border-muted' },
}
const getStatusCfg = (s: string) =>
  STATUS_CONFIG[s] ?? { label: s, icon: Cpu, pill: 'bg-muted/40 text-muted-foreground border-muted' }

const JOB_TYPE_LABEL: Record<string, string> = {
  entity_brief_regen: 'Brief regen',
}
const jobTypeLabel = (t: string) => JOB_TYPE_LABEL[t] ?? t.replace(/_/g, ' ')

const FILTERS: { value: StatusFilter; label: string }[] = [
  { value: 'all',       label: 'All' },
  { value: 'pending',   label: 'Pending' },
  { value: 'running',   label: 'Running' },
  { value: 'failed',    label: 'Failed' },
  { value: 'succeeded', label: 'Done' },
]

export function AutomationAuditPanel({
  workspaceId,
  limit = 20,
  compact = false,
}: AutomationAuditPanelProps) {
  const {
    automationJobs, isLoadingAutomationJobs, automationJobsTotal,
    fetchAutomationJobs,
  } = useKnowledgeStore()

  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')

  const doFetch = (status?: StatusFilter) => {
    fetchAutomationJobs({
      workspaceId,
      status: (!status || status === 'all') ? undefined : status,
      limit,
    })
  }

  useEffect(() => {
    if (workspaceId) doFetch(statusFilter)
  }, [workspaceId, statusFilter, limit, fetchAutomationJobs])

  const rows = automationJobs

  return (
    <div className="rounded-lg border bg-background/60 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 border-b bg-muted/30">
        <div className="flex items-center gap-1.5">
          <Cpu className="w-3 h-3 text-violet-600" />
          <span className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
            Automation Audit
          </span>
          {automationJobsTotal > 0 && (
            <span className="text-[9px] font-bold px-1.5 py-0.5 rounded-full bg-violet-500/10 text-violet-700 dark:text-violet-400 border border-violet-300 dark:border-violet-700">
              {automationJobsTotal}
            </span>
          )}
          {isLoadingAutomationJobs && <Loader2 className="w-3 h-3 animate-spin text-muted-foreground" />}
        </div>
        <button
          type="button"
          onClick={() => doFetch(statusFilter)}
          className="p-1 rounded hover:bg-muted/60 transition-colors"
          title="Refresh"
        >
          <RefreshCw className="w-3 h-3 text-muted-foreground" />
        </button>
      </div>

      {/* Status filter tabs */}
      <div className="flex items-center gap-1 px-3 py-1.5 border-b bg-muted/10 overflow-x-auto no-scrollbar">
        {FILTERS.map(f => (
          <button
            key={f.value}
            type="button"
            onClick={() => setStatusFilter(f.value)}
            className={cn(
              "text-[9px] font-bold uppercase tracking-widest px-2 py-0.5 rounded border transition-colors shrink-0",
              statusFilter === f.value
                ? "bg-foreground/10 border-foreground/30 text-foreground"
                : "border-transparent text-muted-foreground hover:text-foreground hover:border-muted"
            )}
          >
            {f.label}
          </button>
        ))}
      </div>

      {rows.length === 0 && !isLoadingAutomationJobs && (
        <p className="text-[11px] text-muted-foreground italic p-3">No automation jobs found.</p>
      )}

      <div className="divide-y max-h-64 overflow-y-auto">
        {rows.map(job => {
          const cfg = getStatusCfg(job.status)
          const StatusIcon = cfg.icon
          const isEntityScope = job.scope_type === 'knowledge_entity'
          const ago = (() => {
            const ts = job.finished_at || job.started_at || job.scheduled_at || job.created_at
            try { return ts ? formatDistanceToNow(new Date(ts), { addSuffix: true }) : '' }
            catch { return '' }
          })()

          return (
            <div
              key={job.id}
              className={cn("p-2.5 space-y-1 hover:bg-muted/20 transition-colors", compact && "py-1.5")}
            >
              <div className="flex items-center gap-1.5 flex-wrap">
                {/* Status badge */}
                <span className={cn(
                  "inline-flex items-center gap-1 text-[9px] font-bold uppercase tracking-widest px-1.5 py-0.5 rounded border",
                  cfg.pill
                )}>
                  <StatusIcon className="w-2.5 h-2.5 shrink-0" />
                  {cfg.label}
                </span>

                {/* Job type chip */}
                <span className="text-[9px] font-mono bg-muted/40 px-1.5 py-0.5 rounded border border-muted text-muted-foreground">
                  {jobTypeLabel(job.job_type)}
                </span>

                {/* Scope link → entity detail */}
                {isEntityScope && job.scope_id && (
                  <Link
                    href={`/workspace/knowledge/${job.scope_id}`}
                    className="inline-flex items-center gap-0.5 text-[9px] text-sky-700 dark:text-sky-400 hover:underline"
                  >
                    entity
                    <ExternalLink className="w-2 h-2 shrink-0" />
                  </Link>
                )}

                {job.attempt_count > 1 && (
                  <span className="text-[9px] text-muted-foreground ml-auto">
                    ×{job.attempt_count} attempts
                  </span>
                )}
              </div>

              {!compact && job.last_error && (
                <p className="text-[10px] text-rose-600 dark:text-rose-400 line-clamp-1 pl-0.5">
                  {job.last_error}
                </p>
              )}

              <div className="flex items-center gap-1.5">
                {job.trigger_reason && (
                  <span className="text-[9px] text-muted-foreground capitalize">{job.trigger_reason.replace(/_/g, ' ')}</span>
                )}
                {ago && (
                  <span className="text-[9px] text-muted-foreground ml-auto">{ago}</span>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
