# Architecture

## Monorepo shape

```txt
apps/web      worker-facing Vue app
services/api  Go + SQLite API
studio        repo-local preview, API lab, and docs shell
```

## Runtime flow

```txt
apps/web --> GET /v1/summary
apps/web --> GET /v1/tasks
apps/web --> PATCH /v1/tasks/{id}

studio --> iframe --> apps/web
studio --> live endpoint runner --> services/api
```

## Data model

```txt
tasks
  current task state
  checklist_state
  lifecycle timestamps

task_events
  append-only state transition log
  notes and resolution history
  events_today source
```

## Interaction model

```txt
mobile
  list -> detail

tablet / desktop
  list + detail split view

operator rule
  one live task at a time
  switching work returns the previous live task to pending
```

## Boundaries

```txt
Go     owns validation, transitions, persistence
Vue    owns presentation, interaction, local view state
Studio owns preview, docs, and endpoint testing
```
