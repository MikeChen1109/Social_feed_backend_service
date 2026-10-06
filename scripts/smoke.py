#!/usr/bin/env python3
"""Exercise real local services without additional Python packages."""
from concurrent.futures import ThreadPoolExecutor
from threading import Barrier
import json
from http.client import RemoteDisconnected
import os
import time
import uuid
from urllib.request import Request, urlopen
from urllib.error import HTTPError, URLError
BASE = os.getenv('API_BASE_URL', 'http://localhost:2000').rstrip('/')
def request(method, path, payload=None, token=None):
    headers = {'Content-Type': 'application/json'}
    if token:
        headers['Authorization'] = 'Bearer ' + token
    data = json.dumps(payload).encode() if payload is not None else None
    try:
        with urlopen(Request(BASE + path, data=data, headers=headers, method=method), timeout=10) as response:
            return response.status, json.load(response)
    except HTTPError as error:
        return error.code, json.load(error)
def ok(method, path, payload=None, token=None):
    status, body = request(method, path, payload, token)
    assert 200 <= status < 300, f'{method} {path}: HTTP {status}: {body}'
    print(f'PASS {method} {path}')
    return body
for attempt in range(60):
    try:
        status, _ = request('GET', '/')
        if status == 200:
            break
    except (URLError, TimeoutError, ConnectionError, RemoteDisconnected):
        pass
    time.sleep(1)
else:
    raise SystemExit('API unavailable. Run make local-up first, then inspect make local-logs.')
credentials = {'username': 'smoke_' + uuid.uuid4().hex, 'password': 'LocalSmoke123!'}
ok('POST', '/api/user/signup', credentials)
tokens = ok('POST', '/api/user/login', credentials)
other = {'username': 'smoke_other_' + uuid.uuid4().hex, 'password': 'LocalSmoke123!'}
ok('POST', '/api/user/signup', other)
other_tokens = ok('POST', '/api/user/login', other)
feed_id = None
try:
    feed = ok('POST', '/api/feed/create', {'title': 'Local smoke test', 'content': 'Created by scripts/smoke.py'}, tokens['token'])
    feed_id = feed['ID']
    for method in ('PUT', 'DELETE'):
        status, _ = request(method, f'/api/feed/{feed_id}', {'title': 'unauthorized', 'content': 'unauthorized'}, other_tokens['token'])
        assert status == 403, f'Non-author {method} should return 403, got {status}'
    print('PASS non-author update/delete denied')
    ok('PUT', f'/api/feed/{feed_id}', {'title': 'Local smoke test', 'content': 'Updated by author'}, tokens['token'])
    retrieved = ok('GET', f'/api/feed/{feed_id}')
    assert retrieved['Title'] == 'Local smoke test'
    page = ok('GET', '/api/feed/paginated?page=1&limit=10')
    assert any(item['ID'] == feed_id for item in page['data'])
    old_refresh = tokens['refreshToken']
    barrier = Barrier(8)
    def refresh_once(_):
        barrier.wait()
        return request('POST', '/api/user/refresh', {'refreshToken': old_refresh})
    with ThreadPoolExecutor(max_workers=8) as pool:
        outcomes = list(pool.map(refresh_once, range(8)))
    winners = [body for status, body in outcomes if status == 200]
    assert len(winners) == 1, f'Expected one concurrent refresh winner, got {[status for status, _ in outcomes]}'
    assert all(status in (200, 401) for status, _ in outcomes)
    tokens = winners[0]
    print('PASS concurrent refresh: one success, seven rejected')
finally:
    if feed_id is not None:
        ok('DELETE', f'/api/feed/{feed_id}', token=tokens['token'])
ok('POST', '/api/user/logout', {'refreshToken': other_tokens['refreshToken']})
ok('POST', '/api/user/logout', {'refreshToken': tokens['refreshToken']})
status, _ = request('POST', '/api/user/refresh', {'refreshToken': tokens['refreshToken']})
assert status == 401, f'Revoked refresh token should return 401, got {status}'
print('PASS revoked refresh token rejected')
print('All smoke checks passed. Test feed removed; the two unique test accounts remain in the local database.')

for port in (2000, 3000, 4000):
    for path in ('/healthz', '/readyz'):
        with urlopen(f'http://localhost:{port}{path}', timeout=5) as response:
            assert response.status == 200
print('PASS health and readiness for all three services')
with urlopen(BASE + '/metrics', timeout=5) as response:
    metrics = response.read().decode()
assert 'http_requests_total ' in metrics and 'http_request_duration_seconds_count ' in metrics
print('PASS gateway metrics')
