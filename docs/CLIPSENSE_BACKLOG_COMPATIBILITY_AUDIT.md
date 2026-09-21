# ClipSense Backlog Compatibility Audit

## 1. Executive Verdict

**Compatibility percentage:** 68%

**Short verdict:** The proposed product backlog is directionally compatible with ClipSense, but it is not compatible enough to become the official source of truth as-is. It captures the broad product arc from local ZIP processing to AI analysis, creator workspace, remote ingestion, narrative intelligence, and export. It does not sufficiently capture current repo reality: auth, CORS, local runtime blockers, route bugs, Docker packaging bugs, upload security gaps, CI/test details, dependency vulnerabilities, legacy SQLite scripts, and verification gates.

**Acceptable as-is:** No.

**Needs updates before becoming official:** Yes. The backlog should be updated before GitHub issue creation so it does not accidentally mark partially working source as complete product capability.

The backlog is strongest at describing the long-term product direction. It is weakest at describing the current stabilized ZIP MVP state, known bugs, local environment blockers, and hardening work required before moving into Phase 3 or Phase 4.

## 2. Current Repo Product Identity

Based on repo evidence, ClipSense is currently an AI-powered video repurposing and creator workflow MVP in stabilization.

Supported by repo evidence:

- The repo has a Next.js creator-facing web app in `apps/web`.
- The repo has a Go API in `apps/api`.
- The repo has a Python AI worker in `apps/api/ai_worker`.
- The Docker Compose file defines Postgres, Redis, Qdrant, API, worker, and web services.
- The Go API supports register/login, authenticated batch list/detail/export routes, multipart batch upload, Postgres schema bootstrap, Redis enqueue, CORS, and health checks.
- The Python worker supports Redis job consumption, safe ZIP extraction, ffmpeg audio extraction, Whisper transcription, basic summary/mood/role/topic classification, embeddings, KMeans ordering, storyline persistence, and Qdrant upsert in source.
- Current implemented intake path is ZIP upload only.

Not fully supported yet:

- Direct video upload does not exist.
- Remote link intake does not exist.
- Semantic clipping as timestamped clip segmentation is not implemented; current "clips" map to extracted video files.
- Transcript segments with timestamps are not implemented.
- Multi-pane/timeline creator workspace is not implemented.
- Advanced graph-based narrative sequencing is not implemented.
- Rust video core is not implemented.
- Production cloud upload is not implemented.
- Full Docker runtime verification is blocked because Docker is unavailable in the current shell.

The supported framing is:

> ClipSense is an AI-powered video repurposing and creator workflow platform that ingests ZIP video batches, extracts/transcribes/analyzes them, generates basic clip metadata/storyline structure, and prepares creator-visible review/export output.

The broader phrasing in the backlog is valid as a target direction, but it overstates the current repo if read as implemented product status.

## 3. Existing Repo Capability Map

| Capability | Exists? | Evidence | Status |
| ---------- | ------- | -------- | ------ |
| Next.js web app | Yes | `apps/web/package.json`, `apps/web/src/app`, `npm run build` passed | Review |
| Go API | Yes | `apps/api/main.go`, `apps/api/go.mod`, `go test ./...` passed | Done |
| Python worker | Yes | `apps/api/ai_worker/main.py`, `requirements.txt`, worker tests passed | Review |
| Docker Compose | Yes | `docker-compose.yml` defines web, API, worker, Postgres, Redis, Qdrant | Blocked |
| Postgres | Yes | Compose uses `postgres:15-alpine`; API uses pgx; schema created in `migrate()` | Review |
| Redis | Yes | Compose uses `redis:7-alpine`; API `RPush`; worker `blpop` | Review |
| Qdrant | Yes | Compose uses `qdrant/qdrant:v1.8.1`; worker `QdrantClient` upsert | Review |
| Auth/register/login | Yes | `/api/auth/register`, `/api/auth/login`, JWT middleware, frontend login/register form | Review |
| ZIP upload | Yes | `createBatch` accepts multipart `file`; frontend accepts `.zip` via dropzone | Review |
| File storage | Yes | API stores upload under `UPLOAD_DIR`; Compose shares `/data` volume | Review |
| Redis queue | Yes | API enqueues `jobs:batch`; worker consumes `jobs:batch` | Review |
| Worker ZIP extraction | Yes | `zip_safety.py`, tests cover traversal, absolute paths, limits, non-video ignore | Done |
| Worker transcription | Yes in source | `extract_audio()`, `transcribe()` using Whisper | Review |
| Worker classification | Basic | `classify_mood`, `classify_role`, topic hardcoded to `general` | Partial |
| Worker storyline generation | Basic | KMeans ordering and `storylines`/`storyline_clips` inserts | Partial |
| Qdrant upsert | Yes in source | `ensure_qdrant_collection()`, `upsert_embeddings()` | Review |
| Batch list/detail | Yes | `GET /api/batches`, `GET /api/batches/{id}`, frontend pages | Review |
| Export | Basic | `GET /api/batches/{id}/export`, CSV/JSON behavior | Bugged |
| CI | Yes | `.github/workflows/ci.yml` | Review |
| Tests | Yes | `apps/api/main_test.go`, `apps/api/ai_worker/tests/test_zip_safety.py` | Done |
| Docs | Yes | `docs/MVP_STABILIZATION.md`, `docs/CLIPSENSE_KANBAN_STATUS_REPORT.md`, `CS.md`, `1807ish/pbl.md` | Review |
| Environment config | Yes | `.env.example`, Compose env, app env usage | Review |
| Security limits | Partial | ZIP extraction limits exist; API upload body limit and ZIP validation incomplete | Partial |
| Failure visibility | Partial | Failed batches marked failed; failure reason not persisted/displayed | Partial |

