"use client"

import { useState } from "react"
import { AnalysisListDraft } from "@/lib/analysis-list-draft"
import { Button } from "@/components/ui/button"
import { Check, Loader2, AlertCircle, ArrowLeft, ExternalLink } from "lucide-react"
import { cn } from "@/lib/utils"

interface AnalysisListDraftPreviewProps {
  draft: AnalysisListDraft
  onConfirm: () => Promise<string | null>
  onCancel: () => void
  onOpenList: (listId: string) => void
}

export function AnalysisListDraftPreview({
  draft,
  onConfirm,
  onCancel,
  onOpenList,
}: AnalysisListDraftPreviewProps) {
  const [isConfirming, setIsConfirming] = useState(false)
  const [createdListId, setCreatedListId] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleConfirm = async () => {
    setIsConfirming(true)
    setError(null)
    try {
      const listId = await onConfirm()
      if (listId) {
        setCreatedListId(listId)
      } else {
        setError("Failed to create list. Please try again.")
      }
    } catch (err: any) {
      setError(err.message || "An unexpected error occurred.")
    } finally {
      setIsConfirming(false)
    }
  }

  if (createdListId) {
    return (
      <div className="p-4 bg-emerald-500/10 dark:bg-emerald-500/5 border border-emerald-500/20 rounded-xl space-y-4 animate-in fade-in zoom-in-95 duration-300">
        <div className="flex items-center gap-3 text-emerald-700 dark:text-emerald-400">
          <div className="w-10 h-10 rounded-full bg-emerald-500/20 flex items-center justify-center shrink-0">
            <Check className="w-6 h-6" />
          </div>
          <div>
            <h4 className="font-bold text-sm">List created successfully!</h4>
            <p className="text-xs opacity-80">Your execution plan is now a workspace list.</p>
          </div>
        </div>

        <div className="flex gap-2">
          <Button 
            className="flex-1 gap-2 h-9 text-xs font-bold uppercase tracking-wider"
            onClick={() => onOpenList(createdListId)}
          >
            <ExternalLink className="w-4 h-4" />
            Open List
          </Button>
          <Button 
            variant="ghost" 
            className="flex-1 h-9 text-xs font-bold uppercase tracking-wider"
            onClick={onCancel}
          >
            Back to analysis
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4 animate-in fade-in slide-in-from-right-4 duration-300">
      <div className="flex items-center justify-between">
        <Button 
          variant="ghost" 
          size="sm" 
          className="h-7 px-2 text-[10px] font-black uppercase tracking-widest gap-1.5"
          onClick={onCancel}
          disabled={isConfirming}
        >
          <ArrowLeft className="w-3 h-3" />
          Cancel
        </Button>
        <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
          Draft Preview
        </span>
      </div>

      <section className="bg-white dark:bg-[#1a1d21] border rounded-lg overflow-hidden shadow-sm">
        <div className="px-3 py-2 border-b bg-muted/30 flex items-center gap-2">
          <ListCircle className="w-4 h-4 text-purple-500" />
          <h4 className="font-bold text-sm truncate">{draft.title}</h4>
        </div>
        
        <div className="p-3 space-y-2 max-h-[240px] overflow-y-auto">
          {draft.items.map((item, i) => (
            <div key={i} className="flex gap-2.5 items-start py-1">
              <div className="w-4 h-4 rounded border border-muted-foreground/30 mt-0.5 shrink-0" />
              <span className="text-xs leading-relaxed">{item.title}</span>
            </div>
          ))}
        </div>

        <div className="px-3 py-2 border-t bg-muted/10 flex items-center justify-between">
          <span className="text-[9px] font-bold text-muted-foreground uppercase">
            Destination
          </span>
          <span className="text-[9px] font-black text-purple-600 dark:text-purple-400 uppercase">
            # {draft.channel_id}
          </span>
        </div>
      </section>

      {error && (
        <div className="flex items-start gap-2 p-2.5 bg-red-500/10 border border-red-500/20 rounded-lg text-red-600 dark:text-red-400 text-xs">
          <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
          <p>{error}</p>
        </div>
      )}

      <Button 
        className="w-full h-10 gap-2 text-sm font-black uppercase tracking-widest shadow-lg"
        onClick={handleConfirm}
        disabled={isConfirming}
      >
        {isConfirming ? (
          <>
            <Loader2 className="w-4 h-4 animate-spin" />
            Creating...
          </>
        ) : (
          <>
            <Check className="w-4 h-4" />
            Confirm and Create
          </>
        )}
      </Button>
      
      <p className="text-[10px] text-center text-muted-foreground px-4">
        This will create one list with {draft.items.length} items in the specified channel.
      </p>
    </div>
  )
}

function ListCircle(props: any) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <circle cx="12" cy="12" r="10" />
      <line x1="10" y1="8" x2="16" y2="8" />
      <line x1="10" y1="12" x2="16" y2="12" />
      <line x1="10" y1="16" x2="16" y2="16" />
      <line x1="8" y1="8" x2="8" y2="8.01" />
      <line x1="8" y1="12" x2="8" y2="12.01" />
      <line x1="8" y1="16" x2="8" y2="16.01" />
    </svg>
  )
}
