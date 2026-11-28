const { test, expect } = require('./fixtures');

test('capture chooses a Markdown file, can retry, and appends the original text and source', async ({ page, context, extensionId }) => {
  const background = context.serviceWorkers().find(worker => worker.url().includes(extensionId));
  await background.evaluate(async () => {
    await chrome.storage.sync.set({ instanceUrl: 'https://snipo.test', apiKey: 'test-token', interactiveMode: true });
    self.noteRequests = [];
    self.fetch = async (url, options) => {
      const request = new URL(url);
      if (request.pathname.endsWith('/note-targets')) {
        const query = request.searchParams.get('q');
        const page = Number(request.searchParams.get('page'));
        const targets = page === 2
          ? [{ snippet_id: 'legacy', title: 'Legacy runbook', language: 'markdown' }]
          : [{ snippet_id: 'runbook', title: query ? 'Selected runbook' : 'Other runbook', file_id: 7, filename: 'journal.md', language: 'markdown' }];
        return Response.json({ data: { targets, has_more: page === 1 && !query } });
      }
      if (request.pathname.endsWith('/append')) {
        self.noteRequests.push({ path: request.pathname, body: JSON.parse(options.body) });
        if (self.noteRequests.length === 1) {
          return Response.json({ error: { message: 'Temporary failure' } }, { status: 503 });
        }
        return new Response(null, { status: 204 });
      }
      return Response.json({ data: null });
    };
  });
  await page.route('https://code.test/**', route => route.fulfill({
    contentType: 'text/html',
    body: '<title>Capture</title><pre><code class="language-javascript">  console.log("hello");\n</code></pre>',
  }));
  await page.goto('https://code.test/example_(1)?v=2');
  const pending = context.waitForEvent('page');
  await page.locator('.snipo-btn').click();
  const dialog = await pending;
  await dialog.waitForURL(/\/save\/save.html\?draft=/);
  const open = dialog.getByRole('button', { name: 'Append to a note' });
  await open.click();
  await dialog.getByRole('combobox', { name: 'Append to' }).selectOption('snippet');
  const list = dialog.getByRole('listbox', { name: 'Destination' });
  await expect(list.getByRole('option', { name: 'Other runbook · journal.md' })).toBeAttached();
  await dialog.getByRole('button', { name: 'Load more' }).click();
  await expect(list.getByRole('option', { name: 'Legacy runbook · Markdown' })).toBeAttached();
  await dialog.getByRole('button', { name: 'Back' }).click();
  await expect(dialog.getByRole('button', { name: 'Save Snippet' })).toBeVisible();
  await open.click();
  await dialog.getByRole('searchbox', { name: 'Find by title or filename' }).fill('selected');
  await expect(list.getByRole('option', { name: 'Selected runbook · journal.md' })).toBeAttached();
  const submit = dialog.getByRole('button', { name: 'Append selection' });
  await expect(submit).toBeDisabled();
  await list.selectOption('runbook');
  await submit.click();
  await expect(dialog.getByRole('status')).toHaveText('Temporary failure');
  await expect(submit).toBeEnabled();
  const closed = dialog.waitForEvent('close');
  await submit.click();
  await closed;
  const requests = await background.evaluate(() => self.noteRequests);
  expect(requests[0].body.request_id).toMatch(/^[0-9a-f-]{36}:7$/);
  expect(requests).toEqual(Array(2).fill({
    path: '/api/v1/snippets/runbook/append',
    body: {
      file_id: 7,
      language: 'markdown',
      request_id: requests[0].body.request_id,
      content: '~~~javascript\n  console.log("hello");\n~~~\n\n[Source](<https://code.test/example_(1)?v=2>)',
    },
  }));
});

for (const includeSource of [true, false]) {
  test(`edits a private plain text capture ${includeSource ? 'with' : 'without'} its source`, async ({ page, context, extensionId }) => {
    const background = context.serviceWorkers().find(worker => worker.url().includes(extensionId));
    await background.evaluate(async () => {
      await chrome.storage.sync.set({ instanceUrl: 'https://snipo.test', apiKey: 'test-token', interactiveMode: true });
      self.createdNote = null;
      self.fetch = async (url, options) => {
        const request = new URL(url);
        if (request.pathname === '/api/v1/snippets' && options.method === 'POST') {
          self.createdNote = JSON.parse(options.body);
          return Response.json({ data: { ...self.createdNote, id: 'new-note' } }, { status: 201 });
        }
        return Response.json({ data: [] });
      };
    });
    await page.route('https://code.test/**', route => route.fulfill({
      contentType: 'text/html',
      body: '<title>Capture</title><pre><code class="language-javascript">console.log("hello");</code></pre>',
    }));
    await page.goto('https://code.test/runbook');
    const pending = context.waitForEvent('page');
    await page.locator('.snipo-btn').click();
    const dialog = await pending;
    await dialog.waitForURL(/\/save\/save.html\?draft=/);
    await dialog.getByRole('combobox', { name: 'Save as' }).selectOption('note');
    await dialog.getByRole('combobox', { name: 'Format' }).selectOption('plaintext');
    await expect(dialog.locator('#snipo-filename')).toBeHidden();
    await dialog.locator('#snipo-title').fill('Runbook');
    const body = dialog.getByRole('textbox', { name: 'Note body' });
    await expect(body).toHaveValue('console.log("hello");');
    await body.fill('');
    await dialog.getByRole('button', { name: 'Save Note' }).click();
    await expect(dialog.locator('#save-status')).toContainText('Write some content');
    await body.fill('  Edited selection\nSecond line');
    if (!includeSource) await dialog.getByRole('checkbox', { name: 'Include source link' }).uncheck();
    await dialog.getByRole('combobox', { name: 'Save as' }).selectOption('snippet');
    await expect(body).toBeHidden();
    await dialog.getByRole('combobox', { name: 'Save as' }).selectOption('note');
    await expect(body).toHaveValue('  Edited selection\nSecond line');
    const closed = dialog.waitForEvent('close');
    await dialog.getByRole('button', { name: 'Save Note' }).click();
    await closed;
    expect(await background.evaluate(() => self.createdNote)).toEqual({
      kind: 'note',
      title: 'Runbook',
      language: 'plaintext',
      content: '  Edited selection\nSecond line' + (includeSource ? '\n\nSource: https://code.test/runbook' : ''),
      tags: [],
      is_public: false,
    });
  });
}
