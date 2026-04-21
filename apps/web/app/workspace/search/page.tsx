"use client"

import { useState } from "react"
import { useCitationStore } from "@/stores/citation-store"
import { CitationCard } from "@/components/citation/citation-card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Badge } from "@/components/ui/badge"
import { Quote, Search, Loader2, FileText, MessageSquare, MessageCircle, Layout, X } from "lucide-react"
import { cn } from "@/lib/utils"
import type { EvidenceKind } from "@/types"

const KIND_FILTERS: { value: EvidenceKind | 'all'; label: string; icon: React.ElementType }[] = [
  { value: 'all', label: 'All', icon: Quote },
  { value: 'file_chunk', label: 'Files', icon: FileText },
  { value: 'message', label: 'Messages', icon: MessageSquare },
  { value: 'thread', label: 'Threads', icon: MessageCircle },
  { value: 'artifact_section', label: 'Artifacts', icon: Layout },
]

export default function SearchPage() {
  const { results, isSearching, lastQuery, filterKind, lookupCitations, setFilterKind, clearResults } = useCitationStore()
  const [inputValue, setInputValue] = useState("")

  const handleInput = (value: string) => {
    setInputValue(value)
    if (!value.trim()) { clearResults(); return }
    lookupCitations(value)
  }

  const filtered = filterKind === 'all' ? results : results.filter(r => r.evidence_kind === filterKind)

  const kindCounts = results.reduce<Record<string, number>>((acc, r) => {
    acc[r.evidence_kind] = (acc[r.evidence_kind] || 0) + 1
    return acc
  }, {})

  return (
    <div className="flex-1 flex flex-col bg-white dark:bg-[#1a1d21] h-full overflow-hidden">
      <header className="h-14 px-6 flex items-center border-b shrink-0 gap-3">
        <Quote className="w-5 h-5 text-violet-600 shrink-0" />
        <h1 className="text-lg font-black tracking-tight uppercase">Citation Search</h1>
        <Badge variant="secondary" className="text-[9px] font-black uppercase tracking-wider h-5 px-2">
          Knowledge Evidence
        </Badge>
      </header>

      {/* Search Bar */}
      <div className="px-6 py-4 border-b bg-violet-500/5 space-y-3">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-violet-500" />
          <Input
            placeholder="Search across files, messages, threads, and artifacts..."
            className="pl-9 border-violet-300 dark:border-violet-700 focus-visible:ring-violet-500"
            value={inputValue}
            onChange={e => handleInput(e.target.value)}
          />
          {isSearching && (
            <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 animate-spin text-violet-500" />
          )}
          {inputValue && !isSearching && (
            <button
              className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              onClick={() => { setInputValue(""); clearResults() }}
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>

        {/* Kind filter pills */}
        {results.length > 0 && (
          <div className="flex items-center gap-2 flex-wrap">
            {KIND_FILTERS.map(f => {
              const count = f.value === 'all' ? results.length : (kindCounts[f.value] || 0)
              if (f.value !== 'all' && count === 0) return null
              return (
                <Button
                  key={f.value}
                  variant={filterKind === f.value ? "secondary" : "ghost"}
                  size="sm"
                  className={cn("h-7 text-xs gap-1.5 font-bold", filterKind === f.value && "ring-1 ring-violet-400")}
                  onClick={() => setFilterKind(f.value)}
                >
                  <f.icon className="w-3 h-3" />
                  {f.label}
                  <span className="text-[10px] opacity-60">({count})</span>
                </Button>
              )
            })}
          </div>
        )}
      </div>

      <ScrollArea className="flex-1">
        <div className="p-6 space-y-3">
          {/* Empty states */}
          {!lastQuery && (
            <div className="py-24 flex flex-col items-center gap-4 text-center">
              <div className="w-16 h-16 rounded-2xl bg-violet-500/10 flex items-center justify-center">
                <Quote className="w-8 h-8 text-violet-400" />
              </div>
              <div className="space-y-1 max-w-sm">
                <p className="font-black uppercase text-xs tracking-widest text-muted-foreground">Knowledge Evidence Search</p>
                <p className="text-[11px] text-muted-foreground italic leading-relaxed">
                  Search across extracted file content, channel messages, thread replies, and artifact sections.
                  Results include snippets, locators, and entity references.
                </p>
              </div>
            </div>
          )}

          {lastQuery && !isSearching && filtered.length === 0 && (
            <div className="py-16 text-center space-y-2">
              <Quote className="w-10 h-10 text-muted-foreground/20 mx-auto" />
              <p className="text-xs text-muted-foreground italic">
                No {filterKind === 'all' ? '' : filterKind.replace('_', ' ')} evidence found for &ldquo;{lastQuery}&rdquo;
              </p>
            </div>
          )}

          {/* Results */}
          {filtered.length > 0 && (
            <>
              <p className="text-[10px] text-muted-foreground font-black uppercase tracking-widest pb-1">
                {filtered.length} result{filtered.length !== 1 ? 's' : ''} for &ldquo;{lastQuery}&rdquo;
              </p>
              {filtered.map(citation => (
                <CitationCard key={citation.id} citation={citation} />
              ))}
            </>
          )}
        </div>
      </ScrollArea>
    </div>
  )
}
