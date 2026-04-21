import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"
import { User } from "@/types"

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
  createdByUser?: User
  updatedByUser?: User
}

export interface ArtifactVersion {
  id: string
  artifactId: string
  version: number
  title: string
  content: string
  updatedAt: string
  updatedByUser?: User
}

export interface ArtifactDiff {
  fromVersion: number
  toVersion: number
  fromContent: string
  toContent: string
  unifiedDiff: string
  spans?: {
    kind: 'addition' | 'deletion' | 'context' | 'header'
    content: string
    fromLine?: number
    toLine?: number
  }[]
  summary: {
    added: number
    removed: number
  }
}

interface ArtifactState {
  artifacts: Artifact[]
  activeArtifact: Artifact | null
  versions: ArtifactVersion[]
  currentDiff: ArtifactDiff | null
  references: { message: any, user: any, channel: any }[]
  templates: { id: string, title: string, description: string, type: string }[]
  isLoading: boolean
  isHistoryLoading: boolean
  isDiffLoading: boolean
  isReferencesLoading: boolean
  isTemplatesLoading: boolean
  fetchArtifacts: (channelId: string) => Promise<void>
  fetchArtifactDetail: (id: string) => Promise<void>
  fetchVersions: (id: string) => Promise<void>
  fetchVersionDetail: (id: string, version: number) => Promise<ArtifactVersion | null>
  fetchDiff: (id: string, fromVersion: number, toVersion: number) => Promise<void>
  fetchReferences: (id: string) => Promise<void>
  fetchTemplates: () => Promise<void>
  restoreVersion: (id: string, version: number) => Promise<void>
  createArtifact: (data: Partial<Artifact>) => Promise<Artifact | null>
  createArtifactFromTemplate: (templateId: string, channelId: string, userId: string) => Promise<Artifact | null>
  updateArtifact: (id: string, updates: Partial<Artifact>) => Promise<void>
  generateAIArtifact: (prompt: string, channelId: string) => Promise<void>
  setActiveArtifact: (artifact: Artifact | null) => void
  updateArtifactLocally: (artifact: Artifact) => void
  clearDiff: () => void
}

const mapArtifact = (a: any): Artifact | null => {
  if (!a) return null
  return {
    ...a,
    channelId: a.channel_id,
    userId: a.user_id,
    createdAt: a.created_at,
    updatedAt: a.updated_at,
    createdByUser: a.created_by_user,
    updatedByUser: a.updated_by_user
  }
}

const mapVersion = (v: any): ArtifactVersion => ({
  ...v,
  artifactId: v.artifact_id,
  updatedAt: v.updated_at,
  updatedByUser: v.updated_by_user
})

