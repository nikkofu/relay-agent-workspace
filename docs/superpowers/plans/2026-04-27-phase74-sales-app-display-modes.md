# Phase 74 Sales App Display Modes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refine Sales App into the reusable business-app reference surface with shared search and five display styles: List, Card Grid, Kanban, Calendar, and Stat.

**Architecture:** Gemini updates the app data contract so search is a shared filter, not a mode, and adds Calendar event + Stat chart aggregate projections. Windsurf updates the Sales App shell and views so all modes share one search/filter region while the content area switches between List/Card Grid/Kanban/Calendar/Stat. Codex owns contract review and release closeout.

**Tech Stack:** Go/Gin API, GORM/SQLite, Next.js 16, React 19, TypeScript, Zustand, existing Sales App route, existing business app store/helpers, lightweight CSS/SVG chart/calendar rendering.

---

## Contract Freeze

- Valid display modes:
  - `list`
  - `card_grid`
  - `kanban`
  - `calendar`
  - `stat`
- Deprecated mode aliases:
  - `search` => `list` with `q`
  - `stats` => `stat`
- Search is a shared filter above every mode.
- Calendar subviews:
  - `day`
  - `week`
  - `month`
- Calendar business time fields:
  - `expected_close_date`
  - `order_date`
  - `due_date`
  - `last_activity_at`
- Stat chart styles:
  - `summary`
  - `bar`
  - `funnel`
  - `timeline`
- Do not introduce heavy calendar/chart dependencies unless necessary. A lightweight implementation is preferred for Phase 74.

## Task 0: Codex Contract Preflight

**Owner:** Codex

- [ ] **Step 1: Confirm Phase 74 replaces Phase 73 display labels**

Verify:
- `search` is removed from the mode switcher.
- `stats` is renamed/replaced by `stat`.
- all modes retain the same search/filter controls.

- [ ] **Step 2: Keep worker questions inside this contract**

Requirement:
- Gemini and Windsurf must not add Sales mutations, drag/drop status updates, or full BI/calendar runtime behavior in Phase 74.

## Task 1: Backend Mode And Metadata Contract

**Owner:** Gemini  
**Files:**
- Modify: `apps/api/internal/handlers/business_apps.go`
- Modify: `apps/api/internal/handlers/phase73_business_apps_test.go`
- Create/Modify: `apps/api/internal/handlers/phase74_sales_display_modes_test.go`

- [ ] **Step 1: Write failing tests**

Cover:
- `GET /api/v1/apps/sales` returns modes `list`, `card_grid`, `kanban`, `calendar`, `stat`
- `search` query mode aliases to `list`
- `stats` query mode aliases to `stat`
- invalid mode still returns 400
- shared `q`, `stage`, `status`, `date_from`, `date_to` filters work for all valid modes

- [ ] **Step 2: Run failing tests**

```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase73BusinessApps|TestPhase74SalesDisplayModes' -count=1
```

Expected:
- FAIL before implementation

- [ ] **Step 3: Implement mode contract**

Requirements:
- update app metadata modes
- normalize aliases before validation
- keep response shape backwards tolerant
- do not add write endpoints

- [ ] **Step 4: Re-run tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers/business_apps.go apps/api/internal/handlers/phase73_business_apps_test.go apps/api/internal/handlers/phase74_sales_display_modes_test.go
git commit -m "feat(api): update sales app display mode contract"
```

## Task 2: Backend Calendar Event Projection

**Owner:** Gemini  
**Files:**
- Modify: `apps/api/internal/handlers/business_apps.go`
- Modify: `apps/api/internal/handlers/phase74_sales_display_modes_test.go`

- [ ] **Step 1: Write failing calendar projection tests**

Cover:
- `mode=calendar&calendar_time_field=expected_close_date` returns `calendar_events`
- event includes `id`, `record_id`, `title`, `start`, `end`, `time_field`, `stage`, `status`, `amount`, and `record`
- `calendar_time_field=order_date|due_date|last_activity_at` works
- invalid `calendar_time_field` returns 400
- `calendar_view=day|week|month` accepted
- invalid `calendar_view` returns 400
- date range filters affect both records and events

- [ ] **Step 2: Run failing tests**

```bash
cd apps/api && go test ./internal/handlers -run TestPhase74SalesCalendar -count=1
```

Expected:
- FAIL before implementation

- [ ] **Step 3: Implement calendar event projection**

Requirements:
- project each Sales Order into an event using the selected business time field
- default `calendar_time_field=expected_close_date`
- default `calendar_view=month`
- all-day event behavior is acceptable for date-only fields
- omit records missing the selected time field from `calendar_events`, but keep them in regular records if filters allow

- [ ] **Step 4: Re-run tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers/business_apps.go apps/api/internal/handlers/phase74_sales_display_modes_test.go
git commit -m "feat(api): add sales calendar event projection"
```

