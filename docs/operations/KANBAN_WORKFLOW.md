# ClipSense Kanban Workflow

Report date: 2026-07-26

## Source Of Truth

The authoritative board is:

- GitHub repository: `vebbaybi/ClipSense`
- GitHub Issues: stable work-item identity and acceptance content
- GitHub Project: `ClipSense`, user project number 19
- Direction: repository evidence and accepted plans update GitHub; GitHub status does
  not generate or overwrite runtime source.

At inspection, Project #19 contained all 66 repository issues. The live issue taxonomy
uses `CS-001` through `CS-066`, which differs from the tracked rebuild script.

## Tool Classification

| Tool | State | Decision |
|---|---|---|
| `rebuild-clipsense-kanban.ps1` | Tracked; self-test passes; embedded titles conflict with live board | Disabled by default in Sprint 0. Do not run against Project #19. |
| `sync-clipsense-kanban.ps1` | Untracked protected user work; self-test passes | Candidate only. It is not available from a clean checkout and remains untouched. |
| `1807ish/kanban.ps`, `kanban2.ps` | Untracked generator variants | Historical/candidate sources; do not execute. |
| `1807ish/pbl.md` | Untracked backlog representation | Useful comparison evidence; not clean-checkout truth. |
| GitHub CLI | Installed and authenticated | Approved for explicit, scoped reads and Sprint card updates. |

There is no `CSUS.md`. The board is not generated from Markdown in the tracked
repository. Existing synchronization is one-way toward GitHub and can overwrite
project field/status values, so it is not safe until the script taxonomy is reconciled.

## Requirements

- GitHub CLI authenticated with `repo` and `project` scopes.
- Explicit repository `vebbaybi/ClipSense`.
- Explicit project owner `vebbaybi` and number `19`.
- Exact issue-title matching and duplicate detection.
- Read-only inspection before every write session.

## Safe Commands

```powershell
gh auth status
gh repo view --json nameWithOwner,defaultBranchRef
gh issue list --repo vebbaybi/ClipSense --state all --limit 200
gh project view 19 --owner vebbaybi --format json
gh project field-list 19 --owner vebbaybi --format json
powershell -NoProfile -ExecutionPolicy Bypass -File rebuild-clipsense-kanban.ps1 -SelfTest
```

Creating/editing an issue or project item is allowed only for a scoped, reviewed
Sprint task with a structured body and evidence. Capture returned issue URLs.

## Prohibited Commands

- Running `rebuild-clipsense-kanban.ps1 -AllowLegacyTaxonomy` without approved migration.
- Running any untracked generator/sync script against GitHub.
- Deleting, closing, or recreating issues to repair status.
- Bulk replacing issue bodies.
- Using fuzzy title matches.
- Treating a dry run or placeholder document as acceptance evidence.

## Status Definitions

| Status | Meaning |
|---|---|
| New Issue | Newly captured work awaiting triage |
| Ice Box | Accepted future work outside the active delivery path |
| Product Backlog | Valid work not ready to start |
| Ready / Sprint | Independently actionable and selected for delivery |
| In progress | Active implementation |
| Review / QA | Implementation exists; acceptance/validation remains |
| Bugged | Started work stopped by a verified defect or unsafe workflow |
| Done | Implemented, validated, reviewed, and accepted |

## Sprint And Evidence Rules

- A Sprint assignment is a planning commitment, not completion.
- A card body must contain scope, out of scope, acceptance criteria, validation, owner,
  risk, and evidence links.
- Done requires the required file/implementation, passed validation, a commit/PR when
  available, and no unresolved acceptance criterion.
- A plan, ADR proposal, or placeholder cannot close a product story.
- Sprint 0 documentation remains `Review / QA` until all validation and board evidence
  are attached.

## Regeneration And Recovery

There is currently no approved deterministic regeneration command. Sprint 0 deliberately
does not bless the competing scripts. Recovery procedure:

1. Stop all sync runs.
2. Export/read Project #19 items and fields using GitHub CLI.
3. Compare exact issue numbers/titles against repository evidence.
4. Repair one item at a time without deleting issues or bodies.
5. Re-run script self-tests and a no-write dry run only after a single adapter is reviewed.
6. Adopt that adapter in a focused commit and archive/deprecate alternatives.

Kanban adapter consolidation remains an In Review/Blocked Sprint 0 item because the
best candidate is protected untracked work.
