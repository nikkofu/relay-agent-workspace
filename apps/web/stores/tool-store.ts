import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"

// Phase 66 T07 — frozen Codex contract (Q3):
//   writeback_target = "message" | "list_item"
//   writeback payload is per-target ("message" → { channel_id, thread_id? }, "list_item" → { list_id })
export type WritebackTarget = "message" | "list_item"

export interface WritebackMessageInput {
  channel_id: string
  thread_id?: string
}

export interface WritebackListItemInput {
  list_id: string
}

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
  // Phase 66 T07: writeback-target telemetry persisted on each run
  writebackTarget?: WritebackTarget
  writeback?: Record<string, any>
}

interface ToolState {
  toolRuns: ToolRun[]
  activeRun: ToolRun | null
  isLoading: boolean
  
  fetchToolRuns: (channelId?: string) => Promise<void>
  fetchRunDetail: (id: string) => Promise<void>
  addRunLocally: (run: ToolRun) => void
  updateRunLocally: (run: ToolRun) => void
  executeTool: (
    toolId: string,
    input: any,
    channelId?: string,
    writeback?: { target: WritebackTarget, payload: WritebackMessageInput | WritebackListItemInput }
  ) => Promise<ToolRun | null>
}

// Exported so the realtime WS handler (`use-websocket.ts`) can normalise
// raw snake_case `tool.run.*` payloads into the `ToolRun` shape the store
// expects before invoking the local-only update actions. Without this,
// WS-delivered runs lose `toolId` / `startedAt` / `durationMs` / etc.
export const mapToolRun = (r: any): ToolRun => ({
  ...r,
  toolId: r.tool_id,
  toolName: r.tool_name,
  toolKey: r.tool_key,
  startedAt: r.started_at,
  finishedAt: r.finished_at,
  durationMs: r.duration_ms,
  userId: r.user_id,
  channelId: r.channel_id,
  // Phase 66 T07: hydrate writeback telemetry
  writebackTarget: r.writeback_target || undefined,
  writeback: r.writeback || undefined,
})

export const useToolStore = create<ToolState>((set, get) => ({
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

  addRunLocally: (run) => {
    set((state) => {
      if (state.toolRuns.find(r => r.id === run.id)) return state
      return { toolRuns: [run, ...state.toolRuns].sort((a, b) => 
        new Date(b.startedAt).getTime() - new Date(a.startedAt).getTime()
      ) }
    })
  },

  updateRunLocally: (run) => {
    set((state) => ({
      toolRuns: state.toolRuns.map(r => r.id === run.id ? run : r),
      activeRun: state.activeRun?.id === run.id ? run : state.activeRun
    }))
  },

  executeTool: async (toolId, input, channelId, writeback) => {
    try {
      const body: Record<string, any> = { input, channel_id: channelId }
      if (writeback) {
        body.writeback_target = writeback.target
        body.writeback = writeback.payload
      }
      const response = await fetch(`${API_BASE_URL}/tools/${toolId}/execute`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body)
      })
      const data = await response.json()
      if (!response.ok) throw new Error(data.message || "Execution failed")

      const newRun = mapToolRun(data.run)
      
      // Phase 67B: addRunLocally ensures immediate UI feedback
      get().addRunLocally(newRun)
      
      const wbLabel = writeback?.target === "message"
        ? " → message"
        : writeback?.target === "list_item"
          ? " → list item"
          : ""
      toast.success(`${newRun.toolName} execution started${wbLabel}`)
      return newRun
    } catch (error) {
      console.error("Failed to execute tool:", error)
      toast.error("Tool execution failed")
      return null
    }
  }
}))
