# ClipSense Complete Product Backlog

## 1. Document Purpose

This is the authoritative ClipSense backlog for the complete product journey from product foundation through functional prototype, functional MVP, production-ready MVP, stabilization, public production release, progressive post-MVP releases, mature product operation, maintenance, deprecation, and responsible retirement.

This document does not begin from the current repository state. The repository is evidence. Existing code, docs, scripts, and tests are mapped to the backlog item where that capability belongs in the professional development lifecycle, then assessed as verified, partial, hardening required, defective, missing validation, not implemented, or legacy.

`CSUS.md` is intended to guide product planning, issue creation, release boundaries, engineering validation, security and privacy hardening, accessibility, DevOps readiness, supportability, and long-term product evolution.

## 2. Product Vision

ClipSense is an AI-assisted pre-editing platform for creators, streamers, reaction creators, video teams, and editors who need to turn raw or disorganized footage into structured, analyzed, reviewable, and story-ready material before conventional editing begins.

### Consumer Problem

Creators collect hours of footage, clips, reaction videos, game captures, and recording sessions, but the costly work often happens before editing: finding what is in each clip, identifying the useful moments, understanding tone and topic, ordering clips into a story, and exporting a usable plan for an editor.

### Target Consumers

- Solo creators and streamers who need faster clip organization.
- Reaction and commentary creators managing many source clips.
- Editors who receive disorganized footage from creators or agencies.
- Small video teams that need shared review, correction, and export workflows.
- Future professional teams with repeatable intake, review, retention, support, and interoperability needs.

### Product Value

ClipSense should reduce manual review and organization time by converting input media into trustworthy transcripts, summaries, classifications, searchable metadata, storyline suggestions, correction controls, and exportable structures.

### Product Boundaries

ClipSense is a pre-editing intelligence layer. It is not a full non-linear editor, renderer, live broadcasting tool, rights-management system, or permanent media archive unless future evidence justifies those directions. Advanced infrastructure such as Kubernetes, Terraform, cloud object storage, billing, enterprise SSO, or editor plug-ins should be introduced only when product and operational need is proven.

### Key Terminology

- Source input: a ZIP archive, direct video file, or user-provided link submitted by a consumer.
- Batch: the tracked unit of source intake and processing.
- Source asset: the original uploaded or resolved input stored under defined retention rules.
- Clip: a video asset extracted or normalized from a source input.
- Transcript: text generated from clip audio, with future segment and correction support.
- Analysis: summaries, topics, roles, mood, embeddings, quality signals, and other derived metadata.
- Storyline: an ordered arrangement of clips intended to support a narrative or review workflow.
- Export: a consumer-usable output such as JSON, CSV, EDL, XML, or editor-oriented handoff data.
- Retention: the explicit lifecycle for source media, extracted media, transcripts, embeddings, exports, logs, and backups.

### Intended Full Product Direction

The mature product direction is a secure, accessible, operable, and reviewable AI pre-editor that supports multiple intake types, reliable large-batch processing, transparent AI analysis, creator correction, storyline construction, searchable media intelligence, professional exports, controlled account and project management, retention and deletion controls, operational resilience, model-quality evaluation, compatibility management, and responsible feature evolution.

## 3. Product Principles

- Deliver complete value in vertical increments that a creator or operator can verify.
- Prioritize consumer safety, privacy, and trust over speed of feature expansion.
- Build upload, media, authentication, authorization, and queue behavior securely from the start.
- Validate every increment with tests, runtime checks, and documented evidence proportional to risk.
- Automate quality gates and avoid release decisions that depend on memory or heroics.
- Keep current documentation separate from aspirational planning.
- Release progressively: stabilize before expanding the intake, AI, export, and collaboration surface.
- Avoid speculative architecture until there is a product or operational reason.
- Keep ClipSense operable, observable, recoverable, maintainable, and supportable.
- Treat AI output as assistive, reviewable, and fallible, not as unquestioned truth.

## 4. Verified Repository Assessment

### Existing Implementation

The repository currently contains a compact MVP-oriented implementation:

- Web UI in `apps/web` using Next.js 14, React, Tailwind, and client-side token storage.
- Go API in `apps/api/main.go` using chi, Postgres, Redis, JWT, bcrypt, multipart upload handling, batch CRUD, batch detail, and JSON/CSV export endpoints.
- Python worker in `apps/api/ai_worker` using Redis, Postgres, FFmpeg, Whisper, sentence-transformers, scikit-learn KMeans, and Qdrant.
- Docker Compose in `docker-compose.yml` for Postgres, Redis, Qdrant, API, worker, and web.
- CI in `.github/workflows/ci.yml` for Go tests, worker compile/tests, web build, and Compose config.
- Historical and aspirational planning in `CS.md`, `CS.txt`, `apps/app`, `apps/package/packages`, `apps/web/src/source`, `apps/api/apis`, `1807ish/roadmap.eos`, and `1807ish/whitepaper.block`.

### Working Capabilities

- Go API health endpoint reports database and Redis health as JSON and returns `503` when dependencies are unavailable.
- Registration and login issue JWTs and use bcrypt for password hashing.
- Auth middleware rejects missing or invalid bearer tokens.
- Authenticated users can list, create, retrieve, and export their own batches through API queries that filter by `user_id`.
- The API persists uploaded archives to a shared upload directory and enqueues Redis list jobs.
- Worker extracts supported video entries from ZIP files, extracts audio, transcribes, summarizes, classifies simple mood/role values, creates embeddings, clusters clips, creates a basic storyline, writes rows to Postgres, and upserts vectors to Qdrant.
- ZIP extraction has tests for supported videos, ignored non-video entries, path traversal, absolute paths, excessive file counts, and decompressed-size limits.

### Partial Capabilities and Confirmed Defects

- The web CSV export link points directly to the API without attaching the bearer token, so authenticated export from the browser is defective unless another auth mechanism is introduced.
- Browser tokens are stored in `localStorage`; logout exists in the hook but no visible logout control or session expiry handling exists.
- Upload validation relies on multipart parsing size and downstream ZIP extraction. There is no API-level content signature verification, MIME verification, malformed archive rejection path with safe user messaging, or persisted validation detail.
- Batch states are coarse (`pending`, `processing`, `complete`, `failed`) and lack per-stage progress, failure reasons, partial success, cancellation, retry, or poison-job states.
- Redis queue uses a list with `RPush` and worker `BLPOP`, without acknowledgement, visibility timeout, retry limit, or dead-letter handling.
- Worker marks failed batches but does not persist actionable failure reasons or per-clip failures.
- Inline database schema creation in `migrate()` is not a production migration system and lacks foreign keys, cascade rules, indexes, constraints, ownership enforcement at the database layer, or rollback.
- Worker and API both connect directly to the database, increasing coupling and making authorization and schema evolution harder.
- Qdrant is used for vector storage, but there is no search API, backup/snapshot routine, or vector lifecycle cleanup.
- Dockerfiles do not use non-root runtime users, reproducible dependency installation everywhere, health checks, or vulnerability scanning.
- Compose uses fixed development credentials and lacks production secret handling.
- Database backup/restore scripts target a SQLite file while the current Compose runtime uses Postgres.
- Monitoring and deployment scripts are placeholders or thin wrappers, not production-ready operations.
- Web accessibility is basic and unverified; forms rely on placeholders, status feedback is not consistently announced, and drag/drop needs keyboard and screen-reader validation.
- Documentation is contradictory: several blueprint files describe folders, route handlers, WebSockets, SDKs, OpenAPI, workspaces, and editor integrations that do not exist.

### Verification Performed

- `go test -mod=readonly ./...` in `apps/api`: passed.
- `python -m py_compile main.py zip_safety.py` in `apps/api/ai_worker`: passed.
- `python -m unittest discover -s tests` in `apps/api/ai_worker`: passed 7 tests.
- `go list -mod=readonly -m -u all` in `apps/api`: completed and showed available updates for chi, jwt, pgx, go-redis, `golang.org/x/*`, and other dependencies.
- `npm ci` in `apps/web`: attempted, emitted deprecated package warnings, then stalled with no completion; partial `node_modules` was removed.
- `npm audit --package-lock-only --json` in `apps/web`: attempted, stalled with no completion, and the audit process was stopped.
- `docker --version`: failed because Docker is not on PATH in this environment, so Compose config/build/runtime validation could not be completed locally.

### Current Technical Debt

Current debt clusters around security defaults, auth/session hardening, reliable queue semantics, state accuracy, validation depth, data lifecycle, migrations, observability, deployment hardening, documentation hygiene, dependency remediation, and evidence-based accessibility/AI-quality validation.

## 5. Complete Product Story Map

| Stage | Consumer and Engineering Journey | Backlog Coverage |
| --- | --- | --- |
| Foundation | Define consumers, product boundaries, terminology, repository environment, security/privacy/accessibility baselines, CI, and documentation truth. | Ranks 001-007 |
| Functional Prototype | Demonstrate the first vertical ZIP-to-result loop: account access, guidance, upload, storage, batch state, safe extraction, queue handoff, first analysis result, and accessible upload/status basics. This proves the concept but is not safe for public release. | Ranks 008-020 |
| Functional MVP | Complete the minimum creator journey: user-owned results, reviewable clips, vector-backed basic storyline, authenticated export, lifecycle baseline, failure messaging, and health visibility. This is usable for controlled internal validation but not publicly releasable. | Ranks 021-029 |
| Production-Ready MVP | Add the controls required for controlled real-user use: unsafe-default removal, session hardening, truthful partial states, reliable queueing, idempotency, accessibility validation, E2E tests, dependency remediation, hardened containers, production config, privacy/deletion, contracts, runtime validation, AI quality, migrations, backups, observability, quotas, cost controls, abuse prevention, and media-rights communication. | Ranks 030-049 |
| MVP Stabilization | Use controlled-release evidence to repair defects, tune limits, resolve support findings, and validate release gates before public launch. | Rank 050 |
| Public Production Release | Publish the production-ready MVP with release versioning, support runbooks, incident communication, service-status communication, rollback criteria, and consumer-facing limitation notes. | Rank 051 |
| Post-MVP Release 1 | Expand intake progressively: media compatibility, direct uploads, resumable uploads, and user-provided links. | Ranks 052-055 |
| Post-MVP Release 2 | Improve creator workflow and intelligence: transcript-linked playback/correction, configurable classification, editable storylines, accessibility regression, search, professional exports, and reprocessing. | Ranks 056-062 |
| Post-MVP Release 3 | Add account/workspace and collaboration controls only after single-user authorization and deletion are mature. | Ranks 063-064 |
| Mature Product | Manage model upgrades, compatibility, long-term maintainability, deprecation, migration, and responsible retirement. | Ranks 065-066 |

## 6. MVP Definition

### Functional Prototype

The functional prototype proves the ClipSense concept with a demonstrable ZIP intake path, account access, batch creation, safe extraction basics, worker processing, first analysis output, and visible status. It may be suitable for local development and internal demos only. It is not releasable to public or uncontrolled real users.

### Functional MVP

The functional MVP is the smallest complete creator journey with meaningful product value:

1. A creator can register, log in, and access a protected dashboard.
2. The app explains exactly what ZIP inputs are accepted.
3. The creator submits supported footage and receives immediate validation.
4. The system creates a user-owned batch, stores the source, safely extracts media, enqueues processing, and produces reviewable results.
5. The creator can see truthful status, review clips and generated metadata, inspect a basic storyline, export through an authenticated flow, and understand data lifecycle basics.

The functional MVP is not public-release-ready until production-ready controls below are completed.

### Production-Ready MVP

The production-ready MVP is the functional MVP plus all release-blocking controls for controlled real-user use: secure defaults, hardened browser session behavior, authenticated export, truthful failure and partial-success state, reliable queue acknowledgement/retry/dead-letter handling, idempotent processing, retention and deletion controls, accessibility validation, dependency remediation, container hardening, production configuration, API contracts, runtime validation, AI quality evaluation, migrations, backup/recovery, observability, usage and storage limits, processing cost controls, abuse prevention, media-rights communication, and end-to-end release evidence.

### MVP Stabilization

MVP stabilization happens after production-ready MVP validation or controlled release. It is not a license to postpone known release blockers. It is the stage for issues discovered through real runtime evidence, creator support, incident review, telemetry, and release-gate findings.

### Public Production Release

Public production release is the production-ready MVP plus release versioning, support runbooks, rollback criteria, incident communication, service-status communication, privacy and acceptable-use communication, and operational ownership.

Direct video upload, link intake, resumable uploads, collaboration, billing, editor plug-ins, advanced narrative intelligence, and production-scale infrastructure remain post-MVP unless they become necessary to satisfy the production-ready MVP boundaries above.

## 7. Full Product Definition

The mature ClipSense platform should support robust multi-source intake, large-batch processing, accurate and transparent AI-assisted analysis, creator review and correction, advanced storyline construction, searchable media intelligence, professional export interoperability, secure account/project/workspace management, scalable processing, privacy and data-lifecycle controls, mature accessibility, observability, resilience, disaster recovery, model evaluation, safe feature evolution, compatibility management, supportability, and long-term maintenance.

This is a direction, not a claim that every advanced feature is already justified. Each post-MVP capability must be introduced only when it improves creator value, reduces operational risk, or enables a release objective.

## 8. Ranked Backlog

## Product Vision and Consumer Boundaries

**Business Rank:** 001  
**Release Stage:** Product Foundation  
**Fibonacci Estimate:** 3  
**Current Implementation Assessment:** Partially Documented; Requires Consolidation

### Purpose

Define the consumer, problem, product boundaries, value proposition, success measures, and non-goals that every future ClipSense increment must trace to.

### Consumer, Business, or Risk-Reduction Value

This enables every later story by preventing the product from drifting into a generic editor, storage service, or speculative infrastructure platform.

### Repository Evidence

`CS.md`, `docs/MVP_STABILIZATION.md`, `1807ish/roadmap.eos`, and `1807ish/whitepaper.block` describe ClipSense as a pre-editor for creators, but some documents mix current implementation with future architecture.

### Scope and Required Behaviour

Document target consumers, core jobs to be done, supported and unsupported workflows, MVP value, future value hypotheses, and measurable success signals such as time-to-first-result, successful batch completion, export usefulness, and correction rate.

### Security, Privacy, Accessibility, or Operational Requirements

The foundation must state that media, transcripts, embeddings, and exports are sensitive creator data and that every consumer journey must be accessible and recoverable.

### Dependencies

None.

### Acceptance Criteria

- [ ] **Given** a new contributor reads the product foundation, **when** they plan a feature, **then** they can identify the target consumer and outcome it serves.
- [ ] **Given** a proposed feature is outside pre-editing, **when** it is reviewed, **then** the document shows whether it is out of scope, post-MVP, or a justified future extension.
- [ ] **Given** current and aspirational docs disagree, **when** planning uses them, **then** current implementation evidence is clearly separated from intent.

### Validation Evidence

- [ ] Product foundation reviewed against `CS.md`, `docs/MVP_STABILIZATION.md`, and current routes/API files.
- [ ] Success measures and non-goals captured in authoritative docs.

### Out of Scope

Detailed UI design, API contracts, or implementation.

### Definition of Ready

- [ ] Existing product notes have been inventoried.
- [ ] The intended MVP consumer journey has an agreed owner.

### Definition of Done

- [ ] Product boundaries are documented and linked from repository entry points.
- [ ] No contradictory roadmap file is treated as current implementation truth.

## Authoritative Vocabulary and Roadmap Hygiene

**Business Rank:** 002  
**Release Stage:** Product Foundation  
**Fibonacci Estimate:** 3  
**Current Implementation Assessment:** Legacy or Obsolete

### Purpose

Replace fragmented blueprint files with an explicit documentation hierarchy that distinguishes current state, product intent, backlog, architecture decisions, and historical artifacts.

### Consumer, Business, or Risk-Reduction Value

This reduces planning errors where non-existent files, SDKs, routes, WebSockets, object storage, or editor integrations are mistaken for shipped functionality.

### Repository Evidence

`CS.txt`, `apps/app`, `apps/package/packages`, `apps/web/src/source`, and `apps/api/apis` are extensionless structure blueprints. They mention many absent folders and endpoints. Placeholder files in `apps/web/src/components`, `features`, `stores`, `styles`, and `types` also explain that implementation is inline for MVP.

### Scope and Required Behaviour

Create a docs index, mark legacy blueprints as historical, define accepted terminology, and keep future roadmap decisions separate from verified implementation notes.

### Dependencies

Product Vision and Consumer Boundaries.

### Acceptance Criteria

- [ ] **Given** a repository reader opens a blueprint file, **when** it describes absent code, **then** it is clearly marked as historical or aspirational.
- [ ] **Given** a backlog item references a capability, **when** repository evidence is cited, **then** the citation points to actual files or explicitly says not implemented.
- [ ] **Given** terminology changes, **when** docs are updated, **then** batch, clip, source asset, transcript, analysis, storyline, export, retention, and deletion remain consistent.

### Validation Evidence

- [ ] Documentation inventory confirms no legacy structure file is the only source of truth.
- [ ] Vocabulary appears consistently in product, API, worker, and UI documentation.

### Out of Scope

Deleting historical files before a migration plan is approved.

### Definition of Ready

- [ ] Legacy documentation files are listed with current relevance.
- [ ] Current implementation owners agree on doc categories.

### Definition of Done

- [ ] Repository docs distinguish current state, roadmap, and archive.
- [ ] Backlog references use current evidence instead of obsolete folder trees.

## Security and Privacy Baseline

**Business Rank:** 003  
**Release Stage:** Product Foundation  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### Purpose

Establish baseline security and privacy expectations before media intake, authentication, processing, exports, and retention expand.

### Consumer, Business, or Risk-Reduction Value

Creators upload private footage and generated transcripts. A breach, exposed token, leaked archive, or misleading deletion behavior would damage trust before the product proves value.

### Repository Evidence

`apps/api/main.go` has bcrypt password hashing, JWT issuance, bearer auth, user-filtered batch queries, CORS allowlist logic, and a `dev-secret` default. `docker-compose.yml` uses development database credentials. No privacy notice, deletion policy, rate limits, CSRF/session strategy, or structured secret validation exists.

### Scope and Required Behaviour

Define secure defaults, auth/session requirements, password requirements, secret management, upload threat model, authorization model, safe error handling, log data restrictions, retention categories, deletion expectations, dependency security, and release-blocking security gates.

### Security, Privacy, Accessibility, or Operational Requirements

Use OWASP API, file upload, authentication, and session guidance; use NIST SSDF and Privacy Framework principles for secure development and data lifecycle.

### Dependencies

Product Vision and Consumer Boundaries.

### Acceptance Criteria

- [ ] **Given** a shared or production environment starts, **when** required secrets are missing or development defaults are present, **then** startup fails with a safe configuration error.
- [ ] **Given** media, transcript, embedding, or export data is stored, **when** retention is reviewed, **then** ownership, purpose, retention, deletion, and backup behavior are documented.
- [ ] **Given** an API error occurs, **when** a response is returned, **then** the user sees a safe message and logs omit secrets, tokens, private media contents, and unnecessary personal data.

### Validation Evidence

- [ ] Threat model covers upload, auth, queue, worker, database, vector store, export, and logs.
- [ ] Security checklist is enforced before production-ready MVP work and public production release.

### Out of Scope

Enterprise compliance certifications and paid security tooling selection.

### Definition of Ready

- [ ] Required data categories and trust boundaries are identified.
- [ ] Development and production environments are distinguished.

### Definition of Done

- [ ] Baseline controls are documented with release gates.
- [ ] Known unsafe defaults are linked to specific hardening backlog items.

## Accessibility Baseline

**Business Rank:** 004  
**Release Stage:** Product Foundation  
**Fibonacci Estimate:** 3  
**Current Implementation Assessment:** Missing Validation

### Purpose

Define accessibility expectations for authentication, upload, status, results review, correction, export, settings, errors, and support paths.

### Consumer, Business, or Risk-Reduction Value

Creators and editors must be able to use ClipSense with keyboard navigation, assistive technology, understandable status updates, visible focus, and accessible error recovery.

### Repository Evidence

`apps/web/src/app/page.tsx` uses placeholders instead of explicit labels. `apps/web/src/app/(dashboard)/batches/new/page.tsx` uses drag/drop with limited keyboard-specific guidance. No accessibility tests, axe checks, or WCAG conformance notes exist.

### Scope and Required Behaviour

Adopt WCAG 2.2 AA as the target baseline; define form labels, focus management, keyboard operation, status announcement, error identification, contrast, responsive behavior, reduced motion, and accessible data views.

### Dependencies

Product Vision and Consumer Boundaries.

### Acceptance Criteria

- [ ] **Given** a keyboard-only user signs in and uploads a batch, **when** they navigate the flow, **then** every control is reachable and has a visible focus state.
- [ ] **Given** processing state changes, **when** assistive technology is active, **then** meaningful status is announced without requiring a visual-only cue.
- [ ] **Given** validation fails, **when** an error appears, **then** it is associated with the field or control that needs attention.

### Validation Evidence

- [ ] WCAG 2.2 AA checklist exists for MVP routes.
- [ ] Accessibility checks are included in UI and E2E validation planning.

### Out of Scope

Full third-party accessibility audit before the first functional increment.

### Definition of Ready

- [ ] MVP screens and interaction types are listed.
- [ ] Accessibility target and test approach are agreed.

### Definition of Done

- [ ] Baseline accessibility requirements are documented.
- [ ] Every UI backlog item includes relevant accessibility acceptance criteria.

## Repository and Environment Baseline

**Business Rank:** 005  
**Release Stage:** Product Foundation  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### Purpose

Make local development, configuration, dependency installation, and service startup reproducible enough for contributors and CI.

### Consumer, Business, or Risk-Reduction Value

Reliable setup shortens time to fix product defects and reduces hidden environment drift before the first user-visible increment.

### Repository Evidence

`.env.example`, `docker-compose.yml`, service Dockerfiles, `apps/web/package-lock.json`, `apps/api/go.mod`, `apps/api/go.sum`, and `apps/api/ai_worker/requirements.txt` exist. Scripts under `scripts/dev` wrap Compose. No root README exists, Docker is required for full runtime, and web dependency installation stalled in this environment.

### Scope and Required Behaviour

Document prerequisites, supported OS assumptions, required tool versions, environment variables, service ports, model download behavior, local data paths, and clean reset behavior.

### Dependencies

Authoritative Vocabulary and Roadmap Hygiene.

### Acceptance Criteria

- [ ] **Given** a new contributor follows setup docs, **when** prerequisites are installed, **then** they can start the stack or understand exactly which dependency blocks it.
- [ ] **Given** environment variables are missing, **when** a service starts, **then** development defaults are allowed only for local use and production-required variables are explicit.
- [ ] **Given** dependency installation fails or stalls, **when** troubleshooting is needed, **then** the docs identify cache, registry, and platform checks.

