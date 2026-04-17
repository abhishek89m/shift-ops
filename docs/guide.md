# Guide

## Main commands

```bash
pnpm install
pnpm check
pnpm dev:web
pnpm dev:api
pnpm dev:studio
```

## Podman stack

```bash
pnpm podman:up
pnpm podman:down
```

## Local URLs

```txt
app    http://localhost:4174
studio http://localhost:4175
api    http://localhost:8080
```

## Recommended review flow

```txt
1. Open Studio -> App
2. Check mobile / tablet / desktop
3. Open Studio -> API
4. Run summary + tasks
5. Try reset / seed
6. Send a PATCH update
7. Confirm the app reflects the changed data
```

## Verification

```txt
pnpm check
  lint
  typecheck
  web build
  studio build
  api tests
  api build
```

Studio is the fast manual verification surface beyond CLI checks.
