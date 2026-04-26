"use client"

import { useEffect } from "react"
import { RefreshCw, Loader2, AlertCircle, CheckCircle2, Clock } from "lucide-react"
import { useShallow } from "zustand/react/shallow"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { formatDistanceToNow } from "date-fns"
import type { AIAutomationJobStatus } from "@/types"

// Phase 63H: Visual status map for AIAutomationJob.status
const STATUS_META: Record<
  AIAutomationJobStatus,
  { label: string; color: string; icon: React.ReactNode }
> = {
  pending:   { label: 'Queued',    color: 'bg-sky-500/15 text-sky-700 dark:text-sky-300',     icon: <Clock className="h-3 w-3" /> },
  running:   { label: 'Running',   color: 'bg-violet-500/15 text-violet-700 dark:text-violet-300', icon: <Loader2 className="h-3 w-3 animate-spin" /> },
  succeeded: { label: 'Succeeded', color: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300', icon: <CheckCircle2 className="h-3 w-3" /> },
  failed:    { label: 'Failed',    color: 'bg-red-500/15 text-red-700 dark:text-red-300',      icon: <AlertCircle className="h-3 w-3" /> },
  cancelled: { label: 'Cancelled', color: 'bg-muted/40 text-muted-foreground',                 icon: <AlertCircle className="h-3 w-3" /> },
}

function statusMeta(status: string) {
  return STATUS_META[status] ?? { label: status, color: 'bg-muted/40 text-muted-foreground', icon: <Clock className="h-3 w-3" /> }
}

interface Props {
  entityId: string
}

export function EntityBriefAutomationBanner({ entityId }: Props) {
  const {
    entityBriefAutomation,
    isLoadingEntityBriefAutomation,
    fetchEntityBriefAutomation,
    runEntityBriefAutomation,
    retryEntityBriefAutomation,
  } = useKnowledgeStore(useShallow(s => ({
    entityBriefAutomation: s.entityBriefAutomation,
    isLoadingEntityBriefAutomation: s.isLoadingEntityBriefAutomation,
    fetchEntityBriefAutomation: s.fetchEntityBriefAutomation,
    runEntityBriefAutomation: s.runEntityBriefAutomation,
    retryEntityBriefAutomation: s.retryEntityBriefAutomation,
  })))

  const job = entityBriefAutomation[entityId]   // undefined = not yet loaded; null = never queued
  const isLoading = isLoadingEntityBriefAutomation[entityId]
  const notFetched = job === undefined && !isLoading

  useEffect(() => {
    if (notFetched) {
      fetchEntityBriefAutomation(entityId)
    }
  }, [entityId, notFetched, fetchEntityBriefAutomation])

  if (isLoading && job === undefined) {
    return (
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <Loader2 className="h-3 w-3 animate-spin" />
        Loading automation state…
      </div>
    )
  }

  const isActive = job?.status === 'pending' || job?.status === 'running'
  const canRetry = job?.status === 'failed' || job?.status === 'cancelled'
  const meta = job ? statusMeta(job.status) : null

  return (
    <div className="rounded-lg border bg-muted/20 px-3 py-2 space-y-1.5">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-xs font-semibold">
          <RefreshCw className="h-3.5 w-3.5 text-muted-foreground" />
          Brief Automation
        </div>
        <div className="flex items-center gap-1.5">
          {meta && (
            <Badge
              variant="secondary"
              className={`flex items-center gap-1 rounded px-1.5 py-0 text-[10px] font-medium ${meta.color}`}
            >
              {meta.icon}
              {meta.label}
            </Badge>
          )}
          {!job && !isLoading && (
            <Button
              variant="outline"
              size="sm"
              className="h-6 px-2 text-[11px]"
              onClick={() => runEntityBriefAutomation(entityId)}
            >
              <RefreshCw className="mr-1 h-3 w-3" />
              Regenerate
            </Button>
          )}
          {isActive && (
            <Button variant="outline" size="sm" className="h-6 px-2 text-[11px]" disabled>
              <Loader2 className="mr-1 h-3 w-3 animate-spin" />
              {job?.status === 'pending' ? 'Queued' : 'Running…'}
            </Button>
          )}
          {canRetry && (
            <Button
              variant="outline"
              size="sm"
              className="h-6 px-2 text-[11px]"
              onClick={() => retryEntityBriefAutomation(entityId)}
            >
              <RefreshCw className="mr-1 h-3 w-3" />
              Retry
            </Button>
          )}
          {job && !isActive && job.status === 'succeeded' && (
            <Button
              variant="ghost"
              size="sm"
              className="h-6 px-2 text-[11px]"
              onClick={() => runEntityBriefAutomation(entityId)}
            >
              <RefreshCw className="mr-1 h-3 w-3" />
              Re-run
            </Button>
          )}
        </div>
      </div>
      {job && (
        <div className="flex flex-wrap items-center gap-x-3 gap-y-0.5 text-[10px] text-muted-foreground">
          <span>#{job.id.slice(-8)}</span>
          <span className="capitalize">{job.trigger_reason.replace(/_/g, ' ')}</span>
          {job.attempt_count > 1 && <span>attempt {job.attempt_count}</span>}
          {job.started_at && <span>started {formatDistanceToNow(new Date(job.started_at), { addSuffix: true })}</span>}
          {job.finished_at && job.status === 'succeeded' && (
            <span>finished {formatDistanceToNow(new Date(job.finished_at), { addSuffix: true })}</span>
          )}
          {job.last_error && (
            <span className="text-red-500 dark:text-red-400 truncate max-w-xs" title={job.last_error}>
              {job.last_error}
            </span>
          )}
        </div>
      )}
    </div>
  )
}
