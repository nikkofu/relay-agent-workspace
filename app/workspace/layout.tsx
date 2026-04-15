"use client"

import { PrimaryNav } from "@/components/layout/primary-nav"
import { ChannelSidebar } from "@/components/layout/channel-sidebar"
import { MessageArea } from "@/components/layout/message-area"
import { ThreadPanel } from "@/components/layout/thread-panel"
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable"
import { useUIStore } from "@/stores/ui-store"

export default function WorkspaceLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const { isThreadOpen } = useUIStore()

  return (
    <div className="flex h-screen w-full overflow-hidden bg-[#3f0e40] dark:bg-[#1a1d21]">
      <PrimaryNav />
      <div className="flex-1 flex overflow-hidden rounded-tl-lg bg-white dark:bg-[#1a1d21]">
        <ChannelSidebar />
        <ResizablePanelGroup direction="horizontal" className="flex-1">
          <ResizablePanel defaultSize={isThreadOpen ? 65 : 100} minSize={30}>
            <MessageArea>{children}</MessageArea>
          </ResizablePanel>
          
          {isThreadOpen && (
            <>
              <ResizableHandle withHandle />
              <ResizablePanel defaultSize={35} minSize={25}>
                <ThreadPanel />
              </ResizablePanel>
            </>
          )}
        </ResizablePanelGroup>
      </div>
    </div>
  )
}
