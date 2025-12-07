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
    loading: true,
    showEditor: false,
    showDeleteModal: false,
    deleteTarget: null,
    
    async init() {
      await Promise.all([
        this.loadSnippets(),
        this.loadTags(),
        this.loadFolders()
      ]);
      // Store total snippets count on initial load
      this.totalSnippets = this.pagination.total;
      this.loading = false;
      // Highlight code after initial load
      this.$nextTick(() => this.highlightAll());
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
      this.filter.tagId = tagId;
      this.filter.folderId = null;
      this.pagination.page = 1;
      await this.loadSnippets();
    },
    
    async filterByFolder(folderId) {
      this.filter.folderId = folderId;
      this.filter.tagId = null;
      this.pagination.page = 1;
      await this.loadSnippets();
    },
    
    async clearFilters() {
      this.filter = { query: '', tagId: null, folderId: null, language: '', isFavorite: null };
      this.pagination.page = 1;
      await this.loadSnippets();
      // Update total count when showing all
      this.totalSnippets = this.pagination.total;
    },
    
    async showFavorites() {
      this.filter.isFavorite = true;
      this.filter.tagId = null;
      this.filter.folderId = null;
      this.pagination.page = 1;
      await this.loadSnippets();
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
    },
    
    async editSnippet(snippet) {
      // Fetch full snippet with tags/folders
      const result = await api.get(`/api/v1/snippets/${snippet.id}`);
      if (result) {
        this.editingSnippet = {
          ...result,
          tags: (result.tags || []).map(t => t.name),
          folder_id: result.folders?.[0]?.id || null
        };
        this.showEditor = true;
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
        await this.loadSnippets();
        await this.loadTags(); // Refresh tags in case new ones were created
      } else if (result?.error) {
        showToast(result.error.message || 'Error saving snippet', 'error');
      }
    },
    
    cancelEdit() {
      this.showEditor = false;
      this.resetEditingSnippet();
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
