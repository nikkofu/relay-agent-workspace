"use client"

import { useState, useEffect, useMemo } from "react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"
import { cn } from "@/lib/utils"
import {
  Hash, Users, CheckCircle2, Clock, Zap, Activity,
  Columns, MessageSquare, BarChart3, UserCircle2,
  TrendingUp, Calendar, Star
} from "lucide-react"
import { TASKS, MEMBERS, ACTIVE_SUPERPOWERS, MEMBER_MAP, TASK_TYPE_META, type Task, type Member, type AgentPower } from "./agent-collab-data"
import { useCollabStore } from "@/stores/collab-store"
import { AgentCollabKanban } from "./agent-collab-kanban"
import { AgentCollabCommLog } from "./agent-collab-comm-log"

type Tab = 'overview' | 'kanban' | 'comms' | 'stats'

const TABS: { id: Tab; label: string; icon: React.ElementType }[] = [
  { id: 'overview', label: 'Overview',   icon: Activity },
  { id: 'kanban',   label: 'Kanban',     icon: Columns },
  { id: 'comms',    label: 'Comm Log',   icon: MessageSquare },
  { id: 'stats',    label: 'Statistics', icon: BarChart3 },
]

// ─── Utility ─────────────────────────────────────────────────────────────────

function MemberAvatar({ name, size = 'md' }: { name: string; size?: 'sm' | 'md' | 'lg' }) {
  const member = MEMBER_MAP[name]
  const bg = member?.bgClass ?? 'bg-slate-500'
  const initials = member?.avatar ?? name.substring(0, 2).toUpperCase()
  const sz = size === 'sm' ? 'w-6 h-6 text-[9px]' : size === 'lg' ? 'w-12 h-12 text-base' : 'w-8 h-8 text-xs'
  return (
    <div className={cn('rounded-lg flex items-center justify-center font-black text-white shrink-0 ring-2 ring-white/20', bg, sz)}>
      {initials}
    </div>
  )
}

// ─── Sub-components for Overview ─────────────────────────────────────────────

function MemberCard({ member, power }: { member: Member; power?: AgentPower }) {
  const tasks = TASKS.filter(t => t.assignedTo.includes(member.name))
  const done = tasks.filter(t => t.status === 'done').length
  const ready = tasks.filter(t => t.status === 'ready').length

  return (
    <div className="bg-white dark:bg-[#222529] rounded-2xl border border-border/60 p-5 flex flex-col gap-4 hover:shadow-lg hover:-translate-y-0.5 transition-all duration-200">
      <div className="flex items-start gap-3">
        <div className={cn('w-12 h-12 rounded-2xl flex items-center justify-center font-black text-white text-base shrink-0 shadow-md', member.bgClass)}>
          {member.avatar}
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="font-black text-sm">{member.name}</span>
            {member.type === 'ai' ? (
              <Badge className="text-[9px] h-4 px-1.5 bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20 font-bold">AI</Badge>
            ) : (
              <Badge className="text-[9px] h-4 px-1.5 bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/20 font-bold">Human</Badge>
            )}
          </div>
          <p className="text-[11px] text-muted-foreground font-semibold mt-0.5">{member.role}</p>
        </div>
      </div>

      <p className="text-xs text-muted-foreground leading-relaxed">{member.specialty}</p>

      <div className="flex flex-wrap gap-1.5">
        {member.tools.map(t => (
          <span key={t} className="text-[10px] font-mono bg-muted/60 border border-border/40 rounded px-1.5 py-0.5 text-muted-foreground">{t}</span>
        ))}
      </div>

      <div className="grid grid-cols-3 gap-2 pt-1 border-t border-border/40">
        <div className="text-center">
          <p className="text-lg font-black">{tasks.length}</p>
          <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-wider">Tasks</p>
        </div>
        <div className="text-center">
          <p className="text-lg font-black text-emerald-600">{done}</p>
          <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-wider">Done</p>
        </div>
        <div className="text-center">
          <p className="text-lg font-black text-amber-500">{ready}</p>
          <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-wider">Ready</p>
        </div>
      </div>

      {power && (
        <div className="bg-muted/40 rounded-xl p-3 flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <span className="text-[10px] font-black uppercase tracking-wider text-muted-foreground">Active Superpower</span>
            <div className={cn(
              'w-2 h-2 rounded-full',
              power.status === 'active' ? 'bg-green-500 animate-pulse' :
              power.status === 'done'   ? 'bg-emerald-500' :
              power.status === 'thinking' ? 'bg-amber-500 animate-pulse' : 'bg-slate-400'
            )} />
          </div>
          <span className="text-[11px] font-mono text-foreground/80 bg-background/60 rounded px-2 py-0.5 border border-border/30">{power.skill}</span>
          <p className="text-[11px] text-muted-foreground line-clamp-2">{power.task}</p>
          <Progress value={power.progress} className="h-1" />
          <span className="text-[10px] text-muted-foreground text-right">{power.progress}%</span>
        </div>
      )}
    </div>
  )
}

