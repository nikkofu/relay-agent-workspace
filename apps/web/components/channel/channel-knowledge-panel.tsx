"use client"

import { useEffect, useRef } from "react"
import { useRouter } from "next/navigation"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Globe, Loader2, User2, Briefcase, Lightbulb,
  Building2, FileText, Layout, Tag, ArrowUpRight,
  RefreshCw, Zap,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { format } from "date-fns"
import type { ChannelKnowledgeRef } from "@/types"

const KIND_CONFIG: Record<string, { label: string; icon: React.ElementType; badgeClass: string; bgClass: string }> = {
  person:       { label: "Person",       icon: User2,     badgeClass: "bg-sky-500/10 text-sky-700 border-sky-300 dark:border-sky-700",               bgClass: "bg-sky-500/5" },
  project:      { label: "Project",      icon: Briefcase, badgeClass: "bg-violet-500/10 text-violet-700 border-violet-300 dark:border-violet-700",     bgClass: "bg-violet-500/5" },
  concept:      { label: "Concept",      icon: Lightbulb, badgeClass: "bg-amber-500/10 text-amber-700 border-amber-300 dark:border-amber-700",         bgClass: "bg-amber-500/5" },
  organization: { label: "Organization", icon: Building2, badgeClass: "bg-emerald-500/10 text-emerald-700 border-emerald-300 dark:border-emerald-700", bgClass: "bg-emerald-500/5" },
  file:         { label: "File",         icon: FileText,  badgeClass: "bg-blue-500/10 text-blue-700 border-blue-300 dark:border-blue-700",             bgClass: "bg-blue-500/5" },
  artifact:     { label: "Artifact",     icon: Layout,    badgeClass: "bg-orange-500/10 text-orange-700 border-orange-300 dark:border-orange-700",     bgClass: "bg-orange-500/5" },
}
const getKindCfg = (kind: string) =>
  KIND_CONFIG[kind] ?? { label: kind, icon: Tag, badgeClass: "bg-muted text-muted-foreground border-muted", bgClass: "bg-muted/20" }

const REF_KIND_COLOR: Record<string, string> = {
  message:  "bg-emerald-500",
  file:     "bg-sky-500",
  thread:   "bg-violet-500",
  artifact: "bg-amber-500",
}

interface ChannelKnowledgePanelProps {
  channelId: string
}

