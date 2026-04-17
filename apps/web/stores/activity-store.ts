import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"

export interface ActivityItem {
  id: string
  type: 'mention' | 'reaction' | 'reply' | 'channel_join'
  user: any
  channel?: any
  message?: any
  target?: string
  summary: string
  occurredAt: string
}

interface ActivityState {
  activities: ActivityItem[]
  fetchActivities: () => Promise<void>
}

export const useActivityStore = create<ActivityState>((set) => ({
  activities: [],
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
  }
}))
