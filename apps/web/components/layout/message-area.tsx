"use client"

import { useChannelStore } from "@/stores/channel-store"
import { useMessageStore } from "@/stores/message-store"
import { Hash, Lock, Info, Phone, Video, Star } from "lucide-react"
import { Button } from "@/components/ui/button"
import { MessageComposer } from "@/components/message/message-composer"
import { useEffect, useRef } from "react"
import { AgentCollabDashboard } from "./agent-collab-dashboard"
import { ChannelInfo } from "@/components/channel/channel-info"
import { TypingIndicator } from "@/components/message/typing-indicator"
import { cn } from "@/lib/utils"

export function MessageArea({ children }: { children?: React.ReactNode }) {
  const { currentChannel, toggleStar } = useChannelStore()
  const { fetchMessages, sendMessage } = useMessageStore()
  const lastFetchedId = useRef<string | null>(null)
  
  useEffect(() => {
    if (currentChannel && currentChannel.id !== lastFetchedId.current) {
      lastFetchedId.current = currentChannel.id
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
          <ChannelInfo trigger={
            <Button variant="ghost" size="sm" className="px-1 hover:bg-muted font-bold text-lg text-foreground">
              <Icon className="w-4 h-4 mr-1" />
              <span className="truncate">{currentChannel.name}</span>
            </Button>
          } />
          <Button 
            variant="ghost" 
            size="icon" 
            className={cn("h-6 w-6", currentChannel.isStarred && "text-yellow-500")}
            onClick={() => toggleStar(currentChannel.id)}
          >
            <Star className={cn("w-4 h-4", currentChannel.isStarred && "fill-current")} />
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
          <ChannelInfo trigger={
            <Button variant="ghost" size="icon" className="h-8 w-8">
              <Info className="w-4 h-4" />
            </Button>
          } />
        </div>
      </header>

      {/* Messages List Area */}
      <div className="flex-1 overflow-hidden relative flex flex-col">
        {currentChannel.name === "agent-collab" && <AgentCollabDashboard />}
        <div className="flex-1 overflow-hidden relative flex flex-col">
          {children}
        </div>
      </div>

      <TypingIndicator scope={`channel:${currentChannel.id}`} />

      {/* Message Composer */}
      <MessageComposer 
        placeholder={`Message #${currentChannel.name}`} 
        scope={`channel:${currentChannel.id}`}
        onSend={(content, artifactIds, fileIds) => {
          if (currentChannel) {
            sendMessage(currentChannel.id, content, "user-1", undefined, artifactIds, fileIds)
          }
        }}
      />
    </div>
  )
}
