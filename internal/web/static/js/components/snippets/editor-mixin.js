// Editor mixin - handles snippet editing operations
import { api } from '../../modules/api.js';
import { showToast } from '../../modules/toast.js';
import { loadEditor, editorEngines } from '../../modules/editors/index.js';
import { theme } from '../../modules/theme.js';
import { isArabicText, isPredominantlyArabic, sanitizeHTML } from '../../utils/helpers.js';

export const editorMixin = {
  // Editor operations (imported from original app.js)
  // This file contains editor-related methods and state
  // Methods: viewSnippet, editSnippet, newSnippet, saveSnippet, startEditing, cancelEditing, etc.

  saveState: 'idle',
  saving: false,
  viewRequest: 0,

  editorLabel(engine) {
    return { ace: 'Ace', codemirror: 'CodeMirror 6', monaco: 'Monaco' }[engine] || engine;
  },

  preserveDraft() {
    if (this.isEditing) {
      clearTimeout(this.autoSaveTimeout);
      this.saveDraft();
      if (this.saveState === 'local') {
        this.hasDraft = true;
        this.draftSnippet = JSON.parse(JSON.stringify(this.editingSnippet));
      }
    }
  },

  async viewSnippet(snippet) {
    if (this.saving) return;
    this.preserveDraft();
    this.showSettings = false;
    const request = ++this.viewRequest;
    let result;
    try { result = await api.get(`/api/v1/snippets/${snippet.id}`); }
    catch { showToast('Could not open snippet. Try again.', 'error'); return; }
    if (result && !result.error && request === this.viewRequest) {
      this.destroyCodeEditor();
      this.saveState = 'saved';
      this.editingSnippet = {
        ...result,
        expires_at: this.resolveToShorthand(result.expires_at),
        tags: (result.tags || []).map(t => t.name),
        folder_id: result.folders?.[0]?.id || null,
        files: result.files || []
      };
      this.activeFileIndex = 0;
      this.showEditor = true;
      this.isEditing = false;
      this.updateUrl({ snippet: snippet.id });
      this.$nextTick(() => {
        this.highlightAll();
        this.updateTextDirection();
      });
    }
  },

  newSnippet() {
    if (this.saving) return;
    this.preserveDraft();
    ++this.viewRequest;
    this.destroyCodeEditor();
    this.saveState = 'idle';
    this.showSettings = false;
    // Calculate default expiration if auto-archive is enabled and default days is set
    let defaultExpires = '';
    if (this.settings?.auto_archive_enabled && this.settings?.default_expiration_days > 0) {
      const d = new Date();
      d.setDate(d.getDate() + this.settings.default_expiration_days);
      defaultExpires = d.toISOString().split('T')[0];
    }

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
      expires_at: defaultExpires,
      sync_to_gist: this.gistConfig?.auto_sync_enabled || false,
      files: [{
        id: 0,
        filename: 'snippet.txt',
        content: '',
        language: 'plaintext'
      }]
    };
    this.activeFileIndex = 0;
    this._resetFileManagerState();
    this.showEditor = true;
    this.isEditing = true;
    this.updateUrl({ edit: true });

    this.$nextTick(() => {
      this.updateCodeEditor();
      const titleInput = document.querySelector('.editor-title-input');
      if (titleInput) {
        titleInput.focus();
        titleInput.select();
      }
      this.updateTextDirection();
    });
  },

  async editSnippet(snippet) {
    const result = await api.get(`/api/v1/snippets/${snippet.id}`);
    if (result) {
      this.editingSnippet = {
        ...result,
        expires_at: this.resolveToShorthand(result.expires_at),
        tags: (result.tags || []).map(t => t.name),
        folder_id: result.folders?.[0]?.id || null,
        files: result.files || []
      };
      this.activeFileIndex = 0;
      this._resetFileManagerState();
      this._ensureEditableFiles();
      this.showEditor = true;
      this.isEditing = true;
      this.updateUrl({ snippet: snippet.id, edit: true });
      this.$nextTick(() => {
        this.updateCodeEditor();
        this.highlightAll();
        this.updateTextDirection();
      });
    }
  },

  startEditing() {
    this.isEditing = true;
    this._resetFileManagerState();
    this._ensureEditableFiles();
    if (this.editingSnippet?.id) {
      this.updateUrl({ snippet: this.editingSnippet.id, edit: true });
    }
    this.$nextTick(() => {
      this.updateCodeEditor();
      this.highlightAll();
      this.updateTextDirection();
    });
  },

  cancelEditing() {
    if (this.saving) return;
    this.preserveDraft();
    this.isEditing = false;
    if (this.editingSnippet?.id) {
      this.viewSnippet(this.editingSnippet);
    } else {
      this.showEditor = false;
      this.resetEditingSnippet();
      this.updateUrl({ snippet: null, edit: null });
    }
  },

  async saveSnippet() {
    if (this.saving) return;
    if (!this.editingSnippet.title?.trim()) {
      showToast('Give this snippet a title before saving.', 'warning');
      document.querySelector('.editor-title-input')?.focus();
      return;
    }
    this.saving = true;
    this.saveState = 'saving';
    clearTimeout(this.autoSaveTimeout);
    try {
      if (this.codeEditor) {
        if (this.editingSnippet.files?.length && this.syncCurrentContent) this.syncCurrentContent();
        else this.editingSnippet.content = this.codeEditor.getValue();
      }
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

      // Resolve expiration shorthand (7d, 30d, etc.) to actual date
      let expiresAt = this.editingSnippet.expires_at || null;
      if (expiresAt && typeof expiresAt === 'string' && expiresAt.endsWith('d')) {
        const days = parseInt(expiresAt, 10);
        const d = new Date();
        d.setDate(d.getDate() + days);
        expiresAt = d.toISOString().split('T')[0];
      }

      const data = {
        title: this.editingSnippet.title,
        description: this.editingSnippet.description || '',
        content: primaryContent,
        language: primaryLanguage,
        tags: this.editingSnippet.tags || [],
        folder_id: folderId,
        is_public: this.editingSnippet.is_public || false,
        is_archived: this.editingSnippet.is_archived || false,
        expires_at: expiresAt,
        files: files
      };

      let result;
      if (this.editingSnippet.id) {
        result = await api.put(`/api/v1/snippets/${this.editingSnippet.id}`, data);
      } else {
        result = await api.post('/api/v1/snippets', data);
      }

      if (result && !result.error) {
        const wasUpdate = !!this.editingSnippet.id;
        const snippetId = result.id || this.editingSnippet.id;
        const shouldSyncToGist = this.editingSnippet.sync_to_gist;

        showToast(wasUpdate ? 'Snippet updated' : 'Snippet created');
        this.isEditing = false;
        this.saveState = 'saved';
        this.destroyCodeEditor();
        this.clearDraft();
        this.hasDraft = false;
        this.draftSnippet = null;
        // Keep the saved snippet open, with the library context intact.
        this.editingSnippet = { ...this.editingSnippet, ...result,
          tags: (result.tags || []).map(t => typeof t === 'string' ? t : t.name),
          folder_id: result.folders?.[0]?.id || folderId };
        this.updateUrl({ snippet: snippetId });
        this.$nextTick(() => this.highlightAll());
        await Promise.allSettled([this.loadSnippets(), this.loadTags(), this.loadFavoritesCount(), this.loadTotalCount()]);

        // Auto-sync to gist if sync is enabled for this snippet
        if (wasUpdate && snippetId && this.isGistSyncEnabled && this.isGistSyncEnabled(snippetId)) {
          this.syncSnippetToGist(snippetId);
        } else if (!wasUpdate && snippetId && shouldSyncToGist && this.isGistConfigured && this.isGistConfigured()) {
          if (this.enableGistSyncForSnippet) {
            this.enableGistSyncForSnippet(snippetId);
          }
        }
      } else {
        throw new Error(result?.error?.message || 'Could not save. Check your connection and try again.');
      }
    } catch (error) {
      this.saveDraft();
      this.saveState = 'error';
      showToast(error.message || 'Could not save snippet', 'error');
    } finally {
      this.saving = false;
    }
  },

  cancelEdit() {
    if (this.saving) return;
    this.preserveDraft();
    ++this.viewRequest;
    this.showEditor = false;
    this.isEditing = false;
    this.destroyCodeEditor();
    this.resetEditingSnippet();
    this.editorExpanded = false;

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
      expires_at: '',
      sync_to_gist: false,
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
    this.deleteFromGist = false;
    this.showDeleteModal = true;
  },

  async deleteSnippet() {
    if (!this.deleteTarget) return;

    if (this.deleteFromGist) {
      const result = await api.delete(`/api/v1/gist/sync/snippet/${this.deleteTarget.id}`);
      if (result && result.error) {
        showToast('Failed to delete Gist from GitHub: ' + (result.error.message || 'Unknown error'), 'error');
        return; // Abort deleting the local snippet if gist deletion failed
      }
    }

    const permanent = this.filter?.isDeleted === true;
    const url = `/api/v1/snippets/${this.deleteTarget.id}` + (permanent ? '?permanent=true' : '');

    await api.delete(url);
    showToast(permanent ? 'Snippet permanently deleted' : 'Snippet moved to trash');
    this.showDeleteModal = false;
    this.showEditor = false;
    this.deleteTarget = null;
    // Don't nullify immediately to avoid Alpine errors during transition
    // this.editingSnippet = null; 
    setTimeout(() => {
      this.editingSnippet = { files: [], tags: [] };
    }, 100);

    await Promise.all([
      this.loadSnippets(),
      this.loadTags(),
      this.loadFolders(),
      this.loadFavoritesCount()
    ]);

    await this.loadTotalCount();
  },

  async restoreSnippet(snippet) {
    const result = await api.post(`/api/v1/snippets/${snippet.id}/restore`);
    if (result) {
      showToast('Snippet restored');
      await Promise.all([this.loadSnippets(), this.loadTotalCount()]);
    }
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
      await Promise.all([this.loadSnippets(), this.loadTotalCount()]);
    }
  },

  async copyToClipboard(snippet) {
    try {
      // Get content from either legacy content field or first file
      let contentToCopy = snippet === this.editingSnippet ? this.activeFile.content : snippet.content || '';
      if (snippet !== this.editingSnippet && !contentToCopy && snippet.files && snippet.files.length > 0) {
        contentToCopy = snippet.files[0].content || '';
      }

      // Check if setting to exclude first line is enabled
      if (this.settings?.exclude_first_line_on_copy) {
        const lines = contentToCopy.split('\n');
        if (lines.length > 1) {
          contentToCopy = lines.slice(1).join('\n');
        }
      }

      await navigator.clipboard.writeText(contentToCopy);
      showToast('Copied to clipboard');
    } catch (err) {
      showToast('Failed to copy', 'error');
    }
  },

  async copyAsRichText(snippet) {
    try {
      if (typeof Prism === 'undefined') {
        showToast('Syntax highlighter not available', 'error');
        return;
      }

      // Get content from either legacy content field or first file
      let content = snippet === this.editingSnippet ? this.activeFile.content : snippet.content || '';
      let language = snippet === this.editingSnippet ? this.activeFile.language : snippet.language || 'plaintext';

      if (snippet !== this.editingSnippet && !content && snippet.files && snippet.files.length > 0) {
        content = snippet.files[0].content || '';
        language = snippet.files[0].language || language;
      }

      // Check if setting to exclude first line is enabled
      if (this.settings?.exclude_first_line_on_copy) {
        const lines = content.split('\n');
        if (lines.length > 1) {
          content = lines.slice(1).join('\n');
        }
      }

      // Map language to Prism grammar
      const prismLangMap = {
        'javascript': 'javascript', 'js': 'javascript',
        'typescript': 'typescript', 'ts': 'typescript',
        'python': 'python', 'py': 'python',
        'go': 'go', 'golang': 'go',
        'bash': 'bash', 'sh': 'bash', 'shell': 'bash', 'zsh': 'bash',
        'powershell': 'powershell', 'ps': 'powershell', 'ps1': 'powershell',
        'sql': 'sql',
        'json': 'json',
        'yaml': 'yaml', 'yml': 'yaml',
        'markdown': 'markdown', 'md': 'markdown',
        'cuda': 'cuda', 'cu': 'cuda',
      };

      const prismLang = prismLangMap[language?.toLowerCase()] || null;
      const grammar = prismLang ? (Prism.languages[prismLang] || Prism.languages.plain) : Prism.languages.plain;
      const highlighted = Prism.highlight(content, grammar, prismLang || 'plain');

      // Build themed HTML with inline styles for clipboard compatibility
      const htmlContent = `<pre style="background:#1e1e1e;color:#d4d4d4;padding:1rem;border-radius:6px;font-family:'Fira Code','JetBrains Mono',Consolas,Monaco,monospace;font-size:13px;line-height:1.5;overflow-x:auto;tab-size:2;"><code class="language-${language || 'plaintext'}">${highlighted}</code></pre>`;

      // Write both HTML and plain text to clipboard
      await navigator.clipboard.write([
        new ClipboardItem({
          'text/html': new Blob([htmlContent], { type: 'text/html' }),
          'text/plain': new Blob([content], { type: 'text/plain' })
        })
      ]);
      showToast('Copied as rich text!');
    } catch (err) {
      console.error('Failed to copy as rich text:', err);
      showToast('Failed to copy as rich text', 'error');
    }
  },

  async copyShareLink(snippet) {
    if (!snippet?.id || !snippet?.is_public) {
      showToast('Snippet must be public to share', 'warning');
      return;
    }
    try {
      const shareUrl = `${window.location.origin}${window.SNIPO_CONFIG?.basePath || ''}/s/${snippet.id}`;
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
    const tagName = (tag || '').trim();
    if (!tagName) return;

    if (!this.editingSnippet.tags) {
      this.editingSnippet.tags = [];
    }

    if (!this.editingSnippet.tags.includes(tagName)) {
      this.editingSnippet.tags.push(tagName);
      this.scheduleAutoSave();
    }
  },

  removeTag(index) {
    this.editingSnippet.tags.splice(index, 1);
    this.scheduleAutoSave();
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

  async updateCodeEditor() {
    const container = this.$refs.codeEditor;
    if (!container || !this.isEditing) return;
    const request = ++this.editorRequest;
    this.editorLoading = true;
    const selected = this.editorEngine;
    let create;
    let active = selected;
    try {
      try {
        create = await loadEditor(selected);
      } catch (error) {
        if (request !== this.editorRequest || !this.isEditing) return;
        console.error('Editor loading failed:', error);
        active = 'ace';
        create = await loadEditor('ace');
        if (request === this.editorRequest) showToast('Could not load the selected editor. Using Ace; your content is safe.', 'error');
      }
      if (request !== this.editorRequest || !this.isEditing || container !== this.$refs.codeEditor) return;
      if (this.editingSnippet?.files?.length && !this.editingSnippet.files[this.activeFileIndex]) this.activeFileIndex = 0;
      const file = this.activeFile || this.editingSnippet;
      if (this.codeEditor) {
        this.editorIgnoreChange = true;
        this.codeEditor.setValue(file.content || '');
        this.codeEditor.setLanguage(file.language || 'plaintext');
      } else {
        this.activeEditorEngine = active;
        this.codeEditor = create(container, {
          content: file.content || '', language: file.language || 'plaintext',
          settings: this.settings, dark: theme.get() === 'dark',
          onSave: () => { if (this.isEditing) this.saveSnippet(); },
          onChange: value => {
            if (this.editorIgnoreChange) return;
            if (this.editingSnippet.files?.length) this.updateActiveFileContent(value);
            else { this.editingSnippet.content = value; this.scheduleAutoSave(); }
            this.updateTextDirection();
          }
        });
      }
      this.applyEditorSettings();
      this.updateTextDirection();
      this.$nextTick(() => this.codeEditor?.resize());
    } catch (error) {
      if (request !== this.editorRequest || !this.isEditing) return;
      console.error('Editor initialization failed:', error);
      this.destroyCodeEditor();
      this.editorIgnoreChange = false;
      showToast('Could not open the editor. Your snippet and draft are unchanged.', 'error');
    } finally {
      if (request === this.editorRequest) {
        this.editorIgnoreChange = false;
        this.editorLoading = false;
      }
    }
  },

  async changeEditor(engine) {
    if (!editorEngines.includes(engine)) return;
    if (this.codeEditor) {
      if (this.editingSnippet.files?.length) this.syncCurrentContent();
      else this.editingSnippet.content = this.codeEditor.getValue();
      this.saveDraft();
    }
    this.editorEngine = engine;
    try { localStorage.setItem('snipo-editor', engine); } catch { /* Session-only preference. */ }
    this.destroyCodeEditor();
    if (this.isEditing) await this.updateCodeEditor();
  },

  applyEditorSettings() {
    this.codeEditor?.configure(this.settings || {}, theme.get() === 'dark');
  },

  destroyCodeEditor() {
    ++this.editorRequest;
    this.editorLoading = false;
    this.codeEditor?.destroy();
    this.codeEditor = null;
    const container = this.$refs?.codeEditor;
    if (container) {
      container.replaceChildren();
      container.className = 'code-editor-container';
    }
  },

  // Arabic/RTL text detection and support
  updateTextDirection() {
    this.$nextTick(() => {
      const title = this.editingSnippet?.title || '';
      const description = this.editingSnippet?.description || '';
      const content = this.activeFile?.content ?? this.editingSnippet?.content ?? '';

      const isTitleArabic = isArabicText(title);
      const isDescriptionArabic = isArabicText(description);
      const isContentArabic = isArabicText(content);
      const isContentPredominantlyArabic = isPredominantlyArabic(content);

      this.codeEditor?.setDirection(isContentArabic, !isContentPredominantlyArabic);

      // Update title input direction
      const titleInput = document.querySelector('.editor-title-input');
      if (titleInput) {
        titleInput.classList.toggle('rtl', isTitleArabic);
        titleInput.classList.toggle('ltr', !isTitleArabic);
      }

      // Update title display direction
      const titleDisplay = document.querySelector('.editor-title-display');
      if (titleDisplay) {
        titleDisplay.classList.toggle('rtl', isTitleArabic);
        titleDisplay.classList.toggle('ltr', !isTitleArabic);
      }

      // Update description textarea direction
      const descriptionTextarea = document.querySelector('#snippet-description');
      if (descriptionTextarea) {
        descriptionTextarea.classList.toggle('rtl', isDescriptionArabic);
        descriptionTextarea.classList.toggle('ltr', !isDescriptionArabic);
      }

      // Update markdown preview direction
      const markdownPreview = document.querySelector('.preview-markdown-scroll');
      if (markdownPreview) {
        const isMarkdown = (this.activeFile?.language || this.editingSnippet?.language) === 'markdown';
        if (isMarkdown && isContentArabic) {
          markdownPreview.classList.add('rtl');
          // Add mixed class for better bidirectional handling
          if (isContentPredominantlyArabic) {
            markdownPreview.classList.remove('mixed');
          } else {
            markdownPreview.classList.add('mixed');
          }
        } else {
          markdownPreview.classList.remove('rtl');
          markdownPreview.classList.remove('mixed');
        }
      }

      // Update filename input direction (always LTR for filenames)
      const filenameInput = document.querySelector('.filename-input');
      if (filenameInput) {
        filenameInput.classList.add('ltr');
        filenameInput.classList.remove('rtl');
      }

      // Update tags inline direction
      const tagsInline = document.querySelector('.tags-inline');
      if (tagsInline) {
        tagsInline.classList.toggle('rtl', isTitleArabic || isDescriptionArabic);
      }
    });
  },

  renderMarkdown(content) {
    if (typeof marked === 'undefined' || !content) {
      return content || '';
    }

    try {
      const html = sanitizeHTML(marked.parse(content));

      // Check if content is Arabic and apply RTL class
      if (this.isArabicText(content)) {
        // Create a temporary div to parse the HTML
        const tempDiv = document.createElement('div');
        tempDiv.innerHTML = html;

        // Add RTL class to the markdown preview if content is Arabic
        this.$nextTick(() => {
          const preview = document.querySelector('.preview-markdown-scroll');
          if (preview) {
            preview.classList.add('rtl');
          }
        });
      }

      return html;
    } catch (e) {
      console.error('Error rendering markdown:', e);
      return content;
    }
  },

  // Resolve actual date to shorthand for select dropdown
  resolveToShorthand(expiresAt) {
    if (!expiresAt) return '';
    // Already a shorthand value
    if (typeof expiresAt === 'string' && expiresAt.endsWith('d')) return expiresAt;
    // Not a valid date string (e.g., random string) — return as-is
    if (isNaN(Date.parse(expiresAt))) return '';
    const shorthandDays = [7, 30, 90, 180, 365];
    const now = new Date();
    now.setHours(0, 0, 0, 0);
    const exp = new Date(expiresAt);
    exp.setHours(0, 0, 0, 0);
    for (const days of shorthandDays) {
      const target = new Date();
      target.setDate(target.getDate() + days);
      target.setHours(0, 0, 0, 0);
      if (exp.getTime() === target.getTime()) {
        return `${days}d`;
      }
    }
    // Not a match — return empty for custom
    return '';
  },

  // Expiration helpers
  todayDate() {
    const d = new Date();
    return d.toISOString().split('T')[0];
  },

  setExpirationDays(days) {
    const d = new Date();
    d.setDate(d.getDate() + days);
    this.editingSnippet.expires_at = d.toISOString().split('T')[0];
  },

  expirationMatchesDays(expiresAt, days) {
    if (!expiresAt || expiresAt === '') return false;
    // Handle shorthand values
    if (expiresAt.endsWith('d')) {
      return parseInt(expiresAt, 10) === days;
    }
    const target = new Date();
    target.setDate(target.getDate() + days);
    target.setHours(0, 0, 0, 0);
    const exp = new Date(expiresAt);
    exp.setHours(0, 0, 0, 0);
    return exp.getTime() === target.getTime();
  },

  // Resolve shorthand to display label
  resolveExpirationLabel(expiresAt) {
    if (!expiresAt || expiresAt === '') return 'Never';
    if (expiresAt.endsWith('d')) {
      const days = parseInt(expiresAt, 10);
      if (days === 180) return '6 months';
      if (days === 365) return '1 year';
      return `${days} days`;
    }
    const exp = new Date(expiresAt);
    const now = new Date();
    now.setHours(0, 0, 0, 0);
    exp.setHours(0, 0, 0, 0);
    const diffDays = Math.ceil((exp - now) / (1000 * 60 * 60 * 24));
    if (diffDays < 0) return 'Expired';
    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Tomorrow';
    if (diffDays <= 30) return `${diffDays} days`;
    return exp.toLocaleDateString();
  },

  isExpired(expiresAt) {
    if (!expiresAt || expiresAt === '') return false;
    if (expiresAt.endsWith('d')) {
      return false;
    }
    const exp = new Date(expiresAt);
    const now = new Date();
    now.setHours(0, 0, 0, 0);
    return exp < now;
  },

  formatExpiration(expiresAt) {
    return this.resolveExpirationLabel(expiresAt);
  }
};
