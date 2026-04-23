"use client"

import { PrimaryNav } from "@/components/layout/primary-nav"
import { ChannelSidebar } from "@/components/layout/channel-sidebar"
import { ThreadPanel } from "@/components/layout/thread-panel"
import { AIChatPanel } from "@/components/ai-chat/ai-chat-panel"
import { CanvasPanel } from "@/components/layout/canvas-panel"
import { SearchDialog } from "@/components/search/search-dialog"
import { DockedChatContainer } from "@/components/dm/docked-chat-container"
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable"
import { useUIStore } from "@/stores/ui-store"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { useChannelStore } from "@/stores/channel-store"
import { useUserStore } from "@/stores/user-store"
import { usePresenceStore } from "@/stores/presence-store"
import { useEffect, useRef, useState, Suspense } from "react"
import { useWebsocket } from "@/hooks/use-websocket"
import { useDraftStore } from "@/stores/draft-store"

export default function WorkspaceLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const { isThreadOpen, isAIPanelOpen, isCanvasOpen, closeThread, closeAIPanel, closeCanvas } = useUIStore()
  const { fetchWorkspaces, currentWorkspace } = useWorkspaceStore()
  const { fetchChannels, currentChannel } = useChannelStore()
  const { fetchMe, fetchUsers } = useUserStore()
  const { fetchPresence, sendHeartbeat, fetchScopedPresence } = usePresenceStore()
  const { fetchDrafts } = useDraftStore()
  const [mounted, setMounted] = useState(false)
  const showRightPanel = isThreadOpen || isAIPanelOpen || isCanvasOpen
  
  useWebsocket()
  
  const workspacesFetched = useRef(false)
  const lastChannelsWorkspaceId = useRef<string | null>(null)

  const heartbeatIntervalRef = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    setMounted(true)
    if (!workspacesFetched.current) {
      workspacesFetched.current = true
      fetchMe()
      fetchUsers()
      fetchWorkspaces()
      fetchPresence()
      fetchDrafts()
    }
  }, [fetchMe, fetchUsers, fetchWorkspaces, fetchPresence, fetchDrafts])

  useEffect(() => {
    if (heartbeatIntervalRef.current) clearInterval(heartbeatIntervalRef.current)
    heartbeatIntervalRef.current = setInterval(() => sendHeartbeat(), 30000)
    return () => {
      if (heartbeatIntervalRef.current) {
        clearInterval(heartbeatIntervalRef.current)
        heartbeatIntervalRef.current = null
      }
    }
  }, [sendHeartbeat])

  useEffect(() => {
    if (currentWorkspace && currentWorkspace.id !== lastChannelsWorkspaceId.current) {
      lastChannelsWorkspaceId.current = currentWorkspace.id
      fetchChannels(currentWorkspace.id)
    }
  }, [currentWorkspace, fetchChannels])

  useEffect(() => {
    if (currentChannel) {
      fetchScopedPresence(currentChannel.id)
    }
  }, [currentChannel, fetchScopedPresence])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        if (isAIPanelOpen) closeAIPanel()
        if (isThreadOpen) closeThread()
        if (isCanvasOpen) closeCanvas()
      }
    }
    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [isAIPanelOpen, isThreadOpen, isCanvasOpen, closeAIPanel, closeThread, closeCanvas])

  return (
    <div className="flex h-screen w-full overflow-hidden bg-[#3f0e40] dark:bg-[#1a1d21]">
      <Suspense fallback={<aside className="w-[60px] bg-[#3f0e40] dark:bg-[#1a1d21] shrink-0" />}>
        <PrimaryNav />
      </Suspense>
      <div className="flex-1 flex overflow-hidden rounded-tl-lg bg-white dark:bg-[#1a1d21]">
        <Suspense fallback={<nav className="w-[260px] bg-[#3f0e40] dark:bg-[#19171d] shrink-0" />}>
          <ChannelSidebar />
        </Suspense>
        {mounted ? (
          <ResizablePanelGroup direction="horizontal" className="flex-1">
            <ResizablePanel defaultSize={showRightPanel ? (isCanvasOpen ? 50 : 65) : 100} minSize={30}>
              {children}
            </ResizablePanel>
            
            {showRightPanel && (
              <>
                <ResizableHandle withHandle />
                <ResizablePanel defaultSize={isCanvasOpen ? 50 : 35} minSize={25}>
                  {isThreadOpen && <ThreadPanel />}
                  {isAIPanelOpen && <AIChatPanel />}
                  {isCanvasOpen && <CanvasPanel />}
                </ResizablePanel>
              </>
            )}
          </ResizablePanelGroup>
        ) : (
          <div className="flex-1 overflow-hidden">
            {children}
          </div>
        )}
      </div>
      <SearchDialog />
      <DockedChatContainer />
    </div>
  )
}
