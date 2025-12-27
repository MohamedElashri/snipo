// Ace Editor utilities
import { theme } from '../modules/theme.js';

// Ace Editor language mode mapping
export function getAceMode(language) {
  const modeMap = {
    'javascript': 'ace/mode/javascript',
    'typescript': 'ace/mode/typescript',
    'python': 'ace/mode/python',
    'go': 'ace/mode/golang',
    'rust': 'ace/mode/rust',
    'java': 'ace/mode/java',
    'csharp': 'ace/mode/csharp',
    'cpp': 'ace/mode/c_cpp',
    'cuda': 'ace/mode/cuda',
    'ruby': 'ace/mode/ruby',
    'php': 'ace/mode/php',
    'swift': 'ace/mode/swift',
    'kotlin': 'ace/mode/kotlin',
    'html': 'ace/mode/html',
    'css': 'ace/mode/css',
    'sql': 'ace/mode/sql',
    'bash': 'ace/mode/sh',
    'powershell': 'ace/mode/powershell',
    'json': 'ace/mode/json',
    'yaml': 'ace/mode/yaml',
    'markdown': 'ace/mode/markdown',
    'tex': 'ace/mode/tex',
    'bibtex': 'ace/mode/tex',
    'plaintext': 'ace/mode/text'
  };
  return modeMap[language] || 'ace/mode/text';
}

// Get file extension for language
export function getFileExtension(language) {
  const extMap = {
    'javascript': 'js', 'typescript': 'ts', 'python': 'py', 'go': 'go',
    'rust': 'rs', 'java': 'java', 'csharp': 'cs', 'cpp': 'cpp', 'cuda': 'cu',
    'ruby': 'rb', 'php': 'php', 'swift': 'swift', 'kotlin': 'kt',
    'html': 'html', 'css': 'css', 'sql': 'sql', 'bash': 'sh',
    'powershell': 'ps1', 'json': 'json', 'yaml': 'yaml', 'markdown': 'md', 'tex': 'tex',
    'bibtex': 'bib', 'plaintext': 'txt'
  };
  return extMap[language] || 'txt';
}

// Detect language from filename
export function detectLanguageFromFilename(filename, options = {}) {
  const { isOnlyFile = false } = options;
  
  // List of common documentation filenames that should be ignored
  // unless they're the only file or explicitly requested
  const docFiles = [
    'README.md', 'readme.md', 'Readme.md',
    'LICENSE', 'LICENSE.md', 'license', 'license.md',
    'CHANGELOG.md', 'changelog.md', 'Changelog.md',
    'CONTRIBUTING.md', 'contributing.md',
    'CODE_OF_CONDUCT.md', 'code_of_conduct.md',
    'SECURITY.md', 'security.md'
  ];
  
  // If this is a documentation file and not the only file, return null
  // to prevent it from affecting the default language detection
  if (!isOnlyFile && docFiles.includes(filename)) {
    return null;
  }
  
  // Check for special filenames without extensions first
  const lowerFilename = filename.toLowerCase();
  const specialFiles = {
    'makefile': 'plaintext',
    'dockerfile': 'plaintext',
    'rakefile': 'ruby',
    'gemfile': 'ruby',
    'podfile': 'ruby'
  };
  
  if (specialFiles[lowerFilename]) {
    return specialFiles[lowerFilename];
  }
  
  // Get file extension
  const parts = filename.split('.');
  if (parts.length === 1) {
    // No extension found
    return null;
  }
  
  const ext = parts.pop()?.toLowerCase();
  
  const langMap = {
    // JavaScript/TypeScript
    'js': 'javascript', 'mjs': 'javascript', 'cjs': 'javascript', 'jsx': 'javascript',
    'ts': 'typescript', 'tsx': 'typescript',
    // Python
    'py': 'python', 'pyw': 'python', 'pyi': 'python',
    // Go
    'go': 'go',
    // Rust
    'rs': 'rust',
    // Java/Kotlin
    'java': 'java', 'kt': 'kotlin', 'kts': 'kotlin',
    // C/C++
    'c': 'cpp', 'h': 'cpp', 'cpp': 'cpp', 'cc': 'cpp', 'cxx': 'cpp',
    'hpp': 'cpp', 'hh': 'cpp', 'hxx': 'cpp',
    // C#
    'cs': 'csharp', 'csx': 'csharp',
    // CUDA
    'cu': 'cuda', 'cuh': 'cuda',
    // Ruby
    'rb': 'ruby', 'rake': 'ruby',
    // PHP
    'php': 'php', 'phtml': 'php', 'php3': 'php', 'php4': 'php', 'php5': 'php',
    // Swift
    'swift': 'swift',
    // Web
    'html': 'html', 'htm': 'html',
    'css': 'css', 'scss': 'css', 'sass': 'css', 'less': 'css',
    // SQL
    'sql': 'sql',
    // Shell/Bash
    'sh': 'bash', 'bash': 'bash', 'zsh': 'bash',
    // PowerShell
    'ps1': 'powershell', 'psm1': 'powershell', 'psd1': 'powershell',
    // Data formats
    'json': 'json', 'jsonc': 'json',
    'yaml': 'yaml', 'yml': 'yaml',
    'xml': 'html',
    // Documentation
    'md': 'markdown', 'markdown': 'markdown',
    'tex': 'tex',
    'bib': 'bibtex',
    // Plain text
    'txt': 'plaintext', 'text': 'plaintext'
  };
  
  return langMap[ext] || null;
}
