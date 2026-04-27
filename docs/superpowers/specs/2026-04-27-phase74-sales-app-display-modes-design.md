# Phase 74 Sales App Display Modes Design

## Product Definition

Sales App should become the reference implementation for future business-app surfaces. The same Sales Orders data source must render through five display styles: **List**, **Card Grid**, **Kanban**, **Calendar**, and **Stat**. Every display style shares the same search and filter area at the top, and only the content region below changes.

Phase 74 refines Phase 73. It does not build a full CRM, app builder, charting platform, or autonomous sales workflow engine.

## Goals

- Replace Phase 73 display modes with:
  - `list`
  - `card_grid`
  - `kanban`
  - `calendar`
  - `stat`
- Keep one shared search/filter bar above every mode.
- Add Card Grid for dense business cards.
- Upgrade Calendar from date buckets to a business-time event model.
- Support Calendar subviews:
  - `day`
  - `week`
  - `month`
- Upgrade Stat to support multiple chart styles:
  - `summary`
  - `bar`
  - `funnel`
  - `timeline`
- Preserve one Sales Orders data contract across all modes.
- Keep AI-native actions metadata-only.

## Non-Goals

- Do not add write/mutation flows for Sales Orders.
- Do not add drag-and-drop Kanban stage updates.
- Do not add a full calendar scheduling product.
- Do not add a full charting/BI engine.
- Do not require `react-big-calendar` as a dependency in this phase.
- Do not remove existing Slack-like workspace routes or navigation.

## Mode Contract

Allowed `mode` values:

- `list`
- `card_grid`
- `kanban`
- `calendar`
- `stat`

Deprecated aliases:

- `search` should be treated as `list` with `q` populated.
- `stats` should be treated as `stat`.

The URL remains the source of display state:

- `?mode=list&q=acme`
- `?mode=card_grid&stage=proposal`
- `?mode=calendar&calendar_view=week&calendar_time_field=expected_close_date`
- `?mode=stat&chart_style=funnel`

## Shared Search And Filter Area

All modes must render the same top control region:

- search input
- search submit/clear
- stage filter
- status filter
- date range filter
- mode switcher

Search is not a separate display mode in Phase 74. It is a filter state that works above every mode.

## Calendar Business-Time Contract

Calendar mode converts business records into events.

Supported first-release time fields for Sales Orders:

- `expected_close_date`
- `order_date`
- `due_date`
- `last_activity_at`

Query params:

- `calendar_view=day|week|month`
- `calendar_time_field=expected_close_date|order_date|due_date|last_activity_at`
- `date_from`
- `date_to`

Backend response should include `calendar_events` when `mode=calendar`:

```json
{
  "calendar_events": [
    {
      "id": "sales-order:so-1:expected_close_date",
      "record_id": "so-1",
      "title": "SO-2026-001 · TechFlow Systems",
      "start": "2026-05-12T00:00:00Z",
      "end": "2026-05-12T23:59:59Z",
      "time_field": "expected_close_date",
      "amount": 12500,
      "stage": "negotiation",
      "status": "open",
      "record": {}
    }
  ]
}
```

The Web should implement a lightweight calendar shell inspired by `react-big-calendar` concepts:

- event array
- `start` / `end` accessors
- day/week/month subview switching
- visible date window derived from selected subview

Direct dependency on `react-big-calendar` is optional and should not be introduced unless Windsurf verifies it fits Next.js 16, React 19, bundle size, and styling constraints.

## Stat Chart Contract

Stat mode supports chart-style switching through:

- `chart_style=summary|bar|funnel|timeline`

Backend response should include enough aggregate data for all chart styles:

- summary totals
- amount/count by stage
- funnel ordered by sales stage
- expected close timeline buckets
- at-risk/won/lost/open counts

Charts can be CSS/SVG/lightweight components in Phase 74. Do not introduce a heavy charting dependency unless necessary.

## Backend Responsibilities

Gemini owns:

- mode alias normalization
- `card_grid` and `stat` support in app metadata
- Calendar event projection from selected business time field
- Calendar subview/date-window compatible filtering
- Stat aggregates for `summary`, `bar`, `funnel`, and `timeline`
- tests for mode aliases, calendar events, invalid calendar fields, invalid chart styles, and shared search/filter semantics

Recommended focused test:

```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase73BusinessApps|TestPhase74SalesDisplayModes' -count=1
```

## Web Responsibilities

Windsurf owns:

- update mode switcher to List/Card Grid/Kanban/Calendar/Stat
- keep shared search/filter controls above every display
- add Sales Order Card Grid component
- upgrade Calendar component with day/week/month subviews and business time field selector
- upgrade Stat component with chart-style selector and multiple chart renderers
- keep route/query state reload-safe
- preserve App Hub and Sales App navigation

Recommended checks:

```bash
cd apps/web && pnpm exec tsc --noEmit
cd apps/web && pnpm lint
```

## Acceptance Criteria

- Sales App no longer shows `Search` as a mode.
- Sales App no longer shows `Stats`; it shows `Stat`.
- Search works above every mode.
- Card Grid mode renders Sales Orders as business cards.
- Calendar mode supports day/week/month subviews.
- Calendar mode can switch between supported business time fields.
- Calendar content is derived from Sales Order business timestamps.
- Stat mode supports at least summary, bar, funnel, and timeline chart styles.
- Existing Sales App route and App Hub links remain stable.
- Existing Slack-like workspace surfaces remain reachable.

