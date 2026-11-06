const assert = require('node:assert/strict');
const http = require('node:http');

exports.checkMarkdownSecurity = async function (page, origin, password) {
  const victim = await page.context().newPage();
  try {
    const created = await page.request.post(origin + '/api/v1/snippets', { data: {
      title: 'Security regression', language: 'markdown', is_public: true,
      content: '<div><img src="/__security_missing__" onerror="window.__snipoSecurityProbe=true"></div>'
    } });
    assert.equal(created.ok(), true);
    const { data: snippet } = await created.json();
    await victim.goto(origin + '/?snippet=' + snippet.id);
    await victim.locator('.preview-markdown-scroll').waitFor({ state: 'visible' });
    await victim.waitForTimeout(150);
    assert.equal(await victim.evaluate(() => window.__snipoSecurityProbe), undefined,
      'Untrusted nested HTML must not execute in a stored snippet');
    assert.equal(await victim.locator('[onerror]').count(), 0,
      'Sanitizer must inspect descendants before unwrapping elements');

    const fallback = await victim.evaluate(() => {
      const app = Alpine.$data(document.querySelector('.app-container'));
      const saved = window.marked;
      window.marked = undefined;
      try { return app.renderMarkdown('<img src=x onerror="window.__snipoSecurityProbe=true">'); }
      finally { window.marked = saved; }
    });
    assert.ok(fallback.includes('&lt;img'), 'Missing Markdown parser must fall back to escaped text');

    // A real second loopback server gives the browser a distinct origin on the
    // same site. Requests use normal browser cookie and Fetch Metadata behavior.
    const source = http.createServer((_req, res) => {
      res.setHeader('Content-Type', 'text/html');
      res.end('<title>Isolated same-site origin</title>');
    });
    await new Promise(resolve => source.listen(0, '127.0.0.1', resolve));
    const attacker = await page.context().newPage();
    try {
      await attacker.goto('http://127.0.0.1:' + source.address().port + '/csrf-probe');
      const denied = attacker.waitForResponse(response =>
        response.url() === origin + '/api/v1/snippets' && response.request().method() === 'POST');
      await attacker.evaluate(async origin => {
        await fetch(origin + '/api/v1/snippets', {
          method: 'POST', mode: 'no-cors', credentials: 'include',
          body: new TextEncoder().encode(JSON.stringify({
            title: 'CSRF security marker', content: 'test-only', language: 'plaintext'
          }))
        });
      }, origin);
      assert.equal((await denied).status(), 403, 'Server must reject the cross-origin browser write');
      const response = await page.request.get(origin + '/api/v1/snippets?q=CSRF%20security%20marker');
      assert.equal(response.status(), 200);
      assert.equal(((await response.json()).data || []).length, 0,
        'Another same-site origin must not write using session cookies');
    } finally {
      await attacker.close();
      source.closeAllConnections();
      await new Promise(resolve => source.close(resolve));
    }

    // Exercise the actual authentication middleware: only a validated token
    // bypasses browser-origin checks, and its permission level still applies.
    for (const permission of ['write', 'read']) {
      const minted = await page.request.post(origin + '/api/v1/tokens', { data: {
        name: 'Security regression ' + permission, permissions: permission, password
      } });
      assert.equal(minted.status(), 201);
      const token = (await minted.json()).data.token;
      const result = await page.request.post(origin + '/api/v1/snippets', {
        headers: { Authorization: 'Bearer ' + token, Origin: 'https://client.example.test' },
        data: { title: 'Explicit token ' + permission, content: 'test-only', language: 'plaintext' }
      });
      assert.equal(result.status(), permission === 'write' ? 201 : 403);
    }
    const forged = await page.request.post(origin + '/api/v1/snippets', {
      headers: { Authorization: 'Bearer invalid-test-token', Origin: 'https://client.example.test' },
      data: { title: 'Invalid token', content: 'test-only', language: 'plaintext' }
    });
    assert.equal(forged.status(), 403, 'An invalid explicit token must not bypass session protection');

    const anonymous = await page.context().browser().newContext();
    try {
      const url = origin + '/api/v1/snippets/public/' + snippet.id;
      const before = await anonymous.request.get(url);
      assert.equal(before.status(), 200);
      const body = await before.json();
      const filename = body.data.files?.[0]?.filename || 'snippet.md';
      const removed = await page.request.delete(origin + '/api/v1/snippets/' + snippet.id);
      assert.equal(removed.ok(), true);
      assert.equal((await anonymous.request.get(url)).status(), 404,
        'Trash must revoke anonymous snippet access');
      assert.equal((await anonymous.request.get(url + '/files/' + encodeURIComponent(filename))).status(), 404,
        'Trash must revoke anonymous file access');
    } finally { await anonymous.close(); }
  } finally { await victim.close(); }
};

