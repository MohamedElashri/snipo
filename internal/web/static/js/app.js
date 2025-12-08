// Snipo - Alpine.js Application

// Theme management
const theme = {
  get() {
    return localStorage.getItem('snipo-theme') || 'dark';
  },
  set(value) {
    localStorage.setItem('snipo-theme', value);
    document.documentElement.setAttribute('data-theme', value);
    // Update Prism theme
    this.updatePrismTheme(value);
  },
  toggle() {
    const current = this.get();
    this.set(current === 'dark' ? 'light' : 'dark');
  },
  init() {
    const saved = this.get();
    document.documentElement.setAttribute('data-theme', saved);
    this.updatePrismTheme(saved);
  },
  updatePrismTheme(themeName) {
    const prismLink = document.getElementById('prism-theme');
    if (prismLink) {
      if (themeName === 'dark') {
        prismLink.href = 'https://cdn.jsdelivr.net/npm/prismjs@1.29.0/themes/prism-tomorrow.min.css';
      } else {
        prismLink.href = 'https://cdn.jsdelivr.net/npm/prismjs@1.29.0/themes/prism.min.css';
      }
    }
  }
};

// Initialize theme on load
theme.init();

// API helper
const api = {
  async request(method, url, data = null) {
    const options = {
      method,
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include'
    };
    if (data) options.body = JSON.stringify(data);
    
    const response = await fetch(url, options);
    if (response.status === 401) {
      window.location.href = '/login';
      return null;
    }
    if (response.status === 204) return null;
    return response.json();
  },
  
  get: (url) => api.request('GET', url),
  post: (url, data) => api.request('POST', url, data),
  put: (url, data) => api.request('PUT', url, data),
  delete: (url) => api.request('DELETE', url)
};

// Toast notifications
function showToast(message, type = 'success') {
  const container = document.getElementById('toast-container');
  if (!container) return;
  
  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  toast.innerHTML = `
    <span>${message}</span>
    <button onclick="this.parentElement.remove()" style="background:none;border:none;cursor:pointer;padding:0;margin-left:0.5rem;">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="18" y1="6" x2="6" y2="18"></line>
        <line x1="6" y1="6" x2="18" y2="18"></line>
      </svg>
    </button>
  `;
  container.appendChild(toast);
  setTimeout(() => toast.remove(), 5000);
}

