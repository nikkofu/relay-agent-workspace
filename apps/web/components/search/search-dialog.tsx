"use client"

import { useEffect, useState } from "react"
import { useUIStore } from "@/stores/ui-store"
import { useSearchStore } from "@/stores/search-store"
import { useChannelStore } from "@/stores/channel-store"
import { useRouter } from "next/navigation"
import { CommandDialog, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Sparkles, Hash, User, MessageSquare, Search as SearchIcon, ArrowRight, Zap } from "lucide-react"

export function SearchDialog() {
  const { isSearchOpen, toggleSearch, closeSearch, openDockedChat } = useUIStore()
  const { results, isSearching, suggestions, fetchSuggestions, search, clearResults } = useSearchStore()
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
          {isSearching ? (
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

        {suggestions.length > 0 && query.length > 0 && (
          <CommandGroup heading="Suggestions">
            {suggestions.map((s) => (
              <CommandItem 
                key={`${s.type}-${s.id}`} 
                onSelect={() => {
                  if (s.type === 'channel') navigateToChannel(s.id)
                  else if (s.type === 'user') navigateToUser(s.id)
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
