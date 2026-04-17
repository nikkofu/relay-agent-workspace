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

interface AIProviderConfig {
  id: string
  models: string[]
}

interface AISettingsProps {
  provider: string
  setProvider: (val: string) => void
  mode: "fast" | "planning"
  setMode: (val: "fast" | "planning") => void
  model: string
  setModel: (val: string) => void
  availableProviders: AIProviderConfig[]
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
  const currentProviderConfig = availableProviders.find(p => p.id === provider)
  const models = currentProviderConfig?.models || []

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="h-8 w-8 text-white/70 hover:text-white hover:bg-white/10 rounded-full transition-all">
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
                  <Button variant="outline" size="sm" className="h-7 px-2 text-[11px] capitalize min-w-[100px] justify-between">
                    {provider}
                    <ChevronRight className="ml-1 h-3 w-3 rotate-90 opacity-50" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-40">
                  {availableProviders.map(p => (
                    <DropdownMenuItem 
                      key={p.id} 
                      onClick={() => setProvider(p.id)}
                      className="capitalize flex items-center justify-between"
                    >
                      {p.id}
                      {provider === p.id && <Check className="h-3.5 w-3.5 text-purple-500" />}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            {/* Model Selection */}
            <div className="flex items-center justify-between">
              <Label htmlFor="model" className="text-xs font-medium">Model</Label>
              <DropdownMenu>
                <DropdownMenuTrigger asChild disabled={models.length === 0}>
                  <Button variant="outline" size="sm" className="h-7 px-2 text-[11px] min-w-[100px] justify-between">
                    {model || (models.length > 0 ? "Select model" : "No models")}
                    <ChevronRight className="ml-1 h-3 w-3 rotate-90 opacity-50" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-56">
                  {models.map(m => (
                    <DropdownMenuItem 
                      key={m} 
                      onClick={() => setModel(m)}
                      className="text-[11px] flex items-center justify-between"
                    >
                      {m}
                      {model === m && <Check className="h-3.5 w-3.5 text-purple-500" />}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            {/* Planning Mode Toggle */}
            <div className="flex items-center justify-between pt-1 border-t border-border/50 mt-1">
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
