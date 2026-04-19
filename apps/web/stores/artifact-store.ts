import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"

export interface Artifact {
  id: string
  title: string
  content: string
  type: string
  channelId: string
  userId: string
  version: number
  createdAt: string
  updatedAt: string
}

interface ArtifactState {
  artifacts: Artifact[]
  activeArtifact: Artifact | null
  isLoading: boolean
  fetchArtifacts: (channelId: string) => Promise<void>
  fetchArtifactDetail: (id: string) => Promise<void>
  createArtifact: (data: Partial<Artifact>) => Promise<Artifact | null>
  updateArtifact: (id: string, updates: Partial<Artifact>) => Promise<void>
  generateAIArtifact: (prompt: string, channelId: string) => Promise<void>
  setActiveArtifact: (artifact: Artifact | null) => void
  updateArtifactLocally: (artifact: Artifact) => void
}

const mapArtifact = (a: any): Artifact => ({
  ...a,
  channelId: a.channel_id,
  userId: a.user_id,
  createdAt: a.created_at,
  updatedAt: a.updated_at
})

export const useArtifactStore = create<ArtifactState>((set, get) => ({
  artifacts: [],
  activeArtifact: null,
  isLoading: false,

  fetchArtifacts: async (channelId) => {
    try {
      set({ isLoading: true })
      const response = await fetch(`${API_BASE_URL}/artifacts?channel_id=${channelId}`)
      const data = await response.json()
      set({ artifacts: (data.artifacts || []).map(mapArtifact), isLoading: false })
    } catch (error) {
      console.error("Failed to fetch artifacts:", error)
      set({ isLoading: false })
    }
  },

  fetchArtifactDetail: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/artifacts/${id}`)
      const data = await response.json()
      const artifact = mapArtifact(data.artifact)
      set({ activeArtifact: artifact })
    } catch (error) {
      console.error("Failed to fetch artifact detail:", error)
    }
  },

  createArtifact: async (data) => {
    try {
      const response = await fetch(`${API_BASE_URL}/artifacts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          ...data,
          channel_id: data.channelId
        })
      })
      const result = await response.json()
      const newArtifact = mapArtifact(result.artifact)
      set((state) => ({ artifacts: [newArtifact, ...state.artifacts] }))
      return newArtifact
    } catch (error) {
      console.error("Failed to create artifact:", error)
      return null
    }
  },

  updateArtifact: async (id, updates) => {
    try {
      const response = await fetch(`${API_BASE_URL}/artifacts/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(updates)
      })
      const data = await response.json()
      const updated = mapArtifact(data.artifact)
      set((state) => ({
        artifacts: state.artifacts.map(a => a.id === id ? updated : a),
        activeArtifact: state.activeArtifact?.id === id ? updated : state.activeArtifact
      }))
      toast.success("Artifact updated")
    } catch (error) {
      console.error("Failed to update artifact:", error)
      toast.error("Failed to update artifact")
    }
  },

  generateAIArtifact: async (prompt, channelId) => {
    try {
      set({ isLoading: true })
      const response = await fetch(`${API_BASE_URL}/ai/canvas/generate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt, channel_id: channelId })
      })
      const data = await response.json()
      const artifact = mapArtifact(data.artifact)
      set((state) => ({ 
        artifacts: [artifact, ...state.artifacts],
        activeArtifact: artifact,
        isLoading: false
      }))
      toast.success("AI Canvas generated")
    } catch (error) {
      console.error("Failed to generate AI artifact:", error)
      set({ isLoading: false })
      toast.error("Failed to generate AI artifact")
    }
  },

  setActiveArtifact: (artifact) => set({ activeArtifact: artifact }),

  updateArtifactLocally: (artifact) => {
    const mapped = mapArtifact(artifact)
    set((state) => ({
      artifacts: state.artifacts.map(a => a.id === mapped.id ? mapped : a),
      activeArtifact: state.activeArtifact?.id === mapped.id ? mapped : state.activeArtifact
    }))
  }
}))