## Task 3: Backend Stat Chart Aggregates

**Owner:** Gemini  
**Files:**
- Modify: `apps/api/internal/handlers/business_apps.go`
- Modify: `apps/api/internal/handlers/phase74_sales_display_modes_test.go`

- [ ] **Step 1: Write failing stat tests**

Cover:
- `GET /api/v1/apps/sales/stats?chart_style=summary`
- `chart_style=bar`
- `chart_style=funnel`
- `chart_style=timeline`
- invalid `chart_style` returns 400
- stats response includes enough data for all styles:
  - totals
  - `by_stage`
  - `funnel`
  - `timeline_buckets`

- [ ] **Step 2: Run failing tests**

```bash
cd apps/api && go test ./internal/handlers -run TestPhase74SalesStatsCharts -count=1
```

Expected:
- FAIL before implementation

- [ ] **Step 3: Implement stat aggregate contract**

Requirements:
- support chart style query param
- compute stage counts/sums
- compute ordered funnel using sales stage order
- compute expected-close timeline buckets from Sales Orders
- preserve existing stats fields consumed by `v0.6.70`

- [ ] **Step 4: Re-run tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers/business_apps.go apps/api/internal/handlers/phase74_sales_display_modes_test.go
git commit -m "feat(api): add sales stat chart aggregates"
```

## Task 4: Web Types, Store, And Mode Helpers

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/lib/business-apps.ts`
- Modify: `apps/web/stores/business-app-store.ts`
- Modify: `apps/web/types/index.ts`

- [ ] **Step 1: Write failing type/helper checks**

Cover:
- mode set is `list|card_grid|kanban|calendar|stat`
- `search` aliases to `list`
- `stats` aliases to `stat`
- query builder preserves `calendar_view`, `calendar_time_field`, and `chart_style`
- calendar event normalization
- stats chart aggregate normalization

- [ ] **Step 2: Run focused checks**

```bash
cd apps/web && pnpm exec tsc --noEmit
```

Expected:
- FAIL until helpers/types are updated

- [ ] **Step 3: Implement helpers/types**

Requirements:
- update `BusinessAppMode`
- add `BusinessCalendarView`
- add `BusinessCalendarTimeField`
- add `BusinessChartStyle`
- add `BusinessCalendarEvent`
- keep old URL aliases tolerant

- [ ] **Step 4: Re-run checks**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/lib/business-apps.ts apps/web/stores/business-app-store.ts apps/web/types/index.ts
git commit -m "feat(web): update sales app display mode types"
```

## Task 5: Web Shared Search And Card Grid

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/app/workspace/apps/sales/page.tsx`
- Modify: `apps/web/components/apps/business-app-shell.tsx`
- Modify: `apps/web/components/apps/business-view-switcher.tsx`
- Create: `apps/web/components/apps/sales/sales-order-card-grid-view.tsx`
- Modify: `apps/web/components/apps/sales/sales-order-list-view.tsx`

- [ ] **Step 1: Write or run rendering checks**

Cover:
- shared search/filter bar renders for List/Card Grid/Kanban/Calendar/Stat
- mode switcher no longer shows Search
- mode switcher shows Card Grid and Stat
- search submit does not force mode to `search`
- Card Grid renders Sales Order cards

- [ ] **Step 2: Implement Web mode shell changes**

Requirements:
- route `search` alias to `list`
- route `stats` alias to `stat`
- add Card Grid component
- keep search/filter controls visually above content
- do not remove existing App Hub navigation

- [ ] **Step 3: Verify**

```bash
cd apps/web && pnpm exec tsc --noEmit
cd apps/web && pnpm lint
```

Expected:
- PASS

- [ ] **Step 4: Commit**

```bash
git add apps/web/app/workspace/apps/sales/page.tsx apps/web/components/apps/business-app-shell.tsx apps/web/components/apps/business-view-switcher.tsx apps/web/components/apps/sales
git commit -m "feat(web): add sales card grid and shared search modes"
```

