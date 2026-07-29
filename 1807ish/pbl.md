# ClipSense Product Backlog

ClipSense is an **AI-powered video repurposing and creator workflow platform**.

It turns uploaded or linked video sources into structured creator-ready assets through ingestion, transcription, semantic clipping, AI classification, storyline sequencing, review, and export.

## User Stories

These user stories translate the ClipSense product backlog into creator, developer, and operator value.

Agile operating rule: this backlog is the source of board flow. Repository evidence can inform a card, but it does not make a card Done. A card becomes Done only after it is pulled, verified against acceptance criteria in this delivery cycle, reviewed, and accepted.

They are split into:

- **Stabilized ZIP MVP stories:** compatible with the current repo and required for MVP stabilization.
- **Roadmap stories:** compatible with the broader product direction, but not required before the ZIP MVP is verified.

---

# Stabilized ZIP MVP User Stories

## US-000: Transform Raw Creator Footage Into Story-Ready Output

**As a creator or editor,**
I want to submit raw video sources and have ClipSense organize them into analyzed, story-aware creator-ready material,
so that I can move from chaotic footage to a structured edit plan without manually watching and sorting everything first.

**Long-term product acceptance criteria:**

- User can submit supported source inputs: public/user-provided video links, direct video uploads, and ZIP batches.
- System stores the submitted source asset.
- System extracts usable video/audio assets.
- System generates transcript or processing output.
- System produces clip metadata and AI classification.
- System creates a basic storyline, sequence, or timeline structure.
- User can review the result in the web UI.
- User can export or hand off the structured result.
- Failures are visible and understandable.

**Current stabilized ZIP MVP interpretation:**

**As a creator,**
I want to upload a ZIP of raw video clips and see analyzed clips plus a basic storyline,
so that I can quickly understand and organize my footage before editing.

**Current MVP acceptance criteria:**

- User can register and log in.
- User can upload a ZIP containing supported video files.
- API stores the source ZIP and creates a batch.
- Redis queues a worker job.
- Worker consumes the job.
- Worker safely extracts supported videos.
- Worker transcribes or otherwise processes the media.
- Worker writes clips and storyline data to Postgres.
- Worker upserts vectors to Qdrant or reports Qdrant failure clearly.
- Batch detail shows analyzed clips, status, and storyline results or a clear failure reason.

**Related backlog items:** CS-101, CS-103, CS-104, CS-110, CS-111, CS-114, CS-201, CS-203, CS-204, CS-205, CS-207, CS-208, CS-301, CS-305, CS-306, CS-405
**Priority:** P0
**Status:** In Progress
**Kanban column:** In Progress

**Known risks to resolve through ordered backlog cards:**

- Docker runtime verification.
- Worker Docker packaging.
- Frontend route/auth bugs.
- Upload hardening.
- Failure reason persistence.

---

## US-001: ZIP Batch Upload

**As a creator,**
I want to upload a ZIP file containing multiple videos,
so that ClipSense can process a full batch of source material at once.

**Related backlog items:** CS-103, CS-111
**Priority:** P0
**Status:** Backlog
**Kanban column:** Backlog

**Acceptance Criteria:**

- User can select or drag a `.zip` file in the web app.
- API accepts valid ZIP uploads.
- API rejects oversized uploads safely.
- API rejects non-ZIP files before enqueueing.
- A successful upload creates a batch record.
- A successful upload queues a Redis job.

---

## US-002: Safe ZIP Extraction

**As a creator,**
I want ClipSense to safely extract only supported videos from my ZIP,
so that unsafe files, junk files, or corrupted archives do not break processing.

**Related backlog items:** CS-104, CS-110
**Priority:** P0
**Status:** Backlog
**Kanban column:** Backlog

**Acceptance Criteria:**

- Worker safely rejects path traversal.
- Worker rejects absolute paths.
- Worker ignores unsupported non-video files.
- Worker respects file count and decompressed size limits.
- Worker container includes all required Python modules.
- Worker does not crash because of missing `zip_safety.py`.

---

## US-003: AI Video Processing

**As a creator,**
I want ClipSense to process my uploaded videos with AI,
so that I can get transcripts, clip summaries, mood, role, topic, and storyline structure without watching everything manually.

**Related backlog items:** CS-203, CS-204, CS-205
**Priority:** P1
**Status:** Backlog
**Kanban column:** Backlog

**Acceptance Criteria:**

- Worker extracts audio from supported videos.
- Worker produces transcript text.
- Worker creates basic clip metadata.
- Worker classifies mood, topic, and clip role.
- Worker generates basic storyline ordering.
- Worker upserts vectors to Qdrant or reports failure clearly.

---

## US-004: Batch Result Review

**As a creator,**
I want to view my processed batch results in the web app,
so that I can review clips, transcripts, analysis, and storyline output after processing.

**Related backlog items:** CS-301, CS-305, CS-306
**Priority:** P0
**Status:** Backlog
**Kanban column:** Backlog

**Acceptance Criteria:**

- User can navigate to the batch list.
- User can open a batch detail page.
- Batch detail page displays clips returned by the API.
- Batch detail page displays storyline data where available.
- Frontend links point to real Next.js routes.
- Logged-in users are not falsely redirected during token hydration.

---

