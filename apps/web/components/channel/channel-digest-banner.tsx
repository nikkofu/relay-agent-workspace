"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  Newspaper, Pin, Loader2, ChevronDown, ChevronUp, X, Globe,
  TrendingUp, TrendingDown, Zap,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { format } from "date-fns"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import type { ChannelKnowledgeDigest } from "@/types"

type DigestWindow = 'daily' | 'weekly' | 'monthly'

interface ChannelDigestBannerProps {
  channelId: string
}

export function ChannelDigestBanner({ channelId }: ChannelDigestBannerProps) {
  const router = useRouter()
  const { fetchChannelDigest, publishChannelDigest } = useKnowledgeStore()
  const [expanded, setExpanded] = useState(false)
  const [dismissed, setDismissed] = useState(false)
  const [digest, setDigest] = useState<ChannelKnowledgeDigest | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [window, setWindow] = useState<DigestWindow>('weekly')
  const [isPublishing, setIsPublishing] = useState(false)

  useEffect(() => {
    setDismissed(false)
    setExpanded(false)
    setDigest(null)
  }, [channelId])

  useEffect(() => {
    if (!channelId || dismissed) return
    let cancelled = false
    setIsLoading(true)
    fetchChannelDigest(channelId, window, 5).then(data => {
      if (!cancelled) {
        setDigest(data)
        setIsLoading(false)
      }
    })
    return () => { cancelled = true }
  }, [channelId, window, fetchChannelDigest, dismissed])

  const handlePublish = async () => {
    if (!channelId) return
    setIsPublishing(true)
    const msg = await publishChannelDigest(channelId, { window, limit: 5, pin: true })
    setIsPublishing(false)
    if (msg) {
      setDismissed(true)
    }
  }

  if (dismissed) return null
  if (isLoading && !digest) {
    return (
      <div className="px-4 py-2 border-b bg-muted/10 flex items-center gap-2 text-muted-foreground text-xs">
        <Loader2 className="w-3 h-3 animate-spin" />
        <span>Generating knowledge digest...</span>
      </div>
    )
  }
  if (!digest || !digest.entries || digest.entries.length === 0) return null

  const deltaIcon = (digest.delta ?? 0) > 0 ? TrendingUp : (digest.delta ?? 0) < 0 ? TrendingDown : Zap
  const DeltaIcon = deltaIcon

  return (
    <div className="border-b bg-gradient-to-r from-emerald-500/5 via-sky-500/5 to-violet-500/5">
      <div className="px-4 py-2">
        <div className="flex items-center gap-2">
          <div className="w-7 h-7 rounded-lg bg-emerald-500/10 border border-emerald-400/30 flex items-center justify-center shrink-0">
            <Newspaper className="w-3.5 h-3.5 text-emerald-600" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="text-[10px] font-black uppercase tracking-widest text-emerald-700">
                {window} knowledge digest
              </span>
              <Badge variant="secondary" className="text-[9px] font-black h-4 px-1.5">
                {digest.entries.length}
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
            <p className="text-xs font-bold leading-tight truncate mt-0.5">
              {digest.headline || `Top ${digest.entries.length} entities referenced this ${window === 'daily' ? 'day' : window === 'weekly' ? 'week' : 'month'}`}
            </p>
          </div>

          {/* Window picker */}
          <div className="inline-flex gap-0.5 border rounded-md overflow-hidden">
            {(['daily', 'weekly', 'monthly'] as DigestWindow[]).map(w => (
              <button
                key={w}
                onClick={() => setWindow(w)}
                className={cn(
                  "text-[9px] font-black uppercase tracking-widest px-1.5 py-0.5 transition-colors",
                  window === w
                    ? "bg-emerald-500/20 text-emerald-800"
                    : "text-muted-foreground hover:bg-muted"
                )}
              >
                {w[0]}
              </button>
            ))}
          </div>

          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7"
            onClick={() => setExpanded(v => !v)}
            title={expanded ? "Collapse" : "Expand"}
          >
            {expanded ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
          </Button>

          <Button
            size="sm"
            className="h-7 text-[10px] font-black bg-emerald-600 hover:bg-emerald-600/90 text-white gap-1"
            onClick={handlePublish}
            disabled={isPublishing}
          >
            {isPublishing ? <Loader2 className="w-3 h-3 animate-spin" /> : <Pin className="w-3 h-3" />}
            Publish & Pin
          </Button>

          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-muted-foreground"
            onClick={() => setDismissed(true)}
            title="Dismiss"
          >
            <X className="w-3.5 h-3.5" />
          </Button>
        </div>

        {/* Expanded entry list */}
        {expanded && (
          <div className="mt-2 space-y-1.5 pl-9">
            <p className="text-[9px] text-muted-foreground">
              Generated {format(new Date(digest.generated_at), "MMM d, h:mm a")} · {digest.total_refs} total refs
            </p>
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
        )}
      </div>
    </div>
  )
}
