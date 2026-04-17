#!/bin/sh

set -eu

pnpm lint:web
pnpm typecheck:web
pnpm build:web
pnpm build:playground

fmt_out=$(gofmt -l services/api)
if [ -n "$fmt_out" ]; then
  echo "$fmt_out"
  exit 1
fi

pnpm test:api
pnpm build:api
