import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"

interface NotificationPreferences {
  inboxEnabled: boolean
  mentionsEnabled: boolean
  dmEnabled: boolean
  muteAll: boolean
  muteRules: string[]
}

interface NotificationState {
  preferences: NotificationPreferences | null
  isLoading: boolean
  fetchPreferences: () => Promise<void>
  updatePreferences: (updates: Partial<NotificationPreferences>) => Promise<void>
}

const mapPreferences = (p: any): NotificationPreferences => ({
  inboxEnabled: p.inbox_enabled,
  mentionsEnabled: p.mentions_enabled,
  dmEnabled: p.dm_enabled,
  muteAll: p.mute_all,
  muteRules: p.mute_rules || []
})

const unmapPreferences = (p: Partial<NotificationPreferences>): any => {
  const result: any = {}
  if (p.inboxEnabled !== undefined) result.inbox_enabled = p.inboxEnabled
  if (p.mentionsEnabled !== undefined) result.mentions_enabled = p.mentionsEnabled
  if (p.dmEnabled !== undefined) result.dm_enabled = p.dmEnabled
  if (p.muteAll !== undefined) result.mute_all = p.muteAll
  if (p.muteRules !== undefined) result.mute_rules = p.muteRules
  return result
}

export const useNotificationStore = create<NotificationState>((set) => ({
  preferences: null,
  isLoading: false,

  fetchPreferences: async () => {
    try {
      set({ isLoading: true })
      const response = await fetch(`${API_BASE_URL}/notifications/preferences`)
      const data = await response.json()
      set({ preferences: mapPreferences(data.preferences), isLoading: false })
    } catch (error) {
      console.error("Failed to fetch notification preferences:", error)
      set({ isLoading: false })
    }
  },

  updatePreferences: async (updates) => {
    try {
      const response = await fetch(`${API_BASE_URL}/notifications/preferences`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(unmapPreferences(updates))
      })
      const data = await response.json()
      set({ preferences: mapPreferences(data.preferences) })
    } catch (error) {
      console.error("Failed to update notification preferences:", error)
    }
  }
}))
