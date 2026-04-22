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
  ref_kind?: string
  ref_id?: string
  role?: string
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
  source_kind?: string
  ref_kind?: string
  ref_id?: string
  role?: string
  weight?: number
}

export interface KnowledgeGraphEdge {
  from_id: string
  to_id: string
  rel: string
  weight?: number
  direction?: 'in' | 'out' | 'both'
  role?: string
}

export interface KnowledgeGraph {
  center: KnowledgeEntity
  nodes: KnowledgeGraphNode[]
  edges?: KnowledgeGraphEdge[]
}

export type KnowledgeUpdateType =
  | 'entity.created'
  | 'entity.updated'
  | 'ref.created'
  | 'event.created'
  | 'link.created'
  | 'digest.published'

export type KnowledgeUpdate =
  | { type: 'entity.created' | 'entity.updated'; entityId: string; payload: KnowledgeEntity; ts: number }
  | { type: 'ref.created'; entityId: string; payload: KnowledgeEntityRef; ts: number }
  | { type: 'event.created'; entityId: string; payload: KnowledgeEvent; ts: number }
  | { type: 'digest.published'; entityId: string; payload: { channel_id?: string; message?: unknown; digest?: unknown }; ts: number }
  | { type: 'link.created'; entityId: string; payload: KnowledgeEntityLink; ts: number }

export interface ChannelKnowledgeRef {
  entity_id: string
  entity_title: string
  entity_kind: string
  ref_kind: string
  ref_id: string
  role?: string
  source_title?: string
  source_snippet?: string
  created_at: string
}

export interface ChannelKnowledgeTopEntity {
  entity_id: string
  entity_title: string
  entity_kind: string
  ref_count: number
  message_ref_count: number
  file_ref_count: number
  last_ref_at: string
  trend: { date: string; count: number }[]
}

export interface KnowledgeVelocity {
  recent_window_days: number
  previous_ref_count: number
  recent_ref_count: number
  delta: number
  is_spiking: boolean
}

export interface ChannelKnowledgeSummary {
  channel_id: string
  window_days: number
  total_refs: number
  recent_ref_count: number
  top_entities: ChannelKnowledgeTopEntity[]
  velocity?: KnowledgeVelocity
}

export interface EntitySuggestResult {
  entity_id: string
  entity_title: string
  entity_kind: string
  summary?: string
  ref_count?: number
}

export interface RelatedChannel {
  channel_id: string
  channel_name?: string
  ref_count: number
  last_ref_at?: string
}

export interface EntityHoverCard {
  entity_id: string
  entity_title: string
  entity_kind: string
  summary?: string
  ref_count: number
  channel_ref_count: number
  message_ref_count: number
  file_ref_count: number
  recent_ref_count: number
  last_activity_at?: string
  related_channels: RelatedChannel[]
}

export type MatchSource = 'knowledge_ref' | 'explicit_mention' | 'title_match'

export type FollowNotificationLevel = 'all' | 'digest_only' | 'silent'

export interface KnowledgeEntityFollow {
  id: string
  workspace_id: string
  entity_id: string
  user_id: string
  notification_level: FollowNotificationLevel
  last_alerted_at?: string
  created_at: string
}

export interface EntitySpikePayload {
  entity: KnowledgeEntity
  user_ids: string[]
  channel_id?: string
  related_channel_ids?: string[]
  recent_ref_count: number
  previous_ref_count: number
  delta: number
  occurred_at: string
}

export interface FollowedEntity {
  follow: KnowledgeEntityFollow
  entity: KnowledgeEntity
  is_following: boolean
}

export interface WorkspaceKnowledgeSettings {
  workspace_id: string
  spike_threshold: number
  spike_cooldown_minutes: number
}

export interface EntityActivityBucket {
  date: string
  ref_count: number
}

export interface EntityActivity {
  entity_id: string
  workspace_id: string
  days: number
  buckets: EntityActivityBucket[]
}

export interface TrendingEntity {
  entity: KnowledgeEntity
  recent_ref_count: number
  previous_ref_count: number
  velocity_delta: number
  related_channel_ids: string[]
  last_ref_at?: string
}

export interface FollowedEntityStatsKindCount {
  kind: string
  count: number
}

export interface FollowedEntityStats {
  total_count: number
  spiking_count: number
  muted_count: number
  by_kind: FollowedEntityStatsKindCount[]
}

export interface SharedEntityLink {
  entity_id: string
  workspace_id: string
  title: string
  url: string
  short_url: string
  relative_path: string
}

// ── Phase 61 / 63A ───────────────────────────────────────────────────────────