### Validation Evidence

- [ ] Setup docs verified against current files and Compose service names.
- [ ] Environment variable reference matches `.env.example`, Compose, API, worker, and web usage.

### Out of Scope

Cloud deployment or production secrets.

### Definition of Ready

- [ ] Tool versions and service dependencies are known.
- [ ] Current setup gaps have repository evidence.

### Definition of Done

- [ ] Local setup is reproducible or clearly blocked by documented external prerequisites.
- [ ] No setup step depends on stale blueprint-only files.

## Testing Strategy Baseline

**Business Rank:** 006  
**Release Stage:** Product Foundation  
**Fibonacci Estimate:** 3  
**Current Implementation Assessment:** Partially Implemented; Missing Coverage

### Purpose

Define the layered test strategy for API, worker, UI, integration, end-to-end, security, accessibility, media processing, queue behavior, AI quality, migrations, containers, and release validation.

### Consumer, Business, or Risk-Reduction Value

Creators need a dependable product; early automated validation prevents silent regressions in upload, analysis, export, and data lifecycle behavior.

### Repository Evidence

`apps/api/main_test.go` covers health dependency failure, CORS preflight, missing/invalid auth, and missing upload file. `apps/api/ai_worker/tests/test_zip_safety.py` covers ZIP safety. `.github/workflows/ci.yml` runs Go tests, worker compile/tests, web build, and Compose config. No UI tests, E2E tests, integration tests with Postgres/Redis/Qdrant, security tests, accessibility tests, migration tests, or Docker runtime tests exist.

### Scope and Required Behaviour

Define required test layers, fixtures, media samples, malformed inputs, CI gates, local commands, risk-based coverage, and release evidence.

### Dependencies

Repository and Environment Baseline.

### Acceptance Criteria

- [ ] **Given** a new backlog item is prepared, **when** its risk is assessed, **then** it names the smallest adequate validation layer.
- [ ] **Given** upload or processing code changes, **when** CI runs, **then** malformed-input and failure-path tests are required.
- [ ] **Given** production release is considered, **when** quality gates are checked, **then** unit-only evidence is insufficient.

### Validation Evidence

- [ ] Test strategy maps existing tests to risk areas.
- [ ] Missing test layers are represented by ranked backlog items.

### Out of Scope

Writing every future test in the foundation stage.

### Definition of Ready

- [ ] Existing tests and CI jobs have been inventoried.
- [ ] High-risk flows are identified.

### Definition of Done

- [ ] Test strategy is documented with required evidence by release stage.
- [ ] Gaps are traceable to backlog ranks.

## CI and Documentation Foundation

**Business Rank:** 007  
**Release Stage:** Product Foundation  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### Purpose

Provide a reliable baseline for automated checks and contributor-facing documentation without pretending CI is production readiness.

### Consumer, Business, or Risk-Reduction Value

CI catches broken increments before they reach creators, while accurate docs reduce operational confusion.

### Repository Evidence

`.github/workflows/ci.yml` exists and runs service checks. `docs/MVP_STABILIZATION.md` documents current stabilization. No root README, architecture decision records, API docs, operational runbooks, release checklist, or dependency-security workflow exists.

### Scope and Required Behaviour

Add root documentation entry points, CI status expectations, required checks, dependency cache strategy, workflow permissions, artifact handling, and a distinction between build checks and release approval.

### Security, Privacy, Accessibility, or Operational Requirements

GitHub Actions should use least permissions, pinned or reviewed actions, secret safety, and clear artifact retention rules.

### Dependencies

Testing Strategy Baseline.

### Acceptance Criteria

- [ ] **Given** a pull request changes API, worker, web, Compose, or docs, **when** CI runs, **then** the relevant checks execute and failures are visible.
- [ ] **Given** a contributor starts at repository root, **when** they need setup, architecture, testing, or release guidance, **then** the root docs route them to current information.
- [ ] **Given** a workflow runs, **when** permissions are inspected, **then** it uses the least privileges needed for its jobs.

### Validation Evidence

- [ ] CI workflow reviewed for service coverage and permission hardening.
- [ ] Documentation index links only to current or clearly historical docs.

### Out of Scope

Full production deployment automation.

### Definition of Ready

- [ ] Current CI jobs and docs are inventoried.
- [ ] Required checks for MVP are known.

### Definition of Done

- [ ] CI and docs support first functional increment development.
- [ ] Known CI gaps are captured in later ranked items.

## Creator Account Access and Recovery

**Business Rank:** 008  
**Release Stage:** Functional Prototype  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### User Story

**As a** creator  
**I need** to create, verify, recover, and sign in to an account  
**So that** my batches and generated results are separated from other users without locking me out permanently.

### Product Outcome

A creator can register, verify account ownership where required, log in, recover access safely, receive an authenticated session, and access the dashboard.

### Business Value

Account access is the first privacy boundary and enables every user-owned media workflow.

### Repository Evidence

`apps/api/main.go` implements `/api/auth/register` and `/api/auth/login` with bcrypt and JWTs. `apps/web/src/app/page.tsx` provides login/register UI. There is no password policy, email verification, account recovery, duplicate email handling UX, visible logout, session expiration messaging, or rate protection.

### Functional Requirements and Business Rules

Require valid email format, minimum password rules, duplicate account handling, optional email verification for public release, safe password reset, login failure normalization, token expiry handling, and a visible session control.

### Consumer Safety and Trust

Do not reveal whether an email exists during login or recovery. Explain session expiration and recovery outcomes without exposing token details.

### Data and State Requirements

User records must have unique emails, password hashes only, created timestamps, verification/recovery state where enabled, and future support for account deletion and downloadable personal data.

### Failure and Fallback Behaviour

Invalid credentials and recovery requests return safe, consistent messages. Duplicate registration returns a user-actionable message. API or email delivery unavailability is shown as a service issue.

### Edge Cases

Invalid JSON, empty email, weak password, duplicate email, unverified email where verification is required, expired token, expired recovery token, network failure, and browser refresh after login.

### Security and Privacy Requirements

Use bcrypt or approved password hashing, safe error messages, token expiration, single-use recovery tokens, no password or recovery-token logging, and rate protection in hardening.

### Accessibility Requirements

Inputs require labels, error association, keyboard submission, visible focus, and screen-reader announcement of form errors.

### Performance and Reliability Requirements

Auth and recovery responses should be quick enough for interactive use, resilient to dependency errors, and consistent enough to avoid account enumeration.

### Observability and Operational Requirements

Log auth and recovery failures as counters without recording passwords, reset tokens, verification tokens, or full JWTs.

### Assumptions

Email/password is sufficient for MVP; SSO is post-MVP unless product need changes. Email verification and recovery are required before broad public account creation, not necessarily for a local prototype.

### Dependencies

Security and Privacy Baseline; Accessibility Baseline.

### Acceptance Criteria

- [ ] **Given** a creator submits a valid new email and password, **when** registration succeeds, **then** a token is issued and the dashboard loads.
- [ ] **Given** a creator submits invalid credentials, **when** login fails, **then** the response is safe and does not reveal whether the email exists.
- [ ] **Given** a creator starts account recovery, **when** the email exists or does not exist, **then** the user-facing response remains consistent and any valid recovery token is single-use and time-limited.
- [ ] **Given** public registration requires verification, **when** an unverified creator tries protected actions, **then** the product clearly explains the verification requirement without exposing internal token state.
- [ ] **Given** a token expires, **when** the creator opens a protected page, **then** they are redirected to sign in with an understandable message.
- [ ] **Given** a keyboard-only creator uses the form, **when** they tab and submit, **then** every control and error is accessible.

### Validation Evidence

- [ ] Unit tests cover registration validation, optional verification, login success, duplicate email, invalid credentials, recovery request, single-use recovery token, and expired token.
- [ ] UI or E2E tests cover sign-in, registration, verification or recovery messaging, error display, and logout.
- [ ] Security test confirms passwords, JWTs, verification tokens, and recovery tokens are not logged.

### Out of Scope

OAuth, SSO, multi-factor authentication, and teams.

### Definition of Ready

- [ ] Password and email validation rules are agreed.
- [ ] Session storage strategy is selected for the MVP stage.

### Definition of Done

- [ ] Account access works end to end with safe errors.
- [ ] Acceptance criteria and auth tests pass.
- [ ] No known release-blocking account access defect remains.

## Authenticated API Boundary

**Business Rank:** 009  
**Release Stage:** Functional Prototype  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### Purpose

Protect batch, clip, storyline, export, and future data APIs behind a consistent authorization boundary.

### Consumer, Business, or Risk-Reduction Value

Creators must not see or export another user's footage, transcripts, analysis, or metadata.

### Repository Evidence

`authMiddleware` in `apps/api/main.go` validates HMAC JWTs and stores `user_id` in request context. Batch list/detail/export queries filter by `user_id`. There are no table-level foreign keys, no authorization tests for cross-user access, and token parsing assumes `jwt.MapClaims` after parse.

### Scope and Required Behaviour

Normalize bearer token parsing, validate claim presence and expiration, require ownership checks on every resource, add cross-user negative tests, and document authorization rules per endpoint.

### Security, Privacy, Accessibility, or Operational Requirements

Return `401` for unauthenticated requests and `404` or safe `403` for unauthorized resource access without leaking resource existence.

### Dependencies

Creator Account Access and Recovery.

### Acceptance Criteria

- [ ] **Given** no bearer token is supplied, **when** a protected API is called, **then** the response is `401` and no handler side effect occurs.
- [ ] **Given** user A requests user B's batch, **when** the ID exists, **then** user A cannot infer private batch data.
- [ ] **Given** a malformed or unexpected JWT claim set is supplied, **when** middleware parses it, **then** the request is rejected safely.

### Validation Evidence

- [ ] API tests cover missing token, invalid token, expired token, malformed claims, and cross-user access.
- [ ] Endpoint authorization matrix is documented.

### Out of Scope

Workspace roles and shared access.

### Definition of Ready

- [ ] Protected endpoint inventory is complete.
- [ ] Ownership rules are defined for batches, clips, storylines, exports, and future assets.

### Definition of Done

- [ ] Authorization checks exist for every protected resource.
- [ ] Negative authorization tests pass.

## Supported ZIP Batch Guidance

**Business Rank:** 010  
**Release Stage:** Functional Prototype  
**Fibonacci Estimate:** 3  
**Current Implementation Assessment:** Partially Implemented; Missing Validation

### User Story

**As a** creator  
**I need** clear guidance on the first supported ZIP input  
**So that** I know what ClipSense can accept before I upload footage.

### Product Outcome

The UI and docs explain ZIP requirements, accepted video extensions, limits, unsupported cases, and what happens to ignored files.

### Business Value

Clear guidance reduces failed uploads and builds trust before processing begins.

### Repository Evidence

`apps/web/src/app/(dashboard)/batches/new/page.tsx` says "Drop a .zip of clips here". `zip_safety.py` allows `.mp4`, `.mov`, `.mkv`, `.webm`, and `.avi`, limits file count to 200 and extracted bytes to 10 GB. The UI does not display those limits.

### Functional Requirements and Business Rules

Show accepted archive type, video extensions, maximum file count, maximum upload size, extracted-size behavior, ignored non-video entries, and privacy/retention summary.

### Consumer Safety and Trust

Do not imply all media formats or codecs are supported. Warn that AI analysis may be imperfect.

### Data and State Requirements

Guidance must match API and worker constants or be sourced from a shared capability endpoint.

### Failure and Fallback Behaviour

Unsupported input guidance should appear before upload and again in validation errors.

### Edge Cases

Mixed media and non-media files, nested folders, duplicate filenames, huge archives, corrupted archives, and unsupported codecs.

### Security and Privacy Requirements

Guidance must not encourage password-protected or encrypted archives unless support exists.

### Accessibility Requirements

Guidance must be visible text associated with the upload control, not only toast messages.

### Performance and Reliability Requirements

Limits should reflect actual processing capacity and be updated as capacity changes.

### Observability and Operational Requirements

Track validation failure categories without logging filenames unnecessarily.

### Assumptions

ZIP remains the first MVP intake type.

### Dependencies

Product Vision and Consumer Boundaries.

### Acceptance Criteria

- [ ] **Given** a creator opens the new batch page, **when** they inspect requirements, **then** accepted extensions, limits, and unsupported cases are visible.
- [ ] **Given** a ZIP includes ignored non-video files, **when** processing completes, **then** the creator can understand that non-video entries were ignored.
- [ ] **Given** implementation limits change, **when** the UI renders guidance, **then** it does not show stale constants.

### Validation Evidence

- [ ] UI test verifies guidance text and accessible association.
- [ ] API/worker capability contract test verifies guidance matches enforcement.

### Out of Scope

Direct video upload and link intake.

### Definition of Ready

- [ ] Current ZIP limits are confirmed across API and worker.
- [ ] Product wording is approved for unsupported cases.

### Definition of Done

- [ ] Guidance is visible, accurate, accessible, and tested.
- [ ] No mismatch remains between guidance and enforcement.

## Batch Schema and State Model

**Business Rank:** 011  
**Release Stage:** Functional Prototype  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### Purpose

Define durable batch, clip, storyline, job, and processing-state models before the first upload flow depends on them.

### Consumer, Business, or Risk-Reduction Value

Accurate state prevents creators from believing a batch is complete when files failed, the queue stalled, or results are partial.

### Repository Evidence

`migrate()` in `apps/api/main.go` creates `users`, `batches`, `clips`, `storylines`, and `storyline_clips` tables inline. Batch status is free-form text. There are no foreign keys, job records, per-stage events, failure reasons, or indexes beyond primary keys.

### Scope and Required Behaviour

Define allowed states, state transitions, ownership, timestamps, clip counts, duration aggregation, failure reason fields, partial success, retry state, cancellation state, and database constraints.

### Security, Privacy, Accessibility, or Operational Requirements

State must not expose private file paths to unauthorized users. Operational state should support safe support diagnosis.

### Dependencies

Authenticated API Boundary.

### Acceptance Criteria

- [ ] **Given** a batch is created, **when** it enters processing, **then** allowed state transitions are enforced.
- [ ] **Given** a worker fails one clip, **when** remaining clips succeed, **then** the batch can represent partial success without false completion.
- [ ] **Given** a batch belongs to a user, **when** clips and storylines are persisted, **then** referential integrity prevents orphaned or cross-user data.

### Validation Evidence

- [ ] Migration or schema tests cover constraints and allowed states.
- [ ] API tests cover state transitions and failure state serialization.

### Out of Scope

Advanced workflow orchestration and multi-worker scheduling.

### Definition of Ready

- [ ] MVP states and transitions are agreed.
- [ ] Current schema limitations are documented.

### Definition of Done

- [ ] Batch state model is explicit, constrained, and tested.
- [ ] UI, API, and worker use the same state vocabulary.

## Secure ZIP Batch Upload

**Business Rank:** 012  
**Release Stage:** Functional Prototype  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### User Story

**As a** creator  
**I need** to upload a supported ZIP batch safely  
**So that** ClipSense can start organizing footage without risking my account or the service.

### Product Outcome

A creator submits one supported ZIP archive, receives immediate validation, and a user-owned batch is created only when the upload is acceptable and enqueueable.

### Business Value

This is the first core product action and the highest early security-risk surface.

### Repository Evidence

`createBatch` parses multipart form data with `200 << 20`, stores the uploaded file as `{batchID}.zip`, inserts a `pending` batch, and enqueues Redis. It does not verify ZIP signature, normalize original filenames beyond storage name, detect corrupted archives before persistence, use `MaxBytesReader`, or clean up the stored archive if database insertion fails after file creation.

### Functional Requirements and Business Rules

Accept exactly one archive in the `file` field; validate request size, archive type, content signature, and user ownership; create a batch only if storage and queue operations produce a consistent result.

### Consumer Safety and Trust

Reject unsafe archives with clear user-facing messages. Do not expose server paths or internal stack details.

### Data and State Requirements

Store source archive with a generated ID path, retain original filename as metadata, record validation outcome, and clean up orphan files on failure.

### Failure and Fallback Behaviour

If storage, database insert, or queue enqueue fails, return a safe failure and leave no misleading `pending` batch.

### Edge Cases

Missing file, multiple files, wrong extension, bad ZIP header, empty archive, upload interruption, huge multipart body, duplicate original names, and queue outage.

### Security and Privacy Requirements

Follow OWASP file upload controls: allowlist types, size limits, server-side validation, safe filenames, storage outside web root, malware scanning when justified, and resource limits.

### Accessibility Requirements

Upload control must support keyboard selection, visible status, and accessible error messages.

### Performance and Reliability Requirements

Validation should fail early for unsupported archives and avoid reading unbounded data into memory.

### Observability and Operational Requirements

Emit structured events for upload accepted, validation rejected, storage failed, and enqueue failed without logging private media content.

### Assumptions

MVP uses local shared storage under Compose; cloud object storage is post-MVP unless required for deployment.

### Dependencies

Supported ZIP Batch Guidance; Batch Schema and State Model.

### Acceptance Criteria

- [ ] **Given** a supported ZIP within size limits, **when** the creator uploads it, **then** a user-owned batch is created and enqueued exactly once.
- [ ] **Given** a non-ZIP or corrupted archive is uploaded, **when** validation runs, **then** no processable batch is created and the creator sees a safe reason.
- [ ] **Given** Redis enqueue fails, **when** upload storage and database insert already occurred, **then** batch state and stored files are reconciled without leaving a stuck pending batch.
- [ ] **Given** an unauthenticated request uploads a file, **when** it reaches the API, **then** no file is stored and `401` is returned.

### Validation Evidence

- [ ] API unit tests cover missing file, wrong content type, bad ZIP, over-size body, and enqueue failure.
- [ ] Integration test verifies stored file cleanup on database or queue failure.
- [ ] Accessibility test verifies keyboard upload and error association.

### Out of Scope

Direct uploads, resumable uploads, virus scanning service selection, and cloud object storage.

### Definition of Ready

- [ ] ZIP validation rules and limits are agreed.
- [ ] Storage cleanup behavior is designed.
- [ ] Batch state transition contract is available.

### Definition of Done

- [ ] Upload creates consistent storage, database, and queue state.
- [ ] Security and failure acceptance criteria pass.
- [ ] Documentation states supported archive behavior accurately.

## Local Source Storage Baseline

**Business Rank:** 013  
**Release Stage:** Functional Prototype  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### Purpose

Establish safe local storage behavior for source archives and processing files during MVP development.

### Consumer, Business, or Risk-Reduction Value

Creators need assurance that uploaded footage is stored under ownership, access, retention, and cleanup rules, even in local MVP deployments.

### Repository Evidence

`UPLOAD_DIR` defaults to `data/uploads` in the API and `/data/uploads` in Compose. Worker `PROCESS_DIR` defaults to `data/processing` or `/data/processing`. `docker-compose.yml` shares `clipsense-data` between API and worker. No retention, cleanup, encryption-at-rest, disk usage guard, or permission policy exists.

### Scope and Required Behaviour

Define storage directory permissions, generated object naming, metadata, cleanup on failure, processing workspace cleanup, disk usage limits, and separation between source and derived artifacts.

### Security, Privacy, Accessibility, or Operational Requirements

Media files must not be web-served directly, logged, or exposed through exports without authorization.

### Dependencies

Secure ZIP Batch Upload.

### Acceptance Criteria

- [ ] **Given** a source archive is stored, **when** the path is generated, **then** it cannot be influenced by the user-supplied filename.
- [ ] **Given** processing completes or fails, **when** retention rules apply, **then** temporary processing files are cleaned or retained only for documented diagnosis windows.
- [ ] **Given** disk usage approaches a configured threshold, **when** new uploads arrive, **then** the API fails safely before exhausting the host.

### Validation Evidence

- [ ] Storage tests cover generated names, cleanup, and path confinement.
- [ ] Operational check documents upload and processing volume usage.

### Out of Scope

Cloud object storage and cross-region replication.

### Definition of Ready

- [ ] Storage paths and ownership rules are defined.
- [ ] Retention expectations for MVP are known.

### Definition of Done

- [ ] Local storage behavior is documented, constrained, and validated.
- [ ] No user-controlled path reaches filesystem writes.

## Batch Submission and Tracking

**Business Rank:** 014  
**Release Stage:** Functional Prototype  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Missing Validation

### User Story

**As a** creator  
**I need** each accepted upload tracked as a batch  
**So that** I can return to its status and results later.

### Product Outcome

Accepted uploads create durable batch records visible in the dashboard and detail page.

### Business Value

Tracking turns upload into a product workflow instead of a one-shot processing request.

### Repository Evidence

`listBatches`, `createBatch`, and `getBatch` exist in `apps/api/main.go`. The dashboard fetches `/api/batches`, and the new batch page redirects to `/dashboard/batches/{id}`. There are no integration tests for list/detail ownership, empty states, or post-upload navigation.

### Functional Requirements and Business Rules

Batches must have stable IDs, names, ownership, state, created/updated timestamps, clip counts, duration, and user-visible status labels.

### Consumer Safety and Trust

Status must not imply processing has started if the job is not enqueued or the worker is unavailable.

### Data and State Requirements

Batch list and detail responses must exclude server file paths unless needed for internal operators.

### Failure and Fallback Behaviour

If dashboard fetch fails, show a retryable error. If a batch ID is unknown or unauthorized, show a safe not-found path.

### Edge Cases

Empty batch list, deleted batch, failed batch, large batch name, duplicate names, and stale browser token.

### Security and Privacy Requirements

All batch queries are scoped to the current user.

### Accessibility Requirements

Batch cards must be keyboard reachable links with understandable names and statuses.

### Performance and Reliability Requirements

Batch list should paginate or limit results once history grows.

### Observability and Operational Requirements

Track batch creation count, list errors, detail errors, and status distribution.

### Assumptions

Single-user ownership is sufficient for MVP.

### Dependencies

Batch Schema and State Model; Authenticated API Boundary.

### Acceptance Criteria

- [ ] **Given** an accepted upload, **when** the creator opens the dashboard, **then** the new batch appears with name, status, timestamp, and counts.
- [ ] **Given** a creator opens a batch detail URL, **when** the batch belongs to them, **then** batch metadata is shown.
- [ ] **Given** another user's batch ID is requested, **when** the API handles the request, **then** private data is not returned.
- [ ] **Given** the batch list is empty, **when** the dashboard loads, **then** an accessible empty state points to new batch creation.

### Validation Evidence

- [ ] API integration tests cover create, list, detail, not-found, and unauthorized cases.
- [ ] UI tests cover empty, pending, processing, complete, and failed list states.

### Out of Scope

Project folders, search, sharing, and collaboration.

### Definition of Ready

- [ ] Batch response shape is agreed.
- [ ] Dashboard state labels are defined.

### Definition of Done

