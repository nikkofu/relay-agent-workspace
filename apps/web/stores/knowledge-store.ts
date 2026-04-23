import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"
import type {
  KnowledgeEntity,
  KnowledgeEntityRef,
  KnowledgeEntityLink,
  KnowledgeEvent,
  KnowledgeGraph,
  KnowledgeUpdate,
  ChannelKnowledgeRef,
  ChannelKnowledgeSummary,
  EntitySuggestResult,
  EntityHoverCard,
  MessageByEntityResult,
  ChannelKnowledgeDigest,
  Message,
  DigestSchedule,
  DigestScheduleInput,
  KnowledgeInboxItem,
  KnowledgeInboxDetail,
  KnowledgeInboxScope,
  DigestSchedulePreview,
  FollowedEntity,
  EntityTextMatch,
  FollowNotificationLevel,
  WorkspaceKnowledgeSettings,
  EntityActivity,
  TrendingEntity,
  FollowedEntityStats,
  SharedEntityLink,
  EntityBrief,
  WeeklyBrief,
  ActivityBackfillStatus,
  EntityAnswer,
  SharedWeeklyBriefLink,
  StaleBriefNotice,
  ComposeResponse,
  ComposeSuggestion,
  ComposeContextEntity,
  Citation,
  ComposeFeedbackValue,
  ComposeStreamingState,
  ComposeIntent,
  ComposeScope,
  ComposeFeedbackSummary,
  EntityAskHistoryItem,
  EntityAskHistoryResponse,
  EntityAskStreamingState,
  AISummary,
  ChannelAutoSummarySetting,
  ChannelAutoSummarizeInput,
  ChannelAutoSummarizeResponse,
  AIComposeActivity,
  AIComposeActivityDigest,
  AIComposeActivityDigestFilters,
  AIAutomationJob,
  EntityBriefAutomationState,
  AIScheduleBooking,
  AIScheduleBookingInput,
  KnowledgeAskRecentItem,
  KnowledgeAskRecentFilters,
  AIAutomationJobsFilters,
  AIAutomationJobsResponse,
} from "@/types"

interface KnowledgeState {
  entities: KnowledgeEntity[]
  isLoading: boolean
  liveUpdate: KnowledgeUpdate | null
  spikingEntityIds: Record<string, boolean>
  channelKnowledge: ChannelKnowledgeRef[]
  channelKnowledgeId: string | null
  isLoadingChannelKnowledge: boolean
  channelSummary: ChannelKnowledgeSummary | null
  isLoadingChannelSummary: boolean
  entitySuggestions: EntitySuggestResult[]
  isLoadingSuggestions: boolean
  digestSchedules: Record<string, DigestSchedule | null>
  knowledgeInbox: KnowledgeInboxItem[]
  knowledgeInboxScope: KnowledgeInboxScope
  knowledgeInboxUnreadCount: number
  isLoadingInbox: boolean
  followedEntities: FollowedEntity[]
  followedEntityIds: Record<string, boolean>
  isLoadingFollowed: boolean
  // ── Phase 59 ─────────────────────────────────────────────────────────────
  workspaceKnowledgeSettings: Record<string, WorkspaceKnowledgeSettings>
  trendingEntities: TrendingEntity[]
  trendingWorkspaceId: string | null
  trendingLastUpdatedAt: number | null
  isLoadingTrending: boolean
  entityActivity: Record<string, EntityActivity>
  // ── Phase 60 ─────────────────────────────────────────────────────────────
  followedStats: FollowedEntityStats | null
  // ── Phase 61 ─────────────────────────────────────────────────────────────
  entityBriefs: Record<string, EntityBrief>
  isGeneratingBrief: Record<string, boolean>
  weeklyBrief: WeeklyBrief | null
  isGeneratingWeeklyBrief: boolean
  backfillStatuses: Record<string, ActivityBackfillStatus>
  isBackfilling: Record<string, boolean>
  // ── Phase 63A ───────────────────────────────────────────────────────────
  entityAnswers: Record<string, EntityAnswer[]>
  isAskingEntity: Record<string, boolean>
  isSharingWeeklyBrief: boolean
  staleBriefs: Record<string, StaleBriefNotice>
  // ── Phase 63B ───────────────────────────────────────────────────────────
  composeResults: Record<string, ComposeResponse>
  isComposing: Record<string, boolean>
  // ── Phase 63C ───────────────────────────────────────────────────────────
  composeStreaming: Record<string, ComposeStreamingState | null>
  composeFeedback: Record<string, ComposeFeedbackValue>
  // ── Phase 63D ───────────────────────────────────────────────────────────
  composeFeedbackSummary: Record<string, ComposeFeedbackSummary>
  // ── Phase 63E: persisted entity Ask history + SSE streaming state ─────
  entityAskHistory: Record<string, EntityAskHistoryItem[]>
  isLoadingAskHistory: Record<string, boolean>
  entityAskStreaming: Record<string, EntityAskStreamingState | null>
  // ── Phase 63F: always-on channel auto-summarize ───────────────────────
  channelAutoSummarize: Record<string, ChannelAutoSummarizeResponse>
  isLoadingAutoSummarize: Record<string, boolean>
  isRunningAutoSummarize: Record<string, boolean>
  // Phase 63F + 63G: Rolling co-drafting activity log. Hydrated on demand from
  // `GET /api/v1/ai/compose/activity` and extended live by the
  // `knowledge.compose.suggestion.generated` WS event (preferring `payload.activity`).
  // Capped list so an observer surface (e.g. #agent-collab, channel info) can render
  // recent activity without unbounded growth. Newest first, de-duped by `compose_id`.
  // Shape matches backend `domain.AIComposeActivity` exactly so persisted rows and
  // live rows are interchangeable.
  composeSuggestionActivity: AIComposeActivity[]
  // Phase 63G: per-scope hydration flags so multiple mounts of ComposeActivityPane
  // don't re-fetch the same scope repeatedly. Keys are produced by `composeActivityScopeKey`.
  isLoadingComposeActivity: Record<string, boolean>
  hasHydratedComposeActivity: Record<string, boolean>

  // ── Phase 63H: compose activity digest analytics ──────────────────────
  // Keyed by a `digestCacheKey(filters)` so multiple strips (different
  // window/group_by) can coexist without collision. Empty default.
  composeActivityDigests: Record<string, AIComposeActivityDigest>
  isLoadingComposeActivityDigest: Record<string, boolean>

  // ── Phase 63H: entity brief automation (background regen job) ─────────
  // Keyed by entity id. `null` means "no job has ever been created for this
  // entity" (maps to backend `{ job: null }` response); `undefined` means
  // "not yet fetched".
  entityBriefAutomation: Record<string, AIAutomationJob | null>
  isLoadingEntityBriefAutomation: Record<string, boolean>

  // ── Phase 63H: AI schedule booking lifecycle ─────────────────────────
  // One flat list scoped to the current user (backend filters by
  // requested_by). Newest-first, deduped by booking id. Scoped UI (e.g. per
  // channel) filters this list client-side, mirroring Phase 63G compose
  // activity UX. The WS `schedule.event.booked|cancelled` handlers upsert
  // rows here so any open mount stays fresh.
  scheduleBookings: AIScheduleBooking[]
  isLoadingScheduleBookings: Record<string, boolean>
  hasHydratedScheduleBookings: Record<string, boolean>
  // Tracks the most recently booked compose_id so MessageComposer can flip
  // the calendar chip to a "Booked ✓" pill immediately on success.
  lastBookedComposeIds: Record<string, string> // compose_id → booking_id

  // ── Phase 63I: shared knowledge ask recent feed ───────────────────────
  // Newest-first cross-entity ask feed. Hydrated from GET /knowledge/ask/recent
  // and extended live by `knowledge.entity.ask.answered` WS events.
  // Keyed by workspace (one flat list per workspace; scoped mounts filter
  // client-side). Capped at 50 entries.
  knowledgeAskRecent: KnowledgeAskRecentItem[]
  isLoadingAskRecent: boolean
  hasHydratedAskRecent: Record<string, boolean> // workspaceId → boolean

  // ── Phase 63I: workspace automation audit ─────────────────────────────
  // Flat list of AIAutomationJob rows from GET /api/v1/ai/automation/jobs.
  // May cover multiple job_types. UI renders status filter tabs.
  automationJobs: AIAutomationJob[]
  isLoadingAutomationJobs: boolean
  automationJobsTotal: number