## 4. Backlog Item Compatibility Table

| ID | Backlog Item | Compatibility | Repo Evidence | Current Status | Required Update |
| -- | ------------ | ------------- | ------------- | -------------- | --------------- |
| CS-101 | Unified Docker Orchestration and Multi-Stage Environment Layout | Partial match | `docker-compose.yml`, three Dockerfiles | Bugged/Blocked | Add current blocker: Docker unavailable; worker Dockerfile omits `zip_safety.py`; Compose boot unverified. |
| CS-102 | Relational Postgres Schema and DB Migration Harness | Partial match | API `migrate()` creates users, batches, clips, storylines, storyline_clips | Review | Add missing `source_assets`/processing records, foreign keys, cascade behavior, and real migration workflow. |
| CS-105 | Postgres Connection Pooling and Health-Check Handshake | Partial match | API health checks DB; `sql.DB` pool exists by default | Review | Add startup retry/backoff requirement; current API can crash if Postgres is late. |
| CS-103 | Go API Multipart File Intake | Partial match | `createBatch` accepts multipart `file`, saves ZIP path, creates batch | Bugged/Review | Add true request size cap, ZIP validation, orphan cleanup, status-code cleanup. |
| CS-104 | Python Worker Local Extraction and Audio Prep Pipeline | Mostly match | `zip_safety.py`, `extract_audio()`, ffmpeg command, worker tests | Review/Bugged | Fix worker Dockerfile packaging; verify corrupted-video handling and audio prep at runtime. |
| CS-106 | Disk-Space Janitor and TTL Cleanup Daemon | Not found | No cleanup daemon found; worker removes prior workdir only for same batch | Missing | Add as backlog or defer until after basic runtime verification. |
| CS-201 | Go API Redis Task Publisher and Job State Machine | Partial match | API `RPush` to `jobs:batch`; worker `blpop`; statuses pending/processing/complete/failed | Review | Align status naming with backlog (`completed` vs current `complete`), add retry/dead-letter/failure reason. |
| CS-202 | Go API Progress Broadcaster | Not found | No WebSocket/SSE/progress route found | Too advanced for current MVP | Move to Backlog/Icebox until ZIP MVP is verified. |
| CS-206 | Poison Pill Queue Catch-All and Dead Letter Queue | Partial match | Worker catches processing exception and marks failed | Backlog | Add persisted failure reason and real DLQ/retry strategy; current Redis list is not acknowledged. |
| CS-203 | Python Worker Transcription Integration | Partial match | Whisper load/transcribe source exists | Review | Add timestamped transcript segments and persistence model; verify runtime/model behavior. |
| CS-204 | Qdrant Vector Upsert Pipeline | Partial match | Worker encodes texts and upserts to Qdrant | Review | Verify runtime; handle Qdrant failure clearly; decide search/API recovery. |
| CS-205 | AI Context Classifier and Clip Taxonomy | Partial match | Basic mood/role functions; topic hardcoded | Partial | Rename to "Basic classifier" or split advanced structured taxonomy into later work. |
| CS-301 | Multi-Pane Creator Core Interface | Partial match | Basic dashboard/list/detail/upload pages exist | Bugged/Partial | Fix route mismatch and auth hydration; multi-pane workspace is not implemented. |
| CS-302 | Native Video Segment Player | Not found | No video player component found in implemented app files | Too advanced for current MVP | Move to Backlog/Icebox. |
| CS-303 | Timeline Workspace State Store | Not found | No timeline store/editor found | Too advanced for current MVP | Move to Backlog/Icebox. |
| CS-304 | Frontend Network Interceptor and Global Error Toast System | Partial match | `api()` wrapper exists; upload page uses `react-hot-toast` locally | Backlog | Add global error strategy; fix infinite-loading/route/auth cases. |
| CS-401 | Remote Video Link Ingestion Worker | Not found | No link intake route or worker path found | Missing from repo but valid roadmap | Keep in Phase 4/Icebox; not a ZIP MVP blocker. |
| CS-402 | Cloud Multipart Upload Pipeline | Not found | Local disk upload only | Missing from repo but valid roadmap | Keep Phase 4/Icebox; not current MVP. |
| CS-403 | Graph-Based Narrative Connector | Not found | Current storyline uses KMeans ordering | Too advanced for current MVP | Keep post-MVP or later roadmap. |
| CS-404 | Rust Video Core for Scene Boundary Detection | Not found | No Rust files/crates found | Too advanced for current MVP | Keep post-MVP/Icebox. |
| CS-405 | Narrative Configuration Export Module | Partial match | API exports JSON/CSV; frontend has CSV link | Bugged/Partial | Fix authenticated export; split MVP CSV/JSON from advanced editor-ready export. |

