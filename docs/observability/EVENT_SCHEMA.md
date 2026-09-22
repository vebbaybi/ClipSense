# Kaufman Event Schema

Required: timestamp (UTC), level, service, environment, version, event.
API/worker version is a 40-character build SHA or `unknown`; deploy tooling supplies
SOURCE_SHA. Web image build receives the same SHA through NEXT_PUBLIC_BUILD_SHA.
Version fields aid attribution but are not cryptographic build identity.
Optional: request_id, correlation_id, batch_id (canonical UUIDs); operation (fixed
API route); stage (fixed enum); duration_ms (nonnegative); status (HTTP code);
dependency (database/redis/qdrant), retry_count (only when actually known).
API startup failures include a fixed error_code category (configuration, database,
migration, redis, listener, runtime), never the original library exception.

No opaque error string, message template arguments, user ID, filename or payload.
Safe classification is the event and stage/status, not a serialized exception.

Taxonomy:
- service.started/stopped/failed; http.completed/failed
- auth.login.success/failed; auth.rate_limit.triggered; authorization.denied
- upload.started/accepted/rejected; upload.cleanup.failed
- job.enqueued/claimed/invalid/failed/completed
- processing.started/completed; processing.stage.started/completed/failed
- dependency.degraded/recovered
- export.started/failed/completed

Browser: upload.started/rejected, batch.fetch.failed, export.failed,
application.error; fixed category and validated correlation only.
Registration currently shares auth outcome counters with login. A 401/403 records
denial, not proof of complete authorization coverage. Export completion means
handler output production, not client receipt. Existing plain-anchor browser export
cannot report failures through the API wrapper; authenticated export remains open.

Worker stage values: validation, ffmpeg, transcription, embedding, qdrant,
persistence, classification, models, processing. Fields are optional, not fabricated.
The application adapters drop unknown fields/events; schema changes require tests
and privacy review. Third-party dependency logs are not automatically Kaufman events.
