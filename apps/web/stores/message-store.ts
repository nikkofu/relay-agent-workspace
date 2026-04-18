import { create } from "zustand"
import { Message } from "@/types"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"

interface MessageState {
  messages: Message[]
  currentThreadMessages: Message[]
  pinnedMessages: { message: Message, channel: any, user: any }[]
  currentThreadSummary: string | null
  isSummaryLoading: boolean
  fetchMessages: (channelId: string) => Promise<void>
  fetchDMMessages: (dmId: string) => Promise<void>
  fetchThread: (messageId: string) => Promise<void>
  fetchPins: (channelId?: string) => Promise<void>
  fetchThreadSummary: (messageId: string) => Promise<void>
  generateThreadSummary: (messageId: string) => Promise<void>
  addMessage: (message: Message) => void
  sendMessage: (channelId: string, content: string, userId: string, threadId?: string) => Promise<void>
  sendDMMessage: (dmId: string, content: string, userId: string) => Promise<void>
  getMessagesByChannel: (channelId: string) => Message[]
  getMessagesByDM: (dmId: string) => Message[]
  // ... rest of actions
  addReaction: (messageId: string, emoji: string) => Promise<void>
  deleteMessage: (messageId: string) => Promise<void>
  pinMessage: (messageId: string) => Promise<void>
  saveForLater: (messageId: string) => Promise<void>
  markAsUnread: (messageId: string) => Promise<void>
  deleteMessageLocally: (messageId: string) => void
  updateMessageLocally: (rawMessage: any) => void
}

