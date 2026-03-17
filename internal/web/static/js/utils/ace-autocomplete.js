// Enhanced Ace Editor autocomplete functionality
export function setupEnhancedAutocomplete(editor) {
  if (!editor) return;

  try {
    // Enable live autocompletion
    editor.setOptions({
      enableBasicAutocompletion: true,
      enableLiveAutocompletion: true,
      enableSnippets: true
    });

    // Add word completer for better suggestions
    const wordCompleter = {
      getCompletions: function(editor, session, pos, prefix, callback) {
        // Get all words from the document
        const text = session.getValue();
        const wordRegex = /\b\w+\b/g;
        const words = {};
        let match;
        
        while ((match = wordRegex.exec(text)) !== null) {
          const word = match[0];
          if (word.length > 1 && word.indexOf(prefix) === 0) {
            words[word] = true;
          }
        }
        
        const wordList = Object.keys(words);
        callback(null, wordList.map(function(word) {
          return {
            caption: word,
            value: word,
            meta: 'local'
          };
        }));
      }
    };

    // Register the completer
    const langTools = ace.require('ace/ext/language_tools');
    if (langTools) {
      langTools.addCompleter(wordCompleter);
    }

    // Add keyboard shortcut for triggering autocomplete (Ctrl+Space like VSCode)
    editor.commands.addCommand({
      name: 'triggerAutocomplete',
      bindKey: { win: 'Ctrl-Space', mac: 'Ctrl-Space' },
      exec: function(editor) {
        editor.execCommand('startAutocomplete');
      }
    });

    // Add command for format code
    editor.commands.addCommand({
      name: 'beautify',
      bindKey: { win: 'Ctrl-Shift-B', mac: 'Ctrl-Shift-B' },
      exec: function(editor) {
        try {
          const beautify = ace.require('ace/ext/beautify');
          beautify.beautify(editor.session);
        } catch (e) {
          console.warn('Beautify not available');
        }
      }
    });

    // Add command for toggle comment (Ctrl+/)
    editor.commands.addCommand({
      name: 'toggleComment',
      bindKey: { win: 'Ctrl-/', mac: 'Ctrl-/' },
      exec: function(editor) {
        editor.toggleComment();
      }
    });

    // Add command for duplicate line (Shift+Alt+Down)
    editor.commands.addCommand({
      name: 'duplicateSelection',
      bindKey: { win: 'Shift-Alt-Down', mac: 'Shift-Option-Down' },
      exec: function(editor) {
        editor.copyLinesDown();
      }
    });

    // Add command for move line up (Alt+Up)
    editor.commands.addCommand({
      name: 'moveLineUp',
      bindKey: { win: 'Alt-Up', mac: 'Option-Up' },
      exec: function(editor) {
        editor.moveLinesUp();
      }
    });

    // Add command for move line down (Alt+Down)
    editor.commands.addCommand({
      name: 'moveLineDown',
      bindKey: { win: 'Alt-Down', mac: 'Option-Down' },
      exec: function(editor) {
        editor.moveLinesDown();
      }
    });

    // Add command for delete line (Ctrl+D)
    editor.commands.addCommand({
      name: 'removeline',
      bindKey: { win: 'Ctrl-D', mac: 'Command-D' },
      exec: function(editor) {
        editor.removeLines();
      }
    });

    // Add command for find (Ctrl+F)
    editor.commands.addCommand({
      name: 'find',
      bindKey: { win: 'Ctrl-F', mac: 'Command-F' },
      exec: function(editor) {
        editor.execCommand('find');
      }
    });

    // Add command for replace (Ctrl+H)
    editor.commands.addCommand({
      name: 'replace',
      bindKey: { win: 'Ctrl-H', mac: 'Command-Option-F' },
      exec: function(editor) {
        editor.execCommand('replace');
      }
    });

    // Add command for goto line (Ctrl+G)
    editor.commands.addCommand({
      name: 'gotoLine',
      bindKey: { win: 'Ctrl-G', mac: 'Command-G' },
      exec: function(editor) {
        editor.execCommand('gotoline');
      }
    });

    // Add command for select all (Ctrl+A)
    editor.commands.addCommand({
      name: 'selectAll',
      bindKey: { win: 'Ctrl-A', mac: 'Command-A' },
      exec: function(editor) {
        editor.selectAll();
      }
    });

    // Add command for undo (Ctrl+Z)
    editor.commands.addCommand({
      name: 'undo',
      bindKey: { win: 'Ctrl-Z', mac: 'Command-Z' },
      exec: function(editor) {
        editor.undo();
      }
    });

    // Add command for redo (Ctrl+Y or Ctrl+Shift+Z)
    editor.commands.addCommand({
      name: 'redo',
      bindKey: { win: 'Ctrl-Y', mac: 'Command-Shift-Z' },
      exec: function(editor) {
        editor.redo();
      }
    });

    // Enable multiple cursors (Alt+Click is default, add keyboard shortcuts)
    editor.commands.addCommand({
      name: 'addCursorAbove',
      bindKey: { win: 'Ctrl-Alt-Up', mac: 'Ctrl-Option-Up' },
      exec: function(editor) {
        editor.selection.addCursorAbove();
      }
    });

    editor.commands.addCommand({
      name: 'addCursorBelow',
      bindKey: { win: 'Ctrl-Alt-Down', mac: 'Ctrl-Option-Down' },
      exec: function(editor) {
        editor.selection.addCursorBelow();
      }
    });

  } catch (e) {
    console.warn('Failed to setup enhanced autocomplete:', e);
  }
}

export function getEnhancedCompleters() {
  const completers = [];
  
  try {
    const langTools = ace.require('ace/ext/language_tools');
    if (langTools) {
      // Add text completers
      completers.push(langTools.textCompleter);
      // Add keyword completers
      completers.push(langTools.keyWordCompleter);
      // Add snippet completers
      completers.push(langTools.snippetCompleter);
    }
  } catch (e) {
    // Language tools not available
  }

  return completers;
}
