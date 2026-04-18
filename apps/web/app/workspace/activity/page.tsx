"use client"

import { AtSign, ThumbsUp, MessageSquare, UserPlus, Mail, CheckCheck } from "lucide-react"
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

const TYPE_ICONS = {
  mention: AtSign,
  reaction: ThumbsUp,
  reply: MessageSquare,
  thread_reply: MessageSquare,
  channel_join: UserPlus,
  dm_message: Mail
}

export default function ActivityPage() {
  const { 
    activities, inboxItems, mentionItems, 
    fetchActivities, fetchInbox, fetchMentions, markAsRead 
  } = useActivityStore()
  const { setCurrentChannelById } = useChannelStore()
  const { conversations, setCurrentConversation } = useDMStore()
  const router = useRouter()
  const [activeTab, setActiveTab] = useState("all")

  useEffect(() => {
    fetchActivities()
    fetchInbox()
    fetchMentions()
  }, [fetchActivities, fetchInbox, fetchMentions])

  const handleItemClick = (item: ActivityItem) => {
    if (!item.isRead) {
      markAsRead([item.id])
    }

    if (item.channel?.id) {
      setCurrentChannelById(item.channel.id)
      router.push(`/workspace?c=${item.channel.id}`)
    } else if (item.message?.dm_id) {
      const conv = conversations.find(c => c.id === item.message.dm_id)
      if (conv) {
        setCurrentConversation(conv)
      }
      router.push(`/workspace/dms?id=${item.message.dm_id}`)
    }
  }

  const handleMarkAllRead = () => {
    let itemsToMark: ActivityItem[] = []
    if (activeTab === "all") itemsToMark = activities
    else if (activeTab === "inbox") itemsToMark = inboxItems
    else if (activeTab === "mentions") itemsToMark = mentionItems

    const unreadIds = itemsToMark.filter(i => !i.isRead).map(i => i.id)
    if (unreadIds.length > 0) {
      markAsRead(unreadIds)
    }
  }

  const renderItemList = (items: ActivityItem[], emptyText: string) => {
    if (items.length === 0) {
      return (
        <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
          <p className="text-sm italic">{emptyText}</p>
        </div>
      )
    }

    return (
      <div className="flex flex-col gap-1 p-4">
        {items.map((item) => {
          const Icon = TYPE_ICONS[item.type as keyof typeof TYPE_ICONS] || MessageSquare
          return (
            <div 
              key={item.id} 
              onClick={() => handleItemClick(item)}
              className={cn(
                "flex items-start gap-4 p-3 rounded-lg cursor-pointer group transition-colors border border-transparent relative",
                item.isRead 
                  ? "hover:bg-muted/30" 
                  : "bg-blue-50/50 dark:bg-blue-900/10 hover:bg-blue-50 dark:hover:bg-blue-900/20 border-blue-100/50 dark:border-blue-900/30"
              )}
            >
              {!item.isRead && (
                <div className="absolute left-1.5 top-1/2 -translate-y-1/2 w-1.5 h-1.5 rounded-full bg-blue-500" />
              )}
              <div className="w-8 h-8 rounded-full bg-muted flex items-center justify-center shrink-0">
                <Icon className={cn("w-4 h-4", !item.isRead ? "text-blue-500" : "text-foreground")} />
              </div>
              <div className="flex-1 min-w-0">
                <p className={cn("text-sm leading-tight", !item.isRead ? "text-foreground font-semibold" : "text-muted-foreground")}>
                  <span className="font-bold text-foreground">{item.user?.name || "Someone"}</span> {item.summary} <span className="font-bold text-blue-500">{item.target}</span>
                </p>
                <span className="text-[10px] text-muted-foreground mt-1 block uppercase tracking-wider font-medium">
                  {formatDistanceToNow(new Date(item.occurredAt), { addSuffix: true })}
                </span>
              </div>
            </div>
          )
        })}
      </div>
    )
  }

  const getUnreadCount = (items: ActivityItem[]) => items.filter(i => !i.isRead).length
  const allUnread = getUnreadCount(activities)
  const inboxUnread = getUnreadCount(inboxItems)
  const mentionsUnread = getUnreadCount(mentionItems)

  return (
    <div className="flex-1 flex flex-col h-full bg-white dark:bg-[#1a1d21]">
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0 bg-white dark:bg-[#1a1d21] z-10">
        <h2 className="font-bold text-lg text-foreground">Activity & Notifications</h2>
        <Button 
          variant="ghost" 
          size="sm" 
          className="text-xs text-muted-foreground hover:text-foreground gap-1.5"
          onClick={handleMarkAllRead}
        >
          <CheckCheck className="w-3.5 h-3.5" />
          Mark all as read
        </Button>
      </header>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="px-4 border-b bg-muted/30">
          <TabsList className="bg-transparent h-12 p-0 gap-6">
            <TabsTrigger value="all" className="relative rounded-none border-b-2 border-transparent data-[state=active]:border-[#1164a3] data-[state=active]:bg-transparent px-0 text-sm font-medium">
              All Activity
              {allUnread > 0 && (
                <span className="ml-1.5 px-1.5 py-0.5 rounded-full bg-blue-500 text-white text-[10px] font-bold">
                  {allUnread}
                </span>
              )}
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

        <ScrollArea className="flex-1">
          <TabsContent value="all" className="m-0 border-none outline-none">
            {renderItemList(activities, "No recent activity found.")}
          </TabsContent>
          <TabsContent value="inbox" className="m-0 border-none outline-none">
            {renderItemList(inboxItems, "Your inbox is empty.")}
          </TabsContent>
          <TabsContent value="mentions" className="m-0 border-none outline-none">
            {renderItemList(mentionItems, "No mentions yet.")}
          </TabsContent>
        </ScrollArea>
      </Tabs>
    </div>
  )
}
