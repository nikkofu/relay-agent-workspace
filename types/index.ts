export type UserStatus = "online" | "away" | "offline" | "busy"

export interface User {
  id: string
  name: string
  email: string
  avatar?: string
  status: UserStatus
  statusText?: string
  lastSeen?: string
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
  type: "public" | "private"
  workspaceId: string
  isStarred?: boolean
  unreadCount?: number
  memberCount: number
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
  type: "image" | "file" | "link"
  url: string
  name: string
  size?: number
  mimeType?: string
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
  role: "user" | "assistant" | "system" | "data"
  content: string
  createdAt: string
  reasoning?: string
  tools?: any[]
  sources?: any[]
}
