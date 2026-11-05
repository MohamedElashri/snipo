const { test, expect } = require('./fixtures');

test('extension background script is running and popup opens', async ({ page, extensionId }) => {
  // Wait for the background service worker or extension to be ready.
  // The fixture `extensionId` resolves after the service worker or background script is ready.
  expect(extensionId).toBeTruthy();

  // Navigate to the extension's popup HTML
  await page.goto(`chrome-extension://${extensionId}/options/options.html`);

  // Verify the page title or a specific element to ensure it loaded correctly
  await expect(page).toHaveTitle(/Snipo/i);
});

test('quick save prevents duplicate requests and recovers after failure', async ({ page, context, extensionId }) => {
  const background = context.serviceWorkers().find(worker => worker.url().includes(extensionId));
  await background.evaluate(async () => {
    await chrome.storage.sync.set({ instanceUrl: 'https://snipo.test', apiKey: 'test-token', interactiveMode: false });
    self.saveRequests = 0;
    self.fetch = async () => {
      self.saveRequests++;
      await new Promise(resolve => setTimeout(resolve, 200));
      return Response.json({ error: { message: 'Temporary failure' } }, { status: 503 });
    };
  });
  await page.route('https://code.test/**', route => route.fulfill({
    contentType: 'text/html', body: '<html><title>Code</title><body><pre><code>console.log("hello");</code></pre></body></html>'
  }));
  await page.goto('https://code.test/example');
  const button = page.locator('.snipo-btn');
  await expect(button).toHaveCount(1);
  await button.evaluate(element => { element.click(); element.click(); });
  await expect(button).toBeDisabled();
  await expect(page.locator('.snipo-toast')).toContainText('Temporary failure');
  expect(await background.evaluate(() => self.saveRequests)).toBe(1);
  await expect(button).toBeEnabled();
  await button.evaluate(element => element.click());
  await expect.poll(() => background.evaluate(() => self.saveRequests)).toBe(2);
});

test('saving preserves an explicitly empty description and defaults an omitted one', async ({ page, context, extensionId }) => {
  const background = context.serviceWorkers().find(worker => worker.url().includes(extensionId));
  await background.evaluate(async () => {
    await chrome.storage.sync.set({ instanceUrl: 'https://snipo.test', apiKey: 'test-token' });
    self.savedPayloads = [];
    self.fetch = async (_url, options) => {
      self.savedPayloads.push(JSON.parse(options.body));
      return Response.json({ data: { id: 'saved' } });
    };
  });
  await page.goto(`chrome-extension://${extensionId}/options/options.html`);
  for (const description of ['', undefined]) {
    const response = await page.evaluate(description => chrome.runtime.sendMessage({
      action: 'saveSnippet', code: '    indented();\n', title: 'Example', description
    }), description);
    expect(response.success).toBe(true);
  }
  const payloads = await background.evaluate(() => self.savedPayloads);
  expect(payloads[0].description).toBe('');
  expect(payloads[0].files[0].content).toBe('    indented();\n');
  expect(payloads[1].description).toMatch(/^Saved from /);
});

test('code-block saves retain leading spaces and trailing newlines', async ({page,context,extensionId}) => {
  const bg=context.serviceWorkers().find(w=>w.url().includes(extensionId));
  await bg.evaluate(async()=>{await chrome.storage.sync.set({instanceUrl:'https://snipo.test',apiKey:'test',interactiveMode:false});self.payload=null;self.fetch=async(_u,o)=>{self.payload=JSON.parse(o.body);return Response.json({data:{id:'saved'}})}});
  await page.route('https://code.test/**',r=>r.fulfill({contentType:'text/html',body:'<pre><code>    indented();\n\n</code></pre>'}));
  await page.goto('https://code.test/whitespace');await page.locator('.snipo-btn').click();
  await expect.poll(()=>bg.evaluate(()=>self.payload?.files[0].content)).toBe('    indented();\n\n');
});

test('repeated interactive clicks open only one save dialog',async({page,context,extensionId})=>{
  const bg=context.serviceWorkers().find(w=>w.url().includes(extensionId));
  await bg.evaluate(async()=>{await chrome.storage.sync.set({instanceUrl:'https://snipo.test',apiKey:'test',interactiveMode:true});self.fetch=async()=>{await new Promise(r=>setTimeout(r,300));return Response.json({data:[]})}});
  await page.route('https://code.test/**',r=>r.fulfill({contentType:'text/html',body:'<pre><code>hello</code></pre>'}));
  await page.goto('https://code.test/dialog');await page.locator('.snipo-btn').evaluate(b=>{b.click();setTimeout(()=>b.click(),50)});
  await page.waitForTimeout(1000);await expect(page.locator('.snipo-modal-overlay')).toHaveCount(1);
});

test('connection check times out and restores the test button',async({page,extensionId})=>{
  await page.goto(`chrome-extension://${extensionId}/options/options.html`);
  test.setTimeout(45000);
  await page.locator('[data-tab="tab-settings"]').click();
  await page.route('https://stalled.test/**',()=>{});
  await page.locator('#instanceUrl').fill('https://stalled.test');await page.locator('#apiKey').fill('test');
  await page.locator('#test-btn').click();
  await expect(page.locator('#test-btn')).toBeEnabled({timeout:25000});await expect(page.locator('#status')).toContainText(/timed out/i);
});
