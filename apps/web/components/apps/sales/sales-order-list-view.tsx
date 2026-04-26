"use client"

import Link from "next/link"
import { Loader2, MessageSquareText, Search, Sparkles } from "lucide-react"
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
  mode: "list" | "search"
  records: SalesOrder[]
  schema: BusinessAppSchema | null
  query?: string
  isLoading?: boolean
  error?: string | null
  nextCursor?: string | null
  onLoadMore?: () => void
  isLoadingMore?: boolean
}

function SearchResultCard({ order }: { order: SalesOrder }) {
  const sourceHref = resolveSalesOrderSourceHref(order)

  return (
    <div className="rounded-2xl border bg-background/70 p-4 shadow-sm">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="space-y-1">
          <div className="flex flex-wrap items-center gap-2">
            <p className="text-sm font-black tracking-tight">{order.customer_name}</p>
            <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
              {order.order_number}
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground">{order.summary ?? "No summary available."}</p>
        </div>
        <div className="text-right">
          <p className="text-sm font-black">{formatSalesCurrency(order.amount, order.currency)}</p>
          <p className="text-[11px] text-muted-foreground">Close {formatBusinessAppDate(order.expected_close_date)}</p>
        </div>
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <Badge className={`text-[10px] font-black uppercase tracking-widest ${getSalesOrderStageTone(order.stage)}`}>
          {getSalesOrderStageLabel(order.stage)}
        </Badge>
        <Badge className={`text-[10px] font-black uppercase tracking-widest ${getSalesOrderStatusTone(order.status)}`}>
          {getSalesOrderStatusLabel(order.status)}
        </Badge>
        {order.owner_name || order.owner_user_id ? (
          <span className="text-[11px] text-muted-foreground">Owner {order.owner_name ?? order.owner_user_id}</span>
        ) : null}
        {sourceHref ? (
          <Link href={sourceHref} className="inline-flex items-center gap-1 text-[11px] font-semibold text-violet-600 hover:underline">
            <MessageSquareText className="h-3.5 w-3.5" />
            Source channel
          </Link>
        ) : null}
      </div>

      {order.tags && order.tags.length > 0 && (
        <div className="mt-3 flex flex-wrap gap-1.5">
          {order.tags.map((tag) => (
            <Badge key={`${order.id}-${tag}`} variant="secondary" className="text-[10px]">
              {tag}
            </Badge>
          ))}
        </div>
      )}
    </div>
  )
}

export function SalesOrderListView({
  mode,
  records,
  schema,
  query,
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

  if (mode === "search" && !query?.trim()) {
    return (
      <div className="flex min-h-[260px] flex-col items-center justify-center gap-3 rounded-3xl border border-dashed bg-muted/10 text-center">
        <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-violet-500/10 text-violet-500">
          <Search className="h-6 w-6" />
        </div>
        <div className="space-y-1">
          <p className="text-sm font-black tracking-tight">Search Sales Orders</p>
          <p className="max-w-md text-sm text-muted-foreground">Search by customer, order number, summary, or tags. Results stay on the same Sales data contract as every other mode.</p>
        </div>
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

  if (mode === "search") {
    return (
      <div className="space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <p className="text-[11px] font-black uppercase tracking-widest text-muted-foreground">
            {records.length} result{records.length === 1 ? "" : "s"}{query ? ` for “${query}”` : ""}
          </p>
          {schema?.entity && <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">{schema.entity}</Badge>}
        </div>
        <div className="grid grid-cols-1 gap-3 xl:grid-cols-2">
          {records.map((order) => (
            <SearchResultCard key={order.id} order={order} />
          ))}
        </div>
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

  return (
    <div className="space-y-3">
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
