#!/usr/bin/env bash
set -euo pipefail

# Install dependencies and run API locally
cd "$(dirname "$0")"
go mod tidy
GO111MODULE=on go run main.go
