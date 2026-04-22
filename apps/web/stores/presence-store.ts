import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"
import { useUserStore } from "./user-store"

interface TypingState {
  userId: string
  isTyping: boolean
  channelId?: string
  dmId?: string
  threadId?: string
}

interface PresenceState {
  typingIndicators: Record<string, TypingState[]> // key is channelId, dmId, or threadId
  fetchPresence: () => Promise<void>
  fetchScopedPresence: (channelId: string) => Promise<void>
  bulkHydratePresence: (channelId?: string) => Promise<void>
  updatePresence: (user: any) => void
  setTyping: (data: TypingState) => void
  sendTyping: (scope: { channelId?: string, dmId?: string, threadId?: string }, isTyping: boolean) => Promise<void>
  sendHeartbeat: () => Promise<void>
}

export const usePresenceStore = create<PresenceState>((set, get) => ({
  typingIndicators: {},

  fetchPresence: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/presence`)
      const data = await response.json()
      if (data.users) {
        // We can update the users in userStore if needed
        const { users: existingUsers } = useUserStore.getState()
        const updatedUsers = existingUsers.map(u => {
          const presence = data.users.find((pu: any) => pu.id === u.id)
          if (presence) {
            return {
              ...u,
              status: presence.status,
              statusText: presence.status_text,
              lastSeen: presence.last_seen_at
            }
          }
          return u
        })
        useUserStore.setState({ users: updatedUsers })
      }
    } catch (error) {
      console.error("Failed to fetch presence:", error)
    }
  },

  fetchScopedPresence: async (channelId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/presence?channel_id=${channelId}`)
      const data = await response.json()
      if (data.users) {
        const { users: existingUsers } = useUserStore.getState()
        const updatedUsers = existingUsers.map(u => {
          const presence = data.users.find((pu: any) => pu.id === u.id)
          if (presence) {
            return {
              ...u,
              status: presence.status,
              statusText: presence.status_text,
              lastSeen: presence.last_seen_at
            }
          }
          return u
        })
        useUserStore.setState({ users: updatedUsers })
      }
    } catch (error) {
      console.error("Failed to fetch scoped presence:", error)
    }
  },

  // Phase 61: bulk presence hydration on reconnect or channel-switch
  bulkHydratePresence: async (channelId?: string) => {
    try {
      const url = channelId
        ? `${API_BASE_URL}/presence/bulk?channel_id=${channelId}`
        : `${API_BASE_URL}/presence/bulk`
      const response = await fetch(url)
      if (!response.ok) return
      const data = await response.json()
      if (data.users && Array.isArray(data.users)) {
        const { users: existingUsers } = useUserStore.getState()
        const presenceMap = new Map<string, any>(data.users.map((u: any) => [u.id, u]))
        const updatedUsers = existingUsers.map(u => {
          const presence = presenceMap.get(u.id)
          if (presence) {
            return {
              ...u,
              status: presence.status,
              statusText: presence.status_text,
              lastSeen: presence.last_seen_at,
            }
          }
          return u
        })
        useUserStore.setState({ users: updatedUsers })
      }
    } catch (error) {
      console.error("Failed to bulk hydrate presence:", error)
    }
  },

  updatePresence: (userData: any) => {
    // This will be called from websocket
    const mappedUser = {
      ...userData,
      statusText: userData.status_text,
      lastSeen: userData.last_seen_at,
      aiInsight: userData.ai_insight,
      aiProvider: userData.ai_provider,
      aiModel: userData.ai_model,
      aiMode: userData.ai_mode
    }
    
    // Update users list and currentUser if it's them
    useUserStore.setState((state) => ({
      users: state.users.map(u => u.id === mappedUser.id ? { ...u, ...mappedUser } : u),
      currentUser: state.currentUser?.id === mappedUser.id ? { ...state.currentUser, ...mappedUser } : state.currentUser
    }))
  },

  setTyping: (data: TypingState) => {
    const key = data.threadId ? `thread:${data.threadId}` : 
                data.dmId ? `dm:${data.dmId}` : 
                data.channelId ? `channel:${data.channelId}` : 'global'
    
    set((state) => {
      const current = state.typingIndicators[key] || []
      const others = current.filter(t => t.userId !== data.userId)
      
      if (data.isTyping) {
        return {
          typingIndicators: {
            ...state.typingIndicators,
            [key]: [...others, data]
          }
        }
      } else {
        return {
          typingIndicators: {
            ...state.typingIndicators,
            [key]: others
          }
        }
      }
    })

    // Auto-clear typing after 5 seconds if no stop event arrives
    if (data.isTyping) {
      setTimeout(() => {
        get().setTyping({ ...data, isTyping: false })
      }, 5000)
    }
  },

  sendTyping: async (scope, isTyping) => {
    try {
      await fetch(`${API_BASE_URL}/typing`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          channel_id: scope.channelId,
          dm_id: scope.dmId,
          thread_id: scope.threadId,
          is_typing: isTyping,
          user_id: useUserStore.getState().currentUser?.id
        })
      })
    } catch {
      // Silent error for typing
    }
  },

  sendHeartbeat: async () => {
    try {
      await fetch(`${API_BASE_URL}/presence/heartbeat`, {
        method: "POST"
      })
    } catch {
      // Silent error for heartbeat
    }
  }
}))
