"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { cn } from "@/lib/utils"
import {
  Bell, BellOff, BellRing, Newspaper, VolumeX,
  Zap, ChevronDown, ChevronRight, Loader2, Search,
  Tag, User2, Building2, BookOpen, FileText, Layout,
  BellMinus, BellPlus,
} from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import type { FollowNotificationLevel, FollowedEntity } from "@/types"

const KIND_CONFIG: Record<string, { label: string; icon: React.ElementType; color: string; badgeClass: string }> = {
  person:       { label: "Person",       icon: User2,       color: "text-sky-600",    badgeClass: "bg-sky-500/10 text-sky-700 border-sky-300 dark:border-sky-700" },
  project:      { label: "Project",      icon: BookOpen,    color: "text-emerald-600",badgeClass: "bg-emerald-500/10 text-emerald-700 border-emerald-300 dark:border-emerald-700" },
  concept:      { label: "Concept",      icon: Tag,         color: "text-violet-600", badgeClass: "bg-violet-500/10 text-violet-700 border-violet-300 dark:border-violet-700" },
  organization: { label: "Organization", icon: Building2,   color: "text-amber-600",  badgeClass: "bg-amber-500/10 text-amber-700 border-amber-300 dark:border-amber-700" },
  file:         { label: "File",         icon: FileText,    color: "text-rose-600",   badgeClass: "bg-rose-500/10 text-rose-700 border-rose-300 dark:border-rose-700" },
  artifact:     { label: "Artifact",     icon: Layout,      color: "text-orange-600", badgeClass: "bg-orange-500/10 text-orange-700 border-orange-300 dark:border-orange-700" },
}
const getKindCfg = (kind: string) =>
  KIND_CONFIG[kind] ?? { label: kind, icon: Tag, color: "text-muted-foreground", badgeClass: "bg-muted text-muted-foreground border-muted" }

const LEVEL_META: Record<FollowNotificationLevel, { label: string; short: string; icon: React.ElementType; class: string }> = {
  all:         { label: "All alerts",   short: "All",    icon: BellRing,   class: "bg-purple-500/10 text-purple-700 dark:text-purple-400 border-purple-300 dark:border-purple-700" },
  digest_only: { label: "Digest only",  short: "Digest", icon: Newspaper,  class: "bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-300 dark:border-blue-700" },
  silent:      { label: "Silent",       short: "Silent", icon: VolumeX,    class: "bg-muted text-muted-foreground border-muted" },
}