## US-005: Clear Processing Failure Reasons

**As a creator,**
I want to see why a batch failed,
so that I know whether the issue was my file, the upload, the worker, transcription, or another system problem.

**Related backlog items:** CS-207, CS-304
**Priority:** P1
**Status:** Backlog
**Kanban column:** Backlog

**Acceptance Criteria:**

- Failed batches store a safe failure reason.
- API returns the failure reason on batch detail.
- Web app displays the failure reason clearly.
- Raw stack traces or secrets are not exposed.
- User is not left with an infinite loading state.

---

## US-006: Poison Pill Job Handling

**As a developer,**
I want corrupted or fatal worker jobs to fail safely,
so that one broken upload does not trap the worker in a retry loop or stop the queue.

**Related backlog items:** CS-206, CS-207
**Priority:** P1
**Status:** Backlog
**Kanban column:** Backlog

**Acceptance Criteria:**

- Worker catches fatal processing errors.
- Failed jobs are marked `failed` in Postgres.
- Safe failure reason is persisted.
- Worker moves on to the next job.
- Corrupted media cannot block the queue forever.
- Dead-letter or terminal failure behavior is documented.

---

## US-007: Disk Cleanup For Large Media

**As an operator,**
I want temporary uploads, extracted videos, and audio files cleaned safely,
so that repeated video processing does not fill the server disk.

**Related backlog items:** CS-106
**Priority:** P2
**Status:** Backlog
**Kanban column:** Backlog

**Acceptance Criteria:**

- Temporary processing folders are cleaned after successful processing.
- Failed batches retain only necessary diagnostic artifacts.
- Cleanup does not delete active in-progress files.
- Cleanup behavior is documented.
- Disk usage does not grow forever during repeated uploads.

---

## US-008: Full Local Runtime Verification

**As a developer,**
I want Docker Compose to boot the full ClipSense stack,
so that the ZIP upload-to-analysis MVP can be verified end to end.

**Related backlog items:** CS-101, CS-114, CS-208
**Priority:** P0
**Status:** Backlog
**Kanban column:** Backlog

**Acceptance Criteria:**

- `docker --version` works.
- `docker compose version` works.
- `docker compose config` validates.
- `docker compose up --build` starts web, API, worker, Postgres, Redis, and Qdrant.
- API health endpoint responds.
- Web app loads locally.
- Worker starts without missing module errors.

---

## US-009: Upload And API Error Clarity

**As a creator,**
I want upload, auth, and API errors to be clear,
so that I understand what went wrong instead of seeing a generic failure.

**Related backlog items:** CS-112, CS-304
**Priority:** P1
**Status:** Backlog
**Kanban column:** Backlog

**Acceptance Criteria:**

- Bad input returns a 400-level response.
- Missing or invalid auth returns 401.
- Oversized uploads return 413.
- Duplicate or conflict cases return 409 where applicable.
- Frontend displays safe error messages.
- API does not return 500 for normal user mistakes.

---

## US-010: Authenticated Export Download

**As a creator,**
I want to export my processed batch results,
so that I can use the analysis outside ClipSense.

**Related backlog items:** CS-405
**Priority:** P1/P2
**Status:** Backlog
**Kanban column:** Backlog

**Acceptance Criteria:**

- Export works for authenticated users.
- Frontend does not rely on a plain anchor that cannot send the bearer token.
- Export includes batch, clips, and storyline data where available.
- CSV or JSON export works at MVP quality.
- Advanced editor-ready export remains future scope.

---

# Roadmap User Stories

These stories are compatible with the product direction, but they should not block the stabilized ZIP MVP.

---

## US-011: Direct Video Upload

**As a creator,**
I want to upload a single video file directly,
so that I do not need to create a ZIP for one source video.

**Related backlog items:** Direct Video Upload Intake
**Priority:** P3
**Status:** Icebox
**Kanban column:** Icebox

**Acceptance Criteria:**

- User can upload supported single video files.
- Supported formats are documented.
- Direct video upload normalizes into the same processing pipeline as ZIP uploads.
- Unsupported formats fail safely.

---

## US-012: Public Or User-Provided Link Intake

**As a creator,**
I want to submit a video link,
so that ClipSense can ingest source material without manual download and upload.

**Related backlog items:** CS-401
**Priority:** P3
**Status:** Icebox
**Kanban column:** Icebox

**Acceptance Criteria:**

- User can submit a supported source link.
- Worker extracts media safely.
- Unsupported links fail clearly.
- Link intake normalizes into the same asset pipeline as uploads.

---

## US-013: Video Segment Preview

**As a creator,**
I want to preview the exact video segment linked to a transcript or clip,
so that I can quickly judge whether a moment is useful.

**Related backlog items:** CS-302
**Priority:** P3
**Status:** Icebox
**Kanban column:** Icebox

**Acceptance Criteria:**

- Video player loads the selected source asset.
- Clicking a clip or transcript segment seeks to the correct timestamp.
- Missing media shows a safe fallback.
- Preview does not block ZIP MVP stabilization.

---

## US-014: Timeline Story Workspace

**As a creator,**
I want to arrange clips into a timeline or storyline workspace,
so that I can shape the final reaction or repurposed content sequence.

**Related backlog items:** CS-303
**Priority:** P3
**Status:** Icebox
**Kanban column:** Icebox

