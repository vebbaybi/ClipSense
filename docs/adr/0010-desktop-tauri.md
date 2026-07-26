# ADR 0010: Desktop Architecture With Tauri 2

Status: Proposed

## Context
The product intends a Windows, macOS, and Linux local processing workstation.

## Current Implementation
No desktop directory, Rust crate, renderer, installer, or native service exists.

## Problem
A website wrapper alone cannot own folders, local jobs, FFmpeg, hardware, updates, or recovery.

## Requirements
Cross-platform installers, native file access, least privilege, background jobs, updates,
shared contracts/UI where useful, and modest footprint.

## Options Considered
Tauri 2 + Vite/React; Electron; native apps; browser PWA.

## Benefits
Tauri offers native capabilities and smaller distribution than bundled Chromium approaches.

## Costs
Rust skills, WebView variance, signing/notarization, multi-platform CI.

## Risks
Premature desktop work would duplicate unstable web/API contracts.

## Operational Impact
Adds release channels, signing keys, updater infrastructure, and platform CI.

## Windows Impact
WebView2, installer signing, GPU/FFmpeg discovery.

## macOS Impact
WebKit behavior, notarization, sandbox permissions, universal binaries.

## Linux Impact
WebKitGTK/package variability and distribution formats.

## Web Impact
Share contracts/components selectively; do not force web-only routing into desktop.

## Security Impact
Tauri command allowlists and filesystem scopes require strict review.

## Migration Path
Stabilize contracts, prototype one folder-intake shell, then add local job state.

## Reversal Strategy
Keep web/API independent and retire the desktop shell without data loss.

## Recommendation
Tauri 2 with Vite/React is the leading future option, not approved for Sprint 0/1.

## Approval Required
Product, security, UI, operations, and desktop owners.

