# ClipSense Current State

Verified on 2026-07-26 at starting commit
`e9bb25fbd00b541f935a5668033915f97a201f6e`.

## Implemented

- Next.js 14 App Router web source with login/register, batch list, ZIP submission,
  batch detail, basic storyline display, and static settings.
- Go/chi API with health, custom JWT registration/login, protected batch routes,
  multipart ZIP storage, Redis enqueue, batch retrieval, and JSON/CSV export.
- Python worker source with safe ZIP extraction, FFmpeg audio extraction, Whisper,
  sentence-transformer embeddings, KMeans ordering, Postgres writes, and Qdrant upsert.
- Postgres, Redis, Qdrant, API, worker, and web definitions in Docker Compose.
- Go handler tests, Python ZIP-safety tests, and GitHub Actions jobs.

## Partially Implemented Or Defective

- The worker container omits `zip_safety.py` and cannot import its required module.
- `(dashboard)` is a route group, so physical pages resolve to `/`, `/batches/new`,
  `/batches/[id]`, and `/settings`; UI navigation incorrectly uses `/dashboard/...`.
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

## Routes In Source

| Source | Actual route |
|---|---|
| `src/app/page.tsx` | `/` |
| `src/app/(dashboard)/page.tsx` | `/` (route collision) |
| `src/app/(dashboard)/batches/new/page.tsx` | `/batches/new` |
| `src/app/(dashboard)/batches/[id]/page.tsx` | `/batches/[id]` |
| `src/app/(dashboard)/settings/page.tsx` | `/settings` |

Sprint 1 must choose whether to add a real `dashboard/` URL segment or change all
navigation to the actual routes, then test the result.

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

Use the HP development machine for Sprint 1 Docker integration if it has Docker
Engine and FFmpeg. Use the Mac later for Safari/macOS validation; no desktop code
exists to validate during Sprint 0.

