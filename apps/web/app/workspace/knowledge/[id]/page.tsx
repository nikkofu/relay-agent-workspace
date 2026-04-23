"use client"

import { use, useEffect, useRef, useState, Suspense } from "react"
import { useRouter } from "next/navigation"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import {
  Globe, ChevronLeft, Loader2, Edit2, CheckCircle2, X,
  BookOpen, Clock, Network, FileText, MessageSquare,
  Layout, Tag, ArrowRight, ArrowLeft, User2, Briefcase,
  Lightbulb, Building2, AlertCircle, Zap, Send, Plus, Share2,
  Sparkles, RefreshCw, DatabaseZap, CheckCheck, HelpCircle,
} from "lucide-react"
import { toast } from "sonner"
import { cn } from "@/lib/utils"
import { format } from "date-fns"
import type { KnowledgeEntity, KnowledgeEntityRef, KnowledgeEntityLink, KnowledgeEvent, KnowledgeGraph, KnowledgeGraphEdge } from "@/types"
import { EntityFollowButton } from "@/components/knowledge/entity-follow-button"
import { EntityActivitySparkline } from "@/components/knowledge/entity-activity-sparkline"

const KIND_CONFIG: Record<string, { label: string; icon: React.ElementType; color: string; badgeClass: string; bgClass: string }> = {
  person:       { label: "Person",       icon: User2,       color: "text-sky-600",    badgeClass: "bg-sky-500/10 text-sky-700 border-sky-300 dark:border-sky-700",         bgClass: "bg-sky-500/5" },
  project:      { label: "Project",      icon: Briefcase,   color: "text-violet-600", badgeClass: "bg-violet-500/10 text-violet-700 border-violet-300 dark:border-violet-700", bgClass: "bg-violet-500/5" },
  concept:      { label: "Concept",      icon: Lightbulb,   color: "text-amber-600",  badgeClass: "bg-amber-500/10 text-amber-700 border-amber-300 dark:border-amber-700",   bgClass: "bg-amber-500/5" },
  organization: { label: "Organization", icon: Building2,   color: "text-emerald-600",badgeClass: "bg-emerald-500/10 text-emerald-700 border-emerald-300 dark:border-emerald-700", bgClass: "bg-emerald-500/5" },
  file:         { label: "File",         icon: FileText,    color: "text-blue-600",   badgeClass: "bg-blue-500/10 text-blue-700 border-blue-300 dark:border-blue-700",       bgClass: "bg-blue-500/5" },
  artifact:     { label: "Artifact",     icon: Layout,      color: "text-orange-600", badgeClass: "bg-orange-500/10 text-orange-700 border-orange-300 dark:border-orange-700",bgClass: "bg-orange-500/5" },
}
const getKindConfig = (kind: string) =>
  KIND_CONFIG[kind] ?? { label: kind, icon: Tag, color: "text-muted-foreground", badgeClass: "bg-muted text-muted-foreground border-muted", bgClass: "bg-muted/20" }

const SOURCE_KIND_ICON: Record<string, React.ElementType> = {
  file: FileText, message: MessageSquare, artifact: Layout, thread: MessageSquare,
}
const EVENT_KIND_COLOR: Record<string, string> = {
  created: "bg-emerald-500", updated: "bg-sky-500", linked: "bg-violet-500",
  referenced: "bg-amber-500", archived: "bg-slate-400",
}