## 5. Backlog Items Already Covered By Repo

### CS-101: Unified Docker Orchestration and Multi-Stage Environment Layout

- **Repo files/folders:** `docker-compose.yml`, `apps/web/Dockerfile`, `apps/api/Dockerfile`, `apps/api/ai_worker/Dockerfile`.
- **Evidence:** Compose defines all expected services; each app has a Dockerfile.
- **Done or Review:** Review/Bugged, not Done.
- **Remaining verification needed:** Docker CLI/runtime must work; `docker compose up --build` must boot all services; worker Dockerfile must include `zip_safety.py`.

### CS-102: Relational Postgres Schema and DB Migration Harness

- **Repo files/folders:** `apps/api/main.go`.
- **Evidence:** `migrate()` creates `users`, `batches`, `clips`, `storylines`, and `storyline_clips`.
- **Done or Review:** Review.
- **Remaining verification needed:** Runtime migration in Postgres; relationships/foreign keys/cascades; source asset and processing records are missing from backlog definition.

### CS-103: Go API Multipart File Intake

- **Repo files/folders:** `apps/api/main.go`, `apps/web/src/app/(dashboard)/batches/new/page.tsx`.
- **Evidence:** API parses multipart form, saves uploaded file, inserts batch, enqueues Redis job; frontend uploads `.zip`.
- **Done or Review:** Review/Bugged.
- **Remaining verification needed:** True upload size limit, ZIP type validation, orphan cleanup, runtime upload through web/API.

### CS-104: Python Worker Local Extraction and Audio Prep Pipeline

- **Repo files/folders:** `apps/api/ai_worker/main.py`, `apps/api/ai_worker/zip_safety.py`, `apps/api/ai_worker/tests/test_zip_safety.py`.
- **Evidence:** Safe ZIP extraction and ffmpeg audio conversion source exist; tests pass for ZIP safety.
- **Done or Review:** Review/Bugged.
- **Remaining verification needed:** Docker packaging fix, runtime ffmpeg/Whisper verification, corrupted media behavior.

### CS-105: Postgres Connection Pooling and Health-Check Handshake

- **Repo files/folders:** `apps/api/main.go`, `apps/api/main_test.go`.
- **Evidence:** `/api/health` reports DB/Redis health and tests cover degraded JSON.
- **Done or Review:** Review.
- **Remaining verification needed:** Startup retry/backoff is not implemented.

### CS-201: Go API Redis Task Publisher and Job State Machine

- **Repo files/folders:** `apps/api/main.go`, `apps/api/ai_worker/main.py`.
- **Evidence:** API enqueues `jobs:batch`; worker consumes; statuses include `pending`, `processing`, `complete`, `failed`.
- **Done or Review:** Review.
- **Remaining verification needed:** Runtime queue consumption, status naming decision, failure reason persistence, retry/DLQ strategy.

### CS-203: Python Worker Transcription Integration

- **Repo files/folders:** `apps/api/ai_worker/main.py`.
- **Evidence:** Whisper model load and `model.transcribe()` exist.
- **Done or Review:** Review.
- **Remaining verification needed:** Timestamped segments, persistence, runtime model behavior.

### CS-204: Qdrant Vector Upsert Pipeline

- **Repo files/folders:** `apps/api/ai_worker/main.py`, `docker-compose.yml`.
- **Evidence:** Worker creates/recreates collection and upserts points.
- **Done or Review:** Review.
- **Remaining verification needed:** Runtime Qdrant compatibility and failure handling.

### CS-205: AI Context Classifier and Clip Taxonomy

- **Repo files/folders:** `apps/api/ai_worker/main.py`, API clip schema.
- **Evidence:** Clip rows include mood, role, topic; basic mood/role functions exist.
- **Done or Review:** Partial.
- **Remaining verification needed:** Structured taxonomy, JSON validation, richer topics, runtime verification.

### CS-304: Frontend Network Interceptor and Global Error Toast System

- **Repo files/folders:** `apps/web/src/lib/api/client.ts`, `apps/web/src/app/(dashboard)/batches/new/page.tsx`.
- **Evidence:** API wrapper throws on non-OK responses; upload page uses toasts.
- **Done or Review:** Partial.
- **Remaining verification needed:** Global handling, auth failure behavior, worker failure display, infinite loading prevention.

