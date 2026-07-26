# Sprint 0 Validation Report

Validation date: 2026-07-26

## Passed

| Validation | Command | Result |
|---|---|---|
| Worktree inspection | `git status`, `git diff`, `git branch -vv`, `git remote -v`, `git stash list` | Starting tracked/staged state clean; seven protected untracked files classified |
| Physical inventory | `git ls-files`, `git ls-files --others --exclude-standard`, `rg --files` | Runtime, proposal, placeholder, historical, and untracked paths distinguished |
| Repository doctor | `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/doctor.ps1` | Exit 0; reports tools, OS/AMD64, ports, env absence, and OneDrive warning |
| Go unit tests | `go test ./...` in `apps/api` | Pass |
| Python compilation | `python -m py_compile main.py zip_safety.py` | Pass |
| Python unit tests | `python -m unittest discover -s tests` | Pass: 7 tests |
| TypeScript checking | `npx tsc --noEmit` in `apps/web` | Pass |
| Dependency audit | `npm audit --omit=dev --json` | Command completed; findings recorded, not remediated |
| Kanban self-test | `rebuild-clipsense-kanban.ps1 -SelfTest` | Pass: 17 safeguards |
| Kanban write guard | `rebuild-clipsense-kanban.ps1 -DryRun` without override | Expected exit 1 before GitHub access |
| PowerShell parsing | Parser check over repository `*.ps1` files | 33 files parsed, 0 syntax errors |
| ADR completeness | Required-section check | 14 ADRs, 0 missing required sections |
| Documentation local links | Markdown local-link existence check | 0 missing local targets |
| Git whitespace | `git diff --check` | Pass |
| Kanban live reconciliation | GitHub CLI reads/writes scoped to Project #19 | Issues #67-#78 created, added, and assigned non-Done statuses |

The dependency audit reported one critical aggregate for the pinned Next.js dependency
and one high aggregate involving PostCSS. Sprint 0 intentionally does not mutate
dependencies.

## Failed Or Not A Valid Gate

| Validation | Result | Impact |
|---|---|---|
| `npm run lint` | Opened Next.js interactive first-time configuration; no lint run occurred despite process exit 0 | CI lint is not reproducible; configure in Sprint 1 |
| `gofmt -l .` | Reported pre-existing `apps/api/main.go` formatting drift | No runtime file formatting was changed in Sprint 0 |

## Blocked

| Validation | Exact blocker | Sprint 0 impact | Recommended environment |
|---|---|---|---|
| Docker Compose config/runtime | Docker CLI and Engine unavailable on current Acer machine | Does not block documentation foundation; blocks runtime claims | HP with Docker Desktop/Engine |
| Full ZIP MVP smoke test | Requires working Compose, worker image repair, FFmpeg/models | Sprint 1 acceptance gate | HP |
| Next.js production build | Earlier build compiled/type-checked but did not exit after trace finalization in OneDrive; bounded Sprint 0 attempt was inconclusive | Type checking passed, but no production-build pass is claimed | HP outside OneDrive or CI |
| FFmpeg/media checks | FFmpeg unavailable on current machine | No media runtime claim | HP |
| PowerShell Core portability | `pwsh` unavailable | Windows PowerShell path only is evidenced | HP/Mac with PowerShell Core |
| Kanban deterministic sync | Candidate sync script is protected untracked work; tracked taxonomy conflicts with live Project #19 | Card remains Bugged; live board was updated through scoped GitHub CLI commands | Review candidate work before adoption |

## Not Applicable

- Generated API/client drift: no approved generator or generated destination exists.
- Database migration drift: no versioned migration system exists.
- Desktop validation: no desktop implementation exists.
- Browser E2E: no E2E suite exists.

## Validation Integrity

No Docker startup, production web build, lint pass, end-to-end workflow, or Kanban
sync success is claimed. Generated `apps/web/tsconfig.tsbuildinfo` created by the
type check was removed after verifying its path was inside the workspace.
