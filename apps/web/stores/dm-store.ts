import { create } from "zustand"
import { DirectMessage } from "@/types"
import { API_BASE_URL } from "@/lib/constants"

interface DMState {
  conversations: DirectMessage[]
  currentConversation: DirectMessage | null
  setCurrentConversation: (conversation: DirectMessage | null) => void
  fetchConversations: () => Promise<void>
  createConversation: (userIds: string[]) => Promise<DirectMessage | null>
}

export const useDMStore = create<DMState>((set, get) => ({
  conversations: [],
  currentConversation: null,
  setCurrentConversation: (conversation) => set({ currentConversation: conversation }),
  fetchConversations: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/dms`)
      const data = await response.json()
      set({ conversations: data.conversations })
    } catch (error) {
      console.error("Failed to fetch DM conversations:", error)
    }
  },
  createConversation: async (userIds) => {
    try {
      const response = await fetch(`${API_BASE_URL}/dms`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user_ids: userIds })
      })
      const data = await response.json()
      const newConv = data.conversation
      set((state) => ({ 
        conversations: [newConv, ...state.conversations.filter(c => c.id !== newConv.id)] 
      }))
      return newConv
    } catch (error) {
      console.error("Failed to create DM conversation:", error)
      return null
    }
  }
}))
