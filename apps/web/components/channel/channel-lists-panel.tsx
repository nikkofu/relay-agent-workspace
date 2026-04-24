"use client"

// ── Phase 66 T02: Channel Execution Hub — Lists Panel Shell ──────────────────
//
// Reads only existing list-store fields (WorkspaceList / WorkspaceListItem).
// Message-linked source fields (source_message_id, source_channel_id, snippet)
// are Gemini T01/T02 backend work — NOT consumed here yet. This panel will be
// extended in T07 once the backend contract is frozen.

import { useEffect, useRef } from "react"
import { Loader2, ListTodo, CheckCircle2, Circle, RefreshCw, Plus } from "lucide-react"
import { useListStore } from "@/stores/list-store"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { formatDistanceToNow } from "date-fns"

interface ChannelListsPanelProps {
  channelId: string
}

export function ChannelListsPanel({ channelId }: ChannelListsPanelProps) {
  const { lists, isLoading, fetchLists } = useListStore()
  const didFetch = useRef<string>("")

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

  if (isLoading && channelLists.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground gap-2">
        <Loader2 className="w-5 h-5 animate-spin" />
        <span className="text-xs">Loading lists…</span>
      </div>
    )
  }

  if (channelLists.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 px-4 text-center gap-2">
        <div className="w-12 h-12 rounded-full bg-violet-500/10 flex items-center justify-center">
          <ListTodo className="w-6 h-6 text-violet-600" />
        </div>
        <p className="text-sm font-semibold text-foreground">No lists in this channel</p>
        <p className="text-[11px] text-muted-foreground max-w-[220px] leading-snug">
          Lists turn channel conversations into tracked work. Create one from a message or start fresh.
        </p>
        <Button
          variant="outline"
          size="sm"
          className="mt-2 gap-1.5 text-xs font-bold h-7"
          disabled
          title="Available after Phase 66 T07"
        >
          <Plus className="w-3 h-3" />
          New List
        </Button>
        <span className="text-[9px] text-muted-foreground/60 uppercase tracking-widest mt-1">
          Create available after backend contract lands
        </span>
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
            onClick={handleRefresh}
            className="p-1 rounded hover:bg-muted/60 transition-colors"
            title="Refresh"
          >
            <RefreshCw className={cn("w-3 h-3 text-muted-foreground", isLoading && "animate-spin")} />
          </button>
        </div>
      </div>

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
                    <div className="mt-1.5 space-y-0.5">
                      {list.items.slice(0, 3).map(item => (
                        <div key={item.id} className="flex items-center gap-1.5 text-[11px]">
                          {item.isCompleted
                            ? <CheckCircle2 className="w-3 h-3 text-emerald-500 shrink-0" />
                            : <Circle className="w-3 h-3 text-muted-foreground shrink-0" />}
                          <span className={cn(
                            "truncate",
                            item.isCompleted ? "text-muted-foreground line-through" : "text-foreground/80"
                          )}>
                            {item.content}
                          </span>
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
