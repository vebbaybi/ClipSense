# Sprint 0 Physical Repository Inventory

Report date: 2026-07-26

This inventory is based on physical filesystem entries, `git ls-files`, imports,
CI configuration, and script references. Directory diagrams in blueprint files
are not evidence that their pictured paths exist.

## Runtime And Tests

| Path | Classification | Status | Referenced by | Owner | Recommendation | Risk / evidence |
|---|---|---|---|---|---|---|
| `apps/web/` | Runtime implementation | Active, partial | Compose, CI | Web application | Retain; repair in Sprint 1 | Next.js App Router routes physically exist. |
| `apps/web/src/app/page.tsx` | Runtime route `/` | Active | Next.js | Web application | Retain | Login/register screen. |
| `apps/web/src/app/(dashboard)/page.tsx` | Runtime route `/dashboard` | Active, defective | Next.js | Web application | Sprint 1 route/auth repair | Route group does not add a URL segment; actual route is `/`, colliding with `app/page.tsx`. |
| `apps/web/src/app/(dashboard)/batches/new/page.tsx` | Runtime route `/batches/new` | Active, mislinked | Next.js | Web application | Sprint 1 route repair | UI links to `/dashboard/batches/new`. |
| `apps/web/src/app/(dashboard)/batches/[id]/page.tsx` | Runtime route `/batches/[id]` | Active, mislinked | Next.js | Web application | Sprint 1 route/export repair | UI links to `/dashboard/batches/{id}`; anchor export cannot send bearer token. |
| `apps/web/src/app/(dashboard)/settings/page.tsx` | Runtime route `/settings` | Active, static | Next.js | Web application | Retain | UI links to `/dashboard/settings`. |
| `apps/web/src/lib/api/client.ts` | Runtime API adapter | Active | Web pages | Web application | Retain until generated-client ADR is approved | One hand-written fetch wrapper. |
| `apps/web/src/hooks/useAuthToken.ts` | Runtime auth adapter | Active | Web pages | Web application | Sprint 1 security review | Browser `localStorage` token owner. |
| `apps/api/main.go` | Runtime implementation | Active, partial | Dockerfile, CI | Go API | Retain; modularize only with tests | Owns routes, auth, schema creation, queries, upload, exports, health. |
| `apps/api/main_test.go` | Unit tests | Active | CI | Go API | Expand in Sprint 1 | Focused handler tests; no database/worker integration. |
| `apps/api/ai_worker/main.py` | Runtime processor | Active source, container blocked | Worker Dockerfile | Python processor | Retain; repair packaging in Sprint 1 | Imports `zip_safety`, but image omits that file. |
| `apps/api/ai_worker/zip_safety.py` | Runtime media validation | Active in source | Worker, tests | Python processor | Retain as ZIP extraction owner | Seven safety tests cover it. |
| `apps/api/ai_worker/tests/` | Unit tests | Active | CI | Python processor | Retain and expand | ZIP validation only. |

## Build, CI, Infrastructure, And Operations

