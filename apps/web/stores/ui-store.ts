import { create } from "zustand"

interface UIState {
  isSidebarOpen: boolean
  isThreadOpen: boolean
  activeThreadId: string | null
  isAIPanelOpen: boolean
  isSearchOpen: boolean
  isCanvasOpen: boolean
  activeCanvasId: string | null
  dockedChats: string[] // List of user IDs with open chats
  toggleSidebar: () => void
  openThread: (threadId: string) => void
  closeThread: () => void
  toggleAIPanel: () => void
  openAIPanel: () => void
  closeAIPanel: () => void
  toggleSearch: () => void
  openSearch: () => void
  closeSearch: () => void
  openCanvas: (canvasId: string) => void
  closeCanvas: () => void
  openDockedChat: (userId: string) => void
  closeDockedChat: (userId: string) => void
}

export const useUIStore = create<UIState>((set) => ({
  isSidebarOpen: true,
  isThreadOpen: false,
  activeThreadId: null,
  isAIPanelOpen: false,
  isSearchOpen: false,
  isCanvasOpen: false,
  activeCanvasId: null,
  dockedChats: [],
  toggleSidebar: () => set((state) => ({ isSidebarOpen: !state.isSidebarOpen })),
  openThread: (threadId) => set({ isThreadOpen: true, activeThreadId: threadId, isAIPanelOpen: false, isCanvasOpen: false }),
  closeThread: () => set({ isThreadOpen: false, activeThreadId: null }),
  toggleAIPanel: () => set((state) => ({ isAIPanelOpen: !state.isAIPanelOpen, isThreadOpen: false, isCanvasOpen: false })),
  openAIPanel: () => set({ isAIPanelOpen: true, isThreadOpen: false, isCanvasOpen: false }),
  closeAIPanel: () => set({ isAIPanelOpen: false }),
  toggleSearch: () => set((state) => ({ isSearchOpen: !state.isSearchOpen })),
  openSearch: () => set({ isSearchOpen: true }),
  closeSearch: () => set({ isSearchOpen: false }),
  openCanvas: (canvasId) => set({ isCanvasOpen: true, activeCanvasId: canvasId, isThreadOpen: false, isAIPanelOpen: false }),
  closeCanvas: () => set({ isCanvasOpen: false, activeCanvasId: null }),
  openDockedChat: (userId) => set((state) => ({ 
    dockedChats: state.dockedChats.includes(userId) 
      ? [userId, ...state.dockedChats.filter(id => id !== userId)] // Bring to front
      : [userId, ...state.dockedChats].slice(0, 3) // Max 3 windows
  })),
  closeDockedChat: (userId) => set((state) => ({ 
    dockedChats: state.dockedChats.filter(id => id !== userId) 
  })),
}))
