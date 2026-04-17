# API Docs

## Base

```txt
Base URL   http://localhost:8080
Content    application/json
State      SQLite-backed
```

## Endpoints

```txt
GET   /healthz
GET   /v1/tasks
GET   /v1/summary
PATCH /v1/tasks/{id}
POST  /v1/dev/seed
POST  /v1/dev/reset
```

## GET /healthz

```http
GET /healthz
```

```json
{
  "service": "shift-ops-api",
  "status": "ok",
  "version": "0.1.0"
}
```

## GET /v1/tasks

```http
GET /v1/tasks
```

```json
{
  "tasks": [
    {
      "id": "task_quality_01",
      "type": "quality_check",
      "status": "in_progress",
      "title": "Finish quality check already started",
      "vehicle_id": "VOI-3044",
      "location_label": "Sankt Paulsgatan corner",
      "distance_meters": 80,
      "battery_level": 48,
      "urgency": 2,
      "blocked_access_severity": 0,
      "notes": "",
      "checklist_state": [true, false, false],
      "started_at": "2026-04-17T08:15:00Z",
      "completed_at": null,
      "completed_by": null,
      "resolution_code": null
    }
  ]
}
```

## GET /v1/summary

```http
GET /v1/summary
```

```json
{
  "pending": 3,
  "in_progress": 1,
  "completed": 0,
  "skipped": 0,
  "total": 4,
  "events_today": 1,
  "recommended_task": {
    "task": {
      "id": "task_quality_01",
      "status": "in_progress",
      "title": "Finish quality check already started"
    },
    "reasons": [
      "already active",
      "close by"
    ]
  }
}
```

## PATCH /v1/tasks/{id}

```http
PATCH /v1/tasks/task_quality_01
Content-Type: application/json
```

## Mutable body fields

```json
{
  "status": "in_progress",
  "notes": "Operator note",
  "checklist_state": [true, false, false],
  "completed_by": "shift-worker-1",
  "resolution_code": "checked_ok"
}
```

## Transition rules

```txt
pending      -> in_progress
in_progress  -> pending
in_progress  -> completed
in_progress  -> skipped
```

## Terminal resolution codes

```txt
battery_swapped
reparked
checked_ok
collected_for_repair
unable_to_locate
blocked_access
duplicate_task
other
```

## Example: start a task

```json
{
  "status": "in_progress"
}
```

## Example: persist checklist progress

```json
{
  "checklist_state": [true, true, false]
}
```

## Example: complete a task

```json
{
  "status": "completed",
  "completed_by": "shift-worker-1",
  "resolution_code": "checked_ok",
  "notes": "Lights and brakes checked."
}
```

## Example error

```json
{
  "error": "invalid task update: completed_by and resolution_code are only valid for terminal updates"
}
```

## POST /v1/dev/seed

```http
POST /v1/dev/seed
```

```txt
Seeds the database only when it is empty.
```

## POST /v1/dev/reset

```http
POST /v1/dev/reset
```

```txt
Drops current task data and reseeds the SQLite file.
```
