# Telemetry Redaction Policy

Default deny: only documented fields and constrained values reach operational
logging. This is safer than guessing every credential's shape with regular expressions.
Never pass exceptions, request/response objects or job payloads into logging.

Exclude passwords, JWTs, Authorization, signing keys, DSNs and credentials, cloud
secrets, request bodies, transcripts, media, email, filenames and filesystem paths.
UUID diagnostic identifiers are pseudonymous operational data: restrict log access.
No user identifier is currently logged. Browser telemetry is opt-in local only.

API replaces panic/library errors with fixed events; no panic stack is emitted.
Worker never serializes exception messages/stack arguments. FFmpeg/FFprobe stderr
is suppressed on private-media paths. Keep model/dependency verbose logging off.
Dependency libraries may emit their own startup diagnostics; do not assume these
are redacted by the application adapter. Review diagnostics before external sharing.

Automated checks cover injected extra fields, invalid/newline IDs, exception text,
private-data metric labels and browser records. Hosted checks include sentinel
credentials/token exclusion from API logs and worker fixture events. This does not
prove all third-party output for all inputs; future changes require fresh tests.

Local Docker logs: rotate at 10 MB x 3 per application container. CI artifacts:
14 days; access follows repository visibility. Never upload resolved Compose
environment or full docker inspect Config.Env. SBOMs can expose package inventory,
not credentials. A leaked real secret requires revocation and incident handling;
deleting a log is not remediation. CI fixture values are disposable, never reused.
