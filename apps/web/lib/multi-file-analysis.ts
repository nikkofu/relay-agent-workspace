/**
 * Phase 69 / 70B: Multi-File Canvas AI Analysis Contract
 *
 * Shared Web-side types for the analyze request and response.
 * Phase 70B adds typed execution targets: analysis-level default and per-step
 * override. Normalization is handled by lib/execution-target.ts.
 */

import type { ExecutionTarget } from "./execution-target"

export interface MultiFileAnalysisRequest {
  artifact_id: string
  file_refs: { file_id: string }[]
  mode?: "multi_file_analysis"
}

export interface MultiFileAnalysisResponse {
  analysis: {
    summary: string
    /** Phase 70B: analysis-level default execution target. */
    default_execution_target?: ExecutionTarget
    observations: string[]
    next_steps: {
      text: string
      rationale: string
      action_hint: "summarize" | "compare" | "decide" | "share" | "plan" | "investigate" | "custom"
      /** Phase 70B: step-level override. Wins over analysis default when present. */
      execution_target?: ExecutionTarget
    }[]
  }
}

/**
 * Type guard for the analysis response.
 */
export function isMultiFileAnalysisResponse(data: any): data is MultiFileAnalysisResponse {
  return (
    data &&
    data.analysis &&
    typeof data.analysis.summary === "string" &&
    Array.isArray(data.analysis.observations) &&
    Array.isArray(data.analysis.next_steps)
  )
}
