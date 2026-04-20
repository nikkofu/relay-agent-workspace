# Changelog

All notable changes to Relay Agent Workspace are documented in this file.

## [0.5.47] - 2026-04-20

This release focuses on stability, engineering standards, and codebase cleanup after the Next.js 16 upgrade.

### Fixed

- **Store Syntax Errors**: Resolved critical parsing errors in `UserStore` and `WorkspaceStore` caused by placeholder leaks.
- **Lint Sanitization**: Achieved zero errors and zero warnings across the entire web application.
- **Redundancy Cleanup**: Removed duplicate `HomeDashboard` implementation in `WorkspacePage` and consolidated into a reusable component.
- **Type Safety**: Fixed a missing field in `SearchStore` initialization that caused build failures.
- **Image Optimization**: Migrated remaining native `<img>` tags to `next/image` for better performance and LCP.
- **Dead Code Removal**: Pruned over 10 unused variables, imports, and parameters across stores and hooks.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.45] - 2026-04-20


This release expands the Slack-parity foundation beyond channels and messages into home, people, groups, status, tools, and workflow surfaces.

### Added

- `GET /api/v1/home`
- `GET /api/v1/users/:id`
- `PATCH /api/v1/users/:id/status`
- `GET /api/v1/user-groups`
- `GET /api/v1/user-groups/:id`
- `GET /api/v1/workflows`
- `GET /api/v1/tools`

### Data Model

- `user_groups`
- `user_group_members`
- `workflow_definitions`
- `tool_definitions`
- richer `users` fields for:
  - `title`
  - `department`
  - `timezone`
  - `working_hours`

### Behavior Notes

- `GET /api/v1/home` returns a Slack-style landing payload with:
  - `user`
  - `profile`
  - `activity`
  - `starred_channels`
  - `recent_dms`
  - `drafts`
  - `tools`
  - `workflows`
- `GET /api/v1/users/:id` now returns a hydrated personal profile with local time, working hours, focus areas, top channels, and recent artifacts
- `PATCH /api/v1/users/:id/status` persists `status` and `status_text`, refreshes presence expiry, and emits websocket `presence.updated`
- user groups now expose both list and detail contracts, including hydrated member users
- workflows and tools are now explicit backend registries instead of UI-only assumptions

### Documentation

