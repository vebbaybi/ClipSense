# ADR 0016: Authentication And Upload Boundary Controls

Status: Implemented for Sprint 1 Gates 2/3; acceptance requires exact-candidate CI.

Use existing Go JWT/bcrypt/Redis and standard-library multipart/ZIP APIs. Redis
owns only expiring attempt counters, never durable authentication identity. Keep
media probing and filesystem extraction in the Python worker using zipfile and
FFprobe. No new service, identity provider, job engine, or schema is introduced.

Fail closed on missing production secret or auth-limiter outage. Opt into local
development explicitly. Rate limits trust socket peers, not forwarded headers.
Use finite secret-keyed Redis buckets and upload concurrency/storage quotas to
bound resource use. Quotas are intentionally single-API-instance policy.

Validate archives before enqueue and again at extraction. Use a pending-state
claim to coordinate source-file ownership on ambiguous Redis handoff failures.
Retain bounded, recorded uncertain handoffs for operator reconciliation rather
than deleting media a worker might already own. Reliable retries/outbox, identity
isolation and production deployment are separate gates, not solved here.

Detailed policy, compatibility restrictions, tests and remaining limitations:
[Work Unit 2 report](../audits/WORK_UNIT_2_AUTH_UPLOAD.md).
