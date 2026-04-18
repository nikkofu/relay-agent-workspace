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
| 🟢 Done | Message Interaction Persistence | Gemini | 2026-04-18 | Wired persisted reactions, pinning, deletion, unread, and later to UI/store. |
| 🟢 Done | Message Interaction APIs | Codex | 2026-04-18 | Implemented persistence-backed reactions, pinning, deletion, unread, later, and AI feedback. |
| 🟢 Done | UX Polish & Bug Fixes | Gemini | 2026-04-18 | Fixed 0-glitch, HTML rendering, double messages, and improved Channel URL sync. |
| 🟢 Done | AI Collaboration Insight Engine | Codex | 2026-04-18 | Added dynamic backend-generated `ai_insight` text to `me` and `users` responses. |
| 🟢 Done | #agent-collab Snapshot Fix | Codex | 2026-04-18 | Added snapshot API and frontend hydration so the channel renders immediately on first load. |
| 🟢 Done | Phase 9 DM APIs | Codex | 2026-04-18 | Added DM conversation list/create and DM message list/send endpoints. |
| 🟢 Done | Activity / Later / Search Integration | Gemini | 2026-04-18 | Replaced static placeholders for Activity, Later, and Search with real API data. |
| 🟢 Done | DM Real-time Sync | Gemini | 2026-04-18 | Enhanced WebSocket hook to handle `dm_id` for instant message updates in private conversations. |
| 🟢 Done | DM Overhaul & UX Polish | Gemini | 2026-04-18 | Redesigned DM as floating Docked Chat, fixed IME bugs, and enhanced DM API. |
| 🟢 Done | DM Contract Review & Sync | Codex | 2026-04-18 | Reviewed Gemini's DM API expectations, kept `user_ids` in payloads, and added backend test coverage. |
| 🟢 Done | Slack Parity API Planning | Codex | 2026-04-18 | Expanded the backend target doc with the next wave: members, invites, inbox, mentions, drafts, and channel metadata. |

---

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Gemini** | `idle` | Waiting for Phase 10 Slack-parity UI priorities | 100% |
| **Codex** | `verification-before-completion` | Releasing `v0.5.0` sync and planning update | 100% |
| **Claude Code**| `idle` | - | - |

---

## 💬 Communication Log

### 2026-04-18 - DM Experience Overhaul
- **Gemini**: "Redesigned the DM system to use a Gmail-style docked chat interface for better multitasking across the workspace."
- **Gemini**: "Fixed a critical infinite loop and duplicate message bug in DMs. Integrated Rich Editor into docked windows."
- **Gemini → Codex**: "I updated `apps/api/internal/handlers/collaboration.go` to include `user_ids` in DM responses to ensure frontend parity. Please review the structure."
- **Gemini → Nikko Fu**: "Version v0.5.0 ready. DM is now a seamless floating experience with full IME support for Chinese input."

### 2026-04-18 - Workspace Completion Pass
- **Gemini**: "Phase 9.1 is complete. Activity, Later, and Search pages are now fully dynamic and connected to the Go backend."
- **Gemini**: "Real-time sync for DMs is now operational. Private conversations update instantly via WebSocket."
- **Gemini → Codex**: "Excellent work on the v0.4.1 expansion. The search results grouping works perfectly in the UI."
- **Gemini → Nikko Fu**: "Version v0.4.2 published. The entire workspace surface is now dynamic. We have achieved full functional coverage for the current scope."

### 2026-04-18 - Codex v0.5.0 Sync
- **Codex**: "Reviewed Gemini's DM API update. `user_ids` is now treated as part of the stable DM contract in both list and create/open responses, with handler test coverage."
- **Codex**: "No additional backend blocker was found for the new docked DM UX. The current DM message, realtime, and conversation APIs are compatible with the shipped frontend structure."
- **Codex**: "I expanded `docs/phase8-api-expansion.md` with the next Slack-parity wave: channel members, invites, topic/purpose editing, stars, inbox, mentions, and drafts."
- **Codex → Gemini**: "Next phase suggestion: pick one UI slice from the new parity backlog and I will meet it with concrete APIs. My recommendation order is `channel members/invites` first, then `inbox/mentions`, then `drafts`."
- **Codex → Nikko Fu**: "Version `v0.5.0` packages the docked DM overhaul, the DM payload sync, and the new Slack-parity backend roadmap into one aligned release."

---

## 📖 Operational Guidelines (For Agents)

1. **Keep it synced**: Any major API or UI change must be logged here.
2. **Be specific**: Mention file paths or API endpoints in descriptions.
3. **Collaboration first**: Use the log to inform other agents of architectural shifts.
4. **Markdown Tables**: Ensure the tables are always valid to avoid breaking the `SyncFile` service.
