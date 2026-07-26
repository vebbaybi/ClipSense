# ADR 0001: Public API Contract Source

Status: Proposed

## Context
Go structs and page-local TypeScript types currently describe the same HTTP data.

## Current Implementation
Routes and models live in `apps/api/main.go`; React pages repeat selected types.

## Problem
There is no machine-checkable public contract or drift detection.

## Requirements
One reviewed source, HTTP/error semantics, auth declarations, generation validation,
and compatibility for web and future desktop clients.

## Options Considered
OpenAPI-first; Go-code-first schema extraction; handwritten models; GraphQL.

## Benefits
OpenAPI-first can generate server/client adapters and publish interoperable docs.

## Costs
Schema authoring, generator ownership, generated-diff review, and migration work.

## Risks
An incomplete schema can create false confidence or force awkward domain leakage.

## Operational Impact
CI must validate and regenerate deterministically.

## Windows Impact
Generator binaries must be reproducibly available.

## macOS Impact
Generation must produce identical output.

## Linux Impact
CI is the canonical drift check.

## Web Impact
Enables typed, consistent client adapters.

## Security Impact
Auth requirements and safe error shapes become reviewable; secrets remain out of schema.

## Migration Path
Describe existing routes first, add contract tests, then migrate one endpoint at a time.

## Reversal Strategy
Retain existing handlers and remove generation while keeping the schema as documentation.

## Recommendation
Use OpenAPI as the future public contract source after Sprint 1 route/error stabilization.

## Approval Required
API and web owners must approve before adding schema or generators.

