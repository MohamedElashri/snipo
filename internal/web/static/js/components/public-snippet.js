// Public snippet view component
import { showToast } from '../modules/toast.js';
import { getLanguageColor } from '../utils/helpers.js';

export function initPublicSnippet(Alpine) {
  Alpine.data('publicSnippet', () => ({
    snippet: null,
    loading: true,
    error: false,
    errorMessage: '',

    async init() {
      const path = window.location.pathname;
      const match = path.match(/\/s\/([a-zA-Z0-9]+)/);

      if (!match) {
        this.error = true;
        this.errorMessage = 'Invalid snippet URL';
        this.loading = false;
        return;
      }

      const snippetId = match[1];

      try {
        const response = await fetch(`/api/v1/snippets/public/${snippetId}`);
        const json = await response.json();

        // Handle error response format: { error: { code, message } }
        if (json.error) {
          this.error = true;
          this.errorMessage = json.error.message || 'This snippet is not available or not public';
          return;
        }

        // Handle success response format: { data: {...}, meta }
        if (response.ok && json.data) {
          this.snippet = json.data;
          this.$nextTick(() => {
            if (typeof Prism !== 'undefined') {
              Prism.highlightAll();
            }
          });
        } else {
          this.error = true;
          this.errorMessage = 'This snippet is not available or not public';
        }
      } catch (err) {
        this.error = true;
        this.errorMessage = 'Failed to load snippet';
      }

      this.loading = false;
    },

    async copyCode() {
      if (this.snippet?.content) {
        await navigator.clipboard.writeText(this.snippet.content);
        showToast('Code copied to clipboard');
      }
    },

    getLanguageColor,

    formatDate(dateStr) {
      if (!dateStr) return '';
      const date = new Date(dateStr);
      return date.toLocaleDateString();
    },

    autoResizeInput(element) { window.autoResizeInput(element); },
    autoResizeSelect(element) { window.autoResizeSelect(element); }
  }));
}
