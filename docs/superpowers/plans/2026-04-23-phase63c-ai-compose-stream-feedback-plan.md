# Phase 63C AI Compose Stream And Feedback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add SSE streaming for grounded channel/thread compose suggestions and persist per-suggestion compose feedback without changing the existing sync compose contract.

**Architecture:** Reuse the existing compose grounding pipeline by extracting shared preparation helpers from `ai.go`, then add a streaming handler that emits a minimal SSE contract and a dedicated persistence path for suggestion feedback. Keep sync and stream suggestion parsing aligned by sharing the final compose-response builder.

**Tech Stack:** Go, Gin, GORM, SQLite, existing LLM gateway, SSE, Go handler tests

---

## File Map

- Modify: `apps/api/internal/handlers/ai.go`
  - extract shared compose preparation
  - add `ComposeAIStream`
  - add `SubmitAIComposeFeedback`
  - add any small helper functions for request normalization and SSE events
- Modify: `apps/api/internal/domain/models.go`
  - add `AIComposeFeedback`
- Modify: `apps/api/internal/db/db.go`
  - ensure `AIComposeFeedback` is auto-migrated
- Modify: `apps/api/main.go`
  - register new routes
- Modify: `apps/api/internal/handlers/collaboration_test.go`
  - add red/green handler coverage for compose stream and compose feedback
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/phase8-api-expansion.md`
- Create: `docs/releases/v0.6.16.md`

### Task 1: Add Failing Handler Tests

**Files:**
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] **Step 1: Write the failing tests**

Add tests for:

- `TestPhase63CComposeStreamContract`
- `TestPhase63CComposeFeedbackContract`

Cover:

- `POST /api/v1/ai/compose/stream` emits `start`, at least one `suggestion.done`, and `done`
- thread-scoped stream works with `thread_id`
- stream returns `Content-Type: text/event-stream`
- stream emits an SSE `error` event when the gateway stream fails after headers are written
- stream returns `503` when `AIGateway == nil`
- `POST /api/v1/ai/compose/:id/feedback` creates a row
- repeated feedback from same user and compose id updates the same row
- invalid feedback value returns `400`

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase63C(ComposeStreamContract|ComposeFeedbackContract)' -count=1
```

Expected:

- FAIL because the stream route and compose feedback handler do not exist yet

- [ ] **Step 3: Commit the red tests**

```bash
git add apps/api/internal/handlers/collaboration_test.go
git commit -m "test: add phase63c compose stream coverage"
```

### Task 2: Implement Shared Compose Preparation And Streaming Compose

**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `apps/api/main.go`

- [ ] **Step 1: Extract shared compose preparation**

Create a helper that:

- validates `channel_id`, `thread_id`, and `intent`
- loads `channel`
- loads optional `threadParent`
- loads `recentMessages`
- loads `knowledgeContext`
- loads `entityMatches`
- returns one structured context object consumed by sync and stream handlers

- [ ] **Step 2: Implement `ComposeAIStream`**

Use the shared compose context helper, call `AIGateway.Stream`, and emit:

- `start`
- `suggestion.delta`
- `suggestion.done`
- `done`
- `error`

Do not persist anything.
Ensure stream ids are stable:

- generate one `request_id`
- normalize parsed suggestions so each one has a stable `compose-*` id
- use the first suggestion id for provisional `suggestion.delta` events

- [ ] **Step 3: Wire the route**

Register:

```go
v1.POST("/ai/compose/stream", handlers.ComposeAIStream)
```

- [ ] **Step 4: Run focused tests**

Run:

```bash
cd apps/api && go test ./internal/handlers -run TestPhase63CComposeStreamContract -count=1
```

Expected:

- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers/ai.go apps/api/main.go
git commit -m "feat: add streaming ai compose handler"
```

### Task 3: Implement Compose Feedback Persistence

**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `apps/api/main.go`

- [ ] **Step 1: Add the feedback model**

Add:

```go
type AIComposeFeedback struct {
    ID             string    `gorm:"primaryKey" json:"id"`
    ComposeID      string    `gorm:"index;uniqueIndex:idx_ai_compose_feedback_scope" json:"compose_id"`
    UserID         string    `gorm:"uniqueIndex:idx_ai_compose_feedback_scope" json:"user_id"`
    ChannelID      string    `gorm:"index" json:"channel_id"`
    ThreadID       string    `gorm:"index" json:"thread_id"`
    Intent         string    `json:"intent"`
    Feedback       string    `json:"feedback"`
    SuggestionText string    `json:"suggestion_text"`
    Provider       string    `json:"provider"`
    Model          string    `json:"model"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}
```

- [ ] **Step 2: Auto-migrate the model**

Add `&domain.AIComposeFeedback{}` to the DB migration list.

- [ ] **Step 3: Implement `SubmitAIComposeFeedback`**

Validate:

- current user exists
- `feedback` is one of `up|down|edited`
- `channel_id` is present
- `intent` defaults to `reply`

Persist with upsert semantics by `(compose_id, user_id)`.
Persist `provider` and `model` only from the request payload in this phase.

- [ ] **Step 4: Wire the route**

Register:

```go
v1.POST("/ai/compose/:id/feedback", handlers.SubmitAIComposeFeedback)
```

- [ ] **Step 5: Run focused tests**

Run:

```bash
cd apps/api && go test ./internal/handlers -run TestPhase63CComposeFeedbackContract -count=1
```

Expected:

- PASS

- [ ] **Step 6: Commit**

```bash
git add apps/api/internal/domain/models.go apps/api/internal/db/db.go apps/api/internal/handlers/ai.go apps/api/main.go
git commit -m "feat: persist ai compose feedback"
```

### Task 4: Narrow Docs And Verification

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/phase8-api-expansion.md`
- Create: `docs/releases/v0.6.16.md`

- [ ] **Step 1: Run API verification**

Run:

```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase63C(ComposeStreamContract|ComposeFeedbackContract)' -count=1
cd apps/api && go test ./...
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...
```

Expected:

- PASS

- [ ] **Step 2: Run workspace verification**

Run:

```bash
pnpm --filter relay-agent-workspace lint
pnpm --filter relay-agent-workspace exec tsc --noEmit
```

Expected:

- PASS

- [ ] **Step 3: Update docs**

Document:

- new endpoints
- stream SSE contract
- Windsurf handoff for composer streaming and feedback UI

- [ ] **Step 4: Commit**

```bash
git add README.md CHANGELOG.md docs/AGENT-COLLAB.md docs/phase8-api-expansion.md docs/releases/v0.6.16.md
git commit -m "docs: record phase63c ai compose release"
```
