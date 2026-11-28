import { createHmac, randomUUID, sign } from 'node:crypto';
import { openAsBlob } from 'node:fs';
import { setTimeout as delay } from 'node:timers/promises';

const chromeAPI = 'https://chromewebstore.googleapis.com';
const googleTokenURL = 'https://oauth2.googleapis.com/token';
const amoAPI = 'https://addons.mozilla.org/api/v5/addons';

function requiredEnv(name) {
  if (!process.env[name]) throw new Error(`Configure ${name} for this publishing job.`);
  return process.env[name];
}

function encodeJSON(value) {
  return Buffer.from(JSON.stringify(value)).toString('base64url');
}

async function storeRequest(url, options) {
  return fetch(url, { redirect: 'error', signal: AbortSignal.timeout(60_000), ...options });
}

async function requireJSON(response, operation) {
  if (!response.ok) throw new Error(`${operation} failed (HTTP ${response.status}).`);
  return response.json();
}

async function chromeToken() {
  const credentials = requiredEnv('CHROME_SERVICE_ACCOUNT_JSON');
  let account;
  try {
    account = JSON.parse(credentials);
  } catch {
    throw new Error('CHROME_SERVICE_ACCOUNT_JSON must contain a Google service account key.');
  }
  if (!account?.client_email || !account?.private_key) {
    throw new Error('The Chrome service account key is missing client_email or private_key.');
  }
  const now = Math.floor(Date.now() / 1000);
  const claims = encodeJSON({
    iss: account.client_email, aud: googleTokenURL, iat: now, exp: now + 3600,
    scope: 'https://www.googleapis.com/auth/chromewebstore',
  });
  const unsigned = `${encodeJSON({ alg: 'RS256', typ: 'JWT' })}.${claims}`;
  const assertion = `${unsigned}.${sign('RSA-SHA256', Buffer.from(unsigned), account.private_key).toString('base64url')}`;
  const token = await requireJSON(await storeRequest(googleTokenURL, {
    method: 'POST',
    body: new URLSearchParams({ grant_type: 'urn:ietf:params:oauth:grant-type:jwt-bearer', assertion }),
  }), 'Chrome authentication');
  if (!token.access_token) throw new Error('Google did not return an access token.');
  return token.access_token;
}

export async function publishChrome(version, packagePath) {
  const publisher = encodeURIComponent(requiredEnv('CHROME_PUBLISHER_ID'));
  const item = 'gagllhkjibjhllhkmiphjpahbgnmgana';
  const resource = `/v2/publishers/${publisher}/items/${item}`;
  const headers = { Authorization: `Bearer ${await chromeToken()}` };
  const status = async () => requireJSON(await storeRequest(`${chromeAPI}${resource}:fetchStatus`, { headers }), 'Chrome status');
  const current = await status();
  const hasVersion = revision => revision?.distributionChannels?.some(channel => channel.crxVersion === version);
  if (hasVersion(current.publishedItemRevisionStatus)) return 'Already published';
  const submitted = current.submittedItemRevisionStatus;
  if (submitted?.state === 'PENDING_REVIEW') {
    if (hasVersion(submitted)) return 'Already submitted for review';
    throw new Error('Chrome is reviewing another version. Retry after that review finishes.');
  }
  if (submitted?.state === 'STAGED') {
    throw new Error('Chrome has a staged version. Resolve it in the Developer Dashboard before retrying.');
  }

  const upload = await requireJSON(await storeRequest(`${chromeAPI}/upload${resource}:upload`, {
    method: 'POST', headers, body: await openAsBlob(packagePath, { type: 'application/zip' }),
  }), 'Chrome upload');
  let state = upload.uploadState;
  for (let attempt = 0; state === 'IN_PROGRESS' && attempt < 24; attempt++) {
    await delay(5000);
    state = (await status()).lastAsyncUploadState;
  }
  if (state !== 'SUCCEEDED') throw new Error(`Chrome upload status: ${state || 'unknown'}.`);

  const submission = await requireJSON(await storeRequest(`${chromeAPI}${resource}:publish`, {
    method: 'POST',
    headers: { ...headers, 'Content-Type': 'application/json' },
    body: JSON.stringify({ publishType: 'DEFAULT_PUBLISH' }),
  }), 'Chrome submission');
  switch (submission.state) {
    case 'PENDING_REVIEW': return 'Submitted for review';
    case 'PUBLISHED':
    case 'PUBLISHED_TO_TESTERS': return `Published (${submission.state})`;
    default: throw new Error(`Unexpected Chrome submission status: ${submission.state || 'unknown'}.`);
  }
}

