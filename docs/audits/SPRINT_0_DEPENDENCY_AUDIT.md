# Sprint 0 Dependency And Framework Audit

Report date: 2026-07-26

No dependency is changed by this audit.

## Node And TypeScript

| Dependency | Current | Purpose / usage | Finding | Urgency / destination |
|---|---:|---|---|---|
| `next` | 14.1.4 | App Router/build/runtime | Active; 14.x is unsupported and affected by later advisories | Isolated Sprint 1 upgrade to supported line |
| `react`, `react-dom` | 18.2.0 | UI runtime | Active; coupled to Next upgrade | Sprint 1 with Next |
| `react-dropzone` | ^14.2.3 | ZIP selection | Active | Retain |
| `react-hot-toast` | ^2.4.1 | Upload feedback | Active only on upload page | Retain; decide global error UX later |
| `@headlessui/react` | ^1.7.17 | Headless components | No import found | Remove after clean dependency proof |
| `@heroicons/react` | ^2.1.4 | Icons | No import found | Remove after clean dependency proof |
| `@tanstack/react-query` | ^5.8.4 | Server state | No import found | Either adopt deliberately for polling or remove |
| `clsx` | ^2.1.1 | Class composition | No import found | Remove after proof |
| `framer-motion` | ^11.0.17 | Motion | No import found | Remove unless near-term interaction requires it |
| `zod` | ^3.23.0 | Validation | No import found | Reserve for generated/runtime contract validation decision |
| Tailwind/PostCSS/Autoprefixer | 3.4.x/8.4.x/10.4.x | Styling build | Active | Retain during runtime repair |
| ESLint/Next config | 8.56/14.1.4 | Lint | Active but coupled to unsupported Next | Upgrade with Next |
| TypeScript/types | 5.3.3/current pins | Type checking | Active | Review with Next upgrade |

The official Next.js security guidance recommends supported 15.5.x or 16.x lines:
<https://nextjs.org/blog>. The upgrade requires a focused compatibility test rather
than `npm audit fix --force`.

On 2026-07-26, `npm audit --omit=dev` reported two affected production dependency
groups: one critical aggregate for `next@14.1.4` and one high aggregate involving
PostCSS versions in the installed tree. The report recommends at least Next 14.2.35,
but 14.x is no longer a suitable long-term supported target; Sprint 1 must select a
currently supported line and regression-test it.

## Go

| Dependency | Current | Purpose | Finding | Recommendation |
|---|---:|---|---|---|
| `go-chi/chi/v5` | 5.0.10 | Router/middleware | Active | Retain; review current patch in Sprint 1 |
| `golang-jwt/jwt/v5` | 5.2.1 | Custom JWT | Active | Retain for repair; reconsider under OIDC ADR |
| `google/uuid` | 1.6.0 | IDs | Active | Retain |
| `jackc/pgx/v5` | 5.5.4 | Postgres driver | Active | Retain; evaluate sqlc compatibility |
| `redis/go-redis/v9` | 9.5.1 | Queue/health | Active | Retain until workflow decision |
| `x/crypto` | 0.21.0 | bcrypt | Active | Review current patch/security in Sprint 1 |

No `govulncheck` tool is committed or installed. Add it to CI only through a
reproducible pinned tool workflow.

## Python

| Dependency | Current | Purpose | Finding | Recommendation |
|---|---:|---|---|---|
| `openai-whisper` | 20231117 | Transcription | Active, old and resource-heavy | Compare with faster-whisper after baseline |
| `sentence-transformers` | 2.2.2 | Embeddings | Active; old ecosystem pin | Compatibility/security review |
| `scikit-learn` | 1.4.2 | KMeans | Active | Retain for MVP |
| `numpy` | 1.26.4 | Numeric vectors | Transitive/direct runtime use by libraries | Retain until dependency resolver migration |
| `ffmpeg-python` | 0.2.0 | Duration probe | Active; audio extraction separately uses subprocess | Consolidate media adapter later |
| `python-dotenv` | 1.0.1 | Environment loading | No import found | Remove after clean proof |
| `redis` | 5.0.1 | Queue consumption | Active | Retain until workflow decision |
| `psycopg[binary]` | 3.1.18 | Postgres | Active | Retain |
| `qdrant-client` | 1.9.1 | Vector writes | Active | Retain until vector ADR |