exports.checkLoginRedirectSecurity = async function (browser, origin, password) {
  const context = await browser.newContext();
  const page = await context.newPage();
  try {
    // Intercept the destination: this never contacts a real external website.
    await page.route('**://login-redirect.example.test/**', route => route.fulfill({
      contentType: 'text/html', body: '<title>Redirect marker</title>'
    }));
    await page.goto(origin + '/login?next=' + encodeURIComponent('//login-redirect.example.test/marker'));
    await page.locator('#password').fill(password);
    await page.getByRole('button', { name: 'Sign In', exact: true }).click();
    await page.waitForURL(url => !url.pathname.endsWith('/login'));
    assert.equal(new URL(page.url()).origin, new URL(origin).origin,
      'Login must never redirect to a protocol-relative external destination');
    await context.clearCookies();
    await page.goto(origin + '/login?next=' + encodeURIComponent('?q=return%20query'));
    await page.locator('#password').fill('incorrect-fixture-password');
    await page.getByRole('button', { name: 'Sign In', exact: true }).click();
    await page.waitForFunction(() => {
      const form = Alpine.$data(document.querySelector('.login-container'));
      return !form.loading && !!form.error;
    });
    assert.equal(await page.getByRole('button', { name: 'Sign In', exact: true }).isEnabled(), true,
      'Failed login must leave the form usable');
    // Respect the first failed-attempt delay before checking the successful retry.
    await page.waitForTimeout(1100);
    await page.locator('#password').fill(password);
    await page.getByRole('button', { name: 'Sign In', exact: true }).click();
    await page.waitForURL(url => url.pathname === new URL(origin + '/').pathname);
    assert.equal(new URL(page.url()).searchParams.get('q'), 'return query');
  } finally { await context.close(); }
};

