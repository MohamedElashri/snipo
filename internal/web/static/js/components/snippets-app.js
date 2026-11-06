import { readEditorPreference } from '../modules/editors/index.js';
// Snippets App Alpine component
import { api } from '../modules/api.js';
import { showToast } from '../modules/toast.js';
import {
  highlightAll,
  highlightCode,
  renderMarkdown,
  formatDate,
  formatFileSize,
  getLanguageColor,
  autoResizeInput,
  autoResizeTextarea,
  isArabicText
} from '../utils/helpers.js';
import {
  getFileExtension,
  detectLanguageFromFilename
} from '../utils/ace-utils.js';
import { theme } from '../modules/theme.js';

// Import component mixins
import { sessionMixin } from './snippets/session-mixin.js';
import { workspaceMixin } from './snippets/workspace-mixin.js';
import { editorMixin } from './snippets/editor-mixin.js';
import { fileManagerMixin } from './snippets/file-manager-mixin.js';
import { foldersMixin } from './snippets/folders-mixin.js';
import { tagsMixin } from './snippets/tags-mixin.js';
import { backupMixin } from './snippets/backup-mixin.js';
import { historyMixin } from './snippets/history-mixin.js';
import { draftMixin } from './snippets/draft-mixin.js';
import { settingsMixin } from './snippets/settings-mixin.js';
import { gistSyncMixin } from './snippets/gist-sync-mixin.js';

