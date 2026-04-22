"use client"

import { useEffect, useMemo, useState, useCallback } from "react"
import { useRouter } from "next/navigation"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import {
  Inbox, Hash, Star, Loader2, ChevronRight, CheckCheck, Pin, Newspaper, Clock,
  TrendingUp, TrendingDown, Zap, Globe, MessageSquare,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { formatDistanceToNow } from "date-fns"
import { useLocale, formatLocaleDate } from "@/hooks/use-locale"
import { KnowledgeDigestCard } from "@/components/message/knowledge-digest-card"
import type { KnowledgeInboxItem, KnowledgeInboxScope, KnowledgeInboxDetail } from "@/types"

export default function KnowledgeInboxPage() {
  const router = useRouter()
  const locale = useLocale()
  const {
    knowledgeInbox, knowledgeInboxScope, knowledgeInboxUnreadCount, isLoadingInbox,
    fetchKnowledgeInbox, fetchKnowledgeInboxItem, markInboxRead, liveUpdate,
  } = useKnowledgeStore()
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [detail, setDetail] = useState<KnowledgeInboxDetail | null>(null)
  const [isLoadingDetail, setIsLoadingDetail] = useState(false)

  useEffect(() => {
    fetchKnowledgeInbox(knowledgeInboxScope, 50)
  }, [knowledgeInboxScope, fetchKnowledgeInbox])

  // Live refresh on digest.published
  useEffect(() => {
    if (liveUpdate?.type === 'digest.published') {
      fetchKnowledgeInbox(knowledgeInboxScope, 50)
    }
  }, [liveUpdate, knowledgeInboxScope, fetchKnowledgeInbox])

  const selectedItem = useMemo(
    () => knowledgeInbox.find(i => i.id === selectedId) || knowledgeInbox[0] || null,
    [knowledgeInbox, selectedId]
  )

  const handleSelect = useCallback(async (item: KnowledgeInboxItem) => {
    setSelectedId(item.id)
    if (!item.is_read && item.message?.id) {
      markInboxRead([item.message.id])
    }
    setIsLoadingDetail(true)
    const d = await fetchKnowledgeInboxItem(item.id)
    setDetail(d)
    setIsLoadingDetail(false)
  }, [markInboxRead, fetchKnowledgeInboxItem])

  const handleMarkAllRead = async () => {
    const unread = knowledgeInbox.filter(i => !i.is_read && i.message?.id).map(i => i.message.id!)
    if (unread.length === 0) return
    await markInboxRead(unread)
  }

  const handleJumpToMessage = (item: KnowledgeInboxItem) => {
    if (item.channel?.id && item.message?.id) {
      router.push(`/workspace?c=${item.channel.id}&m=${item.message.id}`)
    }
  }

  return (
    <div className="flex-1 flex flex-col h-full bg-white dark:bg-[#1a1d21] overflow-hidden">
      {/* Header */}
      <header className="h-14 px-6 flex items-center justify-between border-b shrink-0">
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-emerald-500/10 flex items-center justify-center">
            <Inbox className="w-4 h-4 text-emerald-600" />
          </div>
          <div>
            <h1 className="text-sm font-black tracking-tight uppercase">Knowledge Inbox</h1>
            <p className="text-[10px] text-muted-foreground">
              {knowledgeInboxUnreadCount > 0 ? `${knowledgeInboxUnreadCount} unread digest${knowledgeInboxUnreadCount === 1 ? '' : 's'}` : 'All caught up'}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {/* Scope toggle */}
          <div className="inline-flex rounded-md border overflow-hidden">
            {(['all', 'starred'] as KnowledgeInboxScope[]).map(s => (
              <button
                key={s}
                onClick={() => fetchKnowledgeInbox(s, 50)}
                className={cn(
                  "text-[10px] font-black uppercase tracking-widest px-3 py-1 transition-colors inline-flex items-center gap-1",
                  knowledgeInboxScope === s
                    ? "bg-emerald-500/15 text-emerald-700"
                    : "text-muted-foreground hover:bg-muted"
                )}
              >
                {s === 'starred' && <Star className="w-2.5 h-2.5 fill-current" />}
                {s}
              </button>
            ))}
          </div>

          <Button
            variant="outline"
            size="sm"
            className="h-7 text-[10px] font-bold gap-1"
            onClick={handleMarkAllRead}
            disabled={knowledgeInboxUnreadCount === 0 || isLoadingInbox}
          >
            <CheckCheck className="w-3 h-3" />
            Mark all read
          </Button>
        </div>
      </header>

      <div className="flex-1 flex overflow-hidden">
        {/* Left column: inbox list */}
        <div className="w-[380px] border-r flex flex-col bg-muted/5">
          <ScrollArea className="flex-1">
            {isLoadingInbox && knowledgeInbox.length === 0 ? (
              <div className="py-12 flex flex-col items-center gap-2 text-muted-foreground">
                <Loader2 className="w-5 h-5 animate-spin" />
                <p className="text-xs">Loading inbox...</p>
              </div>
            ) : knowledgeInbox.length === 0 ? (
              <div className="py-16 flex flex-col items-center gap-3 text-muted-foreground/60 px-8 text-center">
                <Inbox className="w-10 h-10" />
                <p className="text-xs font-bold">No digests yet</p>
                <p className="text-[10px] italic leading-relaxed">
                  Schedule an auto-publish digest on a channel, or publish one manually from the channel digest banner.
                </p>
              </div>
            ) : (
              <div className="divide-y">
                {knowledgeInbox.map(item => {
                  const isActive = selectedItem?.id === item.id
                  const digest = item.digest
                  const DeltaIcon = (digest?.delta ?? 0) > 0 ? TrendingUp : (digest?.delta ?? 0) < 0 ? TrendingDown : Zap
                  return (
                    <button
                      key={item.id}
                      onClick={() => handleSelect(item)}
                      className={cn(
                        "w-full text-left p-3.5 transition-colors relative group",
                        isActive ? "bg-emerald-500/10" : "hover:bg-muted/60",
                        !item.is_read && !isActive && "bg-emerald-500/5"
                      )}
                    >
                      {!item.is_read && (
                        <span className="absolute top-4 left-1.5 w-1.5 h-1.5 rounded-full bg-emerald-500" />
                      )}
                      <div className="pl-3 space-y-1">
                        <div className="flex items-center gap-1.5 text-[10px]">
                          <Hash className="w-2.5 h-2.5 text-muted-foreground" />
                          <span className="font-bold text-muted-foreground truncate">
                            {item.channel?.name || item.channel?.id}
                          </span>
                          {item.channel?.is_starred && (
                            <Star className="w-2.5 h-2.5 fill-amber-500 text-amber-500" />
                          )}
                          <span className="ml-auto text-muted-foreground font-medium">
                            {formatDistanceToNow(new Date(item.occurred_at), { addSuffix: true })}
                          </span>
                        </div>
                        <div className="flex items-center gap-1.5">
                          <Newspaper className="w-3 h-3 text-emerald-600 shrink-0" />
                          <p className={cn(
                            "text-xs leading-tight truncate",
                            !item.is_read ? "font-black" : "font-bold"
                          )}>
                            {digest?.headline || `${digest?.window || 'weekly'} digest · ${digest?.entries?.length || 0} entities`}
                          </p>
                        </div>
                        <div className="flex items-center gap-1.5 flex-wrap">
                          <Badge variant="outline" className="text-[8px] h-3.5 px-1 uppercase tracking-widest border-emerald-400/40 text-emerald-700">
                            {digest?.window}
                          </Badge>
                          {digest?.total_refs !== undefined && (
                            <span className="text-[10px] text-muted-foreground">
                              {digest.total_refs} refs
                            </span>
                          )}
                          {digest?.delta !== undefined && digest.delta !== 0 && (
                            <span className={cn(
                              "inline-flex items-center gap-0.5 text-[9px] font-black px-1 py-0.5 rounded-full",
                              digest.delta > 0 ? "bg-emerald-500/10 text-emerald-700" : "bg-rose-500/10 text-rose-700"
                            )}>
                              <DeltaIcon className="w-2 h-2" />
                              {digest.delta > 0 ? '+' : ''}{digest.delta}
                            </span>
                          )}
                          {item.message?.isPinned && (
                            <Pin className="w-2.5 h-2.5 text-amber-500 ml-auto" />
                          )}
                        </div>
                      </div>
                    </button>
                  )
                })}
              </div>
            )}
          </ScrollArea>
        </div>

        {/* Right column: selected digest detail */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {selectedItem ? (
            <>
              <div className="px-6 py-3 border-b bg-muted/10 flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-emerald-500/10 flex items-center justify-center shrink-0">
                  <Newspaper className="w-4 h-4 text-emerald-600" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground">
                    <Hash className="w-2.5 h-2.5" />
                    <span className="font-bold truncate">{selectedItem.channel?.name || selectedItem.channel?.id}</span>
                    <span>·</span>
                    <Clock className="w-2.5 h-2.5" />
                    {formatLocaleDate(selectedItem.occurred_at, locale, { year: "numeric", month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" })}
                  </div>
                  <p className="text-sm font-black truncate">
                    {selectedItem.digest?.headline || `${selectedItem.digest?.window} knowledge digest`}
                  </p>
                </div>
                <Button
                  size="sm"
                  className="h-7 text-[10px] font-black gap-1"
                  onClick={() => handleJumpToMessage(selectedItem)}
                >
                  <ChevronRight className="w-3 h-3" />
                  Jump to message
                </Button>
              </div>

              <ScrollArea className="flex-1">
                <div className="p-6 max-w-2xl space-y-6">
                  {isLoadingDetail ? (
                    <div className="py-8 flex items-center justify-center text-muted-foreground gap-2">
                      <Loader2 className="w-4 h-4 animate-spin" />
                      <span className="text-xs">Loading detail…</span>
                    </div>
                  ) : (
                    <>
                      {selectedItem.digest ? (
                        <KnowledgeDigestCard digest={selectedItem.digest} />
                      ) : (
                        <div className="text-muted-foreground italic text-sm">
                          This digest payload is no longer available.
                        </div>
                      )}

                      {/* Entity contexts from API detail */}
                      {detail?.entity_contexts && detail.entity_contexts.length > 0 && (
                        <div className="space-y-3">
                          <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-1">
                            <Zap className="w-3 h-3" /> Entity Activity
                          </p>
                          {detail.entity_contexts.map(ctx => (
                            <div key={ctx.entity_id} className="rounded-xl border bg-muted/10 p-4 space-y-2">
                              <div className="flex items-center justify-between">
                                <span className="text-xs font-black">{ctx.entity_title}</span>
                                <div className="flex items-center gap-1.5">
                                  <Badge variant="outline" className="text-[8px] h-3.5 px-1 uppercase tracking-widest">
                                    {ctx.entity_kind}
                                  </Badge>
                                  {ctx.delta !== 0 && (
                                    <span className={cn(
                                      "inline-flex items-center gap-0.5 text-[9px] font-black px-1.5 py-0.5 rounded-full",
                                      ctx.delta > 0 ? "bg-emerald-500/10 text-emerald-700" : "bg-rose-500/10 text-rose-700"
                                    )}>
                                      {ctx.delta > 0 ? <TrendingUp className="w-2.5 h-2.5" /> : <TrendingDown className="w-2.5 h-2.5" />}
                                      {ctx.delta > 0 ? '+' : ''}{ctx.delta}
                                    </span>
                                  )}
                                </div>
                              </div>
                              {ctx.messages.length > 0 && (
                                <div className="space-y-1">
                                  {ctx.messages.slice(0, 3).map(msg => (
                                    <div key={msg.id} className="text-[11px] text-muted-foreground border-l-2 border-muted pl-2 leading-relaxed">
                                      <MessageSquare className="w-2.5 h-2.5 inline mr-1" />
                                      {msg.content ? msg.content.replace(/<[^>]*>/g, '').slice(0, 120) : '(no content)'}
                                    </div>
                                  ))}
                                  {ctx.messages.length > 3 && (
                                    <p className="text-[10px] text-muted-foreground italic pl-4">
                                      +{ctx.messages.length - 3} more messages
                                    </p>
                                  )}
                                </div>
                              )}
                            </div>
                          ))}
                        </div>
                      )}

                      {/* Message snippet preview */}
                      {selectedItem.message?.content && (
                        <div className="rounded-xl border bg-muted/20 p-4">
                          <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground mb-2 flex items-center gap-1">
                            <Globe className="w-3 h-3" />
                            Published message
                          </p>
                          <div
                            className="text-xs leading-relaxed text-muted-foreground prose prose-xs"
                            dangerouslySetInnerHTML={{ __html: selectedItem.message.content }}
                          />
                        </div>
                      )}
                    </>
                  )}
                </div>
              </ScrollArea>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-muted-foreground/60">
              <div className="text-center space-y-2">
                <Inbox className="w-10 h-10 mx-auto" />
                <p className="text-xs italic">Select a digest to preview</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
