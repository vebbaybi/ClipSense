#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$ROOT_DIR"

echo "Stopping and removing containers/volumes..."
docker compose down -v

echo "Removing data directory..."
rm -rf data
mkdir -p data/uploads data/processing
