import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"
import type { UnifiedActivityFeedItem, ActivityFeedFilters } from "@/types"

export interface ActivityItem {
  id: string
  type: 'mention' | 'reaction' | 'reply' | 'channel_join' | 'dm_message' | 'thread_reply' | 'list_completed' | 'tool_run' | 'file_uploaded'
  user: any
  channel?: any
  message?: any
  target?: string
  summary: string
  occurredAt: string
  isRead: boolean
}

interface ActivityState {
  activities: ActivityItem[]
  inboxItems: ActivityItem[]
  mentionItems: ActivityItem[]
  fetchActivities: () => Promise<void>
  fetchInbox: () => Promise<void>
  fetchMentions: () => Promise<void>
  markAsRead: (itemIds: string[]) => Promise<void>
  markAsReadLocally: (itemIds: string[]) => void
  appendMentionItem: (item: ActivityItem) => void
  // Phase 64A: unified activity feed
  unifiedFeedItems: UnifiedActivityFeedItem[]
  isLoadingUnifiedFeed: boolean
  unifiedFeedCursor: string | null
  hasMoreUnifiedFeed: boolean
  fetchUnifiedFeed: (filters: ActivityFeedFilters) => Promise<void>
  appendUnifiedFeedItem: (item: UnifiedActivityFeedItem) => void
}

const mapActivity = (a: any): ActivityItem => ({
  ...a,
  id: a.id || `${a.type}-${a.message?.id || 'no-msg'}-${a.user?.id || 'no-user'}-${a.target || 'no-target'}-${a.occurred_at}`,
  occurredAt: a.occurred_at,
  isRead: a.is_read
})

const deduplicateItems = (items: ActivityItem[]): ActivityItem[] => {
  const seen = new Set();
  return items.filter(item => {
    if (seen.has(item.id)) return false;
    seen.add(item.id);
    return true;
  });
}

export const useActivityStore = create<ActivityState>((set, get) => ({
  activities: [],
  inboxItems: [],
  mentionItems: [],
  // Phase 64A initial state
  unifiedFeedItems: [],
  isLoadingUnifiedFeed: false,
  unifiedFeedCursor: null,
  hasMoreUnifiedFeed: false,
  fetchActivities: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/activity`)
      const data = await response.json()
      set({ activities: deduplicateItems(data.activities.map(mapActivity)) })
    } catch (error) {
      console.error("Failed to fetch activities:", error)
    }
  },
  fetchInbox: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/inbox`)
      const data = await response.json()
      set({ inboxItems: deduplicateItems(data.items.map(mapActivity)) })
    } catch (error) {
      console.error("Failed to fetch inbox:", error)
    }
  },
  fetchMentions: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/mentions`)
      const data = await response.json()
      set({ mentionItems: deduplicateItems(data.items.map(mapActivity)) })
    } catch (error) {
      console.error("Failed to fetch mentions:", error)
    }
  },
  markAsRead: async (itemIds) => {
    if (itemIds.length === 0) return

    try {
      // Optimistic update
      set((state) => ({
        inboxItems: state.inboxItems.map(item => 
          itemIds.includes(item.id) ? { ...item, isRead: true } : item
        ),
        mentionItems: state.mentionItems.map(item => 
          itemIds.includes(item.id) ? { ...item, isRead: true } : item
        ),
        activities: state.activities.map(item => 
          itemIds.includes(item.id) ? { ...item, isRead: true } : item
        )
      }))

      await fetch(`${API_BASE_URL}/notifications/read`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ item_ids: itemIds })
      })
    } catch (error) {
      console.error("Failed to mark as read:", error)
      // We could revert here, but for simple read state, 
      // it's usually fine to just let the next fetch fix it
    }
  },
  markAsReadLocally: (itemIds) => {
    set((state) => ({
      inboxItems: state.inboxItems.map(item => 
        itemIds.includes(item.id) ? { ...item, isRead: true } : item
      ),
      mentionItems: state.mentionItems.map(item => 
        itemIds.includes(item.id) ? { ...item, isRead: true } : item
      ),
      activities: state.activities.map(item => 
        itemIds.includes(item.id) ? { ...item, isRead: true } : item
      )
    }))
  },
  appendMentionItem: (item) => {
    set(state => {
      if (state.mentionItems.some(i => i.id === item.id)) return state
      return {
        mentionItems: [item, ...state.mentionItems].slice(0, 200),
        activities: state.activities.some(i => i.id === item.id)
          ? state.activities
          : [item, ...state.activities].slice(0, 200),
      }
    })
  },

  // ── Phase 64A: Unified Activity Feed ──────────────────────────────────────
  //
  // Calls GET /api/v1/activity/feed (Phase 64A backend, pending).
  // On 404/failure gracefully sets empty list — the UI will show AI-store
  // data as a fallback in the "AI Events" tab.

  fetchUnifiedFeed: async (filters) => {
    set({ isLoadingUnifiedFeed: true })
    try {
      const params = new URLSearchParams()
      if (filters.workspaceId) params.set('workspace_id', filters.workspaceId)
      if (filters.channelId)   params.set('channel_id', filters.channelId)
      if (filters.dmId)        params.set('dm_id', filters.dmId)
      if (filters.actorId)     params.set('actor_id', filters.actorId)
      if (filters.eventType)   params.set('event_type', filters.eventType)
      if (filters.limit)       params.set('limit', String(filters.limit))
      if (filters.cursor)      params.set('cursor', filters.cursor)

      const res = await fetch(`${API_BASE_URL}/activity/feed?${params}`)
      if (!res.ok) {
        set({ isLoadingUnifiedFeed: false })
        return
      }
      const data = await res.json()
      const incoming: UnifiedActivityFeedItem[] = data.items ?? []
      const cursor = data.next_cursor ?? null
      const { unifiedFeedItems } = get()
      if (filters.cursor) {
        // pagination append
        const existing = new Set(unifiedFeedItems.map(i => i.id))
        const merged = [...unifiedFeedItems, ...incoming.filter(i => !existing.has(i.id))]
        set({ unifiedFeedItems: merged, unifiedFeedCursor: cursor, hasMoreUnifiedFeed: !!cursor, isLoadingUnifiedFeed: false })
      } else {
        set({ unifiedFeedItems: incoming, unifiedFeedCursor: cursor, hasMoreUnifiedFeed: !!cursor, isLoadingUnifiedFeed: false })
      }
    } catch {
      set({ isLoadingUnifiedFeed: false })
    }
  },

  appendUnifiedFeedItem: (item) => {
    set(state => {
      if (state.unifiedFeedItems.some(i => i.id === item.id)) return state
      return { unifiedFeedItems: [item, ...state.unifiedFeedItems].slice(0, 200) }
    })
  },
}))
