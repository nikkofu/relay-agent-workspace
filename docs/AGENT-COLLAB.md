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
| 🟢 Done | Dynamic AI Settings UI Wiring | Gemini | 2026-04-18 | Replaced hardcoded AI settings with dynamic config and persistence. |

---

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Gemini** | `idle` | Integration complete | 100% |
| **Codex** | `idle` | Waiting for new requirements | 100% |
| **Claude Code**| `idle` | - | - |

---

## 💬 Communication Log

### 2026-04-16 - Real-time Sync Accomplished
- **Gemini**: "WebSocket sync for `#agent-collab` is fully integrated! The dashboard now updates live from this document."

### 2026-04-17 - Dynamic AI Config Handoff
- **Codex**: "Delivered `GET /api/v1/ai/config` and `PATCH /api/v1/me/settings`."

### 2026-04-18 - Full Dynamic Settings Integration
- **Gemini**: "Dynamic AI settings are now fully wired. The ⚙️ panel now fetches available providers/models from `/ai/config`."
- **Gemini**: "User preferences for AI (provider, model, mode) are now persisted to the backend via `/me/settings` and hydrated on load."
- **Gemini → Nikko Fu**: "Version v0.2.9 published. The frontend is now fully decoupled from static AI configurations."

---

## 📖 Operational Guidelines (For Agents)

...

---

## 🛠 Backend Specifications (Request for Codex)

...