export function initSnippetsApp(Alpine) {
  Alpine.data('snippetsApp', () => ({
    // State
    snippets: [],
    tags: [],
    folders: [],
    selectedSnippet: null,
    editingSnippet: {
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
    },
    activeFileIndex: 0,
    editorHeaderVisible: true,

    fileManagerState: {
      operationInProgress: false,
      editorDirty: false,
      lastSyncedContent: '',
      pendingOperation: null
    },

    filter: {
      query: '',
      tagId: null,
      folderId: null,
      language: '',
      isFavorite: null,
      isArchived: null,
      isDeleted: null
    },

    pagination: { page: 1, limit: 20, total: 0, totalPages: 0 },
    totalSnippets: 0,
    favoritesCount: 0,
    loading: true,
    loadError: '',
    listRequest: 0,
    restoringUrl: false,
    viewMode: localStorage.getItem('snipo-view-mode') === 'list' ? 'list' : 'grid',
    density: localStorage.getItem('snipo-density') === 'compact' ? 'compact' : 'comfortable',
    sortBy: localStorage.getItem('snipo-sort-by') || 'updated_at',
    showEditor: false,
    isEditing: false,
    showDeleteModal: false,
    deleteTarget: null,
    deleteFromGist: false,
    showSearchHelp: false,

    foldersCollapsed: false,
    tagsCollapsed: false,

    codeEditor: null,
    editorEngine: readEditorPreference(),
    activeEditorEngine: 'ace',
    editorRequest: 0,
    editorLoading: false,
    editorIgnoreChange: false,
    settings: { archive_enabled: false, history_enabled: true, trash_enabled: true },

    // Lifecycle
    async init() {
      this.initSessionGuard();
      this._editorTheme = () => this.applyEditorSettings();
      window.addEventListener('snipo-theme-change', this._editorTheme);
      this._beforeUnload = () => this.saveDraft();
      window.addEventListener('beforeunload', this._beforeUnload);
      this.$nextTick(() => {
        this._editorResize = new ResizeObserver(() => this.codeEditor?.resize());
        const editor = this.$root.querySelector('.editor-container');
        if (editor) this._editorResize.observe(editor);
      });
      await Promise.all([
        this.loadTotalCount(),
        this.loadTags(),
        this.loadFolders(),
        this.loadFavoritesCount(),
        this.loadSettings(),
        this.loadGistConfig(),
        this.loadGistMappings()
      ]);

      this.loading = false;
      this.$nextTick(() => highlightAll());

      await this.restoreFromUrl();
      this._popstate = () => this.restoreFromUrl();
      window.addEventListener('popstate', this._popstate);
      this.loadDraft();

      // Verify gist mappings in background (removes stale mappings for deleted gists)
      this.verifyGistMappings();
    },

    destroy() {
      this.destroySessionGuard();
      this._editorResize?.disconnect();
      window.removeEventListener('snipo-theme-change', this._editorTheme);
      window.removeEventListener('beforeunload', this._beforeUnload);
      window.removeEventListener('popstate', this._popstate);
      clearTimeout(this.autoSaveTimeout);
      this.destroyCodeEditor();
    },

    // Sidebar resize functionality
    initSidebarResize() {
      const savedWidth = localStorage.getItem('sidebarWidth');
      if (savedWidth) {
        document.documentElement.style.setProperty('--sidebar-width', savedWidth + 'px');
      }
    },

    startSidebarResize(event) {
      event.preventDefault();
      const sidebar = event.target.closest('.sidebar');
      const handle = event.target;
      handle.classList.add('resizing');

      const startX = event.clientX;
      const startWidth = sidebar.offsetWidth;
      const minWidth = 180;
      const maxWidth = 500;

      const onMouseMove = (e) => {
        const newWidth = Math.min(maxWidth, Math.max(minWidth, startWidth + e.clientX - startX));
        document.documentElement.style.setProperty('--sidebar-width', newWidth + 'px');
      };

      const onMouseUp = () => {
        handle.classList.remove('resizing');
        const finalWidth = sidebar.offsetWidth;
        localStorage.setItem('sidebarWidth', finalWidth);
        document.removeEventListener('mousemove', onMouseMove);
        document.removeEventListener('mouseup', onMouseUp);
      };

      document.addEventListener('mousemove', onMouseMove);
      document.addEventListener('mouseup', onMouseUp);
    },

    // Core methods
    async loadSnippets() {
      const params = new URLSearchParams();
      params.set('page', this.pagination.page);
      params.set('limit', this.pagination.limit);
      params.set('view', 'summary');

      // Handle sorting
      if (this.sortBy) {
        if (this.sortBy === 'title_desc') {
          params.set('sort', 'title');
          params.set('order', 'desc');
        } else if (this.sortBy === 'title') {
          params.set('sort', 'title');
          params.set('order', 'asc');
        } else {
          params.set('sort', this.sortBy);
        }
      }

      if (this.filter.query) params.set('q', this.filter.query);
      if (this.filter.tagId) params.set('tag_ids', this.filter.tagId);
      if (this.filter.folderId) params.set('folder_ids', this.filter.folderId);
      if (this.filter.language) params.set('language', this.filter.language);
      if (this.filter.isFavorite !== null) params.set('favorite', this.filter.isFavorite);
      if (this.filter.isArchived !== null) params.set('is_archived', this.filter.isArchived);
      if (this.filter.isDeleted !== null) params.set('is_deleted', this.filter.isDeleted);

      const request = ++this.listRequest;
      this.loading = true;
      this.loadError = '';
      try {
        const result = await api.get(`/api/v1/snippets?${params}`);
        if (request !== this.listRequest) return;
        if (!result || result.error) throw new Error(result?.error?.message || 'Unable to load snippets.');
        if (result) {
          this.snippets = result.data || [];
          this.pagination = result.pagination || this.pagination;
        }
      } catch (error) {
        if (request === this.listRequest) this.loadError = error.message || 'Unable to load snippets.';
      } finally {
        if (request === this.listRequest) this.loading = false;
      }
    },

    async loadTags() {
      const result = await api.get('/api/v1/tags');
      if (result) this.tags = result || [];
    },

    async loadFolders() {
      const result = await api.get('/api/v1/folders?tree=true');
      if (result) this.folders = result || [];
    },

    async loadSettings() {
      const result = await api.get('/api/v1/settings');
      if (result && !result.error) {
        this.settings = result;
        // Cache settings for theme updates
        try {
          sessionStorage.setItem('snipo-settings', JSON.stringify(result));
        } catch (e) {
          // Ignore storage errors
        }
        // Apply markdown font size
        this.applyMarkdownFontSize();
        // Apply custom CSS if it exists
        if (result.custom_css) {
          theme.applyCustomCSS(result.custom_css);
        }
      }
    },

    async loadTotalCount() {
      const result = await api.get('/api/v1/snippets?limit=1&view=summary');
      if (result?.pagination) this.totalSnippets = result.pagination.total;
    },

    async loadFavoritesCount() {
      const result = await api.get('/api/v1/snippets?favorite=true&limit=1&view=summary');
      if (result && result.pagination) {
        this.favoritesCount = result.pagination.total;
      }
    },

    async search() {
      if (this.showEditor && window.matchMedia?.('(max-width: 900px)').matches) this.cancelEdit();
      this.pagination.page = 1;
      await this.loadSnippets();
      this.updateUrl(this.showEditor ? { snippet: this.editingSnippet?.id, edit: this.isEditing } : {});
    },

    async filterByTag(tagId) {
      if (!this.prepareLibraryNavigation()) return;
      this.filter.isArchived = null;
      this.filter.isDeleted = null;
      this.filter.tagId = tagId;
      this.filter.folderId = null;
      this.filter.isFavorite = null;
      this.pagination.page = 1;
      await this.loadSnippets();
      this.updateUrl({ tag: tagId });
    },

    async filterByFolder(folderId) {
      if (!this.prepareLibraryNavigation()) return;
      this.filter.isArchived = null;
      this.filter.isDeleted = null;
      this.filter.folderId = folderId;
      this.filter.tagId = null;
      this.filter.isFavorite = null;
      this.pagination.page = 1;
      await this.loadSnippets();
      this.updateUrl({ folder: folderId });
    },

    async clearFilters() {
      if (!this.prepareLibraryNavigation()) return;
      this.filter = { query: '', tagId: null, folderId: null, language: '', isFavorite: null, isArchived: null, isDeleted: null };
      this.pagination.page = 1;
      await this.loadSnippets();
      this.totalSnippets = this.pagination.total;
      this.updateUrl({});
    },

    async showFavorites() {
      if (!this.prepareLibraryNavigation()) return;
      this.filter.isArchived = null;
      this.filter.isDeleted = null;
      this.filter.isFavorite = true;
      this.filter.tagId = null;
      this.filter.folderId = null;
      this.pagination.page = 1;
      await this.loadSnippets();
      this.updateUrl({ favorites: true });
    },

    async showArchive() {
      if (!this.prepareLibraryNavigation()) return;
      this.filter.isArchived = true;
      this.filter.isDeleted = null;
      this.filter.tagId = null;
      this.filter.folderId = null;
      this.filter.isFavorite = null;
      this.pagination.page = 1;
      await this.loadSnippets();
      this.updateUrl();
    },

    async showTrash() {
      if (!this.prepareLibraryNavigation()) return;
      this.filter.isDeleted = true;
      this.filter.isArchived = null;
      this.filter.tagId = null;
      this.filter.folderId = null;
      this.filter.isFavorite = null;
      this.pagination.page = 1;
      await this.loadSnippets();
      this.updateUrl({ trash: true });
    },

    setViewMode(mode) {
      if (!['grid', 'list'].includes(mode)) return;
      this.viewMode = mode;
      localStorage.setItem('snipo-view-mode', mode);
    },

    async setSortBy(sort) {
      this.sortBy = sort;
      localStorage.setItem('snipo-sort-by', sort);
      this.pagination.page = 1;
      await this.loadSnippets();
    },

    // URL routing
    updateUrl(params = {}) {
      if (this.restoringUrl) return;
      const url = new URL(window.location);
      const values = {
        q: this.filter.query, folder: this.filter.folderId, tag: this.filter.tagId,
        language: this.filter.language, favorites: this.filter.isFavorite,
        archive: this.filter.isArchived, trash: this.filter.isDeleted,
        settings: this.showSettings ? this.settingsTab : null, ...params
      };
      for (const key of ['snippet', 'edit', 'folder', 'tag', 'favorites', 'trash', 'archive', 'q', 'language', 'settings']) {
        url.searchParams.delete(key);
        if (values[key]) url.searchParams.set(key, values[key]);
      }
      if (url.href !== window.location.href) history.pushState({}, '', url);
    },

    async restoreFromUrl() {
      this.restoringUrl = true;
      try {
        this.preserveDraft();
        this.destroyCodeEditor();
        const params = new URLSearchParams(window.location.search);
        this.filter = {
          query: params.get('q') || '', language: params.get('language') || '',
          folderId: parseInt(params.get('folder'), 10) || null,
          tagId: parseInt(params.get('tag'), 10) || null,
          isFavorite: params.get('favorites') === 'true' || null,
          isArchived: params.get('archive') === 'true' || null,
          isDeleted: params.get('trash') === 'true' || null
        };
        this.showEditor = false;
        this.isEditing = false;
        this.showSettings = false;
        this.loadDraft();
        await this.loadSnippets();
        const snippetId = params.get('snippet');
        const editing = params.get('edit') === 'true';
        if (editing && this.draftSnippet && (this.draftSnippet.id || null) === snippetId) {
          this.restoreDraft();
        } else if (snippetId && /^[a-zA-Z0-9_-]+$/.test(snippetId)) {
          await this.viewSnippet({ id: snippetId });
          if (this.showEditor && params.get('edit') === 'true') this.startEditing();
        } else if (params.get('edit') === 'true') this.newSnippet();
        if (params.has('settings')) await this.openSettings(params.get('settings'));
      } finally {
        this.restoringUrl = false;
      }
    },

    // Computed
    get activeFile() {
      const files = this.editingSnippet?.files || [];
      if (files.length === 0) {
        return {
          id: 0,
          filename: 'main',
          content: this.editingSnippet?.content || '',
          language: this.editingSnippet?.language || 'javascript'
        };
      }
      return files[this.activeFileIndex] || files[0];
    },

    get hasMultipleFiles() {
      return (this.editingSnippet?.files || []).length > 1;
    },

    get flattenedFolders() {
      const result = [];
      const flatten = (folders, depth = 0) => {
        for (const folder of folders) {
          result.push({
            id: folder.id,
            name: folder.name,
            displayName: '  '.repeat(depth) + (depth > 0 ? '└ ' : '') + folder.name
          });
          if (folder.children && folder.children.length > 0) {
            flatten(folder.children, depth + 1);
          }
        }
      };
      flatten(this.folders);
      return result;
    },

    // Helper methods for snippet cards
    async copySnippetFromCard(snippet) {
      try {
        if (snippet.is_summary) {
          snippet = await api.get(`/api/v1/snippets/${snippet.id}`);
          if (!snippet || snippet.error) throw new Error('Could not load snippet');
        }
        // Get content from either legacy content field or first file
        let contentToCopy = snippet.content || '';
        if (!contentToCopy && snippet.files && snippet.files.length > 0) {
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

    async copyFileFromCard(snippet, file) {
      try {
        if (snippet.is_summary) {
          const full = await api.get(`/api/v1/snippets/${snippet.id}`);
          if (!full || full.error) throw new Error('Could not load snippet');
          file = full.files?.find(candidate => candidate.id === file.id);
          if (!file) throw new Error('File no longer exists');
        }
        let contentToCopy = file.content;

        // Check if setting to exclude first line is enabled
        if (this.settings?.exclude_first_line_on_copy) {
          const lines = contentToCopy.split('\n');
          if (lines.length > 1) {
            contentToCopy = lines.slice(1).join('\n');
          }
        }

        await navigator.clipboard.writeText(contentToCopy);
        showToast(`Copied ${file.filename} to clipboard`);
      } catch (err) {
        showToast('Failed to copy', 'error');
      }
    },

    // Utility methods exposed to component
    highlightAll,
    highlightCode,
    renderMarkdown,
    formatDate,
    formatFileSize,
    getLanguageColor,
    autoResizeInput,
    autoResizeTextarea,
    getFileExtension,
    detectLanguageFromFilename,
    isArabicText,

    // Mix in other functionality
    ...sessionMixin,
    ...workspaceMixin,
    ...editorMixin,
    ...fileManagerMixin,
    ...foldersMixin,
    ...tagsMixin,
    ...backupMixin,
    ...historyMixin,
    ...draftMixin,
    ...settingsMixin,
    ...gistSyncMixin
  }));
}
