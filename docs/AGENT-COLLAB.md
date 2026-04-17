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
| 🔴 Pending | Multi-User Profile API | Codex | 2026-04-17 | Implement `GET /api/v1/users` to support sender avatars/names. |
| 🔴 Pending | Message Threads API | Codex | 2026-04-17 | Implement `thread_id` and `/messages/:id/thread`. |
| 🔴 Pending | AI SSE Execution Layer | Codex | 2026-04-18 | Implement `POST /api/v1/ai/execute` with SSE streaming. |

---

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Gemini** | `idle` | Waiting for API updates | 100% |
| **Codex** | `awaiting` | - | - |
| **Claude Code**| `idle` | - | - |

---

## 💬 Communication Log

### 2026-04-16 - Real-time Sync Accomplished
- **Gemini**: "WebSocket sync for `#agent-collab` is fully integrated! The dashboard now updates live from this document."
- **Gemini → Codex**: "Excellent work on the file watcher. I've posted new requirements for the next phase in the 'Backend Specifications' section below. We need user profiles and thread support to clean up the UI."
- **Nikko Fu**: "Version v0.2.3 published. Gemini is idling until Codex delivers the next batch of APIs."

---

## 📖 Operational Guidelines (For Agents)

1. **Read First**: Always start your session by reading `docs/AGENT-COLLAB.md` to get the latest context.
2. **Update After Task**: When you finish a task, change an API spec, or switch a Superpower skill, update this file.
3. **Be Atomic**: Keep entries concise and focused on technical requirements or progress status.
4. **Collaboration Mode**: If you need an API change, tag **Codex**. If you need a UI change, tag **Gemini**.

---

## 🛠 Backend Specifications (Request for Codex)

To support the next level of frontend integration, please implement the following in `apps/api`:

1. **GET /api/v1/users**: 
   - **Goal**: Fetch all users or a specific user by ID.
   - **Purpose**: Frontend needs this to replace "Unknown User" with real names and avatars in the message list.
   - **Response**: `{ "users": [...] }`

2. **Message Threads**:
   - **Model Update**: Add `ThreadID` (string) and `ReplyCount` (int) to the `Message` domain model.
   - **New Endpoint**: `GET /api/v1/messages/:id/thread` to fetch all replies for a parent message.

3. **AI SSE Execution Scaffolding**:
   - **Endpoint**: `POST /api/v1/ai/execute`
   - **Payload**: `{ "prompt": string, "channel_id": string }`
   - **Response**: Initiate a **Server-Sent Events (SSE)** stream.
   - **Initial Implementation**: Can be a mock stream that sends "Thinking..." followed by a generic response.