exports.checkLogoutSecurity = async function (browser, origin, password) {
  const context = await browser.newContext();
  const page = await context.newPage();
  const marker = 'PRIVATE-LOGOUT-CONTENT-MARKER';
  try {
    await page.goto(origin + '/login');
    await page.locator('#password').fill(password);
    await page.getByRole('button', { name: 'Sign In', exact: true }).click();
    await page.waitForURL(origin + '/');
    await page.waitForFunction(() => window.Alpine && !Alpine.$data(document.querySelector('.app-container')).loading);
    // Session privacy guards must preserve recovery during ordinary navigation.
    await page.evaluate(() => {
      const app = Alpine.$data(document.querySelector('.app-container'));
      app.newSnippet();
      app.editingSnippet.title = 'Navigation recovery fixture';
      app.editingSnippet.files[0].content = 'NORMAL-NAVIGATION-DRAFT-MARKER';
      app.saveDraft();
    });
    await page.goto(origin + '/ping');
    await page.goBack();
    await page.waitForFunction(() => window.Alpine && !Alpine.$data(document.querySelector('.app-container')).loading);
    assert.ok((await page.evaluate(() => sessionStorage.getItem('snipo-draft'))).includes('NORMAL-NAVIGATION-DRAFT-MARKER'));
    assert.equal((await page.request.get(origin + '/api/v1/auth/check')).status(), 200);
    // Return to the list so the recovery action can be exercised explicitly.
    await page.goto(origin + '/');
    await page.waitForFunction(() => window.Alpine && Alpine.$data(document.querySelector('.app-container')).hasDraft);
    await page.evaluate(() => Alpine.$data(document.querySelector('.app-container')).restoreDraft());
    await page.waitForFunction(() => Alpine.$data(document.querySelector('.app-container')).codeEditor?.getValue() === 'NORMAL-NAVIGATION-DRAFT-MARKER');
    const created = await page.request.post(origin + '/api/v1/snippets', { data: {
      title: 'Private logout fixture', content: marker, language: 'plaintext', is_public: false
    } });
    assert.equal(created.status(), 201);
    const { data: snippet } = await created.json();
    const peer = await context.newPage();
    await peer.goto(origin + '/?snippet=' + snippet.id);
    await peer.waitForFunction(marker => document.body.textContent.includes(marker), marker);
    await peer.evaluate(() => {
      const app = Alpine.$data(document.querySelector('.app-container'));
      app.startEditing();
      app.saveDraft();
      app.scheduleAutoSave();
    });
    await page.evaluate(() => {
      const app = Alpine.$data(document.querySelector('.app-container'));
      app.newSnippet();
      app.editingSnippet.title = 'Private local draft';
      app.editingSnippet.files[0].content = 'PRIVATE-DRAFT-LOGOUT-MARKER';
      app.saveDraft();
    });
    assert.ok((await page.evaluate(() => localStorage.getItem('snipo-draft'))).includes('PRIVATE-DRAFT-LOGOUT-MARKER'));
    const privateResponse = await page.request.get(origin + '/api/v1/snippets');
    assert.equal(privateResponse.headers()['cache-control'], 'no-store');
    const assets = await page.evaluate(() => window.SNIPO_CONFIG.assetBasePath);
    const publicAsset = await page.request.get(new URL(assets + '/css/main.css', origin).href);
    assert.ok(publicAsset.headers()['cache-control'].includes('immutable'), 'Public assets must keep their cache policy');

    // A server failure must not be mistaken for a confirmed logout.
    await page.route('**/api/v1/auth/logout', route => route.fulfill({
      status: 500, contentType: 'application/json', body: '{"error":{"message":"Fixture logout failure"}}'
    }));
    await page.evaluate(() => Alpine.$data(document.querySelector('.app-container')).logout());
    assert.notEqual(page.url(), origin + '/login');
    assert.equal((await page.request.get(origin + '/api/v1/auth/check')).status(), 200);
    await page.waitForFunction(() => document.body.textContent.includes('Could not log out.'));
    await page.unroute('**/api/v1/auth/logout');
    await page.evaluate(() => Alpine.$data(document.querySelector('.app-container')).logout());
    await page.waitForURL(origin + '/login');
    assert.equal((await page.request.get(origin + '/api/v1/snippets')).status(), 401,
      'Fixture must actually be logged out at the server');
    await peer.waitForURL(origin + '/login');
    // Let the old autosave deadline pass; it must not recreate discarded code.
    await peer.waitForTimeout(2100);
    const remaining = await page.evaluate(() => ({
      localDraft: (localStorage.getItem('snipo-draft') || '').includes('PRIVATE-DRAFT-LOGOUT-MARKER'),
      tabDraft: (sessionStorage.getItem('snipo-draft') || '').includes('PRIVATE-DRAFT-LOGOUT-MARKER')
    }));
    remaining.peerPrivate = await peer.evaluate(marker => document.body.textContent.includes(marker), marker);
    assert.deepEqual(remaining, { localDraft: false, tabDraft: false, peerPrivate: false },
      'Confirmed logout must remove private recovery data and lock other open tabs');
    assert.equal(await peer.evaluate(() => sessionStorage.getItem('snipo-draft')), null);
    await page.goBack();
    await page.waitForURL(origin + '/login');
    assert.equal(await page.evaluate(marker => document.body.textContent.includes(marker), marker), false,
      'Back navigation after logout must not restore private content');
  } finally { await context.close(); }
};
