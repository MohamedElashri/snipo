// History mixin
import { api } from '../../modules/api.js';
import { showToast } from '../../modules/toast.js';

export const historyMixin = {
  showHistory: false,
  history: [],
  historyLoading: false,
  expandedHistory: null,
  showHistoryDetail: false,
  viewingHistoryEntry: null,
  showRestoreConfirm: false,
  restoringHistoryEntry: null,

  async openHistory(snippet) {
    if (!snippet || !snippet.id) return;
    
    this.showHistory = true;
    this.history = [];
    this.historyLoading = true;
    this.expandedHistory = null;

    try {
      const result = await api.get(`/api/v1/snippets/${snippet.id}/history?limit=50`);
      if (!Array.isArray(result)) throw new Error(result?.error?.message || 'Failed to load history');
      this.history = result;
    } catch (error) {
      console.error('Failed to load history:', error);
      showToast('Failed to load history', 'error');
    } finally {
      this.historyLoading = false;
    }
  },

  closeHistory() {
    this.showHistory = false;
    this.history = [];
    this.expandedHistory = null;
  },

  isCurrentHistoryEntry(entry) {
    const current = this.editingSnippet;
    if (!current || !entry || entry.snippet_id !== current.id ||
        entry.title !== current.title || (entry.description || '') !== (current.description || '')) return false;

    const entryFiles = entry.files || [];
    if (current.kind === 'note') {
      const body = entryFiles[0] || entry;
      return entryFiles.length <= 1 && body.content === current.content && body.language === current.language;
    }

    const files = current.files || [];
    return entry.content === current.content && entry.language === current.language &&
      entry.is_public === current.is_public && entry.is_archived === current.is_archived &&
      entryFiles.length === files.length && entryFiles.every((file, index) =>
        file.filename === files[index].filename && file.content === files[index].content &&
        file.language === files[index].language);
  },

  canRestoreHistoryEntry(entry) {
    if (!entry || this.isCurrentHistoryEntry(entry)) return false;
    if (!this.isNote) return true;
    const files = entry.files || [];
    return files.length <= 1 && ['markdown', 'plaintext'].includes((files[0] || entry).language);
  },

  viewHistoryDetail(historyEntry) {
    if (!historyEntry) return;
    this.viewingHistoryEntry = historyEntry;
    this.showHistoryDetail = true;
  },

  closeHistoryDetail() {
    this.showHistoryDetail = false;
    this.viewingHistoryEntry = null;
  },

  copyHistoryContent(content) {
    if (!content) return;
    navigator.clipboard.writeText(content)
      .then(() => showToast('Content copied to clipboard'))
      .catch(() => showToast('Failed to copy content', 'error'));
  },

  confirmRestoreHistory(historyEntry) {
    if (!historyEntry || !historyEntry.snippet_id || !historyEntry.id) {
      showToast('Invalid history entry', 'error');
      return;
    }

    if (!this.canRestoreHistoryEntry(historyEntry)) {
      showToast(this.isCurrentHistoryEntry(historyEntry) ? 'This is already the current version' :
        'This version cannot be restored as a note', 'info');
      return;
    }

    this.restoringHistoryEntry = historyEntry;
    this.showRestoreConfirm = true;
  },

  async executeRestore() {
    const historyEntry = this.restoringHistoryEntry;
    if (!historyEntry) return;

    if (this.editingSnippet && this.editingSnippet.id !== historyEntry.snippet_id) {
      showToast('Cannot restore: snippet mismatch', 'error');
      this.showRestoreConfirm = false;
      return;
    }

    this.showRestoreConfirm = false;

    try {
      const result = await api.post(
        `/api/v1/snippets/${historyEntry.snippet_id}/history/${historyEntry.id}/restore`
      );
      
      if (result && !result.error) {
        showToast(result.kind === 'note' ? 'Note restored from history' : 'Snippet restored from history');
        
        this.closeHistoryDetail();
        this.closeHistory();
        
        if (this.editingSnippet && this.editingSnippet.id === historyEntry.snippet_id) {
          await this.viewSnippet(result);
        }
        
        await this.loadSnippets();
      } else {
        throw new Error(result?.error?.message || 'No response from server');
      }
    } catch (error) {
      console.error('Failed to restore history:', error);
      
      if (error.message && error.message.includes('404')) {
        showToast('Snippet or history entry not found', 'error');
      } else if (error.message && error.message.includes('403')) {
        showToast('Permission denied', 'error');
      } else if (error.message && error.message.includes('network')) {
        showToast('Network error - please check your connection', 'error');
      } else {
        showToast(error.message || 'Failed to restore from history', 'error');
      }
    } finally {
      this.restoringHistoryEntry = null;
    }
  },

  toggleHistoryPreview(historyId) {
    this.expandedHistory = this.expandedHistory === historyId ? null : historyId;
  }
};
