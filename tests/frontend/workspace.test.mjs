import test, { beforeEach } from 'node:test';
import assert from 'node:assert/strict';

const storage = new Map();
globalThis.localStorage = { getItem: key => storage.get(key) ?? null, setItem: (key, value) => storage.set(key, value), removeItem: key => storage.delete(key) };
globalThis.window = { location: new URL('http://localhost/snipo/'), SNIPO_CONFIG: { basePath: '/snipo' } };
globalThis.document = { getElementById: () => null, querySelector: () => null, querySelectorAll: () => [] };
globalThis.history = { pushState: (_, __, url) => { window.location = new URL(url); } };
const { initSnippetsApp } = await import('../../internal/web/static/js/components/snippets-app.js');
const { api } = await import('../../internal/web/static/js/modules/api.js');
let factory;
initSnippetsApp({ data: (_, fn) => { factory = fn; } });
function app() {
  const instance = factory();
  instance.$nextTick = () => {};
  instance.destroyAceEditor = () => {};
  return instance;
}
beforeEach(() => { storage.clear(); window.location = new URL('http://localhost/snipo/'); });

test('workspace composes without evaluating computed state before initialization', () => {
  const state = app();
  assert.equal(state.libraryTitle(), 'All snippets');
  assert.equal(state.viewMode, 'grid');
  assert.equal(state.hasActiveFilters(), false);
  state.filter.query = 'docker';
  assert.deepEqual(state.activeFilterChips(), [{ key: 'query', label: 'docker' }]);
});

test('newer search results win even when an earlier request finishes last', async () => {
  const state = app();
  const pending = [];
  api.get = () => new Promise(resolve => pending.push(resolve));
  const first = state.loadSnippets();
  state.filter.query = 'new';
  const second = state.loadSnippets();
  pending[1]({ data: [{ id: 2 }], pagination: { total: 1 } });
  await second;
  pending[0]({ data: [{ id: 1 }], pagination: { total: 1 } });
  await first;
  assert.equal(state.snippets[0].id, 2);
  assert.equal(state.loading, false);
});

test('library request failure produces a retry state rather than an empty library', async () => {
  const state = app();
  api.get = async () => { throw new Error('Network unavailable'); };
  await state.loadSnippets();
  assert.equal(state.loadError, 'Network unavailable');
  assert.equal(state.loading, false);
});

test('opening and closing a snippet preserves URL filters', () => {
  const state = app();
  state.filter.query = 'docker compose';
  state.filter.tagId = 7;
  state.updateUrl({ snippet: 42 });
  assert.equal(window.location.searchParams.get('q'), 'docker compose');
  assert.equal(window.location.searchParams.get('tag'), '7');
  assert.equal(window.location.searchParams.get('snippet'), '42');
  state.cancelEdit();
  assert.equal(window.location.searchParams.get('snippet'), null);
  assert.equal(window.location.searchParams.get('tag'), '7');
});

test('closing an edit retains a recoverable local draft', () => {
  const state = app();
  state.isEditing = true;
  state.editingSnippet.title = 'Recovery';
  state.editingSnippet.files[0].content = 'echo preserved';
  state.cancelEdit();
  assert.equal(state.hasDraft, true);
  assert.equal(JSON.parse(storage.get('snipo-draft')).snippet.title, 'Recovery');
  state.restoreDraft();
  assert.equal(state.isEditing, true);
  assert.equal(state.editingSnippet.files[0].content, 'echo preserved');
  assert.ok(storage.has('snipo-draft'));
});

test('save failure keeps editor and draft, and releases the saving lock', async () => {
  const state = app();
  state.isEditing = state.showEditor = true;
  state.editingSnippet.title = 'Recover me';
  api.post = async () => { throw new Error('Offline'); };
  await state.saveSnippet();
  assert.equal(state.saveState, 'error');
  assert.equal(state.isEditing, true);
  assert.equal(state.showEditor, true);
  assert.equal(state.saving, false);
  assert.ok(storage.has('snipo-draft'));
});

test('successful save remains in preview and clears the local draft', async () => {
  const state = app();
  state.isEditing = state.showEditor = true;
  state.editingSnippet.title = 'Saved';
  state.saveDraft();
  api.post = async () => ({ id: 12, title: 'Saved', files: [], tags: [] });
  state.loadSnippets = state.loadTags = state.loadFavoritesCount = async () => {};
  await state.saveSnippet();
  assert.equal(state.showEditor, true);
  assert.equal(state.isEditing, false);
  assert.equal(state.saveState, 'saved');
  assert.equal(state.editingSnippet.id, 12);
  assert.equal(storage.has('snipo-draft'), false);
  assert.equal(window.location.searchParams.get('snippet'), '12');
});

test('back navigation to library resets stale selection without pushing history', async () => {
  const state = app();
  state.showEditor = true;
  state.filter.isDeleted = true;
  state.loadSnippets = async () => {};
  await state.restoreFromUrl();
  assert.equal(state.showEditor, false);
  assert.equal(state.filter.isDeleted, null);
  assert.equal(window.location.href, 'http://localhost/snipo/');
});

test('settings sections are validated and included in the URL', async () => {
  const state = app();
  state.loadApiTokens = async () => {};
  await state.openSettings('appearance');
  assert.equal(window.location.searchParams.get('settings'), 'appearance');
  await state.selectSettingsTab('unknown');
  assert.equal(state.settingsTab, 'general');
  state.closeSettings();
  assert.equal(window.location.searchParams.has('settings'), false);
});

