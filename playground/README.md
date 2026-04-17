# Playground

Repo-local testing shell.

Purpose:

- embed the real app from `apps/web`
- inspect live API summary and task data
- show backend schema shape
- trigger reset / seed actions for testing

Container-first run:

```bash
pnpm podman:up
```

Then:

- app: `http://localhost:4174`
- playground: `http://localhost:4175`

Stop:

```bash
pnpm podman:down
```

Local fallback:

```bash
pnpm dev:api
pnpm dev:web
pnpm dev:playground
```