**Acceptance Criteria:**

- User can add clips to a timeline.
- User can reorder clips.
- Timeline state is represented as structured JSON.
- Timeline can later feed export workflows.

---

## US-015: Advanced Narrative Intelligence

**As a creator,**
I want ClipSense to suggest stronger narrative sequences,
so that my final content has better pacing, setup, escalation, climax, and payoff.

**Related backlog items:** CS-403
**Priority:** P3
**Status:** Icebox
**Kanban column:** Icebox

**Acceptance Criteria:**

- Clips can be scored for narrative compatibility.
- Suggested order uses semantic similarity, mood pacing, and role/taxonomy signals.
- Output improves beyond basic clustering.
- This does not block the current ZIP MVP.

---

## US-016: Scaled Cloud Uploads

**As a creator,**
I want large videos to upload reliably through cloud storage,
so that production-scale media does not overload the API server.

**Related backlog items:** CS-402
**Priority:** P3
**Status:** Icebox
**Kanban column:** Icebox

**Acceptance Criteria:**

- Client can request presigned upload metadata.
- Large files upload directly to object storage.
- API records source metadata after upload completion.
- Local disk upload is no longer required for production-scale files.

---

## US-017: Faster Video Processing Engine

**As an operator,**
I want heavy video processing to be accelerated where needed,
so that ClipSense can handle larger files and more users efficiently.

**Related backlog items:** CS-404
**Priority:** P3
**Status:** Icebox
**Kanban column:** Icebox

**Acceptance Criteria:**

- Performance bottleneck is measured first.
- Rust or native engine improves a real bottleneck.
- Existing Python worker pipeline remains stable.
- Acceleration is not started before MVP runtime proof.

---

## US-018: Editor-Ready Export Package

**As a creator,**
I want an editor-ready export package,
so that I can move ClipSense storylines into tools like Premiere, Resolve, or another editing workflow.

**Related backlog items:** CS-405
**Priority:** P3
**Status:** Icebox
**Kanban column:** Icebox

**Acceptance Criteria:**

- Export includes selected clips.
- Export includes sequence order.
- Export includes transcript/context metadata.
- Export format is documented.
- Advanced export waits until timeline/workspace features exist.

## Product Flow

```text
source input
-> stored source asset
-> extracted video/audio assets
-> transcript
-> semantic clips
-> metadata and AI classification
-> storyline/timeline workspace
-> export package
```

## Implementation Context

The current MVP target path is **ZIP upload only**:

```text
ZIP upload through web
-> Go API stores source archive and creates batch
-> Redis job is queued
-> Python worker consumes job
-> safe ZIP extraction
-> supported videos are processed
-> clips/storyline records are written to Postgres
-> vectors are upserted to Qdrant
-> API/web display batch result
```

Long-term product intake includes:

- Public or user-provided video links.
- Direct uploaded video files such as `.mp4`, `.mov`, `.mkv`, `.webm`, or `.avi`.
- ZIP archives containing multiple video files.

Direct video upload and link intake are **not** current stabilized ZIP MVP requirements. They belong in the roadmap after the ZIP MVP is verified.

## Repository Evidence Notes

These notes inform backlog acceptance criteria only. They do not assign Kanban status.

- Go, Python worker, and frontend build checks have been run previously and can be used as reference commands when the relevant cards are pulled.
- Docker runtime, web/auth runtime, and full ZIP upload-to-analysis verification still need to be accepted through the board.
- Known repo findings such as Docker packaging, route, auth hydration, upload validation, and dependency issues remain card acceptance work, not automatic board status.

## Status Definitions

- **In Progress:** currently being worked.
- **Ready:** next planned work with clear acceptance criteria.
- **Backlog:** ordered planned work that should not be pulled before Ready items.
- **Review:** completed in the current cycle and waiting for acceptance.
- **Blocked:** a pulled card cannot progress because of an external dependency or environment failure.
- **Done:** accepted after current-cycle verification. Done starts empty after this agile reset.
- **Icebox:** valuable future work that must not block the stabilized ZIP MVP.

## Initial Kanban Board

The board starts from agile flow, not from repository inspection.

| Column | Cards |
| --- | --- |
| In Progress | `US-000` |
| Ready | `CS-101`, `CS-102`, `CS-107` |
| Backlog | `CS-108`, `CS-109`, `CS-103`, `CS-111`, `CS-112`, `CS-201`, `CS-114`, `CS-110`, `CS-104`, `CS-203`, `CS-205`, `CS-204`, `CS-207`, `CS-301`, `CS-305`, `CS-306`, `CS-304`, `CS-405`, `CS-113`, `CS-115`, `CS-208`, `CS-117`, `CS-105`, `CS-106`, `CS-116`, `CS-119`, `CS-202`, `CS-206` |
| Review | Empty |
| Blocked | Empty until a pulled card is blocked during execution |
| Done | Empty until accepted in this agile cycle |
| Icebox | `US-011`, `US-012`, `US-013`, `US-014`, `US-015`, `US-016`, `US-017`, `US-018`, `CS-302`, `CS-303`, `CS-401`, `CS-402`, `CS-403`, `CS-404`, advanced scope of `CS-405` |

---

