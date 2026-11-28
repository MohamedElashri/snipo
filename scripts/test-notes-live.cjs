#!/usr/bin/env node
const assert = require('node:assert/strict');
const { execFileSync, spawn } = require('node:child_process');
const { existsSync, mkdtempSync, mkdirSync, readFileSync, openSync, closeSync, rmSync } = require('node:fs');
const { createServer } = require('node:net');
const os = require('node:os');
const path = require('node:path');
const { setTimeout: pause } = require('node:timers/promises');
const { chromium, firefox, webkit } = require('playwright');

const root = path.resolve(__dirname, '..');
const temp = mkdtempSync(path.join(os.tmpdir(), 'snipo-notes-live-'));
const binary = path.join(temp, 'snipo');
const password = 'snipo-notes-live-test-password';
const basePath = process.env.SNIPO_TEST_BASE_PATH ?? '/library';
const browserType = { chromium, firefox, webkit }[process.env.SNIPO_TEST_BROWSER || 'chromium'];
const caseFilter = process.env.SNIPO_TEST_CASE;
const servers = [];

async function startServer(name, executable = binary, database) {
  const socket = createServer();
  await new Promise((resolve, reject) => socket.listen(0, '127.0.0.1', resolve).once('error', reject));
  const port = socket.address().port;
  await new Promise((resolve, reject) => socket.close(error => error ? reject(error) : resolve()));
  const directory = path.join(temp, name);
  mkdirSync(directory, { recursive: true });
  const db = database || path.join(directory, 'notes.db');
  const log = openSync(path.join(directory, 'server.log'), 'a');
  const processHandle = spawn(executable, ['serve'], {
    cwd: directory,
    env: {
      ...process.env,
      SNIPO_HOST: '127.0.0.1', SNIPO_PORT: String(port), SNIPO_BASE_PATH: basePath,
      SNIPO_DB_PATH: db, SNIPO_MASTER_PASSWORD: password,
      SNIPO_DISABLE_AUTH: 'false', SNIPO_DEMO_MODE: 'false',
      SNIPO_ENABLE_PUBLIC_SNIPPETS: 'true', SNIPO_ENABLE_BACKUP_RESTORE: 'true',
      SNIPO_RATE_LIMIT_READ: '1000', SNIPO_RATE_LIMIT_WRITE: '1000', SNIPO_RATE_LIMIT_ADMIN: '1000',
      SNIPO_S3_ENABLED: 'false', SNIPO_SESSION_SECRET: 'notes-live-session-secret',
      SNIPO_ENCRYPTION_SALT: 'notes-live-encryption-salt',
    },
    stdio: ['ignore', log, log],
  });
  closeSync(log);
  const url = `http://127.0.0.1:${port}${basePath}`;
  const server = { processHandle, url, db };
  servers.push(server);
  for (let attempt = 0; attempt < 100; attempt++) {
    if (await fetch(url + '/ping', { signal: AbortSignal.timeout(500) }).then(response => response.ok).catch(() => false)) {
      return server;
    }
    if (processHandle.exitCode !== null) break;
    await pause(100);
  }
  throw new Error(`Server did not start. See ${directory}/server.log`);
}

async function stopServer({ processHandle }) {
  if (processHandle.exitCode !== null || processHandle.signalCode !== null) return;
  await new Promise(resolve => {
    processHandle.once('exit', resolve);
    processHandle.kill();
  });
}

async function signIn(context, server) {
  const page = await context.newPage();
  await page.goto(server.url + '/login');
  await page.locator('#password').fill(password);
  await page.getByRole('button', { name: 'Sign In', exact: true }).click();
  await page.waitForURL(server.url + '/');
  await page.locator('.app-container .library-heading h1').waitFor();
  return page;
}

async function request(page, server, route, method = 'GET', data, status = 200) {
  const response = await page.request.fetch(server.url + '/api/v1' + route, { method, data });
  assert.equal(response.status(), status, `${method} ${route}: ${await response.text()}`);
  if (status === 204) return;
  return (await response.json()).data;
}

function createNote(page, server, fields) {
  return request(page, server, '/snippets', 'POST', {
    kind: 'note', language: 'markdown', content: 'Original body', ...fields,
  }, 201);
}

