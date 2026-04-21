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
import { useArtifactStore } from "@/stores/artifact-store"
import { useState, useEffect } from "react"
import { toast } from "sonner"
import { Pin, FileCode, FileText, MoreVertical, Copy } from "lucide-react"
import { FileAttachmentCard } from "./file-attachment-card"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import Image from "next/image"

interface MessageItemProps {
  message: Message
  sender: User
  isCompact?: boolean
  showActions?: boolean
}

export function MessageItem({ message, sender, isCompact, showActions = true }: MessageItemProps) {
  const { openThread, openCanvas } = useUIStore()
  const { currentUser } = useUserStore()
  const { addReaction, deleteMessage, pinMessage, saveForLater, markAsUnread } = useMessageStore()
  const { duplicateArtifact } = useArtifactStore()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  const safeFormat = (dateStr: string, formatStr: string) => {
    if (!dateStr) return ""
    const date = new Date(dateStr)
    if (isNaN(date.getTime())) return ""
    return format(date, formatStr)
  }

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
            {mounted ? safeFormat(message.createdAt, "h:mm") : null}
          </span>
        )}
      </div>

      <div className="flex-1 min-w-0">
        {/* Sender Name and Time (only for non-compact) */}
        {!isCompact && (
          <div className="flex items-center gap-2 leading-tight">
            <span className="font-bold text-sm hover:underline cursor-pointer">{sender.name}</span>
            <span className="text-[11px] text-muted-foreground">
              {mounted ? safeFormat(message.createdAt, "h:mm a") : null}
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

        {/* Attachments */}
        {message.attachments && message.attachments.length > 0 && (
          <div className="flex flex-wrap gap-3 mt-2">
            {message.attachments.map((attachment) => (
              attachment.type === 'file' || attachment.kind === 'file' ? (
                <FileAttachmentCard
                  key={attachment.id}
                  attachment={attachment}
                  messageId={message.id}
                />
              ) : (
                <div
                  key={attachment.id}
                  className="flex items-center gap-3 p-2 rounded-lg border bg-muted/30 hover:bg-muted/50 transition-all max-w-[300px] group/attachment"
                >
                  <div className="w-8 h-8 rounded bg-blue-500/10 flex items-center justify-center shrink-0 text-blue-600">
                    {attachment.type === 'artifact' ? (
                      <FileCode className="w-4 h-4 text-purple-600" />
                    ) : (
                      <FileText className="w-4 h-4" />
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="text-[11px] font-bold truncate leading-tight">{attachment.name}</p>
                    <p className="text-[9px] text-muted-foreground uppercase font-bold tracking-tighter">
                      {attachment.type} {attachment.size ? `• ${(attachment.size / 1024).toFixed(1)} KB` : ''}
                    </p>
                  </div>
                  <div className="flex items-center gap-1 opacity-0 group-hover/attachment:opacity-100 transition-opacity">
                    {attachment.type === 'artifact' ? (
                      <>
                        <Button variant="ghost" size="icon" className="h-6 w-6" onClick={() => openCanvas(attachment.id)}>
                          <FileCode className="w-3 h-3" />
                        </Button>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-6 w-6">
                              <MoreVertical className="w-3 h-3" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => duplicateArtifact(attachment.id)}>
                              <Copy className="w-3.5 h-3.5 mr-2" />
                              Duplicate
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </>
                    ) : (
                      <Button variant="ghost" size="icon" className="h-6 w-6" asChild>
                        <a href={attachment.url} target="_blank" rel="noopener noreferrer" download={attachment.name}>
                          <FileText className="w-3 h-3" />
                        </a>
                      </Button>
                    )}
                  </div>
                </div>
              )
            ))}
          </div>
        )}

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
              <div className="w-5 h-5 rounded border-2 border-background bg-muted overflow-hidden relative">
                {sender.avatar ? (
                  <Image src={sender.avatar} alt={sender.name} fill className="object-cover" />
                ) : (
                  <div className="w-full h-full bg-muted flex items-center justify-center text-[8px] font-bold">
                    {sender.name.substring(0, 1).toUpperCase()}
                  </div>
                )}
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
