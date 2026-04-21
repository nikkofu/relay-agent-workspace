import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"
import { FileAudit, FileAsset, FileComment, FileShare, FileChunk, FileCitation, FileSearchResult } from "@/types"

interface FileState {
  files: FileAsset[]
  archivedFiles: FileAsset[]
  starredFiles: FileAsset[]
  isUploading: boolean
  isLoadingArchive: boolean
  fetchFiles: (params?: { channelId?: string, q?: string, uploaderId?: string, contentType?: string }) => Promise<void>
  fetchArchivedFiles: (params?: { channelId?: string, q?: string, uploaderId?: string, contentType?: string }) => Promise<void>
  uploadFile: (file: File, channelId?: string) => Promise<FileAsset | null>
  archiveFile: (id: string, isArchived: boolean) => Promise<void>
  deleteFile: (id: string) => Promise<void>
  fetchFileAuditHistory: (id: string) => Promise<FileAudit[]>
  updateFileRetention: (id: string, retentionDays: number) => Promise<void>
  fetchFilePreview: (id: string) => Promise<any>
  fetchFileComments: (id: string) => Promise<FileComment[]>
  createFileComment: (id: string, content: string) => Promise<FileComment | null>
  toggleFileStar: (id: string) => Promise<boolean | null>
  fetchStarredFiles: () => Promise<void>
  shareFile: (id: string, channelId: string, comment?: string, threadId?: string) => Promise<FileShare | null>
  fetchFileShares: (id: string) => Promise<FileShare[]>
  updateFileKnowledge: (id: string, meta: { knowledge_state?: string; source_kind?: string; summary?: string; tags?: string[] }) => Promise<FileAsset | null>
  fetchMessageFiles: (messageId: string) => Promise<FileAsset[]>
  fetchFileExtraction: (id: string) => Promise<any>
  rebuildFileExtraction: (id: string) => Promise<void>
  fetchFileExtractedContent: (id: string) => Promise<{ text?: string; summary?: string } | null>
  fetchFileChunks: (id: string) => Promise<FileChunk[]>
  fetchFileCitations: (id: string) => Promise<FileCitation[]>
  searchFiles: (q: string) => Promise<FileSearchResult[]>
  updateFileLocally: (id: string, updates: Partial<FileAsset>) => void
}