- added [docs/phases/phase-10-home-profiles-groups-workflows.md](./docs/phases/phase-10-home-profiles-groups-workflows.md)
- added [docs/releases/v0.5.46.md](./docs/releases/v0.5.46.md)
- updated `docs/AGENT-COLLAB.md` with the next Gemini handoff for home, people, user groups, and workflow surfaces

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm build`

## [0.5.44] - 2026-04-20

This release implements Phase 26: Intelligent Search and Artifact Backlinks, deepening the connectivity and discoverability of knowledge within the workspace.

### Added

- **Artifact Backlinks**: New "Referencing Messages" sidebar in `CanvasPanel` to see where an artifact is mentioned in discussions.
- **Intelligent Ranked Search**: Integrated AI-native ranked search results at the top of the search palette.
- **Realtime Read State Sync**: Leveraged websocket `notifications.read` to synchronize notification read state across all open client windows.
- **Search Score Badges**: Ranked search results now display their relevance score and match reason.

### Fixed

- **Navigation**: Intelligent search results correctly navigate to channels, users, or artifacts.
- **Store Reconciliation**: Added `markAsReadLocally` to `ActivityStore` for consistent multi-device read states.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.42] - 2026-04-20


This release adds the first ranked search layer and cross-surface backlink/read-sync primitives.

### Added

- `GET /api/v1/artifacts/:id/references`
- `GET /api/v1/search/intelligent?q=...`
- realtime `notifications.read` event broadcast from `POST /api/v1/notifications/read`

### Search Contract

- intelligent search returns ranked typed items with:
  - `type`
  - `id`
  - `label`
  - `reason`
  - `score`

### Documentation

- added [docs/phases/phase-10-intelligent-search-backlinks-and-notification-sync.md](./docs/phases/phase-10-intelligent-search-backlinks-and-notification-sync.md)
- added [docs/releases/v0.5.43.md](./docs/releases/v0.5.43.md)
- updated `docs/AGENT-COLLAB.md` with Gemini handoff for backlinks, ranked search, and notification realtime sync

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.42] - 2026-04-20

This release implements Phase 25: Knowledge References and Knowledge Search, enabling deep linking and discovery of artifacts and files across the workspace.

### Added

- **Message Attachments**: Messages can now include artifact references and file attachments, rendered inline with rich previews.
- **Composer Integration**: The message composer now correctly tracks uploaded files and attaches them to the outgoing message.
- **Knowledge Search**: Expanded the global search to include dedicated sections for artifacts and files, complete with typeahead suggestions.
- **Smart Suggestions**: Global search suggestions now include matching artifacts and files for faster discovery.

### Fixed

- **Type Safety**: Updated `Message` and `Attachment` types to support rich hydrated artifact and file metadata.
- **UI Consistency**: Standardized the rendering of files and artifacts across the channel introduction, message items, and search results.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.40] - 2026-04-20


This release expands Relay's knowledge graph across messages, artifacts, files, and search.

### Added

- `POST /api/v1/messages` now accepts optional:
  - `artifact_ids`
  - `file_ids`
- `GET /api/v1/messages` and `GET /api/v1/messages/:id/thread` now hydrate linked artifacts and files into `metadata.attachments`
- `GET /api/v1/search` now returns:
  - `results.artifacts`
  - `results.files`
- `GET /api/v1/search/suggestions` now includes typed:
  - `artifact` suggestions
  - `file` suggestions

### Data Model

- `message_artifact_references`
- `message_file_attachments`

### Documentation

- added [docs/phases/phase-10-knowledge-references-and-search-expansion.md](./docs/phases/phase-10-knowledge-references-and-search-expansion.md)
- added [docs/releases/v0.5.41.md](./docs/releases/v0.5.41.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for message references and search expansion

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.40] - 2026-04-20

This release completes Phase 24: Artifact Restore and Enhanced History, providing a robust workflow for document version control and visualization.

### Added

- **Official Version Restore**: Wired the "Restore this version" button to the new `POST /api/v1/artifacts/:id/restore/:version` endpoint.
- **Structured Diff Spans**: Integrated the new `diff.spans` payload for more accurate and richer diff rendering in the comparison view.
- **Line Number Hints**: Added hoverable line number markers (`L10 → L12`) in the diff view when provided by the backend.
- **Robust History Hydration**: Automatically refreshes the version history and active artifact state after a successful rollback.

### Fixed

- **Diff Fallback**: Improved the diff rendering logic to gracefully fallback to manual `unified_diff` parsing if structured spans are missing.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.38] - 2026-04-20

This release implements Phase 24: Artifact Restore, allowing users to revert collaborative canvases to any previous version in their history.

### Added

- **Version Restore**: Integrated `POST /api/v1/artifacts/:id/restore/:version` to enable official version rollbacks.
- **Structured Diff Mapping**: Updated `artifact-store.ts` to handle the new `diff.spans` field for richer future diff rendering.
- **Restore CTA**: Wired the "Restore this version" button in the `CanvasPanel` history sidebar to the backend restore flow.

### Fixed

- **History Hydration**: After a successful restore, the version history list is automatically refreshed to reflect the new state.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.36] - 2026-04-20


This release upgrades the canvas history backend from view-only history into a real rollback-ready workflow with restore support and structured diff spans.

### Added

- `POST /api/v1/artifacts/:id/restore/:version`
- structured `diff.spans` in `GET /api/v1/artifacts/:id/diff/:from/:to`

### Artifact Contract

- restore responses include:
  - `artifact`
  - `restored_from_version`
- diff responses now include:
  - `spans[].kind`
  - `spans[].content`
  - optional `spans[].from_line`
  - optional `spans[].to_line`

### Documentation

- added [docs/phases/phase-10-artifact-restore-and-structured-diff.md](./docs/phases/phase-10-artifact-restore-and-structured-diff.md)
- added [docs/releases/v0.5.37.md](./docs/releases/v0.5.37.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for canvas restore and richer compare rendering

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.36] - 2026-04-20

This release implements Phase 23: Search Suggestions and Rich Results, significantly improving the search experience with real-time typeahead and enhanced result metadata.

### Added

- **Search Suggestions**: Integrated `GET /api/v1/search/suggestions` to provide instant typeahead suggestions in the search dialog.
- **Message Snippets**: Search results for messages now render highlighted snippets, making it easier to identify relevant context.
- **Match Reasons**: Channels and users in search results now display the reason for the match (e.g., "Matched in description").
- **Unified Search Store**: Expanded `search-store.ts` to manage suggestions and handle richer result payloads.
- **Improved Navigation**: Clicking on search results now correctly navigates to channels or opens docked chats for users.

### Fixed

- **Dialog UX**: Optimized the search input and results list for a smoother command palette experience.
- **Lint Stability**: Resolved all unused variable and unescaped quote errors in the search component.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.34] - 2026-04-20


This release upgrades workspace search from a raw result dump to a more UI-friendly discovery API with suggestions and richer result shaping.

### Added

- `GET /api/v1/search/suggestions?q=...`

### Search Contract

- `GET /api/v1/search` now returns richer result items including:
  - `messages[].snippet`
  - `channels[].match_reason`
  - `users[].match_reason`
- `GET /api/v1/search/suggestions` returns typed suggestion items with:
  - `type`
  - `id`
  - `label`
  - `hint`

### Documentation

- added [docs/releases/v0.5.35.md](./docs/releases/v0.5.35.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for search suggestions

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.34] - 2026-04-20

This release stabilizes the AI-Canvas collaboration flow by aligning the frontend with recent backend command and sentinel improvements.

### Added

- **Command Forwarding**: Explicit slash commands (like `/canvas`) are now forwarded to the AI execute endpoint, skipping unnecessary classification logic.
- **Robust Diff Mapping**: Enhanced `artifact-store.ts` to handle the standardized `{ diff: ... }` response envelope for artifact comparisons.

### Fixed

- **Canvas Creation**: Corrected the save flow for new documents (`new-doc`) to ensure they are created as real artifacts upon first save.
- **Data Integrity**: Improved `mapArtifact` with better null-safety and field mapping for `created_by_user` and `updated_by_user`.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.32] - 2026-04-20


This release closes the loop between the latest Gemini UI work and the backend by removing the `new-doc` artifact 404 path and adding explicit AI command forwarding.

### Added

- `POST /api/v1/ai/execute` now accepts optional `command`

### Fixed

- opening a new canvas no longer triggers `GET /api/v1/artifacts/new-doc`
- saving a new canvas now creates a real artifact instead of trying to patch a sentinel id
- artifact diff store mapping now correctly reads the backend `{ diff: ... }` envelope

### Documentation

- added [docs/releases/v0.5.33.md](./docs/releases/v0.5.33.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for command forwarding and canvas sentinel handling

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.32] - 2026-04-20

This release improves the AI Assistant's responsiveness and stability, specifically around slash commands and panel layout.

### Added

- **Slash Command Filtering**: Integrated dynamic filtering in the `AISlashCommand` menu. Commands now filter in real-time as you type after the slash.
- **Smart /canvas Interception**: The UI now intercepts `/canvas` commands locally to provide instant visual feedback (opening the canvas) while the AI processes the prompt.

### Fixed

- **AI Panel Scrolling**: Resolved a recurring issue where the AI Assistant dialogue wouldn't show a scrollbar by stabilizing the flex-height inheritance.
- **HTML Command Leaks**: Fixed a bug where slash commands entered in the rich editor would leak as raw HTML tags in the message stream.
- **Artifact Store Safety**: Added robust null-checks in `mapArtifact` to prevent client-side crashes on malformed API responses.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.30] - 2026-04-20

This release implements Phase 22: Artifact Diff UI, allowing users to compare different versions of collaborative canvases side-by-side.

### Added

- **Artifact Comparison UI**: New `ArtifactDiffView` component to render unified diffs with line-level highlighting.
- **Compare Mode**: Enhanced `CanvasPanel` with a "Compare Versions" toggle in the history sidebar.
- **Multi-Version Selection**: Users can now select any two historical versions to generate a visual diff.
- **Diff Summary**: Real-time statistics showing the number of lines added and removed between versions.
- **Lazy Diff Loading**: Optimized network usage by fetching diff data only when a specific comparison is requested.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.28] - 2026-04-20


This release adds artifact diff APIs so Relay can compare any two stored canvas versions and power a real visual comparison view.

### Added

- `GET /api/v1/artifacts/:id/diff/:from/:to`

### Diff Contract

- diff responses include:
  - `artifact_id`
  - `from_version`
  - `to_version`
  - `from_content`
  - `to_content`
  - `unified_diff`
  - `summary.added_lines`
  - `summary.removed_lines`

### Documentation

- added [docs/phases/phase-10-artifact-diff.md](./docs/phases/phase-10-artifact-diff.md)
- added [docs/releases/v0.5.29.md](./docs/releases/v0.5.29.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for artifact comparison views

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.28] - 2026-04-20

This release implements Phase 21: Artifact Version History, enabling users to browse and restore previous versions of collaborative canvases.

### Added

- **Version History UI**: Integrated a history sidebar in the `CanvasPanel` to browse all previous versions of an artifact.
- **Version Preview**: Users can now view the content, title, and metadata of any historical version without affecting the current live state.
- **One-Click Restore**: Added a "Restore this version" capability to quickly rollback the active canvas to a prior state.
- **History Metadata**: Each version now displays the timestamp and the user who made the change.
- **Expanded Artifact Store**: Updated `artifact-store.ts` to manage version lists and lazy-loaded version details.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.26] - 2026-04-19


This release adds artifact version history so Relay canvases are now auditable and ready for a real history panel in the UI.

### Added

- `GET /api/v1/artifacts/:id/versions`
- `GET /api/v1/artifacts/:id/versions/:version`
- persisted `artifact_versions` snapshots on artifact create, update, and AI canvas generation

### Artifact Contract

- artifact payloads now include:
  - `version`
- version payloads include:
  - `artifact_id`
  - `version`
  - `title`
  - `type`
  - `status`
  - `content`
  - `updated_by`
  - `updated_by_user`
  - `created_at`

### Documentation

- added [docs/phases/phase-10-artifact-version-history.md](./docs/phases/phase-10-artifact-version-history.md)
- added [docs/releases/v0.5.27.md](./docs/releases/v0.5.27.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for artifact version history

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.26] - 2026-04-19

This release implements Phase 20: Presence Refinements, making user status more dynamic and context-aware.

### Added

- **Presence Heartbeat**: Implemented a 30-second client-side heartbeat to maintain active session presence.
- **Enriched Last Seen**: Added "Last seen" timestamps using `formatDistanceToNow` for offline users.
- **Scoped Presence**: Added automatic presence fetching for channel members when switching channels.
- **Status Metadata**: Integrated `status_text` support in avatars and profiles.

### Fixed

- UI Consistency: Unified presence mapping across `presence-store.ts` and `user-store.ts`.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.24] - 2026-04-19


This release upgrades Relay presence with heartbeat refresh, derived last-seen metadata, and scoped presence lists.

### Added

- `POST /api/v1/presence/heartbeat`
- scoped `GET /api/v1/presence?channel_id=...`

### Presence Contract

- user payloads now include:
  - `status_text`
  - `last_seen_at`
- presence freshness is tracked with heartbeat expiry timestamps
- websocket `presence.updated` payloads use the enriched user structure

### Documentation

- added [docs/phases/phase-10-presence-refinements.md](./docs/phases/phase-10-presence-refinements.md)
- added [docs/releases/v0.5.25.md](./docs/releases/v0.5.25.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for presence refinements

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.24] - 2026-04-19

This release implements Phase 19: File Assets and Artifact Identity, enabling real file uploads and richer collaboration context.

### Added

- **Real File Uploads**: Integrated multipart file upload API in the message composer.
- **File Attachments**: Uploaded files are automatically linked in the editor for instant sharing.
- **Channel Files Tab**: New "Files" tab in the `ChannelInfo` panel to browse and download all assets shared in a channel.
- **Unified File Store**: Added `file-store.ts` for managing file uploads and listing.
- **Artifact Identity**: The `CanvasPanel` now displays the user who last edited the artifact.
- **Hydrated Metadata**: The `Artifact` model now includes full `createdByUser` and `updatedByUser` objects.

### Fixed

- **Store Types**: Updated `Artifact` interface to support rich user metadata.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.21] - 2026-04-19


This release adds file asset APIs and upgrades artifact responses so Relay can support real attachments and richer canvas identity context.

### Added

- `POST /api/v1/files/upload`
- `GET /api/v1/files`
- `GET /api/v1/files/:id`
- `GET /api/v1/files/:id/content`

### Artifact Contract

- artifact responses now include:
  - `created_by_user`
  - `updated_by_user`
- websocket `artifact.updated` payloads use the same hydrated artifact structure

### Documentation

- added [docs/phases/phase-10-file-assets.md](./docs/phases/phase-10-file-assets.md)
- added [docs/releases/v0.5.22.md](./docs/releases/v0.5.22.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for files and artifact identity hydration
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped file baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.21] - 2026-04-19

This release implements Phase 18: Artifact Lifecycle integration, bringing real dynamic canvases and AI generation to the workspace.

### Added

- **Dynamic CanvasPanel**: Replaced static demo with a full Artifact-backed surface for code and documents.
- **AI Canvas Generation**: Integrated `POST /api/v1/ai/canvas/generate` into the assistant flow, allowing AI to create new artifacts on demand.
- **Real-time Sync**: Leveraged websocket `artifact.updated` events for seamless live collaboration on canvases.
- **Unified Artifact Store**: Added `artifact-store.ts` for managing artifact CRUD and loading states.
- **Channel Header Artifacts**: Dynamic "Recent Artifacts" bar in the channel view for quick access to collaboration assets.

### Fixed

- **Engineering Excellence**: Enforced strict linting in `pnpm dev` and `pnpm build` scripts.
- **Lint Cleanup**: Resolved all remaining unused variable and import errors across the frontend.
- **Event Handling**: Stabilized websocket event mapping for unified transient state management.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.18] - 2026-04-18


This release adds backend-backed artifact lifecycle APIs so Relay's canvas surface can create, persist, update, and AI-generate collaborative outputs.

### Added

- `GET /api/v1/artifacts`
- `POST /api/v1/artifacts`
- `GET /api/v1/artifacts/:id`
- `PATCH /api/v1/artifacts/:id`
- `POST /api/v1/ai/canvas/generate`

### Realtime And Hardening

- artifact mutations now broadcast `artifact.updated`
- activity feed reaction items now use stable unique IDs, avoiding duplicate-key collisions when the same emoji appears multiple times on one message

### Documentation

- added [docs/phases/phase-10-artifact-lifecycle.md](./docs/phases/phase-10-artifact-lifecycle.md)
- added [docs/releases/v0.5.20.md](./docs/releases/v0.5.20.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for canvas and artifacts
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped artifact baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.18] - 2026-04-18

This release focuses on UX polish, branding consistency, and critical UI bug fixes including hydration errors and list rendering stability.

### Added

- **Sidebar Interactivity**: Activated "Add channels" and "Add teammates" buttons with dedicated creation/invitation dialogs.
- **Unified Branding**: Completed the full rebranding from "Acme Corp" to "Relay" across frontend, backend seed data, and documentation.
- **Starred & Pinned Page**: Created a dedicated view at `/workspace/starred` to aggregate all important collaboration assets.

### Fixed

- **Hydration Errors**: Resolved `ResizablePanel` and `DialogTitle` hydration mismatches by implementing stable mounting guards.
- **List Rendering**: Eliminated duplicate key errors in Activity and Later pages through enhanced ID mapping and deduplication logic.
- **Scroll Stability**: Fixed missing scrollbars in DM docked windows by optimizing nested flex containers and removing redundant scroll areas.
- **Date Safety**: Resolved "Invalid time value" crashes by adding defensive date formatting and reliable store-level fallbacks.
- **Syntax & Compilation**: Fixed a parsing error and duplicate `useState` import in the channel creation dialog.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.17] - 2026-04-18


This release implements Phase 17: AI Summaries integration, providing instant insights for both message threads and channels.

### Added

- **AI Thread Summaries**: Integrated persistent thread summary generation in the Thread Panel.
- **AI Channel Summaries**: Added a dedicated AI summary section in the Channel Info panel for high-level context.
- **Dynamic Summarization**: Real-time summary generation and regeneration supported via the backend AI flow.
- **Store Integration**: Expanded `MessageStore` and `ChannelStore` to handle summary state and automatic hydration.

### Fixed

- **UI Polish**: Replaced static thread summary mocks with real API-driven content and loading states.
- **Hydration**: Summaries are now automatically fetched when opening threads or switching channels.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.16] - 2026-04-18


This release implements Phase 16: AI Conversation Persistence, allowing users to browse and resume previous AI assistant chats.

### Added

- **AI History View**: Integrated a history list in the AI Assistant panel to browse previous conversations.
- **Session Continuation**: AI chats now persist across sessions and can be resumed by passing `conversation_id`.
- **New Chat**: Added a "New Chat" button to quickly reset the assistant context.
- **AI Store**: Introduced `ai-store.ts` for centralized management of AI messages and conversation history.

### Fixed

- **Store Sync**: Unified message handling between `useAIChat` and the persistent conversation list.
- **UI Interaction**: History toggle automatically fetches the latest list when opened.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.15] - 2026-04-18


This release adds persistent AI summaries for threads and channels so Relay can surface reusable context snapshots instead of relying on frontend-only mock summary cards.

### Added

- `GET /api/v1/messages/:id/summary`
- `POST /api/v1/messages/:id/summary`
- `GET /api/v1/channels/:id/summary`
- `POST /api/v1/channels/:id/summary`

### Persistence

- added `ai_summaries`
- each summary is keyed by `scope_type + scope_id`
- stored summary fields include:
  - `provider`
  - `model`
  - `content`
  - `reasoning`
  - `message_count`
  - `last_message_at`

### Documentation

- added [docs/phases/phase-10-ai-summaries.md](./docs/phases/phase-10-ai-summaries.md)
- added [docs/releases/v0.5.15.md](./docs/releases/v0.5.15.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for AI summaries
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped AI summary baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.13] - 2026-04-18


This release adds persistent AI conversation history so Relay's assistant panel can recover prior prompts, responses, and reasoning instead of acting as a purely ephemeral stream.

### Added

- `GET /api/v1/ai/conversations`
- `GET /api/v1/ai/conversations/:id`

### Execute Behavior

- `POST /api/v1/ai/execute` now persists:
  - the user prompt
  - the assistant response
  - reasoning text
  - provider/model metadata
- a `conversation_id` can be passed to continue an existing AI conversation
- if `conversation_id` is omitted, a new one is created and returned in the stream

### Documentation

- added [docs/phases/phase-10-ai-conversation-persistence.md](./docs/phases/phase-10-ai-conversation-persistence.md)
- added [docs/releases/v0.5.13.md](./docs/releases/v0.5.13.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for AI conversation persistence
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped AI history baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.12] - 2026-04-18

This release implements persistent notification read states across the activity feed, inbox, and mentions views.

### Added

- **Unread Indicators**: Integrated blue dot indicators and distinctive styling for unread notification items.
- **Mark as Read**: Clicking an activity item now automatically marks it as read in the backend.
- **Bulk Read Action**: Added a "Mark all as read" button to clear notifications in the current view.
- **Unread Counts**: Live unread count badges on Activity, Inbox, and Mentions tabs.
- **Store Support**: Updated `activity-store.ts` with optimistic read state updates and persistent API synchronization.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.11] - 2026-04-18


This release adds persistent notification read state so Relay inbox and mentions surfaces can track which collaboration signals have already been acknowledged.

### Added

- `POST /api/v1/notifications/read`

### Feed Behavior

- `GET /api/v1/inbox` now includes `is_read`
- `GET /api/v1/mentions` now includes `is_read`

### Documentation

- added [docs/phases/phase-10-notification-read-state.md](./docs/phases/phase-10-notification-read-state.md)
- added [docs/releases/v0.5.11.md](./docs/releases/v0.5.11.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for notification read state
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped notification baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.10] - 2026-04-18

This release implements Phase 14: Stars and Pins integration, providing persistent surfaces for important channels and messages.

### Added

- **Channel Starring**: Users can now star/unstar channels directly from the message area header.
- **Pinned Messages Tab**: Integrated a new "Pins" tab in the `ChannelInfo` panel to display all pinned messages for the current channel.
- **Unpin Action**: Conveniently unpin messages directly from the `ChannelInfo` panel.
- **Store Expansion**: Added `toggleStar` to `ChannelStore` and `fetchPins` to `MessageStore` with optimistic UI updates.

### Fixed

- **Header UI**: Synchronized the header star button state with the backend `isStarred` attribute.
- **Data Integrity**: Automatically refreshes the pinned messages list when a message's pin state is toggled.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.9] - 2026-04-18


This release adds the first starred and pinned discovery APIs so Relay can surface saved channels and pinned references as dedicated collaboration destinations.

### Added

- `GET /api/v1/starred`
- `POST /api/v1/channels/:id/star`
- `GET /api/v1/pins`

### Behavior

- starred channels are returned from persisted `channels.is_starred`
- channel star toggle returns the updated channel plus `is_starred`
- pins return hydrated message items with:
  - `message`
  - `channel`
  - `user`

### Documentation

- added [docs/phases/phase-10-stars-pins.md](./docs/phases/phase-10-stars-pins.md)
- added [docs/releases/v0.5.9.md](./docs/releases/v0.5.9.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for starred and pinned surfaces
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped discovery baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.8] - 2026-04-18

This release implements Phase 13: Presence and Typing indicators, bringing real-time social signals to the workspace.

### Added

- **Live Presence**: Real-time user status (Online, Away, Busy, Offline) across the entire UI.
- **Typing Indicators**: Visual feedback when teammates are typing in Channels, DMs, or Threads.
- **Presence Store**: New `presence-store.ts` for managing transient real-time states.
- **Websocket Expansion**: Integrated `presence.updated` and `typing.updated` events into the unified websocket hook.

### Fixed

- UI consistency: Status dots now appear correctly on all user avatars and navigation elements.
- Typing debounce: Efficiently manages typing broadcast state to minimize network overhead.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.7] - 2026-04-18


This release adds the first realtime presence layer for Relay by shipping presence APIs and typing broadcasts for channels, DMs, and threads.

### Added

- `GET /api/v1/presence`
- `POST /api/v1/presence`
- `POST /api/v1/typing`

### Realtime

- `POST /api/v1/presence` broadcasts `presence.updated`
- `POST /api/v1/typing` broadcasts `typing.updated`

### Documentation

- added [docs/phases/phase-10-presence-typing.md](./docs/phases/phase-10-presence-typing.md)
- added [docs/releases/v0.5.7.md](./docs/releases/v0.5.7.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for presence and typing
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped realtime baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.6] - 2026-04-18

This release implements the Phase 12 Drafts integration, bringing autosave persistence to all composer surfaces.

### Added

- **Drafts Persistence**: Integrated `GET /api/v1/drafts` and `PUT /api/v1/drafts/:scope` for reliable message autosaving.
- **Rich Thread Replies**: Replaced the basic thread textarea with the rich `MessageComposer`, enabling bold, italic, and slash commands in threads.
- **Unified Draft Store**: Added `draft-store.ts` to manage draft state across Channels, DMs, and Threads.
- **Autosave & Restore**: Content is automatically saved as you type and restored when you return to a conversation.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.5] - 2026-04-18


This release starts the next Slack-parity backend wave by adding draft persistence APIs for channel, DM, and thread composers.

### Added

- `GET /api/v1/drafts`
- `PUT /api/v1/drafts/:scope`

### Behavior

- drafts are returned for the current user only
- drafts are ordered by `updated_at desc`
- one draft is stored per `user_id + scope`
- recommended scope keys:
  - `channel:<channelId>`
  - `dm:<dmId>`
  - `thread:<messageId>`

### Persistence

- added `drafts`
- seeded one example draft for local development visibility

### Documentation

- added [docs/phases/phase-10-drafts.md](./docs/phases/phase-10-drafts.md)
- added [docs/releases/v0.5.5.md](./docs/releases/v0.5.5.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for drafts
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped drafts baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.4] - 2026-04-18

This release implements the Phase 11 notification UI wave, introducing a unified Activity, Inbox, and Mentions experience.

### Added

- **Inbox & Mentions UI**: Integrated new notification tabs in the Activity page.
- **Unified Activity Store**: Expanded `activity-store.ts` to support fetching Inbox and Mentions from backend endpoints.
- **Smart Navigation**: Activity items now support clicking to jump directly to the relevant channel or DM conversation.
- **Interactive Tabs**: Seamless switching between "All Activity", "Inbox", and "Mentions" views.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.3] - 2026-04-18


This release unifies Gemini's `v0.5.2` channel-management frontend work with the next Codex backend wave by adding inbox and mentions APIs.

### Added

- `GET /api/v1/inbox`
- `GET /api/v1/mentions`

### Behavior

- inbox returns aggregated collaboration signals for the current user:
  - direct mentions
  - thread replies
  - reactions on your messages
  - DM activity
- mentions returns the direct-mention subset only

### Documentation

- added [docs/phases/phase-10-inbox-mentions.md](./docs/phases/phase-10-inbox-mentions.md)
- added [docs/releases/v0.5.3.md](./docs/releases/v0.5.3.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for inbox and mentions
- updated `docs/phase8-api-expansion.md` shipped baseline to include the new endpoints

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.2] - 2026-04-18

This release completes the Phase 10 frontend integration for channel management, introducing member lists, metadata editing, and a unified channel info panel.

### Added

- `ChannelInfo` panel for managing channel metadata and membership
- support for editing channel `topic` and `purpose`
- member management UI (list, add, and remove members)
- extended `Channel` and `ChannelMember` types for robust metadata handling
- integrated `ChannelInfo` into the main message area header

### Fixed

- channel metadata sync: ensures `topic` and `purpose` are correctly updated across the UI
- member list hydration: automatically fetches and updates members when switching channels

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.1] - 2026-04-18

This release starts Phase 10 with the first Slack-parity backend wave and adds explicit stage documentation for ongoing Codex + Gemini collaboration.

### Added

- `GET /api/v1/channels/:id/members`
- `POST /api/v1/channels/:id/members`
- `DELETE /api/v1/channels/:id/members/:userId`
- `PATCH /api/v1/channels/:id`
- `GET /api/v1/workspaces/:id/invites`
- `POST /api/v1/workspaces/:id/invites`

### Data Model

- added `channel_members`
- added `workspace_invites`
- extended `channels` with:
  - `topic`
  - `purpose`
  - `is_archived`

### Documentation

- added [docs/phases/phase-10-slack-parity-foundation.md](./docs/phases/phase-10-slack-parity-foundation.md)
- added [docs/releases/v0.5.1.md](./docs/releases/v0.5.1.md)
- updated `docs/AGENT-COLLAB.md` with Gemini handoff for members, invites, and metadata editing
- updated `docs/phase8-api-expansion.md` shipped baseline to include the new endpoints

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.0] - 2026-04-18

This release synchronizes Gemini's docked DM overhaul with the backend contract, confirms the DM payload shape in handler tests, and documents the next Slack-parity API wave for Relay.

### Added

- DM conversation responses now formally include `user_ids`:
  - `GET /api/v1/dms`
  - `POST /api/v1/dms`
- handler coverage now verifies:
  - DM list payloads include `user_ids`
  - DM create/open payloads include `user_ids`

### Frontend Integration Sync

- `/workspace/dms` now redirects back into the unified workspace shell
- docked DM chat windows are mounted from the workspace layout
- DM store now maps `user_ids` and defensive `last_message_at` fallbacks
- message-store duplicate guards reduce repeated DM and thread inserts
- IME-safe Enter handling was added to the rich message composer

### Planning And Documentation

- updated `docs/phase8-api-expansion.md` with the next Slack-parity API layer:
  - channel members
  - invites
  - channel topic and purpose
  - stars and pins
  - inbox and mentions
  - drafts
- updated `docs/AGENT-COLLAB.md` with the `v0.5.0` handoff and next-phase objectives
- updated `README.md` to reflect the docked DM experience and the new planning track

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.4.1] - 2026-04-18

This release closes the DM realtime gap and expands the post-Phase-8 backend with Activity, Later, and Search APIs so Gemini can finish the remaining workspace pages against live data.

### Added

- `GET /api/v1/activity`
  - returns recent collaboration signals for the current user
  - includes mentions, thread replies, reactions on your messages, and DM activity
- `GET /api/v1/later`
  - returns the current user's saved messages with channel and sender context
- `GET /api/v1/search?q=...`
  - returns grouped results across:
    - `channels`
    - `users`
    - `messages`
    - `dms`

### Realtime And DM Improvements

- `POST /api/v1/dms/:id/messages` now broadcasts `message.created`
- DM websocket payloads now include:
  - `id`
  - `dm_id`
  - `user_id`
  - `content`
  - `created_at`
- `POST /api/v1/dms` now accepts either:
  - `user_id`
  - `user_ids`
  which keeps the backend compatible with the current frontend store shape

### Documentation Sync

- updated `README.md` current status to the `v0.4.1` release line
- updated `docs/phase8-api-expansion.md` so DM, Activity, Later, and Search are no longer listed as missing
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for page integration

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.3.9] - 2026-04-18

This release starts the first post-Phase-8 backend expansion by adding the minimum DM API surface needed for Gemini to replace the static DMs page.

### Added

- `GET /api/v1/dms`
  - returns recent DM conversations for the current user
  - includes the counterpart user plus last message preview
- `POST /api/v1/dms`
  - creates or reopens a 1:1 DM conversation
  - payload: `{ "user_id": string }`
- `GET /api/v1/dms/:id/messages`
  - returns ordered DM history
- `POST /api/v1/dms/:id/messages`
  - sends a DM message
  - payload:
    - `content`
    - `user_id`

### Data Model

- added:
  - `dm_conversations`
  - `dm_members`
  - `dm_messages`
- seeded one initial DM thread between `user-1` and `user-2`

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.3.7] - 2026-04-18

This release adds a real backend collaboration insight path for user profiles and fixes `#agent-collab` so it renders meaningful content on first load.