export const useArtifactStore = create<ArtifactState>((set, get) => ({
  artifacts: [],
  activeArtifact: null,
  versions: [],
  currentDiff: null,
  references: [],
  templates: [],
  isLoading: false,
  isHistoryLoading: false,
  isDiffLoading: false,
  isReferencesLoading: false,
  isTemplatesLoading: false,

  fetchArtifacts: async (channelId) => {
    try {
      set({ isLoading: true })
      const response = await fetch(`${API_BASE_URL}/artifacts?channel_id=${channelId}`)
      const data = await response.json()
      const mapped = (data.artifacts || []).map(mapArtifact).filter(Boolean) as Artifact[]
      set({ artifacts: mapped, isLoading: false })
    } catch (error) {
      console.error("Failed to fetch artifacts:", error)
      set({ isLoading: false })
    }
  },

  fetchArtifactDetail: async (id) => {
    try {
      const channelId = get().activeArtifact?.channelId
      const url = id === "new-doc" 
        ? `${API_BASE_URL}/artifacts/new-doc${channelId ? `?channel_id=${channelId}` : ''}`
        : `${API_BASE_URL}/artifacts/${id}`
        
      const response = await fetch(url)
      const data = await response.json()
      const artifact = mapArtifact(data.artifact)
      if (artifact) set({ activeArtifact: artifact })
      if (id === "new-doc") set({ versions: [] })
    } catch (error) {
      console.error("Failed to fetch artifact detail:", error)
    }
  },

  fetchVersions: async (id) => {
    try {
      set({ isHistoryLoading: true })
      const response = await fetch(`${API_BASE_URL}/artifacts/${id}/versions`)
      const data = await response.json()
      set({ versions: (data.versions || []).map(mapVersion), isHistoryLoading: false })
    } catch (error) {
      console.error("Failed to fetch artifact versions:", error)
      set({ isHistoryLoading: false })
    }
  },

  fetchVersionDetail: async (id, version) => {
    try {
      const response = await fetch(`${API_BASE_URL}/artifacts/${id}/versions/${version}`)
      const data = await response.json()
      return mapVersion(data.version)
    } catch (error) {
      console.error("Failed to fetch artifact version detail:", error)
      return null
    }
  },

  fetchDiff: async (id, from, to) => {
    try {
      set({ isDiffLoading: true, currentDiff: null })
      const response = await fetch(`${API_BASE_URL}/artifacts/${id}/diff/${from}/${to}`)
      const data = await response.json()
      const diff = data.diff || data
      set({ 
        currentDiff: {
          fromVersion: diff.from_version,
          toVersion: diff.to_version,
          fromContent: diff.from_content,
          toContent: diff.to_content,
          unifiedDiff: diff.unified_diff,
          spans: diff.spans,
          summary: {
            added: diff.summary?.added_lines ?? diff.summary?.added ?? 0,
            removed: diff.summary?.removed_lines ?? diff.summary?.removed ?? 0,
          }
        },
        isDiffLoading: false 
      })
    } catch (error) {
      console.error("Failed to fetch artifact diff:", error)
      set({ isDiffLoading: false })
      toast.error("Failed to load comparison data")
    }
  },

  fetchReferences: async (id) => {
    try {
      set({ isReferencesLoading: true })
      const response = await fetch(`${API_BASE_URL}/artifacts/${id}/references`)
      const data = await response.json()
      set({ references: data.references || [], isReferencesLoading: false })
    } catch (error) {
      console.error("Failed to fetch artifact references:", error)
      set({ isReferencesLoading: false })
    }
  },

  restoreVersion: async (id, version) => {
    try {
      set({ isLoading: true })
      const response = await fetch(`${API_BASE_URL}/artifacts/${id}/restore/${version}`, {
        method: "POST"
      })
      if (!response.ok) throw new Error("Restore failed")
      
      const data = await response.json()
      const restored = mapArtifact(data.artifact)
      if (restored) {
        set((state) => ({
          activeArtifact: restored,
          artifacts: state.artifacts.map(a => a.id === id ? restored : a),
          isLoading: false
        }))
        // Refresh versions history
        get().fetchVersions(id)
        toast.success(`Restored to version ${version}`)
      }
    } catch (error) {
      console.error("Failed to restore artifact version:", error)
      set({ isLoading: false })
      toast.error("Failed to restore version")
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
      if (newArtifact) {
        set((state) => ({ artifacts: [newArtifact, ...state.artifacts], activeArtifact: newArtifact }))
      }
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
      if (updated) {
        set((state) => ({
          artifacts: state.artifacts.map(a => a.id === id ? updated : a),
          activeArtifact: state.activeArtifact?.id === id ? updated : state.activeArtifact
        }))
      }
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
      if (artifact) {
        set((state) => ({ 
          artifacts: [artifact, ...state.artifacts],
          activeArtifact: artifact,
          isLoading: false
        }))
      } else {
        set({ isLoading: false })
      }
      toast.success("AI Canvas generated")
    } catch (error) {
      console.error("Failed to generate AI artifact:", error)
      set({ isLoading: false })
      toast.error("Failed to generate AI artifact")
    }
  },

  fetchTemplates: async () => {
    try {
      set({ isTemplatesLoading: true })
      const response = await fetch(`${API_BASE_URL}/artifacts/templates`)
      const data = await response.json()
      set({ templates: data.templates || [], isTemplatesLoading: false })
    } catch (error) {
      console.error("Failed to fetch artifact templates:", error)
      set({ isTemplatesLoading: false })
    }
  },

  createArtifactFromTemplate: async (templateId, channelId, userId) => {
    try {
      set({ isLoading: true })
      const response = await fetch(`${API_BASE_URL}/artifacts/from-template`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          template_id: templateId,
          channel_id: channelId,
          user_id: userId
        })
      })
      const result = await response.json()
      const newArtifact = mapArtifact(result.artifact)
      if (newArtifact) {
        set((state) => ({ artifacts: [newArtifact, ...state.artifacts], activeArtifact: newArtifact, isLoading: false }))
      } else {
        set({ isLoading: false })
      }
      return newArtifact
    } catch (error) {
      console.error("Failed to create artifact from template:", error)
      set({ isLoading: false })
      return null
    }
  },

  setActiveArtifact: (artifact) => set({ activeArtifact: artifact }),

  updateArtifactLocally: (artifact) => {
    const mapped = mapArtifact(artifact)
    if (mapped) {
      set((state) => ({
        artifacts: state.artifacts.map(a => a.id === mapped.id ? mapped : a),
        activeArtifact: state.activeArtifact?.id === mapped.id ? mapped : state.activeArtifact
      }))
    }
  },

  clearDiff: () => set({ currentDiff: null })
}))
