"use client"

import { useRouter } from "next/navigation"
import { Newspaper, Globe, TrendingUp, TrendingDown, Zap } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import type { ChannelKnowledgeDigest } from "@/types"

interface KnowledgeDigestCardProps {
  digest: ChannelKnowledgeDigest
}

export function KnowledgeDigestCard({ digest }: KnowledgeDigestCardProps) {
  const router = useRouter()
  const DeltaIcon = (digest.delta ?? 0) > 0 ? TrendingUp : (digest.delta ?? 0) < 0 ? TrendingDown : Zap

  return (
    <div className="rounded-xl border bg-gradient-to-br from-emerald-500/5 via-sky-500/5 to-violet-500/5 p-3 mt-2 max-w-xl">
      {/* Header */}
      <div className="flex items-center gap-2 mb-2">
        <div className="w-7 h-7 rounded-lg bg-emerald-500/10 border border-emerald-400/30 flex items-center justify-center shrink-0">
          <Newspaper className="w-3.5 h-3.5 text-emerald-600" />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1.5">
            <span className="text-[10px] font-black uppercase tracking-widest text-emerald-700">
              {digest.window} digest
            </span>
            <Badge variant="secondary" className="text-[9px] font-black h-4 px-1.5">
              {digest.total_refs} refs
            </Badge>
            {digest.delta !== undefined && digest.delta !== 0 && (
              <span
                className={cn(
                  "inline-flex items-center gap-0.5 text-[9px] font-black px-1.5 py-0.5 rounded-full",
                  digest.delta > 0 ? "bg-emerald-500/10 text-emerald-700" : "bg-rose-500/10 text-rose-700"
                )}
              >
                <DeltaIcon className="w-2.5 h-2.5" />
                {digest.delta > 0 ? '+' : ''}{digest.delta}
              </span>
            )}
          </div>
          {digest.headline && (
            <p className="text-xs font-bold leading-tight mt-0.5 line-clamp-2">{digest.headline}</p>
          )}
        </div>
      </div>

      {/* Entry list */}
      <div className="space-y-1">
        {digest.entries.map(entry => (
          <button
            key={entry.entity_id}
            onClick={() => router.push(`/workspace/knowledge/${entry.entity_id}`)}
            className="w-full text-left rounded-lg border bg-white/50 dark:bg-black/15 px-2.5 py-1.5 hover:border-blue-300 hover:bg-white/70 dark:hover:bg-black/25 transition-colors flex items-center gap-2"
          >
            <Globe className="w-3 h-3 text-emerald-600 shrink-0" />
            <span className="text-[11px] font-bold truncate flex-1">{entry.entity_title}</span>
            <Badge variant="outline" className="text-[8px] h-3.5 px-1 uppercase tracking-widest">
              {entry.entity_kind}
            </Badge>
            <span className="text-[10px] font-black text-muted-foreground">
              {entry.ref_count}
            </span>
            {entry.delta !== undefined && entry.delta !== 0 && (
              <span className={cn(
                "text-[9px] font-bold",
                entry.delta > 0 ? "text-emerald-700" : "text-rose-700"
              )}>
                {entry.delta > 0 ? '+' : ''}{entry.delta}
              </span>
            )}
          </button>
        ))}
      </div>
    </div>
  )
}
