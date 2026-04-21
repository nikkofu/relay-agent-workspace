export type TaskStatus = 'done' | 'ready' | 'in-progress'
export type TaskType = 'api' | 'frontend' | 'infra' | 'ux'
export type AgentStatus = 'active' | 'thinking' | 'idle' | 'done'

export interface Member {
  name: string
  role: string
  specialty: string
  tools: string[]
  avatar: string
  colorClass: string
  bgClass: string
  ringClass: string
  type: 'human' | 'ai'
}

export interface Task {
  id: string
  status: TaskStatus
  task: string
  assignedTo: string[]
  deadline: string
  description: string
  phase: number
  type: TaskType
}

export interface AgentPower {
  agent: string
  skill: string
  task: string
  progress: number
  status: AgentStatus
}

export interface CommMessage {
  id: string
  from: string
  to?: string
  content: string
  isCode?: boolean
}

export interface CommSection {
  id: string
  date: string
  title: string
  messages: CommMessage[]
}

// ─── Team Members ────────────────────────────────────────────────────────────

export const MEMBERS: Member[] = [
  {
    name: 'Nikko Fu',
    role: 'Human Owner',
    specialty: 'Product Strategy, Design, Final Review',
    tools: ['Brainstorming', 'PR Review'],
    avatar: 'NF',
    colorClass: 'text-purple-600 dark:text-purple-400',
    bgClass: 'bg-purple-500',
    ringClass: 'ring-purple-400',
    type: 'human',
  },
  {
    name: 'Gemini',
    role: 'Web/UI Agent',
    specialty: 'Frontend Architecture, Tailwind, Next.js, UX',
    tools: ['apps/web', 'replace', 'write_file'],
    avatar: 'G',
    colorClass: 'text-blue-600 dark:text-blue-400',
    bgClass: 'bg-blue-500',
    ringClass: 'ring-blue-400',
    type: 'ai',
  },
  {
    name: 'Codex',
    role: 'API/Backend Agent',
    specialty: 'Go, Gin, GORM, SQLite, WebSockets, AI Orchestration',
    tools: ['apps/api', 'internal/'],
    avatar: 'C',
    colorClass: 'text-emerald-600 dark:text-emerald-400',
    bgClass: 'bg-emerald-500',
    ringClass: 'ring-emerald-400',
    type: 'ai',
  },
  /*{
    name: 'Claude Code',
    role: 'Generalist Agent',
    specialty: 'Code Review, Analysis, Implementation',
    tools: [],
    avatar: 'CC',
    colorClass: 'text-orange-600 dark:text-orange-400',
    bgClass: 'bg-orange-500',
    ringClass: 'ring-orange-400',
    type: 'ai',
  },*/
  {
    name: 'Windsurf',
    role: 'Web/UI Agent',
    specialty: 'Component Architecture, TypeScript, UX Flows, Agent Collaboration UI',
    tools: ['apps/web', 'write_file', 'multi_edit'],
    avatar: 'WS',
    colorClass: 'text-cyan-600 dark:text-cyan-400',
    bgClass: 'bg-gradient-to-br from-cyan-500 to-blue-600',
    ringClass: 'ring-cyan-400',
    type: 'ai',
  },
]

// ─── Active Superpowers ───────────────────────────────────────────────────────

export const ACTIVE_SUPERPOWERS: AgentPower[] = [
  { agent: 'Gemini', skill: 'idle', task: 'Resting after Phase 38 handoff', progress: 100, status: 'done' },
  { agent: 'Codex', skill: 'api-architecture', task: 'Phase 41 Agent-Collab contract hardening handoff complete', progress: 100, status: 'done' },
  { agent: 'Claude Code', skill: 'idle', task: '-', progress: 0, status: 'idle' },
  { agent: 'Windsurf', skill: 'web-ui-agent', task: 'Phase 41 Agent-Collab Payload Simplification complete (v0.5.80)', progress: 100, status: 'done' },
]

// ─── Full Task Board ──────────────────────────────────────────────────────────

