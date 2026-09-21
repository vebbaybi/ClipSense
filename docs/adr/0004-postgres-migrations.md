# ADR 0004: PostgreSQL Migration Tooling

Status: Proposed

## Context
The API executes `CREATE TABLE IF NOT EXISTS` on startup.

## Current Implementation
Schema statements are embedded in `migrate()` in `apps/api/main.go`.

## Problem
Changes are unversioned, lack rollback, and cannot be audited independently.

## Requirements
Postgres support, ordered SQL, transactional behavior, rollback policy, CLI portability,
and compatibility with CI/containers.

## Options Considered
Goose; golang-migrate; Atlas; application startup DDL.

## Benefits
Goose is small, Go-native, and supports SQL migrations.

## Costs
Tool pinning, initial baseline, deployment ordering, rollback discipline.

## Risks
Incorrectly baselining existing databases can reapply or skip schema.

## Operational Impact
Deployments must run migrations as an explicit step.

## Windows Impact
Go binary/tool command is suitable.

## macOS Impact
Suitable with Go or pinned binary.

## Linux Impact
Suitable for CI and container jobs.

## Web Impact
Indirect; stable schemas improve API behavior.

## Security Impact
Migration credentials should not be broader than necessary.

## Migration Path
Snapshot current DDL as version 1, test fresh/existing database cases, remove startup DDL.

## Reversal Strategy
Restore startup baseline temporarily and preserve SQL migration history.

## Recommendation
Approve Goose for Sprint 1 only after existing-data baseline tests are designed.

## Approval Required
API and operations owners.
