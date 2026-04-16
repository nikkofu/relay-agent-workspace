# Relay Agent Workspace Repository Migration Plan

## Purpose

This document captures the concrete migration path from the current repository identity to the new GitHub repository:

- Current repository identity: `acim-ui`
- Target repository: `https://github.com/nikkofu/relay-agent-workspace`

## Migration Strategy

The migration uses an in-place rename strategy:

1. Update branding and repo-facing materials inside the current codebase
2. Preserve history and current implementation documents
3. Review migration-specific changes
4. Switch the git remote only after explicit approval

## Stage 1: Prepare In Place

Complete these changes in the current repository before touching remotes:

- create `README.md`
- add product positioning document
- add repository migration document
- rename package metadata to `relay-agent-workspace`
- update core docs to `Relay Agent Workspace`
- scan for old `acim-ui` and `ACIM-UI` references

## Stage 2: Review Scope

Before switching remotes:

- verify which modified files belong to the migration
- separate migration edits from unrelated local work
- confirm repository-facing copy and GitHub metadata
- confirm no accidental functional code changes were introduced

## Stage 3: Switch Remote

Run only after explicit approval:

```bash
git remote set-url origin https://github.com/nikkofu/relay-agent-workspace
git remote -v
git push -u origin main
```

## Validation Checklist

- `README.md` exists and uses `Relay Agent Workspace`
- `package.json` uses `relay-agent-workspace`
- core docs no longer present `ACIM-UI` as the active product name
- repo description and topics are ready for GitHub settings
- remote switch commands are prepared but not yet executed

## GitHub Settings To Apply

### Repository Name

- `relay-agent-workspace`

### Description

- `AI-native collaboration workspace for humans and agents, combining messaging, artifacts, and orchestration.`

### Topics

- `ai`
- `agents`
- `multi-agent`
- `collaboration`
- `workspace`
- `messaging`
- `realtime`
- `nextjs`
- `typescript`
- `ui`

## Risks

### Dirty Worktree

This repository already contains local modifications unrelated to the rename. Review `git status --short` carefully before staging or pushing.

### Historical Naming Residue

Some implementation docs may continue to mention `acim-ui` or `ACIM-UI` as historical context. That is acceptable if the active product identity is clearly `Relay Agent Workspace`.

### Remote Timing

Do not switch the remote before branding and documentation are in a reviewable state.

