// ── Unified AI Side-Channel Contract — Web client ────────────────────────────
//
// Frozen by Codex (`docs/superpowers/specs/2026-04-25-unified-ai-message-
// sidechannel-design.md`) and implemented backend-side by Gemini in
// `v0.6.51` (`metadata.ai_sidecar`, normative SSE envelope). This module
// defines the canonical Web type for that contract and a single
// `normalizeAISidecar()` entry point that EVERY AI surface in the app
// should funnel its metadata through, so reasoning / tool / usage
// rendering stays consistent across:
//
//   • AI DM bubbles (`app/workspace/dms/[id]/page.tsx`)
//   • Channel `/ask` replies (`components/message/message-item.tsx`)
//   • Canvas AI Dock (`components/layout/canvas-ai-dock.tsx`)
//
// The normalizer accepts:
//   1. The canonical `metadata.ai_sidecar` object.
//   2. Gemini's looser persisted shape (`reasoning: string`,
//      `tool_calls[].arguments/result`).
//   3. Codex's spec shape (`reasoning: { summary, segments[] }`,
//      `tool_calls[].input_summary/output_summary/status`).
//   4. Legacy flat `metadata.reasoning` / `metadata.tool_calls` /
//      `metadata.usage` from in-flight rollouts.
//
// Anything malformed degrades silently — UIs render whatever survives.
//
// `parseAIStreamEvent()` does the same dual-mode trick for SSE: it prefers
// the new `{ kind, message_id, payload }` envelope but transparently
// falls back to the legacy `event:`-name + `{ text }` shape so the canvas
// AI Dock keeps working against any backend that hasn't switched yet.

// ─── Canonical persisted shape ───────────────────────────────────────────────

export type ReasoningSegmentKind = "thought" | "step" | "note"

export interface ReasoningSegment {
  text: string
  kind: ReasoningSegmentKind
}

export interface AIReasoning {
  summary?: string
  segments: ReasoningSegment[]
}

export type AIToolStatus = "pending" | "running" | "success" | "failed"

export interface AIToolCall {
  id: string
  name: string
  status: AIToolStatus
  input_summary?: string
  output_summary?: string
  duration_ms?: number
}

export interface AIUsage {
  input_tokens?: number
  output_tokens?: number
  total_tokens?: number
  cost_usd?: number
}

export interface AISidecar {
  reasoning?: AIReasoning
  tool_calls?: AIToolCall[]
  usage?: AIUsage
  analysis?: any
}

// ─── Normalizer ──────────────────────────────────────────────────────────────

// Strict-typing-friendly truthy check.
const isObj = (v: unknown): v is Record<string, unknown> =>
  typeof v === "object" && v !== null && !Array.isArray(v)

function normalizeReasoning(input: unknown): AIReasoning | undefined {
  if (!input) return undefined
  // Spec shape: { summary, segments[] }
  if (isObj(input)) {
    const summary = typeof input.summary === "string" ? input.summary : undefined
    const segs = Array.isArray(input.segments) ? input.segments : []
    const segments: ReasoningSegment[] = segs
      .map((s: any) => {
        if (typeof s === "string") return { text: s, kind: "thought" as const }
        if (isObj(s) && typeof s.text === "string") {
          const rawKind = typeof s.kind === "string" ? s.kind : ""
          const kind: ReasoningSegmentKind = (["thought", "step", "note"] as const)
            .includes(rawKind as ReasoningSegmentKind)
            ? (rawKind as ReasoningSegmentKind)
            : "thought"
          return { text: s.text, kind }
        }
        return null
      })
      .filter((s): s is ReasoningSegment => !!s)
    if (!summary && segments.length === 0) return undefined
    return { summary, segments }
  }
  // Gemini's flat-string shape — collapse into a single thought segment so
  // the spec-driven UI can render it without a special branch.
  if (typeof input === "string" && input.trim()) {
    return { summary: input, segments: [{ text: input, kind: "thought" }] }
  }
  return undefined
}

