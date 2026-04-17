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

- monorepo/runtime shell is in place
- CI and repo quality bar exist
- backend MVP exists in `services/api`
- web MVP exists in `apps/web`
- `studio/` works as the repo-local review surface
- repo docs and Playwright smoke coverage exist

## Next Work

### 1. Hardening

- make task switching atomic or truly rollback-safe
- revert checklist UI on PATCH failure
- finish docs drift cleanup in:
  - `README.md`
  - `docs/api.md`
  - `docs/guide.md`
- expand browser coverage beyond current smoke paths

### 2. Code Cleanup

- split `services/api/task_store.go` further if time still exists
- add web unit tests for composables and task-action rules
- decide whether browser tests should join `pnpm check` or stay separate

### 3. Optional MVP 2

- tiny internal task-load/create surface
- richer skip reason flow
- slightly better shift summary/accountability details

## Current Priority Order

1. fix docs drift and keep the repo explainable
2. make task switching safer
3. revert checklist UI on patch failure
4. deepen browser and web test coverage
5. only then do extra cleanup or MVP 2 work

## Review Tickets

### Open

- `RT-002` make task switching atomic or rollback-safe
  Scope: `apps/web/src/App.vue`, `apps/web/src/composables/useShiftOps.ts`, `services/api/handlers.go`, `services/api/store.go`, `services/api/main_test.go`
  Status: partial fix landed; keep open until the multi-step switch gap is closed
- `RT-005` revert optimistic checklist UI on patch failure
  Scope: `apps/web/src/App.vue`, `apps/web/src/composables/useShiftOps.ts`, `tests/e2e/app.spec.ts`

### Done

- `RT-001` keep recommended work visible without making tasks disappear from sections
  Scope: `apps/web/src/App.vue`, `apps/web/src/components/TaskListPanel.vue`, `tests/e2e/app.spec.ts`
  Status: resolved by keeping active work visible in `In progress` and recommended pending work visible in `Remaining`
- `RT-003` disable invalid terminal actions in task detail
  Scope: `apps/web/src/components/TaskDetailPanel.vue`
- `RT-004` clear current web lint warnings so `pnpm check` passes
  Scope: `apps/web/src/components/TaskDetailPanel.vue`, `apps/web/src/components/TaskListPanel.vue`

## Carry Forward Into Next Work

- recommendation should stay primary
- task list should feel operational, not dashboard-like
- task detail should stay compact and decisive
- shell should stay calm and mobile-first
- recommended tasks should not disappear from grouped sections

## Verification Gates

After each meaningful step:

- `pnpm check`
- `npx playwright test` when web interaction rules change
- manual app review in browser
- API happy-path check when backend changes land

## Stop Conditions

Pause and review before moving on if:

- API shape becomes hard to explain in interview
- frontend state starts to need a real store
- preview tooling starts duplicating product logic
- any change weakens the mobile-first task flow
