#!/usr/bin/env node

/**
 * verify-vendor.js
 *
 * Usage:
 *   node scripts/verify-vendor.js           — Verify all expected files exist
 *   node scripts/verify-vendor.js --cleanup — Remove orphaned files
 *   node scripts/verify-vendor.js --status  — Show current versions + update status
 */

const fs = require('fs-extra');
const path = require('path');

const VENDOR_DIR = path.join(__dirname, '../internal/web/static/vendor');
const NODE_MODULES = path.join(__dirname, '../node_modules');
const PACKAGE_JSON = path.join(__dirname, '../package.json');

// Re-use the same config from sync-vendor.js
const { syncVendor, vendorConfig } = require('./sync-vendor.js');

// Build flat list of expected files with their vendor subdir
function getExpectedFiles() {
  const files = [];
  for (const [subdir, mapping] of Object.entries(vendorConfig)) {
    for (const destFile of Object.keys(mapping)) {
      files.push(path.join(subdir, destFile));
    }
  }
  // Also include files NOT managed by sync-vendor.js (custom/manual)
  const manualFiles = ['fonts/FiraCode-Bold.woff2', 'fonts/FiraCode-Medium.woff2', 'fonts/FiraCode-Regular.woff2', 'css/fira_code.css', 'css/fonts.css'];
  for (const f of manualFiles) {
    files.push(f);
  }
  return files;
}

function getActualFiles() {
  const files = [];
  function walk(dir, prefix) {
    if (!fs.existsSync(dir)) return;
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const entry of entries) {
      const full = path.join(dir, entry.name);
      const rel = prefix ? path.join(prefix, entry.name) : entry.name;
      if (entry.isDirectory()) {
        walk(full, rel);
      } else {
        files.push(rel);
      }
    }
  }
  walk(VENDOR_DIR, '');
  return files;
}

// ── Verify mode ──────────────────────────────────────────────────────
function verify() {
  const expected = getExpectedFiles();
  const missing = [];

  for (const f of expected) {
    if (!fs.existsSync(path.join(VENDOR_DIR, f))) {
      missing.push(f);
    }
  }

  if (missing.length > 0) {
    console.error(`❌ ${missing.length} vendor file(s) missing:`);
    missing.forEach(f => console.error(`   - ${f}`));
    console.error('\n  Run: make vendor-sync');
    process.exit(1);
  }
}

// ── Cleanup mode ─────────────────────────────────────────────────────
function cleanup() {
  const expected = new Set(getExpectedFiles());
  const actual = getActualFiles();
  const orphaned = actual.filter(f => !expected.has(f));

  if (orphaned.length === 0) {
    console.log('  No orphaned files found.');
    return;
  }

  for (const f of orphaned) {
    const fullPath = path.join(VENDOR_DIR, f);
    fs.removeSync(fullPath);
    console.log(`  Removed: ${f}`);
  }

  // Remove empty directories
  function pruneEmpty(dir) {
    if (!fs.existsSync(dir)) return;
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const entry of entries) {
      if (entry.isDirectory()) {
        pruneEmpty(path.join(dir, entry.name));
      }
    }
    const remaining = fs.readdirSync(dir);
    if (remaining.length === 0) {
      fs.removeSync(dir);
    }
  }
  pruneEmpty(VENDOR_DIR);

  console.log(`  Cleaned up ${orphaned.length} orphaned file(s).`);
}

// ── Status mode ──────────────────────────────────────────────────────
function status() {
  const pkg = JSON.parse(fs.readFileSync(PACKAGE_JSON, 'utf8'));
  const deps = { ...pkg.dependencies };

  const vendorPkgs = [
    { name: 'htmx.org', pkgName: 'htmx.org' },
    { name: 'alpinejs', pkgName: 'alpinejs' },
    { name: '@alpinejs/collapse', pkgName: '@alpinejs/collapse' },
    { name: 'ace-builds', pkgName: 'ace-builds' },
    { name: 'prismjs', pkgName: 'prismjs' },
    { name: 'marked', pkgName: 'marked' },
    { name: '@picocss/pico', pkgName: '@picocss/pico' },
  ];

  const pad = (str, n) => String(str).padEnd(n, ' ');

  console.log('  %s  %s  %s  %s', pad('Package', 25), pad('Installed', 12), pad('Wanted', 12), pad('Latest', 12));
  console.log('  %s  %s  %s  %s', '─'.repeat(25), '─'.repeat(12), '─'.repeat(12), '─'.repeat(12));

  for (const v of vendorPkgs) {
    const pkgDir = path.join(NODE_MODULES, v.pkgName);
    let installed = '—';
    if (fs.existsSync(path.join(pkgDir, 'package.json'))) {
      try {
        const pkgMeta = JSON.parse(fs.readFileSync(path.join(pkgDir, 'package.json'), 'utf8'));
        installed = pkgMeta.version;
      } catch {
        installed = 'error';
      }
    }

    const wanted = deps[v.pkgName] || '—';
    console.log('  %s  %s  %s  %s', pad(v.pkgName, 25), pad(installed, 12), pad(wanted, 12), '');
  }
}

// ── Main ─────────────────────────────────────────────────────────────
const mode = process.argv[2] || '';

if (!fs.existsSync(NODE_MODULES)) {
  console.error('node_modules/ not found. Run: make vendor-install');
  process.exit(1);
}

if (mode === '--cleanup') {
  cleanup();
} else if (mode === '--status') {
  status();
} else {
  verify();
}
