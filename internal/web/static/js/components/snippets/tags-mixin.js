// Tags mixin
import { api } from '../../modules/api.js';
import { showToast } from '../../modules/toast.js';

export const tagsMixin = {
  showTagModal: false,
  editingTag: { id: null, name: '' },

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
};
