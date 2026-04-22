"use client"

import { useEffect, useMemo, useState } from "react"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { Loader2, Activity } from "lucide-react"
import { cn } from "@/lib/utils"
import type { EntityActivityBucket } from "@/types"

interface EntityActivitySparklineProps {
  entityId: string
  days?: number
  width?: number
  height?: number
  className?: string
}

/**
 * Inline SVG sparkline rendering the daily ref_count buckets for an entity.
 * Calls GET /api/v1/knowledge/entities/:id/activity and caches per-entity via the store.
 */
export function EntityActivitySparkline({
  entityId,
  days = 30,
  width = 220,
  height = 42,
  className,
}: EntityActivitySparklineProps) {
  const { entityActivity, fetchEntityActivity } = useKnowledgeStore()
  const [loading, setLoading] = useState(false)
  const activity = entityActivity[entityId]

  useEffect(() => {
    if (!entityId || activity) return
    setLoading(true)
    fetchEntityActivity(entityId, days).finally(() => setLoading(false))
  }, [entityId, days, activity, fetchEntityActivity])

  const { path, maxCount, totalRefs, lastBucket, areaPath } = useMemo(() => {
    const buckets: EntityActivityBucket[] = activity?.buckets || []
    if (buckets.length === 0) {
      return { path: "", areaPath: "", maxCount: 0, totalRefs: 0, lastBucket: null }
    }
    const max = Math.max(1, ...buckets.map(b => b.ref_count))
    const stepX = width / Math.max(1, buckets.length - 1)
    const points = buckets.map((b, i) => {
      const x = i * stepX
      const y = height - (b.ref_count / max) * (height - 6) - 3
      return `${x.toFixed(1)},${y.toFixed(1)}`
    })
    const p = `M ${points.join(" L ")}`
    const a = `M 0,${height} L ${points.join(" L ")} L ${width},${height} Z`
    const total = buckets.reduce((acc, b) => acc + b.ref_count, 0)
    return { path: p, areaPath: a, maxCount: max, totalRefs: total, lastBucket: buckets[buckets.length - 1] }
  }, [activity, width, height])

  if (loading && !activity) {
    return (
      <div className={cn("flex items-center gap-2 text-[10px] text-muted-foreground", className)}>
        <Loader2 className="w-3 h-3 animate-spin" /> loading activity…
      </div>
    )
  }

  if (!activity || !path) {
    return (
      <div className={cn("flex items-center gap-1.5 text-[10px] text-muted-foreground opacity-60", className)}>
        <Activity className="w-3 h-3" /> no activity in last {days}d
      </div>
    )
  }

  return (
    <div className={cn("flex flex-col gap-1", className)} title={`${totalRefs} references over last ${days} days`}>
      <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
        <Activity className="w-3 h-3 text-purple-500" />
        <span className="font-bold">{totalRefs}</span>
        <span>refs · last {days}d</span>
        {lastBucket && lastBucket.ref_count > 0 && (
          <span className="ml-auto font-bold text-purple-600">
            +{lastBucket.ref_count} today
          </span>
        )}
      </div>
      <svg
        width={width}
        height={height}
        viewBox={`0 0 ${width} ${height}`}
        className="overflow-visible"
        preserveAspectRatio="none"
      >
        <defs>
          <linearGradient id={`spark-grad-${entityId}`} x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="rgb(168 85 247)" stopOpacity="0.35" />
            <stop offset="100%" stopColor="rgb(168 85 247)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={areaPath} fill={`url(#spark-grad-${entityId})`} />
        <path
          d={path}
          fill="none"
          stroke="rgb(168 85 247)"
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        {/* last point dot */}
        {lastBucket && (
          <circle
            cx={width}
            cy={height - (lastBucket.ref_count / maxCount) * (height - 6) - 3}
            r={2.5}
            fill="rgb(168 85 247)"
          />
        )}
      </svg>
    </div>
  )
}
