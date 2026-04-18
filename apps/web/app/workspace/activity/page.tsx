"use client"

import { AtSign, ThumbsUp, MessageSquare, UserPlus, Inbox, Bell, Mail } from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useActivityStore, ActivityItem } from "@/stores/activity-store"
import { useChannelStore } from "@/stores/channel-store"
import { useDMStore } from "@/stores/dm-store"
import { useEffect } from "react"
import { formatDistanceToNow } from "date-fns"
import { useRouter } from "next/navigation"

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
    fetchActivities, fetchInbox, fetchMentions 
  } = useActivityStore()
  const { setCurrentChannelById } = useChannelStore()
  const { conversations, setCurrentConversation } = useDMStore()
  const router = useRouter()

  useEffect(() => {
    fetchActivities()
    fetchInbox()
    fetchMentions()
  }, [fetchActivities, fetchInbox, fetchMentions])

  const handleItemClick = (item: ActivityItem) => {
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
              className="flex items-start gap-4 p-3 hover:bg-muted/50 rounded-lg cursor-pointer group transition-colors border border-transparent hover:border-border/50"
            >
              <div className="w-8 h-8 rounded-full bg-muted flex items-center justify-center shrink-0">
                <Icon className="w-4 h-4 text-foreground" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm text-foreground leading-tight">
                  <span className="font-bold">{item.user?.name || "Someone"}</span> {item.summary} <span className="font-bold text-blue-500">{item.target}</span>
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

  return (
    <div className="flex-1 flex flex-col h-full bg-white dark:bg-[#1a1d21]">
      <header className="h-14 px-4 flex flex-col justify-center border-b shrink-0 bg-white dark:bg-[#1a1d21] z-10">
        <h2 className="font-bold text-lg text-foreground">Activity & Notifications</h2>
      </header>

      <Tabs defaultValue="all" className="flex-1 flex flex-col">
        <div className="px-4 border-b bg-muted/30">
          <TabsList className="bg-transparent h-12 p-0 gap-6">
            <TabsTrigger value="all" className="rounded-none border-b-2 border-transparent data-[state=active]:border-[#1164a3] data-[state=active]:bg-transparent px-0 text-sm font-medium">All Activity</TabsTrigger>
            <TabsTrigger value="inbox" className="rounded-none border-b-2 border-transparent data-[state=active]:border-[#1164a3] data-[state=active]:bg-transparent px-0 text-sm font-medium">Inbox</TabsTrigger>
            <TabsTrigger value="mentions" className="rounded-none border-b-2 border-transparent data-[state=active]:border-[#1164a3] data-[state=active]:bg-transparent px-0 text-sm font-medium">Mentions</TabsTrigger>
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