### CS-405: Narrative Configuration Export Module

- **Repo files/folders:** `apps/api/main.go`, `apps/web/src/app/(dashboard)/batches/[id]/page.tsx`.
- **Evidence:** API can export JSON/CSV; UI has CSV link.
- **Done or Review:** Bugged/Partial.
- **Remaining verification needed:** Authenticated export download, metadata completeness, editor-ready package scope.

## 6. Backlog Items Partially Covered

### CS-101

- **What exists:** Compose and Dockerfiles.
- **What is missing:** Verified boot; worker image includes all source; Docker CLI/runtime availability.
- **MVP-critical:** Yes.
- **Suggested Kanban status:** Bugged/Blocked.

### CS-102

- **What exists:** Bootstrap schema in Go API.
- **What is missing:** source assets table, semantic clip distinction, processing records, foreign keys, cascade behavior, migration tool/history.
- **MVP-critical:** Partly. Current tables may be enough for ZIP MVP, but backlog definition is broader than repo.
- **Suggested Kanban status:** Review.

### CS-103

- **What exists:** Multipart upload and batch creation.
- **What is missing:** hard upload limit, ZIP validation, orphan cleanup, correct client error statuses.
- **MVP-critical:** Yes.
- **Suggested Kanban status:** Bugged.

### CS-104

- **What exists:** Safe extraction, video filtering, ffmpeg audio conversion source.
- **What is missing:** Docker packaging, runtime proof, cleanup, robust corrupted-video policy.
- **MVP-critical:** Yes.
- **Suggested Kanban status:** Bugged/Review.

### CS-105

- **What exists:** Health endpoint and DB check.
- **What is missing:** startup retry/backoff.
- **MVP-critical:** Yes for Docker reliability.
- **Suggested Kanban status:** Review.

### CS-201

- **What exists:** Redis publish/consume and status updates.
- **What is missing:** acknowledged queue, retry/DLQ, persisted failure reason, status naming consistency.
- **MVP-critical:** Yes.
- **Suggested Kanban status:** Review.

### CS-203

- **What exists:** Whisper transcription source.
- **What is missing:** timestamped segments and durable transcript segment schema.
- **MVP-critical:** Basic transcription is MVP-critical; timestamped segment quality may be backlog.
- **Suggested Kanban status:** Review.

### CS-204

- **What exists:** Embedding generation and Qdrant upsert source.
- **What is missing:** runtime verification, search path, resilient failure handling.
- **MVP-critical:** Upsert verification is MVP-critical if Qdrant remains in Compose.
- **Suggested Kanban status:** Review.

### CS-205

- **What exists:** Basic keyword mood/role classification and topic field.
- **What is missing:** structured taxonomy, JSON output, validation, richer roles/topics.
- **MVP-critical:** Basic fields are MVP-useful; advanced taxonomy is not current blocker.
- **Suggested Kanban status:** Partial.

### CS-301

- **What exists:** Dashboard/list/detail/upload pages.
- **What is missing:** actual multi-pane workspace, working `/dashboard` routing, stable auth hydration, responsive creator workflow verification.
- **MVP-critical:** Route/auth fixes are critical; multi-pane workspace is not.
- **Suggested Kanban status:** Bugged/Partial.

### CS-304

- **What exists:** fetch wrapper and localized upload toast.
- **What is missing:** global interceptor, auth error strategy, safe worker failure display.
- **MVP-critical:** Error display is MVP-critical; full global interceptor can follow.
- **Suggested Kanban status:** Backlog.

### CS-405

- **What exists:** CSV/JSON batch export route and CSV UI link.
- **What is missing:** working authenticated UI export, selected clips/timeline edits, editor-ready package.
- **MVP-critical:** Basic export verification is useful but not before ZIP runtime.
- **Suggested Kanban status:** Bugged/Backlog.

## 7. Backlog Items Not Yet Started

### CS-106: Disk-Space Janitor and TTL Cleanup Daemon

- **Why it is valid:** Repeated uploads and processing artifacts will consume disk.
- **Backlog or Icebox:** Backlog.
- **What must happen before it starts:** Verify current upload/worker runtime and define retention behavior.

### CS-202: Go API Progress Broadcaster

- **Why it is valid:** Long-running video processing needs creator-visible progress.
- **Backlog or Icebox:** Backlog, not current P0.
- **What must happen before it starts:** Stabilize ZIP processing and persist processing stages.

### CS-206: Poison Pill Queue Catch-All and Dead Letter Queue

- **Why it is valid:** Redis list queue has weak failure semantics.
- **Backlog or Icebox:** Backlog.
- **What must happen before it starts:** Persist failure reasons and decide retry/DLQ design.

### CS-302: Native Video Segment Player

- **Why it is valid:** Creator review needs clip playback.
- **Backlog or Icebox:** Icebox/Backlog after ZIP MVP.
- **What must happen before it starts:** Persist accessible video asset references and timestamped segments.

