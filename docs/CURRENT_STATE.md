# ClipSense Current State

Updated on 2026-09-21 for Gate 1 runtime acceptance. PR #84 is merged.
The integrated candidate `ce9d9929dfec43639a0e1a57d72c0ee10352eb66` passed
[CI and Docker verification](https://github.com/vebbaybi/ClipSense/actions/runs/35644022417).
See [the acceptance record](audits/GATE_1_RUNTIME_ACCEPTANCE.md) for scope,
warning classification, and the final-main sign-off process. Work Unit 2 is unstarted.

## Implemented

- Next.js 14 App Router web source with login/register, batch list, ZIP submission,
  batch detail, basic storyline display, and static settings.
- Go/chi API with separate liveness and dependency readiness, custom JWT
  registration/login, protected batch routes,
  multipart ZIP storage, Redis enqueue, batch retrieval, and JSON/CSV export.
- Python worker source with safe ZIP extraction, FFmpeg audio extraction, Whisper,
  sentence-transformer embeddings, KMeans ordering, Postgres writes, and Qdrant upsert.
- Postgres, Redis, Qdrant, API, worker, and web definitions in Docker Compose,
  including dependency health checks and health-gated service startup.
- Go handler tests, Python ZIP-safety tests, and GitHub Actions jobs.
- A real `/dashboard` App Router segment with batch intake, batch detail, and
  settings routes beneath it.
- An explicit Go HTTP server with bounded request timeouts and graceful
  interrupt/termination handling.
- A GitHub-hosted Ubuntu Docker build and runtime verification gate with exact commit
  attribution, bounded health/lifecycle checks, summaries, and retained diagnostics.
- Authoritative candidate run `35644022417` passed at
  `ce9d9929dfec43639a0e1a57d72c0ee10352eb66`: all application images built; the
  six-service stack became healthy; canonical health/routes, worker boot/stability,
  Redis degradation/recovery, API SIGTERM and restart, evidence upload, and clean shutdown passed.

## Partially Implemented Or Defective

- Build dependency audit reported 14 vulnerabilities, including one critical.
  Runtime acceptance does not establish dependency security or release readiness.

- The export endpoint is protected, but the browser anchor cannot attach its token.
- `ParseMultipartForm` does not enforce the intended total request byte limit.
- Compose omits `JWT_SECRET`, causing the API's `dev-secret` fallback.
- Processing jobs are removed before acknowledgement and are not idempotent.
- Schema changes are unversioned startup DDL.
- The web app has no automatic processing-status refresh.
- Deployment and database backup scripts do not match the active repository layout.

## Not Implemented

- Desktop/Tauri application and local processing controller.
- Direct-video or link intake.
- Shared package workspace or generated API contracts/clients.
- Project/workspace collaboration model.
- Durable processing workflow, retry/DLQ, stage history, or failure reason storage.
- Transcript correction, real media review, semantic search UI, editable storylines,
  editor-ready exports, deletion/retention controls, billing, or cloud deployment.
- Integration and browser end-to-end tests.

## Public Routes

| Source | Public route |
|---|---|
| `src/app/page.tsx` | `/` |
| `src/app/dashboard/page.tsx` | `/dashboard` |
| `src/app/dashboard/batches/new/page.tsx` | `/dashboard/batches/new` |
| `src/app/dashboard/batches/[id]/page.tsx` | `/dashboard/batches/{id}` |
| `src/app/dashboard/settings/page.tsx` | `/dashboard/settings` |

## Startup Configuration

The API reads `API_ADDR`, `DATABASE_URL`, `REDIS_URL`, `UPLOAD_DIR`,
`CORS_ALLOWED_ORIGINS`, and `JWT_SECRET`. The worker reads `DB_DRIVER`,
`DB_PATH` for its legacy SQLite path, `DATABASE_URL`, `UPLOAD_DIR`,
`PROCESS_DIR`, `WHISPER_MODEL`, `REDIS_URL`, `QDRANT_HOST`, `QDRANT_PORT`, and
`QDRANT_COLLECTION`. The browser build reads `NEXT_PUBLIC_API_URL`.

`DB_DRIVER` in the API container is unused, and the worker's `UPLOAD_DIR` is
currently read but unused. Browser-facing `NEXT_PUBLIC_API_URL` must remain a
host-reachable URL, while Compose service URLs use container service names.
`JWT_SECRET` still has a predictable development fallback; secret enforcement is
a Work Unit 2 blocker and this stack is not production-secure.

## Health And Lifecycle

- `/api/health/live` reports only whether the API can answer HTTP.
- `/api/health/ready` checks the API's required Postgres and Redis dependencies
  with a two-second bound and does not expose raw dependency errors.
- `/api/health` remains a compatibility alias for readiness.
- The API uses a 5-second header timeout, 5-minute read/write timeouts to retain
  current upload behavior, a 1-minute idle timeout, and a 15-second shutdown
  deadline.

## Services And Ports

| Service | Host port | Notes |
|---|---:|---|
| Web | 3000 | Browser-facing Next.js |
| API | 8080 | Public local HTTP API |
| Postgres | 5432 | Exposed by current development Compose |
| Redis | 6379 | Exposed by current development Compose |
| Qdrant | 6333 | Exposed by current development Compose |

## Current Machine

- Available: Git, authenticated GitHub CLI, Node 22, npm 10, Go 1.26, Python 3.10,
  Windows PowerShell.
- Unavailable at inspection: Docker CLI/Engine, FFmpeg, PowerShell Core (`pwsh`).
- No root `.env` or `apps/web/.env.local`.
- Workspace is under OneDrive.

GitHub Actions is authoritative for Linux container verification. Docker Desktop on
the HP or Acer is optional supplemental Windows QA, not a mandatory completion gate.
Windows/macOS/Linux desktop application testing remains separate because no desktop
code exists. Full media processing and cold model downloads may require a future
manual or scheduled smoke workflow.