const mapMessage = (m: any): Message => {
  let reactions = []
  let attachments = []
  if (m.metadata) {
    try {
      const meta = typeof m.metadata === 'string' ? JSON.parse(m.metadata) : m.metadata
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
    dmId: m.dm_id,
    threadId: m.thread_id,
    createdAt: m.created_at,
    reactions,
    attachments,
    replyCount: m.reply_count,
    lastReplyAt: m.last_reply_at,
    isPinned: m.is_pinned
  }
}

export const useMessageStore = create<MessageState>((set, get) => ({
  messages: [],
  currentThreadMessages: [],
  pinnedMessages: [],
  currentThreadSummary: null,
  isSummaryLoading: false,
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
  fetchDMMessages: async (dmId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/dms/${dmId}/messages`)
      const data = await response.json()
      const mappedMessages = data.messages.map(mapMessage)
      set({ messages: mappedMessages })
    } catch (error) {
      console.error("Failed to fetch DM messages:", error)
    }
  },
  fetchThread: async (messageId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/messages/${messageId}/thread`)
      const data = await response.json()
      const parent = mapMessage(data.parent)
      const replies = data.replies.map(mapMessage)
      set({ currentThreadMessages: [parent, ...replies] })
      
      // Fetch summary when opening thread
      get().fetchThreadSummary(messageId)
    } catch (error) {
      console.error("Failed to fetch thread:", error)
    }
  },
  fetchThreadSummary: async (messageId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/messages/${messageId}/summary`)
      const data = await response.json()
      set({ currentThreadSummary: data.summary })
    } catch (error) {
      console.error("Failed to fetch thread summary:", error)
    }
  },
  generateThreadSummary: async (messageId) => {
    try {
      set({ isSummaryLoading: true })
      const response = await fetch(`${API_BASE_URL}/messages/${messageId}/summary`, {
        method: "POST"
      })
      const data = await response.json()
      set({ currentThreadSummary: data.summary, isSummaryLoading: false })
      toast.success("Thread summary generated")
    } catch (error) {
      console.error("Failed to generate thread summary:", error)
      set({ isSummaryLoading: false })
      toast.error("Failed to generate summary")
    }
  },
  fetchPins: async (channelId) => {
    try {
      const url = channelId ? `${API_BASE_URL}/pins?channel_id=${channelId}` : `${API_BASE_URL}/pins`
      const response = await fetch(url)
      const data = await response.json()
      set({ pinnedMessages: (data.items || []).map((item: any) => ({
        ...item,
        message: mapMessage(item.message)
      })) })
    } catch (error) {
      console.error("Failed to fetch pins:", error)
    }
  },
  addMessage: (message) => set((state) => {
    if (state.messages.find(m => m.id === message.id)) return state;
    return { messages: [...state.messages, message] };
  }),
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
        set((state) => {
          if (state.currentThreadMessages.find(m => m.id === msg.id)) return state;
          return { 
            currentThreadMessages: [...state.currentThreadMessages, msg],
            messages: state.messages.map(m => m.id === threadId ? { ...m, replyCount: (m.replyCount || 0) + 1 } : m)
          };
        })
      } else {
        set((state) => {
          if (state.messages.find(m => m.id === msg.id)) return state;
          return { messages: [...state.messages, msg] };
        })
      }
    } catch (error) {
      console.error("Failed to send message:", error)
    }
  },
  sendDMMessage: async (dmId, content, userId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/dms/${dmId}/messages`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content, user_id: userId })
      })
      const data = await response.json()
      const msg = mapMessage(data.message)
      set((state) => {
        if (state.messages.find(m => m.id === msg.id)) return state;
        return { messages: [...state.messages, msg] };
      })
    } catch (error) {
      console.error("Failed to send DM message:", error)
    }
  },
  getMessagesByChannel: (channelId) => get().messages.filter((m) => m.channelId === channelId),
  getMessagesByDM: (dmId) => get().messages.filter((m) => m.dmId === dmId),
  
  deleteMessageLocally: (messageId: string) => {
    set(state => ({
      messages: state.messages.filter(m => m.id !== messageId),
      currentThreadMessages: state.currentThreadMessages.filter(m => m.id !== messageId)
    }))
  },

  updateMessageLocally: (rawMessage: any) => {
    const updatedMsg = mapMessage(rawMessage)
    set(state => ({
      messages: state.messages.map(m => m.id === updatedMsg.id ? updatedMsg : m),
      currentThreadMessages: state.currentThreadMessages.map(m => m.id === updatedMsg.id ? updatedMsg : m)
    }))
  },

  addReaction: async (messageId, emoji) => {
    try {
      const response = await fetch(`${API_BASE_URL}/messages/${messageId}/reactions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ emoji })
      })
      const data = await response.json()
      const updatedMsg = mapMessage(data.message)
      
      set(state => ({
        messages: state.messages.map(m => m.id === messageId ? updatedMsg : m),
        currentThreadMessages: state.currentThreadMessages.map(m => m.id === messageId ? updatedMsg : m)
      }))
      
      toast.success(data.added ? `Added ${emoji} reaction` : `Removed ${emoji} reaction`)
    } catch (error) {
      console.error("Failed to toggle reaction:", error)
    }
  },
  deleteMessage: async (messageId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/messages/${messageId}`, {
        method: "DELETE"
      })
      const data = await response.json()
      
      if (data.deleted) {
        set(state => ({
          messages: state.messages.filter(m => m.id !== messageId),
          currentThreadMessages: state.currentThreadMessages.filter(m => m.id !== messageId)
        }))
        toast.success("Message deleted")
      }
    } catch (error) {
      console.error("Failed to delete message:", error)
    }
  },
  pinMessage: async (messageId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/messages/${messageId}/pin`, {
        method: "POST"
      })
      const data = await response.json()
      const updatedMsg = mapMessage(data.message)
      
      set(state => ({
        messages: state.messages.map(m => m.id === messageId ? updatedMsg : m),
        currentThreadMessages: state.currentThreadMessages.map(m => m.id === messageId ? updatedMsg : m)
      }))
      
      // Refresh pins list
      get().fetchPins(updatedMsg.channelId)

      toast.success(data.is_pinned ? "Message pinned" : "Message unpinned")
    } catch (error) {
      console.error("Failed to toggle pin:", error)
    }
  },
  saveForLater: async (messageId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/messages/${messageId}/later`, {
        method: "POST"
      })
      const data = await response.json()
      toast.success(data.saved ? "Saved for later" : "Removed from later")
    } catch (error) {
      console.error("Failed to toggle save for later:", error)
    }
  },
  markAsUnread: async (messageId) => {
    try {
      await fetch(`${API_BASE_URL}/messages/${messageId}/unread`, {
        method: "POST"
      })
      toast.success("Marked as unread")
    } catch (error) {
      console.error("Failed to mark as unread:", error)
    }
  }
}))
