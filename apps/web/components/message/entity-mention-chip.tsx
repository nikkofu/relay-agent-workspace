"use client"

import { useRouter } from "next/navigation"
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@/components/ui/hover-card"
import {
  User2, Briefcase, Lightbulb, Building2, FileText,
  Layout, Tag, ArrowUpRight, Globe,
} from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import type { MessageEntityMention } from "@/types"

const KIND_CONFIG: Record<string, { label: string; icon: React.ElementType; color: string; bg: string }> = {
  person:       { label: "Person",       icon: User2,     color: "text-sky-700 dark:text-sky-400",     bg: "bg-sky-500/10 border-sky-300 dark:border-sky-700" },
  project:      { label: "Project",      icon: Briefcase, color: "text-violet-700 dark:text-violet-400", bg: "bg-violet-500/10 border-violet-300 dark:border-violet-700" },
  concept:      { label: "Concept",      icon: Lightbulb, color: "text-amber-700 dark:text-amber-400",  bg: "bg-amber-500/10 border-amber-300 dark:border-amber-700" },
  organization: { label: "Organization", icon: Building2, color: "text-emerald-700 dark:text-emerald-400", bg: "bg-emerald-500/10 border-emerald-300 dark:border-emerald-700" },
  file:         { label: "File",         icon: FileText,  color: "text-blue-700 dark:text-blue-400",    bg: "bg-blue-500/10 border-blue-300 dark:border-blue-700" },
  artifact:     { label: "Artifact",     icon: Layout,    color: "text-orange-700 dark:text-orange-400", bg: "bg-orange-500/10 border-orange-300 dark:border-orange-700" },
}
const getCfg = (kind: string) =>
  KIND_CONFIG[kind] ?? { label: kind, icon: Tag, color: "text-muted-foreground", bg: "bg-muted/50 border-muted" }

interface EntityMentionChipProps {
  mention: MessageEntityMention
}

export function EntityMentionChip({ mention }: EntityMentionChipProps) {
  const router = useRouter()
  const cfg = getCfg(mention.entity_kind)
  const Icon = cfg.icon

  return (
    <HoverCard openDelay={300} closeDelay={100}>
      <HoverCardTrigger asChild>
        <button
          className={cn(
            "inline-flex items-center gap-1 text-[11px] font-bold px-1.5 py-0.5 rounded-md border transition-colors cursor-pointer",
            cfg.bg, cfg.color,
            "hover:brightness-95 dark:hover:brightness-125"
          )}
          onClick={() => router.push(`/workspace/knowledge/${mention.entity_id}`)}
        >
          <Icon className="w-2.5 h-2.5 shrink-0" />
          {mention.mention_text || mention.entity_title}
        </button>
      </HoverCardTrigger>
      <HoverCardContent align="start" className="w-72 p-3 space-y-2.5">
        {/* Header */}
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2 min-w-0">
            <div className={cn("w-7 h-7 rounded-lg flex items-center justify-center shrink-0", cfg.bg)}>
              <Icon className={cn("w-3.5 h-3.5", cfg.color)} />
            </div>
            <div className="min-w-0">
              <p className="text-sm font-black truncate">{mention.entity_title}</p>
              <Badge
                variant="outline"
                className={cn("text-[8px] h-3.5 px-1 mt-0.5", cfg.bg, cfg.color)}
              >
                {cfg.label}
              </Badge>
            </div>
          </div>
          <button
            className="shrink-0 text-muted-foreground hover:text-blue-600 transition-colors mt-0.5"
            onClick={() => router.push(`/workspace/knowledge/${mention.entity_id}`)}
            title="Open entity wiki"
          >
            <ArrowUpRight className="w-3.5 h-3.5" />
          </button>
        </div>

        {/* Footer link */}
        <div className="border-t pt-2 flex items-center gap-1.5">
          <Globe className="w-3 h-3 text-emerald-600 shrink-0" />
          <button
            className="text-[10px] font-bold text-emerald-700 hover:text-emerald-600 transition-colors"
            onClick={() => router.push(`/workspace/knowledge/${mention.entity_id}`)}
          >
            Open in Knowledge Wiki
          </button>
        </div>
      </HoverCardContent>
    </HoverCard>
  )
}
