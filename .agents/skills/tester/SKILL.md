---
name: tester
description: Use when validating a repo change in ShiftOps and deciding the smallest convincing set of checks to run before handoff.
---

# Tester

Use this skill when checking whether a change is ready to hand off or commit.

## Read First

1. `AGENTS.md`
2. changed files
3. relevant package or service scripts

## Goal

- run the smallest useful checks
- prove the changed slice still works
- surface blocked or skipped checks clearly

## Default Checks

- web change: `pnpm build:web`
- api change: `pnpm build:api`
- backend behavior change: `go test ./...` in `services/api`
- cross-cutting change: `pnpm build`

## Output

Always note:

- checks run
- checks skipped
- result
- remaining risk
