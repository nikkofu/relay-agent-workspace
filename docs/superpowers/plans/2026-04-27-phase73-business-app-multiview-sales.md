# Phase 73 Business App Multiview Sales Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a reusable business-app multiview surface and ship the first App Hub instance: Sales App with Sales Orders shown as search, list, calendar, kanban, and stats modes.

**Architecture:** Gemini adds a narrow backend business-app contract with deterministic Sales Order demo data, filtering, cursor pagination, and stats. Windsurf builds one reusable Web shell and Sales App page that switches display modes over the same data source. Codex owns contract review, sequencing, and release closeout.

**Tech Stack:** Go/Gin API, GORM/SQLite, Next.js 16, React 19, TypeScript, Zustand, existing workspace shell, existing Home Apps & Tools, existing primary navigation.

---

## Contract Freeze

- Phase 73 is not a full CRM, ERP, app builder, or BI runtime.
- The first app is `sales`.
- The first entity is `sales_order`.
- The reusable page supports these modes:
  - `search`
  - `list`
  - `calendar`
  - `kanban`
  - `stats`
- Every mode renders the same Sales Order record contract.
- Calendar, Kanban, and Stats are display modes only.
- No drag-to-update, no order mutation workflow, and no autonomous agent execution unless a later phase freezes a write contract.

## File Structure

### Backend files likely involved

- Create: `apps/api/internal/handlers/business_apps.go`
- Create/Modify: `apps/api/internal/handlers/phase73_business_apps_test.go`
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Modify: `apps/api/main.go`

### Web files likely involved

- Create: `apps/web/app/workspace/apps/sales/page.tsx`
- Create: `apps/web/components/apps/business-app-shell.tsx`
- Create: `apps/web/components/apps/business-view-switcher.tsx`
- Create: `apps/web/components/apps/sales/sales-order-list-view.tsx`
- Create: `apps/web/components/apps/sales/sales-order-calendar-view.tsx`
- Create: `apps/web/components/apps/sales/sales-order-kanban-view.tsx`
- Create: `apps/web/components/apps/sales/sales-order-stats-view.tsx`
- Create: `apps/web/lib/business-apps.ts`
- Create/Modify: `apps/web/stores/business-app-store.ts`
- Modify: `apps/web/types/index.ts`
- Modify: `apps/web/components/layout/primary-nav.tsx`
- Modify: `apps/web/components/layout/home-apps-tools-section.tsx`
- Modify: `apps/web/lib/workspace-views.ts`

## Task 0: Codex Contract Preflight

**Owner:** Codex

- [ ] **Step 1: Confirm Phase 73 boundary**

Verify:
- Sales App is one app instance, not a generic app builder.
- Sales Orders are read/search/filter/display only.
- AI-native actions are metadata-only.
- App Hub link and Sales App route are required.

- [ ] **Step 2: Answer worker questions by spec/plan patch**

Requirement:
- Gemini and Windsurf must not invent incompatible app contracts independently.

## Task 1: Backend Sales App Data Contract

**Owner:** Gemini  
**Files:**
- Create: `apps/api/internal/handlers/business_apps.go`
- Create/Modify: `apps/api/internal/handlers/phase73_business_apps_test.go`
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Modify: `apps/api/main.go`

- [ ] **Step 1: Write failing backend contract tests**

Cover:
- `GET /api/v1/apps` includes `sales`
- `GET /api/v1/apps/sales` returns app metadata and available modes
- `GET /api/v1/apps/sales/data?mode=list` returns Sales Order records and schema
- `GET /api/v1/apps/sales/data?mode=search&q=...` filters by customer/order/summary/tags
- `GET /api/v1/apps/sales/data?mode=calendar` returns date bucket groups from `expected_close_date`
- `GET /api/v1/apps/sales/data?mode=kanban` returns stage groups
- `GET /api/v1/apps/sales/stats` returns aggregate totals
- invalid app key returns 404
- invalid mode returns 400
- `limit` defaults to 20 and maxes at 50
- `next_cursor` is returned when more rows exist

