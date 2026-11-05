// Adapters own library instances outside Alpine's reactive proxy. Application code
// only uses getValue/setValue/setLanguage/configure/focus/resize/destroy.
export const editorEngines = ['ace', 'codemirror', 'monaco'];
export function readEditorPreference() {
  try {
    const saved = localStorage.getItem('snipo-editor');
    return editorEngines.includes(saved) ? saved : 'ace';
  } catch { return 'ace'; }
}
let monacoStyles;
function loadMonacoStyles() {
  if (!monacoStyles) monacoStyles = new Promise((resolve, reject) => {
    const link = document.createElement('link');
    link.id = 'monaco-styles'; link.rel = 'stylesheet';
    link.href = new URL('../../../vendor/editors/monaco.css', import.meta.url).href;
    link.onload = resolve;
    link.onerror = () => { link.remove(); monacoStyles = null; reject(new Error('Monaco styles failed to load')); };
    document.head.append(link);
  });
  return monacoStyles;
}
export async function loadEditor(engine) {
  switch (engine) {
    case 'codemirror': return (await import('../../../vendor/editors/codemirror.js')).createEditor;
    case 'monaco': {
      await loadMonacoStyles();
      return (await import('../../../vendor/editors/monaco.js')).createEditor;
    }
    default: return (await import('./ace.js')).createEditor;
  }
}
