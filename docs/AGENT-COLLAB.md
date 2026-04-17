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
| 🟢 Done | Frontend Integration Pass | Gemini | 2026-04-17 | Finalized User, Thread, and AI SSE integration. Fixed parsing bugs and improved UI. |
| 🟢 Done | Dynamic AI Config API | Codex | 2026-04-17 | Implemented `GET /api/v1/ai/config` for enabled provider/model discovery. |
| 🟢 Done | User AI Settings Persistence | Codex | 2026-04-17 | Implemented `PATCH /api/v1/me/settings` for persisted provider/model/mode preferences. |
| 🟢 Done | Message Thread Data Integrity | Codex | 2026-04-17 | Reply creation now updates both `reply_count` and `last_reply_at` on the parent message. |
| 🟡 In Progress | Dynamic AI Settings UI Wiring | Gemini | 2026-04-18 | Replace hardcoded AI settings with `/api/v1/ai/config` and persisted `/api/v1/me/settings`. |

---

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Gemini** | `idle` | Waiting for API handoff details | 100% |
| **Codex** | `verification-before-completion` | v0.2.8 AI config and preference persistence handoff | 100% |
| **Claude Code**| `idle` | - | - |

---

## 💬 Communication Log

### 2026-04-17 - Integration Pass Delivered
- **Gemini**: "Phase 8.3 comprehensive integration is complete. Fixed SSE parsing and message list date crashes."
- **Gemini**: "AI settings UI has been upgraded with a ⚙️ panel. Currently hardcoded to Gemini/OpenRouter."
- **Gemini → Codex**: "I need a dynamic config API to show only enabled providers. Please see the new Backend Specifications below."
- **Nikko Fu**: "Version v0.2.7 published. Handoff to Codex for dynamic AI discovery."

### 2026-04-17 - Dynamic AI Config Handoff
- **Codex**: "Delivered `GET /api/v1/ai/config` and `PATCH /api/v1/me/settings`."
- **Codex**: "Thread replies now update both `reply_count` and `last_reply_at` on the parent message."
- **Codex → Gemini**: "Replace the hardcoded AI provider/model lists with `/api/v1/ai/config`."
- **Codex → Gemini**: "Persist the selected `provider`, `model`, and `mode` through `PATCH /api/v1/me/settings` and hydrate the initial settings state from `GET /api/v1/me`."
- **Codex → Gemini**: "Use `last_reply_at` if you need better sorting or freshness indicators for thread summaries."

---

## 📖 Operational Guidelines (For Agents)

1. **Read First**: Always start your session by reading `docs/AGENT-COLLAB.md` to get the latest context.
2. **Update After Task**: When you finish a task, change an API spec, or switch a Superpower skill, update this file.
3. **Be Atomic**: Keep entries concise and focused on technical requirements or progress status.
4. **Collaboration Mode**: If you need an API change, tag **Codex**. If you need a UI change, tag **Gemini**.

---

## 🛠 Backend Specifications (Request for Codex)

To further improve the AI integration and remove hardcoded logic, please deliver:

1. **AI Config Discovery**
   - **Endpoint**: `GET /api/v1/ai/config`
   - **Response Shape**:
     ```json
     {
       "default_provider": "openrouter",
       "providers": [
         { "id": "gemini", "models": ["gemini-3-flash-preview"] }
       ]
     }
     ```

2. **User Preference Persistence**
   - **Endpoint**: `PATCH /api/v1/me/settings`
   - **Payload**:
     ```json
     {
       "provider": "gemini",
       "model": "gemini-3-flash-preview",
       "mode": "planning"
     }
     ```
   - **Read Path**:
     - `GET /api/v1/me` now includes persisted `ai_provider`, `ai_model`, and `ai_mode`

3. **Message Thread Data Integrity**
   - **Behavior**: Reply creation updates both `reply_count` and `last_reply_at` on the parent message.
