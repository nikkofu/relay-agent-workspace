"use client"

// ── Phase 66 T09: Home Execution Blocks ──────────────────────────────────────
//
// Consumes the frozen Q4 contract top-level keys on GET /api/v1/home:
//   - open_list_work[]             : { item: WorkspaceListItem, channel_id, list_title }
//   - tool_runs_needing_attention[]: hydrated tool-run rows (status ∈ running|failed)
//   - channel_execution_pulse[]    : { channel_id, channel_name, open_item_count, overdue_count, summary }
//
// Cards deep-link using the IDs already present in each row (no synthetic
// client-side IDs — contract-correct). Each block renders nothing when empty
// so Home stays backward-compatible for existing consumers.

import { useRouter } from "next/navigation"
import {
  ListTodo, Terminal, Zap, ArrowRight, AlertCircle, Clock, CheckCircle2, XCircle,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { formatDistanceToNow } from "date-fns"

interface HomeExecutionBlocksProps {
  homeData: any
}

export function HomeExecutionBlocks({ homeData }: HomeExecutionBlocksProps) {
  const router = useRouter()

  const openWork: any[] = homeData?.open_list_work || []
  const toolRuns: any[] = homeData?.tool_runs_needing_attention || []
  const pulse: any[] = homeData?.channel_execution_pulse || []

  const hasAny = openWork.length > 0 || toolRuns.length > 0 || pulse.length > 0
  if (!hasAny) return null

  return (
    <div className="space-y-5">
      <h3 className="text-lg font-black tracking-tight flex items-center gap-2">
        <Zap className="w-5 h-5 text-violet-600" />
        Channel Execution
      </h3>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        {/* Open List Work */}
        {openWork.length > 0 && (
          <ExecutionCard
            title="Open List Work"
            subtitle={`${openWork.length} open`}
            icon={<ListTodo className="w-4 h-4" />}
            tint="violet"
          >
            <div className="space-y-1.5">
              {openWork.slice(0, 5).map((row: any) => {
                const item = row.item || row
                const dueAgo = item?.due_at ? safeRelative(item.due_at) : null
                const isOverdue = dueAgo && new Date(item.due_at).getTime() < Date.now()
                return (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => row.channel_id && router.push(`/workspace?c=${row.channel_id}`)}
                    className="w-full text-left px-2 py-1.5 rounded hover:bg-violet-500/10 transition-colors group"
                  >
                    <div className="flex items-center gap-1.5">
                      <p className="text-[11px] font-semibold truncate flex-1 group-hover:text-violet-700 dark:group-hover:text-violet-300">
                        {item.content || item.title || "Untitled item"}
                      </p>
                      {isOverdue && (
                        <span className="text-[8px] font-black uppercase tracking-widest px-1 py-0.5 rounded bg-rose-500/10 text-rose-700 dark:text-rose-300 shrink-0">
                          Overdue
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-1.5 mt-0.5">
                      {row.list_title && (
                        <span className="text-[9px] text-muted-foreground/80 uppercase tracking-widest truncate">
                          {row.list_title}
                        </span>
                      )}
                      {dueAgo && (
                        <span className={cn(
                          "text-[9px] uppercase tracking-widest shrink-0",
                          isOverdue ? "text-rose-600" : "text-muted-foreground/60",
                        )}>
                          <Clock className="w-2 h-2 inline mr-0.5" />
                          {dueAgo}
                        </span>
                      )}
                    </div>
                  </button>
                )
              })}
            </div>
          </ExecutionCard>
        )}

        {/* Tool Runs Needing Attention */}
        {toolRuns.length > 0 && (
          <ExecutionCard
            title="Tool Runs Needing Attention"
            subtitle={`${toolRuns.length} active`}
            icon={<Terminal className="w-4 h-4" />}
            tint="amber"
          >
            <div className="space-y-1.5">
              {toolRuns.slice(0, 5).map((run: any) => {
                const status = run.status as string
                const StatusIcon = status === "failed" ? XCircle
                  : status === "running" ? Clock
                  : status === "success" ? CheckCircle2
                  : AlertCircle
                const statusColor = status === "failed" ? "text-rose-600"
                  : status === "running" ? "text-amber-600"
                  : status === "success" ? "text-emerald-600"
                  : "text-muted-foreground"
                const startedAgo = safeRelative(run.started_at || run.startedAt)
                return (
                  <button
                    key={run.id}
                    type="button"
                    onClick={() => run.channel_id && router.push(`/workspace?c=${run.channel_id}`)}
                    className="w-full text-left px-2 py-1.5 rounded hover:bg-amber-500/10 transition-colors group"
                  >
                    <div className="flex items-center gap-1.5">
                      <StatusIcon className={cn("w-3 h-3 shrink-0", statusColor, status === "running" && "animate-pulse")} />
                      <p className="text-[11px] font-semibold truncate flex-1 group-hover:text-amber-700 dark:group-hover:text-amber-300">
                        {run.tool_name || run.tool_key || "Tool run"}
                      </p>
                    </div>
                    {startedAgo && (
                      <p className="text-[9px] text-muted-foreground/70 uppercase tracking-widest pl-4 mt-0.5">
                        {startedAgo}
                      </p>
                    )}
                  </button>
                )
              })}
            </div>
          </ExecutionCard>
        )}

        {/* Channel Execution Pulse */}
        {pulse.length > 0 && (
          <ExecutionCard
            title="Channel Execution Pulse"
            subtitle={`${pulse.length} active`}
            icon={<Zap className="w-4 h-4" />}
            tint="sky"
          >
            <div className="space-y-1.5">
              {pulse.slice(0, 5).map((row: any) => (
                <button
                  key={row.channel_id}
                  type="button"
                  onClick={() => row.channel_id && router.push(`/workspace?c=${row.channel_id}`)}
                  className="w-full text-left px-2 py-1.5 rounded hover:bg-sky-500/10 transition-colors group"
                >
                  <div className="flex items-center justify-between gap-1.5">
                    <p className="text-[11px] font-semibold truncate group-hover:text-sky-700 dark:group-hover:text-sky-300">
                      #{row.channel_name}
                    </p>
                    {row.overdue_count > 0 && (
                      <span className="text-[8px] font-black uppercase tracking-widest px-1 py-0.5 rounded bg-rose-500/10 text-rose-700 dark:text-rose-300 shrink-0">
                        {row.overdue_count} late
                      </span>
                    )}
                  </div>
                  {row.summary && (
                    <p className="text-[9px] text-muted-foreground/80 leading-snug mt-0.5 line-clamp-2">
                      {row.summary}
                    </p>
                  )}
                </button>
              ))}
            </div>
          </ExecutionCard>
        )}
      </div>
    </div>
  )
}

function ExecutionCard({
  title, subtitle, icon, tint, children,
}: {
  title: string
  subtitle: string
  icon: React.ReactNode
  tint: "violet" | "amber" | "sky"
  children: React.ReactNode
}) {
  const tintMap = {
    violet: { bg: "bg-violet-500/5", border: "border-violet-500/20", text: "text-violet-700 dark:text-violet-300", iconBg: "bg-violet-500/10" },
    amber:  { bg: "bg-amber-500/5",  border: "border-amber-500/20",  text: "text-amber-700 dark:text-amber-300",  iconBg: "bg-amber-500/10" },
    sky:    { bg: "bg-sky-500/5",    border: "border-sky-500/20",    text: "text-sky-700 dark:text-sky-300",      iconBg: "bg-sky-500/10" },
  }
  const t = tintMap[tint]
  return (
    <div className={cn("rounded-2xl border p-3 space-y-2", t.bg, t.border)}>
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2 min-w-0">
          <div className={cn("w-7 h-7 rounded-lg flex items-center justify-center shrink-0", t.iconBg, t.text)}>
            {icon}
          </div>
          <div className="min-w-0">
            <p className={cn("text-[10px] font-black uppercase tracking-widest truncate", t.text)}>{title}</p>
            <p className="text-[9px] text-muted-foreground uppercase tracking-widest">{subtitle}</p>
          </div>
        </div>
      </div>
      {children}
    </div>
  )
}

function safeRelative(ts: string | undefined | null): string | null {
  if (!ts) return null
  try {
    const d = new Date(ts)
    if (isNaN(d.getTime())) return null
    return formatDistanceToNow(d, { addSuffix: true })
  } catch {
    return null
  }
}
