const { test, expect } = require('./fixtures');

async function configuredPage(page, context, extensionId, interactiveMode) {
  const background = context.serviceWorkers().find(worker => worker.url().includes(extensionId));
  await background.evaluate(async interactiveMode => {
    await chrome.storage.sync.set({ instanceUrl: 'https://snipo.test', apiKey: 'test-only-token', interactiveMode });
    self.saveRequests = 0;
    self.savedPayload = null;
    self.fetch = async (url, options) => {
      if (options.method === 'POST') { self.saveRequests++; self.savedPayload = JSON.parse(options.body); }
      return Response.json({ data: String(url).includes('/folders') ? [{ id: 1, name: 'PRIVATE-FOLDER-MARKER' }] : [] });
    };
  }, interactiveMode);
  await page.route('https://code.test/**', route => route.fulfill({
    contentType: 'text/html', body: '<title>Untrusted website</title><pre><code>sample</code></pre>'
  }));
  await page.goto('https://code.test/security');
  await expect(page.locator('.snipo-btn')).toHaveCount(1);
  return background;
}

test('website scripts cannot trigger authenticated quick saves', async ({ page, context, extensionId }) => {
  const background = await configuredPage(page, context, extensionId, false);
  await page.evaluate(() => {
    const button = document.querySelector('.snipo-btn');
    button.click();
    button.dispatchEvent(new MouseEvent('click', { bubbles: true }));
  });
  await page.waitForTimeout(400);
  expect(await background.evaluate(() => self.saveRequests)).toBe(0);
  await page.locator('.snipo-btn').click();
  await expect.poll(() => background.evaluate(() => self.saveRequests)).toBe(1);
});

test('interactive save metadata and edits stay outside the website document', async ({ page, context, extensionId }) => {
  const background = await configuredPage(page, context, extensionId, true);
  const pending = context.waitForEvent('page');
  await page.locator('.snipo-btn').click();
  const dialog = await pending;
  await dialog.waitForURL(/\/save\/save.html\?draft=/);
  await expect(dialog.locator('#snipo-folder')).toContainText('PRIVATE-FOLDER-MARKER');
  expect(await page.evaluate(() => document.body.textContent.includes('PRIVATE-FOLDER-MARKER'))).toBe(false);
  await expect(page.locator('#snipo-save')).toHaveCount(0);
  await page.evaluate(() => {
    window.capturedKeys = [];
    document.addEventListener('keydown', event => window.capturedKeys.push(event.key), true);
    document.body.insertAdjacentHTML('beforeend', '<input id="snipo-title" value="Website tampering"><button id="snipo-save">Fake save</button>');
    document.querySelector('#snipo-save').click();
  });
  await dialog.locator('#snipo-title').fill('Private dialog title');
  expect(await page.evaluate(() => window.capturedKeys)).toEqual([]);
  expect(await background.evaluate(() => self.saveRequests)).toBe(0);
  await dialog.locator('#snipo-save').click();
  await expect.poll(() => background.evaluate(() => self.savedPayload?.title)).toBe('Private dialog title');
  expect(await background.evaluate(() => self.savedPayload?.description)).toBe('Saved from https://code.test/security');
  await expect.poll(() => background.evaluate(async () => Object.keys(await chrome.storage.session.get(null)).filter(k => k.startsWith('snipoSaveDraft:')).length)).toBe(0);
});

test('website scripts cannot open interactive save windows or fetch metadata', async ({ page, context, extensionId }) => {
  const background = await configuredPage(page, context, extensionId, true);
  const existingPages = context.pages().length;
  await page.evaluate(() => document.querySelector('.snipo-btn').click());
  await page.waitForTimeout(400);
  expect(context.pages()).toHaveLength(existingPages);
  expect(await background.evaluate(async () => Object.keys(await chrome.storage.session.get(null)).filter(k => k.startsWith('snipoSaveDraft:')).length)).toBe(0);
});

test('closing a popup before window creation returns cleans its draft and permits another save', async ({ page, context, extensionId }) => {
  const background = await configuredPage(page, context, extensionId, true);
  await background.evaluate(() => {
    const create = chrome.windows.create.bind(chrome.windows);
    chrome.windows.create = async options => {
      const popup = await create(options);
      await new Promise(resolve => { self.finishWindowCreation = resolve; });
      return popup;
    };
  });
  const pending = context.waitForEvent('page');
  await page.locator('.snipo-btn').click();
  const dialog = await pending;
  await expect(dialog.locator('#snipo-save')).toBeVisible();
  await dialog.close();
  await background.evaluate(() => self.finishWindowCreation());
  await expect.poll(() => background.evaluate(async () => Object.keys(await chrome.storage.session.get(null)).filter(key => key.startsWith('snipoSaveDraft:')).length)).toBe(0);
  const reopened = context.waitForEvent('page');
  await page.locator('.snipo-btn').click();
  const next = await reopened;
  await expect(next.locator('#snipo-save')).toBeVisible();
  // Release the second creation too, then exercise the ordinary save path.
  await background.evaluate(() => self.finishWindowCreation());
  await next.locator('#snipo-save').click();
  await expect.poll(() => background.evaluate(() => self.saveRequests)).toBe(1);
  await expect.poll(() => background.evaluate(async () => Object.keys(await chrome.storage.session.get(null)).filter(key => key.startsWith('snipoSaveDraft:')).length)).toBe(0);
});
