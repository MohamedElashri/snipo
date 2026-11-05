const { test, afterEach } = require('node:test');
const assert = require('node:assert/strict');
const { SnipoClient } = require('../dist/index.js');
const originalFetch = global.fetch;
afterEach(() => { global.fetch = originalFetch; });
const snippet = { id: 'abc-123', title: 'Saved', content: 'hello' };

test('detail, create, and update unwrap server envelopes without losing files', async () => {
  const saved = { ...snippet, files: [{ filename: 'main.js', content: 'code' }] };
  global.fetch = async () => Response.json({ data: saved, meta: { version: '1.0' } });
  const client = new SnipoClient({ baseURL: 'https://example.test/snipo/' });
  assert.deepEqual(await client.getSnippet(saved.id), saved);
  assert.deepEqual(await client.createSnippet(saved), saved);
  assert.deepEqual(await client.updateSnippet(saved.id, saved), saved);
  global.fetch = async () => Response.json(saved);
  assert.deepEqual(await client.getSnippet(saved.id), saved);
});

test('preserves base paths, filters, raw wrapper envelopes, and clears credentials', async () => {
  const client = new SnipoClient({ baseURL: ' https://example.test/snipo/// ', token: 'old' });
  client.updateToken(undefined);
  global.fetch = async (url, options) => {
    assert.equal(url, 'https://example.test/snipo/api/v1/snippets?q=a%26b');
    assert.equal(options.headers.Authorization, undefined);
    return Response.json({ data: [snippet], pagination: { totalPages: 2 } });
  };
  const response = await client.client.get('/api/v1/snippets', { params: { q: 'a&b' } });
  assert.equal(response.data.pagination.totalPages, 2);
});

function pendingFetch(_url, { signal }) {
  return new Promise((_, reject) => {
    if (signal.aborted) reject(signal.reason);
    else signal.addEventListener('abort', () => reject(signal.reason), { once: true });
  });
}

test('honors request timeouts and caller cancellation', async () => {
  global.fetch = pendingFetch;
  const client = new SnipoClient({ baseURL: 'https://example.test' });
  await assert.rejects(client.client.get('/slow', { timeout: 10 }), { name: 'TimeoutError' });
  const controller = new AbortController();
  const request = client.getSnippets('search', controller.signal);
  controller.abort();
  await assert.rejects(request, { name: 'AbortError' });
  await assert.rejects(client.getSnippets('search', controller.signal), { name: 'AbortError' });
});

test('retains structured API errors for client feedback', async () => {
  const error = { error: { message: 'Token expired' } };
  global.fetch = async () => Response.json(error, { status: 401 });
  await assert.rejects(new SnipoClient({ baseURL: 'https://example.test' }).getSnippet('1'), e => {
    assert.equal(e.response.status, 401);
    assert.deepEqual(e.response.data, error);
    return true;
  });
});
