import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"

interface LaterItem {
  id: string
  message: any
  channel: any
  user: any
  savedAt: string
}

interface LaterState {
  items: LaterItem[]
  fetchLaterItems: () => Promise<void>
}

export const useLaterStore = create<LaterState>((set) => ({
  items: [],
  fetchLaterItems: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/later`)
      const data = await response.json()
      set({ items: data.items.map((i: any) => ({
        ...i,
        id: i.id || i.message?.id,
        savedAt: i.saved_at
      })) })
    } catch (error) {
      console.error("Failed to fetch later items:", error)
    }
  }
}))
