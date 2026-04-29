"use client"

import { Hash, Lock, ChevronDown, Plus, Search, Sparkles, Settings, UserPlus, Pencil } from "lucide-react"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import { cn } from "@/lib/utils"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { useChannelStore } from "@/stores/channel-store"
import { useDMStore } from "@/stores/dm-store"
import { isAIUserLike, useUserStore } from "@/stores/user-store"
import { useUIStore } from "@/stores/ui-store"
import { HuddleBar } from "@/components/huddle/huddle-bar"
import { useState, useEffect } from "react"
import { Badge } from "@/components/ui/badge"
import { useRouter } from "next/navigation"
import type { Channel, DirectMessage } from "@/types"

import { CreateChannelDialog } from "@/components/channel/create-channel-dialog"
import { InviteMemberDialog } from "@/components/workspace/invite-member-dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

export function ChannelSidebar() {
  const { currentWorkspace } = useWorkspaceStore()
  const { channels, currentChannel, setCurrentChannel } = useChannelStore()
  const { conversations, fetchConversations, currentConversation, setCurrentConversation } = useDMStore()
  const { users, currentUser } = useUserStore()
  const { openSearch, openDockedChat } = useUIStore()
  const [openSections, setOpenSections] = useState({ starred: true, channels: true, dms: true })
  const [mounted, setMounted] = useState(false)
  const [showCreateChannel, setShowCreateChannel] = useState(false)
  const [showInviteMember, setShowInviteMember] = useState(false)
  const router = useRouter()

  useEffect(() => {
    setMounted(true)
    fetchConversations()
  }, [fetchConversations])

  const starredChannels = channels.filter(c => c.isStarred)
  const regularChannels = channels.filter(c => !c.isStarred)

  const handleChannelClick = (channel: Channel) => {
    setCurrentChannel(channel)
    setCurrentConversation(null)
    router.push(`/workspace?c=${channel.id}`)
  }

  // Clicking a DM in the nav now navigates to the full-page DM view
  // (`/workspace/dms/:dmId`) instead of opening a small docked chat. This
  // keeps the experience consistent with the new WhatsApp-style two-pane
  // DMs surface — request #4.
  const handleDMClick = (conv: DirectMessage) => {
    setCurrentConversation(conv)
    if (conv.id) {
      router.push(`/workspace/dms/${conv.id}`)
    } else {
      // Fallback for placeholder conversations that don't have an id yet.
      const otherUserId = conv.userIds?.find(id => id !== currentUser?.id)
      if (otherUserId) openDockedChat(otherUserId)
    }
  }

  if (!mounted) {
    return (
      <nav className="w-[260px] bg-[#3f0e40] dark:bg-[#19171d] flex flex-col h-full text-[#cfc3cf] dark:text-[#9b999b] shrink-0 border-r border-white/5 shadow-xl relative z-10" />
    )
  }

  return (
    <nav className="w-[260px] bg-[#3f0e40] dark:bg-[#19171d] flex flex-col h-full text-[#cfc3cf] dark:text-[#9b999b] shrink-0 border-r border-white/5 shadow-xl relative z-10">
      {/* Workspace Header */}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <div className="h-14 px-4 flex items-center justify-between hover:bg-white/10 cursor-pointer transition-colors border-b border-white/5">
            <h2 className="font-bold text-lg text-white truncate flex items-center gap-1">
              {mounted ? (currentWorkspace?.name || "Workspace") : "Workspace"}
              <ChevronDown className="w-4 h-4" />
            </h2>
            <div
              className="bg-white rounded-full p-1.5 shrink-0 hover:bg-purple-100 transition-colors"
              onClick={(e) => { e.stopPropagation(); setShowCreateChannel(true) }}
              title="New message / channel"
            >
              <Plus className="w-4 h-4 text-[#3f0e40]" />
            </div>
          </div>
        </DropdownMenuTrigger>
        <DropdownMenuContent side="bottom" align="start" className="w-56">
          <DropdownMenuItem onClick={() => router.push('/workspace/settings')}>
            <Settings className="w-4 h-4 mr-2" /> Workspace settings
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => setShowInviteMember(true)}>
            <UserPlus className="w-4 h-4 mr-2" /> Invite people
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={() => setShowCreateChannel(true)}>
            <Pencil className="w-4 h-4 mr-2" /> Create a channel
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => router.push('/workspace/dms')}>
            <Sparkles className="w-4 h-4 mr-2" /> Browse direct messages
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <ScrollArea className="flex-1">
        <div className="px-2 py-4 flex flex-col gap-4">
          {/* Quick Search */}
          <div className="px-2 mb-2">
            <Button 
              variant="ghost" 
              className="w-full justify-start text-white/50 bg-white/10 hover:bg-white/20 h-8 px-2 font-normal rounded-md"
              onClick={openSearch}
            >
              <Search className="w-4 h-4 mr-2" />
              Search Relay
              <span className="ml-auto text-[10px] opacity-50">⌘K</span>
            </Button>
          </div>

          {/* Starred Section */}
          <Collapsible open={openSections.starred} onOpenChange={(open) => setOpenSections(s => ({ ...s, starred: open }))}>
            <div className="group flex items-center px-2 py-1 hover:bg-white/10 rounded cursor-pointer mb-1" onClick={() => router.push('/workspace/starred')}>
              <CollapsibleTrigger asChild onClick={(e) => e.stopPropagation()}>
                <ChevronDown className={cn("w-4 h-4 mr-1 transition-transform", !openSections.starred && "-rotate-90")} />
              </CollapsibleTrigger>
              <span className="text-sm font-semibold uppercase tracking-wider text-[11px] flex-1">Starred</span>
            </div>
            <CollapsibleContent className="flex flex-col gap-[2px]">
              {starredChannels.map(channel => (
                <SidebarItem 
                  key={channel.id} 
                  channel={channel} 
                  isActive={currentChannel?.id === channel.id}
                  onClick={() => handleChannelClick(channel)}
                />
              ))}
            </CollapsibleContent>
          </Collapsible>

          {/* Channels Section */}
          <Collapsible open={openSections.channels} onOpenChange={(open) => setOpenSections(s => ({ ...s, channels: open }))}>
            <div className="group flex items-center px-2 py-1 hover:bg-white/10 rounded cursor-pointer mb-1">
              <CollapsibleTrigger asChild>
                <ChevronDown className={cn("w-4 h-4 mr-1 transition-transform", !openSections.channels && "-rotate-90")} />
              </CollapsibleTrigger>
              <span className="text-sm font-semibold uppercase tracking-wider text-[11px] flex-1">Channels</span>
              <Button 
                variant="ghost" 
                size="icon" 
                className="h-6 w-6 text-[#cfc3cf] hover:text-white hover:bg-white/10 opacity-0 group-hover:opacity-100 transition-all"
                onClick={() => setShowCreateChannel(true)}
              >
                <Plus className="w-4 h-4" />
              </Button>
            </div>
            <CollapsibleContent className="flex flex-col gap-[2px]">
              {regularChannels.map(channel => (
                <SidebarItem 
                  key={channel.id} 
                  channel={channel} 
                  isActive={currentChannel?.id === channel.id}
                  onClick={() => handleChannelClick(channel)}
                />
              ))}
              <div 
                className="flex items-center px-2 py-1.5 hover:bg-white/10 rounded cursor-pointer text-sm gap-2"
                onClick={() => setShowCreateChannel(true)}
              >
                <div className="bg-white/10 rounded p-0.5">
                  <Plus className="w-3.5 h-3.5" />
                </div>
                <span>Add channels</span>
              </div>
            </CollapsibleContent>
          </Collapsible>

          {/* Direct Messages Section */}
          <Collapsible open={openSections.dms} onOpenChange={(open) => setOpenSections(s => ({ ...s, dms: open }))}>
            <div className="group flex items-center px-2 py-1 hover:bg-white/10 rounded cursor-pointer mb-1">
              <CollapsibleTrigger asChild>
                <ChevronDown className={cn("w-4 h-4 mr-1 transition-transform", !openSections.dms && "-rotate-90")} />
              </CollapsibleTrigger>
              <span className="text-sm font-semibold uppercase tracking-wider text-[11px] flex-1">Direct Messages</span>
              <Plus className="w-4 h-4 opacity-0 group-hover:opacity-100 transition-opacity" />
            </div>
            <CollapsibleContent className="flex flex-col gap-[2px]">
              {conversations.map(conv => {
                const otherUserId = conv.userIds?.find(id => id !== currentUser?.id)
                if (!otherUserId) return null
                
                const user = users.find(u => u.id === otherUserId)
                if (!user) return null

                return (
                  <div 
                    key={conv.id} 
                    onClick={() => handleDMClick(conv)}
                    className={cn(
                      "flex items-center px-2 py-1.5 hover:bg-[#1164a3] hover:text-white rounded cursor-pointer text-sm gap-2 transition-colors",
                      currentConversation?.id === conv.id && "bg-[#1164a3] text-white"
                    )}
                  >
                    <div className="relative shrink-0">
                      <div className={cn(
                        "w-5 h-5 rounded flex items-center justify-center",
                        user.status === 'online' && "bg-green-500/20",
                        user.status === 'away' && "bg-amber-500/20",
                        user.status === 'busy' && "bg-red-500/20",
                        user.status === 'offline' && "bg-white/10"
                      )}>
                        {isAIUserLike(user) ? (
                          <Sparkles className="w-3 h-3 text-purple-400" />
                        ) : (
                          <div className={cn(
                            "w-2 h-2 rounded-full", 
                            user.status === 'online' && "bg-green-500",
                            user.status === 'away' && "bg-amber-500",
                            user.status === 'busy' && "bg-red-500",
                            user.status === 'offline' && "bg-white/30"
                          )} />
                        )}
                      </div>
                    </div>
                    <span className="flex-1 truncate">{user.name}</span>
                  </div>
                )
              })}
              <div 
                onClick={() => setShowInviteMember(true)}
                className="flex items-center px-2 py-1.5 hover:bg-white/10 rounded cursor-pointer text-sm gap-2 opacity-70"
              >
                <div className="bg-white/10 rounded p-0.5">
                  <Plus className="w-3.5 h-3.5" />
                </div>
                <span>Add teammates</span>
              </div>
            </CollapsibleContent>
          </Collapsible>
        </div>
      </ScrollArea>
      <HuddleBar />

      <CreateChannelDialog open={showCreateChannel} onOpenChange={setShowCreateChannel} />
      <InviteMemberDialog open={showInviteMember} onOpenChange={setShowInviteMember} />
    </nav>
  )
}

interface SidebarChannel {
  name: string
  type: string
  unreadCount?: number
  id: string
}

function SidebarItem({ channel, isActive, onClick }: { channel: SidebarChannel, isActive: boolean, onClick: () => void }) {
  const Icon = channel.type === "private" ? Lock : Hash

  return (
    <div 
      className={cn(
        "flex items-center px-2 py-1.5 hover:bg-[#1164a3] hover:text-white rounded cursor-pointer text-sm gap-2 transition-colors",
        isActive && "bg-[#1164a3] text-white"
      )}
      onClick={onClick}
    >
      <Icon className={cn("w-4 h-4 shrink-0 opacity-70", isActive && "opacity-100")} />
      <span className={cn("flex-1 truncate font-medium", channel.unreadCount && "font-bold text-white")}>
        {channel.name}
      </span>
      {channel.unreadCount && channel.unreadCount > 0 ? (
        <Badge className="bg-[#e01e5a] text-white hover:bg-[#e01e5a] h-5 px-1.5 min-w-5 justify-center border-none text-[10px]">
          {channel.unreadCount}
        </Badge>
      ) : null}
    </div>
  )
}
