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
  updatePresence: (user: any) => void
  setTyping: (data: TypingState) => void
  sendTyping: (scope: { channelId?: string, dmId?: string, threadId?: string }, isTyping: boolean) => Promise<void>
}

export const usePresenceStore = create<PresenceState>((set, get) => ({
  typingIndicators: {},

  fetchPresence: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/presence`)
      const data = await response.json()
      // data.presence is likely a list of users or statuses
      // We update the user-store with this info
      // For now let's just assume it returns user objects with updated status
      if (data.users) {
        // We might want a way to bulk update in userStore
        // But for simplicity let's just trigger fetchUsers or handle individual updates
      }
    } catch (error) {
      console.error("Failed to fetch presence:", error)
    }
  },

  updatePresence: (userData: any) => {
    // This will be called from websocket
    const userStore = useUserStore.getState()
    const mappedUser = {
      ...userData,
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
          ...scope,
          is_typing: isTyping,
          // user_id is handled by backend from session or we pass it
          user_id: useUserStore.getState().currentUser?.id
        })
      })
    } catch (error) {
      // Silent error for typing
    }
  }
}))
