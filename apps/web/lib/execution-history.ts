import { API_BASE_URL } from "@/lib/constants"
import type { ExecutionTargetType } from "@/lib/execution-target"

export type ExecutionHistoryEventType = "draft_generated" | "confirmed" | "created" | "published" | "failed"
export type ExecutionHistoryStatus = "success" | "failed"

export interface ExecutionHistoryEvent {
  id: string
  event_type: ExecutionHistoryEventType
  status: ExecutionHistoryStatus
  actor_user_id?: string
  analysis_snapshot_id: string
  next_step_id?: string
  step_index?: number
  execution_target_type?: ExecutionTargetType
  draft_id?: string
  draft_type?: string
  created_object_id?: string
  created_object_type?: string
  failure_stage?: string
  error_message?: string
  created_at: string
}

export interface AnalysisExecutionProjectionItem {
  event: ExecutionHistoryEvent
  status: ExecutionHistoryEventType
  label: string
  stepIndex?: number
  executionTargetType?: ExecutionTargetType
  createdObjectId?: string
  createdObjectType?: string
  createdObjectHref?: string | null
  failureStage?: string
  errorMessage?: string
  occurredAt: string
}

export interface AnalysisExecutionProjection {
  analysis: AnalysisExecutionProjectionItem | null
  steps: Record<number, AnalysisExecutionProjectionItem>
  events: ExecutionHistoryEvent[]
}

const KNOWN_EVENT_TYPES = new Set<ExecutionHistoryEventType>([
  "draft_generated",
  "confirmed",
  "created",
  "published",
  "failed",
])

const KNOWN_STATUSES = new Set<ExecutionHistoryStatus>(["success", "failed"])

const TARGET_NOUNS: Record<ExecutionTargetType, string> = {
  list: "list",
  workflow: "workflow",
  channel_message: "message",
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null
}

function normalizeEventType(value: unknown): ExecutionHistoryEventType | null {
  return typeof value === "string" && KNOWN_EVENT_TYPES.has(value as ExecutionHistoryEventType)
    ? value as ExecutionHistoryEventType
    : null
}

function normalizeStatus(value: unknown): ExecutionHistoryStatus | null {
  return typeof value === "string" && KNOWN_STATUSES.has(value as ExecutionHistoryStatus)
    ? value as ExecutionHistoryStatus
    : null
}

function normalizeExecutionTargetType(value: unknown): ExecutionTargetType | undefined {
  return value === "list" || value === "workflow" || value === "channel_message"
    ? value
    : undefined
}

function normalizeOptionalString(value: unknown): string | undefined {
  return typeof value === "string" && value.trim() ? value : undefined
}

function normalizeOptionalNumber(value: unknown): number | undefined {
  return typeof value === "number" && Number.isFinite(value) ? value : undefined
}

export function normalizeExecutionHistoryEvent(raw: unknown): ExecutionHistoryEvent | null {
  if (!isRecord(raw)) return null
  const id = normalizeOptionalString(raw.id)
  const eventType = normalizeEventType(raw.event_type)
  const status = normalizeStatus(raw.status)
  const snapshotId = normalizeOptionalString(raw.analysis_snapshot_id)
  const createdAt = normalizeOptionalString(raw.created_at)
  if (!id || !eventType || !status || !snapshotId || !createdAt) return null
  return {
    id,
    event_type: eventType,
    status,
    actor_user_id: normalizeOptionalString(raw.actor_user_id),
    analysis_snapshot_id: snapshotId,
    next_step_id: normalizeOptionalString(raw.next_step_id),
    step_index: normalizeOptionalNumber(raw.step_index),
    execution_target_type: normalizeExecutionTargetType(raw.execution_target_type),
    draft_id: normalizeOptionalString(raw.draft_id),
    draft_type: normalizeOptionalString(raw.draft_type),
    created_object_id: normalizeOptionalString(raw.created_object_id),
    created_object_type: normalizeOptionalString(raw.created_object_type),
    failure_stage: normalizeOptionalString(raw.failure_stage),
    error_message: normalizeOptionalString(raw.error_message),
    created_at: createdAt,
  }
}

export function isAnalysisExecutionHistoryResponse(data: unknown): data is { events: ExecutionHistoryEvent[] } {
  return isRecord(data) && Array.isArray(data.events)
}

