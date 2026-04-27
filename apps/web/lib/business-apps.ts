import type {
  BusinessApp,
  BusinessAppDataGroup,
  BusinessAppDataResponse,
  BusinessAppMode,
  BusinessAppQueryState,
  BusinessAppSchema,
  BusinessAppSchemaField,
  BusinessAppStats,
  BusinessAppStatsResponse,
  BusinessAppStatsStageBucket,
  BusinessCalendarEvent,
  BusinessCalendarTimeField,
  BusinessCalendarView,
  BusinessChartStyle,
  BusinessAppTimelineBucket,
  SalesOrder,
} from "@/types"

export const APPS_HUB_ROUTE = "/workspace/apps"
export const SALES_APP_KEY = "sales"
export const SALES_APP_ROUTE = "/workspace/apps/sales"

export const BUSINESS_APP_MODES: BusinessAppMode[] = ["list", "card_grid", "kanban", "calendar", "stat"]
export const BUSINESS_CALENDAR_VIEWS: BusinessCalendarView[] = ["day", "week", "month"]
export const BUSINESS_CALENDAR_TIME_FIELDS: BusinessCalendarTimeField[] = ["expected_close_date", "order_date", "due_date", "last_activity_at"]
export const BUSINESS_CHART_STYLES: BusinessChartStyle[] = ["summary", "bar", "funnel", "timeline"]
export const DEFAULT_BUSINESS_APP_MODE: BusinessAppMode = "list"
export const DEFAULT_BUSINESS_CALENDAR_VIEW: BusinessCalendarView = "month"
export const DEFAULT_BUSINESS_CALENDAR_TIME_FIELD: BusinessCalendarTimeField = "expected_close_date"
export const DEFAULT_BUSINESS_CHART_STYLE: BusinessChartStyle = "summary"
export const SALES_STAGE_ORDER = ["lead", "qualified", "proposal", "negotiation", "closed_won", "closed_lost"] as const
export const SALES_STATUS_ORDER = ["open", "at_risk", "won", "lost", "active", "archived"] as const

const LEGACY_MODE_ALIASES: Record<string, BusinessAppMode> = {
  search: "list",
  stats: "stat",
}

const STAGE_LABELS: Record<string, string> = {
  lead: "Lead",
  qualified: "Qualified",
  proposal: "Proposal",
  negotiation: "Negotiation",
  closed_won: "Closed Won",
  closed_lost: "Closed Lost",
}

const STATUS_LABELS: Record<string, string> = {
  open: "Open",
  at_risk: "At Risk",
  won: "Won",
  lost: "Lost",
  active: "Active",
  archived: "Archived",
}

const CALENDAR_TIME_FIELD_LABELS: Record<BusinessCalendarTimeField, string> = {
  expected_close_date: "Expected Close",
  order_date: "Order Date",
  due_date: "Due Date",
  last_activity_at: "Last Activity",
}

const CHART_STYLE_LABELS: Record<BusinessChartStyle, string> = {
  summary: "Summary",
  bar: "Bar",
  funnel: "Funnel",
  timeline: "Timeline",
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null
}

function normalizeOptionalString(value: unknown): string | undefined {
  return typeof value === "string" && value.trim() ? value.trim() : undefined
}

function normalizeNumber(value: unknown): number | undefined {
  if (typeof value === "number" && Number.isFinite(value)) return value
  if (typeof value === "string" && value.trim()) {
    const parsed = Number(value)
    if (Number.isFinite(parsed)) return parsed
  }
  return undefined
}

function normalizeStringArray(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.filter((item): item is string => typeof item === "string" && item.trim().length > 0).map(item => item.trim())
  }
  if (typeof value === "string" && value.trim()) {
    return value
      .split(",")
      .map(item => item.trim())
      .filter(Boolean)
  }
  return []
}

function normalizeModeAlias(value: unknown): BusinessAppMode | null {
  if (typeof value !== "string") return null
  if (LEGACY_MODE_ALIASES[value]) return LEGACY_MODE_ALIASES[value]
  return BUSINESS_APP_MODES.includes(value as BusinessAppMode) ? (value as BusinessAppMode) : null
}

function normalizeMode(value: unknown): BusinessAppMode | null {
  return normalizeModeAlias(value)
}

