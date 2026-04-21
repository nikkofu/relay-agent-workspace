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
  extraction_status?: string
  content_summary?: string
  last_indexed_at?: string
  needs_ocr?: boolean
  ocr_provider?: string
  ocr_is_mock?: boolean
  is_searchable?: boolean
  is_citable?: boolean
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

export interface FileChunk {
  id: string
  file_id: string
  chunk_index: number
  content: string
  char_count?: number
  token_estimate?: number
}

export interface FileCitation {
  id: string
  file_id: string
  message_id?: string
  artifact_id?: string
  snippet: string
  created_at: string
}

export type EvidenceKind = 'file_chunk' | 'message' | 'thread' | 'artifact_section'

export interface CitationEvidence {
  id: string
  evidence_kind: EvidenceKind
  title?: string
  snippet: string
  locator?: string
  source_kind?: string
  source_ref?: string
  ref_kind?: string
  entity_id?: string
  entity_title?: string
  score?: number
}

export interface KnowledgeEntity {
  id: string
  title: string
  kind: string
  summary?: string
  tags?: string[]
  source_kind?: string
  ref_count?: number
  created_at: string
  updated_at?: string
}

export interface KnowledgeEntityRef {
  id: string
  entity_id: string
  source_kind: string
  source_id: string
  snippet?: string
  created_at: string
}

export interface KnowledgeEntityLink {
  id: string
  from_entity_id: string
  to_entity_id: string
  rel: string
  created_at: string
}

export interface KnowledgeEvent {
  id: string
  entity_id: string
  event_kind: string
  title: string
  description?: string
  occurred_at: string
  created_at: string
}

export interface KnowledgeGraphNode {
  entity: KnowledgeEntity
  rel?: string
  direction?: 'in' | 'out'
}

export interface KnowledgeGraph {
  center: KnowledgeEntity
  nodes: KnowledgeGraphNode[]
}

export interface FileSearchResult {
  id: string
  name: string
  type: string
  size: number
  url?: string
  extraction_status?: string
  is_searchable?: boolean
  is_citable?: boolean
  snippet?: string
  match_reason?: string
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
