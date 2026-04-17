# Shift Ops

Shift Ops is a small full-stack monorepo for a mobile-first field-operations app.

The easiest way to review the current result is to run `studio/`, which acts as the repo-local review surface for:

- app demos
- viewport previews
- API inspection
- repo docs and review flow

## Workspace

```text
apps/web
services/api
studio/
scripts/version.mjs
compose.yml
```

## Studio shape

The repo-local studio keeps the same split we want in the final output:

- `apps/web` for frontend iteration
- `services/api` for backend iteration
- `studio/` for preview/test helpers that wrap those live surfaces without duplicating them

## Quick Start

### Easiest full demo

This is the fastest way to see the full experience, including:

- the worker app
- the Studio App preview
- the Studio API playground
- the Studio docs page

```bash
pnpm install
pnpm podman:up
```

Then open:

- studio: `http://localhost:4175`
- app: `http://localhost:4174`
- api: `http://localhost:8080`
- future ideas: `docs/future-work.md`

Stop the stack with:

```bash
pnpm podman:down
```

### Local dev, all together

If you prefer to run the services separately, use 3 terminals:

```bash
pnpm install
pnpm dev:api
pnpm dev:web
pnpm dev:studio
```

Then open:

- studio: `http://localhost:4175`

Important:

- Studio `App` embeds the real web app from `http://localhost:4174`
- Studio `API` talks to the backend at `http://localhost:8080`
- so for the full Studio demo, both `dev:web` and `dev:api` should be running
- `docs/future-work.md` captures the most useful next steps that were deliberately left out of this timed build

## Run Individually

### Web app

```bash
pnpm install
pnpm dev:api
pnpm dev:web
```

Open:

- app: `http://localhost:4174`

### API

```bash
pnpm install
pnpm dev:api
```

Open:

- api: `http://localhost:8080`

### Studio only

```bash
pnpm install
pnpm dev:studio
```

Open:

- studio: `http://localhost:4175`

Note:

- this is useful for reading docs and checking the shell itself
- the Studio `App` tab still expects the real web app on `http://localhost:4174`
- the Studio `API` tab still expects the backend on `http://localhost:8080`

## Checks

```bash
pnpm check
```

## Container files

The repo includes:

- `apps/web/Dockerfile`
- `services/api/Dockerfile`
- `studio/Dockerfile`
- `compose.yml`

Use either Docker Compose or Podman Compose in the environment you prefer.

### Podman stack

```bash
pnpm podman:up
```

Then use:

- studio: `http://localhost:4175`
- app: `http://localhost:4174`
- api: `http://localhost:8080`

Stop it with:

```bash
pnpm podman:down
```

## Version sync

Update the root package version, then run:

```bash
pnpm version:sync
```

That syncs version values into:

- `apps/web/package.json`
- `services/api/internal/buildinfo/version.go`
