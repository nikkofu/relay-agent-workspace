"use client"

import { useUIStore } from "@/stores/ui-store"
import { X, Hash, Lock } from "lucide-react"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useChannelStore } from "@/stores/channel-store"
import { useMessageStore } from "@/stores/message-store"
import { useUserStore } from "@/stores/user-store"
import { AIThreadSummary } from "@/components/ai-chat/ai-thread-summary"
import { MessageItem } from "@/components/message/message-item"
import { useEffect, useState } from "react"

export function ThreadPanel() {
  const { isThreadOpen, closeThread, activeThreadId } = useUIStore()
  const { currentChannel } = useChannelStore()
  const { currentThreadMessages, fetchThread, sendMessage } = useMessageStore()
  const { currentUser, users } = useUserStore()
  const [replyText, setReplyText] = useState("")

  useEffect(() => {
    if (isThreadOpen && activeThreadId) {
      fetchThread(activeThreadId)
    }
  }, [isThreadOpen, activeThreadId, fetchThread])

  if (!isThreadOpen) return null

  const Icon = currentChannel?.type === "private" ? Lock : Hash

  const handleReply = async () => {
    if (replyText.trim() && activeThreadId && currentChannel && currentUser) {
      await sendMessage(currentChannel.id, replyText, currentUser.id, activeThreadId)
      setReplyText("")
    }
  }

  const parentMessage = currentThreadMessages[0]
  const replies = currentThreadMessages.slice(1)

  return (
    <div className="w-full h-full flex flex-col bg-white dark:bg-[#1a1d21] border-l animate-in slide-in-from-right duration-300">
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0">
        <div className="flex flex-col">
          <h3 className="font-bold text-sm">Thread</h3>
          <div className="flex items-center text-[10px] text-muted-foreground font-medium">
            <Icon className="w-3 h-3 mr-1" />
            {currentChannel?.name}
          </div>
        </div>
        <Button variant="ghost" size="icon" className="h-8 w-8 rounded-full" onClick={closeThread}>
          <X className="w-4 h-4" />
        </Button>
      </header>
      
      <AIThreadSummary />

      <ScrollArea className="flex-1">
        <div className="flex flex-col">
          {parentMessage && (
            <div className="border-b pb-2 mb-2">
              <MessageItem 
                message={parentMessage} 
                sender={users.find(u => u.id === parentMessage.senderId) || {
                  id: parentMessage.senderId,
                  name: "Unknown User",
                  email: "unknown@example.com",
                  avatar: "https://github.com/shadcn.png",
                  status: "offline",
                }}
                showActions={false}
              />
            </div>
          )}
          
          <div className="px-4 py-2 flex items-center gap-2">
            <span className="text-[11px] text-muted-foreground font-bold uppercase tracking-wider">
              {replies.length} {replies.length === 1 ? "Reply" : "Replies"}
            </span>
            <div className="h-[1px] bg-border flex-1" />
          </div>

          <div className="flex flex-col">
            {replies.map((msg) => (
              <MessageItem 
                key={msg.id}
                message={msg} 
                sender={users.find(u => u.id === msg.senderId) || {
                  id: msg.senderId,
                  name: "Unknown User",
                  email: "unknown@example.com",
                  avatar: "https://github.com/shadcn.png",
                  status: "offline",
                }}
              />
            ))}
          </div>
        </div>
      </ScrollArea>

      <div className="p-4 border-t bg-white dark:bg-[#1a1d21]">
        <div className="border rounded-lg p-2 focus-within:ring-1 focus-within:ring-ring transition-all bg-muted/20">
          <textarea 
            placeholder="Reply..." 
            value={replyText}
            onChange={(e) => setReplyText(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault()
                handleReply()
              }
            }}
            className="w-full resize-none bg-transparent border-none focus:ring-0 text-sm h-20 outline-none placeholder:text-muted-foreground/50"
          />
          <div className="flex justify-end mt-1">
            <Button 
              size="sm" 
              className="h-7 text-[11px] bg-[#007a5a] hover:bg-[#007a5a]/90"
              onClick={handleReply}
              disabled={!replyText.trim()}
            >
              Reply
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
