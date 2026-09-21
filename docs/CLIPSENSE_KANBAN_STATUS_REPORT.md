# ClipSense Kanban Status Report

## 1. Operating Rule

This board is the source of execution truth for ClipSense agile delivery.

Repository evidence can inform a card, but it does not make a card Done. A card is Done only after it is pulled, implemented or verified against its acceptance criteria, reviewed, and accepted during this delivery cycle.

The current agile goal is the stabilized ZIP MVP:

```text
ZIP upload through web
-> Go API stores source archive and creates batch
-> Redis job is queued
-> Python worker consumes job
-> safe ZIP extraction
-> supported videos are processed
-> clips/storyline records are written to Postgres
-> vectors are upserted to Qdrant
-> API/web display batch result
```

Long-term ClipSense intake still includes links, direct video files, and ZIP batches, but direct video upload and link ingestion stay in Icebox until the ZIP MVP is accepted.

## 2. Board Columns

| Column | Meaning |
| --- | --- |
| In Progress | The card currently being worked. Keep this narrow. |
| Ready | The next cards with clear acceptance criteria and no product-scope ambiguity. |
| Backlog | Ordered work that matters, but should not be pulled before Ready items. |
| Review | Work completed in the current cycle and waiting for acceptance. |
| Blocked | A pulled card that cannot progress because of an external dependency or environment failure. |
| Done | Accepted after current-cycle verification. Starts empty for this agile reset. |
| Icebox | Valid future work that must not distract from the ZIP MVP. |

## 3. Current Board

| Column | Cards |
| --- | --- |
| In Progress | `US-000` Transform Raw Creator Footage Into Story-Ready Output |
| Ready | `CS-101` Unified Docker Orchestration; `CS-102` Postgres Schema and Migration Baseline; `CS-107` Local JWT Auth and Protected Routes |
| Backlog | `CS-108`, `CS-109`, `CS-103`, `CS-111`, `CS-112`, `CS-201`, `CS-114`, `CS-110`, `CS-104`, `CS-203`, `CS-205`, `CS-204`, `CS-207`, `CS-301`, `CS-305`, `CS-306`, `CS-304`, `CS-405`, `CS-113`, `CS-115`, `CS-208`, `CS-117`, `CS-105`, `CS-106`, `CS-116`, `CS-119`, `CS-202`, `CS-206` |
| Review | Empty |
| Blocked | Empty until a pulled card is blocked during execution |
| Done | Empty until accepted in this agile cycle |
| Icebox | `US-011`, `US-012`, `US-013`, `US-014`, `US-015`, `US-016`, `US-017`, `US-018`, `CS-302`, `CS-303`, `CS-401`, `CS-402`, `CS-403`, `CS-404`, advanced scope of `CS-405` |

## 4. Immediate Board Order

1. `US-000` Define the core product outcome and MVP done gate.
2. `CS-101` Unified Docker orchestration.
3. `CS-102` Postgres schema and migration baseline.
4. `CS-107` Local JWT auth and protected routes.
5. `CS-108` Local browser CORS policy.
6. `CS-109` JWT secret configuration and production guardrails.
7. `CS-103` ZIP multipart upload intake.
8. `CS-111` Upload size and ZIP validation hardening.
9. `CS-112` API error response semantics.
10. `CS-201` Redis job publisher, worker consumer, and batch state machine.
11. `CS-114` Local Docker runtime repair and verification, pulled only if orchestration cannot be accepted without environment repair.
12. `CS-110` Worker Docker packaging.
13. `CS-104` Worker ZIP extraction and audio preparation.
14. `CS-203` Transcription integration.
15. `CS-205` Basic MVP clip classification.
16. `CS-204` Qdrant vector upsert.
17. `CS-207` Persist and display batch failure reasons.
18. `CS-301` MVP dashboard and batch review views.
19. `CS-305` Dashboard route structure.
20. `CS-306` Frontend auth hydration.
21. `CS-304` Frontend error handling.
22. `CS-405` Basic authenticated MVP export.
23. `CS-113` Dependency vulnerability cleanup.
24. `CS-115` CI build/test validation.
25. `CS-208` End-to-end ZIP MVP verification gate.
26. `CS-117` Stabilization documentation refresh.

## 5. Sprint Discipline

- Pull from Ready in order unless a blocker forces a small unblocker card.
- Do not pull Icebox work into the ZIP MVP sprint.
- Do not call a card Done from repo inspection alone.
- A card can move to Review only when its acceptance criteria have been executed or explicitly documented as blocked.
- A card can move to Done only after Review is accepted.
- If implementation discovers a defect inside the pulled card's acceptance criteria, fix it inside that card.
- If implementation discovers unrelated work, create or reference a separate backlog card.

## 6. ZIP MVP Done Gate

The stabilized ZIP MVP is Done only when these are accepted through board workflow:

- Docker Compose boots web, API, worker, Postgres, Redis, and Qdrant.
- Web loads at `http://localhost:3000`.
- Browser-facing API calls use `http://localhost:8080`.
- User can register and log in.
- Protected routes reject missing or invalid tokens.
- User can upload a ZIP containing supported video files.
- API stores the ZIP and creates a batch.
- Redis queues the worker job.
- Worker consumes the job.
- Worker safely extracts supported videos.
- Worker processes media without crashing the processing loop.
- Clips and storylines persist to Postgres.
- Qdrant vector upsert works or failure is clearly handled.
- Batch detail displays real results or a clear failure reason.
- Core build/test/security checks pass or residual risks are documented.
- Docs match accepted behavior.

## 7. Icebox Guardrail

These are intentionally not part of the current ZIP MVP sprint:

- Direct single-video upload.
- Public or user-provided link ingestion.
- Advanced narrative intelligence.
- Timeline workspace.
- Native video segment preview.
- Rust video engine.
- Cloud multipart uploads.
- Platform integrations.
- Enterprise workspaces.
- Editor-ready export packages beyond basic authenticated MVP export.
