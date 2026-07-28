# ClipSense

ClipSense is an early AI-assisted video pre-editing prototype. The intended product
will eventually include a desktop processing workstation and a web workspace, but
this repository currently contains only a partial web MVP: a Next.js browser app,
a Go HTTP API, a Python processing worker, and local infrastructure definitions.

The repository is in a controlled re-foundation. Start with
[Current State](docs/CURRENT_STATE.md) before attempting runtime work.

## Physical Runtime

| Component | Path | Current purpose |
|---|---|---|
| Web | `apps/web` | Login, ZIP submission, batch list/detail UI |
| API | `apps/api` | Auth, batches, upload, retrieval, export |
| Processor | `apps/api/ai_worker` | ZIP extraction, transcription, analysis, vectors |
| Local stack | `docker-compose.yml` | Postgres, Redis, Qdrant, API, worker, web |

There is no desktop application, shared package workspace, OpenAPI contract,
versioned migration system, or end-to-end test suite yet. Extensionless files that
show directory trees are historical blueprints, not packages or applications.

## Local Runtime

The root `docker-compose.yml` is the authoritative local stack definition. On a
machine with Docker and Docker Compose:

```sh
docker compose config
docker compose build --no-cache
docker compose up -d
docker compose ps
```

The web application is served at `http://localhost:3000`. The public landing page
is `/`; authenticated application routes begin at `/dashboard`. The API exposes:

- `GET http://localhost:8080/api/health/live` for process liveness.
- `GET http://localhost:8080/api/health/ready` for Postgres and Redis readiness.
- `GET http://localhost:8080/api/health` as a compatibility alias for readiness.

Stop the stack normally with:

```sh
docker compose down
```

The authoritative Sprint 1 Linux container gate is
`.github/workflows/docker-runtime.yml`. It builds the application images, starts the
canonical Compose stack, checks dependency and route behavior, verifies readiness
degradation/recovery and graceful termination, and retains commit-attributed
diagnostics. Docker Desktop on the HP or Acer is optional supplemental Windows QA.
Future desktop application testing remains a separate cross-platform concern.

The current Compose defaults are for local development only. In particular,
production-sensitive secret and port hardening remains Sprint 1 Work Unit 2 work.

## Current Limitations

The Docker MVP is not yet demonstrated end to end until the authoritative GitHub
Actions runtime gate passes. Full media processing and cold model acquisition are
outside that boot/integration gate and may require a future manual or scheduled smoke.
Authenticated CSV export is broken, the upload cap is ineffective, and Compose does
not provide a secure JWT secret; those remain Work Unit 2 concerns.

## Inspect And Validate

```powershell
# Repository/environment report
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/doctor.ps1

# Go unit tests
Set-Location apps/api
go test ./...

# Python syntax and ZIP-safety tests
Set-Location ../api/ai_worker
python -m py_compile main.py zip_safety.py
python -m unittest discover -s tests

# Web build
Set-Location ../../../web
npm ci
npm run build
```

Docker commands require Docker Engine and Compose:

```bash
docker compose config
docker compose up --build
```

Do not interpret a successful Compose configuration parse as proof that containers
start or that the product workflow succeeds.

## Repository Truth

- [Documentation index](docs/README.md)
- [Current verified state](docs/CURRENT_STATE.md)
- [Service ownership](docs/architecture/SERVICE_OWNERSHIP.md)
- [Repository inventory](docs/audits/SPRINT_0_REPOSITORY_INVENTORY.md)
- [Capability ownership](docs/audits/SPRINT_0_CAPABILITY_OWNERSHIP.md)
- [Contribution rules](CONTRIBUTING.md)
- [Agent guardrails](AGENTS.md)
- [Kanban workflow](docs/operations/KANBAN_WORKFLOW.md)