- [ ] Batch tracking works across upload, dashboard, and detail views.
- [ ] Ownership and empty-state criteria are tested.

## Upload Request Size and Type Enforcement

**Business Rank:** 015  
**Release Stage:** Functional Prototype  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### Purpose

Enforce request and archive limits at the earliest safe boundary to prevent memory, disk, and worker exhaustion.

### Consumer, Business, or Risk-Reduction Value

Creators get fast, understandable feedback while the service avoids avoidable outages from oversized or malformed inputs.

### Repository Evidence

`createBatch` uses `r.ParseMultipartForm(200 << 20)`. Worker ZIP extraction limits decompressed video bytes and file count. There is no `http.MaxBytesReader`, no shared constant endpoint, no upload timeout, and no API-level ZIP member preflight.

### Scope and Required Behaviour

Add explicit max request body size, upload timeout, content-type and magic-number validation, archive preflight, consistent error codes, and shared published limits.

### Security, Privacy, Accessibility, or Operational Requirements

Limit failures should not expose stack traces or filesystem paths and should produce metrics by failure category.

### Dependencies

Secure ZIP Batch Upload.

### Acceptance Criteria

- [ ] **Given** a request exceeds the configured max body size, **when** it reaches the API, **then** the connection is rejected safely before full buffering.
- [ ] **Given** a file is named `.zip` but is not a ZIP, **when** validation runs, **then** no batch is created and the user sees an unsupported archive message.
- [ ] **Given** UI guidance displays limits, **when** API limits are changed, **then** the displayed values update from the same source or contract.

### Validation Evidence

- [ ] API tests cover max body size, content sniffing, malformed multipart, and timeout behavior.
- [ ] Security test covers resource-exhaustion attempts within safe fixture sizes.

### Out of Scope

Resumable upload protocol and external malware scanning.

### Definition of Ready

- [ ] MVP size limits are selected.
- [ ] Error response contract is defined.

### Definition of Done

- [ ] API and worker enforce complementary limits.
- [ ] Limit behavior is documented and tested.

## Processing Status Visibility

**Business Rank:** 016  
**Release Stage:** Functional Prototype  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### User Story

**As a** creator  
**I need** to understand processing status  
**So that** I know whether ClipSense is waiting, working, done, failed, or needs my action.

### Product Outcome

Batch detail shows trustworthy status, counts, and failure or partial-success information.

### Business Value

Processing is asynchronous. Without accurate status, creators cannot trust the product or plan editing work.

### Repository Evidence

Dashboard and detail pages display `batch.status`. The worker updates `processing`, `complete`, or `failed`. There is no polling loop on the detail page, no stage progress, no failure reason, and no partial state.

### Functional Requirements and Business Rules

Expose status labels for queued, processing, complete, failed, partial, cancelled, and retrying when supported. Show last update time and clear next action. Treat email, push, or background-completion notifications as post-MVP unless controlled-release evidence shows creators cannot use the product without them.

### Consumer Safety and Trust

Never show complete until required output is persisted and internally consistent.

### Data and State Requirements

State transitions must come from the shared state model and include timestamps.

### Failure and Fallback Behaviour

If status fetch fails, show stale status with retry instead of clearing results.

### Edge Cases

Worker offline, queue backlog, batch partially processed, browser reconnect, and result rows inserted after status update.

### Security and Privacy Requirements

Status details must not expose private filesystem paths or another user's job state.

### Accessibility Requirements

State changes must be announced through accessible status regions.

### Performance and Reliability Requirements

Polling or eventing must avoid excessive requests and stop when terminal state is reached.

### Observability and Operational Requirements

Emit metrics for state duration, queue wait, processing duration, failed state counts, and later notification-delivery outcomes if external notifications are added.

### Assumptions

Polling is enough for the functional MVP; WebSockets and external completion notifications remain future enhancements until justified.

### Dependencies

Batch Schema and State Model.

### Acceptance Criteria

- [ ] **Given** a batch is queued, **when** the detail page loads, **then** the creator sees that processing has not completed.
- [ ] **Given** processing progresses, **when** the state changes, **then** the UI updates without requiring a full manual navigation cycle.
- [ ] **Given** processing fails, **when** the detail page loads, **then** the creator sees a safe reason and any available recovery action.
- [ ] **Given** assistive technology is active, **when** status changes, **then** the change is announced.

### Validation Evidence

- [ ] API tests cover all MVP states.
- [ ] UI tests cover status polling, terminal states, and accessible announcements.
- [ ] Worker integration test covers status transitions.

### Out of Scope

Real-time WebSockets and detailed per-frame progress.

### Definition of Ready

- [ ] State vocabulary and UI labels are approved.
- [ ] Polling cadence or event strategy is selected.

### Definition of Done

- [ ] Status display is accurate, accessible, and tested.
- [ ] No false complete state remains in supported flows.

## Redis Job Enqueue and Worker Startup

**Business Rank:** 017  
**Release Stage:** Functional Prototype  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### Purpose

Ensure accepted batches are handed to the worker and the worker starts only when required dependencies are available.

### Consumer, Business, or Risk-Reduction Value

The first useful result depends on reliable movement from upload to processing.

### Repository Evidence

API uses `rdb.RPush` to `jobs:batch`; worker uses `rdb.blpop('jobs:batch', timeout=0)`. API pings Redis at startup. Worker initializes Whisper, SentenceTransformer, Redis, and Qdrant at module import. There is no startup readiness endpoint, model availability check, job schema versioning, or retry acknowledgement.

### Scope and Required Behaviour

Define job payload contract, enqueue timeout, worker startup validation, dependency failure behavior, model load failure behavior, and safe invalid-job handling.

### Security, Privacy, Accessibility, or Operational Requirements

Job payloads should include IDs and storage references only, not private transcript content or secrets.

### Dependencies

Secure ZIP Batch Upload; Local Source Storage Baseline.

### Acceptance Criteria

- [ ] **Given** Redis is unavailable, **when** the API starts, **then** startup fails or readiness reports unavailable according to environment policy.
- [ ] **Given** a job payload is invalid, **when** the worker consumes it, **then** it is rejected and logged without crashing the worker.
- [ ] **Given** required AI models cannot load, **when** the worker starts, **then** it fails visibly instead of silently consuming jobs it cannot process.

### Validation Evidence

- [ ] Unit tests cover job payload validation.
- [ ] Integration test covers successful enqueue and worker consume with a minimal fixture.
- [ ] Operational check documents worker startup dependencies.

### Out of Scope

Dead-letter queue and acknowledged retry semantics, covered later.

### Definition of Ready

- [ ] Job payload fields and versioning are defined.
- [ ] Worker dependency list is documented.

### Definition of Done

- [ ] Enqueue and worker startup behavior are deterministic.
- [ ] Invalid jobs and unavailable dependencies are safely handled.

## Safe ZIP Extraction and Resource Limits

**Business Rank:** 018  
**Release Stage:** Functional Prototype  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Verified for Core Cases; Requires Hardening

### Purpose

Safely extract supported videos from ZIP archives without path traversal, absolute-path writes, unbounded file counts, or decompression resource exhaustion.

### Consumer, Business, or Risk-Reduction Value

This protects creator data and service availability at the most dangerous media-ingestion boundary.

### Repository Evidence

`apps/api/ai_worker/zip_safety.py` rejects absolute paths, Windows drive paths, empty/current/parent segments, extraction outside destination, more than 200 files, and more than 10 GB extracted bytes. Tests in `test_zip_safety.py` passed locally.

### Scope and Required Behaviour

Keep safe extraction, add tests for duplicate output paths, symlinks if applicable, corrupted ZIPs, nested folders, zero-byte media, encrypted archives, and API preflight alignment.

### Security, Privacy, Accessibility, or Operational Requirements

Extraction logs must avoid unnecessary original path leakage and must provide failure categories for operators.

### Dependencies

Secure ZIP Batch Upload.

### Acceptance Criteria

- [ ] **Given** a ZIP member tries path traversal, **when** extraction runs, **then** no file escapes the processing directory.
- [ ] **Given** a ZIP exceeds file or extracted-byte limits, **when** extraction runs, **then** extraction stops and partial files are cleaned.
- [ ] **Given** a corrupted or encrypted ZIP is submitted, **when** worker validation runs, **then** the batch fails safely with a persisted validation reason.

### Validation Evidence

- [ ] Existing ZIP safety tests remain passing.
- [ ] Additional malformed archive tests cover duplicate paths, corrupted files, encrypted archives, and nested folders.
- [ ] Integration evidence confirms API guidance matches worker enforcement.

### Out of Scope

Malware scanning and media transcoding.

### Definition of Ready

- [ ] Additional malicious ZIP fixtures are defined.
- [ ] Failure reason schema is available.

### Definition of Done

- [ ] ZIP safety coverage includes current and credible malformed cases.
- [ ] Worker failures are surfaced to batch state safely.

## First Clip Analysis Result

**Business Rank:** 019  
**Release Stage:** Functional Prototype  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Partially Implemented; Missing Validation

### User Story

**As a** creator  
**I need** ClipSense to produce an understandable first analysis result  
**So that** I can see value from uploaded footage before advanced storytelling exists.

### Product Outcome

At least one safely extracted clip produces transcript text where possible, a summary, simple mood, simple role, topic, duration, and visible result data.

### Business Value

This proves the core pre-editing loop: source footage becomes structured review material.

### Repository Evidence

Worker `process_batch` extracts audio with FFmpeg, transcribes with Whisper, summarizes by truncation, classifies simple mood and role, calculates duration through `ffmpeg.probe`, and inserts clips. UI displays title, summary, mood, role, and duration but not full transcript.

### Functional Requirements and Business Rules

Process only videos that passed safe ZIP extraction into clip records with filename, title, transcript, summary, mood, role, topic, duration, and created timestamp. Store empty transcript only with an explicit warning or per-clip failure state.

### Consumer Safety and Trust

Mark AI analysis as generated and potentially imperfect. Do not hide failed transcription behind a neutral-looking summary.

### Data and State Requirements

Clip records must belong to one batch and be ordered predictably. Duration should aggregate into batch duration.

### Failure and Fallback Behaviour

If one clip fails transcription, the batch should record a per-clip failure and continue where safe.

### Edge Cases

No audio track, unsupported codec, very short clips, very long clips, non-English speech, silence, FFmpeg probe failure, and model timeout.

### Security and Privacy Requirements

Transcripts are sensitive generated data and must be protected by the same authorization as source media.

### Accessibility Requirements

Results should be readable with semantic headings and not depend only on color-coded mood/role.

### Performance and Reliability Requirements

Processing time must be bounded by documented prototype expectations and large-file limits.

### Observability and Operational Requirements

Record per-stage durations and failure categories for extract, transcribe, analyze, persist, and vector upsert.

### Assumptions

Local Whisper and heuristic summary/classification are acceptable for the prototype.

### Dependencies

Safe ZIP Extraction and Resource Limits; Redis Job Enqueue and Worker Startup; Batch Schema and State Model.

### Acceptance Criteria

- [ ] **Given** a supported ZIP with one valid safely extracted video, **when** processing completes, **then** a clip record exists with transcript or explicit transcript failure, summary, role, mood, topic, and duration.
- [ ] **Given** FFmpeg cannot extract audio for one clip, **when** processing continues, **then** the clip failure is visible and does not masquerade as successful analysis.
- [ ] **Given** a creator opens batch detail, **when** clip results exist, **then** the visible cards expose the meaningful first result.

### Validation Evidence

- [ ] Worker integration test processes a tiny media fixture or mocked FFmpeg/Whisper path.
- [ ] API test verifies clip result serialization.
- [ ] UI test verifies result rendering for successful and failed clip analysis.

### Out of Scope

Advanced summaries, speaker diarization, timestamps, or visual scene analysis.

### Definition of Ready

- [ ] Minimal clip analysis fields are agreed.
- [ ] Media fixture or mock strategy is available.

### Definition of Done

- [ ] First result is useful, honest, persisted, and visible.
- [ ] Success and per-clip failure evidence is tested.

## Accessible Upload and Status UX

**Business Rank:** 020  
**Release Stage:** Functional Prototype  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Missing Validation

### User Story

**As a** creator using keyboard or assistive technology  
**I need** upload and status controls to be accessible  
**So that** I can complete the first ClipSense workflow without drag/drop or visual-only cues.

### Product Outcome

The new batch flow supports keyboard file selection, labeled controls, progress and status announcements, and accessible errors.

### Business Value

The first functional increment should be usable by the intended audience without excluding disabled creators.

### Repository Evidence

`useDropzone` provides file input props in `batches/new/page.tsx`, but visible labels, accepted limits, role/status announcements, and upload error associations are incomplete. CSS has focus rings on `.btn`, but no validation has been run.

### Functional Requirements and Business Rules

Provide an explicit file input label, keyboard-operable select button, drag/drop as enhancement, `aria-live` status, actionable validation errors, and disabled-state messaging.

### Consumer Safety and Trust

Errors should explain what the user can do next and not vanish before being read.

### Data and State Requirements

Upload UI state must distinguish idle, selected, uploading, accepted, rejected, failed, and redirecting.

### Failure and Fallback Behaviour

If drag/drop fails or JavaScript errors, the file input path should remain usable where possible.

### Edge Cases

No token, rejected file, multiple files, network failure, slow upload, and repeated submission.

### Security and Privacy Requirements

Do not show full local file paths.

### Accessibility Requirements

Meet WCAG 2.2 AA for form labels, keyboard operation, error identification, focus order, status messages, and contrast.

### Performance and Reliability Requirements

Disable duplicate upload submission while preserving focus and status messaging.

### Observability and Operational Requirements

Track upload start, validation rejection, and client-side failure categories without file content.

### Assumptions

No full design system exists yet, so route-level controls can be improved directly.

### Dependencies

Supported ZIP Batch Guidance; Accessibility Baseline.

### Acceptance Criteria

- [ ] **Given** a keyboard-only creator opens new batch, **when** they tab through controls, **then** they can select a file and submit without drag/drop.
- [ ] **Given** upload begins, **when** the state changes, **then** assistive technology receives an upload status message.
- [ ] **Given** the API rejects the file, **when** the error is displayed, **then** focus and message association make the failure discoverable.

### Validation Evidence

- [ ] UI accessibility test covers keyboard path and status announcement.
- [ ] Manual screen-reader smoke test is documented for the new batch flow.

### Out of Scope

Full accessibility audit of every future feature.

### Definition of Ready

- [ ] Upload states and visible copy are defined.
- [ ] Accessibility target checklist is available.

### Definition of Done

- [ ] Upload and status UX is accessible and validated.
- [ ] Drag/drop is not the only successful path.

## User-Owned Batch Authorization

**Business Rank:** 021  
**Release Stage:** Functional MVP  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Missing Validation

### Purpose

Ensure every batch, clip, storyline, export, and future derived result is accessible only to its owner unless explicit sharing exists.

### Consumer, Business, or Risk-Reduction Value

Privacy of creator footage and generated transcripts is a release blocker.

### Repository Evidence

API `loadBatch` filters by `id` and `user_id`; `listBatches` filters by `user_id`. `clips`, `storylines`, and `storyline_clips` tables do not encode user ownership directly and no cross-user tests exist.

### Scope and Required Behaviour

Add authorization tests for list, detail, export, clip access, storyline access, and future delete/retry actions. Consider database constraints or joins that prevent orphan data access.

### Security, Privacy, Accessibility, or Operational Requirements

Unauthorized access attempts should be counted as security events without leaking private data.

### Dependencies

Authenticated API Boundary; Batch Schema and State Model.

### Acceptance Criteria

- [ ] **Given** user A and user B each have batches, **when** user A lists batches, **then** only user A's batches appear.
- [ ] **Given** user A requests user B's export, **when** the API authorizes the request, **then** no export content is returned.
- [ ] **Given** a clip or storyline is orphaned, **when** it is queried through a batch endpoint, **then** it cannot bypass batch ownership.

### Validation Evidence

- [ ] Integration tests cover cross-user list, detail, export, clips, and storylines.
- [ ] Database integrity tests cover orphan prevention or safe query behavior.

### Out of Scope

Shared workspaces and collaboration.

### Definition of Ready

- [ ] Test fixtures can create multiple users and batches.
- [ ] Ownership query strategy is selected.

### Definition of Done

- [ ] User-owned data isolation is tested for all MVP endpoints.
- [ ] No unowned derived result is reachable.

## Batch Results Review

**Business Rank:** 022  
**Release Stage:** Functional MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### User Story

**As a** creator  
**I need** to review processed clips and generated metadata  
**So that** I can decide what footage is useful before editing.

### Product Outcome

Batch detail presents clips, summaries, transcript access, roles, mood, duration, and processing warnings in a scannable, accessible way.

### Business Value

Review is where ClipSense turns processing into creator value.

### Repository Evidence

`apps/web/src/app/(dashboard)/batches/[id]/page.tsx` displays cards with duration, title, summary, mood, and role. Transcript is returned by the API but not shown. No filters, warnings, per-clip failures, loading skeleton, or retry action exists.

### Functional Requirements and Business Rules

Show clip count, duration, each clip's title/filename, transcript, summary, mood, role, topic, per-clip status, and whether the result is AI-generated or user-corrected.

### Consumer Safety and Trust

Distinguish empty transcript from successful silence or transcription failure.

### Data and State Requirements

Clip display should not depend on insertion timestamp alone once ordering metadata exists.

### Failure and Fallback Behaviour

If results are incomplete, show partial state rather than hiding all results.

### Edge Cases

No clips, all clips failed, partial failures, long summaries, very long filenames, missing duration, and loading with stale data.

### Security and Privacy Requirements

Results require current authorization and should not expose source paths.

### Accessibility Requirements

Use semantic sections, headings, keyboard focus, accessible tables/lists, and non-color-only status.

### Performance and Reliability Requirements

Large batches should paginate, virtualize, or progressively render once needed.

### Observability and Operational Requirements

Track result load latency and error categories.

### Assumptions

MVP result review can be read-only.

### Dependencies

First Clip Analysis Result; User-Owned Batch Authorization.

### Acceptance Criteria

- [ ] **Given** a completed batch, **when** the creator opens detail, **then** each processed clip shows core metadata and transcript access.
- [ ] **Given** some clips failed, **when** results render, **then** successful clips remain visible and failed clips show safe reasons.
- [ ] **Given** a long filename or transcript, **when** rendered on mobile and desktop, **then** text does not overlap or break controls.
- [ ] **Given** a screen-reader user navigates results, **when** they move through clips, **then** structure and status are understandable.

### Validation Evidence

- [ ] UI tests cover complete, empty, partial, and failed results.
- [ ] Accessibility test covers keyboard and semantic navigation.
- [ ] API test verifies transcript and failure fields are serialized safely.

### Out of Scope

Transcript correction, clip preview playback, and search.

### Definition of Ready

- [ ] MVP result fields and per-clip statuses are defined.
- [ ] Responsive layout constraints are specified.

### Definition of Done

- [ ] Results review is complete enough for MVP decision-making.
- [ ] Accessibility, partial-state, and authorization criteria pass.

## Worker Processing Integration Tests

**Business Rank:** 023  
**Release Stage:** Functional MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Missing Validation

### Purpose

Verify the worker pipeline across extraction, audio handling, transcription fallback, persistence, storyline creation, vector upsert, and failure status.

### Consumer, Business, or Risk-Reduction Value

The most important MVP promise depends on a worker that can process real or realistic media reliably.

### Repository Evidence

Worker tests cover ZIP safety only. There are no tests for `process_batch`, database writes, Redis payloads, FFmpeg extraction, Whisper failures, Qdrant upsert, or state updates.

### Scope and Required Behaviour

Create integration tests with mocked heavy dependencies or tiny fixtures, validate database writes, status changes, partial failures, empty ZIP behavior, and vector upsert contracts.

### Security, Privacy, Accessibility, or Operational Requirements

Test logs must avoid fixture media content where not necessary and must clean temporary files.

### Dependencies

First Clip Analysis Result; Safe ZIP Extraction and Resource Limits.

### Acceptance Criteria

- [ ] **Given** a valid job and fixture ZIP, **when** the worker processes it, **then** clips and a storyline are persisted and the batch reaches a terminal successful or partial state.
- [ ] **Given** transcription raises an exception for one clip, **when** the worker continues, **then** failure is recorded without losing successful clip results.
- [ ] **Given** Qdrant is unavailable, **when** vector upsert fails, **then** the batch state reflects whether reviewable results are available or blocked.

### Validation Evidence

- [ ] Worker integration tests run in CI or a documented local integration profile.
- [ ] Temporary directories, database rows, and vector points are cleaned after tests.
- [ ] Failure-path tests cover FFmpeg, model, database, Redis, and Qdrant failures.

### Out of Scope

Benchmarking and model-quality scoring.

### Definition of Ready

- [ ] Fixture or mock strategy avoids expensive model downloads in normal CI.
- [ ] Test database setup is available.

### Definition of Done

- [ ] Worker pipeline has repeatable integration evidence.
- [ ] Failure states are validated, not inferred from logs.

## Embedding and Vector Storage Baseline

**Business Rank:** 024  
**Release Stage:** Functional MVP  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### Purpose

Create a safe, maintainable baseline for embeddings and vector storage used by storyline generation and future search.

### Consumer, Business, or Risk-Reduction Value

Embeddings drive similarity and later search, but they are derived data that must be lifecycle-managed and recoverable.

### Repository Evidence

Worker uses `SentenceTransformer('all-MiniLM-L6-v2')`, Qdrant collection `clipsense_clips`, `recreate_collection`, and point payloads. Compose starts `qdrant/qdrant:v1.8.1` with a named volume. No vector cleanup, snapshots, schema versioning, embedding model version, or search endpoint exists.

### Scope and Required Behaviour

Record embedding model/version, vector collection config, payload schema, batch ownership metadata, upsert idempotency, cleanup on deletion, snapshot/backup needs, and migration path for model changes.

### Security, Privacy, Accessibility, or Operational Requirements

Embeddings are derived creator data and must follow retention/deletion policy.

### Dependencies

First Clip Analysis Result.

### Acceptance Criteria

- [ ] **Given** embeddings are created, **when** points are stored, **then** model version and ownership metadata are available for lifecycle management.
- [ ] **Given** a batch is deleted, **when** deletion completes, **then** related vector points are removed or marked according to policy.
- [ ] **Given** the embedding model changes, **when** new vectors are written, **then** old and new vectors can be distinguished.

### Validation Evidence

- [ ] Worker tests verify vector payload shape and model version recording.
- [ ] Deletion test verifies vector cleanup once deletion exists.
- [ ] Operational evidence documents Qdrant snapshot or recovery approach.

### Out of Scope

Advanced semantic search UX.

### Definition of Ready

- [ ] Vector payload schema is defined.
- [ ] Data lifecycle policy covers embeddings.

### Definition of Done

