#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$ROOT_DIR"

docker compose build
VERSION=${1:-dev}

mkdir -p releases

echo "Packaging docker images with version $VERSION"
docker save $(docker compose config --services | xargs -I{} docker compose images --quiet {}) > "releases/clipsense-${VERSION}.tar"

echo "Release artifact: releases/clipsense-${VERSION}.tar"
