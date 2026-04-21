import { create } from "zustand"
import { UserGroup, Workflow, Tool, WorkflowRun } from "@/types"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"

const mapRun = (r: any): WorkflowRun => ({
  ...r,
  workflowId: r.workflow_id,
  workflowName: r.workflow_name,
  startedAt: r.started_at,
  finishedAt: r.finished_at,
  durationMs: r.duration_ms,
  triggeredBy: r.triggered_by
})

interface DirectoryState {
  userGroups: UserGroup[]
  activeGroup: UserGroup | null
  workflows: Workflow[]
  workflowRuns: any[]
  tools: Tool[]
  isLoading: boolean
  isWorkflowLoading: boolean
  
  fetchUserGroups: () => Promise<void>
  fetchGroupDetail: (id: string) => Promise<void>
  fetchWorkflows: () => Promise<void>
  fetchWorkflowRuns: () => Promise<void>
  triggerWorkflow: (id: string) => Promise<void>
  fetchTools: () => Promise<void>
  createGroup: (name: string, handle: string, description?: string) => Promise<void>
  updateGroup: (id: string, updates: Partial<UserGroup>) => Promise<void>
  deleteGroup: (id: string) => Promise<void>
  fetchGroupMembers: (id: string) => Promise<void>
  addGroupMember: (groupId: string, userId: string, role?: string) => Promise<void>
  removeGroupMember: (groupId: string, userId: string) => Promise<void>
  fetchGroupMentions: (q: string) => Promise<UserGroup[]>
  fetchWorkflowRunDetail: (id: string) => Promise<WorkflowRun | null>
  cancelWorkflowRun: (id: string) => Promise<void>
  retryWorkflowRun: (id: string) => Promise<void>
  fetchWorkflowRunLogs: (id: string) => Promise<string[]>
  deleteWorkflowRun: (id: string) => Promise<void>
}

export const useDirectoryStore = create<DirectoryState>((set, get) => ({
  userGroups: [],
  activeGroup: null,
  workflows: [],
  workflowRuns: [],
  tools: [],
  isLoading: false,
  isWorkflowLoading: false,

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

  fetchWorkflowRuns: async () => {
    try {
      set({ isWorkflowLoading: true })
      const response = await fetch(`${API_BASE_URL}/workflows/runs`)
      const data = await response.json()
      set({ workflowRuns: (data.runs || []).map(mapRun), isWorkflowLoading: false })
    } catch (error) {
      console.error("Failed to fetch workflow runs:", error)
      set({ isWorkflowLoading: false })
    }
  },

  triggerWorkflow: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/workflows/${id}/runs`, {
        method: "POST"
      })
      if (!response.ok) throw new Error("Trigger failed")
      toast.success("Workflow triggered successfully")
    } catch (error) {
      console.error("Failed to trigger workflow:", error)
      toast.error("Failed to trigger workflow")
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
  },

  createGroup: async (name, handle, description) => {
    try {
      const response = await fetch(`${API_BASE_URL}/user-groups`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, handle, description })
      })
      if (!response.ok) throw new Error("Create failed")
      await get().fetchUserGroups()
      toast.success("Group created successfully")
    } catch (error) {
      console.error("Failed to create group:", error)
      toast.error("Failed to create group")
    }
  },

  updateGroup: async (id, updates) => {
    try {
      const response = await fetch(`${API_BASE_URL}/user-groups/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(updates)
      })
      if (!response.ok) throw new Error("Update failed")
      await get().fetchUserGroups()
      toast.success("Group updated successfully")
    } catch (error) {
      console.error("Failed to update group:", error)
      toast.error("Failed to update group")
    }
  },

  deleteGroup: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/user-groups/${id}`, {
        method: "DELETE"
      })
      if (!response.ok) throw new Error("Delete failed")
      set((state) => ({
        userGroups: state.userGroups.filter(g => g.id !== id),
        activeGroup: state.activeGroup?.id === id ? null : state.activeGroup
      }))
      toast.success("Group deleted successfully")
    } catch (error) {
      console.error("Failed to delete group:", error)
      toast.error("Failed to delete group")
    }
  },

  fetchGroupMembers: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/user-groups/${id}/members`)
      const data = await response.json()
      if (get().activeGroup?.id === id) {
        set({ activeGroup: { ...get().activeGroup!, members: data.members } })
      }
    } catch (error) {
      console.error("Failed to fetch group members:", error)
    }
  },

  addGroupMember: async (groupId, userId, role = "member") => {
    try {
      const response = await fetch(`${API_BASE_URL}/user-groups/${groupId}/members`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user_id: userId, role })
      })
      if (!response.ok) throw new Error("Add member failed")
      await get().fetchGroupMembers(groupId)
      toast.success("Member added to group")
    } catch (error) {
      console.error("Failed to add group member:", error)
      toast.error("Failed to add member")
    }
  },

  removeGroupMember: async (groupId, userId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/user-groups/${groupId}/members/${userId}`, {
        method: "DELETE"
      })
      if (!response.ok) throw new Error("Remove member failed")
      await get().fetchGroupMembers(groupId)
      toast.success("Member removed from group")
    } catch (error) {
      console.error("Failed to remove group member:", error)
      toast.error("Failed to remove member")
    }
  },

  fetchGroupMentions: async (q) => {
    try {
      const response = await fetch(`${API_BASE_URL}/user-groups/mentions?q=${encodeURIComponent(q)}`)
      const data = await response.json()
      return data.groups || []
    } catch (error) {
      console.error("Failed to fetch group mentions:", error)
      return []
    }
  },

  fetchWorkflowRunDetail: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/workflows/runs/${id}`)
      const data = await response.json()
      return mapRun(data.run)
    } catch (error) {
      console.error("Failed to fetch workflow run detail:", error)
      return null
    }
  },

  cancelWorkflowRun: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/workflows/runs/${id}/cancel`, {
        method: "POST"
      })
      if (!response.ok) throw new Error("Cancel failed")
      toast.success("Workflow run cancellation requested")
    } catch (error) {
      console.error("Failed to cancel workflow run:", error)
      toast.error("Failed to cancel run")
    }
  },

  retryWorkflowRun: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/workflows/runs/${id}/retry`, {
        method: "POST"
      })
      if (!response.ok) throw new Error("Retry failed")
      toast.success("Workflow run retry requested")
    } catch (error) {
      console.error("Failed to retry workflow run:", error)
      toast.error("Failed to retry run")
    }
  },

  fetchWorkflowRunLogs: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/workflows/runs/${id}/logs`)
      const data = await response.json()
      return data.logs || []
    } catch (error) {
      console.error("Failed to fetch workflow run logs:", error)
      return []
    }
  },

  deleteWorkflowRun: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/workflows/runs/${id}`, {
        method: "DELETE"
      })
      if (!response.ok) throw new Error("Delete failed")
      set((state) => ({
        workflowRuns: state.workflowRuns.filter(r => r.id !== id)
      }))
      toast.success("Workflow run log deleted")
    } catch (error) {
      console.error("Failed to delete workflow run:", error)
      toast.error("Failed to delete log")
    }
  }
}))
