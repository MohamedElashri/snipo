// Global shortcuts are scoped to the authenticated workspace.
export function initKeyboardShortcuts() {
  document.addEventListener('keydown', (event) => {
    const root = document.querySelector('[x-data="snippetsApp()"]');
    if (!root) return;
    const app = window.Alpine.$data(root);
    const typing = event.target.closest('input, textarea, [contenteditable="true"], .ace_editor');
    const command = event.ctrlKey || event.metaKey;
    const key = event.key.toLowerCase();

    if ((!typing && key === '/') || (command && key === 'k')) {
      event.preventDefault();
      if (app.showSettings) app.closeSettings();
      const input = root.querySelector('.search-input');
      input?.focus();
      input?.select();
    } else if (command && key === 'n') {
      event.preventDefault();
      app.newSnippet();
    } else if (command && key === 's' && app.isEditing) {
      event.preventDefault();
      app.saveSnippet();
    } else if (key === 'escape') {
      // Close only the frontmost layer, leaving the editor behind it intact.
      if (app.tokenPasswordAction) app.cancelTokenPassword();
      else if (app.showDisableLoginPassword) app.cancelDisableLoginPassword();
      else if (app.showDeleteModal) app.showDeleteModal = false;
      else if (app.showSearchHelp) app.showSearchHelp = false;
      else if (app.showFolderModal) app.showFolderModal = false;
      else if (app.showSettings) app.closeSettings();
      else if (event.target.closest('details[open]')) event.target.closest('details').open = false;
      else if (app.showEditor) app.cancelEdit();
    }
  });
}
