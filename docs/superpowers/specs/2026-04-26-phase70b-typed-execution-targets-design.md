# Phase 70B Typed Execution Targets Design

## Summary

Phase 70B gives Relay a typed execution taxonomy for AI output:

`AI result -> typed execution targets -> deterministic execution routing`

The contract applies across all AI reply surfaces, but the first real execution UX lands in Canvas AI. This phase prevents Web from guessing execution intent from prose and prepares a consistent path from AI proposal to execution object.

## Product Definition

`All AI replies may carry typed execution targets. Each analysis can declare a default execution target, and each next step can override it. The execution UX lands in Canvas first.`

## Scope

### In Scope

- define one shared execution-target contract across:
  - Canvas AI
  - channel `/ask`
  - AI DM
- support two layers of execution target declaration:
  - analysis-level default target
  - step-level override target
- freeze deterministic inheritance:
  - `step override > analysis default`
- support first-release target types:
  - `list`
  - `workflow`
  - `channel_message`
- land real execution UX in Canvas AI first
- keep other AI surfaces contract-compatible even if their execution UX remains light

### Out of Scope

- full execution UX parity across DM, `/ask`, and Canvas in the same release
- automatic execution without user confirmation
- natural-language target inference in Web
- broad workflow-builder UX expansion
- arbitrary multi-hop routing trees

## User Experience

### First-Release Strategy

Phase 70B deliberately separates protocol rollout from full interaction rollout:

- protocol:
  - all AI reply surfaces may return typed execution targets
- interaction:
  - Canvas AI gets the first real execution entry points

This prevents contract fragmentation while avoiding three simultaneous heavy UI implementations.

### Canvas UX

When an AI result appears in Canvas:

- the Dock may show a default execution target for the overall analysis
- each next step may show:
  - inherited default target
  - or an explicit override target
- users should be able to understand, without guesswork, where a suggestion is intended to go next

### Target Semantics

#### `list`

The suggestion is best converted into a list/list-item execution flow.

#### `workflow`

The suggestion is best converted into a workflow-oriented execution path.

#### `channel_message`

The suggestion is best published back into the current channel after explicit confirmation.

## System Design

### Canonical Contract

The contract is layered:

```json
{
  "analysis": {
    "default_execution_target": {
      "type": "list|workflow|channel_message"
    }
  },
  "next_steps": [
    {
      "text": "string",
      "rationale": "string",
      "action_hint": "summarize|compare|decide|share|plan|investigate|custom",
      "execution_target": {
        "type": "list|workflow|channel_message"
      }
    }
  ]
}
```

### Inheritance Rule

Execution target resolution must be deterministic:

1. if `next_steps[i].execution_target` exists, use it
2. otherwise use `analysis.default_execution_target`
3. Web must not invent a target from prose when both are absent

This rule is mandatory across every AI reply surface.

### Malformed Target Rule

Malformed targets must be handled consistently:

- backend should emit only valid target types from the allowed set:
  - `list`
  - `workflow`
  - `channel_message`
- if backend cannot validate a target, it must omit that target field rather than emit an ambiguous partial object
- shared Web normalization must treat malformed or unknown target objects as absent
- after normalization, fallback to the analysis default target is allowed only when the step target is absent
- Web must never preserve a malformed step target as if it were authoritative

### First-Release Contract Depth

For Phase 70B, target typing should be intentionally shallow:

- required:
  - `type`
- allowed but ignored in first release:
  - extra target-specific payload fields
- not required yet:
  - deep target-specific payloads
  - routing trees
  - nested execution chains

The goal is to freeze the taxonomy and inheritance rules first, then deepen each target family's payload later.

First-release pass-through rule:

- backend may reserve future target-specific payload fields
- Web must preserve unknown extra fields in normalized data where practical
- first-release behavior must be driven only by `type`
- Web must not invent first-release UX from extra payload fields

### Current Target Types

#### `list`

This target connects directly to the Phase 70A execution path.

Expected behavior:

- Canvas may surface `Create list from plan`
- if a step-level target is `list`, it explicitly confirms that this suggestion is list-shaped

First-release required real execution flow:

- `list` is the mandatory target type that must have a true end-to-end execution path in Canvas during Phase 70B

#### `workflow`

This target signals workflow-oriented execution intent.

Expected first-release behavior:

- freeze the typed target in backend and UI
- provide clear Canvas-facing affordance or placeholder for workflow execution
- do not require a fully featured workflow builder in the same phase

#### `channel_message`

This target means:

- publish to the current channel
- only after explicit user confirmation

It does not mean:

- auto-send
- multi-destination routing
- DM fallback

## Cross-Surface Applicability