async function openNote(page, server, id, edit = false) {
  await page.goto(server.url + '/?kind=note&snippet=' + id + (edit ? '&edit=true' : ''));
  if (edit) await page.getByRole('textbox', { name: 'Note body' }).waitFor();
  else await page.getByRole('button', { name: 'Edit', exact: true }).waitFor();
}

function save(page) {
  return page.getByRole('button', { name: 'Save', exact: true }).click();
}

const cases = [
  ['creates and reads notes, downloads a plain-text note, then shares and revokes access',
    async ({ page, browser, server }) => {
      await page.getByRole('button', { name: 'Notes', exact: true }).click();
      await page.getByRole('heading', { name: 'Notes', exact: true }).waitFor();
      await page.getByRole('button', { name: 'New note', exact: true }).first().click();
      const body = page.getByRole('textbox', { name: 'Note body' });
      await body.waitFor();
      assert.equal(await body.evaluate(node => node === document.activeElement), true);
      await body.fill('# Release checklist\n\n\x60\x60\x60sh\necho ready\n\x60\x60\x60');
      await page.getByRole('button', { name: 'Preview', exact: true }).click();
      await page.getByRole('button', { name: 'Copy code block' }).waitFor();
      await save(page);
      const edit = page.getByRole('button', { name: 'Edit', exact: true });
      await edit.waitFor();
      await page.locator('.note-reader').getByRole('heading', { name: 'Release checklist' }).waitFor();
      const id = new URL(page.url()).searchParams.get('snippet');
      assert.ok(id, 'A saved note should have a URL');
      await edit.click();
      await page.getByRole('combobox', { name: 'Format' }).selectOption('plaintext');
      const text = '  First line\n\nSecond <literal> line\n';
      await body.fill(text);
      await page.getByRole('button', { name: 'Details', exact: true }).click();
      await page.getByRole('checkbox', { name: 'Make public', exact: true }).check();
      await save(page);
      const reader = page.locator('.note-reader .note-plain');
      await reader.waitFor();
      assert.equal(await reader.textContent(), text);
      await page.getByRole('region', { name: 'Note workspace' }).getByText('More', { exact: true }).click();
      const downloadPromise = page.waitForEvent('download');
      await page.getByRole('button', { name: 'Download note', exact: true }).click();
      const download = await downloadPromise;
      assert.equal(download.suggestedFilename(), 'Release checklist.txt');
      assert.equal(readFileSync(await download.path(), 'utf8'), text);

      const visitor = await browser.newContext();
      try {
        const share = await visitor.newPage();
        const markdown = await createNote(page, server, {
          title: 'Shareable Markdown', content: '## Reading view\n\n```sh\necho ready\n```', is_public: true,
        });
        await share.goto(server.url + '/s/' + markdown.id);
        const document = share.locator('.note-document.markdown-content');
        await document.getByRole('heading', { name: 'Reading view' }).waitFor();
        assert.equal(await document.evaluate(el => getComputedStyle(el).position), 'static');
        await share.getByRole('button', { name: 'Copy code block' }).click();
        await share.getByTitle('Copy note').click();

        await share.goto(server.url + '/s/' + id);
        await share.locator('.note-plain').waitFor();
        assert.equal(await share.locator('.note-plain').textContent(), text);
        assert.equal(await share.locator('literal').count(), 0);
        await page.goto(server.url + '/s/' + id);
        await page.getByTitle('Make Private').click();
        await page.waitForURL(server.url + '/');
        await share.reload();
        await share.getByRole('heading', { name: 'Shared item unavailable' }).waitFor();
      } finally {
        await visitor.close();
      }
    }],
  ['fills the note workspace and keeps editing controls reachable in both Markdown modes',
    async ({ page, server }) => {
      const code = 'echo ' + 'x'.repeat(250);
      const content = ['## Start of document', '```bash\n' + code + '\n```', ...Array.from({ length: 40 }, (_, i) => 'Paragraph ' + i), '## End of document'].join('\n\n');
      const note = await createNote(page, server, { title: 'Long note '.repeat(16), content });
      const scroll = page.locator('.note-scroll');
      const writer = page.getByRole('textbox', { name: 'Note body' });
      const previewButton = page.getByRole('button', { name: 'Preview', exact: true });
      const writeButton = page.getByRole('button', { name: 'Write', exact: true });

      for (const [width, height] of [[1440, 900], [1030, 720], [390, 844], [320, 680]]) {
        await page.setViewportSize({ width, height });
        await openNote(page, server, note.id, true);
        assert.equal(await page.locator('.workspace > .snippet-list').isVisible(), false);
        const workspaceWidth = (await page.locator('.workspace').boundingBox()).width;
        const editor = await writer.boundingBox();
        assert.ok(editor.width >= workspaceWidth - 2, 'The writer should fill the workspace at ' + width);
        assert.ok(editor.height >= (await scroll.boundingBox()).height - 2, 'The writer should fill the available height');

        await writer.fill(content + '\n\nLast edit');
        await writer.evaluate(input => { input.selectionStart = 12; input.selectionEnd = 21; });
        const writingStart = await writer.evaluate(el => el.getBoundingClientRect().left + parseFloat(getComputedStyle(el).paddingLeft));
        await previewButton.click();
        await page.locator('.note-reader').getByRole('heading', { name: 'Start of document' }).waitFor();
        const document = page.locator('.note-reader .note-document.markdown-content');
        const preview = await document.evaluate(el => {
          const style = getComputedStyle(el);
          const rect = el.getBoundingClientRect();
          return { position: style.position, width: rect.width, start: rect.left + parseFloat(style.paddingLeft) };
        });
        assert.equal(preview.position, 'static');
        assert.ok(preview.width >= await scroll.evaluate(el => el.clientWidth) - 2, 'Preview should fill the available width');
        assert.ok(Math.abs(preview.start - writingStart) < 2, 'Preview should line up with the writing area');
        assert.equal(await page.getByRole('button', { name: 'Copy code block' }).isVisible(), true);
        await scroll.evaluate(el => { el.scrollTop = el.scrollHeight; });
        assert.ok(await scroll.evaluate(el => el.scrollTop > 0), 'Long Markdown should scroll beneath the toolbar');
        assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true);
        await writeButton.click();
        assert.equal(await writer.inputValue(), content + '\n\nLast edit');
        assert.deepEqual(await writer.evaluate(el => [el.selectionStart, el.selectionEnd]), [12, 21]);
        await previewButton.click();
        await save(page);
        const edit = page.getByRole('button', { name: 'Edit', exact: true });
        await edit.waitFor();
        await page.locator('.note-reader').getByRole('heading', { name: 'End of document' }).waitFor();
        await scroll.evaluate(el => { el.scrollTop = el.scrollHeight; });
        const details = page.getByRole('button', { name: 'Details', exact: true });
        await details.click();
        assert.equal(await details.getAttribute('aria-expanded'), 'true');
        await page.waitForFunction(() => document.querySelector('.note-scroll').scrollTop === 0);
        await page.keyboard.press('Escape');
        assert.equal(await details.getAttribute('aria-expanded'), 'false');
        await edit.click();
        await writer.waitFor();
        if (width === 320) {
          await writer.fill('');
          await previewButton.click();
          await page.getByText('Nothing to preview yet.').waitFor();
          await writeButton.click();
          await writer.waitFor();
        }
      }
    }],
  ['resizes the navigation sidebar and reflows notes, snippets, and the library',
    async ({ page, server }) => {
      await page.setViewportSize({ width: 1440, height: 900 });
      await page.emulateMedia({ reducedMotion: 'reduce' });
      const note = await createNote(page, server, { title: 'Resizable note', content: '## Resizing the workspace' });
      const snippet = await request(page, server, '/snippets', 'POST', {
        title: 'Resizable snippet', language: 'bash', content: 'echo resized',
      }, 201);
      await openNote(page, server, note.id, true);

      const divider = page.getByRole('separator', { name: 'Resize navigation sidebar' });
      const writer = page.getByRole('textbox', { name: 'Note body' });
      const width = locator => locator.evaluate(el => el.getBoundingClientRect().width);
      const checkSidebarWidth = async expected => {
        await page.waitForFunction(width => {
          const sidebar = document.querySelector('#navigation-sidebar').getBoundingClientRect();
          const main = document.querySelector('.main-content').getBoundingClientRect();
          return sidebar.width === width && main.left === sidebar.right && main.right === innerWidth;
        }, expected);
        assert.equal(await divider.getAttribute('aria-valuenow'), String(expected));
      };
      const drag = async (distance, expected) => {
        const { x, y, width, height } = await divider.boundingBox();
        const startX = x + width / 2;
        await page.mouse.move(startX, y + height / 2);
        await page.mouse.down();
        await page.mouse.move(startX + distance, y + height * 0.7, { steps: 8 });
        await checkSidebarWidth(expected);
        await page.mouse.up();
      };

      await checkSidebarWidth(220);
      const before = await width(writer);
      await drag(140, 360);
      assert.equal(await width(writer), before - 140);
      await page.getByRole('button', { name: 'Preview', exact: true }).click();
      await page.getByRole('heading', { name: 'Resizing the workspace' }).waitFor();
      const preview = page.locator('.note-reader .markdown-content');
      const previewBefore = await width(preview);
      await drag(90, 450);
      assert.equal(await width(preview), previewBefore - 90);
      await page.getByRole('button', { name: 'Write', exact: true }).click();
      assert.equal(await writer.inputValue(), '## Resizing the workspace');

      await page.reload();
      await writer.waitFor();
      await checkSidebarWidth(450);
      const toggle = page.getByTitle('Toggle Sidebar', { exact: true });
      await toggle.click();
      await divider.waitFor({ state: 'hidden' });
      assert.equal(await width(writer), 1440);
      await toggle.click();
      await checkSidebarWidth(450);

      await page.goto(server.url + '/?kind=snippet&snippet=' + snippet.id);
      await page.getByRole('button', { name: 'Edit', exact: true }).waitFor();
      const library = page.locator('.workspace > .snippet-list');
      await library.waitFor();
      await page.setViewportSize({ width: 1030, height: 900 });
      await checkSidebarWidth(412);
      await library.waitFor({ state: 'hidden' });
      assert.equal(await width(page.locator('.editor-container')), 618);
      assert.equal(await divider.getAttribute('aria-valuemax'), '412');
      await drag(200, 412);
      assert.equal(await page.evaluate(() => localStorage.getItem('sidebarWidth')), '450');
      await drag(-600, 180);
      await library.waitFor();
      await divider.press('ArrowRight');
      await checkSidebarWidth(196);

      await page.setViewportSize({ width: 390, height: 844 });
      await divider.waitFor({ state: 'hidden' });
      const mobileToggle = page.getByRole('button', { name: 'Toggle navigation' });
      assert.equal(await mobileToggle.getAttribute('aria-expanded'), 'false');
      await mobileToggle.click();
      assert.equal(await mobileToggle.getAttribute('aria-expanded'), 'true');
      assert.equal(await divider.isVisible(), false);
      assert.equal(await width(page.locator('.main-content')), 390);
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true);

      await page.setViewportSize({ width: 1440, height: 900 });
      await checkSidebarWidth(196);
    }],
  ['keeps multiple local drafts after switching notes and saving another item',
    async ({ page, server }) => {
      const first = await createNote(page, server, { title: 'First draft' });
      const second = await createNote(page, server, { title: 'Second draft' });
      const unrelated = await createNote(page, server, { title: 'Unrelated note' });
      for (const [note, text] of [[first, 'Unsaved first change'], [second, 'Unsaved second change']]) {
        await openNote(page, server, note.id, true);
        await page.getByRole('textbox', { name: 'Note body' }).fill(text);
        await page.getByRole('button', { name: 'Close', exact: true }).click();
        await page.getByRole('button', { name: 'Edit', exact: true }).waitFor();
      }
      await openNote(page, server, unrelated.id, true);
      await page.getByRole('textbox', { name: 'Note body' }).fill('Saved unrelated edit');
      await save(page);
      await page.getByRole('button', { name: 'Edit', exact: true }).waitFor();
      await openNote(page, server, first.id, true);
      assert.equal(await page.getByRole('textbox', { name: 'Note body' }).inputValue(), 'Unsaved first change');
      await openNote(page, server, second.id, true);
      assert.equal(await page.getByRole('textbox', { name: 'Note body' }).inputValue(), 'Unsaved second change');
    }],
  ['preserves a new note and its draft when searching and reloading',
    async ({ page }) => {
      await page.getByRole('button', { name: 'New note', exact: true }).first().click();
      await page.getByRole('textbox', { name: 'Note body' }).fill('Writing while searching');
      await page.locator('.search-input').fill('deployment');
      await page.waitForURL(url => url.searchParams.get('q') === 'deployment');
      await page.reload();
      await page.getByRole('button', { name: 'Save', exact: true }).waitFor();
      assert.equal(await page.getByRole('textbox', { name: 'Note body' }).inputValue(), 'Writing while searching');
    }],
  ['a stale draft cannot silently undo a privacy change or lose a browser capture',
    async ({ page, server }) => {
      const note = await createNote(page, server, { title: 'Shared operations', is_public: true });
      await openNote(page, server, note.id, true);
      const body = page.getByRole('textbox', { name: 'Note body' });
      await body.fill('My private additions');
      await request(page, server, '/snippets/' + note.id, 'PUT', {
        kind: 'note', revision: note.revision, title: note.title, content: note.content,
        language: note.language, is_public: false,
      });
      await request(page, server, '/snippets/' + note.id + '/append', 'POST', {
        content: 'Browser capture',
      }, 204);
      await save(page);
      await page.getByText('A newer version was saved.').waitFor();
      assert.equal(await body.inputValue(), 'My private additions');
      assert.match(await page.locator('.note-edit-notice pre').textContent(), /Browser capture/);
      await body.fill('My private additions\n\nBrowser capture');
      await page.getByRole('button', { name: 'I’ve reviewed the latest version' }).click();
      await save(page);
      await page.getByRole('button', { name: 'Edit', exact: true }).waitFor();
      const saved = await request(page, server, '/snippets/' + note.id);
      assert.equal(saved.content, 'My private additions\n\nBrowser capture');
      assert.equal(saved.is_public, false);
    }],
  ['deduplicates a retried capture and accepts independent concurrent captures',
    async ({ page, server }) => {
      const note = await createNote(page, server, { title: 'Capture replay', language: 'plaintext' });
      const route = '/snippets/' + note.id + '/append';
      const capture = { content: 'Captured once', language: 'plaintext', request_id: 'text-capture-retry' };
      await Promise.all([
        request(page, server, route, 'POST', capture, 204),
        request(page, server, route, 'POST', capture, 204),
        request(page, server, route, 'POST', { ...capture, content: 'Another capture', request_id: 'second-capture' }, 204),
      ]);
      await request(page, server, route, 'POST', { ...capture, content: 'Different payload' }, 409);
      const saved = await request(page, server, '/snippets/' + note.id);
      assert.equal(saved.content.match(/Captured once/g).length, 1);
      assert.equal(saved.content.match(/Another capture/g).length, 1);
      assert.equal(saved.revision, note.revision + 2);
    }],
  ['converts only eligible items and protects the folder overview assignment',
    async ({ page, server }) => {
      const folder = await request(page, server, '/folders', 'POST', { name: 'Conversion project' }, 201);
      const legacy = await request(page, server, '/snippets', 'POST', {
        title: 'Legacy runbook', description: 'Legacy description', tags: ['ops'], folder_id: folder.id,
        files: [{ filename: 'legacy.md', content: '# Existing runbook', language: 'markdown' }],
      }, 201);
      const bundle = await request(page, server, '/snippets', 'POST', {
        title: 'Two files', files: [
          { filename: 'one.md', content: 'one', language: 'markdown' },
          { filename: 'two.txt', content: 'two', language: 'plaintext' },
        ],
      }, 201);
      await request(page, server, '/snippets/' + bundle.id + '/convert', 'POST', { kind: 'note' }, 422);
      const converted = await request(page, server, '/snippets/' + legacy.id + '/convert', 'POST', { kind: 'note' });
      assert.equal(converted.id, legacy.id);
      assert.equal(converted.files[0].id, legacy.files[0].id);
      assert.equal(converted.tags[0].name, 'ops');
      assert.equal(converted.folders[0].id, folder.id);
      await request(page, server, '/folders/' + folder.id + '/overview', 'PUT', { snippet_id: legacy.id });
      await openNote(page, server, legacy.id, true);
      await page.getByRole('combobox', { name: 'Format' }).selectOption('plaintext');
      await save(page);
      await page.getByRole('button', { name: 'Remove overview and save' }).waitFor();
      assert.equal((await request(page, server, '/folders/' + folder.id)).overview_snippet_id, legacy.id);
      await page.getByRole('button', { name: 'Remove overview and save' }).click();
      await page.locator('.note-reader .note-plain').waitFor();
      assert.equal((await request(page, server, '/folders/' + folder.id)).overview_snippet_id, undefined);
      const fresh = await request(page, server, '/snippets/' + legacy.id);
      assert.equal(fresh.language, 'plaintext');
      assert.equal(fresh.files[0].language, 'plaintext');
    }],
  ['restores the latest previous version and sends the note through archive and trash',
    async ({ page, server }) => {
      const note = await createNote(page, server, { title: 'History runbook', content: '# Before editing' });
      await openNote(page, server, note.id, true);
      await page.getByRole('textbox', { name: 'Note body' }).fill('# After editing');
      await save(page);
      await page.locator('.note-reader').getByRole('heading', { name: 'After editing' }).waitFor();
      const editor = page.getByRole('region', { name: 'Note workspace' });
      await editor.getByText('More', { exact: true }).click();
      await page.getByRole('button', { name: 'History', exact: true }).click();
      await page.locator('.history-entry').first().getByRole('button', { name: 'Restore', exact: true }).click();
      await page.getByRole('button', { name: 'Restore Now', exact: true }).click();
      await page.locator('.note-reader').getByRole('heading', { name: 'Before editing' }).waitFor();

      await request(page, server, '/snippets/' + note.id + '/archive', 'POST');
      const archived = await request(page, server, '/snippets?kind=note&is_archived=true&q=History+runbook');
      assert.deepEqual(archived.map(item => item.id), [note.id]);
      const targets = await request(page, server, '/snippets/note-targets?kind=note&q=History+runbook');
      assert.deepEqual(targets.targets, []);

      await request(page, server, '/snippets/' + note.id + '/archive', 'POST');
      await request(page, server, '/snippets/' + note.id, 'DELETE', undefined, 204);
      await page.goto(server.url + '/?trash=true&kind=note&snippet=' + note.id);
      await editor.getByRole('button', { name: 'Restore', exact: true }).click();
      await editor.getByRole('button', { name: 'Edit', exact: true }).waitFor();
      assert.equal((await request(page, server, '/snippets/' + note.id)).kind, 'note');
    }],
  ['recovers separate new-note drafts on mobile and clears their local copies on logout',
    async ({ page, server }) => {
      await page.setViewportSize({ width: 390, height: 844 });
      await page.reload();
      const draftUrls = [];
      for (const text of ['Mobile draft one', 'Mobile draft two']) {
        await page.getByRole('button', { name: 'New note', exact: true }).first().click();
        await page.getByRole('textbox', { name: 'Note body' }).fill(text);
        draftUrls.push(page.url());
        await page.getByRole('button', { name: 'Close', exact: true }).click();
      }
      assert.notEqual(draftUrls[0], draftUrls[1]);
      for (const [index, url] of draftUrls.entries()) {
        await page.goto(url);
        const body = page.getByRole('textbox', { name: 'Note body' });
        await body.waitFor();
        assert.equal(await body.inputValue(), 'Mobile draft ' + (index === 0 ? 'one' : 'two'));
        assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true);
        await page.getByRole('button', { name: 'Close', exact: true }).click();
      }
      await page.getByText('2 unsaved note drafts', { exact: true }).waitFor();
      await page.getByRole('button', { name: 'Toggle navigation' }).click();
      await page.getByTitle('Log out and clear local drafts').click();
      await page.waitForURL(server.url + '/login');
      const keys = await page.evaluate(() => [
        ...Object.keys(localStorage), ...Object.keys(sessionStorage),
      ].filter(key => key.startsWith('snipo-note-draft:')));
      assert.deepEqual(keys, []);
    }],
  ['restores ZIP backups with notes in distinct, same-named nested folders',
    async ({ page, browser, server }) => {
      const projects = [
        { name: 'Service A', bodies: ['First runbook', 'Another runbook in a same-named sibling'] },
        { name: 'Service B', bodies: ['Second runbook'] },
      ];
      for (const project of projects) {
        const parent = await request(page, server, '/folders', 'POST', { name: project.name }, 201);
        for (const content of project.bodies) {
          const child = await request(page, server, '/folders', 'POST', { name: 'Runbooks', parent_id: parent.id }, 201);
          await createNote(page, server, { title: 'Runbook', content, folder_id: child.id });
        }
      }
      const backup = await page.request.get(server.url + '/api/v1/backup/export?format=zip');
      assert.equal(backup.status(), 200, await backup.text());
      const destination = await startServer('restore');
      const destinationContext = await browser.newContext();
      try {
        const target = await signIn(destinationContext, destination);
        const backupBytes = await backup.body();
        const upload = () => target.request.post(destination.url + '/api/v1/backup/import', {
          multipart: {
            file: { name: 'notes.zip', mimeType: 'application/zip', buffer: backupBytes },
            strategy: 'merge',
          },
        });
        const imported = await upload();
        assert.equal(imported.status(), 200, await imported.text());
        assert.deepEqual((await imported.json()).data.errors ?? [], []);
        const retried = await upload();
        assert.equal(retried.status(), 200, await retried.text());
        const importedAgain = (await retried.json()).data;
        assert.equal(importedAgain.snippets_imported, 0);
        assert.equal(importedAgain.folders_imported, 0);
        const tree = await request(target, destination, '/folders?tree=true');
        for (const project of projects) {
          const parent = tree.find(folder => folder.name === project.name);
          assert.deepEqual(parent.children.map(folder => folder.name), project.bodies.map(() => 'Runbooks'));
          const bodies = [];
          for (const child of parent.children) {
            const list = await request(target, destination, '/snippets?kind=note&folder_ids=' + child.id);
            assert.equal(list.length, 1);
            bodies.push(list[0].content);
          }
          assert.deepEqual(bodies.sort(), [...project.bodies].sort());
        }
      } finally {
        await destinationContext.close();
      }
    }],
];

