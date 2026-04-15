"use client"

import { Home, MessageSquare, Bell, Bookmark, MoreHorizontal, Plus, Sparkles } from "lucide-react"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"
import { useWorkspaceStore } from "@/stores/workspace-store"

const NAV_ITEMS = [
  { icon: Home, label: "Home", active: true },
  { icon: MessageSquare, label: "DMs", count: 2 },
  { icon: Bell, label: "Activity" },
  { icon: Bookmark, label: "Later" },
  { icon: MoreHorizontal, label: "More" },
]

export function PrimaryNav() {
  const { workspaces, currentWorkspace } = useWorkspaceStore()

  return (
    <aside className="w-[60px] bg-[#3f0e40] dark:bg-[#1a1d21] flex flex-col items-center py-3 gap-4 border-r border-white/10 shrink-0">
      {/* Workspace Switcher Button */}
      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="ghost" size="icon" className="w-9 h-9 p-0 hover:bg-white/10">
            <div className="w-8 h-8 rounded-lg bg-white/20 flex items-center justify-center text-white font-bold text-sm">
              {currentWorkspace?.logo}
            </div>
          </Button>
        </TooltipTrigger>
        <TooltipContent side="right">
          <p>{currentWorkspace?.name}</p>
        </TooltipContent>
      </Tooltip>

      <div className="flex flex-col gap-1 w-full items-center">
        {NAV_ITEMS.map((item, idx) => (
          <Tooltip key={idx}>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className={cn(
                  "w-9 h-9 rounded-lg text-white/70 hover:bg-white/10 hover:text-white transition-colors relative",
                  item.active && "bg-white/10 text-white"
                )}
              >
                <item.icon className="w-5 h-5" />
                {item.count && (
                  <span className="absolute -top-1 -right-1 w-4 h-4 bg-[#e01e5a] text-white text-[10px] rounded-full flex items-center justify-center border-2 border-[#3f0e40] dark:border-[#1a1d21]">
                    {item.count}
                  </span>
                )}
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">
              <p>{item.label}</p>
            </TooltipContent>
          </Tooltip>
        ))}
      </div>

      <div className="mt-auto flex flex-col gap-4 items-center">
        {/* AI Assistant Entrance */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="w-9 h-9 rounded-lg text-white/70 hover:bg-white/10 hover:text-white transition-colors"
            >
              <Sparkles className="w-5 h-5 text-purple-400" />
            </Button>
          </TooltipTrigger>
          <TooltipContent side="right">
            <p>AI Assistant</p>
          </TooltipContent>
        </Tooltip>

        <Button variant="ghost" size="icon" className="w-9 h-9 rounded-lg text-white/70 hover:bg-white/10">
          <Plus className="w-5 h-5" />
        </Button>

        <Tooltip>
          <TooltipTrigger asChild>
            <Avatar className="w-8 h-8 cursor-pointer border border-white/10">
              <AvatarImage src="https://github.com/nikkofu.png" />
              <AvatarFallback>NF</AvatarFallback>
            </Avatar>
          </TooltipTrigger>
          <TooltipContent side="right">
            <p>Nikko Fu (Online)</p>
          </TooltipContent>
        </Tooltip>
      </div>
    </aside>
  )
}
