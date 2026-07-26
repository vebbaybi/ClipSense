# Contributing To ClipSense

## Branches And Commits

- Use `codex/<scope>` for agent branches and `<type>/<scope>` for contributor branches.
- Start from an inspected, clean tracked state and record pre-existing untracked work.
- Keep commits focused: audit, documentation, governance, ADR, tooling, runtime, and
  validation changes should be independently reviewable where practical.
- Do not commit secrets, media uploads, models, runtime databases, or unrelated user files.

## Pull Requests

A pull request must state scope, behavior changes, ownership boundary, risks,
dependencies, migrations, validation results, blocked checks, documentation updates,
and Kanban evidence. Link the relevant GitHub issue and Project #19 item.

## Definition Of Done

Work is Done only when:

- Acceptance criteria are satisfied by implementation, not a placeholder.
- Focused automated tests pass.
- Required runtime/integration checks pass, or the item remains In Review/Blocked.
- Public contracts, schema changes, configuration, and operator behavior are documented.
- The Kanban item links to files, commands, commit/PR, and remaining risks.
- No user work or unrelated behavior was altered.

## Tests

- Go: `go test ./...` from `apps/api`.
- Python: compile plus `python -m unittest discover -s tests` from
  `apps/api/ai_worker` until a Python tooling ADR is approved.
- Web: `npm ci` then `npm run build` from `apps/web`.
- Compose: `docker compose config`; runtime claims additionally require actual container
  startup and a smoke test.
- A build that hangs before process completion is not a pass.

## Generated Code And Dependencies

- A generated destination must have one source and one reproducible command.
- Do not hand-edit generated output.
- New direct dependencies require purpose, license/platform review, maintenance/security
  assessment, alternatives, and owner approval.
- Major runtime upgrades and infrastructure frameworks require an accepted ADR or an
  explicitly isolated security repair.

## Migrations And ADRs

Schema changes require a versioned migration and rollback/reversal note once the
migration tool is approved. Changes to contracts, persistence, auth, workflow
orchestration, vector storage, desktop architecture, observability, or task running
require an ADR decision before adoption.

## Documentation And Kanban

Update `docs/CURRENT_STATE.md` when implementation truth changes. Keep product intent,
historical material, and current evidence distinct. Follow
`docs/operations/KANBAN_WORKFLOW.md`; no task is Done without acceptance evidence.