### Added

- `GET /api/v1/agent-collab/snapshot`
  - returns:
    - `active_superpowers`
    - `task_board`
- dynamic `ai_insight` generation on existing user payloads:
  - `GET /api/v1/me`
  - `GET /api/v1/users`

### Fixed

- `#agent-collab` no longer depends on catching a startup websocket event
- frontend collab store now correctly maps backend lowercase JSON keys
- the dashboard now renders both:
  - agent cards
  - task board rows

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.3.4] - 2026-04-18

This release closes the audit loop on the original LLM/thread/user API delivery plan and synchronizes repository-facing docs with the shipped backend state.

### Audit Result

- `docs/superpowers/plans/2026-04-17-llm-thread-user-api.md` has been audited against the live codebase
- conclusion: the original plan scope is complete
- the plan file now reflects shipped status instead of stale unchecked tasks

### Documentation Sync

- updated `README.md` to point at the latest release line
- updated `docs/AGENT-COLLAB.md` with the audit result and Gemini handoff note

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm build`

## [0.3.3] - 2026-04-18

This release aligns the backend with Gemini's latest realtime AI integration pass, adds reasoning-aware SSE support, and refreshes the repository docs for GitHub-facing consumption.

### Added

- `POST /api/v1/ai/execute` may now emit `event: reasoning` in addition to:
  - `start`
  - `chunk`
  - `done`
  - `error`

### Realtime Coverage

- reaction mutations now broadcast `reaction.updated`
- pin toggles now broadcast `message.updated`
- deletions now broadcast `message.deleted`

### Documentation Refresh

- updated `README.md` with current product positioning and shipped backend surface
- replaced broken local-only links with GitHub-safe relative links
- rewrote `docs/phase8-api-expansion.md` to reflect the current monorepo and broader backend target
- updated `docs/AGENT-COLLAB.md` with Gemini handoff notes for reasoning and websocket coverage

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm build`

