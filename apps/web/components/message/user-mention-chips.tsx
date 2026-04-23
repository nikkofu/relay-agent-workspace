"use client"

import { AtSign } from "lucide-react"
import { useRouter } from "next/navigation"
import { useDMStore } from "@/stores/dm-store"
import { useUserStore } from "@/stores/user-store"
import type { MessageUserMention } from "@/types"
import { cn } from "@/lib/utils"

interface UserMentionChipsProps {
  mentions: MessageUserMention[]
  className?: string
}

export function UserMentionChips({ mentions, className }: UserMentionChipsProps) {
  const router = useRouter()
  const { conversations } = useDMStore()
  const { currentUser } = useUserStore()

  if (!mentions || mentions.length === 0) return null

  const handleMentionClick = (userId: string) => {
    const conv = conversations.find(c =>
      c.userIds?.includes(userId) && c.userIds?.includes(currentUser?.id ?? '')
    )
    if (conv?.id) {
      router.push(`/workspace/dms/${conv.id}`)
    }
  }

  return (
    <div className={cn("flex flex-wrap gap-1 mt-1", className)}>
      {mentions.map((mention, i) => (
        <button
          key={`${mention.user_id}-${i}`}
          type="button"
          onClick={() => handleMentionClick(mention.user_id)}
          className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] font-semibold bg-fuchsia-500/10 text-fuchsia-700 dark:text-fuchsia-300 border border-fuchsia-400/30 hover:bg-fuchsia-500/20 transition-colors"
        >
          <AtSign className="w-2.5 h-2.5 shrink-0" />
          {mention.name}
        </button>
      ))}
    </div>
  )
}
