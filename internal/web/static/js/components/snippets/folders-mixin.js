// Folders mixin
import { api } from '../../modules/api.js';
import { showToast } from '../../modules/toast.js';

export const foldersMixin = {
  showFolderModal: false,
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
