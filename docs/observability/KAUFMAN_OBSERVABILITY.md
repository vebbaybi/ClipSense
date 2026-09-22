# Kaufman Observability

Authoritative bounded Work Unit 2.5 contract. See ADR 0013 and the pre-change
[truth audit](../audits/WORK_UNIT_2_5_TRUTH.md). Acceptance requires hosted evidence;
the presence of these files is not proof of runtime behavior.

## Architecture

Go slog and Python standard logging emit JSON to stdout/stderr through allowlisted
adapters. Browser diagnostics are opt-in local console events only; no collector,
third-party analytics or private payload capture. Prometheus client libraries expose
API metrics on internal 9091 and worker metrics on internal 9092. Do not publish
these ports. Compose's optional `observability` profile adds only Prometheus/Grafana.
No production deployment, log index, pager, tracing backend or collector is implied.

Set `SOURCE_SHA` to `git rev-parse HEAD`, `APP_ENV` and signing material as documented
for Gate 2. Set an explicit `GRAFANA_ADMIN_PASSWORD` of at least 16 characters.
Run `docker compose --profile observability up --build -d`.
Prometheus: http://127.0.0.1:9090; Grafana: http://127.0.0.1:3001 (admin).
Dashboard and data source provision automatically. Never expose these unaudited
local endpoints beyond loopback. Prometheus retention is seven days/1 GB; Docker
log rotation is configured separately. These are local capacity bounds, not SLOs.

## Correlation

Browser sends random UUID request/correlation IDs. API validates canonical UUIDs,
generates replacements for invalid/missing values, returns them in exposed response
headers, and places them in JSON job metadata. Batch UUID is assigned by the API.
Worker validates context fields again. IDs grant no authority and do not establish
user ownership. Existing jobs without context get a new correlation ID, so an old
job cannot retroactively establish browser correlation. No schema migration.

Example: UUID A (browser correlation) -> request UUID B -> API upload.accepted
with batch UUID C -> job.enqueued with A/B/C -> worker job.claimed and stage events
with A/B/C. Look up C in the API log to recover A; context is not stored in Postgres.

## Metrics

| Metric | Purpose / bounded labels |
|---|---|
| clipsense_http_requests_total | API volume/error classes; fixed route operation and status class |
| clipsense_http_duration_seconds | HTTP latency; fixed route operation |
| clipsense_events_total | Fixed event taxonomy, auth/upload/queue/job/export/cleanup outcomes |
| clipsense_rejections_total | API error categories including archive, multipart, too_large; no raw error text |
| clipsense_active_uploads | In-flight upload HTTP requests, including validation |
| clipsense_dependency_up | Latest database/Redis observation from health/scrape |
| clipsense_queue_depth | Redis pending list length; -1 means observation failed |
| clipsense_stage_duration_seconds | Fixed worker stage and completed/failed outcome |
| clipsense_job_duration_seconds | Consumer execution duration, including failed or unclaimed attempts |
| clipsense_storage_available_bytes / total_bytes | Worker processing filesystem pressure |
| up | Prometheus target availability; not application readiness or useful model progress |

No UUIDs, filenames, raw routes, emails, private data, or arbitrary error strings
are metric labels. A queue of zero does not prove completion: BLPOP removes jobs
before completion. Oldest age, durable in-flight age, restart counters, Qdrant
availability polling and host metrics remain gaps. Qdrant failures have stage
events/metrics; Compose diagnostics retain restart counts and dependency probes.
Database stage timing currently covers connection acquisition, not every SQL call.
Retries are not invented: there is no durable retry counter until that gate exists.

## Tracing Direction

Full OpenTelemetry tracing is deferred. Future spans should cover HTTP, Redis
enqueue/consume and major stages, use W3C trace context with validation, and retain
request/correlation IDs as log fields. Never synthesize trace IDs from user IDs.
Correlation and duration histograms are not distributed traces or per-request spans.

## Operator Visibility

One dashboard covers targets, dependencies, uploads, queue/jobs, stage/job/API
latency, auth failures, HTTP classes, disk space and cleanup. Idle histograms can
show no data; that is not a monitoring failure. Logs remain Compose logs, not a
Grafana log data source. Alerts are visible in Prometheus; no delivery receiver is
configured. See [alert policy](ALERT_POLICY.md) and the incident runbook.
