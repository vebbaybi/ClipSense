# ADR 0015: Authoritative Docker Runtime Verification

Status: Accepted

## Context

Sprint 1 Work Unit 1 requires repeatable evidence that the Compose configuration,
three application images, six-service runtime, health behavior, canonical routes,
worker environment, and API lifecycle work together. Docker Desktop startup on the
available Windows machines is slow and hardware-dependent, and cannot provide a
consistent routine gate tied automatically to a source commit.

## Decision

GitHub-hosted Ubuntu runners are the authoritative Linux container build and runtime
verification environment. The `docker-runtime` job in `.github/workflows/ci.yml`
owns orchestration, security, triggers, attribution, summary, and artifact retention.
`scripts/ci/verify-docker-runtime.sh` is the single owner of reusable Compose build
and runtime assertions.

The other jobs in `.github/workflows/ci.yml` remain the fast source-quality checks.
Local Windows Docker Desktop runs are optional supplemental platform QA and do not
replace the GitHub gate. Desktop application testing on Windows, macOS, and Linux is
separate because no desktop application exists.

## Consequences

- Work Unit 1 cannot be Done without a successful, non-cancelled run for the exact
  source commit.
- Failures retain sanitized service state, logs, image inventory, HTTP results, and
  shutdown evidence for 14 days.
- Pull requests to `main`, pushes to `main`, and manual branch dispatches can run the
  gate; path filters avoid unrelated documentation-only work.
- The workflow does not publish images, use production secrets, or receive write
  permissions.
- Full media processing and cold model acquisition are not claimed by this boot and
  integration gate. A future manual or scheduled media smoke may own that expensive
  evidence.

## Alternatives Considered

- Windows Docker Desktop as the mandatory gate: rejected because startup time and
  machine state are not reproducible.
- Duplicate per-image and runtime workflows: rejected because they duplicate builds
  and split ownership.
- A third-party integration action: rejected because Compose and official GitHub and
  Docker actions provide the required behavior.

## Reversal

Replace the runner or workflow only after another environment supplies equally
reproducible, commit-attributed build/runtime evidence. Remove this ADR, workflow,
and its single owned script together so no competing gate remains.