  pushLiveUpdate: (update: KnowledgeUpdate) => void
  handleEntityCreated: (entity: KnowledgeEntity) => void
  handleEntityUpdated: (entity: KnowledgeEntity) => void
  fetchChannelKnowledge: (channelId: string) => Promise<ChannelKnowledgeRef[]>
  fetchChannelKnowledgeSummary: (channelId: string, days?: number, limit?: number) => Promise<ChannelKnowledgeSummary | null>
  suggestEntities: (q: string, channelId?: string, limit?: number) => Promise<EntitySuggestResult[]>
  fetchEntityHover: (entityId: string, channelId?: string, days?: number) => Promise<EntityHoverCard | null>
  searchMessagesByEntity: (entityId: string, channelId?: string, limit?: number) => Promise<MessageByEntityResult[]>
  fetchChannelDigest: (channelId: string, window?: 'daily' | 'weekly' | 'monthly', limit?: number) => Promise<ChannelKnowledgeDigest | null>
  publishChannelDigest: (channelId: string, params: { window?: 'daily' | 'weekly' | 'monthly'; limit?: number; pin?: boolean }) => Promise<Message | null>
  fetchDigestSchedule: (channelId: string) => Promise<DigestSchedule | null>
  upsertDigestSchedule: (channelId: string, input: DigestScheduleInput) => Promise<DigestSchedule | null>
  deleteDigestSchedule: (channelId: string) => Promise<boolean>
  fetchKnowledgeInbox: (scope?: KnowledgeInboxScope, limit?: number) => Promise<KnowledgeInboxItem[]>
  fetchKnowledgeInboxItem: (id: string) => Promise<KnowledgeInboxDetail | null>
  previewDigestSchedule: (channelId: string, input: DigestScheduleInput) => Promise<DigestSchedulePreview | null>
  markInboxRead: (messageIds: string[]) => Promise<void>
  applyDigestPublished: (payload: { channel_id?: string; message?: Message; digest?: ChannelKnowledgeDigest }) => void
  markEntitySpiking: (entityId: string, ttlMs?: number) => void
  ingestEvent: (data: {
    entity_id: string
    event_type: string
    title: string
    body?: string
    actor_user_id?: string
    source_kind?: string
    source_ref?: string
  }) => Promise<KnowledgeEvent | null>
  fetchEntities: (q?: string) => Promise<KnowledgeEntity[]>
  createEntity: (data: { title: string; kind: string; summary?: string; tags?: string[] }) => Promise<KnowledgeEntity | null>
  fetchEntity: (id: string) => Promise<KnowledgeEntity | null>
  updateEntity: (id: string, data: Partial<KnowledgeEntity>) => Promise<KnowledgeEntity | null>
  fetchEntityRefs: (id: string) => Promise<KnowledgeEntityRef[]>
  addEntityRef: (id: string, data: { source_kind: string; source_id: string; snippet?: string }) => Promise<KnowledgeEntityRef | null>
  fetchEntityTimeline: (id: string) => Promise<KnowledgeEvent[]>
  addEntityEvent: (id: string, data: { event_kind: string; title: string; description?: string; occurred_at?: string }) => Promise<KnowledgeEvent | null>
  fetchEntityLinks: (id: string) => Promise<KnowledgeEntityLink[]>
  createLink: (data: { from_entity_id: string; to_entity_id: string; rel: string }) => Promise<KnowledgeEntityLink | null>
  fetchEntityGraph: (id: string) => Promise<KnowledgeGraph | null>
  fetchFollowedEntities: () => Promise<FollowedEntity[]>
  followEntity: (entityId: string) => Promise<boolean>
  unfollowEntity: (entityId: string) => Promise<boolean>
  updateFollowNotificationLevel: (followId: string, entityId: string, level: FollowNotificationLevel) => Promise<boolean>
  bulkUpdateFollowNotificationLevel: (entityIds: string[], level: FollowNotificationLevel) => Promise<boolean>
  matchEntitiesInText: (workspaceId: string, text: string, limit?: number) => Promise<EntityTextMatch[]>
  // ── Phase 59 ─────────────────────────────────────────────────────────────
  fetchWorkspaceKnowledgeSettings: (workspaceId: string) => Promise<WorkspaceKnowledgeSettings | null>
  updateWorkspaceKnowledgeSettings: (workspaceId: string, spikeThreshold: number, spikeCooldownMinutes: number) => Promise<WorkspaceKnowledgeSettings | null>
  fetchEntityActivity: (entityId: string, days?: number) => Promise<EntityActivity | null>
  fetchTrendingEntities: (workspaceId: string, days?: number, limit?: number) => Promise<TrendingEntity[]>
  // ── Phase 60 ─────────────────────────────────────────────────────────────
  fetchFollowedStats: () => Promise<FollowedEntityStats | null>
  shareEntity: (entityId: string) => Promise<SharedEntityLink | null>
  applyTrendingChanged: (payload: { workspace_id: string; days: number; items: TrendingEntity[] }) => void
  // ── Phase 61 ─────────────────────────────────────────────────────────────
  generateEntityBrief: (entityId: string, force?: boolean) => Promise<EntityBrief | null>
  generateWeeklyBrief: (workspaceId: string, force?: boolean) => Promise<WeeklyBrief | null>
  fetchBackfillStatus: (entityId: string) => Promise<ActivityBackfillStatus | null>
  triggerBackfill: (entityId: string) => Promise<{ refs_created: number; duration_ms: number } | null>
  applyFollowedStatsChanged: (stats: FollowedEntityStats) => void
  // ── Phase 62 ─────────────────────────────────────────────────────────────
  fetchEntityBrief: (entityId: string) => Promise<EntityBrief | null>
  fetchWeeklyBrief: (workspaceId: string) => Promise<WeeklyBrief | null>
  applyEntityBriefGenerated: (brief: EntityBrief) => void
  applyNotificationsBulkRead: (itemIds: string[]) => void
  // ── Phase 63A ───────────────────────────────────────────────────────────
  askEntity: (entityId: string, question: string) => Promise<EntityAnswer | null>
  shareWeeklyBrief: (briefId: string) => Promise<SharedWeeklyBriefLink | null>
  applyEntityBriefChanged: (notice: StaleBriefNotice) => void
  clearEntityAnswers: (entityId: string) => void
  // ── Phase 63B/D ───────────────────────────────────────────────────────────
  suggestCompose: (scope: ComposeScope, draft: string, intent?: ComposeIntent, limit?: number) => Promise<ComposeResponse | null>
  clearComposeResult: (scope: ComposeScope) => void
  // ── Phase 63C/D ──────────────────────────────────────────────────────────
  suggestComposeStream: (scope: ComposeScope, draft: string, intent?: ComposeIntent, limit?: number) => Promise<ComposeResponse | null>
  sendComposeFeedback: (composeId: string, args: { scope: ComposeScope; feedback: ComposeFeedbackValue; suggestionText?: string; provider?: string; model?: string }) => Promise<boolean>
  // ── Phase 63D ───────────────────────────────────────────────────────────
  fetchComposeFeedbackSummary: (composeId: string) => Promise<ComposeFeedbackSummary | null>
  // ── Phase 63E ───────────────────────────────────────────────────────────
  fetchEntityAskHistory: (entityId: string, limit?: number) => Promise<EntityAskHistoryResponse | null>
  askEntityStream: (entityId: string, question: string) => Promise<EntityAnswer | null>
  // ── Phase 63F ───────────────────────────────────────────────────────────
  fetchChannelAutoSummarize: (channelId: string) => Promise<ChannelAutoSummarizeResponse | null>
  updateChannelAutoSummarize: (channelId: string, input: ChannelAutoSummarizeInput) => Promise<ChannelAutoSummarySetting | null>
  runChannelAutoSummarize: (channelId: string, input?: ChannelAutoSummarizeInput) => Promise<ChannelAutoSummarizeResponse | null>
  applyChannelSummaryUpdated: (payload: { channel_id: string; workspace_id?: string; reason?: string; summary?: AISummary | null; setting?: ChannelAutoSummarySetting }) => void
  // Phase 63G: receives the full ws `knowledge.compose.suggestion.generated` payload.
  // Prefers the persisted `activity` row from the server and falls back to synthesizing
  // one from `compose` if the backend ever emits compose-only (defensive parity).
  applyComposeSuggestionGenerated: (payload: { compose?: ComposeResponse; activity?: AIComposeActivity }) => void
  // Phase 63G: GET /api/v1/ai/compose/activity — hydrates the activity log
  // scoped by channel / DM / workspace / intent. Merges into the single shared
  // `composeSuggestionActivity` list so WS appends stay consistent.
  fetchComposeActivity: (filters: { channelId?: string; dmId?: string; workspaceId?: string; intent?: string; limit?: number }) => Promise<AIComposeActivity[] | null>

  // ── Phase 63H actions ───────────────────────────────────────────────────

  // GET /api/v1/ai/compose/activity/digest — counts-by-(intent|user|channel|
  // dm|workspace|provider|model) over a rolling window. Requires exactly one
  // scope (workspace/channel/dm). Cached per `digestCacheKey(filters)`.
  fetchComposeActivityDigest: (filters: AIComposeActivityDigestFilters) => Promise<AIComposeActivityDigest | null>

  // GET /api/v1/knowledge/entities/:id/brief/automation
  fetchEntityBriefAutomation: (entityId: string) => Promise<EntityBriefAutomationState | null>
  // POST /api/v1/knowledge/entities/:id/brief/automation/run — queues
  // (idempotent: if one is already pending/running, server returns the
  // existing job with HTTP 200 instead of 202).
  runEntityBriefAutomation: (entityId: string) => Promise<AIAutomationJob | null>
  // POST /api/v1/knowledge/entities/:id/brief/automation/retry — only valid
  // when the latest job is in `failed` / `cancelled` state; server returns
  // 409 otherwise.
  retryEntityBriefAutomation: (entityId: string) => Promise<AIAutomationJob | null>
  // WS fan-in for knowledge.entity.brief.regen.{queued,started,failed}.
  // Unified handler so consumers don't have to special-case per-event.
  applyEntityBriefAutomationEvent: (eventType: string, payload: { job?: AIAutomationJob; entity?: KnowledgeEntity; reason?: string }) => void

  // POST /api/v1/ai/schedule/book — creates a booking from a ComposeProposedSlot
  bookAISchedule: (input: AIScheduleBookingInput) => Promise<AIScheduleBooking | null>
  // GET /api/v1/ai/schedule/bookings?channel_id|dm_id
  fetchAIScheduleBookings: (filters?: { channelId?: string; dmId?: string }) => Promise<AIScheduleBooking[] | null>
  // POST /api/v1/ai/schedule/bookings/:id/cancel (idempotent on server)
  cancelAIScheduleBooking: (bookingId: string) => Promise<AIScheduleBooking | null>
  // WS fan-in for schedule.event.{booked,cancelled} — upserts the booking
  // and, on booked events, stamps `lastBookedComposeIds` for the composer.
  applyScheduleBookingEvent: (eventType: string, payload: { booking?: AIScheduleBooking }) => void

  // ── Phase 63I actions ────────────────────────────────────────────────

  // GET /api/v1/knowledge/ask/recent — cross-entity ask feed.
  fetchKnowledgeAskRecent: (filters: KnowledgeAskRecentFilters) => Promise<KnowledgeAskRecentItem[] | null>
  // WS fan-in for knowledge.entity.ask.answered — prepends the new item to
  // the recent feed so open panes update without a refresh.
  applyEntityAskAnswered: (payload: { item?: KnowledgeAskRecentItem }) => void

  // GET /api/v1/ai/automation/jobs — workspace automation audit list.
  fetchAutomationJobs: (filters: AIAutomationJobsFilters) => Promise<AIAutomationJobsResponse | null>
}

// Phase 63D: normalize a compose scope into a stable key used for
// `composeResults`, `isComposing`, and `composeStreaming` maps.
// Channel scopes keep threadId for nested-thread independence.
function composeScopeKey(scope: ComposeScope): string | null {
  if (scope.dmId) return `dm:${scope.dmId}`
  if (scope.channelId) return `ch:${scope.channelId}:${scope.threadId || ''}`
  return null
}

// Phase 63G: normalize activity filters into a stable hydration-tracking key.
// Keeps hydration state granular so a workspace-wide mount doesn't clobber a
// per-channel hydration flag and vice versa.
function composeActivityScopeKey(filters: {
  channelId?: string
  dmId?: string
  workspaceId?: string
  intent?: string
}): string {
  const parts: string[] = []
  if (filters.channelId) parts.push(`ch:${filters.channelId}`)
  if (filters.dmId) parts.push(`dm:${filters.dmId}`)
  if (filters.workspaceId) parts.push(`ws:${filters.workspaceId}`)
  if (filters.intent) parts.push(`intent:${filters.intent}`)
  return parts.length > 0 ? parts.join('|') : 'all'
}

// Phase 63H: normalize a digest filter set into a stable cache key. Includes
// window + group_by so different strips (e.g. 24h-by-user vs 7d-by-intent)
// get independent cache slots and don't overwrite each other.
function digestCacheKey(filters: AIComposeActivityDigestFilters): string {
  const scope = filters.channelId
    ? `ch:${filters.channelId}`
    : filters.dmId
      ? `dm:${filters.dmId}`
      : filters.workspaceId
        ? `ws:${filters.workspaceId}`
        : 'none'
  const win = filters.window || '24h'
  const groupBy = filters.groupBy || 'intent'
  const intent = filters.intent ? `|intent:${filters.intent}` : ''
  const custom = filters.window === 'custom' ? `|${filters.startAt || ''}-${filters.endAt || ''}` : ''
  return `${scope}|${win}|${groupBy}${intent}${custom}`
}

// Phase 63H: scope key for hydration tracking of GET /ai/schedule/bookings.
function scheduleBookingsScopeKey(filters?: { channelId?: string; dmId?: string }): string {
  if (!filters) return 'all'
  if (filters.channelId) return `ch:${filters.channelId}`
  if (filters.dmId) return `dm:${filters.dmId}`
  return 'all'
}