const extensionPath = process.env.SNIPO_TEST_EXTENSION;
if (extensionPath && browserType === chromium) {
  cases.push(['runs the Chrome extension against Snipo, edits a capture, and retries an append after a lost reply',
    async ({ page, server }) => {
      const note = await createNote(page, server, { title: 'Extension live target', language: 'plaintext' });
      const token = await request(page, server, '/tokens', 'POST', {
        name: 'Live extension test', password, permissions: 'write',
      }, 201);
      const context = await chromium.launchPersistentContext(path.join(temp, 'extension-profile'), {
        channel: 'chromium', headless: true,
        args: [`--disable-extensions-except=${extensionPath}`, `--load-extension=${extensionPath}`],
      });
      try {
        const worker = context.serviceWorkers()[0] || await context.waitForEvent('serviceworker');
        await worker.evaluate(async ({ url, apiKey }) => {
          await chrome.storage.sync.set({ instanceUrl: url, apiKey, interactiveMode: true });
          const originalFetch = self.fetch.bind(self);
          let loseResponse = true;
          self.fetch = async (input, options) => {
            const response = await originalFetch(input, options);
            if (loseResponse && response.ok && new URL(input).pathname.endsWith('/append')) {
              loseResponse = false;
              return Response.json({ error: { message: 'Connection closed after saving' } }, { status: 503 });
            }
            return response;
          };
        }, { url: server.url, apiKey: token.token });
        const source = await context.newPage();
        await source.route('https://capture.example.test/**', route => route.fulfill({
          contentType: 'text/html',
          body: '<title>Capture fixture</title><pre><code>Live browser capture</code></pre>',
        }));
        await source.goto('https://capture.example.test/runbook');
        const opened = context.waitForEvent('page');
        await source.locator('.snipo-btn').click();
        const dialog = await opened;
        await dialog.getByRole('button', { name: 'Append to a note' }).click();
        await dialog.getByRole('listbox', { name: 'Destination' }).selectOption(note.id);
        const append = dialog.getByRole('button', { name: 'Append selection' });
        await append.click();
        await dialog.getByRole('status').getByText('Connection closed after saving').waitFor();
        const closed = dialog.waitForEvent('close');
        await append.click();
        await closed;
        const saved = await request(page, server, '/snippets/' + note.id);
        assert.equal(saved.content, 'Original body\n\nLive browser capture\n\nSource: https://capture.example.test/runbook');
        assert.equal(saved.revision, note.revision + 1);

        await source.locator('.snipo-btn').waitFor();
        const reopened = context.waitForEvent('page');
        await source.locator('.snipo-btn').click();
        const create = await reopened;
        await create.getByRole('combobox', { name: 'Save as' }).selectOption('note');
        await create.getByRole('combobox', { name: 'Format' }).selectOption('plaintext');
        await create.locator('#snipo-title').fill('Edited browser capture');
        await create.getByRole('textbox', { name: 'Note body' }).fill('  Edited before saving\nNext line');
        await create.getByRole('checkbox', { name: 'Include source link' }).uncheck();
        const created = create.waitForEvent('close');
        await create.getByRole('button', { name: 'Save Note' }).click();
        await created;
        const notes = await request(page, server, '/snippets?kind=note&q=Edited+browser+capture');
        assert.equal(notes.length, 1);
        const capture = await request(page, server, '/snippets/' + notes[0].id);
        assert.equal(capture.content, '  Edited before saving\nNext line');
        assert.equal(capture.language, 'plaintext');
        assert.equal(capture.is_public, false);
      } finally {
        await context.close();
      }
    }]);
}