function normalizeModes(value: unknown): BusinessAppMode[] {
  if (!Array.isArray(value)) return BUSINESS_APP_MODES
  const modes = value
    .map(normalizeMode)
    .filter((mode): mode is BusinessAppMode => mode !== null)
  return modes.length > 0 ? Array.from(new Set(modes)) : BUSINESS_APP_MODES
}

function normalizeCalendarView(value: unknown): BusinessCalendarView | null {
  return typeof value === "string" && BUSINESS_CALENDAR_VIEWS.includes(value as BusinessCalendarView)
    ? (value as BusinessCalendarView)
    : null
}

function normalizeCalendarTimeField(value: unknown): BusinessCalendarTimeField | null {
  return typeof value === "string" && BUSINESS_CALENDAR_TIME_FIELDS.includes(value as BusinessCalendarTimeField)
    ? (value as BusinessCalendarTimeField)
    : null
}

function normalizeChartStyle(value: unknown): BusinessChartStyle | null {
  return typeof value === "string" && BUSINESS_CHART_STYLES.includes(value as BusinessChartStyle)
    ? (value as BusinessChartStyle)
    : null
}

function normalizeDateString(value: unknown): string | undefined {
  const normalized = normalizeOptionalString(value)
  if (!normalized) return undefined
  const parsed = new Date(normalized)
  return Number.isNaN(parsed.getTime()) ? undefined : parsed.toISOString()
}

function normalizeSchemaField(raw: unknown): BusinessAppSchemaField | null {
  if (!isRecord(raw)) return null
  const key = normalizeOptionalString(raw.key)
  const label = normalizeOptionalString(raw.label)
  const type = normalizeOptionalString(raw.type)
  if (!key || !label || !type) return null
  return { key, label, type }
}

function normalizeSchema(raw: unknown): BusinessAppSchema | null {
  if (!isRecord(raw)) return null
  const entity = normalizeOptionalString(raw.entity)
  const fields = Array.isArray(raw.fields)
    ? raw.fields.map(normalizeSchemaField).filter((field): field is BusinessAppSchemaField => field !== null)
    : []
  if (!entity) return null
  return { entity, fields }
}

export function isBusinessAppMode(value: string | null | undefined): value is BusinessAppMode {
  return !!value && normalizeModeAlias(value) !== null
}

export function getBusinessAppModeLabel(mode: BusinessAppMode): string {
  if (mode === "card_grid") return "Card Grid"
  if (mode === "kanban") return "Kanban"
  if (mode === "stat") return "Stat"
  return mode.charAt(0).toUpperCase() + mode.slice(1)
}

export function getBusinessCalendarViewLabel(view: BusinessCalendarView): string {
  return view.charAt(0).toUpperCase() + view.slice(1)
}

export function getBusinessCalendarTimeFieldLabel(timeField: BusinessCalendarTimeField): string {
  return CALENDAR_TIME_FIELD_LABELS[timeField]
}

export function getBusinessChartStyleLabel(chartStyle: BusinessChartStyle): string {
  return CHART_STYLE_LABELS[chartStyle]
}

export function parseBusinessAppQueryState(searchParams: URLSearchParams, fallbackMode: BusinessAppMode = DEFAULT_BUSINESS_APP_MODE): BusinessAppQueryState {
  const modeParam = searchParams.get("mode")
  const mode = normalizeModeAlias(modeParam) ?? fallbackMode
  const q = normalizeOptionalString(searchParams.get("q"))
  const stage = normalizeOptionalString(searchParams.get("stage"))
  const status = normalizeOptionalString(searchParams.get("status"))
  const owner_user_id = normalizeOptionalString(searchParams.get("owner_user_id"))
  const date_from = normalizeOptionalString(searchParams.get("date_from"))
  const date_to = normalizeOptionalString(searchParams.get("date_to"))
  const calendar_view = normalizeCalendarView(searchParams.get("calendar_view")) ?? DEFAULT_BUSINESS_CALENDAR_VIEW
  const calendar_time_field = normalizeCalendarTimeField(searchParams.get("calendar_time_field")) ?? DEFAULT_BUSINESS_CALENDAR_TIME_FIELD
  const chart_style = normalizeChartStyle(searchParams.get("chart_style")) ?? DEFAULT_BUSINESS_CHART_STYLE
  return {
    mode,
    ...(q ? { q } : {}),
    ...(stage ? { stage } : {}),
    ...(status ? { status } : {}),
    ...(owner_user_id ? { owner_user_id } : {}),
    ...(date_from ? { date_from } : {}),
    ...(date_to ? { date_to } : {}),
    ...(mode === "calendar" ? { calendar_view, calendar_time_field } : {}),
    ...(mode === "stat" ? { chart_style } : {}),
  }
}

