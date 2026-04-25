# Unified AI Message Side-Channel Contract Design

## Goal

Define one durable, replayable, and reusable side-channel contract for all AI-generated messages in Relay so reasoning, tool activity, and usage telemetry stop living as page-specific heuristics.

This phase intentionally covers one cross-cutting product seam:

1. unify AI side-channel metadata under one stable message contract
2. support both live streaming and persisted replay
3. make AI DM, channel `/ask`, and canvas AI consume the same semantics

## Why This Phase Exists

Relay now has multiple AI surfaces:

- AI DM
- channel `/ask`
- canvas AI

The UI is already beginning to expose richer AI context:

- thinking/reasoning panels
- tool timelines
- token and usage chips

But those signals are not yet consistently modeled as product data.

Current problems:

- some surfaces infer token counts heuristically
- some live views can show intermediate AI state that disappears on reload
- AI side-channel semantics are drifting toward per-surface behavior
- backend and frontend risk building separate interpretations of "reasoning", "tool activity", and "usage"

This phase prevents that fragmentation by making AI side-channel metadata a first-class part of the AI message contract.

## Scope

### Included

- define a single `metadata.ai_sidecar` shape for AI-generated messages
- cover three side-channel groups:
  - `reasoning`
  - `tool_calls`
  - `usage`
- define streaming semantics for side-channel updates
- define persistence semantics for final message storage
- apply the contract to:
  - AI DM
  - channel `/ask`
  - canvas AI

### Excluded

- provider-specific raw transcript dumps
- fully general agent workflow execution graphs
- prompt logging or prompt persistence
- new billing dashboards
- a redesign of the base message schema outside additive metadata

## Product Definition

### 1. AI Messages Carry A Stable Sidecar

Any AI-generated message may include a stable metadata object:

- `metadata.ai_sidecar`

This is the only supported place for durable reasoning/tool/usage metadata in this phase.

Do not create surface-specific alternatives such as:

- `metadata.reasoning_panel`
- `metadata.tokens`
- `metadata.canvas_debug`

All of those concerns belong inside the single sidecar object.

### 2. The Sidecar Must Survive Reload

The phase is not complete if rich AI context only exists during streaming.

Users must be able to:

- watch the data live while the model is responding
- refresh the page later
- still see the same reasoning/tool/usage structure on the final persisted AI message

### 3. The Contract Is Cross-Surface

The same sidecar semantics apply to:

- AI DM replies
- channel `/ask` replies
- canvas AI replies

The UI treatment may differ by surface, but the underlying message contract must remain the same.

## Data Model

Every eligible AI message may include:

```json
{
  "metadata": {
    "ai_sidecar": {
      "reasoning": {
        "summary": "string",
        "segments": [
          { "text": "string", "kind": "thought" }
        ]
      },
      "tool_calls": [
        {
          "id": "string",
          "name": "string",
          "status": "pending",
          "input_summary": "string",
          "output_summary": "string",
          "duration_ms": 1200
        }
      ],
      "usage": {
        "input_tokens": 123,
        "output_tokens": 456,
        "total_tokens": 579,
        "cost_usd": 0.0123
      }
    }
  }
}
```

### A. `reasoning`

Purpose:

- display a compact reasoning summary
- optionally render structured live/persisted reasoning segments

Fields:

- `summary` optional
- `segments[]` optional

Segment fields:

- `text`
- `kind`

Allowed `kind` values in Phase 1:

- `thought`
- `step`
- `note`

### B. `tool_calls`

Purpose:

- show AI-invoked tool activity in a stable timeline

Fields per entry:

- `id`
- `name`
- `status`
- `input_summary`
- `output_summary`
- `duration_ms`

Allowed `status` values:

- `pending`
- `running`
- `success`
- `failed`

### C. `usage`

Purpose:

- expose accurate usage data when available

Fields:

- `input_tokens`
- `output_tokens`
- `total_tokens`
- `cost_usd`

All fields are optional individually, but the object key `usage` must be stable when usage data is supplied.

## Eligibility Rules

The contract applies to AI-generated messages only.

Phase 1 eligible sources:

- AI DM responses
- channel `/ask` responses
- canvas AI responses

Human-authored messages must not be forced to carry empty AI sidecar objects.

## Streaming Contract

This phase must support live rendering of side-channel information during generation.

### Streaming Principle

The live stream is a preview layer.
The persisted message is the final source of truth.

### Event Envelope

Streaming payloads must distinguish side-channel payloads by explicit kind, not by surface-specific event names.

Normative stream envelope:

