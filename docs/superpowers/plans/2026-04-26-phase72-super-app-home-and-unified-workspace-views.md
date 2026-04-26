# Phase 72 Super App Home And Unified Workspace Views Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild Home below Workspace Overview into a Slack-like workbench centered on Today and My Work, while adding a lightweight WorkspaceView registry for future Super App list/calendar/search/report/form/channel-message surfaces.

**Architecture:** Extend existing Home aggregation instead of replacing it wholesale. Add a shallow WorkspaceView registry with bounded CRUD/query APIs and metadata-only actions. Windsurf preserves the current Workspace Overview top section, replaces the lower Home layout with backend-ranked workbench sections, and exposes Apps & Tools plus lightweight WorkspaceView entries.

**Tech Stack:** Go/Gin API, GORM, SQLite, existing Home/Activity/List/Workflow handlers, Next.js 16, React 19, TypeScript, Zustand stores, existing workspace shell and Home dashboard components.

---

## File Structure

### Existing files likely involved

- `apps/api/internal/handlers/home.go`
  - extend Home aggregation fields for Today/My Work/Recent Channels/AI Suggestions/Apps & Tools/Activity
- `apps/api/internal/domain/*`
  - add lightweight WorkspaceView model
- `apps/api/main.go`
  - route WorkspaceView APIs
- `apps/web/components/layout/home-dashboard.tsx`
  - preserve Workspace Overview and rebuild lower Home layout
- `apps/web/components/layout/home-execution-blocks.tsx`
  - inspect/reuse execution data patterns
- `apps/web/stores/workspace-store.ts`
  - Home response typing/fetching
- `apps/web/types/index.ts`
  - add Home section and WorkspaceView types

### New files recommended

- `apps/api/internal/handlers/workspace_views.go`
  - WorkspaceView CRUD/query handlers
- `apps/api/internal/handlers/phase72_home_test.go`
  - Home aggregation contract tests
- `apps/api/internal/handlers/phase72_workspace_views_test.go`
  - WorkspaceView API tests
- `apps/web/components/layout/home-workbench.tsx`
  - lower Home workbench composition
- `apps/web/components/layout/home-today-section.tsx`
  - Today section
- `apps/web/components/layout/home-my-work-section.tsx`
  - My Work section
- `apps/web/components/layout/home-apps-tools-section.tsx`
  - Apps & Tools section
- `apps/web/lib/workspace-views.ts`
  - WorkspaceView types/helpers if not colocated in global types

## Contract Freeze

- **Codex authority:** Phase 72 is frozen around Home workbench + shallow WorkspaceView registry.
- **Codex preflight gate:** Before Gemini/Windsurf implementation starts, Codex owns this contract and must answer implementation questions by updating this plan/spec rather than allowing ad-hoc API or UI drift.
- Home top:
  - preserve current Workspace Overview.
- Home lower workbench section limits:
  - Today: 8 items
  - My Work: 8 items
  - Recent Channels: 6 channels
  - AI Suggestions: 5 suggestions
  - Activity: 10 events
- Home section priority for duplicate candidates:
  - Today
  - My Work
  - Activity
- Home deterministic ownership:
  - the same logical item ID must not appear in more than one of Today, My Work, and Activity
  - if an item qualifies for multiple sections, the backend assigns it to the highest-priority section only
- WorkspaceView types:
  - `list`
  - `calendar`
  - `search`
  - `report`
  - `form`
  - `channel_messages`
- WorkspaceView APIs:
  - `GET /api/v1/workspace/views`
  - `POST /api/v1/workspace/views`
  - `GET /api/v1/workspace/views/:id`
  - `PATCH /api/v1/workspace/views/:id`
- WorkspaceView first-release action types:
  - `open`
  - `summarize`
  - `update_filters`
  - `open_channel`
- WorkspaceView compatibility:
  - `source` is a persisted/queryable field
  - web clients must tolerate unknown future `view_type` values by rendering a generic registry entry instead of failing
- Actions are metadata only in Phase 72; they are not executed from Home.

## Task 0: Codex Contract Preflight

**Owner:** Codex  
**Files:**
- Reference: `docs/superpowers/specs/2026-04-26-phase72-super-app-home-and-unified-workspace-views-design.md`
- Reference: this plan

- [ ] **Step 1: Confirm frozen contract before implementation dispatch**

Verify the implementation contract includes:
- Home section limits and deterministic ownership
- no duplicate logical item ID across Today, My Work, and Activity
- WorkspaceView `source` persistence/querying
- unknown future `view_type` web fallback behavior
- Home partial rendering when WorkspaceView APIs fail

- [ ] **Step 2: Resolve worker questions by plan/spec patch**

Requirement:
- do not let Gemini or Windsurf change API paths, response ownership rules, or app-surface scope informally

## Task 1: Backend Home Workbench Aggregation

**Owner:** Gemini  
**Files:**
- Modify: `apps/api/internal/handlers/home.go`
- Create/Modify: `apps/api/internal/handlers/phase72_home_test.go`

