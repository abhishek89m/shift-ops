#!/bin/sh

set -eu

podman rm -f shift-ops-studio >/dev/null 2>&1 || true
podman rm -f shift-ops-web >/dev/null 2>&1 || true
podman rm -f shift-ops-api >/dev/null 2>&1 || true
