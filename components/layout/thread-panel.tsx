"use client"

import { useUIStore } from "@/stores/ui-store"
import { X, Hash, Lock } from "lucide-react"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useChannelStore } from "@/stores/channel-store"

export function ThreadPanel() {
  const { isThreadOpen, closeThread, activeThreadId } = useUIStore()
  const { currentChannel } = useChannelStore()

  if (!isThreadOpen) return null

  const Icon = currentChannel?.type === "private" ? Lock : Hash

  return (
    <div className="w-full h-full flex flex-col bg-white dark:bg-[#1a1d21] border-l">
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0">
        <div className="flex flex-col">
          <h3 className="font-bold">Thread</h3>
          <div className="flex items-center text-xs text-muted-foreground">
            <Icon className="w-3 h-3 mr-1" />
            {currentChannel?.name}
          </div>
        </div>
        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={closeThread}>
          <X className="w-4 h-4" />
        </Button>
      </header>
      
      <ScrollArea className="flex-1 p-4">
        <div className="text-sm text-muted-foreground text-center py-8">
          Thread content for {activeThreadId} would go here.
        </div>
      </ScrollArea>

      <div className="p-4 border-t">
        <div className="border rounded-lg p-2">
          <textarea 
            placeholder="Reply..." 
            className="w-full resize-none bg-transparent border-none focus:ring-0 text-sm h-20 outline-none"
          />
        </div>
      </div>
    </div>
  )
}
