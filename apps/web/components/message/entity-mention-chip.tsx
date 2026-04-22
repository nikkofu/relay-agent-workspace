"use client"

import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@/components/ui/hover-card"
import {
  User2, Briefcase, Lightbulb, Building2, FileText,
  Layout, Tag, ArrowUpRight, Globe, MessageSquare, Hash, Loader2, Clock,
} from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { formatDistanceToNow } from "date-fns"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { useChannelStore } from "@/stores/channel-store"
import { EntityMessagesSheet } from "@/components/knowledge/entity-messages-sheet"
import type { MessageEntityMention, EntityHoverCard } from "@/types"

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
  const { fetchEntityHover } = useKnowledgeStore()
  const { currentChannel } = useChannelStore()
  const [open, setOpen] = useState(false)
  const [hover, setHover] = useState<EntityHoverCard | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [drilldownOpen, setDrilldownOpen] = useState(false)
  const cfg = getCfg(mention.entity_kind)
  const Icon = cfg.icon

  // Lazy fetch hover card on first open
  useEffect(() => {
    if (!open || hover || isLoading) return
    setIsLoading(true)
    fetchEntityHover(mention.entity_id, currentChannel?.id, 7).then(data => {
      setHover(data)
      setIsLoading(false)
    })
  }, [open, hover, isLoading, mention.entity_id, currentChannel?.id, fetchEntityHover])

  return (
    <>
      <HoverCard openDelay={300} closeDelay={100} open={open} onOpenChange={setOpen}>
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
        <HoverCardContent align="start" className="w-80 p-3 space-y-2.5">
          {/* Header */}
          <div className="flex items-start justify-between gap-2">
            <div className="flex items-center gap-2 min-w-0">
              <div className={cn("w-7 h-7 rounded-lg flex items-center justify-center shrink-0", cfg.bg)}>
                <Icon className={cn("w-3.5 h-3.5", cfg.color)} />
              </div>
              <div className="min-w-0">
                <p className="text-sm font-black truncate">{hover?.entity_title || mention.entity_title}</p>
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

          {/* Summary */}
          {hover?.summary && (
            <p className="text-[11px] text-muted-foreground leading-relaxed line-clamp-2 italic">
              {hover.summary}
            </p>
          )}

          {/* Loading */}
          {isLoading && !hover && (
            <div className="flex items-center gap-1.5 text-muted-foreground py-1">
              <Loader2 className="w-3 h-3 animate-spin" />
              <span className="text-[10px]">Loading activity...</span>
            </div>
          )}

          {/* Stats */}
          {hover && (
            <div className="grid grid-cols-3 gap-1.5 pt-1">
              <StatCell label="Total" value={hover.ref_count} />
              <StatCell label="Messages" value={hover.message_ref_count} />
              <StatCell label="Files" value={hover.file_ref_count} />
            </div>
          )}

          {/* Recent + last activity */}
          {hover && (hover.recent_ref_count > 0 || hover.last_activity_at) && (
            <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
              {hover.recent_ref_count > 0 && (
                <span className="inline-flex items-center gap-1 font-bold text-emerald-700">
                  <MessageSquare className="w-2.5 h-2.5" />
                  +{hover.recent_ref_count} in 7d
                </span>
              )}
              {hover.last_activity_at && (
                <span className="inline-flex items-center gap-1 ml-auto">
                  <Clock className="w-2.5 h-2.5" />
                  {formatDistanceToNow(new Date(hover.last_activity_at), { addSuffix: true })}
                </span>
              )}
            </div>
          )}

          {/* Related channels */}
          {hover && hover.related_channels.length > 0 && (
            <div className="space-y-1 pt-1">
              <p className="text-[9px] font-black uppercase tracking-widest text-muted-foreground">Related channels</p>
              <div className="flex flex-wrap gap-1">
                {hover.related_channels.slice(0, 4).map(rc => (
                  <button
                    key={rc.channel_id}
                    onClick={() => router.push(`/workspace?c=${rc.channel_id}`)}
                    className="inline-flex items-center gap-0.5 text-[10px] font-bold px-1.5 py-0.5 rounded-md border bg-muted/50 hover:bg-muted transition-colors"
                  >
                    <Hash className="w-2 h-2" />
                    {rc.channel_name || rc.channel_id.slice(0, 6)}
                    <span className="text-muted-foreground font-normal ml-0.5">{rc.ref_count}</span>
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Footer actions */}
          <div className="border-t pt-2 flex items-center justify-between gap-2">
            <button
              className="inline-flex items-center gap-1 text-[10px] font-bold text-emerald-700 hover:text-emerald-600 transition-colors"
              onClick={() => router.push(`/workspace/knowledge/${mention.entity_id}`)}
            >
              <Globe className="w-3 h-3" />
              Open Wiki
            </button>
            <button
              className="inline-flex items-center gap-1 text-[10px] font-bold text-blue-600 hover:text-blue-500 transition-colors"
              onClick={() => { setOpen(false); setDrilldownOpen(true) }}
            >
              <MessageSquare className="w-3 h-3" />
              View messages
            </button>
          </div>
        </HoverCardContent>
      </HoverCard>

      <EntityMessagesSheet
        open={drilldownOpen}
        onOpenChange={setDrilldownOpen}
        entityId={mention.entity_id}
        entityTitle={mention.entity_title}
        channelId={currentChannel?.id}
      />
    </>
  )
}

function StatCell({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-md border bg-muted/30 px-2 py-1.5 text-center">
      <p className="text-sm font-black leading-none">{value}</p>
      <p className="text-[8px] font-bold uppercase tracking-widest text-muted-foreground mt-0.5">{label}</p>
    </div>
  )
}
