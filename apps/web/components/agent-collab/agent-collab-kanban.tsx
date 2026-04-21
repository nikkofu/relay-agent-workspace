"use client"

import { useState, useMemo } from "react"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { CheckCircle2, Clock, Loader, ChevronDown, ChevronRight, Search, Filter } from "lucide-react"
import { TASKS, MEMBER_MAP, TASK_TYPE_META, type Task, type TaskStatus, type TaskType } from "./agent-collab-data"

// ─── Types & Config ───────────────────────────────────────────────────────────

const COLUMN_META: Record<TaskStatus, { label: string; icon: React.ElementType; color: string; dotColor: string; headerBg: string }> = {
  'ready':      { label: 'Ready',       icon: Clock,         color: 'text-amber-600 dark:text-amber-400', dotColor: 'bg-amber-400',  headerBg: 'bg-amber-500/8 border-amber-500/20' },
  'in-progress':{ label: 'In Progress', icon: Loader,        color: 'text-blue-600 dark:text-blue-400',   dotColor: 'bg-blue-400 animate-pulse', headerBg: 'bg-blue-500/8 border-blue-500/20' },
  'done':       { label: 'Done',        icon: CheckCircle2,  color: 'text-emerald-600 dark:text-emerald-400', dotColor: 'bg-emerald-400', headerBg: 'bg-emerald-500/8 border-emerald-500/20' },
}

// ─── Task Card ────────────────────────────────────────────────────────────────

function TaskCard({ task }: { task: Task }) {
  const typeMeta = TASK_TYPE_META[task.type]
  const [expanded, setExpanded] = useState(false)

  return (
    <div
      className={cn(
        'bg-white dark:bg-[#222529] rounded-xl border border-border/60 p-3 space-y-2.5 cursor-pointer hover:shadow-md hover:-translate-y-px transition-all duration-150 group',
        task.status === 'ready' && 'border-l-2 border-l-amber-400'
      )}
      onClick={() => setExpanded(e => !e)}
    >
      {/* Header Row */}
      <div className="flex items-start gap-2">
        <div className={cn('w-4 h-4 shrink-0 mt-0.5', COLUMN_META[task.status].color)}>
          {task.status === 'done'       && <CheckCircle2 className="w-4 h-4" />}
          {task.status === 'ready'      && <Clock className="w-4 h-4" />}
          {task.status === 'in-progress'&& <Loader className="w-4 h-4 animate-spin" />}
        </div>
        <p className={cn(
          'text-xs font-bold leading-snug flex-1',
          task.status === 'done' && 'text-muted-foreground line-through decoration-emerald-500/50',
          task.status === 'ready' && 'text-foreground'
        )}>{task.task}</p>
        <span className="text-muted-foreground opacity-0 group-hover:opacity-60 transition-opacity">
          {expanded ? <ChevronDown className="w-3 h-3" /> : <ChevronRight className="w-3 h-3" />}
        </span>
      </div>

      {/* Phase + Type + Date */}
      <div className="flex items-center gap-1.5 flex-wrap">
        <span className="text-[10px] font-black text-muted-foreground bg-muted/60 border border-border/30 rounded px-1.5 py-0.5">
          P{task.phase}
        </span>
        <span className={cn('text-[10px] font-black px-1.5 py-0.5 rounded border', typeMeta.color)}>
          {typeMeta.label}
        </span>
        <span className="text-[10px] text-muted-foreground ml-auto">{task.deadline}</span>
      </div>

      {/* Assignees */}
      <div className="flex items-center gap-1">
        {task.assignedTo.map(name => {
          const m = MEMBER_MAP[name]
          return (
            <div key={name} className={cn('w-5 h-5 rounded-md flex items-center justify-center text-[9px] font-black text-white shrink-0', m?.bgClass ?? 'bg-slate-500')} title={name}>
              {m?.avatar ?? name[0]}
            </div>
          )
        })}
        <span className="text-[10px] text-muted-foreground ml-1">{task.assignedTo.join(', ')}</span>
      </div>

      {/* Expanded Description */}
      {expanded && (
        <p className="text-[11px] text-muted-foreground leading-relaxed border-t border-border/40 pt-2 mt-1">
          {task.description}
        </p>
      )}
    </div>
  )
}

// ─── Column ───────────────────────────────────────────────────────────────────

