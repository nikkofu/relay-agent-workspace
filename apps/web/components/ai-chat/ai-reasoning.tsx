"use client"

import { Brain, ChevronDown } from "lucide-react"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"

export function AIReasoning({ content }: { content: string }) {
  return (
    <Collapsible className="my-1.5 overflow-hidden rounded-lg border border-purple-500/10 bg-purple-500/[0.02]">
      <CollapsibleTrigger className="flex items-center gap-2 px-3 py-1.5 text-[10px] font-bold text-purple-600/70 dark:text-purple-400/70 hover:text-purple-600 dark:hover:text-purple-400 transition-colors group w-full">
        <Brain className="w-3 h-3" />
        <span className="uppercase tracking-wider">Reasoning Process</span>
        <ChevronDown className="w-3 h-3 ml-auto group-data-[state=open]:rotate-180 transition-transform" />
      </CollapsibleTrigger>
      <CollapsibleContent className="px-3 pb-3 pt-1 border-t border-purple-500/5 text-xs text-muted-foreground/80 italic leading-relaxed font-normal bg-white/40 dark:bg-black/20">
        {content}
      </CollapsibleContent>
    </Collapsible>
  )
}
