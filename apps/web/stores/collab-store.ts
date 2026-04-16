import { create } from "zustand"

export interface AgentState {
  name: string
  skill: string
  task: string
  progress: number
  status: 'active' | 'thinking' | 'idle'
}

export interface TaskItem {
  status: string
  task: string
  assignedTo: string
  deadline: string
  description: string
}

interface CollabState {
  agents: AgentState[]
  tasks: TaskItem[]
  setCollabData: (data: { active_superpowers: any[], task_board: any[] }) => void
}

export const useCollabStore = create<CollabState>((set) => ({
  agents: [],
  tasks: [],
  setCollabData: (data) => {
    // Map backend active_superpowers to frontend AgentState
    const mappedAgents: AgentState[] = data.active_superpowers.map((row: any) => {
      // Basic heuristic for status based on progress or skill
      let status: 'active' | 'thinking' | 'idle' = 'idle'
      const progress = parseInt(row.Progress) || 0
      
      if (progress > 0 && progress < 100) status = 'active'
      if (row.Skill !== 'idle' && row.Skill !== 'awaiting') status = 'active'
      if (row.Skill === 'awaiting') status = 'thinking'
      
      return {
        name: row.Agent,
        skill: row["Current Skill"],
        task: row["Active Task"],
        progress: progress,
        status: status
      }
    })

    // Map backend task_board to frontend TaskItem
    const mappedTasks: TaskItem[] = data.task_board.map((row: any) => ({
      status: row.Status,
      task: row.Task,
      assignedTo: row["Assigned To"],
      deadline: row.Deadline,
      description: row.Description
    }))

    set({ agents: mappedAgents, tasks: mappedTasks })
  },
}))
