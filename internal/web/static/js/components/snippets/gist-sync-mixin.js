// GitHub Gist Sync mixin
import { api } from '../../modules/api.js';
import { showToast } from '../../modules/toast.js';

export const gistSyncMixin = {
  gistConfig: {
    enabled: false,
    github_username: '',
    has_token: false,
    auto_sync_enabled: true,
    sync_interval_minutes: 15,
    conflict_resolution_strategy: 'manual',
    last_full_sync_at: ''
  },
  gistTokenInput: '',
  gistTestingConnection: false,
  gistSyncing: false,
  gistMappings: [],
  gistConflicts: [],
  gistLogs: [],
  showGistTokenInput: false,

  async loadGistConfig() {
    const result = await api.get('/api/v1/gist/config');
    if (result && !result.error) {
      this.gistConfig = result;
    }
  },

  async testGistConnection() {
    this.gistTestingConnection = true;
    const result = await api.post('/api/v1/gist/config/test');
    this.gistTestingConnection = false;

    if (result && !result.error) {
      showToast(`Connected as ${result.username}`, 'success');
    } else {
      showToast(result?.error?.message || 'Connection failed', 'error');
    }
  },

  async saveGistConfig() {
    const payload = {
      enabled: this.gistConfig.enabled,
      auto_sync_enabled: this.gistConfig.auto_sync_enabled,
      sync_interval_minutes: parseInt(this.gistConfig.sync_interval_minutes),
      conflict_resolution_strategy: this.gistConfig.conflict_resolution_strategy
    };

    if (this.gistTokenInput) {
      payload.github_token = this.gistTokenInput;
      // Auto-enable sync when providing a new token
      payload.enabled = true;
    }

    const result = await api.post('/api/v1/gist/config', payload);
    if (result && !result.error) {
      showToast('GitHub Gist configuration saved', 'success');
      this.gistTokenInput = '';
      this.showGistTokenInput = false;
      await this.loadGistConfig();
    } else {
      showToast(result?.error?.message || 'Failed to save configuration', 'error');
    }
  },

  async clearGistConfig() {
    if (!confirm('Are you sure you want to disconnect GitHub Gist? This will disable sync but keep existing mappings.')) {
      return;
    }

    const result = await api.delete('/api/v1/gist/config');
    if (result && !result.error) {
      showToast('GitHub Gist disconnected', 'success');
      await this.loadGistConfig();
    } else {
      showToast(result?.error?.message || 'Failed to disconnect', 'error');
    }
  },

  async enableSyncForAll() {
    if (!confirm('This will create a GitHub Gist for ALL your snippets. Continue?')) {
      return;
    }

    this.gistSyncing = true;
    const result = await api.post('/api/v1/gist/sync/enable-all');
    this.gistSyncing = false;

    if (result && !result.error) {
      showToast(`Enabled sync for ${result.enabled} snippets${result.errors > 0 ? ` (${result.errors} errors)` : ''}`, 'success');
      await this.loadGistMappings();
    } else {
      showToast(result?.error?.message || 'Failed to enable sync', 'error');
    }
  },

  async syncAllGists() {
    this.gistSyncing = true;
    const result = await api.post('/api/v1/gist/sync/all');
    this.gistSyncing = false;

    if (result && !result.error) {
      const message = result.synced > 0 || result.conflicts > 0 || result.errors > 0
        ? `Sync complete: ${result.synced} synced, ${result.conflicts} conflicts, ${result.errors} errors`
        : 'No snippets are synced yet. Use "Enable Sync for All" first.';
      showToast(message, result.synced > 0 ? 'success' : 'info');
      await this.loadGistMappings();
      await this.loadGistConflicts();
    } else {
      showToast(result?.error?.message || 'Sync failed', 'error');
    }
  },

  async loadGistMappings() {
    const result = await api.get('/api/v1/gist/mappings');
    if (result && !result.error) {
      this.gistMappings = result;
    }
  },

  async deleteGistMapping(mappingId) {
    if (!confirm('Remove this gist mapping? The gist will remain on GitHub.')) {
      return;
    }

    const result = await api.delete(`/api/v1/gist/mappings/${mappingId}`);
    if (result && !result.error) {
      showToast('Mapping removed', 'success');
      await this.loadGistMappings();
    } else {
      showToast(result?.error?.message || 'Failed to remove mapping', 'error');
    }
  },

  async loadGistConflicts() {
    const result = await api.get('/api/v1/gist/conflicts');
    if (result && !result.error) {
      this.gistConflicts = result;
    }
  },

  async resolveGistConflict(conflictId, resolution) {
    const result = await api.post(`/api/v1/gist/conflicts/${conflictId}/resolve`, {
      resolution: resolution
    });

    if (result && !result.error) {
      showToast('Conflict resolved', 'success');
      await this.loadGistConflicts();
      await this.loadGistMappings();
    } else {
      showToast(result?.error?.message || 'Failed to resolve conflict', 'error');
    }
  },

  async loadGistLogs() {
    const result = await api.get('/api/v1/gist/logs?limit=50');
    if (result && !result.error) {
      this.gistLogs = result;
    }
  },

  async enableGistSyncForSnippet(snippetId) {
    const result = await api.post(`/api/v1/gist/sync/enable/${snippetId}`);
    if (result && !result.error) {
      showToast('Gist sync enabled for snippet', 'success');
      await this.loadGistMappings();
    } else {
      showToast(result?.error?.message || 'Failed to enable sync', 'error');
    }
  },

  async disableGistSyncForSnippet(snippetId) {
    const result = await api.post(`/api/v1/gist/sync/disable/${snippetId}`);
    if (result && !result.error) {
      showToast('Gist sync disabled for snippet', 'success');
      await this.loadGistMappings();
    } else {
      showToast(result?.error?.message || 'Failed to disable sync', 'error');
    }
  },

  async syncSnippetToGist(snippetId) {
    const result = await api.post(`/api/v1/gist/sync/snippet/${snippetId}`);
    if (result && !result.error) {
      showToast('Snippet synced to gist', 'success');
      await this.loadGistMappings();
    } else {
      showToast(result?.error?.message || 'Failed to sync snippet', 'error');
    }
  },

  getGistSyncStatus(snippetId) {
    const mapping = this.gistMappings.find(m => m.snippet_id === snippetId);
    if (!mapping) return null;
    return mapping.sync_status;
  },

  getGistUrl(snippetId) {
    const mapping = this.gistMappings.find(m => m.snippet_id === snippetId);
    return mapping?.gist_url || null;
  },

  formatGistDate(dateStr) {
    if (!dateStr) return 'Never';
    return new Date(dateStr).toLocaleString();
  },

  getConflictStrategyLabel(strategy) {
    const labels = {
      'manual': 'Manual Resolution',
      'snipo_wins': 'Snipo Always Wins',
      'gist_wins': 'Gist Always Wins',
      'newest_wins': 'Newest Version Wins'
    };
    return labels[strategy] || strategy;
  },

  getSyncStatusBadge(status) {
    const badges = {
      'synced': { icon: '✓', class: 'text-green-600', label: 'Synced' },
      'pending': { icon: '⟳', class: 'text-yellow-600', label: 'Pending' },
      'conflict': { icon: '⚠', class: 'text-orange-600', label: 'Conflict' },
      'error': { icon: '✗', class: 'text-red-600', label: 'Error' }
    };
    return badges[status] || { icon: '?', class: 'text-gray-600', label: 'Unknown' };
  }
};
