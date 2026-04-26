"use client"

import { useRouter } from "next/navigation"
import type { LucideIcon } from "lucide-react"
import {
  AlertCircle,
  BarChart3,
  CalendarDays,
  ChevronRight,
  FileSignature,
  Files,
  LayoutGrid,
  ListTodo,
  MessageSquare,
  Search,
  Wrench,
  Workflow,
} from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import type { HomeAppToolItem, WorkspaceView } from "@/types"
import {
  getWorkspaceViewKindLabel,
  getWorkspaceViewSourceLabel,
  resolveHomeAppToolHref,
  resolveWorkspaceViewHref,
} from "@/lib/workspace-views"

interface HomeAppsToolsSectionProps {
  items: HomeAppToolItem[]
  workspaceViews: WorkspaceView[]
  isLoadingWorkspaceViews: boolean
  workspaceViewsError: string | null
}

const iconByType: Record<string, LucideIcon> = {
  list: ListTodo,
  calendar: CalendarDays,
  search: Search,
  report: BarChart3,
  form: FileSignature,
  workflow: Workflow,
  file: Files,
  tool: Wrench,
  channel_messages: MessageSquare,
}

function getIcon(type: string): LucideIcon {
  return iconByType[type] || LayoutGrid
}

export function HomeAppsToolsSection({
  items,
  workspaceViews,
  isLoadingWorkspaceViews,
  workspaceViewsError,
}: HomeAppsToolsSectionProps) {
  const router = useRouter()

  return (
    <div className="rounded-3xl border bg-white dark:bg-[#222529] shadow-sm">
      <div className="px-5 py-4 border-b flex items-center justify-between gap-3">
        <div>
          <h3 className="text-sm font-black tracking-tight">Apps & Tools</h3>
          <p className="text-xs text-muted-foreground mt-1">Compact entry points for current and future workspace surfaces.</p>
        </div>
        <Badge variant="secondary" className="text-[10px] font-black uppercase tracking-widest">Super App</Badge>
      </div>

      <div className="p-5 space-y-5">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {items.map((item) => {
            const Icon = getIcon(item.view_type)
            const href = resolveHomeAppToolHref(item)
            const disabled = !href
            return (
              <button
                key={item.id}
                type="button"
                disabled={disabled}
                onClick={() => href && router.push(href)}
                className="rounded-2xl border px-4 py-3 text-left transition-colors hover:border-violet-300 hover:bg-violet-500/5 disabled:cursor-default disabled:opacity-80"
              >
                <div className="flex items-start gap-3">
                  <div className="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-violet-500/10 text-violet-600">
                    <Icon className="w-4 h-4" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center justify-between gap-2">
                      <p className="text-sm font-bold truncate">{item.title}</p>
                      <ChevronRight className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                    </div>
                    <div className="mt-2 flex flex-wrap items-center gap-2">
                      <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
                        {getWorkspaceViewKindLabel(item.view_type)}
                      </Badge>
                      <span className="text-[10px] text-muted-foreground uppercase tracking-widest font-bold">
                        {disabled ? "Registry only" : "Open page"}
                      </span>
                    </div>
                  </div>
                </div>
              </button>
            )
          })}
        </div>

        <div className="space-y-3">
          <div className="flex items-center justify-between gap-3">
            <div>
              <h4 className="text-xs font-black uppercase tracking-widest text-muted-foreground">Workspace Views</h4>
              <p className="text-xs text-muted-foreground mt-1">Saved registry entries created by people, agents, tools, or the system.</p>
            </div>
            <Button variant="ghost" size="sm" className="text-[10px] font-black uppercase tracking-widest" disabled>
              Metadata only
            </Button>
          </div>

          {isLoadingWorkspaceViews && (
            <div className="rounded-2xl border border-dashed px-4 py-5 text-xs text-muted-foreground">
              Loading workspace views...
            </div>
          )}

          {!isLoadingWorkspaceViews && workspaceViewsError && (
            <div className="rounded-2xl border border-amber-500/20 bg-amber-500/5 px-4 py-3 text-xs text-amber-700 dark:text-amber-300 flex items-start gap-2">
              <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
              <div>
                <p className="font-bold">Workspace views are temporarily unavailable.</p>
                <p className="mt-1">Core Home sections still load normally. You can retry on the next Home refresh.</p>
              </div>
            </div>
          )}

          {!isLoadingWorkspaceViews && !workspaceViewsError && workspaceViews.length === 0 && (
            <div className="rounded-2xl border border-dashed px-4 py-5 text-xs text-muted-foreground">
              No saved workspace views yet.
            </div>
          )}

          {!isLoadingWorkspaceViews && workspaceViews.length > 0 && (
            <div className="space-y-2">
              {workspaceViews.map((view) => {
                const Icon = getIcon(view.view_type)
                const href = resolveWorkspaceViewHref(view)
                return (
                  <button
                    key={view.id}
                    type="button"
                    disabled={!href}
                    onClick={() => href && router.push(href)}
                    className="w-full rounded-2xl border px-4 py-3 text-left transition-colors hover:border-violet-300 hover:bg-violet-500/5 disabled:cursor-default"
                  >
                    <div className="flex items-start gap-3">
                      <div className="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-muted text-muted-foreground">
                        <Icon className="w-4 h-4" />
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center justify-between gap-2">
                          <p className="text-sm font-bold truncate">{view.title}</p>
                          <span className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground shrink-0">
                            {href ? "Open" : "Registry"}
                          </span>
                        </div>
                        <div className="mt-2 flex flex-wrap items-center gap-2">
                          <Badge variant="outline" className="text-[10px] font-black uppercase tracking-widest">
                            {getWorkspaceViewKindLabel(view.view_type)}
                          </Badge>
                          <Badge variant="secondary" className="text-[10px] font-black uppercase tracking-widest">
                            {getWorkspaceViewSourceLabel(view.source)}
                          </Badge>
                          {view.primary_channel_id && (
                            <span className="text-[10px] text-muted-foreground uppercase tracking-widest font-bold truncate max-w-[140px]">
                              channel {view.primary_channel_id}
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                  </button>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
