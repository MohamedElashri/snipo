// Settings mixin
import { api } from '../../modules/api.js';
import { showToast } from '../../modules/toast.js';

export const settingsMixin = {
  showSettings: false,
  settingsTab: 'password',
  apiTokens: [],
  newToken: { name: '', permissions: 'read', expires_in_days: 30 },
  createdToken: null,
  passwordForm: { current: '', new: '', confirm: '' },
  passwordError: '',
  passwordSuccess: '',

  async openSettings() {
    this.showSettings = true;
    this.settingsTab = 'password';
    this.passwordForm = { current: '', new: '', confirm: '' };
    this.passwordError = '';
    this.passwordSuccess = '';
    this.createdToken = null;
    await this.loadApiTokens();
  },

  openEditorSettings() {
    this.showSettings = true;
    this.settingsTab = 'editor';
    this.passwordForm = { current: '', new: '', confirm: '' };
    this.passwordError = '';
    this.passwordSuccess = '';
    this.createdToken = null;
  },

  closeSettings() {
    this.showSettings = false;
    this.createdToken = null;
  },

  async loadApiTokens() {
    const result = await api.get('/api/v1/tokens');
    if (result) {
      this.apiTokens = result;
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

  async updateSettings() {
    const result = await api.put('/api/v1/settings', this.settings);
    if (result) {
      this.settings = result;
      // Cache settings for theme updates
      try {
        sessionStorage.setItem('snipo-settings', JSON.stringify(result));
      } catch (e) {
        // Ignore storage errors
      }
      showToast('Settings updated');
    }
  },

  applyMarkdownFontSize() {
    if (!this.settings) return;
    
    const fontSize = this.settings.markdown_font_size || 14;
    document.documentElement.style.setProperty('--markdown-font-size', `${fontSize}px`);
  }
};
