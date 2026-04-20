import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"

export interface FileAsset {
  id: string
  name: string
  type: string
  size: number
  url: string
  channelId?: string
  userId: string
  createdAt: string
  isArchived?: boolean
}

interface FileState {
  files: FileAsset[]
  archivedFiles: FileAsset[]
  isUploading: boolean
  isLoadingArchive: boolean
  fetchFiles: (params?: { channelId?: string, q?: string, uploaderId?: string, contentType?: string }) => Promise<void>
  fetchArchivedFiles: (params?: { channelId?: string, q?: string, uploaderId?: string, contentType?: string }) => Promise<void>
  uploadFile: (file: File, channelId?: string) => Promise<FileAsset | null>
  archiveFile: (id: string, isArchived: boolean) => Promise<void>
  deleteFile: (id: string) => Promise<void>
}

export const useFileStore = create<FileState>((set) => ({
  files: [],
  archivedFiles: [],
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
  }
}))