function normalizeToolCalls(input: unknown): AIToolCall[] | undefined {
  if (!Array.isArray(input)) return undefined
  const out: AIToolCall[] = []
  input.forEach((raw: any, i: number) => {
    if (!isObj(raw)) return
    const id = typeof raw.id === "string" ? raw.id : `tc-${i}`
    const name = typeof raw.name === "string" ? raw.name
      : typeof raw.tool_name === "string" ? raw.tool_name
      : typeof raw.label === "string" ? raw.label
      : "tool"
    const rawStatus = (typeof raw.status === "string" ? raw.status : "success").toLowerCase()
    const status: AIToolStatus = (["pending", "running", "success", "failed"] as const).includes(rawStatus as AIToolStatus)
      ? (rawStatus as AIToolStatus)
      : "success"
    const input_summary = typeof raw.input_summary === "string" ? raw.input_summary
      : typeof raw.arguments === "string" ? raw.arguments
      : typeof raw.input === "string" ? raw.input
      : undefined
    const output_summary = typeof raw.output_summary === "string" ? raw.output_summary
      : typeof raw.result === "string" ? raw.result
      : typeof raw.output === "string" ? raw.output
      : undefined
    const duration_ms = typeof raw.duration_ms === "number" ? raw.duration_ms
      : typeof raw.durationMs === "number" ? raw.durationMs
      : undefined
    out.push({ id, name, status, input_summary, output_summary, duration_ms })
  })
  return out.length > 0 ? out : undefined
}

function normalizeUsage(input: unknown): AIUsage | undefined {
  if (!isObj(input)) return undefined
  const u: AIUsage = {}
  if (typeof input.input_tokens === "number") u.input_tokens = input.input_tokens
  if (typeof input.output_tokens === "number") u.output_tokens = input.output_tokens
  if (typeof input.total_tokens === "number") u.total_tokens = input.total_tokens
  if (typeof input.cost_usd === "number") u.cost_usd = input.cost_usd
  // Some providers spell the totals differently; accept aliases.
  if (u.input_tokens === undefined && typeof input.prompt_tokens === "number") u.input_tokens = input.prompt_tokens
  if (u.output_tokens === undefined && typeof input.completion_tokens === "number") u.output_tokens = input.completion_tokens
  if (u.total_tokens === undefined && (u.input_tokens ?? 0) + (u.output_tokens ?? 0) > 0) {
    u.total_tokens = (u.input_tokens ?? 0) + (u.output_tokens ?? 0)
  }
  return Object.keys(u).length > 0 ? u : undefined
}

/**
 * Returns a canonical `AISidecar` view of a message's metadata bag, accepting
 * both the v0.6.51 canonical shape and any legacy / Gemini-flat / spec-only
 * partials. Returns `null` (rather than an empty object) when nothing
 * AI-side-channel-shaped is present so call sites can do a fast nullish check.
 */
export function normalizeAISidecar(metadata: unknown): AISidecar | null {
  if (!isObj(metadata)) return null

  // Canonical bag (`metadata.ai_sidecar`) wins when present per the spec
  // ("`ai_sidecar` is the source of truth").
  const canonical = isObj(metadata.ai_sidecar) ? metadata.ai_sidecar : undefined

  const reasoning = normalizeReasoning(canonical?.reasoning ?? metadata.reasoning ?? metadata.thinking)
  const tool_calls = normalizeToolCalls(canonical?.tool_calls ?? metadata.tool_calls ?? metadata.tools)
  const usage = normalizeUsage(canonical?.usage ?? metadata.usage)
  const analysis = canonical?.analysis ?? metadata.analysis

  if (!reasoning && !tool_calls && !usage && !analysis) return null
  return { reasoning, tool_calls, usage, analysis }
}

// ─── Streaming envelope parser ───────────────────────────────────────────────

/**
 * Normative kinds defined by the unified contract. `answer` is the streamed
 * answer body (legacy `event: chunk`); the rest are side-channel updates.
 */
export type AIStreamKind = "reasoning" | "tool_call" | "usage" | "answer" | "error"

export interface AIStreamEvent {
  kind: AIStreamKind
  message_id?: string
  /** Plain text payload — set for `answer` / `reasoning` kinds. */
  text?: string
  /** Tool-call delta — set for `tool_call` kind. May be partial; merge by `id`. */
  tool?: Partial<AIToolCall>
  /** Usage payload — set for `usage` kind. */
  usage?: AIUsage
  /** Error message — set for `kind: "error"`. */
  error?: string
}

