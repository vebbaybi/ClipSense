# ADR 0013: Observability Strategy

Status: Proposed

## Context
Go logs text, Python prints, and health checks cover only API database/Redis dependencies.

## Current Implementation
No trace propagation, metrics, structured schema, collector, or dashboards.

## Problem
Cross-service job failures cannot be correlated or diagnosed reliably.

## Requirements
Job/request correlation, structured logs, metrics, traces, privacy filtering,
local visibility, production export, and bounded overhead.

## Options Considered
OpenTelemetry SDKs/collector; structured logs plus Prometheus; vendor-specific agents.

## Benefits
OpenTelemetry is vendor-neutral and supports Go/Python/web boundaries.

## Costs
Instrumentation, collector operation, storage, sampling, dashboards.

## Risks
Media/transcript data or tokens could leak into telemetry.

## Operational Impact
Adds collector/export configuration and retention responsibilities.

## Windows Impact
Local collector packaging must be optional.

## macOS Impact
Local development should work without background system installation.

## Linux Impact
Collector is production-friendly in containers.

## Web Impact
Browser telemetry requires consent, sampling, and strict data rules.

## Security Impact
Redaction and least-privilege exporters are mandatory.

## Migration Path
Define event fields/correlation IDs, add structured logs, then traces/metrics and collector.

## Reversal Strategy
Retain structured logs and disable exporters/collector.

## Recommendation
Adopt OpenTelemetry conventions incrementally after Sprint 1 runtime repair.

## Approval Required
Security, operations, API, processor, and web owners.
