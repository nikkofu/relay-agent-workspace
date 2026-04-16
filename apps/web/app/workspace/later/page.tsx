"use client"

import { Bookmark } from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"

export default function LaterPage() {
  return (
    <div className="flex-1 flex flex-col h-full bg-white dark:bg-[#1a1d21]">
      <header className="h-14 px-4 flex items-center border-b shrink-0">
        <h2 className="font-bold text-lg">Later</h2>
      </header>

      <ScrollArea className="flex-1">
        <div className="flex flex-col items-center justify-center p-8 mt-20 text-center gap-4">
          <div className="w-16 h-16 rounded-full bg-muted flex items-center justify-center">
            <Bookmark className="w-8 h-8 opacity-20" />
          </div>
          <div className="max-w-xs">
            <h3 className="font-bold text-lg mb-1">Your saved items</h3>
            <p className="text-sm text-muted-foreground">
              Save messages and files to come back to them later. Everything you save will appear here.
            </p>
          </div>
        </div>
      </ScrollArea>
    </div>
  )
}