- [ ] Vector storage is versioned, authorized through batch ownership, and lifecycle-aware.
- [ ] Qdrant operational needs are captured for production release.

## Basic Storyline Suggestion

**Business Rank:** 025  
**Release Stage:** Functional MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### User Story

**As a** creator  
**I need** an initial suggested clip order  
**So that** I can begin shaping a story instead of reviewing random footage.

### Product Outcome

ClipSense creates and displays one basic AI-assisted storyline for processed clips.

### Business Value

Storyline organization is the differentiator between a transcription utility and a pre-editor.

### Repository Evidence

Worker creates one storyline titled `AI Sequence for {name}` using sentence-transformer embeddings and KMeans cluster labels. UI displays storylines and ordered clip chips. There is no explanation, quality scoring, manual correction, alternate strategies, or validation of sequence usefulness.

### Functional Requirements and Business Rules

Create a storyline only when enough result data and the MVP vector baseline exist, persist ordered clips with positions, label it as AI-generated, and expose its limitations.

### Consumer Safety and Trust

Do not overclaim narrative intelligence. Make clear the order is a suggestion.

### Data and State Requirements

Storyline positions must be stable, unique per storyline, and tied to clips in the same batch.

### Failure and Fallback Behaviour

If storyline generation fails but clips exist, show clips and mark storyline unavailable.

### Edge Cases

One clip, many clips, empty transcripts, identical embeddings, KMeans errors, and vector store outage.

### Security and Privacy Requirements

Storylines inherit batch authorization and must not include private paths.

### Accessibility Requirements

Ordered sequence must be readable as an ordered list, not only visual chips.

### Performance and Reliability Requirements

Generation must avoid unbounded clustering cost for large batches.

### Observability and Operational Requirements

Record generation duration, clip count, strategy, and failure reason.

### Assumptions

One basic AI sequence is enough for MVP; advanced strategies are post-MVP.

### Dependencies

First Clip Analysis Result; Embedding and Vector Storage Baseline.

### Acceptance Criteria

- [ ] **Given** a processed batch with clips and vector baseline available, **when** storyline generation succeeds, **then** the API returns a storyline with ordered clips from the same batch.
- [ ] **Given** only one clip exists, **when** generation runs, **then** a valid one-clip storyline or clear no-storyline state is produced.
- [ ] **Given** generation fails after clip processing, **when** the creator opens results, **then** clip review still works and storyline failure is visible.

### Validation Evidence

- [ ] Worker tests cover one-clip, multi-clip, and generation failure cases.
- [ ] API/UI tests cover storyline serialization and accessible ordered display.

### Out of Scope

Manual reordering, explanation generation, and multiple storyline strategies.

### Definition of Ready

- [ ] MVP storyline strategy and fallback are defined.
- [ ] Storyline schema constraints are available.

### Definition of Done

- [ ] Basic storyline appears when possible and fails safely when not.
- [ ] Acceptance and validation evidence cover edge cases.

## Export Processed Batch Results

**Business Rank:** 026  
**Release Stage:** Functional MVP  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Defective in Browser Flow

### User Story

**As a** creator  
**I need** to export processed results  
**So that** I can use ClipSense output in my editing or planning workflow.

### Product Outcome

A creator can export authorized batch results as JSON and CSV through a working UI and API path.

### Business Value

Export turns review into action outside ClipSense and completes the MVP loop.

### Repository Evidence

`exportBatch` supports JSON default and CSV query format. The web detail page uses an `<a>` link to `/api/batches/{id}/export?format=csv` without an Authorization header, which is defective with bearer-token auth.

### Functional Requirements and Business Rules

Export only the creator's batch, include stable fields, support JSON and CSV, handle empty or partial results, and initiate downloads through an authenticated client flow.

### Consumer Safety and Trust

Clearly identify generated AI fields and omit internal paths.

### Data and State Requirements

Exports should include batch ID/name, clip IDs, titles, summaries, roles, mood, topics, duration, transcripts where selected, storyline positions, and generation timestamp.

### Failure and Fallback Behaviour

Unauthorized, not-found, unsupported format, and empty result exports must return safe messages.

### Edge Cases

CSV escaping, newline characters in transcripts, very long text, partial batches, unsupported format, and browser pop-up/download behavior.

### Security and Privacy Requirements

Bearer-token export must not be bypassed by direct links, and downloaded files should not include server paths or secrets.

### Accessibility Requirements

Export controls require button labels, progress/error status, and keyboard operation.

### Performance and Reliability Requirements

MVP exports can be generated on request with response size limits and later streaming if needed.

### Observability and Operational Requirements

Track export format, success/failure, and size without logging exported content.

### Assumptions

JSON and CSV are enough for MVP.

### Dependencies

Batch Results Review; User-Owned Batch Authorization.

### Acceptance Criteria

- [ ] **Given** a completed authorized batch, **when** the creator chooses CSV export, **then** the browser downloads a CSV using authenticated API access.
- [ ] **Given** an unauthorized user requests export, **when** the API handles it, **then** no batch data is returned.
- [ ] **Given** export text contains commas or newlines, **when** CSV is generated, **then** the output remains valid CSV.
- [ ] **Given** an unsupported format is requested, **when** the API responds, **then** a safe `400` response is returned.

### Validation Evidence

- [ ] API tests cover JSON, CSV, unsupported format, escaping, empty result, and unauthorized export.
- [ ] UI/E2E test verifies authenticated download path.
- [ ] Security test confirms no internal paths appear in export.

### Out of Scope

EDL, XML, NLE plug-ins, and cloud export storage.

### Definition of Ready

- [ ] Export field set is agreed.
- [ ] Authenticated browser download strategy is selected.

### Definition of Done

- [ ] JSON and CSV exports work securely from UI and API.
- [ ] Existing unauthenticated link defect is resolved.

## Data Lifecycle Baseline

**Business Rank:** 027  
**Release Stage:** Functional MVP  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Not Implemented

### Purpose

Define and implement minimum retention, cleanup, and deletion expectations for uploaded and generated data.

### Consumer, Business, or Risk-Reduction Value

Creators need to know what ClipSense keeps, for how long, and how it can be deleted.

### Repository Evidence

Uploads and processing files are stored under shared volumes. There is no deletion endpoint, retention schedule, cleanup job, backup lifecycle, or vector cleanup. Docs mention retention generally but do not define behavior.

### Scope and Required Behaviour

Classify source archives, extracted clips, temporary audio, transcripts, summaries, embeddings, storylines, exports, logs, and backups. Define retention defaults, deletion semantics, cleanup timing, and user communication.

### Security, Privacy, Accessibility, or Operational Requirements

Deletion must include database rows, local files, processing files, vector points, generated exports, and future backups according to documented limits.

### Dependencies

Local Source Storage Baseline; Embedding and Vector Storage Baseline.

### Acceptance Criteria

- [ ] **Given** a creator uploads media, **when** they review data policy, **then** each data category has an owner, purpose, retention rule, and deletion behavior.
- [ ] **Given** temporary processing files remain after failure, **when** cleanup runs, **then** expired files are removed without deleting active work.
- [ ] **Given** a user requests deletion, **when** MVP deletion exists, **then** source, derived, and vector data are removed or documented exceptions are shown.

### Validation Evidence

- [ ] Lifecycle matrix covers media, transcripts, embeddings, exports, logs, and backups.
- [ ] Cleanup tests cover expired processing directories and active job protection.

### Out of Scope

Legal compliance by jurisdiction and enterprise retention policies.

### Definition of Ready

- [ ] Data categories are inventoried.
- [ ] MVP retention defaults are approved.

### Definition of Done

- [ ] Lifecycle rules are documented and enforced for MVP data.
- [ ] Future deletion work is traceable to exact stores.

## Failure Recovery Messaging

**Business Rank:** 028  
**Release Stage:** Functional MVP  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### User Story

**As a** creator  
**I need** understandable failure messages and recovery actions  
**So that** I can fix upload or processing problems without guessing.

### Product Outcome

Failures from validation, storage, queue, worker, AI model, export, or dependency outages produce safe, actionable user messages and operator diagnostics.

### Business Value

Honest recovery preserves trust when AI and media processing inevitably fail.

### Repository Evidence

API `httpError` returns raw `err.Error()` with status `500`. Worker logs failures and marks batch `failed`, but no failure reason is persisted. UI toasts raw response text on upload failure.

### Functional Requirements and Business Rules

Normalize error codes, user messages, retryability, support IDs, and operator details. Persist batch and clip failure reasons with safe categories.

### Consumer Safety and Trust

Never expose stack traces, DSNs, filesystem paths, tokens, or private media content in user errors.

### Data and State Requirements

Store failure category, safe user message, internal correlation ID, retryability, and timestamp.

### Failure and Fallback Behaviour

If error persistence fails, log a safe event and present a generic failure message.

### Edge Cases

Malformed uploads, Redis outage, Postgres outage, Qdrant outage, FFmpeg failure, model load failure, export failure, and unauthorized requests.

### Security and Privacy Requirements

Separate user-safe messages from internal diagnostic logs.

### Accessibility Requirements

Errors must be announced and associated with the relevant control or result region.

### Performance and Reliability Requirements

Error handling must not block worker shutdown or keep jobs in ambiguous states.

### Observability and Operational Requirements

Emit structured logs and metrics by failure category and correlation ID.

### Assumptions

Support IDs can be internal correlation IDs for MVP.

### Dependencies

Batch Schema and State Model; Processing Status Visibility.

### Acceptance Criteria

- [ ] **Given** an upload validation error occurs, **when** the API responds, **then** the creator receives a safe, actionable message and correct status code.
- [ ] **Given** worker processing fails, **when** the batch is viewed, **then** failure category and retryability are visible without internal details.
- [ ] **Given** a server exception occurs, **when** logs are inspected, **then** operators can correlate the failure without exposing secrets to the user.

### Validation Evidence

- [ ] API tests cover normalized error responses.
- [ ] Worker tests cover persisted failure categories.
- [ ] UI accessibility tests cover error announcement and focus.

### Out of Scope

Full customer support ticketing system.

### Definition of Ready

- [ ] Error taxonomy and response contract are defined.
- [ ] Failure fields exist in state model.

### Definition of Done

- [ ] Supported failure paths produce safe messages and diagnostics.
- [ ] Raw internal errors are not returned to consumers.

## Runtime Health Checks

**Business Rank:** 029  
**Release Stage:** Functional MVP  
**Fibonacci Estimate:** 3  
**Current Implementation Assessment:** Partially Implemented; Requires Expansion

### Purpose

Expose health and readiness signals for the API and processing dependencies used by the MVP.

### Consumer, Business, or Risk-Reduction Value

Operators need to know whether ClipSense can accept and process batches before creators hit failures.

### Repository Evidence

`/api/health` checks database and Redis and returns JSON. Tests cover dependency failure. Worker has no health/readiness endpoint or startup probe. Qdrant and model readiness are not reflected in API health.

### Scope and Required Behaviour

Distinguish liveness from readiness, include database, Redis, storage, Qdrant, worker availability, and model readiness where applicable, and document operational use.

### Security, Privacy, Accessibility, or Operational Requirements

Health endpoints must not expose secrets, DSNs, tokens, file paths, or private metadata.

### Dependencies

Redis Job Enqueue and Worker Startup.

### Acceptance Criteria

- [ ] **Given** Postgres or Redis is unavailable, **when** API health is called, **then** the response returns unavailable dependency status without secrets.
- [ ] **Given** Qdrant or the worker is unavailable, **when** readiness is checked, **then** operators can see that analysis cannot complete.
- [ ] **Given** the service is alive but not ready, **when** deployment checks run, **then** traffic can be held back.

### Validation Evidence

- [ ] API tests cover liveness and readiness states.
- [ ] Operational docs describe health-check use in local and production deployment.

### Out of Scope

Full metrics, traces, and alerting.

### Definition of Ready

- [ ] Dependencies required for accepting versus processing jobs are defined.
- [ ] Health response shape is agreed.

### Definition of Done

- [ ] Health/readiness signals reflect MVP dependencies.
- [ ] Tests prove unavailable dependencies are reported safely.

## Remove Unsafe Development Secrets and Defaults

**Business Rank:** 030  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Requires Hardening

### Purpose

Prevent development credentials, JWT defaults, and insecure DSNs from being used in shared or production environments.

### Consumer, Business, or Risk-Reduction Value

Unsafe defaults can expose every creator account and batch in a deployed environment.

### Repository Evidence

API defaults `JWT_SECRET` to `dev-secret` and `DATABASE_URL` to `postgres://clipsense:clipsense@postgres:5432/clipsense?sslmode=disable`. Compose uses `clipsense` as Postgres password. `.env.example` says `JWT_SECRET=replace-me-for-local-dev`.

### Scope and Required Behaviour

Introduce environment mode, required secret validation outside local dev, minimum JWT secret entropy, production DSN requirements, and safe startup failure messages.

### Security, Privacy, Accessibility, or Operational Requirements

No production process should start with known development credentials, missing JWT secret, or unaudited public CORS origins.

### Dependencies

Security and Privacy Baseline.

### Acceptance Criteria

- [ ] **Given** `APP_ENV=production` and `JWT_SECRET` is missing or default, **when** the API starts, **then** startup fails with a safe configuration error.
- [ ] **Given** production database configuration uses development credentials, **when** config validation runs, **then** the process refuses to start.
- [ ] **Given** local development starts without custom secrets, **when** docs are followed, **then** local-only defaults are clearly identified.

### Validation Evidence

- [ ] Config unit tests cover local, test, and production secret validation.
- [ ] CI includes a configuration validation check.

### Out of Scope

Cloud secret manager selection.

### Definition of Ready

- [ ] Environment mode names are agreed.
- [ ] Required production config keys are listed.

### Definition of Done

- [ ] Unsafe defaults are blocked outside local dev.
- [ ] Documentation explains local versus production configuration.

## Harden Browser Session Handling

**Business Rank:** 031  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Requires Hardening

### User Story

**As a** creator  
**I need** my signed-in session handled safely  
**So that** my batches are not exposed through avoidable token theft or confusing session state.

### Product Outcome

Browser session handling supports visible logout, expiry recovery, safer token storage or cookie strategy, and consistent auth failure behavior.

### Business Value

Session risk directly affects private media and transcript access.

### Repository Evidence

`useAuthToken.ts` stores JWTs in `localStorage`. `clear` exists but no UI uses it. Protected pages redirect when no token exists. No refresh token, cookie settings, CSRF strategy, or token revocation exists.

### Functional Requirements and Business Rules

Choose MVP session strategy, add visible logout, handle token expiry, remove invalid tokens, avoid exposing tokens in URLs, and document tradeoffs.

### Consumer Safety and Trust

Session expiration should not lose visible work or show confusing blank pages.

### Data and State Requirements

Client auth state must reset safely on logout or `401` responses.

### Failure and Fallback Behaviour

If an API returns `401`, the app clears stale session state and routes to sign in with context.

### Edge Cases

Expired token, localStorage unavailable, multi-tab logout, clock skew, XSS risk, and direct export download.

### Security and Privacy Requirements

Follow OWASP session/JWT storage guidance; if localStorage remains for MVP, document risk and XSS mitigations.

### Accessibility Requirements

Logout and session-expired notices must be keyboard accessible and announced.

### Performance and Reliability Requirements

Auth checks should not cause render loops or repeated redirects.

### Observability and Operational Requirements

Track session-expired events and unauthorized API responses without logging tokens.

### Assumptions

MVP can remain single-session if risks are accepted and documented.

### Dependencies

Creator Account Access and Recovery; Remove Unsafe Development Secrets and Defaults.

### Acceptance Criteria

- [ ] **Given** a creator is signed in, **when** they use logout, **then** client session state is cleared and protected routes require sign-in.
- [ ] **Given** a token is expired, **when** a protected API returns `401`, **then** the app clears stale credentials and shows a session-expired message.
- [ ] **Given** token storage is reviewed, **when** the MVP ships, **then** the chosen strategy and residual risks are documented.

### Validation Evidence

- [ ] UI tests cover logout, expired token, and unauthorized API handling.
- [ ] Security review records token storage decision and mitigations.

### Out of Scope

SSO, MFA, device management, and enterprise session policies.

### Definition of Ready

- [ ] Session strategy decision is made.
- [ ] Auth error handling contract is available.

### Definition of Done

- [ ] Session lifecycle is visible, recoverable, and tested.
- [ ] Token handling risk is reduced or explicitly accepted for MVP.

## Fix Authenticated CSV Export

**Business Rank:** 032  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 3  
**Current Implementation Assessment:** Defective

### Purpose

Correct the browser export path so CSV export works with the same authentication model as other protected APIs.

### Consumer, Business, or Risk-Reduction Value

The MVP is incomplete if a creator cannot export through the UI after login.

### Repository Evidence

`apps/web/src/app/(dashboard)/batches/[id]/page.tsx` uses a plain anchor to the API export URL. The API requires `Authorization: Bearer`, so the browser request lacks the token.

### Scope and Required Behaviour

Replace direct API link with an authenticated fetch that downloads a Blob, or implement a safe signed one-time download mechanism. Preserve authorization and safe filename behavior.

### Security, Privacy, Accessibility, or Operational Requirements

Do not put bearer tokens in query strings. Export button must report download progress/error accessibly.

### Dependencies

Export Processed Batch Results; Harden Browser Session Handling.

### Acceptance Criteria

- [ ] **Given** a signed-in creator views a completed batch, **when** they select CSV export, **then** the request includes valid authentication without exposing the token in the URL.
- [ ] **Given** the token is expired, **when** export is selected, **then** the app shows a session-expired path instead of downloading an error page.
- [ ] **Given** export fails, **when** the API returns an error, **then** the UI announces a safe failure message.

### Validation Evidence

- [ ] UI test verifies authenticated export request headers or signed download flow.
- [ ] API test verifies unauthorized CSV export is rejected.

### Out of Scope

New export formats.

### Definition of Ready

- [ ] Browser download approach is selected.
- [ ] Export API response headers are defined.

### Definition of Done

- [ ] CSV export works from the UI after login.
- [ ] No token appears in URLs, logs, or downloaded content.

## Persist Failure Reasons and Partial Processing State

**Business Rank:** 033  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Requires Hardening

### Purpose

Persist safe failure reasons and partial success information for batches and clips.

### Consumer, Business, or Risk-Reduction Value

Creators should know whether retrying, changing input, or using partial results is appropriate.

### Repository Evidence

Worker logs per-video failures but still creates clip rows with empty transcripts. Batch failure sets only `status='failed'`. No failure columns or per-clip statuses exist.

### Scope and Required Behaviour

Add failure category, safe message, retryability, per-clip processing status, partial batch state, and support correlation ID.

### Security, Privacy, Accessibility, or Operational Requirements

Failure text must not include server paths, stack traces, model internals, or private content.

### Dependencies

Batch Schema and State Model; Failure Recovery Messaging.

### Acceptance Criteria

- [ ] **Given** one clip fails transcription, **when** the worker persists results, **then** the failed clip has a safe failure reason and the batch can show partial success.
- [ ] **Given** all clips fail, **when** processing ends, **then** the batch is failed with a safe, actionable reason.
- [ ] **Given** an operator reviews logs, **when** they search by correlation ID, **then** they can locate diagnostic detail without exposing it to the user.

### Validation Evidence

- [ ] Worker tests cover per-clip failure and all-failed batch behavior.
- [ ] API/UI tests cover partial and failed state serialization/rendering.

### Out of Scope

Automated retry policy, handled separately.

### Definition of Ready

- [ ] Failure taxonomy is approved.
- [ ] Schema changes are planned.

### Definition of Done

- [ ] Failure reasons are persisted and visible safely.
- [ ] Empty transcript is no longer the only failure signal.

## Reliable Queue Acknowledgement and Dead-Letter Handling

**Business Rank:** 034  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 13  
**Current Implementation Assessment:** Requires Hardening

### Purpose

Replace or augment Redis list queue behavior with acknowledgement, retry, timeout, and dead-letter semantics appropriate for media jobs.

### Consumer, Business, or Risk-Reduction Value

Creators should not lose batches silently if a worker dies after `BLPOP`, and operators need a path for poison jobs.

### Repository Evidence

API uses `RPush`; worker uses blocking `BLPOP`, which removes the job before processing. There is no in-progress set, retry count, visibility timeout, dead-letter queue, poison-job handling, or idempotency key.

### Scope and Required Behaviour

Define reliable queue design using Redis Streams, lists with processing queues, or another justified mechanism. Include acknowledgement, retry limits, backoff, poison queue, requeue on worker crash, and operator visibility.

### Security, Privacy, Accessibility, or Operational Requirements

Queue payloads must avoid unnecessary private data and support safe operational inspection.

### Dependencies

Redis Job Enqueue and Worker Startup; Persist Failure Reasons and Partial Processing State.

### Acceptance Criteria

- [ ] **Given** a worker crashes after claiming a job, **when** recovery runs, **then** the job is retried or marked failed according to policy.
- [ ] **Given** a job fails repeatedly, **when** retry limit is reached, **then** it moves to a dead-letter state with safe failure evidence.
- [ ] **Given** duplicate job delivery occurs, **when** the worker handles it, **then** processing does not create duplicate clips or storylines.

### Validation Evidence

- [ ] Queue integration tests cover acknowledgement, retry, crash recovery, and dead-letter behavior.
- [ ] Operational runbook explains how to inspect and requeue failed jobs.

### Out of Scope

Replacing Redis with another broker unless justified by this design.

### Definition of Ready

- [ ] Queue reliability pattern is selected.
- [ ] Retry and dead-letter policy is approved.

### Definition of Done

- [ ] Queue delivery has tested acknowledgement and poison-job handling.
- [ ] Operators can diagnose stuck and failed jobs.

## Idempotent Processing and Duplicate Job Protection

**Business Rank:** 035  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Not Implemented

### Purpose

Make worker processing safe when a batch job is retried, duplicated, interrupted, or resumed.

### Consumer, Business, or Risk-Reduction Value

Reliable retries require idempotency; otherwise hardening the queue can create duplicate clips, storylines, and vector points.

### Repository Evidence

Worker deletes and recreates `PROCESS_DIR / batch_id`, inserts new clip IDs on every run, and upserts embeddings by new clip ID. No job idempotency key, unique source asset mapping, or prior-result cleanup transaction exists.

### Scope and Required Behaviour

Define idempotency keys, batch processing lock, retry-safe cleanup, unique constraints, transaction boundaries, vector replacement, and behavior for partial prior results.

### Security, Privacy, Accessibility, or Operational Requirements

Idempotency records should not expose media content and must be scoped by batch ownership.

### Dependencies

Reliable Queue Acknowledgement and Dead-Letter Handling.

### Acceptance Criteria

