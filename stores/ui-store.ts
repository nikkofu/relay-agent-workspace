import { create } from "zustand"

interface UIState {
  isSidebarOpen: boolean
  isThreadOpen: boolean
  activeThreadId: string | null
  toggleSidebar: () => void
  openThread: (threadId: string) => void
  closeThread: () => void
}

export const useUIStore = create<UIState>((set) => ({
  isSidebarOpen: true,
  isThreadOpen: false,
  activeThreadId: null,
  toggleSidebar: () => set((state) => ({ isSidebarOpen: !state.isSidebarOpen })),
  openThread: (threadId) => set({ isThreadOpen: true, activeThreadId: threadId }),
  closeThread: () => set({ isThreadOpen: false, activeThreadId: null }),
}))
