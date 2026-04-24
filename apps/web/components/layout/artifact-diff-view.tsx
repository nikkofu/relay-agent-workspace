"use client"

import { useMemo, useState } from "react"
import { ArtifactDiff, ArtifactVersion } from "@/stores/artifact-store"
import { ScrollArea } from "@/components/ui/scroll-area"
import { cn } from "@/lib/utils"
import { format, formatDistanceToNow } from "date-fns"
import { UserAvatar } from "@/components/common/user-avatar"
import { Rows, Columns } from "lucide-react"

interface ArtifactDiffViewProps {
  diff: ArtifactDiff
  // Optional version metadata (looked up from the versions list) so the diff
  // header can show timestamps + authors like a real Git/Office compare view.
  fromVersionMeta?: ArtifactVersion
  toVersionMeta?: ArtifactVersion
}

type ViewMode = "inline" | "side-by-side"

export function ArtifactDiffView({ diff, fromVersionMeta, toVersionMeta }: ArtifactDiffViewProps) {
  const [viewMode, setViewMode] = useState<ViewMode>("inline")
  // If we have spans, use them. Otherwise, fallback to parsing unifiedDiff.
  const renderContent = () => {
    if (diff.spans && diff.spans.length > 0) {
      return diff.spans.map((span, i) => (
        <div 
          key={i} 
          className={cn(
            "px-2 py-0.5 rounded whitespace-pre-wrap break-all relative group",
            span.kind === 'addition' && "bg-green-500/10 text-green-700 dark:text-green-400 border-l-2 border-green-500",
            span.kind === 'deletion' && "bg-red-500/10 text-red-700 dark:text-red-400 border-l-2 border-red-500 line-through opacity-70",
            span.kind === 'header' && "bg-blue-500/5 text-blue-500/50 my-2 font-bold py-1",
            span.kind === 'context' && "text-muted-foreground opacity-50"
          )}
        >
          {/* Line numbers if available */}
          {(span.fromLine || span.toLine) && (
            <span className="absolute right-2 top-0.5 text-[8px] text-muted-foreground/30 select-none hidden group-hover:inline">
              {span.fromLine && `L${span.fromLine}`}
              {span.fromLine && span.toLine && ' → '}
              {span.toLine && `L${span.toLine}`}
            </span>
          )}
          {span.content}
        </div>
      ))
    }

    // Fallback to unifiedDiff parsing
    const lines = diff.unifiedDiff.split('\n')
    return lines.map((line, i) => {
      const isAddition = line.startsWith('+') && !line.startsWith('+++')
      const isDeletion = line.startsWith('-') && !line.startsWith('---')
      const isHeader = line.startsWith('@@')
      const isFileInfo = line.startsWith('---') || line.startsWith('+++')

      if (isFileInfo) return null

      return (
        <div 
          key={i} 
          className={cn(
            "px-2 py-0.5 rounded whitespace-pre-wrap break-all",
            isAddition && "bg-green-500/10 text-green-700 dark:text-green-400 border-l-2 border-green-500",
            isDeletion && "bg-red-500/10 text-red-700 dark:text-red-400 border-l-2 border-red-500 line-through opacity-70",
            isHeader && "bg-blue-500/5 text-blue-500/50 my-2 font-bold py-1",
            !isAddition && !isDeletion && !isHeader && "text-muted-foreground opacity-50"
          )}
        >
          {line}
        </div>
      )
    })
  }
  
  // Split content by \n so side-by-side can render line-by-line with a faint
  // row-index gutter. Using plain text is safe here because spans already
  // encode the diff semantics and the source content is rendered verbatim.
  const fromLines = useMemo(() => (diff.fromContent || "").split("\n"), [diff.fromContent])
  const toLines   = useMemo(() => (diff.toContent || "").split("\n"), [diff.toContent])

  // Build a quick lookup of which lines changed on each side using the spans.
  // Keys are 1-indexed line numbers. Falls back to empty sets when spans are
  // absent; in that case side-by-side still renders the raw content without
  // per-line highlights.
  const { fromChangedLines, toChangedLines } = useMemo(() => {
    const fromSet = new Set<number>()
    const toSet = new Set<number>()
    if (diff.spans) {
      for (const s of diff.spans) {
        if (s.kind === "deletion" && s.fromLine) fromSet.add(s.fromLine)
        if (s.kind === "addition" && s.toLine) toSet.add(s.toLine)
      }
    }
    return { fromChangedLines: fromSet, toChangedLines: toSet }
  }, [diff.spans])

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-white dark:bg-[#1a1d21]">
      {/* Version timeline header — shows WHAT we're comparing, WHEN, and WHO */}
      <div className="border-b bg-muted/20 shrink-0">
        <div className="px-4 pt-3 pb-2 flex items-start justify-between gap-4">
          <div className="flex items-center gap-3 min-w-0 flex-1">
            <VersionBadge label="FROM" version={diff.fromVersion} meta={fromVersionMeta} tint="red" />
            <div className="text-muted-foreground/40 text-lg font-light shrink-0">→</div>
            <VersionBadge label="TO" version={diff.toVersion} meta={toVersionMeta} tint="green" />
          </div>
          <div className="flex items-center gap-3 shrink-0 pt-1">
            <div className="flex items-center gap-2">
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 rounded-full bg-green-500" />
                <span className="text-[10px] font-black text-green-600">+{diff.summary.added}</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 rounded-full bg-red-500" />
                <span className="text-[10px] font-black text-red-600">-{diff.summary.removed}</span>
              </div>
            </div>
            {/* View-mode toggle */}
            <div className="flex items-center rounded-md border overflow-hidden">
              <button
                type="button"
                onClick={() => setViewMode("inline")}
                className={cn(
                  "h-6 px-2 text-[10px] font-black uppercase tracking-widest flex items-center gap-1 transition-colors",
                  viewMode === "inline" ? "bg-purple-600 text-white" : "hover:bg-muted text-muted-foreground",
                )}
                title="Inline view"
              >
                <Rows className="w-3 h-3" />
                Inline
              </button>
              <button
                type="button"
                onClick={() => setViewMode("side-by-side")}
                className={cn(
                  "h-6 px-2 text-[10px] font-black uppercase tracking-widest flex items-center gap-1 transition-colors border-l",
                  viewMode === "side-by-side" ? "bg-purple-600 text-white" : "hover:bg-muted text-muted-foreground",
                )}
                title="Side-by-side view"
              >
                <Columns className="w-3 h-3" />
                Side-by-side
              </button>
            </div>
          </div>
        </div>
      </div>

      {viewMode === "inline" ? (
        <ScrollArea className="flex-1 font-mono text-xs">
          <div className="p-6 max-w-4xl mx-auto space-y-0.5">
            {renderContent()}
          </div>
        </ScrollArea>
      ) : (
        <div className="flex-1 min-h-0 grid grid-cols-2 divide-x">
          <SidePane
            label={`v${diff.fromVersion}`}
            sublabel={fromVersionMeta?.updatedAt ? safeFormatDate(fromVersionMeta.updatedAt) : undefined}
            tint="red"
            lines={fromLines}
            changedLines={fromChangedLines}
          />
          <SidePane
            label={`v${diff.toVersion}`}
            sublabel={toVersionMeta?.updatedAt ? safeFormatDate(toVersionMeta.updatedAt) : undefined}
            tint="green"
            lines={toLines}
            changedLines={toChangedLines}
          />
        </div>
      )}
    </div>
  )
}

