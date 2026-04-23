"use client"

import { useEffect } from "react"
import { CalendarCheck, CalendarX, Download, Loader2, XCircle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { formatDistanceToNow, format } from "date-fns"
import type { AIScheduleBooking } from "@/types"

// Phase 63H: download a booking's ICS content as a .ics file
function downloadICS(booking: AIScheduleBooking) {
  if (!booking.ics_content) return
  const blob = new Blob([booking.ics_content], { type: 'text/calendar;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${booking.title.replace(/[^a-z0-9]/gi, '-').toLowerCase() || 'booking'}.ics`
  a.click()
  URL.revokeObjectURL(url)
}

function formatSlot(booking: AIScheduleBooking): string {
  try {
    const start = new Date(booking.starts_at)
    const end = new Date(booking.ends_at)
    return `${format(start, 'EEE, MMM d · h:mm a')} – ${format(end, 'h:mm a')} (${booking.timezone})`
  } catch {
    return booking.starts_at
  }
}

interface Props {
  channelId?: string
  dmId?: string
  compact?: boolean
  className?: string
}

export function ScheduleBookingsPanel({ channelId, dmId, compact = false, className = '' }: Props) {
  const {
    scheduleBookings,
    isLoadingScheduleBookings,
    hasHydratedScheduleBookings,
    fetchAIScheduleBookings,
    cancelAIScheduleBooking,
  } = useKnowledgeStore(s => ({
    scheduleBookings: s.scheduleBookings,
    isLoadingScheduleBookings: s.isLoadingScheduleBookings,
    hasHydratedScheduleBookings: s.hasHydratedScheduleBookings,
    fetchAIScheduleBookings: s.fetchAIScheduleBookings,
    cancelAIScheduleBooking: s.cancelAIScheduleBooking,
  }))

  const scopeKey = channelId ? `ch:${channelId}` : dmId ? `dm:${dmId}` : 'all'
  const isLoading = isLoadingScheduleBookings[scopeKey]
  const hasHydrated = hasHydratedScheduleBookings[scopeKey]

  useEffect(() => {
    if (!hasHydrated && !isLoading) {
      fetchAIScheduleBookings({ channelId, dmId })
    }
  }, [channelId, dmId, fetchAIScheduleBookings, hasHydrated, isLoading])

  // Client-filter to current scope
  const filtered = scheduleBookings.filter(b => {
    if (channelId) return b.channel_id === channelId
    if (dmId) return b.dm_id === dmId
    return true
  })

  if (isLoading && !hasHydrated) {
    return (
      <div className={`flex items-center gap-2 text-xs text-muted-foreground ${className}`}>
        <Loader2 className="h-3 w-3 animate-spin" />
        Loading bookings…
      </div>
    )
  }

  return (
    <div className={`space-y-2 ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-1.5 text-[10px] font-bold uppercase tracking-wider text-muted-foreground">
          <CalendarCheck className="h-3 w-3" />
          Schedule Bookings
          {filtered.length > 0 && (
            <Badge variant="secondary" className="rounded px-1.5 py-0 text-[10px]">
              {filtered.length}
            </Badge>
          )}
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="h-5 px-1.5 text-[10px] text-muted-foreground"
          onClick={() => fetchAIScheduleBookings({ channelId, dmId })}
          disabled={!!isLoading}
        >
          {isLoading ? <Loader2 className="h-2.5 w-2.5 animate-spin" /> : 'Refresh'}
        </Button>
      </div>

      {filtered.length === 0 ? (
        <p className="text-[11px] text-muted-foreground italic">
          No schedule bookings yet. When you book a slot from an AI Suggest response, it will appear here.
        </p>
      ) : (
        <div className="space-y-2">
          {filtered.slice(0, compact ? 3 : 20).map(booking => (
            <div
              key={booking.id}
              className={`rounded-lg border px-3 py-2 space-y-1 text-xs ${
                booking.status === 'cancelled'
                  ? 'opacity-50 bg-muted/10'
                  : 'bg-background/80'
              }`}
            >
              <div className="flex items-start justify-between gap-2">
                <div className="space-y-0.5 min-w-0">
                  <div className="font-semibold leading-tight truncate">{booking.title}</div>
                  <div className="text-[10px] text-muted-foreground">{formatSlot(booking)}</div>
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  <Badge
                    variant="secondary"
                    className={`rounded px-1.5 py-0 text-[10px] font-medium ${
                      booking.status === 'booked'
                        ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300'
                        : 'bg-muted/40 text-muted-foreground'
                    }`}
                  >
                    {booking.status === 'booked'
                      ? <><CalendarCheck className="inline mr-0.5 h-2.5 w-2.5" />Booked</>
                      : <><CalendarX className="inline mr-0.5 h-2.5 w-2.5" />Cancelled</>
                    }
                  </Badge>
                </div>
              </div>
              <div className="flex items-center justify-between gap-2">
                <div className="text-[10px] text-muted-foreground">
                  {booking.attendee_ids?.length
                    ? `${booking.attendee_ids.length} attendee${booking.attendee_ids.length > 1 ? 's' : ''}`
                    : null}
                  {booking.provider && <span className="ml-2 font-mono">{booking.provider}</span>}
                  <span className="ml-2">{formatDistanceToNow(new Date(booking.created_at), { addSuffix: true })}</span>
                </div>
                <div className="flex items-center gap-1">
                  {booking.ics_content && (
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-5 w-5 text-muted-foreground"
                      title="Download .ics"
                      onClick={() => downloadICS(booking)}
                    >
                      <Download className="h-2.5 w-2.5" />
                    </Button>
                  )}
                  {booking.status === 'booked' && (
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-5 w-5 text-destructive/70 hover:text-destructive"
                      title="Cancel booking"
                      onClick={() => cancelAIScheduleBooking(booking.id)}
                    >
                      <XCircle className="h-2.5 w-2.5" />
                    </Button>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
