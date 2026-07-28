# Sprint 1 Work Unit 1 Report

## 1. Scope

Runtime startup and routing baseline only: worker image packaging, canonical web
routes, API liveness/readiness, Go server lifecycle, startup configuration
inventory, Compose health behavior, validation, documentation, and Kanban.
Sprint 1 Work Unit 2 security and upload hardening is excluded.

## 2. Starting Branch

`backlog-repair-refinement`

## 3. Starting Commit

`49596612fae8cc1bc20675b722efdba107647f40`

## 4. Worktree Condition

The starting branch was six commits ahead of its upstream and contained seven
untracked user files. They were not modified, moved, staged, or committed. Work
continued on `codex/clipsense-sprint-1-runtime-baseline` from the exact starting
commit. The workspace is under OneDrive.

## 5. Worker Packaging Findings

`main.py` imports the local `zip_safety.py`, but the old Dockerfile copied only
`main.py`. The image now copies both runtime modules and uses a worker-specific
`.dockerignore` to exclude virtual environments, caches, tests, models, media,
data, secrets, and temporary files. The build checks Python compilation, module
imports, and `ffmpeg -version`; the command remains `python main.py`.

Runtime initialization was moved behind `initialize_runtime()` so importing the
module does not download models. Actual worker startup still loads Whisper model
`small` and sentence-transformer `all-MiniLM-L6-v2`; their libraries use the
container user's standard caches (normally under `/root/.cache`), and a first
startup blocks while downloads complete.

## 6. Route Findings

The former `(dashboard)` group created no URL segment and collided with `/`.
Its five source files were moved beneath the real `src/app/dashboard` segment.
Navigation and redirects now target the canonical dashboard paths. Auth-token
hydration is explicit so client redirects do not execute during prerender.
The Next.js production build reports `/`, `/dashboard`,
`/dashboard/batches/new`, `/dashboard/batches/[id]`, and
`/dashboard/settings` without a duplicate route.

## 7. API Lifecycle Findings

The former dependency-coupled health handler is now split into process liveness
and Postgres/Redis readiness. Dependency checks have a two-second deadline and
responses expose only safe check states. `/api/health` remains a readiness alias.
The server now has explicit timeouts, handles interrupt and termination, performs
a bounded 15-second shutdown, and closes database and Redis resources.

## 8. Configuration Findings

API reads `API_ADDR`, `DATABASE_URL`, `REDIS_URL`, `UPLOAD_DIR`,
`CORS_ALLOWED_ORIGINS`, and `JWT_SECRET`. Worker reads `DB_DRIVER`, `DB_PATH`,
`DATABASE_URL`, `UPLOAD_DIR`, `PROCESS_DIR`, `WHISPER_MODEL`, `REDIS_URL`,
`QDRANT_HOST`, `QDRANT_PORT`, and `QDRANT_COLLECTION`. Web reads
`NEXT_PUBLIC_API_URL`. Compose's API `DB_DRIVER` value is unused; worker
`UPLOAD_DIR` is currently unused. Container dependency addresses and the public
browser API URL intentionally differ. The predictable JWT fallback remains an
explicit Work Unit 2 blocker.

## 9. Compose Findings

The root `docker-compose.yml` is authoritative. Postgres and Redis use native
probes; Qdrant checks its HTTP `/readyz`; API readiness uses
`/api/health/ready`; web checks `/`. API waits for Postgres and Redis, worker
waits for Postgres, Redis, and Qdrant, and web waits for API readiness. No
circular dependency was introduced. Host-port hardening is deferred.

## 10. Changed Files

- `apps/api/ai_worker/Dockerfile`
- `apps/api/ai_worker/main.py`
- `apps/api/main.go`
- `apps/api/main_test.go`
- `apps/web/src/app/page.tsx`
- `apps/web/src/hooks/useAuthToken.ts`
- `docker-compose.yml`
- `scripts/utils/health-check.sh`
- `README.md`
- `docs/CURRENT_STATE.md`
- `docs/planning/SPRINT_1_BACKLOG.md`

## 11. Added Files

- `apps/api/ai_worker/.dockerignore`
- `docs/audits/SPRINT_1_WORK_UNIT_1_REPORT.md`

## 12. Moved Files

Five files moved from `apps/web/src/app/(dashboard)` to the corresponding paths
under `apps/web/src/app/dashboard`: layout, dashboard page, batch intake, dynamic
batch detail, and settings.

