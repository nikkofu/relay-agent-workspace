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
| 🔴 Pending | Message Interaction APIs | Codex | 2026-04-18 | Implement backend handlers for reactions, pinning, deletion, and feedback. |

---

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Gemini** | `idle` | Integration pass complete | 100% |
| **Codex** | `idle` | Waiting for new requirements | 100% |
| **Claude Code**| `idle` | - | - |

---

## 💬 Communication Log

### 2026-04-18 - AI Chat UX & Feedback Refined
- **Gemini**: "AI Chat header now shows full `Provider • Model • Mode` details. 'Openrouter' renamed to 'Open Router'."
- **Gemini**: "Fixed duplicate `/ai/config` calls and ensured model auto-selection on first load."
- **Gemini**: "Message actions (Add reaction, Delete, Pin, Save) are now wired in the UI. I've added the API requirements for Codex."
- **Gemini → Codex**: "Please implement the new interaction endpoints listed in 'Backend Specifications'. We need to move from toasts to real persistence."
- **Nikko Fu**: "Version v0.3.0 published. Moving towards full message interaction parity with Slack."

---

## 📖 Operational Guidelines (For Agents)

1. **Read First**: Always start your session by reading `docs/AGENT-COLLAB.md` to get the latest context.
2. **Update After Task**: When you finish a task, change an API spec, or switch a Superpower skill, update this file.
3. **Be Atomic**: Keep entries concise and focused on technical requirements or progress status.
4. **Collaboration Mode**: If you need an API change, tag **Codex**. If you need a UI change, tag **Gemini**.

---

## 🛠 Backend Specifications (Request for Codex)

To support real message interactions and AI quality monitoring, please deliver:

1. **AI Feedback API**:
   - **Endpoint**: `POST /api/v1/ai/feedback`
   - **Payload**: `{ "message_id": string, "is_good": boolean }`

2. **Message Reactions**:
   - **Endpoint**: `POST /api/v1/messages/:id/reactions`
   - **Payload**: `{ "emoji": string }`
   - **Behavior**: Add/Toggle reaction for the current user.

3. **Message Lifecycle**:
   - **Endpoints**:
     - `DELETE /api/v1/messages/:id`
     - `POST /api/v1/messages/:id/pin` (Toggle)
     - `POST /api/v1/messages/:id/later` (Save for later)

4. **Message Status**:
   - **Endpoint**: `POST /api/v1/messages/:id/unread` (Mark from this point as unread)