function KanbanColumn({ status, tasks }: { status: TaskStatus; tasks: Task[] }) {
  const meta = COLUMN_META[status]
  const [collapsed, setCollapsed] = useState(status === 'done')

  const byDate = useMemo(() => {
    const map: Record<string, Task[]> = {}
    for (const t of tasks) {
      if (!map[t.deadline]) map[t.deadline] = []
      map[t.deadline].push(t)
    }
    return Object.entries(map).sort(([a], [b]) => b.localeCompare(a))
  }, [tasks])

  return (
    <div className="flex flex-col min-w-[300px] max-w-[340px] shrink-0">
      {/* Column Header */}
      <div className={cn('rounded-2xl border px-4 py-3 mb-3 flex items-center gap-2', meta.headerBg)}>
        <div className={cn('w-2 h-2 rounded-full shrink-0', meta.dotColor)} />
        <meta.icon className={cn('w-4 h-4', meta.color)} />
        <span className={cn('text-xs font-black uppercase tracking-wider', meta.color)}>{meta.label}</span>
        <Badge className="ml-auto h-5 px-1.5 text-[10px] font-black bg-white/50 dark:bg-black/20 border-border/30 text-foreground">
          {tasks.length}
        </Badge>
        {status === 'done' && (
          <button onClick={() => setCollapsed(c => !c)} className="text-muted-foreground hover:text-foreground transition-colors ml-1">
            {collapsed ? <ChevronRight className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
          </button>
        )}
      </div>

      {/* Cards */}
      {!collapsed && (
        <div className="flex flex-col gap-2 overflow-y-auto pr-1" style={{ maxHeight: 'calc(100vh - 280px)' }}>
          {status === 'done' ? (
            byDate.map(([date, dateTasks]) => (
              <DateGroup key={date} date={date} tasks={dateTasks} />
            ))
          ) : (
            tasks.map(t => <TaskCard key={t.id} task={t} />)
          )}
        </div>
      )}

      {collapsed && (
        <button
          onClick={() => setCollapsed(false)}
          className="rounded-xl border border-dashed border-border/40 px-4 py-6 text-xs font-bold text-muted-foreground hover:text-foreground hover:border-border/60 transition-colors text-center"
        >
          Show {tasks.length} completed tasks →
        </button>
      )}
    </div>
  )
}

// ─── Date Group (for Done column) ─────────────────────────────────────────────

function DateGroup({ date, tasks }: { date: string; tasks: Task[] }) {
  const [open, setOpen] = useState(date >= '2026-04-21')
  const label = new Date(date).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })

  return (
    <div className="space-y-1.5">
      <button
        onClick={() => setOpen(o => !o)}
        className="w-full flex items-center gap-2 px-2 py-1 text-[10px] font-black uppercase tracking-wider text-muted-foreground hover:text-foreground transition-colors rounded-lg hover:bg-muted/40"
      >
        {open ? <ChevronDown className="w-3 h-3" /> : <ChevronRight className="w-3 h-3" />}
        {label}
        <span className="ml-auto">{tasks.length}</span>
      </button>
      {open && tasks.map(t => <TaskCard key={t.id} task={t} />)}
    </div>
  )
}

// ─── Filters ──────────────────────────────────────────────────────────────────

type FilterState = { assignee: string; type: TaskType | 'all'; search: string }

function FilterBar({ filters, onChange }: { filters: FilterState; onChange: (f: FilterState) => void }) {
  const assignees = ['All', 'Gemini', 'Codex', 'Nikko Fu']
  const types: Array<TaskType | 'all'> = ['all', 'api', 'frontend', 'infra', 'ux']

  return (
    <div className="flex items-center gap-3 flex-wrap mb-5">
      <div className="flex items-center gap-2 bg-muted/40 border border-border/40 rounded-xl px-3 h-9 flex-1 min-w-[200px] max-w-xs">
        <Search className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
        <input
          type="text"
          placeholder="Search tasks..."
          value={filters.search}
          onChange={e => onChange({ ...filters, search: e.target.value })}
          className="bg-transparent text-xs flex-1 outline-none placeholder:text-muted-foreground"
        />
      </div>
      <div className="flex items-center gap-1">
        <Filter className="w-3.5 h-3.5 text-muted-foreground" />
        {assignees.map(a => (
          <button
            key={a}
            onClick={() => onChange({ ...filters, assignee: a === 'All' ? '' : a })}
            className={cn(
              'text-[10px] font-black px-2.5 h-7 rounded-lg border transition-all',
              (a === 'All' ? !filters.assignee : filters.assignee === a)
                ? 'bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/30'
                : 'text-muted-foreground border-transparent hover:border-border/40 hover:text-foreground'
            )}
          >{a}</button>
        ))}
      </div>
      <div className="flex items-center gap-1">
        {types.map(t => (
          <button
            key={t}
            onClick={() => onChange({ ...filters, type: t })}
            className={cn(
              'text-[10px] font-black px-2.5 h-7 rounded-lg border transition-all capitalize',
              filters.type === t
                ? t === 'all'
                  ? 'bg-slate-500/10 text-foreground border-slate-500/30'
                  : cn(TASK_TYPE_META[t as TaskType]?.color ?? '')
                : 'text-muted-foreground border-transparent hover:border-border/40 hover:text-foreground'
            )}
          >{t}</button>
        ))}
      </div>
    </div>
  )
}

// ─── Main Kanban ──────────────────────────────────────────────────────────────

export function AgentCollabKanban() {
  const [filters, setFilters] = useState<FilterState>({ assignee: '', type: 'all', search: '' })

  const filtered = useMemo(() => {
    return TASKS.filter(t => {
      if (filters.assignee && !t.assignedTo.includes(filters.assignee)) return false
      if (filters.type !== 'all' && t.type !== filters.type) return false
      if (filters.search) {
        const q = filters.search.toLowerCase()
        if (!t.task.toLowerCase().includes(q) && !t.description.toLowerCase().includes(q)) return false
      }
      return true
    })
  }, [filters])

  const columns: TaskStatus[] = ['ready', 'in-progress', 'done']

  return (
    <div className="space-y-2">
      <FilterBar filters={filters} onChange={setFilters} />
      <div className="flex gap-4 overflow-x-auto pb-4 items-start">
        {columns.map(status => (
          <KanbanColumn
            key={status}
            status={status}
            tasks={filtered.filter(t => t.status === status)}
          />
        ))}
      </div>
    </div>
  )
}
