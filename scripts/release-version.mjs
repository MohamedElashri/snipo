import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

export function validateVersion(version) {
  if (!/^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.test(version) ||
      version.split('.').some(n => Number(n) > 65535) || version === '0.0.0' || /\s/.test(version)) {
    throw new Error('Use X.Y.Z without a v prefix, suffix, or leading zeroes (components <= 65535).');
  }
  return version;
}

export function synchronize(root, version, check = false) {
  validateVersion(version);
  const changes = new Map();
  for (const [name, expression] of [
    ['internal/version/version.go', /(const Current = ")[^"]+(";?)/],
    ['tui/cmd/snippy/main.go', /(Version = ")[^"]+(";?)/],
  ]) {
    const original = fs.readFileSync(path.join(root, name), 'utf8');
    if (!expression.test(original)) throw new Error(`Missing version in ${name}`);
    changes.set(name, original.replace(expression, (_, before, after) => before + version + after));
  }
  for (const name of ['extension/package.json', 'extension/manifest.json',
    'extension/manifest-chrome.json', 'vscode-extension/package.json']) {
    const original = fs.readFileSync(path.join(root, name), 'utf8');
    const data = JSON.parse(original);
    if (typeof data.version !== 'string') throw new Error(`Missing version in ${name}`);
    changes.set(name, original.replace(/("version"\s*:\s*")[^"]+("\s*[,}])/, (_, before, after) => before + version + after));
  }
  const lockName = 'package-lock.json';
  const lock = JSON.parse(fs.readFileSync(path.join(root, lockName), 'utf8'));
  for (const workspace of ['extension', 'vscode-extension']) {
    if (!lock.packages?.[workspace]) throw new Error(`Missing ${workspace} workspace in lockfile`);
    lock.packages[workspace].version = version;
  }
  changes.set(lockName, JSON.stringify(lock, null, 2) + '\n');
  const mismatches = [...changes].filter(([name, content]) =>
    fs.readFileSync(path.join(root, name), 'utf8') !== content);
  if (check && mismatches.length) {
    throw new Error(`Version ${version} is not synchronized: ${mismatches.map(([name]) => name).join(', ')}`);
  }
  if (!check) for (const [name, content] of mismatches) fs.writeFileSync(path.join(root, name), content);
  return mismatches.map(([name]) => name);
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    const args = process.argv.slice(2);
    const check = args[0] === '--check';
    if (args.length !== (check ? 2 : 1)) throw new Error('Usage: bump-version.sh [--check] X.Y.Z');
    const version = args[check ? 1 : 0];
    const root = fileURLToPath(new URL('../', import.meta.url));
    const changed = synchronize(root, version, check);
    console.log(check ? `All product versions match ${version}.` : `Version ${version}: updated ${changed.length} files.`);
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}
