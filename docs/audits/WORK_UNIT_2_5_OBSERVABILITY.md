# Work Unit 2.5: Kaufman Observability And DevSecOps

Date: 2026-09-22. Scope: bounded foundation, not a product-feature sprint.
Decision: **CONDITIONAL**. Functional foundation has hosted evidence; security debt
and the proposed temporary inventory policy still require owner disposition.
PR: https://github.com/vebbaybi/ClipSense/pull/86 (draft; not merged).

## 1. Baseline

PR #85 was merged before implementation. Accepted main:
`99cb659243113623b42e6ae6d713977259cbb098`, CI `35723984517` passed.
Local main was fast-forwarded to that commit from the older runtime baseline;
the working tree was clean before creating `codex/sprint1-kaufman-observability`.
Work Unit 2 was not recreated. Its candidate was
`0b09c22561c5e8cd14232d889dbb0bf55bba2931`.

The [pre-implementation truth audit](WORK_UNIT_2_5_TRUTH.md) records Go handler,
auth/upload/security tests with sqlmock/miniredis; 12 Python safety/media tests;
web build-only validation; and hosted real-service runtime/security probes.
Reachable Go vulnerability checks blocked CI. npm audit recorded eight findings.
Python/image/SAST/secret/SBOM checks and coherent operational telemetry were missing.

## 2. Misconceptions Corrected

ClipSense did not have "no tests". It had limited but real polyglot unit, security,
integration and runtime coverage. Missing layers include component/browser/a11y
testing, complete creator E2E, migration evolution, durable recovery and AI quality
evaluation. Python is not the testing framework for the whole application.

## 3. Architecture

[Kaufman Observability](../observability/KAUFMAN_OBSERVABILITY.md) is the shared
contract: native Go slog and Python standard logging adapters, allowlisted JSON,
canonical UUID correlation, bounded metrics, private scrape endpoints, optional
Prometheus/Grafana Compose profile and one provisioned dashboard. Logs rotate;
Prometheus retention is bounded. No external browser analytics were added.
OTel boundary/field compatibility is documented; actual spans, Collector, Loki,
Tempo and paging integrations are deferred rather than claimed as implemented.

## 4. Files And Components

- `apps/api/observability.go`, `rejection_metrics.go`, `main.go` and tests: HTTP,
  upload/export/dependency events, IDs, private metrics and safe errors.
- `apps/api/ai_worker/kaufmanlogger.py`, `main.py`, tests and Dockerfile: standard
  logger adapter, real job context, stage timing and safe failure emission.
- Web telemetry module/declarations/tests, Diagnostics component, API client and
  upload page: opt-in safe browser events and request IDs; Node tests/type checks.
- Web Dockerfile and `.dockerignore`: `npm ci`, clean build context and build SHA.
- `docker-compose.yml`, `infra/observability/**`, monitoring setup script: optional
  local collection, provisioned dashboard, rules, explicit Grafana credential.
- CI/security workflows, `.semgrep.yml`, `scripts/ci/*`: scans, positive scanner
  fixtures, policy tests, runtime telemetry probes and retained evidence/SBOM.
- Testing/security/observability/runbook docs, ADR 0013, current state and Sprint 1
  backlog: authoritative contracts, limitations and conditional acceptance.

New runtime libraries: Go Prometheus client 1.23.2 and its module dependencies;
Python prometheus-client 0.22.1. No schema migration, intake or creator feature.

## 5. Logging

API emits lifecycle, HTTP status/duration, auth outcomes, denials, upload, queue,
dependency and export events. Worker emits claim/outcome, validation, FFmpeg,
transcription, embedding, Qdrant and model/connection stages. Safe event/stage/status
categories replace raw exception text. Startup errors have bounded Go error codes.
Browser emits allowlisted upload/rejection/fetch/export/application-error events
only when diagnostics are enabled. Existing plain-anchor export is still a gap;
the telemetry wrapper does not magically make that path authenticated.

## 6. Correlation

One real hosted fixture from run `35729642613`:
`request -> batch 193f3e54-50c0-4bc5-bdb2-f2d29aa2c704 -> jobs:batch -> worker`.
Shared correlation: `4b909981-14fe-471c-94d0-9a6e149557c6`.
The API validated/generated UUIDs, returned headers and enqueued the context;
the worker claimed the real Postgres row, rejected invalid media and emitted the
same context. These IDs identify diagnostics, never authority. This isolated
consumer probe bypasses model initialization and is not full ML/E2E acceptance.

