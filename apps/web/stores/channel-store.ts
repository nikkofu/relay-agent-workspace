import { create } from "zustand"
import { Channel, ChannelMember } from "@/types"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"

interface ChannelState {
  channels: Channel[]
  currentChannel: Channel | null
  members: ChannelMember[]
  setCurrentChannel: (channel: Channel | null) => void
  fetchChannels: (workspaceId: string) => Promise<void>
  addChannel: (channel: Channel) => void
  setCurrentChannelById: (id: string) => void
  toggleStar: (id: string) => Promise<void>
  fetchStarredChannels: () => Promise<void>
  updateChannel: (id: string, updates: Partial<Channel>) => Promise<void>
  fetchMembers: (id: string) => Promise<void>
  addMember: (channelId: string, userId: string) => Promise<void>
  removeMember: (channelId: string, userId: string) => Promise<void>
}

export const useChannelStore = create<ChannelState>((set, get) => ({
  channels: [],
  currentChannel: null,
  members: [],
  setCurrentChannel: (channel) => {
    set({ currentChannel: channel })
    if (channel) {
      get().fetchMembers(channel.id)
    }
  },
  fetchChannels: async (workspaceId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/channels?workspace_id=${workspaceId}`)
      const data = await response.json()
      set({ 
        channels: data.channels,
        currentChannel: data.channels[0] || null
      })
      if (data.channels[0]) {
        get().fetchMembers(data.channels[0].id)
      }
    } catch (error) {
      console.error("Failed to fetch channels:", error)
    }
  },
  addChannel: (channel) => set((state) => ({ channels: [...state.channels, channel] })),
  setCurrentChannelById: (id) => {
    const channel = get().channels.find(c => c.id === id)
    if (channel) {
      set({ currentChannel: channel })
      get().fetchMembers(id)
    }
  },
  toggleStar: async (id) => {
    const channel = get().channels.find(c => c.id === id)
    if (!channel) return

    // Optimistic update
    const newStarred = !channel.isStarred
    set((state) => ({
      channels: state.channels.map(c => c.id === id ? { ...c, isStarred: newStarred } : c),
      currentChannel: state.currentChannel?.id === id ? { ...state.currentChannel, isStarred: newStarred } : state.currentChannel
    }))

    try {
      await fetch(`${API_BASE_URL}/channels/${id}/star`, {
        method: "POST"
      })
      toast.success(newStarred ? "Channel starred" : "Channel unstarred")
    } catch (error) {
      // Revert on error
      set((state) => ({
        channels: state.channels.map(c => c.id === id ? { ...c, isStarred: !newStarred } : c),
        currentChannel: state.currentChannel?.id === id ? { ...state.currentChannel, isStarred: !newStarred } : state.currentChannel
      }))
      console.error("Failed to toggle star:", error)
      toast.error("Failed to update star status")
    }
  },
  fetchStarredChannels: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/starred`)
      const data = await response.json()
      // Update local isStarred state based on server response if needed
      // but for now we just rely on fetchChannels bringing all data
    } catch (error) {
      console.error("Failed to fetch starred channels:", error)
    }
  },
  updateChannel: async (id, updates) => {
    try {
      const response = await fetch(`${API_BASE_URL}/channels/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(updates)
      })
      const data = await response.json()
      set((state) => ({
        channels: state.channels.map(c => c.id === id ? { ...c, ...updates } : c),
        currentChannel: state.currentChannel?.id === id ? { ...state.currentChannel, ...updates } : state.currentChannel
      }))
      toast.success("Channel updated")
    } catch (error) {
      console.error("Failed to update channel:", error)
      toast.error("Failed to update channel")
    }
  },
  fetchMembers: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/channels/${id}/members`)
      const data = await response.json()
      set({ members: data.members })
    } catch (error) {
      console.error("Failed to fetch channel members:", error)
    }
  },
  addMember: async (channelId, userId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/channels/${channelId}/members`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user_id: userId })
      })
      await get().fetchMembers(channelId)
      toast.success("Member added")
    } catch (error) {
      console.error("Failed to add member:", error)
      toast.error("Failed to add member")
    }
  },
  removeMember: async (channelId, userId) => {
    try {
      await fetch(`${API_BASE_URL}/channels/${channelId}/members/${userId}`, {
        method: "DELETE"
      })
      set((state) => ({
        members: state.members.filter(m => m.user.id !== userId)
      }))
      toast.success("Member removed")
    } catch (error) {
      console.error("Failed to remove member:", error)
      toast.error("Failed to remove member")
    }
  }
}))
