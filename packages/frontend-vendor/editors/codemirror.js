import { EditorState, Compartment } from '@codemirror/state';
import { EditorView, keymap, lineNumbers, highlightActiveLine, drawSelection } from '@codemirror/view';
import { history, historyKeymap, defaultKeymap, indentWithTab } from '@codemirror/commands';
import { syntaxHighlighting, defaultHighlightStyle, indentUnit, bracketMatching, StreamLanguage } from '@codemirror/language';
import { autocompletion, completionKeymap, closeBrackets } from '@codemirror/autocomplete';
import { searchKeymap, highlightSelectionMatches } from '@codemirror/search';
import { oneDark } from '@codemirror/theme-one-dark';
import { javascript } from '@codemirror/lang-javascript';
import { php } from '@codemirror/lang-php';
import { python } from '@codemirror/lang-python';
import { json } from '@codemirror/lang-json';
import { html } from '@codemirror/lang-html';
import { css } from '@codemirror/lang-css';
import { markdown } from '@codemirror/lang-markdown';
import { sql } from '@codemirror/lang-sql';
import { go } from '@codemirror/legacy-modes/mode/go';
import { rust } from '@codemirror/legacy-modes/mode/rust';
import { shell } from '@codemirror/legacy-modes/mode/shell';
import { cpp, c, java, csharp, kotlin, scala } from '@codemirror/legacy-modes/mode/clike';
import { ruby } from '@codemirror/legacy-modes/mode/ruby';
import { swift } from '@codemirror/legacy-modes/mode/swift';
import { powerShell } from '@codemirror/legacy-modes/mode/powershell';
import { stex } from '@codemirror/legacy-modes/mode/stex';
import { yaml } from '@codemirror/legacy-modes/mode/yaml';

function languageSupport(language) {
  const modern = { javascript, typescript: () => javascript({ typescript: true }), python, php, json, html, css, markdown, sql };
  if (modern[language]) return modern[language]();
  const legacy = { go, rust, bash: shell, shell, yaml, cpp, c, cuda: cpp, java, csharp, kotlin, scala, ruby, swift, powershell: powerShell, tex: stex, bibtex: stex };
  return legacy[language] ? StreamLanguage.define(legacy[language]) : [];
}

export function createEditor(container, { content, language, settings, dark, onChange, onSave }) {
  const preferences = new Compartment();
  const syntax = new Compartment();
  const direction = new Compartment();
  let options = settings;
  let isDark = dark;
  let currentLanguage = language;
  function configuration() {
    const size = options.editor_tab_size || 2;
    return [
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
      isDark ? oneDark : [],
      EditorView.theme({ '&': { height: '100%', fontSize: `${options.editor_font_size || 14}px` },
        '.cm-scroller': { overflow: 'auto', fontFamily: 'var(--font-mono)' },
        '.cm-content': { minHeight: '100%' }, '&.cm-focused': { outline: 'none' } }),
      EditorState.tabSize.of(size), indentUnit.of(options.editor_use_soft_tabs === false ? '\t' : ' '.repeat(size)),
      options.editor_word_wrap !== false ? EditorView.lineWrapping : [],
      options.editor_show_gutter !== false ? lineNumbers() : [],
      options.editor_highlight_active_line !== false ? highlightActiveLine() : [],
      autocompletion({ activateOnTyping: !!options.editor_enable_live_autocompletion })
    ];
  }
  const extensions = () => [history(), drawSelection(), bracketMatching(), closeBrackets(), highlightSelectionMatches(),
    keymap.of([{ key: 'Mod-s', run: () => { onSave(); return true; } }, ...defaultKeymap, ...historyKeymap, ...searchKeymap, ...completionKeymap, indentWithTab]),
    preferences.of(configuration()), syntax.of(languageSupport(currentLanguage)), direction.of([]),
    EditorView.contentAttributes.of({ 'aria-label': 'Snippet code' }),
    EditorView.updateListener.of(update => { if (update.docChanged) onChange(update.state.doc.toString()); })];
  const view = new EditorView({ parent: container, state: EditorState.create({ doc: content, extensions: extensions() }) });
  return {
    getValue: () => view.state.doc.toString(),
    // New file = new undo history, so undo can never restore another file's code.
    setValue: value => view.setState(EditorState.create({ doc: value, extensions: extensions() })),
    setLanguage(value) { currentLanguage = value; view.dispatch({ effects: syntax.reconfigure(languageSupport(value)) }); },
    configure(s, d) { options = s; isDark = d; view.dispatch({ effects: preferences.reconfigure(configuration()) }); },
    setDirection(rtl, mixed) { view.dispatch({ effects: direction.reconfigure(EditorView.editorAttributes.of({ dir: rtl && !mixed ? 'rtl' : 'ltr' })) }); },
    focus: () => view.focus(), resize: () => view.requestMeasure(), destroy: () => view.destroy()
  };
}
