"use client"

import { MultiFileAnalysisResponse } from "@/lib/multi-file-analysis"
import {
  normalizeExecutionTarget,
  resolveExecutionTarget,
  TARGET_LABELS,
  TARGET_STYLES,
  isDraftFirstTarget,
  type ExecutionTarget,
} from "@/lib/execution-target"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { Check, ClipboardList, Lightbulb, Zap, ListPlus, ArrowRight, GitBranch, MessageSquare } from "lucide-react"

interface FileGroupAnalysisResultProps {
  result: MultiFileAnalysisResponse["analysis"]
  onInsertSummary: (text: string) => void
  onInsertObservations: (observations: string[]) => void
  onInsertPlan: (steps: MultiFileAnalysisResponse["analysis"]["next_steps"]) => void
  onCreateList?: () => void
  /** Phase 70C: start workflow draft for the analysis default or a specific step. */
  onStartWorkflow?: (stepIndex?: number) => void
  /** Phase 70C: post to channel draft for the analysis default or a specific step. */
  onPostToChannel?: (stepIndex?: number) => void
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
  onStartWorkflow,
  onPostToChannel,
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
                {/* Phase 70C: per-step draft action for workflow/channel_message */}
                {resolved && isDraftFirstTarget(resolved.type) && (
                  <div className="pl-4 mt-1.5">
                    {resolved.type === "workflow" && onStartWorkflow && (
                      <button
                        type="button"
                        onClick={() => onStartWorkflow(i)}
                        className="inline-flex items-center gap-1 text-[9px] font-black uppercase tracking-widest px-2 py-1 rounded bg-amber-500/10 text-amber-700 dark:text-amber-300 hover:bg-amber-500/20 transition-colors"
                      >
                        <GitBranch className="w-2.5 h-2.5" />
                        Start Workflow
                      </button>
                    )}
                    {resolved.type === "channel_message" && onPostToChannel && (
                      <button
                        type="button"
                        onClick={() => onPostToChannel(i)}
                        className="inline-flex items-center gap-1 text-[9px] font-black uppercase tracking-widest px-2 py-1 rounded bg-sky-500/10 text-sky-700 dark:text-sky-300 hover:bg-sky-500/20 transition-colors"
                      >
                        <MessageSquare className="w-2.5 h-2.5" />
                        Post to Channel
                      </button>
                    )}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </section>

      {/* Phase 70A/70B: Create list action */}
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
          {analysisDefault?.type === "list" && (
            <span className="ml-1 text-[9px] font-medium normal-case tracking-normal opacity-80">
              · suggested
            </span>
          )}
        </Button>
      )}

      {/* Phase 70C: analysis-level workflow draft CTA */}
      {analysisDefault?.type === "workflow" && onStartWorkflow && (
        <Button
          className="w-full h-9 gap-2 text-xs font-black uppercase tracking-wider bg-amber-600 hover:bg-amber-700 text-white shadow-md hover:scale-[1.01] active:scale-95 transition-all"
          onClick={() => onStartWorkflow(undefined)}
        >
          <GitBranch className="w-4 h-4" />
          Start Workflow
          <span className="ml-1 text-[9px] font-medium normal-case tracking-normal opacity-80">
            · suggested
          </span>
        </Button>
      )}

      {/* Phase 70C: analysis-level channel_message draft CTA */}
      {analysisDefault?.type === "channel_message" && onPostToChannel && (
        <Button
          className="w-full h-9 gap-2 text-xs font-black uppercase tracking-wider bg-sky-600 hover:bg-sky-700 text-white shadow-md hover:scale-[1.01] active:scale-95 transition-all"
          onClick={() => onPostToChannel(undefined)}
        >
          <MessageSquare className="w-4 h-4" />
          Post to Channel
          <span className="ml-1 text-[9px] font-medium normal-case tracking-normal opacity-80">
            · suggested
          </span>
        </Button>
      )}
    </div>
  )
}