## [0.3.1] - 2026-04-18

This release completes the first persistent message interaction APIs for Gemini's channel message actions UI.

### Added

- `POST /api/v1/ai/feedback`
  - Payload: `{ "message_id": string, "is_good": boolean }`
  - Response: `{ "feedback": { ... } }`
- `POST /api/v1/messages/:id/reactions`
  - Payload: `{ "emoji": string }`
  - Toggle behavior for the current user
  - Response:
    - `message`
    - `added`
- `DELETE /api/v1/messages/:id`
  - Deletes the target message
  - Deletes child replies when the target is a thread parent
  - Response: `{ "deleted": true, "message_id": string }`
- `POST /api/v1/messages/:id/pin`
  - Toggle message pin state
  - Response:
    - `message`
    - `is_pinned`
- `POST /api/v1/messages/:id/later`
  - Toggle save-for-later state for the current user
  - Response:
    - `message_id`
    - `saved`
- `POST /api/v1/messages/:id/unread`
  - Marks the message as an unread checkpoint for the current user
  - Response:
    - `message_id`
    - `unread`

### Data Model And Behavior

- `Message` now includes:
  - `is_pinned`
- New persistence tables:
  - `message_reactions`
  - `saved_messages`
  - `unread_markers`
  - `ai_feedback`
