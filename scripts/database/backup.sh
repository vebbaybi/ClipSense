#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

DB_PATH=${DB_PATH:-data/clipsense.db}
BACKUP_DIR=backups
STAMP=$(date +%Y%m%d_%H%M%S)
mkdir -p "$BACKUP_DIR"
cp "$DB_PATH" "$BACKUP_DIR/clipsense_${STAMP}.db"
echo "Backup written to $BACKUP_DIR/clipsense_${STAMP}.db"
