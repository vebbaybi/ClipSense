# Sprint 0 Capability Ownership

Report date: 2026-07-26

## Active Capability Map

| Capability | Existing locations | Current owner | Conflict / risk | Recommended owner | Action / sprint |
|---|---|---|---|---|---|
| Authentication and JWT | `apps/api/main.go`, web login and `useAuthToken.ts` | Go API issues tokens; web stores them | Rules, validation, and UI concerns are mixed; custom auth is weak | Go API until auth ADR is approved; web only adapts sessions | Harden Sprint 1; OIDC decision later |
| Authorization/user identity | `authMiddleware`, user-filtered batch queries | Go API | No project/workspace model | Go API | Retain; integration-test Sprint 1 |
| CORS | `localDevCORSMiddleware`, `.env.example` | Go API | Local-only defaults can be deployed accidentally | Go API configuration | Harden Sprint 1 |
| API errors | `httpError`, web `api()` wrapper, direct upload `fetch` | Go API and web | Raw server errors leak; upload bypasses wrapper | Go API response contract; web adapter | Repair Sprint 1 |
| API client/types | `client.ts`; page-local `Batch`, `Clip`, `Storyline` types; Go structs | Duplicated between pages/API | Contract drift already possible | Future OpenAPI contract, generated Go/TS adapters if ADR approved | Decide Sprint 1+, no replacement in Sprint 0 |
| Batch/clip/storyline models | Go structs, SQL DDL, Python tuples/SQL, TS page types | No single schema owner | Four handwritten representations | Versioned Postgres schema + public API contract | Migration/contract baseline Sprint 1 |
| Project model | Proposal documents only | None | Product term is not implemented | Future domain contract | Deferred |
| Schema/migrations | `migrate()` in `main.go` | Go API startup | Unversioned DDL; no rollback/constraints | Versioned migration tool after ADR | Sprint 1 design/implementation |
| Database queries | Go inline SQL, Python inline SQL | Each service | Shared writes create coupling and drift | Service-owned typed queries; processor uses explicit internal contract | sqlc evaluation, later migration |
| Queue submission/consumption | Go `RPush`, Python `BLPOP` | Split Go API/Python processor | Payload is an unversioned map; no acknowledgement | Internal processing contract; durable workflow owner after ADR | Runtime repair Sprint 1/2 |
| Status/retry | Python `update_status`; Go enqueue failure update | Split | No durable stages, retry, or failure reason | Processing orchestration owner | Sprint 1 backlog |
| ZIP/media validation | `zip_safety.py`; upload form extension filter | Python processor | API accepts unbounded/unverified bytes | Processor owns archive extraction; API owns request limits/intake validation | Repair Sprint 1 |
| FFmpeg/media probe | Python subprocess and `ffmpeg-python` | Python processor | Two invocation styles, no limits | Python media adapter | Consolidate after tests |
| Transcription | `whisper` calls in worker | Python processor | Startup download/resource concerns | Python processor | Tooling/model ADR; later |
| Summarization/classification | Handwritten worker functions | Python processor | Heuristic placeholders are described as AI output | Python processor | Retain as MVP behavior; label honestly |
| Embeddings/clustering/order | sentence-transformers, sklearn, Qdrant calls | Python processor | Qdrant failure fails whole batch; no tenant payload | Python processor with storage contract | Sprint 1/2 |
| Exports | Go `exportBatch`; browser anchor | Go API | Authenticated browser path broken | Go API generates representation; client downloads via API adapter | Sprint 1 |
| Upload limits/storage paths | Go env/defaults, Compose env, Python env/defaults | Split by service | Values duplicated; upload cap ineffective | Per-service config with shared documented reference | Sprint 1 |
| Logging/health | Go standard logs/health, Python `print`, shell health check | Per service | No structured correlation; worker/Qdrant absent from health | Each service emits telemetry; operations aggregates | Observability ADR, later |
| Deployment | Compose, infra deploy script, build scripts | Repository operations | Deploy script root is wrong; Compose is development-grade | Repository operations | Sprint 1 |
| Backups | SQLite copy scripts | Obsolete | Conflicts with Postgres/Qdrant stack | Repository operations with datastore-native tools | Sprint 1 |
| Kanban | Live GitHub Issues/Project #19, tracked rebuild script, untracked sync/generator variants | Conflicted | Script taxonomies differ; rebuild could create duplicates | Live GitHub Issues + Project #19; one reviewed sync adapter later | Sprint 0 documents and blocks unsafe adapter |

## Ownership Rules

1. The Go API owns public HTTP behavior, authorization, and public representations.
2. PostgreSQL owns durable relational truth; startup DDL is temporary technical debt,
   not a second schema authority.
3. The Python processor owns media execution and analysis implementation. It must not
   become a second public API or independently redefine user authorization.
4. The web app owns browser interaction only. Page-local types are adapters, not
   authoritative domain contracts.
5. Redis and Qdrant are infrastructure implementations, not domain contract owners.
6. GitHub Issues and Project `ClipSense` number 19 own Kanban state. Repository scripts
   are adapters and must fail safely when their embedded model is stale.

## Duplicate And Obsolete Content

- `CS.txt`, `apps/app`, `apps/api/apis`, `apps/package/packages`, and
  `apps/web/src/source` repeat proposed trees. None is imported or executed.
- Small extensionless frontend markers duplicate directory intentions but provide no
  runtime capability.
- Database configuration is repeated in Go, Python, Compose, and `.env.example`.
- Model definitions are repeated across Go, Python tuples/SQL, and page-local TypeScript.
- `rebuild-clipsense-kanban.ps1`, untracked `sync-clipsense-kanban.ps1`, and untracked
  `1807ish/kanban*.ps` represent competing Kanban workflows.

No replacement runtime code is introduced in Sprint 0. Consolidation work is assigned
to Sprint 1 or an approved ADR-guided later sprint.
