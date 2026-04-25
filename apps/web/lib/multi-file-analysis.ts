/**
 * Phase 69: Multi-File Canvas AI Analysis Contract
 * 
 * Shared Web-side types for the analyze request and response.
 */

export interface MultiFileAnalysisRequest {
  artifact_id: string
  file_refs: { file_id: string }[]
  mode?: "multi_file_analysis"
}

export interface MultiFileAnalysisResponse {
  analysis: {
    summary: string
    observations: string[]
    next_steps: {
      text: string
      rationale: string
      action_hint: "summarize" | "compare" | "decide" | "share" | "plan" | "investigate" | "custom"
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
