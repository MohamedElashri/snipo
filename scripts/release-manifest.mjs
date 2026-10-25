import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { validateVersion } from './release-version.mjs';

const [directory, version, commit] = process.argv.slice(2);
validateVersion(version);
if (!/^[a-f0-9]{40}$/.test(commit)) throw new Error('Expected a full commit SHA');
const expected = [
  ...['amd64', 'arm64'].map(arch => `snipo_${version}_linux_${arch}.tar.gz`),
  ...['linux', 'darwin', 'windows'].flatMap(os => ['amd64', 'arm64'].map(arch =>
    `snippy_${version}_${os}_${arch}.${os === 'windows' ? 'zip' : 'tar.gz'}`)),
  ...['chrome', 'firefox', 'source'].map(name => `snipo-${name}-v${version}.zip`),
  `snipo-${version}.vsix`,
];
for (const name of expected) {
  if (!fs.statSync(path.join(directory, name)).isFile()) throw new Error(`Missing release asset: ${name}`);
}
for (const arch of ['amd64', 'arm64']) for (const format of ['deb', 'rpm']) {
  if (!fs.readdirSync(directory).some(name => name.startsWith('snippy_') && name.includes(arch) && name.endsWith(`.${format}`))) {
    throw new Error(`Missing Snippy ${arch} ${format}`);
  }
}
const assets = fs.readdirSync(directory).filter(name => /\.(tar\.gz|zip|vsix|deb|rpm|json)$/.test(name) && name !== 'release-manifest.json')
  .sort().map(name => ({ name, sha256: crypto.createHash('sha256').update(fs.readFileSync(path.join(directory, name))).digest('hex') }));
fs.writeFileSync(path.join(directory, 'release-manifest.json'), JSON.stringify({
  tag: `v${version}`, version, commit, assets,
  stores: { vscode: 'pending', chrome: 'pending', firefox: 'pending' },
}, null, 2) + '\n');
const names = [...assets.map(a => a.name), 'release-manifest.json'];
fs.writeFileSync(path.join(directory, 'checksums.txt'), names.map(name =>
  `${crypto.createHash('sha256').update(fs.readFileSync(path.join(directory, name))).digest('hex')}  ${name}`).join('\n') + '\n');
