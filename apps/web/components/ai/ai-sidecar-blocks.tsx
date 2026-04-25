"use client"

// ── Shared AI-side-channel rendering blocks ──────────────────────────────────
//
// One implementation of the "Thinking" panel + tool timeline + usage chip,
// reused by every AI surface in the app (DM bubbles, channel `/ask`
// replies, canvas AI Dock). All three surfaces are required by the
// Unified AI Side-Channel Contract (`docs/superpowers/specs/2026-04-25-…`)
// to render the same canonical shape, so it's important they share one
// implementation rather than each rolling its own copy.

import { useEffect, useRef, useState } from "react"
import { Brain, ChevronDown, ChevronRight, Wrench, Hash, Check, Loader2, X } from "lucide-react"
import { cn } from "@/lib/utils"
import type { AIReasoning, AIToolCall, AIUsage } from "@/lib/ai-sidecar"

// ─── Reasoning panel ────────────────────────────────────────────────────────

interface ReasoningPanelProps {
  reasoning: AIReasoning
  /** When true, auto-expanded with a pulsing indicator (live streaming). */
  live?: boolean
  /** When true, expanded by default (e.g. canvas AI Dock) — otherwise the
   *  user has to click the header to expand the persisted thoughts. */
  defaultOpen?: boolean
  /** Optional duration label rendered on the right of the header. */
  durationMs?: number
  /** Compact variant flips the styling tokens for inline DM bubbles. */
  variant?: "default" | "amber" | "violet"
  className?: string
}

export function ReasoningPanel({
  reasoning, live, defaultOpen, durationMs, variant = "default", className,
}: ReasoningPanelProps) {
  const [open, setOpen] = useState(!!live || !!defaultOpen)
  const scrollerRef = useRef<HTMLDivElement>(null)

  // Auto-expand the moment a stream goes live so the user sees the
  // chain-of-thought without an extra click.
  useEffect(() => { if (live) setOpen(true) }, [live])

  // Pin the scroller to the bottom while new tokens are arriving so the
  // most-recent reasoning is always visible.
  useEffect(() => {
    if (!live) return
    const el = scrollerRef.current
    if (!el) return
    el.scrollTop = el.scrollHeight
  }, [reasoning.summary, reasoning.segments, live, open])

  const text = (reasoning.summary && reasoning.summary.trim())
    || reasoning.segments.map(s => s.text).filter(Boolean).join("\n")
    || ""
  if (!text) return null

  const wordCount = text.trim().split(/\s+/).filter(Boolean).length

  const tone = variant === "amber"
    ? "border-amber-400/30 bg-amber-50/40 dark:bg-amber-950/10 text-amber-900/80 dark:text-amber-100/70"
    : variant === "violet" || live
      ? "border-violet-500/30 bg-violet-50/50 dark:bg-violet-950/20"
      : "border-border bg-muted/40"

  return (
    <div className={cn("rounded-xl border text-[11px] leading-relaxed", tone, className)}>
      <button
        type="button"
        onClick={() => setOpen(o => !o)}
        className="w-full flex items-center gap-1.5 px-3 py-1.5 hover:bg-black/5 dark:hover:bg-white/5 transition-colors rounded-t-xl"
      >
        {open ? <ChevronDown className="w-3 h-3" /> : <ChevronRight className="w-3 h-3" />}
        <Brain className={cn(
          "w-3 h-3",
          live ? "text-violet-600 animate-pulse" : "text-muted-foreground",
        )} />
        <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
          {live ? "Thinking…" : "Thoughts"}
        </span>
        <span className="text-[9px] text-muted-foreground/70 ml-auto inline-flex items-center gap-1.5">
          {wordCount > 0 && <span>{wordCount}w</span>}
          {!live && durationMs !== undefined && (
            <span>{(durationMs / 1000).toFixed(1)}s</span>
          )}
        </span>
      </button>
      {open && (
        <div
          ref={scrollerRef}
          className="max-h-48 overflow-y-auto px-3 pb-2 pt-0.5 whitespace-pre-wrap font-mono text-[11px]"
        >
          {reasoning.segments.length > 0
            ? (
              <div className="space-y-1">
                {reasoning.segments.map((s, i) => (
                  <div key={i} className={cn(
                    s.kind === "step" && "text-foreground",
                    s.kind === "note" && "italic text-muted-foreground",
                    s.kind === "thought" && "text-muted-foreground",
                  )}>
                    {s.kind === "step" && <span className="text-violet-600 font-black mr-1">▸</span>}
                    {s.text}
                  </div>
                ))}
              </div>
            )
            : <p className="text-muted-foreground">{reasoning.summary}</p>}
        </div>
      )}
    </div>
  )
}

// ─── Tool timeline ──────────────────────────────────────────────────────────

interface ToolTimelineProps {
  toolCalls: AIToolCall[]
  className?: string
}

const STATUS_TINT: Record<AIToolCall["status"], string> = {
  pending: "bg-muted text-muted-foreground",
  running: "bg-blue-500/15 text-blue-600 dark:text-blue-400",
  success: "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400",
  failed: "bg-rose-500/15 text-rose-600 dark:text-rose-400",
}