function EntityDetailContent({ id }: { id: string }) {
  const router = useRouter()
  const {
    fetchEntity, updateEntity, fetchEntityRefs, fetchEntityTimeline, fetchEntityLinks, fetchEntityGraph,
    ingestEvent, liveUpdate, spikingEntityIds, shareEntity,
    entityBriefs, isGeneratingBrief, generateEntityBrief, fetchEntityBrief,
    backfillStatuses, isBackfilling, fetchBackfillStatus, triggerBackfill,
    staleBriefs, entityAnswers, isAskingEntity, clearEntityAnswers,
    // Phase 63E
    fetchEntityAskHistory, askEntityStream, entityAskStreaming, isLoadingAskHistory,
  } = useKnowledgeStore()
  const [sharing, setSharing] = useState(false)
  const [askInput, setAskInput] = useState("")

  const handleShare = async () => {
    if (!id || sharing) return
    setSharing(true)
    const share = await shareEntity(id)
    setSharing(false)
    if (share?.url) {
      try {
        await navigator.clipboard.writeText(share.url)
        toast.success("Share link copied", { description: share.url })
      } catch {
        toast("Share link ready", { description: share.url })
      }
    }
  }

  const [entity, setEntity] = useState<KnowledgeEntity | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [refs, setRefs] = useState<KnowledgeEntityRef[]>([])
  const [timeline, setTimeline] = useState<KnowledgeEvent[]>([])
  const [links, setLinks] = useState<KnowledgeEntityLink[]>([])
  const [graph, setGraph] = useState<KnowledgeGraph | null>(null)
  const [tab, setTab] = useState("overview")
  const [isEditing, setIsEditing] = useState(false)
  const [editTitle, setEditTitle] = useState("")
  const [editSummary, setEditSummary] = useState("")
  const [editTags, setEditTags] = useState("")
  const [isSaving, setIsSaving] = useState(false)
  const [isLoadingRefs, setIsLoadingRefs] = useState(false)
  const [isLoadingTimeline, setIsLoadingTimeline] = useState(false)
  const [isLoadingGraph, setIsLoadingGraph] = useState(false)
  const [liveFlash, setLiveFlash] = useState<string | null>(null)
  const prevLiveTs = useRef<number>(0)
  const [showIngest, setShowIngest] = useState(false)
  const [ingestEventType, setIngestEventType] = useState("updated")
  const [ingestTitle, setIngestTitle] = useState("")
  const [ingestBody, setIngestBody] = useState("")
  const [ingestSourceKind, setIngestSourceKind] = useState("")
  const [isIngesting, setIsIngesting] = useState(false)

  useEffect(() => {
    if (!id) return
    setIsLoading(true)
    fetchEntity(id).then(e => {
      setEntity(e); setIsLoading(false)
      if (e) { setEditTitle(e.title); setEditSummary(e.summary || ""); setEditTags((e.tags || []).join(", ")) }
    })
    fetchBackfillStatus(id).catch(() => {})
    // Phase 62: hydrate cached brief without invoking the LLM
    fetchEntityBrief(id).catch(() => {})
    // Phase 63E: hydrate persisted Ask AI Q&A history (per-user, newest first)
    fetchEntityAskHistory(id, 20).catch(() => {})
  }, [id, fetchEntity, fetchBackfillStatus, fetchEntityBrief, fetchEntityAskHistory])

  useEffect(() => {
    if (!liveUpdate || liveUpdate.ts === prevLiveTs.current) return
    prevLiveTs.current = liveUpdate.ts
    if (liveUpdate.entityId !== id) return
    setLiveFlash(liveUpdate.type)
    setTimeout(() => setLiveFlash(null), 2500)
    if (liveUpdate.type === 'entity.updated') {
      setEntity(liveUpdate.payload)
    } else if (liveUpdate.type === 'ref.created') {
      setRefs(prev => prev.some(r => r.id === liveUpdate.payload.id) ? prev : [liveUpdate.payload, ...prev])
    } else if (liveUpdate.type === 'event.created') {
      setTimeline(prev => prev.some(e => e.id === liveUpdate.payload.id) ? prev : [...prev, liveUpdate.payload])
    } else if (liveUpdate.type === 'link.created') {
      setLinks(prev => prev.some(l => l.id === liveUpdate.payload.id) ? prev : [...prev, liveUpdate.payload])
    }
  }, [liveUpdate, id])

  const handleIngestEvent = async () => {
    if (!ingestTitle.trim()) return
    setIsIngesting(true)
    await ingestEvent({
      entity_id: id,
      event_type: ingestEventType,
      title: ingestTitle.trim(),
      body: ingestBody.trim() || undefined,
      source_kind: ingestSourceKind || undefined,
    })
    setIsIngesting(false)
    setShowIngest(false)
    setIngestTitle(""); setIngestBody(""); setIngestSourceKind("")
  }

  const handleTabChange = async (value: string) => {
    setTab(value)
    if (!id) return
    if (value === "refs" && refs.length === 0 && !isLoadingRefs) {
      setIsLoadingRefs(true); setRefs(await fetchEntityRefs(id)); setIsLoadingRefs(false)
    }
    if (value === "timeline" && timeline.length === 0 && !isLoadingTimeline) {
      setIsLoadingTimeline(true); setTimeline(await fetchEntityTimeline(id)); setIsLoadingTimeline(false)
    }
    if (value === "graph" && !graph && !isLoadingGraph) {
      setIsLoadingGraph(true)
      const [g, l] = await Promise.all([fetchEntityGraph(id), fetchEntityLinks(id)])
      setGraph(g); setLinks(l); setIsLoadingGraph(false)
    }
  }

  const handleSave = async () => {
    if (!entity || !id) return
    setIsSaving(true)
    const updated = await updateEntity(id, {
      title: editTitle.trim(), summary: editSummary.trim() || undefined,
      tags: editTags ? editTags.split(",").map(t => t.trim()).filter(Boolean) : [],
    })
    if (updated) { setEntity(updated); setIsEditing(false) }
    setIsSaving(false)
  }

  if (isLoading) return (
    <div className="flex-1 flex items-center justify-center bg-white dark:bg-[#1a1d21]">
      <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
    </div>
  )

  if (!entity) return (
    <div className="flex-1 flex flex-col items-center justify-center bg-white dark:bg-[#1a1d21] gap-3">
      <AlertCircle className="w-10 h-10 text-muted-foreground/30" />
      <p className="text-sm text-muted-foreground">Entity not found</p>
      <Button variant="ghost" size="sm" onClick={() => router.back()}><ChevronLeft className="w-4 h-4 mr-1" /> Back</Button>
    </div>
  )

  const cfg = getKindConfig(entity.kind)
  const Icon = cfg.icon

  return (
    <div className="flex-1 flex flex-col bg-white dark:bg-[#1a1d21] h-full overflow-hidden">
      <header className={cn("px-6 py-4 border-b shrink-0", cfg.bgClass)}>
        <div className="flex items-center gap-2 mb-3">
          <Button variant="ghost" size="sm" className="h-7 text-xs gap-1 text-muted-foreground -ml-2" onClick={() => router.push("/workspace/knowledge")}>
            <ChevronLeft className="w-3.5 h-3.5" /> Knowledge
          </Button>
        </div>
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-center gap-3 min-w-0">
            <div className="w-10 h-10 rounded-xl flex items-center justify-center shrink-0 bg-white/60 dark:bg-black/20 border">
              <Icon className={cn("w-5 h-5", cfg.color)} />
            </div>
            <div className="min-w-0">
              <h1 className="text-xl font-black truncate">{entity.title}</h1>
              <div className="flex items-center gap-2 mt-1 flex-wrap">
                <Badge className={cn("text-[9px] h-4 px-1.5", cfg.badgeClass)}>{cfg.label}</Badge>
                {entity.source_kind && <span className="text-[9px] text-muted-foreground uppercase font-black tracking-wider">{entity.source_kind}</span>}
                <span className="text-[9px] text-muted-foreground">Created {format(new Date(entity.created_at), "MMM d, yyyy")}</span>
                {liveFlash && (
                  <Badge className="text-[9px] h-4 px-1.5 gap-1 bg-emerald-500/10 text-emerald-700 border-emerald-300 animate-pulse">
                    <Zap className="w-2 h-2" /> Live
                  </Badge>
                )}
              </div>
            </div>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            <EntityActivitySparkline entityId={entity.id} days={30} width={180} height={38} className="hidden md:flex" />
            <EntityFollowButton entityId={entity.id} isSpiking={!!spikingEntityIds[entity.id]} />
            <Button variant="outline" size="sm" className="text-xs gap-1.5" onClick={handleShare} disabled={sharing} title="Copy shareable link">
              {sharing ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Share2 className="w-3.5 h-3.5" />}
              Share
            </Button>
            <Button variant="outline" size="sm" className="text-xs gap-1.5" onClick={() => setIsEditing(!isEditing)}>
              {isEditing ? <X className="w-3.5 h-3.5" /> : <Edit2 className="w-3.5 h-3.5" />}
              {isEditing ? "Cancel" : "Edit"}
            </Button>
          </div>
        </div>
        {!isEditing && (entity.tags || []).length > 0 && (
          <div className="flex flex-wrap gap-1.5 mt-3">
            {entity.tags!.map(tag => <span key={tag} className="text-[9px] px-1.5 py-0.5 rounded bg-white/50 dark:bg-black/20 border font-mono text-muted-foreground">#{tag}</span>)}
          </div>
        )}
      </header>

      <Tabs value={tab} onValueChange={handleTabChange} className="flex flex-col flex-1 overflow-hidden">
        <TabsList className="mx-6 mt-3 mb-0 h-8 shrink-0 bg-muted/40 w-fit">
          <TabsTrigger value="overview" className="text-xs h-7 px-3 gap-1"><Globe className="w-3 h-3" />Overview</TabsTrigger>
          <TabsTrigger value="refs" className="text-xs h-7 px-3 gap-1"><BookOpen className="w-3 h-3" />Refs{refs.length > 0 && <span className="ml-0.5 text-[10px] font-black">{refs.length}</span>}</TabsTrigger>
          <TabsTrigger value="timeline" className="text-xs h-7 px-3 gap-1"><Clock className="w-3 h-3" />Timeline{timeline.length > 0 && <span className="ml-0.5 text-[10px] font-black">{timeline.length}</span>}{liveFlash === 'event.created' && <Zap className="w-2.5 h-2.5 text-emerald-500 ml-0.5" />}</TabsTrigger>
          <TabsTrigger value="graph" className="text-xs h-7 px-3 gap-1"><Network className="w-3 h-3" />Graph</TabsTrigger>
        </TabsList>

        <ScrollArea className="flex-1">
          <TabsContent value="overview" className="p-6 m-0 space-y-5">
            {isEditing ? (
              <div className="space-y-4">
                <div className="space-y-1.5"><Label className="text-xs font-bold uppercase tracking-wider">Title</Label><Input value={editTitle} onChange={e => setEditTitle(e.target.value)} /></div>
                <div className="space-y-1.5"><Label className="text-xs font-bold uppercase tracking-wider">Summary</Label><Textarea className="resize-none h-24 text-sm" value={editSummary} onChange={e => setEditSummary(e.target.value)} /></div>
                <div className="space-y-1.5"><Label className="text-xs font-bold uppercase tracking-wider">Tags</Label><Input value={editTags} onChange={e => setEditTags(e.target.value)} placeholder="tag1, tag2" /></div>
                <div className="flex gap-2">
                  <Button size="sm" className="font-bold gap-1.5" disabled={isSaving} onClick={handleSave}>
                    {isSaving ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <CheckCircle2 className="w-3.5 h-3.5" />}Save
                  </Button>
                  <Button variant="ghost" size="sm" onClick={() => setIsEditing(false)}>Cancel</Button>
                </div>
              </div>
            ) : (
              <>
                {entity.summary ? (
                  <div className="rounded-xl bg-muted/20 border p-4"><p className="text-sm text-foreground/80 leading-relaxed">{entity.summary}</p></div>
                ) : (
                  <div className="rounded-xl bg-muted/10 border border-dashed p-6 text-center"><p className="text-xs text-muted-foreground italic">No summary yet. Click Edit to add one.</p></div>
                )}
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1"><p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Kind</p><p className="text-sm font-bold capitalize">{entity.kind}</p></div>
                  {entity.source_kind && <div className="space-y-1"><p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Source Kind</p><p className="text-sm font-bold capitalize">{entity.source_kind}</p></div>}
                  {entity.ref_count !== undefined && <div className="space-y-1"><p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Evidence Refs</p><p className="text-sm font-bold">{entity.ref_count}</p></div>}
                  {entity.updated_at && <div className="space-y-1"><p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Last Updated</p><p className="text-sm font-bold">{format(new Date(entity.updated_at), "PPp")}</p></div>}
                </div>

                {/* Phase 61/63A: AI Brief Card with stale pulse */}
                {(() => {
                  const brief = entityBriefs[id]
                  const generating = !!isGeneratingBrief[id]
                  const stale = staleBriefs[id]
                  return (
                    <div className={cn(
                      "rounded-xl border bg-gradient-to-br from-violet-500/5 to-indigo-500/5 p-4 space-y-3 transition-shadow",
                      stale && "ring-2 ring-amber-400/40 shadow-[0_0_0_4px_rgba(251,191,36,0.08)]"
                    )}>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <Sparkles className="w-3.5 h-3.5 text-violet-500" />
                          <p className="text-[10px] font-black uppercase tracking-widest text-violet-600 dark:text-violet-400">AI Brief</p>
                          {brief && <span className="text-[9px] text-muted-foreground">{format(new Date(brief.generated_at), "MMM d")}</span>}
                          {stale && (
                            <span className="inline-flex items-center gap-1 text-[9px] font-bold text-amber-700 dark:text-amber-400 bg-amber-500/10 border border-amber-400/30 rounded px-1.5 py-0.5">
                              <AlertCircle className="w-2.5 h-2.5" /> Stale
                            </span>
                          )}
                        </div>
                        <Button
                          variant="outline"
                          size="sm"
                          className={cn(
                            "h-6 text-[10px] gap-1 px-2",
                            stale
                              ? "border-amber-400/50 text-amber-700 dark:text-amber-400 hover:bg-amber-500/10"
                              : "border-violet-400/40 text-violet-700 dark:text-violet-400 hover:bg-violet-500/10"
                          )}
                          disabled={generating}
                          onClick={() => generateEntityBrief(id, !!brief)}
                        >
                          {generating ? <Loader2 className="w-3 h-3 animate-spin" /> : brief ? <RefreshCw className="w-3 h-3" /> : <Sparkles className="w-3 h-3" />}
                          {generating ? 'Generating…' : stale ? 'Refresh' : brief ? 'Regenerate' : 'Generate'}
                        </Button>
                      </div>
                      {stale?.reason && (
                        <p className="text-[10px] text-amber-700 dark:text-amber-400 italic">
                          Brief may be outdated: {stale.reason}
                        </p>
                      )}
                      {brief ? (
                        <p className="text-xs text-foreground/90 leading-relaxed whitespace-pre-wrap">{brief.content}</p>
                      ) : (
                        <p className="text-xs text-muted-foreground italic">Generate an AI-powered brief grounded in this entity&apos;s refs and timeline.</p>
                      )}
                      {brief && (brief.ref_count !== undefined || brief.event_count !== undefined) && (
                        <div className="flex items-center gap-3 text-[10px] text-muted-foreground">
                          {brief.ref_count !== undefined && <span>{brief.ref_count} refs</span>}
                          {brief.event_count !== undefined && <span>{brief.event_count} events</span>}
                          {brief.provider && <span className="ml-auto font-mono">{brief.provider}{brief.model ? ` / ${brief.model}` : ''}</span>}
                        </div>
                      )}
                    </div>
                  )
                })()}

                {/* Phase 63A/E: Ask AI module — streaming + persisted history */}
                {(() => {
                  const answers = entityAnswers[id] || []
                  const asking = !!isAskingEntity[id]
                  const loadingHistory = !!isLoadingAskHistory[id]
                  const streaming = entityAskStreaming[id]
                  const submit = async () => {
                    const q = askInput.trim()
                    if (!q || asking) return
                    setAskInput("")
                    // Phase 63E: prefer SSE streaming; the store falls back to
                    // the sync /ask endpoint on non-OK status or network error.
                    await askEntityStream(id, q)
                  }
                  return (
                    <div className="rounded-xl border bg-gradient-to-br from-sky-500/5 to-cyan-500/5 p-4 space-y-3">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <HelpCircle className="w-3.5 h-3.5 text-sky-500" />
                          <p className="text-[10px] font-black uppercase tracking-widest text-sky-600 dark:text-sky-400">Ask AI</p>
                          {answers.length > 0 && (
                            <span className="text-[9px] text-muted-foreground">{answers.length} answered</span>
                          )}
                          {loadingHistory && (
                            <span className="text-[9px] text-muted-foreground inline-flex items-center gap-1">
                              <Loader2 className="w-2.5 h-2.5 animate-spin" /> loading history
                            </span>
                          )}
                        </div>
                        {answers.length > 0 && (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-6 text-[10px] px-2 text-muted-foreground"
                            onClick={() => clearEntityAnswers(id)}
                          >
                            Clear
                          </Button>
                        )}
                      </div>
                      <div className="flex items-center gap-2">
                        <Input
                          placeholder="Ask a question grounded in this entity…"
                          className="h-8 text-xs"
                          value={askInput}
                          onChange={e => setAskInput(e.target.value)}
                          onKeyDown={e => {
                            if (e.key === 'Enter' && !e.shiftKey) {
                              e.preventDefault()
                              submit()
                            }
                          }}
                          disabled={asking}
                        />
                        <Button
                          size="sm"
                          className="h-8 text-xs gap-1.5 font-bold"
                          disabled={asking || !askInput.trim()}
                          onClick={submit}
                        >
                          {asking ? <Loader2 className="w-3 h-3 animate-spin" /> : <Send className="w-3 h-3" />}
                          Ask
                        </Button>
                      </div>

                      {/* Phase 63E: progressive SSE streaming card — visible until 'done' fires */}
                      {streaming && (
                        <div className="rounded-lg border border-sky-400/40 bg-background/60 p-3 space-y-1.5">
                          <div className="flex items-center gap-1.5">
                            <Loader2 className="w-3 h-3 animate-spin text-sky-500" />
                            <span className="text-[9px] font-black uppercase tracking-widest text-sky-700 dark:text-sky-400">Streaming…</span>
                            {(streaming.provider || streaming.model) && (
                              <span className="text-[9px] font-mono text-muted-foreground ml-auto">
                                {streaming.provider}{streaming.model ? ` / ${streaming.model}` : ''}
                              </span>
                            )}
                          </div>
                          <p className="text-[11px] font-bold text-sky-700 dark:text-sky-400">Q: {streaming.question}</p>
                          <p className="text-xs text-foreground/85 leading-relaxed whitespace-pre-wrap">
                            {streaming.text || <span className="text-muted-foreground italic">Thinking…</span>}
                            <span className="inline-block w-1.5 h-3 align-text-bottom bg-sky-500/70 animate-pulse ml-0.5" />
                          </p>
                        </div>
                      )}

                      {answers.length > 0 && (
                        <div className="space-y-3 pt-1">
                          {answers.map((a, i) => {
                            // Phase 63E: rows hydrated from /ask/history carry citation_count but no citations[]
                            const isHydratedHistory = (a.citations?.length ?? 0) === 0 && typeof a.citation_count === 'number' && a.citation_count > 0
                            const key = a.history_id || `${a.question}-${a.answered_at}-${i}`
                            return (
                              <div key={key} className="rounded-lg border bg-background/60 p-3 space-y-2">
                                <p className="text-[11px] font-bold text-sky-700 dark:text-sky-400">Q: {a.question}</p>
                                <p className="text-xs text-foreground/90 leading-relaxed whitespace-pre-wrap">{a.answer}</p>
                                {a.citations && a.citations.length > 0 && (
                                  <div className="space-y-1 pt-1 border-t">
                                    <p className="text-[9px] font-black uppercase tracking-widest text-muted-foreground">Citations ({a.citations.length})</p>
                                    {a.citations.slice(0, 5).map((c, j) => (
                                      <div key={c.id || j} className="flex items-start gap-1.5 text-[10px]">
                                        <Badge variant="secondary" className="text-[9px] h-4 px-1.5 shrink-0">{c.source_kind || c.ref_kind}</Badge>
                                        <span className="text-foreground/75 italic line-clamp-2">&ldquo;{c.snippet}&rdquo;</span>
                                      </div>
                                    ))}
                                  </div>
                                )}
                                {isHydratedHistory && (
                                  <div className="pt-1 border-t">
                                    <p className="text-[9px] font-black uppercase tracking-widest text-muted-foreground">
                                      {a.citation_count} citation{a.citation_count === 1 ? '' : 's'} · from history
                                    </p>
                                  </div>
                                )}
                                <div className="flex items-center gap-2 text-[9px] text-muted-foreground">
                                  <span>{format(new Date(a.answered_at), "MMM d, HH:mm")}</span>
                                  {a.history_id && (
                                    <span className="font-mono text-muted-foreground/60" title={`History id: ${a.history_id}`}>
                                      #{a.history_id.slice(-6)}
                                    </span>
                                  )}
                                  {a.provider && <span className="font-mono ml-auto">{a.provider}{a.model ? ` / ${a.model}` : ''}</span>}
                                </div>
                              </div>
                            )
                          })}
                        </div>
                      )}
                    </div>
                  )
                })()}
              </>
            )}
          </TabsContent>

          <TabsContent value="refs" className="p-6 m-0 space-y-3">
            {isLoadingRefs && <div className="flex items-center gap-2 text-xs text-muted-foreground py-4"><Loader2 className="w-4 h-4 animate-spin" />Loading refs...</div>}
            {!isLoadingRefs && refs.length === 0 && <div className="py-12 text-center space-y-2"><BookOpen className="w-8 h-8 text-muted-foreground/20 mx-auto" /><p className="text-xs text-muted-foreground italic">No evidence refs yet.</p></div>}
            {refs.map((ref, i) => {
              const RefIcon = SOURCE_KIND_ICON[ref.source_kind] ?? FileText
              return (
                <div key={ref.id || i} className="rounded-xl border bg-muted/10 p-3 space-y-2">
                  <div className="flex items-center gap-2">
                    <Badge variant="secondary" className="gap-1 text-[9px] h-4 px-1.5"><RefIcon className="w-2.5 h-2.5" />{ref.source_kind}</Badge>
                    <span className="text-[10px] text-muted-foreground font-mono truncate">{ref.source_id}</span>
                    <span className="ml-auto text-[9px] text-muted-foreground">{format(new Date(ref.created_at), "MMM d")}</span>
                  </div>
                  {ref.snippet && <p className="text-[11px] text-foreground/80 leading-relaxed bg-muted/30 rounded px-2 py-1.5 italic">&ldquo;{ref.snippet}&rdquo;</p>}
                </div>
              )
            })}
          </TabsContent>

          <TabsContent value="timeline" className="p-6 m-0">
            <div className="flex items-center justify-between mb-4">
              <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Event Log</p>
              <div className="flex items-center gap-2">
                {/* Phase 61: Backfill status + trigger */}
                {(() => {
                  const bs = backfillStatuses[id]
                  const backfilling = !!isBackfilling[id]
                  if (!bs) return null
                  return bs.is_backfilled ? (
                    <div className="flex items-center gap-1 text-[9px] text-emerald-600 font-bold">
                      <CheckCheck className="w-3 h-3" /> Backfill complete
                    </div>
                  ) : (
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-7 text-[10px] gap-1.5 border-amber-400/40 text-amber-700 dark:text-amber-400 hover:bg-amber-500/10"
                      disabled={backfilling}
                      onClick={() => triggerBackfill(id)}
                    >
                      {backfilling ? <Loader2 className="w-3 h-3 animate-spin" /> : <DatabaseZap className="w-3 h-3" />}
                      {backfilling ? 'Backfilling…' : `Backfill (${bs.missing_ref_count} missing)`}
                    </Button>
                  )
                })()}
                <Button variant="outline" size="sm" className="h-7 text-xs gap-1.5 font-bold" onClick={() => setShowIngest(true)}>
                  <Plus className="w-3 h-3" /> Log Event
                </Button>
              </div>
            </div>
            {isLoadingTimeline && <div className="flex items-center gap-2 text-xs text-muted-foreground py-4"><Loader2 className="w-4 h-4 animate-spin" />Loading...</div>}
            {!isLoadingTimeline && timeline.length === 0 && <div className="py-12 text-center space-y-2"><Clock className="w-8 h-8 text-muted-foreground/20 mx-auto" /><p className="text-xs text-muted-foreground italic">No timeline events yet.</p></div>}
            {timeline.length > 0 && (
              <div className="relative pl-4 space-y-0">
                <div className="absolute left-[7px] top-2 bottom-2 w-0.5 bg-border" />
                {timeline.map((event, i) => (
                  <div key={event.id || i} className="relative pl-6 pb-5">
                    <div className={cn("absolute left-0 top-1.5 w-3 h-3 rounded-full border-2 border-background", EVENT_KIND_COLOR[event.event_kind] || "bg-muted-foreground")} />
                    <div className="space-y-0.5">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="font-bold text-xs">{event.title}</span>
                        <Badge variant="secondary" className="text-[9px] h-4 px-1.5">{event.event_kind}</Badge>
                        <span className="ml-auto text-[9px] text-muted-foreground">{format(new Date(event.occurred_at), "PPp")}</span>
                      </div>
                      {event.description && <p className="text-[11px] text-muted-foreground leading-relaxed">{event.description}</p>}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </TabsContent>

          <TabsContent value="graph" className="p-6 m-0 space-y-4">
            {isLoadingGraph && <div className="flex items-center gap-2 text-xs text-muted-foreground py-4"><Loader2 className="w-4 h-4 animate-spin" />Loading graph...</div>}
            {!isLoadingGraph && !graph && links.length === 0 && (
              <div className="py-12 text-center space-y-2">
                <Network className="w-8 h-8 text-muted-foreground/20 mx-auto" />
                <p className="text-xs text-muted-foreground italic">No connections yet.</p>
              </div>
            )}
            {(graph || links.length > 0) && (
              <>
                <div className="flex items-center justify-center py-2">
                  <div className={cn("rounded-xl border-2 px-4 py-2.5 flex items-center gap-2 font-bold text-sm", cfg.bgClass, "border-current/20")}>
                    <Icon className={cn("w-4 h-4", cfg.color)} />{entity.title}
                    <Badge className={cn("text-[9px] h-4 px-1.5 ml-1", cfg.badgeClass)}>{cfg.label}</Badge>
                  </div>
                </div>

                {/* Phase 47: graph.edges — richer weighted directional edges */}
                {graph?.edges && graph.edges.length > 0 && (
                  <div className="space-y-2">
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-1"><Network className="w-3 h-3" /> Edges</p>
                    {graph.edges.map((edge: KnowledgeGraphEdge, i: number) => {
                      const isOut = edge.from_id === id
                      const otherId = isOut ? edge.to_id : edge.from_id
                      const otherNode = graph.nodes.find(n => n.entity.id === otherId)
                      const weight = edge.weight ?? 1
                      const weightBar = Math.min(Math.round(weight * 10), 10)
                      return (
                        <div key={`edge-${i}`} className="flex items-center gap-2 rounded-lg border bg-muted/10 p-2.5">
                          {!isOut && <ArrowLeft className="w-3 h-3 text-muted-foreground shrink-0" />}
                          <button className="text-xs font-bold text-blue-600 hover:underline truncate" onClick={() => router.push("/workspace/knowledge/" + otherId)}>
                            {otherNode?.entity.title || otherId}
                          </button>
                          {isOut && <ArrowRight className="w-3 h-3 text-muted-foreground shrink-0" />}
                          <div className="flex items-center gap-1 ml-auto flex-wrap">
                            <Badge variant="secondary" className="text-[9px] h-4 px-1.5 shrink-0">{edge.rel}</Badge>
                            {edge.role && <Badge variant="outline" className="text-[9px] h-4 px-1.5 shrink-0">{edge.role}</Badge>}
                            {edge.direction && <Badge variant="outline" className="text-[9px] h-4 px-1 shrink-0">{edge.direction}</Badge>}
                            {weight !== 1 && (
                              <div className="flex items-center gap-0.5" title={`weight: ${weight}`}>
                                {Array.from({ length: weightBar }).map((_, j) => (
                                  <div key={j} className="w-1 h-2.5 rounded-sm bg-emerald-500/60" />
                                ))}
                              </div>
                            )}
                          </div>
                        </div>
                      )
                    })}
                  </div>
                )}

                {/* Legacy links */}
                {!graph?.edges && links.filter(l => l.from_entity_id === id).length > 0 && (
                  <div className="space-y-2">
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-1"><ArrowRight className="w-3 h-3" /> Outgoing</p>
                    {links.filter(l => l.from_entity_id === id).map((link, i) => (
                      <div key={link.id || i} className="flex items-center gap-2 rounded-lg border bg-muted/10 p-2.5">
                        <Badge variant="secondary" className="text-[9px] h-4 px-1.5 shrink-0">{link.rel}</Badge>
                        <ArrowRight className="w-3 h-3 text-muted-foreground shrink-0" />
                        <button className="text-xs font-bold text-blue-600 hover:underline truncate" onClick={() => router.push("/workspace/knowledge/" + link.to_entity_id)}>
                          {graph?.nodes.find(n => n.entity.id === link.to_entity_id)?.entity.title || link.to_entity_id}
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                {!graph?.edges && links.filter(l => l.to_entity_id === id).length > 0 && (
                  <div className="space-y-2">
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-1"><ArrowLeft className="w-3 h-3" /> Incoming</p>
                    {links.filter(l => l.to_entity_id === id).map((link, i) => (
                      <div key={link.id || i} className="flex items-center gap-2 rounded-lg border bg-muted/10 p-2.5">
                        <button className="text-xs font-bold text-blue-600 hover:underline truncate" onClick={() => router.push("/workspace/knowledge/" + link.from_entity_id)}>
                          {graph?.nodes.find(n => n.entity.id === link.from_entity_id)?.entity.title || link.from_entity_id}
                        </button>
                        <ArrowRight className="w-3 h-3 text-muted-foreground shrink-0" />
                        <Badge variant="secondary" className="text-[9px] h-4 px-1.5 shrink-0">{link.rel}</Badge>
                      </div>
                    ))}
                  </div>
                )}

                {graph && graph.nodes.length > 0 && (
                  <div className="space-y-2">
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-1"><Network className="w-3 h-3" /> Related Entities</p>
                    <div className="grid grid-cols-2 gap-2">
                      {graph.nodes.map((node, i) => {
                        const nodeCfg = getKindConfig(node.entity.kind)
                        const NodeIcon = nodeCfg.icon
                        return (
                          <button key={node.entity.id || i} className={cn("text-left rounded-lg border p-2.5 hover:bg-muted/30 transition-colors space-y-1", nodeCfg.bgClass)} onClick={() => router.push("/workspace/knowledge/" + node.entity.id)}>
                            <div className="flex items-center gap-1.5"><NodeIcon className={cn("w-3.5 h-3.5 shrink-0", nodeCfg.color)} /><span className="text-xs font-bold truncate">{node.entity.title}</span></div>
                            <div className="flex items-center gap-1.5 flex-wrap">
                              <Badge className={cn("text-[8px] h-3.5 px-1", nodeCfg.badgeClass)}>{nodeCfg.label}</Badge>
                              {node.rel && <span className="text-[9px] text-muted-foreground">{node.rel}</span>}
                              {node.role && <Badge variant="outline" className="text-[8px] h-3.5 px-1">{node.role}</Badge>}
                              {node.source_kind && <span className="text-[9px] text-muted-foreground font-mono">{node.source_kind}</span>}
                              {node.weight !== undefined && node.weight !== 1 && (
                                <span className="text-[9px] text-muted-foreground">w:{node.weight.toFixed(1)}</span>
                              )}
                            </div>
                          </button>
                        )
                      })}
                    </div>
                  </div>
                )}
              </>
            )}
          </TabsContent>
        </ScrollArea>
      </Tabs>

      {/* Event Ingest Dialog */}
      <Dialog open={showIngest} onOpenChange={setShowIngest}>
        <DialogContent className="sm:max-w-[420px]">
          <DialogHeader>
            <DialogTitle className="text-base font-black flex items-center gap-2">
              <Zap className="w-4 h-4 text-emerald-600" /> Log Live Event
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-3 py-2">
            <div className="space-y-1.5">
              <Label className="text-xs font-bold uppercase tracking-wider">Event Type</Label>
              <Select value={ingestEventType} onValueChange={setIngestEventType}>
                <SelectTrigger className="h-8 text-sm"><SelectValue /></SelectTrigger>
                <SelectContent>
                  {["created", "updated", "linked", "referenced", "archived", "milestone", "alert"].map(t => (
                    <SelectItem key={t} value={t} className="text-sm">{t}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs font-bold uppercase tracking-wider">Title</Label>
              <Input
                placeholder="Event title..."
                value={ingestTitle}
                onChange={e => setIngestTitle(e.target.value)}
                onKeyDown={e => e.key === "Enter" && handleIngestEvent()}
              />
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs font-bold uppercase tracking-wider">Body <span className="font-normal text-muted-foreground normal-case">(optional)</span></Label>
              <Textarea className="resize-none h-20 text-sm" placeholder="Event details..." value={ingestBody} onChange={e => setIngestBody(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs font-bold uppercase tracking-wider">Source Kind <span className="font-normal text-muted-foreground normal-case">(optional)</span></Label>
              <Input placeholder="message / file / artifact..." value={ingestSourceKind} onChange={e => setIngestSourceKind(e.target.value)} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" size="sm" onClick={() => setShowIngest(false)}>Cancel</Button>
            <Button size="sm" className="font-bold gap-1.5" disabled={!ingestTitle.trim() || isIngesting} onClick={handleIngestEvent}>
              {isIngesting ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Send className="w-3.5 h-3.5" />}
              Ingest
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function EntityDetailLoader({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  return <EntityDetailContent id={id} />
}

export default function EntityDetailPage({ params }: { params: Promise<{ id: string }> }) {
  return (
    <Suspense fallback={
      <div className="flex-1 flex items-center justify-center bg-white dark:bg-[#1a1d21]">
        <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
      </div>
    }>
      <EntityDetailLoader params={params} />
    </Suspense>
  )
}
