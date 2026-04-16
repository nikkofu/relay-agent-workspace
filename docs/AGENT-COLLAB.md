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
| 🟡 In Progress | Real-time WebSocket Integration | Gemini | 2026-04-17 | Integrate `/api/v1/realtime` for live messaging. |
| 🔴 Pending | Agent Shadow Dept Logic | Codex | TBD | Implement SOP engine and Agent handover logic. |

---

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Gemini** | `brainstorming` | #agent-collab UI Design | 90% |
| **Codex** | `awaiting` | - | - |
| **Claude Code**| `idle` | - | - |

---

## 💬 Communication Log

### 2026-04-16 - Integration Checkpoint
- **Gemini**: "Frontend is now successfully fetching data from Go backend. CORS fixed. Hydration errors resolved. Ready for WebSocket implementation."
- **Nikko Fu**: "Great, created `AGENT-COLLAB.md` for our collaboration. Let's sync this to the UI."
- **Gemini**: "Renamed file to `AGENT-COLLAB.md`. Designing a hybrid mode for the #agent-collab channel to show Skill status and Task board."
- **Codex (Incoming Request)**: *Please implement a file watcher for `docs/AGENT-COLLAB.md` and broadcast updates to `#agent-collab`.*

---

## 📖 Operational Guidelines (For Agents)

1. **Read First**: Always start your session by reading `docs/AGENT-COLLAB.md` to get the latest context.
2. **Update After Task**: When you finish a task, change an API spec, or switch a Superpower skill, update this file.
3. **Be Atomic**: Keep entries concise and focused on technical requirements or progress status.
4. **Collaboration Mode**: If you need an API change, tag **Codex**. If you need a UI change, tag **Gemini**.

---

## 🛠 Backend Specifications (Request for Codex)

To support the real-time #agent-collab view, please implement:

1. **AGENT-COLLAB.md Watcher**: A service to watch this file for changes.
2. **Markdown Parser**: Extract structured data from the `⚡️ Active Superpowers` and `📋 Task Board` sections.
3. **WS Event Broadcaster**: When this file changes, broadcast an `agent_collab.sync` event via WebSocket to the `#agent-collab` channel (id: `ch-collab` or similar).
   - Payload should include: `active_superpowers`, `task_board`.
4. **Member State API**: (Optional) Persistence for Agent thinking/active states in SQLite.
