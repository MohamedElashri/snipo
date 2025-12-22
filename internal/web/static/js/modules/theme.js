// Theme management module
export const theme = {
  get() {
    return localStorage.getItem('snipo-theme') || 'dark';
  },
  
  set(value) {
    localStorage.setItem('snipo-theme', value);
    document.documentElement.setAttribute('data-theme', value);
    this.updatePrismTheme(value);
    this.updateAceTheme(value);
  },
  
  toggle() {
    const current = this.get();
    this.set(current === 'dark' ? 'light' : 'dark');
  },
  
  init() {
    const saved = this.get();
    document.documentElement.setAttribute('data-theme', saved);
    this.updatePrismTheme(saved);
  },
  
  updatePrismTheme(themeName) {
    const prismLink = document.getElementById('prism-theme');
    if (prismLink) {
      if (themeName === 'dark') {
        prismLink.href = '/static/vendor/css/prism-tomorrow.min.css';
      } else {
        prismLink.href = '/static/vendor/css/prism.min.css';
      }
    }
  },
  
  updateAceTheme(themeName) {
    if (typeof ace !== 'undefined') {
      const editors = document.querySelectorAll('.ace_editor');
      editors.forEach(editorEl => {
        const editor = ace.edit(editorEl);
        if (editor) {
          // Check if there's a stored editor theme preference
          try {
            const settingsStr = sessionStorage.getItem('snipo-settings');
            const settings = settingsStr ? JSON.parse(settingsStr) : null;
            const editorTheme = settings?.editor_theme || 'auto';
            
            let aceTheme = 'ace/theme/chrome';
            if (editorTheme === 'auto') {
              aceTheme = themeName === 'dark' ? 'ace/theme/monokai' : 'ace/theme/chrome';
            } else {
              aceTheme = `ace/theme/${editorTheme}`;
            }
            editor.setTheme(aceTheme);
          } catch (e) {
            // Fallback to default behavior
            const aceTheme = themeName === 'dark' ? 'ace/theme/monokai' : 'ace/theme/chrome';
            editor.setTheme(aceTheme);
          }
        }
      });
    }
  }
};
