# ADR 0012: Cross-Platform Task Runner

Status: Proposed

## Context
Current root workflows are Bash scripts plus PowerShell Kanban tools.

## Current Implementation
No root command surface; contributors invoke service commands directly.

## Problem
Windows contributors cannot use Bash scripts consistently and duplicated wrappers can drift.

## Requirements
Windows/macOS/Linux support, readable tasks, dependency checks, no false commands,
CI parity, and low installation burden.

## Options Considered
Taskfile; Make; npm scripts; PowerShell Core; language-specific commands only.

## Benefits
Taskfile can provide a declarative cross-platform command surface.

## Costs
Another prerequisite and migration of existing scripts.

## Risks
Wrapping broken behavior can falsely imply reproducibility.

## Operational Impact
CI and docs would call the same tasks.

## Windows Impact
Strong fit if Task is installed reproducibly.

## macOS Impact
Package/binary availability is good.

## Linux Impact
Suitable for CI.

## Web Impact
Can wrap npm without owning web behavior.

## Security Impact
Tasks must not install silently or expose secrets.

## Migration Path
Keep direct commands authoritative; add only validated tasks after approval.

## Reversal Strategy
Document and invoke underlying commands directly.

## Recommendation
Do not adopt in Sprint 0; evaluate Taskfile after Sprint 1 commands are proven.

## Approval Required
Repository operations and contributor owners.