// Main app store
document.addEventListener('alpine:init', () => {
  // Global app state
  Alpine.store('app', {
    sidebarOpen: window.innerWidth > 768,
    currentView: 'snippets', // snippets, editor, settings
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
  
  // Snippets data
  Alpine.data('snippetsApp', () => ({
    snippets: [],
    tags: [],
    folders: [],
    selectedSnippet: null,
    // Initialize with empty object to prevent null access errors
    editingSnippet: {
      id: null,
      title: '',
      description: '',
      content: '',
      language: 'javascript',
      tags: [],
      folder_id: null,
      is_public: false,
      is_favorite: false
    },
    filter: {
      query: '',
      tagId: null,
      folderId: null,
      language: '',
      isFavorite: null
    },
    pagination: { page: 1, limit: 20, total: 0, totalPages: 0 },
    totalSnippets: 0, // Total count for "All Snippets" (unfiltered)
    favoritesCount: 0, // Count of favorite snippets
    loading: true,
    showEditor: false,
    isEditing: false, // false = preview mode, true = edit mode
    showDeleteModal: false,
    deleteTarget: null,
    showSettings: false,
    settingsTab: 'password', // 'password' or 'apikeys'
    apiTokens: [],
    newToken: { name: '', permissions: 'read', expires_in_days: 30 },
    createdToken: null, // Stores newly created token for display
    passwordForm: { current: '', new: '', confirm: '' },
    passwordError: '',
    passwordSuccess: '',
    // Sidebar collapse state
    foldersCollapsed: false,
    tagsCollapsed: false,
    // Folder/Tag management
    showFolderModal: false,
    showTagModal: false,
    editingFolder: { name: '', parent_id: '' },
    editingTag: { id: null, name: '' },
    // Draft auto-save
    hasDraft: false,
    draftSnippet: null,
    draftSavedAt: null,
    autoSaveTimeout: null,
    
    async init() {
      await Promise.all([
        this.loadSnippets(),
        this.loadTags(),
        this.loadFolders(),
        this.loadFavoritesCount()
      ]);
      // Store total snippets count on initial load
      this.totalSnippets = this.pagination.total;
      this.loading = false;
      // Highlight code after initial load
      this.$nextTick(() => this.highlightAll());
      
      // Restore state from URL
      await this.restoreFromUrl();
      
      // Handle browser back/forward
      window.addEventListener('popstate', () => this.restoreFromUrl());
      
      // Load draft if exists
      this.loadDraft();
    },
    
    // URL routing
    updateUrl(params = {}) {
      const url = new URL(window.location);
      
      // Clear existing params
      url.searchParams.delete('snippet');
      url.searchParams.delete('edit');
      url.searchParams.delete('folder');
      url.searchParams.delete('tag');
      url.searchParams.delete('favorites');
      
      // Set new params
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
      
      // Restore filter state
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
      
      // Restore snippet view/edit state
      if (snippetId) {
        const result = await api.get(`/api/v1/snippets/${snippetId}`);
        if (result && !result.error) {
          this.editingSnippet = {
            ...result,
            tags: (result.tags || []).map(t => t.name),
            folder_id: result.folders?.[0]?.id || null
          };
          this.showEditor = true;
          this.isEditing = isEdit;
          this.$nextTick(() => this.highlightAll());
        }
      }
    },
    
    async loadFavoritesCount() {
      const result = await api.get('/api/v1/snippets?is_favorite=true&limit=1');
      if (result && result.pagination) {
        this.favoritesCount = result.pagination.total;
      }
    },
    
    highlightAll() {
      // Re-run Prism highlighting on all code blocks
      if (typeof Prism !== 'undefined') {
        Prism.highlightAll();
      }
    },
    
    async loadSnippets() {
      const params = new URLSearchParams();
      params.set('page', this.pagination.page);
      params.set('limit', this.pagination.limit);
      if (this.filter.query) params.set('q', this.filter.query);
      if (this.filter.tagId) params.set('tag_id', this.filter.tagId);
      if (this.filter.folderId) params.set('folder_id', this.filter.folderId);
      if (this.filter.language) params.set('language', this.filter.language);
      if (this.filter.isFavorite !== null) params.set('favorite', this.filter.isFavorite);
      
      const result = await api.get(`/api/v1/snippets?${params}`);
      if (result) {
        this.snippets = result.data || [];
        this.pagination = result.pagination || this.pagination;
        // Highlight code after snippets load
        this.$nextTick(() => this.highlightAll());
      }
    },
    
    async loadTags() {
      const result = await api.get('/api/v1/tags');
      if (result) this.tags = result.data || [];
    },
    
    async loadFolders() {
      const result = await api.get('/api/v1/folders?tree=true');
      if (result) this.folders = result.data || [];
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
      this.filter = { query: '', tagId: null, folderId: null, language: '', isFavorite: null };
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
    
    newSnippet() {
      this.editingSnippet = {
        id: null,
        title: '',
        description: '',
        content: '',
        language: 'javascript',
        tags: [],
        folder_id: null,
        is_public: false,
        is_favorite: false
      };
      this.showEditor = true;
      this.isEditing = true;
      this.updateUrl({ edit: true });
    },
    
    async viewSnippet(snippet) {
      const result = await api.get(`/api/v1/snippets/${snippet.id}`);
      if (result) {
        this.editingSnippet = {
          ...result,
          tags: (result.tags || []).map(t => t.name),
          folder_id: result.folders?.[0]?.id || null
        };
        this.showEditor = true;
        this.isEditing = false;
        this.updateUrl({ snippet: snippet.id });
        this.$nextTick(() => this.highlightAll());
      }
    },
    
    startEditing() {
      this.isEditing = true;
      if (this.editingSnippet?.id) {
        this.updateUrl({ snippet: this.editingSnippet.id, edit: true });
      }
      this.$nextTick(() => this.highlightAll());
    },
    
    async editSnippet(snippet) {
      const result = await api.get(`/api/v1/snippets/${snippet.id}`);
      if (result) {
        this.editingSnippet = {
          ...result,
          tags: (result.tags || []).map(t => t.name),
          folder_id: result.folders?.[0]?.id || null
        };
        this.showEditor = true;
        this.isEditing = true;
        this.updateUrl({ snippet: snippet.id, edit: true });
        // Trigger syntax highlighting after editor opens
        this.$nextTick(() => this.highlightAll());
      }
    },
    
    async saveSnippet() {
      // Convert empty string folder_id to null, and string numbers to integers
      let folderId = this.editingSnippet.folder_id;
      if (folderId === '' || folderId === null || folderId === undefined) {
        folderId = null;
      } else {
        folderId = parseInt(folderId, 10) || null;
      }
      
      const data = {
        title: this.editingSnippet.title,
        description: this.editingSnippet.description || '',
        content: this.editingSnippet.content,
        language: this.editingSnippet.language,
        tags: this.editingSnippet.tags || [],
        folder_id: folderId,
        is_public: this.editingSnippet.is_public || false
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
        this.resetEditingSnippet();
        this.clearDraft(); // Clear draft on successful save
        this.updateUrl({}); // Clear URL params
        await this.loadSnippets();
        await this.loadTags();
        await this.loadFavoritesCount();
      } else if (result?.error) {
        showToast(result.error.message || 'Error saving snippet', 'error');
      }
    },
    
    cancelEdit() {
      this.showEditor = false;
      this.resetEditingSnippet();
      this.clearDraft();
      // Restore URL to current filter state
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
        language: 'javascript',
        tags: [],
        folder_id: null,
        is_public: false,
        is_favorite: false
      };
    },
    
    // Draft auto-save functionality
    saveDraft() {
      if (!this.isEditing) return;
      
      const draft = {
        snippet: this.editingSnippet,
        savedAt: new Date().toISOString()
      };
      localStorage.setItem('snipo-draft', JSON.stringify(draft));
    },
    
    loadDraft() {
      try {
        const draftJson = localStorage.getItem('snipo-draft');
        if (!draftJson) return;
        
        const draft = JSON.parse(draftJson);
        if (!draft.snippet) return;
        
        // Check if draft is less than 24 hours old
        const savedAt = new Date(draft.savedAt);
        const now = new Date();
        const hoursDiff = (now - savedAt) / (1000 * 60 * 60);
        
        if (hoursDiff > 24) {
          this.clearDraft();
          return;
        }
        
        // Only restore if we're not already viewing a snippet from URL
        if (this.showEditor) return;
        
        // Show draft recovery prompt
        if (draft.snippet.content || draft.snippet.title) {
          this.hasDraft = true;
          this.draftSnippet = draft.snippet;
          this.draftSavedAt = savedAt;
        }
      } catch (e) {
        this.clearDraft();
      }
    },
    
    restoreDraft() {
      if (this.draftSnippet) {
        this.editingSnippet = { ...this.draftSnippet };
        this.showEditor = true;
        this.isEditing = true;
        this.hasDraft = false;
        showToast('Draft restored');
        this.$nextTick(() => this.highlightAll());
      }
    },
    
    discardDraft() {
      this.clearDraft();
      this.hasDraft = false;
      this.draftSnippet = null;
    },
    
    clearDraft() {
      localStorage.removeItem('snipo-draft');
    },
    
    // Auto-save on content change (debounced)
    scheduleAutoSave() {
      if (this.autoSaveTimeout) {
        clearTimeout(this.autoSaveTimeout);
      }
      this.autoSaveTimeout = setTimeout(() => {
        this.saveDraft();
      }, 2000); // Save after 2 seconds of inactivity
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
      this.deleteTarget = null;
      await this.loadSnippets();
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
    
    formatDate(dateStr) {
      const date = new Date(dateStr);
      const now = new Date();
      const diff = now - date;
      
      if (diff < 60000) return 'Just now';
      if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
      if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
      if (diff < 604800000) return `${Math.floor(diff / 86400000)}d ago`;
      
      return date.toLocaleDateString();
    },
    
    getLanguageColor(lang) {
      const colors = {
        javascript: '#f7df1e',
        typescript: '#3178c6',
        python: '#3776ab',
        go: '#00add8',
        rust: '#dea584',
        java: '#b07219',
        csharp: '#178600',
        cpp: '#f34b7d',
        ruby: '#cc342d',
        php: '#4f5d95',
        swift: '#fa7343',
        kotlin: '#a97bff',
        html: '#e34c26',
        css: '#563d7c',
        sql: '#e38c00',
        bash: '#89e051',
        json: '#292929',
        yaml: '#cb171e',
        markdown: '#083fa1',
        plaintext: '#6b7280'
      };
      return colors[lang] || '#6b7280';
    },
    
    addTag(tag) {
      if (tag && !this.editingSnippet.tags.includes(tag)) {
        this.editingSnippet.tags.push(tag);
      }
    },
    
    removeTag(index) {
      this.editingSnippet.tags.splice(index, 1);
    },
    
    async logout() {
      await api.post('/api/v1/auth/logout');
      window.location.href = '/login';
    },
    
    // Settings functions
    async openSettings() {
      this.showSettings = true;
      this.settingsTab = 'password';
      this.passwordForm = { current: '', new: '', confirm: '' };
      this.passwordError = '';
      this.passwordSuccess = '';
      this.createdToken = null;
      await this.loadApiTokens();
    },
    
    closeSettings() {
      this.showSettings = false;
      this.createdToken = null;
    },
    
    async loadApiTokens() {
      const result = await api.get('/api/v1/tokens');
      if (result && result.data) {
        this.apiTokens = result.data;
      }
    },
    
    async changePassword() {
      this.passwordError = '';
      this.passwordSuccess = '';
      
      if (this.passwordForm.new !== this.passwordForm.confirm) {
        this.passwordError = 'New passwords do not match';
        return;
      }
      
      if (this.passwordForm.new.length < 6) {
        this.passwordError = 'Password must be at least 6 characters';
        return;
      }
      
      const result = await api.post('/api/v1/auth/change-password', {
        current_password: this.passwordForm.current,
        new_password: this.passwordForm.new
      });
      
      if (result && !result.error) {
        this.passwordSuccess = 'Password changed successfully. Logging out...';
        this.passwordForm = { current: '', new: '', confirm: '' };
        // Logout after successful password change
        setTimeout(async () => {
          await this.logout();
        }, 1500);
      } else {
        this.passwordError = result?.error?.message || 'Failed to change password';
      }
    },
    
    async createApiToken() {
      if (!this.newToken.name.trim()) {
        showToast('Token name is required', 'error');
        return;
      }
      
      const result = await api.post('/api/v1/tokens', {
        name: this.newToken.name,
        permissions: this.newToken.permissions,
        expires_in_days: parseInt(this.newToken.expires_in_days) || null
      });
      
      if (result && !result.error) {
        this.createdToken = result.token;
        this.newToken = { name: '', permissions: 'read', expires_in_days: 30 };
        await this.loadApiTokens();
        showToast('API token created');
      } else {
        showToast(result?.error?.message || 'Failed to create token', 'error');
      }
    },
    
    async deleteApiToken(tokenId) {
      if (!confirm('Are you sure you want to delete this API token?')) return;
      
      const result = await api.delete(`/api/v1/tokens/${tokenId}`);
      if (!result || !result.error) {
        await this.loadApiTokens();
        showToast('API token deleted');
      } else {
        showToast('Failed to delete token', 'error');
      }
    },
    
    copyToken() {
      if (this.createdToken) {
        navigator.clipboard.writeText(this.createdToken);
        showToast('Token copied to clipboard');
      }
    },
    
    formatTokenDate(dateStr) {
      if (!dateStr) return 'Never';
      return new Date(dateStr).toLocaleDateString();
    },
    
    // Computed: Flatten folders for select dropdown (includes nested with indentation)
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
    
    // Folder management
    showNewFolderModal() {
      this.editingFolder = { name: '', parent_id: '' };
      this.showFolderModal = true;
    },
    
    renameFolder(folder) {
      this.editingFolder = { id: folder.id, name: folder.name, parent_id: folder.parent_id || '' };
      this.showFolderModal = true;
    },
    
    async saveFolder() {
      if (!this.editingFolder.name.trim()) {
        showToast('Folder name is required', 'error');
        return;
      }
      
      const data = {
        name: this.editingFolder.name,
        parent_id: this.editingFolder.parent_id ? parseInt(this.editingFolder.parent_id) : null
      };
      
      let result;
      if (this.editingFolder.id) {
        // Update existing folder
        result = await api.put(`/api/v1/folders/${this.editingFolder.id}`, data);
      } else {
        // Create new folder
        result = await api.post('/api/v1/folders', data);
      }
      
      if (result && !result.error) {
        this.showFolderModal = false;
        await this.loadFolders();
        showToast(this.editingFolder.id ? 'Folder renamed' : 'Folder created');
      } else {
        showToast(result?.error?.message || 'Failed to save folder', 'error');
      }
    },
    
    async deleteFolder(folder) {
      if (!confirm(`Delete folder "${folder.name}"? Snippets in this folder will not be deleted.`)) return;
      
      const result = await api.delete(`/api/v1/folders/${folder.id}`);
      if (!result || !result.error) {
        await this.loadFolders();
        if (this.filter.folderId === folder.id) {
          this.clearFilters();
        }
        showToast('Folder deleted');
      } else {
        showToast('Failed to delete folder', 'error');
      }
    },
    
    // Tag management
    renameTag(tag) {
      this.editingTag = { id: tag.id, name: tag.name };
      this.showTagModal = true;
    },
    
    async saveTag() {
      if (!this.editingTag.name.trim()) {
        showToast('Tag name is required', 'error');
        return;
      }
      
      const result = await api.put(`/api/v1/tags/${this.editingTag.id}`, {
        name: this.editingTag.name
      });
      
      if (result && !result.error) {
        this.showTagModal = false;
        await this.loadTags();
        showToast('Tag renamed');
      } else {
        showToast(result?.error?.message || 'Failed to rename tag', 'error');
      }
    },
    
    async deleteTag(tag) {
      if (!confirm(`Delete tag "${tag.name}"? This will remove the tag from all snippets.`)) return;
      
      const result = await api.delete(`/api/v1/tags/${tag.id}`);
      if (!result || !result.error) {
        await this.loadTags();
        if (this.filter.tagId === tag.id) {
          this.clearFilters();
        }
        showToast('Tag deleted');
      } else {
        showToast('Failed to delete tag', 'error');
      }
    }
  }));
  
  // Login form
  Alpine.data('loginForm', () => ({
    password: '',
    error: '',
    loading: false,
    
    async login() {
      this.loading = true;
      this.error = '';
      
      try {
        const response = await fetch('/api/v1/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ password: this.password })
        });
        
        const result = await response.json();
        
        if (result.success) {
          window.location.href = '/';
        } else {
          this.error = result.error?.message || 'Invalid password';
        }
      } catch (err) {
        this.error = 'Connection error';
      }
      
      this.loading = false;
    }
  }));
  
  // Public snippet view
  Alpine.data('publicSnippet', () => ({
    snippet: null,
    loading: true,
    error: false,
    errorMessage: '',
    
    async init() {
      // Get snippet ID from URL path: /s/{id}
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
        const result = await response.json();
        
        if (response.ok && result) {
          this.snippet = result;
          this.$nextTick(() => {
            if (typeof Prism !== 'undefined') {
              Prism.highlightAll();
            }
          });
        } else {
          this.error = true;
          this.errorMessage = result.error?.message || 'This snippet is not available or not public';
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
    
    getLanguageColor(lang) {
      const colors = {
        javascript: '#f7df1e',
        typescript: '#3178c6',
        python: '#3776ab',
        go: '#00add8',
        rust: '#dea584',
        java: '#b07219',
        csharp: '#239120',
        cpp: '#f34b7d',
        ruby: '#cc342d',
        php: '#777bb4',
        swift: '#fa7343',
        kotlin: '#a97bff',
        html: '#e34c26',
        css: '#1572b6',
        sql: '#e38c00',
        bash: '#4eaa25',
        json: '#292929',
        yaml: '#cb171e',
        markdown: '#083fa1',
        plaintext: '#6b7280'
      };
      return colors[lang] || '#6b7280';
    },
    
    formatDate(dateStr) {
      if (!dateStr) return '';
      const date = new Date(dateStr);
      return date.toLocaleDateString();
    }
  }));
});

// Keyboard shortcuts
document.addEventListener('keydown', (e) => {
  // Ctrl/Cmd + K: Focus search
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault();
    document.querySelector('.search-input')?.focus();
  }
  
  // Ctrl/Cmd + N: New snippet
  if ((e.ctrlKey || e.metaKey) && e.key === 'n') {
    e.preventDefault();
    const app = Alpine.$data(document.querySelector('[x-data="snippetsApp()"]'));
    if (app) app.newSnippet();
  }
  
  // Escape: Close editor/modal
  if (e.key === 'Escape') {
    const app = Alpine.$data(document.querySelector('[x-data="snippetsApp()"]'));
    if (app?.showEditor) app.cancelEdit();
    if (app?.showDeleteModal) app.showDeleteModal = false;
  }
});