- `GET /api/v1/messages` and `GET /api/v1/messages/:id/thread` now rebuild `metadata.reactions` from persisted reaction rows
- deleting a thread reply now recomputes the parent message:
  - `reply_count`
  - `last_reply_at`

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm build`

## [0.2.8] - 2026-04-17

This release removes the last hardcoded AI settings path for the frontend and adds persistence for user AI preferences plus stronger thread reply metadata.

### Added

- `GET /api/v1/ai/config`
  - Returns enabled providers and configured models
  - Response shape:
    - `default_provider`
    - `providers[]`
- `PATCH /api/v1/me/settings`
  - Persists AI preference fields on the current user:
    - `provider`
    - `model`
    - `mode`

### Thread Integrity Improvements

- reply creation now updates:
  - `reply_count`
  - `last_reply_at`
- `Message` model now stores:
  - `last_reply_at`
- `User` model now stores:
  - `ai_provider`
  - `ai_model`
  - `ai_mode`

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`

## [0.2.5] - 2026-04-17

This release hardens local LLM configuration loading and records the first real upstream validation pass against the configured providers.

### Fixed

- layered LLM config merge now preserves provider defaults while applying local overrides
- local YAML parsing now tolerates tab characters in `llm.local.yaml`

### Verified

- local `gemini` provider configuration successfully streamed real SSE output through `POST /api/v1/ai/execute`
- local `openrouter` configuration loaded correctly, but the selected free model returned upstream `429` rate limiting during validation

