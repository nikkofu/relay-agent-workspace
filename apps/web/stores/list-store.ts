import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"
import { User } from "@/types"

export interface WorkspaceList {
  id: string
  title: string
  channelId: string
  userId: string
  itemCount: number
  completedCount: number
  createdAt: string
  updatedAt: string
  assignedUser?: User
  items?: WorkspaceListItem[]
}

export interface WorkspaceListItem {
  id: string
  listId: string
  content: string
  isCompleted: boolean
  userId: string
  position: number
  createdAt: string
  updatedAt: string
}

interface ListState {
  lists: WorkspaceList[]
  activeList: WorkspaceList | null
  isLoading: boolean
  
  fetchLists: (channelId: string) => Promise<void>
  fetchListDetail: (id: string) => Promise<void>
  createList: (data: { title: string, channelId: string, userId: string }) => Promise<void>
  updateList: (id: string, updates: Partial<WorkspaceList>) => Promise<void>
  deleteList: (id: string) => Promise<void>
  
  addItem: (listId: string, content: string) => Promise<void>
  toggleItem: (listId: string, itemId: string, isCompleted: boolean) => Promise<void>
  deleteItem: (listId: string, itemId: string) => Promise<void>
}

const mapList = (l: any): WorkspaceList => ({
  ...l,
  channelId: l.channel_id,
  userId: l.user_id,
  itemCount: l.item_count,
  completedCount: l.completed_count,
  createdAt: l.created_at,
  updatedAt: l.updated_at,
  assignedUser: l.assigned_user,
  items: (l.items || []).map(mapItem)
})

const mapItem = (i: any): WorkspaceListItem => ({
  ...i,
  listId: i.list_id,
  userId: i.user_id,
  isCompleted: i.is_completed,
  createdAt: i.created_at,
  updatedAt: i.updated_at
})

export const useListStore = create<ListState>((set, get) => ({
  lists: [],
  activeList: null,
  isLoading: false,

  fetchLists: async (channelId) => {
    try {
      set({ isLoading: true })
      const response = await fetch(`${API_BASE_URL}/lists?channel_id=${channelId}`)
      const data = await response.json()
      set({ lists: (data.lists || []).map(mapList), isLoading: false })
    } catch (error) {
      console.error("Failed to fetch lists:", error)
      set({ isLoading: false })
    }
  },

  fetchListDetail: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/lists/${id}`)
      const data = await response.json()
      set({ activeList: mapList(data.list) })
    } catch (error) {
      console.error("Failed to fetch list detail:", error)
    }
  },

  createList: async (data) => {
    try {
      const response = await fetch(`${API_BASE_URL}/lists`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          title: data.title,
          channel_id: data.channelId,
          user_id: data.userId
        })
      })
      if (!response.ok) throw new Error("Create failed")
      await get().fetchLists(data.channelId)
      toast.success("List created")
    } catch (error) {
      console.error("Failed to create list:", error)
      toast.error("Failed to create list")
    }
  },

  updateList: async (id, updates) => {
    try {
      const response = await fetch(`${API_BASE_URL}/lists/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(updates)
      })
      if (!response.ok) throw new Error("Update failed")
      const data = await response.json()
      const updated = mapList(data.list)
      set((state) => ({
        lists: state.lists.map(l => l.id === id ? updated : l),
        activeList: state.activeList?.id === id ? updated : state.activeList
      }))
    } catch (error) {
      console.error("Failed to update list:", error)
    }
  },

  deleteList: async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/lists/${id}`, {
        method: "DELETE"
      })
      if (!response.ok) throw new Error("Delete failed")
      set((state) => ({
        lists: state.lists.filter(l => l.id !== id),
        activeList: state.activeList?.id === id ? null : state.activeList
      }))
      toast.success("List deleted")
    } catch (error) {
      console.error("Failed to delete list:", error)
    }
  },

  addItem: async (listId, content) => {
    try {
      const response = await fetch(`${API_BASE_URL}/lists/${listId}/items`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content })
      })
      if (!response.ok) throw new Error("Add item failed")
      await get().fetchListDetail(listId)
    } catch (error) {
      console.error("Failed to add list item:", error)
    }
  },

  toggleItem: async (listId, itemId, isCompleted) => {
    try {
      const response = await fetch(`${API_BASE_URL}/lists/${listId}/items/${itemId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ is_completed: isCompleted })
      })
      if (!response.ok) throw new Error("Toggle item failed")
      await get().fetchListDetail(listId)
    } catch (error) {
      console.error("Failed to toggle list item:", error)
    }
  },

  deleteItem: async (listId, itemId) => {
    try {
      const response = await fetch(`${API_BASE_URL}/lists/${listId}/items/${itemId}`, {
        method: "DELETE"
      })
      if (!response.ok) throw new Error("Delete item failed")
      await get().fetchListDetail(listId)
    } catch (error) {
      console.error("Failed to delete list item:", error)
    }
  }
}))
