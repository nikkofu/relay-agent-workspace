"use client"

import { Home, MessageSquare, Bell, Bookmark, Plus, Sparkles, Users, Folder, Zap, Settings } from "lucide-react"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { useUIStore } from "@/stores/ui-store"
import { useUserStore } from "@/stores/user-store"
import { useDMStore } from "@/stores/dm-store"
import { useChannelStore } from "@/stores/channel-store"
import { usePathname, useRouter } from "next/navigation"
import { useState, useEffect } from "react"

const NAV_ITEMS = [
  { icon: Home, label: "Home", href: "/workspace" },
  { icon: MessageSquare, label: "DMs", href: "/workspace/dms" },
  { icon: Bell, label: "Activity", href: "/workspace/activity" },
  { icon: Users, label: "People", href: "/workspace/people" },
  { icon: Folder, label: "Files", href: "/workspace/files" },
  { icon: Zap, label: "Workflows", href: "/workspace/workflows" },
  { icon: Bookmark, label: "Later", href: "/workspace/later" },
  { icon: Settings, label: "Settings", href: "/workspace/settings" },
]

export function PrimaryNav() {
  const { currentWorkspace } = useWorkspaceStore()
  const { toggleAIPanel, isAIPanelOpen } = useUIStore()
  const { currentUser } = useUserStore()
  const { conversations } = useDMStore()
  const { setCurrentChannel } = useChannelStore()
  const router = useRouter()
  const pathname = usePathname()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  const handleNavItemClick = (item: typeof NAV_ITEMS[0]) => {
    if (item.label === "Home") {
      setCurrentChannel(null)
    }
    router.push(item.href)
  }

  if (!mounted) {
    return (
      <aside className="w-[60px] bg-[#3f0e40] dark:bg-[#1a1d21] flex flex-col items-center py-3 gap-4 border-r border-white/10 shrink-0 relative z-20" />
    )
  }

  return (
    <aside className="w-[60px] bg-[#3f0e40] dark:bg-[#1a1d21] flex flex-col items-center py-3 gap-4 border-r border-white/10 shrink-0 relative z-20">
      {/* Workspace Switcher Button */}
      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="ghost" size="icon" className="w-9 h-9 p-0 hover:bg-white/10">
            <div className="w-8 h-8 rounded-lg bg-white/20 flex items-center justify-center text-white font-bold text-sm text-foreground">
              {currentWorkspace?.logo}
            </div>
          </Button>
        </TooltipTrigger>
        <TooltipContent side="right">
          <p>{currentWorkspace?.name}</p>
        </TooltipContent>
      </Tooltip>

      <div className="flex flex-col gap-1 w-full items-center">
        {NAV_ITEMS.map((item, idx) => {
          const isActive = pathname === item.href
          let unreadCount = 0
          if (item.label === "DMs") {
            unreadCount = conversations.reduce((acc, c) => acc + (c.unreadCount || 0), 0)
          }

          return (
            <Tooltip key={idx}>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className={cn(
                    "w-9 h-9 rounded-lg text-white/70 hover:bg-white/10 hover:text-white transition-colors relative",
                    isActive && "bg-white/10 text-white"
                  )}
                  onClick={() => handleNavItemClick(item)}
                >
                  <item.icon className="w-5 h-5" />
                  {unreadCount > 0 && (
                    <span className="absolute -top-1 -right-1 w-4 h-4 bg-[#e01e5a] text-white text-[10px] rounded-full flex items-center justify-center border-2 border-[#3f0e40] dark:border-[#1a1d21]">
                      {unreadCount}
                    </span>
                  )}
                </Button>
              </TooltipTrigger>
              <TooltipContent side="right">
                <p>{item.label}</p>
              </TooltipContent>
            </Tooltip>
          )
        })}
      </div>

      <div className="mt-auto flex flex-col gap-4 items-center">
        {/* AI Assistant Entrance */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className={cn(
                "w-9 h-9 rounded-lg text-white/70 hover:bg-white/10 hover:text-white transition-colors",
                isAIPanelOpen && "bg-white/10 text-white ring-1 ring-purple-400/50"
              )}
              onClick={toggleAIPanel}
            >
              <Sparkles className={cn("w-5 h-5", isAIPanelOpen ? "text-purple-300" : "text-purple-400")} />
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
            <div className="relative cursor-pointer">
              <Avatar className="w-8 h-8 border border-white/10">
                <AvatarImage src={currentUser?.avatar} />
                <AvatarFallback>{currentUser?.name?.substring(0, 2).toUpperCase() || "NF"}</AvatarFallback>
              </Avatar>
              <div 
                className={cn(
                  "absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-[#3f0e40] dark:border-[#1a1d21]",
                  currentUser?.status === "online" && "bg-green-500",
                  currentUser?.status === "away" && "bg-amber-500",
                  currentUser?.status === "busy" && "bg-red-500",
                  currentUser?.status === "offline" && "bg-slate-500"
                )}
              />
            </div>
          </TooltipTrigger>
          <TooltipContent side="right">
            <p>{currentUser?.name} ({currentUser?.status || "offline"})</p>
          </TooltipContent>
        </Tooltip>
      </div>
    </aside>
  )
}
