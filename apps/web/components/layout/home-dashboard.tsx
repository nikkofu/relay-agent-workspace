"use client"

import { useEffect, useState } from "react"
import { 
  Layout, MessageSquare, Plus, UserPlus, Smile, Sparkles, Clock, Inbox
} from "lucide-react"
import { useRouter } from "next/navigation"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { HomeWorkbench } from "@/components/layout/home-workbench"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { useUserStore } from "@/stores/user-store"
import { CreateChannelDialog } from "@/components/channel/create-channel-dialog"
import { InviteMemberDialog } from "@/components/workspace/invite-member-dialog"
import { SetStatusDialog } from "@/components/workspace/set-status-dialog"

export function HomeDashboard() {
  const router = useRouter()
  const {
    homeData,
    fetchHome,
    fetchWorkspaceViews,
    workspaceViews,
    isLoadingWorkspaceViews,
    workspaceViewsError,
    currentWorkspace,
    isHomeExecutionStale,
  } = useWorkspaceStore()
  const { currentUser } = useUserStore()
  const [mounted, setMounted] = useState(false)
  const [showCreateChannel, setShowCreateChannel] = useState(false)
  const [showInviteMember, setShowInviteMember] = useState(false)
  const [showSetStatus, setShowSetStatus] = useState(false)

  useEffect(() => {
    setMounted(true)
    fetchHome()
    fetchWorkspaceViews()
  }, [fetchHome, fetchWorkspaceViews])

  // Phase 67B: Silently refresh Home if execution data is marked stale via WS
  useEffect(() => {
    if (mounted && isHomeExecutionStale) {
      console.log("Home execution data is stale, refreshing...")
      fetchHome()
    }
  }, [mounted, isHomeExecutionStale, fetchHome])

  if (!mounted) return null

  return (
    <div className="h-full w-full flex flex-col bg-white dark:bg-[#1a1d21] overflow-hidden">
      {/* Slack-style Floating Header */}
      <header className="h-14 px-6 flex items-center border-b shrink-0 bg-white/80 dark:bg-[#1a1d21]/80 backdrop-blur-md sticky top-0 z-10">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 rounded bg-purple-600 flex items-center justify-center">
            <Layout className="w-3.5 h-3.5 text-white" />
          </div>
          <h1 className="text-sm font-black tracking-tight uppercase">{currentWorkspace?.name || 'Relay'} Home</h1>
        </div>
      </header>

      <ScrollArea className="flex-1">
        <div className="flex flex-col">
          {/* Hero Welcome Section */}
          <div className="px-8 pt-10 pb-12 bg-gradient-to-br from-purple-600 via-purple-700 to-indigo-800 text-white relative overflow-hidden">
            <div className="absolute top-0 right-0 p-8 opacity-10 pointer-events-none">
              <Sparkles className="w-64 h-64 rotate-12" />
            </div>
            
            <div className="max-w-5xl mx-auto relative z-10">
              <div className="flex flex-col md:flex-row md:items-center justify-between gap-8">
                <div className="space-y-4">
                  <Badge className="bg-white/20 hover:bg-white/30 text-white border-none text-[10px] font-black uppercase tracking-widest px-2 py-0.5">
                    Workspace Overview
                  </Badge>
                  <h2 className="text-4xl font-black tracking-tight leading-tight">
                    Good {new Date().getHours() < 12 ? 'morning' : 'afternoon'},<br />
                    {currentUser?.name || 'Team member'}.
                  </h2>
                  <p className="text-purple-100/80 max-w-lg text-lg font-medium leading-relaxed">
                    Here&apos;s what&apos;s happening in <span className="text-white font-bold">{currentWorkspace?.name || 'Relay'}</span> today. 
                    You have <span className="text-white font-bold">{homeData?.stats?.pending_actions || 0} pending actions</span> to review.
                  </p>
                </div>
                
                <div className="flex flex-col gap-3 min-w-[200px]">
                  <Button onClick={() => setShowCreateChannel(true)} className="bg-white text-purple-700 hover:bg-purple-50 font-bold shadow-xl border-none h-11 justify-start gap-3 px-5">
                    <Plus className="w-4 h-4" /> Create Channel
                  </Button>
                  <Button variant="ghost" onClick={() => setShowInviteMember(true)} className="text-white hover:bg-white/10 font-bold h-11 justify-start gap-3 px-5 border border-white/20">
                    <UserPlus className="w-4 h-4" /> Invite Teammates
                  </Button>
                  <Button variant="ghost" onClick={() => setShowSetStatus(true)} className="text-white hover:bg-white/10 font-bold h-11 justify-start gap-3 px-5 border border-white/20">
                    <Smile className="w-4 h-4" /> Set a Status
                  </Button>
                </div>
              </div>
            </div>
          </div>

          <div className="p-8 max-w-6xl mx-auto w-full space-y-12 pb-32">
            {/* Quick Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6 -mt-20 relative z-20">
              <Card
                className="bg-white dark:bg-[#222529] border-none shadow-2xl hover:translate-y-[-4px] transition-all duration-300 cursor-pointer"
                onClick={() => router.push('/workspace/activity')}
              >
                <CardContent className="p-6 flex items-center gap-5">
                  <div className="w-12 h-12 rounded-2xl bg-purple-500/10 flex items-center justify-center text-purple-600 shrink-0">
                    <MessageSquare className="w-6 h-6" />
                  </div>
                  <div>
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Unread Mentions</p>
                    {/* Phase 65C: prefer activity_summary.unread_mention_count (durable inbox) over legacy pending_actions */}
                    <p className="text-2xl font-black">{
                      homeData?.activity?.unread_mention_count
                        ?? homeData?.activity_summary?.unread_mention_count
                        ?? homeData?.unread_mention_count
                        ?? homeData?.stats?.pending_actions
                        ?? 0
                    }</p>
                  </div>
                </CardContent>
              </Card>
              <Card
                className="bg-white dark:bg-[#222529] border-none shadow-2xl hover:translate-y-[-4px] transition-all duration-300 cursor-pointer"
                onClick={() => router.push('/workspace/activity')}
              >
                <CardContent className="p-6 flex items-center gap-5">
                  <div className="w-12 h-12 rounded-2xl bg-blue-500/10 flex items-center justify-center text-blue-600 shrink-0">
                    <Clock className="w-6 h-6" />
                  </div>
                  <div>
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Active Threads</p>
                    <p className="text-2xl font-black">{homeData?.stats?.active_threads || 0}</p>
                  </div>
                </CardContent>
              </Card>
              <Card
                className="bg-white dark:bg-[#222529] border-none shadow-2xl hover:translate-y-[-4px] transition-all duration-300 cursor-pointer"
                onClick={() => router.push('/workspace/knowledge')}
              >
                <CardContent className="p-6 flex items-center gap-5">
                  <div className="w-12 h-12 rounded-2xl bg-green-500/10 flex items-center justify-center text-green-600 shrink-0">
                    <Sparkles className="w-6 h-6" />
                  </div>
                  <div>
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">New Knowledge</p>
                    <p className="text-2xl font-black">{homeData?.recent_artifacts?.length || 0}</p>
                  </div>
                </CardContent>
              </Card>
              <Card
                className="bg-white dark:bg-[#222529] border-none shadow-2xl hover:translate-y-[-4px] transition-all duration-300 cursor-pointer"
                onClick={() => router.push('/workspace/knowledge/inbox')}
              >
                <CardContent className="p-6 flex items-center gap-5">
                  <div className="w-12 h-12 rounded-2xl bg-emerald-500/10 flex items-center justify-center text-emerald-600 shrink-0 relative">
                    <Inbox className="w-6 h-6" />
                    {(homeData?.knowledge_inbox_count || 0) > 0 && (
                      <span className="absolute -top-1 -right-1 min-w-[18px] h-[18px] px-1 rounded-full bg-emerald-500 text-white text-[9px] font-black flex items-center justify-center border-2 border-white dark:border-[#222529]">
                        {homeData?.knowledge_inbox_count}
                      </span>
                    )}
                  </div>
                  <div>
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Knowledge Inbox</p>
                    <p className="text-2xl font-black">{homeData?.knowledge_inbox_count || 0}</p>
                  </div>
                </CardContent>
              </Card>
            </div>

            <HomeWorkbench
              homeData={homeData}
              workspaceViews={workspaceViews}
              isLoadingWorkspaceViews={isLoadingWorkspaceViews}
              workspaceViewsError={workspaceViewsError}
            />
          </div>
        </div>
      </ScrollArea>
      <CreateChannelDialog open={showCreateChannel} onOpenChange={setShowCreateChannel} />
      <InviteMemberDialog open={showInviteMember} onOpenChange={setShowInviteMember} />
      <SetStatusDialog open={showSetStatus} onOpenChange={setShowSetStatus} />
    </div>
  )
}