The worker also imports `whisper`, model artifacts, and FFmpeg at runtime but has no
locked transitive environment. Evaluate `pyproject.toml` + `uv`, Ruff, Pyright, and
pytest as one approved tooling migration; do not mix it with worker behavior repair.
Python 3.11 is used by CI/container while the current workstation has Python 3.10.

## Containers And GitHub Actions

| Item | Current | Finding |
|---|---|---|
| Postgres | `15-alpine` | Active; floating patch tag, development credentials |
| Redis | `7-alpine` | Active; floating patch tag, no auth |
| Qdrant | `v1.8.1` | Active, older pinned minor |
| Go builder/runtime | `golang:1.21-alpine`, `alpine:3.19` | Build runs `go mod tidy`, allowing manifest drift inside image |
| Python | `python:3.11-slim` | Matches CI, not current workstation |
| Node | `node:20-alpine` | Supported baseline; Dockerfile uses `npm install`, CI uses `npm ci` |
| GitHub Actions | checkout v4, setup-go v5, setup-python v5, setup-node v4 | Active major tags; no least-privilege `permissions` block |

Pinning by digest improves reproducibility but increases update overhead. Address
container integrity, non-root users, health checks, dependency install consistency,
and image scanning together after the Sprint 1 runtime baseline.

## PowerShell And Shell

- Tracked Kanban logic uses Windows PowerShell-compatible syntax and GitHub CLI.
- Existing developer/build/database utilities are Bash-only.
- `pwsh` is not available on the current machine, so PowerShell Core portability is
  not verified.
- `scripts/doctor.ps1` is read-only and compatible with Windows PowerShell; validate
  with PowerShell Core on HP/Mac before claiming cross-platform support.
- Taskfile is not adopted. ADR 0012 remains Proposed.

## Framework Evaluation Summary

| Candidate | Fit | Decision point |
|---|---|---|
| OpenAPI | Strong fit for one public contract | Prototype after current routes/errors stabilize |
| oapi-codegen | Strong Go contract fit; generation adds workflow | Pair with OpenAPI decision |
| Orval | Strong typed React client option | Compare output/runtime validation and React Query coupling |
| Goose | Good small Go/Postgres migration tool | Recommended candidate for Sprint 1 ADR approval |
| sqlc | Good typed SQL without ORM | Adopt after schema migration owner exists |
| Testcontainers | Good integration isolation; Docker required | Sprint 1/2 after HP Docker baseline |
| Temporal | Durable but operationally heavy for current MVP | Defer; first specify workflow requirements |
| pgvector | Simplifies services; ties vectors to Postgres | Benchmark/operational comparison with Qdrant |
| OIDC Authorization Code + PKCE | Best long-term web/desktop auth boundary | Defer provider selection; repair current auth first |
| Tauri 2 + Vite/React | Good future cross-platform desktop fit | Post-web baseline; ADR approval required |
| faster-whisper | Potential speed/resource improvement | Benchmark on target CPU/GPU later |
| OpenTelemetry | Appropriate standard | Add only with collector/export plan |
| Playwright | Appropriate browser E2E tool | Sprint 1 smoke test candidate |
| Taskfile | Cross-platform command surface | Do not adopt until available tools/installation policy approved |

## Custom Logic To Replace Or Encapsulate Later

- Startup DDL -> versioned migrations.
- Repeated SQL -> typed queries after schema ownership.
- Handwritten public types/client -> generated contract adapters.
- Redis list queue -> durable acknowledged workflow mechanism.
- Custom auth -> standards-based provider if approved.
- Ad hoc logging -> OpenTelemetry-compatible structured telemetry.
- Mixed subprocess/probe calls -> tested media adapter.
- Page-local validation/types -> contract-derived validation where appropriate.
