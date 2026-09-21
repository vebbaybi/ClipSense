# ADR 0002: Go API Code Generation

Status: Proposed

## Context
The Go API currently defines handlers and public structs by hand.

## Current Implementation
One `main.go` file owns routing, middleware, models, queries, and handlers.

## Problem
Handwritten handlers can drift from a future OpenAPI contract.

## Requirements
Strict interface checking, small runtime footprint, chi compatibility, readable diffs,
and pinned generation.

## Options Considered
oapi-codegen strict server; ogen; manual handlers with contract tests; framework rewrite.

## Benefits
oapi-codegen can retain chi and provide models/interfaces without replacing business logic.

## Costs
Generated code, mapping layers, tool pinning, and contributor learning.

## Risks
Generating transport and domain models together can spread API concerns into persistence.

## Operational Impact
CI gains a generated-drift check; runtime topology is unchanged.

## Windows Impact
Pinned Go tool invocation must work without Bash.

## macOS Impact
Go-native generation is suitable.

## Linux Impact
CI can run the same command.

## Web Impact
No direct effect; shares the OpenAPI source.

## Security Impact
Strict request parsing helps, but authorization remains explicit handler logic.

## Migration Path
Generate interfaces for one stabilized endpoint and wrap existing behavior.

## Reversal Strategy
Return to manual interfaces while preserving contract tests.

## Recommendation
Evaluate oapi-codegen after ADR 0001 is approved.

## Approval Required
API owner approval and a reviewed generated-code policy.
