# ADR 0008: PostgreSQL pgvector Versus Qdrant

Status: Proposed

## Context
Clip embeddings are written to a dedicated Qdrant service.

## Current Implementation
Python creates/upserts one Qdrant collection; no active search endpoint exists.

## Problem
The service adds operations and consistency failure modes before semantic search is used.

## Requirements
Tenant filtering, vector search quality, backup, transactional expectations, scale,
local/desktop compatibility, and manageable operations.

## Options Considered
Keep Qdrant; PostgreSQL with pgvector; abstract support for both; no vectors in MVP.

## Benefits
pgvector may reduce services and align metadata/backup; Qdrant specializes in vector search.

## Costs
Migration/re-embedding, query benchmarking, operational changes.

## Risks
Choosing without workload data can optimize the wrong bottleneck.

## Operational Impact
pgvector simplifies Compose; Qdrant allows independent scaling.

## Windows Impact
Local Postgres extension availability must be tested.

## macOS Impact
Container/native extension packaging must be tested.

## Linux Impact
Both are well-supported.

## Web Impact
Only API search behavior should be visible.

## Security Impact
Tenant filters must be enforced server-side regardless of store.

## Migration Path
Define vector record contract, benchmark representative datasets, dual-write only temporarily.

## Reversal Strategy
Re-embed from stored transcript/model metadata into the previous store.

## Recommendation
Defer selection until semantic search requirements and benchmarks exist.

## Approval Required
Data, processor, API, operations, and cost owners.
