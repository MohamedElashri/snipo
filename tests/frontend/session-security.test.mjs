import test, { beforeEach, afterEach } from 'node:test';
import assert from 'node:assert/strict';
import { initSnippetsApp } from '../../internal/web/static/js/components/snippets-app.js';
import { api } from '../../internal/web/static/js/modules/api.js';

const local = new Map();
const tab = new Map();
const events = new Map();
let navigated, reloads, state;
function storage(map) { return { getItem: key => map.get(key) ?? null, setItem: (key, value) => map.set(key, value), removeItem: key => map.delete(key) }; }
let factory;
initSnippetsApp({ data: (_name, fn) => { factory = fn; } });
beforeEach(() => {
  local.clear(); tab.clear(); events.clear(); navigated = null; reloads = 0;
  globalThis.localStorage = storage(local);
  globalThis.sessionStorage = storage(tab);
  globalThis.document = { getElementById: () => null };
  globalThis.window = {
    SNIPO_CONFIG: { basePath: '/snipo' },
    location: { replace: url => { navigated = url; }, reload: () => { reloads++; } },
    addEventListener: (name, fn) => events.set(name, fn), removeEventListener: name => events.delete(name)
  };
  globalThis.BroadcastChannel = class { postMessage() {} close() {} };
  state = factory(); state.destroyCodeEditor = () => {};
  state.$root = { hidden: false, style: { setProperty() {} } };
  state.initSessionGuard();
  state.isEditing = true;
  state.editingSnippet.title = 'Private draft';
  state.editingSnippet.files[0].content = 'PRIVATE-SESSION-MARKER';
  state.saveDraft();
  local.set('snipo-theme', 'dark');
  tab.set('snipo-settings', '{"custom_css":"private"}');
});
afterEach(() => { state.destroySessionGuard(); });

test('confirmed logout clears private recovery data and prevents unload from recreating it', async () => {
  api.post = async () => ({ success: true });
  await state.logout();
  assert.equal(navigated, '/snipo/login');
  assert.equal(state.$root.hidden, true);
  assert.equal(local.has('snipo-draft'), false);
  assert.equal(tab.has('snipo-draft'), false);
  assert.equal(tab.has('snipo-settings'), false);
  assert.equal(local.get('snipo-theme'), 'dark');
  state.isEditing = true;
  state.editingSnippet.title = 'A stale callback';
  state.saveDraft();
  assert.equal(local.has('snipo-draft'), false);
});

test('failed logout keeps the session usable and does not pretend to sign out', async () => {
  for (const failure of [async () => ({ error: { message: 'Failed' } }), async () => { throw new Error('Offline'); }]) {
    api.post = failure;
    await state.logout();
    assert.equal(navigated, null);
    assert.equal(state._sessionEnded, false);
    assert.equal(state.loggingOut, false);
    assert.ok(local.get('snipo-draft').includes('PRIVATE-SESSION-MARKER'));
  }
});

test('other-tab logout events clear this tab too', () => {
  events.get('storage')({ key: 'snipo-session-ended', newValue: 'another-tab' });
  assert.equal(navigated, '/snipo/login');
  assert.equal(tab.has('snipo-draft'), false);
});

test('history snapshots hide private views and refresh authorization on restore', () => {
  events.get('pagehide')({ persisted: true });
  assert.equal(state.$root.hidden, true);
  assert.ok(local.has('snipo-draft'), 'ordinary navigation must retain draft recovery');
  local.set('snipo-session-ended', 'logged-out-while-frozen');
  events.get('pageshow')({ persisted: true });
  assert.equal(reloads, 1);
  assert.equal(tab.has('snipo-draft'), false);
  assert.equal(local.has('snipo-draft'), false);
});

test('a blocked notification channel cannot skip local logout cleanup', async () => {
  state._sessionChannel.postMessage = () => { throw new Error('Channel blocked'); };
  api.post = async () => ({ success: true });
  await state.logout();
  assert.equal(navigated, '/snipo/login');
  assert.equal(local.has('snipo-draft'), false);
});
