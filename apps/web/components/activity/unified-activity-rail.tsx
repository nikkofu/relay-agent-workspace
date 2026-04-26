"use client"

import { useEffect, useMemo, useState } from "react"
import { useActivityStore } from "@/stores/activity-store"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { cn } from "@/lib/utils"
import {
  Loader2, RefreshCw, Activity,
  MessageSquare, AtSign, ThumbsUp, UserPlus, Mail,
  Paperclip, Terminal, ListTodo, Cpu, Wand2, BookOpen,
  CalendarCheck, ExternalLink, ChevronDown,
  Tag, User2, Building2, FileText, Layout,
  Sparkles,
} from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import Link from "next/link"
import type { UnifiedActivityFeedItem, UnifiedActivityEventType } from "@/types"
import {
  buildExecutionHistoryBody,
  buildExecutionHistorySummary,
  getCreatedObjectHref,
  normalizeExecutionHistoryEvent,
} from "@/lib/execution-history"

// ── Phase 64C: Unified Activity Rail ─────────────────────────────────────────
//
// Aggregates ALL workspace signals into one chronological feed:
//  • "All"      → GET /api/v1/activity/feed (live since v0.6.32, Codex Phase 64B)
//  • "AI Events"→ compose activity + Ask AI feed + automation jobs, no extra API.
//  • "Messages" → traditional activity items (mentions, replies, joins…)
//  • "Files"    → unified feed filtered to file_uploaded events
//  • "Bookings" → unified feed filtered to schedule_booking events
//
// Phase 64C adds: actor_name display, Files/Bookings wired to unified feed,
// WS live-append via appendUnifiedFeedItem (wired in use-websocket.ts).

interface UnifiedActivityRailProps {
  workspaceId?: string
  /** Show compact row variant for dashboard sidebar usage */
  compact?: boolean
  className?: string
  /** Override initial tab */
  defaultTab?: TabId
}

type TabId = 'ai' | 'all' | 'messages' | 'mentions' | 'files' | 'bookings'

const TABS: { id: TabId; label: string; icon: React.ElementType }[] = [
  { id: 'ai',       label: 'AI Events',  icon: Wand2 },
  { id: 'all',      label: 'All',        icon: Activity },
  { id: 'messages', label: 'Messages',   icon: MessageSquare },
  { id: 'mentions', label: 'Mentions',   icon: AtSign },
  { id: 'files',    label: 'Files',      icon: Paperclip },
  { id: 'bookings', label: 'Bookings',   icon: CalendarCheck },
]

// ── event-type display config ─────────────────────────────────────────────────

type EventCfg = { label: string; icon: React.ElementType; pill: string }

const MENTION_USER_CFG: EventCfg  = { label: 'Mention',        icon: AtSign,    pill: 'bg-fuchsia-500/10 text-fuchsia-700 dark:text-fuchsia-300 border-fuchsia-400/30' }
const MENTION_ENTITY_CFG: EventCfg = { label: 'Entity Mention', icon: Tag,       pill: 'bg-violet-500/10 text-violet-700 dark:text-violet-300 border-violet-400/30' }

