// Editor mixin - handles snippet editing operations
import { api } from '../../modules/api.js';
import { showToast } from '../../modules/toast.js';
import { getAceMode } from '../../utils/ace-utils.js';
import { theme } from '../../modules/theme.js';

export const editorMixin = {
  // Editor operations (imported from original app.js)
  // This file contains editor-related methods and state
  // Methods: viewSnippet, editSnippet, newSnippet, saveSnippet, startEditing, cancelEditing, etc.
  
  async viewSnippet(snippet) {
    const result = await api.get(`/api/v1/snippets/${snippet.id}`);
    if (result) {
      this.editingSnippet = {
        ...result,
        tags: (result.tags || []).map(t => t.name),
        folder_id: result.folders?.[0]?.id || null,
        files: result.files || []
      };
      this.activeFileIndex = 0;
      this.showEditor = true;
      this.isEditing = false;
      this.updateUrl({ snippet: snippet.id });
      this.$nextTick(() => this.highlightAll());
    }
  },

  newSnippet() {
    this.editingSnippet = {
      id: null,
      title: '',
      description: '',
      content: '',
      language: 'plaintext',
      tags: [],
      folder_id: null,
      is_public: false,
      is_favorite: false,
      files: [{
        id: 0,
        filename: 'snippet.txt',
        content: '',
        language: 'plaintext'
      }]
    };
    this.activeFileIndex = 0;
    this.showEditor = true;
    this.isEditing = true;
    this.updateUrl({ edit: true });

    this.$nextTick(() => {
      this.updateAceEditor();
      const input = document.querySelector('.filename-input');
      if (input) {
        input.focus();
        input.select();
      }
    });
  },

  async editSnippet(snippet) {
    const result = await api.get(`/api/v1/snippets/${snippet.id}`);
    if (result) {
      this.editingSnippet = {
        ...result,
        tags: (result.tags || []).map(t => t.name),
        folder_id: result.folders?.[0]?.id || null,
        files: result.files || []
      };
      this.activeFileIndex = 0;
      this.showEditor = true;
      this.isEditing = true;
      this.updateUrl({ snippet: snippet.id, edit: true });
      this.$nextTick(() => {
        this.updateAceEditor();
        this.highlightAll();
      });
    }
  },

  startEditing() {
    this.isEditing = true;
    if (this.editingSnippet?.id) {
      this.updateUrl({ snippet: this.editingSnippet.id, edit: true });
    }
    this.$nextTick(() => {
      this.updateAceEditor();
      this.highlightAll();
    });
  },

  cancelEditing() {
    this.isEditing = false;
    if (this.editingSnippet?.id) {
      this.viewSnippet(this.editingSnippet);
    } else {
      this.showEditor = false;
      this.editingSnippet = null;
      this.updateUrl({ snippet: null, edit: null });
    }
  },

  async saveSnippet() {
    let folderId = this.editingSnippet.folder_id;
    if (folderId === '' || folderId === null || folderId === undefined) {
      folderId = null;
    } else {
      folderId = parseInt(folderId, 10) || null;
    }

    let files = null;
    if (this.editingSnippet.files && this.editingSnippet.files.length > 0) {
      files = this.editingSnippet.files.map(f => ({
        id: f.id || 0,
        filename: f.filename,
        content: f.content,
        language: f.language
      }));
    }

    const primaryContent = files && files.length > 0 ? files[0].content : this.editingSnippet.content;
    const primaryLanguage = files && files.length > 0 ? files[0].language : this.editingSnippet.language;

    const data = {
      title: this.editingSnippet.title,
      description: this.editingSnippet.description || '',
      content: primaryContent,
      language: primaryLanguage,
      tags: this.editingSnippet.tags || [],
      folder_id: folderId,
      is_public: this.editingSnippet.is_public || false,
      files: files
    };

    let result;
    if (this.editingSnippet.id) {
      result = await api.put(`/api/v1/snippets/${this.editingSnippet.id}`, data);
    } else {
      result = await api.post('/api/v1/snippets', data);
    }

    if (result && !result.error) {
      showToast(this.editingSnippet.id ? 'Snippet updated' : 'Snippet created');
      this.showEditor = false;
      this.isEditing = false;
      this.destroyAceEditor();
      this.resetEditingSnippet();
      this.clearDraft();
      this.updateUrl({});
      await this.loadSnippets();
      await this.loadTags();
      await this.loadFavoritesCount();
    } else if (result?.error) {
      showToast(result.error.message || 'Error saving snippet', 'error');
    }
  },

  cancelEdit() {
    this.showEditor = false;
    this.isEditing = false;
    this.destroyAceEditor();
    this.resetEditingSnippet();
    this.clearDraft();
    
    if (this.filter.folderId) {
      this.updateUrl({ folder: this.filter.folderId });
    } else if (this.filter.tagId) {
      this.updateUrl({ tag: this.filter.tagId });
    } else if (this.filter.isFavorite) {
      this.updateUrl({ favorites: true });
    } else {
      this.updateUrl({});
    }
  },

  resetEditingSnippet() {
    this.editingSnippet = {
      id: null,
      title: '',
      description: '',
      content: '',
      language: 'plaintext',
      tags: [],
      folder_id: null,
      is_public: false,
      is_favorite: false,
      files: [{
        id: 0,
        filename: 'snippet.txt',
        content: '',
        language: 'plaintext'
      }]
    };
    this.activeFileIndex = 0;
  },

  confirmDelete(snippet) {
    this.deleteTarget = snippet;
    this.showDeleteModal = true;
  },

  async deleteSnippet() {
    if (!this.deleteTarget) return;

    await api.delete(`/api/v1/snippets/${this.deleteTarget.id}`);
    showToast('Snippet deleted');
    this.showDeleteModal = false;
    this.showEditor = false;
    this.deleteTarget = null;

    await Promise.all([
      this.loadSnippets(),
      this.loadTags(),
      this.loadFolders(),
      this.loadFavoritesCount()
    ]);

    this.totalSnippets = this.snippets.length;
  },

  async toggleFavorite(snippet) {
    const result = await api.post(`/api/v1/snippets/${snippet.id}/favorite`);
    if (result) {
      snippet.is_favorite = result.is_favorite;
      showToast(result.is_favorite ? 'Added to favorites' : 'Removed from favorites');
    }
  },

  async duplicateSnippet(snippet) {
    const result = await api.post(`/api/v1/snippets/${snippet.id}/duplicate`);
    if (result && !result.error) {
      showToast('Snippet duplicated');
      await this.loadSnippets();
    }
  },

  async copyToClipboard(snippet) {
    try {
      await navigator.clipboard.writeText(snippet.content);
      showToast('Copied to clipboard');
    } catch (err) {
      showToast('Failed to copy', 'error');
    }
  },

  async copyShareLink(snippet) {
    if (!snippet?.id || !snippet?.is_public) {
      showToast('Snippet must be public to share', 'warning');
      return;
    }
    try {
      const shareUrl = `${window.location.origin}/s/${snippet.id}`;
      await navigator.clipboard.writeText(shareUrl);
      showToast('Share link copied to clipboard');
    } catch (err) {
      showToast('Failed to copy link', 'error');
    }
  },

  async toggleArchive(snippet) {
    const result = await api.post(`/api/v1/snippets/${snippet.id}/archive`);
    if (result) {
      if (this.editingSnippet.id === snippet.id) {
        this.editingSnippet.is_archived = result.is_archived;
      }

      if (this.filter.isArchived !== null) {
        this.snippets = this.snippets.filter(s => s.id !== snippet.id);
      } else {
        const idx = this.snippets.findIndex(s => s.id === snippet.id);
        if (idx !== -1) {
          this.snippets[idx].is_archived = result.is_archived;
          if (result.is_archived && this.filter.isArchived === null) {
            this.snippets.splice(idx, 1);
          }
        }
      }

      showToast(result.is_archived ? 'Snippet archived' : 'Snippet unarchived');
    }
  },

  addTag(tag) {
    if (tag && !this.editingSnippet.tags.includes(tag)) {
      this.editingSnippet.tags.push(tag);
    }
  },

  removeTag(index) {
    this.editingSnippet.tags.splice(index, 1);
  },

  validateFilename() {
    if (this.editingSnippet.files && this.editingSnippet.files.length > 0) {
      const file = this.editingSnippet.files[this.activeFileIndex];
      if (!file.filename || !file.filename.trim()) {
        file.filename = 'untitled.txt';
        showToast('Filename cannot be empty', 'warning');
      }
    }
  },

  // Ace Editor management
  updateAceEditor() {
    const container = this.$refs.codeEditor;
    if (!container) return;

    const content = (this.editingSnippet.files && this.editingSnippet.files.length > 0)
      ? (this.activeFile?.content || '')
      : (this.editingSnippet.content || '');

    const language = (this.editingSnippet.files && this.editingSnippet.files.length > 0)
      ? (this.activeFile?.language || 'javascript')
      : (this.editingSnippet.language || 'javascript');

    const aceMode = getAceMode(language);
    const isDark = theme.get() === 'dark';
    const aceTheme = isDark ? 'ace/theme/monokai' : 'ace/theme/chrome';

    if (!this.aceEditor) {
      if (typeof ace === 'undefined') {
        console.error('Ace Editor not loaded');
        return;
      }

      try {
        ace.config.set('basePath', '/static/vendor/js/ace');

        if (!container.id) {
          container.id = 'ace-editor-' + Date.now();
        }

        this.aceEditor = ace.edit(container.id);
        this.aceEditor.setTheme(aceTheme);
        this.aceEditor.session.setMode(aceMode);
        this.aceEditor.setValue(content, -1);

        this.aceEditor.setOptions({
          fontSize: '14px',
          showPrintMargin: false,
          tabSize: 2,
          useSoftTabs: true,
          wrap: true
        });

        const self = this;
        this.aceEditor.session.on('change', () => {
          if (self.aceIgnoreChange) return;

          const value = self.aceEditor.getValue();
          if (self.editingSnippet.files && self.editingSnippet.files.length > 0) {
            self.updateActiveFileContent(value);
          } else {
            self.editingSnippet.content = value;
            self.scheduleAutoSave();
          }
        });
      } catch (e) {
        console.error('Ace Editor initialization error:', e);
        this.aceEditor = null;
        return;
      }
    } else {
      this.aceIgnoreChange = true;
      try {
        this.aceEditor.setValue(content, -1);
        this.aceEditor.session.setMode(aceMode);
        this.aceEditor.setTheme(aceTheme);
      } catch (e) {
        console.warn('Ace Editor update error:', e);
      }
      this.aceIgnoreChange = false;
    }

    this.$nextTick(() => {
      if (this.aceEditor) {
        this.aceEditor.resize();
      }
    });
  },

  destroyAceEditor() {
    if (this.aceEditor) {
      try {
        this.aceEditor.destroy();
        const container = this.$refs.codeEditor;
        if (container) {
          container.innerHTML = '';
          container.className = 'ace-editor-container';
        }
      } catch (e) {
        // Ignore errors during cleanup
      }
      this.aceEditor = null;
    }
  }
};
