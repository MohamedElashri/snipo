// Presentation and navigation shared by the library and editor.
export const workspaceMixin = {
  editorExpanded: false,
  tagSearch: '',
  showAllTags: false,
  visibleTags() {
    const query = this.tagSearch.trim().toLocaleLowerCase();
    if (query) return this.tags.filter(tag => tag.name.toLocaleLowerCase().includes(query));
    if (this.showAllTags || this.tags.length <= 12) return this.tags;
    const visible = this.tags.slice(0, 10);
    const selected = this.tags.find(tag => tag.id === this.filter.tagId);
    if (selected && !visible.includes(selected)) visible.push(selected);
    return visible;
  },
  async goToPage(page) {
    if (this.loading || page < 1 || page > this.pagination.totalPages) return;
    this.pagination.page = page;
    await this.loadSnippets();
    this.$root?.querySelector('.snippet-list')?.scrollTo({ top: 0 });
  },
  libraryTitle() {
    if (this.filter.isDeleted) return 'Trash';
    if (this.filter.isArchived) return 'Archive';
    if (this.filter.isFavorite) return 'Favorites';
    if (this.filter.folderId) return this.flattenedFolders.find(f => f.id == this.filter.folderId)?.name || 'Folder';
    if (this.filter.tagId) return this.tags.find(t => t.id == this.filter.tagId)?.name || 'Tag';
    return this.filter.query ? 'Search results' : 'All snippets';
  },
  activeFilterChips() {
    const labels = {
      query: this.filter.query,
      language: this.filter.language,
      folderId: this.flattenedFolders.find(f => f.id == this.filter.folderId)?.name,
      tagId: this.tags.find(t => t.id == this.filter.tagId)?.name,
      isFavorite: 'Favorites', isArchived: 'Archive', isDeleted: 'Trash'
    };
    return Object.entries(this.filter).filter(([, value]) => !!value)
      .map(([key, value]) => ({ key, label: labels[key] || String(value) }));
  },
  hasActiveFilters() { return this.activeFilterChips().length > 0; },
  async removeFilter(key) {
    this.filter[key] = ['query', 'language'].includes(key) ? '' : null;
    await this.search();
    this.updateUrl(this.showEditor ? { snippet: this.editingSnippet?.id } : {});
  },
  prepareLibraryNavigation() {
    if (this.saving) return false;
    this.preserveDraft();
    this.isEditing = false;
    this.showSettings = false;
    this.closeMobileNavigation();
    this.destroyCodeEditor();
    ++this.viewRequest;
    this.showEditor = false;
    return true;
  },
  closeMobileNavigation() {
    if (window.matchMedia?.('(max-width: 768px)').matches && this.$store?.app) {
      this.$store.app.sidebarOpen = false;
    }
  },
  setLibraryLayout(layout) {
    if (!['grid', 'list', 'compact'].includes(layout)) return;
    this.setViewMode(layout === 'grid' ? 'grid' : 'list');
    this.density = layout === 'compact' ? 'compact' : 'comfortable';
    localStorage.setItem('snipo-density', this.density);
  },
  toggleEditorExpanded() {
    this.editorExpanded = !this.editorExpanded;
    this.$nextTick(() => this.codeEditor?.resize());
  },
  focusFirstResult() { document.querySelector('.snippet-open')?.focus(); },
  focusResult(event, direction) {
    const results = [...document.querySelectorAll('.snippet-open')];
    const index = results.indexOf(event.currentTarget);
    results[Math.max(0, Math.min(results.length - 1, index + direction))]?.focus();
  },
  snippetExcerpt(snippet) {
    if (snippet.is_summary) return snippet.preview || '';
    const content = snippet.content || snippet.files?.[0]?.content || '';
    const query = this.filter.query.trim().toLowerCase();
    const lines = content.split('\n');
    const match = query ? lines.findIndex(line => line.toLowerCase().includes(query)) : -1;
    return lines.slice(Math.max(0, match), Math.max(0, match) + 2).join('\n');
  }
};