# Phase 1: Foundation, Security, and Local Runtime

**Focus:** stabilize local runtime, database, auth, upload path, Docker packaging, security baseline, and verification harness.

## Epic 1: Architecture and Infrastructure Foundation

### CS-101: Unified Docker Orchestration and Multi-Stage Environment Layout

**Status:** Ready  
**Priority:** P0

**Description:**  
Standardize local development with `docker-compose.yml` containing the web app, Go API, Python worker, Postgres, Redis, and Qdrant. Configure Dockerfiles for Next.js, Go API, and Python worker services.

**Current repo evidence:**

- `docker-compose.yml` exists and defines web, API, worker, Postgres, Redis, and Qdrant.
- `apps/web/Dockerfile`, `apps/api/Dockerfile`, and `apps/api/ai_worker/Dockerfile` exist.
- Docker is currently unavailable on `PATH`, so runtime boot is blocked.
- Worker Dockerfile currently omits `zip_safety.py`, which `main.py` imports.

**Acceptance Criteria:**

- `docker compose config` validates from repo root.
- `docker compose up --build` starts the full local stack.
- App, API, worker, Postgres, Redis, and Qdrant run on an isolated local network.
- Service environment variables are correctly passed through Compose.
- Each service has its own Dockerfile or clear build path.
- Worker image includes every Python module required at runtime.

---

### CS-102: Relational Postgres Schema and DB Migration Harness

**Status:** Ready  
**Priority:** P1

**Description:**  
Build the core Postgres schema for users, batches, source assets, semantic clips, storylines, and related processing records. Keep current MVP schema honest: the repo currently bootstraps basic tables in the Go API, not a full migration system.

**Current repo evidence:**

- `apps/api/main.go` creates `users`, `batches`, `clips`, `storylines`, and `storyline_clips`.
- Explicit source asset, transcript segment, processing record, and failure reason tables are not present.
- Foreign keys/cascade behavior are not currently enforced.

**Acceptance Criteria:**

- Database initializes cleanly.
- Required MVP tables are created automatically or through documented migration commands.
- Relationships between users, batches, clips, and storylines are enforced or intentionally documented.
- Deleting a parent batch safely handles related child records.
- Any broader source asset/processing schema work is split into separate backlog items.

---

### CS-105: Postgres Connection Pooling and Health-Check Handshake

**Status:** Backlog  
**Priority:** P1

**Description:**  
Implement resilient database connection handling so the Go API does not crash if Postgres takes longer to boot inside the Docker network.

**Current repo evidence:**

- `/api/health` reports database and Redis checks.
- Go tests cover degraded dependency health JSON.
- Startup retry/backoff is not currently implemented.

**Acceptance Criteria:**

- API retries Postgres connection during startup with bounded backoff.
- API exposes DB health through `/api/health`.
- API does not crash immediately because Postgres starts late.
- Failed DB connection eventually returns a clear health failure.

---

### CS-107: Local JWT Auth and Protected Batch Routes

**Status:** Ready  
**Priority:** P1

**Description:**  
Track the existing local JWT auth capability as an explicit MVP backlog item.

**Current repo evidence:**

- Go API exposes `/api/auth/register` and `/api/auth/login`.
- Protected batch routes use JWT middleware.
- Frontend has login/register form and stores token in localStorage.

**Acceptance Criteria:**

- Register returns a token.
- Login returns a token.
- Protected routes reject missing or invalid tokens.
- Protected routes accept a valid token.
- Runtime auth flow is verified against the running Docker stack.

---

### CS-108: Local Browser CORS Policy

**Status:** Backlog  
**Priority:** P1

**Description:**  
Maintain the local development CORS policy that allows the web app to call the Go API.

**Current repo evidence:**

- `apps/api/main.go` allows `http://localhost:3000` and `http://127.0.0.1:3000`.
- Go tests cover allowed and rejected preflight behavior.
- `.env.example` documents `CORS_ALLOWED_ORIGINS`.

**Acceptance Criteria:**

- Local browser requests from the web app are allowed.
- Auth headers are allowed.
- Upload requests are allowed.
- Unexpected origins are rejected.

---

### CS-109: JWT Secret Configuration and Production Guardrails

**Status:** Backlog  
**Priority:** P1

**Description:**  
Document and guard local JWT defaults so `dev-secret` is not mistaken for a production-safe value.

**Current repo evidence:**

- API defaults to `dev-secret`.
- `.env.example` uses `JWT_SECRET=replace-me-for-local-dev`.
- Docker Compose does not set `JWT_SECRET`, so local Compose would use the dev default.

**Acceptance Criteria:**

- Docs clearly state `dev-secret` is local-only.
- Shared or deployed environments require explicit `JWT_SECRET`.
- The API logs or fails clearly if production mode is later introduced without a safe secret.

---

### CS-110: Worker Docker Packaging Fix

**Status:** Backlog  
**Priority:** P0

**Description:**  
Fix the Python worker Docker image so it includes every local module imported by `main.py`.

**Current repo evidence:**

- `apps/api/ai_worker/main.py` imports `zip_safety`.
- `apps/api/ai_worker/Dockerfile` copies `main.py` but not `zip_safety.py`.

**Acceptance Criteria:**

