"use client"

// Phase 63G: Reusable co-drafting activity pane.
//
// Hydrates from `GET /api/v1/ai/compose/activity` on mount (scoped by channel /
// DM / workspace / intent) and stays live via the existing
// `knowledge.compose.suggestion.generated` websocket event (wired in
// use-websocket.ts). One shared `composeSuggestionActivity` list in the
// knowledge store powers both historical display and WS-appended rows, so
// every mount of this component — wherever it lives — sees the same canonical
// AIComposeActivity rows.
//
// Two common mount points:
//   - ChannelInfo sheet  → { channelId } (per-channel co-drafting audit)
//   - AgentCollabDashboard → { workspaceId } (workspace-wide firehose)

import { useEffect, useMemo, useCallback } from "react"
import { Loader2, RefreshCw, Sparkles, MessageSquareText, MessageCircle, CalendarClock, Wand2, ListChecks } from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import { Button } from "@/components/ui/button"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { useChannelStore } from "@/stores/channel-store"
import { useUserStore } from "@/stores/user-store"
import { cn } from "@/lib/utils"
import type { AIComposeActivity } from "@/types"

interface ComposeActivityPaneProps {
  // Filter the server query + the client-side rendered list.
  // Exactly one of channelId / dmId / workspaceId is typical; combinations are allowed.
  channelId?: string
  dmId?: string
  workspaceId?: string
  intent?: string
  limit?: number
  // Compact variant: smaller row padding, hides the header title and the manual Refresh action.
  // Intended for embedding in already-dense surfaces like the agent-collab dashboard.
  compact?: boolean
  // Override the heading copy. Defaults to "Co-drafting activity".
  title?: string
  // Override the empty-state copy. Defaults to a sensible message.
  emptyLabel?: string
  // Show the full channel name label (uses channel-store). Enable on workspace-wide mounts;
  // redundant when the pane itself is scoped to a single channel.
  showChannelLabel?: boolean
}

const INTENT_META: Record<string, { label: string; Icon: React.ComponentType<{ className?: string }>; color: string }> = {
  reply: { label: 'Reply', Icon: MessageSquareText, color: 'sky' },
  summarize: { label: 'Summarize', Icon: ListChecks, color: 'violet' },
  followup: { label: 'Follow-up', Icon: MessageCircle, color: 'emerald' },
  schedule: { label: 'Schedule', Icon: CalendarClock, color: 'amber' },
}

function intentMetaFor(intent: string) {
  return INTENT_META[intent] || { label: intent || 'compose', Icon: Wand2, color: 'slate' }
}

// Tailwind can't resolve string-interpolated class names, so hand-map the
// small palette we actually use for intent pills.
const INTENT_PILL_CLASSES: Record<string, string> = {
  sky: 'bg-sky-500/10 border-sky-400/40 text-sky-700 dark:text-sky-400',
  violet: 'bg-violet-500/10 border-violet-400/40 text-violet-700 dark:text-violet-400',
  emerald: 'bg-emerald-500/10 border-emerald-400/40 text-emerald-700 dark:text-emerald-400',
  amber: 'bg-amber-500/10 border-amber-400/40 text-amber-700 dark:text-amber-400',
  slate: 'bg-slate-500/10 border-slate-400/40 text-slate-700 dark:text-slate-400',
}

// Phase 63H: Resolves a user_id to display name via the user store.
// Renders nothing when user_id is empty/undefined (historical pre-63H rows).
function UserChip({ userId, compact }: { userId?: string; compact?: boolean }) {
  const users = useUserStore(s => s.users)
  if (!userId) return null
  const user = users.find(u => u.id === userId)
  const label = user?.name || user?.email || userId.slice(-8)
  return (
    <span
      className={cn(
        "inline-flex items-center rounded border bg-sky-500/10 text-sky-700 dark:text-sky-300 px-1 py-0.5 shrink-0 font-medium",
        compact ? "text-[8px]" : "text-[9px]",
      )}
      title={`User: ${userId}`}
    >
      {label}
    </span>
  )
}

