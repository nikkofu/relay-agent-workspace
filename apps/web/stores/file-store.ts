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
}

interface FileState {
  files: FileAsset[]
  isUploading: boolean
  fetchFiles: (channelId: string) => Promise<void>
  uploadFile: (file: File, channelId?: string) => Promise<FileAsset | null>
}

export const useFileStore = create<FileState>((set, get) => ({
  files: [],
  isUploading: false,

  fetchFiles: async (channelId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/files?channel_id=${channelId}`)
      const data = await response.json()
      set({ files: data.files || [] })
    } catch (error) {
      console.error("Failed to fetch files:", error)
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
}))
