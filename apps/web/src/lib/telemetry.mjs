const events = new Set(['upload.started', 'upload.rejected', 'batch.fetch.failed', 'export.failed', 'application.error']);
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

export function telemetryRecord(event, context = {}) {
  if (!events.has(event)) return null;
  const build = process.env.NEXT_PUBLIC_BUILD_SHA || '';
  const record = { timestamp: new Date().toISOString(), level: 'INFO', service: 'web', environment: process.env.NODE_ENV === 'production' ? 'production' : 'development', version: /^[a-f0-9]{40}$/.test(build) ? build : 'unknown', event };
  if (typeof context.correlation_id === 'string' && uuid.test(context.correlation_id)) record.correlation_id = context.correlation_id;
  if (['network', 'invalid', 'too_large', 'unavailable', 'unauthenticated'].includes(context.category)) record.category = context.category;
  return record;
}

export function telemetry(event, context = {}) {
  const record = telemetryRecord(event, context);
  // Local troubleshooting only. No ingestion endpoint, analytics vendor, or payload capture.
  if (process.env.NEXT_PUBLIC_DIAGNOSTICS === 'true' && record) console.info(JSON.stringify(record));
}

export function requestHeaders() {
  return { 'X-Correlation-ID': crypto.randomUUID(), 'X-Request-ID': crypto.randomUUID() };
}
