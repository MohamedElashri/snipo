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
      if (result && result.data) {
        this.history = result.data;
      }
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

  viewHistoryDetail(historyEntry) {
    if (!historyEntry) return;
    this.viewingHistoryEntry = { 
      ...historyEntry, 
      is_current: this.history.indexOf(historyEntry) === 0 
    };
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

    if (this.history.indexOf(historyEntry) === 0) {
      showToast('This is already the current version', 'info');
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
      
      if (result) {
        showToast('Snippet successfully restored from history');
        
        this.closeHistoryDetail();
        this.closeHistory();
        
        if (this.editingSnippet && this.editingSnippet.id === historyEntry.snippet_id) {
          await this.loadSnippet(historyEntry.snippet_id);
        }
        
        await this.loadSnippets();
      } else {
        throw new Error('No response from server');
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
        showToast('Failed to restore from history', 'error');
      }
    } finally {
      this.restoringHistoryEntry = null;
    }
  },

  toggleHistoryPreview(historyId) {
    this.expandedHistory = this.expandedHistory === historyId ? null : historyId;
  }
};