export async function fetchAnalysisExecutionHistory(snapshotId: string): Promise<ExecutionHistoryEvent[]> {
  const params = new URLSearchParams({ analysis_snapshot_id: snapshotId })
  const res = await fetch(`${API_BASE_URL}/ai/canvas/analysis-execution-history?${params.toString()}`)
  if (!res.ok) {
    throw new Error("Failed to load analysis execution history")
  }
  const data = await res.json()
  if (!isAnalysisExecutionHistoryResponse(data)) return []
  return data.events
    .map(normalizeExecutionHistoryEvent)
    .filter((event): event is ExecutionHistoryEvent => event !== null)
}

export function getExecutionHistoryEventLabel(eventType: ExecutionHistoryEventType): string {
  switch (eventType) {
    case "draft_generated":
      return "Draft ready"
    case "confirmed":
      return "Confirmed"
    case "created":
      return "Created"
    case "published":
      return "Published"
    case "failed":
      return "Failed"
  }
}

export function getCreatedObjectHref(event: Pick<ExecutionHistoryEvent, "created_object_id" | "created_object_type" | "execution_target_type">): string | null {
  const objectId = event.created_object_id
  const objectType = event.created_object_type ?? event.execution_target_type
  if (!objectId || !objectType) return null
  if (objectType === "list") return `/workspace/lists?id=${objectId}`
  if (objectType === "workflow") return "/workspace/workflows"
  return null
}

function getResolvedStepIndex(
  event: ExecutionHistoryEvent,
  draftStepIndexById: Map<string, number>,
): number | undefined {
  if (typeof event.step_index === "number") return event.step_index
  if (event.draft_id && draftStepIndexById.has(event.draft_id)) {
    return draftStepIndexById.get(event.draft_id)
  }
  return undefined
}

function toProjectionItem(
  event: ExecutionHistoryEvent,
  stepIndex: number | undefined,
): AnalysisExecutionProjectionItem {
  return {
    event,
    status: event.event_type,
    label: getExecutionHistoryEventLabel(event.event_type),
    stepIndex,
    executionTargetType: event.execution_target_type,
    createdObjectId: event.created_object_id,
    createdObjectType: event.created_object_type,
    createdObjectHref: getCreatedObjectHref(event),
    failureStage: event.failure_stage,
    errorMessage: event.error_message,
    occurredAt: event.created_at,
  }
}

export function projectAnalysisExecutionHistory(events: ExecutionHistoryEvent[]): AnalysisExecutionProjection {
  const normalized = events
    .map(normalizeExecutionHistoryEvent)
    .filter((event): event is ExecutionHistoryEvent => event !== null)
    .sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())

  const draftStepIndexById = new Map<string, number>()
  const steps: Record<number, AnalysisExecutionProjectionItem> = {}
  let analysis: AnalysisExecutionProjectionItem | null = null

  for (const event of normalized) {
    if (event.draft_id && typeof event.step_index === "number") {
      draftStepIndexById.set(event.draft_id, event.step_index)
    }
    const stepIndex = getResolvedStepIndex(event, draftStepIndexById)
    const projection = toProjectionItem(event, stepIndex)
    if (typeof stepIndex === "number") {
      steps[stepIndex] = projection
    } else {
      analysis = projection
    }
  }

  return { analysis, steps, events: normalized }
}

export function buildExecutionHistorySummary(event: Pick<ExecutionHistoryEvent, "event_type" | "execution_target_type" | "created_object_id" | "error_message">): string {
  const noun = event.execution_target_type ? TARGET_NOUNS[event.execution_target_type] : "execution"
  if (event.event_type === "failed") {
    return event.error_message ? `AI ${noun} failed` : "AI execution failed"
  }
  if (event.event_type === "published") return `AI published ${noun}`
  if (event.event_type === "created") return `AI created ${noun}`
  if (event.event_type === "confirmed") return `AI confirmed ${noun}`
  return `AI drafted ${noun}`
}

export function buildExecutionHistoryBody(event: Pick<ExecutionHistoryEvent, "created_object_id" | "error_message" | "failure_stage" | "created_object_type" | "execution_target_type">): string | undefined {
  if (event.error_message) return event.error_message
  if (event.created_object_id) {
    const label = event.created_object_type ?? event.execution_target_type ?? "object"
    return `${label}: ${event.created_object_id}`
  }
  if (event.failure_stage) {
    return event.failure_stage.replace(/_/g, " ")
  }
  return undefined
}

export function splitHomeAIExecutions(events: ExecutionHistoryEvent[]) {
  const normalized = events
    .map(normalizeExecutionHistoryEvent)
    .filter((event): event is ExecutionHistoryEvent => event !== null)
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())

  return {
    recent: normalized,
    failed: normalized.filter(event => event.event_type === "failed" || event.status === "failed"),
  }
}
