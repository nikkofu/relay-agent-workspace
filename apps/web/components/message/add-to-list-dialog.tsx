"use client"

// ── Phase 66 T08: Message → Add to List dialog ───────────────────────────────
//
// Implements the frozen Codex Q2 contract:
//   POST /api/v1/ai/lists/draft
//   → { ok: boolean, fallback?: "manual_entry", suggestion: { title, assignee_user_id, due_at, rationale, source_message_id, source_channel_id, source_snippet } }
//
// The suggestion block is ALWAYS usable as default form values, even on soft
// failure (`ok=false`). On soft failure we render a subtle "manual entry"
// banner so the user knows the AI pass fell back, but the form still works.
//
// Persistence goes through `useListStore.addItemWithSource` which forwards the
// flat source-message fields (Q1 contract) to the backend so the saved item
// durably references the originating channel message.

import { useEffect, useState } from "react"
import { ListTodo, Sparkles, Loader2, Plus, AlertTriangle } from "lucide-react"
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { useListStore, ListItemDraftSuggestion } from "@/stores/list-store"
import { useUserStore } from "@/stores/user-store"
import type { Message } from "@/types"
import { cn } from "@/lib/utils"

interface AddToListDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  message: Message
}

export function AddToListDialog({ open, onOpenChange, message }: AddToListDialogProps) {
  const { lists, fetchLists, addItemWithSource, aiDraftListItem } = useListStore()
  const { users } = useUserStore()

  const [selectedListId, setSelectedListId] = useState<string>("")
  const [title, setTitle] = useState<string>("")
  const [assigneeId, setAssigneeId] = useState<string>("")
  const [dueAt, setDueAt] = useState<string>("")
  const [rationale, setRationale] = useState<string>("")
  const [isDrafting, setIsDrafting] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [softFallback, setSoftFallback] = useState(false)
  const [suggestion, setSuggestion] = useState<ListItemDraftSuggestion | null>(null)

  // Seed lists for this channel on open
  useEffect(() => {
    if (open && message.channelId) fetchLists(message.channelId)
  }, [open, message.channelId, fetchLists])

  // Default the title to the message content on open (pre-AI baseline)
  useEffect(() => {
    if (open) {
      setTitle(message.content || "")
      setSoftFallback(false)
      setSuggestion(null)
      setAssigneeId("")
      setDueAt("")
      setRationale("")
    }
  }, [open, message.content])

  const channelLists = lists.filter(l => l.channelId === message.channelId)

  // Auto-pick the first list once loaded so the Draft/Submit buttons are enabled
  useEffect(() => {
    if (open && !selectedListId && channelLists.length > 0) {
      setSelectedListId(channelLists[0].id)
    }
  }, [open, selectedListId, channelLists])

  const handleAIDraft = async () => {
    if (!selectedListId) return
    setIsDrafting(true)
    const result = await aiDraftListItem({
      messageId: message.id,
      listId: selectedListId,
      channelId: message.channelId,
      context: message.content,
    })
    setIsDrafting(false)
    if (!result) return

    setSuggestion(result.suggestion)
    setSoftFallback(!result.ok)
    // Always apply suggestion as defaults — contract guarantees usable shape
    if (result.suggestion.title) setTitle(result.suggestion.title)
    if (result.suggestion.assignee_user_id) setAssigneeId(result.suggestion.assignee_user_id)
    if (result.suggestion.due_at) {
      // Convert ISO → local "YYYY-MM-DDTHH:mm" for <input type=datetime-local>
      try {
        const d = new Date(result.suggestion.due_at)
        if (!isNaN(d.getTime())) {
          const pad = (n: number) => String(n).padStart(2, "0")
          setDueAt(`${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`)
        }
      } catch { /* noop */ }
    }
    if (result.suggestion.rationale) setRationale(result.suggestion.rationale)
  }

  const handleSubmit = async () => {
    if (!selectedListId || !title.trim()) return
    setIsSubmitting(true)
    const saved = await addItemWithSource(selectedListId, {
      content: title.trim(),
      assignedTo: assigneeId || undefined,
      dueAt: dueAt ? new Date(dueAt).toISOString() : undefined,
      sourceMessageId: message.id,
      sourceChannelId: message.channelId,
      sourceSnippet: message.content,
    })
    setIsSubmitting(false)
    if (saved) onOpenChange(false)
  }

  const canSubmit = !!selectedListId && !!title.trim()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-sm font-black uppercase tracking-tight">
            <ListTodo className="w-4 h-4 text-violet-600" />
            Add to List
          </DialogTitle>
          <DialogDescription className="text-xs">
            Turn this message into a tracked list item. The saved item will link back to the source message.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3 py-2">
          {/* Source snippet */}
          <div className="rounded-md border border-violet-400/30 bg-violet-500/5 px-2.5 py-1.5">
            <p className="text-[9px] font-black uppercase tracking-widest text-violet-700 dark:text-violet-300 mb-0.5">
              Source message
            </p>
            <p className="text-[11px] text-foreground/80 italic line-clamp-2">
              “{message.content}”
            </p>
          </div>

          {/* Target list */}
          <div className="space-y-1">
            <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Target list</Label>
            <select
              value={selectedListId}
              onChange={e => setSelectedListId(e.target.value)}
              className="w-full h-8 px-2 text-xs font-medium rounded border bg-white dark:bg-[#19171d] focus:outline-none focus:ring-1 focus:ring-violet-500"
              disabled={isSubmitting}
            >
              <option value="">Select a list…</option>
              {channelLists.map(l => (
                <option key={l.id} value={l.id}>{l.title}</option>
              ))}
            </select>
            {channelLists.length === 0 && (
              <p className="text-[10px] text-muted-foreground">
                No lists in this channel — create one from the Execution panel first.
              </p>
            )}
          </div>

          {/* AI draft CTA */}
          <div className="flex items-center justify-between gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleAIDraft}
              disabled={!selectedListId || isDrafting || isSubmitting}
              className="h-7 gap-1.5 text-[11px] font-bold border-violet-500/30 text-violet-700 dark:text-violet-300 hover:bg-violet-500/10"
            >
              {isDrafting ? <Loader2 className="w-3 h-3 animate-spin" /> : <Sparkles className="w-3 h-3" />}
              AI Draft
            </Button>
            {suggestion && (
              <span className={cn(
                "text-[9px] font-black uppercase tracking-widest px-1.5 py-0.5 rounded",
                softFallback
                  ? "bg-amber-500/10 text-amber-700 dark:text-amber-300"
                  : "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
              )}>
                {softFallback ? "Manual entry" : "AI suggestion"}
              </span>
            )}
          </div>
          {softFallback && (
            <div className="flex items-start gap-1.5 rounded-md border border-amber-400/30 bg-amber-500/5 px-2.5 py-1.5">
              <AlertTriangle className="w-3 h-3 text-amber-600 mt-0.5 shrink-0" />
              <p className="text-[10px] text-amber-700 dark:text-amber-300 leading-snug">
                AI draft unavailable — we pre-filled fields from the message so you can edit and save manually.
              </p>
            </div>
          )}

          {/* Title */}
          <div className="space-y-1">
            <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Title</Label>
            <input
              value={title}
              onChange={e => setTitle(e.target.value)}
              className="w-full h-8 px-2 text-xs font-medium rounded border bg-white dark:bg-[#19171d] focus:outline-none focus:ring-1 focus:ring-violet-500"
              disabled={isSubmitting}
            />
          </div>

          {/* Assignee */}
          <div className="space-y-1">
            <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
              Assignee <span className="text-muted-foreground/60 normal-case tracking-normal">(optional)</span>
            </Label>
            <select
              value={assigneeId}
              onChange={e => setAssigneeId(e.target.value)}
              className="w-full h-8 px-2 text-xs font-medium rounded border bg-white dark:bg-[#19171d] focus:outline-none focus:ring-1 focus:ring-violet-500"
              disabled={isSubmitting}
            >
              <option value="">Unassigned</option>
              {users.map(u => (
                <option key={u.id} value={u.id}>{u.name}</option>
              ))}
            </select>
          </div>

          {/* Due date */}
          <div className="space-y-1">
            <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
              Due date <span className="text-muted-foreground/60 normal-case tracking-normal">(optional)</span>
            </Label>
            <input
              type="datetime-local"
              value={dueAt}
              onChange={e => setDueAt(e.target.value)}
              className="w-full h-8 px-2 text-xs font-medium rounded border bg-white dark:bg-[#19171d] focus:outline-none focus:ring-1 focus:ring-violet-500"
              disabled={isSubmitting}
            />
          </div>

          {rationale && (
            <div className="rounded-md border border-violet-400/30 bg-violet-500/5 px-2.5 py-1.5">
              <p className="text-[9px] font-black uppercase tracking-widest text-violet-700 dark:text-violet-300 mb-0.5">
                AI rationale
              </p>
              <p className="text-[11px] text-foreground/80 leading-snug">{rationale}</p>
            </div>
          )}
        </div>

        <DialogFooter className="gap-2">
          <Button variant="ghost" size="sm" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button
            size="sm"
            onClick={handleSubmit}
            disabled={!canSubmit || isSubmitting}
            className="bg-violet-600 hover:bg-violet-700 text-white gap-1.5"
          >
            {isSubmitting ? <Loader2 className="w-3 h-3 animate-spin" /> : <Plus className="w-3 h-3" />}
            Add to list
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
