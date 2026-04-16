import { create } from "zustand"
import { Workspace } from "@/types"
import { API_BASE_URL } from "@/lib/constants"

interface WorkspaceState {
  workspaces: Workspace[]
  currentWorkspace: Workspace | null
  setCurrentWorkspace: (workspace: Workspace) => void
  fetchWorkspaces: () => Promise<void>
}

export const useWorkspaceStore = create<WorkspaceState>((set) => ({
  workspaces: [],
  currentWorkspace: null,
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
}))
