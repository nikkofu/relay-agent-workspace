import { APPS_HUB_ROUTE, SALES_APP_ROUTE } from "@/lib/business-apps"
import type { HomeAppToolItem, WorkspaceView, WorkspaceViewAction, WorkspaceViewKnownType, WorkspaceViewType } from "@/types"

const KNOWN_WORKSPACE_VIEW_TYPES: WorkspaceViewKnownType[] = [
  "list",
  "calendar",
  "search",
  "report",
  "form",
  "channel_messages",
]

const SUPPORTED_APP_TOOL_ROUTES = new Set<string>([
  APPS_HUB_ROUTE,
  SALES_APP_ROUTE,
  "/workspace/search",
  "/workspace/workflows",
  "/workspace/files",
])

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null
}

function normalizeOptionalString(value: unknown): string | undefined {
  return typeof value === "string" && value.trim() ? value : undefined
}

function normalizeString(value: unknown, fallback = ""): string {
  return typeof value === "string" ? value : fallback
}

function normalizeFilters(value: unknown): Record<string, unknown> {
  return isRecord(value) ? value : {}
}

function normalizeActions(value: unknown): WorkspaceViewAction[] {
  if (!Array.isArray(value)) return []
  return value.filter((item): item is WorkspaceViewAction => isRecord(item) && typeof item.type === "string")
}

export function isKnownWorkspaceViewType(value: string): value is WorkspaceViewKnownType {
  return KNOWN_WORKSPACE_VIEW_TYPES.includes(value as WorkspaceViewKnownType)
}

export function getWorkspaceViewKindLabel(viewType: WorkspaceViewType): string {
  if (!isKnownWorkspaceViewType(viewType)) return "View"
  if (viewType === "channel_messages") return "Channel"
  return viewType.charAt(0).toUpperCase() + viewType.slice(1)
}

export function normalizeWorkspaceView(raw: unknown): WorkspaceView | null {
  if (!isRecord(raw)) return null
  const id = normalizeOptionalString(raw.id)
  const title = normalizeOptionalString(raw.title)
  const viewType = normalizeOptionalString(raw.view_type)
  const source = normalizeOptionalString(raw.source)
  const createdBy = normalizeOptionalString(raw.created_by)
  const createdAt = normalizeOptionalString(raw.created_at)
  const updatedAt = normalizeOptionalString(raw.updated_at)
  if (!id || !title || !viewType || !source || !createdBy || !createdAt || !updatedAt) {
    return null
  }

  return {
    id,
    title,
    view_type: viewType,
    source,
    primary_channel_id: normalizeOptionalString(raw.primary_channel_id),
    filters: normalizeFilters(raw.filters),
    actions: normalizeActions(raw.actions),
    created_by: createdBy,
    created_at: createdAt,
    updated_at: updatedAt,
  }
}

export function normalizeWorkspaceViewsResponse(data: unknown): WorkspaceView[] {
  if (!isRecord(data) || !Array.isArray(data.views)) return []
  return data.views
    .map(normalizeWorkspaceView)
    .filter((view): view is WorkspaceView => view !== null)
}

export function resolveWorkspaceViewHref(view: Pick<WorkspaceView, "view_type" | "primary_channel_id">): string | null {
  if (view.view_type === "search") return "/workspace/search"
  if (view.view_type === "channel_messages" && view.primary_channel_id) {
    return `/workspace?c=${view.primary_channel_id}`
  }
  return null
}

export function resolveHomeAppToolHref(item: Pick<HomeAppToolItem, "route">): string | null {
  if (!item.route) return null
  return SUPPORTED_APP_TOOL_ROUTES.has(item.route) ? item.route : null
}

export function getWorkspaceViewSourceLabel(source: string): string {
  const normalized = normalizeString(source).toLowerCase()
  if (!normalized) return "Unknown"
  if (normalized === "manual") return "Manual"
  if (normalized === "agent") return "Agent"
  if (normalized === "tool") return "Tool"
  if (normalized === "system") return "System"
  return normalized.charAt(0).toUpperCase() + normalized.slice(1)
}
