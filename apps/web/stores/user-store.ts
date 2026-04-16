import { create } from "zustand"
import { User } from "@/types"
import { API_BASE_URL } from "@/lib/constants"

interface UserState {
  currentUser: User | null
  users: User[]
  fetchMe: () => Promise<void>
  fetchUsers: () => Promise<void>
}

export const useUserStore = create<UserState>((set) => ({
  currentUser: null,
  users: [],
  fetchMe: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/me`)
      const data = await response.json()
      set({ currentUser: data.user })
    } catch (error) {
      console.error("Failed to fetch current user:", error)
    }
  },
  fetchUsers: async () => {
    // Note: We'll need this endpoint soon
    // For now, let's keep it empty or mock it if needed
  }
}))
