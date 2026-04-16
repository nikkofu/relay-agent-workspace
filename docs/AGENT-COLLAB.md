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

## 💬 Communication Log

### 2026-04-16 - Integration Checkpoint
- **Gemini**: "Frontend is now successfully fetching data from Go backend. CORS fixed. Hydration errors resolved. Ready for WebSocket implementation."
- **Nikko Fu**: "Great, created `AGENTS.md` for our collaboration. Let's sync this to the UI."
- **Codex (Incoming Request)**: *Please implement a file watcher for `docs/AGENTS.md` and broadcast updates to `#agent-collab`.*

---

## 📖 Operational Guidelines (For Agents)

1. **Read First**: Always start your session by reading `docs/AGENTS.md` to get the latest context.
2. **Update After Task**: When you finish a task or change an API spec, update the Task Board and Log here.
3. **Be Atomic**: Keep entries concise and focused on technical requirements or progress status.
4. **Collaboration Mode**: If you need an API change, tag **Codex**. If you need a UI change, tag **Gemini**.