export function ChannelKnowledgePanel({ channelId }: ChannelKnowledgePanelProps) {
  const router = useRouter()
  const {
    channelKnowledge, isLoadingChannelKnowledge, fetchChannelKnowledge, liveUpdate,
  } = useKnowledgeStore()
  const prevLiveTs = useRef<number>(0)
  const didFetch = useRef<string>("")

  useEffect(() => {
    if (didFetch.current !== channelId) {
      didFetch.current = channelId
      fetchChannelKnowledge(channelId)
    }
  }, [channelId, fetchChannelKnowledge])

  useEffect(() => {
    if (!liveUpdate || liveUpdate.ts === prevLiveTs.current) return
    prevLiveTs.current = liveUpdate.ts
  }, [liveUpdate])

  // Group refs by entity_id
  const grouped = channelKnowledge.reduce<Record<string, { refs: ChannelKnowledgeRef[]; meta: ChannelKnowledgeRef }>>((acc, ref) => {
    if (!acc[ref.entity_id]) acc[ref.entity_id] = { refs: [], meta: ref }
    acc[ref.entity_id].refs.push(ref)
    return acc
  }, {})
  const entityGroups = Object.values(grouped)

  return (
    <div className="flex flex-col h-full bg-white dark:bg-[#1a1d21] border-l w-72 shrink-0">
      {/* Header */}
      <div className="h-14 px-4 flex items-center justify-between border-b shrink-0">
        <div className="flex items-center gap-2">
          <Globe className="w-4 h-4 text-emerald-600" />
          <span className="text-sm font-black uppercase tracking-tight">Knowledge</span>
          {channelKnowledge.length > 0 && (
            <Badge variant="secondary" className="text-[9px] font-black h-4 px-1.5">
              {channelKnowledge.length}
            </Badge>
          )}
        </div>
        <Button
          variant="ghost"
          size="icon"
          className="h-7 w-7 text-muted-foreground"
          onClick={() => fetchChannelKnowledge(channelId)}
          disabled={isLoadingChannelKnowledge}
          title="Refresh"
        >
          {isLoadingChannelKnowledge
            ? <Loader2 className="w-3.5 h-3.5 animate-spin" />
            : <RefreshCw className="w-3.5 h-3.5" />
          }
        </Button>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-3 space-y-2">
          {isLoadingChannelKnowledge && entityGroups.length === 0 && (
            <div className="py-10 flex flex-col items-center gap-2 text-muted-foreground">
              <Loader2 className="w-5 h-5 animate-spin" />
              <p className="text-xs">Loading channel knowledge...</p>
            </div>
          )}

          {!isLoadingChannelKnowledge && entityGroups.length === 0 && (
            <div className="py-10 flex flex-col items-center gap-2 text-muted-foreground/60">
              <Globe className="w-7 h-7" />
              <p className="text-xs text-center italic leading-relaxed">
                No knowledge entities linked to this channel yet.
                <br />They appear automatically when messages or files mention entity titles.
              </p>
            </div>
          )}

          {entityGroups.map(({ meta, refs }) => {
            const cfg = getKindCfg(meta.entity_kind)
            const Icon = cfg.icon
            return (
              <div
                key={meta.entity_id}
                className={cn("rounded-xl border p-3 space-y-2 transition-colors", cfg.bgClass)}
              >
                {/* Entity header */}
                <div className="flex items-start justify-between gap-2">
                  <div className="flex items-center gap-1.5 min-w-0">
                    <Icon className="w-3.5 h-3.5 shrink-0 text-muted-foreground" />
                    <span className="text-xs font-bold truncate">{meta.entity_title || meta.entity_id}</span>
                  </div>
                  <button
                    className="shrink-0 text-muted-foreground hover:text-blue-600 transition-colors"
                    onClick={() => router.push(`/workspace/knowledge/${meta.entity_id}`)}
                    title="Open entity"
                  >
                    <ArrowUpRight className="w-3.5 h-3.5" />
                  </button>
                </div>
                <div className="flex items-center gap-1.5 flex-wrap">
                  <Badge className={cn("text-[8px] h-3.5 px-1", cfg.badgeClass)}>{cfg.label}</Badge>
                  <span className="text-[9px] text-muted-foreground">{refs.length} ref{refs.length > 1 ? "s" : ""}</span>
                </div>

                {/* Refs list */}
                <div className="space-y-1.5 mt-1">
                  {refs.slice(0, 3).map((ref, i) => (
                    <div key={`${ref.ref_id}-${i}`} className="rounded-lg bg-white/50 dark:bg-black/15 border border-black/5 dark:border-white/5 px-2.5 py-2 space-y-1">
                      <div className="flex items-center gap-1.5 flex-wrap">
                        <div className={cn("w-1.5 h-1.5 rounded-full shrink-0", REF_KIND_COLOR[ref.ref_kind] || "bg-muted-foreground")} />
                        <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">{ref.ref_kind}</span>
                        {ref.role && <Badge variant="outline" className="text-[8px] h-3 px-1">{ref.role}</Badge>}
                        <span className="ml-auto text-[9px] text-muted-foreground">
                          {format(new Date(ref.created_at), "MMM d")}
                        </span>
                      </div>
                      {ref.source_title && (
                        <p className="text-[10px] font-medium truncate">{ref.source_title}</p>
                      )}
                      {ref.source_snippet && (
                        <p className="text-[9px] text-muted-foreground leading-relaxed line-clamp-2 italic">
                          &ldquo;{ref.source_snippet}&rdquo;
                        </p>
                      )}
                    </div>
                  ))}
                  {refs.length > 3 && (
                    <p className="text-[9px] text-muted-foreground text-center py-0.5">
                      +{refs.length - 3} more
                    </p>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      </ScrollArea>

      {/* Footer */}
      <div className="border-t px-4 py-2.5 shrink-0">
        <button
          className="text-[10px] font-bold text-emerald-700 hover:text-emerald-600 flex items-center gap-1.5 transition-colors"
          onClick={() => router.push("/workspace/knowledge")}
        >
          <Zap className="w-3 h-3" /> Browse all entities
        </button>
      </div>
    </div>
  )
}
