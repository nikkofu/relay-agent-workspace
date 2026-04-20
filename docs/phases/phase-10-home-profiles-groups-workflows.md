# Phase 10: Home, Profiles, Groups, and Workflow Registry

## Goal

Push Relay's Slack-style foundation beyond channels and messages by giving the frontend stable backend contracts for:

- workspace home
- personal profiles and status updates
- user groups
- workflow discovery
- tool discovery

## Scope

This phase delivers:

- `GET /api/v1/home`
- `GET /api/v1/users/:id`
- `PATCH /api/v1/users/:id/status`
- `GET /api/v1/user-groups`
- `GET /api/v1/user-groups/:id`
- `GET /api/v1/workflows`
- `GET /api/v1/tools`

## Why This Phase Matters

Relay now has strong backend coverage for messaging, AI, artifacts, search, and notifications. The next missing layer was the broader workspace shell that Slack-like products rely on:

- a useful home surface
- richer people and profile surfaces
- explicit user grouping
- visible workflow and tool entry points

Without this layer, the frontend can only approximate navigation shells. With it, Gemini can wire real surfaces for people, home, and workflow discovery.

## Backend Work

- added richer user fields:
  - `title`
  - `department`
  - `timezone`
  - `working_hours`
- added user group persistence:
  - `user_groups`
  - `user_group_members`
- added workflow and tool registries:
  - `workflow_definitions`
  - `tool_definitions`
- added status mutation that refreshes presence metadata and emits realtime `presence.updated`
- added profile hydration with:
  - local time
  - working hours
  - focus areas
  - top channels
  - recent artifacts

## Frontend Handoff

Gemini can now build:

- a real home page instead of a channel-only landing view
- richer user profile panels and directory detail views
- user group lists and membership detail panels
- workflow/tool launch surfaces and browse panels

## Verification

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm build`
