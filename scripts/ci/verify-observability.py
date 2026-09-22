"""Hosted integration evidence, without model inference or private inputs."""
import base64
import io
import json
import os
import subprocess
import time
import urllib.error
import urllib.request
import uuid
import zipfile
from pathlib import Path

evidence = Path(os.getenv('EVIDENCE_DIR', 'artifacts/docker-runtime'))
correlation = str(uuid.uuid4())

def http(url, data=None, headers=None):
    request = urllib.request.Request(url, data=data, headers=headers or {})
    try:
        with urllib.request.urlopen(request, timeout=5) as response:
            return response.status, response.read(), response.headers
    except urllib.error.HTTPError as error:
        return error.code, error.read(), error.headers

for _ in range(60):
    try:
        status, body, _ = http('http://localhost:9090/api/v1/targets')
        targets = json.loads(body)['data']['activeTargets']
        if status == 200 and len(targets) == 2 and all(t['health'] == 'up' for t in targets):
            break
    except (OSError, KeyError):
        pass
    time.sleep(2)
else:
    raise RuntimeError('Prometheus did not scrape both services')
evidence.joinpath('prometheus-targets.json').write_bytes(body)
auth = base64.b64encode(('admin:' + os.environ['GRAFANA_ADMIN_PASSWORD']).encode()).decode()
status, dashboard, _ = http('http://localhost:3001/api/dashboards/uid/clipsense-operations', headers={'Authorization':'Basic ' + auth})
assert status == 200 and len(json.loads(dashboard)['dashboard']['panels']) >= 10
evidence.joinpath('dashboard.json').write_bytes(dashboard)

credentials = {'email':f'fixture-{uuid.uuid4()}@example.test', 'password':'test-only-password-9182'}
status, body, _ = http('http://localhost:8080/api/auth/register', json.dumps(credentials).encode(), {'Content-Type':'application/json'})
assert status == 200
token = json.loads(body)['token']
status, _, _ = http('http://localhost:8080/api/batches', headers={'Authorization':'Bearer invalid-secret-sentinel'})
assert status == 401
archive = io.BytesIO()
with zipfile.ZipFile(archive, 'w') as output:
    output.writestr('fixture.mp4', b'not actual video; validation must reject')
boundary = 'fixture-boundary'
subprocess.run(['docker','compose','stop','worker'], check=True, capture_output=True, timeout=40)
body = (f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="fixture.zip"\r\nContent-Type: application/zip\r\n\r\n'.encode() + archive.getvalue() + f'\r\n--{boundary}--\r\n'.encode())
status, result, response_headers = http('http://localhost:8080/api/batches', body, {'Authorization':'Bearer '+token, 'Content-Type':'multipart/form-data; boundary='+boundary, 'X-Correlation-ID':correlation})
assert status == 200 and response_headers['X-Correlation-ID'] == correlation
batch = json.loads(result)['batch']['id']
# Run the actual job consumer boundary in the worker image, without loading models.
# This is isolated integration, not proof of the long-running ML consumer or full E2E.
program = '''
import json, os, redis
from main import execute_job
from prometheus_client import generate_latest
r = redis.from_url(os.environ['REDIS_URL'])
item = r.blpop('jobs:batch', timeout=5)
assert item is not None
data = json.loads(item[1])
assert data['correlation_id'] == %r and data['id'] == %r
execute_job(data)
print(generate_latest().decode())
''' % (correlation, batch)
try:
    result = subprocess.run(['docker','compose','run','--rm','--no-deps','-T','worker','python','-'], input=program, text=True, capture_output=True, check=True, timeout=60)
finally:
    subprocess.run(['docker','compose','start','worker'], check=True, capture_output=True, timeout=40)
assert 'job.claimed' in result.stderr and 'job.failed' in result.stderr and correlation in result.stderr
assert 'clipsense_events_total{event="job.failed"} 1.0' in result.stdout
worker_events = [json.loads(line) for line in result.stderr.splitlines() if line.startswith('{')]
assert all(e['correlation_id'] == correlation for e in worker_events)
evidence.joinpath('worker-correlation.jsonl').write_text('\n'.join(json.dumps(e) for e in worker_events))
evidence.joinpath('worker-fixture-metrics.txt').write_text(result.stdout)
logs = subprocess.run(['docker','compose','logs','--no-color','api'], capture_output=True, text=True, check=True).stdout
assert correlation in logs and 'job.enqueued' in logs
assert 'dependency.degraded' in logs and 'dependency.recovered' in logs
for secret in (token, credentials['password'], credentials['email'], 'invalid-secret-sentinel', os.environ['JWT_SECRET']):
    assert secret not in logs + result.stderr
status, metrics, _ = http('http://localhost:9090/api/v1/query?query=clipsense_events_total')
assert status == 200
evidence.joinpath('prometheus-events.json').write_bytes(metrics)
assert correlation not in metrics.decode() and batch not in metrics.decode()
status, rejected, _ = http('http://localhost:8080/api/batches', b'not-multipart', {'Authorization':'Bearer '+token, 'Content-Type':'text/plain'})
assert status == 400
for _ in range(20):
    status, metrics, _ = http('http://localhost:9090/api/v1/query?query=clipsense_events_total')
    samples = json.loads(metrics)['data']['result']
    observed = {s['metric']['event']:float(s['value'][1]) for s in samples if s['metric'].get('job') == 'clipsense-api'}
    if observed.get('upload.rejected', 0) > 0 and observed.get('authorization.denied', 0) > 0:
        break
    time.sleep(2)
else:
    raise RuntimeError('Prometheus did not observe rejection metrics')
evidence.joinpath('prometheus-events.json').write_bytes(metrics)
print(json.dumps({'result':'PASS', 'correlation_id':correlation, 'batch_id':batch, 'scope':'isolated real queue/DB/worker validation; not model processing'}))
