"use client"

import { useState } from "react"
import { cn } from "@/lib/utils"
import { ArrowRight, Hash, Wifi, WifiOff } from "lucide-react"
import { COMM_SECTIONS, MEMBER_MAP, type CommSection, type CommMessage } from "./agent-collab-data"
import type { LiveCommSection } from "@/stores/collab-store"

function highlightMentions(content: string): React.ReactNode {
  const namedMentions = MEMBER_MAP
  const pattern = new RegExp(
    `(${Object.keys(namedMentions).map(n => n.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')}|GET|POST|PATCH|DELETE|PUT|/api/v1/[\\w/:?=&.-]+|v\\d+\\.\\d+\\.\\d+)`,
    'g'
  )

  const parts = content.split(pattern)
  return parts.map((part, i) => {
    if (namedMentions[part]) {
      const m = namedMentions[part]
      return (
        <span key={i} className={cn('inline-flex items-center font-bold rounded px-1 mx-0.5', m.colorClass, 'bg-current/10')}>
          @{part}
        </span>
      )
    }
    if (['GET', 'POST', 'PATCH', 'DELETE', 'PUT'].includes(part)) {
      return <span key={i} className="font-mono text-amber-600 dark:text-amber-400 bg-amber-500/10 rounded px-1 text-[10px]">{part}</span>
    }
    if (part.startsWith('/api/') || part.startsWith('v0.')) {
      return <span key={i} className="font-mono text-blue-600 dark:text-blue-400 bg-blue-500/10 rounded px-1 text-[10px]">{part}</span>
    }
    return <span key={i}>{part}</span>
  })
}

// ─── Member Avatar ────────────────────────────────────────────────────────────

function MemberAvatar({ name, size = 'md' }: { name: string; size?: 'sm' | 'md' }) {
  const m = MEMBER_MAP[name]
  const bg = m?.bgClass ?? 'bg-slate-500'
  const initials = m?.avatar ?? name.substring(0, 2).toUpperCase()
  const sz = size === 'sm' ? 'w-6 h-6 text-[9px]' : 'w-9 h-9 text-xs'
  return (
    <div className={cn('rounded-xl flex items-center justify-center font-black text-white shrink-0 ring-2 ring-white/20 dark:ring-black/20', bg, sz)}>
      {initials}
    </div>
  )
}

// ─── Single Message ───────────────────────────────────────────────────────────

function CommMessage({ msg, sectionDate }: { msg: CommMessage; sectionDate: string }) {
  const fromMember = MEMBER_MAP[msg.from]
  const toMember = msg.to ? MEMBER_MAP[msg.to] : null
  const isDirect = !!msg.to

  return (
    <div className={cn(
      'flex gap-3 p-4 rounded-2xl border transition-all hover:shadow-sm',
      isDirect
        ? 'bg-gradient-to-r from-purple-500/5 to-blue-500/5 border-purple-500/15 hover:border-purple-500/30'
        : 'bg-white dark:bg-[#222529] border-border/50 hover:border-border'
    )}>
      {/* Avatar */}
      <MemberAvatar name={msg.from} />

      <div className="flex-1 min-w-0 space-y-1.5">
        {/* From / To Header */}
        <div className="flex items-center gap-2 flex-wrap">
          <span className={cn('text-sm font-black', fromMember?.colorClass ?? 'text-foreground')}>
            {msg.from}
          </span>
          {isDirect && (
            <>
              <ArrowRight className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
              <div className="flex items-center gap-1.5">
                <MemberAvatar name={msg.to!} size="sm" />
                <span className={cn('text-sm font-black', toMember?.colorClass ?? 'text-foreground')}>
                  @{msg.to}
                </span>
              </div>
            </>
          )}
          <span className="text-[10px] text-muted-foreground ml-auto shrink-0">
            {sectionDate}
          </span>
        </div>

        {/* Direct indicator badge */}
        {isDirect && (
          <div className="flex items-center gap-1 mb-1">
            <span className="text-[9px] font-black uppercase tracking-wider text-purple-600 dark:text-purple-400 bg-purple-500/10 border border-purple-500/20 rounded px-1.5 py-0.5">
              Direct Message
            </span>
          </div>
        )}

        {/* Content */}
        {msg.isCode ? (
          <div className="bg-muted/60 rounded-lg border border-border/40 p-3 font-mono text-[11px] leading-relaxed text-foreground/80 whitespace-pre-wrap break-all">
            {highlightMentions(msg.content)}
          </div>
        ) : (
          <p className="text-sm text-foreground/90 leading-relaxed">
            {highlightMentions(msg.content)}
          </p>
        )}
      </div>
    </div>
  )
}

