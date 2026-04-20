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
| 🟢 Done | Phase 10 Foundation APIs | Codex | 2026-04-18 | Added channel members, workspace invites, and channel metadata APIs. |
| 🟢 Done | Phase 10 Frontend Integration | Gemini | 2026-04-18 | Completed member lists, metadata editing, and channel info UI integration. |
| 🟢 Done | Inbox And Mentions APIs | Codex | 2026-04-18 | Added `GET /api/v1/inbox` and `GET /api/v1/mentions` backend endpoints. |
| 🟢 Done | Phase 11 Notification UI | Gemini | 2026-04-18 | Built Inbox and Mentions tabs with direct message/channel navigation. |
| 🟢 Done | Phase 12 Drafts APIs | Codex | 2026-04-18 | Added draft persistence APIs for channel, DM, and thread composer scopes. |
| 🟢 Done | Phase 12 Drafts Integration | Gemini | 2026-04-18 | Wired autosave/restore to Channel, DM, and Thread composers using the new Drafts API. |
| 🟢 Done | Phase 13 Presence And Typing APIs | Codex | 2026-04-18 | Added persisted presence endpoints and websocket typing broadcasts. |
| 🟢 Done | Phase 13 Presence And Typing Integration | Gemini | 2026-04-18 | Integrated live status indicators and real-time typing feedback across the workspace UI. |
| 🟢 Done | Phase 14 Stars And Pins APIs | Codex | 2026-04-18 | Added starred channel and pinned message discovery APIs. |
| 🟢 Done | Phase 14 Stars And Pins Integration | Gemini | 2026-04-18 | Built channel starring UI and a dedicated Pins tab in the ChannelInfo panel. |
| 🟢 Done | Phase 15 Notification Read State APIs | Codex | 2026-04-18 | Added persistent read state for inbox and mentions items. |
| 🟢 Done | Phase 15 Notification Read State Integration | Gemini | 2026-04-18 | Wired unread indicators, mark-as-read on click, and bulk read actions to Activity, Inbox, and Mentions surfaces. |
| 🟢 Done | Phase 16 AI Conversation Persistence APIs | Codex | 2026-04-18 | Added persisted AI conversations and detail APIs behind the existing execute flow. |
| 🟢 Done | Phase 16 AI History Integration | Gemini | 2026-04-18 | Built AI history browsing UI, session continuation logic, and centralized AI state management. |
| 🟢 Done | Phase 17 AI Summaries APIs | Codex | 2026-04-18 | Added persistent thread and channel summary generation APIs. |
| 🟢 Done | Phase 17 AI Summaries Integration | Gemini | 2026-04-18 | Wired real-time thread and channel summary generation into the UI using persistent backend APIs. |
| 🟢 Done | UI Bug Bash & UX Refinements | Gemini | 2026-04-18 | Fixed critical hydration errors, duplicate keys, scrolling bugs, and completed branding unification. |
| 🟢 Done | Phase 18 Artifact Lifecycle APIs | Codex | 2026-04-19 | Added artifact CRUD, AI canvas generation, realtime artifact updates, stable activity IDs, and channel creation support. |
| 🟢 Done | Phase 18 Artifact Lifecycle Integration | Gemini | 2026-04-19 | Connected CanvasPanel to real artifact APIs, implemented AI canvas generation flow, and enabled real-time sync. |
| 🟢 Done | Phase 19 File Assets APIs | Codex | 2026-04-19 | Added file upload/list/detail/content APIs and hydrated artifact editor user objects. |
| 🟢 Done | Phase 19 File Assets Integration | Gemini | 2026-04-19 | Built file upload UI, channel asset listing, and enriched artifact identity with user metadata. |
| 🟢 Done | Phase 20 Presence Refinements APIs | Codex | 2026-04-19 | Added heartbeat refresh, scoped presence queries, and enriched presence metadata. |
| 🟢 Done | Phase 20 Presence Refinements Integration | Gemini | 2026-04-19 | Integrated 30s heartbeat interval, scoped member presence fetching, and "Last seen" UI metadata. |
| 🟢 Done | Phase 21 Artifact Version History APIs | Codex | 2026-04-20 | Added persisted artifact snapshots plus version list/detail APIs for canvas history. |
| 🟢 Done | Phase 21 Artifact Version History Integration | Gemini | 2026-04-20 | Built the History panel for artifacts with version browsing and one-click restoration. |
| 🟢 Done | Phase 22 Artifact Diff APIs | Codex | 2026-04-20 | Added version-to-version diff API for canvas comparison views. |
| 🟢 Done | Phase 22 Artifact Diff Integration | Gemini | 2026-04-20 | Built a visual comparison UI for artifacts using unified diff payloads and multi-version history selection. |
| 🟢 Done | AI UI Stability & Slash Commands | Gemini | 2026-04-20 | Fixed AI panel scrolling, rich-text command leaks, and implemented dynamic slash command filtering. |
| 🟢 Done | AI & Canvas Flow Stabilization | Gemini | 2026-04-20 | Integrated AI command forwarding, fixed `new-doc` save flow, and aligned diff mapping with the backend. |
| 🟢 Done | Phase 23 Search Suggestions APIs | Codex | 2026-04-20 | Added typed search suggestions plus richer snippet and match-reason search payloads. |
| 🟢 Done | Phase 23 Search Suggestions Integration | Gemini | 2026-04-20 | Built real-time search suggestions UI and integrated rich result metadata (snippets, match reasons). |
| 🟢 Done | Phase 24 Artifact Restore APIs | Codex | 2026-04-20 | Added version restore support plus structured diff spans for richer canvas history workflows. |
| 🟢 Done | Phase 24 Artifact Restore Integration | Gemini | 2026-04-20 | Wired the official restore CTA and implemented richer diff rendering using structured spans and line numbers. |

---

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Gemini** | `idle` | Waiting for next phase handoff (e.g. references or semantic search) | 100% |
| **Codex** | `idle` | Waiting for next phase handoff | 100% |
| **Claude Code**| `idle` | - | - |

---

## 💬 Communication Log

### 2026-04-20 - Artifact Restore & Structured Diff Completion
- **Gemini**: "Phase 24 frontend is complete. The 'Restore this version' button is now fully operational using the official backend endpoint."
- **Gemini**: "Upgraded `ArtifactDiffView` to use structured spans for rendering. Added hoverable line number hints for better precision. Version `v0.5.40` published."
- **Gemini → Codex**: "Structured spans are a great addition. They make the diff rendering logic much cleaner. No immediate backend needs for this wave."
- **Gemini → Nikko Fu**: "Rollbacks are now official. You can revert any document to a previous state with full history preservation and clearer visual diffs."

### 2026-04-20 - Artifact Restore API Handoff
...
 Process Group PGID: 16001
