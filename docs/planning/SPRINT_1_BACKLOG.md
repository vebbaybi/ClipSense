# Sprint 1 Backlog: Runtime Baseline Repair

Sprint 1 goal: prove one small ZIP can move through the current web/API/processor stack
and produce an authenticated, reviewable, exportable result using a clean environment.

## Ordered Backlog

### Work Unit 1: Runtime Startup And Routing Baseline

Status: **Review / QA**

Source implementation and available local validation are complete. The exact-commit
GitHub Actions Docker runtime gate must pass before this work unit can move to Done.

- Worker image packages `main.py` and `zip_safety.py`, excludes local artifacts,
  installs FFmpeg, and performs compilation/import/FFmpeg build smoke checks.
- Canonical public routes are `/`, `/dashboard`,
  `/dashboard/batches/new`, `/dashboard/batches/{id}`, and
  `/dashboard/settings`.
- API liveness, readiness, bounded dependency checks, explicit server timeouts,
  signal handling, and graceful shutdown logic are implemented and unit tested.
- Compose health-gates Postgres, Redis, Qdrant, API, worker, and web startup.
- GitHub-hosted Linux is the authoritative Docker verification environment. Windows
  Docker Desktop validation is optional supplemental QA.

Evidence: `docs/audits/SPRINT_1_WORK_UNIT_1_REPORT.md`.

### Work Unit 2: Security And Upload Hardening

Status: **Ready / Sprint**, but should begin only after the Work Unit 1 GitHub
Actions runtime gate passes.

Known inputs include JWT secret enforcement, upload hard limits, authenticated
export repair, safe API error responses, and the other explicitly approved Work
Unit 2 items. Work Unit 1 does not implement them.

| ID | Value and exact scope | Out of scope | Dependencies | Acceptance and validation | Risk | Estimate | Owner | Status |
|---|---|---|---|---|---|---:|---|---|
| CS-017 | Repair worker image by packaging all imported source; verify import/startup | Model upgrade | Docker Engine | Image builds; worker imports; startup reaches dependency connection | Large ML image/download | 2 | Processor | Ready |
| CS-022 | Repair dashboard route contract and all internal links/redirects | UI redesign | Web build | One non-colliding route strategy; navigation tests/build | App Router behavior | 3 | Web | Ready |
| CS-032 | Download CSV through authenticated client adapter | New formats | Route/auth baseline | Authorized succeeds; unauthenticated fails; browser test | Token exposure | 2 | Web/API | Ready |
| CS-015 | Enforce total HTTP body limit, ZIP type/signature, partial-file cleanup | Direct video | API tests | Oversize returns 413; invalid type 4xx; disk remains clean | Proxy/body semantics | 3 | API | Ready |
| CS-030 | Require non-development JWT secret outside explicit local mode | OIDC migration | Configuration design | Startup fails securely; Compose uses non-default secret; tests | Local UX break | 2 | API/Ops | Ready |
| CS-028 | Replace raw internal errors with stable safe JSON errors | Full contract generation | Error taxonomy | Validation/auth/conflict/server tests; logs retain cause | Client compatibility | 3 | API/Web | Ready |
| CS-029 | Add liveness/readiness split, HTTP timeouts, graceful shutdown, request IDs | Full telemetry | API refactor | Unit tests; SIGTERM test; readiness covers required dependencies | Startup sequencing | 5 | API/Ops | Ready |
| CS-043 | Add Compose health checks/dependency conditions and repair deploy path | Production deployment | Prior service health work | `compose config`; clean HP startup; documented results | Docker unavailable on current machine | 5 | Ops | Ready |
| CS-047 | Replace SQLite copy scripts with Postgres backup/restore; document Qdrant gap | Automated disaster recovery | Compose database | Disposable backup/restore proves row recovery | Data loss if misused | 5 | Ops/Data | Ready |
| CS-038 | Upgrade Next.js/React to a supported security line in an isolated change | UI features | Route tests | Clean install/build; browser smoke; audit reviewed | Breaking framework changes | 5 | Web | Ready |
| CS-037 | Add synthetic tiny-media Compose smoke test for register/upload/process/review/export | Performance/load testing | All preceding repairs, Docker, FFmpeg | Repeatable test passes twice from clean state | Model runtime/time | 8 | QA/All | Ready |
| CS-035 | Make processing stages idempotent and define partial-failure cleanup for smoke path | Temporal adoption | Migration/job contract decision | Requeued job creates no duplicate clips/storylines; failure test | Cross-store consistency | 8 | Processor/API | Needs Refinement |

## Sprint Guardrails

- Runtime changes are separate from the Sprint 0 documentation commit.
- Do not introduce desktop, link/direct-video intake, advanced AI, collaboration,
  editor integrations, Kubernetes, billing, or Temporal.
- Approve the migration baseline before changing schema.
- Use GitHub-hosted Linux for authoritative Docker/FFmpeg integration. The HP and Acer
  may provide supplemental Windows Docker evidence. Mac browser validation is useful
  after routes and framework upgrade, but is not a substitute for the Docker gate.
