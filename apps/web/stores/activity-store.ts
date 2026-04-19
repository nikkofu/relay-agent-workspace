import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"

export interface ActivityItem {
  id: string
  type: 'mention' | 'reaction' | 'reply' | 'channel_join' | 'dm_message' | 'thread_reply'
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
  }
}))
