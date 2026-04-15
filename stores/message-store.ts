import { create } from "zustand"
import { Message } from "@/types"
import { MESSAGES } from "@/lib/mock-data"

interface MessageState {
  messages: Message[]
  addMessage: (message: Message) => void
  removeMessage: (id: string) => void
  getMessagesByChannel: (channelId: string) => Message[]
}

export const useMessageStore = create<MessageState>((set, get) => ({
  messages: MESSAGES,
  addMessage: (message) => set((state) => ({ messages: [...state.messages, message] })),
  removeMessage: (id) => set((state) => ({ messages: state.messages.filter((m) => m.id !== id) })),
  getMessagesByChannel: (channelId) => get().messages.filter((m) => m.channelId === channelId),
}))
