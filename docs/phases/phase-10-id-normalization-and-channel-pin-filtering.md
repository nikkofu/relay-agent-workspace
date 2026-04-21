# Phase 10 ID Normalization And Channel Pin Filtering

## Goal

Normalize runtime-generated string primary keys to UUID-style identifiers and fix channel-scoped pin retrieval so Slack-like channel sidebars show only relevant pinned messages.

## Delivered

- prefixed UUID-style IDs for newly created:
  - channels
  - user groups
  - workflow runs
  - lists
  - tool runs
  - files
  - artifacts
  - AI conversations
  - AI conversation messages
- `GET /api/v1/pins?channel_id=...` now correctly filters pinned messages by channel

## Backend Notes

- existing seed/demo IDs remain unchanged; this phase normalizes newly created runtime records
- frontends should treat all IDs as opaque strings and avoid inferring time or type details beyond the prefix contract
- file upload storage remains compatible because file names are still derived from the generated file ID

## Gemini Handoff

1. verify the ChannelInfo Pins tab against at least two different channels
2. verify no frontend logic assumes newly created channel IDs are timestamp-shaped
3. keep channel/message/list/tool/file IDs opaque in client code

## Verification

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm --filter relay-agent-workspace lint`
