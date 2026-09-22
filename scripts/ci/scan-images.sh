#!/usr/bin/env bash
set -Eeuo pipefail
mkdir -p artifacts/security
trivy_image="aquasec/trivy:0.74.0"
docker pull "$trivy_image"
for service in api worker web; do
  image="$(docker compose images -q "$service" | head -n 1)"
  test -n "$image"
  docker run --rm -v trivy-cache:/root/.cache -v /var/run/docker.sock:/var/run/docker.sock -v "$PWD/artifacts/security:/evidence" \
    "$trivy_image" image --timeout 15m --parallel 2 --scanners vuln,secret --format json --output "/evidence/$service.json" "$image"
  docker run --rm -v trivy-cache:/root/.cache -v /var/run/docker.sock:/var/run/docker.sock -v "$PWD/artifacts/security:/evidence" \
    "$trivy_image" image --timeout 15m --parallel 2 --format cyclonedx --output "/evidence/$service.cdx.json" "$image"
  python scripts/ci/security-policy.py container "artifacts/security/$service.json"
  docker inspect --format '{{.Id}} user={{.Config.User}} created={{.Created}}' "$image" > "artifacts/security/$service-image.txt"
done
# Audit the actual resolved worker environment without changing that environment.
docker compose run --rm --no-deps worker python -m pip list --format=json > artifacts/security/python-installed.json
python - <<'PY'
import json
from pathlib import Path
packages = json.loads(Path('artifacts/security/python-installed.json').read_text())
Path('artifacts/security/python-resolved.txt').write_text(''.join(f"{p['name']}=={p['version']}\n" for p in packages))
PY
python -m pip install pip-audit==2.9.0
status=0
pip-audit --no-deps --disable-pip -r artifacts/security/python-resolved.txt -f json -o artifacts/security/python-audit.json || status=$?
test "$status" -le 1
python scripts/ci/security-policy.py python artifacts/security/python-audit.json
# Configuration findings are evidence; release review must resolve root/exposed-port debt.
docker run --rm -v "$PWD:/src:ro" -v "$PWD/artifacts/security:/evidence" "$trivy_image" config --format json --output /evidence/iac.json /src
