"use client"

import { useEffect, useState } from "react"
import { Bell, BellOff, BellRing, Loader2, ChevronDown, Newspaper, VolumeX, Zap } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { cn } from "@/lib/utils"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import type { FollowNotificationLevel } from "@/types"

interface EntityFollowButtonProps {
  entityId: string
  variant?: "chip" | "default"
  className?: string
  isSpiking?: boolean
  onChange?: (isFollowing: boolean) => void
}

const LEVEL_LABELS: Record<FollowNotificationLevel, { label: string; icon: React.ReactNode; desc: string }> = {
  all:        { label: "All alerts",    icon: <BellRing className="w-3 h-3" />,   desc: "Real-time spike alerts + digests" },
  digest_only:{ label: "Digest only",  icon: <Newspaper className="w-3 h-3" />,  desc: "Mentioned in digest summaries" },
  silent:     { label: "Silent",        icon: <VolumeX className="w-3 h-3" />,    desc: "Track activity, no notifications" },
}

/**
 * Phase 55/57 — reusable Follow/Unfollow toggle for a knowledge entity.
 * When following, a ChevronDown opens a notification level picker and unfollow option.
 * "chip" variant = compact pill; "default" variant = full-sized button for detail headers.
 * `isSpiking` adds a purple pulse indicator when the entity is spiking.
 */
