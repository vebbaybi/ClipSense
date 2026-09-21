# ADR 0003: TypeScript API Client Generation

Status: Proposed

## Context
The web app uses one fetch wrapper and page-local response types.

## Current Implementation
`apps/web/src/lib/api/client.ts` adds base URL/auth; upload and export bypass it.

## Problem
Types, auth behavior, and error parsing are inconsistent.

## Requirements
Browser compatibility, typed requests/responses, no forced state library, testable auth,
and deterministic output.

## Options Considered
Orval; openapi-typescript plus a small fetch adapter; generated Axios; handwritten client.

## Benefits
Generation removes repeated page types and catches contract drift.

## Costs
Generator configuration, generated output, runtime adapter decisions.

## Risks
Orval can introduce React Query or Axios coupling if configured carelessly.

## Operational Impact
CI must check generated drift.

## Windows Impact
Node-based generator must work through npm scripts.

## macOS Impact
Expected portable Node workflow.

## Linux Impact
CI is canonical.

## Web Impact
Directly improves request and error consistency.

## Security Impact
Central auth attachment reduces bypasses; token storage is a separate decision.

## Migration Path
Generate plain fetch types/client for stabilized endpoints, then replace page-local types.

## Reversal Strategy
Keep generated types as reference and restore the small handwritten adapter.

## Recommendation
Compare Orval plain-fetch output with openapi-typescript before selection.

## Approval Required
Web and API owners must approve after ADR 0001.
