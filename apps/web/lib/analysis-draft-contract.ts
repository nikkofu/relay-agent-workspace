/**
 * Phase 70C: Draft-First Workflow and Channel Message Contract
 *
 * Frozen by Codex. Both target families use an immutable draft-ID chain:
 *   analysis snapshot -> typed target -> generate-draft -> confirm -> execute
 *
 * Canvas AI Dock owns the full execution UX for first release.
 * /ask and DM consume the contract but stay light.
 */

// ── Workflow Draft ────────────────────────────────────────────────────────────

export interface WorkflowDraftStep {
  title: string
}

export interface AnalysisWorkflowDraft {
  draft_id: string
  title: string
  goal: string
  steps: WorkflowDraftStep[]
}

export interface GenerateWorkflowDraftRequest {
  artifact_id: string
  channel_id: string
  analysis_snapshot_id: string
  /** Omit to use the analysis-level default_execution_target. */
  step_index?: number
}

export interface GenerateWorkflowDraftResponse {
  draft: AnalysisWorkflowDraft
}

export interface ConfirmCreateWorkflowRequest {
  draft_id: string
}

export interface ConfirmCreateWorkflowResponse {
  workflow_id: string
}

// ── Message Draft ─────────────────────────────────────────────────────────────

export interface AnalysisMessageDraft {
  draft_id: string
  channel_id: string
  body: string
}

export interface GenerateMessageDraftRequest {
  artifact_id: string
  channel_id: string
  analysis_snapshot_id: string
  /** Omit to use the analysis-level default_execution_target. */
  step_index?: number
}

export interface GenerateMessageDraftResponse {
  draft: AnalysisMessageDraft
}

export interface ConfirmPublishMessageRequest {
  draft_id: string
}

export interface ConfirmPublishMessageResponse {
  message_id: string
}

// ── Type guards ───────────────────────────────────────────────────────────────

export function isGenerateWorkflowDraftResponse(data: any): data is GenerateWorkflowDraftResponse {
  return (
    data?.draft &&
    typeof data.draft.draft_id === "string" &&
    typeof data.draft.title === "string" &&
    Array.isArray(data.draft.steps)
  )
}

export function isGenerateMessageDraftResponse(data: any): data is GenerateMessageDraftResponse {
  return (
    data?.draft &&
    typeof data.draft.draft_id === "string" &&
    typeof data.draft.body === "string"
  )
}
