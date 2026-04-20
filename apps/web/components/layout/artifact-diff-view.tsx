"use client"

import { ArtifactDiff } from "@/stores/artifact-store"
import { ScrollArea } from "@/components/ui/scroll-area"
import { cn } from "@/lib/utils"

interface ArtifactDiffViewProps {
  diff: ArtifactDiff
}

export function ArtifactDiffView({ diff }: ArtifactDiffViewProps) {
  // Parse the unified diff into chunks for rendering
  const lines = diff.unifiedDiff.split('\n')
  
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-white dark:bg-[#1a1d21]">
      <div className="p-4 border-b bg-muted/20 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-4">
          <span className="text-xs font-bold uppercase tracking-wider text-muted-foreground">Comparing v{diff.fromVersion} → v{diff.toVersion}</span>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1">
              <div className="w-2 h-2 rounded-full bg-green-500" />
              <span className="text-[10px] font-bold text-green-600">+{diff.summary.added}</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-2 h-2 rounded-full bg-red-500" />
              <span className="text-[10px] font-bold text-red-600">-{diff.summary.removed}</span>
            </div>
          </div>
        </div>
      </div>

      <ScrollArea className="flex-1 font-mono text-xs">
        <div className="p-6 max-w-4xl mx-auto space-y-0.5">
          {lines.map((line, i) => {
            const isAddition = line.startsWith('+') && !line.startsWith('+++')
            const isDeletion = line.startsWith('-') && !line.startsWith('---')
            const isHeader = line.startsWith('@@')
            const isFileInfo = line.startsWith('---') || line.startsWith('+++')

            if (isFileInfo) return null

            return (
              <div 
                key={i} 
                className={cn(
                  "px-2 py-0.5 rounded whitespace-pre-wrap break-all",
                  isAddition && "bg-green-500/10 text-green-700 dark:text-green-400 border-l-2 border-green-500",
                  isDeletion && "bg-red-500/10 text-red-700 dark:text-red-400 border-l-2 border-red-500 line-through opacity-70",
                  isHeader && "bg-blue-500/5 text-blue-500/50 my-2 font-bold py-1",
                  !isAddition && !isDeletion && !isHeader && "text-muted-foreground opacity-50"
                )}
              >
                {line}
              </div>
            )
          })}
        </div>
      </ScrollArea>
    </div>
  )
}
