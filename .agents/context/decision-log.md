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
- keep the broader workspace review shell outside the repo; the repo itself ships a smaller `studio/` for demos, docs, and API testing
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
- show worker-facing task codes like `TP-001` instead of raw storage ids in the UI
- header behaves like full-width app chrome, with content scrolling underneath it
- header starts inset and rounded, then snaps into flat top-nav chrome after scroll
- checklist progress persists through the API so refreshes and second devices keep task-step state
- terminal actions require a task to enter `in_progress` first; `pending` can no longer close directly
- the recommended card should be reconciled against the live task list, so it never claims an active task when the current lanes say otherwise

## Backend Decisions

- validation stays authoritative in Go
- no fake shared schema across TS + Go
- keep SQL small and inspectable
- no Swagger UI in the timed window
- repo-local playground should focus on live API endpoint testing before raw schema reference
- moving an active task back to `pending` must clear stale lifecycle metadata while keeping checklist progress
- playground also carries a lightweight docs page backed by repo docs for architecture, guide, and API review
- repo-local preview shell inside the output repo is now named `studio`
- studio docs should read like technical reference, with commands and endpoint examples shown as code blocks
- studio API actions should distinguish read-only refresh from destructive reset/seed controls

## Non-Goals For Now

- auth
- full map
- turn-by-turn navigation
- VROOM integration
- A/B infra
- warehouse logistics engine