export function ComposeActivityPane({
  channelId,
  dmId,
  workspaceId,
  intent,
  limit = 50,
  compact = false,
  title,
  emptyLabel,
  showChannelLabel = false,
}: ComposeActivityPaneProps) {
  const {
    composeSuggestionActivity,
    isLoadingComposeActivity,
    hasHydratedComposeActivity,
    fetchComposeActivity,
  } = useKnowledgeStore()
  const channels = useChannelStore(s => s.channels)

  // Stable hydration key mirroring the store-side composeActivityScopeKey.
  // Kept local so we don't import a private helper. The exact string doesn't
  // have to match the store's — both are deterministic on identical inputs.
  const scopeKey = useMemo(() => {
    const parts: string[] = []
    if (channelId) parts.push(`ch:${channelId}`)
    if (dmId) parts.push(`dm:${dmId}`)
    if (workspaceId) parts.push(`ws:${workspaceId}`)
    if (intent) parts.push(`intent:${intent}`)
    return parts.length > 0 ? parts.join('|') : 'all'
  }, [channelId, dmId, workspaceId, intent])

  const isLoading = !!isLoadingComposeActivity[scopeKey]
  const hasHydrated = !!hasHydratedComposeActivity[scopeKey]

  // Fetch on mount / whenever the scope changes.
  useEffect(() => {
    fetchComposeActivity({ channelId, dmId, workspaceId, intent, limit }).catch(() => { /* best-effort */ })
  }, [channelId, dmId, workspaceId, intent, limit, fetchComposeActivity])

  // Filter the shared in-memory list to only rows matching this scope. The
  // store also holds rows from other scopes (e.g. from a workspace-wide pane
  // mounted in parallel), so per-channel / per-dm panes need to self-filter.
  const rows = useMemo<AIComposeActivity[]>(() => {
    return composeSuggestionActivity.filter(r => {
      if (channelId && r.channel_id !== channelId) return false
      if (dmId && r.dm_id !== dmId) return false
      if (workspaceId && r.workspace_id && r.workspace_id !== workspaceId) return false
      if (intent && r.intent !== intent) return false
      return true
    }).slice(0, limit)
  }, [composeSuggestionActivity, channelId, dmId, workspaceId, intent, limit])

  const channelNameById = useMemo(() => {
    const map: Record<string, string> = {}
    for (const ch of channels) map[ch.id] = ch.name
    return map
  }, [channels])

  const handleRefresh = useCallback(() => {
    fetchComposeActivity({ channelId, dmId, workspaceId, intent, limit }).catch(() => { /* best-effort */ })
  }, [channelId, dmId, workspaceId, intent, limit, fetchComposeActivity])

  const headerTitle = title || 'Co-drafting activity'

  return (
    <div
      className={cn(
        "rounded-xl border bg-gradient-to-br from-purple-500/5 to-sky-500/5 border-purple-200 dark:border-purple-900/40",
        compact ? "p-2.5" : "p-4",
      )}
    >
      <div className={cn("flex items-center justify-between gap-2", compact ? "mb-1.5" : "mb-2.5")}>
        <div className="flex items-center gap-2 min-w-0">
          <div className="relative shrink-0">
            <Wand2 className={cn(compact ? "w-3 h-3" : "w-3.5 h-3.5", "text-purple-600")} />
            {rows.length > 0 && (
              <span className="absolute -top-0.5 -right-0.5 w-1.5 h-1.5 rounded-full bg-purple-500 animate-pulse" />
            )}
          </div>
          <div className="min-w-0">
            <div className="flex items-center gap-1.5">
              <span className={cn("font-bold text-foreground truncate", compact ? "text-[11px]" : "text-sm")}>
                {headerTitle}
              </span>
              {rows.length > 0 && (
                <span className="text-[9px] font-black uppercase tracking-widest px-1.5 py-0.5 rounded-full border bg-purple-500/10 border-purple-400/40 text-purple-700 dark:text-purple-400">
                  {rows.length}
                </span>
              )}
            </div>
            {!compact && (
              <p className="text-[10px] text-muted-foreground">
                Persisted AI Suggest sessions from this workspace. Updates live.
              </p>
            )}
          </div>
        </div>
        {!compact && (
          <Button
            variant="ghost"
            size="sm"
            className="h-6 text-[9px] gap-1 text-muted-foreground shrink-0"
            onClick={handleRefresh}
            disabled={isLoading}
            title="Re-fetch from GET /api/v1/ai/compose/activity"
          >
            <RefreshCw className={cn("w-2.5 h-2.5", isLoading && "animate-spin")} />
            Refresh
          </Button>
        )}
      </div>

      {isLoading && !hasHydrated && rows.length === 0 ? (
        <div className="flex items-center gap-2 text-[11px] text-muted-foreground italic py-3">
          <Loader2 className="w-3 h-3 animate-spin" />
          Loading co-drafting activity…
        </div>
      ) : rows.length === 0 ? (
        <div
          className={cn(
            "rounded-lg border border-dashed border-muted-foreground/30 bg-background/40 text-[11px] text-muted-foreground italic",
            compact ? "px-2 py-2" : "px-3 py-3",
          )}
        >
          {emptyLabel || "No AI co-drafting yet. When someone asks AI Suggest for help drafting a reply, it will show up here."}
        </div>
      ) : (
        <ul className={cn("space-y-1", compact && "space-y-0.5")}>
          {rows.map(row => {
            const meta = intentMetaFor(row.intent)
            const pillClass = INTENT_PILL_CLASSES[meta.color] || INTENT_PILL_CLASSES.slate
            const when = row.created_at ? (() => {
              const d = new Date(row.created_at)
              return isNaN(d.getTime()) ? row.created_at : formatDistanceToNow(d, { addSuffix: true })
            })() : 'just now'
            const scopeLabel = row.channel_id
              ? (showChannelLabel && channelNameById[row.channel_id] ? `#${channelNameById[row.channel_id]}` : 'channel')
              : row.dm_id
                ? 'DM'
                : 'workspace'
            return (
              <li
                key={row.id || row.compose_id}
                className={cn(
                  "flex items-center gap-2 rounded-md border bg-background/70 hover:bg-background transition-colors",
                  compact ? "px-2 py-1" : "px-2.5 py-1.5",
                )}
              >
                <span
                  className={cn(
                    "inline-flex items-center gap-1 rounded-md border px-1.5 py-0.5 shrink-0",
                    compact ? "text-[9px]" : "text-[10px]",
                    pillClass,
                  )}
                  title={`Intent: ${meta.label}`}
                >
                  <meta.Icon className="w-2.5 h-2.5" />
                  <span className="font-bold uppercase tracking-wide">{meta.label}</span>
                </span>

                <span className={cn("text-muted-foreground shrink-0", compact ? "text-[9px]" : "text-[10px]")}>
                  ×{row.suggestion_count || 1}
                </span>

                {row.thread_id && (
                  <span
                    className={cn(
                      "inline-flex items-center rounded border bg-muted/40 text-muted-foreground px-1 py-0.5 shrink-0",
                      compact ? "text-[8px]" : "text-[9px]",
                    )}
                    title={`Thread ${row.thread_id}`}
                  >
                    <Sparkles className="w-2 h-2 mr-0.5" /> thread
                  </span>
                )}

                <UserChip userId={row.user_id} compact={compact} />

                <span
                  className={cn(
                    "truncate text-foreground/80",
                    compact ? "text-[10px]" : "text-[11px]",
                  )}
                  title={row.provider && row.model ? `${row.provider} / ${row.model}` : (row.provider || row.model || '')}
                >
                  {scopeLabel}
                  {row.provider && (
                    <span className="text-muted-foreground font-mono ml-1">
                      · {row.provider}{row.model ? `/${row.model}` : ''}
                    </span>
                  )}
                </span>

                <span
                  className={cn(
                    "ml-auto text-muted-foreground shrink-0 tabular-nums",
                    compact ? "text-[9px]" : "text-[10px]",
                  )}
                  title={row.created_at}
                >
                  {when}
                </span>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
