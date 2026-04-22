"use client"

import { useEffect, useRef, useState } from "react"
import { useRouter } from "next/navigation"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Globe, Search, Plus, Loader2, User2, Briefcase, Lightbulb, Building2,
  FileText, Layout, Tag, Hash, ChevronRight, BookOpen, Zap, Bell,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { format } from "date-fns"
import { EntityFollowButton } from "@/components/knowledge/entity-follow-button"

const KIND_CONFIG: Record<string, { label: string; icon: React.ElementType; color: string; badgeClass: string }> = {
  person:       { label: 'Person',       icon: User2,       color: 'text-sky-600',    badgeClass: 'bg-sky-500/10 text-sky-700 border-sky-300 dark:border-sky-700' },
  project:      { label: 'Project',      icon: Briefcase,   color: 'text-violet-600', badgeClass: 'bg-violet-500/10 text-violet-700 border-violet-300 dark:border-violet-700' },
  concept:      { label: 'Concept',      icon: Lightbulb,   color: 'text-amber-600',  badgeClass: 'bg-amber-500/10 text-amber-700 border-amber-300 dark:border-amber-700' },
  organization: { label: 'Organization', icon: Building2,   color: 'text-emerald-600',badgeClass: 'bg-emerald-500/10 text-emerald-700 border-emerald-300 dark:border-emerald-700' },
  file:         { label: 'File',         icon: FileText,    color: 'text-blue-600',   badgeClass: 'bg-blue-500/10 text-blue-700 border-blue-300 dark:border-blue-700' },
  artifact:     { label: 'Artifact',     icon: Layout,      color: 'text-orange-600', badgeClass: 'bg-orange-500/10 text-orange-700 border-orange-300 dark:border-orange-700' },
}

const getKindConfig = (kind: string) =>
  KIND_CONFIG[kind] ?? { label: kind, icon: Tag, color: 'text-muted-foreground', badgeClass: 'bg-muted text-muted-foreground border-muted' }

const KIND_OPTIONS = ['person', 'project', 'concept', 'organization', 'file', 'artifact']

