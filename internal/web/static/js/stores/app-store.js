// Global app store for Alpine.js
import { theme } from '../modules/theme.js';

export function initAppStore(Alpine) {
  Alpine.store('app', {
    sidebarOpen: window.innerWidth > 768,
    currentView: 'snippets',
    loading: false,
    darkMode: theme.get() === 'dark',

    toggleSidebar() {
      this.sidebarOpen = !this.sidebarOpen;
    },

    toggleTheme() {
      theme.toggle();
      this.darkMode = theme.get() === 'dark';
    }
  });
}
