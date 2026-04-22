"use client"

import { useEffect, useState } from "react"
import { useDMStore } from "@/stores/dm-store"
import { useUserStore } from "@/stores/user-store"
import { useUIStore } from "@/stores/ui-store"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { MessageSquare, Search, Plus, Sparkles } from "lucide-react"
import { cn } from "@/lib/utils"
import { formatDistanceToNow } from "date-fns"
import { InviteMemberDialog } from "@/components/workspace/invite-member-dialog"

function statusDot(status?: string) {
  if (status === "online") return "bg-green-500"
  if (status === "away") return "bg-amber-500"
  if (status === "busy") return "bg-red-500"
  return "bg-slate-400"
}

export default function DMsPage() {
  const { conversations, fetchConversations } = useDMStore()
  const { users, currentUser } = useUserStore()
  const { openDockedChat } = useUIStore()
  const [q, setQ] = useState("")
  const [showInvite, setShowInvite] = useState(false)

  useEffect(() => {
    fetchConversations()
  }, [fetchConversations])

  const filtered = conversations.filter(conv => {
    const otherUserId = conv.userIds?.find(id => id !== currentUser?.id)
    const user = users.find(u => u.id === otherUserId)
    if (!user) return false
    return !q || user.name.toLowerCase().includes(q.toLowerCase())
  })

  return (
    <div className="flex-1 flex flex-col bg-white dark:bg-[#1a1d21] h-full overflow-hidden">
      {/* Header */}
      <header className="h-14 px-6 flex items-center justify-between border-b shrink-0">
        <div className="flex items-center gap-2">
          <MessageSquare className="w-5 h-5 text-purple-600" />
          <h1 className="text-lg font-black tracking-tight uppercase">Direct Messages</h1>
          {conversations.length > 0 && (
            <Badge variant="secondary" className="text-[10px] font-black">{conversations.length}</Badge>
          )}
        </div>
        <Button
          size="sm"
          className="h-8 gap-1.5 font-bold bg-purple-600 hover:bg-purple-700 text-white"
          onClick={() => setShowInvite(true)}
        >
          <Plus className="w-3.5 h-3.5" />
          New Message
        </Button>
      </header>

      {/* Search */}
      <div className="px-6 py-3 border-b bg-muted/5">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search conversations…"
            className="pl-9 h-8 text-sm bg-white dark:bg-[#1a1d21]"
            value={q}
            onChange={e => setQ(e.target.value)}
          />
        </div>
      </div>

      <ScrollArea className="flex-1">
        <div className="divide-y">
          {filtered.length === 0 ? (
            <div className="py-20 flex flex-col items-center gap-3 text-muted-foreground/60">
              <MessageSquare className="w-10 h-10" />
              <p className="text-sm font-bold">
                {q ? "No conversations match your search." : "No direct messages yet."}
              </p>
              {!q && (
                <Button
                  variant="outline"
                  size="sm"
                  className="mt-2 gap-1.5"
                  onClick={() => setShowInvite(true)}
                >
                  <Plus className="w-3.5 h-3.5" />
                  Start a conversation
                </Button>
              )}
            </div>
          ) : (
            filtered.map(conv => {
              const otherUserId = conv.userIds?.find(id => id !== currentUser?.id)
              const user = users.find(u => u.id === otherUserId)
              if (!user) return null
              const isAI = user.id === "user-2"
              return (
                <button
                  key={conv.id}
                  onClick={() => openDockedChat(user.id)}
                  className="w-full flex items-center gap-4 px-6 py-4 hover:bg-muted/40 transition-colors text-left group"
                >
                  <div className="relative shrink-0">
                    <Avatar className="w-10 h-10">
                      {isAI ? (
                        <div className="w-full h-full rounded-full bg-purple-100 dark:bg-purple-900/20 flex items-center justify-center">
                          <Sparkles className="w-5 h-5 text-purple-600" />
                        </div>
                      ) : (
                        <>
                          <AvatarImage src={user.avatar} />
                          <AvatarFallback className="text-xs font-bold">
                            {user.name.substring(0, 2).toUpperCase()}
                          </AvatarFallback>
                        </>
                      )}
                    </Avatar>
                    <span className={cn(
                      "absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-white dark:border-[#1a1d21]",
                      statusDot(user.status)
                    )} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <span className={cn(
                        "text-sm font-bold truncate",
                        (conv.unreadCount ?? 0) > 0 && "text-foreground"
                      )}>
                        {user.name}
                        {isAI && (
                          <Badge className="ml-1.5 text-[8px] font-black bg-purple-500/10 text-purple-600 border-none h-4 px-1">AI</Badge>
                        )}
                      </span>
                      {conv.lastMessageAt && (
                        <span className="text-[10px] text-muted-foreground shrink-0 ml-2">
                          {formatDistanceToNow(new Date(conv.lastMessageAt), { addSuffix: true })}
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground truncate mt-0.5">
                      {user.statusText || user.title || user.department || "No recent messages"}
                    </p>
                  </div>
                  {(conv.unreadCount ?? 0) > 0 && (
                    <Badge className="bg-purple-600 text-white text-[10px] font-black border-none h-5 px-1.5">
                      {conv.unreadCount}
                    </Badge>
                  )}
                </button>
              )
            })
          )}
        </div>
      </ScrollArea>

      <InviteMemberDialog open={showInvite} onOpenChange={setShowInvite} />
    </div>
  )
}
