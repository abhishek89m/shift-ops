---
schema: codex-doc/v1
doc_kind: repo_agents
scope: internal
status: active
audience:
  - abhishek
  - codex
last_updated: 2026-04-17
---

# AGENTS

Repo ops for `ShiftOps/`.

## Goal

- ship a small, credible field-worker app
- keep every commit working and explainable
- use AI heavily, keep ownership human-led
- prefer simple systems and low dependency weight

## Repo Shape

- `apps/web` = Vue + TypeScript + Vite frontend
- `services/api` = Go API
- `studio/` = repo-local preview + API docs shell
- `.agents/context/` = repo-local AI memory that travels with the repo
- `.agents/skills/` = repo-local agent skills

## Boundary

- product-facing code and copy belong in this repo
- interview prep, review shells, and temporary workspace-only tooling stay outside this repo
- if a worker would see it in final output, bias it into `apps/web` or `services/api`

## Product Bias

- mobile-first field-worker UX
- route-aware, not map-heavy
- accountability over ceremony
- boring backend over clever backend
- calm, low-noise UI

## Stack

- frontend: Vue 3 + TypeScript + Vite
- backend: Go + SQLite
- package manager: `pnpm`
- vcs: colocated `jj` + `git`

## VCS Flow

- prefer branch-first work, not direct edits on `main`
- default flow:
  - implement on a short-lived branch
  - let Abhishek review
  - commit only after review
  - push branch
  - open PR
- keep `main` clean and releasable
- use `git` for remote compatibility and PR flow
- use `jj` for local history shaping and workspace management
- prefer `jj` for local checkpoints and iterative shaping before a reviewable commit exists
- use `jj new` / `jj describe` freely during local iteration if it reduces friction
- use `jj workspace` when parallel lanes would otherwise collide in one worktree
- use `jj workspace add` only for bounded parallel lanes with low overlap
- do not create extra workspaces unless they materially reduce conflict or context switching

## Run

- `pnpm install`
- `pnpm dev:web`
- `pnpm dev:api`
- `pnpm check`
- `pnpm build`
- `pnpm build:web`
- `pnpm build:api`

## Working Rules

- one logical change per commit
- keep commits vertical and inspectable
- let each commit tell one clear story in the PR
- do not duplicate product logic in review-only surfaces
- do not add heavy infra unless it clearly lowers risk
- keep validation authoritative in Go
- if a later-scope change lands early, update the plan/docs to match reality instead of pretending otherwise

## AI Notes

- default to repo-local `$smol` for chat and machine-facing docs
- keep repo docs concise normal English
- keep the current internal plan in `.agents/context/shift-ops.plan.md`
- log material AI mistakes or corrections in `.agents/context/ai-audit-log.md`
- keep repo-local decisions in `.agents/context/decision-log.md`

## Docs

- human-facing repo docs live in `docs/` when needed
- repo-local plan and agent memory stay in `.agents/context/`

## Verification

- run relevant local checks before handoff
- prefer `pnpm check` as the default repo-wide verification entrypoint
- when reviewing non-trivial diffs, leave inline comments for actionable findings
- if a check is skipped or blocked, say so clearly
- prefer the smallest convincing verification for the current commit
