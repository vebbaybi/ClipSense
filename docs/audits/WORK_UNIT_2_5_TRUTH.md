# DevSecOps And Observability Truth Report

Inspected 2026-09-22 before implementation. PR #85 merged as
`99cb659243113623b42e6ae6d713977259cbb098`; main CI `35723984517` passed.
Local main was `ecd3131db43829800195328055323dd7cce9b919` and was fast-forwarded.
The previous clean candidate branch was `0b09c22561c5e8cd14232d889dbb0bf55bba2931`,
one merge commit behind remote main. New work starts from the merged main SHA.

## Testing Reality

| Layer | Existing evidence | Gap |
|---|---|---|
| Go unit/security | main_test, security_test, upload_test; sqlmock and miniredis; go test/vet | Broader authorization, persistence and contract coverage |
| Python unit/security | 12 unittest ZIP/media safety tests; py_compile | Processing/model/dependency tests |
| Web | Next production build and its type checks | Unit/component/browser/a11y tests and API mocks |
| Integration | Hosted real Postgres/Redis/Qdrant startup, health, HTTP auth/upload rejection, FFprobe fixture | Actual full processing, migration and durable queue tests |
| Reliability | Redis interruption/recovery, API SIGTERM/restart, worker stability | Crash/retry/duplicates/disk pressure |
| E2E/performance/AI quality | No accepted complete workflow or evaluation suite | Separate later gates |

This is limited polyglot testing, not absence of tests. Build success is not browser
behavioral acceptance. Docker runtime probes are integration tests, not creator E2E.

## Security Pipeline Before Changes

| Control | State | Enforcement |
|---|---|---|
| govulncheck | Implemented | Reachable findings fail CI |
| npm audit JSON | Partial | Evidence only; eight findings previously recorded |
| Python/container/SAST/secret scans | Missing | None |
| Dockerfile/IaC scan, SBOM | Missing | None |
| Provenance/signing | Missing/deferred | SHA-attributed runtime artifacts are not attestations |
| Tests/build/runtime | Implemented | Fail CI |
| Workflow permissions | Implemented | contents:read; third-party Actions use mutable major tags |
| Branch protection | Missing | GitHub API reports main is not protected (404) |

## Observability Before Changes

API uses Go text logs and safe fixed auth/upload errors, but readiness and other
failures log raw library errors. Worker prints messages including filenames and
exception strings. Web upload/API adapter can expose raw response text. Docker
captures stdout/stderr and runtime artifacts. API live/ready endpoints exist.
No request/job correlation, metrics, dashboard, traces, alerts or redaction tests.
ADR 0013 was proposed; monitoring setup script was a TODO. CS-048 (#48) also asks
for broader product analytics and alert delivery, which this subset cannot close.

## Bounded Implementation Plan

Native structured logging, allowlisted fields, UUID correlation in Redis metadata,
private Prometheus endpoints, optional Prometheus/Grafana profile, one dashboard,
actionable local alert rules and incident runbook. Defer collectors/log storage,
distributed tracing, paging integrations and workload-specific thresholds.
Add scan evidence and explicit merge/release policies, not fictional deployment.
Final exact-SHA acceptance is recorded on the work-unit PR after hosted verification.
