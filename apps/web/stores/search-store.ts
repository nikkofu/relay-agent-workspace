import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"

export interface SearchSuggestion {
  id: string
  type: 'channel' | 'user' | 'message' | 'artifact' | 'file'
  label: string
  sublabel?: string
}

interface SearchState {
  results: {
    channels: any[]
    users: any[]
    messages: any[]
    dms: any[]
    artifacts: any[]
    files: any[]
  }
  intelligentResults: any[]
  suggestions: SearchSuggestion[]
  isSearching: boolean
  isIntelligentSearching: boolean
  isLoadingSuggestions: boolean
  query: string
  search: (q: string) => Promise<void>
  intelligentSearch: (q: string) => Promise<void>
  fetchSuggestions: (q: string) => Promise<void>
  clearResults: () => void
}

export const useSearchStore = create<SearchState>((set, get) => ({
  results: { channels: [], users: [], messages: [], dms: [], artifacts: [], files: [] },
  intelligentResults: [],
  suggestions: [],
  isSearching: false,
  isIntelligentSearching: false,
  isLoadingSuggestions: false,
  query: "",

  search: async (q: string) => {
    if (!q.trim()) {
      set({ results: { channels: [], users: [], messages: [], dms: [], artifacts: [], files: [] }, query: q, isSearching: false })
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

  intelligentSearch: async (q: string) => {
    if (!q.trim()) {
      set({ intelligentResults: [], query: q, isIntelligentSearching: false })
      return
    }
    set({ isIntelligentSearching: true, query: q })
    try {
      const response = await fetch(`${API_BASE_URL}/search/intelligent?q=${encodeURIComponent(q)}`)
      const data = await response.json()
      set({ intelligentResults: data.results || [], isIntelligentSearching: false })
    } catch (error) {
      console.error("Intelligent search failed:", error)
      set({ isIntelligentSearching: false })
    }
  },

  fetchSuggestions: async (q: string) => {
    if (!q.trim() || q.length < 2) {
      set({ suggestions: [] })
      return
    }
    set({ isLoadingSuggestions: true })
    try {
      const response = await fetch(`${API_BASE_URL}/search/suggestions?q=${encodeURIComponent(q)}`)
      const data = await response.json()
      set({ suggestions: data.suggestions || [], isLoadingSuggestions: false })
    } catch (error) {
      console.error("Failed to fetch suggestions:", error)
      set({ isLoadingSuggestions: false })
    }
  },

  clearResults: () => set({ 
    results: { channels: [], users: [], messages: [], dms: [], artifacts: [], files: [] }, 
    intelligentResults: [],
    suggestions: [],
    query: "" 
  })
}))