export function buildBusinessAppSearchParams(
  state: Partial<BusinessAppQueryState>,
  options?: { includeCursor?: boolean; includeLimit?: boolean }
): URLSearchParams {
  const params = new URLSearchParams()
  const includeCursor = options?.includeCursor ?? true
  const includeLimit = options?.includeLimit ?? true

  const append = (key: string, value?: string | number | null) => {
    if (value === undefined || value === null) return
    const normalized = typeof value === "number" ? String(value) : value.trim()
    if (!normalized) return
    params.set(key, normalized)
  }

  if (state.mode) append("mode", state.mode)
  append("q", state.q)
  append("stage", state.stage)
  append("status", state.status)
  append("owner_user_id", state.owner_user_id)
  append("date_from", state.date_from)
  append("date_to", state.date_to)
  if (state.mode === "calendar") {
    append("calendar_view", state.calendar_view)
    append("calendar_time_field", state.calendar_time_field)
  }
  if (state.mode === "stat") {
    append("chart_style", state.chart_style)
  }
  if (includeLimit) append("limit", state.limit)
  if (includeCursor) append("cursor", state.cursor)
  return params
}

export function buildBusinessAppQueryString(
  state: Partial<BusinessAppQueryState>,
  options?: { includeCursor?: boolean; includeLimit?: boolean }
): string {
  return buildBusinessAppSearchParams(state, options).toString()
}

export function buildBusinessAppCacheKey(appKey: string, state: Partial<BusinessAppQueryState>): string {
  const query = buildBusinessAppQueryString(state)
  return query ? `${appKey}?${query}` : appKey
}

export function normalizeBusinessApp(raw: unknown): BusinessApp | null {
  if (!isRecord(raw)) return null
  const key = normalizeOptionalString(raw.key) ?? normalizeOptionalString(raw.id)
  const title = normalizeOptionalString(raw.title)
  if (!key || !title) return null
  return {
    id: normalizeOptionalString(raw.id) ?? key,
    key,
    title,
    description: normalizeOptionalString(raw.description),
    icon: normalizeOptionalString(raw.icon),
    modes: normalizeModes(raw.modes),
    primary_entity: normalizeOptionalString(raw.primary_entity),
  }
}

export function normalizeBusinessAppsResponse(data: unknown): BusinessApp[] {
  if (!isRecord(data) || !Array.isArray(data.apps)) return []
  return data.apps
    .map(normalizeBusinessApp)
    .filter((app): app is BusinessApp => app !== null)
}

export function normalizeSalesOrder(raw: unknown): SalesOrder | null {
  if (!isRecord(raw)) return null
  const id = normalizeOptionalString(raw.id)
  const orderNumber = normalizeOptionalString(raw.order_number)
  const customerName = normalizeOptionalString(raw.customer_name)
  if (!id || !orderNumber || !customerName) return null
  return {
    id,
    order_number: orderNumber,
    customer_name: customerName,
    customer_id: normalizeOptionalString(raw.customer_id),
    owner_user_id: normalizeOptionalString(raw.owner_user_id),
    owner_name: normalizeOptionalString(raw.owner_name),
    stage: normalizeOptionalString(raw.stage),
    status: normalizeOptionalString(raw.status),
    priority: normalizeOptionalString(raw.priority),
    amount: normalizeNumber(raw.amount),
    currency: normalizeOptionalString(raw.currency),
    probability: normalizeNumber(raw.probability),
    expected_close_date: normalizeOptionalString(raw.expected_close_date) ?? null,
    order_date: normalizeOptionalString(raw.order_date) ?? null,
    due_date: normalizeOptionalString(raw.due_date) ?? null,
    last_activity_at: normalizeOptionalString(raw.last_activity_at) ?? null,
    source_channel_id: normalizeOptionalString(raw.source_channel_id),
    source_message_id: normalizeOptionalString(raw.source_message_id),
    summary: normalizeOptionalString(raw.summary),
    tags: normalizeStringArray(raw.tags),
    created_at: normalizeOptionalString(raw.created_at),
    updated_at: normalizeOptionalString(raw.updated_at),
  }
}

