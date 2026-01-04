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
  autoResizeTextarea
} from '../utils/helpers.js';
import { 
  getAceMode, 
  getFileExtension, 
  detectLanguageFromFilename 
} from '../utils/ace-utils.js';
import { theme } from '../modules/theme.js';

// Import component mixins
import { editorMixin } from './snippets/editor-mixin.js';
import { fileManagerMixin } from './snippets/file-manager-mixin.js';
import { foldersMixin } from './snippets/folders-mixin.js';
import { tagsMixin } from './snippets/tags-mixin.js';
import { backupMixin } from './snippets/backup-mixin.js';
import { historyMixin } from './snippets/history-mixin.js';
import { draftMixin } from './snippets/draft-mixin.js';
import { settingsMixin } from './snippets/settings-mixin.js';

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
      isArchived: null
    },
    
    pagination: { page: 1, limit: 20, total: 0, totalPages: 0 },
    totalSnippets: 0,
    favoritesCount: 0,
    loading: true,
    viewMode: localStorage.getItem('snipo-view-mode') || 'grid',
    sortBy: localStorage.getItem('snipo-sort-by') || 'updated_at',
    showEditor: false,
    isEditing: false,
    showDeleteModal: false,
    deleteTarget: null,
    showSearchHelp: false,
    
    foldersCollapsed: false,
    tagsCollapsed: false,
    
    aceEditor: null,
    aceIgnoreChange: false,
    settings: { archive_enabled: false, history_enabled: true },

    // Lifecycle
    async init() {
      await Promise.all([
        this.loadSnippets(),
        this.loadTags(),
        this.loadFolders(),
        this.loadFavoritesCount(),
        this.loadSettings()
      ]);
      
      this.totalSnippets = this.pagination.total;
      this.loading = false;
      this.$nextTick(() => highlightAll());

      await this.restoreFromUrl();
      window.addEventListener('popstate', () => this.restoreFromUrl());
      this.loadDraft();
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
      
      // Handle sorting
      if (this.sortBy) {
        if (this.sortBy === 'title_desc') {
          params.set('sort', 'title');
          params.set('order', 'desc');
        } else {
          params.set('sort', this.sortBy);
        }
      }
      
      if (this.filter.query) params.set('q', this.filter.query);
      if (this.filter.tagId) params.set('tag_id', this.filter.tagId);
      if (this.filter.folderId) params.set('folder_id', this.filter.folderId);
      if (this.filter.language) params.set('language', this.filter.language);
      if (this.filter.isFavorite !== null) params.set('favorite', this.filter.isFavorite);
      if (this.filter.isArchived !== null) params.set('is_archived', this.filter.isArchived);

      const result = await api.get(`/api/v1/snippets?${params}`);
      if (result) {
        this.snippets = result.data || [];
        this.pagination = result.pagination || this.pagination;
        this.$nextTick(() => highlightAll());
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
      if (result) {
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

    async loadFavoritesCount() {
      const result = await api.get('/api/v1/snippets?favorite=true&limit=1');
      if (result && result.pagination) {
        this.favoritesCount = result.pagination.total;
      }
    },

    async search() {
      this.pagination.page = 1;
      await this.loadSnippets();
    },

    async filterByTag(tagId) {
      this.showEditor = false;
      this.filter.tagId = tagId;
      this.filter.folderId = null;
      this.filter.isFavorite = null;
      this.pagination.page = 1;
      await this.loadSnippets();
      this.updateUrl({ tag: tagId });
    },

    async filterByFolder(folderId) {
      this.showEditor = false;
      this.filter.folderId = folderId;
      this.filter.tagId = null;
      this.filter.isFavorite = null;
      this.pagination.page = 1;
      await this.loadSnippets();
      this.updateUrl({ folder: folderId });
    },

    async clearFilters() {
      this.showEditor = false;
      this.filter = { query: '', tagId: null, folderId: null, language: '', isFavorite: null, isArchived: null };
      this.pagination.page = 1;
      await this.loadSnippets();
      this.totalSnippets = this.pagination.total;
      this.updateUrl({});
    },

    async showFavorites() {
      this.showEditor = false;
      this.filter.isFavorite = true;
      this.filter.tagId = null;
      this.filter.folderId = null;
      this.pagination.page = 1;
      await this.loadSnippets();
      this.updateUrl({ favorites: true });
    },

    async showArchive() {
      this.showEditor = false;
      this.filter.isArchived = true;
      this.filter.tagId = null;
      this.filter.folderId = null;
      this.filter.isFavorite = null;
      this.pagination.page = 1;
      await this.loadSnippets();
    },

    setViewMode(mode) {
      this.viewMode = mode;
      localStorage.setItem('snipo-view-mode', mode);
    },

    async setSortBy(sort) {
      this.sortBy = sort;
      localStorage.setItem('snipo-sort-by', sort);
      this.pagination.page = 1;
      await this.loadSnippets();
    },

    async logout() {
      try {
        await api.post('/api/v1/auth/logout');
      } catch (err) {
        console.error('logout failed:', err);
      }
      // Redirect to login page to ensure clean state
      // This works with auth proxies and handles the case when login is disabled
      window.location.href = '/login';
    },

    // URL routing
    updateUrl(params = {}) {
      const url = new URL(window.location);
      url.searchParams.delete('snippet');
      url.searchParams.delete('edit');
      url.searchParams.delete('folder');
      url.searchParams.delete('tag');
      url.searchParams.delete('favorites');

      if (params.snippet) url.searchParams.set('snippet', params.snippet);
      if (params.edit) url.searchParams.set('edit', 'true');
      if (params.folder) url.searchParams.set('folder', params.folder);
      if (params.tag) url.searchParams.set('tag', params.tag);
      if (params.favorites) url.searchParams.set('favorites', 'true');

      history.pushState({}, '', url);
    },

    async restoreFromUrl() {
      const params = new URLSearchParams(window.location.search);
      const snippetId = params.get('snippet');
      const isEdit = params.get('edit') === 'true';
      const folderId = params.get('folder');
      const tagId = params.get('tag');
      const favorites = params.get('favorites') === 'true';

      if (folderId) {
        this.filter.folderId = parseInt(folderId);
        this.filter.tagId = null;
        this.filter.isFavorite = null;
        await this.loadSnippets();
      } else if (tagId) {
        this.filter.tagId = parseInt(tagId);
        this.filter.folderId = null;
        this.filter.isFavorite = null;
        await this.loadSnippets();
      } else if (favorites) {
        this.filter.isFavorite = true;
        this.filter.tagId = null;
        this.filter.folderId = null;
        await this.loadSnippets();
      }

      if (snippetId) {
        const result = await api.get(`/api/v1/snippets/${snippetId}`);
        if (result && !result.error) {
          this.editingSnippet = {
            ...result,
            tags: (result.tags || []).map(t => t.name),
            folder_id: result.folders?.[0]?.id || null,
            files: result.files || []
          };
          this.activeFileIndex = 0;
          this.showEditor = true;
          this.isEditing = isEdit;
          this.$nextTick(() => {
            if (isEdit) this.updateAceEditor();
            highlightAll();
          });
        }
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
        let contentToCopy = snippet.content;
        
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
    autoResizeTextarea,
    getAceMode,
    getFileExtension,
    detectLanguageFromFilename,

    // Mix in other functionality
    ...editorMixin,
    ...fileManagerMixin,
    ...foldersMixin,
    ...tagsMixin,
    ...backupMixin,
    ...historyMixin,
    ...draftMixin,
    ...settingsMixin
  }));
}
