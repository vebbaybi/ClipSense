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

## Current Limitations

The Docker MVP is not presently demonstrated end to end. The worker image omits a
required source file, dashboard links do not match the physical App Router URLs,
authenticated CSV export is broken, the upload cap is ineffective, and Compose does
not provide a secure JWT secret. These are planned for Sprint 1, not silently fixed
in the Sprint 0 foundation.

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

