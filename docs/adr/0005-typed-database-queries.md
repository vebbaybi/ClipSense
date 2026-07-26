# ADR 0005: Typed Go Database Queries

Status: Proposed

## Context
Go and Python issue handwritten SQL against shared tables.

## Current Implementation
Queries are inline in `main.go` and `ai_worker/main.py`.

## Problem
Schema drift and scan/order errors are caught late.

## Requirements
Postgres fidelity, explicit SQL, generated types, transaction support, and low runtime cost.

## Options Considered
sqlc; pgx helpers; ORM; retain handwritten SQL.

## Benefits
sqlc keeps reviewed SQL while generating typed Go access.

## Costs
Query files, generation, mapping, and only partial help for Python.

## Risks
Generated database models can leak into public API models.

## Operational Impact
No new runtime service; CI generation required.

## Windows Impact
Pinned sqlc availability must be solved.

## macOS Impact
Supported by available binaries/container.

## Linux Impact
CI-friendly.

## Web Impact
None directly.

## Security Impact
Parameterization remains explicit; authorization filters require tests.

## Migration Path
Adopt after versioned migrations, starting with read-only batch queries.

## Reversal Strategy
Retain SQL files and return to pgx scanning.

## Recommendation
Use sqlc for Go after ADR 0004; define a separate internal processor write boundary.

## Approval Required
API and data owners.

