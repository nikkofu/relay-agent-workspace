# Phase 10: Knowledge References And Search Expansion

## Goal

Connect Relay messages, canvases, files, and search into one usable knowledge surface.

## Scope

- allow messages to reference artifacts
- allow messages to attach uploaded files through stable relations
- hydrate those references back into `metadata.attachments`
- extend search and search suggestions to include artifacts and files

## API Deliverables

- `POST /api/v1/messages`
  - add optional `artifact_ids`
  - add optional `file_ids`
- `GET /api/v1/messages`
  - hydrated `metadata.attachments`
- `GET /api/v1/messages/:id/thread`
  - hydrated `metadata.attachments`
- `GET /api/v1/search`
  - add `results.artifacts`
  - add `results.files`
- `GET /api/v1/search/suggestions`
  - add typed `artifact` and `file` suggestions

## Data Model

- `message_artifact_references`
- `message_file_attachments`

## Collaboration Split

- Codex:
  - schema and handler work
  - search contract expansion
  - docs and release
- Gemini:
  - composer support for sending `artifact_ids` and `file_ids`
  - message rendering for hydrated references
  - search UI support for artifact/file result groups and suggestions

## Acceptance

- a message can persist artifact and file references
- references survive fetch and thread reloads
- search surfaces can discover artifacts and files without special-case endpoints
