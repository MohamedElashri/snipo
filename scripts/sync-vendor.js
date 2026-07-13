#!/usr/bin/env node
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const VENDOR_DIR = path.join(__dirname, '..', 'internal', 'web', 'static', 'vendor');
const NODE_MODULES = path.join(__dirname, '..', 'node_modules');
const PACKAGE_JSON = path.join(__dirname, '..', 'package.json');
const PATCHES_DIR = path.join(__dirname, '..', 'patches');

// Vendor patches to apply after syncing (array of {file, patch})

const vendorConfig = {
  js: {
    'htmx.min.js': 'htmx.org/dist/htmx.min.js',
    'alpine.min.js': 'alpinejs/dist/cdn.min.js',
    'alpine-collapse.min.js': '@alpinejs/collapse/dist/cdn.min.js',
    'marked.min.js': 'marked/lib/marked.umd.js',
    'prism.min.js': 'prismjs/prism.js',
    'prism-bash.min.js': 'prismjs/components/prism-bash.min.js',
    'prism-powershell.min.js': 'prismjs/components/prism-powershell.min.js',
    'prism-python.min.js': 'prismjs/components/prism-python.min.js',
    'prism-javascript.min.js': 'prismjs/components/prism-javascript.min.js',
    'prism-go.min.js': 'prismjs/components/prism-go.min.js',
    'prism-json.min.js': 'prismjs/components/prism-json.min.js',
    'prism-yaml.min.js': 'prismjs/components/prism-yaml.min.js',
    'prism-sql.min.js': 'prismjs/components/prism-sql.min.js',
    'prism-markdown.min.js': 'prismjs/components/prism-markdown.min.js',
  },
  css: {
    'pico.min.css': '@picocss/pico/css/pico.min.css',
    'prism.min.css': 'prismjs/themes/prism.min.css',
    'prism-tomorrow.min.css': 'prismjs/themes/prism-tomorrow.min.css',
  },
  'js/ace': {
    'ace.js': 'ace-builds/src-min-noconflict/ace.js',
    'ext-language_tools.js': 'ace-builds/src-min-noconflict/ext-language_tools.js',
    'ext-searchbox.js': 'ace-builds/src-min-noconflict/ext-searchbox.js',
    'mode-c_cpp.js': 'ace-builds/src-min-noconflict/mode-c_cpp.js',
    'mode-csharp.js': 'ace-builds/src-min-noconflict/mode-csharp.js',
    'mode-css.js': 'ace-builds/src-min-noconflict/mode-css.js',
    'mode-golang.js': 'ace-builds/src-min-noconflict/mode-golang.js',
    'mode-html.js': 'ace-builds/src-min-noconflict/mode-html.js',
    'mode-java.js': 'ace-builds/src-min-noconflict/mode-java.js',
    'mode-javascript.js': 'ace-builds/src-min-noconflict/mode-javascript.js',
    'mode-json.js': 'ace-builds/src-min-noconflict/mode-json.js',
    'mode-kotlin.js': 'ace-builds/src-min-noconflict/mode-kotlin.js',
    'mode-markdown.js': 'ace-builds/src-min-noconflict/mode-markdown.js',
    'mode-php.js': 'ace-builds/src-min-noconflict/mode-php.js',
    'mode-powershell.js': 'ace-builds/src-min-noconflict/mode-powershell.js',
    'mode-python.js': 'ace-builds/src-min-noconflict/mode-python.js',
    'mode-ruby.js': 'ace-builds/src-min-noconflict/mode-ruby.js',
    'mode-rust.js': 'ace-builds/src-min-noconflict/mode-rust.js',
    'mode-sh.js': 'ace-builds/src-min-noconflict/mode-sh.js',
    'mode-sql.js': 'ace-builds/src-min-noconflict/mode-sql.js',
    'mode-swift.js': 'ace-builds/src-min-noconflict/mode-swift.js',
    'mode-tex.js': 'ace-builds/src-min-noconflict/mode-tex.js',
    'mode-text.js': 'ace-builds/src-min-noconflict/mode-text.js',
    'mode-typescript.js': 'ace-builds/src-min-noconflict/mode-typescript.js',
    'mode-yaml.js': 'ace-builds/src-min-noconflict/mode-yaml.js',
    'theme-ambiance.js': 'ace-builds/src-min-noconflict/theme-ambiance.js',
    'theme-chaos.js': 'ace-builds/src-min-noconflict/theme-chaos.js',
    'theme-chrome.js': 'ace-builds/src-min-noconflict/theme-chrome.js',
    'theme-clouds.js': 'ace-builds/src-min-noconflict/theme-clouds.js',
    'theme-cobalt.js': 'ace-builds/src-min-noconflict/theme-cobalt.js',
    'theme-dracula.js': 'ace-builds/src-min-noconflict/theme-dracula.js',
    'theme-github.js': 'ace-builds/src-min-noconflict/theme-github.js',
    'theme-kuroir.js': 'ace-builds/src-min-noconflict/theme-kuroir.js',
    'theme-monokai.js': 'ace-builds/src-min-noconflict/theme-monokai.js',
    'theme-textmate.js': 'ace-builds/src-min-noconflict/theme-textmate.js',
    'theme-twilight.js': 'ace-builds/src-min-noconflict/theme-twilight.js',
    'theme-xcode.js': 'ace-builds/src-min-noconflict/theme-xcode.js',
    'worker-css.js': 'ace-builds/src-min-noconflict/worker-css.js',
    'worker-html.js': 'ace-builds/src-min-noconflict/worker-html.js',
    'worker-javascript.js': 'ace-builds/src-min-noconflict/worker-javascript.js',
    'worker-json.js': 'ace-builds/src-min-noconflict/worker-json.js',
    'worker-php.js': 'ace-builds/src-min-noconflict/worker-php.js',
    'worker-xml.js': 'ace-builds/src-min-noconflict/worker-xml.js',
  },
};

