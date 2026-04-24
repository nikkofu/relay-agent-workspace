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
  // Phase 66 T02+T07 source-message link (flat shape per frozen contract)
  sourceMessageId?: string
  sourceChannelId?: string
  sourceSnippet?: string
  // Phase 66 T03 assignee + due (flat)
  assignedTo?: string
  dueAt?: string
}

// Phase 66 T08: AI list-item draft suggestion (matches /api/v1/ai/lists/draft)
export interface ListItemDraftSuggestion {
  title: string
  assignee_user_id?: string
  due_at?: string | null
  rationale?: string
  source_message_id?: string
  source_channel_id?: string
  source_snippet?: string
}

export interface ListItemDraftResult {
  ok: boolean
  fallback?: string
  suggestion: ListItemDraftSuggestion
}

export interface AddItemWithSourceInput {
  content: string
  assignedTo?: string
  dueAt?: string | null
  sourceMessageId?: string
  sourceChannelId?: string
  sourceSnippet?: string
}

interface ListState {
  lists: WorkspaceList[]
  activeList: WorkspaceList | null
  isLoading: boolean
  
  fetchLists: (channelId: string) => Promise<void>
  fetchListDetail: (id: string) => Promise<void>
  createList: (data: { title: string, channelId: string, userId: string }) => Promise<WorkspaceList | null>
  updateList: (id: string, updates: Partial<WorkspaceList>) => Promise<void>
  deleteList: (id: string) => Promise<void>
  
  addItem: (listId: string, content: string) => Promise<void>
  addItemWithSource: (listId: string, input: AddItemWithSourceInput) => Promise<WorkspaceListItem | null>
  addItemLocally: (item: WorkspaceListItem) => void
  updateItemLocally: (listId: string, item: WorkspaceListItem) => void
  removeItemLocally: (listId: string, itemId: string) => void
  aiDraftListItem: (input: { messageId: string, listId: string, channelId?: string, context?: string }) => Promise<ListItemDraftResult | null>
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
  items: (l.items || []).map(mapListItem)
})

// Exported so realtime WS handlers (`use-websocket.ts`) can normalise raw
// snake_case payloads into the `WorkspaceListItem` shape the store
// internally relies on. Without this, WS-delivered items end up missing
// `listId` / `isCompleted` / `createdAt` etc. and slip out of every
// downstream computation that depends on them.
export const mapListItem = (i: any): WorkspaceListItem => ({
  ...i,
  listId: i.list_id,
  userId: i.user_id,
  isCompleted: i.is_completed,
  createdAt: i.created_at,
  updatedAt: i.updated_at,
  // Phase 66 T07: flat source-message fields
  sourceMessageId: i.source_message_id || undefined,
  sourceChannelId: i.source_channel_id || undefined,
  sourceSnippet: i.source_snippet || undefined,
  assignedTo: i.assigned_to || undefined,
  dueAt: i.due_at || undefined,
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
      const payload = await response.json().catch(() => ({}))
      const created = payload?.list ? mapList(payload.list) : null
      await get().fetchLists(data.channelId)
      toast.success("List created")
      return created
    } catch (error) {
      console.error("Failed to create list:", error)
      toast.error("Failed to create list")
      return null
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

  addItemWithSource: async (listId, input) => {
    try {
      const body: Record<string, any> = { content: input.content }
      if (input.assignedTo) body.assigned_to = input.assignedTo
      if (input.dueAt) body.due_at = input.dueAt
      if (input.sourceMessageId) body.source_message_id = input.sourceMessageId
      if (input.sourceChannelId) body.source_channel_id = input.sourceChannelId
      if (input.sourceSnippet) body.source_snippet = input.sourceSnippet

      const response = await fetch(`${API_BASE_URL}/lists/${listId}/items`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      })
      if (!response.ok) throw new Error("Add item failed")
      const data = await response.json().catch(() => ({}))
      const created = data?.item ? mapListItem(data.item) : null
      
      // Phase 67B: addItemLocally will be called by use-websocket if enabled,
      // but manual call here ensures immediate feedback if WS is slow.
      if (created) get().addItemLocally(created)
      
      toast.success(input.sourceMessageId ? "Added to list from message" : "Item added")
      return created
    } catch (error) {
      console.error("Failed to add list item with source:", error)
      toast.error("Failed to add to list")
      return null
    }
  },

  addItemLocally: (item) => {
    set((state) => {
      const existingList = state.lists.find(l => l.id === item.listId)
      if (!existingList) return state

      // Avoid duplicates
      if (existingList.items?.find(i => i.id === item.id)) return state

      const updatedItems = [...(existingList.items || []), item].sort((a, b) => a.position - b.position)
      const updatedList = { ...existingList, items: updatedItems, itemCount: (existingList.itemCount || 0) + 1 }
      
      return {
        lists: state.lists.map(l => l.id === item.listId ? updatedList : l),
        activeList: state.activeList?.id === item.listId ? updatedList : state.activeList
      }
    })
  },

  updateItemLocally: (listId, item) => {
    set((state) => {
      const existingList = state.lists.find(l => l.id === listId)
      if (!existingList) return state

      const updatedItems = (existingList.items || []).map(i => i.id === item.id ? item : i)
      const updatedList = { 
        ...existingList, 
        items: updatedItems,
        completedCount: updatedItems.filter(i => i.isCompleted).length
      }

      return {
        lists: state.lists.map(l => l.id === listId ? updatedList : l),
        activeList: state.activeList?.id === listId ? updatedList : state.activeList
      }
    })
  },

  removeItemLocally: (listId, itemId) => {
    set((state) => {
      const existingList = state.lists.find(l => l.id === listId)
      if (!existingList) return state

      const updatedItems = (existingList.items || []).filter(i => i.id !== itemId)
      const updatedList = { 
        ...existingList, 
        items: updatedItems,
        itemCount: Math.max(0, (existingList.itemCount || 0) - 1),
        completedCount: updatedItems.filter(i => i.isCompleted).length
      }

      return {
        lists: state.lists.map(l => l.id === listId ? updatedList : l),
        activeList: state.activeList?.id === listId ? updatedList : state.activeList
      }
    })
  },

  // Phase 66 T08: call AI draft endpoint. Per frozen Codex contract (Q2) the
  // backend may return { ok: false, fallback: "manual_entry", suggestion: {...} }
  // even on soft-failure — the suggestion is still usable as form defaults.
  aiDraftListItem: async ({ messageId, listId, channelId, context }) => {
    try {
      const response = await fetch(`${API_BASE_URL}/ai/lists/draft`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          message_id: messageId,
          list_id: listId,
          channel_id: channelId,
          context,
        }),
      })
      if (!response.ok) throw new Error("AI draft failed")
      const data = await response.json()
      return {
        ok: Boolean(data?.ok),
        fallback: data?.fallback,
        suggestion: data?.suggestion || { title: "" },
      }
    } catch (error) {
      console.error("Failed to draft list item:", error)
      return null
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
