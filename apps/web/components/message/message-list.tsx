"use client"
import { Message } from "@/types"
import { MessageItem } from "./message-item"
import { ScrollArea } from "@/components/ui/scroll-area"
import { format, isSameDay } from "date-fns"
import { useEffect, useRef, useState } from "react"
import { useUserStore } from "@/stores/user-store"
import { useUnreadStore } from "@/stores/unread-store"
import { useChannelStore } from "@/stores/channel-store"

interface MessageListProps {
  messages: Message[]
}

export function MessageList({ messages }: MessageListProps) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const [mounted, setMounted] = useState(false)
  const { users } = useUserStore()
  const { currentChannel } = useChannelStore()
  const { getLastRead, markAsRead } = useUnreadStore()

  const lastRead = currentChannel ? getLastRead(currentChannel.id) : null
  let unreadDividerShown = false

  useEffect(() => {
    setMounted(true)
    return () => {
      if (currentChannel) markAsRead(currentChannel.id)
    }
  }, [currentChannel, markAsRead])

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    if (scrollRef.current) {
      const scrollContainer = scrollRef.current.querySelector('[data-radix-scroll-area-viewport]')
      if (scrollContainer) {
        scrollContainer.scrollTop = scrollContainer.scrollHeight
      }
    }
  }, [messages])

  return (
    <ScrollArea ref={scrollRef} className="h-full">
      <div className="p-4 flex flex-col gap-0.5 min-h-full justify-end">
        {messages.map((msg, idx) => {
          const sender = users.find(u => u.id === msg.senderId) || {
            id: msg.senderId,
            name: "Unknown User",
            email: "unknown@example.com",
            avatar: "https://github.com/shadcn.png",
            status: "offline" as const,
          }
          const prevMsg = idx > 0 ? messages[idx - 1] : null
          
          // Logic for date separator
          const showDateSeparator = !prevMsg || !isSameDay(new Date(msg.createdAt), new Date(prevMsg.createdAt))
          
          // Logic for compact mode (same sender, same day, and within 5 minutes)
          const isCompact = !showDateSeparator && 
            prevMsg?.senderId === msg.senderId && 
            (new Date(msg.createdAt).getTime() - new Date(prevMsg.createdAt).getTime()) < 5 * 60 * 1000

          // Logic for unread divider
          const isUnread = lastRead && new Date(msg.createdAt) > new Date(lastRead)
          const showUnreadDivider = isUnread && !unreadDividerShown
          if (showUnreadDivider) unreadDividerShown = true

          return (
            <div key={msg.id}>
              {showDateSeparator && (
                <div className="flex items-center gap-4 my-6 relative">
                  <div className="h-[1px] bg-border flex-1" />
                  <div className="px-4 py-1 border rounded-full text-xs font-bold bg-white dark:bg-[#1a1d21] shadow-sm sticky top-0 z-10">
                    {mounted ? format(new Date(msg.createdAt), "EEEE, MMMM do") : null}
                  </div>
                  <div className="h-[1px] bg-border flex-1" />
                </div>
              )}

              {showUnreadDivider && (
                <div className="flex items-center gap-2 my-4 px-4">
                  <div className="h-[1px] bg-[#e01e5a] flex-1 opacity-50" />
                  <span className="text-[10px] font-black text-[#e01e5a] uppercase tracking-widest bg-white dark:bg-[#1a1d21] px-3 py-1 border border-[#e01e5a]/20 rounded-full">New Messages</span>
                  <div className="h-[1px] bg-[#e01e5a] flex-1 opacity-50" />
                </div>
              )}

              <MessageItem 
                message={msg} 
                sender={sender} 
                isCompact={isCompact}
              />
            </div>
          )
        })}
      </div>
    </ScrollArea>
  )
}
