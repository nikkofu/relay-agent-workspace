"use client"

import { useEffect, useMemo, useState } from "react"
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Input } from "@/components/ui/input"
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select"
import { Badge } from "@/components/ui/badge"
import { Calendar, Clock, Pin, Trash2, Loader2 } from "lucide-react"
import { format } from "date-fns"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import type { DigestScheduleInput, DigestWindow } from "@/types"

const DOW_LABELS = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"]

const DEFAULT_INPUT = (tz: string): DigestScheduleInput => ({
  window: 'weekly',
  timezone: tz,
  day_of_week: 1, // Monday
  day_of_month: 1,
  hour: 9,
  minute: 0,
  limit: 5,
  pin: true,
  is_enabled: true,
})

interface DigestScheduleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  channelId: string
  channelName?: string
}

export function DigestScheduleDialog({ open, onOpenChange, channelId, channelName }: DigestScheduleDialogProps) {
  const { digestSchedules, fetchDigestSchedule, upsertDigestSchedule, deleteDigestSchedule } = useKnowledgeStore()
  const existing = digestSchedules[channelId] || null
  const browserTz = useMemo(() => Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC", [])
  const [form, setForm] = useState<DigestScheduleInput>(() => DEFAULT_INPUT(browserTz))
  const [isLoadingInit, setIsLoadingInit] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  useEffect(() => {
    if (!open) return
    setIsLoadingInit(true)
    fetchDigestSchedule(channelId).then(schedule => {
      if (schedule) {
        setForm({
          window: schedule.window,
          timezone: schedule.timezone || browserTz,
          day_of_week: schedule.day_of_week ?? 1,
          day_of_month: schedule.day_of_month ?? 1,
          hour: schedule.hour,
          minute: schedule.minute,
          limit: schedule.limit,
          pin: schedule.pin,
          is_enabled: schedule.is_enabled,
        })
      } else {
        setForm(DEFAULT_INPUT(browserTz))
      }
      setIsLoadingInit(false)
    })
  }, [open, channelId, browserTz, fetchDigestSchedule])

  const updateField = <K extends keyof DigestScheduleInput>(key: K, value: DigestScheduleInput[K]) => {
    setForm(prev => ({ ...prev, [key]: value }))
  }

  const handleSave = async () => {
    setIsSaving(true)
    const result = await upsertDigestSchedule(channelId, form)
    setIsSaving(false)
    if (result) onOpenChange(false)
  }

  const handleDisable = async () => {
    setIsDeleting(true)
    const ok = await deleteDigestSchedule(channelId)
    setIsDeleting(false)
    if (ok) onOpenChange(false)
  }

  const pad = (n: number) => String(n).padStart(2, '0')

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Calendar className="w-4 h-4 text-emerald-600" />
            Digest schedule
            {channelName && (
              <Badge variant="secondary" className="text-[10px] ml-1">#{channelName}</Badge>
            )}
          </DialogTitle>
          <DialogDescription className="text-xs">
            Auto-publish a knowledge digest on a schedule. Teams will see it as a pinned channel message.
          </DialogDescription>
        </DialogHeader>

        {isLoadingInit ? (
          <div className="py-8 flex items-center justify-center text-muted-foreground text-xs gap-2">
            <Loader2 className="w-4 h-4 animate-spin" />
            Loading schedule...
          </div>
        ) : (
          <div className="space-y-4">
            {/* Enable */}
            <div className="flex items-center justify-between rounded-lg border px-3 py-2">
              <div className="flex flex-col">
                <Label className="text-xs font-black">Enable auto-publish</Label>
                <span className="text-[10px] text-muted-foreground">Relay publishes automatically at the scheduled time.</span>
              </div>
              <Switch
                checked={form.is_enabled}
                onCheckedChange={(v) => updateField('is_enabled', v)}
              />
            </div>

            {/* Window + cadence */}
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label className="text-[10px] font-black uppercase tracking-widest">Cadence</Label>
                <Select value={form.window} onValueChange={(v) => updateField('window', v as DigestWindow)}>
                  <SelectTrigger className="h-8 text-xs"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="daily">Daily</SelectItem>
                    <SelectItem value="weekly">Weekly</SelectItem>
                    <SelectItem value="monthly">Monthly</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {form.window === 'weekly' ? (
                <div className="space-y-1.5">
                  <Label className="text-[10px] font-black uppercase tracking-widest">Day of week</Label>
                  <Select value={String(form.day_of_week ?? 1)} onValueChange={(v) => updateField('day_of_week', Number(v))}>
                    <SelectTrigger className="h-8 text-xs"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      {DOW_LABELS.map((label, idx) => (
                        <SelectItem key={idx} value={String(idx)}>{label}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              ) : form.window === 'monthly' ? (
                <div className="space-y-1.5">
                  <Label className="text-[10px] font-black uppercase tracking-widest">Day of month</Label>
                  <Input
                    type="number"
                    min={1}
                    max={31}
                    className="h-8 text-xs"
                    value={form.day_of_month ?? 1}
                    onChange={(e) => updateField('day_of_month', Math.max(1, Math.min(31, Number(e.target.value) || 1)))}
                  />
                </div>
              ) : (
                <div className="space-y-1.5">
                  <Label className="text-[10px] font-black uppercase tracking-widest">Frequency</Label>
                  <div className="h-8 px-2 text-xs border rounded-md flex items-center text-muted-foreground italic">
                    Every day
                  </div>
                </div>
              )}
            </div>

            {/* Time + timezone */}
            <div className="grid grid-cols-3 gap-3">
              <div className="space-y-1.5">
                <Label className="text-[10px] font-black uppercase tracking-widest">Hour</Label>
                <Input
                  type="number"
                  min={0}
                  max={23}
                  className="h-8 text-xs"
                  value={form.hour}
                  onChange={(e) => updateField('hour', Math.max(0, Math.min(23, Number(e.target.value) || 0)))}
                />
              </div>
              <div className="space-y-1.5">
                <Label className="text-[10px] font-black uppercase tracking-widest">Minute</Label>
                <Input
                  type="number"
                  min={0}
                  max={59}
                  className="h-8 text-xs"
                  value={form.minute}
                  onChange={(e) => updateField('minute', Math.max(0, Math.min(59, Number(e.target.value) || 0)))}
                />
              </div>
              <div className="space-y-1.5">
                <Label className="text-[10px] font-black uppercase tracking-widest">Time</Label>
                <div className="h-8 px-2 text-xs border rounded-md flex items-center font-mono">
                  <Clock className="w-3 h-3 mr-1 text-muted-foreground" />
                  {pad(form.hour)}:{pad(form.minute)}
                </div>
              </div>
            </div>

            <div className="space-y-1.5">
              <Label className="text-[10px] font-black uppercase tracking-widest">Timezone</Label>
              <Input
                className="h-8 text-xs"
                value={form.timezone}
                onChange={(e) => updateField('timezone', e.target.value)}
                placeholder="e.g. America/Los_Angeles"
              />
            </div>

            {/* Limit + pin */}
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label className="text-[10px] font-black uppercase tracking-widest">Top entities</Label>
                <Input
                  type="number"
                  min={1}
                  max={20}
                  className="h-8 text-xs"
                  value={form.limit}
                  onChange={(e) => updateField('limit', Math.max(1, Math.min(20, Number(e.target.value) || 5)))}
                />
              </div>
              <div className="space-y-1.5">
                <Label className="text-[10px] font-black uppercase tracking-widest">Pin digest</Label>
                <div className="flex items-center gap-2 h-8 px-3 border rounded-md">
                  <Switch
                    checked={form.pin}
                    onCheckedChange={(v) => updateField('pin', v)}
                  />
                  <span className="text-xs text-muted-foreground">
                    <Pin className="w-3 h-3 inline mr-1" />
                    Pin to channel
                  </span>
                </div>
              </div>
            </div>

            {/* Meta */}
            {existing && (existing.last_published_at || existing.next_run_at) && (
              <div className="rounded-lg bg-muted/30 border px-3 py-2 text-[10px] space-y-0.5">
                {existing.next_run_at && (
                  <p><span className="font-black uppercase tracking-widest text-muted-foreground">Next run:</span> {format(new Date(existing.next_run_at), "MMM d, yyyy h:mm a")}</p>
                )}
                {existing.last_published_at && (
                  <p><span className="font-black uppercase tracking-widest text-muted-foreground">Last published:</span> {format(new Date(existing.last_published_at), "MMM d, yyyy h:mm a")}</p>
                )}
              </div>
            )}
          </div>
        )}

        <DialogFooter className="gap-2 sm:justify-between">
          <div>
            {existing && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleDisable}
                disabled={isDeleting || isSaving}
                className="text-rose-600 hover:text-rose-500 hover:bg-rose-500/10"
              >
                {isDeleting ? <Loader2 className="w-3 h-3 animate-spin mr-1" /> : <Trash2 className="w-3 h-3 mr-1" />}
                Remove schedule
              </Button>
            )}
          </div>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => onOpenChange(false)} disabled={isSaving}>
              Cancel
            </Button>
            <Button size="sm" onClick={handleSave} disabled={isSaving || isLoadingInit} className="bg-emerald-600 hover:bg-emerald-600/90">
              {isSaving ? <Loader2 className="w-3 h-3 animate-spin mr-1" /> : null}
              Save schedule
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

