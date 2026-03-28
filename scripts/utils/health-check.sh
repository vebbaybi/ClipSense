#!/usr/bin/env bash
set -euo pipefail
BASE=${1:-http://localhost:8080}
if curl -fsS "$BASE/api/health" >/dev/null; then
  echo "API healthy at $BASE"
else
  echo "API NOT reachable at $BASE" >&2
  exit 1
fi
