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
- `playground/` = repo-local notes for preview/test helpers
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

## Run

- `pnpm install`
- `pnpm dev:web`
- `pnpm dev:api`
- `pnpm build`
- `pnpm build:web`
- `pnpm build:api`

## Working Rules

- one logical change per commit
- keep commits vertical and inspectable
- do not duplicate product logic in review-only surfaces
- do not add heavy infra unless it clearly lowers risk
- keep validation authoritative in Go

## AI Notes

- default to repo-local `$smol` for chat and machine-facing docs
- keep repo docs concise normal English
- log material AI mistakes or corrections in `.agents/context/ai-audit-log.md`
- keep repo-local decisions in `.agents/context/decision-log.md`

## Verification

- run relevant local checks before handoff
- if a check is skipped or blocked, say so clearly
- prefer the smallest convincing verification for the current commit
