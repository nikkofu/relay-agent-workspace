"use client"

import type { ElementType } from "react"
import { BarChart3, CalendarDays, Columns3, LayoutGrid, TableProperties } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { getBusinessAppModeLabel } from "@/lib/business-apps"
import type { BusinessAppMode } from "@/types"

interface BusinessViewSwitcherProps {
  mode: BusinessAppMode
  availableModes: BusinessAppMode[]
  onModeChange: (mode: BusinessAppMode) => void
  className?: string
}

const MODE_ICONS: Record<BusinessAppMode, ElementType> = {
  list: TableProperties,
  card_grid: LayoutGrid,
  calendar: CalendarDays,
  kanban: Columns3,
  stat: BarChart3,
}

export function BusinessViewSwitcher({ mode, availableModes, onModeChange, className }: BusinessViewSwitcherProps) {
  return (
    <div className={cn("flex flex-wrap items-center gap-1.5", className)}>
      {availableModes.map((item) => {
        const Icon = MODE_ICONS[item]
        const active = item === mode
        return (
          <Button
            key={item}
            type="button"
            variant={active ? "secondary" : "ghost"}
            size="sm"
            className={cn(
              "h-8 gap-1.5 rounded-full px-3 text-[11px] font-black uppercase tracking-widest",
              active && "ring-1 ring-violet-400/50"
            )}
            onClick={() => onModeChange(item)}
          >
            <Icon className="h-3.5 w-3.5" />
            {getBusinessAppModeLabel(item)}
          </Button>
        )
      })}
    </div>
  )
}
