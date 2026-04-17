import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"

interface SearchState {
  results: {
    channels: any[]
    users: any[]
    messages: any[]
    dms: any[]
  }
  isSearching: boolean
  query: string
  search: (q: string) => Promise<void>
  clearResults: () => void
}

export const useSearchStore = create<SearchState>((set) => ({
  results: { channels: [], users: [], messages: [], dms: [] },
  isSearching: false,
  query: "",
  search: async (q: string) => {
    if (!q.trim()) {
      set({ results: { channels: [], users: [], messages: [], dms: [] }, query: q, isSearching: false })
      return
    }
    set({ isSearching: true, query: q })
    try {
      const response = await fetch(`${API_BASE_URL}/search?q=${encodeURIComponent(q)}`)
      const data = await response.json()
      set({ results: data.results, isSearching: false })
    } catch (error) {
      console.error("Search failed:", error)
      set({ isSearching: false })
    }
  },
  clearResults: () => set({ results: { channels: [], users: [], messages: [], dms: [] }, query: "" })
}))