export function EntityFollowButton({
  entityId,
  variant = "default",
  className,
  isSpiking = false,
  onChange,
}: EntityFollowButtonProps) {
  const {
    followedEntityIds, followedEntities,
    followEntity, unfollowEntity,
    updateFollowNotificationLevel,
    fetchFollowedEntities,
  } = useKnowledgeStore()

  const isFollowing = !!followedEntityIds[entityId]
  const followedEntity = followedEntities.find(f => f.entity.id === entityId)
  const currentLevel: FollowNotificationLevel = followedEntity?.follow?.notification_level ?? 'all'
  const [isPending, setIsPending] = useState(false)
  const [isLevelPending, setIsLevelPending] = useState(false)

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

  const handleLevelChange = async (level: FollowNotificationLevel) => {
    if (!followedEntity?.follow?.id || isLevelPending) return
    setIsLevelPending(true)
    await updateFollowNotificationLevel(followedEntity.follow.id, entityId, level)
    setIsLevelPending(false)
  }

  const levelInfo = LEVEL_LABELS[currentLevel]

  // ── chip variant ──────────────────────────────────────────────────────────
  if (variant === "chip") {
    if (!isFollowing) {
      return (
        <button
          onClick={handleToggle}
          disabled={isPending}
          className={cn(
            "inline-flex items-center gap-1 text-[10px] font-bold px-1.5 py-0.5 rounded-md border transition-colors",
            "bg-muted/50 text-muted-foreground hover:bg-muted hover:text-foreground border-muted",
            isPending && "opacity-60 cursor-wait",
            className,
          )}
        >
          {isPending ? <Loader2 className="w-2.5 h-2.5 animate-spin" /> : <BellOff className="w-2.5 h-2.5" />}
          Follow
        </button>
      )
    }

    return (
      <div className={cn("inline-flex items-center rounded-md border border-purple-300 dark:border-purple-700 overflow-hidden", className)}>
        <button
          onClick={handleToggle}
          disabled={isPending}
          className={cn(
            "inline-flex items-center gap-1 text-[10px] font-bold px-1.5 py-0.5 transition-colors",
            "bg-purple-500/10 text-purple-700 dark:text-purple-400 hover:bg-purple-500/20",
            isPending && "opacity-60 cursor-wait",
          )}
        >
          {isPending ? (
            <Loader2 className="w-2.5 h-2.5 animate-spin" />
          ) : isSpiking ? (
            <span className="relative flex h-2.5 w-2.5 mr-0.5">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-purple-400 opacity-75" />
              <Bell className="relative w-2.5 h-2.5" />
            </span>
          ) : (
            <Bell className="w-2.5 h-2.5" />
          )}
          Following
        </button>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button
              disabled={isLevelPending}
              className="px-1 py-0.5 border-l border-purple-300 dark:border-purple-700 bg-purple-500/10 text-purple-700 dark:text-purple-400 hover:bg-purple-500/20 transition-colors"
            >
              {isLevelPending ? <Loader2 className="w-2 h-2 animate-spin" /> : <ChevronDown className="w-2 h-2" />}
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48 text-xs">
            {(Object.entries(LEVEL_LABELS) as [FollowNotificationLevel, typeof LEVEL_LABELS[FollowNotificationLevel]][]).map(([level, info]) => (
              <DropdownMenuItem
                key={level}
                onClick={() => handleLevelChange(level)}
                className={cn("gap-2 text-xs", currentLevel === level && "bg-purple-500/10 text-purple-700")}
              >
                {info.icon}
                <span className="flex-1">{info.label}</span>
                {currentLevel === level && <span className="text-purple-500">✓</span>}
              </DropdownMenuItem>
            ))}
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleToggle} className="gap-2 text-xs text-destructive focus:text-destructive">
              <BellOff className="w-3 h-3" />
              Unfollow
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    )
  }

  // ── default variant ───────────────────────────────────────────────────────
  if (!isFollowing) {
    return (
      <Button
        size="sm"
        variant="default"
        disabled={isPending}
        onClick={handleToggle}
        className={cn("h-8 gap-1.5 text-xs font-bold bg-purple-600 hover:bg-purple-700 text-white", className)}
      >
        {isPending ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <BellOff className="w-3.5 h-3.5" />}
        Follow
      </Button>
    )
  }

  return (
    <div className={cn("inline-flex items-center rounded-md border border-purple-500/40 overflow-hidden", className)}>
      <Button
        size="sm"
        variant="ghost"
        disabled={isPending}
        onClick={handleToggle}
        className="h-8 gap-1.5 text-xs font-bold text-purple-600 dark:text-purple-400 hover:bg-purple-500/10 rounded-none border-0 px-3"
      >
        {isPending ? (
          <Loader2 className="w-3.5 h-3.5 animate-spin" />
        ) : isSpiking ? (
          <span className="relative flex h-3.5 w-3.5">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-purple-400 opacity-75" />
            <Zap className="relative w-3.5 h-3.5 text-purple-500" />
          </span>
        ) : (
          levelInfo.icon
        )}
        {isSpiking ? "Spiking" : "Following"}
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            size="sm"
            variant="ghost"
            disabled={isLevelPending}
            className="h-8 w-7 px-0 rounded-none border-0 border-l border-purple-500/40 text-purple-600 dark:text-purple-400 hover:bg-purple-500/10"
          >
            {isLevelPending ? <Loader2 className="w-3 h-3 animate-spin" /> : <ChevronDown className="w-3 h-3" />}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-52">
          <div className="px-2 py-1.5 text-[10px] font-black uppercase tracking-widest text-muted-foreground">
            Alert level
          </div>
          {(Object.entries(LEVEL_LABELS) as [FollowNotificationLevel, typeof LEVEL_LABELS[FollowNotificationLevel]][]).map(([level, info]) => (
            <DropdownMenuItem
              key={level}
              onClick={() => handleLevelChange(level)}
              className={cn("gap-2 text-xs", currentLevel === level && "bg-purple-500/10 text-purple-700 dark:text-purple-400")}
            >
              {info.icon}
              <div className="flex-1 min-w-0">
                <div className="font-bold leading-tight">{info.label}</div>
                <div className="text-[10px] text-muted-foreground truncate">{info.desc}</div>
              </div>
              {currentLevel === level && <span className="text-purple-500 text-xs">✓</span>}
            </DropdownMenuItem>
          ))}
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={handleToggle} className="gap-2 text-xs text-destructive focus:text-destructive">
            <BellOff className="w-3.5 h-3.5" />
            Unfollow entity
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
