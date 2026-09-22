"""HTTP-only Gate 2/3 checks; no media processing or cross-user acceptance."""
import io
import json
import urllib.error
import urllib.request
import uuid
import zipfile

BASE = 'http://localhost:8080/api'


def request(route, expected, data=None, token=None, content_type='application/json'):
    headers = {'Content-Type': content_type}
    if token is not None:
        headers['Authorization'] = 'Bearer ' + token
    req = urllib.request.Request(BASE + route, data=data, headers=headers)
    try:
        response = urllib.request.urlopen(req, timeout=10)
    except urllib.error.HTTPError as exc:
        response = exc
    with response:
        status, body = response.status, response.read()
    assert status == expected, f'{route}: expected {expected}, got {status}'
    if expected >= 400:
        assert b'error' in body and b'Traceback' not in body and b'/data/' not in body
    return json.loads(body)


credentials = json.dumps({'email': f'{uuid.uuid4()}@example.test', 'password': 'gate-test-password'}).encode()
token = request('/auth/register', 200, credentials)['token']
request('/auth/login', 200, credentials)
request('/batches', 200, token=token)
request('/batches', 401)
request('/batches', 401, token='invalid.token')
request('/auth/login', 400, b'{}')

for member in ('../escape.mp4', 'C:escape.mp4', 'notes.txt'):
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, 'w') as archive:
        archive.writestr(member, b'invalid media')
    body = b'--boundary\r\nContent-Disposition: form-data; name="file"; filename="batch.zip"\r\n\r\n' + buf.getvalue() + b'\r\n--boundary--\r\n'
    request('/batches', 400, body, token, 'multipart/form-data; boundary=boundary')
request('/batches', 400, b'broken', token, 'multipart/form-data; boundary=boundary')
request('/batches', 413, b'x' * (1048576 + 1), token, 'multipart/form-data; boundary=boundary')
assert not request('/batches', 200, token=token)['batches']
limited = False
for _ in range(12):
    try:
        request('/auth/login', 400, b'{}')
    except AssertionError as exc:
        assert 'got 429' in str(exc)
        limited = True
        break
assert limited, 'authentication attempt limit not enforced'
print('PASS: live credentials/token, missing/invalid auth, rejected archives/multipart, 413, no rejected batch rows, attempt threshold')
