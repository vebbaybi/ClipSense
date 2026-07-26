# ClipSense Service Ownership

Status: Current foundation boundary

## Components

| Component | Owns | Must not own |
|---|---|---|
| Web application | Browser UI, interaction state, invoking public API, presenting media intelligence | Database access, queue access, media processing, authoritative API models |
| Go API | Public HTTP API, authorization, intake boundary, public exports, orchestration requests | Transcription/model execution, direct UI state, ad hoc infrastructure provisioning |
| Python processor | Archive extraction, FFmpeg/media probing, transcription, analysis, embedding generation | Public authentication, user-facing HTTP API, independent product schema |
| PostgreSQL | Durable users, batches, clips, storylines, job/state metadata after migration | Media blobs, transient queue ownership |
| Redis | Current transient job transport | Durable workflow history or business truth |
| Qdrant | Current semantic vectors and search payloads | Relational ownership, authorization decisions |
| Media storage | Original uploads and generated media artifacts | User/account metadata |
| Repository operations | Compose, CI, doctor, deployment/backup adapters, Kanban adapter | Product runtime behavior |
| GitHub Issues/Project #19 | Backlog item identity, workflow status, evidence links | Runtime requirements source by itself |

## Allowed Communication

```text
web -> Go public API
Go API -> PostgreSQL
Go API -> processing orchestration/queue
Go API -> media storage
processor -> internal job contract
processor -> media storage
processor -> PostgreSQL (current violation tolerated pending internal contract)
processor -> Qdrant
operations -> service health and deployment interfaces
```

Direct browser access to PostgreSQL, Redis, Qdrant, worker internals, or media filesystem
paths is forbidden. Qdrant payloads must never be trusted for authorization.

## Current Violations

- Go and Python both write the product schema with handwritten SQL.
- The processing job is an unversioned JSON map shared implicitly across languages.
- Public API models are independently retyped in React pages.
- Compose exposes infrastructure ports and development credentials.
- The web export adapter bypasses the authenticated API client.
- Startup migration code makes the Go API both runtime service and schema migrator.

## Future Boundaries

A desktop application is not present. If approved later, it will own native file
selection, local project state, local job control, and OS integration. It will consume
shared public/internal contracts and must not copy cloud authorization or processing
rules.

A local processor is also not present. It should reuse the processing contract and
media rules while keeping cloud credentials and cloud-only orchestration outside the
desktop trust boundary.

No new service should be created until an existing component cannot own the
responsibility cleanly and an ADR records the operational cost.