export default function KnowledgePage() {
  const router = useRouter()
  const { entities, isLoading, fetchEntities, createEntity, liveUpdate, fetchFollowedEntities, followedEntityIds, followedEntities } = useKnowledgeStore()
  const [q, setQ] = useState("")
  const [filterKind, setFilterKind] = useState("all")
  const [onlyFollowed, setOnlyFollowed] = useState(false)
  const [showCreate, setShowCreate] = useState(false)
  const [newTitle, setNewTitle] = useState("")
  const [newKind, setNewKind] = useState("concept")
  const [newSummary, setNewSummary] = useState("")
  const [newTags, setNewTags] = useState("")
  const [isCreating, setIsCreating] = useState(false)
  const [liveFlash, setLiveFlash] = useState(false)
  const prevLiveUpdate = useRef<number>(0)

  useEffect(() => { fetchEntities() }, [fetchEntities])
  useEffect(() => { fetchFollowedEntities() }, [fetchFollowedEntities])

  useEffect(() => {
    if (!liveUpdate) return
    if (liveUpdate.ts === prevLiveUpdate.current) return
    if (liveUpdate.type === 'entity.created' || liveUpdate.type === 'entity.updated') {
      prevLiveUpdate.current = liveUpdate.ts
      setLiveFlash(true)
      setTimeout(() => setLiveFlash(false), 2000)
    }
  }, [liveUpdate])

  const filtered = entities.filter(e => {
    const matchesQ = !q || e.title.toLowerCase().includes(q.toLowerCase()) || e.summary?.toLowerCase().includes(q.toLowerCase())
    const matchesKind = filterKind === 'all' || e.kind === filterKind
    const matchesFollowed = !onlyFollowed || !!followedEntityIds[e.id]
    return matchesQ && matchesKind && matchesFollowed
  })

  const kindCounts = entities.reduce<Record<string, number>>((acc, e) => {
    acc[e.kind] = (acc[e.kind] || 0) + 1
    return acc
  }, {})

  const handleCreate = async () => {
    if (!newTitle.trim()) return
    setIsCreating(true)
    const entity = await createEntity({
      title: newTitle.trim(),
      kind: newKind,
      summary: newSummary.trim() || undefined,
      tags: newTags ? newTags.split(',').map(t => t.trim()).filter(Boolean) : undefined,
    })
    setIsCreating(false)
    if (entity) {
      setShowCreate(false)
      setNewTitle(""); setNewSummary(""); setNewTags("")
      router.push(`/workspace/knowledge/${entity.id}`)
    }
  }

  return (
    <div className="flex-1 flex flex-col bg-white dark:bg-[#1a1d21] h-full overflow-hidden">
      <header className="h-14 px-6 flex items-center border-b shrink-0 justify-between">
        <div className="flex items-center gap-2">
          <Globe className="w-5 h-5 text-emerald-600" />
          <h1 className="text-lg font-black tracking-tight uppercase">Knowledge</h1>
          {entities.length > 0 && (
            <Badge variant="secondary" className="text-[9px] font-black h-5 px-2">{entities.length} entities</Badge>
          )}
          {liveFlash && (
            <Badge className="text-[9px] font-black h-5 px-2 gap-1 bg-emerald-500/10 text-emerald-700 border-emerald-300 animate-pulse">
              <Zap className="w-2.5 h-2.5" /> Live
            </Badge>
          )}
        </div>
        <Button size="sm" className="text-xs font-bold gap-1.5" onClick={() => setShowCreate(true)}>
          <Plus className="w-3.5 h-3.5" /> New Entity
        </Button>
      </header>

      {/* Filters */}
      <div className="px-6 py-3 border-b bg-muted/10 flex flex-wrap gap-2 items-center">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search entities..."
            className="pl-9 bg-white dark:bg-[#1a1d21] h-8 text-sm"
            value={q}
            onChange={e => setQ(e.target.value)}
          />
        </div>
        <div className="flex items-center gap-1.5 flex-wrap">
          <Button
            variant={onlyFollowed ? "secondary" : "ghost"}
            size="sm"
            className={cn(
              "h-7 text-xs gap-1 font-bold",
              onlyFollowed && "ring-1 ring-purple-400 bg-purple-500/10 text-purple-700 dark:text-purple-400"
            )}
            onClick={() => setOnlyFollowed(!onlyFollowed)}
          >
            <Bell className="w-3 h-3" /> Following ({followedEntities.length})
          </Button>
          <div className="h-5 w-px bg-border mx-0.5" />
          <Button
            variant={filterKind === 'all' ? "secondary" : "ghost"}
            size="sm" className="h-7 text-xs gap-1 font-bold"
            onClick={() => setFilterKind('all')}
          >
            <Hash className="w-3 h-3" /> All ({entities.length})
          </Button>
          {Object.entries(kindCounts).map(([kind, count]) => {
            const cfg = getKindConfig(kind)
            return (
              <Button
                key={kind}
                variant={filterKind === kind ? "secondary" : "ghost"}
                size="sm" className="h-7 text-xs gap-1 font-bold"
                onClick={() => setFilterKind(filterKind === kind ? 'all' : kind)}
              >
                <cfg.icon className={cn("w-3 h-3", cfg.color)} />
                {cfg.label} ({count})
              </Button>
            )
          })}
        </div>
      </div>

      <ScrollArea className="flex-1">
        {isLoading && (
          <div className="flex items-center justify-center py-20 gap-2 text-muted-foreground">
            <Loader2 className="w-5 h-5 animate-spin" />
            <span className="text-sm">Loading entities...</span>
          </div>
        )}
        {!isLoading && filtered.length === 0 && (
          <div className="py-24 flex flex-col items-center gap-4 text-center">
            <div className="w-16 h-16 rounded-2xl bg-emerald-500/10 flex items-center justify-center">
              <Globe className="w-8 h-8 text-emerald-400" />
            </div>
            <div className="space-y-1 max-w-xs">
              <p className="font-black uppercase text-xs tracking-widest text-muted-foreground">Knowledge Base</p>
              <p className="text-[11px] text-muted-foreground italic leading-relaxed">
                {onlyFollowed && followedEntities.length === 0
                  ? "You aren't following any entities yet. Click the Follow button on any entity to subscribe to its updates."
                  : q || filterKind !== 'all' || onlyFollowed
                    ? "No entities match your filters."
                    : "No entities yet. Create the first one to start building your knowledge graph."}
              </p>
            </div>
            {!q && filterKind === 'all' && !onlyFollowed && (
              <Button size="sm" className="text-xs font-bold gap-1.5" onClick={() => setShowCreate(true)}>
                <Plus className="w-3.5 h-3.5" /> New Entity
              </Button>
            )}
          </div>
        )}
        {!isLoading && filtered.length > 0 && (
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {filtered.map(entity => {
              const cfg = getKindConfig(entity.kind)
              const Icon = cfg.icon
              return (
                <div
                  key={entity.id}
                  className="group rounded-xl border bg-muted/10 hover:bg-muted/30 p-4 cursor-pointer transition-colors space-y-3"
                  onClick={() => router.push(`/workspace/knowledge/${entity.id}`)}
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex items-center gap-2 min-w-0">
                      <div className={cn("w-8 h-8 rounded-lg flex items-center justify-center shrink-0 bg-muted/30")}>
                        <Icon className={cn("w-4 h-4", cfg.color)} />
                      </div>
                      <div className="min-w-0">
                        <p className="font-bold text-sm truncate">{entity.title}</p>
                        <Badge className={cn("text-[9px] h-4 px-1.5 mt-0.5", cfg.badgeClass)}>{cfg.label}</Badge>
                      </div>
                    </div>
                    <div className="flex items-center gap-1.5 shrink-0">
                      <EntityFollowButton entityId={entity.id} variant="chip" />
                      <ChevronRight className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
                    </div>
                  </div>

                  {entity.summary && (
                    <p className="text-[11px] text-muted-foreground leading-relaxed line-clamp-2">{entity.summary}</p>
                  )}

                  <div className="flex items-center gap-2 flex-wrap">
                    {(entity.tags || []).slice(0, 3).map(tag => (
                      <span key={tag} className="text-[9px] px-1.5 py-0.5 rounded bg-muted font-mono text-muted-foreground">#{tag}</span>
                    ))}
                    {entity.ref_count !== undefined && entity.ref_count > 0 && (
                      <span className="ml-auto text-[10px] text-muted-foreground flex items-center gap-1">
                        <BookOpen className="w-3 h-3" />{entity.ref_count} refs
                      </span>
                    )}
                  </div>
                  <p className="text-[9px] text-muted-foreground">{format(new Date(entity.created_at), 'MMM d, yyyy')}</p>
                </div>
              )
            })}
          </div>
        )}
      </ScrollArea>

      {/* Create Dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent className="sm:max-w-[440px]">
          <DialogHeader>
            <DialogTitle className="text-base font-black flex items-center gap-2">
              <Globe className="w-4 h-4 text-emerald-600" /> New Knowledge Entity
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-1.5">
              <Label className="text-xs font-bold uppercase tracking-wider">Title</Label>
              <Input
                placeholder="Entity title..."
                value={newTitle}
                onChange={e => setNewTitle(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && handleCreate()}
              />
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs font-bold uppercase tracking-wider">Kind</Label>
              <div className="flex flex-wrap gap-1.5">
                {KIND_OPTIONS.map(k => {
                  const cfg = getKindConfig(k)
                  return (
                    <Button
                      key={k} variant={newKind === k ? "secondary" : "outline"}
                      size="sm" className={cn("h-7 text-xs gap-1 font-bold", newKind === k && "ring-1 ring-emerald-400")}
                      onClick={() => setNewKind(k)}
                    >
                      <cfg.icon className={cn("w-3 h-3", cfg.color)} />{cfg.label}
                    </Button>
                  )
                })}
              </div>
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs font-bold uppercase tracking-wider">Summary</Label>
              <Textarea
                placeholder="Brief description..."
                className="resize-none h-20 text-sm"
                value={newSummary}
                onChange={e => setNewSummary(e.target.value)}
              />
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs font-bold uppercase tracking-wider">Tags <span className="font-normal text-muted-foreground normal-case">(comma-separated)</span></Label>
              <Input placeholder="tag1, tag2, tag3" value={newTags} onChange={e => setNewTags(e.target.value)} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" size="sm" onClick={() => setShowCreate(false)}>Cancel</Button>
            <Button size="sm" className="font-bold gap-1.5" disabled={!newTitle.trim() || isCreating} onClick={handleCreate}>
              {isCreating ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Plus className="w-3.5 h-3.5" />}
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
