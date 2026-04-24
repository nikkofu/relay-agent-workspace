"use client"

// ── Phase 66 T07: Channel Execution Hub — Lists Panel (wired) ────────────────
//
// T02 shell is now wired to Gemini's v0.6.39 backend contract (frozen Q1):
//   - flat source_message_id / source_channel_id / source_snippet on each item
//     → rendered as a violet deep-link chip that scrolls to the originating
//       message. Items without a source chip still render the normal row.
//   - "New List" CTA now enabled with an inline name input + channel-scoped
//     createList action. Channel members can now turn a conversation into a
//     new list without leaving the channel.

import { useEffect, useRef, useState } from "react"
import { useRouter } from "next/navigation"
import {
  Loader2, ListTodo, CheckCircle2, Circle, RefreshCw, Plus, MessageSquare, X,
} from "lucide-react"
import { useListStore } from "@/stores/list-store"
import { useUserStore } from "@/stores/user-store"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { formatDistanceToNow } from "date-fns"

interface ChannelListsPanelProps {
  channelId: string
}

export function ChannelListsPanel({ channelId }: ChannelListsPanelProps) {
  const router = useRouter()
  const { lists, isLoading, fetchLists, createList } = useListStore()
  const { currentUser } = useUserStore()
  const didFetch = useRef<string>("")

  // Phase 66 T07: inline "New List" create form state
  const [isCreating, setIsCreating] = useState(false)
  const [newListTitle, setNewListTitle] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (didFetch.current !== channelId) {
      didFetch.current = channelId
      fetchLists(channelId)
    }
  }, [channelId, fetchLists])

  const channelLists = lists.filter(l => l.channelId === channelId)

  const handleRefresh = () => {
    didFetch.current = ""
    fetchLists(channelId)
  }

  const handleOpenSource = (e: React.MouseEvent, sourceChannelId: string, sourceMessageId: string) => {
    e.stopPropagation()
    if (!sourceChannelId || !sourceMessageId) return
    router.push(`/workspace?c=${sourceChannelId}#msg-${sourceMessageId}`)
  }

  const handleCreateList = async () => {
    const title = newListTitle.trim()
    if (!title || !currentUser) return
    setIsSubmitting(true)
    await createList({ title, channelId, userId: currentUser.id })
    setIsSubmitting(false)
    setNewListTitle("")
    setIsCreating(false)
  }

  const inlineCreator = (
    <div className="px-3 py-2 border-b bg-violet-500/5">
      <div className="flex items-center gap-1.5">
        <input
          autoFocus
          value={newListTitle}
          onChange={e => setNewListTitle(e.target.value)}
          onKeyDown={e => {
            if (e.key === "Enter") handleCreateList()
            if (e.key === "Escape") { setIsCreating(false); setNewListTitle("") }
          }}
          placeholder="List name…"
          className="flex-1 px-2 py-1 text-xs font-medium rounded border border-violet-400/40 bg-white dark:bg-[#19171d] focus:outline-none focus:ring-1 focus:ring-violet-500"
          disabled={isSubmitting}
        />
        <button
          type="button"
          onClick={handleCreateList}
          disabled={!newListTitle.trim() || isSubmitting}
          className="h-7 px-2 text-[10px] font-bold uppercase tracking-wider rounded bg-violet-600 text-white hover:bg-violet-700 disabled:opacity-40 disabled:cursor-not-allowed"
        >
          {isSubmitting ? "…" : "Create"}
        </button>
        <button
          type="button"
          onClick={() => { setIsCreating(false); setNewListTitle("") }}
          className="h-7 w-7 flex items-center justify-center rounded hover:bg-muted"
          title="Cancel"
        >
          <X className="w-3 h-3 text-muted-foreground" />
        </button>
      </div>
    </div>
  )

  if (isLoading && channelLists.length === 0) {
    return (
      <div className="flex flex-col h-full">
        {isCreating && inlineCreator}
        <div className="flex flex-col items-center justify-center py-12 text-muted-foreground gap-2">
          <Loader2 className="w-5 h-5 animate-spin" />
          <span className="text-xs">Loading lists…</span>
        </div>
      </div>
    )
  }

  if (channelLists.length === 0) {
    return (
      <div className="flex flex-col h-full">
        {isCreating && inlineCreator}
        <div className="flex flex-col items-center justify-center py-12 px-4 text-center gap-2">
          <div className="w-12 h-12 rounded-full bg-violet-500/10 flex items-center justify-center">
            <ListTodo className="w-6 h-6 text-violet-600" />
          </div>
          <p className="text-sm font-semibold text-foreground">No lists in this channel</p>
          <p className="text-[11px] text-muted-foreground max-w-[220px] leading-snug">
            Lists turn channel conversations into tracked work. Create one from a message or start fresh.
          </p>
          {!isCreating && (
            <Button
              variant="outline"
              size="sm"
              className="mt-2 gap-1.5 text-xs font-bold h-7"
              onClick={() => setIsCreating(true)}
            >
              <Plus className="w-3 h-3" />
              New List
            </Button>
          )}
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header row */}
      <div className="flex items-center justify-between px-3 py-2 border-b bg-muted/20">
        <span className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
          {channelLists.length} {channelLists.length === 1 ? "List" : "Lists"}
        </span>
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={() => setIsCreating(true)}
            className="p-1 rounded hover:bg-violet-500/15 text-violet-600 transition-colors"
            title="New List"
          >
            <Plus className="w-3.5 h-3.5" />
          </button>
          <button
            type="button"
            onClick={handleRefresh}
            className="p-1 rounded hover:bg-muted/60 transition-colors"
            title="Refresh"
          >
            <RefreshCw className={cn("w-3 h-3 text-muted-foreground", isLoading && "animate-spin")} />
          </button>
        </div>
      </div>

      {/* Inline creator */}
      {isCreating && inlineCreator}

      {/* List rows */}
      <div className="flex-1 overflow-y-auto">
        {channelLists.map(list => {
          const completion = list.itemCount > 0
            ? Math.round((list.completedCount / list.itemCount) * 100)
            : 0
          const updatedAgo = (() => {
            try { return formatDistanceToNow(new Date(list.updatedAt), { addSuffix: true }) }
            catch { return "" }
          })()

          return (
            <div
              key={list.id}
              className="px-3 py-2.5 border-b hover:bg-muted/30 transition-colors cursor-pointer group"
            >
              <div className="flex items-start gap-2">
                <div className="w-7 h-7 rounded-lg bg-violet-500/10 flex items-center justify-center shrink-0 mt-0.5">
                  <ListTodo className="w-3.5 h-3.5 text-violet-600" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <p className="text-[12px] font-semibold text-foreground truncate leading-tight">
                      {list.title}
                    </p>
                    {list.itemCount > 0 && (
                      <span className="text-[9px] font-black uppercase tracking-widest text-muted-foreground shrink-0">
                        {list.completedCount}/{list.itemCount}
                      </span>
                    )}
                  </div>
                  {list.itemCount > 0 && (
                    <div className="h-1 bg-muted rounded-full mt-1.5 overflow-hidden">
                      <div
                        className={cn(
                          "h-full rounded-full transition-all",
                          completion === 100 ? "bg-emerald-500" : "bg-violet-500"
                        )}
                        style={{ width: `${completion}%` }}
                      />
                    </div>
                  )}
                  {updatedAgo && (
                    <p className="text-[9px] text-muted-foreground/60 uppercase tracking-widest mt-1">
                      Updated {updatedAgo}
                    </p>
                  )}
                  {/* Preview first 3 items if present */}
                  {list.items && list.items.length > 0 && (
                    <div className="mt-1.5 space-y-1">
                      {list.items.slice(0, 3).map(item => (
                        <div key={item.id} className="text-[11px]">
                          <div className="flex items-center gap-1.5">
                            {item.isCompleted
                              ? <CheckCircle2 className="w-3 h-3 text-emerald-500 shrink-0" />
                              : <Circle className="w-3 h-3 text-muted-foreground shrink-0" />}
                            <span className={cn(
                              "truncate flex-1",
                              item.isCompleted ? "text-muted-foreground line-through" : "text-foreground/80"
                            )}>
                              {item.content}
                            </span>
                            {/* Phase 66 T07: source-message chip (flat contract Q1) */}
                            {item.sourceMessageId && item.sourceChannelId && (
                              <button
                                type="button"
                                onClick={e => handleOpenSource(e, item.sourceChannelId!, item.sourceMessageId!)}
                                className="inline-flex items-center gap-0.5 text-[8px] font-black uppercase tracking-widest px-1 py-0.5 rounded bg-violet-500/10 text-violet-700 dark:text-violet-300 hover:bg-violet-500/20 shrink-0"
                                title={item.sourceSnippet || "Open source message"}
                              >
                                <MessageSquare className="w-2 h-2" />
                                From msg
                              </button>
                            )}
                          </div>
                          {item.sourceSnippet && item.sourceMessageId && (
                            <p className="text-[9px] text-muted-foreground/70 italic pl-4 truncate">
                              “{item.sourceSnippet}”
                            </p>
                          )}
                        </div>
                      ))}
                      {list.items.length > 3 && (
                        <p className="text-[9px] text-muted-foreground/60 pl-4">
                          +{list.items.length - 3} more
                        </p>
                      )}
                    </div>
                  )}
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
