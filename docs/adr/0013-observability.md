# ADR 0013: Observability Strategy

Status: Accepted for bounded Work Unit 2.5 implementation; runtime acceptance pending

## Work Unit 2.5 Decision

Use Go log/slog and Python logging behind an allowlisted Kaufman event contract.
Use maintained Prometheus client libraries, a private metrics listener per service,
and an optional local Compose profile containing Prometheus and Grafana only.
No log shipper, third-party browser analytics, collector, or tracing backend now.
Correlation UUIDs flow through the existing Redis job, without schema migration.
Trace IDs are reserved for a future validated OpenTelemetry context; correlation
is not a claim of distributed tracing. Metrics never label individual operations
with user, batch, filename, request, or correlation identifiers.

Browser telemetry is local, opt-in, allowlisted diagnostic events only. Logs are
JSON in all environments to make privacy and parsing behavior deterministic.
Third-party model/library output is outside the application event contract and
must not be enabled in verbose mode on private inputs.

The historical proposal below records alternatives, not implemented capabilities.

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