function normalizeCalendarEvent(raw: unknown): BusinessCalendarEvent | null {
  if (!isRecord(raw)) return null
  const id = normalizeOptionalString(raw.id)
  const record_id = normalizeOptionalString(raw.record_id)
  const title = normalizeOptionalString(raw.title)
  const start = normalizeDateString(raw.start)
  const end = normalizeDateString(raw.end)
  const time_field = normalizeCalendarTimeField(raw.time_field)
  if (!id || !record_id || !title || !start || !end || !time_field) return null
  return {
    id,
    record_id,
    title,
    start,
    end,
    time_field,
    amount: normalizeNumber(raw.amount),
    stage: normalizeOptionalString(raw.stage),
    status: normalizeOptionalString(raw.status),
    record: normalizeSalesOrder(raw.record) ?? null,
  }
}

function buildFallbackCalendarEvents(records: SalesOrder[], timeField: BusinessCalendarTimeField): BusinessCalendarEvent[] {
  return records.flatMap((record) => {
    const value = record[timeField]
    const start = normalizeDateString(value)
    if (!start) return []
    const startDate = new Date(start)
    const endDate = new Date(startDate)
    endDate.setUTCHours(23, 59, 59, 999)
    return [{
      id: `sales-order:${record.id}:${timeField}`,
      record_id: record.id,
      title: `${record.order_number} · ${record.customer_name}`,
      start,
      end: endDate.toISOString(),
      time_field: timeField,
      amount: record.amount,
      stage: record.stage,
      status: record.status,
      record,
    }]
  })
}

function normalizeDataGroup(raw: unknown): BusinessAppDataGroup | null {
  if (!isRecord(raw)) return null
  const key = normalizeOptionalString(raw.key) ?? normalizeOptionalString(raw.id) ?? normalizeOptionalString(raw.title)
  const title = normalizeOptionalString(raw.title) ?? key
  if (!key || !title) return null
  const rawRecords = Array.isArray(raw.records) ? raw.records : []
  const records = rawRecords.map(normalizeSalesOrder).filter((record): record is SalesOrder => record !== null)
  return {
    id: normalizeOptionalString(raw.id) ?? key,
    key,
    title,
    count: normalizeNumber(raw.count) ?? records.length,
    records,
  }
}

export function getBusinessAppRoute(appKey: string): string {
  return appKey === SALES_APP_KEY ? SALES_APP_ROUTE : APPS_HUB_ROUTE
}

export function getSalesOrderStageLabel(stage?: string): string {
  if (!stage) return "Unstaged"
  return STAGE_LABELS[stage] ?? stage.replace(/_/g, " ")
}

export function getSalesOrderStatusLabel(status?: string): string {
  if (!status) return "Unknown"
  return STATUS_LABELS[status] ?? status.replace(/_/g, " ")
}

export function getSalesOrderStageTone(stage?: string): string {
  if (stage === "closed_won") return "bg-emerald-500/15 text-emerald-700 dark:text-emerald-300"
  if (stage === "closed_lost") return "bg-rose-500/15 text-rose-700 dark:text-rose-300"
  if (stage === "negotiation") return "bg-violet-500/15 text-violet-700 dark:text-violet-300"
  if (stage === "proposal") return "bg-sky-500/15 text-sky-700 dark:text-sky-300"
  if (stage === "qualified") return "bg-amber-500/15 text-amber-700 dark:text-amber-300"
  return "bg-muted/60 text-muted-foreground"
}

export function getSalesOrderStatusTone(status?: string): string {
  if (status === "won") return "bg-emerald-500/15 text-emerald-700 dark:text-emerald-300"
  if (status === "lost") return "bg-rose-500/15 text-rose-700 dark:text-rose-300"
  if (status === "at_risk") return "bg-amber-500/15 text-amber-700 dark:text-amber-300"
  if (status === "open" || status === "active") return "bg-sky-500/15 text-sky-700 dark:text-sky-300"
  return "bg-muted/60 text-muted-foreground"
}