export const useFileStore = create<FileState>((set) => ({
  files: [],
  archivedFiles: [],
  starredFiles: [],
  isUploading: false,
  isLoadingArchive: false,

  fetchFiles: async (params) => {
    try {
      const queryParams = new URLSearchParams()
      if (params?.channelId) queryParams.append('channel_id', params.channelId)
      if (params?.q) queryParams.append('q', params.q)
      if (params?.uploaderId) queryParams.append('uploader_id', params.uploaderId)
      if (params?.contentType) queryParams.append('content_type', params.contentType)

      const url = `${API_BASE_URL}/files${queryParams.toString() ? `?${queryParams.toString()}` : ''}`
      const response = await fetch(url)
      const data = await response.json()
      set({ files: data.files || [] })
    } catch (error) {
      console.error("Failed to fetch files:", error)
    }
  },

  fetchArchivedFiles: async (params) => {
    try {
      set({ isLoadingArchive: true })
      const queryParams = new URLSearchParams()
      if (params?.channelId) queryParams.append('channel_id', params.channelId)
      if (params?.q) queryParams.append('q', params.q)
      if (params?.uploaderId) queryParams.append('uploader_id', params.uploaderId)
      if (params?.contentType) queryParams.append('content_type', params.contentType)

      const url = `${API_BASE_URL}/files/archive${queryParams.toString() ? `?${queryParams.toString()}` : ''}`
      const response = await fetch(url)
      const data = await response.json()
      set({ archivedFiles: data.files || [], isLoadingArchive: false })
    } catch (error) {
      console.error("Failed to fetch archived files:", error)
      set({ isLoadingArchive: false })
    }
  },

  uploadFile: async (file, channelId) => {
    try {
      set({ isUploading: true })
      const formData = new FormData()
      formData.append("file", file)
      if (channelId) formData.append("channel_id", channelId)

      const response = await fetch(`${API_BASE_URL}/files/upload`, {
        method: "POST",
        body: formData,
      })

      if (!response.ok) throw new Error("Upload failed")

      const data = await response.json()
      const newFile = data.file
      set((state) => ({ 
        files: [newFile, ...state.files],
        isUploading: false 
      }))
      toast.success("File uploaded successfully")
      return newFile
    } catch (error) {
      console.error("Failed to upload file:", error)
      set({ isUploading: false })
      toast.error("Failed to upload file")
      return null
    }
  },

  archiveFile: async (id, isArchived) => {
    try {
      const response = await fetch(`${API_BASE_URL}/files/${id}/archive`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ is_archived: isArchived })
      })
      if (!response.ok) throw new Error("Archive failed")

      const data = await response.json()
      const updated = data.file
      set((state) => ({
        files: state.files.filter(f => f.id !== id),
        archivedFiles: isArchived 
          ? [updated, ...state.archivedFiles] 
          : state.archivedFiles.filter(f => f.id !== id)
      }))
      toast.success(isArchived ? "File archived" : "File restored")
    } catch (error) {
      console.error("Failed to archive file:", error)
      toast.error("Failed to archive file")
    }
  },

  deleteFile: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/files/${id}`, {
        method: "DELETE"
      })
      if (!response.ok) throw new Error("Delete failed")

      set((state) => ({
        files: state.files.filter(f => f.id !== id),
        archivedFiles: state.archivedFiles.filter(f => f.id !== id)
      }))
      toast.success("File deleted permanently")
    } catch (error) {
      console.error("Failed to delete file:", error)
      toast.error("Failed to delete file")
    }
  },

  fetchFileAuditHistory: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/files/${id}/audit`)
      const data = await response.json()
      return data.events || data.audit_history || []
    } catch (error) {
      console.error("Failed to fetch file audit history:", error)
      return []
    }
  },

  updateFileRetention: async (id, retentionDays) => {
    try {
      const response = await fetch(`${API_BASE_URL}/files/${id}/retention`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ retention_days: retentionDays })
      })
      if (!response.ok) throw new Error("Retention update failed")
      toast.success("Retention policy updated")
    } catch (error) {
      console.error("Failed to update file retention:", error)
      toast.error("Failed to update retention")
    }
  },

  fetchFilePreview: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/files/${id}/preview`)
      const data = await response.json()
      return data.preview
    } catch (error) {
      console.error("Failed to fetch file preview:", error)
      return null
    }
  },

  fetchFileComments: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/files/${id}/comments`)
      const data = await response.json()
      return data.comments || []
    } catch (error) {
      console.error("Failed to fetch file comments:", error)
      return []
    }
  },

  createFileComment: async (id, content) => {
    try {
      const response = await fetch(`${API_BASE_URL}/files/${id}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content }),
      })
      if (!response.ok) throw new Error("Failed to create comment")
      const data = await response.json()
      return data.comment || null
    } catch (error) {
      console.error("Failed to create file comment:", error)
      toast.error("Failed to post comment")
      return null
    }
  },

  toggleFileStar: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/files/${id}/star`, { method: "POST" })
      if (!response.ok) throw new Error("Failed to toggle star")
      const data = await response.json()
      const starred: boolean = data.starred
      set((state) => ({
        files: state.files.map(f => f.id === id ? { ...f, starred } : f),
        starredFiles: starred
          ? [...state.starredFiles.filter(f => f.id !== id), data.file]
          : state.starredFiles.filter(f => f.id !== id),
      }))
      toast.success(starred ? "File starred" : "Star removed")
      return starred
    } catch (error) {
      console.error("Failed to toggle file star:", error)
      toast.error("Failed to update star")
      return null
    }
  },

  fetchStarredFiles: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/files/starred`)
      const data = await response.json()
      set({ starredFiles: data.files || [] })
    } catch (error) {
      console.error("Failed to fetch starred files:", error)
    }
  },

  shareFile: async (id, channelId, comment, threadId) => {
    try {
      const body: any = { channel_id: channelId }
      if (comment) body.comment = comment
      if (threadId) body.thread_id = threadId
      const response = await fetch(`${API_BASE_URL}/files/${id}/share`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      })
      if (!response.ok) throw new Error("Failed to share file")
      const data = await response.json()
      toast.success("File shared to channel")
      return data.share || null
    } catch (error) {
      console.error("Failed to share file:", error)
      toast.error("Failed to share file")
      return null
    }
  },

  fetchFileShares: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/files/${id}/shares`)
      const data = await response.json()
      return data.shares || []
    } catch (error) {
      console.error("Failed to fetch file shares:", error)
      return []
    }
  },

  updateFileLocally: (id, updates) => {
    set((state) => ({
      files: state.files.map(f => f.id === id ? { ...f, ...updates } : f),
      starredFiles: state.starredFiles.map(f => f.id === id ? { ...f, ...updates } : f),
    }))
  },

  fetchFileExtraction: async (id) => {
    try {
      const res = await fetch(`${API_BASE_URL}/files/${id}/extraction`)
      if (!res.ok) return null
      const data = await res.json()
      const ext = data.extraction || data
      set((state) => ({
        files: state.files.map(f => f.id === id ? { ...f, extraction_status: ext.status, last_indexed_at: ext.last_indexed_at, needs_ocr: ext.needs_ocr, is_searchable: ext.is_searchable, is_citable: ext.is_citable, content_summary: ext.content_summary } : f),
      }))
      return ext
    } catch (error) {
      console.error("Failed to fetch file extraction:", error)
      return null
    }
  },

  rebuildFileExtraction: async (id) => {
    try {
      await fetch(`${API_BASE_URL}/files/${id}/extraction/rebuild`, { method: "POST" })
      toast.success("Extraction rebuild triggered")
    } catch (error) {
      console.error("Failed to rebuild extraction:", error)
      toast.error("Failed to trigger rebuild")
    }
  },

  fetchFileExtractedContent: async (id) => {
    try {
      const res = await fetch(`${API_BASE_URL}/files/${id}/extracted-content`)
      if (!res.ok) return null
      const data = await res.json()
      return { text: data.text || data.content, summary: data.summary }
    } catch (error) {
      console.error("Failed to fetch extracted content:", error)
      return null
    }
  },

  fetchFileChunks: async (id) => {
    try {
      const res = await fetch(`${API_BASE_URL}/files/${id}/chunks`)
      if (!res.ok) return []
      const data = await res.json()
      return data.chunks || []
    } catch (error) {
      console.error("Failed to fetch file chunks:", error)
      return []
    }
  },

  fetchFileCitations: async (id) => {
    try {
      const res = await fetch(`${API_BASE_URL}/files/${id}/citations`)
      if (!res.ok) return []
      const data = await res.json()
      return data.citations || []
    } catch (error) {
      console.error("Failed to fetch file citations:", error)
      return []
    }
  },

  searchFiles: async (q) => {
    try {
      const res = await fetch(`${API_BASE_URL}/search/files?q=${encodeURIComponent(q)}`)
      if (!res.ok) return []
      const data = await res.json()
      return data.results || data.files || []
    } catch (error) {
      console.error("Failed to search files:", error)
      return []
    }
  },

  fetchMessageFiles: async (messageId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/messages/${messageId}/files`)
      if (!response.ok) return []
      const data = await response.json()
      return data.files || []
    } catch (error) {
      console.error("Failed to fetch message files:", error)
      return []
    }
  },

  updateFileKnowledge: async (id, meta) => {
    try {
      const response = await fetch(`${API_BASE_URL}/files/${id}/knowledge`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(meta),
      })
      if (!response.ok) throw new Error("Failed to update knowledge")
      const data = await response.json()
      const updated: FileAsset = data.file
      set((state) => ({
        files: state.files.map(f => f.id === id ? { ...f, ...updated } : f),
      }))
      toast.success("Knowledge metadata updated")
      return updated
    } catch (error) {
      console.error("Failed to update file knowledge:", error)
      toast.error("Failed to update knowledge metadata")
      return null
    }
  },
}))
