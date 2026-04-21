"use client"

import { useEffect } from "react"
import { useDirectoryStore } from "@/stores/directory-store"
import { Layout, Users, Zap, Briefcase, ChevronRight, ListTodo, Terminal, FileText, CheckCircle2 } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { formatDistanceToNow } from "date-fns"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { useWorkspaceStore } from "@/stores/workspace-store"

export function HomeDashboard() {
  const { userGroups, workflows, tools, fetchUserGroups, fetchWorkflows, fetchTools } = useDirectoryStore()
  const { homeData, fetchHome } = useWorkspaceStore()

  useEffect(() => {
    fetchUserGroups()
    fetchWorkflows()
    fetchTools()
    fetchHome()
  }, [fetchUserGroups, fetchWorkflows, fetchTools, fetchHome])

  return (
    <div className="flex-1 flex flex-col h-full bg-white dark:bg-[#1a1d21]">
      <header className="h-14 px-6 flex items-center border-b shrink-0">
        <Layout className="w-5 h-5 mr-2 text-purple-600" />
        <h1 className="text-lg font-black tracking-tight uppercase">Workspace Home</h1>
      </header>

      <ScrollArea className="flex-1">
        <div className="p-8 max-w-5xl mx-auto space-y-10 pb-20">
          {/* Quick Stats */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Card className="bg-purple-500/5 border-purple-500/20 shadow-none">
              <CardHeader className="pb-2">
                <CardTitle className="text-xs font-bold uppercase tracking-widest text-purple-600">Pending Actions</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-black">{homeData?.activity?.unread_count || 0}</div>
              </CardContent>
            </Card>
            <Card className="bg-blue-500/5 border-blue-500/20 shadow-none">
              <CardHeader className="pb-2">
                <CardTitle className="text-xs font-bold uppercase tracking-widest text-blue-600">Active Threads</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-black">{homeData?.activity?.draft_count || 0}</div>
              </CardContent>
            </Card>
            <Card className="bg-green-500/5 border-green-500/20 shadow-none">
              <CardHeader className="pb-2">
                <CardTitle className="text-xs font-bold uppercase tracking-widest text-green-600">Recent Artifacts</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-black">{homeData?.profile?.recent_artifacts?.length || 0}</div>
              </CardContent>
            </Card>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Recent Work Aggregation */}
            <section className="space-y-6">
              {/* Recent Lists */}
              <div className="space-y-4">
                <h2 className="text-sm font-bold flex items-center gap-2 px-1 text-muted-foreground uppercase tracking-widest">
                  <ListTodo className="w-4 h-4" />
                  Recent Lists
                </h2>
                <div className="grid gap-2">
                  {(homeData?.recent_lists || []).map((list: any) => (
                    <div key={list.id} className="p-3 border rounded-xl flex items-center justify-between hover:bg-muted/30 transition-colors cursor-pointer group">
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded bg-purple-500/10 flex items-center justify-center text-purple-600">
                          <ListTodo className="w-4 h-4" />
                        </div>
                        <div>
                          <p className="text-sm font-bold">{list.title}</p>
                          <p className="text-[10px] text-muted-foreground">
                            {list.completed_count}/{list.item_count} items completed
                          </p>
                        </div>
                      </div>
                      <Badge variant="outline" className="text-[8px] h-4 font-black">#{list.channel_name || 'general'}</Badge>
                    </div>
                  ))}
                  {(!homeData?.recent_lists || homeData.recent_lists.length === 0) && (
                    <div className="p-4 border border-dashed rounded-xl text-center text-[10px] text-muted-foreground italic">
                      No recent checklists found.
                    </div>
                  )}
                </div>
              </div>

              {/* Recent Tool Runs */}
              <div className="space-y-4">
                <h2 className="text-sm font-bold flex items-center gap-2 px-1 text-muted-foreground uppercase tracking-widest">
                  <Terminal className="w-4 h-4" />
                  Recent Automations
                </h2>
                <div className="grid gap-2">
                  {(homeData?.recent_tool_runs || []).map((run: any) => (
                    <div key={run.id} className="p-3 border rounded-xl flex items-center justify-between hover:bg-muted/30 transition-colors cursor-pointer group">
                      <div className="flex items-center gap-3">
                        <div className={cn(
                          "w-8 h-8 rounded bg-muted flex items-center justify-center",
                          run.status === 'success' ? "text-green-600 bg-green-500/10" : "text-amber-600 bg-amber-500/10"
                        )}>
                          <Zap className="w-4 h-4" />
                        </div>
                        <div>
                          <p className="text-sm font-bold">{run.tool_name}</p>
                          <p className="text-[10px] text-muted-foreground">
                            {run.status} • {formatDistanceToNow(new Date(run.started_at), { addSuffix: true })}
                          </p>
                        </div>
                      </div>
                      {run.status === 'success' && <CheckCircle2 className="w-3.5 h-3.5 text-green-600" />}
                    </div>
                  ))}
                  {(!homeData?.recent_tool_runs || homeData.recent_tool_runs.length === 0) && (
                    <div className="p-4 border border-dashed rounded-xl text-center text-[10px] text-muted-foreground italic">
                      No recent tool executions.
                    </div>
                  )}
                </div>
              </div>

              {/* Latest Files */}
              <div className="space-y-4">
                <h2 className="text-sm font-bold flex items-center gap-2 px-1 text-muted-foreground uppercase tracking-widest">
                  <FileText className="w-4 h-4" />
                  Latest Files
                </h2>
                <div className="grid gap-2">
                  {(homeData?.recent_files || []).map((file: any) => (
                    <div key={file.id} className="p-3 border rounded-xl flex items-center justify-between hover:bg-muted/30 transition-colors cursor-pointer group">
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded bg-blue-500/10 flex items-center justify-center text-blue-600">
                          <FileText className="w-4 h-4" />
                        </div>
                        <div>
                          <p className="text-sm font-bold truncate max-w-[200px]">{file.name}</p>
                          <p className="text-[10px] text-muted-foreground">
                            {file.type} • {(file.size / 1024).toFixed(1)} KB
                          </p>
                        </div>
                      </div>
                      <Badge variant="outline" className="text-[8px] h-4 font-black">#{file.channel_name || 'general'}</Badge>
                    </div>
                  ))}
                  {(!homeData?.recent_files || homeData.recent_files.length === 0) && (
                    <div className="p-4 border border-dashed rounded-xl text-center text-[10px] text-muted-foreground italic">
                      No files uploaded recently.
                    </div>
                  )}
                </div>
              </div>
            </section>

            <section className="space-y-10">
              {/* User Groups */}
              <div className="space-y-4">
                <h2 className="text-sm font-bold flex items-center gap-2 px-1 text-muted-foreground uppercase tracking-widest">
                  <Users className="w-4 h-4" />
                  User Groups
                </h2>
                <div className="grid gap-2">
                  {userGroups.map(group => (
                    <div key={group.id} className="p-3 border rounded-xl flex items-center justify-between hover:bg-muted/30 transition-colors cursor-pointer group">
                      <div>
                        <p className="text-sm font-bold group-hover:text-purple-600 transition-colors">{group.name}</p>
                        <p className="text-[10px] text-muted-foreground">{group.memberCount} members</p>
                      </div>
                      <ChevronRight className="w-4 h-4 text-muted-foreground" />
                    </div>
                  ))}
                </div>
              </div>

              {/* Workflows & Tools */}
              <div className="space-y-6">
                <div className="space-y-4">
                  <h2 className="text-sm font-bold flex items-center gap-2 px-1 text-muted-foreground uppercase tracking-widest">
                    <Zap className="w-4 h-4 text-amber-500" />
                    Workflows
                  </h2>
                  <div className="grid gap-2">
                    {workflows.map(wf => (
                      <div key={wf.id} className="p-3 border rounded-xl bg-amber-500/5 border-amber-500/10 flex items-center gap-3">
                        <Zap className="w-4 h-4 text-amber-600" />
                        <span className="text-sm font-bold">{wf.name}</span>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="space-y-4">
                  <h2 className="text-sm font-bold flex items-center gap-2 px-1 text-muted-foreground uppercase tracking-widest">
                    <Briefcase className="w-4 h-4 text-blue-500" />
                    Integration Tools
                  </h2>
                  <div className="grid grid-cols-2 gap-2">
                    {tools.map(tool => (
                      <button key={tool.id} className="p-3 border rounded-xl hover:bg-muted transition-colors text-left group">
                        <p className="text-xs font-bold group-hover:text-blue-600 transition-colors">{tool.name}</p>
                        <p className="text-[9px] text-muted-foreground">{tool.category}</p>
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </section>
          </div>
        </div>
      </ScrollArea>
    </div>
  )
}
