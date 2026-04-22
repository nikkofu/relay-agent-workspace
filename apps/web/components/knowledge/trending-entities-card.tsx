"use client"

import { useEffect } from "react"
import { useRouter } from "next/navigation"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import {
  TrendingUp, TrendingDown, Flame, Loader2, ChevronRight, Tag,
  User2, BookOpen, Building2, FileText, Layout, Minus,
} from "lucide-react"
import { formatDistanceToNow } from "date-fns"

const KIND_ICONS: Record<string, React.ElementType> = {
  person: User2, project: BookOpen, concept: Tag,
  organization: Building2, file: FileText, artifact: Layout,
}

interface TrendingEntitiesCardProps {
  workspaceId: string
  days?: number
  limit?: number
  compact?: boolean
  className?: string
}

export function TrendingEntitiesCard({
  workspaceId,
  days = 7,
  limit = 5,
  compact = false,
  className,
}: TrendingEntitiesCardProps) {
  const router = useRouter()
  const { trendingEntities, isLoadingTrending, fetchTrendingEntities } = useKnowledgeStore()

  useEffect(() => {
    if (!workspaceId) return
    fetchTrendingEntities(workspaceId, days, limit)
  }, [workspaceId, days, limit, fetchTrendingEntities])

  if (isLoadingTrending && trendingEntities.length === 0) {
    return (
      <div className={cn("flex items-center justify-center py-8 text-muted-foreground", className)}>
        <Loader2 className="w-4 h-4 animate-spin mr-2" />
        <span className="text-xs">Loading trending…</span>
      </div>
    )
  }

  if (trendingEntities.length === 0) {
    return (
      <div className={cn(
        "flex items-center gap-2 p-4 rounded-lg border border-dashed text-xs text-muted-foreground",
        className
      )}>
        <Flame className="w-3.5 h-3.5" />
        <span>No trending entities in the last {days} days.</span>
      </div>
    )
  }

  return (
    <div className={cn("rounded-xl border bg-gradient-to-br from-orange-500/5 to-amber-500/5 dark:from-orange-950/20 dark:to-amber-950/20 overflow-hidden", className)}>
      <div className="flex items-center gap-2 px-4 py-2.5 border-b border-orange-500/10">
        <div className="w-6 h-6 rounded-md bg-orange-500/15 flex items-center justify-center">
          <Flame className="w-3.5 h-3.5 text-orange-600" />
        </div>
        <p className="text-xs font-black tracking-tight uppercase">Trending in Knowledge</p>
        <Badge className="text-[9px] h-4 px-1.5 bg-orange-500/10 text-orange-700 border-orange-300 ml-auto">
          Last {days}d
        </Badge>
      </div>
      <ul className="divide-y divide-orange-500/10">
        {trendingEntities.map((item, idx) => {
          const Icon = KIND_ICONS[item.entity.kind] || Tag
          const delta = item.velocity_delta
          const TrendIcon = delta > 0 ? TrendingUp : delta < 0 ? TrendingDown : Minus
          const trendColor =
            delta > 0 ? "text-emerald-600 bg-emerald-500/10 border-emerald-300" :
            delta < 0 ? "text-rose-600 bg-rose-500/10 border-rose-300" :
            "text-muted-foreground bg-muted border-muted"
          return (
            <li
              key={item.entity.id}
              onClick={() => router.push(`/workspace/knowledge/${item.entity.id}`)}
              className="flex items-center gap-3 px-4 py-2.5 hover:bg-orange-500/5 cursor-pointer transition-colors group"
            >
              <span className="w-5 text-[10px] font-black text-muted-foreground shrink-0">#{idx + 1}</span>
              <div className={cn(
                "w-7 h-7 rounded-md flex items-center justify-center shrink-0",
                idx === 0 ? "bg-orange-500/15" : "bg-muted/30"
              )}>
                <Icon className={cn(
                  "w-3.5 h-3.5",
                  idx === 0 ? "text-orange-600" : "text-muted-foreground"
                )} />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-bold truncate">{item.entity.title}</p>
                {!compact && (
                  <div className="flex items-center gap-2 text-[10px] text-muted-foreground mt-0.5">
                    <span>{item.recent_ref_count} refs</span>
                    {item.related_channel_ids.length > 0 && (
                      <>
                        <span>·</span>
                        <span>{item.related_channel_ids.length} channel{item.related_channel_ids.length === 1 ? "" : "s"}</span>
                      </>
                    )}
                    {item.last_ref_at && (
                      <>
                        <span>·</span>
                        <span>{formatDistanceToNow(new Date(item.last_ref_at), { addSuffix: true })}</span>
                      </>
                    )}
                  </div>
                )}
              </div>
              <Badge className={cn("text-[10px] font-black h-5 px-1.5 gap-0.5 shrink-0", trendColor)}>
                <TrendIcon className="w-2.5 h-2.5" />
                {delta > 0 ? `+${delta}` : delta}
              </Badge>
              <ChevronRight className="w-3.5 h-3.5 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity shrink-0" />
            </li>
          )
        })}
      </ul>
    </div>
  )
}
