import { test } from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { validateVersion, synchronize } from './release-version.mjs';

test('reject versions incompatible with release tags or browser stores', () => {
  for (const version of ['v2.0.0', '02.0.0', '2.0', '2.0.0-rc.1', '2.0.0+build',
    '65536.0.0', '0.0.0', '2.0.0\n', '2.0.0\r', '2.0.0\u2028', '2.0.0; echo bad']) {
    assert.throws(() => validateVersion(version), version);
  }
  assert.equal(validateVersion('2.0.0'), '2.0.0');
});

test('version changes are complete, checked without writes, and idempotent', () => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'snipo-version-'));
  try {
    const files = {
      'internal/version/version.go': 'const Current = "1.0.0"\n',
      'tui/cmd/snippy/main.go': 'Version = "0.1.0"\n',
      'extension/package.json': '{"version":"1.0.0"}\n',
      'extension/manifest.json': '{"version":"1.0.0"}\n',
      'extension/manifest-chrome.json': '{"version":"1.0.0"}\n',
      'vscode-extension/package.json': '{"version":"0.0.3","other":"keep"}\n',
      'package-lock.json': JSON.stringify({ packages: {
        extension: { version: '1.0.0' }, 'vscode-extension': { version: '0.0.3' },
      } }),
    };
    for (const [name, content] of Object.entries(files)) {
      fs.mkdirSync(path.dirname(path.join(root, name)), { recursive: true });
      fs.writeFileSync(path.join(root, name), content);
    }
    assert.throws(() => synchronize(root, '2.0.0', true));
    for (const [name, content] of Object.entries(files)) assert.equal(fs.readFileSync(path.join(root, name), 'utf8'), content);
    assert.equal(synchronize(root, '2.0.0').length, 7);
    assert.deepEqual(synchronize(root, '2.0.0'), []);
    assert.deepEqual(synchronize(root, '2.0.0', true), []);
    assert.equal(JSON.parse(fs.readFileSync(path.join(root, 'vscode-extension/package.json'))).other, 'keep');
    fs.unlinkSync(path.join(root, 'extension/manifest.json'));
    assert.throws(() => synchronize(root, '2.0.1'));
    assert.match(fs.readFileSync(path.join(root, 'internal/version/version.go'), 'utf8'), /2\.0\.0/);
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
});