export function ToolTimeline({ toolCalls, className }: ToolTimelineProps) {
  if (!toolCalls.length) return null
  return (
    <div className={cn("rounded-xl border bg-muted/40 px-3 py-2 space-y-1", className)}>
      <div className="flex items-center gap-1.5">
        <Wrench className="w-3 h-3 text-muted-foreground" />
        <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
          Tools used · {toolCalls.length}
        </span>
      </div>
      <ol className="space-y-1">
        {toolCalls.map((c, i) => (
          <li key={c.id || i} className="flex items-start gap-1.5 text-[11px]">
            <span className={cn(
              "inline-flex w-3.5 h-3.5 rounded-full items-center justify-center shrink-0 mt-0.5",
              STATUS_TINT[c.status],
            )}>
              {c.status === "success" ? <Check className="w-2 h-2" />
                : c.status === "running" ? <Loader2 className="w-2 h-2 animate-spin" />
                : c.status === "failed" ? <X className="w-2 h-2" />
                : <span className="text-[9px] font-black">{i + 1}</span>}
            </span>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-1.5">
                <span className="font-semibold truncate">{c.name}</span>
                {c.duration_ms !== undefined && (
                  <span className="text-muted-foreground/70 ml-auto text-[9px] shrink-0">
                    {c.duration_ms < 1000 ? `${c.duration_ms}ms` : `${(c.duration_ms / 1000).toFixed(1)}s`}
                  </span>
                )}
              </div>
              {(c.input_summary || c.output_summary) && (
                <div className="text-[10px] text-muted-foreground/80 truncate">
                  {c.input_summary && <span>← {c.input_summary}</span>}
                  {c.input_summary && c.output_summary && <span className="mx-1">·</span>}
                  {c.output_summary && <span>→ {c.output_summary}</span>}
                </div>
              )}
            </div>
          </li>
        ))}
      </ol>
    </div>
  )
}

// ─── Usage chip ─────────────────────────────────────────────────────────────

interface UsageChipProps {
  usage: AIUsage
  /** Optional fallback estimate to render when usage has no totals. The
   *  contract (Frontend Consumption Rule §2) only allows fallback display
   *  WHEN no usage payload exists, so callers must check `usage` is null
   *  before passing the estimate; we surface it here only as a last-resort
   *  display when a payload exists but is empty. */
  estimateTotal?: number
  className?: string
}

export function UsageChip({ usage, estimateTotal, className }: UsageChipProps) {
  const total = usage.total_tokens
    ?? ((usage.input_tokens ?? 0) + (usage.output_tokens ?? 0) || estimateTotal)
  if (total === undefined || total <= 0) return null

  const breakdown: string[] = []
  if (usage.input_tokens !== undefined) breakdown.push(`${usage.input_tokens.toLocaleString()} in`)
  if (usage.output_tokens !== undefined) breakdown.push(`${usage.output_tokens.toLocaleString()} out`)
  const cost = usage.cost_usd

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 text-[9px] text-muted-foreground/80",
        className,
      )}
      title={
        breakdown.length
          ? `${breakdown.join(" · ")}${cost ? ` · $${cost.toFixed(4)}` : ""}`
          : (estimateTotal ? "Client-side estimate (no usage payload yet)" : "Reported by backend")
      }
    >
      <Hash className="w-2.5 h-2.5" />
      {total.toLocaleString()} tok
      {cost !== undefined && cost > 0 && (
        <span className="ml-0.5">· ${cost.toFixed(cost < 0.01 ? 4 : 3)}</span>
      )}
    </span>
  )
}

// ─── Convenience: render an entire sidecar in one go ────────────────────────

import type { AISidecar } from "@/lib/ai-sidecar"

interface AISidecarBlocksProps {
  sidecar: AISidecar
  live?: boolean
  /** Where to put the usage chip — `inline` (after the answer) or `none`. */
  usagePlacement?: "inline" | "none"
  className?: string
}

/**
 * Renders the `Thoughts` + `Tools used` panels in the canonical order. Use
 * this in DM bubbles, /ask replies, etc. The usage chip is rendered
 * separately by the host because every surface places it slightly
 * differently (DM footer, canvas dock chip line, message-item meta row).
 */
export function AISidecarBlocks({
  sidecar, live, usagePlacement = "none", className,
}: AISidecarBlocksProps) {
  if (!sidecar.reasoning && !sidecar.tool_calls && !sidecar.usage) return null
  return (
    <div className={cn("space-y-1", className)}>
      {sidecar.reasoning && (
        <ReasoningPanel reasoning={sidecar.reasoning} live={live} />
      )}
      {sidecar.tool_calls && sidecar.tool_calls.length > 0 && (
        <ToolTimeline toolCalls={sidecar.tool_calls} />
      )}
      {usagePlacement === "inline" && sidecar.usage && (
        <UsageChip usage={sidecar.usage} />
      )}
    </div>
  )
}
