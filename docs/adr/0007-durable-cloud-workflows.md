# ADR 0007: Durable Cloud Workflow Orchestration

Status: Proposed

## Context
Large media jobs must survive worker crashes and eventually support retries/cancellation.

## Current Implementation
Go `RPush` plus Python blocking `BLPOP` on a Redis list.

## Problem
Jobs are removed before acknowledgement and have no durable history or retry policy.

## Requirements
At-least-once semantics, idempotency, leases, retries, cancellation, progress, DLQ,
large-job duration, and Python/Go support.

## Options Considered
Temporal; Redis Streams; Postgres job table; managed queue plus workflow state.

## Benefits
Temporal provides durable workflow history, retries, timers, and multi-language SDKs.

## Costs
Meaningful operational complexity and a new control-plane service.

## Risks
Temporal may be disproportionate for the current single-worker MVP.

## Operational Impact
New service, persistence, monitoring, deployment, and upgrade burden.

## Windows Impact
Local development is simplest through Docker, currently unavailable here.

## macOS Impact
Docker resource use matters.

## Linux Impact
Best-supported production environment.

## Web Impact
Enables reliable progress/cancellation APIs.

## Security Impact
Workflow payloads must avoid secrets and enforce tenant boundaries.

## Migration Path
First define idempotent stages and contracts; prototype Postgres/Redis Streams and Temporal.

## Reversal Strategy
Keep orchestration behind an interface and replay state from durable batch/job records.

## Recommendation
Do not adopt Temporal yet; specify and repair the MVP durability contract first.

## Approval Required
API, processor, operations, and cost owners.