/**
 * Parse one SSE record (`event:` name + `data:` JSON) into a normalized
 * `AIStreamEvent`. Prefers the new normative envelope shape
 * (`{ kind, message_id, payload }`) but transparently degrades to:
 *   • Legacy `event: chunk` / `event: reasoning` / `event: error` with
 *     `{ text }` payloads (canvas SSE today).
 *   • Plain `{ text, content, delta }` payloads on the default channel.
 *
 * Returns null if the record carries no usable signal so callers can ignore
 * heartbeats safely.
 */
export function parseAIStreamEvent(eventName: string, dataJson: string): AIStreamEvent | null {
  let parsed: any
  try { parsed = JSON.parse(dataJson) } catch { return null }
  if (!isObj(parsed)) return null

  // 1. Normative envelope: `{ kind, message_id, payload }`.
  const k = typeof parsed.kind === "string" ? parsed.kind.toLowerCase() : null
  if (k && ["reasoning", "tool_call", "usage", "answer", "error", "chunk"].includes(k)) {
    const kind: AIStreamKind = (k === "chunk" ? "answer" : (k as AIStreamKind))
    const message_id = typeof parsed.message_id === "string" ? parsed.message_id : undefined
    const payload = isObj(parsed.payload) ? parsed.payload : parsed
    if (kind === "answer" || kind === "reasoning") {
      const text = typeof payload.text === "string" ? payload.text
        : typeof payload.content === "string" ? payload.content
        : typeof payload.delta === "string" ? payload.delta
        : ""
      return { kind, message_id, text }
    }
    if (kind === "tool_call") {
      const tcs = normalizeToolCalls([payload])
      return { kind, message_id, tool: tcs?.[0] ?? { id: "tc", name: "tool", status: "pending" } }
    }
    if (kind === "usage") {
      return { kind, message_id, usage: normalizeUsage(payload) ?? {} }
    }
    if (kind === "error") {
      return {
        kind: "error",
        message_id,
        error: typeof payload.error === "string" ? payload.error
          : typeof payload.message === "string" ? payload.message
          : "Unknown error",
      }
    }
  }

  // 2. Legacy `event: <name>` form — translate the name to a normative kind.
  const legacyKindMap: Record<string, AIStreamKind | undefined> = {
    chunk: "answer", message: "answer", token: "answer",
    reasoning: "reasoning", thought: "reasoning",
    tool: "tool_call", tool_call: "tool_call",
    usage: "usage",
    error: "error",
  }
  const kind = legacyKindMap[eventName.toLowerCase()]
  if (!kind) return null

  if (kind === "answer" || kind === "reasoning") {
    const text = typeof parsed.text === "string" ? parsed.text
      : typeof parsed.content === "string" ? parsed.content
      : typeof parsed.delta === "string" ? parsed.delta
      : ""
    if (!text) return null
    return { kind, text }
  }
  if (kind === "tool_call") {
    const tcs = normalizeToolCalls([parsed])
    return { kind, tool: tcs?.[0] }
  }
  if (kind === "usage") {
    return { kind, usage: normalizeUsage(parsed) ?? {} }
  }
  if (kind === "error") {
    return {
      kind: "error",
      error: typeof parsed.error === "string" ? parsed.error
        : typeof parsed.message === "string" ? parsed.message
        : "Unknown error",
    }
  }
  return null
}

// ─── Token-estimate helpers (UI fallback only — see spec §"Frontend") ────────

/**
 * Rough token estimate at ~4 chars / token (OpenAI guidance). The unified
 * contract requires UIs to PREFER `metadata.ai_sidecar.usage` when present;
 * this helper is the documented fallback when the backend hasn't supplied
 * real usage telemetry yet.
 */
export function estimateTokens(s: string): number {
  if (!s) return 0
  return Math.max(1, Math.round(s.length / 4))
}

/** Strip HTML tags before estimating tokens. */
export function plainTextOf(html: string): string {
  if (typeof document === "undefined") return html.replace(/<[^>]+>/g, "")
  const tmp = document.createElement("div")
  tmp.innerHTML = html
  return tmp.textContent || ""
}
