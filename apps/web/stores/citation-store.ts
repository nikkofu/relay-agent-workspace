import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"
import type { CitationEvidence, EvidenceKind } from "@/types"

interface CitationState {
  results: CitationEvidence[]
  isSearching: boolean
  lastQuery: string
  filterKind: EvidenceKind | 'all'
  lookupCitations: (q: string) => Promise<CitationEvidence[]>
  setFilterKind: (kind: EvidenceKind | 'all') => void
  clearResults: () => void
}

export const useCitationStore = create<CitationState>((set) => ({
  results: [],
  isSearching: false,
  lastQuery: '',
  filterKind: 'all',

  lookupCitations: async (q) => {
    if (!q.trim()) {
      set({ results: [], lastQuery: '' })
      return []
    }
    set({ isSearching: true, lastQuery: q })
    try {
      const res = await fetch(`${API_BASE_URL}/citations/lookup?q=${encodeURIComponent(q)}`)
      if (!res.ok) { set({ isSearching: false }); return [] }
      const data = await res.json()
      const results: CitationEvidence[] = data.citations || data.results || []
      set({ results, isSearching: false })
      return results
    } catch (error) {
      console.error("Failed to lookup citations:", error)
      set({ isSearching: false })
      return []
    }
  },

  setFilterKind: (kind) => set({ filterKind: kind }),

  clearResults: () => set({ results: [], lastQuery: '' }),
}))
