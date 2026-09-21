# Gate 1 Runtime Baseline Acceptance

## Repository Truth

Inspection date: 2026-09-21. Repository root:
`C:/Users/Wildf/OneDrive/Desktop/The 1807/ClipSense`.

Branch `codex/clipsense-sprint-1-runtime-baseline` started at
`02e5a4eff9e5e943d29401c2fcacc40a047f3b04`, matching its upstream on remote
`clipsense` (`https://github.com/vebbaybi/ClipSense.git`). Staged, unstaged,
and untracked changes were empty. Main was
`e112d28edde10c98c3c9fc3382bfe024317bcfe9`. The branch had 21 unique commits
and main seven; `git cherry` found no patch-equivalent branch commits on main.
PR #84 was open, draft, and conflicting.

## PR Review And Integration

Required runtime work includes worker module packaging and dependency compatibility,
deferred model initialization, bounded Redis reconnect waits, canonical dashboard
routes and auth hydration, API liveness/readiness/timeouts/graceful shutdown,
Compose health dependencies, and the CI Docker verification script. Go health and
lifecycle tests are retained; existing Python ZIP tests are retained.

Sprint 0 ownership/ADR/audit documents provide context, not new implemented features.
The tip also contains historical board scripts, backlog documents, and a branding
asset. These existing user commits are preserved without executing the scripts or
claiming them as Gate 1 delivery. No equivalent runtime implementation exists on main.

Main's issue templates, CSUS.md, and license are preserved. The sole merge conflict
was `rebuild-clipsense-kanban.ps1`: the branch added an execution guard while main
deleted the obsolete script. The deletion is preserved; restoring a legacy board
generator is unnecessary for runtime acceptance. No application conflict required
choosing between competing runtime implementations.

The verification script now also restarts the API after SIGTERM and asserts healthy
state, liveness, and readiness. This closes a required acceptance coverage gap.

## Acceptance Status

The reviewed integrated candidate `ce9d9929dfec43639a0e1a57d72c0ee10352eb66`
passed all five CI jobs in [run 35644022417](https://github.com/vebbaybi/ClipSense/actions/runs/35644022417).
PR #84 was then merged as `4f073e5fa6f10beccd8bd8ad6e289ec9b76b373b`.
Local Docker is unavailable on PATH; GitHub-hosted Linux is authoritative under
ADR 0015. There is no local Docker result to claim or compare.

This record accepts the tested candidate's runtime scope. The documentation-only
reconciliation commit must also pass the same CI on main before final Gate 1
sign-off. Its SHA and run are recorded in the PR acceptance follow-up, avoiding a
self-referential documentation commit. Earlier July runs are historical only.

## Validation And Review

- Local `go test ./...`, `go vet ./...`, `python -m py_compile main.py zip_safety.py`,
  and `python -m unittest discover -s tests` passed (seven Python tests).
- CI Go API, Python Worker, Web Build, Docker Compose Config, and Build and verify
  Compose runtime all passed on the candidate above.
- CI ran `npm ci`, `npm run build`, `docker compose config`, and
  `bash scripts/ci/verify-docker-runtime.sh` with `config`, `build`, `runtime`, and
  `diagnostics` modes. Image build took 379 seconds; runtime checks took 72 seconds.
- Artifact: `docker-runtime-ce9d9929dfec43639a0e1a57d72c0ee10352eb66`.
  Reviewed build/Compose logs, HTTP results/bodies, service transitions, worker
  smoke output, SIGTERM result, and full-stack shutdown output.
- API, worker, web images: PASS. Worker imports and FFmpeg: PASS.
- Six-service startup and Postgres/Redis/Qdrant connectivity: PASS.
- Five canonical HTTP routes: PASS (200). This is not authenticated browser E2E.
- API live/ready/compatibility: PASS (200). Redis outage: live 200, ready 503 with
  safe degraded JSON. Recovery: ready 200 with both dependencies reporting ok.
- Worker survives outage and resumes blocking consumption: PASS; no restart loop.
- API SIGTERM exits 0: PASS. API restart, live and ready 200: PASS.
- Full-stack shutdown: PASS, no remaining Compose containers.
- Final runtime diff reviewed for configuration consistency, retained tests,
  imports, lifecycle behavior, accidental deletions, debug code and merge markers.
  No new runtime blocker identified. No credentials or generated runtime artifacts
  were added by this reconciliation. Historical default credentials remain unsafe.
- Focused `git diff --check` passed. Inherited Markdown hard-break whitespace in
  main's documentation was not reformatted as unrelated cleanup.

## Warning Classification

- Expected fault injection: Redis connection failures while intentionally stopped;
  readiness and worker recovered. No persistent reconnect storm or uncaught error.
- Compose obsolete `version`: maintenance warning, not a startup failure.
- Postgres Alpine locale warning: local image limitation; locale behavior untested.
  Local-socket trust initialization warning is a security configuration follow-up,
  not approval for deployment. Normal initdb shutdown/restart is expected.
- Redis memory-overcommit/default-config warnings: hosting and persistence risks
  remain open; this short startup test is not durability or load acceptance.
- pip root-install and resolver-backtracking notices: build hygiene/reproducibility
  follow-up; install/import succeeded. Dependency security is not certified.
- npm deprecated packages and vulnerable Next.js: build reported 14 vulnerabilities
  (1 low, 1 moderate, 11 high, 1 critical). CS-038/security work remains open;
  this is a release blocker, not a reason to claim the runtime test failed.
- No panic, fatal import/FFmpeg error, migration failure, or persistent dependency
  failure was found in the captured runtime logs.

Work Unit 2 remains unstarted. This gate makes no claim about media processing,
authentication security, upload safety, isolation, backups, or release readiness.
