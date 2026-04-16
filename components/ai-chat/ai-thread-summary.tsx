"use client"

import { useState } from "react"
import { Sparkles, Loader2, RefreshCw } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export function AIThreadSummary() {
  const [summary, setSummary] = useState("")
  const [loading, setLoading] = useState(false)

  const generateSummary = async () => {
    setLoading(true)
    setSummary("")
    // 模拟 AI 处理延迟
    await new Promise(r => setTimeout(r, 1200))
    setSummary("The team is aligning on the Relay Agent Workspace v1.0 release. Key consensus reached on implementing AI Canvas first. Nikko will handle the core layout, and John will provide the initial mock data by EOD.")
    setLoading(false)
  }

  if (!summary && !loading) {
    return (
      <div className="px-4 py-3">
        <Button 
          variant="outline" 
          size="sm" 
          className="w-full justify-start text-[11px] font-bold text-purple-600 bg-purple-50/50 dark:bg-purple-900/10 border-purple-200 dark:border-purple-800 hover:bg-purple-100 hover:text-purple-700 transition-all group"
          onClick={generateSummary}
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
        loading && "animate-pulse"
      )}>
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2 text-purple-600 dark:text-purple-400">
            {loading ? <Loader2 className="w-3 h-3 animate-spin" /> : <Sparkles className="w-3 h-3" />}
            <span className="text-[10px] font-bold uppercase tracking-wider">AI Summary</span>
          </div>
          {!loading && (
            <Button variant="ghost" size="icon" className="h-5 w-5 text-purple-400 hover:text-purple-600 hover:bg-transparent" onClick={generateSummary}>
              <RefreshCw className="w-3 h-3" />
            </Button>
          )}
        </div>
        <p className="text-xs leading-relaxed text-muted-foreground font-medium">
          {loading ? "Analyzing conversation context..." : summary}
        </p>
      </div>
    </div>
  )
}
