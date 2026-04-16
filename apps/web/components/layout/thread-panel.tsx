"use client"

import { useUIStore } from "@/stores/ui-store"
import { X, Hash, Lock } from "lucide-react"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useChannelStore } from "@/stores/channel-store"
import { AIThreadSummary } from "@/components/ai-chat/ai-thread-summary"

export function ThreadPanel() {
  const { isThreadOpen, closeThread, activeThreadId } = useUIStore()
  const { currentChannel } = useChannelStore()

  if (!isThreadOpen) return null

  const Icon = currentChannel?.type === "private" ? Lock : Hash

  return (
    <div className="w-full h-full flex flex-col bg-white dark:bg-[#1a1d21] border-l animate-in slide-in-from-right duration-300">
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0">
        <div className="flex flex-col">
          <h3 className="font-bold text-sm">Thread</h3>
          <div className="flex items-center text-[10px] text-muted-foreground font-medium">
            <Icon className="w-3 h-3 mr-1" />
            {currentChannel?.name}
          </div>
        </div>
        <Button variant="ghost" size="icon" className="h-8 w-8 rounded-full" onClick={closeThread}>
          <X className="w-4 h-4" />
        </Button>
      </header>
      
      <AIThreadSummary />

      <ScrollArea className="flex-1 p-4 pt-0">
        <div className="flex flex-col gap-4">
          <div className="text-xs text-muted-foreground text-center py-4 px-8 bg-muted/30 rounded-lg border border-dashed italic">
            This is the start of a thread. All replies will be grouped here.
          </div>
          {/* Placeholder for thread messages */}
          <div className="text-[11px] text-muted-foreground/50 text-center py-2 uppercase tracking-widest font-bold">
            Thread history for {activeThreadId}
          </div>
        </div>
      </ScrollArea>

      <div className="p-4 border-t bg-white dark:bg-[#1a1d21]">
        <div className="border rounded-lg p-2 focus-within:ring-1 focus-within:ring-ring transition-all">
          <textarea 
            placeholder="Reply..." 
            className="w-full resize-none bg-transparent border-none focus:ring-0 text-sm h-20 outline-none placeholder:text-muted-foreground/50"
          />
          <div className="flex justify-end mt-1">
            <Button size="sm" className="h-7 text-[11px] bg-[#007a5a] hover:bg-[#007a5a]/90">Reply</Button>
          </div>
        </div>
      </div>
    </div>
  )
}
