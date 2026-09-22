# Incident Diagnostics

## Service And Dependencies

For a batch processing for 40 minutes, first record UTC time, batch UUID and the
deployed SOURCE_SHA. Do not ask for tokens, transcripts, passwords or private media.

1. `docker compose ps --all`: running state and restart evidence.
2. `curl -fsS http://localhost:8080/api/health/live` and `/api/health/ready`:
   process response versus database/Redis availability.
3. Open Grafana ClipSense Operations: target/dependency gauges, HTTP errors and queue.
4. `docker compose exec -T redis redis-cli LLEN jobs:batch`: pending count only.
   Do not dump Redis job payloads; they contain private names/paths.
5. `docker compose exec -T postgres pg_isready -U clipsense -d clipsense`.
6. `curl -fsS http://localhost:6333/readyz`: Qdrant readiness, not vector correctness.

## Stuck Batch

Use the authenticated batch API/UI to check status without pasting its token into
shell history. An authorized operator may query only id/status/updated_at from
Postgres for the validated batch UUID (parameterized client or psql bound variable).
Never select transcripts, password hashes or full rows into diagnostic artifacts.

`docker compose logs --since 1h --no-color api worker` gives JSON events. Filter
locally by batch UUID, then correlation_id. Find upload.accepted/job.enqueued,
job.claimed, the latest processing.stage.started/completed/failed and job completion.
Stage duration metrics show aggregate trends, not elapsed time for this exact job.
The last started event's UTC timestamp provides that incident's elapsed time.

- Enqueued but no claim: inspect worker startup/model download and Redis health.
- Claimed but no terminal event: inspect stage, process state and deployment time.
- Failed: identify fixed stage/classification; correlate dependency transitions.
- Queue zero and batch processing: does not imply lost or successful work; current
  BLPOP has no acknowledgement history. Escalate rather than blindly re-enqueueing.
- Missing context: legacy jobs cannot be reconstructed across the queue boundary.

Worker model loading is lazy startup work; metrics availability is not model-ready
proof. Inspect models-stage events. No automatic job retry is provided by this unit.

## Storage And Cleanup

Dashboard shows processing filesystem free/capacity. `docker compose exec -T worker
df -h /data` confirms mount capacity. API intake quota includes retained ZIPs. A
cleanup failure may mean ownership is ambiguous: check batch state and worker logs
before touching the server-generated artifact. Never recursively delete the shared
volume, another batch workspace, or an unverified computed path. Preserve evidence.

## Deployment And Recovery

Compare event version fields with CI exact-SHA artifacts and image inventory.
Record `docker compose images`, service state and safe health results. Avoid full
environment/inspect dumps. Gate 1 evidence includes Redis recovery and API shutdown/
restart; it is not permission to replay jobs. Recovery/backup acceptance is later.

If diagnosis remains unclear, record missing signal and escalate with redacted event
samples. Do not infer completion from healthy containers or fabricate root cause.
