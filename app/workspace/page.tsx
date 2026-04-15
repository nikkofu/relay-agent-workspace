"use client"

import { useChannelStore } from "@/stores/channel-store"
import { useMessageStore } from "@/stores/message-store"
import { ScrollArea } from "@/components/ui/scroll-area"
import { UserAvatar } from "@/components/common/user-avatar"
import { format } from "date-fns"
import { USERS } from "@/lib/mock-data"

export default function WorkspacePage() {
  const { currentChannel } = useChannelStore()
  const { messages } = useMessageStore()

  const channelMessages = messages.filter(m => m.channelId === currentChannel?.id)

  return (
    <ScrollArea className="h-full">
      <div className="p-4 flex flex-col gap-6">
        {/* Channel Introduction */}
        <div className="mb-8">
          <div className="w-12 h-12 rounded bg-muted flex items-center justify-center mb-2 font-bold text-xl">
            #
          </div>
          <h1 className="text-2xl font-bold">This is the very beginning of the #{currentChannel?.name} channel</h1>
          <p className="text-muted-foreground mt-1">
            {currentChannel?.description || "This channel is for everything #"+currentChannel?.name+". Hold meetings, share docs, and make decisions together."}
          </p>
        </div>

        {/* Messages List */}
        <div className="flex flex-col gap-4">
          {channelMessages.map((msg) => {
            const sender = USERS.find(u => u.id === msg.senderId)
            return (
              <div key={msg.id} className="group flex items-start gap-3 hover:bg-muted/50 -mx-4 px-4 py-2 transition-colors">
                <div className="shrink-0 mt-1">
                  <div className="w-9 h-9 rounded overflow-hidden bg-muted">
                    {sender?.avatar ? (
                      <img src={sender.avatar} alt={sender.name} className="w-full h-full object-cover" />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center font-bold text-xs uppercase">
                        {sender?.name.substring(0, 2)}
                      </div>
                    )}
                  </div>
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-bold text-sm hover:underline cursor-pointer">{sender?.name}</span>
                    <span className="text-[10px] text-muted-foreground">
                      {format(new Date(msg.createdAt), "h:mm a")}
                    </span>
                  </div>
                  <div className="text-sm mt-0.5 whitespace-pre-wrap break-words">
                    {msg.content}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </ScrollArea>
  )
}