- [ ] **Given** the same job is delivered twice, **when** the worker processes it, **then** only one coherent set of clips/storylines remains.
- [ ] **Given** processing is interrupted after clips are inserted, **when** retry starts, **then** stale partial results are reconciled before new results are exposed.
- [ ] **Given** vector points from a prior attempt exist, **when** retry succeeds, **then** old points are removed or replaced consistently.

### Validation Evidence

- [ ] Worker integration tests cover duplicate delivery and interrupted retry.
- [ ] Database and vector store checks verify no duplicate result sets remain.

### Out of Scope

User-triggered reprocessing with alternate settings.

### Definition of Ready

- [ ] Queue retry semantics are defined.
- [ ] Result replacement strategy is approved.

### Definition of Done

- [ ] Processing is retry-safe and duplicate-safe.
- [ ] Idempotency behavior is tested across database and vector storage.

## Accessible Results Review and Keyboard Navigation

**Business Rank:** 036  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Missing Validation

### User Story

**As a** creator using keyboard or assistive technology  
**I need** results and storylines to be navigable and understandable  
**So that** I can review ClipSense output without visual-only interaction.

### Product Outcome

Batch detail, clip list, storyline order, export controls, failures, and loading states are accessible.

### Business Value

The MVP cannot be considered releasable if its core review screen excludes users.

### Repository Evidence

Batch detail uses cards and visual chips. Export is a link styled as a button. No axe, Playwright, screen-reader, keyboard, or contrast evidence exists.

### Functional Requirements and Business Rules

Use semantic lists or tables, headings, focus management, keyboard actions, non-color status, accessible export controls, and responsive text constraints.

### Consumer Safety and Trust

AI-generated status and failure warnings must be available to assistive technologies.

### Data and State Requirements

Long text should be expandable without losing keyboard position.

### Failure and Fallback Behaviour

Loading and error states should be announced and recoverable.

### Edge Cases

Empty results, partial results, long transcripts, mobile viewport, high zoom, and reduced motion.

### Security and Privacy Requirements

Accessibility attributes must not reveal hidden private data from other batches.

### Accessibility Requirements

Target WCAG 2.2 AA for MVP result routes.

### Performance and Reliability Requirements

Accessibility improvements must not introduce excessive re-rendering for large result sets.

### Observability and Operational Requirements

CI should collect accessibility test output for release evidence.

### Assumptions

Automated accessibility checks are necessary but not sufficient.

### Dependencies

Batch Results Review; Accessibility Baseline.

### Acceptance Criteria

- [ ] **Given** a keyboard-only creator reviews results, **when** they navigate clips, storylines, and export, **then** focus order is logical and every control is operable.
- [ ] **Given** a screen reader announces a storyline, **when** it encounters ordered clips, **then** order and labels are understandable.
- [ ] **Given** errors or partial states exist, **when** results render, **then** assistive technology receives the status.

### Validation Evidence

- [ ] Automated accessibility test covers dashboard, upload, batch detail, and export.
- [ ] Manual keyboard and screen-reader smoke notes are recorded for MVP.

### Out of Scope

Full third-party audit.

### Definition of Ready

- [ ] MVP routes and key states are listed.
- [ ] Accessibility tooling is selected.

### Definition of Done

- [ ] Core review workflow meets documented accessibility acceptance criteria.
- [ ] Accessibility evidence is included in release readiness.

## End-to-End MVP Smoke Test

**Business Rank:** 037  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Not Implemented

### User Story

**As a** product team  
**I need** an automated smoke test for the MVP journey  
**So that** every release proves a creator can get from login to export.

### Product Outcome

A repeatable test signs in, uploads a tiny supported ZIP or mocked fixture, waits for processing, reviews results, and exports.

### Business Value

This is the release gate for the smallest complete ClipSense promise.

### Repository Evidence

CI builds components separately but no full stack E2E test exists. Docker runtime validation could not be run locally because Docker is not on PATH.

### Functional Requirements and Business Rules

Test must cover account creation/login, upload, batch state transition, result review, storyline visibility, export, and cleanup.

### Consumer Safety and Trust

The test should fail if the UI shows false completion or unauthenticated export.

### Data and State Requirements

Use disposable users, batches, fixture files, and cleanup after test completion.

### Failure and Fallback Behaviour

When a dependency is unavailable, the test reports the exact missing dependency.

### Edge Cases

Slow model startup, worker cold start, queue delay, browser token expiry, and fixture media compatibility.

### Security and Privacy Requirements

Fixtures must contain no private real media.

### Accessibility Requirements

Smoke flow should include basic keyboard path where feasible.

### Performance and Reliability Requirements

Smoke test should have bounded timeouts and be separate from heavier performance tests.

### Observability and Operational Requirements

Collect logs and test artifacts on failure without exposing secrets.

### Assumptions

Heavy AI steps can be mocked or fixture-minimized in CI.

### Dependencies

Worker Processing Integration Tests; Fix Authenticated CSV Export.

### Acceptance Criteria

- [ ] **Given** the Compose stack is available, **when** the E2E smoke test runs, **then** it creates an account, uploads a fixture, observes terminal state, reviews results, and exports.
- [ ] **Given** export auth is broken, **when** the smoke test reaches export, **then** the test fails.
- [ ] **Given** processing stalls beyond timeout, **when** the test ends, **then** logs identify API, worker, queue, database, and vector status.

### Validation Evidence

- [ ] Playwright or equivalent E2E test committed and runnable locally.
- [ ] CI or nightly job records full-stack smoke evidence.
- [ ] Failure artifacts exclude secrets and fixture media content where unnecessary.

### Out of Scope

Load testing and advanced AI quality evaluation.

### Definition of Ready

- [ ] Disposable test fixture and stack startup path are available.
- [ ] Authenticated export is fixed.

### Definition of Done

- [ ] MVP smoke test passes in an agreed environment.
- [ ] Release readiness requires this evidence.

## Dependency Vulnerability Remediation

**Business Rank:** 038  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Requires Hardening

### Purpose

Assess and remediate known dependency vulnerabilities and unsupported package versions across web, API, worker, and containers.

### Consumer, Business, or Risk-Reduction Value

Outdated dependencies in auth, web runtime, media processing, and AI packages can become product security incidents.

### Repository Evidence

`docs/MVP_STABILIZATION.md` reports prior `npm audit` findings including a critical Next.js issue. Local `npm ci` and `npm audit --package-lock-only --json` both stalled in this environment. `go list -m -u all` showed many available updates. Python dependencies are pinned but no successful `pip-audit` evidence exists.

### Scope and Required Behaviour

Run `npm audit`, `govulncheck`, `pip-audit` or equivalent, container scanning, update planning, regression tests, and documented risk acceptance for unfixed items.

### Security, Privacy, Accessibility, or Operational Requirements

Do not blindly auto-upgrade packages that change runtime behavior without regression testing the MVP flow.

### Dependencies

Testing Strategy Baseline; CI and Documentation Foundation.

### Acceptance Criteria

- [ ] **Given** dependency scans run, **when** vulnerabilities are found, **then** each has a fix, mitigation, or explicit risk acceptance.
- [ ] **Given** Next.js, Go, or Python dependencies are upgraded, **when** tests run, **then** auth, upload, processing, and export regressions are checked.
- [ ] **Given** a scan cannot complete, **when** release readiness is reviewed, **then** the exact blocker and residual risk are documented.

### Validation Evidence

- [ ] Current npm, Go, Python, and container scan outputs are attached to release evidence.
- [ ] Updated dependency lock files pass build and tests.

### Out of Scope

Major framework migration unless required by security or support status.

### Definition of Ready

- [ ] Scan tooling and command set are selected.
- [ ] Current stalled npm audit behavior is investigated.

### Definition of Done

- [ ] Known release-blocking vulnerabilities are fixed or accepted with mitigation.
- [ ] Dependency scans are automated for future releases.

## Container Build Hardening

**Business Rank:** 039  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Requires Hardening

### Purpose

Harden API, worker, and web container builds for reproducibility, least privilege, and production readiness.

### Consumer, Business, or Risk-Reduction Value

Containers are the deployment unit; unsafe images expand blast radius and make releases unreliable.

### Repository Evidence

API Dockerfile uses multi-stage build but runs as default root and runs `go mod tidy` during build. Worker Dockerfile installs dependencies as root and copies only `main.py`, not `zip_safety.py`, which appears defective for runtime imports. Web Dockerfile uses `npm install` instead of `npm ci` and runs as root.

### Scope and Required Behaviour

Use reproducible installs, copy required files, non-root users, minimal runtime images, health checks where useful, build cache discipline, no secrets in image layers, and image scan evidence.

### Security, Privacy, Accessibility, or Operational Requirements

Follow Docker best practices for multi-stage builds, least privilege, and secret handling.

### Dependencies

Repository and Environment Baseline; Dependency Vulnerability Remediation.

### Acceptance Criteria

- [ ] **Given** the worker image builds, **when** it starts, **then** `zip_safety.py` is present and importable.
- [ ] **Given** production containers run, **when** process user is inspected, **then** API, worker, and web do not run as root unless explicitly justified.
- [ ] **Given** dependency manifests are unchanged, **when** images build, **then** builds use lockfiles or pinned requirements rather than mutating dependency files.

### Validation Evidence

- [ ] Docker build tests pass for API, worker, and web.
- [ ] Container scan results are reviewed.
- [ ] Runtime smoke test confirms worker imports and service startup.

### Out of Scope

Kubernetes manifests and cloud registry policies.

### Definition of Ready

- [ ] Docker runtime target is known.
- [ ] Current Dockerfile defects are confirmed.

### Definition of Done

- [ ] Images build reproducibly and run as least-privilege users.
- [ ] Worker image includes all required runtime modules.

## Production Configuration and Secret Management

**Business Rank:** 040  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Not Implemented

### Purpose

Define production configuration loading, validation, secret handling, CORS policy, and environment separation.

### Consumer, Business, or Risk-Reduction Value

Production release requires confidence that secrets, origins, databases, queues, and model settings are intentional and safe.

### Repository Evidence

Environment variables are spread across `.env.example`, Compose, API globals, worker globals, and web build-time env. There is no central config validation package, production profile, or secret source.

### Scope and Required Behaviour

Implement typed config validation for API and worker, required production variables, CORS origin validation, secret entropy checks, non-secret config docs, and redaction in logs.

### Security, Privacy, Accessibility, or Operational Requirements

Secrets must never be logged, committed, embedded into public web bundles, or stored in image layers.

### Dependencies

Remove Unsafe Development Secrets and Defaults; Container Build Hardening.

### Acceptance Criteria

- [ ] **Given** production mode starts, **when** required config is missing, **then** startup fails before accepting traffic.
- [ ] **Given** logs include configuration summaries, **when** secrets are present, **then** values are redacted.
- [ ] **Given** web build receives public env vars, **when** the bundle is inspected, **then** no server-only secrets are exposed.

### Validation Evidence

- [ ] Config unit tests cover required, optional, invalid, and redacted values.
- [ ] Deployment checklist includes secret rotation and origin review.

### Out of Scope

Specific cloud secret manager integration unless deployment target requires it.

### Definition of Ready

- [ ] Production environment categories are defined.
- [ ] Required variables are inventoried.

### Definition of Done

- [ ] Production config fails safely and documents every required secret.
- [ ] No development default can reach production unnoticed.

## Privacy Notice, Retention, and Deletion Controls

**Business Rank:** 041  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 13  
**Current Implementation Assessment:** Not Implemented

### Purpose

Provide user-facing privacy communication and functional controls for retention and deletion before production release.

### Consumer, Business, or Risk-Reduction Value

Creators need control over uploaded footage and generated artifacts; production release without deletion is a trust and privacy risk.

### Repository Evidence

No privacy notice, deletion endpoint, retention job, vector cleanup, export cleanup, or account deletion exists.

### Scope and Required Behaviour

Implement batch deletion, derived-data cleanup, retention policy display, privacy notice, deletion confirmation, background cleanup, backup exception disclosure, and audit-safe deletion logs.

### Security, Privacy, Accessibility, or Operational Requirements

Deletion must cover source archives, processing files, clips, transcripts, storylines, embeddings, exports, and future cache entries. UI must be accessible and prevent accidental destructive action.

### Dependencies

Data Lifecycle Baseline; Embedding and Vector Storage Baseline.

### Acceptance Criteria

- [ ] **Given** a creator requests batch deletion, **when** they confirm, **then** source files, derived database rows, vector points, and temporary files are removed or documented exceptions are shown.
- [ ] **Given** deletion fails for one store, **when** the request completes, **then** the user sees a pending or failed deletion state rather than false success.
- [ ] **Given** a creator reads privacy information, **when** they review ClipSense data use, **then** media, transcripts, embeddings, exports, logs, and backups are covered.

### Validation Evidence

- [ ] Integration tests cover deletion across database, filesystem, and Qdrant.
- [ ] UI tests cover confirmation, cancellation, pending deletion, and accessible error handling.
- [ ] Documentation review confirms privacy and retention wording matches implementation.

### Out of Scope

Enterprise legal retention holds and jurisdiction-specific compliance automation.

### Definition of Ready

- [ ] Data stores and retention exceptions are known.
- [ ] Deletion UX and confirmation pattern are approved.

### Definition of Done

- [ ] Deletion is functional, observable, and honestly communicated.
- [ ] Privacy notice matches production behavior.

## API Contract and Schema Validation

**Business Rank:** 042  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Not Implemented

### Purpose

Document and validate the REST API contract for auth, batches, upload, results, storylines, exports, errors, and future lifecycle endpoints.

### Consumer, Business, or Risk-Reduction Value

Stable contracts prevent web/API drift and make release validation repeatable.

### Repository Evidence

There is no OpenAPI spec. Legacy files mention OpenAPI and SDKs that do not exist. Web defines inline TypeScript response types, while API structs define JSON fields.

### Scope and Required Behaviour

Create OpenAPI or equivalent contract, response schemas, error schema, auth requirements, status codes, versioning policy, and contract tests.

### Security, Privacy, Accessibility, or Operational Requirements

Contract must identify protected endpoints and avoid exposing internal file paths in schemas.

### Dependencies

Authenticated API Boundary; Failure Recovery Messaging.

### Acceptance Criteria

- [ ] **Given** the web calls an API endpoint, **when** response shape changes, **then** contract validation catches incompatible changes.
- [ ] **Given** an error response is returned, **when** it is validated, **then** it follows the documented safe error schema.
- [ ] **Given** a protected endpoint is documented, **when** contract is reviewed, **then** auth requirements are explicit.

### Validation Evidence

- [ ] API contract file exists and validates current handlers.
- [ ] Contract tests run in CI.
- [ ] Web client types are generated or checked against the contract.

### Out of Scope

Public third-party SDK release.

### Definition of Ready

- [ ] Contract format is selected.
- [ ] Current endpoint inventory is complete.

### Definition of Done

- [ ] API contract covers MVP and production endpoints.
- [ ] Contract drift is caught automatically.

## Compose Runtime and Performance Validation

**Business Rank:** 043  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Missing Validation

### Purpose

Validate the full Compose runtime, service startup order, health, resource use, and MVP performance envelope.

### Consumer, Business, or Risk-Reduction Value

Production readiness cannot be inferred from unit tests; the services must run together under realistic constraints.

### Repository Evidence

`docker-compose.yml` defines all core services. Local Docker checks could not run because Docker is not on PATH. `docs/MVP_STABILIZATION.md` reports previous daemon availability blockers.

### Scope and Required Behaviour

Run `docker compose config`, build all images, start stack, verify health/readiness, run smoke test, measure upload/processing/export timings, inspect disk and memory use, and document limits.

### Security, Privacy, Accessibility, or Operational Requirements

Runtime logs and artifacts must redact secrets and avoid retaining private test media.

### Dependencies

End-to-End MVP Smoke Test; Container Build Hardening.

### Acceptance Criteria

- [ ] **Given** Docker is installed and running, **when** Compose validation runs, **then** config, build, startup, health, and smoke checks pass.
- [ ] **Given** worker cold start loads models, **when** runtime is measured, **then** startup behavior and expected wait time are documented.
- [ ] **Given** resource limits are exceeded, **when** testing runs, **then** the product fails safely rather than corrupting state.

### Validation Evidence

- [ ] Compose config output, image build logs, health checks, and smoke-test results are stored as release evidence.
- [ ] Performance notes include at least small ZIP processing time and resource use.

### Out of Scope

Cloud load testing and autoscaling.

### Definition of Ready

- [ ] Docker runtime is available.
- [ ] Smoke fixture is ready.

### Definition of Done

- [ ] Full stack runtime is verified and documented.
- [ ] Environment blockers are no longer unresolved for release.

## Responsible AI Quality Evaluation Harness

**Business Rank:** 044  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Not Implemented

### Purpose

Evaluate transcription, summary, classification, clustering, storyline usefulness, confidence/uncertainty communication, and generated-versus-corrected distinctions with repeatable fixtures and human review criteria.

### Consumer, Business, or Risk-Reduction Value

ClipSense value depends on AI output being useful, reviewable, and transparent enough for creators to trust, correct, and challenge without mistaking suggestions for facts.

### Repository Evidence

Current summaries are text truncation; mood and role are keyword heuristics; storylines use KMeans ordering. No evaluation fixtures, metrics, review rubric, confidence representation, generated-versus-corrected labeling, model version records, or regression tests exist.

### Scope and Required Behaviour

Create fixture batches, expected transcript/summary/classification tolerances, storyline review rubric, model version capture, confidence or uncertainty representation, generated-versus-corrected labeling rules, regression comparison, model rollback expectations, and failure thresholds.

### Security, Privacy, Accessibility, or Operational Requirements

Evaluation fixtures must be licensed or synthetic and must not contain private creator data. AI limitation disclosure must be visible wherever generated summaries, classifications, or storylines may affect creator decisions.

### Dependencies

First Clip Analysis Result; Basic Storyline Suggestion.

### Acceptance Criteria

- [ ] **Given** an evaluation fixture is processed, **when** outputs are generated, **then** transcript availability, summary relevance, classification, confidence or uncertainty, and storyline order are scored.
- [ ] **Given** a model or prompt changes, **when** evaluation runs, **then** regressions are visible before release.
- [ ] **Given** AI confidence is low, uncertainty is high, or output is empty, **when** the UI displays results, **then** the creator sees appropriate caveats and can distinguish generated output from user-corrected output.

### Validation Evidence

- [ ] Evaluation harness runs on synthetic or licensed fixtures.
- [ ] Release evidence includes model versions, quality notes, limitation disclosure, and rollback implications.
- [ ] Regression thresholds are documented.

### Out of Scope

Claiming human-level narrative quality, fully automated editorial judgment, or bias-free analysis.

### Definition of Ready

- [ ] Fixture licensing and storage approach are approved.
- [ ] Minimum quality rubric is defined.

### Definition of Done

- [ ] AI quality is measured and regressions are visible.
- [ ] Output caveats are aligned with measured limitations.

## Consumer Deletion and Cleanup Flow

**Business Rank:** 045  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Not Implemented

### User Story

**As a** creator  
**I need** to delete a batch and its generated results  
**So that** I control the lifecycle of footage I uploaded.

### Product Outcome

Creators can delete batches from UI and API, see deletion state, and receive confirmation once source and derived data are removed.

### Business Value

Deletion is a concrete trust feature and a production privacy boundary.

### Repository Evidence

No delete endpoint, UI control, cleanup service, vector cleanup, or deletion tests exist.

### Functional Requirements and Business Rules

Allow deleting only owned batches; require confirmation; remove or mark source archive, processing files, clips, storylines, vectors, and generated exports; handle partial deletion safely.

### Consumer Safety and Trust

Use clear destructive-action copy and avoid accidental deletion.

### Data and State Requirements

Deletion should be transactional where possible and recoverable when multi-store cleanup fails.

### Failure and Fallback Behaviour

If one store fails deletion, expose pending cleanup and retry path without showing false completion.

### Edge Cases

Batch processing in progress, queue job pending, vector store unavailable, file already missing, concurrent delete, and unauthorized delete.

### Security and Privacy Requirements

Delete requires current authorization and logs only deletion metadata.

### Accessibility Requirements

Confirmation dialog or page must support keyboard, focus trap where applicable, error announcement, and non-color-only destructive cues.

### Performance and Reliability Requirements

Large deletions may run async but must give clear status.

### Observability and Operational Requirements

Track deletion requested, completed, failed, retrying, and orphan cleanup.

### Assumptions

Batch-level deletion is required before account-level deletion.

### Dependencies

Privacy Notice, Retention, and Deletion Controls.

### Acceptance Criteria

- [ ] **Given** a creator owns a batch, **when** they confirm deletion, **then** the batch and derived data are removed or placed in a truthful deletion-pending state.
- [ ] **Given** processing is active, **when** deletion is requested, **then** processing is cancelled or blocked from writing new results.
- [ ] **Given** an unauthorized user requests deletion, **when** the API handles it, **then** no data is deleted.

### Validation Evidence

- [ ] API and worker integration tests cover deletion across stores.
- [ ] UI/E2E test covers confirmation, cancellation, success, and failure.
- [ ] Accessibility test covers destructive confirmation.

### Out of Scope

Account deletion and legal hold workflows.

### Definition of Ready

- [ ] Deletion semantics for each store are documented.
- [ ] Cancellation behavior is defined.

### Definition of Done

- [ ] Batch deletion is secure, accessible, and operationally observable.
- [ ] No known orphaned data remains in tested flows.

## Database Migration and Rollback Discipline

**Business Rank:** 046  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Requires Hardening

### Purpose

Replace inline schema creation with controlled migrations, rollback planning, constraints, indexes, and schema test evidence.

### Consumer, Business, or Risk-Reduction Value

Production data must evolve without silent data loss, missing constraints, or unrecoverable deploys.

### Repository Evidence

`migrate()` in `apps/api/main.go` executes `CREATE TABLE IF NOT EXISTS` statements. No migration files, migration runner, down migrations, seed strategy, foreign keys, indexes, or schema version table exists.

### Scope and Required Behaviour

Introduce migration tooling, versioned migration files, rollback rules, data migration tests, foreign keys, indexes, non-null constraints, and release checklist integration.

### Security, Privacy, Accessibility, or Operational Requirements

Migrations must preserve ownership constraints and not dump private data into logs.

### Dependencies

Batch Schema and State Model; API Contract and Schema Validation.

### Acceptance Criteria

- [ ] **Given** a new environment starts, **when** migrations run, **then** schema is created from versioned migrations rather than inline handler code.
- [ ] **Given** a migration fails, **when** deployment checks run, **then** the application does not continue with a partially unknown schema.
- [ ] **Given** rollback is required, **when** release notes are reviewed, **then** reversible and irreversible migrations are identified.

### Validation Evidence

- [ ] Migration tests run against a disposable Postgres database.
- [ ] Schema constraints for ownership and referential integrity are verified.

### Out of Scope

Multi-tenant sharding and zero-downtime migration automation unless required by deployment scale.

