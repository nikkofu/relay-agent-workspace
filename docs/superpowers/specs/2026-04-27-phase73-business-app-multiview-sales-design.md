# Phase 73 Business App Multiview Sales Design

## Product Definition

Relay should provide one reusable business-app page pattern that any future app can use. Phase 73 ships the first instance under App Hub: **Sales App**, backed by a standard Sales Orders data source that can be displayed as search results, list/table, calendar, kanban, and stats without duplicating data or inventing separate app-specific UIs.

This phase turns Phase 72's WorkspaceView registry into a practical business surface while staying narrow: one reusable view shell, one Sales Orders dataset, and metadata-driven display modes.

## Goals

- Add an App Hub entry point and a menu link to Sales App.
- Add a reusable business app page that supports mode switching:
  - Search
  - List/table
  - Calendar
  - Kanban
  - Stats
- Use the same Sales Orders records in every mode.
- Keep display mode state routeable via query params.
- Preserve Slack-like workspace feel: left nav, channel-aware links, compact workbench density, and clear activity affordances.
- Prepare for AI-native actions without executing autonomous workflows in this phase.

## Non-Goals

- Do not build a full CRM.
- Do not build a full ERP app runtime.
- Do not add custom app builders.
- Do not add permissions beyond existing workspace assumptions.
- Do not add write-heavy order lifecycle automation beyond read/search/filter display.
- Do not require Calendar/Reports/Kanban as separate product pages.

## Sales Order Standard Fields

Each Sales Order should expose stable fields that all display modes can use:

- `id`
- `order_number`
- `customer_name`
- `customer_id`
- `owner_user_id`
- `owner_name`
- `stage`
- `status`
- `priority`
- `amount`
- `currency`
- `probability`
- `expected_close_date`
- `order_date`
- `due_date`
- `last_activity_at`
- `source_channel_id`
- `source_message_id`
- `summary`
- `tags`
- `created_at`
- `updated_at`

Allowed first-release `stage` values:

- `lead`
- `qualified`
- `proposal`
- `negotiation`
- `closed_won`
- `closed_lost`

Allowed first-release `status` values:

- `open`
- `at_risk`
- `won`
- `lost`

## Display Mode Semantics

Search mode:

- Uses `q` across order number, customer name, owner name, summary, tags, and source message snippet if available.
- Shows ranked cards or rows with customer, amount, stage, expected close date, and channel/message source link.

List mode:

- Default mode.
- Shows a table-like responsive list with columns: order, customer, stage/status, owner, amount, expected close date, last activity.
- Supports backend filters and cursor pagination.

Calendar mode:

- Groups orders by `expected_close_date`.
- Displays date buckets, not a full scheduling product.
- Missing dates appear in an `Unscheduled` bucket.

Kanban mode:

- Groups orders by `stage`.
- Cards show customer, amount, probability, owner, due/close date, status badge, and source channel/message link.
- Drag-and-drop stage mutation is out of scope unless Gemini ships an explicit patch endpoint in a later phase.

Stats mode:

- Shows backend-computed aggregates:
  - total pipeline amount
  - weighted pipeline amount
  - open/won/lost counts
  - amount by stage
  - at-risk count
  - expected-close buckets
- Stats are read-only.

## Backend Contract

Gemini owns backend/API/test work.

Recommended endpoints:

- `GET /api/v1/apps`
- `GET /api/v1/apps/:app_key`
- `GET /api/v1/apps/:app_key/data`
- `GET /api/v1/apps/:app_key/stats`

For Phase 73, `app_key=sales`.

`GET /api/v1/apps/:app_key/data` query params:

- `mode`
- `q`
- `stage`
- `status`
- `owner_user_id`
- `date_from`
- `date_to`
- `limit`
- `cursor`

Response shape:

```json
{
  "app": {
    "key": "sales",
    "title": "Sales App",
    "primary_entity": "sales_order"
  },
  "view": {
    "mode": "list",
    "available_modes": ["search", "list", "calendar", "kanban", "stats"]
  },
  "schema": {
    "entity": "sales_order",
    "fields": []
  },
  "records": [],
  "groups": [],
  "next_cursor": ""
}
```

`GET /api/v1/apps/:app_key/stats` returns aggregate numbers using the same filter params where applicable.

The backend may seed demo Sales Orders. Demo data must be realistic and deterministic, not random per request.

## Web Contract

Windsurf owns Web/UI work.

Recommended route:

- `/workspace/apps/sales`

Recommended components:

- `apps/web/app/workspace/apps/sales/page.tsx`
- `apps/web/components/apps/business-app-shell.tsx`
- `apps/web/components/apps/business-view-switcher.tsx`
- `apps/web/components/apps/sales/sales-order-list-view.tsx`
- `apps/web/components/apps/sales/sales-order-calendar-view.tsx`
- `apps/web/components/apps/sales/sales-order-kanban-view.tsx`
- `apps/web/components/apps/sales/sales-order-stats-view.tsx`
- `apps/web/stores/business-app-store.ts`
- `apps/web/lib/business-apps.ts`

Navigation:

- Add a primary nav or sidebar link named `Apps` or `App Hub`.
- Add a visible `Sales App` entry under App Hub.
- Home Apps & Tools should route Sales App to `/workspace/apps/sales`.

Route state:

- `?mode=list`
- `?mode=search&q=acme`
- `?stage=proposal&status=open`

Mode changes update the URL without losing filters.

## AI-Native Boundary

Phase 73 should expose action metadata but not execute autonomous changes.

Allowed metadata actions:

- `summarize_order`
- `draft_follow_up`
- `explain_risk`
- `open_source_channel`
- `create_workflow_draft`

UI may show disabled or metadata-only action chips if no backend execution endpoint exists.

Future agents should be able to discover:

- app key
- entity type
- selected filters
- selected record IDs
- current mode
- routeable source links

## Acceptance Criteria

- Sales App is reachable from workspace navigation.
- Sales Orders data source renders in search, list, calendar, kanban, and stats modes.
- All modes use the same backend record contract.
- Mode/filter state is routeable and reload-safe.
- Calendar/Kanban/Stats are scoped display modes, not separate product builds.
- Sales Orders include channel/message linkage where available.
- Web keeps rendering usable empty/loading/error states.
- Backend tests cover filtering, pagination, stats, and mode-compatible response shape.
- Web tests or type checks cover mode switching and rendering of the Sales order fields.

