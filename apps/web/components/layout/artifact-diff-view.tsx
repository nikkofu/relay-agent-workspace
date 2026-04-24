"use client"

// ── Artifact Diff View ───────────────────────────────────────────────────────
//
// Three-way view onto an artifact version comparison:
//
//   1. Review  — Microsoft Word–style word-level inline diff. HTML content
//      is flattened to plain text (so `<p>` / `<br>` / tag bleed never
//      leaks into the diff — request #4) and then Myers-LCS'd at the word
//      level. Insertions render underlined green; deletions render
//      strikethrough red — same visual grammar as Office's Review pane.
//
//   2. Inline — the previous server-provided span/unifiedDiff view, useful
//      for code-type artifacts where line granularity matters.
//
//   3. Side-by-side — two panes with synchronized scrolling, each showing
//      the version's plain-text content with changed lines highlighted.
//
// All three modes share a version-timeline header with author + timestamp
// chips so the user always knows WHAT they're comparing, WHEN, and WHO.

import { useMemo, useState } from "react"
import { ArtifactDiff, ArtifactVersion } from "@/stores/artifact-store"
import { ScrollArea } from "@/components/ui/scroll-area"
import { cn } from "@/lib/utils"
import { format, formatDistanceToNow } from "date-fns"
import { UserAvatar } from "@/components/common/user-avatar"
import { Rows, Columns, FileDiff } from "lucide-react"
import { diffWords, htmlToPlainText, type DiffToken } from "@/lib/word-diff"

interface ArtifactDiffViewProps {
  diff: ArtifactDiff
  fromVersionMeta?: ArtifactVersion
  toVersionMeta?: ArtifactVersion
}

type ViewMode = "review" | "inline" | "side-by-side"

