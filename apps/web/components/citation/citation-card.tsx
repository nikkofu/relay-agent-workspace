"use client"

import { FileText, MessageSquare, MessageCircle, Layout, Quote, MapPin, Tag, User2 } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import type { CitationEvidence, EvidenceKind } from "@/types"

type KindConfig = {
  label: string
  icon: React.ElementType
  badgeClass: string
  borderClass: string
  bgClass: string
}

const KIND_CONFIG: Record<EvidenceKind, KindConfig> = {
  file_chunk: {
    label: 'File Chunk',
    icon: FileText,
    badgeClass: 'bg-sky-500/10 text-sky-700 dark:text-sky-400 border-sky-300 dark:border-sky-700',
    borderClass: 'border-sky-200 dark:border-sky-800',
    bgClass: 'bg-sky-500/5',
  },
  message: {
    label: 'Message',
    icon: MessageSquare,
    badgeClass: 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-300 dark:border-emerald-700',
    borderClass: 'border-emerald-200 dark:border-emerald-800',
    bgClass: 'bg-emerald-500/5',
  },
  thread: {
    label: 'Thread',
    icon: MessageCircle,
    badgeClass: 'bg-violet-500/10 text-violet-700 dark:text-violet-400 border-violet-300 dark:border-violet-700',
    borderClass: 'border-violet-200 dark:border-violet-800',
    bgClass: 'bg-violet-500/5',
  },
  artifact_section: {
    label: 'Artifact',
    icon: Layout,
    badgeClass: 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-300 dark:border-amber-700',
    borderClass: 'border-amber-200 dark:border-amber-800',
    bgClass: 'bg-amber-500/5',
  },
}

interface CitationCardProps {
  citation: CitationEvidence
  compact?: boolean
  onClick?: () => void
}

export function CitationCard({ citation, compact = false, onClick }: CitationCardProps) {
  const config = KIND_CONFIG[citation.evidence_kind] ?? KIND_CONFIG.file_chunk
  const Icon = config.icon

  return (
    <div
      className={cn(
        "rounded-xl border transition-colors",
        config.borderClass,
        config.bgClass,
        onClick && "cursor-pointer hover:brightness-95",
        compact ? "p-2.5" : "p-4"
      )}
      onClick={onClick}
    >
      {/* Header */}
      <div className="flex items-center gap-2 mb-2 flex-wrap">
        <Badge className={cn("gap-1 text-[9px] h-4 px-1.5 font-bold", config.badgeClass)}>
          <Icon className="w-2.5 h-2.5" />
          {config.label}
        </Badge>
        {citation.source_kind && (
          <span className="text-[9px] text-muted-foreground uppercase font-black tracking-wider">{citation.source_kind}</span>
        )}
        {citation.score !== undefined && (
          <span className="ml-auto text-[9px] text-muted-foreground font-mono tabular-nums">
            {Math.round(citation.score * 100)}% match
          </span>
        )}
      </div>

      {/* Title */}
      {citation.title && (
        <p className={cn("font-bold truncate mb-1.5", compact ? "text-xs" : "text-sm")}>
          {citation.title}
        </p>
      )}

      {/* Snippet */}
      <div className={cn(
        "rounded-lg bg-white/50 dark:bg-black/20 border border-black/5 dark:border-white/5 px-2.5 py-2",
        compact ? "text-[10px]" : "text-[11px]"
      )}>
        <Quote className="w-3 h-3 opacity-20 mb-0.5" />
        <p className="text-foreground/80 leading-relaxed line-clamp-4">{citation.snippet}</p>
      </div>

      {/* Footer */}
      <div className="flex items-center gap-3 mt-2 flex-wrap">
        {citation.locator && (
          <span className="flex items-center gap-1 text-[10px] text-muted-foreground">
            <MapPin className="w-3 h-3" />
            {citation.locator}
          </span>
        )}
        {citation.ref_kind && (
          <span className="flex items-center gap-1 text-[10px] text-muted-foreground">
            <Tag className="w-3 h-3" />
            {citation.ref_kind}
          </span>
        )}
        {citation.entity_id && (
          <Badge variant="secondary" className="text-[9px] h-4 px-1.5 gap-1 ml-auto">
            <User2 className="w-2.5 h-2.5" />
            {citation.entity_title || citation.entity_id}
          </Badge>
        )}
      </div>
    </div>
  )
}
