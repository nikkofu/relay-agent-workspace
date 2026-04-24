"use client"

import { useEffect, useState } from "react"
import { useDirectoryStore } from "@/stores/directory-store"
import { 
  Layout, Users, Zap, Briefcase, ChevronRight, FileText, MessageSquare, 
  Plus, UserPlus, Smile, ArrowRight, Sparkles, Hash, Clock, Inbox, Newspaper
} from "lucide-react"
import { useRouter } from "next/navigation"
import { format } from "date-fns"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { useUserStore } from "@/stores/user-store"
import { useChannelStore } from "@/stores/channel-store"
import { useUIStore } from "@/stores/ui-store"
import { useArtifactStore } from "@/stores/artifact-store"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Copy, ExternalLink, MoreVertical } from "lucide-react"
import { CreateChannelDialog } from "@/components/channel/create-channel-dialog"
import { InviteMemberDialog } from "@/components/workspace/invite-member-dialog"
import { SetStatusDialog } from "@/components/workspace/set-status-dialog"
import { TrendingEntitiesCard } from "@/components/knowledge/trending-entities-card"
import { HomeExecutionBlocks } from "@/components/layout/home-execution-blocks"

function stripHtml(html: string): string {
  if (!html) return ""
  return html.replace(/<[^>]*>/g, "").replace(/&amp;/g, "&").replace(/&lt;/g, "<").replace(/&gt;/g, ">").replace(/&quot;/g, '"').replace(/&#039;/g, "'").replace(/&nbsp;/g, " ").trim()
}

export function HomeDashboard() {
  const router = useRouter()
  const { userGroups, workflows, tools, fetchUserGroups, fetchWorkflows, fetchTools } = useDirectoryStore()
  const { homeData, fetchHome, currentWorkspace, isHomeExecutionStale } = useWorkspaceStore()
  const { currentUser } = useUserStore()
  const { setCurrentChannelById } = useChannelStore()
  const { openCanvas } = useUIStore()
  const { duplicateArtifact } = useArtifactStore()
  const [mounted, setMounted] = useState(false)
  const [showCreateChannel, setShowCreateChannel] = useState(false)
  const [showInviteMember, setShowInviteMember] = useState(false)
  const [showSetStatus, setShowSetStatus] = useState(false)

  useEffect(() => {
    setMounted(true)
    fetchUserGroups()
    fetchWorkflows()
    fetchTools()
    fetchHome()
  }, [fetchUserGroups, fetchWorkflows, fetchTools, fetchHome])

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
                      homeData?.activity_summary?.unread_mention_count
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

            <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 pt-4">
              {/* Left Column: Recent Activity & Discussion */}
              <div className="lg:col-span-7 space-y-10">
                <div className="space-y-5">
                  <h3 className="text-lg font-black tracking-tight flex items-center gap-2">
                    <MessageSquare className="w-5 h-5 text-purple-600" />
                    Recent Conversations
                  </h3>
                  <div className="space-y-1">
                    {(homeData?.recent_activity || []).map((item: any) => (
                      <button 
                        key={item.id}
                        onClick={() => item.channel_id && setCurrentChannelById(item.channel_id)}
                        className="w-full flex items-start gap-4 p-4 hover:bg-muted/50 rounded-2xl transition-all text-left group border border-transparent hover:border-border"
                      >
                        <div className="w-10 h-10 rounded-xl bg-muted flex items-center justify-center shrink-0 group-hover:bg-white dark:group-hover:bg-black transition-colors">
                          <Hash className="w-5 h-5 text-muted-foreground group-hover:text-purple-600" />
                        </div>
                        <div className="min-w-0 flex-1">
                          <div className="flex items-center justify-between">
                            <p className="text-sm font-bold truncate leading-tight">#{item.channel_name || 'general'}</p>
                            <span className="text-[10px] text-muted-foreground font-medium opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-1 uppercase">
                              Jump to <ArrowRight className="w-2.5 h-2.5" />
                            </span>
                          </div>
                          <p className="text-sm text-muted-foreground line-clamp-1 mt-1 font-medium">{stripHtml(item.last_message || '')}</p>
                        </div>
                      </button>
                    ))}
                    {(!homeData?.recent_activity || homeData.recent_activity.length === 0) && (
                      <div className="py-12 text-center border-2 border-dashed rounded-3xl opacity-40">
                        <p className="text-sm font-bold italic">No recent activity to show.</p>
                      </div>
                    )}
                  </div>
                </div>

                {/* Trending Knowledge Entities (Phase 59) */}
                {currentWorkspace?.id && (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <h3 className="text-lg font-black tracking-tight flex items-center gap-2">
                        <Sparkles className="w-5 h-5 text-orange-500" />
                        Trending Knowledge
                      </h3>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-[10px] font-black text-orange-600 hover:text-orange-500"
                        onClick={() => router.push('/workspace/knowledge')}
                      >
                        Open wiki
                        <ArrowRight className="w-3 h-3 ml-1" />
                      </Button>
                    </div>
                    <TrendingEntitiesCard workspaceId={currentWorkspace.id} limit={5} />
                  </div>
                )}

                {/* Phase 66 T09: Channel Execution blocks */}
                <HomeExecutionBlocks homeData={homeData} />

                {/* Recent Knowledge Digests */}
                {(homeData?.recent_knowledge_digests?.length || 0) > 0 && (
                  <div className="space-y-5">
                    <div className="flex items-center justify-between">
                      <h3 className="text-lg font-black tracking-tight flex items-center gap-2">
                        <Newspaper className="w-5 h-5 text-emerald-600" />
                        Recent Knowledge Digests
                      </h3>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-[10px] font-black text-emerald-700 hover:text-emerald-600"
                        onClick={() => router.push('/workspace/knowledge/inbox')}
                      >
                        Open inbox
                        <ArrowRight className="w-3 h-3 ml-1" />
                      </Button>
                    </div>
                    <div className="space-y-2">
                      {(homeData?.recent_knowledge_digests || []).slice(0, 4).map((d: any) => (
                        <button
                          key={d.id || d.message?.id}
                          onClick={() => {
                            if (d.channel?.id && d.message?.id) {
                              router.push(`/workspace?c=${d.channel.id}&m=${d.message.id}`)
                            }
                          }}
                          className="w-full flex items-center gap-3 p-3.5 border rounded-2xl hover:border-emerald-300 hover:bg-emerald-500/5 transition-colors text-left group"
                        >
                          <div className="w-9 h-9 rounded-xl bg-emerald-500/10 flex items-center justify-center shrink-0 text-emerald-600">
                            <Newspaper className="w-4 h-4" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground">
                              <Hash className="w-2.5 h-2.5" />
                              <span className="font-bold">{d.channel?.name || d.channel?.id}</span>
                              <span>·</span>
                              <span>{d.digest?.window}</span>
                              {d.occurred_at && (
                                <span className="ml-auto">{format(new Date(d.occurred_at), 'MMM d')}</span>
                              )}
                            </div>
                            <p className="text-xs font-bold truncate mt-0.5">
                              {d.digest?.headline || `Top ${d.digest?.entries?.length || 0} entities`}
                            </p>
                          </div>
                          <ArrowRight className="w-3 h-3 text-muted-foreground group-hover:text-emerald-600 shrink-0" />
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                <div className="space-y-5">
                  <h3 className="text-lg font-black tracking-tight flex items-center gap-2">
                    <Sparkles className="w-5 h-5 text-blue-600" />
                    Latest Knowledge
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {(homeData?.recent_artifacts || []).slice(0, 4).map((art: any) => (
                      <div key={art.id} className="relative group">
                        <button
                          onClick={() => openCanvas(art.id)}
                          className="w-full p-5 bg-blue-500/5 hover:bg-blue-500/10 border border-blue-500/10 hover:border-blue-500/30 rounded-2xl transition-all text-left flex flex-col justify-between h-36"
                        >
                          <div className="w-10 h-10 rounded-xl bg-blue-500/10 flex items-center justify-center text-blue-600 mb-4 group-hover:bg-blue-600 group-hover:text-white transition-all">
                            <FileText className="w-5 h-5" />
                          </div>
                          <div>
                            <p className="text-sm font-bold truncate group-hover:text-blue-600 transition-colors">{art.title}</p>
                            <p className="text-[10px] text-muted-foreground uppercase font-black tracking-tighter mt-1">v{art.version} • {art.type}</p>
                          </div>
                        </button>
                        
                        <div className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="h-8 w-8 rounded-full hover:bg-blue-500/20 text-blue-600">
                                <MoreVertical className="w-4 h-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem onClick={() => openCanvas(art.id)}>
                                <ExternalLink className="w-3.5 h-3.5 mr-2" />
                                Open Canvas
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => duplicateArtifact(art.id)}>
                                <Copy className="w-3.5 h-3.5 mr-2" />
                                Duplicate
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* Right Column: Organization & Tools */}
              <div className="lg:col-span-5 space-y-12">
                {/* User Groups */}
                <div className="space-y-5">
                  <h3 className="text-lg font-black tracking-tight flex items-center gap-2">
                    <Users className="w-5 h-5 text-purple-600" />
                    User Groups
                  </h3>
                  <div className="bg-muted/30 rounded-3xl p-2 border">
                    {userGroups.slice(0, 3).map(group => (
                      <div key={group.id} onClick={() => router.push('/workspace/people')} className="p-4 hover:bg-white dark:hover:bg-black rounded-2xl flex items-center justify-between transition-all cursor-pointer group shadow-sm hover:shadow-md mb-1 last:mb-0">
                        <div>
                          <p className="text-sm font-bold group-hover:text-purple-600 transition-colors">{group.name}</p>
                          <p className="text-[10px] text-muted-foreground font-black uppercase tracking-tighter mt-0.5">{group.memberCount} members</p>
                        </div>
                        <ChevronRight className="w-4 h-4 text-muted-foreground opacity-30 group-hover:opacity-100 group-hover:translate-x-1 transition-all" />
                      </div>
                    ))}
                    <Button variant="ghost" onClick={() => router.push('/workspace/people')} className="w-full text-xs font-black uppercase tracking-widest text-muted-foreground hover:text-purple-600 h-12 rounded-2xl">
                      View All Groups
                    </Button>
                  </div>
                </div>

                {/* Automations */}
                <div className="space-y-5">
                  <h3 className="text-lg font-black tracking-tight flex items-center gap-2">
                    <Zap className="w-5 h-5 text-amber-500" />
                    Workflows
                  </h3>
                  <div className="grid gap-3">
                    {workflows.slice(0, 3).map(wf => (
                      <div key={wf.id} onClick={() => router.push('/workspace/workflows')} className="p-4 border rounded-2xl flex items-center gap-4 hover:border-amber-500/50 hover:bg-amber-500/5 transition-all group cursor-pointer">
                        <div className="w-10 h-10 rounded-xl bg-amber-500/10 flex items-center justify-center text-amber-600 group-hover:scale-110 transition-transform">
                          <Zap className="w-5 h-5 fill-current" />
                        </div>
                        <div>
                          <span className="text-sm font-bold">{wf.name}</span>
                          <p className="text-[10px] text-muted-foreground line-clamp-1 mt-0.5">{wf.description}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Tools */}
                <div className="space-y-5">
                  <h3 className="text-lg font-black tracking-tight flex items-center gap-2">
                    <Briefcase className="w-5 h-5 text-blue-500" />
                    Tools
                  </h3>
                  <div className="grid grid-cols-2 gap-3">
                    {tools.slice(0, 4).map(tool => (
                      <button key={tool.id} onClick={() => router.push('/workspace/tools')} className="p-4 border rounded-2xl hover:bg-muted transition-all text-left group hover:border-blue-500/30">
                        <p className="text-xs font-bold group-hover:text-blue-600 transition-colors">{tool.name}</p>
                        <p className="text-[10px] text-muted-foreground uppercase font-black tracking-tighter mt-1">{tool.category}</p>
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </ScrollArea>
      <CreateChannelDialog open={showCreateChannel} onOpenChange={setShowCreateChannel} />
      <InviteMemberDialog open={showInviteMember} onOpenChange={setShowInviteMember} />
      <SetStatusDialog open={showSetStatus} onOpenChange={setShowSetStatus} />
    </div>
  )
}