// Phase 63E: Convert a persisted history row into an EntityAnswer-shaped object
// so the UI can render both fresh answers and hydrated history in the same list.
// History rows don't carry `citations[]`, only `citation_count`, which we preserve.
function historyItemToEntityAnswer(item: EntityAskHistoryItem, entity: KnowledgeEntity): EntityAnswer {
  return {
    entity,
    question: item.question,
    answer: item.answer,
    reasoning: item.reasoning,
    provider: item.provider,
    model: item.model,
    answered_at: item.answered_at,
    citations: [],
    citation_count: item.citation_count,
    history_id: item.id,
  }
}

export const useKnowledgeStore = create<KnowledgeState>((set, get) => ({
  entities: [],
  isLoading: false,
  liveUpdate: null,
  spikingEntityIds: {},
  channelKnowledge: [],
  channelKnowledgeId: null,
  isLoadingChannelKnowledge: false,
  channelSummary: null,
  isLoadingChannelSummary: false,
  entitySuggestions: [],
  isLoadingSuggestions: false,
  digestSchedules: {},
  knowledgeInbox: [],
  knowledgeInboxScope: 'all' as KnowledgeInboxScope,
  knowledgeInboxUnreadCount: 0,
  isLoadingInbox: false,
  followedEntities: [],
  followedEntityIds: {},
  isLoadingFollowed: false,
  workspaceKnowledgeSettings: {},
  trendingEntities: [],
  trendingWorkspaceId: null,
  trendingLastUpdatedAt: null,
  isLoadingTrending: false,
  entityActivity: {},
  followedStats: null,
  entityBriefs: {},
  isGeneratingBrief: {},
  weeklyBrief: null,
  isGeneratingWeeklyBrief: false,
  backfillStatuses: {},
  isBackfilling: {},
  entityAnswers: {},
  isAskingEntity: {},
  isSharingWeeklyBrief: false,
  staleBriefs: {},
  composeResults: {},
  isComposing: {},
  composeStreaming: {},
  composeFeedback: {},
  composeFeedbackSummary: {},
  entityAskHistory: {},
  isLoadingAskHistory: {},
  entityAskStreaming: {},
  channelAutoSummarize: {},
  isLoadingAutoSummarize: {},
  isRunningAutoSummarize: {},
  composeSuggestionActivity: [],
  isLoadingComposeActivity: {},
  hasHydratedComposeActivity: {},
  // Phase 63H
  composeActivityDigests: {},
  isLoadingComposeActivityDigest: {},
  entityBriefAutomation: {},
  isLoadingEntityBriefAutomation: {},
  scheduleBookings: [],
  isLoadingScheduleBookings: {},
  hasHydratedScheduleBookings: {},
  lastBookedComposeIds: {},
  // Phase 63I
  knowledgeAskRecent: [],
  isLoadingAskRecent: false,
  hasHydratedAskRecent: {},
  automationJobs: [],
  isLoadingAutomationJobs: false,
  automationJobsTotal: 0,

  pushLiveUpdate: (update) => set({ liveUpdate: update }),

  handleEntityCreated: (entity) =>
    set(state => ({
      entities: state.entities.some(e => e.id === entity.id)
        ? state.entities
        : [entity, ...state.entities],
      liveUpdate: { type: 'entity.created', entityId: entity.id, payload: entity, ts: Date.now() },
    })),

  handleEntityUpdated: (entity) =>
    set(state => ({
      entities: state.entities.map(e => e.id === entity.id ? entity : e),
      liveUpdate: { type: 'entity.updated', entityId: entity.id, payload: entity, ts: Date.now() },
    })),

  fetchChannelKnowledgeSummary: async (channelId, days = 7, limit = 5) => {
    set({ isLoadingChannelSummary: true })
    try {
      const res = await fetch(`${API_BASE_URL}/channels/${channelId}/knowledge/summary?days=${days}&limit=${limit}`)
      if (!res.ok) { set({ isLoadingChannelSummary: false }); return null }
      const data = await res.json()
      const summary: ChannelKnowledgeSummary = data.summary || data
      set({ channelSummary: summary, isLoadingChannelSummary: false })
      return summary
    } catch (error) {
      console.error("Failed to fetch channel knowledge summary:", error)
      set({ isLoadingChannelSummary: false })
      return null
    }
  },

  fetchEntityHover: async (entityId, channelId, days = 7) => {
    try {
      const params = new URLSearchParams({ days: String(days) })
      if (channelId) params.set('channel_id', channelId)
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${entityId}/hover?${params}`)
      if (!res.ok) return null
      const data = await res.json()
      return (data.hover || data) as EntityHoverCard
    } catch (error) {
      console.error("Failed to fetch entity hover:", error)
      return null
    }
  },

  searchMessagesByEntity: async (entityId, channelId, limit = 30) => {
    try {
      const params = new URLSearchParams({ entity_id: entityId, limit: String(limit) })
      if (channelId) params.set('channel_id', channelId)
      const res = await fetch(`${API_BASE_URL}/search/messages/by-entity?${params}`)
      if (!res.ok) return []
      const data = await res.json()
      return (data.results || data.messages || []) as MessageByEntityResult[]
    } catch (error) {
      console.error("Failed to search messages by entity:", error)
      return []
    }
  },

  fetchChannelDigest: async (channelId, window = 'weekly', limit = 5) => {
    try {
      const params = new URLSearchParams({ window, limit: String(limit) })
      const res = await fetch(`${API_BASE_URL}/channels/${channelId}/knowledge/digest?${params}`)
      if (!res.ok) return null
      const data = await res.json()
      return (data.digest || data) as ChannelKnowledgeDigest
    } catch (error) {
      console.error("Failed to fetch channel digest:", error)
      return null
    }
  },

  fetchDigestSchedule: async (channelId) => {
    try {
      const res = await fetch(`${API_BASE_URL}/channels/${channelId}/knowledge/digest/schedule`)
      if (res.status === 404) {
        set(state => ({ digestSchedules: { ...state.digestSchedules, [channelId]: null } }))
        return null
      }
      if (!res.ok) return null
      const data = await res.json()
      const schedule = (data.schedule || data) as DigestSchedule
      set(state => ({ digestSchedules: { ...state.digestSchedules, [channelId]: schedule } }))
      return schedule
    } catch (error) {
      console.error("Failed to fetch digest schedule:", error)
      return null
    }
  },

  upsertDigestSchedule: async (channelId, input) => {
    try {
      const res = await fetch(`${API_BASE_URL}/channels/${channelId}/knowledge/digest/schedule`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input),
      })
      if (!res.ok) {
        toast.error(`Failed to save schedule (${res.status})`)
        return null
      }
      const data = await res.json()
      const schedule = (data.schedule || data) as DigestSchedule
      set(state => ({ digestSchedules: { ...state.digestSchedules, [channelId]: schedule } }))
      toast.success(input.is_enabled ? "Digest schedule saved" : "Digest schedule disabled")
      return schedule
    } catch (error) {
      console.error("Failed to save digest schedule:", error)
      toast.error("Failed to save digest schedule")
      return null
    }
  },

  deleteDigestSchedule: async (channelId) => {
    try {
      const res = await fetch(`${API_BASE_URL}/channels/${channelId}/knowledge/digest/schedule`, {
        method: 'DELETE',
      })
      if (!res.ok && res.status !== 404) {
        toast.error(`Failed to remove schedule (${res.status})`)
        return false
      }
      set(state => ({ digestSchedules: { ...state.digestSchedules, [channelId]: null } }))
      toast.success("Digest schedule removed")
      return true
    } catch (error) {
      console.error("Failed to delete digest schedule:", error)
      return false
    }
  },

  fetchKnowledgeInbox: async (scope = 'all', limit = 30) => {
    set({ isLoadingInbox: true, knowledgeInboxScope: scope })
    try {
      const params = new URLSearchParams({ scope, limit: String(limit) })
      const res = await fetch(`${API_BASE_URL}/knowledge/inbox?${params}`)
      if (!res.ok) { set({ isLoadingInbox: false }); return [] }
      const data = await res.json()
      const items: KnowledgeInboxItem[] = data.items || data.inbox || []
      const unread = items.filter(i => !i.is_read).length
      set({ knowledgeInbox: items, knowledgeInboxUnreadCount: unread, isLoadingInbox: false })
      return items
    } catch (error) {
      console.error("Failed to fetch knowledge inbox:", error)
      set({ isLoadingInbox: false })
      return []
    }
  },

  markInboxRead: async (messageIds) => {
    if (messageIds.length === 0) return
    const itemIds = messageIds.map(id => `knowledge-digest-${id}`)
    // Optimistic local update
    set(state => {
      const idSet = new Set(messageIds)
      const next = state.knowledgeInbox.map(item =>
        idSet.has(item.message?.id) ? { ...item, is_read: true } : item
      )
      return { knowledgeInbox: next, knowledgeInboxUnreadCount: next.filter(i => !i.is_read).length }
    })
    try {
      // Phase 62: atomic bulk read endpoint (de-duplicates + single transaction + broadcasts notifications.bulk_read)
      await fetch(`${API_BASE_URL}/notifications/bulk-read`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ item_ids: itemIds }),
      })
    } catch (error) {
      console.error("Failed to mark inbox items read:", error)
    }
  },

  applyDigestPublished: (payload) => {
    // Invalidate channel summary if applicable; inbox will be re-fetched by caller.
    const channelId = payload?.channel_id
    if (!channelId) return
    set(() => ({
      liveUpdate: { type: 'digest.published', entityId: channelId, payload, ts: Date.now() },
    }))
  },

  publishChannelDigest: async (channelId, { window = 'weekly', limit = 5, pin = true }) => {
    try {
      const res = await fetch(`${API_BASE_URL}/channels/${channelId}/knowledge/digest/publish`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ window, limit, pin }),
      })
      if (!res.ok) {
        toast.error(`Failed to publish digest (${res.status})`)
        return null
      }
      const data = await res.json()
      const message = (data.message || data) as Message
      toast.success(`Knowledge digest published${pin ? ' and pinned' : ''}`)
      return message
    } catch (error) {
      console.error("Failed to publish digest:", error)
      toast.error("Failed to publish digest")
      return null
    }
  },

  suggestEntities: async (q, channelId, limit = 8) => {
    if (!q.trim()) { set({ entitySuggestions: [] }); return [] }
    set({ isLoadingSuggestions: true })
    try {
      const params = new URLSearchParams({ q, limit: String(limit) })
      if (channelId) params.set('channel_id', channelId)
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/suggest?${params}`)
      if (!res.ok) { set({ isLoadingSuggestions: false }); return [] }
      const data = await res.json()
      const suggestions: EntitySuggestResult[] = data.entities || data.suggestions || []
      set({ entitySuggestions: suggestions, isLoadingSuggestions: false })
      return suggestions
    } catch (error) {
      console.error("Failed to suggest entities:", error)
      set({ isLoadingSuggestions: false })
      return []
    }
  },

  fetchChannelKnowledge: async (channelId) => {
    set({ isLoadingChannelKnowledge: true, channelKnowledgeId: channelId })
    try {
      const res = await fetch(`${API_BASE_URL}/channels/${channelId}/knowledge`)
      if (!res.ok) { set({ isLoadingChannelKnowledge: false }); return [] }
      const data = await res.json()
      const refs: ChannelKnowledgeRef[] = data.refs || data.knowledge || []
      set({ channelKnowledge: refs, isLoadingChannelKnowledge: false })
      return refs
    } catch (error) {
      console.error("Failed to fetch channel knowledge:", error)
      set({ isLoadingChannelKnowledge: false })
      return []
    }
  },

  ingestEvent: async (payload) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/events/ingest`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })
      if (!res.ok) { toast.error("Failed to ingest event"); return null }
      const data = await res.json()
      toast.success("Event ingested")
      return data.event || data
    } catch (error) {
      console.error("Failed to ingest event:", error)
      toast.error("Failed to ingest event")
      return null
    }
  },

  fetchEntities: async (q) => {
    set({ isLoading: true })
    try {
      const url = q
        ? `${API_BASE_URL}/knowledge/entities?q=${encodeURIComponent(q)}`
        : `${API_BASE_URL}/knowledge/entities`
      const res = await fetch(url)
      if (!res.ok) { set({ isLoading: false }); return [] }
      const data = await res.json()
      const entities: KnowledgeEntity[] = data.entities || data.items || []
      set({ entities, isLoading: false })
      return entities
    } catch (error) {
      console.error("Failed to fetch entities:", error)
      set({ isLoading: false })
      return []
    }
  },

  createEntity: async (payload) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })
      if (!res.ok) { toast.error("Failed to create entity"); return null }
      const data = await res.json()
      const entity: KnowledgeEntity = data.entity || data
      set(state => ({ entities: [entity, ...state.entities] }))
      toast.success("Entity created")
      return entity
    } catch (error) {
      console.error("Failed to create entity:", error)
      toast.error("Failed to create entity")
      return null
    }
  },

  fetchEntity: async (id) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${id}`)
      if (!res.ok) return null
      const data = await res.json()
      return data.entity || data
    } catch (error) {
      console.error("Failed to fetch entity:", error)
      return null
    }
  },

  updateEntity: async (id, payload) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })
      if (!res.ok) { toast.error("Failed to update entity"); return null }
      const data = await res.json()
      const updated: KnowledgeEntity = data.entity || data
      set(state => ({ entities: state.entities.map(e => e.id === id ? updated : e) }))
      toast.success("Entity updated")
      return updated
    } catch (error) {
      console.error("Failed to update entity:", error)
      toast.error("Failed to update entity")
      return null
    }
  },

  fetchEntityRefs: async (id) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${id}/refs`)
      if (!res.ok) return []
      const data = await res.json()
      return data.refs || []
    } catch (error) {
      console.error("Failed to fetch entity refs:", error)
      return []
    }
  },

  addEntityRef: async (id, payload) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${id}/refs`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })
      if (!res.ok) return null
      const data = await res.json()
      return data.ref || data
    } catch (error) {
      console.error("Failed to add entity ref:", error)
      return null
    }
  },

  fetchEntityTimeline: async (id) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${id}/timeline`)
      if (!res.ok) return []
      const data = await res.json()
      return data.events || data.timeline || []
    } catch (error) {
      console.error("Failed to fetch entity timeline:", error)
      return []
    }
  },

  addEntityEvent: async (id, payload) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${id}/events`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })
      if (!res.ok) return null
      const data = await res.json()
      return data.event || data
    } catch (error) {
      console.error("Failed to add entity event:", error)
      return null
    }
  },

  fetchEntityLinks: async (id) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${id}/links`)
      if (!res.ok) return []
      const data = await res.json()
      return data.links || []
    } catch (error) {
      console.error("Failed to fetch entity links:", error)
      return []
    }
  },

  createLink: async (payload) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/links`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })
      if (!res.ok) { toast.error("Failed to create link"); return null }
      const data = await res.json()
      toast.success("Link created")
      return data.link || data
    } catch (error) {
      console.error("Failed to create link:", error)
      toast.error("Failed to create link")
      return null
    }
  },

  fetchEntityGraph: async (id) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${id}/graph`)
      if (!res.ok) return null
      const data = await res.json()
      return data.graph || data
    } catch (error) {
      console.error("Failed to fetch entity graph:", error)
      return null
    }
  },

  // ── Phase 55: Knowledge Entity Follow + Composer Reverse Lookup ───────────
  fetchFollowedEntities: async () => {
    set({ isLoadingFollowed: true })
    try {
      const res = await fetch(`${API_BASE_URL}/users/me/knowledge/followed`)
      if (!res.ok) { set({ isLoadingFollowed: false }); return [] }
      const data = await res.json()
      const items: FollowedEntity[] = data.items || []
      const ids: Record<string, boolean> = {}
      items.forEach(i => { ids[i.entity.id] = true })
      set({ followedEntities: items, followedEntityIds: ids, isLoadingFollowed: false })
      return items
    } catch (error) {
      console.error("Failed to fetch followed entities:", error)
      set({ isLoadingFollowed: false })
      return []
    }
  },

  followEntity: async (entityId) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${entityId}/follow`, {
        method: 'POST',
      })
      if (!res.ok) { toast.error("Failed to follow entity"); return false }
      // Optimistically mark followed locally; refresh list
      set(state => ({
        followedEntityIds: { ...state.followedEntityIds, [entityId]: true },
      }))
      // Refresh the full list so the new follow appears in the Following tab
      try {
        const listRes = await fetch(`${API_BASE_URL}/users/me/knowledge/followed`)
        if (listRes.ok) {
          const data = await listRes.json()
          const items: FollowedEntity[] = data.items || []
          const ids: Record<string, boolean> = {}
          items.forEach(i => { ids[i.entity.id] = true })
          set({ followedEntities: items, followedEntityIds: ids })
        }
      } catch { /* non-fatal */ }
      toast.success("Following entity")
      return true
    } catch (error) {
      console.error("Failed to follow entity:", error)
      toast.error("Failed to follow entity")
      return false
    }
  },

  updateFollowNotificationLevel: async (followId, entityId, level) => {
    try {
      const res = await fetch(`${API_BASE_URL}/users/me/knowledge/followed/${followId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ notification_level: level }),
      })
      if (!res.ok) { toast.error("Failed to update alert preference"); return false }
      // Optimistically update local follow list
      set(state => ({
        followedEntities: state.followedEntities.map(f =>
          f.entity.id === entityId
            ? { ...f, follow: { ...f.follow, notification_level: level } }
            : f
        ),
      }))
      return true
    } catch (error) {
      console.error("Failed to update follow notification level:", error)
      toast.error("Failed to update alert preference")
      return false
    }
  },

  // ── Phase 59: Bulk follow notification level update ─────────────────────
  bulkUpdateFollowNotificationLevel: async (entityIds, level) => {
    if (entityIds.length === 0) return true
    try {
      const res = await fetch(`${API_BASE_URL}/users/me/knowledge/followed/bulk`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ entity_ids: entityIds, notification_level: level }),
      })
      if (!res.ok) { toast.error("Failed to update alert preferences"); return false }
      const set_ = new Set(entityIds)
      set(state => ({
        followedEntities: state.followedEntities.map(f =>
          set_.has(f.entity.id)
            ? { ...f, follow: { ...f.follow, notification_level: level } }
            : f
        ),
      }))
      return true
    } catch (error) {
      console.error("Failed to bulk update follow notification levels:", error)
      toast.error("Failed to update alert preferences")
      return false
    }
  },

  // ── Phase 59: Workspace knowledge settings ──────────────────────────────
  fetchWorkspaceKnowledgeSettings: async (workspaceId) => {
    try {
      const res = await fetch(`${API_BASE_URL}/workspace/settings?workspace_id=${encodeURIComponent(workspaceId)}`)
      if (!res.ok) return null
      const data = await res.json()
      const settings: WorkspaceKnowledgeSettings = data.settings
      if (!settings) return null
      set(state => ({ workspaceKnowledgeSettings: { ...state.workspaceKnowledgeSettings, [workspaceId]: settings } }))
      return settings
    } catch (error) {
      console.error("Failed to fetch workspace knowledge settings:", error)
      return null
    }
  },

  updateWorkspaceKnowledgeSettings: async (workspaceId, spikeThreshold, spikeCooldownMinutes) => {
    try {
      const res = await fetch(`${API_BASE_URL}/workspace/settings`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          workspace_id: workspaceId,
          spike_threshold: spikeThreshold,
          spike_cooldown_minutes: spikeCooldownMinutes,
        }),
      })
      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        toast.error(err?.error || "Failed to save workspace settings")
        return null
      }
      const data = await res.json()
      const settings: WorkspaceKnowledgeSettings = data.settings
      set(state => ({ workspaceKnowledgeSettings: { ...state.workspaceKnowledgeSettings, [workspaceId]: settings } }))
      return settings
    } catch (error) {
      console.error("Failed to update workspace knowledge settings:", error)
      toast.error("Failed to save workspace settings")
      return null
    }
  },

  // ── Phase 59: Entity activity sparkline data ────────────────────────────
  fetchEntityActivity: async (entityId, days = 30) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${entityId}/activity?days=${days}`)
      if (!res.ok) return null
      const data = await res.json()
      const activity: EntityActivity = data.activity
      if (!activity) return null
      set(state => ({ entityActivity: { ...state.entityActivity, [entityId]: activity } }))
      return activity
    } catch (error) {
      console.error("Failed to fetch entity activity:", error)
      return null
    }
  },

  // ── Phase 59: Trending entities for workspace ───────────────────────────
  fetchTrendingEntities: async (workspaceId, days = 7, limit = 5) => {
    set({ isLoadingTrending: true })
    try {
      const url = `${API_BASE_URL}/knowledge/trending?workspace_id=${encodeURIComponent(workspaceId)}&days=${days}&limit=${limit}`
      const res = await fetch(url)
      if (!res.ok) { set({ isLoadingTrending: false }); return [] }
      const data = await res.json()
      const items: TrendingEntity[] = data.items || []
      set({
        trendingEntities: items,
        trendingWorkspaceId: workspaceId,
        trendingLastUpdatedAt: Date.now(),
        isLoadingTrending: false,
      })
      return items
    } catch (error) {
      console.error("Failed to fetch trending entities:", error)
      set({ isLoadingTrending: false })
      return []
    }
  },

  // ── Phase 60: Followed stats (for Following Hub summary strip) ──────────
  fetchFollowedStats: async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/users/me/knowledge/followed/stats`)
      if (!res.ok) return null
      const data = await res.json()
      const stats: FollowedEntityStats = data.stats
      if (!stats) return null
      set({ followedStats: stats })
      return stats
    } catch (error) {
      console.error("Failed to fetch followed stats:", error)
      return null
    }
  },

  // ── Phase 60: Shareable entity deeplink ─────────────────────────────────
  shareEntity: async (entityId) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${entityId}/share`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: '{}',
      })
      if (!res.ok) {
        toast.error("Failed to generate share link")
        return null
      }
      const data = await res.json()
      const share: SharedEntityLink = data.share
      return share || null
    } catch (error) {
      console.error("Failed to share entity:", error)
      toast.error("Failed to generate share link")
      return null
    }
  },

  // ── Phase 60: Live trending update from websocket ───────────────────────
  applyTrendingChanged: (payload) => set(state => {
    // Only apply if the payload workspace matches what the UI is currently showing,
    // or if nothing has been loaded yet (first subscriber wins).
    if (state.trendingWorkspaceId && payload.workspace_id !== state.trendingWorkspaceId) {
      return state
    }
    return {
      ...state,
      trendingEntities: payload.items || [],
      trendingWorkspaceId: payload.workspace_id,
      trendingLastUpdatedAt: Date.now(),
    }
  }),

  unfollowEntity: async (entityId) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${entityId}/follow`, {
        method: 'DELETE',
      })
      if (!res.ok) { toast.error("Failed to unfollow entity"); return false }
      set(state => {
        const nextIds = { ...state.followedEntityIds }
        delete nextIds[entityId]
        return {
          followedEntityIds: nextIds,
          followedEntities: state.followedEntities.filter(f => f.entity.id !== entityId),
        }
      })
      toast.success("Unfollowed entity")
      return true
    } catch (error) {
      console.error("Failed to unfollow entity:", error)
      toast.error("Failed to unfollow entity")
      return false
    }
  },

  // ── Phase 57: Spike tracking ─────────────────────────────────────────────
  markEntitySpiking: (entityId, ttlMs = 5 * 60 * 1000) => {
    set(state => ({ spikingEntityIds: { ...state.spikingEntityIds, [entityId]: true } }))
    setTimeout(() => {
      set(state => {
        const next = { ...state.spikingEntityIds }
        delete next[entityId]
        return { spikingEntityIds: next }
      })
    }, ttlMs)
  },

  // ── Phase 56: Inbox Detail + Digest Preview ─────────────────────────────
  fetchKnowledgeInboxItem: async (id) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/inbox/${id}`)
      if (!res.ok) return null
      const data = await res.json()
      return (data.detail || data) as KnowledgeInboxDetail
    } catch (error) {
      console.error("Failed to fetch inbox item detail:", error)
      return null
    }
  },

  previewDigestSchedule: async (channelId, input) => {
    try {
      const res = await fetch(`${API_BASE_URL}/channels/${channelId}/knowledge/digest/preview-schedule`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input),
      })
      if (!res.ok) return null
      const data = await res.json()
      return (data.preview || data) as DigestSchedulePreview
    } catch (error) {
      console.error("Failed to preview digest schedule:", error)
      return null
    }
  },

  matchEntitiesInText: async (workspaceId, text, limit = 10) => {
    const trimmed = text.trim()
    if (!trimmed) return []
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/match-text`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ workspace_id: workspaceId, text: trimmed, limit }),
      })
      if (!res.ok) return []
      const data = await res.json()
      return (data.matches || []) as EntityTextMatch[]
    } catch (error) {
      console.error("Failed to match entities in text:", error)
      return []
    }
  },

  // ── Phase 61: Entity AI brief ────────────────────────────────────────────
  generateEntityBrief: async (entityId, force = false) => {
    set(state => ({ isGeneratingBrief: { ...state.isGeneratingBrief, [entityId]: true } }))
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${entityId}/brief`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ force }),
      })
      if (!res.ok) { toast.error("Failed to generate brief"); return null }
      const data = await res.json()
      const brief: EntityBrief = data.brief
      if (brief) {
        set(state => {
          const nextStale = { ...state.staleBriefs }
          delete nextStale[entityId]
          return {
            entityBriefs: { ...state.entityBriefs, [entityId]: brief },
            staleBriefs: nextStale,
          }
        })
        toast.success(force ? "Brief regenerated" : "Brief generated")
      }
      return brief || null
    } catch (error) {
      console.error("Failed to generate entity brief:", error)
      toast.error("Failed to generate brief")
      return null
    } finally {
      set(state => ({ isGeneratingBrief: { ...state.isGeneratingBrief, [entityId]: false } }))
    }
  },

  // ── Phase 61: Weekly brief ────────────────────────────────────────────────
  generateWeeklyBrief: async (workspaceId, force = false) => {
    set({ isGeneratingWeeklyBrief: true })
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/weekly-brief`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ workspace_id: workspaceId, force }),
      })
      if (!res.ok) { toast.error("Failed to generate weekly brief"); return null }
      const data = await res.json()
      const brief: WeeklyBrief = data.brief
      if (brief) {
        set({ weeklyBrief: brief })
        toast.success("Weekly knowledge brief ready")
      }
      return brief || null
    } catch (error) {
      console.error("Failed to generate weekly brief:", error)
      toast.error("Failed to generate weekly brief")
      return null
    } finally {
      set({ isGeneratingWeeklyBrief: false })
    }
  },

  // ── Phase 61: Activity backfill ───────────────────────────────────────────
  fetchBackfillStatus: async (entityId) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${entityId}/activity/backfill-status`)
      if (!res.ok) return null
      const data = await res.json()
      const status: ActivityBackfillStatus = data.status
      if (status) {
        set(state => ({ backfillStatuses: { ...state.backfillStatuses, [entityId]: status } }))
      }
      return status || null
    } catch (error) {
      console.error("Failed to fetch backfill status:", error)
      return null
    }
  },

  triggerBackfill: async (entityId) => {
    set(state => ({ isBackfilling: { ...state.isBackfilling, [entityId]: true } }))
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${entityId}/activity/backfill`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: '{}',
      })
      if (!res.ok) { toast.error("Backfill failed"); return null }
      const data = await res.json()
      const status: ActivityBackfillStatus | undefined = data.status
      const created = status?.created_ref_count ?? (Array.isArray(data.created_refs) ? data.created_refs.length : 0)
      toast.success(`Backfill complete — ${created} refs added`)
      if (status) {
        set(state => ({ backfillStatuses: { ...state.backfillStatuses, [entityId]: status } }))
      } else {
        // Fallback: refresh backfill status separately
        const statusRes = await fetch(`${API_BASE_URL}/knowledge/entities/${entityId}/activity/backfill-status`)
        if (statusRes.ok) {
          const sd = await statusRes.json()
          if (sd.status) {
            set(state => ({ backfillStatuses: { ...state.backfillStatuses, [entityId]: sd.status } }))
          }
        }
      }
      return { refs_created: created, duration_ms: 0 }
    } catch (error) {
      console.error("Failed to trigger backfill:", error)
      toast.error("Backfill failed")
      return null
    } finally {
      set(state => ({ isBackfilling: { ...state.isBackfilling, [entityId]: false } }))
    }
  },

  // ── Phase 61: Live followed-stats update from websocket ───────────────────
  applyFollowedStatsChanged: (stats) => set({ followedStats: stats }),

  // ── Phase 62: Cached entity brief (no LLM call) ─────────────────────
  fetchEntityBrief: async (entityId) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${entityId}/brief`)
      if (!res.ok) return null
      const data = await res.json()
      const brief: EntityBrief | null = data.brief
      if (brief) {
        set(state => ({ entityBriefs: { ...state.entityBriefs, [entityId]: brief } }))
      }
      return brief
    } catch (error) {
      console.error("Failed to fetch cached entity brief:", error)
      return null
    }
  },

  // ── Phase 62: Cached weekly brief (no LLM call) ──────────────────────
  fetchWeeklyBrief: async (workspaceId) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/weekly-brief?workspace_id=${encodeURIComponent(workspaceId)}`)
      if (!res.ok) return null
      const data = await res.json()
      const brief: WeeklyBrief | null = data.brief
      if (brief) set({ weeklyBrief: brief })
      return brief
    } catch (error) {
      console.error("Failed to fetch cached weekly brief:", error)
      return null
    }
  },

  // ── Phase 62: Live brief-generated update from websocket (multi-tab sync) ──
  applyEntityBriefGenerated: (brief) => set(state => ({
    entityBriefs: { ...state.entityBriefs, [brief.entity_id]: brief },
    isGeneratingBrief: { ...state.isGeneratingBrief, [brief.entity_id]: false },
  })),

  // ── Phase 63A: Grounded entity Q&A ───────────────────────────
  askEntity: async (entityId, question) => {
    const q = question.trim()
    if (!q) return null
    set(state => ({ isAskingEntity: { ...state.isAskingEntity, [entityId]: true } }))
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${entityId}/ask`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ question: q }),
      })
      if (!res.ok) {
        toast.error(`Ask failed (${res.status})`)
        return null
      }
      const data = await res.json()
      const answer: EntityAnswer | undefined = data.answer
      if (answer) {
        set(state => ({
          entityAnswers: {
            ...state.entityAnswers,
            [entityId]: [answer, ...(state.entityAnswers[entityId] || [])],
          },
        }))
      }
      return answer || null
    } catch (error) {
      console.error("Failed to ask entity:", error)
      toast.error("Ask failed")
      return null
    } finally {
      set(state => ({ isAskingEntity: { ...state.isAskingEntity, [entityId]: false } }))
    }
  },

  clearEntityAnswers: (entityId) => set(state => {
    const next = { ...state.entityAnswers }
    delete next[entityId]
    const nextHistory = { ...state.entityAskHistory }
    delete nextHistory[entityId]
    return { entityAnswers: next, entityAskHistory: nextHistory }
  }),

  // ── Phase 63E: Hydrate persisted entity Ask history ─────────────────
  // Also mirrors the history into `entityAnswers[entityId]` so the existing
  // Ask AI card can render persisted + fresh answers through the same
  // rendering path (via `historyItemToEntityAnswer`).
  fetchEntityAskHistory: async (entityId, limit) => {
    if (!entityId) return null
    set(state => ({ isLoadingAskHistory: { ...state.isLoadingAskHistory, [entityId]: true } }))
    try {
      const qs = typeof limit === 'number' && limit > 0 ? `?limit=${limit}` : ''
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${encodeURIComponent(entityId)}/ask/history${qs}`)
      if (!res.ok) {
        console.warn(`ask/history returned ${res.status}`)
        return null
      }
      const data = await res.json() as EntityAskHistoryResponse
      const items = Array.isArray(data?.items) ? data.items : []
      const entity = data?.entity
      set(state => {
        // Keep any fresher in-memory answers in `entityAnswers` at the top; append
        // history rows below them, deduped by `history_id` so the fresh answer
        // emitted by ask/stream doesn't render twice after hydration.
        const fresh = state.entityAnswers[entityId] || []
        const freshHistoryIds = new Set(fresh.map(a => a.history_id).filter(Boolean))
        const hydrated = entity
          ? items
              .filter(it => !freshHistoryIds.has(it.id))
              .map(it => historyItemToEntityAnswer(it, entity))
          : []
        return {
          entityAskHistory: { ...state.entityAskHistory, [entityId]: items },
          entityAnswers: { ...state.entityAnswers, [entityId]: [...fresh, ...hydrated] },
        }
      })
      return data
    } catch (error) {
      console.error('Failed to load entity ask history:', error)
      return null
    } finally {
      set(state => ({ isLoadingAskHistory: { ...state.isLoadingAskHistory, [entityId]: false } }))
    }
  },

  // ── Phase 63E: Streaming entity Ask via SSE with graceful fallback ─────
  // Progressively renders `answer.delta` tokens in `entityAskStreaming[entityId]`.
  // On `answer.done`, the full EntityAnswer (with citations[] + history_id) is
  // prepended to `entityAnswers[entityId]` so the Ask AI card snaps to a finalized
  // card. Falls back to sync `askEntity` on non-OK status / network error.
  askEntityStream: async (entityId, question) => {
    const q = question.trim()
    if (!entityId || !q) return null
    set(state => ({
      isAskingEntity: { ...state.isAskingEntity, [entityId]: true },
      entityAskStreaming: {
        ...state.entityAskStreaming,
        [entityId]: { entityId, requestId: '', question: q, text: '' },
      },
    }))
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${encodeURIComponent(entityId)}/ask/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ question: q }),
      })
      if (!res.ok || !res.body) {
        console.warn(`ask/stream returned ${res.status}; falling back to /ask`)
        return await get().askEntity(entityId, q)
      }

      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      let streamProvider = ''
      let streamModel = ''
      let streamText = ''
      let requestId = ''
      let finalAnswer: EntityAnswer | null = null

      const processEvent = (event: string, payload: Record<string, unknown>) => {
        if (event === 'start') {
          streamProvider = typeof payload.provider === 'string' ? payload.provider : ''
          streamModel = typeof payload.model === 'string' ? payload.model : ''
          requestId = typeof payload.request_id === 'string' ? payload.request_id : ''
          set(state => ({
            entityAskStreaming: {
              ...state.entityAskStreaming,
              [entityId]: { entityId, requestId, question: q, text: '', provider: streamProvider, model: streamModel },
            },
          }))
        } else if (event === 'answer.delta') {
          const delta = typeof payload.text_delta === 'string' ? payload.text_delta : ''
          streamText += delta
          set(state => ({
            entityAskStreaming: {
              ...state.entityAskStreaming,
              [entityId]: { entityId, requestId, question: q, text: streamText, provider: streamProvider, model: streamModel },
            },
          }))
        } else if (event === 'answer.done') {
          const answer = payload.answer as EntityAnswer | undefined
          const historyId = typeof payload.history_id === 'string' ? payload.history_id : undefined
          if (answer) {
            finalAnswer = { ...answer, history_id: historyId }
            set(state => ({
              entityAnswers: {
                ...state.entityAnswers,
                [entityId]: [finalAnswer as EntityAnswer, ...(state.entityAnswers[entityId] || [])],
              },
            }))
          }
        } else if (event === 'done') {
          set(state => ({
            entityAskStreaming: { ...state.entityAskStreaming, [entityId]: null },
          }))
        } else if (event === 'error') {
          const msg = typeof payload.message === 'string' ? payload.message : 'stream error'
          toast.error(`Ask stream failed: ${msg}`)
        }
      }

      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const parts = buffer.split('\n\n')
        buffer = parts.pop() || ''
        for (const part of parts) {
          if (!part.trim()) continue
          let evName = 'message'
          let dataRaw = ''
          for (const line of part.split('\n')) {
            if (line.startsWith('event: ')) evName = line.slice(7).trim()
            else if (line.startsWith('event:')) evName = line.slice(6).trim()
            else if (line.startsWith('data: ')) dataRaw += line.slice(6)
            else if (line.startsWith('data:')) dataRaw += line.slice(5)
          }
          try {
            const payload = dataRaw ? JSON.parse(dataRaw) : {}
            processEvent(evName, payload)
          } catch (e) {
            console.error('SSE parse error', e, part)
          }
        }
      }
      return finalAnswer
    } catch (error) {
      console.error('Failed to stream entity ask:', error)
      return await get().askEntity(entityId, q)
    } finally {
      set(state => ({
        isAskingEntity: { ...state.isAskingEntity, [entityId]: false },
        entityAskStreaming: { ...state.entityAskStreaming, [entityId]: null },
      }))
    }
  },

  // ── Phase 63F: Channel auto-summarize (always-on rolling summary) ─────
  // GET /channels/:id/knowledge/auto-summarize → { setting, summary }
  fetchChannelAutoSummarize: async (channelId) => {
    if (!channelId) return null
    set(state => ({ isLoadingAutoSummarize: { ...state.isLoadingAutoSummarize, [channelId]: true } }))
    try {
      const res = await fetch(`${API_BASE_URL}/channels/${encodeURIComponent(channelId)}/knowledge/auto-summarize`)
      if (!res.ok) {
        console.warn(`auto-summarize GET returned ${res.status}`)
        return null
      }
      const data = await res.json() as ChannelAutoSummarizeResponse
      if (data?.setting) {
        set(state => ({
          channelAutoSummarize: { ...state.channelAutoSummarize, [channelId]: data },
        }))
      }
      return data
    } catch (error) {
      console.error('Failed to load channel auto-summarize:', error)
      return null
    } finally {
      set(state => ({ isLoadingAutoSummarize: { ...state.isLoadingAutoSummarize, [channelId]: false } }))
    }
  },

  // PUT /channels/:id/knowledge/auto-summarize → { setting }
  // Persists the settings only (does not re-run the summary).
  updateChannelAutoSummarize: async (channelId, input) => {
    if (!channelId) return null
    try {
      const res = await fetch(`${API_BASE_URL}/channels/${encodeURIComponent(channelId)}/knowledge/auto-summarize`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input),
      })
      if (!res.ok) {
        toast.error(`Auto-summarize update failed (${res.status})`)
        return null
      }
      const data = await res.json() as { setting?: ChannelAutoSummarySetting }
      const setting = data?.setting
      if (setting) {
        set(state => {
          const prev = state.channelAutoSummarize[channelId]
          return {
            channelAutoSummarize: {
              ...state.channelAutoSummarize,
              [channelId]: { setting, summary: prev?.summary },
            },
          }
        })
      }
      return setting || null
    } catch (error) {
      console.error('Failed to update channel auto-summarize:', error)
      toast.error('Auto-summarize update failed')
      return null
    }
  },

  // POST /channels/:id/knowledge/auto-summarize → { setting, summary }
  // Runs a summary generation synchronously; backend also fires ws channel.summary.updated.
  runChannelAutoSummarize: async (channelId, input) => {
    if (!channelId) return null
    set(state => ({ isRunningAutoSummarize: { ...state.isRunningAutoSummarize, [channelId]: true } }))
    try {
      const res = await fetch(`${API_BASE_URL}/channels/${encodeURIComponent(channelId)}/knowledge/auto-summarize`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input || {}),
      })
      if (!res.ok) {
        toast.error(`Auto-summarize run failed (${res.status})`)
        return null
      }
      const data = await res.json() as ChannelAutoSummarizeResponse
      if (data?.setting) {
        set(state => ({
          channelAutoSummarize: { ...state.channelAutoSummarize, [channelId]: data },
        }))
      }
      toast.success('Channel summary generated')
      return data
    } catch (error) {
      console.error('Failed to run channel auto-summarize:', error)
      toast.error('Auto-summarize run failed')
      return null
    } finally {
      set(state => ({ isRunningAutoSummarize: { ...state.isRunningAutoSummarize, [channelId]: false } }))
    }
  },

  // Phase 63F: websocket channel.summary.updated handler.
  // Merges the latest setting + summary into channelAutoSummarize[channelId]
  // so any open tab re-renders without a fetch. Null payload fields are ignored
  // so partial updates (e.g. setting-only) don't clobber the other slot.
  applyChannelSummaryUpdated: (payload) => {
    if (!payload || !payload.channel_id) return
    const channelId = payload.channel_id
    set(state => {
      const prev = state.channelAutoSummarize[channelId]
      const setting = payload.setting || prev?.setting
      if (!setting) return state
      const summary = payload.summary !== undefined ? payload.summary : prev?.summary
      return {
        channelAutoSummarize: {
          ...state.channelAutoSummarize,
          [channelId]: { setting, summary: summary || null },
        },
      }
    })
  },

  // Phase 63F + 63G: websocket knowledge.compose.suggestion.generated handler.
  // Appends a capped, newest-first activity log entry. Since Phase 63G, the
  // backend emits `{ compose, activity }` where `activity` is the persisted
  // AIComposeActivity row; we prefer that row verbatim so the UI stays aligned
  // with what `GET /api/v1/ai/compose/activity` would return on refresh. If
  // `activity` is absent (defensive backwards-compat), we synthesize a row
  // from `compose` using the first suggestion id as `compose_id`.
  applyComposeSuggestionGenerated: (payload) => {
    if (!payload) return
    let row: AIComposeActivity | null = null
    if (payload.activity && payload.activity.compose_id) {
      row = payload.activity
    } else if (payload.compose) {
      const compose = payload.compose
      const composeId = compose.suggestions?.[0]?.id || ''
      if (!composeId) return
      row = {
        id: composeId, // synthetic, best-effort — backend sets a real UUID
        compose_id: composeId,
        workspace_id: '',
        channel_id: compose.channel_id,
        dm_id: compose.dm_id,
        thread_id: compose.thread_id,
        intent: compose.intent,
        suggestion_count: compose.suggestions.length,
        provider: compose.provider,
        model: compose.model,
        created_at: new Date().toISOString(),
      }
    }
    if (!row) return
    const composeId = row.compose_id
    set(state => {
      // Dedupe by compose_id so sync + stream paths (or client-initiated sync
      // followed by ws echo) don't double-count.
      const existing = state.composeSuggestionActivity.filter(e => e.compose_id !== composeId)
      return { composeSuggestionActivity: [row!, ...existing].slice(0, 100) }
    })
  },

  // Phase 63G: GET /api/v1/ai/compose/activity
  // Pulls newest-first activity rows filtered by channel/dm/workspace/intent.
  // Merges results into the single shared `composeSuggestionActivity` list so
  // the same list powers both historical display and live WS appends.
  fetchComposeActivity: async (filters) => {
    const scopeKey = composeActivityScopeKey(filters)
    set(state => ({ isLoadingComposeActivity: { ...state.isLoadingComposeActivity, [scopeKey]: true } }))
    try {
      const params = new URLSearchParams()
      if (filters.channelId) params.set('channel_id', filters.channelId)
      if (filters.dmId) params.set('dm_id', filters.dmId)
      if (filters.workspaceId) params.set('workspace_id', filters.workspaceId)
      if (filters.intent) params.set('intent', filters.intent)
      const lim = Math.max(1, Math.min(100, filters.limit ?? 50))
      params.set('limit', String(lim))
      const res = await fetch(`${API_BASE_URL}/ai/compose/activity?${params.toString()}`)
      if (!res.ok) {
        console.warn(`compose activity GET returned ${res.status}`)
        return null
      }
      const data = await res.json() as { items?: AIComposeActivity[] }
      const items = Array.isArray(data?.items) ? data.items : []
      set(state => {
        // Merge fetched rows with any WS-appended rows already present. Dedupe by compose_id.
        const seen = new Set<string>()
        const merged: AIComposeActivity[] = []
        for (const row of [...items, ...state.composeSuggestionActivity]) {
          const key = row.compose_id
          if (!key || seen.has(key)) continue
          seen.add(key)
          merged.push(row)
        }
        // Keep newest-first by created_at (falls back to insertion order for ties).
        merged.sort((a, b) => {
          const ta = Date.parse(a.created_at) || 0
          const tb = Date.parse(b.created_at) || 0
          return tb - ta
        })
        return {
          composeSuggestionActivity: merged.slice(0, 100),
          hasHydratedComposeActivity: { ...state.hasHydratedComposeActivity, [scopeKey]: true },
        }
      })
      return items
    } catch (error) {
      console.error('Failed to load compose activity:', error)
      return null
    } finally {
      set(state => ({ isLoadingComposeActivity: { ...state.isLoadingComposeActivity, [scopeKey]: false } }))
    }
  },

  // ── Phase 63A: Weekly brief share link ────────────────────────
  shareWeeklyBrief: async (briefId) => {
    if (!briefId) { toast.error("No brief snapshot to share"); return null }
    set({ isSharingWeeklyBrief: true })
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/weekly-brief/${briefId}/share`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      })
      if (!res.ok) {
        toast.error(`Share failed (${res.status})`)
        return null
      }
      const data = await res.json()
      const share: SharedWeeklyBriefLink | undefined = data.share
      if (share?.url) {
        try {
          await navigator.clipboard.writeText(share.url)
          toast.success("Weekly brief link copied to clipboard")
        } catch {
          toast.success("Weekly brief share link ready")
        }
      }
      return share || null
    } catch (error) {
      console.error("Failed to share weekly brief:", error)
      toast.error("Share failed")
      return null
    } finally {
      set({ isSharingWeeklyBrief: false })
    }
  },

  // ── Phase 63A: Brief stale invalidation (websocket-driven) ─────────
  applyEntityBriefChanged: (notice) => set(state => ({
    staleBriefs: { ...state.staleBriefs, [notice.entity_id]: notice },
  })),

  // ── Phase 63B/D: Grounded compose suggestions (sync) ─────────────
  // Phase 63D: scope is { channelId [+threadId] } or { dmId }; intent is reply | summarize | followup | schedule
  suggestCompose: async (scope, draft, intent, limit) => {
    const key = composeScopeKey(scope)
    if (!key) return null
    const activeIntent: ComposeIntent = intent || 'reply'
    set(state => ({ isComposing: { ...state.isComposing, [key]: true } }))
    try {
      const body: Record<string, unknown> = {
        intent: activeIntent,
        draft,
      }
      if (scope.dmId) {
        body.dm_id = scope.dmId
      } else if (scope.channelId) {
        body.channel_id = scope.channelId
        if (scope.threadId) body.thread_id = scope.threadId
      }
      if (typeof limit === 'number' && limit > 0) body.limit = limit
      const res = await fetch(`${API_BASE_URL}/ai/compose`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) {
        toast.error(`Compose failed (${res.status})`)
        return null
      }
      const data = await res.json()
      const compose: ComposeResponse | undefined = data.compose
      if (compose) {
        set(state => ({ composeResults: { ...state.composeResults, [key]: compose } }))
      }
      return compose || null
    } catch (error) {
      console.error("Failed to suggest compose:", error)
      toast.error("Compose failed")
      return null
    } finally {
      set(state => ({ isComposing: { ...state.isComposing, [key]: false } }))
    }
  },

  clearComposeResult: (scope) => {
    const key = composeScopeKey(scope)
    if (!key) return
    set(state => {
      const nextResults = { ...state.composeResults }
      delete nextResults[key]
      const nextStreaming = { ...state.composeStreaming }
      delete nextStreaming[key]
      return { composeResults: nextResults, composeStreaming: nextStreaming }
    })
  },

  // ── Phase 63C/D: Streaming compose via SSE with fallback ─────
  suggestComposeStream: async (scope, draft, intent, limit) => {
    const key = composeScopeKey(scope)
    if (!key) return null
    const activeIntent: ComposeIntent = intent || 'reply'
    set(state => ({
      isComposing: { ...state.isComposing, [key]: true },
      composeStreaming: { ...state.composeStreaming, [key]: null },
    }))
    try {
      const body: Record<string, unknown> = {
        intent: activeIntent,
        draft,
      }
      if (scope.dmId) {
        body.dm_id = scope.dmId
      } else if (scope.channelId) {
        body.channel_id = scope.channelId
        if (scope.threadId) body.thread_id = scope.threadId
      }
      if (typeof limit === 'number' && limit > 0) body.limit = limit

      const res = await fetch(`${API_BASE_URL}/ai/compose/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })

      // Fallback: if stream is unsupported (older server, error, no body), use sync compose
      if (!res.ok || !res.body) {
        console.warn(`compose/stream returned ${res.status}; falling back to /ai/compose`)
        return await get().suggestCompose(scope, draft, activeIntent, limit)
      }

      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      let streamProvider = ''
      let streamModel = ''
      let provisionalId = ''
      let streamText = ''
      const finalSuggestions: ComposeSuggestion[] = []
      let finalCitations: Citation[] = []
      let finalContextEntities: ComposeContextEntity[] = []

      const processEvent = (event: string, payload: Record<string, unknown>) => {
        if (event === 'start') {
          streamProvider = typeof payload.provider === 'string' ? payload.provider : ''
          streamModel = typeof payload.model === 'string' ? payload.model : ''
        } else if (event === 'suggestion.delta') {
          const delta = typeof payload.text_delta === 'string' ? payload.text_delta : ''
          const sid = typeof payload.suggestion_id === 'string' ? payload.suggestion_id : ''
          if (sid) provisionalId = sid
          streamText += delta
          set(state => ({
            composeStreaming: {
              ...state.composeStreaming,
              [key]: { suggestionId: provisionalId, text: streamText, index: 0 },
            },
          }))
        } else if (event === 'suggestion.done') {
          const sugg = payload.suggestion as ComposeSuggestion | undefined
          if (sugg) finalSuggestions.push(sugg)
          if (Array.isArray(payload.citations)) finalCitations = payload.citations as Citation[]
          if (Array.isArray(payload.context_entities)) finalContextEntities = payload.context_entities as ComposeContextEntity[]
        } else if (event === 'done') {
          const response: ComposeResponse = {
            channel_id: scope.channelId,
            dm_id: scope.dmId,
            thread_id: scope.threadId,
            intent: activeIntent,
            suggestions: finalSuggestions,
            citations: finalCitations,
            context_entities: finalContextEntities,
            provider: streamProvider,
            model: streamModel,
          }
          set(state => ({
            composeResults: { ...state.composeResults, [key]: response },
            composeStreaming: { ...state.composeStreaming, [key]: null },
          }))
        } else if (event === 'error') {
          const msg = typeof payload.message === 'string' ? payload.message : 'stream error'
          toast.error(`Compose stream failed: ${msg}`)
        }
      }

      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const parts = buffer.split('\n\n')
        buffer = parts.pop() || ''
        for (const part of parts) {
          if (!part.trim()) continue
          let evName = 'message'
          let dataRaw = ''
          for (const line of part.split('\n')) {
            if (line.startsWith('event: ')) evName = line.slice(7).trim()
            else if (line.startsWith('event:')) evName = line.slice(6).trim()
            else if (line.startsWith('data: ')) dataRaw += line.slice(6)
            else if (line.startsWith('data:')) dataRaw += line.slice(5)
          }
          try {
            const payload = dataRaw ? JSON.parse(dataRaw) : {}
            processEvent(evName, payload)
          } catch (e) {
            console.error('SSE parse error', e, part)
          }
        }
      }

      return get().composeResults[key] || null
    } catch (error) {
      console.error("Failed to stream compose:", error)
      // Fallback on network error
      return await get().suggestCompose(scope, draft, activeIntent, limit)
    } finally {
      set(state => ({
        isComposing: { ...state.isComposing, [key]: false },
        composeStreaming: { ...state.composeStreaming, [key]: null },
      }))
    }
  },

  // ── Phase 63C/D: Per-suggestion feedback capture (scope-aware) ────
  sendComposeFeedback: async (composeId, args) => {
    if (!composeId) return false
    try {
      const body: Record<string, unknown> = {
        intent: 'reply',
        feedback: args.feedback,
      }
      if (args.scope.dmId) {
        body.dm_id = args.scope.dmId
      } else if (args.scope.channelId) {
        body.channel_id = args.scope.channelId
        if (args.scope.threadId) body.thread_id = args.scope.threadId
      } else {
        console.warn('compose feedback: scope missing channel/dm id')
        return false
      }
      if (args.suggestionText) body.suggestion_text = args.suggestionText
      if (args.provider) body.provider = args.provider
      if (args.model) body.model = args.model
      const res = await fetch(`${API_BASE_URL}/ai/compose/${encodeURIComponent(composeId)}/feedback`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) {
        console.warn(`compose feedback returned ${res.status}`)
        return false
      }
      set(state => ({
        composeFeedback: { ...state.composeFeedback, [composeId]: args.feedback },
      }))
      // Refresh aggregate signal for this compose (best-effort, non-blocking)
      get().fetchComposeFeedbackSummary(composeId).catch(() => { /* best-effort */ })
      return true
    } catch (error) {
      console.error("Failed to send compose feedback:", error)
      return false
    }
  },

  // ── Phase 63D: Aggregate feedback signal per compose ─────────────
  fetchComposeFeedbackSummary: async (composeId) => {
    if (!composeId) return null
    try {
      const res = await fetch(`${API_BASE_URL}/ai/compose/${encodeURIComponent(composeId)}/feedback/summary`)
      if (!res.ok) {
        console.warn(`compose feedback summary returned ${res.status}`)
        return null
      }
      const data = await res.json()
      const summary: ComposeFeedbackSummary | undefined = data.summary
      if (summary) {
        set(state => ({
          composeFeedbackSummary: { ...state.composeFeedbackSummary, [composeId]: summary },
        }))
      }
      return summary || null
    } catch (error) {
      console.error("Failed to fetch compose feedback summary:", error)
      return null
    }
  },

  // ── Phase 63H: Compose activity digest analytics ─────────────────────────

  fetchComposeActivityDigest: async (filters) => {
    const key = digestCacheKey(filters)
    if (get().isLoadingComposeActivityDigest[key]) return null
    set(state => ({ isLoadingComposeActivityDigest: { ...state.isLoadingComposeActivityDigest, [key]: true } }))
    try {
      const params = new URLSearchParams()
      if (filters.channelId) params.set('channel_id', filters.channelId)
      else if (filters.dmId) params.set('dm_id', filters.dmId)
      else if (filters.workspaceId) params.set('workspace_id', filters.workspaceId)
      if (filters.window) params.set('window', filters.window)
      if (filters.startAt) params.set('start_at', filters.startAt)
      if (filters.endAt) params.set('end_at', filters.endAt)
      if (filters.intent) params.set('intent', filters.intent)
      if (filters.groupBy) params.set('group_by', filters.groupBy)
      if (filters.limit != null) params.set('limit', String(filters.limit))
      const res = await fetch(`${API_BASE_URL}/ai/compose/activity/digest?${params}`)
      if (!res.ok) {
        set(state => ({ isLoadingComposeActivityDigest: { ...state.isLoadingComposeActivityDigest, [key]: false } }))
        return null
      }
      const data = await res.json() as AIComposeActivityDigest
      set(state => ({
        composeActivityDigests: { ...state.composeActivityDigests, [key]: data },
        isLoadingComposeActivityDigest: { ...state.isLoadingComposeActivityDigest, [key]: false },
      }))
      return data
    } catch (err) {
      console.error('Failed to fetch compose activity digest:', err)
      set(state => ({ isLoadingComposeActivityDigest: { ...state.isLoadingComposeActivityDigest, [key]: false } }))
      return null
    }
  },

  // ── Phase 63H: Entity brief automation ───────────────────────────────────

  fetchEntityBriefAutomation: async (entityId) => {
    set(state => ({ isLoadingEntityBriefAutomation: { ...state.isLoadingEntityBriefAutomation, [entityId]: true } }))
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${encodeURIComponent(entityId)}/brief/automation`)
      if (!res.ok) {
        set(state => ({ isLoadingEntityBriefAutomation: { ...state.isLoadingEntityBriefAutomation, [entityId]: false } }))
        return null
      }
      const data = await res.json() as { job: AIAutomationJob | null; entity?: import('@/types').KnowledgeEntity }
      set(state => ({
        entityBriefAutomation: { ...state.entityBriefAutomation, [entityId]: data.job ?? null },
        isLoadingEntityBriefAutomation: { ...state.isLoadingEntityBriefAutomation, [entityId]: false },
      }))
      return data
    } catch (err) {
      console.error('Failed to fetch entity brief automation:', err)
      set(state => ({ isLoadingEntityBriefAutomation: { ...state.isLoadingEntityBriefAutomation, [entityId]: false } }))
      return null
    }
  },

  runEntityBriefAutomation: async (entityId) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${encodeURIComponent(entityId)}/brief/automation/run`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      })
      if (!res.ok) {
        toast.error('Failed to queue brief regeneration')
        return null
      }
      const data = await res.json() as { job: AIAutomationJob; entity?: import('@/types').KnowledgeEntity }
      set(state => ({ entityBriefAutomation: { ...state.entityBriefAutomation, [entityId]: data.job } }))
      return data.job
    } catch (err) {
      console.error('Failed to run entity brief automation:', err)
      toast.error('Failed to queue brief regeneration')
      return null
    }
  },

  retryEntityBriefAutomation: async (entityId) => {
    try {
      const res = await fetch(`${API_BASE_URL}/knowledge/entities/${encodeURIComponent(entityId)}/brief/automation/retry`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      })
      if (res.status === 409) {
        toast.error('Cannot retry: job is not in a failed state')
        return null
      }
      if (!res.ok) {
        toast.error('Failed to retry brief regeneration')
        return null
      }
      const data = await res.json() as { job: AIAutomationJob; entity?: import('@/types').KnowledgeEntity }
      set(state => ({ entityBriefAutomation: { ...state.entityBriefAutomation, [entityId]: data.job } }))
      return data.job
    } catch (err) {
      console.error('Failed to retry entity brief automation:', err)
      toast.error('Failed to retry brief regeneration')
      return null
    }
  },

  // Unified WS handler for knowledge.entity.brief.regen.{queued,started,failed}.
  // Upserts the job into the automation state map and shows a toast on failure.
  applyEntityBriefAutomationEvent: (eventType, payload) => {
    const job = payload.job
    if (!job?.scope_id) return
    set(state => ({
      entityBriefAutomation: { ...state.entityBriefAutomation, [job.scope_id]: job },
    }))
    if (eventType === 'knowledge.entity.brief.regen.failed' && payload.reason) {
      toast.error(`Brief automation failed: ${payload.reason}`)
    }
  },

  // ── Phase 63H: AI schedule booking lifecycle ─────────────────────────────

  bookAISchedule: async (input) => {
    try {
      const body: Record<string, unknown> = {
        compose_id: input.compose_id,
        title: input.title,
        description: input.description,
        provider: input.provider,
        slot: input.slot,
      }
      if (input.channel_id) body['channel_id'] = input.channel_id
      if (input.dm_id) body['dm_id'] = input.dm_id
      const res = await fetch(`${API_BASE_URL}/ai/schedule/book`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        toast.error((data as { error?: string }).error || 'Failed to book schedule slot')
        return null
      }
      const data = await res.json() as { booking: AIScheduleBooking }
      const booking = data.booking
      set(state => ({
        scheduleBookings: [booking, ...state.scheduleBookings.filter(b => b.id !== booking.id)],
        lastBookedComposeIds: { ...state.lastBookedComposeIds, [booking.compose_id]: booking.id },
      }))
      toast.success('Schedule slot booked ✓')
      return booking
    } catch (err) {
      console.error('Failed to book AI schedule:', err)
      toast.error('Failed to book schedule slot')
      return null
    }
  },

  fetchAIScheduleBookings: async (filters) => {
    const key = scheduleBookingsScopeKey(filters)
    if (get().isLoadingScheduleBookings[key]) return null
    set(state => ({ isLoadingScheduleBookings: { ...state.isLoadingScheduleBookings, [key]: true } }))
    try {
      const params = new URLSearchParams()
      if (filters?.channelId) params.set('channel_id', filters.channelId)
      if (filters?.dmId) params.set('dm_id', filters.dmId)
      const res = await fetch(`${API_BASE_URL}/ai/schedule/bookings?${params}`)
      if (!res.ok) {
        set(state => ({
          isLoadingScheduleBookings: { ...state.isLoadingScheduleBookings, [key]: false },
        }))
        return null
      }
      const data = await res.json() as { bookings: AIScheduleBooking[] }
      const fetched: AIScheduleBooking[] = data.bookings || []
      set(state => ({
        // Merge fetched rows with any rows already in state (WS-appended may be newer)
        scheduleBookings: Object.values(
          [...fetched, ...state.scheduleBookings].reduce<Record<string, AIScheduleBooking>>((acc, b) => {
            if (!acc[b.id] || new Date(b.updated_at) >= new Date(acc[b.id].updated_at)) acc[b.id] = b
            return acc
          }, {})
        ).sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()),
        isLoadingScheduleBookings: { ...state.isLoadingScheduleBookings, [key]: false },
        hasHydratedScheduleBookings: { ...state.hasHydratedScheduleBookings, [key]: true },
      }))
      return fetched
    } catch (err) {
      console.error('Failed to fetch schedule bookings:', err)
      set(state => ({ isLoadingScheduleBookings: { ...state.isLoadingScheduleBookings, [key]: false } }))
      return null
    }
  },

  cancelAIScheduleBooking: async (bookingId) => {
    try {
      const res = await fetch(`${API_BASE_URL}/ai/schedule/bookings/${encodeURIComponent(bookingId)}/cancel`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      })
      if (!res.ok) {
        toast.error('Failed to cancel booking')
        return null
      }
      const data = await res.json() as { booking: AIScheduleBooking }
      const booking = data.booking
      set(state => ({
        scheduleBookings: state.scheduleBookings.map(b => b.id === booking.id ? booking : b),
      }))
      return booking
    } catch (err) {
      console.error('Failed to cancel schedule booking:', err)
      toast.error('Failed to cancel booking')
      return null
    }
  },

  // Unified WS handler for schedule.event.{booked,cancelled}.
  // Upserts the booking into `scheduleBookings` and, on booked, stamps
  // `lastBookedComposeIds` so the calendar chip in the composer can flip.
  applyScheduleBookingEvent: (eventType, payload) => {
    const booking = payload.booking
    if (!booking?.id) return
    set(state => {
      const next = state.scheduleBookings.some(b => b.id === booking.id)
        ? state.scheduleBookings.map(b => b.id === booking.id ? booking : b)
        : [booking, ...state.scheduleBookings]
      const nextLastBooked = eventType === 'schedule.event.booked'
        ? { ...state.lastBookedComposeIds, [booking.compose_id]: booking.id }
        : state.lastBookedComposeIds
      return { scheduleBookings: next, lastBookedComposeIds: nextLastBooked }
    })
  },

  // ── Phase 63I: shared knowledge ask recent feed ───────────────────────────────
  fetchKnowledgeAskRecent: async (filters) => {
    const wsKey = filters.workspaceId || 'none'
    if (get().hasHydratedAskRecent[wsKey]) return get().knowledgeAskRecent
    set({ isLoadingAskRecent: true })
    try {
      const params = new URLSearchParams()
      if (filters.workspaceId) params.set('workspace_id', filters.workspaceId)
      if (filters.entityId) params.set('entity_id', filters.entityId)
      if (filters.userId) params.set('user_id', filters.userId)
      params.set('limit', String(filters.limit || 20))
      const res = await fetch(`${API_BASE_URL}/knowledge/ask/recent?${params.toString()}`)
      if (!res.ok) { set({ isLoadingAskRecent: false }); return null }
      const data = await res.json() as { items?: KnowledgeAskRecentItem[] }
      const items = Array.isArray(data?.items) ? data.items : []
      set(state => ({
        isLoadingAskRecent: false,
        hasHydratedAskRecent: { ...state.hasHydratedAskRecent, [wsKey]: true },
        // Merge with any WS-prepended rows, dedupe by id, newest first
        knowledgeAskRecent: (() => {
          const seen = new Set<string>()
          const merged: KnowledgeAskRecentItem[] = []
          for (const row of [...state.knowledgeAskRecent, ...items]) {
            if (!row.id || seen.has(row.id)) continue
            seen.add(row.id)
            merged.push(row)
          }
          return merged.sort((a, b) =>
            new Date(b.answered_at).getTime() - new Date(a.answered_at).getTime()
          ).slice(0, 50)
        })(),
      }))
      return get().knowledgeAskRecent
    } catch (err) {
      console.error('Failed to fetch knowledge ask recent:', err)
      set({ isLoadingAskRecent: false })
      return null
    }
  },

  applyEntityAskAnswered: (payload) => {
    const item = payload.item
    if (!item?.id) return
    set(state => ({
      knowledgeAskRecent: [item, ...state.knowledgeAskRecent.filter(r => r.id !== item.id)].slice(0, 50),
    }))
  },

  // ── Phase 63I: workspace automation audit ──────────────────────────────────
  fetchAutomationJobs: async (filters) => {
    set({ isLoadingAutomationJobs: true })
    try {
      const params = new URLSearchParams()
      if (filters.workspaceId) params.set('workspace_id', filters.workspaceId)
      if (filters.status) params.set('status', filters.status)
      if (filters.jobType) params.set('job_type', filters.jobType)
      if (filters.scopeType) params.set('scope_type', filters.scopeType)
      if (filters.scopeId) params.set('scope_id', filters.scopeId)
      params.set('limit', String(filters.limit || 20))
      const res = await fetch(`${API_BASE_URL}/ai/automation/jobs?${params.toString()}`)
      if (!res.ok) { set({ isLoadingAutomationJobs: false }); return null }
      const data = await res.json() as AIAutomationJobsResponse
      const items = Array.isArray(data?.items) ? data.items : []
      const total = typeof data?.total === 'number' ? data.total : items.length
      set({ isLoadingAutomationJobs: false, automationJobs: items, automationJobsTotal: total })
      return { items, total }
    } catch (err) {
      console.error('Failed to fetch automation jobs:', err)
      set({ isLoadingAutomationJobs: false })
      return null
    }
  },

  // ── Phase 62: Live bulk-read update from websocket (multi-tab inbox sync) ─
  applyNotificationsBulkRead: (itemIds) => set(state => {
    if (!itemIds || itemIds.length === 0) return state
    // item_ids follow the `knowledge-digest-<messageId>` convention for inbox items
    const msgIdSet = new Set<string>()
    for (const id of itemIds) {
      if (id.startsWith('knowledge-digest-')) {
        msgIdSet.add(id.slice('knowledge-digest-'.length))
      }
    }
    if (msgIdSet.size === 0) return state
    const next = state.knowledgeInbox.map(item =>
      item.message?.id && msgIdSet.has(item.message.id) ? { ...item, is_read: true } : item
    )
    return {
      ...state,
      knowledgeInbox: next,
      knowledgeInboxUnreadCount: next.filter(i => !i.is_read).length,
    }
  }),
}))
