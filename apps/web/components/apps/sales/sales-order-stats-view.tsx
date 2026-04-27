"use client"

import { BarChart3, Loader2, TrendingUp } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { formatSalesCurrency, getBusinessChartStyleLabel, getSalesOrderStageLabel } from "@/lib/business-apps"
import type { BusinessAppStats, BusinessChartStyle } from "@/types"

interface SalesOrderStatsViewProps {
  stats: BusinessAppStats | null
  chartStyle: BusinessChartStyle
  isLoading?: boolean
  error?: string | null
}

function SummaryCards({ stats }: { stats: BusinessAppStats }) {
  const summaryCards = [
    { label: "Pipeline amount", value: formatSalesCurrency(stats.total_amount) },
    { label: "Orders", value: stats.order_count?.toLocaleString() ?? "—" },
    { label: "Open", value: stats.open_count?.toLocaleString() ?? "—" },
    { label: "Won", value: stats.won_count?.toLocaleString() ?? "—" },
  ]

  return (
    <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4">
      {summaryCards.map((card) => (
        <div key={card.label} className="rounded-3xl border bg-background/70 p-4 shadow-sm">
          <p className="text-[11px] font-black uppercase tracking-widest text-muted-foreground">{card.label}</p>
          <p className="mt-3 text-2xl font-black tracking-tight">{card.value}</p>
        </div>
      ))}
    </div>
  )
}

function BarChart({ stats }: { stats: BusinessAppStats }) {
  return (
    <section className="rounded-3xl border bg-background/70 p-5 shadow-sm">
      <div className="flex items-center justify-between gap-2 border-b pb-3">
        <div>
          <h3 className="text-sm font-black tracking-tight">Amount by stage</h3>
          <p className="mt-1 text-xs text-muted-foreground">Backend-computed stage totals for the current Sales filters.</p>
        </div>
        <Badge variant="secondary" className="text-[10px] font-black uppercase tracking-widest">
          Bar
        </Badge>
      </div>
      <div className="mt-4 space-y-3">
        {stats.by_stage.map((bucket) => {
          const share = stats.total_amount && stats.total_amount > 0 ? (bucket.sum / stats.total_amount) * 100 : 0
          return (
            <div key={bucket.stage} className="space-y-2">
              <div className="flex items-center justify-between gap-3 text-sm">
                <div>
                  <p className="font-semibold">{getSalesOrderStageLabel(bucket.stage)}</p>
                  <p className="text-xs text-muted-foreground">{bucket.count} order{bucket.count === 1 ? "" : "s"}</p>
                </div>
                <p className="font-semibold">{formatSalesCurrency(bucket.sum)}</p>
              </div>
              <div className="h-2 rounded-full bg-muted">
                <div className="h-2 rounded-full bg-violet-500 transition-all" style={{ width: `${Math.max(share, 6)}%` }} />
              </div>
            </div>
          )
        })}
      </div>
    </section>
  )
}

