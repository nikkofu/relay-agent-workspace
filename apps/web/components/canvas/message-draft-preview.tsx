"use client"

// ── Message Draft Preview ─────────────────────────────────────────────────────
//
// Renders the Dock-local channel-message draft produced by
// POST /ai/canvas/generate-message-draft.
//
// The user reviews channel + body before clicking "Confirm Publish".
// On confirm we call POST /ai/canvas/confirm-publish-message with the
// immutable draft ID. On cancel the draft is discarded locally.

import { useState } from "react"
import { Loader2, CheckCircle2, X, MessageSquare, Hash } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import type { AnalysisMessageDraft } from "@/lib/analysis-draft-contract"

interface MessageDraftPreviewProps {
  draft: AnalysisMessageDraft
  onConfirm: () => Promise<string | null>
  onCancel: () => void
}

export function MessageDraftPreview({ draft, onConfirm, onCancel }: MessageDraftPreviewProps) {
  const [state, setState] = useState<"idle" | "confirming" | "done" | "error">("idle")
  const [messageId, setMessageId] = useState<string | null>(null)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  const handleConfirm = async () => {
    setState("confirming")
    setErrorMsg(null)
    const id = await onConfirm()
    if (id) {
      setMessageId(id)
      setState("done")
    } else {
      setErrorMsg("Failed to publish message. The draft is still saved.")
      setState("error")
    }
  }

  return (
    <div className="rounded-xl border border-sky-500/30 bg-sky-500/5 dark:bg-sky-500/10 overflow-hidden animate-in fade-in slide-in-from-bottom-2 duration-200">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-sky-500/20 bg-sky-500/10">
        <div className="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-widest text-sky-700 dark:text-sky-300">
          <MessageSquare className="w-3 h-3" />
          Message Draft
        </div>
        {state !== "done" && (
          <button
            type="button"
            onClick={onCancel}
            className="p-0.5 rounded hover:bg-sky-500/20 text-sky-600 transition-colors"
            title="Discard draft"
          >
            <X className="w-3 h-3" />
          </button>
        )}
      </div>

      {state === "done" ? (
        <div className="px-3 py-4 flex flex-col items-center gap-2 text-center">
          <CheckCircle2 className="w-7 h-7 text-emerald-500" />
          <p className="text-xs font-bold text-foreground">Message published</p>
          <p className="text-[10px] text-muted-foreground font-mono break-all">{messageId}</p>
          <Button size="sm" variant="ghost" className="h-6 text-[10px] gap-1" onClick={onCancel}>
            Dismiss
          </Button>
        </div>
      ) : (
        <div className="p-3 space-y-3">
          {/* Channel */}
          <div className="flex items-center gap-1.5">
            <Hash className="w-3 h-3 text-sky-600 shrink-0" />
            <span className="text-[11px] text-sky-700 dark:text-sky-300 font-bold truncate">
              {draft.channel_id}
            </span>
          </div>

          {/* Body */}
          <div>
            <p className="text-[9px] font-black uppercase tracking-widest text-muted-foreground mb-1">
              Message
            </p>
            <div className="bg-white dark:bg-[#1a1d21] border rounded-lg px-3 py-2 text-[11px] leading-relaxed text-foreground/90 whitespace-pre-wrap break-words max-h-40 overflow-y-auto">
              {draft.body}
            </div>
          </div>

          {errorMsg && (
            <p className="text-[10px] text-rose-600 leading-snug">{errorMsg}</p>
          )}

          {/* Actions */}
          <div className="flex items-center gap-2 pt-1">
            <Button
              size="sm"
              className={cn(
                "flex-1 h-7 text-[10px] font-black uppercase tracking-wider gap-1",
                "bg-sky-600 hover:bg-sky-700 text-white",
              )}
              onClick={handleConfirm}
              disabled={state === "confirming"}
            >
              {state === "confirming"
                ? <Loader2 className="w-3 h-3 animate-spin" />
                : <MessageSquare className="w-3 h-3" />}
              {state === "confirming" ? "Publishing…" : "Confirm Publish"}
            </Button>
            <Button
              size="sm"
              variant="ghost"
              className="h-7 text-[10px] text-muted-foreground"
              onClick={onCancel}
              disabled={state === "confirming"}
            >
              Cancel
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
