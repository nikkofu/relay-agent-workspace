"use client"

import { useChannelStore } from "@/stores/channel-store"
import { useMessageStore } from "@/stores/message-store"
import { Hash, Lock, Info, Phone, Video, Star } from "lucide-react"
import { Button } from "@/components/ui/button"
import { MessageComposer } from "@/components/message/message-composer"
import { useEffect } from "react"

export function MessageArea({ children }: { children?: React.ReactNode }) {
  const { currentChannel } = useChannelStore()
  const { fetchMessages, sendMessage } = useMessageStore()
  
  useEffect(() => {
    if (currentChannel) {
      fetchMessages(currentChannel.id)
    }
  }, [currentChannel, fetchMessages])

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

      {/* Message Composer */}
      <MessageComposer 
        placeholder={`Message #${currentChannel.name}`} 
        onSend={(content) => {
          if (currentChannel) {
            sendMessage(currentChannel.id, content, "user-1")
          }
        }}
      />
    </div>
  )
}
