// Main entry point for Snipo frontend application
// Import modules
import { theme } from './modules/theme.js';
import { api } from './modules/api.js';
import { showToast } from './modules/toast.js';

// Import stores
import { initAppStore } from './stores/app-store.js';

// Import components
import { initSnippetsApp } from './components/snippets-app.js';
import { initLoginForm } from './components/login-form.js';
import { initPublicSnippet } from './components/public-snippet.js';

// Import utilities
import { initKeyboardShortcuts } from './utils/keyboard.js';

// Initialize theme on load
theme.init();

// Expose globals for compatibility with inline handlers
window.api = api;
window.showToast = showToast;
window.theme = theme;

// Register alpine:init listener BEFORE loading Alpine scripts.
// When Alpine auto-starts it will fire this event and our components
// will be registered in time.
document.addEventListener('alpine:init', () => {
  initAppStore(Alpine);
  initSnippetsApp(Alpine);
  initLoginForm(Alpine);
  initPublicSnippet(Alpine);
});

// Dynamically load Alpine.js (and collapse plugin) so that the
// alpine:init listener above is guaranteed to be in place before
// Alpine auto-starts. This fixes the race condition where defer
// scripts execute before module scripts per the HTML spec.
function loadScript(src) {
  return new Promise((resolve, reject) => {
    const s = document.createElement('script');
    s.src = src;
    s.onload = resolve;
    s.onerror = reject;
    document.body.appendChild(s);
  });
}

const basePath = window.SNIPO_CONFIG?.basePath || '';
await loadScript(basePath + '/static/vendor/js/alpine-collapse.min.js');
await loadScript(basePath + '/static/vendor/js/alpine.min.js');

// Initialize keyboard shortcuts (DOM is already ready when modules execute)
initKeyboardShortcuts();
