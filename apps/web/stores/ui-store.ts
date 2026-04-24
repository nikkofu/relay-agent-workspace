import { create } from "zustand"

interface UIState {
  isSidebarOpen: boolean
  isThreadOpen: boolean
  activeThreadId: string | null
  isAIPanelOpen: boolean
  isSearchOpen: boolean
  isCanvasOpen: boolean
  activeCanvasId: string | null
  // True while the user is actively editing the open canvas. The workspace
  // layout watches this to resize to a 33/33/33 (messages / editor / AI chat)
  // split so all three columns get roughly equal, full-height visibility.
  isCanvasEditing: boolean
  // When the user toggles Maximize on the canvas, the workspace layout
  // collapses the left messages column so the canvas + dock effectively
  // fill the viewport.
  isCanvasMaximized: boolean
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
  setCanvasEditing: (editing: boolean) => void
  toggleCanvasMaximized: () => void
  setCanvasMaximized: (maximized: boolean) => void
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
  isCanvasEditing: false,
  isCanvasMaximized: false,
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
  openCanvas: (canvasId) => set({
    isCanvasOpen: true, activeCanvasId: canvasId,
    isThreadOpen: false, isAIPanelOpen: false,
    isCanvasEditing: false, isCanvasMaximized: false,
  }),
  closeCanvas: () => set({
    isCanvasOpen: false, activeCanvasId: null,
    isCanvasEditing: false, isCanvasMaximized: false,
  }),
  setCanvasEditing: (editing) => set({ isCanvasEditing: editing }),
  toggleCanvasMaximized: () => set((state) => ({ isCanvasMaximized: !state.isCanvasMaximized })),
  setCanvasMaximized: (maximized) => set({ isCanvasMaximized: maximized }),
  openDockedChat: (userId) => set((state) => ({ 
    dockedChats: state.dockedChats.includes(userId) 
      ? [userId, ...state.dockedChats.filter(id => id !== userId)] // Bring to front
      : [userId, ...state.dockedChats].slice(0, 3) // Max 3 windows
  })),
  closeDockedChat: (userId) => set((state) => ({ 
    dockedChats: state.dockedChats.filter(id => id !== userId) 
  })),
}))
