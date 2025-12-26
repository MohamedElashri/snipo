// Keyboard shortcuts
export function initKeyboardShortcuts() {
  document.addEventListener('keydown', (e) => {
    // Forward slash: Focus search (unless typing in input/textarea)
    if (e.key === '/' && !['INPUT', 'TEXTAREA'].includes(e.target.tagName)) {
      e.preventDefault();
      e.stopPropagation();
      const searchInput = document.querySelector('.search-input');
      if (searchInput) {
        searchInput.focus();
        searchInput.select();
      }
      return false;
    }
    
    // Ctrl/Cmd + K: Focus search (alternative)
    if ((e.ctrlKey || e.metaKey) && (e.key === 'k' || e.key === 'K')) {
      e.preventDefault();
      e.stopPropagation();
      const searchInput = document.querySelector('.search-input');
      if (searchInput) {
        searchInput.focus();
        searchInput.select();
      }
      return false;
    }

    // Ctrl/Cmd + N: New snippet
    if ((e.ctrlKey || e.metaKey) && (e.key === 'n' || e.key === 'N')) {
      e.preventDefault();
      e.stopPropagation();
      const app = Alpine.$data(document.querySelector('[x-data="snippetsApp()"]'));
      if (app) app.newSnippet();
      return false;
    }

    // Escape: Close editor/modal
    if (e.key === 'Escape' || e.key === 'Esc') {
      const app = Alpine.$data(document.querySelector('[x-data="snippetsApp()"]'));
      if (app?.showEditor) app.cancelEdit();
      if (app?.showDeleteModal) app.showDeleteModal = false;
      if (app?.showSearchHelp) app.showSearchHelp = false;
    }
  });
}