```json
{
  "kind": "reasoning",
  "message_id": "ai-msg-temp-or-final-id",
  "payload": {}
}
```

Allowed `kind` values in Phase 1:

- `reasoning`
- `tool_call`
- `usage`
- `answer`

This is the required parser contract for all three Phase 1 surfaces.

### Behavior

During streaming:

- reasoning segments may arrive incrementally
- tool call rows may be created and updated
- usage may arrive only near or at completion
- answer text may stream independently of side-channel updates

On completion:

- the final persisted message should include the normalized `metadata.ai_sidecar`
- the frontend should replace transient stream state with the stored sidecar state

## Persistence Contract

When an AI message is finalized and stored:

- `metadata.ai_sidecar` should contain the normalized durable form

That durable form should be replayable through normal message history endpoints for:

- DM history
- channel message history
- canvas history/message surface where applicable

### Persistence Principle

Do not persist provider-specific raw wire payloads if they make the contract unstable.

Persist the normalized product shape only.

## Transition And Mixed-Version Rollout

Phase 1 must support a safe transition from legacy flat metadata fields to the new sidecar shape.

### Legacy Fields In Scope

Existing or in-flight clients may still encounter flat metadata fields such as:

- `metadata.reasoning`
- `metadata.tool_calls`
- `metadata.usage`

### Read Rule

During rollout, backend and frontend should treat `metadata.ai_sidecar` as canonical when present.

If `metadata.ai_sidecar` is absent, they must synthesize an equivalent sidecar view from legacy flat fields for replay and rendering.

That means:

- old rows remain readable
- mixed-version clients do not lose AI context
- replay semantics stay stable while rollout is incomplete

### Write Rule

During rollout, newly finalized AI messages should write:

- canonical `metadata.ai_sidecar`

And may additionally dual-write the legacy flat fields if needed for short-term client compatibility.

The rollout rule is:

- `ai_sidecar` is the source of truth
- legacy flat fields are temporary compatibility output only

### Migration Direction

This phase does not require a blocking historical migration before ship.

It does require:

- dual-read support immediately
- optional dual-write during rollout
- a later cleanup phase can remove legacy writes once all supported clients consume `ai_sidecar`

## Frontend Consumption Rules

### 1. Render What Exists

If a message includes:

- `reasoning`
- `tool_calls`
- `usage`

the UI may render those sections.

If a section is absent, hide it cleanly.

### 2. No Heuristic Truth When Real Data Exists

Frontend heuristics such as "4 chars per token" may remain as temporary fallback behavior only when:

- no `usage` payload exists

Once `metadata.ai_sidecar.usage` exists, that becomes authoritative.

### 3. Streaming Is Temporary

The UI may maintain stream-local state while a response is in flight.

Once the final message is persisted:

- persisted `metadata.ai_sidecar` overrides temporary stream-local state

## Error Handling

- if a provider emits no reasoning data, render only answer/tool/usage sections that exist
- if a provider emits no usage data, usage UI stays hidden or falls back to existing heuristic display where allowed
- if tool activity is absent, no tool timeline is shown
- malformed sidecar payloads should fail soft and not break the underlying message rendering

## Testing Strategy

This phase should land with four test groups.

### 1. Normalization Tests

- provider output is normalized into `metadata.ai_sidecar`
- missing optional fields do not break storage
- ineligible human messages do not receive forced AI sidecars

### 2. Streaming Contract Tests

- stream emits side-channel kinds with stable routing semantics
- final persisted message contains the normalized durable sidecar
- frontend can safely swap from transient stream state to persisted state

### 3. Replay Tests

- AI DM history replays sidecar data
- channel `/ask` history replays sidecar data
- canvas AI history replays sidecar data

### 4. Degradation Tests

- no reasoning still renders valid AI messages
- no usage still renders valid AI messages
- malformed sidecar does not break the whole message component

## Ownership And Delivery Model

### Gemini Ownership

Gemini owns backend/API/test work:

- sidecar schema normalization
- streaming payload contract
- persisted metadata contract
- history replay behavior

### Windsurf Ownership

Windsurf owns Web/UI consumption:

- AI DM sidecar rendering
- channel `/ask` sidecar rendering
- canvas AI sidecar rendering
- stream-to-persisted-state handoff behavior

### Codex Ownership

Codex owns:

- phase decomposition
- contract freezing
- cross-surface consistency review
- collaboration document updates
- integration control

## Release Intent

This phase should ship as a platform-quality release rather than a single-page feature release.

The release story should be:

`AI messages now carry durable reasoning, tool, and usage context everywhere they appear.`
