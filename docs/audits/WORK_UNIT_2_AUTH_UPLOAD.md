# Work Unit 2: Authentication And Upload Boundaries

## Baseline And Scope

Started from clean, synchronized `main` at
`ecd3131db43829800195328055323dd7cce9b919` on 2026-09-21.
Gate 1 remains accepted. Candidate branch: `codex/sprint1-auth-upload-hardening`.
Only Gates 2 and 3 are implemented here. No schema migration or new service.
Exact candidate SHA, final CI run, artifact, and acceptance decision are attached
to the candidate PR after verification; this document does not certify an unrun CI.

Before changes: bcrypt and signed expiring JWTs existed, but credentials were
unbounded, hash/signing errors were ignored, any HMAC algorithm was accepted,
expiry/identity claims were not required, and `dev-secret` was implicit. No attempt
limiter existed. Multipart memory threshold was mistaken for a request cap. API
stored any bytes as ZIP and leaked raw failures. Worker rejected basic traversal
and bounded total extraction, but trusted extensions and lacked comprehensive
archive/resource/cleanup controls.

## Findings And Changes

| Severity | Authentication baseline finding | Resolution |
|---|---|---|
| Critical | Predictable implicit signing key permits forged authority | Fail-closed configuration; explicit local-only development opt-in |
| High | Missing expiry/identity and broad algorithm acceptance | HS256 only; required exp/iat, valid UUID subject, database identity existence |
| High | Unbounded attempts and credential cost | Redis atomic fixed-window limiter, bounded keys, 4 KiB JSON and field limits |
| Medium | Raw errors; ignored hash/sign errors; malformed headers | Safe JSON errors, checked results, strict single Bearer header |
| Low | Session response caching | `Cache-Control: no-store` |

| Severity | Upload baseline finding | Resolution |
|---|---|---|
| High | No real body cap; unbounded accumulated ZIP storage | Streaming MaxBytesReader, early/streamed 413, concurrency/storage quotas |
| High | Archive links, collisions, expansion attacks | Metadata preflight, canonical extraction, exclusive output creation, CRC and expansion bounds |
| High | Extension-only media trust | Worker FFprobe before any transcription/analysis with demuxer/protocol/time limits |
| Medium | Partial files and queue-failure ownership ambiguity | Scoped rollback; pending/processing compare-and-set; bounded retained ambiguous handoffs |
| Medium | Raw internal/client-controlled errors and log data | Fixed error codes and bounded event fields; no request URL/query logger |
| Low | Unsupported files silently ignored | Explicitly reject non-video members; document ZIP-only policy |

## Authentication Policy

- `APP_ENV` accepts development, test, production, or unset (production-like).
- Only explicitly selected development mode permits an empty `JWT_SECRET` and the
  documented local-only key. Never expose that mode to untrusted networks.
- Otherwise `JWT_SECRET` is base64 for 32-64 cryptographically random bytes. Empty,
  short, malformed, or single-repeated-byte material fails before DB startup.
  Generate it once using `openssl rand -base64 32` and persist it securely; the app
  never silently rotates keys on restart. Existing sessions require re-login when
  migrating from the old default/configuration. Tests configure their own key.
