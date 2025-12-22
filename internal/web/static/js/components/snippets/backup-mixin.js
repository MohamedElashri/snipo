// Backup mixin
import { api } from '../../modules/api.js';
import { showToast } from '../../modules/toast.js';

export const backupMixin = {
  backupOptions: { format: 'json', password: '' },
  importOptions: { strategy: 'merge', password: '' },
  backupFile: null,
  backupLoading: false,
  importResult: null,
  s3Status: { enabled: false },
  s3Backups: [],

  async exportBackup() {
    this.backupLoading = true;
    try {
      const params = new URLSearchParams({
        format: this.backupOptions.format
      });
      if (this.backupOptions.password) {
        params.append('password', this.backupOptions.password);
      }

      const response = await fetch(`/api/v1/backup/export?${params}`, {
        credentials: 'include'
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error?.message || 'Export failed');
      }

      const disposition = response.headers.get('Content-Disposition');
      let filename = 'snipo-backup.json';
      if (disposition) {
        const match = disposition.match(/filename="(.+)"/);
        if (match) filename = match[1];
      }

      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      showToast('Backup downloaded successfully');
    } catch (err) {
      showToast(err.message || 'Failed to export backup', 'error');
    }
    this.backupLoading = false;
  },

  async importBackup() {
    if (!this.backupFile) {
      showToast('Please select a backup file', 'error');
      return;
    }

    this.backupLoading = true;
    this.importResult = null;

    try {
      const formData = new FormData();
      formData.append('file', this.backupFile);
      formData.append('strategy', this.importOptions.strategy);
      if (this.importOptions.password) {
        formData.append('password', this.importOptions.password);
      }

      const response = await fetch('/api/v1/backup/import', {
        method: 'POST',
        credentials: 'include',
        body: formData
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error?.message || 'Import failed');
      }

      this.importResult = result;
      this.backupFile = null;

      await Promise.all([
        this.loadSnippets(),
        this.loadTags(),
        this.loadFolders()
      ]);

      showToast('Backup imported successfully');
    } catch (err) {
      showToast(err.message || 'Failed to import backup', 'error');
    }
    this.backupLoading = false;
  },

  async loadS3Status() {
    try {
      const result = await api.get('/api/v1/backup/s3/status');
      if (result) {
        this.s3Status = result;
        if (result.enabled) {
          await this.loadS3Backups();
        }
      }
    } catch (err) {
      console.error('Failed to load S3 status:', err);
    }
  },

  async loadS3Backups() {
    try {
      const result = await api.get('/api/v1/backup/s3/list');
      if (result && result.backups) {
        this.s3Backups = result.backups;
      }
    } catch (err) {
      console.error('Failed to load S3 backups:', err);
    }
  },

  async syncToS3() {
    this.backupLoading = true;
    try {
      const result = await api.post('/api/v1/backup/s3/sync', {
        format: this.backupOptions.format,
        password: this.backupOptions.password
      });

      if (result && !result.error) {
        await this.loadS3Backups();
        showToast('Backup synced to S3 successfully');
      } else {
        throw new Error(result?.error?.message || 'Sync failed');
      }
    } catch (err) {
      showToast(err.message || 'Failed to sync to S3', 'error');
    }
    this.backupLoading = false;
  },

  async restoreFromS3(key) {
    if (!confirm('Restore from this backup? This will import the backup data.')) return;

    this.backupLoading = true;
    try {
      const result = await api.post('/api/v1/backup/s3/restore', {
        key: key,
        strategy: this.importOptions.strategy,
        password: this.importOptions.password
      });

      if (result && !result.error) {
        await Promise.all([
          this.loadSnippets(),
          this.loadTags(),
          this.loadFolders()
        ]);
        showToast('Backup restored successfully');
      } else {
        throw new Error(result?.error?.message || 'Restore failed');
      }
    } catch (err) {
      showToast(err.message || 'Failed to restore from S3', 'error');
    }
    this.backupLoading = false;
  },

  async deleteS3Backup(key) {
    if (!confirm('Delete this backup from S3?')) return;

    try {
      const result = await api.delete(`/api/v1/backup/s3/delete?key=${encodeURIComponent(key)}`);
      if (!result || !result.error) {
        await this.loadS3Backups();
        showToast('Backup deleted from S3');
      } else {
        throw new Error(result?.error?.message || 'Delete failed');
      }
    } catch (err) {
      showToast(err.message || 'Failed to delete backup', 'error');
    }
  }
};
