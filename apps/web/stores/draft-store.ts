import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"

export interface Draft {
  scope: string
  content: string
  updatedAt: string
}

interface DraftState {
  drafts: Record<string, Draft>
  fetchDrafts: () => Promise<void>
  saveDraft: (scope: string, content: string) => Promise<void>
  getDraft: (scope: string) => string | undefined
}

export const useDraftStore = create<DraftState>((set, get) => ({
  drafts: {},
  fetchDrafts: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/drafts`)
      const data = await response.json()
      const draftsMap: Record<string, Draft> = {}
      data.drafts.forEach((d: any) => {
        draftsMap[d.scope] = {
          scope: d.scope,
          content: d.content,
          updatedAt: d.updated_at
        }
      })
      set({ drafts: draftsMap })
    } catch (error) {
      console.error("Failed to fetch drafts:", error)
    }
  },
  saveDraft: async (scope, content) => {
    try {
      // Optimistic update
      set((state) => ({
        drafts: {
          ...state.drafts,
          [scope]: {
            scope,
            content,
            updatedAt: new Date().toISOString()
          }
        }
      }))

      await fetch(`${API_BASE_URL}/drafts/${scope}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content })
      })
    } catch (error) {
      console.error("Failed to save draft:", error)
    }
  },
  getDraft: (scope) => {
    return get().drafts[scope]?.content
  }
}))