- [ ] **Step 2: Run tests and verify failure**

```bash
cd apps/api && go test ./internal/handlers -run TestPhase73BusinessApps -count=1
```

Expected:
- FAIL

- [ ] **Step 3: Implement minimal backend**

Requirements:
- add `SalesOrder` model or equivalent deterministic read model
- seed deterministic demo orders if no rows exist
- implement read-only app metadata/data/stats endpoints
- support filters: `q`, `stage`, `status`, `owner_user_id`, `date_from`, `date_to`, `limit`, `cursor`
- return stable schema metadata for Sales Orders
- include routeable `source_channel_id` and `source_message_id` fields where available
- keep AI action descriptors metadata-only

- [ ] **Step 4: Re-run backend tests**

```bash
cd apps/api && go test ./internal/handlers -run TestPhase73BusinessApps -count=1
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers/business_apps.go apps/api/internal/handlers/phase73_business_apps_test.go apps/api/internal/domain/models.go apps/api/internal/db/db.go apps/api/main.go
git commit -m "feat(api): add sales business app data contract"
```

## Task 2: Web Business App Store And Types

**Owner:** Windsurf  
**Files:**
- Create: `apps/web/lib/business-apps.ts`
- Create/Modify: `apps/web/stores/business-app-store.ts`
- Modify: `apps/web/types/index.ts`

- [ ] **Step 1: Write failing type/store tests or narrow checks**

Cover:
- app metadata normalization
- Sales Order record normalization
- mode/query/filter params preserved
- stats payload typed
- unknown fields do not crash rendering helpers

- [ ] **Step 2: Run focused Web checks**

Use the narrowest available command for Web tests. If no test harness exists, run:

```bash
cd apps/web && pnpm exec tsc --noEmit
```

Expected before implementation:
- FAIL if tests were added, or type errors until code is complete

- [ ] **Step 3: Implement store/types**

Requirements:
- `BusinessApp`
- `BusinessAppMode`
- `SalesOrder`
- `BusinessAppDataResponse`
- `BusinessAppStatsResponse`
- fetch methods for app metadata, data, and stats
- shared query builder for route/filter state

- [ ] **Step 4: Re-run focused checks**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/lib/business-apps.ts apps/web/stores/business-app-store.ts apps/web/types/index.ts
git commit -m "feat(web): add business app store and sales types"
```

## Task 3: Web Sales App Multiview Page

**Owner:** Windsurf  
**Files:**
- Create: `apps/web/app/workspace/apps/sales/page.tsx`
- Create: `apps/web/components/apps/business-app-shell.tsx`
- Create: `apps/web/components/apps/business-view-switcher.tsx`
- Create: `apps/web/components/apps/sales/sales-order-list-view.tsx`
- Create: `apps/web/components/apps/sales/sales-order-calendar-view.tsx`
- Create: `apps/web/components/apps/sales/sales-order-kanban-view.tsx`
- Create: `apps/web/components/apps/sales/sales-order-stats-view.tsx`

- [ ] **Step 1: Write failing rendering/mode tests where available**

Cover:
- default route loads list mode
- switching mode updates `?mode=...`
- search mode preserves `q`
- list/calendar/kanban/stats render the same Sales Order fields with mode-specific emphasis
- empty/loading/error states are usable

- [ ] **Step 2: Run focused Web checks**

```bash
cd apps/web && pnpm exec tsc --noEmit
```

Expected:
- FAIL until components/types are complete

- [ ] **Step 3: Implement reusable shell and Sales views**

Requirements:
- Slack-like compact workbench styling
- header with Sales App identity and mode switcher
- shared search/filter bar
- list view: table-like rows
- calendar view: expected-close date buckets
- kanban view: stage columns
- stats view: aggregate cards and simple charts/lists
- action chips are metadata-only unless existing routeable links are available

- [ ] **Step 4: Re-run checks**

```bash
cd apps/web && pnpm exec tsc --noEmit
cd apps/web && pnpm lint
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/app/workspace/apps/sales apps/web/components/apps
git commit -m "feat(web): add sales app multiview page"
```

## Task 4: Web App Hub Navigation

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/components/layout/primary-nav.tsx`
- Modify: `apps/web/components/layout/home-apps-tools-section.tsx`
- Modify: `apps/web/lib/workspace-views.ts`

