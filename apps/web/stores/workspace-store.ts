import { create } from "zustand"
import { Workspace } from "@/types"
import { API_BASE_URL } from "@/lib/constants"

interface WorkspaceState {
  workspaces: Workspace[]
  currentWorkspace: Workspace | null
  homeData: any | null
  setCurrentWorkspace: (workspace: Workspace) => void
  fetchWorkspaces: () => Promise<void>
  fetchHome: () => Promise<void>
  inviteMember: (email: string) => Promise<void>
}

export const useWorkspaceStore = create<WorkspaceState>((set, get) => ({
  workspaces: [],
  currentWorkspace: null,
  homeData: null,
  setCurrentWorkspace: (workspace) => set({ currentWorkspace: workspace }),
  fetchWorkspaces: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/workspaces`)
      const data = await response.json()
      set({ 
        workspaces: data.workspaces,
        currentWorkspace: data.workspaces[0] || null 
      })
    } catch (error) {
      console.error("Failed to fetch workspaces:", error)
    }
  },
  fetchHome: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/home`)
      const data = await response.json()
      set({ homeData: data.home || null })
    } catch (error) {
      console.error("Failed to fetch home data:", error)
    }
  },
  inviteMember: async (email) => {
    const workspaceId = get().currentWorkspace?.id
    if (!workspaceId) return

    try {
      await fetch(`${API_BASE_URL}/workspaces/${workspaceId}/invites`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email })
      })
    } catch (error) {
      console.error("Failed to invite member:", error)
      throw error
    }
  }
}))