### Definition of Ready

- [ ] Migration tool is selected.
- [ ] Current schema is captured as baseline migration.

### Definition of Done

- [ ] Database schema is versioned, constrained, and release-managed.
- [ ] Inline migration code is removed or limited to migration runner invocation.

## Backup and Restore for Postgres and Qdrant

**Business Rank:** 047  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Defective for Current Runtime

### Purpose

Provide tested backup and restore procedures for structured data and vector data used by ClipSense.

### Consumer, Business, or Risk-Reduction Value

Creators and operators need recoverability after accidental deletion, deployment failure, disk loss, or data corruption.

### Repository Evidence

`scripts/database/backup.sh` and `restore.sh` copy `data/clipsense.db`, but Compose uses Postgres. Qdrant uses a named volume but has no snapshot or restore script. No backup tests exist.

### Scope and Required Behaviour

Create Postgres backup/restore using appropriate tools, Qdrant snapshot/restore procedure, backup retention, encryption/storage policy, restore drills, and RPO/RTO targets.

### Security, Privacy, Accessibility, or Operational Requirements

Backups contain private media metadata and generated data; access, retention, encryption, and deletion exceptions must be documented.

### Dependencies

Database Migration and Rollback Discipline; Embedding and Vector Storage Baseline.

### Acceptance Criteria

- [ ] **Given** Postgres is the active database, **when** backup runs, **then** it captures current schema and data without relying on a SQLite file.
- [ ] **Given** Qdrant stores embeddings, **when** backup policy is applied, **then** vector snapshots are included or explicitly excluded with product impact documented.
- [ ] **Given** a restore drill runs, **when** restored data is checked, **then** batches, clips, storylines, and vector references are consistent.

### Validation Evidence

- [ ] Backup and restore scripts pass against a disposable Compose stack.
- [ ] Restore drill evidence includes integrity checks.
- [ ] Retention and encryption policy is documented.

### Out of Scope

Cross-region disaster recovery until production scale requires it.

### Definition of Ready

- [ ] RPO/RTO targets are agreed.
- [ ] Backup destination and retention are selected.

### Definition of Done

- [ ] Backups and restores are tested for Postgres and vector data.
- [ ] SQLite-oriented scripts are replaced or marked obsolete.

## Observability and Incident Diagnostics

**Business Rank:** 048  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 13  
**Current Implementation Assessment:** Partially Implemented; Requires Hardening

### Purpose

Add structured logs, privacy-respecting product analytics, metrics, and diagnostic events that allow operators to understand upload, queue, worker, model, database, vector, export, and cleanup behavior.

### Consumer, Business, or Risk-Reduction Value

When processing fails or slows, operators need evidence to fix it before creators lose trust.

### Repository Evidence

API uses chi logger and `log.Println`. Worker prints `[worker]` messages. Health endpoint exists. `infra/scripts/monitoring/setup-prometheus.sh` is a TODO placeholder. No metrics, traces, correlation IDs, dashboards, alerts, or redaction policy exists.

### Scope and Required Behaviour

Implement structured logging, correlation IDs, redaction, metrics for request latency, upload failures, queue depth, job duration, processing stage duration, failed jobs, disk usage, export failures, cleanup, and privacy-respecting feature adoption. Define alerts and incident runbook.

### Security, Privacy, Accessibility, or Operational Requirements

Logs, metrics, and product analytics must not include secrets, tokens, private media content, full transcripts, or unnecessary personal data.

### Dependencies

Runtime Health Checks; Failure Recovery Messaging.

### Acceptance Criteria

- [ ] **Given** a batch moves through upload, queue, processing, and export, **when** logs are inspected, **then** events can be correlated without exposing private content.
- [ ] **Given** queue depth or failed jobs exceed thresholds, **when** monitoring runs, **then** operators receive an actionable alert.
- [ ] **Given** an incident occurs, **when** the runbook is followed, **then** health, logs, metrics, and recent deploy evidence are available.
- [ ] **Given** product usage is reviewed, **when** analytics are inspected, **then** operators can see adoption of upload, review, export, deletion, and recovery flows without exposing media or transcript content.

### Validation Evidence

- [ ] Observability tests or smoke checks verify correlation ID propagation.
- [ ] Metrics dashboard or documented query set covers MVP services.
- [ ] Product analytics review verifies adoption signals are aggregated and privacy-respecting.
- [ ] Redaction review verifies sensitive fields are excluded.

### Out of Scope

Distributed tracing everywhere unless justified by operational complexity.

### Definition of Ready

- [ ] Key operational questions are listed.
- [ ] Metrics, analytics events, and log fields are selected.

### Definition of Done

- [ ] Operators can diagnose common MVP incidents and understand feature adoption.
- [ ] Sensitive data is excluded from observability and analytics outputs.

## Usage Limits, Cost Controls, Abuse Prevention, and Media Rights

**Business Rank:** 049  
**Release Stage:** Production-Ready MVP  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Not Implemented

### Purpose

Define and enforce the minimum usage, storage, processing, abuse-prevention, acceptable-use, and media-rights controls required before real-user use.

### Consumer, Business, or Risk-Reduction Value

ClipSense processes large media and expensive AI workloads. Without limits and rights communication, one user, abusive actor, or mistaken upload can exhaust storage, queue capacity, model budget, or legal trust.

### Repository Evidence

The API has a multipart parsing limit and the worker has ZIP extraction limits, but there is no user quota, batch quota, storage quota, processing-time limit, registration abuse control, acceptable-use notice, or media-rights confirmation. Compose uses unbounded local volumes.

### Scope and Required Behaviour

Set MVP limits for upload size, extracted size, file count, active batches, queued jobs, retained storage, daily processing, and retry attempts. Add user-facing media-rights confirmation and acceptable-use guidance for uploaded or linked footage. Add operator controls for disabling abusive accounts or pausing intake.

### Security, Privacy, Accessibility, or Operational Requirements

Limits must protect availability without exposing other users' usage. Media-rights and acceptable-use copy must be accessible before upload and link intake. Rate and quota enforcement should align with OWASP API guidance on unrestricted resource consumption and automated abuse.

### Dependencies

Upload Request Size and Type Enforcement; Privacy Notice, Retention, and Deletion Controls; Observability and Incident Diagnostics.

### Acceptance Criteria

- [ ] **Given** a creator exceeds an active batch, storage, upload, retry, or processing quota, **when** they submit work, **then** the request is rejected or deferred with a clear, safe message.
- [ ] **Given** a creator uploads or submits media, **when** the intake form is used, **then** they receive an accessible media-rights and acceptable-use confirmation appropriate to the source type.
- [ ] **Given** abuse or runaway processing is detected, **when** an operator intervenes, **then** intake can be paused or limited without corrupting existing batches.

### Validation Evidence

- [ ] API tests cover upload, active-batch, retry, and storage quota enforcement.
- [ ] UI tests cover accessible quota and media-rights messaging.
- [ ] Operational evidence shows quota metrics and intake pause behavior.

### Out of Scope

Billing, paid plans, and enterprise policy management.

### Definition of Ready

- [ ] MVP resource limits and acceptable-use wording are approved.
- [ ] Metrics needed for quota enforcement are available.

### Definition of Done

- [ ] Real-user use is bounded by documented, tested limits.
- [ ] Media-rights and acceptable-use expectations are visible before intake.

## Controlled MVP Stabilization and Release Feedback

**Business Rank:** 050  
**Release Stage:** MVP Stabilization  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Not Implemented

### Purpose

Create a controlled stabilization loop after production-ready MVP validation and before public production launch.

### Consumer, Business, or Risk-Reduction Value

This prevents known release blockers from being deferred while still reserving space for defects and usability issues discovered through real runtime evidence.

### Repository Evidence

`docs/MVP_STABILIZATION.md` records earlier stabilization findings, but there is no formal controlled-release feedback process, issue triage rule, evidence checklist, or go/no-go loop before public release.

### Scope and Required Behaviour

Define controlled test cohort, release-gate evidence, defect severity rules, support intake, telemetry and privacy-respecting product analytics review, rollback rehearsal, accessibility review, privacy/deletion verification, performance notes, and public-release go/no-go criteria.

### Security, Privacy, Accessibility, or Operational Requirements

No known release-blocking security, privacy, accessibility, data-loss, deletion, or false-completion defect may be waived silently. Stabilization evidence must avoid private media content.

### Dependencies

Compose Runtime and Performance Validation; Responsible AI Quality Evaluation Harness; Usage Limits, Cost Controls, Abuse Prevention, and Media Rights.

### Acceptance Criteria

- [ ] **Given** a controlled MVP validation run completes, **when** findings are reviewed, **then** each defect is classified as release-blocking, stabilization-required, or post-release with rationale.
- [ ] **Given** a release-blocking defect is found, **when** public launch is considered, **then** the launch is blocked until the defect is fixed or formally risk-accepted by the owner.
- [ ] **Given** creator feedback identifies confusing status, export, privacy, or AI-output behavior, **when** stabilization planning runs, **then** the backlog is updated without duplicating existing items.

### Validation Evidence

- [ ] Controlled-release checklist includes security, privacy, accessibility, queue, deletion, backup, export, AI quality, usage-limit, product-analytics, and support evidence.
- [ ] Stabilization review notes link each finding to an existing or new backlog item.

### Out of Scope

Public growth experiments and broad beta marketing.

### Definition of Ready

- [ ] Production-ready MVP evidence is complete.
- [ ] Controlled validation cohort and feedback channels are defined.

### Definition of Done

- [ ] Stabilization findings are triaged and release-blocking issues are resolved or formally accepted.
- [ ] Public production launch has a clear go/no-go record.

## Release Versioning and Support Runbooks

**Business Rank:** 051  
**Release Stage:** Public Production Release  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Not Implemented

### Purpose

Define versioning, changelog, release notes, rollback decision points, support troubleshooting, service-status communication, and known limitation communication.

### Consumer, Business, or Risk-Reduction Value

Creators and operators need predictable releases and clear support paths when behavior changes.

### Repository Evidence

`scripts/build/package-release.sh` packages Docker images but assumes Docker availability and Compose image discovery. No changelog, release notes template, versioning policy, support guide, or rollback playbook exists.

### Scope and Required Behaviour

Establish version naming, release checklist, release notes, rollback criteria, support triage, service-status communication, known limitations, incident escalation, and maintenance release policy.

### Security, Privacy, Accessibility, or Operational Requirements

Release notes and status communication must disclose material privacy, retention, security, accessibility, availability, and AI behavior changes.

### Dependencies

Controlled MVP Stabilization and Release Feedback; Compose Runtime and Performance Validation; Observability and Incident Diagnostics.

### Acceptance Criteria

- [ ] **Given** a production release is proposed, **when** the checklist is reviewed, **then** tests, security, accessibility, privacy, backup, and rollback evidence are required.
- [ ] **Given** a support issue arrives, **when** the runbook is followed, **then** operators can gather safe diagnostics without requesting private media unnecessarily.
- [ ] **Given** a release changes AI output behavior, **when** notes are published, **then** creators see relevant limitations or migration guidance.
- [ ] **Given** an outage or degraded processing occurs, **when** status communication is needed, **then** creators receive clear availability and recovery information without exposing incident-sensitive details.

### Validation Evidence

- [ ] Release checklist exists and maps to CI, smoke, security, accessibility, backup, and docs evidence.
- [ ] Support runbook includes common upload, processing, auth, export, deletion, quota, and status-communication failures.

### Out of Scope

Paid support tiers and service-level agreements.

### Definition of Ready

- [ ] Production release gates are known.
- [ ] Common support scenarios are identified.
- [ ] Service-status owner, channel, and update cadence are identified.

### Definition of Done

- [ ] Release and support processes are documented and usable.
- [ ] Production release cannot be approved without required evidence.
- [ ] Status communication can be issued and updated during degraded service.

## Media Format and Codec Compatibility Matrix

**Business Rank:** 052  
**Release Stage:** Post-MVP Release 1  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Partially Implemented; Missing Validation

### Purpose

Define supported containers, codecs, audio tracks, duration limits, and failure behavior for ZIP and direct media intake.

### Consumer, Business, or Risk-Reduction Value

Creators need predictable media compatibility before uploading large files.

### Repository Evidence

`ALLOWED_VIDEO_EXTENSIONS` includes `.mp4`, `.mov`, `.mkv`, `.webm`, and `.avi`. FFmpeg is used for extraction and probing. No codec matrix, fixture suite, or user-facing compatibility docs exist.

### Scope and Required Behaviour

Document supported containers, expected common codecs, unsupported codecs, audio requirements, duration/file size limits, test fixtures, and fallback messages.

### Security, Privacy, Accessibility, or Operational Requirements

Compatibility tests should use licensed or synthetic media fixtures.

### Dependencies

First Clip Analysis Result.

### Acceptance Criteria

- [ ] **Given** a listed supported format is uploaded, **when** fixture validation runs, **then** audio extraction and duration probing succeed or failure is documented.
- [ ] **Given** an unsupported codec is uploaded, **when** processing fails, **then** the creator sees an unsupported media reason.
- [ ] **Given** docs list supported media, **when** constants or tests change, **then** docs remain in sync.

### Validation Evidence

- [ ] Fixture test matrix covers supported and unsupported media cases.
- [ ] User docs match tested compatibility.

### Out of Scope

Transcoding every unsupported media type.

### Definition of Ready

- [ ] Candidate media formats are selected from creator use cases.
- [ ] Fixture licensing is cleared.

### Definition of Done

- [ ] Compatibility matrix is tested and documented.
- [ ] Upload guidance reflects real media behavior.

## Direct Video Upload Intake

**Business Rank:** 053  
**Release Stage:** Post-MVP Release 1  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Not Implemented

### User Story

**As a** creator  
**I need** to upload a single video directly  
**So that** I do not have to wrap one clip in a ZIP archive.

### Product Outcome

The intake model accepts a supported direct video file and normalizes it into the same batch/clip pipeline as ZIP entries.

### Business Value

Direct upload is a natural next step after ZIP MVP and reduces friction for small projects.

### Repository Evidence

Docs state direct upload is intended future work. Current web dropzone accepts `application/zip`; API stores all uploads as `.zip`; worker expects `extract_zip`.

### Functional Requirements and Business Rules

Support allowlisted video extensions/codecs from the compatibility matrix, content validation, size limits, metadata extraction, same storage ownership, same processing job type, and same review/export outputs.

### Consumer Safety and Trust

UI must distinguish direct-video and ZIP limits and avoid implying unsupported codecs will work.

### Data and State Requirements

Normalize direct video into a source asset and one clip or batch entry without creating a separate downstream pipeline.

### Failure and Fallback Behaviour

Unsupported codec or extraction failure creates safe failure reason and retry guidance.

### Edge Cases

No audio, unsupported codec, huge file, wrong extension, duplicate upload, interrupted transfer, and mobile browser upload.

### Security and Privacy Requirements

Use the same upload security controls as ZIP plus media content sniffing.

### Accessibility Requirements

File chooser and validation guidance must remain accessible.

### Performance and Reliability Requirements

Large direct files must respect request-size and timeout limits or wait for resumable upload support.

### Observability and Operational Requirements

Track intake type and validation outcomes.

### Assumptions

Direct upload should reuse the unified source pipeline.

### Dependencies

Media Format and Codec Compatibility Matrix; Upload Request Size and Type Enforcement; Usage Limits, Cost Controls, Abuse Prevention, and Media Rights.

### Acceptance Criteria

- [ ] **Given** a creator uploads a supported single video, **when** validation passes, **then** ClipSense creates a batch and processes it through the same worker result model.
- [ ] **Given** a direct video has no audio, **when** processing runs, **then** the result explains transcript limitations instead of failing ambiguously.
- [ ] **Given** a ZIP-only downstream assumption remains, **when** direct upload is tested, **then** the test fails until normalization is fixed.

### Validation Evidence

- [ ] API/worker integration tests cover direct video intake.
- [ ] UI/E2E test covers direct upload guidance, validation, and result review.

### Out of Scope

Resumable upload and link import.

### Definition of Ready

- [ ] Media compatibility matrix is approved.
- [ ] Unified source model supports non-ZIP assets.

### Definition of Done

- [ ] Direct videos use the shared pipeline and MVP controls.
- [ ] ZIP and direct intake are both tested.

## Resumable Large Uploads

**Business Rank:** 054  
**Release Stage:** Post-MVP Release 1  
**Fibonacci Estimate:** 13  
**Current Implementation Assessment:** Not Implemented

### User Story

**As a** creator with large footage batches  
**I need** interrupted uploads to resume  
**So that** a network hiccup does not force me to restart from zero.

### Product Outcome

Large upload support becomes more reliable through a resumable protocol or chunked upload design.

### Business Value

Creator footage is often large; resumability improves real-world usability after MVP safety is stable.

### Repository Evidence

Current upload uses one multipart request and `ParseMultipartForm`. No chunking, resumable state, upload sessions, checksum validation, or partial cleanup exists.

### Functional Requirements and Business Rules

Create upload sessions, chunk validation, resumable offsets, final assembly, checksums, expiration, cleanup, and authorization per upload session.

### Consumer Safety and Trust

Show upload progress and recovery instructions without hiding failed chunks.

### Data and State Requirements

Store upload session state separately from processable batch until final assembly validates.

### Failure and Fallback Behaviour

Expired sessions clean up partial chunks. Corrupt chunks are rejected and retryable.

### Edge Cases

Out-of-order chunks, duplicate chunks, expired session, quota exceeded, lost browser tab, and concurrent finalization.

### Security and Privacy Requirements

Chunks require authentication, ownership, size limits, checksum verification, and path confinement.

### Accessibility Requirements

Progress and retry controls must be announced and keyboard operable.

### Performance and Reliability Requirements

Chunk sizes and concurrency must protect server memory and disk.

### Observability and Operational Requirements

Track active sessions, partial disk usage, expired sessions, and completion rate.

### Assumptions

Resumability is post-MVP because the MVP can limit upload size.

### Dependencies

Direct Video Upload Intake; Upload Request Size and Type Enforcement.

### Acceptance Criteria

- [ ] **Given** an upload is interrupted, **when** the creator resumes within the expiration window, **then** only missing chunks are resent.
- [ ] **Given** a chunk checksum fails, **when** the API validates it, **then** the chunk is rejected without corrupting the final asset.
- [ ] **Given** an upload session expires, **when** cleanup runs, **then** partial chunks are removed safely.

### Validation Evidence

- [ ] API tests cover chunk upload, resume, duplicate chunks, checksum failure, and finalization.
- [ ] UI/E2E test covers interruption and resume.
- [ ] Operational metrics cover active and expired sessions.

### Out of Scope

Peer-to-peer upload acceleration and cloud object storage unless chosen as implementation path.

### Definition of Ready

- [ ] Resumable protocol is selected.
- [ ] Storage and cleanup design is approved.

### Definition of Done

- [ ] Large uploads can resume safely and accessibly.
- [ ] Partial upload cleanup is tested.

## User-Provided Link Intake

**Business Rank:** 055  
**Release Stage:** Post-MVP Release 1  
**Fibonacci Estimate:** 13  
**Current Implementation Assessment:** Not Implemented

### User Story

**As a** creator  
**I need** to submit a supported video link  
**So that** ClipSense can process source footage without requiring a local upload.

### Product Outcome

ClipSense can resolve approved link sources into stored source assets and the same processing pipeline.

### Business Value

Link intake can reduce friction, but it introduces higher security, privacy, reliability, and rights risks than local files.

### Repository Evidence

Product docs mention links as future intent. No API route, resolver, allowlist, SSRF protection, download service, rights messaging, or storage integration exists.

### Functional Requirements and Business Rules

Support only justified link types, validate URLs, block private networks and unsafe schemes, fetch with size/time limits, record source metadata, and require creator acknowledgement of rights/permissions.

### Consumer Safety and Trust

Clearly communicate link support limitations and failure reasons.

### Data and State Requirements

Store resolved media as source assets under the same lifecycle as uploads.

### Failure and Fallback Behaviour

Handle unavailable links, redirects, auth-required content, unsupported platforms, download timeout, and content-size mismatch.

### Edge Cases

Redirect chains, private IPs, localhost URLs, expiring links, age-restricted content, playlists, and partial downloads.

### Security and Privacy Requirements

Implement SSRF protection, scheme allowlist, DNS/IP validation, redirect policy, rate limits, and safe downloader isolation.

### Accessibility Requirements

URL form requires label, validation, and accessible error messages.

### Performance and Reliability Requirements

Downloads must be bounded and cancellable.

### Observability and Operational Requirements

Track source type, validation category, fetch duration, size, and failure reason without logging sensitive query tokens.

### Assumptions

Link intake is introduced only after local upload safety and lifecycle controls are mature.

### Dependencies

Usage Limits, Cost Controls, Abuse Prevention, and Media Rights; Data Lifecycle Baseline; Security and Privacy Baseline.

### Acceptance Criteria

- [ ] **Given** a supported public video link, **when** validation and fetch succeed, **then** ClipSense stores a source asset and processes it through the unified pipeline.
- [ ] **Given** a URL resolves to a private network address, **when** validation runs, **then** the request is blocked before fetching content.
- [ ] **Given** a link download fails, **when** status is shown, **then** the creator receives a safe, actionable reason.

### Validation Evidence

- [ ] Security tests cover SSRF, redirects, unsafe schemes, and private IP ranges.
- [ ] Integration tests cover supported, unsupported, timeout, and oversized link sources.

### Out of Scope

Platform account integrations and DRM bypass.

### Definition of Ready

- [ ] Supported link sources are justified.
- [ ] SSRF and rights-risk controls are designed.

### Definition of Done

- [ ] Link intake is secure, bounded, lifecycle-aware, and unified with the media pipeline.
- [ ] Rights and limitations are communicated to creators.

## Transcript Review, Time-Coded Playback, and Correction

**Business Rank:** 056  
**Release Stage:** Post-MVP Release 2  
**Fibonacci Estimate:** 13  
**Current Implementation Assessment:** Not Implemented

### User Story

**As a** creator  
**I need** to review transcripts alongside the related media and correct transcript text  
**So that** downstream summaries, classifications, search, and storylines reflect what was actually said.

### Product Outcome

Creators can view generated transcripts, navigate media by transcript timecodes where available, edit corrections, and trigger dependent analysis updates.

### Business Value

AI transcription is imperfect; correction turns ClipSense into a trustworthy review tool.

### Repository Evidence

API returns `transcript`; UI does not show it. No media preview, transcript segments, time-coded playback, correction model, edit history, re-analysis trigger, or permissions exist.

### Functional Requirements and Business Rules

Display transcript text, support media preview and transcript-linked navigation when segment timing exists, allow correction, save original and corrected versions, record editor/time, mark downstream analysis stale, and reprocess affected summaries/embeddings/storylines.

### Consumer Safety and Trust

