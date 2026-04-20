"use client"

import { useEffect, useState } from "react"
import { useDirectoryStore } from "@/stores/directory-store"
import { Zap, Play, Clock, CheckCircle2, AlertCircle, Loader2, Settings, History } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { formatDistanceToNow } from "date-fns"
import { Badge } from "@/components/ui/badge"
import { 
  Tabs, 
  TabsContent, 
  TabsList, 
  TabsTrigger 
} from "@/components/ui/tabs"

export default function WorkflowsPage() {
  const { 
    workflows, workflowRuns, isWorkflowLoading,
    fetchWorkflows, fetchWorkflowRuns, triggerWorkflow 
  } = useDirectoryStore()
  
  const [activeTab, setActiveTab] = useState("all")

  useEffect(() => {
    fetchWorkflows()
    fetchWorkflowRuns()
  }, [fetchWorkflows, fetchWorkflowRuns])

  const handleRun = async (id: string) => {
    await triggerWorkflow(id)
    fetchWorkflowRuns()
  }

  return (
    <div className="flex-1 flex flex-col bg-white dark:bg-[#1a1d21] h-full overflow-hidden">
      <header className="h-14 px-6 flex items-center border-b shrink-0 justify-between">
        <div className="flex items-center gap-2">
          <Zap className="w-5 h-5 text-amber-500 fill-amber-500" />
          <h1 className="text-lg font-black tracking-tight uppercase">Automations & Workflows</h1>
        </div>
        <Button variant="outline" size="sm" className="h-8 text-xs font-bold gap-2 border-amber-500/20 bg-amber-500/5 text-amber-600">
          <Plus className="w-3.5 h-3.5" />
          Create Workflow
        </Button>
      </header>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col min-h-0">
        <div className="px-6 border-b bg-muted/5">
          <TabsList className="h-12 bg-transparent gap-6">
            <TabsTrigger value="all" className="h-12 rounded-none border-b-2 border-transparent data-[state=active]:border-purple-600 data-[state=active]:bg-transparent px-0 text-xs font-bold uppercase tracking-widest">Available Workflows</TabsTrigger>
            <TabsTrigger value="history" className="h-12 rounded-none border-b-2 border-transparent data-[state=active]:border-purple-600 data-[state=active]:bg-transparent px-0 text-xs font-bold uppercase tracking-widest">Run History</TabsTrigger>
          </TabsList>
        </div>

        <div className="flex-1 min-h-0">
          <TabsContent value="all" className="h-full m-0">
            <ScrollArea className="h-full">
              <div className="p-8 max-w-4xl mx-auto grid grid-cols-1 md:grid-cols-2 gap-4">
                {workflows.map(wf => (
                  <Card key={wf.id} className="group hover:border-amber-500/50 transition-all shadow-none">
                    <CardHeader className="pb-3">
                      <div className="flex items-start justify-between">
                        <div className="w-10 h-10 rounded-xl bg-amber-500/10 flex items-center justify-center text-amber-600">
                          <Zap className="w-5 h-5" />
                        </div>
                        <Button 
                          variant="ghost" 
                          size="icon" 
                          className="h-8 w-8 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity"
                        >
                          <Settings className="w-4 h-4" />
                        </Button>
                      </div>
                      <CardTitle className="text-sm font-black mt-4">{wf.name}</CardTitle>
                      <CardDescription className="text-[11px] line-clamp-2 min-h-[32px]">{wf.description}</CardDescription>
                    </CardHeader>
                    <CardContent className="pt-0 flex items-center justify-between">
                      <div className="flex items-center gap-1.5 text-[9px] text-muted-foreground font-bold uppercase tracking-tighter">
                        <Clock className="w-3 h-3" />
                        Updated {formatDistanceToNow(new Date(wf.updatedAt), { addSuffix: true })}
                      </div>
                      <Button 
                        size="sm" 
                        variant="ghost"
                        className="h-8 px-3 text-[10px] font-black uppercase tracking-widest gap-2 bg-green-500/5 text-green-600 hover:bg-green-500/10"
                        onClick={() => handleRun(wf.id)}
                      >
                        <Play className="w-3 h-3 fill-current" />
                        Run Now
                      </Button>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </ScrollArea>
          </TabsContent>

          <TabsContent value="history" className="h-full m-0">
            <ScrollArea className="h-full">
              <div className="p-0">
                {isWorkflowLoading ? (
                  <div className="py-20 flex flex-col items-center gap-2">
                    <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
                    <p className="text-xs text-muted-foreground">Fetching run history...</p>
                  </div>
                ) : (
                  <div className="divide-y divide-border">
                    {workflowRuns.map(run => (
                      <div key={run.id} className="flex items-center justify-between p-4 hover:bg-muted/30 transition-colors">
                        <div className="flex items-center gap-4">
                          <div className={cn(
                            "w-8 h-8 rounded-full flex items-center justify-center",
                            run.status === 'success' ? "bg-green-500/10 text-green-600" : "bg-red-500/10 text-red-600"
                          )}>
                            {run.status === 'success' ? <CheckCircle2 className="w-4 h-4" /> : <AlertCircle className="w-4 h-4" />}
                          </div>
                          <div>
                            <p className="text-sm font-bold">{run.workflow_name}</p>
                            <div className="flex items-center gap-2 mt-0.5">
                              <Badge variant="outline" className="text-[8px] h-4 font-black uppercase tracking-tighter border-muted-foreground/20">#{run.id.slice(0, 8)}</Badge>
                              <span className="text-[10px] text-muted-foreground">{formatDistanceToNow(new Date(run.started_at), { addSuffix: true })}</span>
                            </div>
                          </div>
                        </div>
                        <div className="flex items-center gap-6">
                          <div className="text-right">
                            <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">Duration</p>
                            <p className="text-xs font-mono">{run.duration_ms}ms</p>
                          </div>
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <History className="w-4 h-4" />
                          </Button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
                
                {workflowRuns.length === 0 && !isWorkflowLoading && (
                  <div className="py-32 text-center flex flex-col items-center gap-4">
                    <History className="w-16 h-16 text-muted-foreground/10" />
                    <div className="space-y-1">
                      <p className="text-muted-foreground font-black uppercase text-xs tracking-widest">No Run History</p>
                      <p className="text-[11px] text-muted-foreground italic">Workflows you trigger will appear here.</p>
                    </div>
                  </div>
                )}
              </div>
            </ScrollArea>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  )
}

function Plus({ className }: { className?: string }) {
  return (
    <div className={cn("w-4 h-4 rounded-full border-2 border-current flex items-center justify-center font-bold text-[10px]", className)}>
      +
    </div>
  )
}
