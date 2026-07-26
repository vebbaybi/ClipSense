# ADR 0006: Authentication Strategy

Status: Proposed

## Context
ClipSense will serve web and future desktop clients.

## Current Implementation
Custom email/password registration, bcrypt, 24-hour HS256 JWT, browser localStorage.

## Problem
Password policy, recovery, session revocation, secure storage, and desktop flows are absent.

## Requirements
OIDC, Authorization Code with PKCE, web/desktop support, account recovery, revocation,
minimal secret handling, and local-development usability.

## Options Considered
Managed OIDC provider; self-hosted OIDC; improved custom auth; passkeys-first.

## Benefits
Standards-based OIDC reduces custom credential/security ownership.

## Costs
Provider cost/configuration, callback handling, offline behavior, account migration.

## Risks
Provider lock-in and desktop deep-link complexity.

## Operational Impact
Adds identity-provider dependency and key/issuer monitoring.

## Windows Impact
Desktop loopback/deep-link flow must be tested.

## macOS Impact
Keychain and callback behavior must be tested.

## Linux Impact
Browser launch and secret-service availability vary.

## Web Impact
Prefer secure HttpOnly session handling over localStorage bearer tokens.

## Security Impact
Improves credential boundary; PKCE, state, nonce, issuer, audience, and logout remain required.

## Migration Path
Harden current auth first, select provider, support account linking, then retire passwords.

## Reversal Strategy
Keep current auth behind a temporary local-only configuration during migration.

## Recommendation
Adopt OIDC Authorization Code with PKCE long term; do not select a provider in Sprint 0.

## Approval Required
Product, security, API, web, and future desktop owners.

