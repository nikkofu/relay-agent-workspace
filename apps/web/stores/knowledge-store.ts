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
} from "@/types"

interface KnowledgeState {
  entities: KnowledgeEntity[]
  isLoading: boolean
  liveUpdate: KnowledgeUpdate | null
  channelKnowledge: ChannelKnowledgeRef[]
  channelKnowledgeId: string | null
  isLoadingChannelKnowledge: boolean
  channelSummary: ChannelKnowledgeSummary | null
  isLoadingChannelSummary: boolean
  entitySuggestions: EntitySuggestResult[]
  isLoadingSuggestions: boolean

  pushLiveUpdate: (update: KnowledgeUpdate) => void
  handleEntityCreated: (entity: KnowledgeEntity) => void
  handleEntityUpdated: (entity: KnowledgeEntity) => void
  fetchChannelKnowledge: (channelId: string) => Promise<ChannelKnowledgeRef[]>
  fetchChannelKnowledgeSummary: (channelId: string, days?: number, limit?: number) => Promise<ChannelKnowledgeSummary | null>
  suggestEntities: (q: string, channelId?: string, limit?: number) => Promise<EntitySuggestResult[]>
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
}

export const useKnowledgeStore = create<KnowledgeState>((set) => ({
  entities: [],
  isLoading: false,
  liveUpdate: null,
  channelKnowledge: [],
  channelKnowledgeId: null,
  isLoadingChannelKnowledge: false,
  channelSummary: null,
  isLoadingChannelSummary: false,
  entitySuggestions: [],
  isLoadingSuggestions: false,

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
}))