- Worker Docker image includes `main.py` and `zip_safety.py`.
- Worker container starts without `ModuleNotFoundError`.
- Docker build remains reproducible.

---

### CS-113: Frontend Dependency Vulnerability Cleanup

**Status:** Backlog  
**Priority:** P1

**Description:**  
Resolve or explicitly triage known frontend dependency vulnerabilities.

**Current repo evidence:**

- `npm audit --audit-level=moderate` in `apps/web` reports 8 vulnerabilities.
- Findings include critical advisories against `next@14.1.4`.

**Acceptance Criteria:**

- Upgrade plan is reviewed.
- Vulnerable dependencies are upgraded where safe.
- `npm ci`, `npm run build`, and `npm audit` are rerun.
- Any remaining vulnerabilities are documented with accepted risk.

---

### CS-114: Local Docker Runtime Repair and Verification

**Status:** Backlog  
**Priority:** P0

**Description:**  
Restore local Docker/Compose runtime so the MVP stack can be verified.

**Current repo evidence:**

- `docker` and `docker compose` are not found on `PATH` in the current shell.
- Compose runtime verification is blocked.

**Acceptance Criteria:**

- `docker --version` works from repo root.
- `docker compose version` works from repo root.
- `docker info` connects to the daemon.
- `docker compose config` validates.
- `docker compose up --build` boots the stack.

---

### CS-115: CI Build/Test Validation

**Status:** Backlog  
**Priority:** P1

**Description:**  
Keep CI aligned with actual repo commands and dependency boundaries.

**Current repo evidence:**

- `.github/workflows/ci.yml` runs Go API tests, Python worker checks, web build, and Compose config.
- Local Docker validation is blocked, but the commands are real.

**Acceptance Criteria:**

- CI runs on pull requests and pushes to `main`.
- API job runs `go mod download` and `go test ./...` from `apps/api`.
- Worker job runs compile and tests from `apps/api/ai_worker`.
- Web job runs `npm ci` and `npm run build` from `apps/web`.
- Compose job validates `docker compose config`.

---

### CS-116: Replace Legacy SQLite Backup/Restore Scripts

**Status:** Backlog  
**Priority:** P2

**Description:**  
Replace or clearly mark legacy SQLite backup/restore scripts now that the Docker MVP is Postgres-first.

**Current repo evidence:**

- `scripts/database/backup.sh` and `scripts/database/restore.sh` copy `data/clipsense.db`.
- Docker Compose uses Postgres.

**Acceptance Criteria:**

- Scripts either support Postgres or are clearly labeled legacy/out-of-scope.
- Docs do not instruct developers to use SQLite scripts for the active Docker MVP.

---

### CS-117: Stabilization Documentation Refresh

**Status:** Backlog  
**Priority:** P1

**Description:**  
Keep stabilization docs synchronized with current verification results and blockers.

**Current repo evidence:**

- `docs/MVP_STABILIZATION.md` contains stale Go/go.sum blocker language.
- Current inspection shows Go is available, `go.sum` exists, and Go tests pass.

**Acceptance Criteria:**

- Docs reflect current Go status.
- Docs reflect current Docker blocker.
- Docs distinguish ZIP MVP from future direct video/link intake.
- Docs do not overclaim runtime verification.

---

### CS-119: Python Dependency Audit Path

**Status:** Backlog  
**Priority:** P2

**Description:**  
Establish a reliable Python dependency audit path for the worker without faking lockfiles or audit results.

**Current repo evidence:**

- Worker dependencies are listed in `apps/api/ai_worker/requirements.txt`.
- Previous `pip-audit` attempts failed around `openai-whisper` metadata/build behavior.

**Acceptance Criteria:**

- Python worker dependencies can be audited with a documented command, or the current blocker is documented precisely.
- Worker dependencies remain installable.
- Worker compile/tests still pass.
- No fake dependency lockfile is introduced.

---

## Epic 2: File Ingestion and Local Media Preparation

### CS-103: Go API Multipart ZIP Intake

**Status:** Backlog  
**Priority:** P0

**Description:**  
Build and harden the Go API upload endpoint for local multipart ZIP uploads in the current MVP. Save the uploaded binary to local storage and create a batch record in Postgres.

**Current repo evidence:**

- `createBatch` accepts multipart field `file`, saves it under `UPLOAD_DIR`, inserts a `pending` batch, and enqueues Redis.
- The API does not enforce a true request body size cap.
- The API does not validate ZIP file type/signature before enqueueing.

**Acceptance Criteria:**

- API accepts MVP ZIP uploads.
- Oversized files are rejected safely using a true request body cap.
- Non-ZIP files are rejected before enqueueing.
- Upload creates a batch record.
- Upload returns a structured batch ID response.
- Upload failure does not leave orphaned records.
- Client errors return appropriate 4xx status codes.

---

### CS-104: Python Worker Safe Extraction and Audio Prep Pipeline

**Status:** Backlog  
**Priority:** P0

**Description:**  
Implement a Python worker process that receives uploaded archives, safely extracts supported video files, filters junk files, and prepares audio for transcription.

**Current repo evidence:**

- `zip_safety.py` safely extracts ZIP members and has passing tests.
- Worker calls ffmpeg to create 16kHz mono WAV audio.
- Worker Docker packaging is currently bugged because `zip_safety.py` is not copied.

