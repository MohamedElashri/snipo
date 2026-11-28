// Folders mixin
import { api } from '../../modules/api.js';
import { showToast } from '../../modules/toast.js';
import { renderMarkdown } from '../../utils/helpers.js';

export const foldersMixin = {
  showFolderModal: false,
  renderFolderOverview: renderMarkdown,
  folderOverview: null,
  folderOverviewRequest: 0,
  savingFolderOverview: false,

  async loadFolderOverview(folderId) {
    const request = ++this.folderOverviewRequest;
    if (this.folderOverview?.folderId !== folderId) this.folderOverview = null;
    if (!folderId) return;

    try {
      const folder = await api.get(`/api/v1/folders/${folderId}`);
      if (request !== this.folderOverviewRequest) return;
      if (!folder?.overview_snippet_id) {
        this.folderOverview = null;
        return;
      }
      const snippet = await api.get(`/api/v1/snippets/${folder.overview_snippet_id}`);
      if (request !== this.folderOverviewRequest) return;
      const content = snippet?.files?.length
        ? snippet.files.find(file => file.language === 'markdown')?.content
        : (snippet?.language === 'markdown' ? snippet.content : undefined);
      this.folderOverview = snippet?.folders?.some(folder => folder.id === folderId)
        && !snippet.deleted_at && !snippet.is_archived && !this.isExpired(snippet.expires_at)
        && content !== undefined ? { folderId, id: snippet.id, title: snippet.title, content } : null;
    } catch {
      if (request === this.folderOverviewRequest) this.folderOverview = null;
    }
  },

  canSetFolderOverview() {
    const snippet = this.editingSnippet;
    return !this.isEditing && snippet?.id && snippet.folder_id && !snippet.is_archived &&
      !snippet.deleted_at && !this.isExpired(snippet.expires_at) &&
      (snippet.files.length ? snippet.files.some(file => file.language === 'markdown') : snippet.language === 'markdown');
  },

  async setFolderOverview(folderId, snippetId) {
    if (this.savingFolderOverview) return;
    this.savingFolderOverview = true;
    try {
      const result = await api.put(`/api/v1/folders/${folderId}/overview`, { snippet_id: snippetId });
      if (!result || result.error) throw new Error(result?.error?.message || 'Could not update folder overview');
      if (this.filter.folderId === folderId) await this.loadFolderOverview(folderId);
      showToast(snippetId ? 'Folder overview set' : 'Folder overview removed');
    } catch (error) {
      showToast(error.message || 'Could not update folder overview', 'error');
    } finally {
      this.savingFolderOverview = false;
    }
  },
  editingFolder: { name: '', parent_id: '' },

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
      result = await api.put(`/api/v1/folders/${this.editingFolder.id}`, data);
    } else {
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
  }
};
