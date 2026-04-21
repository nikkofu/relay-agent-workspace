export type UserStatus = "online" | "away" | "offline" | "busy"

export interface FileAsset {
  id: string
  name: string
  type: string
  size: number
  url: string
  channelId?: string
  userId: string
  createdAt: string
  isArchived?: boolean
  comment_count?: number
  share_count?: number
  starred?: boolean
  tags?: string[]
  knowledge_state?: string
  source_kind?: string
  summary?: string
}

export interface FileComment {
  id: string
  file_id: string
  user_id: string
  content: string
  created_at: string
  updated_at: string
  user?: User
}

export interface FileShare {
  id: string
  file_id: string
  channel_id?: string
  thread_id?: string
  message_id?: string
  shared_by: string
  comment?: string
  created_at: string
  actor?: User
  message?: Message
}

export interface User {
  id: string
  name: string
  email: string
  avatar?: string
  status: UserStatus
  statusText?: string
  statusEmoji?: string
  statusExpiresAt?: string
  lastSeen?: string
  title?: string
  department?: string
  timezone?: string
  workingHours?: string
  pronouns?: string
  location?: string
  phone?: string
  bio?: string
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
  members?: { user: User, role: string, joinedAt: string }[]
}

export interface WorkflowRun {
  id: string
  workflowId: string
  workflowName: string
  status: 'pending' | 'running' | 'success' | 'failed' | 'cancelled'
  startedAt: string
  finishedAt?: string
  durationMs?: number
  triggeredBy: string
  error?: string
  steps?: {
    name: string
    status: string
    durationMs: number
  }[]
}

export interface FileAudit {
  id: string
  fileId: string
  userId: string
  action: 'upload' | 'download' | 'archive' | 'restore' | 'delete' | 'retention_update'
  occurredAt: string
  metadata?: any
  user?: User
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

export interface MessageAttachmentPreview {
  url?: string
  thumbnail_url?: string
  width?: number
  height?: number
  content_type?: string
}

export interface MessageAttachment {
  id: string
  type: "image" | "file" | "link" | "artifact"
  kind?: string
  url: string
  name: string
  size?: number
  mimeType?: string
  artifact?: any // Hydrated artifact if type is artifact
  file?: Partial<FileAsset> // Hydrated file if type is file (Phase 43 enriched)
  preview?: MessageAttachmentPreview // Preview metadata (Phase 43)
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
