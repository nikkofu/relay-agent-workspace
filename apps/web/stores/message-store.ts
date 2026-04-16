import { create } from "zustand"
import { Message } from "@/types"
import { API_BASE_URL } from "@/lib/constants"

interface MessageState {
  messages: Message[]
  fetchMessages: (channelId: string) => Promise<void>
  addMessage: (message: Message) => void
  sendMessage: (channelId: string, content: string, userId: string) => Promise<void>
  getMessagesByChannel: (channelId: string) => Message[]
}

export const useMessageStore = create<MessageState>((set, get) => ({
  messages: [],
  fetchMessages: async (channelId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/messages?channel_id=${channelId}`)
      const data = await response.json()
      
      const mappedMessages = data.messages.map((m: any) => {
        let reactions = []
        let attachments = []
        if (m.metadata) {
          try {
            const meta = JSON.parse(m.metadata)
            reactions = meta.reactions || []
            attachments = meta.attachments || []
          } catch (e) {
            console.error("Failed to parse metadata", e)
          }
        }
        return {
          id: m.id,
          content: m.content,
          senderId: m.user_id,
          channelId: m.channel_id,
          createdAt: m.created_at,
          reactions,
          attachments
        }
      })
      
      set({ messages: mappedMessages })
    } catch (error) {
      console.error("Failed to fetch messages:", error)
    }
  },
  addMessage: (message) => set((state) => ({ messages: [...state.messages, message] })),
  sendMessage: async (channelId, content, userId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/messages`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ channel_id: channelId, content, user_id: userId })
      })
      const data = await response.json()
      const msg = data.message
      set((state) => ({ messages: [...state.messages, {
        id: msg.id,
        content: msg.content,
        senderId: msg.user_id,
        channelId: msg.channel_id,
        createdAt: msg.created_at,
        reactions: [],
        attachments: []
      }]}))
    } catch (error) {
      console.error("Failed to send message:", error)
    }
  },
  getMessagesByChannel: (channelId) => get().messages.filter((m) => m.channelId === channelId),
}))
