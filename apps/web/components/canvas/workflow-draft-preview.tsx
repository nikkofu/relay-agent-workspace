"use client"

// ── Workflow Draft Preview ────────────────────────────────────────────────────
//
// Renders the Dock-local workflow draft produced by
// POST /ai/canvas/generate-workflow-draft.
//
// The user reviews title, goal, and ordered steps before clicking
// "Confirm Create Workflow". On confirm we call
// POST /ai/canvas/confirm-create-workflow with the immutable draft ID.
// On cancel the draft is discarded locally — the backend draft row
// persists but is harmless (no workflow is ever created without confirm).

import { useState } from "react"
import { Loader2, CheckCircle2, X, GitBranch, Target, List } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import type { AnalysisWorkflowDraft } from "@/lib/analysis-draft-contract"

interface WorkflowDraftPreviewProps {
  draft: AnalysisWorkflowDraft
  onConfirm: () => Promise<string | null>
  onCancel: () => void
}

export function WorkflowDraftPreview({ draft, onConfirm, onCancel }: WorkflowDraftPreviewProps) {
  const [state, setState] = useState<"idle" | "confirming" | "done" | "error">("idle")
  const [workflowId, setWorkflowId] = useState<string | null>(null)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  const handleConfirm = async () => {
    setState("confirming")
    setErrorMsg(null)
    const id = await onConfirm()
    if (id) {
      setWorkflowId(id)
      setState("done")
    } else {
      setErrorMsg("Failed to create workflow. The draft is still saved.")
      setState("error")
    }
  }

  return (
    <div className="rounded-xl border border-amber-500/30 bg-amber-500/5 dark:bg-amber-500/10 overflow-hidden animate-in fade-in slide-in-from-bottom-2 duration-200">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-amber-500/20 bg-amber-500/10">
        <div className="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-widest text-amber-700 dark:text-amber-300">
          <GitBranch className="w-3 h-3" />
          Workflow Draft
        </div>
        {state !== "done" && (
          <button
            type="button"
            onClick={onCancel}
            className="p-0.5 rounded hover:bg-amber-500/20 text-amber-600 transition-colors"
            title="Discard draft"
          >
            <X className="w-3 h-3" />
          </button>
        )}
      </div>

      {state === "done" ? (
        <div className="px-3 py-4 flex flex-col items-center gap-2 text-center">
          <CheckCircle2 className="w-7 h-7 text-emerald-500" />
          <p className="text-xs font-bold text-foreground">Workflow created</p>
          <p className="text-[10px] text-muted-foreground font-mono break-all">{workflowId}</p>
          <Button size="sm" variant="ghost" className="h-6 text-[10px] gap-1" onClick={onCancel}>
            Dismiss
          </Button>
        </div>
      ) : (
        <div className="p-3 space-y-3">
          {/* Title */}
          <div>
            <p className="text-[9px] font-black uppercase tracking-widest text-muted-foreground mb-0.5">
              Title
            </p>
            <p className="text-xs font-bold text-foreground">{draft.title}</p>
          </div>

          {/* Goal */}
          <div>
            <div className="flex items-center gap-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground mb-0.5">
              <Target className="w-2.5 h-2.5" />
              Goal
            </div>
            <p className="text-[11px] text-foreground/80 leading-snug">{draft.goal}</p>
          </div>

          {/* Steps */}
          {draft.steps.length > 0 && (
            <div>
              <div className="flex items-center gap-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground mb-1">
                <List className="w-2.5 h-2.5" />
                Steps ({draft.steps.length})
              </div>
              <div className="space-y-1">
                {draft.steps.map((step, i) => (
                  <div key={i} className="flex gap-2 items-start text-[11px] leading-snug">
                    <span className="text-[9px] font-mono text-muted-foreground/60 shrink-0 mt-0.5 w-4 text-right">
                      {i + 1}.
                    </span>
                    <span className="text-foreground/80">{step.title}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {errorMsg && (
            <p className="text-[10px] text-rose-600 leading-snug">{errorMsg}</p>
          )}

          {/* Actions */}
          <div className="flex items-center gap-2 pt-1">
            <Button
              size="sm"
              className={cn(
                "flex-1 h-7 text-[10px] font-black uppercase tracking-wider gap-1",
                "bg-amber-600 hover:bg-amber-700 text-white",
              )}
              onClick={handleConfirm}
              disabled={state === "confirming"}
            >
              {state === "confirming"
                ? <Loader2 className="w-3 h-3 animate-spin" />
                : <GitBranch className="w-3 h-3" />}
              {state === "confirming" ? "Creating…" : "Confirm Create Workflow"}
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
