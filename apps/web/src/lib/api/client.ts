import {requestHeaders, telemetry} from '../telemetry.mjs';
const base = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// api is a thin fetch wrapper that applies base URL, auth header, and error handling.
export async function api<T>(path: string, init?: RequestInit, token?: string): Promise<T> {
  const diagnosticHeaders = requestHeaders();
  const failureEvent = path.includes('/export') ? 'export.failed' : 'batch.fetch.failed';
  let res: Response;
  try { res = await fetch(`${base}${path}`, {
    ...init,
    headers: {
      ...(init?.headers || {}),
      ...diagnosticHeaders,
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    cache: 'no-store'
  }); } catch {
    telemetry(failureEvent, {category:'network', correlation_id:diagnosticHeaders['X-Correlation-ID']});
    throw new Error('Request unavailable');
  }
  if (!res.ok) {
    telemetry(failureEvent, {category:res.status === 401 ? 'unauthenticated' : 'unavailable', correlation_id:diagnosticHeaders['X-Correlation-ID']});
    throw new Error(`Request failed (${res.status})`);
  }
  return res.json() as Promise<T>;
}
