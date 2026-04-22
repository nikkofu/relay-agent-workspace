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
}

export const useKnowledgeStore = create<KnowledgeState>((set) => ({
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
      await fetch(`${API_BASE_URL}/notifications/read`, {
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
}))
