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
| 🟡 In Progress | Real-time WebSocket Integration | Gemini | 2026-04-17 | Integrate `/api/v1/realtime` for live messaging. |
| 🟢 Done | Agent-Collab Sync Service | Codex | 2026-04-16 | File watcher, Markdown table parser, and `agent_collab.sync` WebSocket broadcast implemented in `apps/api`. |
| 🔴 Pending | Agent Shadow Dept Logic | Codex | TBD | Implement SOP engine and Agent handover logic. |

---

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Gemini** | `executing-plans` | v0.2.1 Release & Handover | 100% |
| **Codex** | `verification-before-completion` | Agent-Collab Sync Service handoff | 100% |
| **Claude Code**| `idle` | - | - |

---

## 💬 Communication Log

### 2026-04-16 - Integration Checkpoint
- **Gemini**: "Frontend is now successfully fetching data from Go backend. CORS fixed. Hydration errors resolved. Ready for WebSocket implementation."
- **Nikko Fu**: "Great, created `AGENT-COLLAB.md` for our collaboration. Let's sync this to the UI."
- **Gemini**: "Renamed file to `AGENT-COLLAB.md`. #agent-collab channel is now visible and has a custom hybrid dashboard."
- **Gemini**: "Version v0.2.1 published. I have updated the UI to be ready for real-time sync. Over to you, **Codex**."
- **Codex**: "Backend sync target is implemented. `apps/api` now watches `docs/AGENT-COLLAB.md`, parses the `Task Board` and `Active Superpowers` Markdown tables into JSON, and broadcasts `agent_collab.sync` on the realtime hub with `channel_id: ch-collab`."
- **Codex → Gemini**: "You can now subscribe to `type === agent_collab.sync` on `/api/v1/realtime`. Payload shape is `{ active_superpowers, task_board }`, both derived directly from this file. Please bind it to the `#agent-collab` dashboard state."

---

## 📖 Operational Guidelines (For Agents)

1. **Read First**: Always start your session by reading `docs/AGENT-COLLAB.md` to get the latest context.
2. **Update After Task**: When you finish a task, change an API spec, or switch a Superpower skill, update this file.
3. **Be Atomic**: Keep entries concise and focused on technical requirements or progress status.
4. **Collaboration Mode**: If you need an API change, tag **Codex**. If you need a UI change, tag **Gemini**.

---

## 🛠 Backend Specifications (Request for Codex)

To support the real-time #agent-collab view, please implement the following and deliver as a working update to `apps/api`:

1. **AGENT-COLLAB.md Watcher**: A service to watch this file for changes using `fsnotify` or similar.
2. **Markdown Parser**: Extract structured data from the `⚡️ Active Superpowers` and `📋 Task Board` sections of this file.
3. **WS Event Broadcaster**: When this file changes, broadcast an `agent_collab.sync` event via WebSocket.
   - **Target Channel**: `#agent-collab` (ID: `ch-collab`).
   - **Payload Shape**: 
     ```json
     {
       "type": "agent_collab.sync",
       "payload": {
         "active_superpowers": [...],
         "task_board": [...]
       }
     }
     ```
4. **Member State API**: (Optional) Persistence for Agent thinking/active states in SQLite.
