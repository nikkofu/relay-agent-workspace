# Relay Agent Workspace: Team Collaboration Hub

This document is the primary communication channel between **Nikko Fu**, **Gemini**, and **Codex**. 

---

## 👥 Member Profiles

| Name | Role | Specialty | Primary Tools |
| :--- | :--- | :--- | :--- |
| **Nikko Fu** | Human Owner | Product Strategy, Design, Final Review | Brainstorming, PR Review |
| **Gemini** | Web/UI Agent | Frontend Architecture, Tailwind, Next.js, UX | `apps/web`, `replace`, `write_file` |
| **Codex** | API/Backend Agent | Go, Gin, GORM, SQLite, WebSockets, AI Orchestration | `apps/api`, `internal/` |

---

## 📋 Task Board

| Status | Task | Assigned To | Deadline | Description |
| :--- | :--- | :--- | :--- | :--- |
| 🟢 Done | Monorepo Migration | Gemini/Codex | 2026-04-16 | Moved all frontend to `apps/web`, created `apps/api`. |
| 🟢 Done | Core API v0.2.0 | Codex | 2026-04-16 | Auth, Workspace, Channel, and Message REST APIs. |
| 🟢 Done | Initial Frontend Integration | Gemini | 2026-04-16 | Connected stores to backend, fixed Hydration/CORS. |
| 🟢 Done | #agent-collab UI Scaffolding | Gemini | 2026-04-16 | Created Dashboard, State Cards, and WS Client. |
| 🟢 Done | Real-time WebSocket Integration | Gemini | 2026-04-16 | Integrated `/api/v1/realtime` for live messaging and sync. |
| 🟢 Done | Agent-Collab Sync Service | Codex | 2026-04-16 | File watcher, Markdown table parser, and `agent_collab.sync` WebSocket broadcast. |
| 🟢 Done | Multi-User Profile API | Codex | 2026-04-17 | Implemented `GET /api/v1/users` for sender resolution. |
| 🟢 Done | Message Threads API | Codex | 2026-04-17 | Added `thread_id` and `/messages/:id/thread` support. |
| 🟢 Done | AI SSE Execution Layer | Codex | 2026-04-17 | Implemented `POST /api/v1/ai/execute` with SSE streaming. |
| 🟢 Done | Frontend Integration Pass | Gemini | 2026-04-17 | Finalized User, Thread, and AI SSE integration. |
| 🟢 Done | Dynamic AI Config API | Codex | 2026-04-17 | Implemented `GET /api/v1/ai/config` for enabled provider/model discovery. |
| 🟢 Done | User AI Settings Persistence | Codex | 2026-04-17 | Implemented `PATCH /api/v1/me/settings` for preferences. |
| 🟢 Done | AI Chat UI Refinements | Gemini | 2026-04-18 | Improved ⚙️ settings UI, fixed SSE parsing, added Copy/Regenerate/Feedback. |
| 🟡 In Progress | Message Actions Integration | Gemini | 2026-04-18 | Bind Add reaction, Delete, Pin, Save for later to UI and store. |
| 🟢 Done | Message Interaction APIs | Codex | 2026-04-18 | Implemented persistence-backed reactions, pinning, deletion, unread, later, and AI feedback. |

---

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Gemini** | `message-actions integration` | Wiring persisted message actions to UI/store | 70% |
| **Codex** | `verification-before-completion` | v0.3.1 handoff and release | 100% |
| **Claude Code**| `idle` | - | - |

---

## 💬 Communication Log

### 2026-04-18 - AI Chat UX & Feedback Refined
- **Gemini**: "AI Chat header now shows full `Provider • Model • Mode` details. 'Openrouter' renamed to 'Open Router'."
- **Gemini**: "Fixed duplicate `/ai/config` calls and ensured model auto-selection on first load."
- **Gemini**: "Message actions (Add reaction, Delete, Pin, Save) are now wired in the UI. I've added the API requirements for Codex."
- **Gemini → Codex**: "Please implement the new interaction endpoints listed in 'Backend Specifications'. We need to move from toasts to real persistence."
- **Nikko Fu**: "Version v0.3.0 published. Moving towards full message interaction parity with Slack."

### 2026-04-18 - Message Interaction Backend Delivered
- **Codex**: "Released `v0.3.1` with persistence-backed message interactions."
- **Codex → Gemini**: "The following endpoints are ready for direct UI wiring: `POST /api/v1/messages/:id/reactions`, `DELETE /api/v1/messages/:id`, `POST /api/v1/messages/:id/pin`, `POST /api/v1/messages/:id/later`, `POST /api/v1/messages/:id/unread`, `POST /api/v1/ai/feedback`."
- **Codex → Gemini**: "Reaction responses return `{ message, added }` and `message.metadata.reactions` is now rebuilt from the database, so existing reaction parsing logic can stay."
- **Codex → Gemini**: "Pin responses return `{ message, is_pinned }`; save-for-later returns `{ message_id, saved }`; unread returns `{ message_id, unread }`; delete returns `{ deleted, message_id }`."
- **Codex → Gemini**: "Thread deletes now recompute the parent `reply_count` and `last_reply_at`. No extra frontend workaround is needed."
- **Codex → Gemini**: "Please finish wiring the message action store methods to these endpoints and report any payload/state mismatch with exact response samples."

---

## 📖 Operational Guidelines (For Agents)

1. **Read First**: Always start your session by reading `docs/AGENT-COLLAB.md` to get the latest context.
2. **Update After Task**: When you finish a task, change an API spec, or switch a Superpower skill, update this file.
3. **Be Atomic**: Keep entries concise and focused on technical requirements or progress status.
4. **Collaboration Mode**: If you need an API change, tag **Codex**. If you need a UI change, tag **Gemini**.

---

## 🛠 Backend Specifications

Latest delivered interaction endpoints for frontend wiring:

1. **AI Feedback API**
   - `POST /api/v1/ai/feedback`
   - Payload: `{ "message_id": string, "is_good": boolean }`
   - Response: `{ "feedback": { "message_id": string, "user_id": string, "is_good": boolean, ... } }`

2. **Message Reactions**
   - `POST /api/v1/messages/:id/reactions`
   - Payload: `{ "emoji": string }`
   - Behavior: toggle reaction for the current user
   - Response: `{ "message": Message, "added": boolean }`

3. **Message Lifecycle**
   - `DELETE /api/v1/messages/:id`
   - Response: `{ "deleted": true, "message_id": string }`
   - `POST /api/v1/messages/:id/pin`
   - Response: `{ "message": Message, "is_pinned": boolean }`
   - `POST /api/v1/messages/:id/later`
   - Response: `{ "message_id": string, "saved": boolean }`

4. **Message Status**
   - `POST /api/v1/messages/:id/unread`
   - Response: `{ "message_id": string, "unread": true }`

5. **Frontend Notes For Gemini**
   - `message.metadata.reactions` remains the source of truth for reaction rendering
   - `message.is_pinned` is now persisted on the API model
   - deleting a reply updates the parent thread metadata automatically
