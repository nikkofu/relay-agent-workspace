import { create } from "zustand"
import { AIMessage } from "@/types"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"

export interface AIConversation {
  id: string
  title: string
  lastMessage?: string
  updatedAt: string
}

interface AIState {
  conversations: AIConversation[]
  currentConversationId: string | null
  messages: AIMessage[]
  isLoading: boolean
  
  fetchConversations: () => Promise<void>
  fetchConversationDetail: (id: string) => Promise<void>
  setCurrentConversation: (id: string | null) => void
  addMessage: (msg: AIMessage) => void
  updateMessage: (id: string, updates: Partial<AIMessage>) => void
  setLoading: (loading: boolean) => void
  setConversations: (convs: AIConversation[]) => void
}

export const useAIStore = create<AIState>((set, get) => ({
  conversations: [],
  currentConversationId: null,
  messages: [],
  isLoading: false,

  fetchConversations: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/ai/conversations`)
      const data = await response.json()
      set({ conversations: data.conversations.map((c: any) => ({
        id: c.id,
        title: c.title,
        lastMessage: c.last_message,
        updatedAt: c.updated_at
      })) })
    } catch (error) {
      console.error("Failed to fetch AI conversations:", error)
    }
  },

  fetchConversationDetail: async (id) => {
    try {
      set({ isLoading: true, currentConversationId: id })
      const response = await fetch(`${API_BASE_URL}/ai/conversations/${id}`)
      const data = await response.json()
      
      const mappedMessages = data.messages.map((m: any) => ({
        id: m.id,
        role: m.role,
        content: m.content,
        reasoning: m.reasoning,
        createdAt: m.created_at,
        isStreaming: false
      }))
      
      set({ messages: mappedMessages, isLoading: false })
    } catch (error) {
      console.error("Failed to fetch AI conversation detail:", error)
      set({ isLoading: false })
    }
  },

  setCurrentConversation: (id) => {
    if (id === null) {
      set({ currentConversationId: null, messages: [] })
    } else {
      get().fetchConversationDetail(id)
    }
  },

  addMessage: (msg) => set((state) => ({ 
    messages: [...state.messages, msg] 
  })),

  updateMessage: (id, updates) => set((state) => ({
    messages: state.messages.map(m => m.id === id ? { ...m, ...updates } : m)
  })),

  setLoading: (loading) => set({ isLoading: loading }),
  
  setConversations: (convs) => set({ conversations: convs })
}))
