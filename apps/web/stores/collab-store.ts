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

export interface LiveMember {
  name: string
  role: string
  specialty: string
  primary_tools: string[]
}

export interface LiveCommMessage {
  id?: string
  from: string
  to?: string
  content: string
  isCode?: boolean
}

export interface LiveCommSection {
  id?: string
  title: string
  date: string
  messages: LiveCommMessage[]
}

interface CollabState {
  agents: AgentState[]
  tasks: TaskItem[]
  members: LiveMember[]
  commLog: LiveCommSection[]
  isLive: boolean
  setCollabData: (data: any) => void
  fetchSnapshot: () => Promise<void>
  fetchMembers: () => Promise<void>
  postCommLog: (entry: { from: string; to?: string; title: string; content: string }) => Promise<void>
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

const parsePrimaryTools = (tools: any): string[] => {
  if (Array.isArray(tools)) return tools
  if (typeof tools === 'string') return tools.split(',').map((s: string) => s.trim()).filter(Boolean)
  return []
}

const groupCommLog = (flatMessages: any[]): LiveCommSection[] => {
  const sectionMap = new Map<string, LiveCommSection>()
  for (const m of flatMessages) {
    const key = `${m.title}|${m.date}`
    if (!sectionMap.has(key)) {
      sectionMap.set(key, {
        id: `cs-${(m.title || '').toLowerCase().replace(/\s+/g, '-')}-${m.date}`,
        title: m.title || '',
        date: m.date || '',
        messages: [],
      })
    }
    const to = m.to && m.to !== '' && m.to !== 'null' ? m.to : undefined
    sectionMap.get(key)!.messages.push({
      id: m.id,
      from: m.from || '',
      to,
      content: m.content || '',
      isCode: m.is_code || false,
    })
  }
  return Array.from(sectionMap.values())
}

export const useCollabStore = create<CollabState>((set) => ({
  agents: [],
  tasks: [],
  members: [],
  commLog: [],
  isLive: false,

  setCollabData: (data) => {
    const mappedAgents: AgentState[] = (data.active_superpowers || []).map(mapAgentRow)
    const mappedTasks: TaskItem[] = (data.task_board || []).map(mapTaskRow)
    const mappedMembers: LiveMember[] = Array.isArray(data.members)
      ? data.members.map((m: any) => ({
          name: m.name || m.Name || "",
          role: m.role || m.Role || "",
          specialty: m.specialty || m.Specialty || "",
          primary_tools: parsePrimaryTools(m.primary_tools || m.primaryTools),
        }))
      : []
    const mappedCommLog: LiveCommSection[] = Array.isArray(data.comm_log)
      ? groupCommLog(data.comm_log)
      : []

    set((state) => ({
      agents: mappedAgents.length > 0 ? mappedAgents : state.agents,
      tasks: mappedTasks.length > 0 ? mappedTasks : state.tasks,
      members: mappedMembers.length > 0 ? mappedMembers : state.members,
      commLog: mappedCommLog.length > 0 ? mappedCommLog : state.commLog,
      isLive: true,
    }))
  },

  fetchSnapshot: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/agent-collab/snapshot`)
      if (!response.ok) return
      const data = await response.json()
      const mappedAgents: AgentState[] = (data.active_superpowers || []).map(mapAgentRow)
      const mappedTasks: TaskItem[] = (data.task_board || []).map(mapTaskRow)
      const mappedMembers: LiveMember[] = Array.isArray(data.members)
        ? data.members.map((m: any) => ({
            name: m.name || m.Name || "",
            role: m.role || m.Role || "",
            specialty: m.specialty || m.Specialty || "",
            primary_tools: parsePrimaryTools(m.primary_tools || m.primaryTools),
          }))
        : []
      const mappedCommLog: LiveCommSection[] = Array.isArray(data.comm_log)
        ? groupCommLog(data.comm_log)
        : []

      set((state) => ({
        agents: mappedAgents.length > 0 ? mappedAgents : state.agents,
        tasks: mappedTasks.length > 0 ? mappedTasks : state.tasks,
        members: mappedMembers.length > 0 ? mappedMembers : state.members,
        commLog: mappedCommLog.length > 0 ? mappedCommLog : state.commLog,
        isLive: true,
      }))
    } catch (error) {
      console.error("Failed to fetch agent collab snapshot:", error)
    }
  },

  fetchMembers: async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/agent-collab/members`)
      if (!response.ok) return
      const data = await response.json()
      const mappedMembers: LiveMember[] = (data.members || []).map((m: any) => ({
        name: m.name || m.Name || "",
        role: m.role || m.Role || "",
        specialty: m.specialty || m.Specialty || "",
        primary_tools: m.primary_tools || m.primaryTools || [],
      }))
      if (mappedMembers.length > 0) {
        set({ members: mappedMembers, isLive: true })
      }
    } catch (error) {
      console.error("Failed to fetch agent-collab members:", error)
    }
  },

  postCommLog: async (entry) => {
    try {
      await fetch(`${API_BASE_URL}/agent-collab/comm-log`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(entry),
      })
    } catch (error) {
      console.error("Failed to post comm-log entry:", error)
    }
  },
}))
