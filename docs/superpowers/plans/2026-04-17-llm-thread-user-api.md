# LLM Gateway, User API, Threading, And SSE Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a provider-based LLM gateway with configurable OpenAI, OpenAI-compatible, OpenRouter, and Gemini support; ship `GET /api/v1/users`, message thread APIs, and a real `POST /api/v1/ai/execute` SSE endpoint.

**Architecture:** Keep the HTTP handlers thin and move provider logic into a dedicated `internal/llm` package. Use configuration files plus env overrides for provider settings, then expose a single gateway interface so later Agentic orchestration can reuse the same streaming abstraction.

**Tech Stack:** Go, Gin, GORM, SQLite, SSE, `google.golang.org/genai`, standard HTTP client, YAML config.

---

### Task 1: Add LLM Config And Gateway Skeleton

**Files:**
- Create: `apps/api/internal/config/config.go`
- Create: `apps/api/internal/config/config_test.go`
- Create: `apps/api/internal/llm/types.go`
- Create: `apps/api/internal/llm/gateway.go`
- Create: `apps/api/internal/llm/gateway_test.go`
- Create: `apps/api/config/llm.base.yaml`
- Create: `apps/api/config/llm.example.yaml`
- Create: `apps/api/config/llm.local.yaml`
- Create: `apps/api/config/llm.secrets.local.yaml`
- Modify: `.gitignore`
- Modify: `apps/api/go.mod`

- [ ] **Step 1: Write failing config tests**
- [ ] **Step 2: Run config tests to verify failure**
- [ ] **Step 3: Implement YAML loading + env override logic**
- [ ] **Step 4: Write failing gateway selection tests**
- [ ] **Step 5: Run gateway tests to verify failure**
- [ ] **Step 6: Implement provider registry and request/stream abstractions**
- [ ] **Step 7: Run config and gateway tests to verify pass**

### Task 2: Add User Listing And Threaded Message Domain Support

**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Modify: `apps/api/internal/db/seed.go`
- Modify: `apps/api/internal/handlers/collaboration.go`
- Create: `apps/api/internal/handlers/collaboration_test.go`
- Modify: `apps/api/main.go`

- [ ] **Step 1: Write failing handler tests for `GET /api/v1/users` and `GET /api/v1/messages/:id/thread`**
- [ ] **Step 2: Run those tests to verify failure**
- [ ] **Step 3: Add `ThreadID` and `ReplyCount` to `Message` and migrate/seed thread data**
- [ ] **Step 4: Implement user list and thread detail handlers**
- [ ] **Step 5: Update message creation logic so replies increment parent `ReplyCount`**
- [ ] **Step 6: Run handler tests to verify pass**

### Task 3: Add Real LLM Providers And SSE Execution Endpoint

**Files:**
- Create: `apps/api/internal/llm/openrouter.go`
- Create: `apps/api/internal/llm/openai_compatible.go`
- Create: `apps/api/internal/llm/gemini.go`
- Create: `apps/api/internal/handlers/ai.go`
- Create: `apps/api/internal/handlers/ai_test.go`
- Modify: `apps/api/main.go`
- Modify: `apps/api/internal/realtime/hub.go` (only if shared event helpers help SSE/AI follow-up)

- [ ] **Step 1: Write failing SSE handler tests for `POST /api/v1/ai/execute`**
- [ ] **Step 2: Run the tests to verify failure**
- [ ] **Step 3: Implement provider adapters**
- [ ] **Step 4: Implement SSE event streaming handler using the gateway**
- [ ] **Step 5: Run AI handler tests to verify pass**

### Task 4: Wire Startup, Docs, And Collaboration Handoff

**Files:**
- Modify: `apps/api/main.go`
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`

- [ ] **Step 1: Wire config + LLM gateway into API startup**
- [ ] **Step 2: Update README and changelog for new config and APIs**
- [ ] **Step 3: Update `AGENT-COLLAB.md` task board, superpower state, and Gemini handoff note**
- [ ] **Step 4: Run final verification (`pnpm build`, `go test ./...`, `go build ./...`)**
- [ ] **Step 5: Commit and tag release**