Always distinguish generated transcript from creator-corrected transcript.

### Data and State Requirements

Store version history or audit fields sufficient to recover the generated baseline and latest correction.

### Failure and Fallback Behaviour

If re-analysis fails, keep correction and show stale derived outputs.

### Edge Cases

Long transcript, empty transcript, missing timecodes, preview playback failure, multiple editors later, special characters, profanity, non-English text, and reprocessing conflict.

### Security and Privacy Requirements

Transcript edits are private batch data and require authorization.

### Accessibility Requirements

Media controls and text editing controls need labels, keyboard operation, visible focus, captions or transcript alternatives where applicable, save status, and error association.

### Performance and Reliability Requirements

Saving corrections should not block on full batch reprocessing.

### Observability and Operational Requirements

Track correction saved, re-analysis queued, re-analysis failed, and stale output.

### Assumptions

Segment timestamps can be added later if not present in MVP transcript output.

### Dependencies

Batch Results Review; Responsible AI Quality Evaluation Harness.

### Acceptance Criteria

- [ ] **Given** a generated transcript exists, **when** a creator edits and saves it, **then** the corrected transcript is stored without losing the original.
- [ ] **Given** transcript segment timecodes exist, **when** the creator selects a transcript segment, **then** media preview navigates to the corresponding time or reports that linked playback is unavailable.
- [ ] **Given** a correction changes text, **when** dependent analysis exists, **then** summaries, embeddings, and storylines are marked stale or reprocessed.
- [ ] **Given** re-analysis fails, **when** the creator views results, **then** the corrected transcript remains saved and stale derived outputs are identified.

### Validation Evidence

- [ ] API tests cover transcript save, authorization, version fields, and stale analysis state.
- [ ] UI tests cover media preview, transcript-linked navigation, edit, validation, save, cancel, and accessible status.
- [ ] Worker tests cover correction-triggered re-analysis.

### Out of Scope

Collaborative simultaneous transcript editing and full non-linear editing controls.

### Definition of Ready

- [ ] Transcript data model supports corrections.
- [ ] Re-analysis invalidation rules are defined.

### Definition of Done

- [ ] Transcript correction is safe, accessible, and linked to downstream analysis state.
- [ ] Original generated transcript remains recoverable.

## Configurable Clip Classification

**Business Rank:** 057  
**Release Stage:** Post-MVP Release 2  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Partially Implemented; Requires Productization

### Purpose

Allow classification taxonomies such as mood, role, topic, and custom creator tags to evolve beyond fixed heuristics.

### Consumer, Business, or Risk-Reduction Value

Different creator workflows need different labels; configurable classification makes results more useful without requiring a new product.

### Repository Evidence

Worker has hardcoded `classify_mood`, `classify_role`, and topic `'general'`. Docs describe richer roles and moods, but no taxonomy model or configuration exists.

### Scope and Required Behaviour

Define default taxonomy, validation, custom labels where justified, confidence/source, reclassification, UI display, and effect on search/storylines.

### Security, Privacy, Accessibility, or Operational Requirements

User-provided labels must be validated and safely rendered to prevent injection.

### Dependencies

Transcript Review, Time-Coded Playback, and Correction; Responsible AI Quality Evaluation Harness.

### Acceptance Criteria

- [ ] **Given** default classification runs, **when** results appear, **then** mood, role, and topic use documented taxonomy values.
- [ ] **Given** a creator changes classification, **when** saved, **then** the corrected label is distinguished from generated output.
- [ ] **Given** taxonomy changes, **when** old batches are viewed, **then** compatibility behavior is documented and safe.

### Validation Evidence

- [ ] Unit tests cover taxonomy validation and safe rendering.
- [ ] Worker evaluation tests cover classification regressions.
- [ ] UI tests cover generated versus corrected labels.

### Out of Scope

Fully personalized ML models.

### Definition of Ready

- [ ] Default taxonomy is approved.
- [ ] Correction and stale-analysis rules exist.

### Definition of Done

- [ ] Classifications are documented, validated, and user-correctable where supported.
- [ ] Taxonomy evolution has compatibility rules.

## Creator Storyline Reordering, Timeline Notes, and Markers

**Business Rank:** 058  
**Release Stage:** Post-MVP Release 2  
**Fibonacci Estimate:** 13  
**Current Implementation Assessment:** Not Implemented

### User Story

**As a** creator  
**I need** to reorder, annotate, and compare storylines  
**So that** AI suggestions become editable story plans for an editor.

### Product Outcome

Creators can duplicate an AI storyline, reorder clips, add lightweight notes/tags/markers, save a manual version, and export the chosen order.

### Business Value

Creative control is essential; ClipSense should assist, not replace, editorial judgment.

### Repository Evidence

UI displays storylines but offers no editor. Database stores storylines and positions but no owner-edited metadata, timeline review state, notes, tags, markers, versioning, locks, or export selection.

### Functional Requirements and Business Rules

Support manual storyline creation from an AI sequence, reorder operations, title edits, notes, tags, timeline markers, version/source metadata, save validation, export selected storyline, and conflict-safe updates.

### Consumer Safety and Trust

Distinguish AI-generated, creator-edited, and stale storylines.

### Data and State Requirements

Positions must remain unique and contiguous per storyline. Storyline must reference clips from the same batch.

### Failure and Fallback Behaviour

If save fails, preserve unsaved local changes where feasible and show retry.

### Edge Cases

Duplicate clip, missing clip after deletion, concurrent edit later, empty storyline, mobile reorder, keyboard-only reorder, long notes, duplicate marker names, and marker positions after clip duration changes.

### Security and Privacy Requirements

Storyline edits require batch ownership.

### Accessibility Requirements

Reordering, notes, tags, and markers must be possible by keyboard, not only pointer drag/drop.

### Performance and Reliability Requirements

Save should be atomic and avoid corrupting order on partial failure.

### Observability and Operational Requirements

Track manual storyline saved, export selected, and save failure categories.

### Assumptions

Single-user editing is sufficient until collaboration ships.

### Dependencies

Basic Storyline Suggestion; Accessible Results Review and Keyboard Navigation.

### Acceptance Criteria

- [ ] **Given** an AI storyline exists, **when** the creator duplicates and reorders it, **then** a manual storyline is saved with valid positions.
- [ ] **Given** a creator adds notes, tags, or markers to a storyline, **when** the storyline is saved and reopened, **then** those annotations remain associated with the intended clip or time position.
- [ ] **Given** a keyboard-only creator reorders clips, **when** they save, **then** the order persists without pointer drag/drop.
- [ ] **Given** a clip is removed or unavailable, **when** a storyline is loaded, **then** the issue is shown without corrupting saved order.

### Validation Evidence

- [ ] API tests cover create, update order, annotations, authorization, and same-batch validation.
- [ ] UI/E2E tests cover mouse and keyboard reordering plus note/tag/marker editing.
- [ ] Accessibility test covers reorder controls and announcements.

### Out of Scope

Multi-user simultaneous editing and full NLE timeline rendering.

### Definition of Ready

- [ ] Storyline editing data model is defined.
- [ ] Accessible reorder interaction is designed.

### Definition of Done

- [ ] Creators can save and export manual storylines.
- [ ] Reorder behavior is accessible and tested.

## Accessibility Regression Program

**Business Rank:** 059  
**Release Stage:** Post-MVP Release 2  
**Fibonacci Estimate:** 5  
**Current Implementation Assessment:** Not Implemented

### Purpose

Prevent accessibility regressions as search, correction, reordering, settings, collaboration, and exports expand.

### Consumer, Business, or Risk-Reduction Value

Accessibility is not a one-time MVP checklist; ongoing releases need automated and manual evidence.

### Repository Evidence

No accessibility tooling exists in package scripts or CI. No route-level accessibility checklist exists.

### Scope and Required Behaviour

Add automated axe or equivalent checks, keyboard smoke tests, manual review cadence, contrast review, reduced motion checks, and release evidence requirements.

### Security, Privacy, Accessibility, or Operational Requirements

Accessibility test artifacts must not expose private media or transcripts.

### Dependencies

Accessible Results Review and Keyboard Navigation.

### Acceptance Criteria

- [ ] **Given** a UI route is added or changed, **when** CI runs, **then** required accessibility checks run for representative states.
- [ ] **Given** automated checks pass, **when** release readiness is reviewed, **then** manual keyboard and screen-reader notes are still required for high-risk interactions.
- [ ] **Given** an accessibility defect is found, **when** triaged, **then** it is linked to the affected workflow and release risk.

### Validation Evidence

- [ ] CI includes accessibility checks for auth, upload, dashboard, batch detail, and editor flows.
- [ ] Manual review template exists and is used for release gates.

### Out of Scope

Replacing human accessibility review with automation only.

### Definition of Ready

- [ ] Accessibility tooling is selected.
- [ ] Representative UI states are listed.

### Definition of Done

- [ ] Accessibility regression checks are part of normal delivery.
- [ ] Release evidence includes automated and manual accessibility results.

## Searchable Clip Intelligence

**Business Rank:** 060  
**Release Stage:** Post-MVP Release 2  
**Fibonacci Estimate:** 13  
**Current Implementation Assessment:** Not Implemented; Earlier Search Removed

### User Story

**As a** creator  
**I need** to search clips by transcript, topic, role, mood, and semantic similarity  
**So that** I can find useful moments across large batches.

### Product Outcome

Creators can search within a batch and later across projects using text and semantic signals.

### Business Value

Search becomes increasingly valuable as ClipSense handles larger batches and histories.

### Repository Evidence

Qdrant upsert exists in the worker. `docs/MVP_STABILIZATION.md` says a Go API search endpoint was removed because it depended on an incompatible Qdrant client API and was outside the current MVP.

### Functional Requirements and Business Rules

Support text search over transcripts/summaries and vector similarity where justified, scope results by ownership, filter by classification, and show why results matched.

### Consumer Safety and Trust

Search should make clear whether matches are exact text, metadata filter, or semantic similarity.

### Data and State Requirements

Indexing must track model version, stale corrections, deleted batches, and authorization.

### Failure and Fallback Behaviour

If vector search is unavailable, text search may still work with a clear degraded message.

### Edge Cases

Empty query, no results, corrected transcripts, stale embeddings, deleted vectors, large result sets, and cross-user access attempts.

### Security and Privacy Requirements

All search results must be scoped to authorized batches and must not reveal existence of another user's clips.

### Accessibility Requirements

Search input, filters, result count, and result cards must be accessible.

### Performance and Reliability Requirements

Search latency and result limits must be defined before cross-batch search expands.

### Observability and Operational Requirements

Track search latency, result counts, degraded mode, and vector errors without logging sensitive query text unless policy allows.

### Assumptions

Batch-level search should precede global search.

### Dependencies

Embedding and Vector Storage Baseline; Transcript Review, Time-Coded Playback, and Correction.

### Acceptance Criteria

- [ ] **Given** a creator searches within a batch, **when** matching clips exist, **then** only authorized clips are returned with match context.
- [ ] **Given** Qdrant is unavailable, **when** text search can still run, **then** the UI shows degraded semantic search status.
- [ ] **Given** a transcript correction changes content, **when** search indexes update, **then** stale results are avoided or identified.

### Validation Evidence

- [ ] API tests cover text, vector, filters, authorization, and no-result cases.
- [ ] UI tests cover accessible search and filters.
- [ ] Operational tests cover Qdrant unavailable behavior.

### Out of Scope

Cross-workspace discovery and public search.

### Definition of Ready

- [ ] Search scope and ranking expectations are defined.
- [ ] Vector model lifecycle exists.

### Definition of Done

- [ ] Search is authorized, explainable, and resilient to vector degradation.
- [ ] Removed search functionality is replaced only with validated contracts.

## Editor Interoperability Exports

**Business Rank:** 061  
**Release Stage:** Post-MVP Release 2  
**Fibonacci Estimate:** 13  
**Current Implementation Assessment:** Not Implemented

### Purpose

Add professional export formats and documentation that help creators move ClipSense storylines into editing workflows.

### Consumer, Business, or Risk-Reduction Value

Editors need outputs that map to their tools, not just generic CSV.

### Repository Evidence

Docs mention EDL, XML, Adobe, and DaVinci ideas. Current API supports JSON and CSV only. No timecode, clip source path mapping, media relink metadata, or format-specific tests exist.

### Scope and Required Behaviour

Research and implement justified exports such as EDL or XML only when required metadata exists. Include format limitations, timecode assumptions, filename mapping, storyline selection, and validation fixtures.

### Security, Privacy, Accessibility, or Operational Requirements

Exports must not leak server paths and must make private data inclusion clear.

### Dependencies

Creator Storyline Reordering, Timeline Notes, and Markers; Media Format and Codec Compatibility Matrix.

### Acceptance Criteria

- [ ] **Given** a selected storyline has required timing metadata, **when** an editor export is generated, **then** the file follows the documented format constraints.
- [ ] **Given** required timecode or source metadata is missing, **when** export is requested, **then** the UI explains why the format is unavailable.
- [ ] **Given** export files are inspected, **when** paths appear, **then** they reference consumer-meaningful names rather than server filesystem paths.

### Validation Evidence

- [ ] Format fixtures validate generated export syntax.
- [ ] Documentation explains supported editor import workflow and limitations.
- [ ] UI test covers unavailable format messaging.

### Out of Scope

Native plug-ins for Adobe, DaVinci, or Final Cut.

### Definition of Ready

- [ ] Required metadata for selected export format is available.
- [ ] Target editor workflow is validated with users or documented need.

### Definition of Done

- [ ] Export format is tested with fixtures and documented limitations.
- [ ] Unsupported editor claims are not made.

## Project History, Saved Review State, and Reprocessing

**Business Rank:** 062  
**Release Stage:** Post-MVP Release 2  
**Fibonacci Estimate:** 13  
**Current Implementation Assessment:** Not Implemented

### User Story

**As a** creator  
**I need** to revisit projects, resume review where I left off, and reprocess results  
**So that** corrections, model improvements, or failed jobs can produce better outputs without losing my review context.

### Product Outcome

Batch history supports saved review state, reprocessing with clear versioning, and preservation or replacement of prior results.

### Business Value

AI systems evolve; creators need a controlled way to benefit from improvements without losing work.

### Repository Evidence

Dashboard lists batches but no project model, saved review position, reviewed/unreviewed state, reprocess action, processing version, model version display, or result versioning exists.

### Functional Requirements and Business Rules

Support project/batch history, saved review position, reviewed/unreviewed markers, reprocess request, versioned results, cancellation or locking during reprocess, stale result display, and rollback to prior output where feasible.

### Consumer Safety and Trust

Warn before replacing manual corrections or storylines.

### Data and State Requirements

Record processing version, model versions, source asset version, correction version, and generated output version.

### Failure and Fallback Behaviour

If reprocess fails, preserve prior successful results and show new failure.

### Edge Cases

Deleted source asset, changed transcript correction, model unavailable, queue duplicate, concurrent delete, and saved review state that references a removed clip.

### Security and Privacy Requirements

Reprocess requires ownership and follows retention policy.

### Accessibility Requirements

Reprocess controls require confirmation and accessible status.

### Performance and Reliability Requirements

Reprocessing must use idempotent queue behavior and avoid duplicate outputs.

### Observability and Operational Requirements

Track review-state saves, reprocess reason, version, duration, and outcome.

### Assumptions

Project history may start as batch history before richer project grouping.

### Dependencies

Idempotent Processing and Duplicate Job Protection; Responsible AI Quality Evaluation Harness.

### Acceptance Criteria

- [ ] **Given** a creator reviews clips and leaves the batch, **when** they return, **then** saved review position or reviewed state is restored without exposing another user's activity.
- [ ] **Given** a creator requests reprocessing, **when** prior results exist, **then** the system preserves or explicitly replaces them according to chosen policy.
- [ ] **Given** reprocessing fails, **when** the creator views the batch, **then** prior usable results remain available.
- [ ] **Given** model versions changed, **when** results are compared, **then** generated output version is visible.

### Validation Evidence

- [ ] API/worker integration tests cover review-state persistence, reprocess success, failure, and duplicate prevention.
- [ ] UI tests cover saved review state, confirmation, progress, and prior-result preservation.

### Out of Scope

Branching collaborative project history.

### Definition of Ready

- [ ] Result versioning strategy is defined.
- [ ] Idempotent processing is implemented.

### Definition of Done

- [ ] Reprocessing is controlled, versioned, and recoverable.
- [ ] Prior creator work is protected from accidental overwrite.

## Account and Workspace Controls

**Business Rank:** 063  
**Release Stage:** Post-MVP Release 3  
**Fibonacci Estimate:** 13  
**Current Implementation Assessment:** Not Implemented

### Purpose

Introduce account settings and workspace/project ownership controls only after single-user MVP data boundaries are stable.

### Consumer, Business, or Risk-Reduction Value

Teams need organized access, but adding workspaces before authorization, deletion, and audit foundations would increase risk.

### Repository Evidence

Settings page only shows API URL guidance. No profile, password change, account deletion, workspace, role, invite, or audit model exists.

### Scope and Required Behaviour

Add profile settings, password change, account deletion path, workspace model, role definitions, invitations, project ownership, and audit-safe membership changes when justified.

### Security, Privacy, Accessibility, or Operational Requirements

Authorization must move from user-only to workspace-aware without leaking legacy single-user data.

### Dependencies

Privacy Notice, Retention, and Deletion Controls; User-Owned Batch Authorization.

### Acceptance Criteria

- [ ] **Given** workspace support is enabled, **when** a user accesses a project, **then** role and membership authorize access.
- [ ] **Given** a user leaves a workspace, **when** ownership is transferred or revoked, **then** project access updates predictably.
- [ ] **Given** account deletion is requested, **when** owned data exists, **then** the user sees the impact on projects and batches.

### Validation Evidence

- [ ] Authorization tests cover owner, member, removed member, and non-member access.
- [ ] Migration tests protect existing single-user batches.
- [ ] UI accessibility tests cover settings and invitation flows.

### Out of Scope

Enterprise SSO, billing, and organization-wide compliance.

### Definition of Ready

- [ ] Product need for multi-user workspace is validated.
- [ ] Migration strategy from user-owned batches is designed.

### Definition of Done

- [ ] Workspace controls are secure, tested, and backward-compatible.
- [ ] Account settings reflect actual lifecycle behavior.

## Collaboration and Sharing Controls

**Business Rank:** 064  
**Release Stage:** Post-MVP Release 3  
**Fibonacci Estimate:** 13  
**Current Implementation Assessment:** Not Implemented

### User Story

**As a** creator or editor in a team  
**I need** to share batches or storylines with controlled permissions  
**So that** review and editing decisions can happen collaboratively.

### Product Outcome

Authorized collaborators can view, comment, or edit according to role without exposing private data broadly.

### Business Value

Collaboration is valuable for teams, but it must be built on mature access control.

### Repository Evidence

No collaboration, sharing links, comments, roles, invitations, or audit trail exist.

### Functional Requirements and Business Rules

Define share targets, roles, permissions, expiration, revocation, comments or notes if justified, audit events, and notification strategy.

### Consumer Safety and Trust

Sharing must be explicit, revocable, and visible to owners.

### Data and State Requirements

Shared access must reference workspace/project ownership and not duplicate media unnecessarily.

### Failure and Fallback Behaviour

Revoked users lose access promptly. Expired links fail safely.

### Edge Cases

Owner deleted, collaborator removed, stale browser session, export permission mismatch, and shared batch deletion.

### Security and Privacy Requirements

Avoid public unguessable links for private media unless risk is accepted and expiration is enforced.

### Accessibility Requirements

Sharing controls, permission tables, and confirmation flows must be keyboard and screen-reader accessible.

### Performance and Reliability Requirements

Permission checks must remain efficient as collaborators grow.

### Observability and Operational Requirements

Audit sharing created, changed, revoked, and access denied events.

### Assumptions

Collaboration follows workspace controls and is not required for MVP.

### Dependencies

Account and Workspace Controls.

### Acceptance Criteria

- [ ] **Given** an owner shares a batch with view permission, **when** the collaborator opens it, **then** they can view but not edit or delete.
- [ ] **Given** access is revoked, **when** the collaborator refreshes or calls the API, **then** private data is no longer returned.
- [ ] **Given** sharing permissions change, **when** audit logs are reviewed, **then** the change is recorded without private media content.

### Validation Evidence

- [ ] Authorization tests cover role matrix and revocation.
- [ ] UI tests cover share, permission change, revoke, and accessible confirmation.

### Out of Scope

Real-time co-editing and external public publishing.

### Definition of Ready

- [ ] Workspace authorization is implemented.
- [ ] Collaboration scope is validated with target users.

### Definition of Done

- [ ] Sharing is explicit, revocable, audited, and permission-tested.
- [ ] Collaboration does not weaken privacy boundaries.

## Model Upgrade and Compatibility Management

**Business Rank:** 065  
**Release Stage:** Mature Product  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Not Implemented

### Purpose

Manage ASR, embedding, classification, and storyline model changes safely over time.

### Consumer, Business, or Risk-Reduction Value

Model upgrades can improve output but also change transcripts, search, storylines, and exports; creators need predictable compatibility.

### Repository Evidence

Worker loads `WHISPER_MODEL` and a hardcoded sentence-transformer. No model version persisted in clip, transcript, embedding, or storyline rows. No compatibility matrix, migration plan, or evaluation gate exists.

### Scope and Required Behaviour

Persist model versions, define upgrade gates, reprocessing policy, embedding compatibility, evaluation regression thresholds, rollback plan, and user-facing change notes.

### Security, Privacy, Accessibility, or Operational Requirements

Model downloads and caches must be controlled and not leak private media to external providers unless explicitly configured and disclosed.

### Dependencies

Responsible AI Quality Evaluation Harness; Embedding and Vector Storage Baseline.

### Acceptance Criteria

- [ ] **Given** a model version changes, **when** new analysis is generated, **then** outputs record the model or algorithm version used.
- [ ] **Given** embeddings from incompatible models exist, **when** search or storyline generation runs, **then** incompatible vectors are not mixed silently.
- [ ] **Given** a model upgrade regresses evaluation fixtures, **when** release gates run, **then** the upgrade is blocked or explicitly accepted.

### Validation Evidence

- [ ] Model version fields are tested in worker output.
- [ ] Evaluation report compares old and new model behavior.
- [ ] Rollback plan documents cache and vector implications.

### Out of Scope

Training custom proprietary models from user data.

### Definition of Ready

- [ ] Current model inventory is complete.
- [ ] Evaluation harness exists.

### Definition of Done

- [ ] Model upgrades are versioned, evaluated, and reversible where practical.
- [ ] Compatibility behavior is documented for users and operators.

## Maintenance, Deprecation, and Responsible Retirement

