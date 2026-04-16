"use client"

import { Brain, ChevronDown } from "lucide-react"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"

export function AIReasoning({ content }: { content: string }) {
  return (
    <Collapsible className="my-2" defaultOpen>
      <CollapsibleTrigger className="flex items-center gap-2 text-[11px] font-medium text-muted-foreground hover:text-foreground transition-colors group">
        <Brain className="w-3.5 h-3.5 text-purple-400" />
        <span>Thought process</span>
        <ChevronDown className="w-3 h-3 group-data-[state=open]:rotate-180 transition-transform" />
      </CollapsibleTrigger>
      <CollapsibleContent className="mt-2 pl-5 border-l-2 border-muted text-xs text-muted-foreground italic leading-relaxed">
        {content}
      </CollapsibleContent>
    </Collapsible>
  )
}
