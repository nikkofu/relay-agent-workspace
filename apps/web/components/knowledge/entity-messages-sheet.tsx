"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription,
} from "@/components/ui/sheet"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Badge } from "@/components/ui/badge"
import {
  Hash, Loader2, MessageSquare, ArrowUpRight, Globe, Link2, Quote, Type,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { format } from "date-fns"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import type { MessageByEntityResult, MatchSource } from "@/types"

const MATCH_SOURCE_CFG: Record<MatchSource, { label: string; icon: React.ElementType; color: string }> = {
  knowledge_ref:    { label: "Ref",     icon: Link2,  color: "bg-emerald-500/10 text-emerald-700 border-emerald-300 dark:border-emerald-700" },
  explicit_mention: { label: "Mention", icon: Quote,  color: "bg-violet-500/10 text-violet-700 border-violet-300 dark:border-violet-700" },
  title_match:      { label: "Title",   icon: Type,   color: "bg-sky-500/10 text-sky-700 border-sky-300 dark:border-sky-700" },
}

interface EntityMessagesSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  entityId: string
  entityTitle: string
  channelId?: string
}

export function EntityMessagesSheet({ open, onOpenChange, entityId, entityTitle, channelId }: EntityMessagesSheetProps) {
  const router = useRouter()
  const { searchMessagesByEntity } = useKnowledgeStore()
  const [results, setResults] = useState<MessageByEntityResult[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [scope, setScope] = useState<'channel' | 'workspace'>('channel')

  useEffect(() => {
    if (!open || !entityId) return
    let cancelled = false
    setIsLoading(true)
    searchMessagesByEntity(entityId, scope === 'channel' ? channelId : undefined).then(data => {
      if (!cancelled) {
        setResults(data)
        setIsLoading(false)
      }
    })
    return () => { cancelled = true }
  }, [open, entityId, channelId, scope, searchMessagesByEntity])

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-[520px] sm:max-w-[520px] p-0 flex flex-col">
        <SheetHeader className="px-5 py-4 border-b shrink-0">
          <div className="flex items-center gap-2">
            <Globe className="w-4 h-4 text-emerald-600 shrink-0" />
            <SheetTitle className="text-sm font-black truncate">{entityTitle}</SheetTitle>
            <Badge variant="secondary" className="text-[9px] font-black h-4 px-1.5 ml-auto">
              {results.length} {results.length === 1 ? 'message' : 'messages'}
            </Badge>
          </div>
          <SheetDescription className="text-[11px] text-muted-foreground">
            Every message referencing this entity.
          </SheetDescription>

          {/* Scope toggle */}
          {channelId && (
            <div className="inline-flex gap-1 mt-1">
              {(['channel', 'workspace'] as const).map(s => (
                <button
                  key={s}
                  onClick={() => setScope(s)}
                  className={cn(
                    "text-[10px] font-bold uppercase tracking-widest px-2 py-0.5 rounded-full border transition-colors",
                    scope === s
                      ? "bg-emerald-500/10 border-emerald-400/40 text-emerald-700"
                      : "bg-transparent border-border text-muted-foreground hover:bg-muted"
                  )}
                >
                  {s === 'channel' ? 'This channel' : 'Workspace'}
                </button>
              ))}
            </div>
          )}
        </SheetHeader>

        <ScrollArea className="flex-1">
          <div className="p-3 space-y-2">
            {isLoading && (
              <div className="py-10 flex flex-col items-center gap-2 text-muted-foreground">
                <Loader2 className="w-5 h-5 animate-spin" />
                <p className="text-xs">Searching messages...</p>
              </div>
            )}

            {!isLoading && results.length === 0 && (
              <div className="py-10 flex flex-col items-center gap-2 text-muted-foreground/60">
                <MessageSquare className="w-7 h-7" />
                <p className="text-xs italic">No messages reference this entity yet.</p>
              </div>
            )}

            {results.map(msg => (
              <button
                key={msg.id}
                onClick={() => {
                  if (msg.channel_id) router.push(`/workspace?c=${msg.channel_id}&m=${msg.id}`)
                  onOpenChange(false)
                }}
                className="w-full text-left rounded-xl border p-3 space-y-1.5 hover:bg-muted/50 hover:border-blue-300 transition-colors group"
              >
                <div className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-1.5 min-w-0">
                    {msg.channel_name && (
                      <span className="inline-flex items-center gap-1 text-[10px] font-bold text-muted-foreground">
                        <Hash className="w-2.5 h-2.5" />
                        {msg.channel_name}
                      </span>
                    )}
                    {msg.sender_name && (
                      <span className="text-[10px] text-muted-foreground">· {msg.sender_name}</span>
                    )}
                  </div>
                  <ArrowUpRight className="w-3 h-3 text-muted-foreground group-hover:text-blue-600 shrink-0" />
                </div>

                <p className="text-xs leading-relaxed line-clamp-3">
                  {msg.snippet || stripHtml(msg.content)}
                </p>

                <div className="flex items-center gap-1.5 flex-wrap pt-0.5">
                  {msg.match_sources.map((src, i) => {
                    const cfg = MATCH_SOURCE_CFG[src]
                    if (!cfg) return null
                    const Icon = cfg.icon
                    return (
                      <Badge key={i} variant="outline" className={cn("text-[8px] h-4 px-1 gap-0.5", cfg.color)}>
                        <Icon className="w-2 h-2" />
                        {cfg.label}
                      </Badge>
                    )
                  })}
                  <span className="ml-auto text-[9px] text-muted-foreground">
                    {format(new Date(msg.created_at), "MMM d · h:mm a")}
                  </span>
                </div>
              </button>
            ))}
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}

function stripHtml(html: string): string {
  if (!html) return ""
  return html.replace(/<[^>]*>/g, '').trim()
}
