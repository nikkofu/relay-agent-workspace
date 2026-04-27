"use client"

import type { FormEvent, ReactNode } from "react"
import { Bot, Filter, Search, Sparkles } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { BusinessViewSwitcher } from "@/components/apps/business-view-switcher"
import { cn } from "@/lib/utils"
import type { BusinessApp, BusinessAppMode } from "@/types"

interface SelectOption {
  value: string
  label: string
}

interface BusinessAppShellProps {
  app: BusinessApp | null
  mode: BusinessAppMode
  availableModes: BusinessAppMode[]
  searchValue: string
  stage?: string
  status?: string
  dateFrom?: string
  dateTo?: string
  onModeChange: (mode: BusinessAppMode) => void
  onSearchValueChange: (value: string) => void
  onSearchSubmit: () => void
  onSearchClear: () => void
  onFiltersChange: (patch: { stage?: string; status?: string; date_from?: string; date_to?: string }) => void
  stageOptions: SelectOption[]
  statusOptions: SelectOption[]
  modeControls?: ReactNode
  summary?: ReactNode
  children: ReactNode
  className?: string
}

export function BusinessAppShell({
  app,
  mode,
  availableModes,
  searchValue,
  stage,
  status,
  dateFrom,
  dateTo,
  onModeChange,
  onSearchValueChange,
  onSearchSubmit,
  onSearchClear,
  onFiltersChange,
  stageOptions,
  statusOptions,
  modeControls,
  summary,
  children,
  className,
}: BusinessAppShellProps) {
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    onSearchSubmit()
  }

  return (
    <div className={cn("flex flex-1 flex-col overflow-hidden bg-white dark:bg-[#1a1d21]", className)}>
      <div className="border-b bg-gradient-to-br from-violet-500/8 via-white to-white px-6 py-5 dark:from-violet-500/10 dark:via-[#1a1d21] dark:to-[#1a1d21]">
        <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
          <div className="space-y-3">
            <div className="flex flex-wrap items-center gap-2">
              <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-violet-500/10 text-violet-600 dark:text-violet-300">
                <Sparkles className="h-5 w-5" />
              </div>
              <div>
                <div className="flex flex-wrap items-center gap-2">
                  <h1 className="text-xl font-black tracking-tight">{app?.title ?? "Business App"}</h1>
                  <Badge variant="secondary" className="text-[10px] font-black uppercase tracking-widest">
                    {app?.primary_entity ?? "sales_order"}
                  </Badge>
                  <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
                    AI-native metadata only
                  </Badge>
                </div>
                <p className="mt-1 max-w-3xl text-sm text-muted-foreground">
                  {app?.description ?? "One reusable multiview business workspace for routeable search, pipeline, calendar, and stats surfaces."}
                </p>
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
              <span className="inline-flex items-center gap-1 rounded-full bg-muted px-2.5 py-1 font-semibold">
                <Bot className="h-3 w-3" />
                Summarize order
              </span>
              <span className="inline-flex items-center gap-1 rounded-full bg-muted px-2.5 py-1 font-semibold">
                <Bot className="h-3 w-3" />
                Draft follow-up
              </span>
              <span className="inline-flex items-center gap-1 rounded-full bg-muted px-2.5 py-1 font-semibold">
                <Bot className="h-3 w-3" />
                Explain risk
              </span>
            </div>
          </div>

          <BusinessViewSwitcher mode={mode} availableModes={availableModes} onModeChange={onModeChange} />
        </div>
      </div>

      <div className="border-b bg-muted/20 px-6 py-4">
        <form className="flex flex-col gap-3" onSubmit={handleSubmit}>
          <div className="flex flex-col gap-2 xl:flex-row xl:items-center">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-violet-500" />
              <Input
                value={searchValue}
                onChange={(event) => onSearchValueChange(event.target.value)}
                className="pl-9"
                placeholder="Search orders, customers, tags, or summaries..."
              />
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Button type="submit" size="sm" className="h-9 px-3 text-[11px] font-black uppercase tracking-widest">
                Search
              </Button>
              <Button type="button" variant="outline" size="sm" className="h-9 px-3 text-[11px] font-black uppercase tracking-widest" onClick={onSearchClear}>
                Clear
              </Button>
            </div>
          </div>

          <div className="flex flex-col gap-2 lg:flex-row lg:flex-wrap lg:items-center">
            <div className="flex items-center gap-2 text-[11px] font-black uppercase tracking-widest text-muted-foreground">
              <Filter className="h-3.5 w-3.5" />
              Filters
            </div>
            <select
              className="h-9 rounded-md border bg-background px-3 text-sm"
              value={stage ?? ""}
              onChange={(event) => onFiltersChange({ stage: event.target.value || undefined })}
            >
              <option value="">All stages</option>
              {stageOptions.map((option) => (
                <option key={option.value} value={option.value}>{option.label}</option>
              ))}
            </select>
            <select
              className="h-9 rounded-md border bg-background px-3 text-sm"
              value={status ?? ""}
              onChange={(event) => onFiltersChange({ status: event.target.value || undefined })}
            >
              <option value="">All statuses</option>
              {statusOptions.map((option) => (
                <option key={option.value} value={option.value}>{option.label}</option>
              ))}
            </select>
            <Input
              type="date"
              className="h-9 w-full lg:w-[180px]"
              value={dateFrom ?? ""}
              onChange={(event) => onFiltersChange({ date_from: event.target.value || undefined })}
            />
            <Input
              type="date"
              className="h-9 w-full lg:w-[180px]"
              value={dateTo ?? ""}
              onChange={(event) => onFiltersChange({ date_to: event.target.value || undefined })}
            />
            {modeControls ? <div className="flex flex-wrap items-center gap-2">{modeControls}</div> : null}
          </div>
        </form>

        {summary && <div className="mt-3">{summary}</div>}
      </div>

      <div className="min-h-0 flex-1 overflow-auto px-6 py-5">{children}</div>
    </div>
  )
}
