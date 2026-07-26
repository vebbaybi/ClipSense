# ClipSense Documentation

Use this order when sources disagree:

1. **Current implementation truth**: `CURRENT_STATE.md`, physical code, tests, and
   validation reports.
2. **Product intent**: `../CS.md`, with its implementation snapshot treated as
   historical unless reconfirmed.
3. **Architecture decisions**: `architecture/` and proposed records in `adr/`.
4. **Backlog and Kanban**: GitHub Issues and GitHub Project `ClipSense` number 19;
   see `operations/KANBAN_WORKFLOW.md`.
5. **Historical material**: `HISTORICAL_ARTIFACTS.md`,
   `MVP_STABILIZATION.md`, `../CS.txt`, and `../1807ish/`.
6. **Generated artifacts**: none are currently authoritative.

## Current Foundation

- [Current State](CURRENT_STATE.md)
- [Service Ownership](architecture/SERVICE_OWNERSHIP.md)
- [Sprint 0 Worktree Safety](audits/SPRINT_0_WORKTREE_SAFETY.md)
- [Sprint 0 Repository Inventory](audits/SPRINT_0_REPOSITORY_INVENTORY.md)
- [Sprint 0 Capability Ownership](audits/SPRINT_0_CAPABILITY_OWNERSHIP.md)
- [Sprint 0 Dependency Audit](audits/SPRINT_0_DEPENDENCY_AUDIT.md)
- [Kanban Workflow](operations/KANBAN_WORKFLOW.md)
- [Sprint 1 Backlog](planning/SPRINT_1_BACKLOG.md)

## Reading Rule

A path shown inside a document is proposed until the filesystem contains it and an
import, build, test, CI job, or documented command uses it. A plan or Kanban status
is not runtime evidence.

