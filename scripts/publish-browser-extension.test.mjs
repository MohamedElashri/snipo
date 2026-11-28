import assert from 'node:assert/strict';
import { generateKeyPairSync } from 'node:crypto';
import { mkdtemp, rm, writeFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { after, before, test } from 'node:test';
import { publishChrome, publishFirefox } from './publish-browser-extension.mjs';

const version = '2.2.6';
const zip = Buffer.from('extension artifact');
const source = Buffer.from('review source');
let dir, packagePath, sourcePath;

before(async () => {
  dir = await mkdtemp(path.join(os.tmpdir(), 'snipo-store-tests-'));
  packagePath = path.join(dir, 'extension.zip');
  sourcePath = path.join(dir, 'source.zip');
  await Promise.all([writeFile(packagePath, zip), writeFile(sourcePath, source)]);
  const { privateKey } = generateKeyPairSync('rsa', { modulusLength: 2048 });
  process.env.CHROME_PUBLISHER_ID = 'test-publisher';
  process.env.CHROME_SERVICE_ACCOUNT_JSON = JSON.stringify({
    client_email: 'test@example.invalid',
    private_key: privateKey.export({ type: 'pkcs8', format: 'pem' }),
  });
  process.env.AMO_JWT_ISSUER = 'test-issuer';
  process.env.AMO_JWT_SECRET = 'test-secret';
});

after(async () => rm(dir, { recursive: true, force: true }));

test('Chrome uploads the release ZIP before submitting it for review', async t => {
  const fetch = t.mock.method(globalThis, 'fetch', async (url, options) => {
    if (url === 'https://oauth2.googleapis.com/token') {
      const claims = JSON.parse(Buffer.from(options.body.get('assertion').split('.')[1], 'base64url'));
      assert.equal(claims.scope, 'https://www.googleapis.com/auth/chromewebstore');
      return Response.json({ access_token: 'test-token' });
    }
    assert.equal(options.headers.Authorization, 'Bearer test-token');
    assert.match(url, /publishers\/test-publisher\/items\/gagllhkjibjhllhkmiphjpahbgnmgana:/);
    if (url.endsWith(':fetchStatus')) return Response.json({});
    if (url.endsWith(':upload')) {
      assert.match(url, /\/upload\/v2\//);
      assert.equal(options.method, 'POST');
      assert.deepEqual(Buffer.from(await options.body.arrayBuffer()), zip);
      return Response.json({ uploadState: 'SUCCEEDED' });
    }
    assert.ok(url.endsWith(':publish'));
    assert.deepEqual(JSON.parse(options.body), { publishType: 'DEFAULT_PUBLISH' });
    return Response.json({ state: 'PENDING_REVIEW' });
  });
  assert.equal(await publishChrome(version, packagePath), 'Submitted for review');
  assert.equal(fetch.mock.callCount(), 4);
});

test('Chrome retries skip the same submitted version and preserve any other review', async t => {
  let submittedVersion = version;
  const fetch = t.mock.method(globalThis, 'fetch', async url => {
    if (url === 'https://oauth2.googleapis.com/token') return Response.json({ access_token: 'test-token' });
    assert.ok(url.endsWith(':fetchStatus'));
    return Response.json({
      submittedItemRevisionStatus: {
        state: 'PENDING_REVIEW', distributionChannels: [{ crxVersion: submittedVersion }],
      },
    });
  });
  assert.equal(await publishChrome(version, packagePath), 'Already submitted for review');
  submittedVersion = '2.2.5';
  await assert.rejects(publishChrome(version, packagePath), /reviewing another version/);
  assert.equal(fetch.mock.callCount(), 4);
});

test('Firefox submits the validated listed ZIP with its matching source in the same request', async t => {
  const tokens = new Set();
  const fetch = t.mock.method(globalThis, 'fetch', async (url, options) => {
    assert.match(options.headers.Authorization, /^JWT /);
    tokens.add(options.headers.Authorization);
    if (url.endsWith(`/versions/${version}/`)) return new Response(null, { status: 404 });
    assert.equal(options.method, 'POST');
    if (url.endsWith('/upload/')) {
      assert.equal(options.body.get('channel'), 'listed');
      assert.deepEqual(Buffer.from(await options.body.get('upload').arrayBuffer()), zip);
      return Response.json({ uuid: 'test-upload', processed: true, valid: true, version });
    }
    assert.ok(url.endsWith('/addon/snipo-code-snippet-manager/versions/'));
    assert.equal(options.body.get('upload'), 'test-upload');
    assert.deepEqual(Buffer.from(await options.body.get('source').arrayBuffer()), source);
    return Response.json({ version, channel: 'listed' }, { status: 201 });
  });
  assert.equal(await publishFirefox(version, packagePath, sourcePath), 'Submitted to Mozilla');
  assert.equal(fetch.mock.callCount(), 3);
  assert.equal(tokens.size, 3);
});

test('Firefox retries accept a listed version with source already in review', async t => {
  const fetch = t.mock.method(globalThis, 'fetch', async () => Response.json({
    channel: 'listed', source: '/review/source.zip', file: { status: 'unreviewed' },
  }));
  assert.equal(await publishFirefox(version, packagePath, sourcePath), 'Already submitted (unreviewed)');
  assert.equal(fetch.mock.callCount(), 1);
});