// ── Subcomponents ────────────────────────────────────────────────────────────

function VersionBadge({
  label, version, meta, tint,
}: {
  label: "FROM" | "TO"
  version: number
  meta?: ArtifactVersion
  tint: "red" | "green"
}) {
  const tintClasses = tint === "red"
    ? "border-red-500/30 bg-red-500/5 text-red-700 dark:text-red-300"
    : "border-green-500/30 bg-green-500/5 text-green-700 dark:text-green-300"
  const labelClasses = tint === "red" ? "text-red-600" : "text-green-600"

  return (
    <div className={cn("flex items-center gap-2 rounded-lg border px-2.5 py-1.5 min-w-0", tintClasses)}>
      <span className={cn("text-[9px] font-black uppercase tracking-widest shrink-0", labelClasses)}>{label}</span>
      <div className="min-w-0">
        <div className="flex items-center gap-1.5">
          <span className="text-xs font-black whitespace-nowrap">v{version}</span>
          {meta?.updatedAt && (
            <span className="text-[9px] text-muted-foreground truncate" title={meta.updatedAt}>
              · {safeFormatDate(meta.updatedAt)} ({safeRelative(meta.updatedAt)})
            </span>
          )}
        </div>
        {meta?.updatedByUser && (
          <div className="flex items-center gap-1 mt-0.5">
            <UserAvatar src={meta.updatedByUser.avatar} name={meta.updatedByUser.name} className="h-3 w-3" />
            <span className="text-[9px] text-muted-foreground truncate">{meta.updatedByUser.name}</span>
          </div>
        )}
      </div>
    </div>
  )
}

function SidePane({
  label, sublabel, tint, lines, changedLines,
}: {
  label: string
  sublabel?: string
  tint: "red" | "green"
  lines: string[]
  changedLines: Set<number>
}) {
  const headerClasses = tint === "red"
    ? "bg-red-500/5 border-red-500/20 text-red-700 dark:text-red-300"
    : "bg-green-500/5 border-green-500/20 text-green-700 dark:text-green-300"
  const changedBg = tint === "red"
    ? "bg-red-500/10 text-red-700 dark:text-red-300 border-l-2 border-red-500"
    : "bg-green-500/10 text-green-700 dark:text-green-300 border-l-2 border-green-500"

  return (
    <div className="flex flex-col min-h-0">
      <div className={cn("border-b px-3 py-1.5 flex items-center justify-between gap-2 shrink-0", headerClasses)}>
        <span className="text-[10px] font-black uppercase tracking-widest">{label}</span>
        {sublabel && <span className="text-[9px] font-bold text-muted-foreground truncate">{sublabel}</span>}
      </div>
      <ScrollArea className="flex-1 font-mono text-xs">
        <div className="py-2">
          {lines.map((line, i) => {
            const lineNum = i + 1
            const changed = changedLines.has(lineNum)
            return (
              <div
                key={i}
                className={cn(
                  "flex items-start gap-2 px-2 py-0.5",
                  changed && changedBg,
                )}
              >
                <span className="text-[9px] text-muted-foreground/40 select-none shrink-0 w-8 text-right tabular-nums pt-0.5">
                  {lineNum}
                </span>
                <span className="whitespace-pre-wrap break-all flex-1">
                  {line || "\u00A0"}
                </span>
              </div>
            )
          })}
        </div>
      </ScrollArea>
    </div>
  )
}

function safeFormatDate(ts: string): string {
  try {
    const d = new Date(ts)
    if (isNaN(d.getTime())) return ""
    return format(d, "MMM d, HH:mm")
  } catch {
    return ""
  }
}

function safeRelative(ts: string): string {
  try {
    const d = new Date(ts)
    if (isNaN(d.getTime())) return ""
    return formatDistanceToNow(d, { addSuffix: true })
  } catch {
    return ""
  }
}
