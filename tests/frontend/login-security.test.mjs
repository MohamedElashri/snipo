import test, { beforeEach } from 'node:test';
import assert from 'node:assert/strict';
import { initLoginForm } from '../../internal/web/static/js/components/login-form.js';
import { api } from '../../internal/web/static/js/modules/api.js';

let factory;
initLoginForm({ data: (_name, fn) => { factory = fn; } });
beforeEach(() => {
  globalThis.window = { location: { origin: 'https://snipo.test', pathname: '/login', search: '', href: 'https://snipo.test/login' }, SNIPO_CONFIG: { basePath: '' } };
});

test('failed login releases the submit button for retry', async () => {
  const form = factory();
  api.post = async () => ({ error: { message: 'Invalid password' } });
  await form.login();
  assert.equal(form.error, 'Invalid password');
  assert.equal(form.loading, false);
});

test('successful login ignores protocol-relative destinations', async () => {
  window.location.search = '?next=%2F%2Freceiver.test%2Fmarker';
  const form = factory();
  api.post = async () => ({ success: true });
  await form.login();
  assert.equal(new URL(window.location.href, 'https://snipo.test').origin, 'https://snipo.test');
});

test('return destinations remain on the same origin under browser URL normalization', async () => {
  const { loginDestination } = await import('../../internal/web/static/js/components/login-form.js');
  for (const path of ['//receiver.test', '///receiver.test', '/\\receiver.test', '/\t/receiver.test', '/..//receiver.test', 'https://receiver.test', 'javascript:alert(1)']) {
    assert.equal(new URL(loginDestination(path)).origin, 'https://snipo.test', path);
  }
  assert.equal(loginDestination('?q=hello%20world', '/snipo'), 'https://snipo.test/snipo/?q=hello%20world');
  assert.equal(loginDestination('/?snippet=example', '/snipo'), 'https://snipo.test/snipo/?snippet=example');
  assert.equal(loginDestination('/../elsewhere', '/snipo'), 'https://snipo.test/snipo/');
});

test('a failed attempt can be retried and query return destinations survive', async () => {
  const form = factory();
  api.post = async () => ({ error: { message: 'Try again' } });
  await form.login();
  api.post = async () => ({ success: true });
  window.location.search = '?next=%3Fq%3Dreturn%2520query';
  await form.login();
  assert.equal(window.location.href, 'https://snipo.test/?q=return%20query');
  assert.equal(form.loading, false);
});

test('duplicate submit does not start another login request', async () => {
  const form = factory();
  let requests = 0;
  let complete;
  api.post = () => { requests++; return new Promise(resolve => { complete = resolve; }); };
  const first = form.login();
  await form.login();
  assert.equal(requests, 1);
  complete({ error: { message: 'Try again' } });
  await first;
  assert.equal(form.loading, false);
});