**Business Rank:** 066  
**Release Stage:** Maintenance and Evolution  
**Fibonacci Estimate:** 8  
**Current Implementation Assessment:** Partially Implemented Through Legacy Files; Requires Process

### Purpose

Create a sustainable process for maintaining ClipSense, deprecating obsolete functionality, migrating data, and retiring features safely.

### Consumer, Business, or Risk-Reduction Value

A mature product must evolve without leaving broken docs, unsupported dependencies, stale models, or confusing legacy behavior.

### Repository Evidence

Legacy blueprint files remain alongside current implementation. `rebuild-clipsense-kanban.ps1` is deleted in the current working tree before this task and was not modified. There is no deprecation policy, compatibility policy, dependency upgrade cadence, data migration policy, or retirement runbook.

### Scope and Required Behaviour

Define maintenance cadence, dependency updates, security patch windows, deprecation notices, migration plans, compatibility tests, data export before removal, obsolete-code cleanup, archived docs, and responsible product retirement.

### Security, Privacy, Accessibility, or Operational Requirements

Retirement plans must preserve user data rights, deletion guarantees, export access, accessibility of migration notices, and supportability.

### Dependencies

Release Versioning and Support Runbooks; Model Upgrade and Compatibility Management.

### Acceptance Criteria

- [ ] **Given** a feature is deprecated, **when** users rely on it, **then** notice, migration path, timeline, and data impact are documented.
- [ ] **Given** a dependency becomes unsupported or vulnerable, **when** maintenance review runs, **then** upgrade, mitigation, or replacement work is ranked.
- [ ] **Given** obsolete docs or code remain, **when** cleanup is planned, **then** current behavior is preserved or migration is completed before removal.

### Validation Evidence

- [ ] Maintenance policy covers dependencies, docs, schemas, models, exports, and support.
- [ ] Deprecation checklist includes user communication, tests, migration, rollback, and deletion impact.

### Out of Scope

Immediate deletion of legacy files without review.

### Definition of Ready

- [ ] Release and support processes exist.
- [ ] Current legacy artifacts are inventoried.

### Definition of Done

- [ ] ClipSense has a documented maintenance and retirement process.
- [ ] Obsolete functionality can be removed without surprising users or operators.

## 9. Release Plan

| Release Stage | Business Ranks | Release Boundary |
| --- | --- | --- |
| Product Foundation | 001-007 | Product, security, privacy, accessibility, environment, testing, CI, and documentation baselines are defined before feature delivery. |
| Functional Prototype | 008-020 | A creator can access the app, understand ZIP intake, upload safely, create a batch, enqueue processing, safely extract media, and see the first useful result with accessible upload/status basics. This is not public-release-ready. |
| Functional MVP | 021-029 | A useful end-to-end ZIP journey exists: owned data, result review, vector-backed basic storyline, authenticated export, lifecycle baseline, recovery messaging, and health checks. This supports controlled internal validation only. |
| Production-Ready MVP | 030-049 | The functional MVP is hardened for controlled real-user use: secrets/session/export/state/queue/idempotency/accessibility/E2E/dependency/container/config/privacy/deletion/contract/runtime/AI/migration/backup/observability/quota/abuse/media-rights controls are complete. |
| MVP Stabilization | 050 | Controlled-release evidence is reviewed, release-blocking defects are resolved or formally accepted, and public-launch go/no-go criteria are recorded. |
| Public Production Release | 051 | Release versioning, support runbooks, rollback, incident communication, service-status communication, consumer limitation notes, and operational ownership are ready. |
| Post-MVP Release 1 | 052-055 | Intake expands to media compatibility, direct video, resumable upload, and user-provided links only after MVP safety and lifecycle controls exist. |
| Post-MVP Release 2 | 056-062 | Product intelligence and creator workflow expand through transcript-linked playback/correction, configurable classification, timeline/storyline notes, accessibility regression, search, editor exports, and reprocessing. |
| Post-MVP Release 3 | 063-064 | Account/workspace and collaboration controls are introduced after single-user authorization and deletion are mature. |
| Mature Product and Maintenance | 065-066 | Model compatibility, long-term maintenance, deprecation, migration, and responsible retirement processes govern ongoing evolution. |

## 10. Coverage and Traceability Matrix

| Coverage Area | Backlog Coverage | Current Coverage Assessment |
| --- | --- | --- |
| Product discovery | 001, 002 | Partially documented; contradictions require cleanup. |
| Consumer journey | 008, 010, 012, 014, 016, 018, 019, 022, 025, 026, 028, 050, 051 | Implemented in thin ZIP prototype form; needs production-ready hardening and evidence. |
| User interface | 008, 010, 014, 016, 020, 022, 026, 031, 036, 045, 056, 058, 060, 061, 064 | Basic UI exists; accessibility, media preview, saved state, and session handling are incomplete. |
| API | 009, 012, 014, 015, 021, 026, 028, 032, 042, 045, 049, 053, 055 | Core endpoints exist; contracts, errors, quotas, lifecycle, and export auth need work. |
| Worker | 017, 018, 019, 023, 024, 025, 033, 035, 044, 052, 062, 065 | Worker exists; integration, failure, idempotency, quality, and model-version gaps remain. |
| Media processing | 018, 019, 023, 052, 053, 054, 055, 056, 061 | ZIP/video processing exists; compatibility, direct intake, playback, and professional export metadata are future. |
| Database | 011, 021, 033, 041, 045, 046, 047, 056, 058, 062 | Inline schema exists; migrations, constraints, deletion, correction, annotations, and backup need production work. |
| Queue | 017, 034, 035, 049, 062 | Redis list exists; reliable queue semantics and usage controls are missing. |
| Vector storage | 024, 041, 047, 060, 065 | Qdrant upsert exists; lifecycle, snapshots, search, and model versioning are incomplete. |
| Authentication | 008, 009, 030, 031 | Basic JWT auth exists; recovery, verification, defaults, and browser sessions need hardening. |
| Authorization | 009, 021, 026, 045, 063, 064 | User filtering exists; cross-user tests and future workspace roles are missing. |
| Security | 003, 008, 012, 015, 018, 030, 031, 034, 038, 039, 040, 049, 055 | ZIP safety partly tested; broader auth, abuse, dependency, and link-intake hardening remain. |
| Privacy | 003, 027, 041, 045, 047, 049, 063, 066 | Retention, deletion, privacy communication, personal data export, and backup exceptions are not implemented. |
| Accessibility | 004, 020, 036, 045, 056, 058, 059, 060, 064 | No validation evidence yet; media controls, timeline editing, and collaboration need accessibility coverage. |
| AI quality | 019, 025, 044, 056, 057, 065 | Heuristic outputs exist; responsible AI evaluation, confidence, feedback, and rollback are absent. |
| Reliability | 016, 017, 029, 033, 034, 035, 037, 043, 048, 050 | Health partly exists; queue, runtime validation, stabilization, and recovery gaps remain. |
| Performance | 015, 019, 043, 049, 054, 060 | No measured performance or cost envelope exists. |
| Observability | 028, 029, 033, 048, 049, 050, 051 | Logs and health are basic; metrics, alerts, support diagnostics, and launch evidence are absent. |
| Testing | 006, 018, 023, 037, 038, 042, 043, 044, 050, 059 | Unit tests exist; integration/E2E/security/accessibility/runtime/AI validation gaps remain. |
| CI/CD | 007, 037, 038, 039, 043, 050, 051 | CI exists; production gates and release evidence are incomplete. |
| Deployment | 039, 040, 043, 050, 051 | Compose exists; production deployment evidence and rollback ownership are not ready. |
| Backup | 047 | Current scripts mismatch Postgres runtime. |
| Recovery | 008, 028, 034, 035, 047, 048, 050, 051 | Account recovery, job recovery, backup recovery, and support recovery are mostly unimplemented. |
| Documentation | 001, 002, 005, 007, 010, 041, 049, 051, 066 | Documentation exists but requires hierarchy, truth cleanup, privacy, acceptable-use, and support content. |
| Operations | 029, 043, 047, 048, 049, 050, 051 | Operational maturity incomplete. |
| Support | 028, 048, 050, 051, 066 | Support runbooks and service communication are absent. |
| Retention | 027, 041, 045, 047 | Not implemented. |
| Deletion | 041, 045, 063, 066 | Not implemented. |
| Maintenance | 038, 051, 065, 066 | No mature maintenance policy yet. |
| Deprecation | 002, 066 | Legacy artifacts exist; deprecation process missing. |

## 11. Dependency Map

| Enables | Blocks or Enables |
| --- | --- |
| Product Vision and Consumer Boundaries | All ranked work, especially 002-007. |
| Security and Privacy Baseline | 008, 009, 012, 021, 030, 031, 040, 041, 049, 055. |
| Accessibility Baseline | 020, 022, 036, 045, 056, 058, 059, 060, 064. |
| Repository and Environment Baseline | 006, 007, 037, 039, 043. |
| Batch Schema and State Model | 012, 014, 016, 021, 028, 033, 046. |
| Secure ZIP Batch Upload | 013, 015, 017, 018, 023. |
| Safe ZIP Extraction and Resource Limits | 019 and all downstream clip analysis. |
| First Clip Analysis Result | 022, 023, 024, 025, 044, 052. |
| Embedding and Vector Storage Baseline | 025, 041, 047, 060, 065. |
| Export Processed Batch Results | 032, 037, 061. |
| Data Lifecycle Baseline | 041, 045, 047, 055, 066. |
| Reliable Queue Acknowledgement and Dead-Letter Handling | 035, 062. |
| Idempotent Processing and Duplicate Job Protection | 062 and safe retry/reprocess capabilities. |
| Container Build Hardening | 040, 043, production-ready MVP readiness. |
| Privacy Notice, Retention, and Deletion Controls | 045, 049, 063, 066. |
| API Contract and Schema Validation | 046, post-MVP client stability. |
| Responsible AI Quality Evaluation Harness | 056, 057, 062, 065. |
| End-to-End MVP Smoke Test | Production-ready MVP approval and regression gating. |
| Usage Limits, Cost Controls, Abuse Prevention, and Media Rights | 050, 053, 055 and public release readiness. |
| Controlled MVP Stabilization and Release Feedback | 051 and public launch approval. |
| Model Upgrade and Compatibility Management | Future AI upgrades and mature maintenance; reprocessing uses earlier AI quality evidence first. |
| Account and Workspace Controls | Collaboration and sharing controls. |

## 12. Existing Implementation Mapping

| Current Repository Capability or File | Backlog Items | Assessment |
| --- | --- | --- |
| `apps/web/src/app/page.tsx` login/register UI | 008, 031 | Partially Implemented; lacks labels, logout, expiry handling, and hardened session strategy. |
| `apps/web/src/hooks/useAuthToken.ts` | 008, 031, 032 | Requires Hardening; stores JWT in `localStorage`, has unused clear function. |
| `apps/api/main.go` auth handlers and middleware | 008, 009, 021, 030 | Partially Implemented; needs validation, safe errors, cross-user tests, and production secret checks. |
| `apps/api/main.go` inline `migrate()` | 011, 046 | Requires Hardening; no migrations, rollback, constraints, foreign keys, or indexes. |
| `createBatch` upload handling | 012, 014, 015, 017 | Partially Implemented; saves archive and enqueues job, but validation, cleanup, and limits need hardening. |
| `loadBatch`, `listBatches`, `getBatch` | 014, 021, 022 | Partially Implemented; user filtering exists, cross-user and partial-state tests missing. |
| `exportBatch` | 026, 032, 061 | Partially Implemented; API supports JSON/CSV, browser CSV path defective. |
| `apps/api/main_test.go` | 006, 009, 029 | Verified for limited cases; coverage is narrow. |
| `apps/api/ai_worker/main.py` | 017, 018, 019, 023, 024, 025, 033, 035, 044, 062, 065 | Partially Implemented; real processing exists, but failure, idempotency, integration, model quality, reprocessing, and versioning gaps remain. |
| `apps/api/ai_worker/zip_safety.py` | 018 | Verified for core ZIP safety tests; additional malformed cases and API alignment needed. |
| `apps/api/ai_worker/tests/test_zip_safety.py` | 018 | Verified; 7 tests passed locally. |
| `docker-compose.yml` | 005, 029, 039, 043, 047 | Partially Implemented; local stack defined, production hardening and runtime validation missing. |
| `apps/api/Dockerfile` | 039 | Requires Hardening; root runtime and `go mod tidy` during build. |
| `apps/api/ai_worker/Dockerfile` | 039 | Defective Risk; does not copy `zip_safety.py` even though `main.py` imports it. |
| `apps/web/Dockerfile` | 039 | Requires Hardening; uses `npm install`, root runtime, no scan evidence. |
| `.github/workflows/ci.yml` | 007, 038, 043 | Partially Implemented; service checks exist but no security/accessibility/E2E gates. |
| `.env.example` | 005, 030, 040 | Partially Implemented; documents env values, needs mode-specific validation and secret handling. |
| `scripts/database/backup.sh` and `restore.sh` | 047 | Defective for current runtime; scripts target SQLite while Compose uses Postgres. |
| `infra/scripts/monitoring/setup-prometheus.sh` | 048 | Not Implemented; placeholder TODO. |
| `infra/scripts/deploy/deploy.sh` and build scripts | 043, 051 | Partial local wrappers; not production deployment or release evidence. |
| `docs/MVP_STABILIZATION.md` | Many early items | Useful evidence; accurately notes several current limitations. |
| `CS.md`, `CS.txt`, `apps/app`, `apps/package/packages`, `apps/web/src/source`, `apps/api/apis`, `1807ish/*` | 001, 002, 066 | Legacy or aspirational; must not be treated as implemented capability. |
| Missing root README | 005, 007, 051 | Not Implemented. |
| Missing delete endpoints and cleanup jobs | 027, 041, 045 | Not Implemented. |
| Missing OpenAPI/contract | 042 | Not Implemented. |
| Missing account verification and recovery flows | 008 | Not Implemented. |
| Missing quota, cost-control, abuse-prevention, and media-rights controls | 049 | Not Implemented. |
| Missing controlled-release feedback and go/no-go loop | 050 | Not Implemented. |
| Missing media preview, transcript-linked playback, and correction state | 056 | Not Implemented. |
| Missing saved review state and reprocessing controls | 062 | Not Implemented. |
| Missing E2E/accessibility/security scans | 037, 038, 059 | Not Implemented or blocked by environment. |

## 13. Research References

These sources materially influenced the backlog:

- OWASP File Upload Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html
- OWASP API Security Top 10 2023: https://owasp.org/API-Security/editions/2023/en/0x00-header/
- OWASP Authentication Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- OWASP Forgot Password Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html
- OWASP Session Management Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
- OWASP JSON Web Token Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html
- OWASP API4:2023 Unrestricted Resource Consumption: https://owasp.org/API-Security/editions/2023/en/0xa4-unrestricted-resource-consumption/
- OWASP Automated Threats to Web Applications: https://owasp.org/www-project-automated-threats-to-web-applications/
- NIST SP 800-218 Secure Software Development Framework: https://csrc.nist.gov/publications/detail/sp/800-218/final
- NIST SP 800-63-4 Digital Identity Guidelines: https://pages.nist.gov/800-63-4/
- NIST SP 800-63B-4 Authentication and Authenticator Management: https://pages.nist.gov/800-63-4/sp800-63b.html
- NIST Privacy Framework: https://www.nist.gov/privacy-framework
- NIST AI Risk Management Framework: https://www.nist.gov/itl/ai-risk-management-framework
- W3C Web Content Accessibility Guidelines 2.2: https://www.w3.org/TR/WCAG22/
- W3C Media Accessibility User Requirements: https://www.w3.org/TR/media-accessibility-reqs/
- W3C WebVTT: The Web Video Text Tracks Format: https://www.w3.org/TR/webvtt1/
- W3C WAI Making Audio and Video Media Accessible, transcripts guidance: https://www.w3.org/WAI/media/av/transcripts/
- WAI Forms Tutorial: https://www.w3.org/WAI/tutorials/forms/
- Next.js deployment documentation: https://nextjs.org/docs/app/building-your-application/deploying
- Go `net/http` package documentation including request body limiting support: https://pkg.go.dev/net/http
- Python `zipfile` documentation and extraction warnings: https://docs.python.org/3/library/zipfile.html
- PostgreSQL backup and restore documentation: https://www.postgresql.org/docs/current/backup.html
- Redis persistence documentation: https://redis.io/docs/latest/operate/oss_and_stack/management/persistence/
- Redis Streams documentation for acknowledged processing patterns: https://redis.io/docs/latest/develop/data-types/streams/
- Qdrant snapshots documentation: https://qdrant.tech/documentation/concepts/snapshots/
- Docker multi-stage build documentation: https://docs.docker.com/build/building/multi-stage/
- Dockerfile reference for `USER` and build behavior: https://docs.docker.com/reference/dockerfile/
- GitHub Actions security hardening documentation: https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions
- FFmpeg official documentation: https://ffmpeg.org/documentation.html
- FFmpeg `ffprobe` documentation: https://ffmpeg.org/ffprobe.html
- npm audit command documentation: https://docs.npmjs.com/cli/v10/commands/npm-audit
- npm ci command documentation: https://docs.npmjs.com/cli/v10/commands/npm-ci
- Adobe Premiere Pro supported file formats and import/export documentation: https://helpx.adobe.com/premiere-pro/using/supported-file-formats.html
- Adobe Premiere Pro EDL export documentation: https://helpx.adobe.com/premiere/desktop/render-and-export/export-files/export-a-project-as-an-edl-file.html
- Apple Final Cut Pro XML interchange documentation: https://developer.apple.com/documentation/professional_video_applications/final_cut_pro_xml

## 14. Backlog Quality Audit

- [x] The backlog begins with product and engineering foundations.
- [x] The backlog reconstructs the entire journey rather than beginning at the repository's current state.
- [x] Existing functionality is represented at its correct lifecycle position.
- [x] Existing code is assessed for hardening, not automatically accepted as complete.
- [x] Every title is plain and unique.
- [x] No title contains `CS-US`, `US`, or an artificial ID.
- [x] Every item has a unique business rank.
- [x] Items appear in strict business-rank order.
- [x] Every mandatory dependency names an earlier item, with no dependency on a later rank.
- [x] No later-ranked item remains a hidden prerequisite for an earlier delivery item.
- [x] Every estimate uses an allowed Fibonacci value.
- [x] No item exceeds `13`.
- [x] Every acceptance criterion uses `- [ ]`.
- [x] Validation evidence uses `- [ ]`.
- [x] Definition of Ready uses `- [ ]`.
- [x] Definition of Done uses `- [ ]`.
- [x] No duplicated outcomes remain; overlapping risks are consolidated into scoped items.
- [x] Every item is independently valuable or independently verifiable.
- [x] No repeated label blocks are used.
- [x] Boilerplate is minimized and item evidence is tied to ClipSense repository behavior.
- [x] Consumer safety is addressed through upload, state, AI transparency, errors, session, export, retention, and deletion items.
- [x] Security requirements are precise for auth, upload, session, queue, containers, dependencies, link intake, and secrets.
- [x] Privacy and data lifecycle are covered through account recovery, retention, deletion, backup, vector cleanup, product analytics limits, and privacy notice items.
- [x] Accessibility is covered in baseline, upload, results, deletion, reordering, search, and regression work.
- [x] AI and processing quality are addressed through first result, embeddings, storyline, evaluation, correction, classification, and model compatibility items.
- [x] Failure and recovery paths are addressed through state, failure reasons, queue reliability, idempotency, backups, observability, and support runbooks.
- [x] Testing is proportional to risk and includes unit, integration, E2E, security, accessibility, contract, runtime, and AI evaluation needs.
- [x] CI/CD and production operations are covered through CI foundation, E2E, dependency remediation, containers, config, runtime validation, release gates, and runbooks.
- [x] Functional prototype, functional MVP, production-ready MVP, stabilization, and public production release boundaries are distinct.
- [x] Functional MVP is usable for controlled validation only and is not described as public-release-ready.
- [x] Production-ready MVP requirements include usage limits, cost controls, abuse prevention, media-rights communication, observability, and AI quality evidence.
- [x] Public production release requirements include release versioning, support runbooks, rollback criteria, incident communication, service-status communication, and limitation notes.
- [x] Post-MVP releases are progressive and do not label advanced intake, correction, search, collaboration, or editor interoperability as MVP.
- [x] Mature product requirements are defined through model compatibility and maintenance/retirement practices.
- [x] Maintenance, migration, deprecation, and product evolution are covered.
- [x] Media preview and transcript-linked review needs are evaluated before transcript correction is considered complete.
- [x] Account recovery, verification, deletion, and downloadable personal data needs are represented.
- [x] Email, push, and background-completion notifications were evaluated and kept post-MVP unless controlled-release evidence makes them release-blocking.
- [x] Product analytics and support/status operations are represented without collecting private media or transcript content.
- [x] Usage, storage, processing cost, abuse, and acceptable-use/media-rights controls are represented before public launch.
- [x] Professional editing interoperability has been evaluated and kept post-MVP through editor export and interchange work.
- [x] Current implementation and future intent are clearly separated.
- [x] Repository evidence supports implementation assessments.
- [x] Research sources are recorded accurately and are limited to sources that influenced backlog requirements.
- [x] The final audit is limited to conditions verified during this refinement pass.

## Backlog Classification Summary

These counts are intentionally cross-cutting except for the total item count and consumer-story count. A single backlog item can legitimately contribute to security, accessibility, QA, DevOps, or maintenance at the same time.

- Total backlog items: 66
- Consumer or operator stories: 23
- Technical enablers: 21
- Defects: 7
- Security and privacy items: 22
- Accessibility-focused items: 4
- QA and validation items: 10
- DevOps and operational items: 11
- Documentation and maintenance items: 7

### Classification by Rank

- Consumer or operator stories: 008, 010, 012, 014, 016, 019, 020, 022, 025, 026, 028, 031, 036, 037, 045, 053, 054, 055, 056, 058, 060, 062, 064.
- Technical enablers: 001, 005, 006, 009, 011, 013, 015, 017, 018, 023, 024, 027, 029, 035, 042, 044, 046, 052, 057, 063, 065.
- Defects: 002, 030, 032, 033, 038, 039, 047.
- Security and privacy items: 003, 008, 009, 012, 015, 018, 021, 027, 030, 031, 034, 038, 039, 040, 041, 045, 047, 049, 053, 055, 063, 064.
- Accessibility-focused items: 004, 020, 036, 059.
- QA and validation items: 006, 018, 023, 037, 038, 042, 043, 044, 050, 059.
- DevOps and operational items: 007, 017, 029, 039, 040, 043, 047, 048, 049, 050, 051.
- Documentation and maintenance primary items: 001, 002, 005, 007, 049, 051, 066.

Note: Accessibility requirements also appear inside many consumer stories; the accessibility-focused count lists items whose main reason for existing is accessibility delivery or regression prevention.
