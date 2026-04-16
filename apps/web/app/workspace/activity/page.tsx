"use client"

import { AtSign, ThumbsUp, MessageSquare } from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"

export default function ActivityPage() {
  const activities = [
    { icon: AtSign, user: "Nikko Fu", action: "mentioned you in", target: "#engineering", time: "10m ago" },
    { icon: ThumbsUp, user: "John Doe", action: "reacted to your message in", target: "#general", time: "2h ago" },
    { icon: MessageSquare, user: "Jane Smith", action: "replied to your thread", target: "", time: "Yesterday" },
  ]

  return (
    <div className="flex-1 flex flex-col h-full bg-white dark:bg-[#1a1d21]">
      <header className="h-14 px-4 flex items-center border-b shrink-0">
        <h2 className="font-bold text-lg">Activity</h2>
      </header>

      <ScrollArea className="flex-1">
        <div className="p-4 flex flex-col gap-4">
          {activities.map((activity, idx) => (
            <div key={idx} className="flex gap-4 p-3 hover:bg-muted/50 rounded-lg cursor-pointer transition-colors border-l-4 border-transparent hover:border-blue-500">
              <div className="w-8 h-8 rounded bg-muted flex items-center justify-center shrink-0 mt-1">
                <activity.icon className="w-4 h-4" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm leading-snug">
                  <span className="font-bold">{activity.user}</span> {activity.action} <span className="font-bold text-blue-500">{activity.target}</span>
                </p>
                <span className="text-xs text-muted-foreground mt-1 block">{activity.time}</span>
              </div>
            </div>
          ))}
        </div>
      </ScrollArea>
    </div>
  )
}
