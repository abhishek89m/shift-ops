# Shift Ops

Shift Ops is a small full-stack monorepo for a mobile-first field-operations app.

This first scaffold commit sets up:

- a Vue + Vite web app shell
- a Go API shell
- container files for both services
- root run/build scripts
- a version sync script instead of a release helper

## Workspace

```text
apps/web
services/api
playground/
scripts/version.mjs
compose.yml
```

## Playground shape

The repo-local playground keeps the same split we want in the final output:

- `apps/web` for frontend iteration
- `services/api` for backend iteration
- `playground/` for notes about how preview/test helpers should wrap those live surfaces without duplicating them

## Run locally

### Web

```bash
pnpm install
pnpm --filter web dev
```

### API

```bash
pnpm dev:api
```

### Build

```bash
pnpm --filter web build
pnpm build:api
pnpm build
```

## Checks

```bash
pnpm check
```

## Container files

The repo includes:

- `apps/web/Dockerfile`
- `services/api/Dockerfile`
- `playground/Dockerfile`
- `compose.yml`

Use either Docker Compose or Podman Compose in the environment you prefer.

### Podman stack

```bash
pnpm podman:up
```

Then use:

- app: `http://localhost:4174`
- playground: `http://localhost:4175`

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