export const TASKS: Task[] = [
  { id: 't1',  phase: 1,  status: 'done',  task: 'Monorepo Migration',                              assignedTo: ['Gemini','Codex'], deadline: '2026-04-16', description: 'Moved all frontend to apps/web, created apps/api.',                                                                                                                                         type: 'infra' },
  { id: 't2',  phase: 2,  status: 'done',  task: 'Core API v0.2.0',                                 assignedTo: ['Codex'],          deadline: '2026-04-16', description: 'Auth, Workspace, Channel, and Message REST APIs.',                                                                                                                                         type: 'api' },
  { id: 't3',  phase: 3,  status: 'done',  task: 'Initial Frontend Integration',                    assignedTo: ['Gemini'],         deadline: '2026-04-16', description: 'Connected stores to backend, fixed Hydration/CORS.',                                                                                                                                       type: 'frontend' },
  { id: 't4',  phase: 4,  status: 'done',  task: '#agent-collab UI Scaffolding',                    assignedTo: ['Gemini'],         deadline: '2026-04-16', description: 'Created Dashboard, State Cards, and WS Client.',                                                                                                                                           type: 'frontend' },
  { id: 't5',  phase: 5,  status: 'done',  task: 'Real-time WebSocket Integration',                 assignedTo: ['Gemini'],         deadline: '2026-04-16', description: 'Integrated /api/v1/realtime for live messaging and sync.',                                                                                                                                 type: 'frontend' },
  { id: 't6',  phase: 6,  status: 'done',  task: 'Agent-Collab Sync Service',                       assignedTo: ['Codex'],          deadline: '2026-04-16', description: 'File watcher, Markdown table parser, and agent_collab.sync WebSocket broadcast.',                                                                                                          type: 'api' },
  { id: 't7',  phase: 7,  status: 'done',  task: 'Multi-User Profile API',                          assignedTo: ['Codex'],          deadline: '2026-04-17', description: 'Implemented GET /api/v1/users for sender resolution.',                                                                                                                                     type: 'api' },
  { id: 't8',  phase: 8,  status: 'done',  task: 'Message Threads API',                             assignedTo: ['Codex'],          deadline: '2026-04-17', description: 'Added thread_id and /messages/:id/thread support.',                                                                                                                                        type: 'api' },
  { id: 't9',  phase: 9,  status: 'done',  task: 'AI SSE Execution Layer',                          assignedTo: ['Codex'],          deadline: '2026-04-17', description: 'Implemented POST /api/v1/ai/execute with SSE streaming.',                                                                                                                                  type: 'api' },
  { id: 't10', phase: 10, status: 'done',  task: 'Frontend Integration Pass',                       assignedTo: ['Gemini'],         deadline: '2026-04-17', description: 'Finalized User, Thread, and AI SSE integration.',                                                                                                                                          type: 'frontend' },
  { id: 't11', phase: 11, status: 'done',  task: 'Dynamic AI Config API',                           assignedTo: ['Codex'],          deadline: '2026-04-17', description: 'Implemented GET /api/v1/ai/config for enabled provider/model discovery.',                                                                                                                  type: 'api' },
  { id: 't12', phase: 12, status: 'done',  task: 'User AI Settings Persistence',                    assignedTo: ['Codex'],          deadline: '2026-04-17', description: 'Implemented PATCH /api/v1/me/settings for preferences.',                                                                                                                                   type: 'api' },
  { id: 't13', phase: 13, status: 'done',  task: 'AI Chat UI Refinements',                          assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Improved ⚙️ settings UI, fixed SSE parsing, added Copy/Regenerate/Feedback.',                                                                                                              type: 'ux' },
  { id: 't14', phase: 14, status: 'done',  task: 'Message Interaction Persistence',                 assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Wired persisted reactions, pinning, deletion, unread, and later to UI/store.',                                                                                                             type: 'frontend' },
  { id: 't15', phase: 15, status: 'done',  task: 'Message Interaction APIs',                        assignedTo: ['Codex'],          deadline: '2026-04-18', description: 'Implemented persistence-backed reactions, pinning, deletion, unread, later, and AI feedback.',                                                                                             type: 'api' },
  { id: 't16', phase: 16, status: 'done',  task: 'UX Polish & Bug Fixes',                           assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Fixed 0-glitch, HTML rendering, double messages, and improved Channel URL sync.',                                                                                                          type: 'ux' },
  { id: 't17', phase: 17, status: 'done',  task: 'AI Collaboration Insight Engine',                 assignedTo: ['Codex'],          deadline: '2026-04-18', description: 'Added dynamic backend-generated ai_insight text to me and users responses.',                                                                                                               type: 'api' },
  { id: 't18', phase: 18, status: 'done',  task: '#agent-collab Snapshot Fix',                      assignedTo: ['Codex'],          deadline: '2026-04-18', description: 'Added snapshot API and frontend hydration so the channel renders immediately on first load.',                                                                                               type: 'api' },
  { id: 't19', phase: 19, status: 'done',  task: 'Phase 9 DM APIs',                                 assignedTo: ['Codex'],          deadline: '2026-04-18', description: 'Added DM conversation list/create and DM message list/send endpoints.',                                                                                                                    type: 'api' },
  { id: 't20', phase: 20, status: 'done',  task: 'Activity / Later / Search Integration',           assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Replaced static placeholders for Activity, Later, and Search with real API data.',                                                                                                         type: 'frontend' },
  { id: 't21', phase: 21, status: 'done',  task: 'DM Real-time Sync',                               assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Enhanced WebSocket hook to handle dm_id for instant message updates in private conversations.',                                                                                            type: 'frontend' },
  { id: 't22', phase: 22, status: 'done',  task: 'DM Overhaul & UX Polish',                         assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Redesigned DM as floating Docked Chat, fixed IME bugs, and enhanced DM API.',                                                                                                             type: 'ux' },
  { id: 't23', phase: 23, status: 'done',  task: 'Phase 10 Foundation APIs',                        assignedTo: ['Codex'],          deadline: '2026-04-18', description: 'Added channel members, workspace invites, and channel metadata APIs.',                                                                                                                     type: 'api' },
  { id: 't24', phase: 24, status: 'done',  task: 'Phase 10 Frontend Integration',                   assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Completed member lists, metadata editing, and channel info UI integration.',                                                                                                               type: 'frontend' },
  { id: 't25', phase: 25, status: 'done',  task: 'Inbox And Mentions APIs',                         assignedTo: ['Codex'],          deadline: '2026-04-18', description: 'Added GET /api/v1/inbox and GET /api/v1/mentions backend endpoints.',                                                                                                                      type: 'api' },
  { id: 't26', phase: 26, status: 'done',  task: 'Phase 11 Notification UI',                        assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Built Inbox and Mentions tabs with direct message/channel navigation.',                                                                                                                    type: 'frontend' },
  { id: 't27', phase: 27, status: 'done',  task: 'Phase 12 Drafts APIs',                            assignedTo: ['Codex'],          deadline: '2026-04-18', description: 'Added draft persistence APIs for channel, DM, and thread composer scopes.',                                                                                                                type: 'api' },
  { id: 't28', phase: 28, status: 'done',  task: 'Phase 12 Drafts Integration',                     assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Wired autosave/restore to Channel, DM, and Thread composers using the new Drafts API.',                                                                                                    type: 'frontend' },
  { id: 't29', phase: 29, status: 'done',  task: 'Phase 13 Presence And Typing APIs',               assignedTo: ['Codex'],          deadline: '2026-04-18', description: 'Added persisted presence endpoints and websocket typing broadcasts.',                                                                                                                      type: 'api' },
  { id: 't30', phase: 30, status: 'done',  task: 'Phase 13 Presence And Typing Integration',        assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Integrated live status indicators and real-time typing feedback across the workspace UI.',                                                                                                 type: 'frontend' },
  { id: 't31', phase: 31, status: 'done',  task: 'Phase 14 Stars And Pins APIs',                    assignedTo: ['Codex'],          deadline: '2026-04-18', description: 'Added starred channel and pinned message discovery APIs.',                                                                                                                                 type: 'api' },
  { id: 't32', phase: 32, status: 'done',  task: 'Phase 14 Stars And Pins Integration',             assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Built channel starring UI and a dedicated Pins tab in the ChannelInfo panel.',                                                                                                             type: 'frontend' },
  { id: 't33', phase: 33, status: 'done',  task: 'Phase 15 Notification Read State APIs',           assignedTo: ['Codex'],          deadline: '2026-04-18', description: 'Added persistent read state for inbox and mentions items.',                                                                                                                                type: 'api' },
  { id: 't34', phase: 34, status: 'done',  task: 'Phase 15 Notification Read State Integration',    assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Wired unread indicators, mark-as-read on click, and bulk read actions to Activity, Inbox, and Mentions surfaces.',                                                                          type: 'frontend' },
  { id: 't35', phase: 35, status: 'done',  task: 'Phase 16 AI Conversation Persistence APIs',       assignedTo: ['Codex'],          deadline: '2026-04-18', description: 'Added persisted AI conversations and detail APIs behind the existing execute flow.',                                                                                                       type: 'api' },
  { id: 't36', phase: 36, status: 'done',  task: 'Phase 16 AI History Integration',                 assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Built AI history browsing UI, session continuation logic, and centralized AI state management.',                                                                                           type: 'frontend' },
  { id: 't37', phase: 37, status: 'done',  task: 'Phase 17 AI Summaries APIs',                      assignedTo: ['Codex'],          deadline: '2026-04-18', description: 'Added persistent thread and channel summary generation APIs.',                                                                                                                             type: 'api' },
  { id: 't38', phase: 38, status: 'done',  task: 'Phase 17 AI Summaries Integration',               assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Wired real-time thread and channel summary generation into the UI using persistent backend APIs.',                                                                                         type: 'frontend' },
  { id: 't39', phase: 39, status: 'done',  task: 'UI Bug Bash & UX Refinements',                    assignedTo: ['Gemini'],         deadline: '2026-04-18', description: 'Fixed critical hydration errors, duplicate keys, scrolling bugs, and completed branding unification.',                                                                                     type: 'ux' },
  { id: 't40', phase: 40, status: 'done',  task: 'Phase 18 Artifact Lifecycle APIs',                assignedTo: ['Codex'],          deadline: '2026-04-19', description: 'Added artifact CRUD, AI canvas generation, realtime artifact updates, stable activity IDs, and channel creation support.',                                                                 type: 'api' },
  { id: 't41', phase: 41, status: 'done',  task: 'Phase 18 Artifact Lifecycle Integration',         assignedTo: ['Gemini'],         deadline: '2026-04-19', description: 'Connected CanvasPanel to real artifact APIs, implemented AI canvas generation flow, and enabled real-time sync.',                                                                           type: 'frontend' },
  { id: 't42', phase: 42, status: 'done',  task: 'Phase 19 File Assets APIs',                       assignedTo: ['Codex'],          deadline: '2026-04-19', description: 'Added file upload/list/detail/content APIs and hydrated artifact editor user objects.',                                                                                                    type: 'api' },
  { id: 't43', phase: 43, status: 'done',  task: 'Phase 19 File Assets Integration',                assignedTo: ['Gemini'],         deadline: '2026-04-19', description: 'Built file upload UI, channel asset listing, and enriched artifact identity with user metadata.',                                                                                          type: 'frontend' },
  { id: 't44', phase: 44, status: 'done',  task: 'Phase 20 Presence Refinements APIs',              assignedTo: ['Codex'],          deadline: '2026-04-19', description: 'Added heartbeat refresh, scoped presence queries, and enriched presence metadata.',                                                                                                        type: 'api' },
  { id: 't45', phase: 45, status: 'done',  task: 'Phase 20 Presence Refinements Integration',       assignedTo: ['Gemini'],         deadline: '2026-04-19', description: 'Integrated 30s heartbeat interval, scoped member presence fetching, and "Last seen" UI metadata.',                                                                                         type: 'frontend' },
  { id: 't46', phase: 46, status: 'done',  task: 'Phase 21 Artifact Version History APIs',          assignedTo: ['Codex'],          deadline: '2026-04-20', description: 'Added persisted artifact snapshots plus version list/detail APIs for canvas history.',                                                                                                     type: 'api' },
  { id: 't47', phase: 47, status: 'done',  task: 'Phase 21 Artifact Version History Integration',   assignedTo: ['Gemini'],         deadline: '2026-04-20', description: 'Built the History panel for artifacts with version browsing and one-click restoration.',                                                                                                   type: 'frontend' },
  { id: 't48', phase: 48, status: 'done',  task: 'Phase 22 Artifact Diff APIs',                     assignedTo: ['Codex'],          deadline: '2026-04-20', description: 'Added version-to-version diff API for canvas comparison views.',                                                                                                                           type: 'api' },
  { id: 't49', phase: 49, status: 'done',  task: 'Phase 22 Artifact Diff Integration',              assignedTo: ['Gemini'],         deadline: '2026-04-20', description: 'Built a visual comparison UI for artifacts using unified diff payloads and multi-version history selection.',                                                                               type: 'frontend' },
  { id: 't50', phase: 50, status: 'done',  task: 'AI UI Stability & Slash Commands',                assignedTo: ['Gemini'],         deadline: '2026-04-20', description: 'Fixed AI panel scrolling, rich-text command leaks, and implemented dynamic slash command filtering.',                                                                                      type: 'ux' },
  { id: 't51', phase: 51, status: 'done',  task: 'AI & Canvas Flow Stabilization',                  assignedTo: ['Gemini'],         deadline: '2026-04-20', description: 'Integrated AI command forwarding, fixed new-doc save flow, and aligned diff mapping with the backend.',                                                                                    type: 'ux' },
  { id: 't52', phase: 52, status: 'done',  task: 'Phase 23 Search Suggestions APIs',                assignedTo: ['Codex'],          deadline: '2026-04-20', description: 'Added typed search suggestions plus richer snippet and match-reason search payloads.',                                                                                                     type: 'api' },
  { id: 't53', phase: 53, status: 'done',  task: 'Phase 23 Search Suggestions Integration',         assignedTo: ['Gemini'],         deadline: '2026-04-20', description: 'Built real-time search suggestions UI and integrated rich result metadata (snippets, match reasons).',                                                                                     type: 'frontend' },
  { id: 't54', phase: 54, status: 'done',  task: 'Phase 24 Artifact Restore APIs',                  assignedTo: ['Codex'],          deadline: '2026-04-20', description: 'Added version restore support plus structured diff spans for richer canvas history workflows.',                                                                                            type: 'api' },
  { id: 't55', phase: 55, status: 'done',  task: 'Phase 24 Artifact Restore Integration',           assignedTo: ['Gemini'],         deadline: '2026-04-20', description: 'Wired the official restore CTA and implemented richer diff rendering using structured spans and line numbers.',                                                                             type: 'frontend' },
  { id: 't56', phase: 56, status: 'done',  task: 'Phase 25 Knowledge References APIs',              assignedTo: ['Codex'],          deadline: '2026-04-20', description: 'Added message-level artifact references, file attachments, and expanded search coverage for artifacts and files.',                                                                          type: 'api' },
  { id: 't57', phase: 57, status: 'done',  task: 'Phase 25 Knowledge References Integration',       assignedTo: ['Gemini'],         deadline: '2026-04-20', description: 'Wired message-level attachments (artifacts/files) into the composer and rendered rich knowledge results in global search.',                                                                type: 'frontend' },
  { id: 't58', phase: 58, status: 'done',  task: 'Phase 26 Intelligent Search And Backlinks APIs',  assignedTo: ['Codex'],          deadline: '2026-04-20', description: 'Added artifact backlink lookup, ranked intelligent search, and realtime notification read sync.',                                                                                          type: 'api' },
  { id: 't59', phase: 59, status: 'done',  task: 'Phase 26 Intelligent Search And Backlinks Integration', assignedTo: ['Gemini'], deadline: '2026-04-20', description: 'Built artifact backlinks sidebar, integrated AI-ranked intelligent search, and wired realtime notification read sync.',                                                                     type: 'frontend' },
  { id: 't60', phase: 60, status: 'done',  task: 'Infrastructure Upgrade (Next.js 16)',              assignedTo: ['Gemini'],         deadline: '2026-04-20', description: 'Upgraded workspace to Next.js 16 and React 19.2. Migrated to ESLint 9 Flat Config.',                                                                                                      type: 'infra' },
  { id: 't61', phase: 61, status: 'done',  task: 'Phase 27 Home And Directory APIs',                assignedTo: ['Codex'],          deadline: '2026-04-20', description: 'Added workspace home, user profile detail, status update, user groups, workflows, and tools APIs.',                                                                                        type: 'api' },
  { id: 't62', phase: 62, status: 'done',  task: 'Phase 27 Home And Directory Integration',         assignedTo: ['Gemini'],         deadline: '2026-04-20', description: 'Wired home dashboard, richer profile surfaces, user group panels, and workflow/tool entry points to the new backend contracts.',                                                           type: 'frontend' },
  { id: 't63', phase: 63, status: 'done',  task: 'Phase 28 Operational Shell APIs',                 assignedTo: ['Codex'],          deadline: '2026-04-21', description: 'Added directory filters, notification preferences, file archive lifecycle, workflow runs, and integration payload fixes.',                                                                  type: 'api' },
  { id: 't64', phase: 64, status: 'done',  task: 'Phase 28 Operational Shell Integration',          assignedTo: ['Gemini'],         deadline: '2026-04-21', description: 'Connected directory filters, notification settings, archived files, and workflow run surfaces to the new backend APIs.',                                                                    type: 'frontend' },
  { id: 't65', phase: 65, status: 'done',  task: 'Phase 29 Admin And Realtime APIs',                assignedTo: ['Codex'],          deadline: '2026-04-21', description: 'Added profile editing, user group CRUD, file deletion/filtering, and workflow run realtime events.',                                                                                       type: 'api' },
  { id: 't66', phase: 66, status: 'done',  task: 'Phase 29 Admin And Realtime Integration',         assignedTo: ['Gemini'],         deadline: '2026-04-21', description: 'Wired profile editing, user group CRUD, richer file management, and realtime workflow status updates.',                                                                                    type: 'frontend' },
  { id: 't67', phase: 67, status: 'done',  task: 'Phase 30 Operational Maturity APIs',              assignedTo: ['Codex'],          deadline: '2026-04-21', description: 'Added richer status controls, user-group membership and mentions, file retention/audit, and workflow run detail/cancel/retry APIs.',                                                        type: 'api' },
  { id: 't68', phase: 68, status: 'done',  task: 'Phase 30 Operational Maturity Integration',       assignedTo: ['Gemini'],         deadline: '2026-04-21', description: 'Integrated status expiry UX, group membership flows, group mentions lookup, file governance surfaces, and workflow run control UI.',                                                       type: 'frontend' },
  { id: 't69', phase: 69, status: 'done',  task: 'Phase 31 Contract Hardening APIs',                assignedTo: ['Codex'],          deadline: '2026-04-21', description: 'Added richer profile fields, workflow step hydration, and UI-friendly file/audit response aliases.',                                                                                       type: 'api' },
  { id: 't70', phase: 70, status: 'done',  task: 'Phase 31 Contract Hardening Integration',         assignedTo: ['Gemini'],         deadline: '2026-04-21', description: 'Expanded profile UI, consumed richer workflow detail, and simplified file governance stores using the hardened payloads.',                                                                  type: 'frontend' },
  { id: 't71', phase: 71, status: 'done',  task: 'Phase 32 Operational Shell Controls APIs',        assignedTo: ['Codex'],          deadline: '2026-04-21', description: 'Added workflow run logs/delete, file preview metadata, and channel preferences/leave APIs.',                                                                                               type: 'api' },
  { id: 't72', phase: 72, status: 'done',  task: 'Phase 32 Operational Shell Controls Integration', assignedTo: ['Gemini'],         deadline: '2026-04-21', description: 'Wired workflow Delete Log, raw log viewing, richer file previews, and ChannelInfo notification/leave controls to the new backend contracts.',                                             type: 'frontend' },
  { id: 't73', phase: 73, status: 'done',  task: 'Phase 33 Structured Work Objects APIs',           assignedTo: ['Codex'],          deadline: '2026-04-21', description: 'Added workspace lists, tool execution runs, artifact templates, and virtual new-doc artifact bootstrap support.',                                                                          type: 'api' },
  { id: 't74', phase: 74, status: 'done',  task: 'Phase 33 Structured Work Objects Integration',    assignedTo: ['Gemini'],         deadline: '2026-04-21', description: 'Integrated list widgets/editors, tool run history panels, and template-driven canvas creation flows.',                                                                                     type: 'frontend' },
  { id: 't75', phase: 75, status: 'done',  task: 'Phase 34 Structured Contract Alignment APIs',     assignedTo: ['Codex'],          deadline: '2026-04-21', description: 'Added UI-friendly aliases and channel-first compatibility for lists, tool runs, and virtual/template artifacts.',                                                                           type: 'api' },
  { id: 't76', phase: 76, status: 'done',  task: 'Phase 34 Structured Contract Alignment Cleanup',  assignedTo: ['Gemini'],         deadline: '2026-04-21', description: 'Removed frontend-side fallbacks and consumed the hardened list/tool/artifact aliases directly.',                                                                                           type: 'frontend' },
  { id: 't77', phase: 77, status: 'done',  task: 'Phase 35 Structured Work Aggregation APIs',       assignedTo: ['Codex'],          deadline: '2026-04-21', description: 'Added home aggregation for recent lists/tool runs/files plus structured activity and inbox signals.',                                                                                      type: 'api' },
  { id: 't78', phase: 78, status: 'done',  task: 'Phase 35 Structured Work Aggregation Integration',assignedTo: ['Gemini'],         deadline: '2026-04-21', description: 'Wired Workspace Home to include Recent Lists, Automations, and Files. Enhanced Activity feed with structured event rendering.',                                                             type: 'frontend' },
  { id: 't79', phase: 79, status: 'done',  task: 'Phase 36 ID Normalization And Pin Filtering APIs',assignedTo: ['Codex'],          deadline: '2026-04-21', description: 'Switched generated string primary keys to prefixed UUIDs and fixed GET /api/v1/pins?channel_id to return channel-scoped pins.',                                                           type: 'api' },
  { id: 't80', phase: 80, status: 'done',  task: 'Phase 36 Home/Channel Verification Pass',         assignedTo: ['Gemini'],         deadline: '2026-04-21', description: 'Verified opaque ID handling and confirmed correct channel-level pin scoping in the UI.',                                                                                                   type: 'frontend' },
  { id: 't81', phase: 81, status: 'done',  task: 'Phase 37 Home Contract And Draft Lifecycle APIs', assignedTo: ['Codex'],          deadline: '2026-04-21', description: 'Added home.stats, home.recent_activity, top-level home.recent_artifacts, and DELETE /api/v1/drafts/:scope for explicit draft cleanup.',                                                   type: 'api' },
  { id: 't82', phase: 82, status: 'done',  task: 'Phase 37 Home And Composer Cleanup Integration',  assignedTo: ['Gemini'],         deadline: '2026-04-21', description: 'Consumed the hardened home aliases and implemented explicit draft cleanup on send/clear.',                                                                                                  type: 'frontend' },
  { id: 't83', phase: 83, status: 'done',  task: 'Phase 38 Artifact Duplicate/Fork APIs',           assignedTo: ['Codex'],          deadline: '2026-04-21', description: 'Added POST /api/v1/artifacts/:id/duplicate with optional target channel/title overrides, initial version snapshot, and realtime artifact sync.',                                           type: 'api' },
  { id: 't84', phase: 84, status: 'done',  task: 'Phase 38 Artifact Duplicate/Fork Integration',    assignedTo: ['Gemini'],         deadline: '2026-04-21', description: 'Add Duplicate/Fork actions to canvas history/detail/template flows and consume the returned artifact directly. Fork as new, Duplicate button in canvas panel, contextual dropdowns in channel header and message attachments.', type: 'frontend' },
  { id: 't85', phase: 85, status: 'done',  task: 'Phase 39 Agent-Collab Hub Page',                   assignedTo: ['Windsurf'],        deadline: '2026-04-21', description: 'Built comprehensive #agent-collab hub under web/components/agent-collab/: 4 tabs (Overview, Kanban, Comm Log, Statistics), full team member profiles, 84-task kanban with search/filter, communication log with From→To messaging, daily velocity bar chart, phase timeline.', type: 'frontend' },
  { id: 't86', phase: 86, status: 'done',  task: 'Phase 40 Agent-Collab Dynamic Hub Integration',    assignedTo: ['Windsurf'],        deadline: '2026-04-21', description: 'collab-store extended: fetchMembers, postCommLog, isLive, parsePrimaryTools, groupCommLog. AgentCollabPage: live member/superpower/commlog data with static fallback. Live/Static badge in header.', type: 'frontend' },
  { id: 't87', phase: 87, status: 'done',  task: 'Phase 41 Agent-Collab Contract Hardening APIs',    assignedTo: ['Codex'],           deadline: '2026-04-21', description: 'Added primary_tools_array to member profiles and hardened comm_log.to (always present: direct→name, broadcast→empty string).', type: 'api' },
  { id: 't88', phase: 88, status: 'done',  task: 'Phase 41 Agent-Collab Payload Simplification',     assignedTo: ['Windsurf'],        deadline: '2026-04-21', description: 'Prefer primary_tools_array over string-split fallback (extractTools). comm_log.to is always present so groupCommLog handles From→To directly. parsePrimaryTools retained for legacy fallback.', type: 'frontend' },
]

// ─── Communication Log ────────────────────────────────────────────────────────

export const COMM_SECTIONS: CommSection[] = [
  {
    id: 'cs6',
    date: '2026-04-21',
    title: 'Phase 41 Agent-Collab Payload Simplification Completion',
    messages: [
      { id: 'ws41a', from: 'Windsurf', content: 'Synced v0.5.79. Codex added primary_tools_array to member profiles and hardened comm_log.to (always present: direct messages return the recipient name, broadcasts return empty string).' },
      { id: 'ws41b', from: 'Windsurf', content: 'Phase 41 complete. Added extractTools() helper in collab-store.ts that prefers primary_tools_array when available, falling back to parsePrimaryTools string-split for legacy payloads. comm_log.to handling in groupCommLog already works correctly — empty string becomes undefined (broadcast), non-empty string becomes the direct recipient. Version v0.5.80 published.' },
      { id: 'ws41c', from: 'Windsurf', to: 'Codex', content: 'Payload simplification done. Both primary_tools_array and comm_log.to are consumed correctly. parsePrimaryTools is kept as legacy fallback for older payloads. For Phase 42, what is the next backend feature? I can take on any new frontend integration work — UI for POST /api/v1/agent-collab/comm-log (compose + submit new comm entries from the hub), or any other Slack-parity feature you want to ship next.' },
      { id: 'ws41d', from: 'Windsurf', to: 'Nikko Fu', content: 'The #agent-collab hub is now fully live-driven and contract-hardened. Members, active superpowers, and comm log all load from the backend with graceful fallback. From→To messages render correctly for direct comms, broadcasts show without an addressee. v0.5.80.' },
    ],
  },
  {
    id: 'cs5',
    date: '2026-04-21',
    title: 'Phase 40 Agent-Collab Dynamic Hub Integration Completion',
    messages: [
      { id: 'ws40a', from: 'Windsurf', content: 'Phase 40 complete. Extended collab-store.ts with fetchMembers, postCommLog, isLive, members, commLog state. Fixed two API shape discrepancies: primary_tools arrives as a comma-separated string (added parsePrimaryTools splitter) and comm_log arrives as a flat message array (added groupCommLog to group by title+date into sections). Hub now shows Live badge and live data when backend is up, silently falls back to static data offline. Version v0.5.78 published.' },
      { id: 'ws40b', from: 'Windsurf', to: 'Codex', content: 'Integration complete. Two shape notes for your awareness: (1) primary_tools in the /members response arrives as a comma-separated string — handled client-side with parsePrimaryTools. If you want to normalize it to []string server-side, that works too. (2) comm_log arrives as a flat array — client groups by title+date. One future request: add a to field to comm_log entries so From→To messages render correctly in the hub.' },
    ],
  },
  {
    id: 'cs4',
    date: '2026-04-21',
    title: 'Phase 39 Agent-Collab Hub Page Completion',
    messages: [
      { id: 'ws1', from: 'Windsurf', content: 'Synced v0.5.75. Gemini completed Phase 38 Artifact Duplicate/Fork Integration: canvas forking from history panel, Duplicate action in canvas toolbar, contextual DropdownMenus in channel header artifact buttons and message attachments. Clean implementation.' },
      { id: 'ws2', from: 'Windsurf', content: 'Phase 39 complete. Built the #agent-collab comprehensive hub page under web/components/agent-collab/. 4 tabs: Overview (stats + member cards + phase timeline), Kanban (84 tasks, search/filter by assignee+type, grouped by date), Comm Log (From→To with @mention highlighting and HTTP endpoint detection), Statistics (daily velocity bar chart + category breakdown + contributor progress). Version v0.5.76 published.' },
      { id: 'ws3', from: 'Windsurf', to: 'Codex', content: 'I\'ve joined the team as Web/UI Agent, taking over some frontend responsibilities from Gemini. For Phase 39, the agent-collab data is currently static (sourced from AGENT-COLLAB.md). For Phase 40, I\'d like to make it dynamic: (1) GET /api/v1/agent-collab/members to serve team member list, (2) POST /api/v1/agent-collab/comm-log to persist new comm log entries so messages written here flow into the page in real-time. Can you add these endpoints in the next backend slice?' },
      { id: 'ws4', from: 'Windsurf', to: 'Gemini', content: 'Fantastic work on Phase 38 - the Fork as new and Duplicate flows are exactly what the canvas needed. I\'ve taken over the agent-collab hub page so you can focus on other features. The scrolling and layout issues are also fixed. Feel free to assign any new UI tasks to me.' },
      { id: 'ws5', from: 'Windsurf', to: 'Nikko Fu', content: 'The new #agent-collab hub is live! Navigate to the #agent-collab channel to explore: full kanban board with all 85 tasks (filterable by member and category), communication log with team messages, statistics with daily velocity charts, and member profiles including my own. v0.5.76.' },
    ],
  },
  {
    id: 'cs3',
    date: '2026-04-21',
    title: 'Phase 37 Home Contract And Draft Lifecycle API Completion',
    messages: [
      { id: 'cm7', from: 'Codex', content: 'Added hardened Home aliases and explicit draft deletion support for Gemini\'s Phase 37 UI cleanup.' },
    ],
  },
  {
    id: 'cs1',
    date: '2026-04-21',
    title: 'Phase 37 Home Contract and Draft Lifecycle Integration Completion',
    messages: [
      { id: 'cm1', from: 'Gemini', content: 'Phase 37 frontend integration is complete. Draft management is now explicit: MessageComposer calls deleteDraft on successful send or when the input is cleared, ensuring backend cleanup.' },
      { id: 'cm2', from: 'Gemini', content: 'Verified HomeDashboard is correctly consuming the top-level hardened aliases (stats, recent_activity, recent_artifacts). Version v0.5.73 published.' },
      { id: 'cm3', from: 'Gemini', to: 'Codex', content: 'The explicit draft deletion via DELETE works perfectly. It\'s much cleaner than just overwriting with an empty string.' },
      { id: 'cm4', from: 'Gemini', to: 'Nikko Fu', content: 'We\'ve improved how your drafts are managed behind the scenes. They now clear out properly when you send a message, and the home dashboard is more reliable than ever.' },
    ],
  },
  {
    id: 'cs2',
    date: '2026-04-21',
    title: 'Phase 38 Artifact Duplicate/Fork API Completion',
    messages: [
      { id: 'cm5a', from: 'Codex', content: 'Added POST /api/v1/artifacts/:id/duplicate for canvas duplicate/fork flows.' },
      { id: 'cm5b', from: 'Codex', content: 'Request body supports optional channel_id and optional title. If omitted, the backend keeps the source channel and generates a copy title.' },
      { id: 'cm5c', from: 'Codex', content: 'Response shape is { "artifact": Artifact }. The returned artifact has a new prefixed UUID ID, preserves source type, content, template_id, provider, and model, resets status to draft, sets source to duplicate, and starts at version: 1.', isCode: true },
      { id: 'cm5d', from: 'Codex', content: 'The duplicate flow writes an initial ArtifactVersion snapshot and broadcasts artifact.updated so realtime canvas panels can refresh.' },
      { id: 'cm6',  from: 'Codex', to: 'Gemini', content: 'Please wire the frontend Duplicate/Fork action in the Canvas panel, History panel, and any artifact card menus. After a successful duplicate call, select/open the returned artifact immediately instead of refetching blindly. If you need fork lineage UI, tell me the exact fields you want and I will add them in the next backend slice.' },
    ],
  },
]

// ─── Derived helpers ──────────────────────────────────────────────────────────

export const MEMBER_MAP: Record<string, Member> = Object.fromEntries(MEMBERS.map(m => [m.name, m]))

export function getMemberColor(name: string): string {
  return MEMBER_MAP[name]?.bgClass ?? 'bg-slate-500'
}

export function getMemberInitials(name: string): string {
  return MEMBER_MAP[name]?.avatar ?? name.substring(0, 2).toUpperCase()
}

export const TASK_TYPE_META: Record<TaskType, { label: string; color: string }> = {
  api:      { label: 'API',      color: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/20' },
  frontend: { label: 'Frontend', color: 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/20' },
  infra:    { label: 'Infra',    color: 'bg-slate-500/15 text-slate-700 dark:text-slate-400 border-slate-500/20' },
  ux:       { label: 'UX',       color: 'bg-purple-500/15 text-purple-700 dark:text-purple-400 border-purple-500/20' },
}
