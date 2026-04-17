"use client"

import { useEffect } from "react"
import { useUIStore } from "@/stores/ui-store"
import { useSearchStore } from "@/stores/search-store"
import { CommandDialog, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Sparkles, Hash, User, MessageSquare, Search as SearchIcon } from "lucide-react"

export function SearchDialog() {
  const { isSearchOpen, toggleSearch, closeSearch } = useUIStore()
  const { results, isSearching, search, clearResults } = useSearchStore()

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
    if (!isSearchOpen) clearResults()
  }, [isSearchOpen, clearResults])

  return (
    <CommandDialog open={isSearchOpen} onOpenChange={closeSearch}>
      <div className="flex items-center border-b px-3 bg-purple-50 dark:bg-purple-900/10">
        {isSearching ? (
          <Sparkles className="mr-2 h-4 w-4 shrink-0 text-purple-500 animate-pulse" />
        ) : (
          <SearchIcon className="mr-2 h-4 w-4 shrink-0 text-muted-foreground opacity-50" />
        )}
        <CommandInput 
          placeholder="Search channels, people, and messages..." 
          className="border-none focus:ring-0" 
          onValueChange={search}
        />
      </div>
      <CommandList>
        <CommandEmpty>{isSearching ? "Searching..." : "No results found."}</CommandEmpty>
        
        {results.channels?.length > 0 && (
          <CommandGroup heading="Channels">
            {results.channels.map(c => (
              <CommandItem key={c.id} className="cursor-pointer">
                <Hash className="mr-2 h-4 w-4 opacity-50" />
                <span>{c.name}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {results.users?.length > 0 && (
          <CommandGroup heading="People">
            {results.users.map(u => (
              <CommandItem key={u.id} className="cursor-pointer">
                <User className="mr-2 h-4 w-4 opacity-50" />
                <span>{u.name}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {results.messages?.length > 0 && (
          <CommandGroup heading="Messages">
            {results.messages.map(m => (
              <CommandItem key={m.id} className="cursor-pointer">
                <MessageSquare className="mr-2 h-4 w-4 opacity-50" />
                <span className="truncate" dangerouslySetInnerHTML={{ __html: m.content }} />
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {results.dms?.length > 0 && (
          <CommandGroup heading="DMs">
            {results.dms.map(d => (
              <CommandItem key={d.id} className="cursor-pointer">
                <User className="mr-2 h-4 w-4 opacity-50" />
                <span>{d.other_user?.name || "Direct Message"}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}
      </CommandList>
    </CommandDialog>
  )
}