export async function publishFirefox(version, packagePath, sourcePath) {
  const issuer = requiredEnv('AMO_JWT_ISSUER');
  const secret = requiredEnv('AMO_JWT_SECRET');
  const versionsURL = `${amoAPI}/addon/snipo-code-snippet-manager/versions/`;
  const amoRequest = (url, method = 'GET', body) => {
    const now = Math.floor(Date.now() / 1000);
    const claims = encodeJSON({ iss: issuer, jti: randomUUID(), iat: now, exp: now + 120 });
    const unsigned = `${encodeJSON({ alg: 'HS256', typ: 'JWT' })}.${claims}`;
    const signature = createHmac('sha256', secret).update(unsigned).digest('base64url');
    return storeRequest(url, { method, body, headers: { Authorization: `JWT ${unsigned}.${signature}` } });
  };

  const existing = await amoRequest(`${versionsURL}${version}/`);
  if (existing.status !== 404) {
    const data = await requireJSON(existing, 'Firefox version lookup');
    if (data.channel === 'listed' && data.source && ['public', 'unreviewed'].includes(data.file?.status)) {
      return `Already submitted (${data.file.status})`;
    }
    throw new Error('Firefox already has this version. Check its review and source status in the developer dashboard.');
  }

  const packageForm = new FormData();
  packageForm.set('channel', 'listed');
  packageForm.set('upload', await openAsBlob(packagePath, { type: 'application/zip' }), `snipo-firefox-v${version}.zip`);
  let upload = await requireJSON(await amoRequest(`${amoAPI}/upload/`, 'POST', packageForm), 'Firefox upload');
  const uuid = upload.uuid;
  if (!uuid) throw new Error('Firefox did not return an upload ID.');
  for (let attempt = 0; !upload.processed && attempt < 24; attempt++) {
    await delay(5000);
    upload = await requireJSON(await amoRequest(`${amoAPI}/upload/${encodeURIComponent(uuid)}/`), 'Firefox validation');
  }
  if (!upload.processed) throw new Error('Firefox validation timed out.');
  if (!upload.valid || upload.version !== version) throw new Error('Firefox validation failed or returned the wrong version.');

  const versionForm = new FormData();
  versionForm.set('upload', uuid);
  versionForm.set('source', await openAsBlob(sourcePath, { type: 'application/zip' }), `snipo-source-v${version}.zip`);
  const submission = await requireJSON(await amoRequest(versionsURL, 'POST', versionForm), 'Firefox submission');
  if (submission.version !== version || submission.channel !== 'listed') {
    throw new Error('Firefox returned an unexpected version or channel.');
  }
  return 'Submitted to Mozilla';
}

if (import.meta.main) {
  const [store, version, packagePath, sourcePath] = process.argv.slice(2);
  if (!/^\d+\.\d+\.\d+$/.test(version) || !packagePath || !['chrome', 'firefox'].includes(store) ||
      (store === 'firefox' && !sourcePath)) {
    throw new Error('Usage: node scripts/publish-browser-extension.mjs chrome|firefox VERSION PACKAGE [FIREFOX_SOURCE]');
  }
  const result = store === 'chrome'
    ? await publishChrome(version, packagePath)
    : await publishFirefox(version, packagePath, sourcePath);
  console.log(result);
}