const previousBinary = process.env.SNIPO_TEST_PREVIOUS_BINARY;
if (previousBinary) {
  cases.push(['upgrades a genuine previous database without reclassifying Markdown snippets',
    async ({ browser }) => {
      assert.ok(existsSync(previousBinary));
      const legacy = await startServer('legacy', previousBinary);
      const oldContext = await browser.newContext();
      let original;
      try {
        const page = await signIn(oldContext, legacy);
        original = await request(page, legacy, '/snippets', 'POST', {
          title: 'Old markdown item', language: 'markdown', content: '# Existing',
        }, 201);
      } finally {
        await oldContext.close();
        await stopServer(legacy);
      }
      const upgraded = await startServer('upgrade', binary, legacy.db);
      const context = await browser.newContext();
      try {
        const page = await signIn(context, upgraded);
        assert.equal((await request(page, upgraded, '/snippets/' + original.id)).kind, 'snippet');
        const converted = await request(page, upgraded, '/snippets/' + original.id + '/convert', 'POST', { kind: 'note' });
        assert.equal(converted.id, original.id);
        assert.equal(converted.content, '# Existing');
      } finally {
        await context.close();
      }
    }]);
}

(async () => {
  let browser;
  let failed = 0;
  let completed = false;
  try {
    assert.ok(browserType, 'Unsupported Playwright browser');
    execFileSync('go', ['build', '-o', binary, './cmd/server'], { cwd: root, stdio: 'inherit' });
    const server = await startServer('primary');
    browser = await browserType.launch();
    const selected = caseFilter ? cases.filter(([name]) => name.includes(caseFilter)) : cases;
    assert.ok(selected.length, 'No matching Notes test cases');
    for (const [name, run] of selected) {
      const context = await browser.newContext({ viewport: { width: 1280, height: 850 } });
      try {
        const page = await signIn(context, server);
        page.setDefaultTimeout(6000);
        const errors = [];
        page.on('pageerror', error => errors.push(error.stack || error.message));
        await run({ page, browser, server });
        assert.deepEqual(errors, [], 'Browser exceptions');
        console.log('PASS ' + name);
      } catch (error) {
        failed++;
        console.error('FAIL ' + name + '\n' + error.stack);
      } finally {
        await context.close();
      }
    }
    assert.equal(failed, 0, `${failed} of ${selected.length} live cases failed`);
    completed = true;
  } finally {
    if (browser) await browser.close();
    await Promise.all(servers.map(stopServer));
    if (!completed) console.log('Test artifacts: ' + temp);
    else rmSync(temp, { recursive: true, force: true });
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
