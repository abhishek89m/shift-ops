---
schema: codex-doc/v1
doc_kind: repo_plan
scope: internal
status: active
audience:
  - abhishek
  - codex
last_updated: 2026-04-17
---

# Shift Ops Plan

## Current Checkpoint

- Commit 1: monorepo + app/API shell + AI scaffold
- Commit 2: quality bar + CI
- Commit 3: repo-local planning context

## Next Work

### 1. API core

Build the first real backend slice in `services/api`:

- SQLite schema
- seed data
- `GET /v1/tasks`
- `GET /v1/summary`
- `PATCH /v1/tasks/:id`
- status transition validation
- task events
- backend tests

### 2. Web MVP

Build the first real app flow in `apps/web`:

- shift overview
- recommended next task card
- grouped task list
- task detail
- `start`, `complete`, `skip`
- summary refresh from API

### 3. Preview harness

Add a review-only preview path that helps test the real app:

- mobile view
- tablet view
- desktop view
- iframe embeds the real app route

## Current Priority Order

1. finish backend task model and endpoints
2. connect the web app to live API data
3. restore the stronger product UX from the earlier prototype work
4. add preview/test harness
5. polish copy, states, and verification

## Review Tickets

### Open

- `RT-001` dedupe recommended task in `apps/web` list sections
  Scope: `apps/web/src/App.vue`, `apps/web/src/components/TaskListPanel.vue`, `tests/e2e/app.spec.ts`
  Status: reopened; current active task is shown in both the recommended card and the in-progress list
- `RT-002` make task switching atomic or rollback-safe
  Scope: `apps/web/src/App.vue`, `apps/web/src/composables/useShiftOps.ts`, `services/api/handlers.go`, `services/api/store.go`, `services/api/main_test.go`
  Status: partial fix landed; keep open until the double-failure gap is closed
- `RT-005` revert optimistic checklist UI on patch failure
  Scope: `apps/web/src/App.vue`, `apps/web/src/composables/useShiftOps.ts`, `tests/e2e/app.spec.ts`

### Done

- `RT-003` disable invalid terminal actions in task detail
  Scope: `apps/web/src/components/TaskDetailPanel.vue`
- `RT-004` clear current web lint warnings so `pnpm check` passes
  Scope: `apps/web/src/components/TaskDetailPanel.vue`, `apps/web/src/components/TaskListPanel.vue`

## Carry Forward Into Next Work

- recommendation should stay primary
- task list should feel operational, not dashboard-like
- task detail should stay compact and decisive
- shell should stay calm and mobile-first

## Verification Gates

After each meaningful step:

- `pnpm check`
- manual app review in browser
- API happy-path check when backend changes land
- keep container flow runnable

## Stop Conditions

Pause and review before moving on if:

- API shape becomes hard to explain in interview
- frontend state starts to need a real store
- preview tooling starts duplicating product logic
- any change weakens the mobile-first task flow
