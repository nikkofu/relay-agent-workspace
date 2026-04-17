"use client"

import { Settings, Check, ChevronRight } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

interface AISettingsProps {
  provider: string
  setProvider: (val: string) => void
  mode: "fast" | "planning"
  setMode: (val: "fast" | "planning") => void
  model: string
  setModel: (val: string) => void
  availableProviders: string[]
}

export function AISettings({ 
  provider, 
  setProvider, 
  mode, 
  setMode, 
  model, 
  setModel,
  availableProviders 
}: AISettingsProps) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="h-8 w-8 text-white/70 hover:text-white hover:bg-white/10 rounded-full">
          <Settings className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-4 bg-white dark:bg-[#1a1d21] border-border shadow-2xl" align="end">
        <div className="space-y-4">
          <div className="flex flex-col gap-1.5">
            <h4 className="font-bold text-sm leading-none">AI Configuration</h4>
            <p className="text-[11px] text-muted-foreground">Adjust your AI model and execution mode.</p>
          </div>

          <div className="space-y-3 pt-2">
            {/* Provider Selection */}
            <div className="flex items-center justify-between">
              <Label htmlFor="provider" className="text-xs font-medium">Provider</Label>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="sm" className="h-7 px-2 text-[11px] capitalize">
                    {provider}
                    <ChevronRight className="ml-1 h-3 w-3 rotate-90" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-40">
                  {availableProviders.map(p => (
                    <DropdownMenuItem 
                      key={p} 
                      onClick={() => setProvider(p)}
                      className="capitalize flex items-center justify-between"
                    >
                      {p}
                      {provider === p && <Check className="h-3.5 w-3.5 text-purple-500" />}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            {/* Model Name */}
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="model" className="text-xs font-medium">Model ID</Label>
              <input 
                id="model"
                type="text"
                placeholder="e.g. gemini-3-flash-preview"
                value={model}
                onChange={(e) => setModel(e.target.value)}
                className="flex h-8 w-full rounded-md border border-input bg-transparent px-3 py-1 text-xs shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              />
            </div>

            {/* Planning Mode Toggle */}
            <div className="flex items-center justify-between pt-1">
              <div className="flex flex-col gap-0.5">
                <Label htmlFor="planning-mode" className="text-xs font-medium leading-none">
                  {mode === 'planning' ? 'Planning Mode' : 'Fast Mode'}
                </Label>
                <p className="text-[10px] text-muted-foreground">
                  {mode === 'planning' ? 'Advanced reasoning, slower.' : 'Quick responses, cost-effective.'}
                </p>
              </div>
              <Switch 
                id="planning-mode"
                checked={mode === 'planning'}
                onCheckedChange={(checked) => setMode(checked ? 'planning' : 'fast')}
              />
            </div>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  )
}
