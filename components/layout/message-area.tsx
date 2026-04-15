"use client"

import { useChannelStore } from "@/stores/channel-store"
import { useMessageStore } from "@/stores/message-store"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Hash, Lock, Info, Phone, Video, Search, Star, MoreVertical } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

export function MessageArea({ children }: { children?: React.ReactNode }) {
  const { currentChannel } = useChannelStore()
  
  if (!currentChannel) {
    return (
      <div className="flex-1 flex items-center justify-center bg-white dark:bg-[#1a1d21] text-muted-foreground">
        Select a channel to start messaging
      </div>
    )
  }

  const Icon = currentChannel.type === "private" ? Lock : Hash

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-white dark:bg-[#1a1d21]">
      {/* Channel Header */}
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0">
        <div className="flex items-center gap-1.5 min-w-0">
          <Button variant="ghost" size="sm" className="px-1 hover:bg-muted font-bold text-lg">
            <Icon className="w-4 h-4 mr-1" />
            <span className="truncate">{currentChannel.name}</span>
          </Button>
          <Button variant="ghost" size="icon" className="h-6 w-6">
            <Star className="w-4 h-4" />
          </Button>
        </div>
        
        <div className="flex items-center gap-2">
          <div className="hidden md:flex items-center -space-x-2 mr-2">
            {[1, 2, 3].map((i) => (
              <div key={i} className="w-6 h-6 rounded border-2 border-white dark:border-[#1a1d21] bg-muted flex items-center justify-center text-[10px] font-bold">
                {i}
              </div>
            ))}
            <div className="pl-3 text-xs font-semibold text-muted-foreground">
              {currentChannel.memberCount}
            </div>
          </div>
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <Phone className="w-4 h-4" />
          </Button>
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <Video className="w-4 h-4" />
          </Button>
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <Info className="w-4 h-4" />
          </Button>
        </div>
      </header>

      {/* Messages List Area */}
      <div className="flex-1 overflow-hidden relative">
        {children}
      </div>

      {/* Message Composer Placeholder */}
      <div className="p-4 pt-0">
        <div className="border rounded-lg p-2 focus-within:ring-1 focus-within:ring-ring">
          <Input 
            placeholder={`Message #${currentChannel.name}`} 
            className="border-none focus-visible:ring-0 px-1"
          />
          <div className="flex items-center justify-between mt-2 pt-2 border-t">
            <div className="flex items-center gap-1">
              {/* Toolbar placeholders */}
              <div className="w-6 h-6 rounded hover:bg-muted" />
              <div className="w-6 h-6 rounded hover:bg-muted" />
              <div className="w-6 h-6 rounded hover:bg-muted" />
            </div>
            <Button size="sm" className="bg-[#007a5a] hover:bg-[#007a5a]/90 h-7">Send</Button>
          </div>
        </div>
      </div>
    </div>
  )
}
