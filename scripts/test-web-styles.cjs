#!/usr/bin/env node
// Browser regression checks for the shared CSS, using the existing Playwright workspace dependency.
// Run with npm run test:web (requires Go and Playwright's Chromium installed).
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const net = require('node:net');
const { spawn, execFileSync } = require('node:child_process');
const { chromium } = require('playwright');

const root = path.resolve(__dirname, '..');
const temp = fs.mkdtempSync(path.join(os.tmpdir(), 'snipo-web-styles-'));
const password = 'web-styles-test-password';
const palettes = {
  light: { background: 'rgb(255, 255, 255)', surface: 'rgb(248, 250, 252)',
    text: 'rgb(30, 41, 59)', muted: 'rgb(100, 116, 139)', border: 'rgb(203, 213, 225)' },
  dark: { background: 'rgb(13, 17, 23)', surface: 'rgb(22, 27, 34)',
    text: 'rgb(201, 209, 217)', muted: 'rgb(139, 148, 158)', border: 'rgb(48, 54, 61)' }
};
const primary = 'rgb(99, 102, 241)';
const primaryHover = 'rgb(79, 70, 229)';

async function expectStyle(locator, property, expected, message = `${property} should match the theme`) {
  await assertEventually(() => style(locator, property), expected, message);
}

async function assertEventually(actual, expected, message) {
  // Reduced motion still permits a short transition; sample the settled style.
  for (let attempt = 0; attempt < 20; attempt++) {
    if (await actual() === expected) return;
    await new Promise(resolve => setTimeout(resolve, 25));
  }
  assert.equal(await actual(), expected, message);
}

async function checkTheme(page, theme) {
  const palette = palettes[theme];
  assert.equal(await page.locator('html').getAttribute('data-theme'), theme);
  await expectStyle(page.locator('html'), 'colorScheme', theme);
  await expectStyle(page.locator('html'), 'backgroundColor', palette.background);
  await expectStyle(page.locator('html'), 'color', palette.text);
}

async function freePort() {
  const server = net.createServer();
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  const port = server.address().port;
  await new Promise(resolve => server.close(resolve));
  return port;
}

async function style(locator, property) {
  return locator.evaluate((element, property) => getComputedStyle(element)[property], property);
}

async function appReady(page) {
  await page.waitForFunction(() => window.Alpine &&
    document.querySelector('.app-container') &&
    !Alpine.$data(document.querySelector('.app-container')).loading);
}

async function appAction(page, method) {
  await page.evaluate(method => Alpine.$data(document.querySelector('.app-container'))[method](), method);
}

async function checkPage(page) {
  assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true,
    'Page should fit the viewport');
  assert.equal(await page.locator('link[href*="pico"]').count(), 0, 'No framework stylesheet');
}

