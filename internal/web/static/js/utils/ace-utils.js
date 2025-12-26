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
    'json': 'json', 'yaml': 'yaml', 'markdown': 'md', 'tex': 'tex',
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
  
  const ext = filename.split('.').pop()?.toLowerCase();
  const langMap = {
    'js': 'javascript', 'ts': 'typescript', 'py': 'python', 'go': 'go',
    'rs': 'rust', 'java': 'java', 'cs': 'csharp', 'cpp': 'cpp', 'c': 'cpp',
    'cu': 'cuda', 'cuh': 'cuda',
    'rb': 'ruby', 'php': 'php', 'swift': 'swift', 'kt': 'kotlin',
    'html': 'html', 'css': 'css', 'sql': 'sql', 'sh': 'bash',
    'json': 'json', 'yaml': 'yaml', 'yml': 'yaml', 'md': 'markdown',
    'tex': 'tex', 'bib': 'bibtex', 'txt': 'plaintext'
  };
  return langMap[ext] || null;
}
