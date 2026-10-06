#!/usr/bin/env python3
"""Validate direct API instrumentation, real scrapes, and Grafana provisioning."""
import json
import math
import time
import uuid
from urllib.request import Request, urlopen
from urllib.parse import urlencode

def get(url):
    with urlopen(url, timeout=10) as r:
        return json.load(r)
def query(expr):
    result = get('http://localhost:9090/api/v1/query?' + urlencode({'query': expr}))
    assert result['status'] == 'success', result
    return result['data']['result']
def post(port, path, payload, token=None, method='POST'):
    headers = {'Content-Type': 'application/json'}
    if token:
        headers['Authorization'] = 'Bearer ' + token
    with urlopen(Request(f'http://localhost:{port}{path}', data=json.dumps(payload).encode(), headers=headers, method=method), timeout=10) as r:
        return json.load(r)
def wait_for(check, description):
    for _ in range(40):
        if check():
            print('PASS ' + description)
            return
        time.sleep(1)
    raise AssertionError(description + ' did not become ready')

targets = get('http://localhost:9090/api/v1/targets')['data']['activeTargets']
assert len(targets) == 3 and all(t['health'] == 'up' for t in targets), targets
print('PASS all three Prometheus targets healthy')
credentials = {'username': 'monitor_' + uuid.uuid4().hex, 'password': 'LocalMonitor123!'}
post(4000, '/user/signup', credentials)
token = post(4000, '/user/login', credentials)['token']
wait_for(lambda: bool(query('http_requests_total{service="user-service",route="/user/login",status="200"}')), 'direct login counter scraped')
# Ensure a second login occurs after a baseline scrape, making rate/P95 observable.
before = float(query('http_requests_total{service="user-service",route="/user/login",status="200"}')[0]['value'][1])
post(4000, '/user/login', credentials)
wait_for(lambda: float(query('http_requests_total{service="user-service",route="/user/login",status="200"}')[0]['value'][1]) > before, 'second direct login updates counter')
feed_id = post(3000, '/feed/create', {'title': 'Monitoring test', 'content': 'Direct API call'}, token)['ID']
try:
    get(f'http://localhost:3000/feed/{feed_id}')
    get(f'http://localhost:2000/api/feed/{feed_id}')
    wait_for(lambda: bool(query('http_requests_total{service="feed-service",route="/feed/create",status="200"}')), 'direct create counter scraped')
    for service in ('feed-service', 'api-gateway'):
        wait_for(lambda: bool(query(f'http_requests_total{{service="{service}",route="/feed/:id",status="200"}}')), service + ' uses normalized ID route')
    avg = query('sum(http_request_duration_seconds_sum{service="user-service",route="/user/login"}) / sum(http_request_duration_seconds_count{service="user-service",route="/user/login"})')
    assert avg and float(avg[0]['value'][1]) > 0
    p95 = query('histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket{service="user-service",route="/user/login"}[1m])))')
    assert p95 and math.isfinite(float(p95[0]['value'][1])) and float(p95[0]['value'][1]) > 0, p95
    print('PASS average latency and P95 have real samples')
    doc = get('http://localhost:3001/api/dashboards/uid/social-feed-api')
    dashboard = doc['dashboard']
    assert dashboard['refresh'] == '5s'
    assert dashboard['templating']['list'][0]['current']['value'] == 'user-service'
    for service in ('api-gateway', 'user-service', 'feed-service'):
        for panel in dashboard['panels']:
            for target in panel['targets']:
                query(target['expr'].replace('$service', service).replace('$__rate_interval', '1m'))
    print('PASS provisioned Grafana dashboard, 5s refresh, and every panel query')
finally:
    post(3000, f'/feed/{feed_id}', {}, token, method='DELETE')
print('Monitoring verification passed; test feed removed, unique test account remains.')
