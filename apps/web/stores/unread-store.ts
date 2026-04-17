import { create } from "zustand"
import { persist } from "zustand/middleware"

interface UnreadState {
  lastReadTimestamps: Record<string, string> // channelId -> ISO timestamp
  markAsRead: (channelId: string) => void
  getLastRead: (channelId: string) => string | null
}

export const useUnreadStore = create<UnreadState>()(
  persist(
    (set, get) => ({
      lastReadTimestamps: {},
      markAsRead: (channelId) => set((state) => ({
        lastReadTimestamps: {
          ...state.lastReadTimestamps,
          [channelId]: new Date().toISOString()
        }
      })),
      getLastRead: (channelId) => get().lastReadTimestamps[channelId] || null
    }),
    {
      name: 'relay-unread-storage'
    }
  )
)
