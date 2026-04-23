"use client"

import { useEffect } from "react"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { cn } from "@/lib/utils"
import {
  MessageSquare, Loader2, RefreshCw, Tag,
  User2, BookOpen, Building2, FileText, Layout,
  HelpCircle, ExternalLink,
} from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import Link from "next/link"

// ── Phase 63I: Cross-entity shared Ask AI feed ───────────────────────────────
//
// Consumes GET /api/v1/knowledge/ask/recent (hydrated once per workspace key)
// and stays live via `knowledge.entity.ask.answered` WS events wired through
// `applyEntityAskAnswered`. Multiple mounts (e.g. Following Hub + AgentCollab)
// share the same store list — client-side filtering is applied when entityId
// is provided.

interface KnowledgeAskFeedPaneProps {
  workspaceId?: string
  entityId?: string
  limit?: number
  compact?: boolean
  title?: string
  emptyLabel?: string
}

const KIND_ICON: Record<string, React.ElementType> = {
  person:       User2,
  project:      BookOpen,
  concept:      Tag,
  organization: Building2,
  file:         FileText,
  artifact:     Layout,
}
const KIND_COLOR: Record<string, string> = {
  person:       'text-sky-600',
  project:      'text-emerald-600',
  concept:      'text-violet-600',
  organization: 'text-amber-600',
  file:         'text-rose-600',
  artifact:     'text-orange-600',
}
const KIND_BG: Record<string, string> = {
  person:       'bg-sky-500/10 border-sky-300 dark:border-sky-700',
  project:      'bg-emerald-500/10 border-emerald-300 dark:border-emerald-700',
  concept:      'bg-violet-500/10 border-violet-300 dark:border-violet-700',
  organization: 'bg-amber-500/10 border-amber-300 dark:border-amber-700',
  file:         'bg-rose-500/10 border-rose-300 dark:border-rose-700',
  artifact:     'bg-orange-500/10 border-orange-300 dark:border-orange-700',
}

function EntityKindChip({ kind, title, entityId }: { kind?: string; title?: string; entityId: string }) {
  const Icon = (kind && KIND_ICON[kind]) ? KIND_ICON[kind] : HelpCircle
  const color = (kind && KIND_COLOR[kind]) ? KIND_COLOR[kind] : 'text-muted-foreground'
  const bg = (kind && KIND_BG[kind]) ? KIND_BG[kind] : 'bg-muted/40 border-muted'
  return (
    <Link
      href={`/workspace/knowledge/${entityId}`}
      className={cn(
        "inline-flex items-center gap-1 text-[9px] font-bold uppercase tracking-widest px-1.5 py-0.5 rounded border",
        "hover:opacity-80 transition-opacity",
        bg, color
      )}
    >
      <Icon className="w-2.5 h-2.5 shrink-0" />
      <span className="truncate max-w-[80px]">{title || entityId.slice(-6)}</span>
      <ExternalLink className="w-2 h-2 shrink-0 opacity-50" />
    </Link>
  )
}

export function KnowledgeAskFeedPane({
  workspaceId,
  entityId,
  limit = 15,
  compact = false,
  title = "Shared Ask AI feed",
  emptyLabel = "No Ask AI activity yet.",
}: KnowledgeAskFeedPaneProps) {
  const {
    knowledgeAskRecent, isLoadingAskRecent,
    fetchKnowledgeAskRecent,
  } = useKnowledgeStore()

  useEffect(() => {
    if (workspaceId || entityId) {
      fetchKnowledgeAskRecent({ workspaceId, entityId, limit })
    }
  }, [workspaceId, entityId, limit, fetchKnowledgeAskRecent])

  // Client-filter by entityId if provided
  const rows = (entityId
    ? knowledgeAskRecent.filter(r => r.entity_id === entityId)
    : knowledgeAskRecent
  ).slice(0, limit)

  return (
    <div className="rounded-lg border bg-background/60 overflow-hidden">
      <div className="flex items-center justify-between px-3 py-2 border-b bg-muted/30">
        <div className="flex items-center gap-1.5">
          <MessageSquare className="w-3 h-3 text-sky-600" />
          <span className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
            {title}
          </span>
          {rows.length > 0 && (
            <span className="text-[9px] font-bold px-1.5 py-0.5 rounded-full bg-sky-500/10 text-sky-700 dark:text-sky-400 border border-sky-300 dark:border-sky-700">
              {rows.length}
            </span>
          )}
          {isLoadingAskRecent && <Loader2 className="w-3 h-3 animate-spin text-muted-foreground" />}
        </div>
        <button
          type="button"
          onClick={() => {
            // Reset hydration flag for this workspace so the next fetch re-hits the API
            useKnowledgeStore.setState(state => ({
              hasHydratedAskRecent: { ...state.hasHydratedAskRecent, [workspaceId || 'none']: false },
            }))
            fetchKnowledgeAskRecent({ workspaceId, entityId, limit })
          }}
          className="p-1 rounded hover:bg-muted/60 transition-colors"
          title="Refresh ask feed"
        >
          <RefreshCw className="w-3 h-3 text-muted-foreground" />
        </button>
      </div>

      {rows.length === 0 && !isLoadingAskRecent && (
        <p className="text-[11px] text-muted-foreground italic p-3">{emptyLabel}</p>
      )}

      <div className="divide-y">
        {rows.map(row => {
          const ago = (() => {
            try { return formatDistanceToNow(new Date(row.answered_at), { addSuffix: true }) }
            catch { return '' }
          })()
          return (
            <div
              key={row.id}
              className={cn("p-2.5 space-y-1 hover:bg-muted/30 transition-colors", compact && "py-1.5")}
            >
              <div className="flex items-center gap-1.5 flex-wrap">
                {/* Entity chip — only shown on cross-entity (workspace) mounts */}
                {!entityId && row.entity_id && (
                  <EntityKindChip
                    kind={row.entity_kind}
                    title={row.entity_title}
                    entityId={row.entity_id}
                  />
                )}
                <span className="text-[10px] font-semibold text-foreground/90 truncate flex-1 min-w-0">
                  {row.question}
                </span>
              </div>

              {!compact && row.answer && (
                <p className="text-[10px] text-muted-foreground leading-relaxed line-clamp-2 pl-0.5">
                  {row.answer}
                </p>
              )}

              <div className="flex items-center gap-1.5 flex-wrap">
                {row.citation_count > 0 && (
                  <span className="text-[9px] text-sky-700 dark:text-sky-400 font-semibold">
                    {row.citation_count} citation{row.citation_count !== 1 ? 's' : ''}
                  </span>
                )}
                {(row.provider || row.model) && (
                  <span className="text-[9px] font-mono text-muted-foreground bg-muted/40 px-1 rounded">
                    {[row.provider, row.model].filter(Boolean).join('/')}
                  </span>
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