export function formatSalesCurrency(amount?: number, currency = "USD"): string {
  if (amount === undefined) return "—"
  try {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: currency || "USD",
      maximumFractionDigits: 0,
    }).format(amount)
  } catch {
    return `${currency || "USD"} ${amount.toLocaleString()}`
  }
}

export function formatBusinessAppDate(value?: string | null): string {
  if (!value) return "—"
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return new Intl.DateTimeFormat("en-US", { month: "short", day: "numeric", year: "numeric" }).format(parsed)
}

export function resolveSalesOrderSourceHref(order: Pick<SalesOrder, "source_channel_id">): string | null {
  if (!order.source_channel_id) return null
  return `/workspace?c=${order.source_channel_id}`
}

export function getSalesStageOptions(): Array<{ value: string; label: string }> {
  return SALES_STAGE_ORDER.map(stage => ({ value: stage, label: getSalesOrderStageLabel(stage) }))
}

export function getSalesStatusOptions(): Array<{ value: string; label: string }> {
  return SALES_STATUS_ORDER.map(status => ({ value: status, label: getSalesOrderStatusLabel(status) }))
}

export function buildBusinessAppGroups(
  mode: BusinessAppMode,
  records: SalesOrder[],
  options?: { calendarTimeField?: BusinessCalendarTimeField }
): BusinessAppDataGroup[] {
  if (mode === "kanban") {
    const groups = new Map<string, SalesOrder[]>()
    for (const stage of SALES_STAGE_ORDER) {
      groups.set(stage, [])
    }
    for (const record of records) {
      const key = record.stage && groups.has(record.stage) ? record.stage : "lead"
      groups.set(key, [...(groups.get(key) ?? []), record])
    }
    return Array.from(groups.entries()).map(([key, bucketRecords]) => ({
      id: key,
      key,
      title: getSalesOrderStageLabel(key),
      count: bucketRecords.length,
      records: bucketRecords,
    }))
  }

  if (mode === "calendar") {
    const timeField = options?.calendarTimeField ?? DEFAULT_BUSINESS_CALENDAR_TIME_FIELD
    const dated = new Map<string, SalesOrder[]>()
    const unscheduled: SalesOrder[] = []
    for (const record of records) {
      const recordDate = record[timeField]
      if (!recordDate) {
        unscheduled.push(record)
        continue
      }
      const parsed = new Date(recordDate)
      if (Number.isNaN(parsed.getTime())) {
        unscheduled.push(record)
        continue
      }
      const key = parsed.toISOString().slice(0, 10)
      dated.set(key, [...(dated.get(key) ?? []), record])
    }

    const groups = Array.from(dated.entries())
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([key, bucketRecords]) => ({
        id: key,
        key,
        title: formatBusinessAppDate(key),
        count: bucketRecords.length,
        records: bucketRecords,
      }))

    if (unscheduled.length > 0) {
      groups.push({
        id: "unscheduled",
        key: "unscheduled",
        title: "Unscheduled",
        count: unscheduled.length,
        records: unscheduled,
      })
    }

    return groups
  }

  return []
}

export function normalizeBusinessAppDataResponse(
  data: unknown,
  fallbackMode: BusinessAppMode = DEFAULT_BUSINESS_APP_MODE,
  fallbackApp?: BusinessApp | null
): BusinessAppDataResponse {
  const root = isRecord(data) ? data : {}
  const app = normalizeBusinessApp(root.app) ?? fallbackApp ?? null
  const viewRoot = isRecord(root.view) ? root.view : {}
  const mode = normalizeMode(viewRoot.mode) ?? fallbackMode
  const availableModes = normalizeModes(viewRoot.available_modes ?? app?.modes)
  const calendar_view = normalizeCalendarView(viewRoot.calendar_view) ?? DEFAULT_BUSINESS_CALENDAR_VIEW
  const calendar_time_field = normalizeCalendarTimeField(viewRoot.calendar_time_field) ?? DEFAULT_BUSINESS_CALENDAR_TIME_FIELD
  const chart_style = normalizeChartStyle(viewRoot.chart_style) ?? DEFAULT_BUSINESS_CHART_STYLE
  const schema = normalizeSchema(root.schema)
  const rawRecords = Array.isArray(root.records) ? root.records : Array.isArray(root.data) ? root.data : []
  const records = rawRecords.map(normalizeSalesOrder).filter((record): record is SalesOrder => record !== null)
  const rawGroups = Array.isArray(root.groups) ? root.groups : []
  const rawCalendarEvents = Array.isArray(root.calendar_events) ? root.calendar_events : []
  const groups = rawGroups.length > 0
    ? rawGroups.map(normalizeDataGroup).filter((group): group is BusinessAppDataGroup => group !== null)
    : buildBusinessAppGroups(mode, records, { calendarTimeField: calendar_time_field })
  const calendar_events = rawCalendarEvents.length > 0
    ? rawCalendarEvents.map(normalizeCalendarEvent).filter((event): event is BusinessCalendarEvent => event !== null)
    : mode === "calendar"
      ? buildFallbackCalendarEvents(records, calendar_time_field)
      : []
  return {
    app,
    view: {
      mode,
      available_modes: availableModes,
      ...(mode === "calendar" ? { calendar_view, calendar_time_field } : {}),
      ...(mode === "stat" ? { chart_style } : {}),
    },
    schema,
    records,
    groups,
    calendar_events,
    next_cursor: normalizeOptionalString(root.next_cursor) ?? null,
  }
}

