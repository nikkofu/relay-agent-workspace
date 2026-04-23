"use client"

import { useEffect } from "react"
import { BarChart2, RefreshCw, Users } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import type { AIComposeActivityDigestFilters, AIComposeActivityDigestGroupBy, AIComposeActivityDigestWindow } from "@/types"

// Phase 63H: intent → color map (mirrors ComposeActivityPane intent pills)
const INTENT_COLORS: Record<string, string> = {
  reply:    'bg-sky-500/15 text-sky-700 dark:text-sky-300',
  summarize:'bg-violet-500/15 text-violet-700 dark:text-violet-300',
  followup: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300',
  schedule: 'bg-amber-500/15 text-amber-700 dark:text-amber-300',
}

function rowColor(key: string, groupBy: AIComposeActivityDigestGroupBy) {
  if (groupBy === 'intent') return INTENT_COLORS[key.toLowerCase()] || 'bg-muted/40 text-muted-foreground'
  return 'bg-muted/40 text-muted-foreground'
}

function keyLabel(key: string, groupBy: AIComposeActivityDigestGroupBy): string {
  if (key === 'unknown') return 'unknown'
  if (groupBy === 'intent') {
    const map: Record<string, string> = { reply: 'Reply', summarize: 'Summarize', followup: 'Follow-up', schedule: 'Schedule' }
    return map[key.toLowerCase()] || key
  }
  if (groupBy === 'channel' && key.startsWith('ch-')) return `#${key}`
  return key
}

interface Props {
  workspaceId?: string
  channelId?: string
  dmId?: string
  window?: AIComposeActivityDigestWindow
  groupBy?: AIComposeActivityDigestGroupBy
  topN?: number
  compact?: boolean
  className?: string
}

export function ComposeActivityDigestStrip({
  workspaceId,
  channelId,
  dmId,
  window: win = '24h',
  groupBy = 'intent',
  topN = 4,
  compact = false,
  className = '',
}: Props) {
  const filters: AIComposeActivityDigestFilters = {
    workspaceId,
    channelId,
    dmId,
    window: win,
    groupBy,
    limit: topN,
  }

  const fetchComposeActivityDigest = useKnowledgeStore(s => s.fetchComposeActivityDigest)
  const composeActivityDigests = useKnowledgeStore(s => s.composeActivityDigests)
  const isLoadingComposeActivityDigest = useKnowledgeStore(s => s.isLoadingComposeActivityDigest)

  // Build cache key to look up the right digest
  const scopePart = channelId ? `ch:${channelId}` : dmId ? `dm:${dmId}` : workspaceId ? `ws:${workspaceId}` : 'none'
  const cacheKey = `${scopePart}|${win}|${groupBy}`
  const digest = composeActivityDigests[cacheKey]
  const loading = isLoadingComposeActivityDigest[cacheKey]

  const hasScope = !!(workspaceId || channelId || dmId)

  useEffect(() => {
    if (hasScope && !digest && !loading) {
      fetchComposeActivityDigest(filters)
    }
  }, [digest, loading, fetchComposeActivityDigest, filters, hasScope])

  if (!hasScope) return null

  const refresh = () => fetchComposeActivityDigest(filters)

  if (loading && !digest) {
    return (
      <div className={`flex items-center gap-2 text-xs text-muted-foreground ${className}`}>
        <RefreshCw className="h-3 w-3 animate-spin" />
        Loading digest…
      </div>
    )
  }

  if (!digest) return null

  const top = digest.breakdown.slice(0, topN)
  const total = digest.summary.total_requests
  const unique = digest.summary.unique_users

  if (total === 0 && !compact) return null

  return (
    <div className={`space-y-1.5 ${className}`}>
      {/* Header row */}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-1.5 text-[10px] font-bold uppercase tracking-wider text-muted-foreground">
          <BarChart2 className="h-3 w-3" />
          AI Co-drafting · {win}
        </div>
        <div className="flex items-center gap-2">
          {unique > 0 && (
            <div className="flex items-center gap-1 text-[10px] text-muted-foreground">
              <Users className="h-3 w-3" />
              {unique}
            </div>
          )}
          <span className="text-[10px] font-semibold tabular-nums">{total} req</span>
          <Button
            variant="ghost"
            size="icon"
            className="h-5 w-5 text-muted-foreground"
            onClick={refresh}
            disabled={!!loading}
          >
            <RefreshCw className={`h-2.5 w-2.5 ${loading ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </div>

      {/* Breakdown bars */}
      {top.length > 0 && (
        <div className="space-y-1">
          {top.map(row => {
            const pct = total > 0 ? Math.round((row.count / total) * 100) : 0
            return (
              <div key={row.key} className="space-y-0.5">
                <div className="flex items-center justify-between gap-2">
                  <Badge
                    variant="secondary"
                    className={`rounded px-1.5 py-0 text-[10px] font-medium ${rowColor(row.key, groupBy)}`}
                  >
                    {keyLabel(row.key, groupBy)}
                  </Badge>
                  <span className="text-[10px] tabular-nums text-muted-foreground">{row.count} · {pct}%</span>
                </div>
                <div className="h-1 w-full rounded-full bg-muted/40">
                  <div
                    className="h-1 rounded-full bg-primary/50 transition-all duration-500"
                    style={{ width: `${pct}%` }}
                  />
                </div>
              </div>
            )
          })}
        </div>
      )}

      {total === 0 && (
        <p className="text-[10px] text-muted-foreground italic">No AI Suggest activity in the last {win}.</p>
      )}
    </div>
  )
}
