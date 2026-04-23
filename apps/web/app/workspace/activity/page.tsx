"use client"

import {
  AtSign, ThumbsUp, MessageSquare, UserPlus, Mail,
  CheckCheck, ListTodo, Terminal, Paperclip, ArrowRight, Activity,
} from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useActivityStore, ActivityItem } from "@/stores/activity-store"
import { useChannelStore } from "@/stores/channel-store"
import { useDMStore } from "@/stores/dm-store"
import { useEffect, useState } from "react"
import { formatDistanceToNow } from "date-fns"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { UnifiedActivityRail } from "@/components/activity/unified-activity-rail"
import { useWorkspaceStore } from "@/stores/workspace-store"

const TYPE_ICONS = {
  mention: AtSign,
  reaction: ThumbsUp,
  reply: MessageSquare,
  thread_reply: MessageSquare,
  channel_join: UserPlus,
  dm_message: Mail,
  list_completed: ListTodo,
  tool_run: Terminal,
  file_uploaded: Paperclip
}

export default function ActivityPage() {
  const {
    activities, inboxItems, mentionItems,
    fetchActivities, fetchInbox, fetchMentions, markAsRead
  } = useActivityStore()
  const { setCurrentChannelById } = useChannelStore()
  const { conversations, setCurrentConversation } = useDMStore()
  const currentWorkspace = useWorkspaceStore(s => s.currentWorkspace)
  const router = useRouter()
  const [activeTab, setActiveTab] = useState("feed")

  useEffect(() => {
    fetchActivities()
    fetchInbox()
    fetchMentions()
  }, [fetchActivities, fetchInbox, fetchMentions])

  const handleItemClick = (item: ActivityItem) => {
    if (!item.isRead) markAsRead([item.id])

    if (item.type === 'tool_run') { router.push('/workspace/workflows'); return }
    if (item.type === 'file_uploaded') { router.push('/workspace/files'); return }

    if (item.channel?.id) {
      setCurrentChannelById(item.channel.id)
      router.push(`/workspace?c=${item.channel.id}`)
    } else if (item.message?.dm_id) {
      const conv = conversations.find(c => c.id === item.message.dm_id)
      if (conv) setCurrentConversation(conv)
      router.push(`/workspace/dms?id=${item.message.dm_id}`)
    }
  }

  const handleMarkAllRead = () => {
    const map: Record<string, ActivityItem[]> = { inbox: inboxItems, mentions: mentionItems }
    const items = map[activeTab] ?? activities
    const ids = items.filter(i => !i.isRead).map(i => i.id)
    if (ids.length > 0) markAsRead(ids)
  }

  const renderItemList = (items: ActivityItem[], emptyText: string) => {
    if (items.length === 0) return (
      <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
        <p className="text-sm italic">{emptyText}</p>
      </div>
    )
    return (
      <div className="flex flex-col gap-1 p-4">
        {items.map((item) => {
          const Icon = TYPE_ICONS[item.type as keyof typeof TYPE_ICONS] || MessageSquare
          const isSpecial = ['list_completed', 'tool_run', 'file_uploaded'].includes(item.type)
          return (
            <div
              key={item.id}
              onClick={() => handleItemClick(item)}
              className={cn(
                "flex items-start gap-4 p-3 rounded-xl cursor-pointer group transition-all border relative",
                item.isRead
                  ? "hover:bg-muted/30 border-transparent"
                  : "bg-blue-50/50 dark:bg-blue-900/10 hover:bg-blue-50 dark:hover:bg-blue-900/20 border-blue-100/50 dark:border-blue-900/30"
              )}
            >
              {!item.isRead && (
                <div className="absolute left-1 top-1/2 -translate-y-1/2 w-1.5 h-6 bg-blue-500 rounded-r-full" />
              )}
              <div className={cn(
                "w-10 h-10 rounded-xl flex items-center justify-center shrink-0",
                item.type === 'tool_run' ? "bg-amber-500/10 text-amber-600" :
                item.type === 'list_completed' ? "bg-purple-500/10 text-purple-600" :
                item.type === 'file_uploaded' ? "bg-blue-500/10 text-blue-600" : "bg-muted text-foreground"
              )}>
                <Icon className="w-5 h-5" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between gap-2">
                  <p className={cn("text-sm leading-tight", !item.isRead ? "text-foreground font-bold" : "text-muted-foreground font-medium")}>
                    <span className="text-foreground">{item.user?.name || "Someone"}</span> {item.summary}
                    {item.target && <span className="text-blue-500 ml-1 font-bold">#{item.target}</span>}
                  </p>
                  {item.type === 'tool_run' && item.message?.status && (
                    <span className={cn(
                      "text-[8px] font-black uppercase px-1.5 py-0.5 rounded-full border",
                      item.message.status === 'success' ? "bg-green-500/10 text-green-600 border-green-500/20" : "bg-amber-500/10 text-amber-600 border-amber-500/20"
                    )}>
                      {item.message.status}
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-2 mt-1.5">
                  <span className="text-[9px] text-muted-foreground uppercase tracking-widest font-black opacity-60">
                    {formatDistanceToNow(new Date(item.occurredAt), { addSuffix: true })}
                  </span>
                  {isSpecial && (
                    <>
                      <span className="text-muted-foreground opacity-30 text-[9px]">•</span>
                      <span className="text-[9px] text-blue-600 dark:text-blue-400 font-bold uppercase tracking-tighter flex items-center gap-1 group-hover:underline">
                        View Context <ArrowRight className="w-2.5 h-2.5" />
                      </span>
                    </>
                  )}
                </div>
              </div>
            </div>
          )
        })}
      </div>
    )
  }

  const getUnreadCount = (items: ActivityItem[]) => items.filter(i => !i.isRead).length
  const inboxUnread = getUnreadCount(inboxItems)
  const mentionsUnread = getUnreadCount(mentionItems)

  return (
    <div className="flex-1 flex flex-col h-full bg-white dark:bg-[#1a1d21]">
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0 bg-white dark:bg-[#1a1d21] z-10">
        <div className="flex items-center gap-2">
          <Activity className="w-5 h-5 text-violet-500" />
          <h2 className="font-bold text-lg text-foreground">Activity</h2>
        </div>
        {(activeTab === 'inbox' || activeTab === 'mentions') && (
          <Button
            variant="ghost"
            size="sm"
            className="text-xs text-muted-foreground hover:text-foreground gap-1.5"
            onClick={handleMarkAllRead}
          >
            <CheckCheck className="w-3.5 h-3.5" />
            Mark all as read
          </Button>
        )}
      </header>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col overflow-hidden">
        <div className="px-4 border-b bg-muted/30 shrink-0">
          <TabsList className="bg-transparent h-12 p-0 gap-6">
            <TabsTrigger value="feed" className="relative rounded-none border-b-2 border-transparent data-[state=active]:border-violet-500 data-[state=active]:bg-transparent px-0 text-sm font-medium">
              Feed
            </TabsTrigger>
            <TabsTrigger value="inbox" className="relative rounded-none border-b-2 border-transparent data-[state=active]:border-[#1164a3] data-[state=active]:bg-transparent px-0 text-sm font-medium">
              Inbox
              {inboxUnread > 0 && (
                <span className="ml-1.5 px-1.5 py-0.5 rounded-full bg-blue-500 text-white text-[10px] font-bold">
                  {inboxUnread}
                </span>
              )}
            </TabsTrigger>
            <TabsTrigger value="mentions" className="relative rounded-none border-b-2 border-transparent data-[state=active]:border-[#1164a3] data-[state=active]:bg-transparent px-0 text-sm font-medium">
              Mentions
              {mentionsUnread > 0 && (
                <span className="ml-1.5 px-1.5 py-0.5 rounded-full bg-blue-500 text-white text-[10px] font-bold">
                  {mentionsUnread}
                </span>
              )}
            </TabsTrigger>
          </TabsList>
        </div>

        <div className="flex-1 overflow-hidden">
          <TabsContent value="feed" className="m-0 border-none outline-none h-full overflow-y-auto">
            <div className="p-4">
              <UnifiedActivityRail
                workspaceId={currentWorkspace?.id}
                defaultTab="ai"
                className="h-full"
              />
            </div>
          </TabsContent>

          <ScrollArea className="h-full">
            <TabsContent value="inbox" className="m-0 border-none outline-none">
              {renderItemList(inboxItems, "Your inbox is empty.")}
            </TabsContent>
            <TabsContent value="mentions" className="m-0 border-none outline-none">
              {renderItemList(mentionItems, "No mentions yet.")}
            </TabsContent>
          </ScrollArea>
        </div>
      </Tabs>
    </div>
  )
}
