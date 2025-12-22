// Keyboard shortcuts
export function initKeyboardShortcuts() {
  document.addEventListener('keydown', (e) => {
    // Ctrl/Cmd + K: Focus search
    if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
      e.preventDefault();
      document.querySelector('.search-input')?.focus();
    }

    // Ctrl/Cmd + N: New snippet
    if ((e.ctrlKey || e.metaKey) && e.key === 'n') {
      e.preventDefault();
      const app = Alpine.$data(document.querySelector('[x-data="snippetsApp()"]'));
      if (app) app.newSnippet();
    }

    // Escape: Close editor/modal
    if (e.key === 'Escape') {
      const app = Alpine.$data(document.querySelector('[x-data="snippetsApp()"]'));
      if (app?.showEditor) app.cancelEdit();
      if (app?.showDeleteModal) app.showDeleteModal = false;
    }
  });
}
