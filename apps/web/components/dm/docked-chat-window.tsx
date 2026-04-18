"use client"

import { useEffect, useState } from "react"
import { X, Minus } from "lucide-react"
import { Button } from "@/components/ui/button"
import { UserAvatar } from "@/components/common/user-avatar"
import { MessageComposer } from "@/components/message/message-composer"
import { MessageList } from "@/components/message/message-list"
import { useUserStore } from "@/stores/user-store"
import { useDMStore } from "@/stores/dm-store"
import { useMessageStore } from "@/stores/message-store"
import { useUIStore } from "@/stores/ui-store"
import { cn } from "@/lib/utils"
import { ScrollArea } from "@/components/ui/scroll-area"

interface DockedChatWindowProps {
  userId: string
  index: number
}

export function DockedChatWindow({ userId, index }: DockedChatWindowProps) {
  const { users, currentUser } = useUserStore()
  const { conversations, createConversation } = useDMStore()
  const { messages, fetchDMMessages, sendDMMessage } = useMessageStore()
  const { closeDockedChat } = useUIStore()
  const [isMinimized, setIsMinimized] = useState(false)
  const [conversation, setConversation] = useState(
    conversations.find(c => c.userIds.includes(userId) && c.userIds.includes(currentUser?.id || ""))
  )

  const user = users.find(u => u.id === userId)

  useEffect(() => {
    if (!conversation && currentUser && userId) {
      createConversation([currentUser.id, userId]).then((nextConversation) => {
        setConversation(nextConversation ?? undefined)
      })
    }
  }, [conversation, currentUser, userId, createConversation])

  useEffect(() => {
    if (conversation) {
      fetchDMMessages(conversation.id)
    }
  }, [conversation, fetchDMMessages])

  if (!user) return null

  const chatMessages = conversation 
    ? messages.filter(m => m.dmId === conversation.id)
    : []

  return (
    <div 
      className={cn(
        "fixed bottom-0 bg-white dark:bg-[#1a1d21] border rounded-t-lg shadow-2xl transition-all duration-200 z-50 flex flex-col w-[320px]",
        isMinimized ? "h-10" : "h-[450px]"
      )}
      style={{ right: `${20 + index * 340}px` }}
    >
      {/* Header */}
      <header className="h-10 px-3 flex items-center justify-between border-b shrink-0 cursor-pointer hover:bg-muted/30 transition-colors rounded-t-lg" onClick={() => setIsMinimized(!isMinimized)}>
        <div className="flex items-center gap-2 min-w-0">
          <UserAvatar src={user.avatar} name={user.name} status={user.status} className="h-6 w-6" />
          <span className="text-sm font-bold truncate text-foreground">{user.name}</span>
        </div>
        <div className="flex items-center gap-0.5">
          <Button variant="ghost" size="icon" className="h-6 w-6" onClick={(e) => {
            e.stopPropagation()
            setIsMinimized(!isMinimized)
          }}>
            <Minus className="h-3 w-3" />
          </Button>
          <Button variant="ghost" size="icon" className="h-6 w-6" onClick={(e) => {
            e.stopPropagation()
            closeDockedChat(userId)
          }}>
            <X className="h-3 w-3" />
          </Button>
        </div>
      </header>

      {!isMinimized && (
        <>
          {/* Messages Area */}
          <div className="flex-1 overflow-hidden flex flex-col min-h-0">
            <ScrollArea className="flex-1">
              <div className="p-2 pt-4">
                <MessageList messages={chatMessages} />
              </div>
            </ScrollArea>
          </div>

          {/* Composer Area */}
          <div className="border-t">
            <MessageComposer 
              placeholder="Message..." 
              onSend={(content) => {
                if (conversation && currentUser) {
                  sendDMMessage(conversation.id, content, currentUser.id)
                }
              }} 
            />
          </div>
        </>
      )}
    </div>
  )
}
