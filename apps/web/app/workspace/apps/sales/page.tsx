"use client"

import Link from "next/link"
import { useEffect, useMemo, useState } from "react"
import { ChevronLeft } from "lucide-react"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import { useShallow } from "zustand/react/shallow"
import { BusinessAppShell } from "@/components/apps/business-app-shell"
import { SalesOrderCardGridView } from "@/components/apps/sales/sales-order-card-grid-view"
import { SalesOrderCalendarView } from "@/components/apps/sales/sales-order-calendar-view"
import { SalesOrderKanbanView } from "@/components/apps/sales/sales-order-kanban-view"
import { SalesOrderListView } from "@/components/apps/sales/sales-order-list-view"
import { SalesOrderStatsView } from "@/components/apps/sales/sales-order-stats-view"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  APPS_HUB_ROUTE,
  buildBusinessAppCacheKey,
  buildBusinessAppGroups,
  buildBusinessAppSearchParams,
  DEFAULT_BUSINESS_APP_MODE,
  getBusinessCalendarTimeFieldLabel,
  getBusinessCalendarViewLabel,
  getBusinessChartStyleLabel,
  BUSINESS_CALENDAR_TIME_FIELDS,
  BUSINESS_CALENDAR_VIEWS,
  BUSINESS_CHART_STYLES,
  getSalesStageOptions,
  getSalesStatusOptions,
  parseBusinessAppQueryState,
  SALES_APP_KEY,
} from "@/lib/business-apps"
import { useBusinessAppStore } from "@/stores/business-app-store"
import type { BusinessAppDataResponse, BusinessAppQueryState, BusinessAppStatsResponse, SalesOrder } from "@/types"

const PAGE_LIMIT = 20

function dedupeOrders(records: SalesOrder[]): SalesOrder[] {
  const seen = new Set<string>()
  return records.filter((record) => {
    if (seen.has(record.id)) return false
    seen.add(record.id)
    return true
  })
}

function createEmptyData(mode: BusinessAppQueryState["mode"]): BusinessAppDataResponse {
  return {
    app: null,
    view: { mode, available_modes: ["list", "card_grid", "kanban", "calendar", "stat"] },
    schema: null,
    records: [],
    groups: [],
    calendar_events: [],
    next_cursor: null,
  }
}

