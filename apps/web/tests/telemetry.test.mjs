import test from 'node:test';
import assert from 'node:assert/strict';
import {telemetryRecord, requestHeaders} from '../src/lib/telemetry.mjs';

test('browser events discard secrets, responses and uncontrolled fields', () => {
  const record = telemetryRecord('upload.rejected', {password:'secret', token:'secret', response:'private transcript', category:'Bearer secret', correlation_id:'bad\nfield'});
  assert.deepEqual(Object.keys(record).sort(), ['event','level','service','timestamp']);
  assert.equal(telemetryRecord('attacker.event', {}), null);
  assert.equal(JSON.stringify(record).includes('secret'), false);
});
test('valid diagnostic identifiers survive; headers contain no authority', () => {
  const headers = requestHeaders();
  const record = telemetryRecord('upload.started', {correlation_id:headers['X-Correlation-ID']});
  assert.equal(record.correlation_id, headers['X-Correlation-ID']);
  assert.equal(Object.keys(headers).length, 2);
});
