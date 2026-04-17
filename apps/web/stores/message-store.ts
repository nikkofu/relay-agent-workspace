import { create } from "zustand"
import { Message } from "@/types"
import { API_BASE_URL } from "@/lib/constants"

interface MessageState {
  messages: Message[]
  currentThreadMessages: Message[]
  fetchMessages: (channelId: string) => Promise<void>
  fetchThread: (messageId: string) => Promise<void>
  addMessage: (message: Message) => void
  sendMessage: (channelId: string, content: string, userId: string, threadId?: string) => Promise<void>
  getMessagesByChannel: (channelId: string) => Message[]
}

const mapMessage = (m: any): Message => {
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
    threadId: m.thread_id,
    createdAt: m.created_at,
    reactions,
    attachments,
    replyCount: m.reply_count,
    lastReplyAt: m.last_reply_at
  }
}

export const useMessageStore = create<MessageState>((set, get) => ({
  messages: [],
  currentThreadMessages: [],
  fetchMessages: async (channelId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/messages?channel_id=${channelId}`)
      const data = await response.json()
      const mappedMessages = data.messages.map(mapMessage)
      set({ messages: mappedMessages })
    } catch (error) {
      console.error("Failed to fetch messages:", error)
    }
  },
  fetchThread: async (messageId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/messages/${messageId}/thread`)
      const data = await response.json()
      const parent = mapMessage(data.parent)
      const replies = data.replies.map(mapMessage)
      set({ currentThreadMessages: [parent, ...replies] })
    } catch (error) {
      console.error("Failed to fetch thread:", error)
    }
  },
  addMessage: (message) => set((state) => ({ messages: [...state.messages, message] })),
  sendMessage: async (channelId, content, userId, threadId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/messages`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ channel_id: channelId, content, user_id: userId, thread_id: threadId })
      })
      const data = await response.json()
      const msg = mapMessage(data.message)
      
      if (threadId) {
        set((state) => ({ 
          currentThreadMessages: [...state.currentThreadMessages, msg],
          // Update parent message reply count locally if needed
          messages: state.messages.map(m => m.id === threadId ? { ...m, replyCount: (m.replyCount || 0) + 1 } : m)
        }))
      } else {
        set((state) => ({ messages: [...state.messages, msg] }))
      }
    } catch (error) {
      console.error("Failed to send message:", error)
    }
  },
  getMessagesByChannel: (channelId) => get().messages.filter((m) => m.channelId === channelId),
}))