### Notes

- with the current local config, Gemini is the verified working provider for frontend integration
- OpenRouter can still be used after switching to a non-rate-limited model or retrying later

## [0.2.4] - 2026-04-17

This release adds the first real LLM gateway architecture plus the backend APIs Gemini requested for user resolution, message threading, and AI streaming.

### Added

- Provider-based LLM gateway in `apps/api/internal/llm`
- Config loading in `apps/api/internal/config`
- Config files:
  - `apps/api/config/llm.base.yaml`
  - `apps/api/config/llm.example.yaml`
  - `apps/api/config/llm.local.yaml`
  - `apps/api/config/llm.secrets.local.yaml`
- Supported provider kinds:
  - `openai`
  - `openai-compatible`
  - `openrouter`
  - `gemini`

### API Surface Added In This Release

- `GET /api/v1/users`
  - Returns all users
  - Supports optional query param `id`
  - Response: `{ "users": [...] }`

- `GET /api/v1/messages/:id/thread`
  - Returns parent message plus replies
  - Response:
    - `parent`
    - `replies`

- `POST /api/v1/ai/execute`
  - Accepts:
    - `prompt`
    - `channel_id`
    - optional `provider`
    - optional `model`
  - Returns `text/event-stream`
  - Stream events currently emitted:
    - `start`
    - `chunk`
    - `done`
    - `error`