- Email must parse as a plain address, at most 254 bytes. Password: 8-72 bytes
  (bcrypt's limit), not all whitespace. No composition rules. JSON max 4096 bytes,
  no trailing JSON/unknown fields. Existing email case semantics are preserved.
- Tokens require HS256, expiry, issued-at not in the future, expiry after issued-at,
  UUID subject, and an existing user. Header/token parsing is capped at 4096 bytes.
- Auth attempts: default 10 per 60 seconds per socket-peer bucket. Configurable
  `AUTH_MAX_ATTEMPTS` (1-1000), `AUTH_WINDOW_SECONDS` (1-3600). Client forwarding
  headers are not trusted. Behind a proxy, clients share its peer budget until a
  separately reviewed trusted-proxy policy exists.
- Redis Lua increments and expires counters atomically. HMAC-keyed 16-bit buckets
  cap cardinality at 65536 keys with TTL, hiding raw peer addresses. Rare bucket
  collisions can share a limit; there is no single global auth lockout. Distributed
  volumetric attacks and account-targeted attacks across many IPs remain beyond
  this simple admission control. Redis outage returns safe 503, never bypasses it.
- 400 invalid request; 401 unauthenticated/invalid credentials; 409 registration
  conflict; 429 rate limit with Retry-After; 503 dependency unavailable. No raw JWT,
  password, secret, SQL text, or filesystem path is returned/logged by these paths.

## Upload Policy And Ownership

- Exactly one multipart `file` named with a .zip suffix; optional single `name`
  up to 256 bytes. Unexpected fields, extra files, zero bytes, malformed bodies,
  and invalid archives are rejected. Client names never determine storage paths.
- `MAX_UPLOAD_REQUEST_BYTES`: default 210763776 (201 MiB), configurable 1 KiB-1 GiB.
  This is the entire request including multipart overhead, not a file-only limit.
  MaxBytesReader is installed before reading. Streaming avoids multipart spool
  files. Unknown Content-Length is tested. Crossing the bound returns 413.
- Two concurrent upload handlers; 32 stored upload artifacts per API-owned upload
  directory, including successful or ambiguous handoffs. Full storage rejects new
  intake rather than deleting data. This development quota assumes one API writer
  for the shared directory; multi-replica quota coordination is not accepted.
- Archive policy matches API and worker: 200 entries (including directories),
  256 MiB/member, 1 GiB expanded total, 100:1 compression ratio, eight path levels,
  512-byte member names, 128 KiB central directory. ZIP64, multi-disk, encryption,
  special files/symlinks, non-video members, zero-byte media, duplicate/case-folded
  names, file/directory conflicts, non-ASCII and Windows reserved names are rejected.
- Central-directory size/count is bounded before standard-library ZIP parsing.
  API streams members to discard for CRC/size verification under a 30-second
  validation context. It never extracts to disk. Worker resolves destinations
  beneath a freshly owned extraction root; refuses pre-existing workspaces.
- Supported media remains mp4/mov/mkv/webm/avi. Worker verifies a video stream and
  matching container with FFprobe: 5-second timeout, 5 MB probe/analysis settings,
  file protocol and supported demuxers only, selected first video stream, no raw
  diagnostic output. This is asynchronous validation before analysis, not a claim
  that an HTTP-accepted ZIP already contains decodable/processable media throughout.
- Failed HTTP validation removes its generated ZIP; no batch is inserted. Failed
  worker validation removes its own extraction workspace and ZIP, then existing
  failure handling marks the batch failed. Successful media retention is unchanged.
- A failed/ambiguous database insert attempts a bounded compensating delete of the
  unique pending row before returning; the not-yet-queued ZIP is then removed.
  Failed DB compensation retains the ZIP under the storage quota and is explicitly
  logged with batch ID for operator reconciliation, bounding repeated failures too.
- Queue failure atomically marks a still-pending row failed and clears its path
  before removing the ZIP. Worker claims only matching pending ID/path rows. If
  already claimed, worker owns the ZIP. If DB reconciliation is unavailable, the
  durable batch and ZIP are retained under the quota and a bounded batch-ID event
  requests operator reconciliation. Do not delete uncertain handoffs blindly.
  Automatic retry/reconciliation and durable queue semantics remain later work.
- Cleanup errors emit explicit operational events; no other workspace is removed.

## Tests And Validation

- `TestSecretConfiguration`: production/test absence and malformed keys fail;
  explicit development works.
- `TestAuthenticationBoundary` and `TestUnsignedDuplicateAndBackendAuthFailures`:
  valid identity accepted; missing/malformed/unsigned/wrong-key/expired/wrong-method/
  invalid-claim/deleted-user tokens rejected; duplicate/oversized headers rejected;
  backend outage fails closed; errors do not disclose signing/token/backend data.
- `TestCredentialValidationAndHandlers`: bounded credentials, checked registration,
  successful password login, safe failure paths. Wrong password also tested.
- `TestAttemptLimitExpiryAndOutage`: threshold, independent peer, TTL expiry,
  Redis failure and recovery; forwarded headers do not control identity.
- Upload tests: cap with/without Content-Length, invalid multipart/ZIP/empty/multiple
  files, traversal, symlinks, duplicate/conflicting names, compression bombs, DB and
  queue cleanup, ownership preservation, and storage quota.
- Python safety tests: path/resource/CRC rollback, existing-workspace preservation,
  media shape/timeout validation, and ZIP/workspace cleanup.
- Hosted `verify-security-boundaries.py`: real register/login/token, rejected HTTP
  requests, 413, no rejected batch rows, and attempt limit.
- Hosted `verify-media-boundary.py`: real FFmpeg fixture accepted by FFprobe; fake
  media rejected and cleaned. Does not run analysis or creator workflow acceptance.
- Local Go tests/vet, Python tests/compile, web build, and diff checks are run before
  committing. Final exact-candidate CI repeats Go tests/vet/scan, Python tests,
  web clean install/build, Compose validation, all Gate 1 runtime checks, and the
  new HTTP/media boundary checks. Local Docker remains unavailable.

## Dependency Qualification

Initial npm audit: 14 vulnerabilities (one critical). Initial reachable Go scan:
12 findings including HTTP/JWT/Redis/database paths. Minimum targeted remediations:

- Go 1.26.6 (including Docker build/CI) for HTTP/TLS/textproto fixes.
- JWT 5.2.2 for [header allocation advisory](https://github.com/golang-jwt/jwt/security/advisories/GHSA-mh63-6h87-95cp).
- Redis 9.6.3 for response ordering; pgx 5.9.2 for SQL sanitization; x/text 0.39.0
  for malformed-input loops. Removed untrusted RealIP middleware use.
- Next.js and matching ESLint config 15.5.24 for request-handling vulnerabilities,
  including [Windows RCE](https://github.com/vercel/next.js/security/advisories/GHSA-p293-qw3h-jr36).
  React remains 18.2; type resolution is constrained to app-local dependencies.
- Added test-only sqlmock/miniredis (and Lua transitive dependency); no runtime
  service added. Lockfiles generated by package tooling, no broad audit-fix.
- Follow-up npm audit reports eight remaining findings (one low, two moderate,
  five high, no critical). Build/tooling/CSS findings remain Gate 7 work: no
  user-supplied CSS or build input is accepted by these auth/upload paths. No full
  Python/container dependency-security acceptance is claimed.
- Follow-up `go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...` passes with
  zero reachable findings; two imported-package and 22 module findings are reported
  as not called. This is reachability qualification, not absence of all module risk.

## Acceptance Boundary

Intermediate candidate `156c6ea2d04d9c50a6b841bc3cce9c305049387c` passed all five
jobs in [run 35649979087](https://github.com/vebbaybi/ClipSense/actions/runs/35649979087),
including full runtime and HTTP/media boundary checks. Initial run `35649859355`
failed Compose syntax validation; the environment-list correction resolved it.
The final follow-up adds database-insert compensation and production missing-key
image verification. [PR #85](https://github.com/vebbaybi/ClipSense/pull/85) records
the exact final SHA/run and gate decisions without a self-referential commit hash.
CS-015 UI limit presentation and CS-030 production database/DSN guardrails remain
outside this accepted subset. No broader story is automatically closed.

Implementation and local checks are not final acceptance without green exact-SHA
hosted evidence. The candidate PR records final Gate 2/Gate 3 and runtime decisions.
User isolation, CORS acceptance, processing reliability, complete dependency
security, recovery, complete creator workflow, and secure-release readiness remain
NOT YET DEMONSTRATED. Do not start Gate 4 automatically.
