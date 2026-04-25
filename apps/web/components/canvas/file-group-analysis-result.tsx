"use client"

import { MultiFileAnalysisResponse } from "@/lib/multi-file-analysis"
import { Button } from "@/components/ui/button"
import { Check, ClipboardList, Lightbulb, Zap } from "lucide-react"

interface FileGroupAnalysisResultProps {
  result: MultiFileAnalysisResponse["analysis"]
  onInsertSummary: (text: string) => void
  onInsertObservations: (text: string) => void
  onInsertPlan: (text: string) => void
}

export function FileGroupAnalysisResult({
  result,
  onInsertSummary,
  onInsertObservations,
  onInsertPlan,
}: FileGroupAnalysisResultProps) {
  const formatObservations = (obs: string[]) => {
    return "### Key Observations\n\n" + obs.map(o => `- ${o}`).join("\n")
  }

  const formatPlan = (steps: MultiFileAnalysisResponse["analysis"]["next_steps"]) => {
    return "### Proposed Plan\n\n" + steps.map((s, i) => (
      `${i + 1}. **${s.text}**\n   - *Rationale*: ${s.rationale}\n   - *Action*: ${s.action_hint}`
    )).join("\n\n")
  }

  return (
    <div className="space-y-4 animate-in fade-in slide-in-from-bottom-2 duration-300">
      {/* Summary Section */}
      <section className="bg-white dark:bg-[#1a1d21] border rounded-lg overflow-hidden shadow-sm">
        <div className="px-3 py-2 border-b bg-muted/30 flex items-center justify-between">
          <div className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-muted-foreground">
            <ClipboardList className="w-3 h-3" />
            Summary
          </div>
          <Button 
            variant="ghost" 
            size="sm" 
            className="h-6 text-[9px] font-bold uppercase gap-1"
            onClick={() => onInsertSummary(result.summary)}
          >
            <Check className="w-2.5 h-2.5" />
            Insert
          </Button>
        </div>
        <div className="p-3 text-xs leading-relaxed">
          {result.summary}
        </div>
      </section>

      {/* Observations Section */}
      <section className="bg-white dark:bg-[#1a1d21] border rounded-lg overflow-hidden shadow-sm">
        <div className="px-3 py-2 border-b bg-muted/30 flex items-center justify-between">
          <div className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-muted-foreground">
            <Lightbulb className="w-3 h-3" />
            Observations
          </div>
          <Button 
            variant="ghost" 
            size="sm" 
            className="h-6 text-[9px] font-bold uppercase gap-1"
            onClick={() => onInsertObservations(formatObservations(result.observations))}
          >
            <Check className="w-2.5 h-2.5" />
            Insert
          </Button>
        </div>
        <div className="p-3 space-y-2">
          {result.observations.map((obs, i) => (
            <div key={i} className="flex gap-2 text-xs leading-relaxed">
              <span className="text-muted-foreground mt-1 text-[10px]">•</span>
              <span>{obs}</span>
            </div>
          ))}
        </div>
      </section>

      {/* Plan Section */}
      <section className="bg-white dark:bg-[#1a1d21] border rounded-lg overflow-hidden shadow-sm">
        <div className="px-3 py-2 border-b bg-muted/30 flex items-center justify-between">
          <div className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-muted-foreground">
            <Zap className="w-3 h-3" />
            Proposed Plan
          </div>
          <Button 
            variant="ghost" 
            size="sm" 
            className="h-6 text-[9px] font-bold uppercase gap-1"
            onClick={() => onInsertPlan(formatPlan(result.next_steps))}
          >
            <Check className="w-2.5 h-2.5" />
            Insert
          </Button>
        </div>
        <div className="p-3 space-y-3">
          {result.next_steps.map((step, i) => (
            <div key={i} className="text-xs leading-relaxed">
              <div className="font-bold flex items-center gap-1.5">
                <span className="text-[10px] text-muted-foreground font-mono">{i + 1}.</span>
                {step.text}
              </div>
              <div className="pl-4 mt-0.5 text-muted-foreground italic text-[11px]">
                {step.rationale}
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}
