import { create } from "zustand"
import { Channel, ChannelMember } from "@/types"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"

interface ChannelState {
  channels: Channel[]
  currentChannel: Channel | null
  members: ChannelMember[]
  currentChannelSummary: string | null
  isSummaryLoading: boolean
  setCurrentChannel: (channel: Channel | null) => void
  fetchChannels: (workspaceId: string) => Promise<void>
  addChannel: (channel: Channel) => void
  createChannel: (name: string, description: string, isPrivate: boolean) => Promise<Channel>
  setCurrentChannelById: (id: string) => void
  toggleStar: (id: string) => Promise<void>
  fetchStarredChannels: () => Promise<void>
  updateChannel: (id: string, updates: Partial<Channel>) => Promise<void>
  fetchMembers: (id: string) => Promise<void>
  addMember: (channelId: string, userId: string) => Promise<void>
  removeMember: (channelId: string, userId: string) => Promise<void>
  fetchChannelSummary: (id: string) => Promise<void>
  generateChannelSummary: (id: string) => Promise<void>
  fetchChannelPreferences: (id: string) => Promise<any>
  updateChannelPreferences: (id: string, updates: any) => Promise<void>
  leaveChannel: (id: string) => Promise<void>
}

export const useChannelStore = create<ChannelState>((set, get) => ({
  channels: [],
  currentChannel: null,
  members: [],
  currentChannelSummary: null,
  isSummaryLoading: false,
  setCurrentChannel: (channel) => {
    set({ currentChannel: channel, currentChannelSummary: null })
    if (channel) {
      get().fetchMembers(channel.id)
      get().fetchChannelSummary(channel.id)
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
  createChannel: async (name: string, description: string, isPrivate: boolean) => {
    const workspaceId = get().channels[0]?.workspaceId || "ws_1"
    try {
      const response = await fetch(`${API_BASE_URL}/channels`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ 
          name, 
          description, 
          type: isPrivate ? "private" : "public",
          workspace_id: workspaceId
        })
      })
      const data = await response.json()
      const newChannel = data.channel
      set((state) => ({ 
        channels: [...state.channels, newChannel]
      }))
      get().setCurrentChannel(newChannel)
      return newChannel
    } catch (error) {
      console.error("Failed to create channel:", error)
      throw error
    }
  },
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
      channels: state.channels.map(c => {
        if (c.id === id) return { ...c, isStarred: newStarred }
        return c
      }),
      currentChannel: get().currentChannel?.id === id ? { ...get().currentChannel!, isStarred: newStarred } : get().currentChannel
    }))

    try {
      await fetch(`${API_BASE_URL}/channels/${id}/star`, {
        method: "POST"
      })
      toast.success(newStarred ? "Channel starred" : "Channel unstarred")
    } catch (error) {
      // Revert on error
      set((state) => ({
        channels: state.channels.map(c => {
          if (c.id === id) return { ...c, isStarred: !newStarred }
          return c
        }),
        currentChannel: get().currentChannel?.id === id ? { ...get().currentChannel!, isStarred: !newStarred } : get().currentChannel
      }))
      console.error("Failed to toggle star:", error)
      toast.error("Failed to update star status")
    }
  },
  fetchStarredChannels: async () => {
    try {
      await fetch(`${API_BASE_URL}/starred`)
      // Current implementation just relies on fetchChannels bringing all data
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
      if (!response.ok) throw new Error("Update failed")
      
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
      if (!response.ok) throw new Error("Add failed")
      await get().fetchMembers(channelId)
      toast.success("Member added")
    } catch (error) {
      console.error("Failed to add member:", error)
      toast.error("Failed to add member")
    }
  },
  removeMember: async (channelId, userId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/channels/${channelId}/members/${userId}`, {
        method: "DELETE"
      })
      if (!response.ok) throw new Error("Remove failed")
      set((state) => ({
        members: state.members.filter(m => m.user.id !== userId)
      }))
      toast.success("Member removed")
    } catch (error) {
      console.error("Failed to remove member:", error)
      toast.error("Failed to remove member")
    }
  },
  fetchChannelSummary: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/channels/${id}/summary`)
      const data = await response.json()
      set({ currentChannelSummary: data.summary })
    } catch (error) {
      console.error("Failed to fetch channel summary:", error)
    }
  },
  generateChannelSummary: async (id) => {
    try {
      set({ isSummaryLoading: true })
      const response = await fetch(`${API_BASE_URL}/channels/${id}/summary`, {
        method: "POST"
      })
      const data = await response.json()
      set({ currentChannelSummary: data.summary, isSummaryLoading: false })
      toast.success("Channel summary generated")
    } catch (error) {
      console.error("Failed to generate channel summary:", error)
      set({ isSummaryLoading: false })
      toast.error("Failed to generate summary")
    }
  },

  fetchChannelPreferences: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/channels/${id}/preferences`)
      const data = await response.json()
      return data.preferences
    } catch (error) {
      console.error("Failed to fetch channel preferences:", error)
      return null
    }
  },

  updateChannelPreferences: async (id, updates) => {
    try {
      const response = await fetch(`${API_BASE_URL}/channels/${id}/preferences`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(updates)
      })
      if (!response.ok) throw new Error("Update failed")
      toast.success("Channel preferences updated")
    } catch (error) {
      console.error("Failed to update channel preferences:", error)
      toast.error("Failed to update preferences")
    }
  },

  leaveChannel: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/channels/${id}/leave`, {
        method: "POST"
      })
      if (!response.ok) throw new Error("Leave failed")
      set((state) => ({
        channels: state.channels.filter(c => c.id !== id),
        currentChannel: state.currentChannel?.id === id ? null : state.currentChannel
      }))
      toast.success("You have left the channel")
    } catch (error) {
      console.error("Failed to leave channel:", error)
      toast.error("Failed to leave channel")
    }
  }
}))
