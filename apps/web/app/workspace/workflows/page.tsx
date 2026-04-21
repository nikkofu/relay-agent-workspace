"use client"

import { useEffect, useState } from "react"
import { useDirectoryStore } from "@/stores/directory-store"
import { Zap, Play, Clock, CheckCircle2, AlertCircle, Loader2, Settings, History, MoreVertical, StopCircle, RotateCw, Trash2, ChevronRight, FileText } from "lucide-react"
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
import { 
  Dialog, 
  DialogContent, 
  DialogHeader, 
  DialogTitle, 
  DialogFooter,
} from "@/components/ui/dialog"
import { 
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger 
} from "@/components/ui/dropdown-menu"

export default function WorkflowsPage() {
  const { 
    workflows, workflowRuns, isWorkflowLoading,
    fetchWorkflows, fetchWorkflowRuns, triggerWorkflow,
    fetchWorkflowRunDetail, cancelWorkflowRun, retryWorkflowRun,
    fetchWorkflowRunLogs, deleteWorkflowRun
  } = useDirectoryStore()
  
  const [activeTab, setActiveTab] = useState("all")
  const [selectedRun, setSelectedRun] = useState<any>(null)
  const [isViewingRun, setIsViewingRun] = useState(false)
  const [isViewingLogs, setIsViewingLogs] = useState(false)
  const [logs, setLogs] = useState<string[]>([])
  const [isLogsLoading, setIsLogsLoading] = useState(false)

  useEffect(() => {
    fetchWorkflows()
    fetchWorkflowRuns()
  }, [fetchWorkflows, fetchWorkflowRuns])

  const handleRun = async (id: string) => {
    await triggerWorkflow(id)
    fetchWorkflowRuns()
  }

  const handleViewRun = async (runId: string) => {
    const run = await fetchWorkflowRunDetail(runId)
    if (run) {
      setSelectedRun(run)
      setIsViewingRun(true)
    }
  }

  const handleViewLogs = async (runId: string) => {
    setIsLogsLoading(true)
    setIsViewingLogs(true)
    const runLogs = await fetchWorkflowRunLogs(runId)
    setLogs(runLogs)
    setIsLogsLoading(false)
  }

  const handleDeleteLog = async (runId: string) => {
    await deleteWorkflowRun(runId)
  }

  const handleCancelRun = async (runId: string) => {
    await cancelWorkflowRun(runId)
    fetchWorkflowRuns()
  }

  const handleRetryRun = async (runId: string) => {
    await retryWorkflowRun(runId)
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
          <PlusBadge className="w-3.5 h-3.5" />
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
                            <p className="text-sm font-bold">{run.workflowName}</p>
                            <div className="flex items-center gap-2 mt-0.5">
                              <Badge variant="outline" className="text-[8px] h-4 font-black uppercase tracking-tighter border-muted-foreground/20">#{run.id.slice(0, 8)}</Badge>
                              <span className="text-[10px] text-muted-foreground">{formatDistanceToNow(new Date(run.startedAt), { addSuffix: true })}</span>
                            </div>
                          </div>
                        </div>
                        <div className="flex items-center gap-6">
                          <div className="text-right">
                            <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">Duration</p>
                            <p className="text-xs font-mono">{run.durationMs || 0}ms</p>
                          </div>
                          
                          <div className="flex items-center gap-1">
                            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleViewRun(run.id)}>
                              <History className="w-4 h-4" />
                            </Button>
                            
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="icon" className="h-8 w-8">
                                  <MoreVertical className="w-4 h-4" />
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                <DropdownMenuItem onClick={() => handleViewRun(run.id)}>View Details</DropdownMenuItem>
                                {run.status === 'running' && (
                                  <DropdownMenuItem className="text-red-600" onClick={() => handleCancelRun(run.id)}>
                                    <StopCircle className="w-3.5 h-3.5 mr-2" />
                                    Cancel Run
                                  </DropdownMenuItem>
                                )}
                                {(run.status === 'failed' || run.status === 'cancelled') && (
                                  <DropdownMenuItem className="text-blue-600" onClick={() => handleRetryRun(run.id)}>
                                    <RotateCw className="w-3.5 h-3.5 mr-2" />
                                    Retry Run
                                  </DropdownMenuItem>
                                )}
                                <DropdownMenuItem onClick={() => handleViewLogs(run.id)}>
                                  <FileText className="w-3.5 h-3.5 mr-2" />
                                  View Raw Logs
                                </DropdownMenuItem>
                                <DropdownMenuItem className="text-red-600" onClick={() => handleDeleteLog(run.id)}>
                                  <Trash2 className="w-3.5 h-3.5 mr-2" />
                                  Delete Log
                                </DropdownMenuItem>
                              </DropdownMenuContent>
                            </DropdownMenu>
                          </div>
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

      {/* Run Detail Dialog */}
      <Dialog open={isViewingRun} onOpenChange={setIsViewingRun}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle className="text-xl font-black">Workflow Run Details</DialogTitle>
          </DialogHeader>
          {selectedRun && (
            <div className="py-4 space-y-6">
              <div className="flex items-center justify-between p-4 bg-muted/30 rounded-xl">
                <div>
                  <p className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">Status</p>
                  <div className="flex items-center gap-2 mt-1">
                    <div className={cn(
                      "w-2 h-2 rounded-full",
                      selectedRun.status === 'success' ? "bg-green-500" :
                      selectedRun.status === 'failed' ? "bg-red-500" :
                      selectedRun.status === 'running' ? "bg-blue-500 animate-pulse" : "bg-slate-500"
                    )} />
                    <span className="text-sm font-black uppercase tracking-tight">{selectedRun.status}</span>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">Triggered By</p>
                  <p className="text-sm font-bold mt-1">{selectedRun.triggeredBy}</p>
                </div>
              </div>

              <div className="space-y-3">
                <h4 className="text-xs font-black uppercase tracking-widest flex items-center gap-2">
                  <ChevronRight className="w-4 h-4 text-purple-600" />
                  Execution Steps
                </h4>
                <div className="space-y-2">
                  {selectedRun.steps?.map((step: any, idx: number) => (
                    <div key={idx} className="flex items-center justify-between p-3 border rounded-lg bg-card/50">
                      <div className="flex items-center gap-3">
                        <div className="w-5 h-5 rounded bg-muted flex items-center justify-center text-[10px] font-bold">{idx + 1}</div>
                        <span className="text-xs font-medium">{step.name}</span>
                      </div>
                      <div className="flex items-center gap-3">
                        <span className="text-[10px] font-mono text-muted-foreground">{step.durationMs}ms</span>
                        <Badge variant="outline" className="text-[8px] h-4 font-black uppercase tracking-tighter">{step.status}</Badge>
                      </div>
                    </div>
                  ))}
                  {(!selectedRun.steps || selectedRun.steps.length === 0) && (
                    <p className="text-xs text-muted-foreground italic p-4 text-center border rounded-lg border-dashed">No step data available for this run.</p>
                  )}
                </div>
              </div>

              {selectedRun.error && (
                <div className="p-4 bg-red-500/5 border border-red-500/20 rounded-xl">
                  <p className="text-[10px] font-bold uppercase tracking-widest text-red-600 mb-2">Error Message</p>
                  <p className="text-xs font-mono text-red-700 bg-red-500/5 p-2 rounded whitespace-pre-wrap">{selectedRun.error}</p>
                </div>
              )}
            </div>
          )}
          <DialogFooter>
            <Button variant="ghost" onClick={() => setIsViewingRun(false)}>Close</Button>
            {selectedRun?.status === 'running' && (
              <Button variant="destructive" onClick={() => handleCancelRun(selectedRun.id)}>Cancel Run</Button>
            )}
            {(selectedRun?.status === 'failed' || selectedRun?.status === 'cancelled') && (
              <Button className="bg-[#3f0e40] text-white" onClick={() => handleRetryRun(selectedRun.id)}>Retry Workflow</Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Raw Logs Dialog */}
      <Dialog open={isViewingLogs} onOpenChange={setIsViewingLogs}>
        <DialogContent className="sm:max-w-[700px] p-0 flex flex-col max-h-[85vh]">
          <DialogHeader className="p-6 pb-2">
            <DialogTitle className="text-xl font-black flex items-center gap-2">
              <FileText className="w-5 h-5 text-purple-600" />
              Raw Execution Logs
            </DialogTitle>
          </DialogHeader>
          <div className="flex-1 min-h-0 bg-[#0d0f12] text-green-500 font-mono text-[11px] p-4 overflow-hidden border-y border-white/5">
            <ScrollArea className="h-full">
              {isLogsLoading ? (
                <div className="flex items-center justify-center py-20 opacity-50">
                  <Loader2 className="w-6 h-6 animate-spin mr-2" />
                  Streaming logs...
                </div>
              ) : logs.length === 0 ? (
                <div className="py-20 text-center opacity-30 italic">No log entries recorded for this run.</div>
              ) : (
                <div className="space-y-1">
                  {logs.map((log, i) => (
                    <div key={i} className="whitespace-pre-wrap break-all border-l-2 border-white/10 pl-3 py-0.5 hover:bg-white/5 transition-colors">
                      {log}
                    </div>
                  ))}
                </div>
              )}
            </ScrollArea>
          </div>
          <DialogFooter className="p-4 bg-muted/20">
            <Button variant="outline" size="sm" onClick={() => setIsViewingLogs(false)}>Close Logs</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function PlusBadge({ className }: { className?: string }) {
  return (
    <div className={cn("w-4 h-4 rounded-full border-2 border-current flex items-center justify-center font-bold text-[10px]", className)}>
      +
    </div>
  )
}