### CS-303: Timeline Workspace State Store

- **Why it is valid:** Storyline review/editing is core product value.
- **Backlog or Icebox:** Icebox/Backlog after runtime MVP.
- **What must happen before it starts:** Stabilize storyline persistence and basic review UI.

### CS-401: Remote Video Link Ingestion Worker

- **Why it is valid:** Product direction includes public/user-provided video links.
- **Backlog or Icebox:** Icebox for current ZIP MVP.
- **What must happen before it starts:** ZIP MVP done; source asset abstraction clarified.

### CS-402: Cloud Multipart Upload Pipeline

- **Why it is valid:** Production-scale uploads should not rely on local disk.
- **Backlog or Icebox:** Icebox.
- **What must happen before it starts:** Local MVP done and storage abstraction defined.

### CS-403: Graph-Based Narrative Connector

- **Why it is valid:** Advanced narrative quality is a differentiator.
- **Backlog or Icebox:** Icebox.
- **What must happen before it starts:** Basic clip metadata, embeddings, and storylines verified.

### CS-404: Rust Video Core for Scene Boundary Detection

- **Why it is valid:** Heavy video processing may need performance acceleration.
- **Backlog or Icebox:** Icebox.
- **What must happen before it starts:** Python MVP pipeline proven and bottlenecks measured.

## 8. Repo Features Missing From The Backlog

### Auth/Register/Login

- **Repo evidence:** `apps/api/main.go` exposes `/api/auth/register` and `/api/auth/login`; frontend login/register form exists.
- **Why it matters:** Current MVP uses protected batch routes.
- **Where to add:** Phase 1 or Phase 2 security/auth epic.
- **Suggested ID:** CS-107: Local JWT Auth and Protected Batch Routes.

### CORS

- **Repo evidence:** `localDevCORSMiddleware` in `apps/api/main.go`; `.env.example` has `CORS_ALLOWED_ORIGINS`.
- **Why it matters:** Browser/API integration depends on it.
- **Where to add:** Phase 1 infrastructure.
- **Suggested ID:** CS-108: Local Browser CORS Policy.

### JWT Secret Handling

- **Repo evidence:** API defaults to `dev-secret`; `.env.example` says replace.
- **Why it matters:** Local-only defaults must not become production defaults.
- **Where to add:** Phase 1/2 security baseline.
- **Suggested ID:** CS-109: JWT Secret Configuration and Production Guardrails.

### Worker Dockerfile Missing `zip_safety.py`

- **Repo evidence:** Worker Dockerfile copies only `main.py`; `main.py` imports `zip_safety`.
- **Why it matters:** Worker container will fail at runtime.
- **Where to add:** Phase 1, under CS-101 or separate bug.
- **Suggested ID:** CS-110: Worker Docker Packaging Fix.

### Frontend Dashboard Route Mismatch

- **Repo evidence:** frontend links use `/dashboard`; Next build routes are `/`, `/batches/new`, `/batches/[id]`, `/settings`.
- **Why it matters:** Main web navigation is broken.
- **Where to add:** Phase 3, before CS-301 acceptance.
- **Suggested ID:** CS-305: Fix Dashboard Route Structure.

### Frontend Auth Hydration Redirect Race

- **Repo evidence:** `useAuthToken` starts with `null`; dashboard pages redirect on `!token`.
- **Why it matters:** Logged-in users can be redirected before localStorage loads.
- **Where to add:** Phase 3 frontend reliability.
- **Suggested ID:** CS-306: Stabilize Frontend Auth Hydration.

### Upload Security

- **Repo evidence:** `ParseMultipartForm(200 << 20)` without `MaxBytesReader`; no ZIP type validation.
- **Why it matters:** Upload endpoint is exposed and currently trusts file shape too much.
- **Where to add:** Phase 1 file ingestion.
- **Suggested ID:** CS-111: Upload Size and ZIP Validation Hardening.

### API Status-Code Cleanup

- **Repo evidence:** `httpError` always returns HTTP 500; invalid credentials/missing file use it.
- **Why it matters:** Client errors and auth errors should be actionable and safe.
- **Where to add:** Phase 1/2 API reliability.
- **Suggested ID:** CS-112: API Error Response Semantics.

### Failure Reason Persistence

- **Repo evidence:** worker logs errors and marks failed; schema lacks error field.
- **Why it matters:** UI cannot show clear failure reason.
- **Where to add:** Phase 2 queue/failure handling.
- **Suggested ID:** CS-207: Persist and Display Batch Failure Reasons.

### Dependency Vulnerability Cleanup

- **Repo evidence:** `npm audit` reports 8 vulnerabilities, including critical `next@14.1.4`.
- **Why it matters:** Security triage is part of MVP hardening.
- **Where to add:** Phase 1/CI/security.
- **Suggested ID:** CS-113: Frontend Dependency Vulnerability Cleanup.

### Docker/WSL Runtime Blocker

