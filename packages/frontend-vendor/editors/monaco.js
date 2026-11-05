import * as monaco from 'monaco-editor/editor/editor.main.js';

self.MonacoEnvironment = {
  getWorker(_id, label) {
    const worker = ({ javascript: 'typescript', typescript: 'typescript', json: 'json', html: 'html',
      handlebars: 'html', razor: 'html', css: 'css', scss: 'css', less: 'css' })[label];
    return new Worker(new URL(worker ? `./monaco-${worker}-worker.js` : './monaco-worker.js', import.meta.url), { type: 'module' });
  }
};
function mode(language) {
  return ({ bash: 'shell', plaintext: 'plaintext', cpp: 'cpp', cuda: 'cpp', tex: 'plaintext', bibtex: 'plaintext' })[language] || language;
}
export function createEditor(container, { content, language, settings, dark, onChange, onSave }) {
  let preferences = settings;
  let darkTheme = dark;
  let model = monaco.editor.createModel(content, mode(language));
  const editor = monaco.editor.create(container, { model, automaticLayout: false, minimap: { enabled: false },
    scrollBeyondLastLine: false, ariaLabel: 'Snippet code', fixedOverflowWidgets: true });
  const listener = editor.onDidChangeModelContent(() => onChange(editor.getValue()));
  editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, onSave);
  const adapter = {
    getValue: () => editor.getValue(),
    setValue(value) {
      const previous = model;
      model = monaco.editor.createModel(value, previous.getLanguageId());
      editor.setModel(model);
      previous.dispose();
      adapter.configure(preferences, darkTheme);
    },
    setLanguage: value => monaco.editor.setModelLanguage(model, mode(value)),
    configure(s, isDark) {
      preferences = s; darkTheme = isDark;
      monaco.editor.setTheme(isDark ? 'vs-dark' : 'vs');
      editor.updateOptions({ fontSize: s.editor_font_size || 14, fontFamily: 'var(--font-mono)',
        wordWrap: s.editor_word_wrap !== false ? 'on' : 'off',
        lineNumbers: s.editor_show_gutter !== false ? 'on' : 'off',
        renderLineHighlight: s.editor_highlight_active_line !== false ? 'line' : 'none',
        guides: { indentation: s.editor_show_indent_guides !== false },
        rulers: s.editor_show_print_margin ? [80] : [],
        quickSuggestions: !!s.editor_enable_live_autocompletion,
        snippetSuggestions: s.editor_enable_snippets !== false ? 'inline' : 'none' });
      model.updateOptions({ tabSize: s.editor_tab_size || 2, insertSpaces: s.editor_use_soft_tabs !== false });
    },
    // Monaco handles bidirectional text internally; keep its layout LTR.
    setDirection() {},
    focus: () => editor.focus(), resize: () => editor.layout(),
    destroy() { listener.dispose(); editor.dispose(); model.dispose(); }
  };
  adapter.configure(settings, dark);
  return adapter;
}
