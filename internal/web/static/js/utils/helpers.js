// Utility helper functions

// Auto-resize input based on content
export function autoResizeInput(element) {
  if (!element) return;
  const val = element.value || element.placeholder || '';
  const length = val.length;
  element.style.width = Math.max(10, Math.ceil(length * 1.5)) + 'ch';
}

// Auto-resize select based on selected option
export function autoResizeSelect(element) {
  if (!element) return;
  const selectedOption = element.options[element.selectedIndex];
  const text = selectedOption ? selectedOption.text : element.value;
  const span = document.createElement('span');
  span.style.font = window.getComputedStyle(element).font;
  span.style.visibility = 'hidden';
  span.style.position = 'absolute';
  span.textContent = text;
  document.body.appendChild(span);
  const width = span.offsetWidth + 35;
  document.body.removeChild(span);
  element.style.width = width + 'px';
}

// Auto-resize textarea
export function autoResizeTextarea(el) {
  if (!el) return;
  el.style.height = 'auto';
  el.style.height = Math.min(el.scrollHeight, 80) + 'px';
}

// Format date for display
export function formatDate(dateStr) {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  const now = new Date();
  const diff = now - date;

  if (diff < 60000) return 'Just now';
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
  if (diff < 604800000) return `${Math.floor(diff / 86400000)}d ago`;

  return date.toLocaleDateString();
}

// Format file size
export function formatFileSize(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

// Get language color
export function getLanguageColor(lang) {
  const colors = {
    javascript: '#f7df1e',
    typescript: '#3178c6',
    python: '#3776ab',
    go: '#00add8',
    rust: '#dea584',
    java: '#b07219',
    csharp: '#178600',
    cpp: '#f34b7d',
    cuda: '#76b900',
    ruby: '#cc342d',
    php: '#4f5d95',
    swift: '#fa7343',
    kotlin: '#a97bff',
    html: '#e34c26',
    css: '#563d7c',
    sql: '#e38c00',
    bash: '#89e051',
    powershell: '#012456',
    json: '#292929',
    yaml: '#cb171e',
    markdown: '#083fa1',
    tex: '#3d6117',
    bibtex: '#778899',
    plaintext: '#6b7280'
  };
  return colors[lang] || '#6b7280';
}

// Highlight code blocks
export function highlightAll() {
  if (typeof Prism !== 'undefined') {
    Prism.highlightAll();
  }
}

// Highlight code manually for a specific language
export function highlightCode(code, language) {
  if (!code) return '';
  if (typeof Prism !== 'undefined' && Prism.languages[language]) {
    return Prism.highlight(code, Prism.languages[language], language);
  }
  // Fallback: escape HTML entities manually
  const div = document.createElement('div');
  div.textContent = code;
  return div.innerHTML;
}

// Render markdown
export function renderMarkdown(content) {
  if (!content) return '';
  if (typeof marked !== 'undefined') {
    marked.setOptions({
      breaks: true,
      gfm: true
    });
    return marked.parse(content);
  }
  return content;
}

// Expose helpers globally
window.autoResizeInput = autoResizeInput;
window.autoResizeSelect = autoResizeSelect;
window.autoResizeTextarea = autoResizeTextarea;
window.highlightCode = highlightCode;