- **Repo evidence:** `docker` unavailable in current shell; WSL previously reported no installed distributions in status report.
- **Why it matters:** Local runtime cannot be verified without it.
- **Where to add:** Phase 1.
- **Suggested ID:** CS-114: Local Docker Runtime Repair and Verification.

### CI Details

- **Repo evidence:** `.github/workflows/ci.yml` exists.
- **Why it matters:** Backlog says environment setup but does not explicitly track CI verification.
- **Where to add:** Phase 1.
- **Suggested ID:** CS-115: CI Build/Test Validation.

### Legacy SQLite Scripts

- **Repo evidence:** `scripts/database/backup.sh` and `restore.sh` copy `data/clipsense.db`.
- **Why it matters:** Current Docker MVP is Postgres-first; scripts are misleading.
- **Where to add:** Phase 1/database cleanup.
- **Suggested ID:** CS-116: Replace Legacy SQLite Backup/Restore Scripts.

### Docs Freshness

- **Repo evidence:** `docs/MVP_STABILIZATION.md` still contains stale Go/go.sum blocker language.
- **Why it matters:** Planning and verification can drift from repo truth.
- **Where to add:** Phase 1/docs.
- **Suggested ID:** CS-117: Stabilization Documentation Refresh.

### MVP Done Gate as Backlog Work

- **Repo evidence:** `1807ish/pbl.md` includes a done gate, but no individual verification issue.
- **Why it matters:** The gate should be tracked as executable QA work.
- **Where to add:** End of Phase 1/2 or separate MVP verification milestone.
- **Suggested ID:** CS-208: End-to-End ZIP MVP Verification Gate.

## 9. Items Marked Done That Become Not Done Under This Backlog

### Docker Compose Exists

- **Current repo feature:** `docker-compose.yml` and Dockerfiles exist.
- **Current status before backlog comparison:** Review/Blocked.
- **Backlog definition:** `docker compose up` starts full stack.
- **Why it is no longer fully Done:** Runtime cannot be verified, Docker CLI is unavailable, and worker Dockerfile omits required `zip_safety.py`.
- **Gap to fill:** Docker runtime repair, worker Dockerfile fix, successful Compose boot.
- **Suggested Kanban status:** Bugged/Blocked.

### Postgres Schema Exists

- **Current repo feature:** API creates basic tables.
- **Current status before backlog comparison:** Review.
- **Backlog definition:** users, batches, source assets, semantic clips, storylines, related processing records, enforced relationships.
- **Why it is no longer fully Done:** Current schema has no explicit source asset table, processing records, foreign keys, or cascade behavior.
- **Gap to fill:** Schema expansion or backlog narrowing for current MVP.
- **Suggested Kanban status:** Partial.

### Multipart Upload Exists

- **Current repo feature:** API accepts multipart file and creates batch.
- **Current status before backlog comparison:** Review/Bugged.
- **Backlog definition:** ZIP uploads, oversized rejection, batch/source record, no orphan records.
- **Why it is no longer fully Done:** True body size cap, ZIP validation, and orphan cleanup are missing.
- **Gap to fill:** Upload hardening.
- **Suggested Kanban status:** Bugged.

### ZIP Safety Exists

- **Current repo feature:** `zip_safety.py` and tests pass.
- **Current status before backlog comparison:** Done for extraction helper.
- **Backlog definition:** full worker extraction and audio prep pipeline.
- **Why it is no longer fully Done:** Audio prep and corrupted-video isolation are not runtime verified, and Docker image misses `zip_safety.py`.
- **Gap to fill:** Worker packaging/runtime verification.
- **Suggested Kanban status:** Review/Bugged.

### Redis Queue Exists

- **Current repo feature:** API publishes Redis job and worker consumes.
- **Current status before backlog comparison:** Review.
- **Backlog definition:** state machine with `pending`, `processing`, `completed`, `failed`; crash safety.
- **Why it is no longer fully Done:** Current code uses `complete`, not `completed`; Redis list has no ack/retry/DLQ.
- **Gap to fill:** Status naming decision, retry/failure semantics.
- **Suggested Kanban status:** Partial.

### Whisper Transcription Exists In Source

- **Current repo feature:** `model.transcribe()` returns text.
- **Current status before backlog comparison:** Review.
- **Backlog definition:** timestamped transcript blocks.
- **Why it is no longer fully Done:** Current schema stores one transcript string per clip, not timestamped segments.
- **Gap to fill:** Transcript segment model and persistence.
- **Suggested Kanban status:** Partial.

### Classification Exists

- **Current repo feature:** Basic mood/role functions and hardcoded topic.
- **Current status before backlog comparison:** Partial.
- **Backlog definition:** structured taxonomy with JSON validation.
- **Why it is no longer fully Done:** Current classification is heuristic and not structured JSON.
- **Gap to fill:** Taxonomy schema and validation.
- **Suggested Kanban status:** Partial.

### Storyline Generation Exists