const manualFiles = [
  'fonts/FiraCode-Bold.woff2',
  'fonts/FiraCode-Medium.woff2',
  'fonts/FiraCode-Regular.woff2',
  'css/fira_code.css',
  'css/fonts.css',
  'js/prism-cuda.min.js',
  'js/ace/mode-cuda.js',
];

// Vendor patches applied after sync (file paths relative to VENDOR_DIR)
const vendorPatches = [
  {
    file: 'js/prism.min.js',
    patch: 'prismjs-worker-origin-check.patch',
  },
];

// Helper: collect all expected vendor-relative paths
function expectedPaths() {
  const paths = [];
  for (const [subdir, files] of Object.entries(vendorConfig)) {
    for (const destFile of Object.keys(files)) {
      paths.push(path.join(subdir, destFile));
    }
  }
  return paths.concat(manualFiles);
}

// ── Sync ────────────────────────────────────────────────────────────────
function sync() {
  let ok = true;
  for (const [subdir, files] of Object.entries(vendorConfig)) {
    const targetDir = path.join(VENDOR_DIR, subdir);
    fs.mkdirSync(targetDir, { recursive: true });
    for (const [destFile, sourceRel] of Object.entries(files)) {
      const src = path.join(NODE_MODULES, sourceRel);
      const dst = path.join(targetDir, destFile);
      try {
        if (!fs.existsSync(src)) {
          console.error(`  MISSING  ${subdir}/${destFile} (source: ${sourceRel})`);
          ok = false;
          continue;
        }
        fs.copyFileSync(src, dst);
        console.log(`  OK       ${subdir}/${destFile}`);
      } catch (err) {
        console.error(`  ERROR    ${subdir}/${destFile}: ${err.message}`);
        ok = false;
      }
    }
  }

  // Apply vendor patches
  for (const { file, patch } of vendorPatches) {
    const target = path.join(VENDOR_DIR, file);
    const patchFile = path.join(PATCHES_DIR, patch);
    if (fs.existsSync(patchFile) && fs.existsSync(target)) {
      try {
        execSync(`patch --forward "${target}" "${patchFile}"`, { stdio: 'ignore' });
        console.log(`  PATCHED  ${file}`);
      } catch {
        // patch might already be applied; ignore
      }
    }
  }

  return ok;
}

// ── Cleanup ─────────────────────────────────────────────────────────────
function cleanup() {
  const expected = new Set(expectedPaths());
  let removed = 0;
  function walk(dir, prefix) {
    if (!fs.existsSync(dir)) return;
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
      const full = path.join(dir, entry.name);
      const rel = prefix ? path.join(prefix, entry.name) : entry.name;
      if (entry.isDirectory()) {
        walk(full, rel);
        if (fs.readdirSync(full).length === 0) {
          fs.rmdirSync(full);
          console.log(`  REMOVED  ${rel}/`);
        }
      } else if (!expected.has(rel)) {
        fs.unlinkSync(full);
        console.log(`  REMOVED  ${rel}`);
        removed++;
      }
    }
  }
  walk(VENDOR_DIR, '');
  if (removed === 0) console.log('  No orphaned files found.');
  return true;
}

// ── Status ──────────────────────────────────────────────────────────────
function status() {
  const pkg = JSON.parse(fs.readFileSync(PACKAGE_JSON, 'utf8'));
  const vendorPkgs = Object.keys(pkg.dependencies);
  const pad = (s, n) => String(s).padEnd(n, ' ');

  console.log(`  ${pad('Package', 25)} ${pad('Installed', 12)} ${pad('Required', 12)}`);
  console.log(`  ${'─'.repeat(25)} ${'─'.repeat(12)} ${'─'.repeat(12)}`);

  for (const name of vendorPkgs) {
    const pkgDir = path.join(NODE_MODULES, name);
    let installed = '—';
    if (fs.existsSync(path.join(pkgDir, 'package.json'))) {
      try {
        const meta = JSON.parse(fs.readFileSync(path.join(pkgDir, 'package.json'), 'utf8'));
        installed = meta.version;
      } catch { installed = 'error'; }
    }
    console.log(`  ${pad(name, 25)} ${pad(installed, 12)} ${pad(pkg.dependencies[name], 12)}`);
  }
  return true;
}

// ── Main ────────────────────────────────────────────────────────────────
const mode = process.argv[2] || '';

if (!fs.existsSync(NODE_MODULES) && mode !== '--status') {
  console.error('node_modules/ not found. Run: npm install');
  process.exit(1);
}

let success;
switch (mode) {
  case '--cleanup':
    success = cleanup();
    break;
  case '--status':
    success = status();
    break;
  default:
    console.log('Syncing vendor files...\n');
    success = sync();
    if (!success) {
      console.error('\n  Some vendor files are missing or errored. Run: npm install');
      process.exit(1);
    }
    console.log('\n  Vendor sync complete.');
}