- [ ] **Step 1: Write the failing Home aggregation tests**

Add tests covering:
- existing Workspace Overview fields remain present
- Home response includes:
  - `today.items`
  - `my_work.items`
  - `recent_channels.items`
  - `ai_suggestions.items`
  - `activity.items`
  - `apps_tools.items`
- backend enforces section limits
- duplicate candidate priority is Today > My Work > Activity
- the same logical item ID never appears in more than one of Today, My Work, and Activity
- section items include routeable target information where available

- [ ] **Step 2: Run the focused Home tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase72Home -count=1
```

Expected:
- FAIL

- [ ] **Step 3: Implement Home workbench aggregation**

Requirements:
- extend existing Home response shape without breaking existing consumers
- use real available data sources
- backend ranks section items
- backend owns dedupe and section assignment before returning the response
- no static placeholder data

- [ ] **Step 4: Re-run focused Home tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers
git commit -m "feat(api): add phase 72 home workbench aggregation"
```

## Task 2: Backend WorkspaceView Registry

**Owner:** Gemini  
**Files:**
- Create: `apps/api/internal/handlers/workspace_views.go`
- Modify: `apps/api/internal/domain/*`
- Modify: `apps/api/main.go`
- Create/Modify: `apps/api/internal/handlers/phase72_workspace_views_test.go`

- [ ] **Step 1: Write the failing WorkspaceView tests**

Add tests covering:
- create valid WorkspaceView
- `source` is persisted when creating/updating WorkspaceView entries
- list with `view_type`, `primary_channel_id`, `source`, `limit`, `cursor`
- detail by ID
- update title/filters/actions/channel
- reject invalid `view_type`
- reject empty title
- enforce default limit 20 and max limit 50
- `filters` must be shallow JSON object
- `actions` must be shallow descriptors and not executed

- [ ] **Step 2: Run the focused WorkspaceView tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase72WorkspaceViews -count=1
```

Expected:
- FAIL

- [ ] **Step 3: Implement WorkspaceView registry**

Requirements:
- shallow model only
- CRUD/query APIs from frozen contract
- persist and query `source`
- preserve unknown action types without executing them
- validate allowed view types

- [ ] **Step 4: Re-run focused WorkspaceView tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers/workspace_views.go apps/api/internal/domain apps/api/main.go
git commit -m "feat(api): add phase 72 workspace view registry"
```

## Task 3: Web Types And Store Wiring

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/stores/workspace-store.ts`
- Modify: `apps/web/types/index.ts`
- Create/Modify: `apps/web/lib/workspace-views.ts`

- [ ] **Step 1: Write the failing Web type/store tests**

Add tests covering:
- Home response can parse workbench sections
- WorkspaceView types support frozen view types
- unknown future WorkspaceView `view_type` values render/fall back as generic entries instead of crashing
- Apps & Tools entries are metadata/navigation only

- [ ] **Step 2: Run focused Web type/store tests**

Run the narrowest test command for workspace-store/types.

Expected:
- FAIL

- [ ] **Step 3: Implement Web types/store support**

Requirements:
- add typed Home workbench sections
- add WorkspaceView types/helpers
- model known view types as a frozen set while preserving unknown strings for forward compatibility
- keep backwards compatibility with existing Home fields

- [ ] **Step 4: Re-run focused tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/stores/workspace-store.ts apps/web/types/index.ts apps/web/lib/workspace-views.ts
git commit -m "feat(web): add phase 72 home and workspace view types"
```

## Task 4: Web Home Workbench Layout

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/components/layout/home-dashboard.tsx`
- Create: `apps/web/components/layout/home-workbench.tsx`
- Create: `apps/web/components/layout/home-today-section.tsx`
- Create: `apps/web/components/layout/home-my-work-section.tsx`

- [ ] **Step 1: Write the failing Home layout tests**

Add tests covering:
- Workspace Overview remains first
- lower Home renders workbench sections
- Today and My Work are visually primary
- no static placeholder content is used when API data exists
- Home still renders Today/My Work/Activity if WorkspaceView or Apps & Tools loading fails
- empty states are compact and not broken

- [ ] **Step 2: Run focused Home layout tests**

Run the narrowest Home component tests.

Expected:
- FAIL

- [ ] **Step 3: Implement Home workbench layout**

Requirements:
- preserve existing top overview
- replace lower dashboard area
- render Today and My Work as primary sections
- isolate WorkspaceView/Apps & Tools failures from core Home sections
- use restrained Slack-like workbench density

- [ ] **Step 4: Re-run focused tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/layout/home-dashboard.tsx apps/web/components/layout/home-workbench.tsx apps/web/components/layout/home-today-section.tsx apps/web/components/layout/home-my-work-section.tsx
git commit -m "feat(web): rebuild phase 72 home workbench"
```

## Task 5: Web Secondary Sections And Apps & Tools

**Owner:** Windsurf  
**Files:**
- Create: `apps/web/components/layout/home-apps-tools-section.tsx`
- Modify: Home workbench components
- Modify: routing/link helpers if needed