- **Current repo feature:** KMeans ordering creates one AI sequence.
- **Current status before backlog comparison:** Partial.
- **Backlog definition:** graph-based narrative connector in CS-403.
- **Why it is no longer fully Done:** Current logic is basic clustering/order, not graph narrative scoring.
- **Gap to fill:** Keep current storyline as MVP basic; defer graph engine.
- **Suggested Kanban status:** Partial/Icebox.

### Export Exists

- **Current repo feature:** API supports JSON/CSV export.
- **Current status before backlog comparison:** Bugged/Partial.
- **Backlog definition:** narrative configuration export package with selected clips, sequence order, transcript/context metadata, editor-ready structure.
- **Why it is no longer fully Done:** UI CSV link lacks auth header, no selected timeline edits, no EDL/editor package.
- **Gap to fill:** Authenticated MVP export first; advanced package later.
- **Suggested Kanban status:** Bugged/Backlog.

## 10. Overreach And Scope Risks

### CS-202: Go API Progress Broadcaster

- **Why valid later:** Creators need live processing updates.
- **Why it should not block current MVP:** Batch status polling/detail can verify ZIP MVP first.
- **Recommended column:** Backlog.

### CS-302: Native Video Segment Player

- **Why valid later:** Review workflow needs playback.
- **Why it should not block current MVP:** Current pipeline must first produce reliable clips/results.
- **Recommended column:** Icebox or Backlog.

### CS-303: Timeline Workspace State Store

- **Why valid later:** Manual storyline arrangement is core product UX.
- **Why it should not block current MVP:** Storyline generation/persistence must be verified first.
- **Recommended column:** Icebox or Backlog.

### CS-401: Remote Video Link Ingestion Worker

- **Why valid later:** Unified intake includes links.
- **Why it should not block current MVP:** Current implemented path is ZIP-only.
- **Recommended column:** Icebox.

### CS-402: Cloud Multipart Upload Pipeline

- **Why valid later:** Required for production-scale large files.
- **Why it should not block current MVP:** Local ZIP upload must be reliable first.
- **Recommended column:** Icebox.

### CS-403: Graph-Based Narrative Connector

- **Why valid later:** Improves product differentiation.
- **Why it should not block current MVP:** Current narrative engine is basic and needs runtime verification first.
- **Recommended column:** Icebox.

### CS-404: Rust Video Core for Scene Boundary Detection

- **Why valid later:** Potential performance optimization.
- **Why it should not block current MVP:** No measured bottleneck yet; Python/ffmpeg MVP path is current.
- **Recommended column:** Icebox.

### CS-405: Narrative Configuration Export Module

- **Why valid later:** Creator handoff/export is important.
- **Why it should not block current MVP:** Current export is basic and bugged; fix MVP export before advanced export packages.
- **Recommended column:** Backlog/Icebox depending on scope.

## 11. Missing MVP Stabilization Items

| Suggested ID | Item | Phase | Column | Priority | Why it matters |
| --- | --- | --- | --- | --- | --- |
| CS-110 | Worker Dockerfile includes all worker modules | Phase 1 | Ready | P0 | Worker container will fail without `zip_safety.py`. |
| CS-114 | Local Docker runtime repair and verification | Phase 1 | Blocked | P0 | Compose cannot boot or be verified. |
| CS-305 | Frontend dashboard route mismatch fix | Phase 3 | Ready | P0 | Main authenticated navigation points to non-existent `/dashboard` routes. |
| CS-306 | Frontend auth hydration redirect fix | Phase 3 | Ready | P0 | Protected pages can redirect before token loads. |
| CS-111 | API true upload request limit | Phase 1 | Ready | P0 | Current upload parsing is not a hard body cap. |
| CS-118 | API ZIP type/signature validation | Phase 1 | Ready | P1 | Non-ZIP files can be accepted into ZIP-only path. |
| CS-112 | API error status semantics cleanup | Phase 1 | Backlog | P1 | Client/auth errors should not return generic 500. |
| CS-208 | End-to-end ZIP upload-to-analysis verification | Phase 2 | Blocked | P0 | Current MVP cannot be called done without real runtime proof. |
| CS-207 | Persist/display worker failure reason | Phase 2 | Backlog | P1 | Failed batches need visible reasons, not logs only. |
| CS-113 | Frontend dependency vulnerability cleanup | Phase 1 | Ready | P1 | `npm audit` reports 8 vulnerabilities, including critical Next.js issues. |
| CS-119 | Python dependency audit path | Phase 1 | Backlog | P2 | Prior audit attempt failed around `openai-whisper` metadata. |
| CS-116 | Replace legacy SQLite backup/restore scripts | Phase 1 | Backlog | P2 | Scripts conflict with Postgres-first Docker MVP. |
| CS-117 | Stabilization docs freshness update | Phase 1 | Ready | P1 | Docs are stale on Go/go.sum status. |
| CS-120 | Auth/JWT local security baseline | Phase 1 | Backlog | P1 | Auth is implemented but absent from backlog and uses local-only defaults. |
| CS-121 | Local CORS baseline | Phase 1 | Backlog | P2 | Browser/API calls depend on configured CORS. |

