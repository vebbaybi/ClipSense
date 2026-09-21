# ADR 0014: Media Storage Strategy

Status: Proposed

## Context
The API and worker share a local Docker volume containing uploads and processing files.

## Current Implementation
Filesystem paths are stored in batches; there is no cleanup, retention, object abstraction,
or desktop/cloud distinction.

## Problem
Media is large, sensitive, failure-prone, and unsuitable for implicit indefinite retention.

## Requirements
Local/cloud modes, explicit asset identity, streaming, integrity, quotas, cleanup,
encryption, deletion, resumability, and future sync.

## Options Considered
Local filesystem adapter; S3-compatible object storage; provider object storage;
database blobs; hybrid asset interface.

## Benefits
An explicit asset contract permits local filesystem and object-store adapters.

## Costs
Metadata model, lifecycle jobs, migration, credentials, multipart upload support.

## Risks
An abstraction added before lifecycle requirements can become empty architecture.

## Operational Impact
Cloud object storage adds buckets, policies, cost monitoring, and backups.

## Windows Impact
Paths, locks, OneDrive, disk quotas, and long filenames require testing.

## macOS Impact
Sandbox permissions and APFS behavior require testing.

## Linux Impact
Container volumes/object stores are straightforward but still need quotas.

## Web Impact
Clients receive authorized URLs/streams, never server filesystem paths.

## Security Impact
Encryption, signed access, malware/media validation, retention, and deletion are central.

## Migration Path
First create asset records and lifecycle rules around current files; add object storage later.

## Reversal Strategy
Export assets plus metadata back to a filesystem adapter.

## Recommendation
Define lifecycle and asset identity in Sprint 1/2; do not add object storage in Sprint 0.

## Approval Required
Product, security, operations, API, processor, and future desktop owners.
