"use client"

import Link from "next/link"
import { CalendarDays, Loader2, MessageSquareText } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import {
  formatBusinessAppDate,
  formatSalesCurrency,
  getBusinessCalendarTimeFieldLabel,
  getSalesOrderStageLabel,
  getSalesOrderStageTone,
  getSalesOrderStatusLabel,
  getSalesOrderStatusTone,
  resolveSalesOrderSourceHref,
} from "@/lib/business-apps"
import type { BusinessCalendarEvent, BusinessCalendarTimeField, BusinessCalendarView } from "@/types"

interface SalesOrderCalendarViewProps {
  events: BusinessCalendarEvent[]
  calendarView: BusinessCalendarView
  calendarTimeField: BusinessCalendarTimeField
  isLoading?: boolean
  error?: string | null
}

function startOfDay(date: Date): Date {
  const next = new Date(date)
  next.setHours(0, 0, 0, 0)
  return next
}

function addDays(date: Date, amount: number): Date {
  const next = new Date(date)
  next.setDate(next.getDate() + amount)
  return next
}

function startOfWeek(date: Date): Date {
  const next = startOfDay(date)
  const day = next.getDay()
  const diff = day === 0 ? -6 : 1 - day
  return addDays(next, diff)
}

function buildDayKey(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`
}

function formatDayLabel(date: Date): string {
  return new Intl.DateTimeFormat("en-US", { weekday: "short", month: "short", day: "numeric" }).format(date)
}

function formatMonthDayLabel(date: Date): string {
  return new Intl.DateTimeFormat("en-US", { month: "short", day: "numeric" }).format(date)
}

function buildAnchorDate(events: BusinessCalendarEvent[]): Date {
  const firstEvent = events[0]
  return firstEvent ? startOfDay(new Date(firstEvent.start)) : startOfDay(new Date())
}

function buildMonthGrid(anchor: Date): Date[] {
  const monthStart = new Date(anchor.getFullYear(), anchor.getMonth(), 1)
  const monthEnd = new Date(anchor.getFullYear(), anchor.getMonth() + 1, 0)
  const gridStart = startOfWeek(monthStart)
  const days: Date[] = []
  let cursor = gridStart
  while (cursor <= monthEnd || cursor.getDay() !== 1 || days.length < 35) {
    days.push(cursor)
    cursor = addDays(cursor, 1)
    if (days.length >= 42 && cursor > monthEnd) break
  }
  return days
}

function buildVisibleDays(view: BusinessCalendarView, anchor: Date): Date[] {
  if (view === "day") return [startOfDay(anchor)]
  if (view === "week") {
    const firstDay = startOfWeek(anchor)
    return Array.from({ length: 7 }, (_, index) => addDays(firstDay, index))
  }
  return buildMonthGrid(anchor)
}

function EventCard({ event, compact = false }: { event: BusinessCalendarEvent; compact?: boolean }) {
  const order = event.record
  const sourceHref = order ? resolveSalesOrderSourceHref(order) : null

  return (
    <div className="rounded-2xl border bg-background/80 p-3 shadow-sm">
      <div className="flex items-start justify-between gap-2">
        <div className="space-y-1">
          <p className={`font-black tracking-tight ${compact ? "text-xs" : "text-sm"}`}>{event.title}</p>
          <p className="text-[11px] text-muted-foreground">{formatBusinessAppDate(event.start)}</p>
        </div>
        <span className={`font-semibold ${compact ? "text-xs" : "text-sm"}`}>{formatSalesCurrency(event.amount, order?.currency)}</span>
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <Badge className={`text-[10px] font-black uppercase tracking-widest ${getSalesOrderStageTone(event.stage)}`}>
          {getSalesOrderStageLabel(event.stage)}
        </Badge>
        <Badge className={`text-[10px] font-black uppercase tracking-widest ${getSalesOrderStatusTone(event.status)}`}>
          {getSalesOrderStatusLabel(event.status)}
        </Badge>
      </div>

      {order?.summary ? <p className="mt-3 text-xs text-muted-foreground">{order.summary}</p> : null}

      {sourceHref ? (
        <div className="mt-3 flex justify-end">
          <Link href={sourceHref} className="inline-flex items-center gap-1 text-[11px] font-semibold text-violet-600 hover:underline">
            <MessageSquareText className="h-3.5 w-3.5" />
            Source
          </Link>
        </div>
      ) : null}
    </div>
  )
}

export function SalesOrderCalendarView({ events, calendarView, calendarTimeField, isLoading, error }: SalesOrderCalendarViewProps) {
  if (error) {
    return <div className="rounded-2xl border border-rose-500/20 bg-rose-500/5 px-4 py-5 text-sm text-rose-700 dark:text-rose-300">{error}</div>
  }

  if (isLoading && events.length === 0) {
    return (
      <div className="flex min-h-[240px] items-center justify-center gap-2 rounded-3xl border border-dashed text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        Loading calendar events...
      </div>
    )
  }

  if (events.length === 0) {
    return (
      <div className="flex min-h-[240px] flex-col items-center justify-center gap-3 rounded-3xl border border-dashed bg-muted/10 text-center">
        <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
          <CalendarDays className="h-6 w-6" />
        </div>
        <div className="space-y-1">
          <p className="text-sm font-black tracking-tight">No calendar events in range</p>
          <p className="max-w-md text-sm text-muted-foreground">Sales Orders with a valid {getBusinessCalendarTimeFieldLabel(calendarTimeField).toLowerCase()} will appear here in {calendarView} view.</p>
        </div>
      </div>
    )
  }

  const anchor = buildAnchorDate(events)
  const visibleDays = buildVisibleDays(calendarView, anchor)
  const eventsByDay = new Map<string, BusinessCalendarEvent[]>()
  for (const event of events) {
    const key = buildDayKey(new Date(event.start))
    eventsByDay.set(key, [...(eventsByDay.get(key) ?? []), event])
  }

  if (calendarView === "day") {
    const day = visibleDays[0]
    const dayEvents = eventsByDay.get(buildDayKey(day)) ?? []
    return (
      <section className="rounded-3xl border bg-background/70 p-5 shadow-sm">
        <div className="border-b pb-3">
          <h3 className="text-sm font-black tracking-tight">{formatDayLabel(day)}</h3>
          <p className="mt-1 text-xs text-muted-foreground">{dayEvents.length} event{dayEvents.length === 1 ? "" : "s"} projected from {getBusinessCalendarTimeFieldLabel(calendarTimeField).toLowerCase()}.</p>
        </div>
        <div className="mt-4 space-y-3">
          {dayEvents.map((event) => (
            <EventCard key={event.id} event={event} />
          ))}
        </div>
      </section>
    )
  }

  if (calendarView === "week") {
    return (
      <div className="grid grid-cols-1 gap-4 xl:grid-cols-7">
        {visibleDays.map((day) => {
          const dayEvents = eventsByDay.get(buildDayKey(day)) ?? []
          return (
            <section key={buildDayKey(day)} className="rounded-3xl border bg-background/70 p-4 shadow-sm">
              <div className="border-b pb-3">
                <h3 className="text-sm font-black tracking-tight">{formatDayLabel(day)}</h3>
                <p className="mt-1 text-[11px] text-muted-foreground">{dayEvents.length} event{dayEvents.length === 1 ? "" : "s"}</p>
              </div>
              <div className="mt-3 space-y-3">
                {dayEvents.length === 0 ? (
                  <div className="rounded-2xl border border-dashed px-3 py-6 text-center text-xs text-muted-foreground">No events.</div>
                ) : (
                  dayEvents.map((event) => <EventCard key={event.id} event={event} compact />)
                )}
              </div>
            </section>
          )
        })}
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="rounded-3xl border bg-background/70 p-4 shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-2 border-b pb-3">
          <div>
            <h3 className="text-sm font-black tracking-tight">{anchor.toLocaleString("en-US", { month: "long", year: "numeric" })}</h3>
            <p className="mt-1 text-xs text-muted-foreground">{events.length} event{events.length === 1 ? "" : "s"} projected from {getBusinessCalendarTimeFieldLabel(calendarTimeField).toLowerCase()}.</p>
          </div>
          <Badge variant="secondary" className="text-[10px] font-black uppercase tracking-widest">
            {calendarView}
          </Badge>
        </div>

        <div className="mt-4 grid grid-cols-7 gap-2 text-[10px] font-black uppercase tracking-widest text-muted-foreground">
          {["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"].map((label) => (
            <div key={label} className="px-2 py-1">{label}</div>
          ))}
        </div>

        <div className="mt-2 grid grid-cols-1 gap-2 lg:grid-cols-7">
          {visibleDays.map((day) => {
            const dayEvents = eventsByDay.get(buildDayKey(day)) ?? []
            const inActiveMonth = day.getMonth() === anchor.getMonth()
            return (
              <div key={buildDayKey(day)} className={`min-h-[180px] rounded-2xl border p-3 ${inActiveMonth ? "bg-background/70" : "bg-muted/20"}`}>
                <div className="mb-3 flex items-center justify-between gap-2">
                  <span className={`text-sm font-black tracking-tight ${inActiveMonth ? "text-foreground" : "text-muted-foreground"}`}>{formatMonthDayLabel(day)}</span>
                  <span className="text-[11px] text-muted-foreground">{dayEvents.length}</span>
                </div>
                <div className="space-y-2">
                  {dayEvents.length === 0 ? (
                    <div className="rounded-xl border border-dashed px-2 py-4 text-center text-[11px] text-muted-foreground">No events</div>
                  ) : (
                    dayEvents.slice(0, 3).map((event) => <EventCard key={event.id} event={event} compact />)
                  )}
                  {dayEvents.length > 3 ? <div className="text-[11px] font-semibold text-violet-600">+{dayEvents.length - 3} more</div> : null}
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
