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

// Initialize Alpine.js components and stores
document.addEventListener('alpine:init', () => {
  // Initialize stores
  initAppStore(Alpine);
  
  // Initialize components
  initSnippetsApp(Alpine);
  initLoginForm(Alpine);
  initPublicSnippet(Alpine);
});

// Initialize keyboard shortcuts after DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  initKeyboardShortcuts();
});

// Expose globals for compatibility with inline handlers
window.api = api;
window.showToast = showToast;
window.theme = theme;
