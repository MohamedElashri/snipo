// Public snippet view component
import { showToast } from '../modules/toast.js';
import { getLanguageColor } from '../utils/helpers.js';

// Sanitize filename by replacing spaces with underscores
function sanitizeFilename(filename) {
  return filename
    .replace(/\s+/g, '_')  // Replace spaces with underscores
    .replace(/_+/g, '_');   // Replace multiple underscores with single underscore
}

export function initPublicSnippet(Alpine) {
  Alpine.data('publicSnippet', () => ({
    snippet: null,
    loading: true,
    error: false,
    errorMessage: '',
    activeFileIndex: 0,
    isAuthenticated: false,

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

      // Check if user is authenticated
      await this.checkAuth();

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

    async checkAuth() {
      try {
        const response = await fetch('/api/v1/auth/check', {
          credentials: 'include'
        });
        const json = await response.json();
        this.isAuthenticated = response.ok && json.authenticated;
      } catch (err) {
        this.isAuthenticated = false;
      }
    },

    async togglePublic() {
      if (!this.isAuthenticated) {
        showToast('Please login to change snippet visibility', 'error');
        return;
      }

      const newPublicState = !this.snippet.is_public;
      
      try {
        const response = await fetch(`/api/v1/snippets/${this.snippet.id}`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json'
          },
          credentials: 'include',
          body: JSON.stringify({
            ...this.snippet,
            is_public: newPublicState
          })
        });

        if (response.ok) {
          this.snippet.is_public = newPublicState;
          showToast(newPublicState ? 'Snippet is now public' : 'Snippet is now private');
          
          // If made private, redirect to home after a delay
          if (!newPublicState) {
            setTimeout(() => {
              window.location.href = '/';
            }, 1500);
          }
        } else {
          const json = await response.json();
          showToast(json.error?.message || 'Failed to update snippet', 'error');
        }
      } catch (err) {
        showToast('Failed to update snippet', 'error');
      }
    },

    getFiles() {
      if (!this.snippet) return [];
      // If snippet has files array with content, use that
      if (this.snippet.files && this.snippet.files.length > 0) {
        return this.snippet.files;
      }
      // Otherwise, create a single file from legacy content
      return [{
        filename: 'snippet.' + (this.snippet.language || 'txt'),
        content: this.snippet.content || '',
        language: this.snippet.language || 'plaintext'
      }];
    },

    hasMultipleFiles() {
      const files = this.getFiles();
      return files.length > 1;
    },

    getCurrentContent() {
      const files = this.getFiles();
      if (files.length === 0) return '';
      return files[this.activeFileIndex]?.content || '';
    },

    getCurrentLanguage() {
      const files = this.getFiles();
      if (files.length === 0) return 'plaintext';
      return files[this.activeFileIndex]?.language || 'plaintext';
    },

    getCurrentFilename() {
      const files = this.getFiles();
      if (files.length === 0) return 'snippet.txt';
      return files[this.activeFileIndex]?.filename || 'snippet.txt';
    },

    async copyCode() {
      const content = this.getCurrentContent();
      if (content) {
        await navigator.clipboard.writeText(content);
        showToast('Code copied to clipboard');
      }
    },

    async downloadFile() {
      const filename = sanitizeFilename(this.getCurrentFilename());
      const content = this.getCurrentContent();
      
      // Create a blob and trigger download
      const blob = new Blob([content], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      
      showToast('File downloaded');
    },

    async copyFileUrl() {
      const filename = sanitizeFilename(this.getCurrentFilename());
      const fileUrl = `${window.location.origin}/api/v1/snippets/public/${this.snippet.id}/files/${encodeURIComponent(filename)}`;
      
      await navigator.clipboard.writeText(fileUrl);
      showToast('File URL copied to clipboard');
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
