# ADR 0009: Python Dependency And Quality Tooling

Status: Proposed

## Context
The worker uses pinned top-level `requirements.txt` and unittest.

## Current Implementation
pip installs dependencies; CI compiles two files and runs seven unittest cases.

## Problem
Transitive resolution, linting, typing, auditability, and developer parity are weak.

## Requirements
Python 3.11 target, deterministic locks, Windows/macOS/Linux support, ML packages,
fast CI, formatting/linting, typing, and tests.

## Options Considered
uv + pyproject; Poetry; pip-tools; retain requirements; Ruff, Pyright, pytest.

## Benefits
uv with pyproject can provide fast reproducible resolution; Ruff/Pyright/pytest improve feedback.

## Costs
One-time metadata migration, ML wheel compatibility, contributor tooling.

## Risks
Whisper/torch platform resolution may differ and produce a large lock.

## Operational Impact
Container and CI install commands change together.

## Windows Impact
CPU/GPU wheels and long paths require validation.

## macOS Impact
Apple Silicon wheel/model behavior requires validation.

## Linux Impact
Container remains deployment baseline.

## Web Impact
None directly.

## Security Impact
Locked transitive dependencies and auditing improve supply-chain visibility.

## Migration Path
Model current requirements in pyproject, lock on supported platforms, run both paths once.

## Reversal Strategy
Export a compatible requirements file from the lock.

## Recommendation
Adopt uv, Ruff, Pyright, and pytest in an isolated tooling sprint after runtime repair.

## Approval Required
Processor and operations owners.

