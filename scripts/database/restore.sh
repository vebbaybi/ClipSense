#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

SRC=${1:-}
DB_PATH=${DB_PATH:-data/clipsense.db}
if [[ -z "$SRC" ]]; then
  echo "Usage: restore.sh <backup-file>" >&2
  exit 1
fi
cp "$SRC" "$DB_PATH"
echo "Restored $SRC to $DB_PATH"
