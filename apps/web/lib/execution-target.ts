/**
 * Phase 70B: Shared Typed Execution Target Contract
 *
 * Frozen by Codex. All AI surfaces (Canvas, /ask, DM) normalize raw execution
 * target payloads through this module before any surface-specific rendering.
 *
 * Resolution rule (deterministic):
 *   step-level override > analysis default > null (no guessed action)
 *
 * Malformed or unknown target types normalize to null — they are never guessed.
 */

export type ExecutionTargetType = "list" | "workflow" | "channel_message"

const KNOWN_TARGET_TYPES = new Set<string>(["list", "workflow", "channel_message"])

export interface ExecutionTarget {
  type: ExecutionTargetType
  /** Extra payload fields — preserved in transport, ignored for first-release behavior. */
  [key: string]: unknown
}

/**
 * Normalize a raw execution-target value from any AI surface API response.
 * Returns null for malformed, absent, or unknown-typed payloads.
 */
export function normalizeExecutionTarget(raw: unknown): ExecutionTarget | null {
  if (!raw || typeof raw !== "object") return null
  const obj = raw as Record<string, unknown>
  if (typeof obj.type !== "string") return null
  if (!KNOWN_TARGET_TYPES.has(obj.type)) return null
  return { ...obj, type: obj.type as ExecutionTargetType }
}

/**
 * Resolve the effective execution target for a single analysis step.
 * Step-level override wins. Falls back to analysis default. Otherwise null.
 */
export function resolveExecutionTarget(
  analysisDefault: ExecutionTarget | null | undefined,
  stepOverride: ExecutionTarget | null | undefined,
): ExecutionTarget | null {
  return stepOverride ?? analysisDefault ?? null
}

export const TARGET_LABELS: Record<ExecutionTargetType, string> = {
  list: "Create List",
  workflow: "Start Workflow",
  channel_message: "Post to Channel",
}

export const TARGET_STYLES: Record<
  ExecutionTargetType,
  { bg: string; text: string; ring: string }
> = {
  list:            { bg: "bg-violet-500/10",  text: "text-violet-700 dark:text-violet-300",  ring: "ring-violet-500/30"  },
  workflow:        { bg: "bg-amber-500/10",   text: "text-amber-700 dark:text-amber-300",    ring: "ring-amber-500/30"   },
  channel_message: { bg: "bg-sky-500/10",     text: "text-sky-700 dark:text-sky-300",        ring: "ring-sky-500/30"     },
}

/**
 * Whether this execution target type has a real end-to-end execution path.
 * Phase 70B: `list` (direct).
 * Phase 70C: `workflow` and `channel_message` (draft-first).
 */
export function isExecutableTarget(type: ExecutionTargetType): boolean {
  return type === "list" || type === "workflow" || type === "channel_message"
}

/**
 * Whether this target uses the draft-first flow (generate → preview → confirm).
 * Distinct from `list` which uses a single generate step.
 */
export function isDraftFirstTarget(type: ExecutionTargetType): boolean {
  return type === "workflow" || type === "channel_message"
}