### Model And Handler Changes

- `Message` now includes:
  - `thread_id`
  - `reply_count`
- `GET /api/v1/messages` now returns top-level channel messages only
- `POST /api/v1/messages` accepts optional `thread_id`
- Reply creation increments parent `reply_count`

### LLM Notes

- OpenAI and OpenAI-compatible providers use configurable `api_style`
  - `responses`
  - `chat_completions`
- OpenRouter defaults to `responses`
- Gemini uses the official native Gemini streaming protocol over `streamGenerateContent?alt=sse`
- Env overrides are supported through:
  - `LLM_DEFAULT_PROVIDER`
  - `LLM_PROVIDER_<NAME>_API_KEY`
  - `LLM_PROVIDER_<NAME>_BASE_URL`
  - `LLM_PROVIDER_<NAME>_MODEL`
  - `LLM_PROVIDER_<NAME>_API_STYLE`
  - `LLM_PROVIDER_<NAME>_ENABLED`

### Verification Used For This Release

- `pnpm build`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`

## [0.2.2] - 2026-04-16

This release adds the backend sync path for the `#agent-collab` workspace view and aligns the repository version with the latest cross-agent handoff.

### Added

- `docs/AGENT-COLLAB.md` file watcher in `apps/api`
- Markdown table parser for:
  - `Task Board`
  - `Active Superpowers`
