"use client"

import { PrimaryNav } from "@/components/layout/primary-nav"
import { ChannelSidebar } from "@/components/layout/channel-sidebar"
import { ThreadPanel } from "@/components/layout/thread-panel"
import { AIChatPanel } from "@/components/ai-chat/ai-chat-panel"
import { CanvasPanel } from "@/components/layout/canvas-panel"
import { SearchDialog } from "@/components/search/search-dialog"
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable"
import { useUIStore } from "@/stores/ui-store"
import { useEffect } from "react"

export default function WorkspaceLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const { isThreadOpen, isAIPanelOpen, isCanvasOpen, closeThread, closeAIPanel, closeCanvas } = useUIStore()
  const showRightPanel = isThreadOpen || isAIPanelOpen || isCanvasOpen

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
      <PrimaryNav />
      <div className="flex-1 flex overflow-hidden rounded-tl-lg bg-white dark:bg-[#1a1d21]">
        <ChannelSidebar />
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
      </div>
      <SearchDialog />
    </div>
  )
}
