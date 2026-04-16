import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"
import { cn } from "@/lib/utils"

interface AgentState {
  name: string
  skill: string
  task: string
  progress: number
  status: 'active' | 'thinking' | 'idle'
}

export function AgentStateCard({ agent }: { agent: AgentState }) {
  return (
    <Card className="min-w-[200px] bg-muted/40 border-white/5">
      <CardContent className="p-3 flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <span className="font-bold text-sm text-white">{agent.name}</span>
          <div className={cn(
            "w-2 h-2 rounded-full",
            agent.status === 'active' ? "bg-green-500 animate-pulse" : 
            agent.status === 'thinking' ? "bg-amber-500 animate-pulse" : "bg-slate-500"
          )} />
        </div>
        <Badge variant="outline" className="w-fit text-[10px] font-mono py-0 text-white/70 border-white/10">{agent.skill}</Badge>
        <p className="text-[10px] text-muted-foreground truncate">{agent.task}</p>
        <Progress value={agent.progress} className="h-1 bg-white/5" />
      </CardContent>
    </Card>
  )
}