- Realtime broadcast for `agent_collab.sync`

### Realtime Event Added

- `agent_collab.sync`
  - Broadcast through `GET /api/v1/realtime`
  - Fixed target channel id: `ch-collab`
  - Payload shape:
    - `active_superpowers`
    - `task_board`

Example payload:

```json
{
  "type": "agent_collab.sync",
  "channel_id": "ch-collab",
  "payload": {
    "active_superpowers": [],
    "task_board": []
  }
}
```

### Backend Notes

- Watch path defaults to `../../docs/AGENT-COLLAB.md` from `apps/api`
- Sync is pushed on service startup and on subsequent file `write/create` events
- Parsing currently targets the two collaboration tables only

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- Automated coverage added for:
  - Markdown table parsing
  - Direct sync broadcast
  - Watcher-triggered broadcast after file write

## [0.2.0] - 2026-04-16

This release turns Relay Agent Workspace into a monorepo with a working backend foundation. It is the first version suitable for frontend-to-backend integration across `apps/web` and `apps/api`.

### Added

- Monorepo structure with `apps/web` and `apps/api`
- Go API service built with Gin
- SQLite persistence using GORM auto-migrations
- Seed data for:
  - `Organization`
  - `Team`
  - `User`
  - `Agent`
  - `Workspace`
  - `Channel`
  - `Message`
- In-memory realtime WebSocket hub
- REST handlers for collaboration and organization workflows

### API Surface In This Release

Base URL:
- Web: `http://localhost:3000`
- API: `http://localhost:8080`

Health:
- `GET /ping`
  - Returns `{"message":"pong"}`

Current user:
- `GET /api/v1/me`
  - Returns the seeded current user
  - Response shape:
    - `user.id`
    - `user.org_id`
    - `user.name`
    - `user.email`
    - `user.avatar`
    - `user.status`

Organizations:
- `GET /api/v1/orgs`
  - Lists organizations visible to the current seeded user
  - Response: `{ "organizations": [...] }`

Teams:
- `GET /api/v1/orgs/:id/teams`
  - Lists teams under an organization
  - Example: `/api/v1/orgs/org_1/teams`
  - Response: `{ "teams": [...] }`

Agents:
- `POST /api/v1/orgs/:id/agents`
  - Creates an agent under an organization
  - Required JSON body:
    - `name`
    - `type`
    - `owner_id`
  - Response: `{ "agent": { ... } }`

Workspaces:
- `GET /api/v1/workspaces`
  - Lists workspaces from SQLite
  - Response: `{ "workspaces": [...] }`

Channels:
- `GET /api/v1/channels`
  - Supports query param:
    - `workspace_id`
  - Example: `/api/v1/channels?workspace_id=ws_1`
  - Response: `{ "channels": [...] }`

Messages:
- `GET /api/v1/messages`
  - Required query param:
    - `channel_id`
  - Example: `/api/v1/messages?channel_id=ch_1`
  - Returns channel messages ordered by `created_at asc`
  - Response: `{ "messages": [...] }`

- `POST /api/v1/messages`
  - Creates a new message and persists it to SQLite
  - Required JSON body:
    - `channel_id`
    - `content`
    - `user_id`
  - Response: `{ "message": { ... } }`
  - Side effect:
    - broadcasts a realtime `message.created` event to websocket clients

Realtime:
- `GET /api/v1/realtime`
  - WebSocket upgrade endpoint
  - Current shipped events:
    - `realtime.connected`
    - `message.created`
  - `message.created` event envelope includes:
    - `id`
    - `type`
    - `workspace_id`
    - `channel_id`
    - `entity_id`
    - `ts`
    - `payload`

### Notes For Frontend Integration

- Persistence uses `apps/api/db/relay.db`
- Database auto-migration runs on API startup
- Seed data is idempotent and backfills newly added records
- Realtime is currently in-memory and single-instance only
- Message pagination, presence, typing, AI execute, and SSE are not part of `0.2.0`

### Verification Used For This Release

- `pnpm build`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- Manual local verification of:
  - `GET /ping`
  - `GET /api/v1/orgs`
  - `GET /api/v1/orgs/org_1/teams`
  - `POST /api/v1/orgs/org_1/agents`
  - `GET /api/v1/workspaces`
  - `GET /api/v1/channels?workspace_id=ws_1`
  - `GET /api/v1/messages?channel_id=ch_1`
  - `POST /api/v1/messages`
  - `GET /api/v1/realtime` websocket upgrade and event receipt

### Internal Commits Included

- `e8e3bba` refactor: migrate to monorepo structure with Go backend boilerplate (Phase 8.1)
- `f628d06` feat: implement phase 8 core collaboration api
- `32202a8` chore: move api sqlite database into db directory
- `c752f9d` feat: add organization team and agent api endpoints
- `a1e8699` feat: add realtime websocket hub
