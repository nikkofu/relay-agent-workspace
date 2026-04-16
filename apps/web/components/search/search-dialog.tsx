"use client"

import { useEffect } from "react"
import { useUIStore } from "@/stores/ui-store"
import { CommandDialog, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Sparkles, Hash, User, Clock } from "lucide-react"

export function SearchDialog() {
  const { isSearchOpen, toggleSearch, closeSearch } = useUIStore()

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

  return (
    <CommandDialog open={isSearchOpen} onOpenChange={closeSearch}>
      <div className="flex items-center border-b px-3 bg-purple-50 dark:bg-purple-900/10">
        <Sparkles className="mr-2 h-4 w-4 shrink-0 text-purple-500 opacity-50" />
        <CommandInput placeholder="Ask AI to search across Acme Corp..." className="border-none focus:ring-0" />
      </div>
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        
        <CommandGroup heading="Recent">
          <CommandItem className="cursor-pointer">
            <Clock className="mr-2 h-4 w-4 opacity-50" />
            <span>AI-native app development</span>
          </CommandItem>
          <CommandItem className="cursor-pointer">
            <Clock className="mr-2 h-4 w-4 opacity-50" />
            <span>#engineering-team</span>
          </CommandItem>
        </CommandGroup>

        <CommandGroup heading="AI Semantic Suggestions">
          <CommandItem className="cursor-pointer">
            <Sparkles className="mr-2 h-4 w-4 text-purple-500" />
            <span>Find messages related to &quot;project timeline&quot;</span>
          </CommandItem>
          <CommandItem className="cursor-pointer">
            <Sparkles className="mr-2 h-4 w-4 text-purple-500" />
            <span>Show me all files about &quot;UI design&quot;</span>
          </CommandItem>
        </CommandGroup>

        <CommandGroup heading="Channels & People">
          <CommandItem className="cursor-pointer">
            <Hash className="mr-2 h-4 w-4 opacity-50" />
            <span>#engineering</span>
          </CommandItem>
          <CommandItem className="cursor-pointer">
            <Hash className="mr-2 h-4 w-4 opacity-50" />
            <span>#ai-lab</span>
          </CommandItem>
          <CommandItem className="cursor-pointer">
            <User className="mr-2 h-4 w-4 opacity-50" />
            <span>Nikko Fu (you)</span>
          </CommandItem>
          <CommandItem className="cursor-pointer">
            <User className="mr-2 h-4 w-4 opacity-50" />
            <span>AI Assistant</span>
          </CommandItem>
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  )
}