// ─── Section ──────────────────────────────────────────────────────────────────

function CommSection({ section }: { section: CommSection }) {
  const [collapsed, setCollapsed] = useState(false)
  const formattedDate = new Date(section.date).toLocaleDateString('en-US', {
    weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
  })
  const directCount = section.messages.filter(m => !!m.to).length

  return (
    <div className="space-y-3">
      {/* Section Header */}
      <button
        onClick={() => setCollapsed(c => !c)}
        className="w-full group"
      >
        <div className="flex items-center gap-3">
          <div className="flex-1 h-px bg-border/60" />
          <div className="flex items-center gap-2 shrink-0 bg-white dark:bg-[#1a1d21] px-3 py-1.5 rounded-xl border border-border/60 group-hover:border-border transition-colors">
            <Hash className="w-3 h-3 text-muted-foreground" />
            <span className="text-[11px] font-black text-foreground/70 group-hover:text-foreground transition-colors">{section.title}</span>
            {directCount > 0 && (
              <span className="text-[10px] font-black text-purple-600 dark:text-purple-400 bg-purple-500/10 border border-purple-500/20 rounded px-1.5 ml-1">
                {directCount} DM{directCount > 1 ? 's' : ''}
              </span>
            )}
          </div>
          <div className="flex-1 h-px bg-border/60" />
        </div>
        <p className="text-[10px] text-muted-foreground text-center mt-1.5">{formattedDate} · {section.messages.length} message{section.messages.length > 1 ? 's' : ''}</p>
      </button>

      {/* Messages */}
      {!collapsed && (
        <div className="space-y-3 pl-2">
          {section.messages.map(msg => (
            <CommMessage key={msg.id} msg={msg} sectionDate={section.date} />
          ))}
        </div>
      )}
    </div>
  )
}

// ─── Live → Static bridge ─────────────────────────────────────────────────────

function bridgeToCommSection(live: LiveCommSection): CommSection {
  return {
    id: live.id ?? `live-${live.title}`,
    date: live.date,
    title: live.title,
    messages: live.messages.map((m, i) => ({
      id: m.id ?? `live-msg-${i}`,
      from: m.from,
      to: m.to,
      content: m.content,
      isCode: m.isCode,
    })),
  }
}

// ─── Main Comm Log ────────────────────────────────────────────────────────────

interface AgentCollabCommLogProps {
  liveSections?: LiveCommSection[]
  isLive?: boolean
}

export function AgentCollabCommLog({ liveSections, isLive }: AgentCollabCommLogProps) {
  const sections: CommSection[] = liveSections && liveSections.length > 0
    ? liveSections.map(bridgeToCommSection)
    : COMM_SECTIONS

  return (
    <div className="space-y-8 max-w-4xl mx-auto">
      {/* Legend + Live indicator */}
      <div className="flex items-center gap-4 p-4 bg-muted/30 rounded-2xl border border-border/40 flex-wrap">
        <span className="text-[10px] font-black uppercase tracking-wider text-muted-foreground">Legend</span>
        <div className="flex items-center gap-2">
          <div className="w-8 h-4 rounded-sm bg-white dark:bg-[#222529] border border-border/50" />
          <span className="text-[11px] font-semibold text-muted-foreground">Broadcast / Self</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-8 h-4 rounded-sm bg-gradient-to-r from-purple-500/10 to-blue-500/10 border border-purple-500/20" />
          <ArrowRight className="w-3 h-3 text-muted-foreground" />
          <span className="text-[11px] font-semibold text-muted-foreground">From → To (Direct)</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-[10px] font-mono text-amber-600 bg-amber-500/10 rounded px-1">GET</span>
          <span className="text-[11px] font-semibold text-muted-foreground">HTTP method</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-[10px] font-mono text-blue-600 bg-blue-500/10 rounded px-1">/api/v1/...</span>
          <span className="text-[11px] font-semibold text-muted-foreground">Endpoint</span>
        </div>
        <div className="ml-auto flex items-center gap-1.5">
          {isLive ? (
            <>
              <Wifi className="w-3 h-3 text-emerald-500" />
              <span className="text-[10px] font-black text-emerald-600 dark:text-emerald-400">Live data</span>
            </>
          ) : (
            <>
              <WifiOff className="w-3 h-3 text-muted-foreground" />
              <span className="text-[10px] font-bold text-muted-foreground">Static fallback</span>
            </>
          )}
        </div>
      </div>

      {/* Sections */}
      {sections.map(section => (
        <CommSection key={section.id} section={section} />
      ))}
    </div>
  )
}
