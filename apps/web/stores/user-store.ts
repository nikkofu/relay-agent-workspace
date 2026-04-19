import { create } from "zustand"
import { User } from "@/types"
import { API_BASE_URL } from "@/lib/constants"

interface UserState {
  currentUser: User | null
  users: User[]
  fetchMe: () => Promise<void>
  fetchUsers: () => Promise<void>
}

const mapUser = (u: any): User => ({
  ...u,
  statusText: u.status_text,
  lastSeen: u.last_seen_at,
  aiInsight: u.ai_insight,
  aiProvider: u.ai_provider,
  aiModel: u.ai_model,
  aiMode: u.ai_mode
})

export const useUserStore = create<UserState>((set) => ({
  currentUser: null,
  users: [],
  fetchMe: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/me`)
      const data = await response.json()
      set({ currentUser: mapUser(data.user) })
    } catch (error) {
      console.error("Failed to fetch current user:", error)
    }
  },
  fetchUsers: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/users`)
      const data = await response.json()
      set({ users: data.users.map(mapUser) })
    } catch (error) {
      console.error("Failed to fetch users:", error)
    }
  }
}))
