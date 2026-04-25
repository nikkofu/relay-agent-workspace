/**
 * Phase 70A: Create List From Analysis Snapshot Contract
 * 
 * Defines shared types for generating list drafts from analysis snapshots
 * and confirming creation from draft IDs.
 */

export interface GenerateListDraftRequest {
  artifact_id: string
  channel_id: string
  analysis_snapshot_id: string
}

export interface ListDraftItem {
  title: string
}

export interface AnalysisListDraft {
  draft_id: string
  channel_id: string
  title: string
  items: ListDraftItem[]
}

export interface GenerateListDraftResponse {
  draft: AnalysisListDraft
}

export interface ConfirmCreateListRequest {
  draft_id: string
}

export interface ConfirmCreateListResponse {
  list_id: string
}

/**
 * Type guard for the list draft response.
 */
export function isGenerateListDraftResponse(data: any): data is GenerateListDraftResponse {
  return (
    data &&
    data.draft &&
    typeof data.draft.draft_id === "string" &&
    typeof data.draft.title === "string" &&
    Array.isArray(data.draft.items)
  )
}

/**
 * Type guard for the confirm create response.
 */
export function isConfirmCreateListResponse(data: any): data is ConfirmCreateListResponse {
  return data && typeof data.list_id === "string"
}
