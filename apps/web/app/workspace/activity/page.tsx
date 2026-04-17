"use client"

import { AtSign, ThumbsUp, MessageSquare, UserPlus } from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useActivityStore } from "@/stores/activity-store"
import { useEffect } from "react"
import { formatDistanceToNow } from "date-fns"

const TYPE_ICONS = {
  mention: AtSign,
  reaction: ThumbsUp,
  reply: MessageSquare,
  channel_join: UserPlus
}

export default function ActivityPage() {
  const { activities, fetchActivities } = useActivityStore()

  useEffect(() => {
    fetchActivities()
  }, [fetchActivities])

  return (
    <div className="flex-1 flex flex-col h-full bg-white dark:bg-[#1a1d21]">
      <header className="h-14 px-4 flex items-center border-b shrink-0 bg-white dark:bg-[#1a1d21] z-10">
        <h2 className="font-bold text-lg text-foreground">Activity</h2>
      </header>

      <ScrollArea className="flex-1">
        <div className="p-4 flex flex-col gap-1">
          {activities.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
              <p className="text-sm italic">No recent activity found.</p>
            </div>
          ) : (
            activities.map((activity) => {
              const Icon = TYPE_ICONS[activity.type] || MessageSquare
              return (
                <div key={activity.id} className="flex items-start gap-4 p-3 hover:bg-muted/50 rounded-lg cursor-pointer group transition-colors border border-transparent hover:border-border/50">
                  <div className="w-8 h-8 rounded-full bg-muted flex items-center justify-center shrink-0">
                    <Icon className="w-4 h-4 text-foreground" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-foreground leading-tight">
                      <span className="font-bold">{activity.user?.name || "Someone"}</span> {activity.summary} <span className="font-bold text-blue-500">{activity.target}</span>
                    </p>
                    <span className="text-[10px] text-muted-foreground mt-1 block uppercase tracking-wider font-medium">
                      {formatDistanceToNow(new Date(activity.occurredAt), { addSuffix: true })}
                    </span>
                  </div>
                </div>
              )
            })
          )}
        </div>
      </ScrollArea>
    </div>
  )
}
