#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
test "${#GRAFANA_ADMIN_PASSWORD}" -ge 16
docker compose --profile observability up -d prometheus grafana
