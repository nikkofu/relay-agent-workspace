import { create } from "zustand"
import { API_BASE_URL } from "@/lib/constants"

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
  fetchSnapshot: () => Promise<void>
}

const mapAgentRow = (row: any): AgentState => {
  let status: 'active' | 'thinking' | 'idle' = 'idle'
  const progress = parseInt(row.progress || row.Progress) || 0
  const skill = row.current_skill || row["Current Skill"] || row.skill || "idle"

  if (progress > 0 && progress < 100) status = 'active'
  if (skill !== 'idle' && skill !== 'awaiting') status = 'active'
  if (skill === 'awaiting') status = 'thinking'

  return {
    name: row.agent || row.Agent,
    skill,
    task: row.active_task || row["Active Task"] || row.task || "",
    progress,
    status
  }
}

const mapTaskRow = (row: any): TaskItem => ({
  status: row.status || row.Status,
  task: row.task || row.Task,
  assignedTo: row.assigned_to || row["Assigned To"],
  deadline: row.deadline || row.Deadline,
  description: row.description || row.Description
})

export const useCollabStore = create<CollabState>((set) => ({
  agents: [],
  tasks: [],
  setCollabData: (data) => {
    const mappedAgents: AgentState[] = data.active_superpowers.map(mapAgentRow)
    const mappedTasks: TaskItem[] = data.task_board.map(mapTaskRow)

    set({ agents: mappedAgents, tasks: mappedTasks })
  },
  fetchSnapshot: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/agent-collab/snapshot`)
      const data = await response.json()
      set((state) => {
        const mappedAgents: AgentState[] = data.active_superpowers.map(mapAgentRow)
        const mappedTasks: TaskItem[] = data.task_board.map(mapTaskRow)

        return {
          agents: mappedAgents.length > 0 ? mappedAgents : state.agents,
          tasks: mappedTasks.length > 0 ? mappedTasks : state.tasks
        }
      })
    } catch (error) {
      console.error("Failed to fetch agent collab snapshot:", error)
    }
  }
}))
