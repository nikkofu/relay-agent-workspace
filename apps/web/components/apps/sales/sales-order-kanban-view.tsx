"use client"

import Link from "next/link"
import { Columns3, Loader2, MessageSquareText } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import {
  formatBusinessAppDate,
  formatSalesCurrency,
  getSalesOrderStageTone,
  getSalesOrderStatusLabel,
  getSalesOrderStatusTone,
  resolveSalesOrderSourceHref,
} from "@/lib/business-apps"
import type { BusinessAppDataGroup } from "@/types"

interface SalesOrderKanbanViewProps {
  groups: BusinessAppDataGroup[]
  isLoading?: boolean
  error?: string | null
}

export function SalesOrderKanbanView({ groups, isLoading, error }: SalesOrderKanbanViewProps) {
  if (error) {
    return <div className="rounded-2xl border border-rose-500/20 bg-rose-500/5 px-4 py-5 text-sm text-rose-700 dark:text-rose-300">{error}</div>
  }

  if (isLoading && groups.length === 0) {
    return (
      <div className="flex min-h-[240px] items-center justify-center gap-2 rounded-3xl border border-dashed text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        Loading pipeline columns...
      </div>
    )
  }

  if (groups.length === 0) {
    return (
      <div className="flex min-h-[240px] flex-col items-center justify-center gap-3 rounded-3xl border border-dashed bg-muted/10 text-center">
        <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
          <Columns3 className="h-6 w-6" />
        </div>
        <div className="space-y-1">
          <p className="text-sm font-black tracking-tight">No pipeline cards yet</p>
          <p className="max-w-md text-sm text-muted-foreground">Orders will group into stage columns here using the same Sales Order contract as list and calendar.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="overflow-x-auto pb-2">
      <div className="flex min-w-max gap-4">
        {groups.map((group) => (
          <section key={group.id} className="flex w-[300px] shrink-0 flex-col rounded-3xl border bg-muted/20 p-3">
            <div className="flex items-center justify-between gap-2 border-b px-1 pb-3">
              <div>
                <h3 className="text-sm font-black tracking-tight">{group.title}</h3>
                <p className="mt-1 text-[11px] text-muted-foreground">{group.count} card{group.count === 1 ? "" : "s"}</p>
              </div>
              <Badge className={`text-[10px] font-black uppercase tracking-widest ${getSalesOrderStageTone(group.key)}`}>
                {group.key.replace(/_/g, " ")}
              </Badge>
            </div>

            <div className="mt-3 space-y-3">
              {group.records.length === 0 ? (
                <div className="rounded-2xl border border-dashed px-3 py-6 text-center text-xs text-muted-foreground">No orders in this stage.</div>
              ) : (
                group.records.map((order) => {
                  const sourceHref = resolveSalesOrderSourceHref(order)
                  return (
                    <div key={order.id} className="rounded-2xl border bg-background px-3 py-3 shadow-sm">
                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <p className="text-sm font-black tracking-tight">{order.customer_name}</p>
                          <p className="mt-1 text-[11px] text-muted-foreground">{order.order_number}</p>
                        </div>
                        <Badge className={`text-[10px] font-black uppercase tracking-widest ${getSalesOrderStatusTone(order.status)}`}>
                          {getSalesOrderStatusLabel(order.status)}
                        </Badge>
                      </div>
                      <p className="mt-3 text-sm font-semibold">{formatSalesCurrency(order.amount, order.currency)}</p>
                      <p className="mt-1 text-[11px] text-muted-foreground">Close {formatBusinessAppDate(order.expected_close_date)}</p>
                      <p className="mt-3 text-xs text-muted-foreground">{order.summary ?? "No summary available."}</p>
                      <div className="mt-3 flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
                        {order.owner_name || order.owner_user_id ? <span>{order.owner_name ?? order.owner_user_id}</span> : null}
                        {sourceHref ? (
                          <Link href={sourceHref} className="inline-flex items-center gap-1 font-semibold text-violet-600 hover:underline">
                            <MessageSquareText className="h-3.5 w-3.5" />
                            Source
                          </Link>
                        ) : null}
                      </div>
                    </div>
                  )
                })
              )}
            </div>
          </section>
        ))}
      </div>
    </div>
  )
}