const EVENT_CONFIG: Partial<Record<UnifiedActivityEventType, EventCfg>> = {
  message:          { label: 'Message',      icon: MessageSquare, pill: 'bg-sky-500/10 text-sky-700 dark:text-sky-300 border-sky-400/30' },
  mention:          MENTION_USER_CFG,
  reaction:         { label: 'Reaction',     icon: ThumbsUp,      pill: 'bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-400/30' },
  reply:            { label: 'Reply',        icon: MessageSquare, pill: 'bg-sky-500/10 text-sky-700 dark:text-sky-300 border-sky-400/30' },
  channel_join:     { label: 'Joined',       icon: UserPlus,      pill: 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-400/30' },
  dm_message:       { label: 'DM',           icon: Mail,          pill: 'bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 border-indigo-400/30' },
  file_uploaded:    { label: 'File',         icon: Paperclip,     pill: 'bg-blue-500/10 text-blue-700 dark:text-blue-300 border-blue-400/30' },
  artifact_updated: { label: 'Artifact',     icon: ListTodo,      pill: 'bg-fuchsia-500/10 text-fuchsia-700 dark:text-fuchsia-300 border-fuchsia-400/30' },
  schedule_booking: { label: 'Booking',      icon: CalendarCheck, pill: 'bg-teal-500/10 text-teal-700 dark:text-teal-300 border-teal-400/30' },
  compose_activity: { label: 'AI Compose',   icon: Wand2,         pill: 'bg-violet-500/10 text-violet-700 dark:text-violet-300 border-violet-400/30' },
  knowledge_ask:    { label: 'Ask AI',       icon: BookOpen,      pill: 'bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 border-indigo-400/30' },
  automation_job:   { label: 'Automation',   icon: Cpu,           pill: 'bg-rose-500/10 text-rose-700 dark:text-rose-300 border-rose-400/30' },
  tool_run:         { label: 'Tool Run',     icon: Terminal,      pill: 'bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-400/30' },
  ai_execution:    { label: 'AI Execute',   icon: Sparkles,      pill: 'bg-violet-500/10 text-violet-700 dark:text-violet-300 border-violet-400/30' },
}
const getEventCfg = (t: UnifiedActivityEventType): EventCfg =>
  EVENT_CONFIG[t] ?? { label: t, icon: Activity, pill: 'bg-muted/40 text-muted-foreground border-muted' }

// ── entity kind chip config ───────────────────────────────────────────────────

const KIND_CONFIG: Record<string, { icon: React.ElementType; color: string }> = {
  person:       { icon: User2,       color: 'text-sky-600' },
  project:      { icon: BookOpen,    color: 'text-emerald-600' },
  concept:      { icon: Tag,         color: 'text-violet-600' },
  organization: { icon: Building2,   color: 'text-amber-600' },
  file:         { icon: FileText,    color: 'text-rose-600' },
  artifact:     { icon: Layout,      color: 'text-orange-600' },
}
const getKindIcon = (kind?: string) => (kind && KIND_CONFIG[kind]?.icon) || Tag
const getKindColor = (kind?: string) => (kind && KIND_CONFIG[kind]?.color) || 'text-muted-foreground'

// ── AI feed aggregation from existing stores ──────────────────────────────────

function normalizeAIExecutionFeedItem(item: UnifiedActivityFeedItem): UnifiedActivityFeedItem {
  if (item.event_type !== 'ai_execution') return item
  const event = normalizeExecutionHistoryEvent({
    id: item.id,
    event_type: item.meta?.event_type,
    status: item.meta?.event_type === 'failed' ? 'failed' : 'success',
    analysis_snapshot_id: item.meta?.analysis_snapshot_id ?? item.id,
    execution_target_type: item.meta?.execution_target_type,
    created_object_id: item.meta?.created_object_id,
    created_object_type: item.meta?.created_object_type,
    failure_stage: item.meta?.failure_stage,
    error_message: item.meta?.error_message,
    created_at: item.occurred_at,
  })
  if (!event) return item
  return {
    ...item,
    title: buildExecutionHistorySummary(event),
    body: buildExecutionHistoryBody(event) ?? item.body,
    link: item.link ?? getCreatedObjectHref(event) ?? undefined,
  }
}

function useAIFeedItems(workspaceId: string | undefined, unifiedFeedItems: UnifiedActivityFeedItem[]): UnifiedActivityFeedItem[] {
  const { composeSuggestionActivity, knowledgeAskRecent, automationJobs } = useKnowledgeStore()

  return useMemo(() => {
    const items: UnifiedActivityFeedItem[] = []

    for (const item of unifiedFeedItems) {
      if (item.event_type !== 'ai_execution') continue
      items.push(normalizeAIExecutionFeedItem(item))
    }

    // Compose activity rows
    for (const a of composeSuggestionActivity) {
      if (workspaceId && a.workspace_id && a.workspace_id !== workspaceId) continue
      items.push({
        id: `compose-${a.compose_id}`,
        event_type: 'compose_activity',
        workspace_id: a.workspace_id,
        channel_id: a.channel_id ?? undefined,
        title: `AI Compose · ${a.intent ?? 'suggest'} · ${a.suggestion_count ?? 1} suggestion${(a.suggestion_count ?? 1) !== 1 ? 's' : ''}`,
        body: a.provider ? `${a.provider}${a.model ? ` / ${a.model}` : ''}` : undefined,
        occurred_at: a.created_at,
        meta: { intent: a.intent, suggestion_count: a.suggestion_count },
      })
    }

    // Ask AI feed rows
    for (const r of knowledgeAskRecent) {
      items.push({
        id: `ask-${r.id}`,
        event_type: 'knowledge_ask',
        entity_id: r.entity_id,
        entity_title: r.entity_title,
        entity_kind: r.entity_kind,
        title: r.entity_title ? `Ask AI · ${r.entity_title}` : `Ask AI · entity`,
        body: r.question,
        link: r.entity_id ? `/workspace/knowledge/${r.entity_id}` : undefined,
        occurred_at: r.answered_at ?? r.created_at ?? new Date().toISOString(),
        meta: { citation_count: r.citation_count, provider: r.provider, model: r.model },
      })
    }

    // Automation job rows (most recent 20)
    for (const j of automationJobs.slice(0, 20)) {
      if (workspaceId && j.workspace_id !== workspaceId) continue
      const ts = j.finished_at || j.started_at || j.scheduled_at || j.created_at
      items.push({
        id: `job-${j.id}`,
        event_type: 'automation_job',
        workspace_id: j.workspace_id,
        entity_id: j.scope_type === 'knowledge_entity' ? j.scope_id : undefined,
        link: j.scope_type === 'knowledge_entity' && j.scope_id
          ? `/workspace/knowledge/${j.scope_id}` : undefined,
        title: `Automation · ${j.job_type.replace(/_/g, ' ')} · ${j.status}`,
        body: j.last_error || j.trigger_reason?.replace(/_/g, ' '),
        occurred_at: ts,
        meta: { status: j.status, attempt_count: j.attempt_count },
      })
    }

    return items.sort((a, b) => new Date(b.occurred_at).getTime() - new Date(a.occurred_at).getTime())
  }, [composeSuggestionActivity, knowledgeAskRecent, automationJobs, unifiedFeedItems, workspaceId])
}

// ── Row component ─────────────────────────────────────────────────────────────

function FeedRow({ item, compact }: { item: UnifiedActivityFeedItem; compact: boolean }) {
  // Phase 65B: distinguish user mentions (fuchsia @) from entity mentions (violet tag)
  const cfg = item.event_type === 'mention'
    ? (item.meta?.mention_kind === 'entity' ? MENTION_ENTITY_CFG : MENTION_USER_CFG)
    : getEventCfg(item.event_type)
  const Icon = cfg.icon
  const KindIcon = getKindIcon(item.entity_kind)
  const kindColor = getKindColor(item.entity_kind)
  const ago = (() => {
    try { return formatDistanceToNow(new Date(item.occurred_at), { addSuffix: true }) }
    catch { return '' }
  })()

  const row = item.link ? (
    <Link href={item.link} className="block px-3 hover:bg-muted/20 transition-colors space-y-0.5 group/row" style={{ paddingTop: compact ? '6px' : '10px', paddingBottom: compact ? '6px' : '10px' }}>
      <RowInner item={item} cfg={cfg} Icon={Icon} KindIcon={KindIcon} kindColor={kindColor} ago={ago} compact={compact} />
    </Link>
  ) : (
    <div className={cn("px-3 hover:bg-muted/20 transition-colors space-y-0.5", compact ? "py-1.5" : "py-2.5")}>
      <RowInner item={item} cfg={cfg} Icon={Icon} KindIcon={KindIcon} kindColor={kindColor} ago={ago} compact={compact} />
    </div>
  )

  return row
}

function RowInner({ item, cfg, Icon, KindIcon, kindColor, ago, compact }: {
  item: UnifiedActivityFeedItem
  cfg: EventCfg
  Icon: React.ElementType
  KindIcon: React.ElementType
  kindColor: string
  ago: string
  compact: boolean
}) {
  return (
    <>
      <div className="flex items-center gap-1.5 flex-wrap">
        {/* event type badge */}
        <span className={cn(
          "inline-flex items-center gap-1 text-[9px] font-bold uppercase tracking-widest px-1.5 py-0.5 rounded border",
          cfg.pill
        )}>
          <Icon className="w-2.5 h-2.5 shrink-0" />
          {cfg.label}
        </span>

        {/* actor name */}
        {item.actor_name && (
          <span className="text-[9px] font-semibold text-foreground/70">{item.actor_name}</span>
        )}

        {/* entity kind chip */}
        {item.entity_id && (
          <span className="inline-flex items-center gap-0.5 text-[9px] text-sky-700 dark:text-sky-400">
            <KindIcon className={cn("w-2.5 h-2.5", kindColor)} />
            <span>{item.entity_title ?? item.entity_id.slice(-6)}</span>
            <ExternalLink className="w-2 h-2" />
          </span>
        )}

        {item.channel_name && (
          <span className="text-[9px] text-muted-foreground">#{item.channel_name}</span>
        )}

        {ago && <span className="text-[9px] text-muted-foreground ml-auto">{ago}</span>}
      </div>

      <p className={cn("text-[11px] text-foreground/80 leading-snug", compact ? "line-clamp-1" : "line-clamp-2")}>
        {item.title}
      </p>

      {!compact && item.body && (
        <p className="text-[10px] text-muted-foreground line-clamp-1">{item.body}</p>
      )}
    </>
  )
}

// ── Main component ────────────────────────────────────────────────────────────

export function UnifiedActivityRail({
  workspaceId,
  compact = false,
  className,
  defaultTab = 'ai',
}: UnifiedActivityRailProps) {
  const [activeTab, setActiveTab] = useState<TabId>(defaultTab)
  const [showMore, setShowMore] = useState(false)

  const {
    unifiedFeedItems, isLoadingUnifiedFeed, hasMoreUnifiedFeed, unifiedFeedCursor,
    activities, fetchActivities, fetchUnifiedFeed,
  } = useActivityStore()

  const currentWorkspaceId = useWorkspaceStore(s => s.currentWorkspace?.id)
  const wsId = workspaceId ?? currentWorkspaceId

  const aiFeedItems = useAIFeedItems(wsId, unifiedFeedItems)

  useEffect(() => {
    if (!wsId) return
    if (activeTab === 'ai') {
      fetchUnifiedFeed({ workspaceId: wsId, eventType: 'ai_execution', limit: 40 })
    } else if (activeTab === 'all') {
      fetchUnifiedFeed({ workspaceId: wsId, limit: 50 })
    } else if (activeTab === 'files') {
      fetchUnifiedFeed({ workspaceId: wsId, eventType: 'file_uploaded', limit: 40 })
    } else if (activeTab === 'bookings') {
      fetchUnifiedFeed({ workspaceId: wsId, eventType: 'schedule_booking', limit: 40 })
    } else if (activeTab === 'mentions') {
      fetchUnifiedFeed({ workspaceId: wsId, eventType: 'mention', limit: 40 })
    } else if (activeTab === 'messages') {
      fetchActivities()
    }
  }, [activeTab, wsId, fetchUnifiedFeed, fetchActivities])

  const messageItems = activities.filter(a =>
    ['mention', 'reply', 'thread_reply', 'channel_join', 'dm_message'].includes(a.type)
  )

  const visibleAI = showMore ? aiFeedItems : aiFeedItems.slice(0, 30)
  const visibleAll = showMore ? unifiedFeedItems : unifiedFeedItems.slice(0, 30)

  const isLoading = ['ai', 'all', 'mentions', 'files', 'bookings'].includes(activeTab) ? isLoadingUnifiedFeed : false
  const isEmpty = {
    ai:       aiFeedItems.length === 0,
    all:      unifiedFeedItems.length === 0 && !isLoadingUnifiedFeed,
    messages: messageItems.length === 0,
    mentions: unifiedFeedItems.length === 0 && !isLoadingUnifiedFeed,
    files:    unifiedFeedItems.length === 0 && !isLoadingUnifiedFeed,
    bookings: unifiedFeedItems.length === 0 && !isLoadingUnifiedFeed,
  }[activeTab]

  const renderRows = () => {
    if (isLoading) return (
      <div className="flex items-center justify-center py-8 text-muted-foreground gap-2">
        <Loader2 className="w-4 h-4 animate-spin" />
        <span className="text-xs">Loading…</span>
      </div>
    )

    if (activeTab === 'ai') {
      if (aiFeedItems.length === 0) return (
        <p className="text-[11px] text-muted-foreground italic px-3 py-4">
          No AI events yet in this workspace.
        </p>
      )
      return (
        <>
          {visibleAI.map(item => <FeedRow key={item.id} item={item} compact={compact} />)}
          {aiFeedItems.length > 30 && !showMore && (
            <button type="button" onClick={() => setShowMore(true)}
              className="w-full py-2 text-[10px] text-muted-foreground hover:text-foreground flex items-center justify-center gap-1 hover:bg-muted/20 transition-colors">
              <ChevronDown className="w-3 h-3" /> Show {aiFeedItems.length - 30} more
            </button>
          )}
        </>
      )
    }

    if (activeTab === 'all') {
      if (unifiedFeedItems.length === 0) return (
        <div className="px-3 py-4 space-y-1">
          <p className="text-[11px] text-muted-foreground italic">
            No activity yet in this workspace.
          </p>
          <p className="text-[10px] text-muted-foreground/60">
            Events will appear here as your team uses channels, files, and AI features.
          </p>
        </div>
      )
      return (
        <>
          {visibleAll.map(item => <FeedRow key={item.id} item={item} compact={compact} />)}
          {(unifiedFeedItems.length > 30 && !showMore) && (
            <button type="button" onClick={() => setShowMore(true)}
              className="w-full py-2 text-[10px] text-muted-foreground hover:text-foreground flex items-center justify-center gap-1 hover:bg-muted/20 transition-colors">
              <ChevronDown className="w-3 h-3" /> Show more
            </button>
          )}
          {showMore && hasMoreUnifiedFeed && (
            <button type="button"
              onClick={() => fetchUnifiedFeed({ workspaceId: wsId, limit: 50, cursor: unifiedFeedCursor ?? undefined })}
              className="w-full py-2 text-[10px] text-muted-foreground hover:text-foreground flex items-center justify-center gap-1 hover:bg-muted/20 transition-colors">
              <ChevronDown className="w-3 h-3" /> Load more
            </button>
          )}
        </>
      )
    }

    if (activeTab === 'mentions') {
      if (unifiedFeedItems.length === 0) return (
        <div className="px-3 py-4 space-y-1">
          <p className="text-[11px] text-muted-foreground italic">No mentions yet.</p>
          <p className="text-[10px] text-muted-foreground/60">You'll see @user and @entity mentions here as they happen.</p>
        </div>
      )
      return (
        <>
          {visibleAll.map(item => <FeedRow key={item.id} item={item} compact={compact} />)}
          {(unifiedFeedItems.length > 30 && !showMore) && (
            <button type="button" onClick={() => setShowMore(true)}
              className="w-full py-2 text-[10px] text-muted-foreground hover:text-foreground flex items-center justify-center gap-1 hover:bg-muted/20 transition-colors">
              <ChevronDown className="w-3 h-3" /> Show {unifiedFeedItems.length - 30} more
            </button>
          )}
          {/* Phase 65C: cursor-based infinite scroll using next_cursor (contract v0.6.40) */}
          {showMore && hasMoreUnifiedFeed && (
            <button type="button"
              onClick={() => fetchUnifiedFeed({ workspaceId: wsId, eventType: 'mention', limit: 40, cursor: unifiedFeedCursor ?? undefined })}
              className="w-full py-2 text-[10px] text-muted-foreground hover:text-foreground flex items-center justify-center gap-1 hover:bg-muted/20 transition-colors">
              <ChevronDown className="w-3 h-3" /> Load more mentions
            </button>
          )}
        </>
      )
    }

    if (activeTab === 'messages') {
      if (messageItems.length === 0) return <p className="text-[11px] text-muted-foreground italic px-3 py-4">No recent message activity.</p>
      return messageItems.slice(0, 40).map(a => {
        const item: UnifiedActivityFeedItem = {
          id: a.id,
          event_type: a.type as UnifiedActivityEventType,
          title: `${a.user?.name ?? 'Someone'} ${a.summary}${a.target ? ` · #${a.target}` : ''}`,
          occurred_at: a.occurredAt,
          channel_name: a.channel?.name,
        }
        return <FeedRow key={item.id} item={item} compact={compact} />
      })
    }

    if (activeTab === 'files') {
      if (unifiedFeedItems.length === 0) return <p className="text-[11px] text-muted-foreground italic px-3 py-4">No file uploads yet.</p>
      return (
        <>
          {visibleAll.map(item => <FeedRow key={item.id} item={item} compact={compact} />)}
          {(unifiedFeedItems.length > 30 && !showMore) && (
            <button type="button" onClick={() => setShowMore(true)}
              className="w-full py-2 text-[10px] text-muted-foreground hover:text-foreground flex items-center justify-center gap-1 hover:bg-muted/20 transition-colors">
              <ChevronDown className="w-3 h-3" /> Show {unifiedFeedItems.length - 30} more
            </button>
          )}
        </>
      )
    }

    if (activeTab === 'bookings') {
      if (unifiedFeedItems.length === 0) return (
        <p className="text-[11px] text-muted-foreground italic px-3 py-4">
          No schedule bookings yet. Book a slot from the message composer.
        </p>
      )
      return (
        <>
          {visibleAll.map(item => <FeedRow key={item.id} item={item} compact={compact} />)}
          {(unifiedFeedItems.length > 30 && !showMore) && (
            <button type="button" onClick={() => setShowMore(true)}
              className="w-full py-2 text-[10px] text-muted-foreground hover:text-foreground flex items-center justify-center gap-1 hover:bg-muted/20 transition-colors">
              <ChevronDown className="w-3 h-3" /> Show {unifiedFeedItems.length - 30} more
            </button>
          )}
        </>
      )
    }

    return null
  }

  return (
    <div className={cn("rounded-lg border bg-background/60 overflow-hidden", className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 border-b bg-muted/30">
        <div className="flex items-center gap-1.5">
          <Activity className="w-3 h-3 text-violet-600" />
          <span className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
            Activity Feed
          </span>
          {activeTab === 'ai' && aiFeedItems.length > 0 && (
            <span className="text-[9px] font-bold px-1.5 py-0.5 rounded-full bg-violet-500/10 text-violet-700 dark:text-violet-400 border border-violet-300 dark:border-violet-700">
              {aiFeedItems.length}
            </span>
          )}
          {isLoading && <Loader2 className="w-3 h-3 animate-spin text-muted-foreground" />}
        </div>
        {activeTab === 'all' && (
          <button
            type="button"
            onClick={() => {
              setShowMore(false)
              if (wsId) fetchUnifiedFeed({ workspaceId: wsId, limit: 50 })
            }}
            className="p-1 rounded hover:bg-muted/60 transition-colors"
            title="Refresh"
          >
            <RefreshCw className="w-3 h-3 text-muted-foreground" />
          </button>
        )}
      </div>

      {/* Tab bar */}
      <div className="flex items-center gap-1 px-3 py-1.5 border-b bg-muted/10 overflow-x-auto no-scrollbar">
        {TABS.map(t => {
          const TabIcon = t.icon
          return (
            <button
              key={t.id}
              type="button"
              onClick={() => { setActiveTab(t.id); setShowMore(false) }}
              className={cn(
                "inline-flex items-center gap-1 text-[9px] font-bold uppercase tracking-widest px-2 py-0.5 rounded border transition-colors shrink-0",
                activeTab === t.id
                  ? "bg-foreground/10 border-foreground/30 text-foreground"
                  : "border-transparent text-muted-foreground hover:text-foreground hover:border-muted"
              )}
            >
              <TabIcon className="w-2.5 h-2.5" />
              {t.label}
              {isEmpty && activeTab === t.id && (
                <span className="w-1.5 h-1.5 rounded-full bg-muted-foreground/30 ml-0.5" />
              )}
            </button>
          )
        })}
      </div>

      {/* Feed rows */}
      <div className="divide-y max-h-80 overflow-y-auto">
        {renderRows()}
      </div>
    </div>
  )
}
