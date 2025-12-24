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
  },

  /**
   * Apply custom CSS from settings
   * @param {string} customCSS - Custom CSS string to apply
   */
  applyCustomCSS(customCSS) {
    // Remove existing custom CSS style tag if it exists
    const existingStyle = document.getElementById('snipo-custom-css');
    if (existingStyle) {
      existingStyle.remove();
    }

    // If no custom CSS provided, just remove the old style
    if (!customCSS || customCSS.trim() === '') {
      return;
    }

    // Create and inject new style tag
    const style = document.createElement('style');
    style.id = 'snipo-custom-css';
    style.textContent = customCSS;
    document.head.appendChild(style);
  },

  /**
   * Validate custom CSS for common issues
   * @param {string} css - CSS string to validate
   * @returns {Object} - { valid: boolean, warnings: string[] }
   */
  validateCustomCSS(css) {
    const warnings = [];
    
    if (!css || css.trim() === '') {
      return { valid: true, warnings: [] };
    }

    // Check for script tags (security)
    if (/<script/i.test(css)) {
      warnings.push('Script tags are not allowed in CSS');
      return { valid: false, warnings };
    }

    // Check for @import with external URLs (potential security issue)
    if (/@import\s+url\s*\(\s*['"]?https?:\/\//i.test(css)) {
      warnings.push('External @import URLs may pose security risks');
    }

    // Check for excessive nesting or size
    if (css.length > 50000) {
      warnings.push('CSS is very large (>50KB). Consider optimizing for performance.');
    }

    // Check for unclosed braces
    const openBraces = (css.match(/\{/g) || []).length;
    const closeBraces = (css.match(/\}/g) || []).length;
    if (openBraces !== closeBraces) {
      warnings.push('Unmatched braces detected. CSS may not apply correctly.');
    }

    return { valid: warnings.length === 0 || !warnings.some(w => w.includes('not allowed')), warnings };
  }
};