test('deep links accept actual string snippet identifiers', async () => {
  const state = app();
  window.location = new URL('http://localhost/snipo/?snippet=6ef5c04f11d15e59&q=docker');
  state.loadSnippets = async () => {};
  api.get = async () => ({ id: '6ef5c04f11d15e59', title: 'Compose', files: [], tags: [] });
  await state.restoreFromUrl();
  assert.equal(state.editingSnippet.id, '6ef5c04f11d15e59');
  assert.equal(state.showEditor, true);
  assert.equal(state.filter.query, 'docker');
});

test('copy uses the active file, including intentionally empty content', async () => {
  const state = app();
  state.editingSnippet.files = [{ content: 'first', language: 'bash' }, { content: '', language: 'python' }];
  state.activeFileIndex = 1;
  let copied;
  Object.defineProperty(globalThis, 'navigator', { configurable: true, value: { clipboard: { writeText: async text => { copied = text; } } } });
  await state.copyToClipboard(state.editingSnippet);
  assert.equal(copied, '');
  state.editingSnippet.files[1].content = 'second';
  await state.copyToClipboard(state.editingSnippet);
  assert.equal(copied, 'second');
});

test('share links respect deployments under a path prefix', async () => {
  const state = app();
  let copied;
  Object.defineProperty(globalThis, 'navigator', { configurable: true, value: { clipboard: { writeText: async text => { copied = text; } } } });
  await state.copyShareLink({ id: 'abc123', is_public: true });
  assert.equal(copied, 'http://localhost/snipo/s/abc123');
});

test('large tag lists stay bounded, searchable, and retain the selected tag', () => {
  const state = app();
  state.tags = Array.from({ length: 60 }, (_, i) => ({ id: i + 1, name: `tag-${i + 1}` }));
  assert.equal(state.visibleTags().length, 10);
  state.filter.tagId = 60;
  assert.equal(state.visibleTags().at(-1).id, 60);
  state.tagSearch = 'TAG-60';
  assert.deepEqual(state.visibleTags().map(t => t.id), [60]);
  state.tagSearch = 'missing';
  assert.equal(state.visibleTags().length, 0);
  state.tagSearch = '';
  state.showAllTags = true;
  assert.equal(state.visibleTags().length, 60);
});

test('pagination rejects invalid pages and scrolls the library after loading', async () => {
  const state = app();
  state.pagination.totalPages = 8;
  state.loading = false;
  let calls = 0, scrolled = false;
  state.loadSnippets = async () => { calls++; };
  state.$root = { querySelector: () => ({ scrollTo: () => { scrolled = true; } }) };
  await state.goToPage(9);
  assert.equal(calls, 0);
  await state.goToPage(8);
  assert.equal(state.pagination.page, 8);
  assert.equal(calls, 1);
  assert.equal(scrolled, true);
});

test('summary card copy fetches full content, including a selected secondary file', async () => {
  const state = app();
  let copied;
  Object.defineProperty(globalThis, 'navigator', { configurable: true, value: { clipboard: { writeText: async text => { copied = text; } } } });
  api.get = async path => {
    assert.equal(path, '/api/v1/snippets/example');
    return { content: 'full primary', files: [{ id: 'secondary', filename: 'README.md', content: 'full secondary' }] };
  };
  await state.copySnippetFromCard({ id: 'example', is_summary: true, preview: 'partial' });
  assert.equal(copied, 'full primary');
  await state.copyFileFromCard({ id: 'example', is_summary: true }, { id: 'secondary', content: '' });
  assert.equal(copied, 'full secondary');
});

test('layout choices persist independently of filters and restore on reload', () => {
  const state = app();
  state.filter.query = 'docker';
  state.pagination.page = 3;
  for (const [layout, mode, density] of [['grid', 'grid', 'comfortable'], ['compact', 'list', 'compact'], ['list', 'list', 'comfortable']]) {
    state.setLibraryLayout(layout);
    assert.equal(state.viewMode, mode);
    assert.equal(state.density, density);
    assert.equal(storage.get('snipo-view-mode'), mode);
    assert.equal(storage.get('snipo-density'), density);
    assert.equal(app().viewMode, mode);
    assert.equal(app().density, density);
    assert.equal(state.filter.query, 'docker');
    assert.equal(state.pagination.page, 3);
  }
  state.setLibraryLayout('invalid');
  assert.equal(state.viewMode, 'list');
});

test('save synchronizes the active Ace buffer before sending files and rejects duplicate saves', async () => {
  const state = app();
  state.isEditing = state.showEditor = true;
  state.editingSnippet.title = 'Buffer sync';
  state.aceEditor = { getValue: () => 'latest code' };
  state.syncCurrentContent = () => { state.editingSnippet.files[0].content = state.aceEditor.getValue(); };
  let finish, requests = 0;
  api.post = async (_path, body) => {
    requests++;
    assert.equal(body.files[0].content, 'latest code');
    return new Promise(resolve => { finish = resolve; });
  };
  state.loadSnippets = state.loadTags = state.loadFavoritesCount = state.loadTotalCount = async () => {};
  const first = state.saveSnippet();
  await state.saveSnippet();
  assert.equal(requests, 1);
  finish({ id: 'buffer', files: state.editingSnippet.files, tags: [] });
  await first;
  assert.equal(state.saving, false);
  assert.equal(state.showEditor, true);
});


test('missing or invalid layout preferences default to Cards', () => {
  for (const saved of [undefined, '', 'unknown']) {
    if (saved !== undefined) storage.set('snipo-view-mode', saved);
    assert.equal(app().viewMode, 'grid');
    assert.equal(app().density, 'comfortable');
  }
});
