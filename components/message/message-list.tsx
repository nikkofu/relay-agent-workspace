"use client"

import { Message, User } from "@/types"
import { MessageItem } from "./message-item"
import { ScrollArea } from "@/components/ui/scroll-area"
import { USERS } from "@/lib/mock-data"
import { format, isSameDay, isSameMinute } from "date-fns"
import { useEffect, useRef } from "react"

interface MessageListProps {
  messages: Message[]
}

export function MessageList({ messages }: MessageListProps) {
  const scrollRef = useRef<HTMLDivElement>(null)

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
          const sender = USERS.find(u => u.id === msg.senderId)!
          const prevMsg = idx > 0 ? messages[idx - 1] : null
          
          // Logic for date separator
          const showDateSeparator = !prevMsg || !isSameDay(new Date(msg.createdAt), new Date(prevMsg.createdAt))
          
          // Logic for compact mode (same sender, same day, and within 5 minutes)
          const isCompact = !showDateSeparator && 
            prevMsg?.senderId === msg.senderId && 
            (new Date(msg.createdAt).getTime() - new Date(prevMsg.createdAt).getTime()) < 5 * 60 * 1000

          return (
            <div key={msg.id}>
              {showDateSeparator && (
                <div className="flex items-center gap-4 my-6 relative">
                  <div className="h-[1px] bg-border flex-1" />
                  <div className="px-4 py-1 border rounded-full text-xs font-bold bg-white dark:bg-[#1a1d21] shadow-sm sticky top-0 z-10">
                    {format(new Date(msg.createdAt), "EEEE, MMMM do")}
                  </div>
                  <div className="h-[1px] bg-border flex-1" />
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
