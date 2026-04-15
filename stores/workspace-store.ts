import { create } from "zustand"
import { Workspace } from "@/types"
import { WORKSPACES } from "@/lib/mock-data"

interface WorkspaceState {
  workspaces: Workspace[]
  currentWorkspace: Workspace | null
  setCurrentWorkspace: (workspace: Workspace) => void
}

export const useWorkspaceStore = create<WorkspaceState>((set) => ({
  workspaces: WORKSPACES,
  currentWorkspace: WORKSPACES[0],
  setCurrentWorkspace: (workspace) => set({ currentWorkspace: workspace }),
}))