| Path | Classification | Status | Owner | Recommendation | Evidence |
|---|---|---|---|---|---|
| `.github/workflows/ci.yml` | CI/CD | Active, incomplete | Repository operations | Retain; add integration/security checks later | Runs Go tests, Python compile/tests, web build, Compose config. |
| `docker-compose.yml` | Local infrastructure | Active definition, runtime unverified | Repository operations | Sprint 1 repair and smoke test | Defines Postgres, Redis, Qdrant, API, worker, web. |
| `apps/*/Dockerfile` | Build configuration | Active definitions | Owning service | Repair worker image; harden later | API and web build definitions exist; worker packaging is incomplete. |
| `.env.example` | Configuration reference | Active, incomplete | Repository operations | Retain and align with secure configuration | No local `.env` exists. |
| `scripts/dev/*.sh` | Developer operations | Active Bash adapters | Repository operations | Retain pending task-runner ADR | Wrap Compose start/stop/reset. |
| `scripts/build/*.sh` | Build/release operations | Partially active | Repository operations | Review package command in Sprint 1+ | Build wraps Compose; release path is unverified. |
| `scripts/database/*.sh` | Database operations | Obsolete for active stack | Repository operations | Replace in Sprint 1 | Copies SQLite while active Compose uses Postgres. |
| `scripts/utils/*.sh` | Utility operations | Active but platform-limited | Repository operations | Retain | Secret generation and HTTP health probe. |
| `infra/scripts/deploy/deploy.sh` | Deployment | Defective | Repository operations | Repair in Sprint 1 | Resolves `infra` as root although Compose is at repository root. |
| `infra/scripts/monitoring/setup-prometheus.sh` | Placeholder | Inactive | None | Do not execute; decide after observability ADR | Contains only TODO output. |
| `infra/scripts/ssl/generate-certs.sh` | Infrastructure utility | Unconnected | Repository operations | Retain as local helper or remove after review | No reverse proxy consumes generated certificates. |
| `rebuild-clipsense-kanban.ps1` | Kanban adapter/generation source | Tracked but unsafe for live board | Repository operations | Disable pending consolidation | Embedded cards conflict with live issue taxonomy. |
| `sync-clipsense-kanban.ps1` | Untracked Kanban adapter candidate | Protected, inactive in clean checkout | Unknown pending review | Leave untouched; review before adoption | Self-test passes; unavailable to contributors after clean checkout. |

## Documentation And Historical Material

| Path | Classification | Status | Owner | Recommendation | Risk / evidence |
|---|---|---|---|---|---|
| `CS.md` | Product specification plus stale implementation snapshot | Mixed | Product documentation | Retain; use only product sections as intent | Snapshot overstates current runnable state. |
| `CS.txt` | Proposed repository tree | Historical blueprint | Product documentation | Retain and label historical | Describes many paths that do not exist. |
| `docs/MVP_STABILIZATION.md` | Historical stabilization report | Stale | Historical documentation | Retain; supersede with `CURRENT_STATE.md` | Claims Go/go.sum unavailable although both exist now. |
| `1807ish/roadmap.eos` | Product roadmap | Historical/aspirational | Product documentation | Retain | Text, not executable `.eos` data. |
| `1807ish/whitepaper.block` | Product whitepaper | Historical/aspirational | Product documentation | Retain | Markdown-like prose, not a runtime block. |
| `apps/app` | Proposed repository tree | Historical blueprint | Historical documentation | Retain for now; never treat as an app | Extensionless regular file, not a directory. |
| `apps/package/packages` | Proposed shared-package tree | Historical blueprint | Historical documentation | Retain for now | Extensionless regular file, not packages. |
| `apps/api/apis` | Proposed API tree | Historical blueprint | Historical documentation | Retain for now | Extensionless regular file, not Go source. |
| `apps/web/src/source` | Proposed web tree | Historical blueprint | Historical documentation | Retain for now | Extensionless regular file. |
| `apps/web/src/{components,features,hooks,lib,stores,styles}/` marker files | Placeholder/obsolete | Inactive | None | Remove only in a dedicated reviewed cleanup | Not imported; several explicitly call themselves retired. |
| `apps/web/src/types/type`, `apps/web/src/app/apps` | Proposed tree fragments | Inactive | None | Review with placeholder cleanup | Regular files, not TypeScript or directories. |

## Dependency Manifests

| Path | Classification | Status | Owner |
|---|---|---|---|
| `apps/web/package.json`, `package-lock.json` | Node dependency source/lock | Active | Web application |
| `apps/api/go.mod`, `go.sum` | Go dependency source/checksums | Active | Go API |
| `apps/api/ai_worker/requirements.txt` | Python dependency pins | Active but legacy workflow | Python processor |

There is no physical `packages/`, `services/`, `database/`, root `tests/`, desktop
application, Rust crate, OpenAPI document, versioned migration directory, generated
API client, or integration/E2E suite.

## Verification Method

Re-run the physical inventory with:

```powershell
git ls-files
git ls-files --others --exclude-standard
rg --files -g '!node_modules' -g '!.next' -g '!.git'
```

