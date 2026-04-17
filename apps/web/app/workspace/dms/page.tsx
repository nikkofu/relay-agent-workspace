"use client"

import { Search, Plus } from "lucide-react"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { UserAvatar } from "@/components/common/user-avatar"
import { useUserStore } from "@/stores/user-store"
import { useDMStore } from "@/stores/dm-store"
import { useMessageStore } from "@/stores/message-store"
import { MessageList } from "@/components/message/message-list"
import { useEffect } from "react"
import { useSearchParams, useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import type { DirectMessage } from "@/types"

import { Suspense } from "react"

function DMsContent() {
  const { users, currentUser } = useUserStore()
  const { conversations, currentConversation, fetchConversations, setCurrentConversation, createConversation } = useDMStore()
  const { messages, fetchDMMessages, sendDMMessage } = useMessageStore()
  const searchParams = useSearchParams()
  const router = useRouter()
  
  const userIdFromUrl = searchParams.get("u")
  const dmIdFromUrl = searchParams.get("id")

  useEffect(() => {
    fetchConversations()
  }, [fetchConversations])

  // Handle URL params
  useEffect(() => {
    if (dmIdFromUrl) {
      const conv = conversations.find(c => c.id === dmIdFromUrl)
      if (conv) setCurrentConversation(conv)
    } else if (userIdFromUrl && currentUser) {
      // Find or create conversation with user
      const existing = conversations.find(c => 
        c.userIds.length === 2 && c.userIds.includes(userIdFromUrl) && c.userIds.includes(currentUser.id)
      )
      if (existing) {
        setCurrentConversation(existing)
        router.replace(`/workspace/dms?id=${existing.id}`)
      } else if (conversations.length > 0) {
        // Only attempt create if we have loaded conversations to avoid duplicates
        createConversation([currentUser.id, userIdFromUrl]).then(conv => {
          if (conv) {
            setCurrentConversation(conv)
            router.replace(`/workspace/dms?id=${conv.id}`)
          }
        })
      }
    }
  }, [dmIdFromUrl, userIdFromUrl, conversations, currentUser, setCurrentConversation, createConversation, router])

  // Fetch messages when conversation changes
  useEffect(() => {
    if (currentConversation) {
      fetchDMMessages(currentConversation.id)
    }
  }, [currentConversation, fetchDMMessages])

  const handleSelectConversation = (conv: DirectMessage) => {
    setCurrentConversation(conv)
    router.push(`/workspace/dms?id=${conv.id}`)
  }

  if (currentConversation) {
    const otherUserId = currentConversation.userIds?.find(id => id !== currentUser?.id)
    const otherUser = users.find(u => u.id === otherUserId)

    return (
      <div className="flex flex-col h-full bg-white dark:bg-[#1a1d21]">
        <header className="h-14 px-4 flex items-center justify-between border-b shrink-0 bg-white dark:bg-[#1a1d21] z-10">
          <div className="flex items-center gap-2">
            <UserAvatar src={otherUser?.avatar} name={otherUser?.name || "User"} status={otherUser?.status} />
            <h2 className="font-bold text-lg text-foreground">{otherUser?.name || "Direct Message"}</h2>
          </div>
        </header>

        <div className="flex-1 overflow-hidden">
          <MessageList messages={messages.filter(m => m.dmId === currentConversation.id)} />
        </div>

        <div className="p-4 border-t bg-white dark:bg-[#1a1d21]">
          <Input 
            placeholder={`Message ${otherUser?.name || "User"}`}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && currentUser) {
                const content = e.currentTarget.value
                if (content.trim()) {
                  sendDMMessage(currentConversation.id, content, currentUser.id)
                  e.currentTarget.value = ""
                }
              }
            }}
          />
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 flex flex-col h-full bg-white dark:bg-[#1a1d21]">
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0 bg-white dark:bg-[#1a1d21] z-10">
        <h2 className="font-bold text-lg text-foreground">Direct Messages</h2>
        <Button variant="ghost" size="icon" className="rounded-full">
          <Plus className="w-4 h-4" />
        </Button>
      </header>
      
      <div className="p-4 border-b">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input placeholder="Search for people" className="pl-9 bg-muted/50 border-none focus-visible:ring-1 text-foreground" />
        </div>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-4 flex flex-col gap-1">
          <h3 className="text-xs font-bold text-muted-foreground uppercase px-2 mb-2">Recent conversations</h3>
          {conversations.map(conv => {
            const otherUserId = conv.userIds?.find(id => id !== currentUser?.id)
            const user = users.find(u => u.id === otherUserId)
            if (!user) return null

            return (
              <div 
                key={conv.id} 
                onClick={() => handleSelectConversation(conv)}
                className="flex items-center gap-3 p-2 hover:bg-muted/50 rounded-lg cursor-pointer group transition-colors"
              >
                <UserAvatar src={user.avatar} name={user.name} status={user.status} className="h-10 w-10" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-bold truncate text-foreground">{user.name}</span>
                    <span className="text-[10px] text-muted-foreground">Just now</span>
                  </div>
                  <p className="text-xs text-muted-foreground truncate leading-tight">
                    {user.statusText || "Online"}
                  </p>
                </div>
              </div>
            )
          })}
        </div>
      </ScrollArea>
    </div>
  )
}

export default function DMsPage() {
  return (
    <Suspense fallback={<div className="flex-1 flex items-center justify-center bg-white dark:bg-[#1a1d21]">Loading conversations...</div>}>
      <DMsContent />
    </Suspense>
  )
}