## Task 6: Web Calendar Day Week Month

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/app/workspace/apps/sales/page.tsx`
- Modify: `apps/web/components/apps/sales/sales-order-calendar-view.tsx`
- Modify: `apps/web/lib/business-apps.ts`

- [ ] **Step 1: Write or run calendar checks**

Cover:
- `calendar_view=day|week|month` renders different layouts
- `calendar_time_field` selector changes event source
- invalid/missing query params normalize safely
- event cards use `start`/`end` from backend calendar events

- [ ] **Step 2: Implement lightweight calendar shell**

Requirements:
- no mandatory `react-big-calendar` dependency in Phase 74
- day view: single-day agenda
- week view: seven-day columns
- month view: month grid
- business time field selector is visible in Calendar mode
- date window controls are route-backed if implemented

- [ ] **Step 3: Verify**

```bash
cd apps/web && pnpm exec tsc --noEmit
cd apps/web && pnpm lint
```

Expected:
- PASS

- [ ] **Step 4: Commit**

```bash
git add apps/web/app/workspace/apps/sales/page.tsx apps/web/components/apps/sales/sales-order-calendar-view.tsx apps/web/lib/business-apps.ts
git commit -m "feat(web): add sales calendar day week month views"
```

## Task 7: Web Stat Chart Styles

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/app/workspace/apps/sales/page.tsx`
- Modify: `apps/web/components/apps/sales/sales-order-stats-view.tsx`
- Modify: `apps/web/lib/business-apps.ts`

- [ ] **Step 1: Write or run stat checks**

Cover:
- `chart_style=summary|bar|funnel|timeline`
- selector updates URL
- each style renders from backend stats data
- empty stats state remains usable

- [ ] **Step 2: Implement chart style selector and renderers**

Requirements:
- summary: cards
- bar: stage amount/count bars
- funnel: ordered stage funnel
- timeline: expected-close buckets
- lightweight CSS/SVG only unless a dependency is explicitly approved later

- [ ] **Step 3: Verify**

```bash
cd apps/web && pnpm exec tsc --noEmit
cd apps/web && pnpm lint
```

Expected:
- PASS

- [ ] **Step 4: Commit**

```bash
git add apps/web/app/workspace/apps/sales/page.tsx apps/web/components/apps/sales/sales-order-stats-view.tsx apps/web/lib/business-apps.ts
git commit -m "feat(web): add sales stat chart style switcher"
```

## Task 8: Codex Contract Audit And Release Closeout

**Owner:** Codex  
**Files:**
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `CHANGELOG.md`
- Create/Modify: `docs/releases/v0.6.xx.md`

- [ ] **Step 1: Review Gemini backend diff**

Verify:
- modes match frozen values
- aliases are backwards tolerant
- calendar events use selected business time field
- invalid calendar fields/chart styles are rejected
- stats include all chart aggregate families

- [ ] **Step 2: Review Windsurf Web diff**

Verify:
- mode switcher shows List/Card Grid/Kanban/Calendar/Stat
- shared search/filter controls stay above all modes
- Calendar day/week/month works from business events
- Stat chart styles work from backend aggregates
- no full CRM/calendar/BI overreach

- [ ] **Step 3: Run final checks**

```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase73BusinessApps|TestPhase74' -count=1
cd apps/web && pnpm exec tsc --noEmit
cd apps/web && pnpm lint
```

- [ ] **Step 4: Update docs and release notes**

Document:
- final Phase 74 behavior
- Gemini/Windsurf completion notes
- next hooks for generic business app reuse and AI-native actions

- [ ] **Step 5: Commit and tag**

```bash
git add docs/AGENT-COLLAB.md CHANGELOG.md docs/releases
git commit -m "docs: close phase 74 sales display modes"
git tag v0.6.xx
```

## Parallelization Guidance

- Gemini Task 1 must complete before Windsurf locks Web mode types.
- Gemini Task 2 and Task 3 can run in parallel only if they do not edit overlapping helper functions in `business_apps.go`.
- Windsurf Task 4 depends on Gemini mode examples.
- Windsurf Task 5 can begin after Task 4.
- Windsurf Task 6 and Task 7 can run in parallel if their write sets stay separated.
- Codex Task 8 happens after Gemini/Windsurf complete their slices.

## Verification Checklist

- [ ] Sales App mode switcher shows List/Card Grid/Kanban/Calendar/Stat.
- [ ] Search/filter bar appears above every mode.
- [ ] Search works without changing mode to Search.
- [ ] Card Grid renders Sales Orders.
- [ ] Calendar supports day/week/month.
- [ ] Calendar can switch business time fields.
- [ ] Stat supports summary/bar/funnel/timeline.
- [ ] Backend rejects invalid calendar and chart params.
- [ ] Existing Sales App/App Hub routes remain stable.
- [ ] Existing Slack-like workspace surfaces remain reachable.