## 12. Recommended Backlog Updates

### Add

- CS-107: Local JWT Auth and Protected Batch Routes.
- CS-108: Local Browser CORS Policy.
- CS-110: Worker Docker Packaging Fix.
- CS-111: Upload Size and ZIP Validation Hardening.
- CS-112: API Error Response Semantics.
- CS-113: Frontend Dependency Vulnerability Cleanup.
- CS-114: Local Docker Runtime Repair and Verification.
- CS-115: CI Build/Test Validation.
- CS-116: Replace Legacy SQLite Backup/Restore Scripts.
- CS-117: Stabilization Documentation Refresh.
- CS-207: Persist and Display Batch Failure Reasons.
- CS-208: End-to-End ZIP MVP Verification Gate.
- CS-305: Fix Dashboard Route Structure.
- CS-306: Stabilize Frontend Auth Hydration.

### Rename

- Rename CS-205 from "AI Context Classifier and Clip Taxonomy" to "Basic MVP Clip Classification and Taxonomy Foundation" if it remains in current MVP.
- Rename CS-301 from "Multi-Pane Creator Core Interface" to "MVP Creator Dashboard and Batch Review Views" for the current phase.
- Rename CS-405 to "MVP Batch Export and Future Narrative Package Export" or split it.

### Move

- Move CS-202 after basic runtime verification and failure persistence.
- Move CS-302 and CS-303 after ZIP MVP done gate.
- Keep CS-401 and CS-402 in Phase 4/Icebox until local ZIP MVP is complete.
- Keep CS-403 and CS-404 post-MVP.

### Split

- Split CS-102 into:
  - basic MVP schema bootstrap,
  - source asset/processing records,
  - relational constraints/cascades,
  - migration workflow.
- Split CS-103 into:
  - upload route,
  - ZIP validation,
  - upload size hard cap,
  - orphan cleanup.
- Split CS-104 into:
  - safe ZIP extraction,
  - Docker packaging,
  - ffmpeg audio prep,
  - corrupted media behavior,
  - cleanup.
- Split CS-201 into:
  - Redis publisher,
  - worker consumer,
  - status model,
  - failure retry/DLQ.
- Split CS-405 into:
  - current CSV/JSON export fix,
  - editor-ready narrative configuration export.

### Defer To Icebox

- CS-302 Native Video Segment Player.
- CS-303 Timeline Workspace State Store.
- CS-401 Remote Video Link Ingestion Worker.
- CS-402 Cloud Multipart Upload Pipeline.
- CS-403 Graph-Based Narrative Connector.
- CS-404 Rust Video Core for Scene Boundary Detection.
- Advanced scope of CS-405.

### Mark As Done

- Go API tests pass.
- API health handler source/tests.
- Local CORS middleware source/tests.
- Worker ZIP safety helper/tests.
- Browser-facing API URL config.
- Frontend production build.
- Root `.gitignore`.

### Mark As Review

- CS-101 Docker Compose overall runtime.
- CS-102 DB schema runtime and relationship completeness.
- CS-104 worker audio/transcription runtime.
- CS-201 queue end-to-end runtime.
- CS-203 transcription runtime and timestamped segment gap.
- CS-204 Qdrant runtime.
- CS-301 web/API runtime after routing fix.
- CI Compose job.

### Mark As Bugged

- Worker Dockerfile missing `zip_safety.py`.
- Frontend `/dashboard` route mismatch.
- Frontend auth hydration redirect race.
- CSV export link lacks authorization header.
- API upload limit is not a true request body limit.
- API accepts non-ZIP files.
- API status codes return 500 for client/auth errors.
- Legacy SQLite backup/restore scripts conflict with Postgres-first MVP.

## 13. Final Compatibility Score

**Final compatibility percentage:** 68%

**Acceptable as-is:** No.

**Should be updated before GitHub issue creation:** Yes.

The backlog is useful and close enough to become the product backbone, but it must be reconciled with current repo truth before it becomes official.

Top 10 corrections needed:

1. Add current P0 bugs: worker Dockerfile missing `zip_safety.py`, frontend route mismatch, frontend auth hydration race.
2. Add Docker runtime repair and Compose boot verification as explicit P0 work.
3. Add upload hardening: true request limit and ZIP validation.
4. Add API status-code cleanup.
5. Add auth/JWT/CORS/security baseline issues.
6. Add failure reason persistence and display.
7. Split broad schema/upload/worker/queue/export items into MVP-sized issues.
8. Mark direct video upload, link intake, graph narrative, Rust, and cloud upload as later roadmap/Icebox.
9. Add dependency vulnerability cleanup for the frontend.
10. Refresh docs so Go/go.sum status, Docker blocker, and MVP done gate match the current repo.