The contract must be available on:

- Canvas AI
- channel `/ask`
- AI DM

But first-release interaction depth differs:

- Canvas AI:
  - real execution-oriented UX
- `/ask` and AI DM:
  - contract-compatible rendering and safe display
  - preserve the normalized typed target object unchanged after shared normalization
  - no requirement for full execution UX parity yet

Non-Canvas preservation rule:

- AI DM and `/ask` must consume the same normalized execution-target contract used by Canvas
- they may render it more lightly
- they must not fork, rename, or reinterpret target semantics by surface

## Backend / API Boundary

Phase 70B backend work should focus on typed contract unification, deterministic inheritance support, and Canvas-first execution routing support.

### Gemini Scope

Gemini owns backend/API/tests only.

Expected backend responsibilities:

- add typed execution-target support to shared AI response contracts
- apply the same execution-target schema across Canvas AI, `/ask`, and AI DM responses
- ensure Canvas AI outputs can carry:
  - analysis default target
  - step-level target overrides
- define minimal server-side target semantics for:
  - `list`
  - `workflow`
  - `channel_message`
- support the Canvas-first execution entry points that consume these targets
- add tests for:
  - target schema presence
  - inheritance-compatible response shapes
  - cross-surface contract stability

### Backend Constraints

- do not require Web to infer targets from text
- do not auto-execute any target
- do not overload `action_hint` as the target itself
- do not make DM or `/ask` execution UX a backend blocker for Canvas-first rollout

## Web / UI Boundary

Windsurf owns all Web delivery.

### Windsurf Scope

- update shared AI result typing/normalization to accept the new execution-target contract
- render default target and step-level target information in Canvas AI Dock
- implement first-release Canvas execution affordances based on resolved target type
- keep `/ask` and AI DM contract-compatible even if their execution UI remains minimal
- preserve deterministic resolution:
  - step target if present
  - otherwise analysis default
  - otherwise no execution target

### Web Constraints

- do not infer targets from prose
- do not auto-execute
- do not fake workflow completeness if backend only provides typed placeholders
- do not fork different target schemas by surface

## AI-Native Behavior

This phase is AI-native because it upgrades AI output from “good-looking advice” into typed, routable execution intent.

The intended product feeling is:

- AI produces recommendations
- those recommendations already know their likely execution shape
- the product can route them consistently without natural-language guesswork

This prepares later phases such as:

- stronger workflow creation
- channel-publish flows from step targets
- richer target-specific payloads
- execution history linked back to AI analysis origin

## Error Handling

Required behaviors:

- if a response has no execution target at either layer, Web shows no execution route rather than guessing
- if a step target is malformed or unknown, shared normalization must treat it as absent before resolution
- after normalization, fallback to the analysis default target is allowed only when the step target is absent
- if Canvas has an execution affordance for one target type but not another yet, unsupported targets must still display clearly without pretending to be executable
- DM and `/ask` must remain render-safe when execution-target fields appear

## Acceptance Criteria

### Core

1. All AI reply surfaces may return the same typed execution-target contract.
2. The contract supports:
   - `analysis.default_execution_target`
   - `next_steps[].execution_target`
3. Execution resolution is deterministic:
   - step override first
   - otherwise analysis default
   - otherwise no target
4. Canvas AI Dock can display both default targets and step-level targets correctly.
5. Canvas lands a real end-to-end `list` target-driven execution flow while `workflow` and `channel_message` remain clearly typed and visible even if their UX is lighter.
6. AI DM and `/ask` remain compatible with the new contract even if their execution UX stays lighter.

### Architecture

1. Target typing is global across AI reply surfaces.
2. Interaction rollout is Canvas-first without fragmenting the protocol or forking target semantics by surface.
3. `action_hint` and `execution_target` remain separate concepts.

## Testing Strategy

### Backend

- cross-surface response-shape tests for execution targets
- inheritance-shape tests for default + override coexistence
- malformed/absent target handling tests
- Canvas-first routing support tests for target-bearing results

### Web

- shared normalization tests for typed execution targets
- Canvas rendering tests for:
  - default target
  - override target
  - inherited target
- contract-compatibility tests for AI DM and `/ask`
- no-guessing tests proving missing targets do not silently become inferred actions

## Release Notes Guidance

This release should be described as:

`AI replies can now carry typed execution targets, allowing Canvas AI to show where recommendations are meant to go next while keeping the contract consistent across DM, channel AI, and canvas.`

## Ownership

- Gemini: backend/API/test support only
- Windsurf: all Web/UI implementation
- Codex: taxonomy definition, contract freeze, collaboration docs, integration control, follow-up phase planning
