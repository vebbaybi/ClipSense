# ADR 0011: Local Desktop Processing

Status: Proposed

## Context
Large/private footage benefits from local CPU/GPU/storage and resumability.

## Current Implementation
Only a cloud-shaped Python worker consuming Redis jobs exists.

## Problem
Local execution needs durable state, model/media management, recovery, and controlled sync.

## Requirements
Offline operation, resumable stages, resource limits, GPU detection, disk budgets,
portable project metadata, explicit cloud sync, and shared processing semantics.

## Options Considered
Bundled Python sidecar; Rust controller plus Python models; embedded native inference;
cloud-only processing.

## Benefits
A controlled Python sidecar reuses the ML ecosystem and current pipeline knowledge.

## Costs
Packaging Python/models/FFmpeg, updates, antivirus/signing, support matrix.

## Risks
Arbitrary subprocess/filesystem access and platform-specific ML failures.

## Operational Impact
Desktop gains a local job database, logs, model cache, and cleanup policy.

## Windows Impact
Primary first target; GPU drivers, paths, process recovery, and signing matter.

## macOS Impact
Apple Silicon acceleration and sandbox permissions matter.

## Linux Impact
Distribution/library variability matters.

## Web Impact
Web must tolerate locally processed metadata without direct local-file access.

## Security Impact
Local footage stays local by default; sync must be explicit and scoped.

## Migration Path
Extract processing contracts, make stages idempotent, prototype on Windows after Tauri approval.

## Reversal Strategy
Projects retain portable metadata and can fall back to cloud processing.

## Recommendation
Use a native controller with a constrained Python processing sidecar initially.

## Approval Required
Desktop, processor, security, product, and release owners.