function StatsBar() {
  const total = TASKS.length
  const done = TASKS.filter(t => t.status === 'done').length
  const ready = TASKS.filter(t => t.status === 'ready').length
  const inProgress = TASKS.filter(t => t.status === 'in-progress').length
  const pct = Math.round((done / total) * 100)

  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
      {[
        { label: 'Total Tasks', value: total, icon: CheckCircle2, color: 'text-foreground', bg: 'bg-muted/50' },
        { label: 'Done',        value: done,  icon: Star,         color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-500/10' },
        { label: 'Ready',       value: ready, icon: Clock,        color: 'text-amber-600 dark:text-amber-400',     bg: 'bg-amber-500/10' },
        { label: 'In Progress', value: inProgress, icon: Zap,     color: 'text-blue-600 dark:text-blue-400',       bg: 'bg-blue-500/10' },
      ].map(s => (
        <div key={s.label} className={cn('rounded-2xl p-4 flex items-center gap-4 border border-border/40', s.bg)}>
          <s.icon className={cn('w-8 h-8 shrink-0', s.color)} strokeWidth={1.5} />
          <div>
            <p className="text-2xl font-black">{s.value}</p>
            <p className="text-[11px] font-black uppercase tracking-wider text-muted-foreground">{s.label}</p>
          </div>
        </div>
      ))}
      <div className="col-span-2 md:col-span-4 bg-white dark:bg-[#222529] rounded-2xl border border-border/60 p-4 flex items-center gap-4">
        <TrendingUp className="w-5 h-5 text-purple-600 shrink-0" />
        <div className="flex-1">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-black">Overall Completion</span>
            <span className="text-sm font-black text-purple-600">{pct}%</span>
          </div>
          <Progress value={pct} className="h-2.5" />
        </div>
      </div>
    </div>
  )
}

