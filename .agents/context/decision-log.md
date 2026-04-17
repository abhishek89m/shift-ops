---
schema: codex-doc/v1
doc_kind: repo_decision_log
scope: internal
status: active
audience:
  - abhishek
  - codex
last_updated: 2026-04-17
---

# Decision Log

Repo-local working memory for `ShiftOps/`.

## Fixed Decisions

- keep `apps/web` and `services/api` as the main playground + final output boundary
- do not recreate separate frontend/backend prototype copies
- keep `studio/` outside the repo as review shell only
- stack stays `Vue + TS + Vite` and `Go + SQLite`
- mobile-first field-worker UX
- AI use is heavy, but ownership stays human-led

## Current MVP

- shift overview
- one recommended next task
- grouped task list
- task detail with `start`, `complete`, `skip`
- summary / progress
- boring API validation and SQLite persistence

## Frontend Decisions

- static copy lives in templates unless there is real i18n value
- real i18n is in scope early: `en` + `sv`
- top-right settings icon opens language choice
- no `Pinia` yet; local state + composables first
- progress track stays inside the primary shell card

## Backend Decisions

- validation stays authoritative in Go
- no fake shared schema across TS + Go
- keep SQL small and inspectable
- no Swagger UI in the timed window

## Non-Goals For Now

- auth
- full map
- turn-by-turn navigation
- VROOM integration
- A/B infra
- warehouse logistics engine