## 7. Redaction

Admission control drops arbitrary fields instead of attempting to scrub arbitrary
payloads. Protected classes: credentials/JWTs/headers/secrets/DSNs, request bodies,
emails, transcripts, media, filenames/paths and raw exception text. Canonical UUID,
enum and numeric validation prevent arbitrary context injection. Tests parse JSON,
reject hostile fields/IDs and inspect bounded metric labels. Hosted sentinel tests
check API/worker telemetry for synthetic password, email, JWT and signing-key leaks.
Third-party library output is not automatically covered by the application adapter;
representative runtime evidence was inspected, not all possible dependency errors.

## 8. Metrics

- `clipsense_events_total`: bounded event counts for auth/upload/enqueue/claim/job
  outcomes, stage failures, cleanup and dependency transitions.
- `clipsense_http_requests_total`, `clipsense_http_duration_seconds`: bounded route
  templates/status classes and latency, never raw paths/IDs.
- `clipsense_rejections_total`, `clipsense_active_uploads`: safe rejection categories
  and currently active upload requests.
- `clipsense_dependency_up`, `clipsense_queue_depth`: DB/Redis availability and list
  depth (-1 means unknown, not zero); no fabricated oldest-job age.
- `clipsense_stage_duration_seconds`, `clipsense_job_duration_seconds`: processing
  timings/outcomes; job duration includes attempts that fail before claim.
- `clipsense_storage_available_bytes`, `clipsense_storage_total_bytes`: processing
  filesystem pressure. Prometheus `up` shows scrape availability, not worker readiness.

Persistence timing currently measures connection acquisition, not every SQL call.
No user/email/file/batch/request/correlation values are Prometheus labels.

## 9. Dashboard

One 12-panel provisioned operational dashboard covers service scrape health,
DB/Redis health, uploads, job outcomes, queue depth, stage/job/API duration,
authentication failures, API response classes, disk and cleanup failures. Hosted
Grafana API retrieval and both Prometheus targets were verified. It is not a
production public monitoring deployment or an end-to-end worker readiness check.

## 10. Alerts

Rules cover target unavailability, DB/Redis outage, observed job/cleanup failures
and exhausted storage. Auth spikes, queue-not-draining and proactive disk thresholds
require baseline measurement before tuning; none are invented production SLOs.
Promtool validates rules. Delivery/escalation ownership is still required for release.

## 11. Security Pipeline

Initial exact-head inventory: `22b15753edc7b721d31134e159d13f8da07668d7`, run
`35729642613`; these counts are point-in-time instances, not unique CVE totals.

| Tool | Observed result | Policy |
|---|---|---|
| Semgrep 1.136.0 | Six local rules, 23 app files, zero findings/errors; positive Go/Python/TS fixtures pass at 7b03485 | Findings/errors block; narrow coverage, not comprehensive taint analysis |
| Gitleaks 8.28.0 | Full history, zero findings | Findings block; exact false-positive review only |
| govulncheck 1.8.0 | Zero reachable vulnerabilities | Reachable findings block |
| npm audit | Eight existing findings recorded | High/critical debt blocks release; temporary inventory integration policy |
| pip-audit 2.9.0 | 79 installed packages; 59 advisory instances | Complete scan required; advisories block release pending triage |
| Trivy 0.74.0 | API 15 HIGH; worker 3 CRITICAL/216 HIGH; web 1 CRITICAL/36 HIGH | Image secrets/errors block now; high/critical debt blocks release |
| Trivy IaC | Three HIGH and three LOW findings | Evidence/review, not an enforced misconfiguration gate |
| CycloneDX | API 42, worker 388, web 636 components | Three nonempty SBOMs retained; no signing/provenance claim |

Critical inventory includes worker torch CVE-2025-32434, transformers CVE-2023-6730,
libxml2 CVE-2026-6653 and web tar CVE-2026-59873. Applicability/reachability still
requires review; do not equate scanner severity with demonstrated exploitability.
All three application images run as default root. No broad CVE suppression added.

The proposed bootstrap inventory window expires 2026-10-06. It is **not approved
risk acceptance**. `security-policy.py --release` blocks known debt immediately;
after expiry it blocks integration too. Missing/skipped/malformed scan evidence
blocks now. Owner disposition of this temporary policy is an acceptance condition.
Main branch protection was absent at audit; workflow failures alone do not prevent
an administrator from merging. Required-check/review configuration remains open.

