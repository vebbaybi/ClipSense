# ClipSense Polyglot Testing Standard

Testing follows service ownership, not a Python-only framework.

| Layer | Implemented / command | Sprint 1 required / later |
|---|---|---|
| Unit | Go handlers/security/telemetry: go test -count=1 ./...; Python unittest safety/telemetry; Node test browser telemetry | React component tests remain missing |
| Integration | Hosted real Postgres/Redis/Qdrant health, HTTP boundaries, FFprobe, isolated worker validation | Actual processing/persistence outcomes and migration tests |
| Contract | Correlation IDs across API/Redis/worker; local field/type expectations | Broader Web/API and versioned job contracts |
| Security | Auth/token/limiter/upload/archive/redaction tests; basic CORS unit tests | Cross-user and complete CORS acceptance in Gate 4 |
| E2E | No complete creator workflow accepted; web build is not E2E | Upload/process/review/export browser workflow in Sprint 1 |
| Reliability | Redis outage/recovery, process shutdown/restart; telemetry failure fixtures | Crash, retries, duplicate jobs, disk pressure and durable recovery |
| Performance | No load suite | Post-correctness baseline; no invented SLOs |
| AI quality | No accepted deterministic evaluation corpus | Later fixtures/metrics for transcript, classification, summary and ordering |

Go uses sqlmock for SQL interaction expectations and miniredis for bounded limiter
tests. These do not replace actual Postgres/Redis semantics. Python uses lightweight
unittest fixtures; no model download for unit tests. Web uses Node's test runner for
pure telemetry logic, TypeScript checks and Next build, not full React rendering.
Browser/a11y tests, API mocking and AI quality remain explicitly missing layers.

Commands: apps/api `go test -count=1 ./...`, `go vet ./...`; worker
`python -m py_compile main.py zip_safety.py kaufmanlogger.py`,
`python -m unittest discover -s tests`; web `npm ci`, `npm test`,
`npm run build`, `npm run typecheck`. Run build before standalone typecheck so Next
route declarations exist. Install prometheus-client from worker requirements for
lightweight tests. Hosted tests run against exact candidate SHA with retained evidence.

Observability acceptance includes JSON/privacy tests, bounded metrics, real scraper
targets, provisioned Grafana dashboard, promtool validation, invalid auth, fake media
rejection and Redis transitions. The worker correlation fixture executes real claim/
validation/persistence code without ML inference. It is not complete creator E2E.