export default function FollowingPage() {
  const router = useRouter()
  const {
    followedEntities, isLoadingFollowed,
    fetchFollowedEntities,
    unfollowEntity,
    updateFollowNotificationLevel,
    bulkUpdateFollowNotificationLevel,
    spikingEntityIds,
    followedStats,
    fetchFollowedStats,
  } = useKnowledgeStore()

  const [q, setQ] = useState("")
  const [pendingIds, setPendingIds] = useState<Record<string, boolean>>({})
  const [mutingAll, setMutingAll] = useState(false)

  useEffect(() => {
    fetchFollowedEntities()
  }, [fetchFollowedEntities])

  // Phase 60: aggregate stats strip
  useEffect(() => {
    fetchFollowedStats()
  }, [fetchFollowedStats, followedEntities.length, spikingEntityIds])

  const filtered = followedEntities.filter(fe =>
    !q || fe.entity.title.toLowerCase().includes(q.toLowerCase()) ||
    fe.entity.kind?.toLowerCase().includes(q.toLowerCase())
  )

  // Spiking first, then alpha
  const sorted = [...filtered].sort((a, b) => {
    const aSpike = !!spikingEntityIds[a.entity.id]
    const bSpike = !!spikingEntityIds[b.entity.id]
    if (aSpike && !bSpike) return -1
    if (!aSpike && bSpike) return 1
    return a.entity.title.localeCompare(b.entity.title)
  })

  const spikingCount = filtered.filter(fe => !!spikingEntityIds[fe.entity.id]).length

  const setPending = (entityId: string, val: boolean) =>
    setPendingIds(p => ({ ...p, [entityId]: val }))

  const handleLevelChange = async (fe: FollowedEntity, level: FollowNotificationLevel) => {
    if (!fe.follow?.id) return
    setPending(fe.entity.id, true)
    await updateFollowNotificationLevel(fe.follow.id, fe.entity.id, level)
    setPending(fe.entity.id, false)
  }

  const handleUnfollow = async (fe: FollowedEntity) => {
    setPending(fe.entity.id, true)
    await unfollowEntity(fe.entity.id)
    setPending(fe.entity.id, false)
  }

  const handleMuteAll = async () => {
    const targetIds = followedEntities
      .filter(fe => fe.follow?.notification_level !== 'silent')
      .map(fe => fe.entity.id)
    if (targetIds.length === 0) return
    setMutingAll(true)
    const ok = await bulkUpdateFollowNotificationLevel(targetIds, 'silent')
    setMutingAll(false)
    if (ok) {
      // toast provided by caller
    }
  }

  const handleUnmuteAll = async () => {
    const targetIds = followedEntities
      .filter(fe => fe.follow?.notification_level === 'silent')
      .map(fe => fe.entity.id)
    if (targetIds.length === 0) return
    setMutingAll(true)
    await bulkUpdateFollowNotificationLevel(targetIds, 'all')
    setMutingAll(false)
  }

  return (
    <div className="flex-1 flex flex-col bg-white dark:bg-[#1a1d21] h-full overflow-hidden">
      {/* Header */}
      <header className="h-14 px-6 flex items-center justify-between border-b shrink-0">
        <div className="flex items-center gap-2">
          <Bell className="w-5 h-5 text-purple-500" />
          <h1 className="text-lg font-black tracking-tight">Following</h1>
          {followedEntities.length > 0 && (
            <Badge className="text-[10px] h-4 px-1.5 bg-purple-500/10 text-purple-700 border-purple-300">
              {followedEntities.length}
            </Badge>
          )}
          {spikingCount > 0 && (
            <Badge className="text-[10px] h-4 px-1.5 gap-1 bg-amber-500/10 text-amber-700 border-amber-300 animate-pulse">
              <Zap className="w-2 h-2" />
              {spikingCount} spiking
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-2">
          {followedEntities.some(fe => fe.follow?.notification_level !== 'silent') && (
            <Button
              variant="outline"
              size="sm"
              className="text-xs gap-1.5 h-7"
              disabled={mutingAll}
              onClick={handleMuteAll}
            >
              {mutingAll ? <Loader2 className="w-3 h-3 animate-spin" /> : <BellMinus className="w-3 h-3" />}
              Mute all
            </Button>
          )}
          {followedEntities.length > 0 &&
            followedEntities.every(fe => fe.follow?.notification_level === 'silent') && (
            <Button
              variant="outline"
              size="sm"
              className="text-xs gap-1.5 h-7 border-purple-400/40 text-purple-700 dark:text-purple-400"
              disabled={mutingAll}
              onClick={handleUnmuteAll}
            >
              {mutingAll ? <Loader2 className="w-3 h-3 animate-spin" /> : <BellPlus className="w-3 h-3" />}
              Restore alerts
            </Button>
          )}
          <Button
            variant="outline"
            size="sm"
            className="text-xs gap-1.5 h-7"
            onClick={() => router.push("/workspace/knowledge")}
          >
            Browse all
            <ChevronRight className="w-3 h-3" />
          </Button>
        </div>
      </header>

      {/* Phase 60: stats strip */}
      {followedStats && followedStats.total_count > 0 && (
        <div className="px-6 pt-3 shrink-0">
          <div className="flex flex-wrap items-center gap-2">
            <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-purple-500/5 border border-purple-500/20">
              <Bell className="w-3 h-3 text-purple-600" />
              <span className="text-[10px] font-bold text-purple-700 dark:text-purple-400">
                {followedStats.total_count} total
              </span>
            </div>
            <div className={cn(
              "flex items-center gap-1.5 px-2.5 py-1 rounded-md border",
              followedStats.spiking_count > 0
                ? "bg-amber-500/10 border-amber-400/40"
                : "bg-muted/30 border-muted"
            )}>
              <Zap className={cn("w-3 h-3", followedStats.spiking_count > 0 ? "text-amber-600" : "text-muted-foreground")} />
              <span className={cn(
                "text-[10px] font-bold",
                followedStats.spiking_count > 0 ? "text-amber-700 dark:text-amber-400" : "text-muted-foreground"
              )}>
                {followedStats.spiking_count} spiking
              </span>
            </div>
            <div className={cn(
              "flex items-center gap-1.5 px-2.5 py-1 rounded-md border",
              followedStats.muted_count > 0
                ? "bg-rose-500/5 border-rose-500/20"
                : "bg-muted/30 border-muted"
            )}>
              <VolumeX className={cn("w-3 h-3", followedStats.muted_count > 0 ? "text-rose-600" : "text-muted-foreground")} />
              <span className={cn(
                "text-[10px] font-bold",
                followedStats.muted_count > 0 ? "text-rose-700 dark:text-rose-400" : "text-muted-foreground"
              )}>
                {followedStats.muted_count} muted
              </span>
            </div>
            {followedStats.by_kind.length > 0 && (
              <div className="h-4 w-px bg-border mx-1" />
            )}
            {followedStats.by_kind.map(bk => {
              const cfg = KIND_CONFIG[bk.kind]
              const Icon = cfg?.icon || Tag
              return (
                <div
                  key={bk.kind}
                  className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-muted/30 border border-muted text-[10px] text-muted-foreground"
                  title={`${bk.count} ${cfg?.label || bk.kind}`}
                >
                  <Icon className={cn("w-3 h-3", cfg?.color || "text-muted-foreground")} />
                  <span className="font-bold">{bk.count}</span>
                  <span>{cfg?.label || bk.kind}</span>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Search */}
      <div className="px-6 py-3 border-b shrink-0">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
          <Input
            value={q}
            onChange={e => setQ(e.target.value)}
            placeholder="Filter followed entities…"
            className="pl-8 h-8 text-xs"
          />
        </div>
      </div>

      {/* Content */}
      <ScrollArea className="flex-1 min-h-0">
        <div className="p-6">
          {isLoadingFollowed ? (
            <div className="flex items-center justify-center py-24 text-muted-foreground">
              <Loader2 className="w-5 h-5 animate-spin mr-2" />
              <span className="text-sm">Loading followed entities…</span>
            </div>
          ) : followedEntities.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-24 text-center gap-3">
              <div className="w-14 h-14 rounded-2xl bg-purple-500/10 flex items-center justify-center">
                <Bell className="w-7 h-7 text-purple-500/60" />
              </div>
              <p className="font-bold text-base">Not following anything yet</p>
              <p className="text-sm text-muted-foreground max-w-xs">
                Follow knowledge entities to get alerted when they spike in activity. Start from the{" "}
                <button
                  onClick={() => router.push("/workspace/knowledge")}
                  className="text-purple-600 underline underline-offset-2"
                >
                  Knowledge wiki
                </button>{" "}
                or by hovering an @mention.
              </p>
            </div>
          ) : sorted.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center gap-2">
              <p className="font-bold text-sm">No results for &ldquo;{q}&rdquo;</p>
              <p className="text-xs text-muted-foreground">Try a different keyword.</p>
            </div>
          ) : (
            <div className="space-y-1.5">
              {/* Spiking section */}
              {spikingCount > 0 && (
                <div className="mb-4">
                  <p className="text-[10px] font-black uppercase tracking-widest text-amber-600 mb-2 flex items-center gap-1">
                    <Zap className="w-3 h-3" /> Spiking now
                  </p>
                  {sorted
                    .filter(fe => !!spikingEntityIds[fe.entity.id])
                    .map(fe => (
                      <FollowRow
                        key={fe.entity.id}
                        fe={fe}
                        isSpiking
                        isPending={!!pendingIds[fe.entity.id]}
                        onLevelChange={handleLevelChange}
                        onUnfollow={handleUnfollow}
                        onNavigate={id => router.push(`/workspace/knowledge/${id}`)}
                      />
                    ))
                  }
                  <div className="border-t my-4" />
                  <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground mb-2">
                    All following
                  </p>
                </div>
              )}

              {sorted
                .filter(fe => !spikingEntityIds[fe.entity.id])
                .map(fe => (
                  <FollowRow
                    key={fe.entity.id}
                    fe={fe}
                    isSpiking={false}
                    isPending={!!pendingIds[fe.entity.id]}
                    onLevelChange={handleLevelChange}
                    onUnfollow={handleUnfollow}
                    onNavigate={id => router.push(`/workspace/knowledge/${id}`)}
                  />
                ))}
            </div>
          )}
        </div>
      </ScrollArea>
    </div>
  )
}

// ── Row component ────────────────────────────────────────────────────────────

interface FollowRowProps {
  fe: FollowedEntity
  isSpiking: boolean
  isPending: boolean
  onLevelChange: (fe: FollowedEntity, level: FollowNotificationLevel) => void
  onUnfollow: (fe: FollowedEntity) => void
  onNavigate: (id: string) => void
}

function FollowRow({ fe, isSpiking, isPending, onLevelChange, onUnfollow, onNavigate }: FollowRowProps) {
  const cfg = getKindCfg(fe.entity.kind || "concept")
  const Icon = cfg.icon
  const level: FollowNotificationLevel = fe.follow?.notification_level ?? 'all'
  const levelMeta = LEVEL_META[level]
  const LevelIcon = levelMeta.icon

  return (
    <div
      className={cn(
        "group flex items-center gap-3 px-3 py-2.5 rounded-lg border transition-all cursor-pointer",
        isSpiking
          ? "border-amber-300 dark:border-amber-700 bg-amber-500/5 hover:bg-amber-500/10"
          : "border-border/50 hover:bg-muted/30"
      )}
      onClick={() => onNavigate(fe.entity.id)}
    >
      {/* Kind icon */}
      <div className={cn(
        "w-8 h-8 rounded-lg flex items-center justify-center shrink-0",
        isSpiking ? "bg-amber-500/10" : "bg-muted/30"
      )}>
        {isSpiking ? (
          <span className="relative flex h-4 w-4">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-60" />
            <Zap className="relative w-4 h-4 text-amber-500" />
          </span>
        ) : (
          <Icon className={cn("w-4 h-4", cfg.color)} />
        )}
      </div>

      {/* Title + meta */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1.5 flex-wrap">
          <p className="font-bold text-sm truncate">{fe.entity.title}</p>
          <Badge className={cn("text-[9px] h-4 px-1.5", cfg.badgeClass)}>{cfg.label}</Badge>
          {isSpiking && (
            <Badge className="text-[9px] h-4 px-1.5 bg-amber-500/10 text-amber-700 border-amber-300 gap-0.5">
              <Zap className="w-2 h-2" /> Spiking
            </Badge>
          )}
        </div>
        {fe.follow?.created_at && (
          <p className="text-[10px] text-muted-foreground mt-0.5">
            Following since {formatDistanceToNow(new Date(fe.follow.created_at), { addSuffix: true })}
          </p>
        )}
      </div>

      {/* Notification level picker */}
      <div className="flex items-center gap-1.5 shrink-0" onClick={e => e.stopPropagation()}>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button
              disabled={isPending}
              className={cn(
                "inline-flex items-center gap-1 text-[10px] font-bold px-1.5 py-0.5 rounded-md border transition-colors",
                levelMeta.class,
                "hover:opacity-80"
              )}
            >
              {isPending ? (
                <Loader2 className="w-2.5 h-2.5 animate-spin" />
              ) : (
                <LevelIcon className="w-2.5 h-2.5" />
              )}
              {levelMeta.short}
              <ChevronDown className="w-2 h-2 opacity-60" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            {(Object.entries(LEVEL_META) as [FollowNotificationLevel, typeof LEVEL_META[FollowNotificationLevel]][]).map(([lvl, meta]) => {
              const LvlIcon = meta.icon
              return (
                <DropdownMenuItem
                  key={lvl}
                  onClick={() => onLevelChange(fe, lvl)}
                  className={cn("gap-2 text-xs", level === lvl && "bg-purple-500/10 text-purple-700 dark:text-purple-400")}
                >
                  <LvlIcon className="w-3 h-3" />
                  <span className="flex-1">{meta.label}</span>
                  {level === lvl && <span className="text-purple-500 text-xs">✓</span>}
                </DropdownMenuItem>
              )
            })}
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => onUnfollow(fe)} className="gap-2 text-xs text-destructive focus:text-destructive">
              <BellOff className="w-3 h-3" />
              Unfollow
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        <ChevronRight className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
      </div>
    </div>
  )
}