## 12. Testing Model

| Layer | Current reality |
|---|---|
| Unit | Go handlers/helpers, 15 Python safety/telemetry tests, two Node telemetry tests |
| Integration | sqlmock/miniredis plus hosted real Postgres/Redis/Qdrant services and worker validation boundary |
| Contract | API-to-job context and response boundary tests; no generated/full contract suite |
| Security | Auth/upload/CORS focused tests and hosted Gates 2-3 probes |
| E2E | Runtime probes only; complete creator browser/ML workflow missing |
| Reliability | Redis interruption/recovery and lifecycle probes; durable retry/duplicate/crash policy deferred |
| Performance | Duration instrumentation, not load/performance acceptance |
| AI quality | No deterministic transcript/classification/storyline evaluation suite yet |

See [Testing](../TESTING.md) for Sprint 1 versus later obligations.

## 13. Validation

Local: `go test ./...`, `go vet ./...`,
`go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...`, Python compile/unittest,
`npm test`, `npm run typecheck`, `npm run build`, policy tests and `git diff --check`
passed. Local Docker was unavailable; no local Compose success is claimed.
Go formatting debt remains in three pre-existing test files; no repository-wide
format gate is claimed. YAML/JSON parsing alone was not used as runtime proof.

Hosted: `bash scripts/ci/verify-docker-runtime.sh {config,build,runtime,diagnostics}`,
`promtool check config`, `python scripts/ci/verify-observability.py`, security/media
boundary probes and `bash scripts/ci/scan-images.sh` passed in the initial run.
Scanner positive fixtures and negative policy tests prove detection/enforcement,
not merely that executables start. The final probe additionally checks persisted
failed batch status. Exact final run attribution is recorded in PR #86 evidence.

## 14. CI Evidence

- Main baseline: `35723984517` at `99cb659243113623b42e6ae6d713977259cbb098`.
- Initial complete exact-head CI: `35729642613` at
  `22b15753edc7b721d31134e159d13f8da07668d7`, all five jobs passed.
- Source scanner/positive fixture proof: `35730678657` at
  `7b03485470fd49e0075a90ce1e6c1995514fa60b`, passed.
- Earlier `35730493564` failed due to Semgrep's fixture-runner IndexError. Replaced
  that test harness with the actual production scan entry point; rules unchanged.

Artifacts include source SHA, redacted scans, resolved Python inventory, image IDs,
three SBOMs, Compose/lifecycle/security diagnostics, worker correlation JSONL,
metrics, Prometheus query/targets and provisioned Grafana dashboard. Retention is
14 days; export required evidence before expiry. Final candidate/run is attached to
PR #86 rather than claiming a self-referential commit hash inside this file.

## 15. Runtime Regression

Initial hosted Gates 1-3 regression passed, including invalid authentication,
rejected uploads and Redis outage/recovery with observable logs/metrics. Later
candidate changes require the final exact-head rerun; no Gate 4 execution or
full creator/ML workflow acceptance is implied by these bounded probes.

## 16. Remaining Gaps

Current work-unit conditions: exact final candidate CI, review/acceptance of the
bounded subset and owner disposition of the temporary dependency inventory policy.
CS-048 remains Review / QA; related testing/container parent issues remain open.

Pre-release blockers: vulnerability triage/remediation or approved specific
exceptions, root images/container hardening, branch protection, missing creator E2E
and later Sprint gates, durable failure handling, alert delivery, public deployment
security and comprehensive security review. Existing npm debt is not erased.

Post-MVP/when justified: distributed tracing/log aggregation, production paging
platform, advanced anomaly detection and load benchmarking. Signing/provenance
are not supplied by SBOM generation and require a release design.

## 17. Work Unit 2.5 Decision

**CONDITIONAL**. The coherent foundation is implemented and initial hosted behavior
is demonstrated, but unconditional acceptance would misrepresent unresolved policy
and security-review conditions. A green inventory scan is not a clean security bill.

## 18. Next Stage

Resolve the work-unit conditions, accept the bounded PR and verify resulting main.
Only after Work Unit 2.5 is accepted should Gate 4 - User Isolation and CORS
Acceptance be authorized. Gate 4 was not started. No production release authorized.