function FunnelChart({ stats }: { stats: BusinessAppStats }) {
  const funnel = stats.funnel ?? stats.by_stage
  const maxAmount = Math.max(...funnel.map((bucket) => bucket.sum), 1)
  return (
    <section className="rounded-3xl border bg-background/70 p-5 shadow-sm">
      <div className="flex items-center justify-between gap-2 border-b pb-3">
        <div>
          <h3 className="text-sm font-black tracking-tight">Pipeline funnel</h3>
          <p className="mt-1 text-xs text-muted-foreground">Ordered stage progression using Gemini&apos;s funnel aggregates.</p>
        </div>
        <Badge variant="secondary" className="text-[10px] font-black uppercase tracking-widest">
          Funnel
        </Badge>
      </div>
      <div className="mt-4 space-y-3">
        {funnel.map((bucket, index) => {
          const width = `${Math.max((bucket.sum / maxAmount) * 100, 22)}%`
          return (
            <div key={bucket.stage} className="space-y-2">
              <div className="flex items-center justify-between gap-3 text-sm">
                <div className="flex items-center gap-2">
                  <span className="inline-flex h-6 w-6 items-center justify-center rounded-full bg-violet-500/10 text-[11px] font-black text-violet-600">{index + 1}</span>
                  <div>
                    <p className="font-semibold">{getSalesOrderStageLabel(bucket.stage)}</p>
                    <p className="text-xs text-muted-foreground">{bucket.count} order{bucket.count === 1 ? "" : "s"}</p>
                  </div>
                </div>
                <p className="font-semibold">{formatSalesCurrency(bucket.sum)}</p>
              </div>
              <div className="flex justify-center">
                <div className="rounded-2xl bg-gradient-to-r from-violet-600 to-fuchsia-500 px-4 py-3 text-center text-sm font-black text-white shadow-sm" style={{ width }}>
                  {bucket.count} · {formatSalesCurrency(bucket.sum)}
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </section>
  )
}

function TimelineChart({ stats }: { stats: BusinessAppStats }) {
  const buckets = stats.timeline_buckets ?? stats.expected_close_buckets ?? []
  const maxAmount = Math.max(...buckets.map((bucket) => bucket.amount), 1)
  return (
    <section className="rounded-3xl border bg-background/70 p-5 shadow-sm">
      <div className="flex items-center justify-between gap-2 border-b pb-3">
        <div>
          <h3 className="text-sm font-black tracking-tight">Revenue timeline</h3>
          <p className="mt-1 text-xs text-muted-foreground">Expected-close monthly buckets projected by the backend.</p>
        </div>
        <Badge variant="secondary" className="text-[10px] font-black uppercase tracking-widest">
          Timeline
        </Badge>
      </div>
      <div className="mt-6 flex items-end gap-3 overflow-x-auto pb-2">
        {buckets.map((bucket) => (
          <div key={bucket.label} className="flex min-w-[96px] flex-col items-center gap-2">
            <div className="text-[11px] font-semibold text-muted-foreground">{formatSalesCurrency(bucket.amount)}</div>
            <div className="flex w-full items-end justify-center rounded-t-2xl bg-violet-500/10 px-2" style={{ height: `${Math.max((bucket.amount / maxAmount) * 220, 24)}px` }}>
              <div className="mb-2 text-[11px] font-black text-violet-700 dark:text-violet-300">{bucket.count}</div>
            </div>
            <div className="text-center text-[11px] font-black uppercase tracking-widest text-muted-foreground">{bucket.label}</div>
          </div>
        ))}
      </div>
    </section>
  )
}

export function SalesOrderStatsView({ stats, chartStyle, isLoading, error }: SalesOrderStatsViewProps) {
  if (error) {
    return <div className="rounded-2xl border border-rose-500/20 bg-rose-500/5 px-4 py-5 text-sm text-rose-700 dark:text-rose-300">{error}</div>
  }

  if (isLoading && !stats) {
    return (
      <div className="flex min-h-[240px] items-center justify-center gap-2 rounded-3xl border border-dashed text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        Loading pipeline stats...
      </div>
    )
  }

  if (!stats) {
    return (
      <div className="flex min-h-[240px] flex-col items-center justify-center gap-3 rounded-3xl border border-dashed bg-muted/10 text-center">
        <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
          <BarChart3 className="h-6 w-6" />
        </div>
        <div className="space-y-1">
          <p className="text-sm font-black tracking-tight">No stats available</p>
          <p className="max-w-md text-sm text-muted-foreground">Once the Sales pipeline has data, aggregate totals and stage rollups will appear here.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
          {getBusinessChartStyleLabel(chartStyle)}
        </Badge>
        <span className="text-xs text-muted-foreground">Phase 74 keeps Sales App read-only while extending AI-native analytical context.</span>
      </div>

      <SummaryCards stats={stats} />

      {chartStyle === "bar" ? <BarChart stats={stats} /> : null}
      {chartStyle === "funnel" ? <FunnelChart stats={stats} /> : null}
      {chartStyle === "timeline" ? <TimelineChart stats={stats} /> : null}
      {chartStyle === "summary" ? (
        <div className="grid grid-cols-1 gap-5 xl:grid-cols-[1.4fr_0.8fr]">
          <BarChart stats={stats} />
          <section className="rounded-3xl border bg-background/70 p-5 shadow-sm">
            <div className="flex items-center justify-between gap-2 border-b pb-3">
              <div>
                <h3 className="text-sm font-black tracking-tight">Health snapshot</h3>
                <p className="mt-1 text-xs text-muted-foreground">Compact read model summary for AI-native follow-up work.</p>
              </div>
              <TrendingUp className="h-4 w-4 text-violet-500" />
            </div>
            <div className="mt-4 space-y-3 text-sm">
              <div className="flex items-center justify-between gap-3">
                <span className="text-muted-foreground">At-risk orders</span>
                <span className="font-semibold">{stats.at_risk_count?.toLocaleString() ?? "—"}</span>
              </div>
              <div className="flex items-center justify-between gap-3">
                <span className="text-muted-foreground">Lost orders</span>
                <span className="font-semibold">{stats.lost_count?.toLocaleString() ?? "—"}</span>
              </div>
              <div className="flex items-center justify-between gap-3">
                <span className="text-muted-foreground">Weighted pipeline</span>
                <span className="font-semibold">{formatSalesCurrency(stats.weighted_pipeline_amount)}</span>
              </div>
              <p className="rounded-2xl bg-muted/40 px-3 py-3 text-xs text-muted-foreground">
                Phase 74 keeps AI actions metadata-only. Codex and Windsurf should continue freezing action contracts before enabling autonomous Sales mutations.
              </p>
            </div>
          </section>
        </div>
      ) : null}

    </div>
  )
}