## 13. Deleted Files

No functionality was deleted. The five obsolete route-group paths were removed
as part of the moves.

## 14. Added Dependencies

None.

## 15. Removed Dependencies

None.

## 16. Validation Commands

```text
go fmt ./...
gofmt -l .
go vet ./...
go test ./...
python -m py_compile main.py zip_safety.py
python -m unittest discover -s tests
python -c "import main; import zip_safety"
npx tsc --noEmit
npm run build
git diff --check
```

## 17. Validation Results

- Passed: Go formatting check, exit 0.
- Passed: `go vet ./...`, exit 0.
- Passed: all Go tests, exit 0, including health and lifecycle tests.
- Passed: Python compilation, exit 0.
- Passed: seven existing Python tests, exit 0.
- Failed locally: worker import smoke test. Acer's global Python 3.10 environment
  loads an incompatible TensorFlow/protobuf combination through transformers.
  The clean Python 3.11 Docker build is the authoritative import environment.
- Passed: TypeScript check, exit 0.
- Passed: Next.js production build, exit 0, with all five intended routes.
- Not applicable: no frontend test script exists.
- Not applicable: no PowerShell file was changed.
- Passed: Git whitespace check, exit 0.

## 18. Authoritative Docker Validation

GitHub-hosted Ubuntu is now the authoritative Docker environment. The
`Docker Runtime Verification` workflow validates Compose metadata, builds all three
application images, starts the canonical six-service stack, verifies dependency
health, API health semantics, web routes, worker imports/FFmpeg/connectivity/stability,
Redis readiness degradation and recovery, SIGTERM handling, evidence collection, and
clean shutdown. It records the exact commit, ref, runner, toolchain, run ID, attempt,
trigger, timings, image inventory, HTTP results, and service state.

The workflow has been implemented but its exact-commit run evidence is pending. Until
that run passes, all Docker-dependent criteria remain Review / QA. HP and Acer Docker
Desktop evidence is optional supplemental Windows QA.

## 19. Kanban Updates

Project #19 was treated as authoritative. Issues #16, #21, #28, and #41 were
updated for worker, route, health, and Compose work. Issues #79 through #83 were
created for the Work Unit 1 epic, Go lifecycle, configuration inventory, local
evidence, and runtime verification. Issue #83 now owns the GitHub Actions Docker
Runtime Verification Gate and remains Review / QA until the authoritative run passes.
Issue #75 remains Bugged and its unsafe workflow was not run.

## 20. Remaining Blockers

Docker configuration validation, clean image build, worker container import and
stability, in-container FFmpeg, full stack health/reachability, and real signal
shutdown evidence require a successful exact-commit GitHub Actions run. That run must
also confirm the pinned Qdrant image supports its readiness probe. These block Work
Unit 1 from Done and are tracked by issue #83.

The first authoritative run, `30405578943` at
`bd9d1666ab0ef065d6a04dbe37062a0572bc0c20`, passed Compose configuration and
source-quality jobs but exposed a worker image build defect before runtime:
`openai-whisper==20231117` imported `pkg_resources` in an isolated build environment
whose current setuptools no longer supplied it. The worker Dockerfile now pins the
compatible build tool to `setuptools<81` and installs the already-declared
requirements without PEP 517 build isolation. The failed run retained artifact
`docker-runtime-bd9d1666ab0ef065d6a04dbe37062a0572bc0c20`.

The second authoritative run, `30405742248` at
`255feb9574f6ac3e6e4e68cc1c040a370f758937`, proved the Whisper wheel correction,
then failed the worker import smoke because the legacy
`sentence-transformers==2.2.2` declaration allowed incompatible current
`transformers` and `huggingface-hub` releases. The requirements now constrain those
existing transitive libraries to the mutually compatible published versions
`transformers==4.30.2` and `huggingface-hub==0.14.1`. No application capability was
added. The failed run also retained commit-attributed diagnostics.

## 21. Deferred Work Unit 2 Risks

Predictable JWT fallback, total upload-size enforcement, authenticated browser
export, safe API error envelopes, password/email rules, authentication rate
limiting, and development port exposure remain deferred. No Work Unit 2 security
claim is made.

## 22. Done Recommendation

**Review / QA, not Done.** Source implementation and available local validation
are complete, but the Definition of Done requires the authoritative exact-commit
GitHub Actions Docker runtime evidence.
