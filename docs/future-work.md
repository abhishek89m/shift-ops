# Future Work

If more time were available, the next valuable work would be:

## Backend

- split [services/api/task_store.go](/Users/abhishek/Documents/Workspace/Companies/Voi/ShiftOps/services/api/task_store.go) further so queries, writes, events, and row decoding live in separate files
- add tighter tests around lifecycle transitions and task-event writes
- add missing handler/error-path coverage for `healthz`, `seed`, unknown task ids, and malformed JSON
- consider making task switching one backend-backed atomic action instead of two sequential PATCH calls

## Web App

- add web unit tests for the composable, task grouping, and task-detail action rules
- make mobile task detail route-backed instead of local state only
- add explicit retry controls for failed loads and updates
- reduce full data refreshes after every PATCH if the interaction model grows

## Browser Coverage

- expand Playwright beyond smoke tests
- cover start, complete, skip, switch-task, checklist persistence, mobile back flow, and tablet split-view behavior
- decide whether browser tests should become part of `pnpm check` or stay separate

## Product Follow-Ups

- add a tiny internal task-load/create surface as `MVP 2`
- add a richer skip flow with reason selection and optional short note
- improve shift summary/accountability details without turning the app into an admin dashboard

## Docs And Repo Shape

- tighten `README.md`, `docs/api.md`, and `docs/guide.md` so they match the live implementation exactly
- add a short human-facing AI/tradeoffs note in the repo
- clean small repo-shell drift, including whether `studio/` stays in the submission shape and whether repo-local skills should exist or be removed from `AGENTS.md`

## Not Planned

- auth
- full map-first experience
- turn-by-turn navigation
- VROOM integration
- experimentation infrastructure
- large admin-dashboard scope
