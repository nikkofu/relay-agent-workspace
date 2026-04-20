export type UserStatus = "online" | "away" | "offline" | "busy"

export interface User {
  id: string
  name: string
  email: string
  avatar?: string
  status: UserStatus
  statusText?: string
  lastSeen?: string
  title?: string
  department?: string
  timezone?: string
  workingHours?: string
  aiInsight?: string
  aiProvider?: string
  aiModel?: string
  aiMode?: "fast" | "planning"
  profile?: {
    localTime: string
    workingHours: string
    focusAreas: string[]
    topChannels: any[]
    recentArtifacts: any[]
  }
}

export interface UserGroup {
  id: string
  name: string
  handle: string
  description?: string
  memberCount: number
  members?: { user: User, role: string }[]
}

export interface Workflow {
  id: string
  name: string
  description?: string
  createdBy: string
  updatedAt: string
}

export interface Tool {
  id: string
  name: string
  description?: string
  category: string
  url?: string
}

export interface Workspace {
  id: string
  name: string
  slug: string
  logo?: string
}

export interface Channel {
  id: string
  name: string
  description?: string
  topic?: string
  purpose?: string
  type: "public" | "private"
  workspaceId: string
  isStarred?: boolean
  unreadCount?: number
  memberCount: number
  isArchived?: boolean
}

export interface ChannelMember {
  user: User
  role: "admin" | "member"
}

export interface DirectMessage {
  id: string
  userIds: string[]
  lastMessageAt: string
  unreadCount?: number
}

export interface Reaction {
  emoji: string
  count: number
  userIds: string[]
}

export interface Message {
  id: string
  content: string
  senderId: string
  channelId?: string
  dmId?: string
  threadId?: string
  createdAt: string
  updatedAt?: string
  reactions?: Reaction[]
  replyCount?: number
  lastReplyAt?: string
  isPinned?: boolean
  attachments?: MessageAttachment[]
}

export interface MessageAttachment {
  id: string
  type: "image" | "file" | "link" | "artifact"
  url: string
  name: string
  size?: number
  mimeType?: string
  artifact?: any // Hydrated artifact if type is artifact
  file?: any // Hydrated file if type is file
}

export interface Thread {
  id: string
  parentMessageId: string
  replies: Message[]
}

export interface AIChat {
  id: string
  messages: AIMessage[]
}

export interface AIMessage {
  id: string
  role: "user" | "assistant" | "system"
  content: string
  createdAt: string
  isStreaming?: boolean
  reasoning?: string // 思考过程
  tools?: {
    name: string
    args: any
    state: "calling" | "result"
    result?: any
  }[]
  sources?: { title: string; url: string }[]
}
