# Relay Agent Workspace Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reposition the current `acim-ui` repository as `Relay Agent Workspace`, create the missing repo-facing materials, update core documentation, and prepare the codebase for migration to the new GitHub repository.

**Architecture:** This migration is a documentation-and-branding refactor performed in-place. The implementation keeps the existing codebase and history intact, adds new top-level materials for the renamed product, updates existing docs selectively to preserve useful historical context, and avoids changing git remotes until a later explicit handoff.

**Tech Stack:** Markdown, JSON (`package.json`), ripgrep, git, Next.js repository docs.

---

## File Map

### Create

- `README.md`
  - New repo landing page for `Relay Agent Workspace`
- `docs/project-positioning.md`
  - Product positioning, naming, audience, and messaging document
- `docs/repo-migration-plan.md`
  - Concrete repository migration checklist from `acim-ui` to `relay-agent-workspace`

### Modify

- `package.json`
  - Rename package from `acim-ui` to `relay-agent-workspace`
- `docs/implementation_plan.md`
  - Update product naming and high-level framing while preserving implementation history
- `docs/task.md`
  - Update project heading and wording to new brand
- `docs/phase8-api-expansion.md`
  - Align current product name and context with `Relay Agent Workspace`

### Inspect

- Repository-wide occurrences of:
  - `acim-ui`
  - `ACIM-UI`
  - `ACIM`
- Existing dirty worktree
  - Ensure migration edits do not overwrite unrelated work

## Task 1: Baseline Scan For Rename Scope

**Files:**
- Inspect: repository-wide text matches

- [ ] **Step 1: Search for old product names**

Run:

```bash
rg -n "acim-ui|ACIM-UI|ACIM" .
```

Expected:
- A list of files containing legacy naming
- Clear separation between migration-relevant docs and unrelated code/data content

- [ ] **Step 2: Review dirty worktree before editing**

Run:

```bash
git status --short
```

Expected:
- Existing unrelated edits are visible
- Migration work can be planned without overwriting user changes

## Task 2: Create Repo-Facing Materials

**Files:**
- Create: `README.md`
- Create: `docs/project-positioning.md`
- Create: `docs/repo-migration-plan.md`

- [ ] **Step 1: Draft `README.md`**

Include:
- Product title: `Relay Agent Workspace`
- One-line positioning statement
- Key capabilities
- Current status
- Local development
- Roadmap
- Repo structure summary

- [ ] **Step 2: Draft `docs/project-positioning.md`**

Include:
- Full name and short name
- Target users
- Product thesis
- Core concepts: workspace, messaging, agents, artifacts, orchestration
- Recommended taglines and GitHub description

- [ ] **Step 3: Draft `docs/repo-migration-plan.md`**

Include:
- Current repo and target repo
- Migration stages
- Remote switching commands
- Validation checklist
- Risk notes for dirty worktree and old naming residue

- [ ] **Step 4: Review all new docs for consistency**

Check:
- Product name is consistent
- README tone is repo-facing, not internal-plan-facing
- Migration doc does not execute remote changes yet

## Task 3: Update Existing Core Docs

**Files:**
- Modify: `package.json`
- Modify: `docs/implementation_plan.md`
- Modify: `docs/task.md`
- Modify: `docs/phase8-api-expansion.md`

- [ ] **Step 1: Update `package.json` package name**

Set:

```json
{
  "name": "relay-agent-workspace"
}
```

- [ ] **Step 2: Update `docs/task.md` title and project wording**

Adjust:
- Top-level project name
- Any explicit `ACIM-UI` references that should now be `Relay Agent Workspace`

- [ ] **Step 3: Update `docs/phase8-api-expansion.md` naming context**

Adjust:
- Current product name references
- Keep API conclusions intact

- [ ] **Step 4: Update `docs/implementation_plan.md` framing**

Adjust:
- Main project naming
- Introductory framing so it reads as historical implementation planning for `Relay Agent Workspace`
- Preserve important implementation details and historical context where useful

- [ ] **Step 5: Re-scan for legacy naming**

Run:

```bash
rg -n "acim-ui|ACIM-UI|ACIM" README.md docs package.json
```

Expected:
- Only intentional historical references remain

## Task 4: Validate Migration Materials

**Files:**
- Verify: `README.md`
- Verify: `docs/project-positioning.md`
- Verify: `docs/repo-migration-plan.md`
- Verify: `package.json`
- Verify: updated core docs

- [ ] **Step 1: Review the final file set**

Run:

```bash
git diff -- README.md docs/project-positioning.md docs/repo-migration-plan.md package.json docs/implementation_plan.md docs/task.md docs/phase8-api-expansion.md
```

Expected:
- Diff contains only migration-related changes

- [ ] **Step 2: Sanity-check package metadata**

Run:

```bash
sed -n '1,80p' package.json
```

Expected:
- `name` is `relay-agent-workspace`

- [ ] **Step 3: Check repository status**

Run:

```bash
git status --short
```

Expected:
- New and modified files are limited to intended migration scope plus pre-existing unrelated edits

## Task 5: Prepare Handoff For Remote Migration

**Files:**
- Reference: `docs/repo-migration-plan.md`

- [ ] **Step 1: Summarize GitHub metadata for copy/paste**

Prepare:
- Repo name
- GitHub description
- Topics
- README heading

- [ ] **Step 2: Summarize next-step remote switch commands without executing**

Prepare commands for later:

```bash
git remote set-url origin https://github.com/nikkofu/relay-agent-workspace
git remote -v
git push -u origin main
```

- [ ] **Step 3: Ask for execution approval before changing remotes**

Do not execute remote changes until the user explicitly requests migration execution.

## Verification Notes

- No code behavior changes are required for this phase
- Verification is documentation integrity plus naming consistency
- Avoid destructive git operations
- Preserve unrelated local changes

## Review Constraints

This environment does not authorize subagent use by default, so plan review should be done inline unless the user explicitly asks for delegated review.
