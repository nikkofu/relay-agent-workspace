"use client"

import { Sparkles, Loader2, RefreshCw } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { useMessageStore } from "@/stores/message-store"

interface AIThreadSummaryProps {
  messageId: string
}

export function AIThreadSummary({ messageId }: AIThreadSummaryProps) {
  const { currentThreadSummary, isSummaryLoading, generateThreadSummary } = useMessageStore()

  const handleGenerate = () => {
    if (messageId) {
      generateThreadSummary(messageId)
    }
  }

  if (!currentThreadSummary && !isSummaryLoading) {
    return (
      <div className="px-4 py-3">
        <Button 
          variant="outline" 
          size="sm" 
          className="w-full justify-start text-[11px] font-bold text-purple-600 bg-purple-50/50 dark:bg-purple-900/10 border-purple-200 dark:border-purple-800 hover:bg-purple-100 hover:text-purple-700 transition-all group"
          onClick={handleGenerate}
        >
          <Sparkles className="w-3 h-3 mr-2 text-purple-500 group-hover:animate-pulse" />
          Summarize this thread
        </Button>
      </div>
    )
  }

  return (
    <div className="px-4 py-3">
      <div className={cn(
        "bg-purple-50/50 dark:bg-purple-500/10 p-3 rounded-lg border border-purple-100 dark:border-purple-500/20 shadow-sm relative overflow-hidden transition-all",
        isSummaryLoading && "animate-pulse"
      )}>
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2 text-purple-600 dark:text-purple-400">
            {isSummaryLoading ? <Loader2 className="w-3 h-3 animate-spin" /> : <Sparkles className="w-3 h-3" />}
            <span className="text-[10px] font-bold uppercase tracking-wider">AI Summary</span>
          </div>
          {!isSummaryLoading && (
            <Button variant="ghost" size="icon" className="h-5 w-5 text-purple-400 hover:text-purple-600 hover:bg-transparent" onClick={handleGenerate}>
              <RefreshCw className="w-3 h-3" />
            </Button>
          )}
        </div>
        <p className="text-xs leading-relaxed text-muted-foreground font-medium">
          {isSummaryLoading ? "Analyzing conversation context..." : currentThreadSummary}
        </p>
      </div>
    </div>
  )
}
