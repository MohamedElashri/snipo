#!/usr/bin/env python3
"""Seed an EMPTY localhost test instance and measure warm, serial API latency.

Never uses production data. Run against a temporary SNIPO_DB_PATH instance.
Example: python3 scripts/scale-preview.py http://127.0.0.1:18081 --tags 12
"""
import argparse
import json
import math
import statistics
import time
import urllib.parse
import urllib.request
from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('url')
parser.add_argument('--snippets', type=int, default=160)
parser.add_argument('--tags', type=int, default=12)
parser.add_argument('--measure-only', action='store_true')
parser.add_argument('--summary', action='store_true', help='Measure the lightweight browser list representation')
parser.add_argument('--output', type=Path)
args = parser.parse_args()
base = args.url.rstrip('/')
if urllib.parse.urlparse(base).hostname not in ('127.0.0.1', 'localhost', '::1'):
    parser.error('Only localhost test instances are supported.')
if not 1 <= args.tags <= 100 or not 1 <= args.snippets <= 1000:
    parser.error('Use 1–100 tags and 1–1000 snippets.')

def request(path, data=None):
    body = None if data is None else json.dumps(data).encode()
    req = urllib.request.Request(base + '/api/v1' + path, body, {'Content-Type': 'application/json'})
    start = time.perf_counter()
    with urllib.request.urlopen(req, timeout=30) as response:
        payload = response.read()
    return json.loads(payload), (time.perf_counter() - start) * 1000, len(payload)

names = ['docker', 'postgresql', 'python', 'bash', 'nginx', 'backup', 'monitoring', 'security', 'networking', 'kubernetes', 'javascript', 'terraform', 'ansible', 'redis', 'sqlite', 'go', 'rust', 'typescript', 'yaml', 'json', 'linux', 'macos', 'windows', 'git', 'github-actions', 'ci-cd', 'deployment', 'observability', 'prometheus', 'grafana', 'logging', 'authentication', 'reverse-proxy', 'ssl-certificates', 'dns', 'wireguard', 'tailscale', 'home-assistant', 'proxmox', 'systemd', 'cron', 'storage', 's3', 'rest-api', 'performance', 'debugging', 'testing', 'documentation', 'migration', 'disaster-recovery', 'production-infrastructure-runbooks', 'development', 'staging', 'personal', 'work', 'automation', 'shell', 'webhooks', 'containers', 'zsh']
names += [f'topic-{i}' for i in range(61, 101)]
names = names[:args.tags]
if not args.measure_only:
    existing, _, _ = request('/snippets?limit=1')
    if existing['pagination']['total']:
        parser.error('Refusing to seed a non-empty database. Use a new temporary instance or --measure-only.')
    topics = [('Docker service health check', 'yaml', 'compose.yaml'), ('PostgreSQL backup and restore', 'bash', 'backup.sh'), ('HTTP retry with exponential backoff', 'python', 'retry.py'), ('Nginx reverse proxy configuration', 'nginx', 'nginx.conf'), ('Deployment checklist', 'markdown', 'README.md'), ('Inspect slow queries', 'sql', 'queries.sql'), ('Application metrics endpoint', 'go', 'metrics.go'), ('Rotate service credentials', 'bash', 'rotate.sh')]
    for i in range(args.snippets):
        title, language, filename = topics[i % len(topics)]
        line_count = 1200 if i % 16 == 0 else (100 if i % 5 == 0 else 20)
        content = '\n'.join(f'# {title}: operational step {j + 1:04d} for service {i + 1:03d}' for j in range(line_count))
        files = [{'filename': filename, 'language': language, 'content': content}]
        if i % 7 == 0:
            files += [{'filename': 'README.md', 'language': 'markdown', 'content': f'# {title}\n\nTest locally before applying.\n\nUnique lookup: canary-{i:03d}'}]
        data = {'title': f'{title} — service {i + 1:03d}', 'description': '' if i % 3 == 0 else f'Reusable {language} recipe for development and self-hosted services. Revision {i + 1}.', 'content': content, 'language': language, 'files': files, 'tags': list(dict.fromkeys([names[i % len(names)], names[(i * 7 + 3) % len(names)], names[(i + 5) % len(names)]]))}
        request('/snippets', data)
    print(f'Seeded {args.snippets} snippets / {args.tags} tags', flush=True)

listing, _, _ = request('/snippets?limit=20')
tags, _, _ = request('/tags')
tag_id = tags['data'][-1]['id']
first_id = listing['data'][0]['id']
paths = {'first_page': '/snippets?limit=20', 'last_page': f'/snippets?limit=20&page={math.ceil(listing["pagination"]["total"] / 20)}', 'title_search': '/snippets?q=PostgreSQL&limit=20', 'content_search': '/snippets?q=canary-000&limit=20', 'tag_filter': f'/snippets?tag_ids={tag_id}&limit=20', 'tags': '/tags', 'detail': f'/snippets/{first_id}'}
report = {'snippets': listing['pagination']['total'], 'tags': len(tags['data']), 'method': '30 warm serial localhost requests per operation; excludes browser rendering and WAN latency', 'operations': {}}
for name, path in paths.items():
    if args.summary and path.startswith('/snippets?'):
        path += '&view=summary'
    request(path)  # warm-up
    samples = [request(path) for _ in range(30)]
    times = sorted(sample[1] for sample in samples)
    report['operations'][name] = {'median_ms': round(statistics.median(times), 2), 'p95_ms': round(times[math.ceil(.95 * len(times)) - 1], 2), 'response_bytes': samples[-1][2], 'returned_items': len(samples[-1][0]['data']) if isinstance(samples[-1][0]['data'], list) else 1}
print(json.dumps(report, indent=2))
if args.output:
    args.output.write_text(json.dumps(report, indent=2) + '\n')
