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
}

interface ActivityState {
  activities: ActivityItem[]
  inboxItems: ActivityItem[]
  mentionItems: ActivityItem[]
  fetchActivities: () => Promise<void>
  fetchInbox: () => Promise<void>
  fetchMentions: () => Promise<void>
}

export const useActivityStore = create<ActivityState>((set) => ({
  activities: [],
  inboxItems: [],
  mentionItems: [],
  fetchActivities: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/activity`)
      const data = await response.json()
      set({ activities: data.activities.map((a: any) => ({
        ...a,
        occurredAt: a.occurred_at
      })) })
    } catch (error) {
      console.error("Failed to fetch activities:", error)
    }
  },
  fetchInbox: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/inbox`)
      const data = await response.json()
      set({ inboxItems: data.items.map((a: any) => ({
        ...a,
        occurredAt: a.occurred_at
      })) })
    } catch (error) {
      console.error("Failed to fetch inbox:", error)
    }
  },
  fetchMentions: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/mentions`)
      const data = await response.json()
      set({ mentionItems: data.items.map((a: any) => ({
        ...a,
        occurredAt: a.occurred_at
      })) })
    } catch (error) {
      console.error("Failed to fetch mentions:", error)
    }
  }
}))
