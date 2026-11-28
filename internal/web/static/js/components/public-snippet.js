// Public snippet view component
import { showToast } from '../modules/toast.js';
import { api } from '../modules/api.js';
import { renderMarkdownWithCopy, copyMarkdownBlock } from '../utils/markdown.js';
import { getLanguageColor, highlightCode } from '../utils/helpers.js';

// Sanitize filename by replacing spaces with underscores
function sanitizeFilename(filename) {
  return filename
    .replace(/\s+/g, '_')  // Replace spaces with underscores
    .replace(/_+/g, '_');   // Replace multiple underscores with single underscore
}

// Arabic/RTL text detection
function isArabicText(text) {
  if (!text) return false;
  const arabicPattern = /[\u0600-\u06FF\u0750-\u077F\u08A0-\u08FF\uFB50-\uFDFF\uFE70-\uFEFF]/;
  return arabicPattern.test(text);
}

export function initPublicSnippet(Alpine) {
  Alpine.data('publicSnippet', () => ({
    snippet: { title: '', description: '', tags: [], files: [], is_public: false },
    loading: true,
    error: false,
    errorMessage: '',
    activeFileIndex: 0,
    isAuthenticated: false,

    async init() {
      const path = window.location.pathname;
      const match = path.match(/\/s\/([a-zA-Z0-9_-]+)\/?$/);

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
        const response = await fetch(`${window.SNIPO_CONFIG?.basePath || ''}/api/v1/snippets/public/${snippetId}`);
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
            // Apply RTL direction if content is Arabic
            this.applyTextDirection();
          });
        } else {
          this.error = true;
          this.errorMessage = 'This snippet is not available or not public';
        }
      } catch (err) {
        this.error = true;
        this.errorMessage = 'Failed to load snippet';
      } finally {
        this.loading = false;
      }
    },

    applyTextDirection() {
      const title = this.snippet?.title || '';
      const description = this.snippet?.description || '';
      const content = this.getCurrentContent();

      const container = document.querySelector('.public-snippet');
      const titleEl = document.querySelector('.public-title');
      const descriptionEl = document.querySelector('.public-description');

      if (container && isArabicText(title)) {
        container.classList.add('rtl');
      }

      if (titleEl && isArabicText(title)) {
        titleEl.classList.add('rtl');
      }

      if (descriptionEl && isArabicText(description)) {
        descriptionEl.classList.add('rtl');
      }
    },

    async checkAuth() {
      try {
        const response = await fetch((window.SNIPO_CONFIG?.basePath || '') + '/api/v1/auth/check', {
          credentials: 'include'
        });
        const json = await response.json();
        this.isAuthenticated = response.ok && json.data?.authenticated;
      } catch (err) {
        this.isAuthenticated = false;
      }
    },

    async togglePublic() {
      if (!this.isAuthenticated) {
        showToast('Sign in to change visibility', 'error');
        return;
      }

      const newPublicState = !this.snippet.is_public;
      
      try {
        const path = `/api/v1/snippets/${this.snippet.id}`;
        const current = await api.get(path);
        if (!current || current.error) throw new Error(current?.error?.message || 'Could not load the latest version.');
        const result = await api.put(path, {
          kind: current.kind, revision: current.revision, title: current.title,
          description: current.description, content: current.content, language: current.language,
          files: current.files?.map(({ id, filename, content, language }) => ({ id, filename, content, language })),
          tags: (current.tags || []).map(tag => tag.name), folder_id: current.folders?.[0]?.id || null,
          is_public: newPublicState, is_archived: current.is_archived, expires_at: current.expires_at,
        });
        if (result && !result.error) {
          this.snippet.is_public = newPublicState;
          showToast(newPublicState ? 'This item is now public' : 'This item is now private');
          
          // If made private, redirect to home after a delay
          if (!newPublicState) {
            setTimeout(() => {
              const basePath = window.SNIPO_CONFIG?.basePath || '';
              window.location.href = basePath + '/';
            }, 1500);
          }
        } else {
          showToast(result?.error?.message || 'Failed to update visibility', 'error');
        }
      } catch (err) {
        showToast(err.message || 'Failed to update visibility', 'error');
      }
    },

    getFiles() {
      if (this.snippet.kind === 'note') {
        return [{ filename: this.snippet.title.replace(/[\\/:*?"<>|]/g, '_') +
          (this.snippet.language === 'markdown' ? '.md' : '.txt'), content: this.snippet.content, language: this.snippet.language }];
      }
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
        showToast(this.snippet.kind === 'note' ? 'Note copied to clipboard' : 'Code copied to clipboard');
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
      const fileUrl = `${window.location.origin}${window.SNIPO_CONFIG?.basePath || ''}/api/v1/snippets/public/${this.snippet.id}/files/${encodeURIComponent(filename)}`;
      
      await navigator.clipboard.writeText(fileUrl);
      showToast('File URL copied to clipboard');
    },

    renderMarkdownWithCopy,
    copyMarkdownBlock,
    getLanguageColor,
    highlightCode,
    isArabicText,

    formatDate(dateStr) {
      if (!dateStr) return '';
      const date = new Date(dateStr);
      return date.toLocaleDateString();
    },
  }));
}