**Acceptance Criteria:**

- ZIP extraction is safe.
- Non-video files are ignored.
- Corrupted videos are isolated without crashing the worker loop.
- Audio is normalized into a transcription-ready format.
- Worker logs clear processing status.
- Worker Docker container can import all local modules.

---

### CS-106: Disk-Space Janitor and TTL Cleanup Daemon

**Status:** Backlog  
**Priority:** P2

**Description:**  
Add a scheduled cleanup process that removes expired raw uploads, extracted video files, temporary audio files, and processing workspace artifacts after successful sync or terminal failure.

**Acceptance Criteria:**

- Temporary extraction folders are cleaned after successful processing.
- Failed batches keep only required diagnostic artifacts.
- Cleanup runs on a predictable interval.
- Cleanup cannot delete active in-progress files.
- Disk usage does not grow forever during repeated uploads.

---

### CS-111: Upload Size and ZIP Validation Hardening

**Status:** Backlog  
**Priority:** P0

**Description:**  
Harden the ZIP upload endpoint with a true request body size limit and ZIP validation.

**Current repo evidence:**

- `createBatch` uses `ParseMultipartForm(200 << 20)` but does not wrap the request with `http.MaxBytesReader`.
- Uploaded files are saved with `.zip` suffix regardless of content.

**Acceptance Criteria:**

- API enforces a clear maximum upload size.
- Oversized uploads return a safe error.
- ZIP extension and/or signature validation runs before persistence/enqueue.
- Existing valid ZIP uploads still work.

---

### CS-112: API Error Response Semantics

**Status:** Backlog  
**Priority:** P1

**Description:**  
Return accurate HTTP status codes for validation, auth, conflict, upload, and server errors.

**Current repo evidence:**

- `httpError` always returns HTTP 500.
- Invalid credentials and missing upload file currently flow through `httpError`.

**Acceptance Criteria:**

- Bad JSON/input returns 400.
- Missing/invalid auth returns 401.
- Duplicate/conflict cases return 409 where applicable.
- Oversized uploads return 413.
- Server faults still return 500.
- Error bodies are safe for users.

---

# Phase 2: Queue Processing and AI Pipeline

**Focus:** background jobs, worker state, transcription, classification, vector upsert, and failure visibility.

## Epic 3: Queue, State Machine, and Failure Handling

### CS-201: Go API Redis Task Publisher and Job State Machine

**Status:** Backlog  
**Priority:** P0

**Description:**  
Wire Redis as the asynchronous task broker. When the API receives a valid upload, it creates a Redis job for the worker and tracks batch status in Postgres.

**Current repo evidence:**

- API enqueues a JSON payload on `jobs:batch`.
- Worker consumes `jobs:batch`.
- Source statuses include `pending`, `processing`, `complete`, and `failed`.
- Backlog status naming should align with code or code should be migrated consistently.

**Acceptance Criteria:**

- API enqueues a JSON job payload.
- API request returns without waiting for full processing.
- Worker can consume the queued job.
- Batch status moves through `pending`, `processing`, `complete`, or `failed`.
- Crashes do not leave batches permanently stuck.
- Runtime queue behavior is verified under Docker Compose.

---

### CS-202: Go API Progress Broadcaster

**Status:** Backlog  
**Priority:** P2

**Description:**  
Add a progress update channel so the frontend can receive live processing status from the API and worker pipeline.

**Acceptance Criteria:**

- Processing stages can be reported to the frontend.
- Status events include stages like `extracting_zip`, `transcribing_audio`, and `analyzing_clips`.
- Authenticated users only receive updates for their own batches.
- Progress updates do not break batch processing if disconnected.

---

### CS-206: Poison Pill Queue Catch-All and Dead Letter Queue

**Status:** Backlog  
**Priority:** P1

**Description:**  
Implement a Redis dead-letter or terminal-failure strategy for jobs that repeatedly fail because of corrupted media, worker crashes, FFmpeg errors, Whisper errors, invalid ZIPs, or unrecoverable processing exceptions.

**Current repo evidence:**

- Worker catches processing exceptions and attempts to mark a batch `failed`.
- Redis queue currently uses a simple list, not an acknowledged retry/dead-letter queue.

**Acceptance Criteria:**

- Worker catches fatal job errors.
- Failed job is marked `failed` in Postgres.
- Safe failure reason is persisted.
- Broken job is moved to a dead-letter queue or terminal failed state.
- Worker continues processing the next job.
- One corrupted upload cannot trap the worker in an infinite retry loop.

---

### CS-207: Persist and Display Batch Failure Reasons

**Status:** Backlog  
**Priority:** P1

**Description:**  
Persist safe batch failure reasons and expose them through the API/web UI.

**Current repo evidence:**

- Worker logs failure reasons.
- Worker marks failed batches as `failed`.
- Schema/API/web do not currently persist or display detailed failure reasons.

**Acceptance Criteria:**

- Failed batches store a safe failure reason.
- API batch detail returns the failure reason.
- Web batch detail displays the reason without raw stack traces.
- Tests cover a worker failure case.

---

### CS-208: End-to-End ZIP MVP Verification Gate

**Status:** Backlog  
**Priority:** P0

**Description:**  
Run and record the real local ZIP MVP runtime verification.

