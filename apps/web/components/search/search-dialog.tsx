"use client"

import { useEffect, useState } from "react"
import { useUIStore } from "@/stores/ui-store"
import { useSearchStore } from "@/stores/search-store"
import { useChannelStore } from "@/stores/channel-store"
import { useRouter } from "next/navigation"
import { CommandDialog, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Sparkles, Hash, User, MessageSquare, Search as SearchIcon, ArrowRight, Zap, FileCode, FileText, Download } from "lucide-react"
import { Button } from "@/components/ui/button"

export function SearchDialog() {
  const { isSearchOpen, toggleSearch, closeSearch, openDockedChat, openCanvas } = useUIStore()
  const { results, isSearching, suggestions, fetchSuggestions, search, intelligentSearch, intelligentResults, isIntelligentSearching, clearResults } = useSearchStore()
  const { setCurrentChannelById } = useChannelStore()
  const [query, setQuery] = useState("")
  const router = useRouter()

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        toggleSearch()
      }
    }
    document.addEventListener("keydown", down)
    return () => document.removeEventListener("keydown", down)
  }, [toggleSearch])

  useEffect(() => {
    if (!isSearchOpen) {
      clearResults()
      setQuery("")
    }
  }, [isSearchOpen, clearResults])

  const handleValueChange = (v: string) => {
    setQuery(v)
    fetchSuggestions(v)
    search(v)
    intelligentSearch(v)
  }

  const navigateToChannel = (id: string) => {
    setCurrentChannelById(id)
    router.push(`/workspace?c=${id}`)
    closeSearch()
  }

  const navigateToUser = (userId: string) => {
    openDockedChat(userId)
    closeSearch()
  }

  const navigateToArtifact = (id: string) => {
    openCanvas(id)
    closeSearch()
  }

  return (
    <CommandDialog open={isSearchOpen} onOpenChange={closeSearch}>
      <div className="flex items-center border-b px-3 bg-muted/20">
        {isSearching ? (
          <Sparkles className="mr-2 h-4 w-4 shrink-0 text-purple-500 animate-pulse" />
        ) : (
          <SearchIcon className="mr-2 h-4 w-4 shrink-0 text-muted-foreground opacity-50" />
        )}
        <CommandInput 
          placeholder="Search channels, people, and messages..." 
          className="border-none focus:ring-0 h-12" 
          value={query}
          onValueChange={handleValueChange}
        />
      </div>
      <CommandList className="max-h-[450px]">
        <CommandEmpty className="py-12 text-center">
          {isSearching || isIntelligentSearching ? (
            <div className="flex flex-col items-center gap-2">
              <LoaderIcon className="h-6 w-6 animate-spin text-muted-foreground" />
              <p className="text-sm text-muted-foreground font-medium">Searching for &quot;{query}&quot;...</p>
            </div>
          ) : (
            <div className="flex flex-col items-center gap-2">
              <SearchIcon className="h-8 w-8 text-muted-foreground opacity-20" />
              <p className="text-sm text-muted-foreground">No results found for &quot;{query}&quot;</p>
            </div>
          )}
        </CommandEmpty>

        {intelligentResults.length > 0 && query.length > 0 && (
          <CommandGroup heading="Top Results (AI-Ranked)">
            {intelligentResults.map((r) => (
              <CommandItem 
                key={`${r.type}-${r.id}`} 
                onSelect={() => {
                  if (r.type === 'channel') navigateToChannel(r.id)
                  else if (r.type === 'user') navigateToUser(r.id)
                  else if (r.type === 'artifact') navigateToArtifact(r.id)
                }}
                className="flex items-center gap-3 py-3 cursor-pointer group border-l-2 border-purple-500/50 ml-1 pl-3 bg-purple-500/5"
              >
                <div className="w-8 h-8 rounded-lg bg-purple-500/10 flex items-center justify-center shrink-0">
                  {r.type === 'channel' ? <Hash className="w-4 h-4 text-purple-600" /> : 
                   r.type === 'user' ? <User className="w-4 h-4 text-purple-600" /> :
                   r.type === 'artifact' ? <FileCode className="w-4 h-4 text-purple-600" /> :
                   <Sparkles className="w-4 h-4 text-purple-600" />}
                </div>
                <div className="flex flex-col min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-bold">{r.label}</span>
                    <span className="text-[8px] bg-purple-500/20 text-purple-700 dark:text-purple-300 px-1.5 py-0.5 rounded-full font-black uppercase tracking-tighter">Ranked {Math.round(r.score * 100)}%</span>
                  </div>
                  {r.reason && <span className="text-[10px] text-muted-foreground truncate italic">{r.reason}</span>}
                </div>
                <ArrowRight className="ml-auto w-4 h-4 opacity-0 group-aria-selected:opacity-100 transition-opacity text-purple-500" />
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {suggestions.length > 0 && query.length > 0 && (
          <CommandGroup heading="Suggestions">
            {suggestions.map((s) => (
              <CommandItem 
                key={`${s.type}-${s.id}`} 
                onSelect={() => {
                  if (s.type === 'channel') navigateToChannel(s.id)
                  else if (s.type === 'user') navigateToUser(s.id)
                  else if (s.type === 'artifact') navigateToArtifact(s.id)
                }}
                className="flex items-center gap-2 py-3 cursor-pointer group"
              >
                <div className="w-6 h-6 rounded bg-purple-500/10 flex items-center justify-center shrink-0">
                  <Zap className="w-3.5 h-3.5 text-purple-600" />
                </div>
                <div className="flex flex-col min-w-0">
                  <span className="text-sm font-bold">{s.label}</span>
                  {s.sublabel && <span className="text-[10px] text-muted-foreground truncate italic">{s.sublabel}</span>}
                </div>
                <ArrowRight className="ml-auto w-3.5 h-3.5 opacity-0 group-aria-selected:opacity-100 transition-opacity text-purple-500" />
              </CommandItem>
            ))}
          </CommandGroup>
        )}
        
        {results.channels?.length > 0 && (
          <CommandGroup heading="Channels">
            {results.channels.map(c => (
              <CommandItem 
                key={c.id} 
                onSelect={() => navigateToChannel(c.id)}
                className="flex flex-col items-start gap-1 py-3 cursor-pointer"
              >
                <div className="flex items-center gap-2 w-full">
                  <Hash className="h-4 w-4 opacity-50" />
                  <span className="text-sm font-bold">{c.name}</span>
                </div>
                {c.match_reason && (
                  <span className="text-[10px] text-muted-foreground ml-6 italic">{c.match_reason}</span>
                )}
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {results.users?.length > 0 && (
          <CommandGroup heading="People">
            {results.users.map(u => (
              <CommandItem 
                key={u.id} 
                onSelect={() => navigateToUser(u.id)}
                className="flex flex-col items-start gap-1 py-3 cursor-pointer"
              >
                <div className="flex items-center gap-2 w-full">
                  <User className="h-4 w-4 opacity-50" />
                  <span className="text-sm font-bold">{u.name}</span>
                </div>
                {u.match_reason && (
                  <span className="text-[10px] text-muted-foreground ml-6 italic">{u.match_reason}</span>
                )}
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {results.artifacts?.length > 0 && (
          <CommandGroup heading="Artifacts">
            {results.artifacts.map(a => (
              <CommandItem 
                key={a.id} 
                onSelect={() => navigateToArtifact(a.id)}
                className="flex flex-col items-start gap-1 py-3 cursor-pointer"
              >
                <div className="flex items-center gap-2 w-full">
                  <FileCode className="h-4 w-4 opacity-50" />
                  <span className="text-sm font-bold">{a.title}</span>
                </div>
                {a.match_reason && (
                  <span className="text-[10px] text-muted-foreground ml-6 italic">{a.match_reason}</span>
                )}
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {results.files?.length > 0 && (
          <CommandGroup heading="Files">
            {results.files.map(f => (
              <CommandItem 
                key={f.id} 
                className="flex items-start justify-between gap-1 py-3 cursor-pointer group"
              >
                <div className="flex flex-col items-start gap-1">
                  <div className="flex items-center gap-2 w-full">
                    <FileText className="h-4 w-4 opacity-50" />
                    <span className="text-sm font-bold">{f.name}</span>
                  </div>
                  {f.match_reason && (
                    <span className="text-[10px] text-muted-foreground ml-6 italic">{f.match_reason}</span>
                  )}
                </div>
                <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground opacity-0 group-aria-selected:opacity-100 transition-opacity" asChild>
                  <a href={f.url} target="_blank" rel="noopener noreferrer" download={f.name}>
                    <Download className="w-4 h-4" />
                  </a>
                </Button>
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {results.messages?.length > 0 && (
          <CommandGroup heading="Messages">
            {results.messages.map(m => (
              <CommandItem 
                key={m.id} 
                onSelect={() => navigateToChannel(m.channel_id)}
                className="flex flex-col items-start gap-1 py-3 cursor-pointer group"
              >
                <div className="flex items-center justify-between w-full">
                  <div className="flex items-center gap-2">
                    <MessageSquare className="h-4 w-4 opacity-50" />
                    <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground group-aria-selected:text-purple-500">In #{m.channel?.name || 'unknown'}</span>
                  </div>
                </div>
                <div className="text-xs ml-6 leading-relaxed line-clamp-2 text-foreground/80 font-medium italic border-l-2 border-muted pl-2 mt-1">
                  {m.snippet ? (
                    <span dangerouslySetInnerHTML={{ __html: m.snippet }} />
                  ) : (
                    <span dangerouslySetInnerHTML={{ __html: m.content }} />
                  )}
                </div>
              </CommandItem>
            ))}
          </CommandGroup>
        )}
      </CommandList>
    </CommandDialog>
  )
}

function LoaderIcon({ className }: { className?: string }) {
  return (
    <svg 
      xmlns="http://www.w3.org/2000/svg" 
      width="24" 
      height="24" 
      viewBox="0 0 24 24" 
      fill="none" 
      stroke="currentColor" 
      strokeWidth="2" 
      strokeLinecap="round" 
      strokeLinejoin="round" 
      className={className}
    >
      <path d="M12 2v4m0 12v4M4.93 4.93l2.83 2.83m8.48 8.48l2.83 2.83M2 12h4m12 0h4M4.93 19.07l2.83-2.83m8.48-8.48l2.83-2.83"/>
    </svg>
  )
}
