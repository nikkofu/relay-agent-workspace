"use client"

import Link from "next/link"
import { CalendarDays, Loader2, MessageSquareText } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import {
  formatSalesCurrency,
  getSalesOrderStageLabel,
  getSalesOrderStageTone,
  getSalesOrderStatusLabel,
  getSalesOrderStatusTone,
  resolveSalesOrderSourceHref,
} from "@/lib/business-apps"
import type { BusinessAppDataGroup } from "@/types"

interface SalesOrderCalendarViewProps {
  groups: BusinessAppDataGroup[]
  isLoading?: boolean
  error?: string | null
}

export function SalesOrderCalendarView({ groups, isLoading, error }: SalesOrderCalendarViewProps) {
  if (error) {
    return <div className="rounded-2xl border border-rose-500/20 bg-rose-500/5 px-4 py-5 text-sm text-rose-700 dark:text-rose-300">{error}</div>
  }

  if (isLoading && groups.length === 0) {
    return (
      <div className="flex min-h-[240px] items-center justify-center gap-2 rounded-3xl border border-dashed text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        Loading calendar buckets...
      </div>
    )
  }

  if (groups.length === 0) {
    return (
      <div className="flex min-h-[240px] flex-col items-center justify-center gap-3 rounded-3xl border border-dashed bg-muted/10 text-center">
        <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
          <CalendarDays className="h-6 w-6" />
        </div>
        <div className="space-y-1">
          <p className="text-sm font-black tracking-tight">No expected close dates in range</p>
          <p className="max-w-md text-sm text-muted-foreground">Orders with an expected close date will appear in daily buckets here. Orders without one move to Unscheduled.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {groups.map((group) => (
        <section key={group.id} className="rounded-3xl border bg-background/70 p-4 shadow-sm">
          <div className="flex flex-wrap items-center justify-between gap-2 border-b pb-3">
            <div>
              <h3 className="text-sm font-black tracking-tight">{group.title}</h3>
              <p className="mt-1 text-xs text-muted-foreground">{group.count} order{group.count === 1 ? "" : "s"}</p>
            </div>
            <Badge variant="secondary" className="text-[10px] font-black uppercase tracking-widest">
              Expected close
            </Badge>
          </div>

          <div className="mt-4 grid grid-cols-1 gap-3 xl:grid-cols-2">
            {group.records.map((order) => {
              const sourceHref = resolveSalesOrderSourceHref(order)
              return (
                <div key={order.id} className="rounded-2xl border bg-muted/10 p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <p className="text-sm font-black tracking-tight">{order.customer_name}</p>
                      <p className="mt-1 text-[11px] text-muted-foreground">{order.order_number}</p>
                    </div>
                    <p className="text-sm font-semibold">{formatSalesCurrency(order.amount, order.currency)}</p>
                  </div>
                  <p className="mt-3 text-xs text-muted-foreground">{order.summary ?? "No summary available."}</p>
                  <div className="mt-3 flex flex-wrap items-center gap-2">
                    <Badge className={`text-[10px] font-black uppercase tracking-widest ${getSalesOrderStageTone(order.stage)}`}>
                      {getSalesOrderStageLabel(order.stage)}
                    </Badge>
                    <Badge className={`text-[10px] font-black uppercase tracking-widest ${getSalesOrderStatusTone(order.status)}`}>
                      {getSalesOrderStatusLabel(order.status)}
                    </Badge>
                    {order.owner_name || order.owner_user_id ? <span className="text-[11px] text-muted-foreground">{order.owner_name ?? order.owner_user_id}</span> : null}
                    {sourceHref ? (
                      <Link href={sourceHref} className="inline-flex items-center gap-1 text-[11px] font-semibold text-violet-600 hover:underline">
                        <MessageSquareText className="h-3.5 w-3.5" />
                        Source
                      </Link>
                    ) : null}
                  </div>
                </div>
              )
            })}
          </div>
        </section>
      ))}
    </div>
  )
}
