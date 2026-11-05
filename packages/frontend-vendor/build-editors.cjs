const path = require('node:path');
const fs = require('node:fs');
const esbuild = require('esbuild');
const output = path.resolve(__dirname, '../../internal/web/static/vendor/editors');
fs.rmSync(output, { recursive: true, force: true });
const result = esbuild.buildSync({
  entryPoints: { codemirror: path.join(__dirname, 'editors/codemirror.js'), monaco: path.join(__dirname, 'editors/monaco.js'),
    'monaco-worker': require.resolve('monaco-editor/editor/editor.worker.js'),
    'monaco-typescript-worker': require.resolve('monaco-editor/language/typescript/ts.worker.js'),
    'monaco-json-worker': require.resolve('monaco-editor/language/json/json.worker.js'),
    'monaco-html-worker': require.resolve('monaco-editor/language/html/html.worker.js'),
    'monaco-css-worker': require.resolve('monaco-editor/language/css/css.worker.js') },
  outdir: output, bundle: true, format: 'esm', splitting: true, minify: true,
  metafile: true, target: 'es2022', loader: { '.ttf': 'file' }, assetNames: '[name]-[hash]',
  chunkNames: 'chunks/[name]-[hash]', legalComments: 'linked', logLevel: 'warning'
});

// Ship the full license text of every package included in these bundles.
const packages = new Set();
for (const input of Object.keys(result.metafile.inputs)) {
  let directory = path.dirname(path.resolve(input));
  while (directory.includes('node_modules')) {
    if (fs.existsSync(path.join(directory, 'package.json'))) { packages.add(directory); break; }
    directory = path.dirname(directory);
  }
}
const licenses = [...packages].sort().map(directory => {
  const metadata = JSON.parse(fs.readFileSync(path.join(directory, 'package.json'), 'utf8'));
  const files = fs.readdirSync(directory).filter(file => /^licen[sc]e/i.test(file));
  return `${metadata.name} ${metadata.version}\n${files.map(file => fs.readFileSync(path.join(directory, file), 'utf8')).join('\n')}`;
});
fs.writeFileSync(path.join(output, 'LICENSES.txt'), licenses.join('\n\n---\n\n'));