**Execution note:** Docker runtime may need repair when this card is pulled.

**Acceptance Criteria:**

- Docker Compose boots all services.
- API health returns valid JSON with DB and Redis health.
- Web loads at `http://localhost:3000`.
- Register/login works.
- ZIP upload creates a pending batch and Redis job.
- Worker consumes the job.
- Worker safely extracts videos.
- Batch becomes `complete` or `failed` with a visible reason.
- Successful run writes clips/storyline to Postgres.
- Qdrant upsert succeeds or failure is documented.

---

## Epic 4: Machine Learning Pipeline

### CS-203: Python Worker Transcription Integration

**Status:** Backlog  
**Priority:** P1

**Description:**  
Add transcription logic using Whisper or a compatible transcription model. Convert prepared audio into transcript output. Timestamped transcript segments are a backlog extension beyond the current single transcript string.

**Current repo evidence:**

- Worker loads Whisper and calls `model.transcribe()`.
- API schema stores one `transcript` string per clip.
- Timestamped transcript segment persistence does not exist.

**Acceptance Criteria:**

- Worker produces transcript text for supported media.
- Transcription failure marks the batch failed with a safe reason or isolates the failed clip according to policy.
- Transcript data is persisted or passed forward reliably.
- Timestamped segment support is tracked separately if not included in current MVP.

---

### CS-204: Qdrant Vector Upsert Pipeline

**Status:** Backlog  
**Priority:** P1

**Description:**  
Embed transcript or clip text with a local embedding model and upsert vectors into Qdrant for semantic search and similarity workflows.

**Current repo evidence:**

- Worker uses SentenceTransformer embeddings.
- Worker creates/recreates a Qdrant collection and upserts points.
- Runtime Qdrant behavior has not been verified.

**Acceptance Criteria:**

- Worker creates embeddings from transcript or clip text.
- Qdrant collection is created or reused safely.
- Vectors are upserted with useful payload metadata.
- Qdrant failure is logged and handled clearly.

---

### CS-205: Basic MVP Clip Classification and Taxonomy Foundation

**Status:** Backlog  
**Priority:** P1

**Description:**  
Classify clips using an MVP taxonomy such as mood, topic, and clip role. Keep current classification expectations honest: the repo currently uses basic heuristics, not a robust AI taxonomy.

**Current repo evidence:**

- Worker has `classify_mood()` and `classify_role()`.
- Topic is currently hardcoded to `general`.

**Acceptance Criteria:**

- Output includes mood, topic, and clip role.
- Classification result is persisted with the clip.
- Invalid classifier output is rejected or normalized if a structured classifier is added later.
- Advanced classifier work is not required before ZIP MVP runtime verification.

---

# Phase 3: Creator Workspace and Review UI

**Focus:** build and stabilize the creator-facing interface where users can view uploads, inspect clips, see processing errors, and review storyline sequences.

## Epic 5: Next.js Workspace Views

### CS-301: MVP Creator Dashboard and Batch Review Views

**Status:** Backlog  
**Priority:** P0

**Description:**  
Build the creator dashboard and batch review surfaces for current ZIP MVP usage. Multi-pane/timeline workspace is future scope after the current route/auth bugs are fixed.

**Current repo evidence:**

- Web app has dashboard, batch list, upload, settings, and batch detail files.
- Links and redirects use `/dashboard/...`, but `(dashboard)` is a route group and does not create `/dashboard` URLs.
- Build output exposes `/`, `/batches/new`, `/batches/[id]`, and `/settings`.

**Acceptance Criteria:**

- Dashboard loads without broken routing.
- User can navigate to upload, batch list, batch detail, and settings.
- Batch detail displays clips and storylines returned by the API.
- Layout is responsive enough for MVP usage.

---

### CS-302: Native Video Segment Player

**Status:** Icebox  
**Priority:** P3

**Description:**  
Add a video player component that can seek to selected clip timestamps and preview semantic segments.

**Acceptance Criteria:**

- Video player loads selected source video.
- Clicking a transcript or clip segment seeks to the correct timestamp.
- Playback does not restart unnecessarily.
- Missing video assets show a safe fallback state.

---

### CS-303: Timeline Workspace State Store

**Status:** Icebox  
**Priority:** P3

**Description:**  
Add frontend state management for arranging clips into a sequence or timeline workspace.

**Acceptance Criteria:**

- User can arrange clips into a storyline order.
- Sequence state is represented as structured JSON.
- Timeline order can be saved or exported.
- Refresh behavior is documented or persisted.

---

### CS-304: Frontend Network Interceptor and Global Error Toast System

**Status:** Backlog  
**Priority:** P1

**Description:**  
Add a global API error handling layer in the Next.js frontend so API failures, auth failures, upload failures, worker failures, and progress disconnects show clear user-facing messages instead of infinite loading states.

**Current repo evidence:**

- `apps/web/src/lib/api/client.ts` wraps fetch and throws on non-OK responses.
- Upload page uses localized `react-hot-toast`.
- There is no global error/toast strategy.

**Acceptance Criteria:**

- API failures trigger visible error toasts.
- Upload failures show clear reasons.
- Auth failures redirect or notify correctly.
- Long-running batch errors do not leave the UI stuck.
- Network disconnects show a recoverable error state.
- Error messages are safe and do not expose secrets.

