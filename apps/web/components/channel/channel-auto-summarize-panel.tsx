"use client"

// Phase 63F: Rolling "always-on" channel summary panel.
// Consumes GET|PUT|POST /channels/:id/knowledge/auto-summarize and the
// websocket `channel.summary.updated` event wired in use-websocket.ts.
//
// This panel is additive — the existing manual "AI Channel Summary" section
// in channel-info.tsx is preserved. This one is about the persistent,
// per-channel rolling summary that regenerates when the channel materially
// changes and broadcasts to all open tabs.

import { useEffect, useState, useCallback } from "react"
import { Loader2, RefreshCw, Radio, Settings2, Zap } from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { cn } from "@/lib/utils"

interface ChannelAutoSummarizePanelProps {
  channelId: string
}

export function ChannelAutoSummarizePanel({ channelId }: ChannelAutoSummarizePanelProps) {
  const {
    channelAutoSummarize,
    isLoadingAutoSummarize,
    isRunningAutoSummarize,
    fetchChannelAutoSummarize,
    updateChannelAutoSummarize,
    runChannelAutoSummarize,
  } = useKnowledgeStore()

  const data = channelAutoSummarize[channelId]
  const setting = data?.setting
  const summary = data?.summary
  const loading = !!isLoadingAutoSummarize[channelId]
  const running = !!isRunningAutoSummarize[channelId]

  // Hydrate on mount / channel switch.
  useEffect(() => {
    if (!channelId) return
    fetchChannelAutoSummarize(channelId).catch(() => { /* best-effort */ })
  }, [channelId, fetchChannelAutoSummarize])

  const handleToggle = useCallback(async (enabled: boolean) => {
    // Send the currently-known knobs back so the backend doesn't reset
    // window/limit/min_new_messages to defaults when the user just flips the toggle.
    await updateChannelAutoSummarize(channelId, {
      is_enabled: enabled,
      window_hours: setting?.window_hours,
      message_limit: setting?.message_limit,
      min_new_messages: setting?.min_new_messages,
      provider: setting?.provider,
      model: setting?.model,
    })
  }, [channelId, setting, updateChannelAutoSummarize])

  const handleRun = useCallback(async () => {
    await runChannelAutoSummarize(channelId, { force: true })
  }, [channelId, runChannelAutoSummarize])

  return (
    <div className="space-y-3 bg-gradient-to-br from-sky-500/5 to-emerald-500/5 p-4 rounded-xl border border-sky-200 dark:border-sky-900/40">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <div className="relative">
            <Radio className="w-4 h-4 text-sky-600" />
            {setting?.is_enabled && (
              <span className="absolute -top-0.5 -right-0.5 w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
            )}
          </div>
          <div>
            <div className="flex items-center gap-1.5">
              <span className="text-sm font-bold text-foreground">Always-On Summary</span>
              {setting?.is_enabled && (
                <span className="text-[9px] font-black uppercase tracking-widest px-1.5 py-0.5 rounded-full border bg-emerald-500/10 border-emerald-400/40 text-emerald-700 dark:text-emerald-400">
                  Live
                </span>
              )}
            </div>
            <p className="text-[10px] text-muted-foreground">
              Rolling LLM summary; regenerates when channel activity materially changes.
            </p>
          </div>
        </div>
        {loading && !setting && (
          <Loader2 className="w-3.5 h-3.5 animate-spin text-muted-foreground" />
        )}
      </div>

      {/* Toggle + settings + run */}
      <div className="flex items-center gap-2 flex-wrap">
        <div className="flex items-center gap-2 rounded-md border bg-background/60 px-2 py-1">
          <Switch
            checked={!!setting?.is_enabled}
            onCheckedChange={handleToggle}
            disabled={!setting || running}
            aria-label="Toggle always-on auto-summarize"
          />
          <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">
            {setting?.is_enabled ? 'Enabled' : 'Disabled'}
          </span>
        </div>

        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              size="sm"
              className="h-7 text-[10px] gap-1.5"
              disabled={!setting}
              title="Tune rolling window and message cap"
            >
              <Settings2 className="w-3 h-3" />
              Settings
            </Button>
          </PopoverTrigger>
          <PopoverContent align="end" className="w-72 p-3 space-y-3">
            <div className="space-y-1.5">
              <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
                Window (hours)
              </Label>
              <Input
                type="number"
                min={1}
                max={720}
                className="h-7 text-xs"
                defaultValue={setting?.window_hours ?? 24}
                onBlur={async (e) => {
                  const v = Math.max(1, Math.min(720, Number(e.target.value) || 24))
                  if (setting && v !== setting.window_hours) {
                    await updateChannelAutoSummarize(channelId, { ...settingToInput(setting), window_hours: v })
                  }
                }}
              />
              <p className="text-[9px] text-muted-foreground">
                Look-back window for rolling summarization. Default 24h, max 720h (30d).
              </p>
            </div>
            <div className="space-y-1.5">
              <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
                Message limit
              </Label>
              <Input
                type="number"
                min={1}
                max={200}
                className="h-7 text-xs"
                defaultValue={setting?.message_limit ?? 50}
                onBlur={async (e) => {
                  const v = Math.max(1, Math.min(200, Number(e.target.value) || 50))
                  if (setting && v !== setting.message_limit) {
                    await updateChannelAutoSummarize(channelId, { ...settingToInput(setting), message_limit: v })
                  }
                }}
              />
              <p className="text-[9px] text-muted-foreground">
                Max messages pulled into each summarization run (1–200).
              </p>
            </div>
            <div className="space-y-1.5">
              <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
                Min new messages
              </Label>
              <Input
                type="number"
                min={1}
                max={100}
                className="h-7 text-xs"
                defaultValue={setting?.min_new_messages ?? 5}
                onBlur={async (e) => {
                  const v = Math.max(1, Math.min(100, Number(e.target.value) || 5))
                  if (setting && v !== setting.min_new_messages) {
                    await updateChannelAutoSummarize(channelId, { ...settingToInput(setting), min_new_messages: v })
                  }
                }}
              />
              <p className="text-[9px] text-muted-foreground">
                Threshold of unseen messages before a new rolling run is triggered.
              </p>
            </div>
            <div className="space-y-1.5">
              <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
                Provider / Model
              </Label>
              <div className="grid grid-cols-2 gap-2">
                <Input
                  type="text"
                  placeholder="provider"
                  className="h-7 text-xs font-mono"
                  defaultValue={setting?.provider || ''}
                  onBlur={async (e) => {
                    const v = e.target.value.trim()
                    if (setting && v !== setting.provider) {
                      await updateChannelAutoSummarize(channelId, { ...settingToInput(setting), provider: v })
                    }
                  }}
                />
                <Input
                  type="text"
                  placeholder="model"
                  className="h-7 text-xs font-mono"
                  defaultValue={setting?.model || ''}
                  onBlur={async (e) => {
                    const v = e.target.value.trim()
                    if (setting && v !== setting.model) {
                      await updateChannelAutoSummarize(channelId, { ...settingToInput(setting), model: v })
                    }
                  }}
                />
              </div>
              <p className="text-[9px] text-muted-foreground">Leave blank to use workspace defaults.</p>
            </div>
          </PopoverContent>
        </Popover>

        <Button
          size="sm"
          className="h-7 text-[10px] gap-1.5 ml-auto bg-sky-600 hover:bg-sky-700 text-white"
          disabled={!setting || running}
          onClick={handleRun}
          title="Run a summary now and broadcast channel.summary.updated"
        >
          {running ? <Loader2 className="w-3 h-3 animate-spin" /> : <Zap className="w-3 h-3" />}
          {running ? 'Running…' : 'Run now'}
        </Button>
      </div>

      {/* Metadata row */}
      {setting && (
        <div className="flex items-center gap-3 flex-wrap text-[9px] text-muted-foreground">
          <span>
            <span className="font-bold text-foreground/80">Last run: </span>
            {setting.last_run_at ? formatDistanceToNow(new Date(setting.last_run_at), { addSuffix: true }) : 'never'}
          </span>
          {setting.last_message_at && (
            <span>
              <span className="font-bold text-foreground/80">Covers: </span>
              up to {formatDistanceToNow(new Date(setting.last_message_at), { addSuffix: true })}
            </span>
          )}
          {summary?.message_count !== undefined && (
            <span>
              <span className="font-bold text-foreground/80">Messages: </span>
              {summary.message_count}
            </span>
          )}
          {(setting.provider || setting.model) && (
            <span className="font-mono ml-auto">
              {setting.provider}{setting.model ? ` / ${setting.model}` : ''}
            </span>
          )}
        </div>
      )}

      {/* Current summary content */}
      {summary?.content ? (
        <div className={cn(
          "rounded-lg border bg-background/60 p-3 text-sm leading-relaxed text-foreground/90 whitespace-pre-wrap",
          running && "opacity-70"
        )}>
          {summary.content}
        </div>
      ) : setting ? (
        <div className="rounded-lg border border-dashed border-muted-foreground/30 bg-muted/20 p-3 text-xs text-muted-foreground italic">
          No summary yet. {setting.is_enabled
            ? "Will run automatically when enough new messages arrive."
            : "Click Run now to generate one, or toggle on to enable rolling summaries."}
        </div>
      ) : null}

      {/* Manual refresh button (re-fetches settings + summary from server) */}
      <div className="flex items-center justify-end pt-1">
        <Button
          variant="ghost"
          size="sm"
          className="h-6 text-[9px] gap-1 text-muted-foreground"
          onClick={() => fetchChannelAutoSummarize(channelId)}
          disabled={loading || running}
          title="Re-fetch settings and current summary"
        >
          <RefreshCw className={cn("w-2.5 h-2.5", loading && "animate-spin")} />
          Refresh
        </Button>
      </div>
    </div>
  )
}

// Copy the durable knobs off an existing setting so PUTs don't reset unset fields.
// `force` is intentionally omitted — it only applies to POST runs.
function settingToInput(s: {
  is_enabled: boolean
  window_hours: number
  message_limit: number
  min_new_messages: number
  provider: string
  model: string
}) {
  return {
    is_enabled: s.is_enabled,
    window_hours: s.window_hours,
    message_limit: s.message_limit,
    min_new_messages: s.min_new_messages,
    provider: s.provider,
    model: s.model,
  }
}
