import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"

export interface ToolRun {
  id: string
  toolId: string
  toolName: string
  toolKey: string
  status: 'pending' | 'running' | 'success' | 'failed'
  input: any
  output?: any
  logs: string[]
  userId: string
  channelId?: string
  startedAt: string
  finishedAt?: string
  durationMs?: number
}

interface ToolState {
  toolRuns: ToolRun[]
  activeRun: ToolRun | null
  isLoading: boolean
  
  fetchToolRuns: (channelId?: string) => Promise<void>
  fetchRunDetail: (id: string) => Promise<void>
  executeTool: (toolId: string, input: any, channelId?: string) => Promise<void>
}

const mapToolRun = (r: any): ToolRun => ({
  ...r,
  toolId: r.tool_id,
  toolName: r.tool_name,
  toolKey: r.tool_key,
  startedAt: r.started_at,
  finishedAt: r.finished_at,
  durationMs: r.duration_ms,
  userId: r.user_id,
  channelId: r.channel_id
})

export const useToolStore = create<ToolState>((set) => ({
  toolRuns: [],
  activeRun: null,
  isLoading: false,

  fetchToolRuns: async (channelId) => {
    try {
      set({ isLoading: true })
      const url = `${API_BASE_URL}/tools/runs${channelId ? `?channel_id=${channelId}` : ''}`
      const response = await fetch(url)
      const data = await response.json()
      set({ toolRuns: (data.runs || []).map(mapToolRun), isLoading: false })
    } catch (error) {
      console.error("Failed to fetch tool runs:", error)
      set({ isLoading: false })
    }
  },

  fetchRunDetail: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/tools/runs/${id}`)
      const data = await response.json()
      set({ activeRun: mapToolRun(data.run) })
    } catch (error) {
      console.error("Failed to fetch tool run detail:", error)
    }
  },

  executeTool: async (toolId, input, channelId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/tools/${toolId}/execute`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ input, channel_id: channelId })
      })
      const data = await response.json()
      if (!response.ok) throw new Error(data.message || "Execution failed")
      
      const newRun = mapToolRun(data.run)
      set((state) => ({ toolRuns: [newRun, ...state.toolRuns] }))
      toast.success(`${newRun.toolName} execution started`)
    } catch (error) {
      console.error("Failed to execute tool:", error)
      toast.error("Tool execution failed")
    }
  }
}))
