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
  const {
    isThreadOpen, isAIPanelOpen, isCanvasOpen,
    isCanvasEditing, isCanvasMaximized,
    closeThread, closeAIPanel, closeCanvas,
  } = useUIStore()
  const { fetchWorkspaces, currentWorkspace } = useWorkspaceStore()
  const { fetchChannels, currentChannel } = useChannelStore()
  const { fetchMe, fetchUsers } = useUserStore()
  const { fetchPresence, sendHeartbeat, fetchScopedPresence } = usePresenceStore()
  const { fetchDrafts } = useDraftStore()
  const [mounted, setMounted] = useState(false)

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
        {mounted ? (() => {
          // Resolve panel layout mode declaratively so the (left messages /
          // right side panel) split adapts to user intent:
          //
          //   • canvas-max  → 5 / 95   (Maximize button — collapses messages)
          //   • canvas-edit → 33 / 67  (Edit clicked → with inner 50/50 dock
          //                            the result is the 33/33/33 split the
          //                            user asked for in request #10)
          //   • canvas      → 50 / 50  (canvas open, read-only preview)
          //   • side        → 65 / 35  (thread or AI panel)
          //   • none        → 100 / 0  (no right panel)
          //
          // We `key={layoutMode}` the group because react-resizable-panels
          // only consults `defaultSize` on mount; remounting on mode change
          // is the cleanest way to apply new defaults without taking on
          // imperative ref forwarding.
          let layoutMode: "none" | "side" | "canvas" | "canvas-edit" | "canvas-max" = "none"
          let leftSize = 100
          let rightSize = 0
          if (isCanvasOpen) {
            if (isCanvasMaximized) { layoutMode = "canvas-max"; leftSize = 5;  rightSize = 95 }
            else if (isCanvasEditing) { layoutMode = "canvas-edit"; leftSize = 33; rightSize = 67 }
            else { layoutMode = "canvas"; leftSize = 50; rightSize = 50 }
          } else if (isThreadOpen || isAIPanelOpen) {
            layoutMode = "side"; leftSize = 65; rightSize = 35
          }
          const showRight = rightSize > 0
          return (
            <ResizablePanelGroup
              key={layoutMode}
              direction="horizontal"
              className="flex-1"
            >
              <ResizablePanel
                defaultSize={leftSize}
                minSize={isCanvasMaximized ? 4 : 25}
                collapsible={isCanvasMaximized}
              >
                {children}
              </ResizablePanel>

              {showRight && (
                <>
                  <ResizableHandle withHandle />
                  <ResizablePanel defaultSize={rightSize} minSize={25}>
                    {isThreadOpen && <ThreadPanel />}
                    {isAIPanelOpen && <AIChatPanel />}
                    {isCanvasOpen && <CanvasPanel />}
                  </ResizablePanel>
                </>
              )}
            </ResizablePanelGroup>
          )
        })() : (
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
