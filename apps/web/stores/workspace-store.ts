import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"
import { normalizeWorkspaceViewsResponse } from "@/lib/workspace-views"
import type { HomeData, Workspace, WorkspaceView } from "@/types"

interface WorkspaceState {
  workspaces: Workspace[]
  currentWorkspace: Workspace | null
  homeData: HomeData | null
  workspaceViews: WorkspaceView[]
  isLoadingWorkspaceViews: boolean
  workspaceViewsError: string | null
  isHomeExecutionStale: boolean
  setCurrentWorkspace: (workspace: Workspace) => void
  fetchWorkspaces: () => Promise<void>
  fetchHome: () => Promise<void>
  fetchWorkspaceViews: () => Promise<void>
  markHomeExecutionStale: () => void
  inviteMember: (email: string) => Promise<void>
}

export const useWorkspaceStore = create<WorkspaceState>((set, get) => ({
  workspaces: [],
  currentWorkspace: null,
  homeData: null,
  workspaceViews: [],
  isLoadingWorkspaceViews: false,
  workspaceViewsError: null,
  isHomeExecutionStale: false,
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
      set({ homeData: (data.home || data) as HomeData, isHomeExecutionStale: false })
    } catch (error) {
      console.error("Failed to fetch home data:", error)
    }
  },
  fetchWorkspaceViews: async () => {
    set({ isLoadingWorkspaceViews: true, workspaceViewsError: null })
    try {
      const response = await fetch(`${API_BASE_URL}/workspace/views?limit=8`)
      if (!response.ok) {
        throw new Error("Failed to fetch workspace views")
      }
      const data = await response.json()
      set({
        workspaceViews: normalizeWorkspaceViewsResponse(data),
        isLoadingWorkspaceViews: false,
        workspaceViewsError: null,
      })
    } catch (error) {
      console.error("Failed to fetch workspace views:", error)
      set({
        workspaceViews: [],
        isLoadingWorkspaceViews: false,
        workspaceViewsError: error instanceof Error ? error.message : "Failed to fetch workspace views",
      })
    }
  },
  markHomeExecutionStale: () => set({ isHomeExecutionStale: true }),
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