function normalizeStageBuckets(value: unknown): BusinessAppStatsStageBucket[] {
  if (!Array.isArray(value)) return []
  return value
    .map((item) => {
      if (!isRecord(item)) return null
      const stage = normalizeOptionalString(item.stage)
      if (!stage) return null
      return {
        stage,
        count: normalizeNumber(item.count) ?? 0,
        sum: normalizeNumber(item.sum) ?? 0,
      }
    })
    .filter((bucket): bucket is BusinessAppStatsStageBucket => bucket !== null)
}

function normalizeTimelineBuckets(value: unknown): BusinessAppTimelineBucket[] {
  if (!Array.isArray(value)) return []
  return value
    .map((item) => {
      if (!isRecord(item)) return null
      const label = normalizeOptionalString(item.label)
      if (!label) return null
      return {
        label,
        count: normalizeNumber(item.count) ?? 0,
        amount: normalizeNumber(item.amount ?? item.sum) ?? 0,
      }
    })
    .filter((bucket): bucket is BusinessAppTimelineBucket => bucket !== null)
}

export function normalizeBusinessAppStatsResponse(data: unknown, fallbackApp?: BusinessApp | null): BusinessAppStatsResponse {
  const root = isRecord(data) ? data : {}
  const app = normalizeBusinessApp(root.app) ?? fallbackApp ?? null
  const rawStats = isRecord(root.stats) ? root.stats : {}
  const byStage = normalizeStageBuckets(rawStats.by_stage)
  const funnel = normalizeStageBuckets(rawStats.funnel)
  const timelineBuckets = normalizeTimelineBuckets(rawStats.timeline_buckets)
  const expectedCloseBuckets = normalizeTimelineBuckets(rawStats.expected_close_buckets)
  const orderCount = normalizeNumber(rawStats.order_count) ?? byStage.reduce((sum, item) => sum + item.count, 0)
  const wonCount = normalizeNumber(rawStats.won_count) ?? (byStage.find(item => item.stage === "closed_won")?.count ?? 0)
  const lostCount = normalizeNumber(rawStats.lost_count) ?? (byStage.find(item => item.stage === "closed_lost")?.count ?? 0)
  const stats: BusinessAppStats = {
    total_amount: normalizeNumber(rawStats.total_amount),
    weighted_pipeline_amount: normalizeNumber(rawStats.weighted_pipeline_amount),
    order_count: orderCount,
    open_count: normalizeNumber(rawStats.open_count) ?? Math.max(orderCount - wonCount - lostCount, 0),
    won_count: wonCount,
    lost_count: lostCount,
    at_risk_count: normalizeNumber(rawStats.at_risk_count),
    by_stage: byStage,
    funnel: funnel.length > 0 ? funnel : undefined,
    timeline_buckets: timelineBuckets.length > 0 ? timelineBuckets : undefined,
    expected_close_buckets: expectedCloseBuckets.length > 0 ? expectedCloseBuckets : timelineBuckets.length > 0 ? timelineBuckets : undefined,
  }
  return {
    app,
    chart_style: normalizeChartStyle(root.chart_style ?? rawStats.chart_style) ?? DEFAULT_BUSINESS_CHART_STYLE,
    stats,
  }
}
