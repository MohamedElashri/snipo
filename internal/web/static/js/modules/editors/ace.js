import { getAceMode } from '../../utils/ace-utils.js';

export function createEditor(container, { content, language, settings, dark, onChange, onSave }) {
  ace.config.set('basePath', window.SNIPO_CONFIG.assetBasePath + '/vendor/js/ace');
  const editor = ace.edit(container);
  editor.setValue(content, -1);
  editor.session.setMode(getAceMode(language));
  editor.commands.addCommand({ name: 'saveSnippet', bindKey: { win: 'Ctrl-S', mac: 'Command-S' }, exec: onSave });
  editor.session.on('change', () => onChange(editor.getValue()));
  const adapter = {
    getValue: () => editor.getValue(),
    setValue(value) { editor.setValue(value, -1); editor.session.getUndoManager().reset(); },
    setLanguage: value => editor.session.setMode(getAceMode(value)),
    focus: () => editor.focus(),
    resize: () => editor.resize(),
    setDirection: (rtl, mixed) => {
      editor.setOption('rtl', rtl);
      container.classList.toggle('rtl', rtl);
      container.classList.toggle('mixed', mixed);
    },
    configure(s, isDark) {
      editor.setOptions({
        fontSize: s.editor_font_size || 14, tabSize: s.editor_tab_size || 2,
        useSoftTabs: s.editor_use_soft_tabs !== false, wrap: s.editor_word_wrap !== false,
        showPrintMargin: !!s.editor_show_print_margin, showGutter: s.editor_show_gutter !== false,
        displayIndentGuides: s.editor_show_indent_guides !== false,
        highlightActiveLine: s.editor_highlight_active_line !== false,
        enableSnippets: s.editor_enable_snippets !== false,
        enableBasicAutocompletion: true,
        enableLiveAutocompletion: !!s.editor_enable_live_autocompletion
      });
      editor.setTheme(`ace/theme/${!s.editor_theme || s.editor_theme === 'auto' ? (isDark ? 'monokai' : 'chrome') : s.editor_theme}`);
    },
    destroy: () => editor.destroy()
  };
  adapter.configure(settings, dark);
  return adapter;
}