export interface EntityBrief {
  entity_id: string
  workspace_id: string
  title: string
  content: string
  reasoning?: string
  provider?: string
  model?: string
  generated_at: string
  ref_count?: number
  event_count?: number
  last_ref_at?: string
  cached?: boolean
}

export interface WeeklyBrief {
  id: string
  user_id: string
  workspace_id: string
  content: string
  reasoning?: string
  provider?: string
  model?: string
  generated_at: string
  stats?: FollowedEntityStats
  trending?: TrendingEntity[]
  followed?: FollowedEntity[]
  cached?: boolean
}

export interface ActivityBackfillStatus {
  entity_id: string
  workspace_id: string
  title: string
  existing_ref_count: number
  message_candidate_count: number
  file_candidate_count: number
  missing_ref_count: number
  created_ref_count: number
  is_backfilled: boolean
  last_ref_at?: string
}

// ── Phase 63A ────────────────────────────────────────────────────────────────

export interface Citation {
  id: string
  evidence_kind: string
  source_kind: string
  source_ref: string
  ref_kind: string
  locator?: string
  snippet: string
  title?: string
  score: number
  entity_id?: string
  entity_title?: string
}

export interface EntityAnswer {
  entity: KnowledgeEntity
  question: string
  answer: string
  reasoning?: string
  provider: string
  model: string
  answered_at: string
  citations: Citation[]
}

export interface SharedWeeklyBriefLink {
  id: string
  user_id: string
  workspace_id: string
  title: string
  url: string
  short_url: string
  relative_path: string
}

export interface StaleBriefNotice {
  entity_id: string
  workspace_id: string
  title?: string
  reason?: string
  changed_at: string
  stale: true
}

export interface EntityTextMatch {
  entity_id: string
  entity_title: string
  entity_kind: string
  source_kind?: string
  matched_text: string
  start: number
  end: number
}

export interface MessageByEntityResult {
  id: string
  channel_id?: string
  channel_name?: string
  sender_id?: string
  sender_name?: string
  content: string
  snippet?: string
  created_at: string
  match_sources: MatchSource[]
  metadata?: Record<string, unknown>
}

export interface ChannelKnowledgeDigestEntry {
  entity_id: string
  entity_title: string
  entity_kind: string
  ref_count: number
  delta?: number
  summary?: string
  last_activity_at?: string
  top_sources?: { source_kind: string; source_id?: string; snippet?: string }[]
}

export interface ChannelKnowledgeDigest {
  channel_id: string
  window: 'daily' | 'weekly' | 'monthly'
  window_days: number
  generated_at: string
  total_refs: number
  delta?: number
  entries: ChannelKnowledgeDigestEntry[]
  headline?: string
  summary?: string
  top_movements?: ChannelKnowledgeDigestEntry[]
}

export type DigestWindow = 'daily' | 'weekly' | 'monthly'

export interface DigestSchedule {
  channel_id: string
  window: DigestWindow
  timezone: string
  day_of_week?: number
  day_of_month?: number
  hour: number
  minute: number
  limit: number
  pin: boolean
  is_enabled: boolean
  last_published_at?: string
  next_run_at?: string
  created_at?: string
  updated_at?: string
}

export interface DigestScheduleInput {
  window: DigestWindow
  timezone: string
  day_of_week?: number
  day_of_month?: number
  hour: number
  minute: number
  limit: number
  pin: boolean
  is_enabled: boolean
}

export type KnowledgeInboxScope = 'all' | 'starred'

export interface KnowledgeInboxChannel {
  id: string
  name?: string
  is_starred?: boolean
}

export interface KnowledgeInboxItem {
  id: string
  channel: KnowledgeInboxChannel
  message: Message
  digest: ChannelKnowledgeDigest
  is_read: boolean
  occurred_at: string
}

export interface KnowledgeInboxEntityContext {
  entity_id: string
  entity_title: string
  entity_kind: string
  delta: number
  messages: Message[]
}

export interface KnowledgeInboxDetail {
  item: KnowledgeInboxItem
  entity_contexts: KnowledgeInboxEntityContext[]
}

export interface DigestScheduleUpcomingRun {
  run_at: string
}

export interface DigestSchedulePreview {
  schedule: DigestSchedule
  upcoming_runs: DigestScheduleUpcomingRun[]
  digest: ChannelKnowledgeDigest
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

export interface MessageEntityMention {
  entity_id: string
  entity_title: string
  entity_kind: string
  source_kind?: string
  mention_text: string
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
  metadata?: {
    entity_mentions?: MessageEntityMention[]
    knowledge_digest?: ChannelKnowledgeDigest
    [key: string]: unknown
  }
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
