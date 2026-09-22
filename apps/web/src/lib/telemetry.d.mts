export function telemetryRecord(event: string, context?: Record<string, unknown>): Record<string, unknown> | null;
export function telemetry(event: string, context?: Record<string, unknown>): void;
export function requestHeaders(): Record<string, string>;