export function ArtifactDiffView({ diff, fromVersionMeta, toVersionMeta }: ArtifactDiffViewProps) {
  // Default to the word-level Review pane — that is what most users reach
  // for when comparing a doc; the spans view is a power-user fallback.
  const [viewMode, setViewMode] = useState<ViewMode>("review")

  // HTML → plain text BEFORE comparing. Without this, TipTap-produced
  // content leaks `<p>` / `</p>` / `<br/>` into the rendered diff, which
  // is exactly the "literal <p></p>" regression the user reported.
  const fromText = useMemo(() => htmlToPlainText(diff.fromContent || ""), [diff.fromContent])
  const toText   = useMemo(() => htmlToPlainText(diff.toContent   || ""), [diff.toContent])

  const reviewTokens = useMemo(() => diffWords(fromText, toText), [fromText, toText])
  const reviewStats = useMemo(() => {
    let add = 0, del = 0
    for (const t of reviewTokens) {
      if (t.kind === "insert") add += wordCount(t.text)
      else if (t.kind === "delete") del += wordCount(t.text)
    }
    return { add, del }
  }, [reviewTokens])

  const fromLines = useMemo(() => fromText.split("\n"), [fromText])
  const toLines   = useMemo(() => toText.split("\n"),   [toText])

  // Build line-change sets from the spans the server returned (when
  // available) so side-by-side still gets per-line tinting — the server
  // payload is the authoritative source for line-level changes.
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

  // The inline "server span" renderer — we keep this for parity so code
  // artifacts (which arrive as unified diffs from the backend) still get
  // their semantic line-by-line view.
  const renderServerSpans = () => {
    if (diff.spans && diff.spans.length > 0) {
      return diff.spans.map((span, i) => (
        <div
          key={i}
          className={cn(
            "px-2 py-0.5 rounded whitespace-pre-wrap break-all relative group",
            span.kind === "addition" && "bg-green-500/10 text-green-700 dark:text-green-400 border-l-2 border-green-500",
            span.kind === "deletion" && "bg-red-500/10 text-red-700 dark:text-red-400 border-l-2 border-red-500 line-through opacity-70",
            span.kind === "header" && "bg-blue-500/5 text-blue-500/50 my-2 font-bold py-1",
            span.kind === "context" && "text-muted-foreground opacity-50",
          )}
        >
          {(span.fromLine || span.toLine) && (
            <span className="absolute right-2 top-0.5 text-[8px] text-muted-foreground/30 select-none hidden group-hover:inline">
              {span.fromLine && `L${span.fromLine}`}
              {span.fromLine && span.toLine && " → "}
              {span.toLine && `L${span.toLine}`}
            </span>
          )}
          {span.content}
        </div>
      ))
    }
    const lines = diff.unifiedDiff.split("\n")
    return lines.map((line, i) => {
      const isAddition = line.startsWith("+") && !line.startsWith("+++")
      const isDeletion = line.startsWith("-") && !line.startsWith("---")
      const isHeader   = line.startsWith("@@")
      const isFileInfo = line.startsWith("---") || line.startsWith("+++")
      if (isFileInfo) return null
      return (
        <div
          key={i}
          className={cn(
            "px-2 py-0.5 rounded whitespace-pre-wrap break-all",
            isAddition && "bg-green-500/10 text-green-700 dark:text-green-400 border-l-2 border-green-500",
            isDeletion && "bg-red-500/10 text-red-700 dark:text-red-400 border-l-2 border-red-500 line-through opacity-70",
            isHeader && "bg-blue-500/5 text-blue-500/50 my-2 font-bold py-1",
            !isAddition && !isDeletion && !isHeader && "text-muted-foreground opacity-50",
          )}
        >
          {line}
        </div>
      )
    })
  }

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-white dark:bg-[#1a1d21]">
      {/* Version timeline header */}
      <div className="border-b bg-muted/20 shrink-0">
        <div className="px-4 pt-3 pb-2 flex items-start justify-between gap-4 flex-wrap">
          <div className="flex items-center gap-3 min-w-0 flex-1">
            <VersionBadge label="FROM" version={diff.fromVersion} meta={fromVersionMeta} tint="red" />
            <div className="text-muted-foreground/40 text-lg font-light shrink-0">→</div>
            <VersionBadge label="TO" version={diff.toVersion} meta={toVersionMeta} tint="green" />
          </div>
          <div className="flex items-center gap-3 shrink-0 pt-1">
            <div className="flex items-center gap-2">
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 rounded-full bg-green-500" />
                <span className="text-[10px] font-black text-green-600">
                  +{viewMode === "review" ? reviewStats.add : diff.summary.added}
                </span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 rounded-full bg-red-500" />
                <span className="text-[10px] font-black text-red-600">
                  -{viewMode === "review" ? reviewStats.del : diff.summary.removed}
                </span>
              </div>
              {viewMode === "review" && (
                <span className="text-[9px] uppercase font-bold tracking-widest text-muted-foreground">words</span>
              )}
            </div>
            {/* View-mode switcher */}
            <div className="flex items-center rounded-md border overflow-hidden">
              <ModeButton active={viewMode === "review"}  onClick={() => setViewMode("review")}  icon={<FileDiff className="w-3 h-3" />} label="Review" />
              <ModeButton active={viewMode === "inline"}  onClick={() => setViewMode("inline")}  icon={<Rows className="w-3 h-3" />}     label="Inline" border />
              <ModeButton active={viewMode === "side-by-side"} onClick={() => setViewMode("side-by-side")} icon={<Columns className="w-3 h-3" />} label="Side" border />
            </div>
          </div>
        </div>
      </div>

      {viewMode === "review" ? (
        <ScrollArea className="flex-1">
          <div className="p-8 max-w-3xl mx-auto">
            <div className="prose dark:prose-invert max-w-none text-[15px] leading-relaxed whitespace-pre-wrap break-words">
              {reviewTokens.length === 0 ? (
                <p className="text-muted-foreground italic text-sm">No textual differences detected.</p>
              ) : (
                reviewTokens.map((t, i) => <ReviewToken key={i} token={t} />)
              )}
            </div>
            <ReviewLegend />
          </div>
        </ScrollArea>
      ) : viewMode === "inline" ? (
        <ScrollArea className="flex-1 font-mono text-xs">
          <div className="p-6 max-w-4xl mx-auto space-y-0.5">
            {renderServerSpans()}
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

// ── Review (Word-style) rendering ────────────────────────────────────────────

function ReviewToken({ token }: { token: DiffToken }) {
  if (token.kind === "equal") {
    return <span>{token.text}</span>
  }
  if (token.kind === "insert") {
    return (
      <span className="bg-green-500/15 text-green-800 dark:text-green-300 underline decoration-green-500 decoration-1 underline-offset-2 px-[2px] rounded-[3px]">
        {token.text}
      </span>
    )
  }
  return (
    <span className="bg-red-500/10 text-red-700 dark:text-red-300 line-through decoration-red-500/60 decoration-1 px-[2px] rounded-[3px]">
      {token.text}
    </span>
  )
}

function ReviewLegend() {
  return (
    <div className="mt-10 pt-4 border-t flex items-center gap-4 text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
      <span className="flex items-center gap-1.5">
        <span className="inline-block w-3 h-3 rounded-sm bg-green-500/20 border border-green-500/40" />
        Insertion
      </span>
      <span className="flex items-center gap-1.5">
        <span className="inline-block w-3 h-3 rounded-sm bg-red-500/15 border border-red-500/40" />
        Deletion
      </span>
      <span className="ml-auto italic font-normal normal-case tracking-normal">
        Word-level diff · HTML markup stripped for readability
      </span>
    </div>
  )
}

// ── Subcomponents ────────────────────────────────────────────────────────────

function ModeButton({ active, onClick, icon, label, border }: {
  active: boolean
  onClick: () => void
  icon: React.ReactNode
  label: string
  border?: boolean
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "h-6 px-2 text-[10px] font-black uppercase tracking-widest flex items-center gap-1 transition-colors",
        active ? "bg-purple-600 text-white" : "hover:bg-muted text-muted-foreground",
        border && "border-l",
      )}
      title={`${label} view`}
    >
      {icon}
      {label}
    </button>
  )
}

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

// ── Helpers ──────────────────────────────────────────────────────────────────

function wordCount(s: string): number {
  const m = s.match(/[A-Za-z0-9\u00C0-\uFFFF_]+/g)
  return m ? m.length : 0
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
