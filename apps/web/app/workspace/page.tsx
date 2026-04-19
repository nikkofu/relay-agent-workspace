"use client"

import { useChannelStore } from "@/stores/channel-store"
import { useMessageStore } from "@/stores/message-store"
import { MessageList } from "@/components/message/message-list"
import { MessageArea } from "@/components/layout/message-area"
import { useSearchParams } from "next/navigation"
import { Suspense, useEffect } from "react"

function WorkspaceContent() {
  const { currentChannel, setCurrentChannelById, channels } = useChannelStore()
  const { messages } = useMessageStore()
  const searchParams = useSearchParams()
  const channelIdFromUrl = searchParams.get("c")

  useEffect(() => {
    if (channelIdFromUrl && channels.length > 0) {
      setCurrentChannelById(channelIdFromUrl)
    }
  }, [channelIdFromUrl, channels, setCurrentChannelById])

  if (!currentChannel) {
    return (
      <div className="flex-1 flex items-center justify-center bg-white dark:bg-[#1a1d21] text-muted-foreground">
        Select a channel to start messaging
      </div>
    )
  }

  const channelMessages = messages.filter(m => m.channelId === currentChannel.id && !m.threadId)

  return (
    <MessageArea>
      <div className="flex flex-col h-full">
        {/* Channel Introduction */}
        <div className="p-4 border-b bg-muted/10">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded bg-muted flex items-center justify-center font-bold">
              #
            </div>
            <div>
              <h1 className="text-sm font-bold">Welcome to #{currentChannel.name}</h1>
              <p className="text-xs text-muted-foreground truncate max-w-md">
                {currentChannel.description || `This is the start of the #${currentChannel.name} channel.`}
              </p>
            </div>
          </div>
        </div>

        {/* Messages List */}
        <div className="flex-1 min-h-0 flex flex-col">
          <MessageList messages={channelMessages} className="flex-1" />
        </div>
      </div>
    </MessageArea>
  )
}

export default function WorkspacePage() {
  return (
    <Suspense fallback={<div className="flex-1 flex items-center justify-center bg-white dark:bg-[#1a1d21] text-muted-foreground">Loading workspace...</div>}>
      <WorkspaceContent />
    </Suspense>
  )
}
