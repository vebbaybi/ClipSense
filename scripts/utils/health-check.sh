#!/usr/bin/env bash
set -euo pipefail
BASE=${1:-http://localhost:8080}
curl -fsS "$BASE/api/health/live" >/dev/null
echo "API live at $BASE"
curl -fsS "$BASE/api/health/ready" >/dev/null
echo "API ready at $BASE"
