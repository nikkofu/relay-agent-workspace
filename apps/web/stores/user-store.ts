import { create } from "zustand"
import { User } from "@/types"
import { API_BASE_URL } from "@/lib/constants"

interface UserState {
  currentUser: User | null
  users: User[]
  userDetail: User | null
  isUserLoading: boolean
  fetchMe: () => Promise<void>
  fetchUsers: (params?: { q?: string, department?: string, status?: string, timezone?: string, userGroupId?: string }) => Promise<void>
  fetchUserDetail: (id: string) => Promise<void>
  updateStatus: (id: string, status: string, statusText: string) => Promise<void>
  updateProfile: (id: string, updates: { title?: string, department?: string, timezone?: string, workingHours?: string }) => Promise<void>
}

const mapUser = (u: any): User => ({
  ...u,
  statusText: u.status_text,
  lastSeen: u.last_seen_at,
  workingHours: u.working_hours,
  aiInsight: u.ai_insight,
  aiProvider: u.ai_provider,
  aiModel: u.ai_model,
  aiMode: u.ai_mode,
  profile: u.profile ? {
    localTime: u.profile.local_time,
    workingHours: u.profile.working_hours,
    focusAreas: u.profile.focus_areas,
    topChannels: u.profile.top_channels,
    recentArtifacts: u.profile.recent_artifacts
  } : undefined
})

export const useUserStore = create<UserState>((set) => ({
  currentUser: null,
  users: [],
  userDetail: null,
  isUserLoading: false,
  fetchMe: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/me`)
      const data = await response.json()
      set({ currentUser: mapUser(data.user) })
    } catch (error) {
      console.error("Failed to fetch current user:", error)
    }
  },
  fetchUsers: async (params) => {
    try {
      const queryParams = new URLSearchParams()
      if (params?.q) queryParams.append('q', params.q)
      if (params?.department) queryParams.append('department', params.department)
      if (params?.status) queryParams.append('status', params.status)
      if (params?.timezone) queryParams.append('timezone', params.timezone)
      if (params?.userGroupId) queryParams.append('user_group_id', params.userGroupId)

      const url = `${API_BASE_URL}/users${queryParams.toString() ? `?${queryParams.toString()}` : ''}`
      const response = await fetch(url)
      const data = await response.json()
      set({ users: data.users.map(mapUser) })
    } catch (error) {
      console.error("Failed to fetch users:", error)
    }
  },
  fetchUserDetail: async (id) => {
    try {
      set({ isUserLoading: true })
      const response = await fetch(`${API_BASE_URL}/users/${id}`)
      const data = await response.json()
      set({ userDetail: mapUser(data.user), isUserLoading: false })
    } catch (error) {
      console.error("Failed to fetch user detail:", error)
      set({ isUserLoading: false })
    }
  },
  updateStatus: async (id, status, statusText) => {
    try {
      const response = await fetch(`${API_BASE_URL}/users/${id}/status`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status, status_text: statusText })
      })
      const data = await response.json()
      const updated = mapUser(data.user)
      set((state) => ({
        users: state.users.map(u => u.id === id ? updated : u),
        currentUser: state.currentUser?.id === id ? updated : state.currentUser,
        userDetail: state.userDetail?.id === id ? updated : state.userDetail
      }))
    } catch (error) {
      console.error("Failed to update status:", error)
    }
  },
  updateProfile: async (id, updates) => {
    try {
      const payload: any = {}
      if (updates.title !== undefined) payload.title = updates.title
      if (updates.department !== undefined) payload.department = updates.department
      if (updates.timezone !== undefined) payload.timezone = updates.timezone
      if (updates.workingHours !== undefined) payload.working_hours = updates.workingHours

      const response = await fetch(`${API_BASE_URL}/users/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      })
      const data = await response.json()
      const updated = mapUser(data.user)
      set((state) => ({
        users: state.users.map(u => u.id === id ? updated : u),
        currentUser: state.currentUser?.id === id ? updated : state.currentUser,
        userDetail: state.userDetail?.id === id ? updated : state.userDetail
      }))
    } catch (error) {
      console.error("Failed to update profile:", error)
    }
  }
}))
