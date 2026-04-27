"use client"

import Link from "next/link"
import { Loader2, MessageSquareText, Sparkles } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  formatBusinessAppDate,
  formatSalesCurrency,
  getSalesOrderStageLabel,
  getSalesOrderStageTone,
  getSalesOrderStatusLabel,
  getSalesOrderStatusTone,
  resolveSalesOrderSourceHref,
} from "@/lib/business-apps"
import type { BusinessAppSchema, SalesOrder } from "@/types"

interface SalesOrderListViewProps {
  records: SalesOrder[]
  schema: BusinessAppSchema | null
  isLoading?: boolean
  error?: string | null
  nextCursor?: string | null
  onLoadMore?: () => void
  isLoadingMore?: boolean
}

export function SalesOrderListView({
  records,
  schema,
  isLoading,
  error,
  nextCursor,
  onLoadMore,
  isLoadingMore,
}: SalesOrderListViewProps) {
  if (error) {
    return <div className="rounded-2xl border border-rose-500/20 bg-rose-500/5 px-4 py-5 text-sm text-rose-700 dark:text-rose-300">{error}</div>
  }

  if (isLoading && records.length === 0) {
    return (
      <div className="flex min-h-[240px] items-center justify-center gap-2 rounded-3xl border border-dashed text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        Loading sales orders...
      </div>
    )
  }

  if (records.length === 0) {
    return (
      <div className="flex min-h-[220px] flex-col items-center justify-center gap-3 rounded-3xl border border-dashed bg-muted/10 text-center">
        <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
          <Sparkles className="h-6 w-6" />
        </div>
        <div className="space-y-1">
          <p className="text-sm font-black tracking-tight">No matching sales orders</p>
          <p className="max-w-md text-sm text-muted-foreground">Try adjusting the search, stage, status, or close-date filters.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-[11px] font-black uppercase tracking-widest text-muted-foreground">
          {records.length} visible order{records.length === 1 ? "" : "s"}
        </p>
        {schema?.entity ? <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">{schema.entity}</Badge> : null}
      </div>

      <div className="hidden grid-cols-[1.1fr_1.1fr_0.9fr_0.9fr_0.9fr_0.9fr_0.9fr] gap-3 rounded-2xl border bg-muted/30 px-4 py-3 text-[10px] font-black uppercase tracking-widest text-muted-foreground lg:grid">
        <span>Order</span>
        <span>Customer</span>
        <span>Stage</span>
        <span>Status</span>
        <span>Owner</span>
        <span>Amount</span>
        <span>Expected Close</span>
      </div>

      {records.map((order) => {
        const sourceHref = resolveSalesOrderSourceHref(order)
        return (
          <div key={order.id} className="rounded-2xl border bg-background/70 px-4 py-4 shadow-sm">
            <div className="grid gap-3 lg:grid-cols-[1.1fr_1.1fr_0.9fr_0.9fr_0.9fr_0.9fr_0.9fr] lg:items-center">
              <div>
                <p className="text-sm font-black tracking-tight">{order.order_number}</p>
                <p className="mt-1 text-[11px] text-muted-foreground">{order.summary ?? "No summary available."}</p>
              </div>
              <div>
                <p className="font-semibold">{order.customer_name}</p>
                {order.tags && order.tags.length > 0 ? <p className="mt-1 text-[11px] text-muted-foreground">{order.tags.join(" • ")}</p> : null}
              </div>
              <div>
                <Badge className={`text-[10px] font-black uppercase tracking-widest ${getSalesOrderStageTone(order.stage)}`}>
                  {getSalesOrderStageLabel(order.stage)}
                </Badge>
              </div>
              <div>
                <Badge className={`text-[10px] font-black uppercase tracking-widest ${getSalesOrderStatusTone(order.status)}`}>
                  {getSalesOrderStatusLabel(order.status)}
                </Badge>
              </div>
              <div className="text-sm text-muted-foreground">{order.owner_name ?? order.owner_user_id ?? "—"}</div>
              <div className="text-sm font-semibold">{formatSalesCurrency(order.amount, order.currency)}</div>
              <div className="space-y-1 text-sm text-muted-foreground">
                <div>{formatBusinessAppDate(order.expected_close_date)}</div>
                {sourceHref ? (
                  <Link href={sourceHref} className="inline-flex items-center gap-1 text-[11px] font-semibold text-violet-600 hover:underline">
                    <MessageSquareText className="h-3.5 w-3.5" />
                    Source
                  </Link>
                ) : null}
              </div>
            </div>
          </div>
        )
      })}

      {nextCursor && onLoadMore && (
        <div className="flex justify-center pt-2">
          <Button variant="outline" onClick={onLoadMore} disabled={!!isLoadingMore}>
            {isLoadingMore ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
            Load more
          </Button>
        </div>
      )}
    </div>
  )
}
