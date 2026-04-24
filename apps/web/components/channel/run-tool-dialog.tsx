"use client"

// ── Phase 66 T07: Run Tool dialog (channel-native, writeback-aware) ──────────
//
// Implements the frozen Codex Q3 contract exactly:
//   POST /api/v1/tools/:id/execute
//   body:
//     input: (free-form JSON)
//     channel_id?
//     writeback_target: "message" | "list_item"
//     writeback:
//       - message:   { channel_id, thread_id? }
//       - list_item: { list_id }
//
// The dialog never guesses around invalid target combinations — missing
// required per-target fields simply disable the Run button. On success the
// tool-run appears in the channel-tools-panel with a writeback-target badge.

import { useEffect, useState } from "react"
import { Terminal, Loader2, MessageSquare, ListTodo, X } from "lucide-react"
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { useDirectoryStore } from "@/stores/directory-store"
import { useListStore } from "@/stores/list-store"
import { useToolStore, WritebackTarget } from "@/stores/tool-store"
import { cn } from "@/lib/utils"

interface RunToolDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  channelId: string
}

export function RunToolDialog({ open, onOpenChange, channelId }: RunToolDialogProps) {
  const { tools, fetchTools } = useDirectoryStore()
  const { lists, fetchLists } = useListStore()
  const { executeTool } = useToolStore()

  const [selectedToolId, setSelectedToolId] = useState<string>("")
  const [inputText, setInputText] = useState<string>("")
  const [writebackTarget, setWritebackTarget] = useState<WritebackTarget | "">("")
  const [selectedListId, setSelectedListId] = useState<string>("")
  const [threadId, setThreadId] = useState<string>("")
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (open) {
      if (tools.length === 0) fetchTools()
      fetchLists(channelId)
    }
  }, [open, channelId, tools.length, fetchTools, fetchLists])

  const channelLists = lists.filter(l => l.channelId === channelId)

  const resetForm = () => {
    setSelectedToolId("")
    setInputText("")
    setWritebackTarget("")
    setSelectedListId("")
    setThreadId("")
    setIsSubmitting(false)
  }

  const parseInput = (): any => {
    const trimmed = inputText.trim()
    if (!trimmed) return {}
    try {
      return JSON.parse(trimmed)
    } catch {
      // Fall back to treating raw text as a simple `text` input
      return { text: trimmed }
    }
  }

  const canSubmit =
    !!selectedToolId &&
    (writebackTarget === "" ||
      (writebackTarget === "message") ||
      (writebackTarget === "list_item" && !!selectedListId))

  const handleSubmit = async () => {
    if (!canSubmit) return
    setIsSubmitting(true)

    const writeback = writebackTarget
      ? writebackTarget === "message"
        ? { target: "message" as const, payload: { channel_id: channelId, ...(threadId ? { thread_id: threadId } : {}) } }
        : { target: "list_item" as const, payload: { list_id: selectedListId } }
      : undefined

    const run = await executeTool(selectedToolId, parseInput(), channelId, writeback)
    setIsSubmitting(false)
    if (run) {
      resetForm()
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) resetForm(); onOpenChange(v) }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-sm font-black uppercase tracking-tight">
            <Terminal className="w-4 h-4 text-amber-600" />
            Run Tool
          </DialogTitle>
          <DialogDescription className="text-xs">
            Execute a workspace tool in this channel. Optionally route the result back as a
            message or as a new list item.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3 py-2">
          {/* Tool picker */}
          <div className="space-y-1">
            <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Tool</Label>
            <select
              value={selectedToolId}
              onChange={e => setSelectedToolId(e.target.value)}
              className="w-full h-8 px-2 text-xs font-medium rounded border bg-white dark:bg-[#19171d] focus:outline-none focus:ring-1 focus:ring-amber-500"
              disabled={isSubmitting}
            >
              <option value="">Select a tool…</option>
              {tools.map(t => (
                <option key={t.id} value={t.id}>{t.name}{t.category ? ` · ${t.category}` : ""}</option>
              ))}
            </select>
            {tools.length === 0 && (
              <p className="text-[10px] text-muted-foreground">No tools available in this workspace.</p>
            )}
          </div>

          {/* Input (JSON or plain text) */}
          <div className="space-y-1">
            <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
              Input <span className="text-muted-foreground/60 normal-case tracking-normal">(JSON or plain text)</span>
            </Label>
            <textarea
              value={inputText}
              onChange={e => setInputText(e.target.value)}
              placeholder='{"topic":"Phase 66"} or just a line of text'
              rows={3}
              className="w-full px-2 py-1.5 text-xs font-mono rounded border bg-white dark:bg-[#19171d] focus:outline-none focus:ring-1 focus:ring-amber-500 resize-none"
              disabled={isSubmitting}
            />
          </div>

          {/* Writeback target */}
          <div className="space-y-1.5">
            <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
              Writeback target <span className="text-muted-foreground/60 normal-case tracking-normal">(optional)</span>
            </Label>
            <div className="grid grid-cols-3 gap-1.5">
              <WritebackOption
                active={writebackTarget === ""}
                onClick={() => setWritebackTarget("")}
                label="None"
                sub="Log only"
                icon={<X className="w-3 h-3" />}
              />
              <WritebackOption
                active={writebackTarget === "message"}
                onClick={() => setWritebackTarget("message")}
                label="Message"
                sub="Post result"
                icon={<MessageSquare className="w-3 h-3" />}
              />
              <WritebackOption
                active={writebackTarget === "list_item"}
                onClick={() => setWritebackTarget("list_item")}
                label="List item"
                sub="Create task"
                icon={<ListTodo className="w-3 h-3" />}
              />
            </div>
          </div>

          {/* Per-target fields */}
          {writebackTarget === "message" && (
            <div className="space-y-1 pl-2 border-l-2 border-sky-500/40">
              <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
                Thread <span className="text-muted-foreground/60 normal-case tracking-normal">(optional)</span>
              </Label>
              <input
                value={threadId}
                onChange={e => setThreadId(e.target.value)}
                placeholder="thread-id (leave empty for main channel)"
                className="w-full h-7 px-2 text-xs font-mono rounded border bg-white dark:bg-[#19171d] focus:outline-none focus:ring-1 focus:ring-sky-500"
                disabled={isSubmitting}
              />
              <p className="text-[9px] text-muted-foreground/70">Will post to <code>#channel</code> unless a thread id is supplied.</p>
            </div>
          )}
          {writebackTarget === "list_item" && (
            <div className="space-y-1 pl-2 border-l-2 border-violet-500/40">
              <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Target list</Label>
              <select
                value={selectedListId}
                onChange={e => setSelectedListId(e.target.value)}
                className="w-full h-7 px-2 text-xs font-medium rounded border bg-white dark:bg-[#19171d] focus:outline-none focus:ring-1 focus:ring-violet-500"
                disabled={isSubmitting}
              >
                <option value="">Select a list…</option>
                {channelLists.map(l => (
                  <option key={l.id} value={l.id}>{l.title}</option>
                ))}
              </select>
              {channelLists.length === 0 && (
                <p className="text-[10px] text-muted-foreground">No lists in this channel yet — create one first.</p>
              )}
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
            className="bg-amber-600 hover:bg-amber-700 text-white gap-1.5"
          >
            {isSubmitting ? <Loader2 className="w-3 h-3 animate-spin" /> : <Terminal className="w-3 h-3" />}
            Run
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function WritebackOption({
  active, onClick, label, sub, icon,
}: { active: boolean; onClick: () => void; label: string; sub: string; icon: React.ReactNode }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "flex flex-col items-start gap-0.5 px-2 py-1.5 rounded border text-left transition-colors",
        active
          ? "border-amber-500/60 bg-amber-500/10 text-amber-700 dark:text-amber-300"
          : "border-border bg-transparent text-muted-foreground hover:bg-muted/40",
      )}
    >
      <span className="flex items-center gap-1 text-[10px] font-black uppercase tracking-wider">{icon}{label}</span>
      <span className="text-[9px] opacity-70">{sub}</span>
    </button>
  )
}
