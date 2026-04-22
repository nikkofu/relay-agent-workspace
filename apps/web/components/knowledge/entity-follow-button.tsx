"use client"

import { useEffect, useState } from "react"
import { Bell, BellOff, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { useKnowledgeStore } from "@/stores/knowledge-store"

interface EntityFollowButtonProps {
  entityId: string
  variant?: "chip" | "default"
  className?: string
  onChange?: (isFollowing: boolean) => void
}

/**
 * Phase 55 — reusable Follow/Unfollow toggle for a knowledge entity.
 *
 * - Reads `followedEntityIds` from the knowledge store for instant state
 * - Calls `followEntity` / `unfollowEntity` which optimistically updates
 *   state and refreshes the followed list
 * - "chip" variant = compact pill for hover cards / dense rows
 * - "default" variant = full-sized button for entity detail headers
 */
export function EntityFollowButton({
  entityId,
  variant = "default",
  className,
  onChange,
}: EntityFollowButtonProps) {
  const { followedEntityIds, followEntity, unfollowEntity, fetchFollowedEntities } = useKnowledgeStore()
  const isFollowing = !!followedEntityIds[entityId]
  const [isPending, setIsPending] = useState(false)

  // Lazily ensure we have the followed list at least once per session
  useEffect(() => {
    if (Object.keys(followedEntityIds).length === 0) {
      fetchFollowedEntities().catch(() => { /* silent */ })
    }
  }, [fetchFollowedEntities, followedEntityIds])

  const handleToggle = async (e: React.MouseEvent) => {
    e.stopPropagation()
    e.preventDefault()
    if (isPending) return
    setIsPending(true)
    const ok = isFollowing
      ? await unfollowEntity(entityId)
      : await followEntity(entityId)
    setIsPending(false)
    if (ok) onChange?.(!isFollowing)
  }

  if (variant === "chip") {
    return (
      <button
        onClick={handleToggle}
        disabled={isPending}
        className={cn(
          "inline-flex items-center gap-1 text-[10px] font-bold px-1.5 py-0.5 rounded-md border transition-colors",
          isFollowing
            ? "bg-purple-500/10 text-purple-700 dark:text-purple-400 border-purple-300 dark:border-purple-700 hover:bg-purple-500/20"
            : "bg-muted/50 text-muted-foreground hover:bg-muted hover:text-foreground border-muted",
          isPending && "opacity-60 cursor-wait",
          className,
        )}
      >
        {isPending ? (
          <Loader2 className="w-2.5 h-2.5 animate-spin" />
        ) : isFollowing ? (
          <Bell className="w-2.5 h-2.5" />
        ) : (
          <BellOff className="w-2.5 h-2.5" />
        )}
        {isFollowing ? "Following" : "Follow"}
      </button>
    )
  }

  return (
    <Button
      size="sm"
      variant={isFollowing ? "outline" : "default"}
      disabled={isPending}
      onClick={handleToggle}
      className={cn(
        "h-8 gap-1.5 text-xs font-bold",
        isFollowing
          ? "border-purple-500/40 text-purple-600 dark:text-purple-400 hover:bg-purple-500/10"
          : "bg-purple-600 hover:bg-purple-700 text-white",
        className,
      )}
    >
      {isPending ? (
        <Loader2 className="w-3.5 h-3.5 animate-spin" />
      ) : isFollowing ? (
        <Bell className="w-3.5 h-3.5" />
      ) : (
        <BellOff className="w-3.5 h-3.5" />
      )}
      {isFollowing ? "Following" : "Follow"}
    </Button>
  )
}