function DateTimeline() {
  const byDate = TASKS.reduce<Record<string, Task[]>>((acc, t) => {
    if (!acc[t.deadline]) acc[t.deadline] = []
    acc[t.deadline].push(t)
    return acc
  }, {})

  const dates = Object.keys(byDate).sort()

  return (
    <div className="space-y-4">
      <h3 className="text-sm font-black uppercase tracking-wider text-muted-foreground flex items-center gap-2">
        <Calendar className="w-4 h-4" /> Phase Timeline
      </h3>
      <div className="relative pl-5">
        <div className="absolute left-2 top-2 bottom-2 w-px bg-gradient-to-b from-purple-500 via-blue-500 to-emerald-500 opacity-30" />
        <div className="flex flex-col gap-4">
          {dates.map(date => {
            const dayTasks = byDate[date]
            const doneCnt = dayTasks.filter(t => t.status === 'done').length
            const readyCnt = dayTasks.filter(t => t.status === 'ready').length
            const formattedDate = new Date(date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
            return (
              <div key={date} className="flex gap-4 items-start">
                <div className="w-3 h-3 rounded-full bg-purple-500 border-2 border-white dark:border-[#1a1d21] shrink-0 mt-1 z-10" />
                <div className="flex-1 bg-white dark:bg-[#222529] rounded-xl border border-border/50 p-3 hover:shadow-md transition-shadow">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-xs font-black text-purple-600 dark:text-purple-400">{formattedDate}</span>
                    <div className="flex items-center gap-2">
                      {doneCnt > 0 && <span className="text-[10px] font-bold text-emerald-600">✓ {doneCnt} done</span>}
                      {readyCnt > 0 && <span className="text-[10px] font-bold text-amber-500">⏳ {readyCnt} ready</span>}
                    </div>
                  </div>
                  <div className="flex flex-wrap gap-1">
                    {dayTasks.slice(0, 6).map(t => (
                      <span key={t.id} className={cn(
                        'text-[10px] font-semibold px-1.5 py-0.5 rounded border',
                        TASK_TYPE_META[t.type].color
                      )}>{TASK_TYPE_META[t.type].label}</span>
                    ))}
                    {dayTasks.length > 6 && (
                      <span className="text-[10px] text-muted-foreground font-semibold px-1.5 py-0.5">+{dayTasks.length - 6} more</span>
                    )}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}

function TypeBreakdown() {
  const types = Object.entries(TASK_TYPE_META).map(([type, meta]) => ({
    type,
    label: meta.label,
    color: meta.color,
    count: TASKS.filter(t => t.type === type).length,
    done: TASKS.filter(t => t.type === type && t.status === 'done').length,
  }))
  const max = Math.max(...types.map(t => t.count))

  return (
    <div className="space-y-4">
      <h3 className="text-sm font-black uppercase tracking-wider text-muted-foreground flex items-center gap-2">
        <BarChart3 className="w-4 h-4" /> By Category
      </h3>
      <div className="flex flex-col gap-3">
        {types.map(t => (
          <div key={t.type} className="space-y-1">
            <div className="flex items-center justify-between">
              <span className={cn('text-[10px] font-black px-2 py-0.5 rounded border', t.color)}>{t.label}</span>
              <span className="text-xs font-bold text-muted-foreground">{t.done}/{t.count}</span>
            </div>
            <div className="h-2 bg-muted rounded-full overflow-hidden">
              <div
                className={cn('h-full rounded-full transition-all duration-500', 
                  t.type === 'api' ? 'bg-emerald-500' : 
                  t.type === 'frontend' ? 'bg-blue-500' : 
                  t.type === 'infra' ? 'bg-slate-500' : 'bg-purple-500'
                )}
                style={{ width: `${(t.count / max) * 100}%` }}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function AssigneeBreakdown() {
  const agents = ['Windsurf', 'Gemini', 'Codex', 'Nikko Fu']
  return (
    <div className="space-y-4">
      <h3 className="text-sm font-black uppercase tracking-wider text-muted-foreground flex items-center gap-2">
        <Users className="w-4 h-4" /> By Contributor
      </h3>
      <div className="flex flex-col gap-3">
        {agents.map(name => {
          const tasks = TASKS.filter(t => t.assignedTo.includes(name))
          const done = tasks.filter(t => t.status === 'done').length
          const pct = tasks.length ? Math.round((done / tasks.length) * 100) : 0
          const member = MEMBER_MAP[name]
          return (
            <div key={name} className="bg-white dark:bg-[#222529] rounded-xl border border-border/50 p-3 flex items-center gap-3">
              <MemberAvatar name={name} size="sm" />
              <div className="flex-1">
                <div className="flex items-center justify-between mb-1.5">
                  <span className="text-xs font-bold">{name}</span>
                  <span className={cn('text-xs font-black', member?.colorClass)}>{pct}%</span>
                </div>
                <Progress value={pct} className="h-1.5" />
                <p className="text-[10px] text-muted-foreground mt-1">{done} of {tasks.length} tasks done</p>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function ContributionHeatmap() {
  const byDate = TASKS.reduce<Record<string, { done: number; total: number }>>((acc, t) => {
    if (!acc[t.deadline]) acc[t.deadline] = { done: 0, total: 0 }
    acc[t.deadline].total++
    if (t.status === 'done') acc[t.deadline].done++
    return acc
  }, {})
  const dates = Object.keys(byDate).sort()
  const maxTotal = Math.max(...Object.values(byDate).map(d => d.total), 1)

  const getColor = (total: number, done: number) => {
    if (total === 0) return 'bg-muted/30'
    const pct = done / total
    const intensity = total / maxTotal
    if (pct === 1) return intensity > 0.6 ? 'bg-emerald-600' : intensity > 0.3 ? 'bg-emerald-500' : 'bg-emerald-400'
    if (pct > 0.5) return intensity > 0.6 ? 'bg-blue-500' : 'bg-blue-400'
    return intensity > 0.6 ? 'bg-amber-500' : 'bg-amber-400/70'
  }

  return (
    <div className="space-y-4">
      <h3 className="text-sm font-black uppercase tracking-wider text-muted-foreground flex items-center gap-2">
        <Activity className="w-4 h-4" /> Contribution Heatmap
      </h3>
      <div className="bg-white dark:bg-[#222529] rounded-2xl border border-border/60 p-5">
        <div className="flex flex-wrap gap-3">
          {dates.map(date => {
            const d = byDate[date]
            const fDate = new Date(date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
            return (
              <div key={date} title={`${fDate}: ${d.done}/${d.total} tasks done`}
                className="flex flex-col items-center gap-1">
                <div className={cn('w-9 h-9 rounded-lg transition-all cursor-default border border-border/20', getColor(d.total, d.done))} />
                <span className="text-[8px] text-muted-foreground font-bold leading-none">{fDate}</span>
              </div>
            )
          })}
        </div>
        <div className="flex items-center gap-2 mt-4 pt-3 border-t border-border/40">
          <span className="text-[10px] text-muted-foreground font-bold">Less</span>
          {['bg-muted/30', 'bg-amber-400/70', 'bg-blue-400', 'bg-emerald-400', 'bg-emerald-600'].map((c, i) => (
            <div key={i} className={cn('w-4 h-4 rounded-sm border border-border/20', c)} />
          ))}
          <span className="text-[10px] text-muted-foreground font-bold">More</span>
          <span className="ml-auto text-[10px] text-muted-foreground">
            <span className="inline-block w-3 h-3 rounded-sm bg-emerald-500 mr-1" />All done
            <span className="inline-block w-3 h-3 rounded-sm bg-blue-400 mx-1" />Partial
            <span className="inline-block w-3 h-3 rounded-sm bg-amber-400/70 mx-1" />Pending
          </span>
        </div>
      </div>
    </div>
  )
}

// ─── Stats Tab ────────────────────────────────────────────────────────────────

function StatsTab() {
  const byDate = TASKS.reduce<Record<string, { done: number; total: number }>>((acc, t) => {
    if (!acc[t.deadline]) acc[t.deadline] = { done: 0, total: 0 }
    acc[t.deadline].total++
    if (t.status === 'done') acc[t.deadline].done++
    return acc
  }, {})
  const dates = Object.keys(byDate).sort()
  const maxTotal = Math.max(...dates.map(d => byDate[d].total))

  return (
    <div className="space-y-8">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="md:col-span-2">
          <h3 className="text-sm font-black uppercase tracking-wider text-muted-foreground mb-4 flex items-center gap-2">
            <BarChart3 className="w-4 h-4" /> Daily Task Velocity
          </h3>
          <div className="bg-white dark:bg-[#222529] rounded-2xl border border-border/60 p-5">
            <div className="relative flex items-end gap-2 h-32">
              {dates.map(date => {
                const d = byDate[date]
                const barH = Math.round((d.total / maxTotal) * 100)
                const doneH = Math.round((d.done / maxTotal) * 100)
                const fDate = new Date(date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
                return (
                  <div key={date} className="flex-1 flex flex-col items-center gap-1">
                    <div className="w-full flex flex-col-reverse gap-px" style={{ height: `${barH}%` }}>
                      <div className="w-full rounded-sm bg-emerald-500" style={{ height: `${(doneH / barH) * 100}%` }} />
                      {d.total > d.done && (
                        <div className="w-full rounded-sm bg-amber-400/40" style={{ height: `${((d.total - d.done) / d.total) * 100}%` }} />
                      )}
                    </div>
                    <span className="text-[9px] text-muted-foreground font-bold -rotate-45 origin-right whitespace-nowrap">{fDate}</span>
                  </div>
                )
              })}
              {/* Cumulative done-rate trend line */}
              {dates.length > 1 && (() => {
                let cumDone = 0, cumTotal = 0
                const pts = dates.map((date, i) => {
                  cumDone += byDate[date].done
                  cumTotal += byDate[date].total
                  const pct = cumTotal > 0 ? (cumDone / cumTotal) * 100 : 0
                  const x = dates.length === 1 ? 50 : (i / (dates.length - 1)) * 100
                  const y = 100 - pct
                  return { x, y, date }
                })
                return (
                  <svg className="absolute inset-0 w-full h-full pointer-events-none" preserveAspectRatio="none" viewBox="0 0 100 100">
                    <polyline
                      points={pts.map(p => `${p.x},${p.y}`).join(' ')}
                      fill="none" stroke="#8b5cf6" strokeWidth="2"
                      strokeLinecap="round" strokeLinejoin="round"
                      vectorEffect="non-scaling-stroke"
                    />
                    {pts.map(p => (
                      <circle key={p.date} cx={p.x} cy={p.y} r="2.5" fill="#8b5cf6" vectorEffect="non-scaling-stroke" />
                    ))}
                  </svg>
                )
              })()}
            </div>
            <div className="flex items-center gap-4 mt-4 pt-3 border-t border-border/40">
              <span className="flex items-center gap-1.5 text-[11px] font-bold text-muted-foreground"><span className="w-3 h-3 rounded-sm bg-emerald-500 inline-block" /> Done</span>
              <span className="flex items-center gap-1.5 text-[11px] font-bold text-muted-foreground"><span className="w-3 h-3 rounded-sm bg-amber-400/40 inline-block" /> Pending</span>
              <span className="flex items-center gap-1.5 text-[11px] font-bold text-muted-foreground"><span className="w-5 h-0.5 bg-purple-500 inline-block rounded-full" /> Cumulative rate</span>
            </div>
          </div>
        </div>
        <div className="space-y-6">
          <TypeBreakdown />
        </div>
      </div>
      <ContributionHeatmap />
      <AssigneeBreakdown />
    </div>
  )
}

// ─── Main Page ────────────────────────────────────────────────────────────────

export function AgentCollabPage() {
  const [activeTab, setActiveTab] = useState<Tab>('overview')
  const [mounted, setMounted] = useState(false)
  const { agents, members: liveMembers, commLog, isLive, fetchSnapshot, fetchMembers } = useCollabStore()

  useEffect(() => { setMounted(true) }, [])

  useEffect(() => {
    fetchSnapshot()
    fetchMembers()
  }, [fetchSnapshot, fetchMembers])

  const displayMembers = useMemo<Member[]>(() => {
    if (liveMembers.length === 0) return MEMBERS
    return liveMembers.map(lm => {
      const s = MEMBER_MAP[lm.name]
      return {
        name: lm.name,
        role: lm.role || s?.role || 'Agent',
        specialty: lm.specialty || s?.specialty || '',
        tools: lm.primary_tools || s?.tools || [],
        avatar: s?.avatar ?? lm.name.substring(0, 2).toUpperCase(),
        colorClass: s?.colorClass ?? 'text-slate-600 dark:text-slate-400',
        bgClass: s?.bgClass ?? 'bg-slate-500',
        ringClass: s?.ringClass ?? 'ring-slate-400',
        type: s?.type ?? 'ai',
      }
    })
  }, [liveMembers])

  const powerMap = useMemo<Record<string, AgentPower>>(() => {
    const base = ACTIVE_SUPERPOWERS.reduce<Record<string, AgentPower>>((m, p) => { m[p.agent] = p; return m }, {})
    if (agents.length === 0) return base
    return agents.reduce<Record<string, AgentPower>>((m, a) => {
      m[a.name] = { agent: a.name, skill: a.skill, task: a.task, progress: a.progress, status: a.status as AgentPower['status'] }
      return m
    }, { ...base })
  }, [agents])

  if (!mounted) return null

  const doneCount = TASKS.filter(t => t.status === 'done').length
  const readyCount = TASKS.filter(t => t.status === 'ready').length

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-white dark:bg-[#1a1d21] overflow-hidden">
      {/* Channel Header */}
      <header className="shrink-0 border-b bg-white/80 dark:bg-[#1a1d21]/80 backdrop-blur-md">
        <div className="h-14 px-5 flex items-center gap-3">
          <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-purple-600 to-indigo-600 flex items-center justify-center shrink-0 shadow-lg">
            <Hash className="w-4 h-4 text-white" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <h1 className="text-sm font-black tracking-tight">agent-collab</h1>
              <Badge className="text-[9px] h-4 px-1.5 bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/20 font-bold">Team Hub</Badge>
            </div>
            <p className="text-[11px] text-muted-foreground truncate">Relay Agent Workspace · Team Collaboration Hub</p>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <div className="flex -space-x-1.5">
                {displayMembers.map(m => (
                  <div key={m.name} className={cn('w-6 h-6 rounded-full border-2 border-white dark:border-[#1a1d21] flex items-center justify-center text-[9px] font-black text-white shrink-0', m.bgClass)}>{m.avatar.substring(0,1)}</div>
                ))}
              </div>
              <span className="font-semibold">{displayMembers.length} members</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-1.5 text-xs">
              <span className="font-bold text-emerald-600">{doneCount} done</span>
              <span className="text-muted-foreground">·</span>
              <span className="font-bold text-amber-500">{readyCount} ready</span>
            </div>
            {isLive ? (
              <div className="flex items-center gap-1 bg-emerald-500/10 border border-emerald-500/20 rounded-full px-2.5 py-1">
                <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                <span className="text-[10px] font-black text-emerald-600 dark:text-emerald-400 uppercase tracking-wider">Live</span>
              </div>
            ) : (
              <div className="flex items-center gap-1 bg-muted/60 border border-border/40 rounded-full px-2.5 py-1">
                <div className="w-1.5 h-1.5 rounded-full bg-slate-400" />
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Static</span>
              </div>
            )}
          </div>
        </div>

        {/* Tab Bar */}
        <div className="flex items-center gap-1 px-5 h-10 border-t border-border/50">
          {TABS.map(tab => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'flex items-center gap-1.5 px-3 h-8 rounded-lg text-xs font-bold transition-all',
                activeTab === tab.id
                  ? 'bg-purple-500/10 text-purple-600 dark:text-purple-400'
                  : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
              )}
            >
              <tab.icon className="w-3.5 h-3.5" />
              {tab.label}
            </button>
          ))}
        </div>
      </header>

      {/* Content Area */}
      <ScrollArea className="flex-1">
        <div className="p-6 max-w-6xl mx-auto pb-24">
          {activeTab === 'overview' && (
            <div className="space-y-8">
              <StatsBar />
              <div>
                <h2 className="text-sm font-black uppercase tracking-wider text-muted-foreground mb-4 flex items-center gap-2">
                  <UserCircle2 className="w-4 h-4" /> Team Members
                </h2>
                <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4">
                  {displayMembers.map(m => <MemberCard key={m.name} member={m} power={powerMap[m.name]} />)}
                </div>
              </div>
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div className="lg:col-span-2">
                  <DateTimeline />
                </div>
                <div>
                  <AssigneeBreakdown />
                </div>
              </div>
            </div>
          )}
          {activeTab === 'kanban'  && <AgentCollabKanban />}
          {activeTab === 'comms'   && <AgentCollabCommLog liveSections={commLog.length > 0 ? commLog : undefined} isLive={isLive} />}
          {activeTab === 'stats'   && <StatsTab />}
        </div>
      </ScrollArea>
    </div>
  )
}