(async () => {
  let server;
  let browser;
  try {
    const binary = path.join(temp, 'snipo');
    execFileSync('go', ['build', '-o', binary, './cmd/server'], { cwd: root, stdio: 'inherit' });
    const port = await freePort();
    const origin = `http://127.0.0.1:${port}`;
    const log = fs.openSync(path.join(temp, 'server.log'), 'a');
    server = spawn(binary, ['serve'], {
      cwd: temp,
      env: { ...process.env, SNIPO_HOST: '127.0.0.1', SNIPO_PORT: String(port),
        SNIPO_DB_PATH: path.join(temp, 'test.db'), SNIPO_MASTER_PASSWORD: password,
        SNIPO_DISABLE_AUTH: 'false', SNIPO_DEMO_MODE: 'false', SNIPO_BASE_PATH: '',
        SNIPO_SESSION_SECRET: 'web-styles-test-session-secret',
        SNIPO_ENCRYPTION_SALT: 'web-styles-test-encryption-salt' },
      stdio: ['ignore', log, log]
    });
    fs.closeSync(log);
    let ready = false;
    for (let attempt = 0; attempt < 100; attempt++) {
      try {
        ready = (await fetch(origin + '/ping', { signal: AbortSignal.timeout(500) })).ok;
        if (ready) break;
      } catch { /* Wait for the isolated test server. */ }
      await new Promise(resolve => setTimeout(resolve, 100));
    }
    assert.ok(ready, `Server did not start. See ${temp}/server.log`);
    browser = await chromium.launch({ headless: true });

    for (const width of [1440, 390]) {
      for (const theme of ['light', 'dark']) {
        const context = await browser.newContext({ viewport: { width, height: 1000 }, reducedMotion: 'reduce' });
        const page = await context.newPage();
        const badAssets = [];
        page.on('response', response => {
          if (response.url().includes('/static/') && !response.ok()) badAssets.push(response.url());
        });
        await page.addInitScript(theme => {
          if (!localStorage.getItem('snipo-theme')) localStorage.setItem('snipo-theme', theme);
        }, theme);
        await page.goto(origin + '/login');
        await page.waitForFunction(() => !!window.Alpine);
        await page.evaluate(() => document.fonts.ready);
        await checkPage(page);
        const rem = width === 1440 ? 20 : 16;
        assert.equal(await style(page.locator('html'), 'fontSize'), rem + 'px');
        assert.equal(await style(page.locator('.login-card'), 'width'), width === 1440 ? '400px' : '326px');
        assert.equal(await style(page.locator('#password'), 'height'), rem * 3.125 + 'px');
        assert.equal(await style(page.locator('label[for="password"]'), 'marginBottom'), rem * 0.375 + 'px');
        await checkTheme(page, theme);
        await expectStyle(page.locator('.login-card'), 'backgroundColor', palettes[theme].surface);
        const signIn = page.getByRole('button', { name: 'Sign In', exact: true });
        await signIn.focus();
        await expectStyle(signIn, 'backgroundColor', primary, 'Focused primary button keeps its brand color');
        await signIn.hover();
        await expectStyle(signIn, 'backgroundColor', primaryHover, 'Primary hover must not turn blue');
        await expectStyle(signIn, 'color', 'rgb(255, 255, 255)');
        await page.screenshot({ path: path.join(temp, `${width}-${theme}-login.png`) });
        await page.locator('#password').fill(password);
        await page.getByRole('button', { name: 'Sign In', exact: true }).click();
        await page.waitForURL(origin + '/');
        await appReady(page);
        await checkPage(page);
        assert.equal(await style(page.locator('.sidebar-nav'), 'display'), 'block');

        // Switch in both directions without reloading; saved colors must survive reload.
        for (const activeTheme of [theme === 'light' ? 'dark' : 'light', theme]) {
          await page.getByTitle('Toggle theme', { exact: true }).click();
          await checkTheme(page, activeTheme);
          assert.equal(await page.evaluate(() => localStorage.getItem('snipo-theme')), activeTheme);
          await page.reload();
          await appReady(page);
          await checkTheme(page, activeTheme);
        }

        // Real controls must remain keyboard accessible, visibly focused, and themed.
        await appAction(page, 'openSettings');
        const modal = page.locator('.settings-modal');
        const inactiveTab = modal.locator('.settings-tab:not(.active)').first();
        await inactiveTab.hover();
        await expectStyle(inactiveTab, 'color', palettes[theme].text, 'Tab hover uses theme text, not button white');
        await expectStyle(inactiveTab, 'backgroundColor', palettes[theme].border);
        const checkbox = modal.locator('input[type="checkbox"]:visible').first();
        await checkbox.focus();
        assert.notEqual(await style(checkbox, 'boxShadow'), 'none');
        const wasChecked = await checkbox.isChecked();
        await checkbox.press('Space');
        assert.equal(await checkbox.isChecked(), !wasChecked);
        await checkbox.press('Space');
        await modal.locator('.settings-tab').filter({ hasText: 'Appearance' }).click();
        const select = modal.locator('select:visible').first();
        await select.focus();
        assert.notEqual(await style(select, 'boxShadow'), 'none', 'Select focus must remain visible');
        await expectStyle(select, 'backgroundColor', palettes[theme].surface, 'Focused component controls keep their surface color');
        await expectStyle(select, 'color', palettes[theme].text);
        const summary = modal.locator('summary');
        await summary.focus();
        await summary.press('Enter');
        assert.notEqual(await style(summary, 'outlineStyle'), 'none');
        assert.equal(await modal.locator('details').getAttribute('open'), '');
        const disabled = modal.locator('button:disabled').first();
        assert.equal(await style(disabled, 'opacity'), '0.5');
        await page.screenshot({ path: path.join(temp, `${width}-${theme}-settings.png`) });
        await page.evaluate(() => document.activeElement.blur());
        await appAction(page, 'closeSettings');

        // Seed through the API, then exercise the actual editor and Markdown preview.
        const created = await page.request.post(origin + '/api/v1/snippets', { data: {
          title: 'Styles regression', description: 'Theme color regression fixture', language: 'markdown', is_public: true,
          content: '# Heading\n\nA **bold** paragraph with `code`.\n\n- First\n- Second\n\n> Quoted text\n\n| A | B |\n|---|---|\n| One | Two |'
        } });
        assert.equal(created.ok(), true);
        const { data: snippet } = await created.json();
        await page.reload();
        await appReady(page);
        const card = page.locator('.snippet-card').filter({ hasText: 'Styles regression' }).first();
        await expectStyle(card, 'backgroundColor', palettes[theme].surface);
        await expectStyle(card, 'borderTopColor', palettes[theme].border);
        await expectStyle(card.locator('.snippet-card-title'), 'color', palettes[theme].text);
        await expectStyle(card.locator('.snippet-card-description'), 'color', palettes[theme].muted);
        await page.screenshot({ path: path.join(temp, `${width}-${theme}-snippets.png`) });
        await card.locator('.snippet-card-title').click();
        await page.locator('.editor-container').waitFor({ state: 'visible' });
        const copyButton = page.getByTitle('Copy options', { exact: true });
        await copyButton.hover();
        await expectStyle(copyButton, 'color', palettes[theme].text, 'Action hover uses theme text');
        await expectStyle(copyButton, 'borderTopColor', palettes[theme].border);
        await page.screenshot({ path: path.join(temp, `${width}-${theme}-editor.png`) });
        assert.equal(await page.locator('.preview-markdown-scroll h1').innerText(), 'Heading');
        assert.equal(await page.locator('.preview-markdown-scroll table').count(), 1);
        await appAction(page, 'startEditing');
        await page.locator('#code-editor').waitFor({ state: 'visible' });
        const title = page.locator('.editor-title-input');
        await title.fill('مقتطف تجريبي');
        await title.press('Tab');
        assert.equal(await style(title, 'direction'), 'rtl');

        // Isolate public-page styling from its pre-existing UUID URL parsing issue.
        await page.route(origin + '/api/v1/snippets/public/*', route => route.fulfill({ json: { data: snippet } }));
        await page.goto(origin + '/s/' + snippet.id);
        await page.waitForFunction(() => window.Alpine &&
          !Alpine.$data(document.querySelector('[x-data]')).loading);
        assert.equal(await page.locator('.public-title').innerText(), 'Styles regression');
        await checkTheme(page, theme);
        await expectStyle(page.locator('.public-header'), 'backgroundColor', palettes[theme].surface);
        const download = page.getByTitle('Download file', { exact: true });
        await download.hover();
        await expectStyle(download, 'backgroundColor', palettes[theme].border, 'Public file actions use the theme hover surface');
        await page.screenshot({ path: path.join(temp, `${width}-${theme}-public.png`) });
        await checkPage(page);
        assert.deepEqual(badAssets, [], 'All stylesheets and scripts must load locally');
        await context.close();
        console.log(`PASS: ${width}px ${theme}: login, settings, native controls, editor, Markdown, RTL, public page`);
      }
    }
  } finally {
    if (browser) await browser.close();
    if (server && server.exitCode === null && server.signalCode === null) {
      const exited = new Promise(resolve => server.once('exit', resolve));
      server.kill();
      await exited;
    }
    console.log('Test artifacts: ' + temp);
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
