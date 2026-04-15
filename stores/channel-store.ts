import { create } from "zustand"
import { Channel } from "@/types"
import { CHANNELS } from "@/lib/mock-data"

interface ChannelState {
  channels: Channel[]
  currentChannel: Channel | null
  setCurrentChannel: (channel: Channel | null) => void
  addChannel: (channel: Channel) => void
}

export const useChannelStore = create<ChannelState>((set) => ({
  channels: CHANNELS,
  currentChannel: CHANNELS[0],
  setCurrentChannel: (channel) => set({ currentChannel: channel }),
  addChannel: (channel) => set((state) => ({ channels: [...state.channels, channel] })),
}))
