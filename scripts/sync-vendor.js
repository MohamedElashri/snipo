#!/usr/bin/env node

const fs = require('fs-extra');
const path = require('path');

const VENDOR_DIR = path.join(__dirname, '../internal/web/static/vendor');
const NODE_MODULES = path.join(__dirname, '../node_modules');

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
    // cuda is custom mode - not available in prismjs
    // 'prism-cuda.min.js': 'prismjs/components/prism-cuda.min.js',
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
    'ext-emmet.js': 'ace-builds/src-min-noconflict/ext-emmet.js',
    'ext-beautify.js': 'ace-builds/src-min-noconflict/ext-beautify.js',
    
    // Snippets (for autocomplete and Emmet)
    'snippets/javascript.js': 'ace-builds/src-min-noconflict/snippets/javascript.js',
    'snippets/typescript.js': 'ace-builds/src-min-noconflict/snippets/typescript.js',
    'snippets/python.js': 'ace-builds/src-min-noconflict/snippets/python.js',
    'snippets/golang.js': 'ace-builds/src-min-noconflict/snippets/golang.js',
    'snippets/rust.js': 'ace-builds/src-min-noconflict/snippets/rust.js',
    'snippets/java.js': 'ace-builds/src-min-noconflict/snippets/java.js',
    'snippets/csharp.js': 'ace-builds/src-min-noconflict/snippets/csharp.js',
    'snippets/c_cpp.js': 'ace-builds/src-min-noconflict/snippets/c_cpp.js',
    'snippets/ruby.js': 'ace-builds/src-min-noconflict/snippets/ruby.js',
    'snippets/php.js': 'ace-builds/src-min-noconflict/snippets/php.js',
    'snippets/swift.js': 'ace-builds/src-min-noconflict/snippets/swift.js',
    'snippets/kotlin.js': 'ace-builds/src-min-noconflict/snippets/kotlin.js',
    'snippets/html.js': 'ace-builds/src-min-noconflict/snippets/html.js',
    'snippets/css.js': 'ace-builds/src-min-noconflict/snippets/css.js',
    'snippets/sql.js': 'ace-builds/src-min-noconflict/snippets/sql.js',
    'snippets/sh.js': 'ace-builds/src-min-noconflict/snippets/sh.js',
    'snippets/powershell.js': 'ace-builds/src-min-noconflict/snippets/powershell.js',
    'snippets/json.js': 'ace-builds/src-min-noconflict/snippets/json.js',
    'snippets/yaml.js': 'ace-builds/src-min-noconflict/snippets/yaml.js',
    'snippets/markdown.js': 'ace-builds/src-min-noconflict/snippets/markdown.js',
    'snippets/tex.js': 'ace-builds/src-min-noconflict/snippets/tex.js',
    'snippets/text.js': 'ace-builds/src-min-noconflict/snippets/text.js',
    // cuda is custom mode - not available in ace-builds
    // 'mode-cuda.js': 'ace-builds/src-min-noconflict/mode-cuda.js',
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
  }
};

async function syncVendor() {
  console.log('Syncing vendor files from node_modules...\n');
  
  let successCount = 0;
  let errorCount = 0;
  const errors = [];

  for (const [subdir, files] of Object.entries(vendorConfig)) {
    const targetDir = path.join(VENDOR_DIR, subdir);
    
    await fs.ensureDir(targetDir);
    
    for (const [destFile, sourcePath] of Object.entries(files)) {
      const source = path.join(NODE_MODULES, sourcePath);
      const dest = path.join(targetDir, destFile);
      
      try {
        if (!await fs.pathExists(source)) {
          throw new Error(`Source file not found: ${sourcePath}`);
        }
        
        await fs.copy(source, dest, { overwrite: true });
        console.log(`${subdir}/${destFile}`);
        successCount++;
      } catch (error) {
        console.error(`❌ ${subdir}/${destFile}: ${error.message}`);
        errors.push({ file: `${subdir}/${destFile}`, error: error.message });
        errorCount++;
      }
    }
  }
  
  console.log(`\nSummary:`);
  console.log(`   Success: ${successCount} files`);
  console.log(`   Errors: ${errorCount} files`);
  
  if (errors.length > 0) {
    console.log('\n  Failed files:');
    errors.forEach(({ file, error }) => {
      console.log(`   - ${file}: ${error}`);
    });
    process.exit(1);
  }
  
  console.log('\n Vendor sync completed successfully!');
}

if (require.main === module) {
  syncVendor().catch(error => {
    console.error(' Sync failed:', error);
    process.exit(1);
  });
}

module.exports = { syncVendor };
