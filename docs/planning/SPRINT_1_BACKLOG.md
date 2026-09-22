# Sprint 1 Backlog: Runtime Baseline Repair

Sprint 1 goal: prove one small ZIP can move through the current web/API/processor stack
and produce an authenticated, reviewable, exportable result using a clean environment.

## Ordered Backlog

### Work Unit 1: Runtime Startup And Routing Baseline

Status: **Done for the tested runtime scope** on
`ce9d9929dfec43639a0e1a57d72c0ee10352eb66`, accepted after source review and
[CI/Docker run 35644022417](https://github.com/vebbaybi/ClipSense/actions/runs/35644022417).
PR #84 merged as `4f073e5fa6f10beccd8bd8ad6e289ec9b76b373b`. Final main
documentation reconciliation receives the same CI before final sign-off; see the
PR follow-up for its exact SHA/run.

Historical runtime run `30407483794` passed at
`94f1911ec73ef4993ad90a9d81a4f0aece8968a8`. It does not certify the current
integration with main. The fresh run above verifies the integrated candidate,
including API restart after graceful shutdown. See
`docs/audits/GATE_1_RUNTIME_ACCEPTANCE.md` for conflict decisions and results.

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

Status: **Review / QA - Gates 2 and 3 implemented; exact-candidate hosted acceptance required**.
User authorization covers authentication and upload hardening only. See
[Work Unit 2 evidence](../audits/WORK_UNIT_2_AUTH_UPLOAD.md) and its candidate PR
for exact tested commit, CI/Docker runs, commands and final gate decisions.

Intermediate candidate `156c6ea2d04d9c50a6b841bc3cce9c305049387c` passed all
five jobs in [run 35649979087](https://github.com/vebbaybi/ClipSense/actions/runs/35649979087),
including the preserved runtime baseline and new HTTP/media boundary checks.
Final follow-up adds ambiguous database-insert cleanup and an image-level missing-key
check. Its exact-SHA result and acceptance decision are recorded on
[PR #85](https://github.com/vebbaybi/ClipSense/pull/85); the intermediate run does not
certify that follow-up. CS-015 still needs shared UI limit presentation; CS-030
still includes production database/DSN guardrails. Neither parent story is closed.

Implemented scope: explicit signing-key configuration, strict credentials/tokens,
bounded auth attempts, request/archive/media validation, scoped cleanup and safe
auth/upload errors. Authenticated export, cross-user/CORS acceptance, processing
reliability, broad dependency modernization and backup/restore remain separate.

| ID | Value and exact scope | Out of scope | Dependencies | Acceptance and validation | Risk | Estimate | Owner | Status |
|---|---|---|---|---|---|---:|---|---|
| CS-017 | Runtime subset: repair worker packaging/import/startup | Job payload validation and actual enqueue/consume | Docker Engine | Image/import/startup and Redis recovery passed in run 35644022417 | Large ML image/download | 2 | Processor | Runtime subset accepted; parent Review / QA |
| CS-022 | Runtime subset: repair dashboard route contract | Full result review and authenticated browser E2E | Web build | Build and five HTTP route checks passed in run 35644022417; links reviewed | App Router behavior | 3 | Web | Routing subset accepted; parent Review / QA |
| CS-032 | Download CSV through authenticated client adapter | New formats | Route/auth baseline | Authorized succeeds; unauthenticated fails; browser test | Token exposure | 2 | Web/API | Ready |
| CS-015 | Enforce total HTTP body limit, ZIP type/signature, partial-file cleanup | Direct video | API tests | Focused tests pass; exact-candidate hosted HTTP/media verification required | Proxy/body semantics | 3 | API | Review / QA; Gate 3 PR evidence required |
| CS-030 | Require non-development JWT secret outside explicit local mode | OIDC migration | Configuration design | Config/token/limiter tests pass; exact-candidate hosted auth verification required | Local UX break | 2 | API/Ops | Review / QA; Gate 2 PR evidence required |
| CS-028 | Replace raw internal errors with stable safe JSON errors | Full contract generation | Error taxonomy | Auth/upload errors covered; other endpoint error contracts remain open | Client compatibility | 3 | API/Web | Auth/upload subset Review / QA; broader scope open |
| CS-029 | Add liveness/readiness split, HTTP timeouts, graceful shutdown, request IDs | Full telemetry | API refactor | Unit tests, health, Redis recovery, SIGTERM/restart passed in run 35644022417; request IDs remain | Startup sequencing | 5 | API/Ops | Runtime subset accepted; request IDs open |
| CS-043 | Add Compose health checks/dependency conditions and repair deploy path | Production deployment | Prior service health work | Compose config and hosted Linux startup passed in run 35644022417; deploy-path repair remains | Local Docker unavailable; hosted CI authoritative | 5 | Ops | Compose subset accepted; deployment repair open |
| CS-047 | Replace SQLite copy scripts with Postgres backup/restore; document Qdrant gap | Automated disaster recovery | Compose database | Disposable backup/restore proves row recovery | Data loss if misused | 5 | Ops/Data | Ready |
| CS-038 | Upgrade Next.js/React to a supported security line in an isolated change | UI features | Route tests | Clean install/build; browser smoke; audit reviewed | Breaking framework changes | 5 | Web | Ready |
| CS-037 | Add synthetic tiny-media Compose smoke test for register/upload/process/review/export | Performance/load testing | All preceding repairs, Docker, FFmpeg | Repeatable test passes twice from clean state | Model runtime/time | 8 | QA/All | Ready |
| CS-035 | Make processing stages idempotent and define partial-failure cleanup for smoke path | Temporal adoption | Migration/job contract decision | Requeued job creates no duplicate clips/storylines; failure test | Cross-store consistency | 8 | Processor/API | Needs Refinement |

### Work Unit 2.5: Bounded DevSecOps And Kaufman Observability

Baseline: merged PR #85, `99cb659243113623b42e6ae6d713977259cbb098`, main CI
`35723984517` passed. Gates 1-3 remain the regression contract; Gate 4 not started.
Status: Review / QA, pending exact-candidate hosted runtime/security evidence.
CS-048 (#48) subset: structured redacted events, cross-job correlation, bounded
metrics, local Prometheus/Grafana, rules/runbook and focused telemetry tests.
Broader analytics, alert delivery, full tracing and incident acceptance remain open.
Related security/testing work: source/dependency/image scans, SBOM and explicit
polyglot test architecture. Scan execution does not grant release approval.
See `docs/audits/WORK_UNIT_2_5_TRUTH.md` and the candidate PR for tested SHA/run,
scan artifacts, validation outcomes and final PASS/CONDITIONAL/FAIL decision.
No parent story or later gate is marked Done by this implementation.

## Sprint Guardrails

- IDs above refer to broader live GitHub stories, not independently completed
  stories. Accepted runtime subsets do not close their remaining criteria. All
  other rows retain their planning status; none are accepted by Gate 1.

- Runtime changes are separate from the Sprint 0 documentation commit.
- Do not introduce desktop, link/direct-video intake, advanced AI, collaboration,
  editor integrations, Kubernetes, billing, or Temporal.
- Approve the migration baseline before changing schema.
- Use GitHub-hosted Linux for authoritative Docker/FFmpeg integration. The HP and Acer
  may provide supplemental Windows Docker evidence. Mac browser validation is useful
  after routes and framework upgrade, but is not a substitute for the Docker gate.
