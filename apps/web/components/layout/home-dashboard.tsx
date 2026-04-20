"use client"

import { useEffect } from "react"
import { useDirectoryStore } from "@/stores/directory-store"
import { Layout, Users, Zap, Briefcase, ChevronRight } from "lucide-react"
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
            {/* User Groups */}
            <section className="space-y-4">
              <h2 className="text-sm font-bold flex items-center gap-2 px-1">
                <Users className="w-4 h-4 text-muted-foreground" />
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
            </section>

            {/* Workflows & Tools */}
            <section className="space-y-6">
              <div className="space-y-4">
                <h2 className="text-sm font-bold flex items-center gap-2 px-1">
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
                <h2 className="text-sm font-bold flex items-center gap-2 px-1">
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
            </section>
          </div>
        </div>
      </ScrollArea>
    </div>
  )
}
