"use client"

import { Message, User } from "@/types"
import { UserAvatar } from "@/components/common/user-avatar"
import { format } from "date-fns"
import { cn } from "@/lib/utils"
import { MessageActions } from "./message-actions"
import { EmojiReaction } from "./emoji-reaction"
import { useUIStore } from "@/stores/ui-store"
import { useUserStore } from "@/stores/user-store"
import { useMessageStore } from "@/stores/message-store"
import { useState, useEffect } from "react"
import { toast } from "sonner"
import { Pin } from "lucide-react"

interface MessageItemProps {
  message: Message
  sender: User
  isCompact?: boolean
  showActions?: boolean
}

export function MessageItem({ message, sender, isCompact, showActions = true }: MessageItemProps) {
  const { openThread } = useUIStore()
  const { currentUser } = useUserStore()
  const { addReaction, deleteMessage, pinMessage, saveForLater, markAsUnread } = useMessageStore()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  return (
    <div className={cn(
      "group flex items-start gap-3 hover:bg-muted/50 -mx-4 px-4 py-1 transition-colors relative",
      isCompact ? "py-0.5" : "py-2 mt-2",
      message.isPinned && "bg-amber-500/5 hover:bg-amber-500/10"
    )}>
      {message.isPinned && (
        <div className="absolute top-1.5 left-1.5 z-10">
          <Pin className="w-3 h-3 text-amber-500 fill-amber-500 rotate-45" />
        </div>
      )}
      {/* Sender Avatar or Time for Compact Mode */}
      <div className="shrink-0 w-9 h-9 flex items-center justify-center">
        {!isCompact ? (
          <UserAvatar src={sender.avatar} name={sender.name} status={sender.status} user={sender} />
        ) : (
          <span className="text-[10px] text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity">
            {mounted ? format(new Date(message.createdAt), "h:mm") : null}
          </span>
        )}
      </div>

      <div className="flex-1 min-w-0">
        {/* Sender Name and Time (only for non-compact) */}
        {!isCompact && (
          <div className="flex items-center gap-2 leading-tight">
            <span className="font-bold text-sm hover:underline cursor-pointer">{sender.name}</span>
            <span className="text-[11px] text-muted-foreground">
              {mounted ? format(new Date(message.createdAt), "h:mm a") : null}
            </span>
          </div>
        )}

        {/* Message Content */}
        <div 
          className={cn(
            "text-sm mt-0.5 break-words leading-normal",
            isCompact && "mt-0"
          )}
          dangerouslySetInnerHTML={{ __html: message.content }}
        />

        {/* Reactions */}
        {message.reactions && message.reactions.length > 0 && (
          <div className="flex flex-wrap gap-1.5 mt-2">
            {message.reactions.map((reaction, idx) => (
              <EmojiReaction 
                key={idx} 
                emoji={reaction.emoji} 
                count={reaction.count} 
                userIds={reaction.userIds}
                isActive={currentUser ? reaction.userIds.includes(currentUser.id) : false}
              />
            ))}
          </div>
        )}

        {/* Thread Info */}
        {message.replyCount !== undefined && message.replyCount > 0 && (
          <div 
            className="flex items-center gap-2 mt-2 group/thread cursor-pointer hover:bg-muted px-2 py-1 rounded w-fit -ml-2 transition-colors"
            onClick={() => openThread(message.id)}
          >
            <div className="flex -space-x-1">
              <div className="w-5 h-5 rounded border-2 border-background bg-muted overflow-hidden">
                <img src={sender.avatar} alt={sender.name} className="w-full h-full object-cover" />
              </div>
            </div>
            <span className="text-xs font-bold text-blue-500 group-hover/thread:underline">
              {message.replyCount} {message.replyCount === 1 ? "reply" : "replies"}
            </span>
            {message.lastReplyAt && (
              <span className="text-[10px] text-muted-foreground hidden group-hover/thread:inline">
                Last reply {format(new Date(message.lastReplyAt), "h:mm a")}
              </span>
            )}
          </div>
        )}
      </div>

      {/* Message Actions (Floating) */}
      {showActions && (
        <div className="absolute -top-4 right-4 z-10 opacity-0 group-hover:opacity-100 transition-opacity">
          <MessageActions 
            onReply={() => openThread(message.id)} 
            onReact={(emoji) => addReaction(message.id, emoji)}
            onDelete={() => deleteMessage(message.id)}
            onPin={() => pinMessage(message.id)}
            onSave={() => saveForLater(message.id)}
            onMarkUnread={() => markAsUnread(message.id)}
            onCopyLink={() => {
              navigator.clipboard.writeText(`${window.location.origin}/workspace?c=${message.channelId}&m=${message.id}`)
              toast.success("Link copied to clipboard")
            }}
            onForward={() => toast.info("Forwarding coming soon")}
          />
        </div>
      )}
    </div>
  )
}
