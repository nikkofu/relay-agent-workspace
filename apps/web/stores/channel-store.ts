import { create } from "zustand"
import { Channel } from "@/types"
import { API_BASE_URL } from "@/lib/constants"

interface ChannelState {
  channels: Channel[]
  currentChannel: Channel | null
  setCurrentChannel: (channel: Channel | null) => void
  fetchChannels: (workspaceId: string) => Promise<void>
  addChannel: (channel: Channel) => void
}

export const useChannelStore = create<ChannelState>((set) => ({
  channels: [],
  currentChannel: null,
  setCurrentChannel: (channel) => set({ currentChannel: channel }),
  fetchChannels: async (workspaceId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/channels?workspace_id=${workspaceId}`)
      const data = await response.json()
      set({ 
        channels: data.channels,
        currentChannel: data.channels[0] || null
      })
    } catch (error) {
      console.error("Failed to fetch channels:", error)
    }
  },
  addChannel: (channel) => set((state) => ({ channels: [...state.channels, channel] })),
}))
