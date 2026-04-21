"use client"

import { useChannelStore } from "@/stores/channel-store"
import { useMessageStore } from "@/stores/message-store"
import { MessageList } from "@/components/message/message-list"
import { MessageArea } from "@/components/layout/message-area"
import { HomeDashboard } from "@/components/layout/home-dashboard"
import { AgentCollabPage } from "@/components/agent-collab/agent-collab-page"
import { useSearchParams } from "next/navigation"
import { Suspense, useEffect } from "react"
import { useArtifactStore } from "@/stores/artifact-store"
import { useUIStore } from "@/stores/ui-store"
import { FileCode, FileText, MoreVertical, Copy, ExternalLink } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

function WorkspaceContent() {
  const { currentChannel, setCurrentChannelById, setCurrentChannel, channels } = useChannelStore()
  const { messages } = useMessageStore()
  const { artifacts, fetchArtifacts, duplicateArtifact } = useArtifactStore()
  const { openCanvas } = useUIStore()
  const searchParams = useSearchParams()
  const channelIdFromUrl = searchParams.get("c")

  useEffect(() => {
    if (channelIdFromUrl) {
      if (channels.length > 0) {
        setCurrentChannelById(channelIdFromUrl)
      }
    } else {
      // If no channel in URL, we should be on the home dashboard
      setCurrentChannel(null)
    }
  }, [channelIdFromUrl, channels, setCurrentChannelById, setCurrentChannel])

  useEffect(() => {
    if (currentChannel) {
      fetchArtifacts(currentChannel.id)
    }
  }, [currentChannel, fetchArtifacts])

  if (!currentChannel) {
    return <HomeDashboard />
  }

  if (currentChannel.name === 'agent-collab') {
    return <AgentCollabPage />
  }

  const channelMessages = messages.filter(m => m.channelId === currentChannel.id && !m.threadId)

  return (
    <MessageArea>
      <div className="flex flex-col h-full">
        {/* Channel Introduction & Artifacts */}
        <div className="p-4 border-b bg-muted/10">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded bg-muted flex items-center justify-center font-bold text-foreground border shadow-sm">
                #
              </div>
              <div>
                <h1 className="text-sm font-bold text-foreground leading-tight">Welcome to #{currentChannel.name}</h1>
                <p className="text-xs text-muted-foreground truncate max-w-md mt-0.5">
                  {currentChannel.description || `This is the start of the #${currentChannel.name} channel.`}
                </p>
              </div>
            </div>
            
            {artifacts.length > 0 && (
              <div className="flex items-center gap-2">
                {artifacts.slice(0, 3).map(a => (
                  <div key={a.id} className="flex items-center">
                    <Button 
                      variant="outline" 
                      size="sm" 
                      className="h-8 text-[10px] font-bold uppercase tracking-wider gap-1.5 border-blue-500/20 bg-blue-500/5 hover:bg-blue-500/10 hover:border-blue-500/40 transition-all text-blue-600 dark:text-blue-400 shadow-xs rounded-r-none border-r-0"
                      onClick={() => openCanvas(a.id)}
                    >
                      {a.type === 'code' ? <FileCode className="w-3 h-3" /> : <FileText className="w-3 h-3" />}
                      {a.title}
                    </Button>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button 
                          variant="outline" 
                          size="icon" 
                          className="h-8 w-6 rounded-l-none border-blue-500/20 bg-blue-500/5 hover:bg-blue-500/10 hover:border-blue-500/40 transition-all text-blue-600 dark:text-blue-400 shadow-xs"
                        >
                          <MoreVertical className="w-3 h-3" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => openCanvas(a.id)}>
                          <ExternalLink className="w-3.5 h-3.5 mr-2" />
                          Open Canvas
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => duplicateArtifact(a.id)}>
                          <Copy className="w-3.5 h-3.5 mr-2" />
                          Duplicate
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Messages List */}
        <div className="flex-1 min-h-0 flex flex-col">
          <MessageList messages={channelMessages} className="flex-1" />
        </div>
      </div>
    </MessageArea>
  )
}

export default function WorkspacePage() {
  return (
    <Suspense fallback={<div className="flex-1 flex items-center justify-center bg-white dark:bg-[#1a1d21] text-muted-foreground">Loading workspace...</div>}>
      <WorkspaceContent />
    </Suspense>
  )
}
