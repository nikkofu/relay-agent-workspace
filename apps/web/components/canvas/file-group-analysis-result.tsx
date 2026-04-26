"use client"

import { MultiFileAnalysisResponse } from "@/lib/multi-file-analysis"
import {
  normalizeExecutionTarget,
  resolveExecutionTarget,
  TARGET_LABELS,
  TARGET_STYLES,
  isExecutableTarget,
  type ExecutionTarget,
} from "@/lib/execution-target"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { Check, ClipboardList, Lightbulb, Zap, ListPlus, ArrowRight } from "lucide-react"

interface FileGroupAnalysisResultProps {
  result: MultiFileAnalysisResponse["analysis"]
  onInsertSummary: (text: string) => void
  onInsertObservations: (observations: string[]) => void
  onInsertPlan: (steps: MultiFileAnalysisResponse["analysis"]["next_steps"]) => void
  onCreateList?: () => void
}

function ExecutionTargetBadge({
  target,
  inherited,
}: {
  target: ExecutionTarget
  inherited?: boolean
}) {
  const style = TARGET_STYLES[target.type]
  return (
    <span
      className={cn(
        "inline-flex items-center gap-0.5 text-[9px] font-black uppercase tracking-widest px-1.5 py-0.5 rounded ring-1",
        style.bg,
        style.text,
        style.ring,
        inherited && "opacity-60",
      )}
      title={inherited ? `Inherited from analysis default` : undefined}
    >
      {inherited && <ArrowRight className="w-2 h-2 opacity-60" />}
      {TARGET_LABELS[target.type]}
    </span>
  )
}

export function FileGroupAnalysisResult({
  result,
  onInsertSummary,
  onInsertObservations,
  onInsertPlan,
  onCreateList,
}: FileGroupAnalysisResultProps) {
  const analysisDefault = normalizeExecutionTarget(result.default_execution_target)

  return (
    <div className="space-y-4 animate-in fade-in slide-in-from-bottom-2 duration-300">
      {/* Summary Section */}
      <section className="bg-white dark:bg-[#1a1d21] border rounded-lg overflow-hidden shadow-sm">
        <div className="px-3 py-2 border-b bg-muted/30 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-widest text-muted-foreground">
              <ClipboardList className="w-3 h-3" />
              Summary
            </div>
            {analysisDefault && (
              <ExecutionTargetBadge target={analysisDefault} />
            )}
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
            onClick={() => onInsertObservations(result.observations)}
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
            onClick={() => onInsertPlan(result.next_steps)}
          >
            <Check className="w-2.5 h-2.5" />
            Insert
          </Button>
        </div>
        <div className="p-3 space-y-3">
          {result.next_steps.map((step, i) => {
            const stepTarget = normalizeExecutionTarget(step.execution_target)
            const resolved = resolveExecutionTarget(analysisDefault, stepTarget)
            const inherited = resolved !== null && stepTarget === null
            return (
              <div key={i} className="text-xs leading-relaxed">
                <div className="font-bold flex items-start gap-1.5 flex-wrap">
                  <span className="text-[10px] text-muted-foreground font-mono shrink-0 mt-0.5">{i + 1}.</span>
                  <span className="flex-1">{step.text}</span>
                  {resolved && (
                    <ExecutionTargetBadge target={resolved} inherited={inherited} />
                  )}
                </div>
                <div className="pl-4 mt-0.5 text-muted-foreground italic text-[11px]">
                  {step.rationale}
                </div>
              </div>
            )
          })}
        </div>
      </section>

      {/* Phase 70A/70B: Create list action — visible when analysis default or any step targets "list" */}
      {onCreateList && (
        <Button
          className={cn(
            "w-full h-9 gap-2 text-xs font-black uppercase tracking-wider shadow-md hover:scale-[1.01] active:scale-95 transition-all",
            analysisDefault?.type === "list"
              ? "bg-violet-600 hover:bg-violet-700 text-white"
              : ""
          )}
          onClick={onCreateList}
        >
          <ListPlus className="w-4 h-4" />
          Create list from plan
          {analysisDefault && isExecutableTarget(analysisDefault.type) && (
            <span className="ml-1 text-[9px] font-medium normal-case tracking-normal opacity-80">
              · suggested
            </span>
          )}
        </Button>
      )}
    </div>
  )
}
