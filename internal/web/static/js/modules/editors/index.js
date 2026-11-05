// Adapters own library instances outside Alpine's reactive proxy. Application code
// only uses getValue/setValue/setLanguage/configure/focus/resize/destroy.
export const editorEngines = ['ace', 'codemirror', 'monaco'];
export function readEditorPreference() {
  try {
    const saved = localStorage.getItem('snipo-editor');
    return editorEngines.includes(saved) ? saved : 'ace';
  } catch { return 'ace'; }
}
// A stalled asset request must not leave the editor covered by a spinner forever.
const loadTimeout = 20000;
function withTimeout(promise) {
  let timer;
  return Promise.race([promise, new Promise((_, reject) => {
    timer = setTimeout(() => reject(new Error('Editor loading timed out')), loadTimeout);
  })]).finally(() => clearTimeout(timer));
}
let monacoStyles;
function loadMonacoStyles() {
  if (!monacoStyles) monacoStyles = new Promise((resolve, reject) => {
    const link = document.createElement('link');
    link.id = 'monaco-styles'; link.rel = 'stylesheet';
    link.href = new URL('../../../vendor/editors/monaco.css', import.meta.url).href;
    const timer = setTimeout(() => link.onerror(), loadTimeout);
    link.onload = () => { clearTimeout(timer); resolve(); };
    link.onerror = () => { clearTimeout(timer); link.remove(); monacoStyles = null; reject(new Error('Monaco styles failed to load')); };
    document.head.append(link);
  });
  return monacoStyles;
}
async function importEditor(engine) {
  switch (engine) {
    case 'codemirror': return (await import('../../../vendor/editors/codemirror.js')).createEditor;
    case 'monaco': {
      await loadMonacoStyles();
      return (await import('../../../vendor/editors/monaco.js')).createEditor;
    }
    default: return (await import('./ace.js')).createEditor;
  }
}

export function loadEditor(engine) {
  return withTimeout(importEditor(engine));
}