export default function SalesAppPage() {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const urlQuery = useMemo(
    () => parseBusinessAppQueryState(new URLSearchParams(searchParams.toString()), DEFAULT_BUSINESS_APP_MODE),
    [searchParams]
  )
  const stageOptions = useMemo(() => getSalesStageOptions(), [])
  const statusOptions = useMemo(() => getSalesStatusOptions(), [])

  const {
    app,
    appError,
    fetchApp,
    fetchAppData,
    fetchAppStats,
  } = useBusinessAppStore(useShallow((state) => ({
    app: state.appsById[SALES_APP_KEY] ?? null,
    appError: state.appErrors[SALES_APP_KEY] ?? null,
    fetchApp: state.fetchApp,
    fetchAppData: state.fetchAppData,
    fetchAppStats: state.fetchAppStats,
  })))

  const [searchValue, setSearchValue] = useState(urlQuery.q ?? "")
  const [dataState, setDataState] = useState<BusinessAppDataResponse>(() => createEmptyData(urlQuery.mode))
  const [statsState, setStatsState] = useState<BusinessAppStatsResponse | null>(null)
  const [isLoadingContent, setIsLoadingContent] = useState(true)
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [dataError, setDataError] = useState<string | null>(null)
  const [statsError, setStatsError] = useState<string | null>(null)

  const baseQuery = useMemo<BusinessAppQueryState>(
    () => ({ ...urlQuery, limit: PAGE_LIMIT }),
    [urlQuery]
  )
  const baseDataCacheKey = useMemo(() => buildBusinessAppCacheKey(SALES_APP_KEY, baseQuery), [baseQuery])
  const baseStatsCacheKey = useMemo(
    () => buildBusinessAppCacheKey(SALES_APP_KEY, {
      mode: "stat",
      q: urlQuery.q,
      stage: urlQuery.stage,
      status: urlQuery.status,
      owner_user_id: urlQuery.owner_user_id,
      date_from: urlQuery.date_from,
      date_to: urlQuery.date_to,
      chart_style: urlQuery.chart_style,
    }),
    [urlQuery.chart_style, urlQuery.date_from, urlQuery.date_to, urlQuery.owner_user_id, urlQuery.q, urlQuery.stage, urlQuery.status]
  )

  useEffect(() => {
    setSearchValue(urlQuery.q ?? "")
  }, [urlQuery.q])

  useEffect(() => {
    fetchApp(SALES_APP_KEY).catch(() => { /* handled in store */ })
  }, [fetchApp])

  useEffect(() => {
    let cancelled = false
    setIsLoadingContent(true)
    setDataError(null)
    setStatsError(null)

    if (urlQuery.mode === "stat") {
      fetchAppStats(SALES_APP_KEY, {
        mode: "stat",
        q: urlQuery.q,
        stage: urlQuery.stage,
        status: urlQuery.status,
        owner_user_id: urlQuery.owner_user_id,
        date_from: urlQuery.date_from,
        date_to: urlQuery.date_to,
        chart_style: urlQuery.chart_style,
      }).then((response) => {
        if (cancelled) return
        setStatsState(response)
        setIsLoadingContent(false)
        if (!response) {
          setStatsError("Sales stats are temporarily unavailable.")
        }
      })
      return () => {
        cancelled = true
      }
    }

    fetchAppData(SALES_APP_KEY, baseQuery).then((response) => {
      if (cancelled) return
      setStatsState(null)
      if (!response) {
        setDataState(createEmptyData(urlQuery.mode))
        setDataError("Sales data is temporarily unavailable.")
      } else {
        setDataState(response)
      }
      setIsLoadingContent(false)
    })

    return () => {
      cancelled = true
    }
  }, [baseDataCacheKey, baseQuery, baseStatsCacheKey, fetchAppData, fetchAppStats, urlQuery.date_from, urlQuery.date_to, urlQuery.mode, urlQuery.owner_user_id, urlQuery.q, urlQuery.stage, urlQuery.status])

  const updateRoute = (patch: Partial<BusinessAppQueryState>) => {
    const nextQuery: BusinessAppQueryState = {
      ...urlQuery,
      ...patch,
      mode: (patch.mode ?? urlQuery.mode) as BusinessAppQueryState["mode"],
      cursor: undefined,
    }
    const params = buildBusinessAppSearchParams(nextQuery, { includeCursor: false, includeLimit: false })
    const queryString = params.toString()
    router.replace(queryString ? `${pathname}?${queryString}` : pathname, { scroll: false })
  }

  const handleLoadMore = async () => {
    if (!dataState.next_cursor) return
    setIsLoadingMore(true)
    const response = await fetchAppData(SALES_APP_KEY, { ...baseQuery, cursor: dataState.next_cursor })
    setIsLoadingMore(false)
    if (!response) {
      setDataError("Sales data is temporarily unavailable.")
      return
    }
    const merged = dedupeOrders([...dataState.records, ...response.records])
    setDataState({
      ...response,
      records: merged,
      groups: buildBusinessAppGroups(urlQuery.mode, merged, { calendarTimeField: urlQuery.calendar_time_field }),
    })
  }

  const modeControls = useMemo(() => {
    if (urlQuery.mode === "calendar") {
      return (
        <>
          <select
            className="h-9 rounded-md border bg-background px-3 text-sm"
            value={urlQuery.calendar_view ?? "month"}
            onChange={(event) => updateRoute({ calendar_view: event.target.value as BusinessAppQueryState["calendar_view"] })}
          >
            {BUSINESS_CALENDAR_VIEWS.map((view) => (
              <option key={view} value={view}>{getBusinessCalendarViewLabel(view)}</option>
            ))}
          </select>
          <select
            className="h-9 rounded-md border bg-background px-3 text-sm"
            value={urlQuery.calendar_time_field ?? "expected_close_date"}
            onChange={(event) => updateRoute({ calendar_time_field: event.target.value as BusinessAppQueryState["calendar_time_field"] })}
          >
            {BUSINESS_CALENDAR_TIME_FIELDS.map((timeField) => (
              <option key={timeField} value={timeField}>{getBusinessCalendarTimeFieldLabel(timeField)}</option>
            ))}
          </select>
        </>
      )
    }

    if (urlQuery.mode === "stat") {
      return (
        <select
          className="h-9 rounded-md border bg-background px-3 text-sm"
          value={urlQuery.chart_style ?? "summary"}
          onChange={(event) => updateRoute({ chart_style: event.target.value as BusinessAppQueryState["chart_style"] })}
        >
          {BUSINESS_CHART_STYLES.map((chartStyle) => (
            <option key={chartStyle} value={chartStyle}>{getBusinessChartStyleLabel(chartStyle)}</option>
          ))}
        </select>
      )
    }

    return null
  }, [urlQuery.calendar_time_field, urlQuery.calendar_view, urlQuery.chart_style, urlQuery.mode])

  const summary = (
    <div className="flex flex-wrap items-center gap-2">
      <Badge variant="secondary" className="text-[10px] font-black uppercase tracking-widest">
        Mode {urlQuery.mode}
      </Badge>
      {urlQuery.stage ? (
        <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
          Stage {urlQuery.stage.replace(/_/g, " ")}
        </Badge>
      ) : null}
      {urlQuery.status ? (
        <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
          Status {urlQuery.status.replace(/_/g, " ")}
        </Badge>
      ) : null}
      {urlQuery.date_from || urlQuery.date_to ? (
        <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
          Date filtered
        </Badge>
      ) : null}
      {urlQuery.q ? (
        <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
          Search active
        </Badge>
      ) : null}
      {urlQuery.mode === "calendar" ? (
        <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
          {getBusinessCalendarViewLabel(urlQuery.calendar_view ?? "month")} · {getBusinessCalendarTimeFieldLabel(urlQuery.calendar_time_field ?? "expected_close_date")}
        </Badge>
      ) : null}
      {urlQuery.mode === "stat" ? (
        <span className="text-xs text-muted-foreground">Read-only metrics sourced from Gemini&apos;s Phase 74 Sales App backend with chart-style switching.</span>
      ) : (
        <span className="text-xs text-muted-foreground">
          {dataState.records.length} visible order{dataState.records.length === 1 ? "" : "s"}
        </span>
      )}
    </div>
  )

  const availableModes = app?.modes ?? dataState.view.available_modes

  return (
    <div className="flex h-full w-full flex-col overflow-hidden">
      <div className="border-b bg-background px-6 py-3">
        <Button asChild variant="ghost" size="sm" className="gap-1.5 text-[11px] font-black uppercase tracking-widest">
          <Link href={APPS_HUB_ROUTE}>
            <ChevronLeft className="h-3.5 w-3.5" />
            App Hub
          </Link>
        </Button>
      </div>

      <BusinessAppShell
        app={app}
        mode={urlQuery.mode}
        availableModes={availableModes}
        searchValue={searchValue}
        stage={urlQuery.stage}
        status={urlQuery.status}
        dateFrom={urlQuery.date_from}
        dateTo={urlQuery.date_to}
        onModeChange={(mode) => updateRoute({ mode })}
        onSearchValueChange={setSearchValue}
        onSearchSubmit={() => updateRoute({ q: searchValue || undefined })}
        onSearchClear={() => {
          setSearchValue("")
          updateRoute({ q: undefined })
        }}
        onFiltersChange={(patch) => updateRoute(patch)}
        stageOptions={stageOptions}
        statusOptions={statusOptions}
        modeControls={modeControls}
        summary={summary}
      >
        {urlQuery.mode === "stat" ? (
          <SalesOrderStatsView
            stats={statsState?.stats ?? null}
            chartStyle={statsState?.chart_style ?? urlQuery.chart_style ?? "summary"}
            isLoading={isLoadingContent}
            error={statsError ?? appError}
          />
        ) : urlQuery.mode === "calendar" ? (
          <SalesOrderCalendarView
            events={dataState.calendar_events}
            calendarView={urlQuery.calendar_view ?? "month"}
            calendarTimeField={urlQuery.calendar_time_field ?? "expected_close_date"}
            isLoading={isLoadingContent}
            error={dataError ?? appError}
          />
        ) : urlQuery.mode === "card_grid" ? (
          <SalesOrderCardGridView
            records={dataState.records}
            schema={dataState.schema}
            isLoading={isLoadingContent}
            error={dataError ?? appError}
            nextCursor={dataState.next_cursor}
            onLoadMore={handleLoadMore}
            isLoadingMore={isLoadingMore}
          />
        ) : urlQuery.mode === "kanban" ? (
          <SalesOrderKanbanView groups={dataState.groups} isLoading={isLoadingContent} error={dataError ?? appError} />
        ) : (
          <SalesOrderListView
            records={dataState.records}
            schema={dataState.schema}
            isLoading={isLoadingContent}
            error={dataError ?? appError}
            nextCursor={dataState.next_cursor}
            onLoadMore={handleLoadMore}
            isLoadingMore={isLoadingMore}
          />
        )}
      </BusinessAppShell>
    </div>
  )
}
