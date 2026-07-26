# ClipSense Agent Guardrails

These rules apply to every coding agent working in this repository.

1. Inspect the physical repository before creating files or abstractions.
2. Search the entire repository for an existing owner of the capability.
3. Follow `docs/architecture/SERVICE_OWNERSHIP.md`.
4. Treat proposal trees, blueprints, and backlog text as intent, not implementation.
5. Do not create empty directories, marker files, unused interfaces, or future services.
6. Prefer approved mature libraries to handwritten infrastructure; record material
   architectural choices in an ADR before adoption.
7. Keep generated files isolated and provide exactly one generation command and source.
8. Avoid broad formatting, dependency churn, and unrelated refactors.
9. Update focused tests and current-state documentation with behavior changes.
10. Update GitHub Project `ClipSense` number 19 through the documented Kanban workflow.
11. Deliver one runnable or internally verifiable vertical increment per sprint.
12. Report every changed file, added dependency, migration, and validation command.
13. Never claim behavior works without test or runtime evidence.
14. Never mark Kanban work Done without acceptance evidence.
15. Preserve untracked and user-authored work; never alter it silently.
16. Never use destructive Git cleanup/reset commands or force push.
17. Keep each capability within its owner; do not scatter a feature across arbitrary paths.
18. Read `docs/CURRENT_STATE.md` before implementation and update it when truth changes.

