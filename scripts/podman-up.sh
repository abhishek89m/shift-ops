#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

cd "$ROOT_DIR"

podman build -f services/api/Dockerfile -t shift-ops-api:dev .
podman build -f apps/web/Dockerfile -t shift-ops-web:dev .
podman build -f playground/Dockerfile -t shift-ops-playground:dev .

podman rm -f shift-ops-api >/dev/null 2>&1 || true
podman rm -f shift-ops-web >/dev/null 2>&1 || true
podman rm -f shift-ops-playground >/dev/null 2>&1 || true

podman run -d --name shift-ops-api -p 8080:8080 shift-ops-api:dev
podman run -d --name shift-ops-web -p 4174:4173 shift-ops-web:dev
podman run -d --name shift-ops-playground -p 4175:4175 shift-ops-playground:dev

printf '\nShift Ops stack is starting:\n'
printf '  app:        http://localhost:4174\n'
printf '  playground: http://localhost:4175\n'
printf '  api:        http://localhost:8080\n'