- [ ] **Step 1: Write the failing secondary-section tests**

Add tests covering:
- Recent Channels renders backend-ranked channels
- AI Suggestions renders only source-backed suggestions
- Activity renders backend activity summary
- Apps & Tools shows:
  - Lists
  - Calendar
  - Search
  - Reports
  - Forms
  - Workflows
  - Files
  - Tools
- Apps & Tools entries are navigation/registry entry points, not full products

- [ ] **Step 2: Run focused secondary-section tests**

Expected:
- FAIL

- [ ] **Step 3: Implement secondary Home sections**

Requirements:
- keep sections compact
- route to existing pages where possible
- do not invent full Calendar/Reports/Forms product UIs

- [ ] **Step 4: Re-run focused tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/layout apps/web
git commit -m "feat(web): add phase 72 home apps tools sections"
```

## Task 6: Web Regression And Existing Surface Safety

**Owner:** Windsurf  
**Files:**
- Existing route/page files only if needed for compatibility

- [ ] **Step 1: Write or run focused route regression checks**

Cover:
- Channel route still loads
- Canvas route still loads
- Activity route still loads
- DM route still loads
- Files route still loads
- Workflows route still loads

- [ ] **Step 2: Fix any Home-caused regressions**

Requirements:
- no unrelated redesigns
- no workspace shell restructure

- [ ] **Step 3: Run static verification**

Run:
```bash
cd apps/web && pnpm exec tsc --noEmit
```

Expected:
- PASS

- [ ] **Step 4: Commit**

```bash
git add apps/web
git commit -m "fix(web): harden phase 72 workspace route compatibility"
```

## Task 7: Codex Contract Audit, Collaboration Sync, And Release Handoff

**Owner:** Codex  
**Files:**
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/releases/v0.6.xx.md`
- Reference: `docs/superpowers/specs/2026-04-26-phase72-super-app-home-and-unified-workspace-views-design.md`

- [ ] **Step 1: Review Gemini backend diff against frozen Phase 72 contract**

Verify:
- Home preserves existing overview fields
- section limits and priority rules are backend-driven
- no logical item ID is duplicated across Today, My Work, and Activity
- WorkspaceView APIs match frozen paths/query params
- WorkspaceView `source` persists and filters correctly
- actions are metadata-only

- [ ] **Step 2: Review Windsurf Web diff against frozen Phase 72 contract**

Verify:
- Workspace Overview stays top
- Today/My Work dominate lower Home
- unknown future WorkspaceView `view_type` values degrade to generic entries
- WorkspaceView/App & Tools errors do not blank the core Home workbench
- Apps & Tools does not pretend Calendar/Reports/Forms are complete products
- existing routes are not broken

- [ ] **Step 3: Run focused verification commands**

Backend:
```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase72Home|TestPhase72WorkspaceViews' -count=1
```

Web:
- run focused checks covering:
  - Home workbench rendering
  - Today/My Work sections
  - Apps & Tools entries
  - WorkspaceView entry rendering
  - route compatibility

Static:
```bash
cd apps/web && pnpm exec tsc --noEmit
```

- [ ] **Step 4: Update collaboration and release docs**

Update:
- `docs/AGENT-COLLAB.md` with completion notes and next-phase hooks
- `docs/releases/v0.6.xx.md` with concise user-facing release notes

- [ ] **Step 5: Commit**

```bash
git add docs/AGENT-COLLAB.md docs/releases
git commit -m "docs: close out phase 72 super app home workspace views"
```

## Parallelization Guidance

Safe parallelism:

- Codex Task 0 must happen before Gemini/Windsurf implementation dispatch.
- Gemini Task 1 and Task 2 should run sequentially if they need shared domain model or serializer changes; parallelize only if the write sets are explicitly disjoint.
- Windsurf Task 3 can begin once Gemini response shapes are frozen, even before all backend handlers are final.
- Windsurf Task 4 depends on Task 3.
- Windsurf Task 5 can proceed after Task 4 establishes the workbench composition.
- Windsurf Task 6 is the final Web hardening pass.
- Codex Task 7 happens after Gemini and Windsurf complete their slices.

## Verification Checklist

- [ ] Workspace Overview remains at top of Home
- [ ] Today and My Work are primary below-overview sections
- [ ] Home sections enforce limits and deterministic dedupe
- [ ] No logical item ID appears in more than one of Today, My Work, and Activity
- [ ] Home uses real API data
- [ ] WorkspaceView registry supports frozen view types
- [ ] WorkspaceView `source` persists and filters correctly
- [ ] Web tolerates unknown future WorkspaceView `view_type` values
- [ ] Home core workbench sections survive WorkspaceView/Apps & Tools load failure
- [ ] WorkspaceView actions are metadata-only in Phase 72
- [ ] Apps & Tools exposes expected entries without full product overreach
- [ ] Existing core workspace routes still work

## Non-Goals Reminder

Do not add in this phase:

- full Super App runtime
- app builder
- complex permission model
- full Calendar/Reports/Forms products
- BI engine
- workflow engine rewrite
