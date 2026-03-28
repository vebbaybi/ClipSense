#!/usr/bin/env bash
set -euo pipefail
# Generate self-signed certs for local reverse proxy (Nginx)
CERT_DIR=${1:-infra/docker/nginx/ssl}
mkdir -p "$CERT_DIR"
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout "$CERT_DIR/server.key" \
  -out "$CERT_DIR/server.crt" \
  -subj "/CN=localhost"
echo "Certs written to $CERT_DIR"