- [ ] **Step 1: Write or run navigation checks**

Cover:
- primary navigation exposes App Hub or Apps
- Sales App is reachable from App Hub/Home Apps & Tools
- existing Files/Workflows/Search links still work
- unknown WorkspaceView route fallback remains safe

- [ ] **Step 2: Implement navigation**

Requirements:
- add a clear menu entry for App Hub or Sales App
- ensure Home Apps & Tools can open `/workspace/apps/sales`
- do not remove existing Slack-like nav items

- [ ] **Step 3: Verify**

```bash
cd apps/web && pnpm exec tsc --noEmit
cd apps/web && pnpm lint
```

Expected:
- PASS

- [ ] **Step 4: Commit**

```bash
git add apps/web/components/layout/primary-nav.tsx apps/web/components/layout/home-apps-tools-section.tsx apps/web/lib/workspace-views.ts
git commit -m "feat(web): add sales app navigation"
```

## Task 5: Codex Contract Audit And Release Closeout

**Owner:** Codex  
**Files:**
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `CHANGELOG.md`
- Create/Modify: `docs/releases/v0.6.xx.md`

- [ ] **Step 1: Review Gemini backend diff**

Verify:
- routes match frozen contract
- records are stable across modes
- search/list/calendar/kanban/stats use the same Sales Order fields
- stats are backend-computed
- pagination and invalid-mode behavior are tested

- [ ] **Step 2: Review Windsurf Web diff**

Verify:
- route `/workspace/apps/sales` exists
- menu/App Hub link exists
- mode switch is URL-backed
- all modes render Sales Orders from one store/data contract
- UI does not pretend to be a full CRM/ERP

- [ ] **Step 3: Run release verification**

```bash
cd apps/api && go test ./internal/handlers -run TestPhase73BusinessApps -count=1
cd apps/web && pnpm exec tsc --noEmit
cd apps/web && pnpm lint
```

- [ ] **Step 4: Update docs and release notes**

Document completion and next hooks:
- write actions for Sales Orders
- App Hub generic registry
- agent-assisted order summarization/follow-up
- channel-linked sales activity feed

- [ ] **Step 5: Commit and tag**

```bash
git add docs/AGENT-COLLAB.md CHANGELOG.md docs/releases
git commit -m "docs: close phase 73 sales app multiview"
git tag v0.6.xx
```

## Parallelization Guidance

- Gemini Task 1 is the backend foundation and should complete before Windsurf locks Web data types.
- Windsurf Task 2 can start once Gemini freezes response examples.
- Windsurf Task 3 depends on Task 2.
- Windsurf Task 4 can proceed in parallel with Task 3 if it only changes navigation and route links.
- Codex Task 5 happens after Gemini/Windsurf implementation slices complete.

## Verification Checklist

- [ ] Sales App link is visible in workspace navigation.
- [ ] `/workspace/apps/sales` loads.
- [ ] Sales Orders render in search, list, calendar, kanban, and stats modes.
- [ ] All modes use the same backend data contract.
- [ ] URL query params preserve mode/search/filter state.
- [ ] Backend filters and cursor pagination work.
- [ ] Stats are computed from Sales Orders.
- [ ] AI action metadata is visible but not autonomously executed.
- [ ] Existing Slack-like surfaces remain available: Files, Canvas, Lists, Workflows, Tools, Activity, DMs, Home, Users, Channels, Messages, Presence, Status, Profiles.

