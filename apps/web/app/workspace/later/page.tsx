"use client"

import { Bookmark } from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useLaterStore } from "@/stores/later-store"
import { useEffect } from "react"
import { MessageItem } from "@/components/message/message-item"

export default function LaterPage() {
  const { items, fetchLaterItems } = useLaterStore()

  useEffect(() => {
    fetchLaterItems()
  }, [fetchLaterItems])

  return (
    <div className="flex-1 flex flex-col h-full bg-white dark:bg-[#1a1d21]">
      <header className="h-14 px-4 flex items-center border-b shrink-0 bg-white dark:bg-[#1a1d21] z-10">
        <h2 className="font-bold text-lg text-foreground">Later</h2>
      </header>

      <ScrollArea className="flex-1">
        <div className="p-4 flex flex-col gap-4">
          {items.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-center">
              <div className="w-16 h-16 rounded-full bg-muted flex items-center justify-center mb-4">
                <Bookmark className="w-8 h-8 text-muted-foreground opacity-20" />
              </div>
              <h3 className="font-bold text-lg mb-1 text-foreground">Your saved items</h3>
              <p className="text-sm text-muted-foreground max-w-xs leading-relaxed">
                Save messages and files to come back to them later. Everything you save will appear here.
              </p>
            </div>
          ) : (
            <div className="flex flex-col gap-6">
              {items.map((item) => (
                <div key={item.id} className="border rounded-xl overflow-hidden bg-muted/20 pb-2">
                  <div className="px-4 py-2 bg-muted/40 border-b flex items-center justify-between">
                    <span className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
                      Saved from #{item.channel?.name || "Unknown"}
                    </span>
                  </div>
                  <MessageItem 
                    message={item.message} 
                    sender={item.user} 
                    showActions={true}
                  />
                </div>
              ))}
            </div>
          )}
        </div>
      </ScrollArea>
    </div>
  )
}
