"use client"

import { cn } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/components/ui/hover-card"
import { useUserStore } from "@/stores/user-store"

interface EmojiReactionProps {
  emoji: string
  count: number
  userIds: string[]
  isActive?: boolean
  onToggle?: () => void
}

export function EmojiReaction({ emoji, count, userIds, isActive, onToggle }: EmojiReactionProps) {
  const { users } = useUserStore()
  const reactingUsers = userIds.map(id => users.find(u => u.id === id)?.name || "Unknown user")
  const tooltipText = reactingUsers.join(", ")

  return (
    <HoverCard openDelay={300}>
      <HoverCardTrigger asChild>
        <Badge
          variant="outline"
          className={cn(
            "h-6 px-1.5 gap-1 cursor-pointer hover:border-blue-500 transition-colors",
            isActive ? "bg-blue-500/10 border-blue-500 text-blue-600 dark:text-blue-400" : "bg-muted/50 border-transparent"
          )}
          onClick={onToggle}
        >
          <span className="text-sm leading-none">{emoji}</span>
          <span className="text-[11px] font-bold">{count}</span>
        </Badge>
      </HoverCardTrigger>
      <HoverCardContent side="top" className="w-auto px-2 py-1 text-[11px] font-medium max-w-xs">
        {tooltipText} reacted with {emoji}
      </HoverCardContent>
    </HoverCard>
  )
}
