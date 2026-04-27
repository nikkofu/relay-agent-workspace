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

interface SalesOrderCardGridViewProps {
  records: SalesOrder[]
  schema: BusinessAppSchema | null
  isLoading?: boolean
  error?: string | null
  nextCursor?: string | null
  onLoadMore?: () => void
  isLoadingMore?: boolean
}

export function SalesOrderCardGridView({
  records,
  schema,
  isLoading,
  error,
  nextCursor,
  onLoadMore,
  isLoadingMore,
}: SalesOrderCardGridViewProps) {
  if (error) {
    return <div className="rounded-2xl border border-rose-500/20 bg-rose-500/5 px-4 py-5 text-sm text-rose-700 dark:text-rose-300">{error}</div>
  }

  if (isLoading && records.length === 0) {
    return (
      <div className="flex min-h-[240px] items-center justify-center gap-2 rounded-3xl border border-dashed text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        Loading sales cards...
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
          <p className="text-sm font-black tracking-tight">No matching sales cards</p>
          <p className="max-w-md text-sm text-muted-foreground">Try adjusting the shared search, stage, status, or date filters.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-[11px] font-black uppercase tracking-widest text-muted-foreground">
          {records.length} card{records.length === 1 ? "" : "s"}
        </p>
        {schema?.entity ? (
          <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
            {schema.entity}
          </Badge>
        ) : null}
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 2xl:grid-cols-3">
        {records.map((order) => {
          const sourceHref = resolveSalesOrderSourceHref(order)
          return (
            <article key={order.id} className="rounded-3xl border bg-background/70 p-5 shadow-sm">
              <div className="flex items-start justify-between gap-3">
                <div className="space-y-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="text-base font-black tracking-tight">{order.customer_name}</p>
                    <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
                      {order.order_number}
                    </Badge>
                  </div>
                  <p className="text-xs text-muted-foreground">{order.summary ?? "No summary available."}</p>
                </div>
                <div className="text-right">
                  <p className="text-base font-black">{formatSalesCurrency(order.amount, order.currency)}</p>
                  <p className="mt-1 text-[11px] text-muted-foreground">{formatBusinessAppDate(order.expected_close_date)}</p>
                </div>
              </div>

              <div className="mt-4 flex flex-wrap items-center gap-2">
                <Badge className={`text-[10px] font-black uppercase tracking-widest ${getSalesOrderStageTone(order.stage)}`}>
                  {getSalesOrderStageLabel(order.stage)}
                </Badge>
                <Badge className={`text-[10px] font-black uppercase tracking-widest ${getSalesOrderStatusTone(order.status)}`}>
                  {getSalesOrderStatusLabel(order.status)}
                </Badge>
                {order.owner_name || order.owner_user_id ? (
                  <span className="text-[11px] text-muted-foreground">{order.owner_name ?? order.owner_user_id}</span>
                ) : null}
              </div>

              {order.tags && order.tags.length > 0 ? (
                <div className="mt-4 flex flex-wrap gap-1.5">
                  {order.tags.map((tag) => (
                    <Badge key={`${order.id}-${tag}`} variant="secondary" className="text-[10px]">
                      {tag}
                    </Badge>
                  ))}
                </div>
              ) : null}

              <div className="mt-5 flex items-center justify-between gap-2 border-t pt-4 text-[11px] text-muted-foreground">
                <span>AI-native metadata only</span>
                {sourceHref ? (
                  <Link href={sourceHref} className="inline-flex items-center gap-1 font-semibold text-violet-600 hover:underline">
                    <MessageSquareText className="h-3.5 w-3.5" />
                    Source
                  </Link>
                ) : null}
              </div>
            </article>
          )
        })}
      </div>

      {nextCursor && onLoadMore ? (
        <div className="flex justify-center pt-2">
          <Button variant="outline" onClick={onLoadMore} disabled={!!isLoadingMore}>
            {isLoadingMore ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
            Load more
          </Button>
        </div>
      ) : null}
    </div>
  )
}
