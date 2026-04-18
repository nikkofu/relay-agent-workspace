"use client"

import { usePresenceStore } from "@/stores/presence-store"
import { useUserStore } from "@/stores/user-store"
import { cn } from "@/lib/utils"

interface TypingIndicatorProps {
  scope: string // e.g. "channel:id", "dm:id", "thread:id"
  className?: string
}

export function TypingIndicator({ scope, className }: TypingIndicatorProps) {
  const { typingIndicators } = usePresenceStore()
  const { users, currentUser } = useUserStore()
  
  const activeTyping = typingIndicators[scope] || []
  const otherTyping = activeTyping.filter(t => t.userId !== currentUser?.id)

  if (otherTyping.length === 0) return null

  const names = otherTyping.map(t => {
    const user = users.find(u => u.id === t.userId)
    return user?.name || "Someone"
  })

  let text = ""
  if (names.length === 1) {
    text = `${names[0]} is typing...`
  } else if (names.length === 2) {
    text = `${names[0]} and ${names[1]} are typing...`
  } else {
    text = `${names[0]}, ${names[1]} and ${names.length - 2} others are typing...`
  }

  return (
    <div className={cn("flex items-center gap-2 px-4 py-1 text-[10px] text-muted-foreground italic h-6 animate-in fade-in duration-300", className)}>
      <div className="flex gap-0.5">
        <span className="w-1 h-1 rounded-full bg-muted-foreground animate-bounce [animation-delay:-0.3s]" />
        <span className="w-1 h-1 rounded-full bg-muted-foreground animate-bounce [animation-delay:-0.15s]" />
        <span className="w-1 h-1 rounded-full bg-muted-foreground animate-bounce" />
      </div>
      {text}
    </div>
  )
}
