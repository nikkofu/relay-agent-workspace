"use client"

// ── DMConversationList ───────────────────────────────────────────────────────
//
// Left rail of the new WhatsApp/WeChat-style two-pane DMs surface (request
// #1). Renders a searchable list of every direct conversation the current
// user has, with online/away/busy presence dots and an unread badge. The
// active conversation is determined from the current URL (the parent
// `dms/[id]` route segment) so clicking a row navigates while the "data
// linkage" is provided by Next.js routing — no duplicate global state.
//
// AI Assistant entries get a violet sparkle badge so the user can spot
// the AI conversation at a glance.

import { useEffect, useMemo, useState } from "react"
import { useRouter, useParams } from "next/navigation"
import { useDMStore } from "@/stores/dm-store"
import { useUserStore } from "@/stores/user-store"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Input } from "@/components/ui/input"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { MessageSquare, Search, Plus, Sparkles } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { formatDistanceToNow } from "date-fns"
import { InviteMemberDialog } from "@/components/workspace/invite-member-dialog"

function statusDot(status?: string) {
  if (status === "online") return "bg-green-500"
  if (status === "away") return "bg-amber-500"
  if (status === "busy") return "bg-red-500"
  return "bg-slate-400"
}

function isAIUser(user: { id: string; name: string; email?: string }) {
  const name = (user.name || "").toLowerCase()
  const email = (user.email ?? "").toLowerCase()
  return name.includes("assistant") || name.includes("ai")
    || email.startsWith("ai@") || user.id === "user-2"
}

export function DMConversationList() {
  const router = useRouter()
  const params = useParams()
  const activeDmId = (params?.id as string | undefined) ?? null

  const { conversations, fetchConversations } = useDMStore()
  const { users, currentUser } = useUserStore()
  const [q, setQ] = useState("")
  const [showInvite, setShowInvite] = useState(false)

  useEffect(() => { fetchConversations() }, [fetchConversations])

  const enriched = useMemo(() => {
    return conversations
      .map(conv => {
        const otherUserId = conv.userIds?.find(id => id !== currentUser?.id)
        const user = otherUserId ? users.find(u => u.id === otherUserId) : null
        return user ? { conv, user } : null
      })
      .filter((x): x is NonNullable<typeof x> => x !== null)
      .filter(({ user }) => !q || user.name.toLowerCase().includes(q.toLowerCase()))
      .sort((a, b) => {
        const aT = a.conv.lastMessageAt ? new Date(a.conv.lastMessageAt).getTime() : 0
        const bT = b.conv.lastMessageAt ? new Date(b.conv.lastMessageAt).getTime() : 0
        return bT - aT
      })
  }, [conversations, users, currentUser, q])

  return (
    <aside className="w-[320px] shrink-0 flex flex-col h-full border-r bg-muted/5">
      {/* Header */}
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0 bg-white dark:bg-[#1a1d21]">
        <div className="flex items-center gap-2 min-w-0">
          <MessageSquare className="w-5 h-5 text-purple-600 shrink-0" />
          <h1 className="text-sm font-black tracking-tight uppercase truncate">Direct Messages</h1>
          {conversations.length > 0 && (
            <Badge variant="secondary" className="text-[10px] font-black shrink-0">{conversations.length}</Badge>
          )}
        </div>
        <Button
          size="icon"
          variant="ghost"
          className="h-8 w-8 shrink-0"
          onClick={() => setShowInvite(true)}
          title="Start a new conversation"
        >
          <Plus className="w-4 h-4" />
        </Button>
      </header>

      {/* Search */}
      <div className="px-3 py-2 border-b bg-muted/5 shrink-0">
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
          <Input
            placeholder="Search…"
            className="pl-8 h-8 text-xs"
            value={q}
            onChange={e => setQ(e.target.value)}
          />
        </div>
      </div>

      {/* List */}
      <ScrollArea className="flex-1">
        {enriched.length === 0 ? (
          <div className="py-12 px-4 flex flex-col items-center gap-2 text-muted-foreground/60 text-center">
            <MessageSquare className="w-8 h-8" />
            <p className="text-xs font-bold">
              {q ? "No matches." : "No direct messages yet."}
            </p>
            {!q && (
              <Button variant="outline" size="sm" className="mt-2 gap-1.5" onClick={() => setShowInvite(true)}>
                <Plus className="w-3.5 h-3.5" />
                Start a conversation
              </Button>
            )}
          </div>
        ) : (
          <div>
            {enriched.map(({ conv, user }) => {
              const isAI = isAIUser(user)
              const isActive = activeDmId && conv.id === activeDmId
              const unread = conv.unreadCount ?? 0
              return (
                <button
                  key={conv.id}
                  onClick={() => conv.id && router.push(`/workspace/dms/${conv.id}`)}
                  className={cn(
                    "w-full flex items-center gap-3 px-3 py-2.5 text-left transition-colors",
                    "hover:bg-purple-500/5",
                    isActive && "bg-purple-500/10 border-l-2 border-purple-600",
                  )}
                >
                  <div className="relative shrink-0">
                    <Avatar className="w-9 h-9">
                      {isAI ? (
                        <div className="w-full h-full rounded-full bg-violet-100 dark:bg-violet-900/20 flex items-center justify-center">
                          <Sparkles className="w-4 h-4 text-violet-600" />
                        </div>
                      ) : (
                        <>
                          <AvatarImage src={(user as any).avatar} />
                          <AvatarFallback className="text-[10px] font-bold">
                            {user.name.substring(0, 2).toUpperCase()}
                          </AvatarFallback>
                        </>
                      )}
                    </Avatar>
                    <span className={cn(
                      "absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-white dark:border-[#1a1d21]",
                      statusDot(user.status),
                    )} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-2">
                      <span className={cn(
                        "text-xs font-bold truncate",
                        unread > 0 && !isActive && "text-foreground",
                      )}>
                        {user.name}
                        {isAI && (
                          <Badge className="ml-1 text-[7px] font-black bg-violet-500/10 text-violet-600 border-none h-3 px-1 align-middle">AI</Badge>
                        )}
                      </span>
                      {conv.lastMessageAt && (
                        <span className="text-[9px] text-muted-foreground shrink-0">
                          {formatDistanceToNow(new Date(conv.lastMessageAt), { addSuffix: false })}
                        </span>
                      )}
                    </div>
                    <p className="text-[10px] text-muted-foreground truncate mt-0.5">
                      {(user as any).statusText || (user as any).title || (isAI ? "Always available" : "No recent messages")}
                    </p>
                  </div>
                  {unread > 0 && (
                    <Badge className="bg-purple-600 text-white text-[9px] font-black border-none h-4 px-1.5 shrink-0">
                      {unread}
                    </Badge>
                  )}
                </button>
              )
            })}
          </div>
        )}
      </ScrollArea>

      <InviteMemberDialog open={showInvite} onOpenChange={setShowInvite} />
    </aside>
  )
}
