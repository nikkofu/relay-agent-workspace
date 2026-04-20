import { create } from "zustand"
import { UserGroup, Workflow, Tool } from "@/types"
import { API_BASE_URL } from "@/lib/constants"

interface DirectoryState {
  userGroups: UserGroup[]
  activeGroup: UserGroup | null
  workflows: Workflow[]
  tools: Tool[]
  isLoading: boolean
  
  fetchUserGroups: () => Promise<void>
  fetchGroupDetail: (id: string) => Promise<void>
  fetchWorkflows: () => Promise<void>
  fetchTools: () => Promise<void>
}

export const useDirectoryStore = create<DirectoryState>((set) => ({
  userGroups: [],
  activeGroup: null,
  workflows: [],
  tools: [],
  isLoading: false,

  fetchUserGroups: async () => {
    try {
      set({ isLoading: true })
      const response = await fetch(`${API_BASE_URL}/user-groups`)
      const data = await response.json()
      set({ userGroups: data.groups || [], isLoading: false })
    } catch (error) {
      console.error("Failed to fetch user groups:", error)
      set({ isLoading: false })
    }
  },

  fetchGroupDetail: async (id) => {
    try {
      set({ isLoading: true })
      const response = await fetch(`${API_BASE_URL}/user-groups/${id}`)
      const data = await response.json()
      set({ activeGroup: data.group || null, isLoading: false })
    } catch (error) {
      console.error("Failed to fetch user group detail:", error)
      set({ isLoading: false })
    }
  },

  fetchWorkflows: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/workflows`)
      const data = await response.json()
      set({ workflows: data.workflows || [] })
    } catch (error) {
      console.error("Failed to fetch workflows:", error)
    }
  },

  fetchTools: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/tools`)
      const data = await response.json()
      set({ tools: data.tools || [] })
    } catch (error) {
      console.error("Failed to fetch tools:", error)
    }
  }
}))
