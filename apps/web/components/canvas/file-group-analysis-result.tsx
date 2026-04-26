"use client"

import Link from "next/link"
import { MultiFileAnalysisResponse } from "@/lib/multi-file-analysis"
import {
  normalizeExecutionTarget,
  resolveExecutionTarget,
  TARGET_LABELS,
  TARGET_STYLES,
  isDraftFirstTarget,
  type ExecutionTarget,
} from "@/lib/execution-target"
import {
  type AnalysisExecutionProjection,
  type AnalysisExecutionProjectionItem,
} from "@/lib/execution-history"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { Check, ClipboardList, Lightbulb, Zap, ListPlus, ArrowRight, GitBranch, MessageSquare, AlertCircle, ExternalLink } from "lucide-react"

interface FileGroupAnalysisResultProps {
  result: MultiFileAnalysisResponse["analysis"]
  execution?: AnalysisExecutionProjection
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

function ExecutionStatusBadge({ item }: { item: AnalysisExecutionProjectionItem }) {
  const style = item.status === "failed"
    ? "bg-rose-500/10 text-rose-700 dark:text-rose-300 ring-rose-500/30"
    : item.status === "published"
    ? "bg-sky-500/10 text-sky-700 dark:text-sky-300 ring-sky-500/30"
    : item.status === "created"
    ? "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 ring-emerald-500/30"
    : item.status === "confirmed"
    ? "bg-amber-500/10 text-amber-700 dark:text-amber-300 ring-amber-500/30"
    : "bg-violet-500/10 text-violet-700 dark:text-violet-300 ring-violet-500/30"

  return (
    <span className={cn(
      "inline-flex items-center gap-1 text-[9px] font-black uppercase tracking-widest px-1.5 py-0.5 rounded ring-1",
      style,
    )}>
      {item.status === "failed" && <AlertCircle className="w-2.5 h-2.5" />}
      {item.label}
    </span>
  )
}

function ExecutionMeta({ item }: { item: AnalysisExecutionProjectionItem }) {
  if (!item.createdObjectHref && !item.createdObjectId && !item.errorMessage && !item.failureStage) return null

  return (
    <div className="mt-1.5 space-y-1">
      {item.createdObjectHref ? (
        <Link
          href={item.createdObjectHref}
          className="inline-flex items-center gap-1 text-[10px] font-semibold text-sky-700 dark:text-sky-300 hover:underline"
        >
          {item.createdObjectType === "list"
            ? "Open created list"
            : item.createdObjectType === "workflow"
            ? "Open workflows"
            : "Open created object"}
          <ExternalLink className="w-2.5 h-2.5" />
        </Link>
      ) : item.createdObjectId ? (
        <div className="text-[10px] text-muted-foreground">
          Created {item.createdObjectType ?? "object"}: {item.createdObjectId}
        </div>
      ) : null}
      {item.errorMessage && (
        <div className="text-[10px] text-rose-700 dark:text-rose-300 leading-relaxed">
          {item.failureStage ? `${item.failureStage.replace(/_/g, " ")}: ` : ""}
          {item.errorMessage}
        </div>
      )}
    </div>
  )
}

export function FileGroupAnalysisResult({
  result,
  execution,
  onInsertSummary,
  onInsertObservations,
  onInsertPlan,
  onCreateList,
  onStartWorkflow,
  onPostToChannel,
}: FileGroupAnalysisResultProps) {
  const analysisDefault = normalizeExecutionTarget(result.default_execution_target)
  const analysisExecution = execution?.analysis ?? null

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
            {analysisExecution && (
              <ExecutionStatusBadge item={analysisExecution} />
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
          {analysisExecution && <ExecutionMeta item={analysisExecution} />}
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
            const stepExecution = execution?.steps[i]
            return (
              <div key={i} className="text-xs leading-relaxed">
                <div className="font-bold flex items-start gap-1.5 flex-wrap">
                  <span className="text-[10px] text-muted-foreground font-mono shrink-0 mt-0.5">{i + 1}.</span>
                  <span className="flex-1">{step.text}</span>
                  {resolved && (
                    <ExecutionTargetBadge target={resolved} inherited={inherited} />
                  )}
                  {stepExecution && (
                    <ExecutionStatusBadge item={stepExecution} />
                  )}
                </div>
                <div className="pl-4 mt-0.5 text-muted-foreground italic text-[11px]">
                  {step.rationale}
                </div>
                {stepExecution && (
                  <div className="pl-4">
                    <ExecutionMeta item={stepExecution} />
                  </div>
                )}
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
        <div className="space-y-2">
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
          {analysisExecution?.executionTargetType === "list" && (
            <div className="rounded-lg border border-violet-500/20 bg-violet-500/5 px-3 py-2">
              <div className="flex items-center gap-2 flex-wrap">
                <ExecutionStatusBadge item={analysisExecution} />
              </div>
              <ExecutionMeta item={analysisExecution} />
            </div>
          )}
        </div>
      )}

      {/* Phase 70C: analysis-level workflow draft CTA */}
      {analysisDefault?.type === "workflow" && onStartWorkflow && (
        <div className="space-y-2">
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
          {analysisExecution?.executionTargetType === "workflow" && (
            <div className="rounded-lg border border-amber-500/20 bg-amber-500/5 px-3 py-2">
              <div className="flex items-center gap-2 flex-wrap">
                <ExecutionStatusBadge item={analysisExecution} />
              </div>
              <ExecutionMeta item={analysisExecution} />
            </div>
          )}
        </div>
      )}

      {/* Phase 70C: analysis-level channel_message draft CTA */}
      {analysisDefault?.type === "channel_message" && onPostToChannel && (
        <div className="space-y-2">
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
          {analysisExecution?.executionTargetType === "channel_message" && (
            <div className="rounded-lg border border-sky-500/20 bg-sky-500/5 px-3 py-2">
              <div className="flex items-center gap-2 flex-wrap">
                <ExecutionStatusBadge item={analysisExecution} />
              </div>
              <ExecutionMeta item={analysisExecution} />
            </div>
          )}
        </div>
      )}
    </div>
  )
}