---

### CS-305: Fix Dashboard Route Structure

**Status:** Backlog  
**Priority:** P0

**Description:**  
Fix frontend links and redirects so they point to real Next.js routes.

**Current repo evidence:**

- Source links use `/dashboard/...`.
- Build output does not include `/dashboard`.

**Acceptance Criteria:**

- Dashboard, upload, settings, and batch detail URLs resolve.
- Auth redirects land on real routes.
- `npm run build` still passes.

---

### CS-306: Stabilize Frontend Auth Hydration

**Status:** Backlog  
**Priority:** P0

**Description:**  
Prevent protected pages from redirecting before the localStorage token has been loaded.

**Current repo evidence:**

- `useAuthToken` initializes token as `null` and loads localStorage in an effect.
- Protected pages redirect immediately when token is null.

**Acceptance Criteria:**

- Existing logged-in users can refresh protected pages without false redirect.
- Missing-token users still redirect to login.
- Auth loading state is explicit.

---

# Phase 4: Scale, Remote Ingestion, and Export

**Focus:** move beyond local ZIP MVP into remote intake, scalable storage, stronger story intelligence, performance acceleration, and editor-ready exports.

## Epic 6: Scaled Video Ingestion

### CS-401: Remote Video Link Ingestion Worker

**Status:** Icebox  
**Priority:** P3

**Description:**  
Extend the worker pipeline to process supported public video links using a controlled ingestion tool such as `yt-dlp`.

**Acceptance Criteria:**

- API accepts a supported link source.
- Worker extracts audio/video from the source safely.
- Unsupported links fail clearly.
- Link intake normalizes into the same internal pipeline as uploads.

---

### CS-402: Cloud Multipart Upload Pipeline

**Status:** Icebox  
**Priority:** P3

**Description:**  
Move large file upload handling to object storage using presigned multipart uploads through S3 or a compatible storage layer.

**Acceptance Criteria:**

- Client can request presigned upload metadata.
- Large files upload directly to object storage.
- API records source asset metadata after upload completion.
- Local disk upload path is no longer required for production-scale files.

---

## Epic 7: Intelligent Story Optimization

### CS-403: Graph-Based Narrative Connector

**Status:** Icebox  
**Priority:** P3

**Description:**  
Build a graph-based storyline engine where clips become nodes and relationships are scored using semantic similarity, mood pacing, taxonomy adjacency, and story structure.

**Acceptance Criteria:**

- Clips are represented as graph nodes.
- Edge weights reflect narrative compatibility.
- System can generate suggested story paths.
- Output improves beyond basic clustering/order logic.

---

## Epic 8: Performance and Export Systems

### CS-404: Rust Video Core for Scene Boundary Detection

**Status:** Icebox  
**Priority:** P3

**Description:**  
Add a Rust-powered video processing core for faster scene boundary detection and visual cut analysis.

**Acceptance Criteria:**

- Rust module can process video scenes faster than Python baseline.
- Scene boundary output is passed back to the worker pipeline.
- Integration does not break existing Python processing.
- Performance improvement is measurable.

---

### CS-405: MVP Batch Export and Future Narrative Package Export

**Status:** Backlog  
**Priority:** P2

**Description:**  
Maintain the current MVP batch JSON/CSV export path and split advanced editor-ready narrative export into future scope.

**Current repo evidence:**

- API supports JSON and CSV export from batch detail.
- Frontend CSV export uses a plain anchor to an authenticated route and cannot attach the bearer token.

**Acceptance Criteria:**

- MVP export works with authentication.
- Export includes batch, clips, and storyline order where available.
- Advanced editor-ready export formats are deferred until timeline/workspace features exist.

---

# Stabilized ZIP MVP Done Gate

The stabilized ClipSense ZIP MVP is done when:

- Docker Compose boots the required services.
- Go API tests pass.
- Python worker tests pass.
- Frontend build passes.
- Web app loads locally.
- User can register and log in.
- User can upload a ZIP file.
- API creates a batch.
- Redis queues a worker job.
- Worker consumes the job.
- Worker safely extracts supported videos.
- Worker handles corrupted videos without crashing the processing loop.
- Worker transcribes or processes the media.
- Worker writes clips/storyline data to Postgres.
- Worker persists clear failure reasons.
- Temporary processing files are cleaned safely or a cleanup limitation is documented.
- Qdrant upsert succeeds or fails clearly.
- Batch detail page shows real results or a clear failure reason.
- Global/frontend error handling prevents infinite loading states for core MVP flows.
- Security risks are triaged.
- Docs match what actually works.

# Broader Product Done Gate

Broader product MVP is done when:

- ZIP upload works.
- Direct video upload works.
- Link intake works for at least one constrained supported source type.
- All intake types normalize into the same internal asset pipeline.
- Analysis output is visible and useful.
- Basic storyline output is understandable.
- Creator can review results in the UI.
- Export or handoff format exists at MVP quality.

# Product Interpretation

ClipSense is an **AI-powered video repurposing and creator workflow platform**.

Its core value is helping creators turn raw videos into structured content assets:

```text
raw video
-> transcript
-> semantic clips
-> context/risk/highlights
-> storyline
-> creator review/export
```
