"use client"

import { useEffect, useRef } from "react"
import { useParams, useRouter } from "next/navigation"
import { useMessageStore } from "@/stores/message-store"
import { useDMStore } from "@/stores/dm-store"
import { useUserStore } from "@/stores/user-store"
import { MessageComposer } from "@/components/message/message-composer"
import { TypingIndicator } from "@/components/message/typing-indicator"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { ArrowLeft, Sparkles } from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import { cn } from "@/lib/utils"
import { UserAvatar } from "@/components/common/user-avatar"

function isAIUserCheck(user: { id: string; name: string; email?: string }) {
  const name = user.name.toLowerCase()
  const email = (user.email ?? "").toLowerCase()
  return name.includes("assistant") || name.includes("ai") || email.startsWith("ai@") || user.id === "user-2"
}

export default function DMConversationPage() {
  const params = useParams()
  const router = useRouter()
  const dmId = params.id as string

  const { messages, fetchDMMessages, sendDMMessage, streamingDMMessages } = useMessageStore()
  const { conversations, fetchConversations } = useDMStore()
  const { users, currentUser } = useUserStore()
  const scrollRef = useRef<HTMLDivElement>(null)

  const conversation = conversations.find(c => c.id === dmId)
  const otherUserId = conversation?.userIds?.find((id: string) => id !== currentUser?.id)
  const otherUser = otherUserId ? users.find(u => u.id === otherUserId) : null
  const isAI = otherUser ? isAIUserCheck(otherUser) : false

  const dmMessages = messages.filter(m => m.dmId === dmId)
  const streamEntries = Object.entries(streamingDMMessages).filter(([, v]) => v.dmId === dmId)

  useEffect(() => {
    if (!conversations.length) fetchConversations()
  }, [conversations.length, fetchConversations])

  useEffect(() => {
    if (dmId) fetchDMMessages(dmId)
  }, [dmId, fetchDMMessages])

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [dmMessages.length, streamEntries.length])

  return (
    <div className="flex flex-col h-full bg-white dark:bg-[#1a1d21]">
      {/* Header */}
      <header className="h-14 px-4 flex items-center gap-3 border-b shrink-0 bg-white dark:bg-[#1a1d21]">
        <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0" onClick={() => router.push('/workspace/dms')}>
          <ArrowLeft className="w-4 h-4" />
        </Button>
        {otherUser ? (
          <>
            <div className={cn(
              "w-9 h-9 rounded-full flex items-center justify-center shrink-0",
              isAI ? "bg-violet-100 dark:bg-violet-900/20" : "bg-muted"
            )}>
              {isAI ? (
                <Sparkles className="w-4.5 h-4.5 text-violet-600" />
              ) : (
                <Avatar className="w-9 h-9">
                  <AvatarImage src={(otherUser as any).avatar} />
                  <AvatarFallback className="text-xs font-bold">{otherUser.name.substring(0, 2).toUpperCase()}</AvatarFallback>
                </Avatar>
              )}
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-1.5">
                <span className="font-bold text-sm truncate">{otherUser.name}</span>
                {isAI && (
                  <Badge className="text-[8px] font-black bg-violet-500/10 text-violet-600 border-none h-4 px-1">AI</Badge>
                )}
              </div>
              <p className="text-[10px] text-muted-foreground leading-none mt-0.5">
                {isAI ? "AI Assistant · Always available" : (otherUser as any).title || "Direct message"}
              </p>
            </div>
          </>
        ) : (
          <span className="font-bold text-sm text-muted-foreground">Loading…</span>
        )}
      </header>

      {/* Messages */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-4 space-y-3">
        {isAI && dmMessages.length === 0 && streamEntries.length === 0 && (
          <div className="flex flex-col items-center gap-4 py-16 text-center">
            <div className="w-16 h-16 rounded-2xl bg-violet-100 dark:bg-violet-900/20 flex items-center justify-center">
              <Sparkles className="w-8 h-8 text-violet-600" />
            </div>
            <div className="space-y-1">
              <h3 className="font-black text-base tracking-tight">Chat with {otherUser?.name ?? "AI Assistant"}</h3>
              <p className="text-sm text-muted-foreground max-w-xs leading-relaxed">
                Ask anything — the AI thinks through your question and responds in real time, showing its reasoning as it goes.
              </p>
            </div>
          </div>
        )}

        {dmMessages.map(msg => {
          const sender = users.find(u => u.id === msg.senderId)
          const senderIsAI = sender ? isAIUserCheck(sender) : false
          const isOwn = msg.senderId === currentUser?.id
          return (
            <div key={msg.id} className={cn("flex gap-2.5 items-end", isOwn && "flex-row-reverse")}>
              {senderIsAI ? (
                <div className="w-7 h-7 rounded-full bg-violet-100 dark:bg-violet-900/20 flex items-center justify-center shrink-0">
                  <Sparkles className="w-3.5 h-3.5 text-violet-600" />
                </div>
              ) : (
                <UserAvatar
                  name={sender?.name ?? "?"}
                  src={(sender as any)?.avatar}
                  className="w-7 h-7 shrink-0"
                />
              )}
              <div className={cn(
                "max-w-[72%] space-y-0.5",
                isOwn && "items-end flex flex-col"
              )}>
                {!isOwn && (
                  <span className="text-[10px] font-semibold text-muted-foreground px-1">
                    {sender?.name ?? "Unknown"}
                  </span>
                )}
                <div className={cn(
                  "px-3.5 py-2 rounded-2xl text-sm leading-relaxed",
                  isOwn
                    ? "bg-violet-600 text-white rounded-br-sm"
                    : "bg-muted dark:bg-muted/60 text-foreground rounded-bl-sm"
                )}>
                  <div dangerouslySetInnerHTML={{ __html: msg.content }} />
                </div>
                <span className={cn(
                  "text-[9px] text-muted-foreground px-1",
                  isOwn && "text-right"
                )}>
                  {msg.createdAt ? formatDistanceToNow(new Date(msg.createdAt), { addSuffix: true }) : ""}
                </span>
              </div>
            </div>
          )
        })}

        {/* Streaming AI response — live chunks before the final message arrives */}
        {streamEntries.map(([tempId, s]) => (
          <div key={tempId} className="flex gap-2.5 items-end">
            <div className="w-7 h-7 rounded-full bg-violet-100 dark:bg-violet-900/20 flex items-center justify-center shrink-0 animate-pulse">
              <Sparkles className="w-3.5 h-3.5 text-violet-600" />
            </div>
            <div className="max-w-[72%] space-y-0.5">
              <span className="text-[10px] font-semibold text-muted-foreground px-1">
                {otherUser?.name ?? "AI Assistant"}
              </span>
              <div className="bg-muted dark:bg-muted/60 px-3.5 py-2 rounded-2xl rounded-bl-sm text-sm leading-relaxed text-foreground">
                <span>{s.text}</span>
                <span className="inline-block w-0.5 h-4 bg-violet-500 ml-0.5 animate-[blink_0.9s_step-end_infinite]" />
              </div>
            </div>
          </div>
        ))}

        {/* Typing indicator (shows "AI Assistant is typing..." from backend typing.updated events) */}
        <TypingIndicator scope={`dm:${dmId}`} />
      </div>

      {/* Composer */}
      <div className="border-t bg-white dark:bg-[#1a1d21] shrink-0">
        <MessageComposer
          placeholder={isAI ? `Message ${otherUser?.name ?? "AI Assistant"}…` : "Message…"}
          scope={`dm:${dmId}`}
          onSend={(content) => {
            if (currentUser && dmId) {
              sendDMMessage(dmId, content, currentUser.id)
            }
          }}
        />
      </div>
    </div>
  )
}
