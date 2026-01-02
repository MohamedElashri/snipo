// Settings mixin
import { api } from '../../modules/api.js';
import { showToast } from '../../modules/toast.js';
import { theme } from '../../modules/theme.js';

export const settingsMixin = {
  showSettings: false,
  settingsTab: 'password',
  showFontSizeHelp: false,
  apiTokens: [],
  newToken: { name: '', permissions: 'read', expires_in_days: 30 },
  createdToken: null,
  tokenPasswordAction: null, // 'create' or 'delete'
  tokenPassword: '',
  pendingTokenData: null, // Stores data for create/delete action
  passwordForm: { current: '', new: '', confirm: '' },
  passwordError: '',
  passwordSuccess: '',
  customCssChanged: false,

  async openSettings() {
    this.showSettings = true;
    this.settingsTab = 'password';
    this.passwordForm = { current: '', new: '', confirm: '' };
    this.passwordError = '';
    this.passwordSuccess = '';
    this.createdToken = null;
    this.customCssChanged = false;
    await this.loadApiTokens();
  },

  openEditorSettings() {
    this.showSettings = true;
    this.settingsTab = 'editor';
    this.passwordForm = { current: '', new: '', confirm: '' };
    this.passwordError = '';
    this.passwordSuccess = '';
    this.createdToken = null;
    this.customCssChanged = false;
  },

  closeSettings() {
    this.showSettings = false;
    this.createdToken = null;
    this.customCssChanged = false;
  },

  async loadApiTokens() {
    const result = await api.get('/api/v1/tokens');
    if (result) {
      // Handle both {data: [...]} and [...] formats
      this.apiTokens = result.data || result;
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

    // Always show password prompt for security
    this.tokenPasswordAction = 'create';
    this.tokenPassword = '';
    this.pendingTokenData = {
      name: this.newToken.name,
      permissions: this.newToken.permissions,
      expires_in_days: parseInt(this.newToken.expires_in_days) || null
    };
  },

  async deleteApiToken(tokenId) {
    // Always show password prompt for security
    this.tokenPasswordAction = 'delete';
    this.tokenPassword = '';
    this.pendingTokenData = tokenId;
  },

  async confirmTokenPassword() {
    if (!this.tokenPassword) {
      showToast('Password is required', 'error');
      return;
    }

    if (this.tokenPasswordAction === 'create') {
      await this.performCreateToken(this.tokenPassword);
    } else if (this.tokenPasswordAction === 'delete') {
      await this.performDeleteToken(this.pendingTokenData, this.tokenPassword);
    }

    this.cancelTokenPassword();
  },

  cancelTokenPassword() {
    this.tokenPasswordAction = null;
    this.tokenPassword = '';
    this.pendingTokenData = null;
  },

  async performCreateToken(password) {
    const payload = { ...this.pendingTokenData || {
      name: this.newToken.name,
      permissions: this.newToken.permissions,
      expires_in_days: parseInt(this.newToken.expires_in_days) || null
    }};

    // Always include password for security
    if (password) {
      payload.password = password;
    }

    const result = await api.post('/api/v1/tokens', payload);

    if (result && !result.error) {
      this.createdToken = result.token;
      this.newToken = { name: '', permissions: 'read', expires_in_days: 30 };
      await this.loadApiTokens();
      showToast('API token created');
    } else {
      showToast(result?.error?.message || 'Failed to create token', 'error');
    }
  },

  async performDeleteToken(tokenId, password) {
    const options = password ? { body: JSON.stringify({ password }) } : {};
    const result = await api.delete(`/api/v1/tokens/${tokenId}`, options);
    if (!result || !result.error) {
      await this.loadApiTokens();
      showToast('API token deleted');
    } else {
      showToast(result?.error?.message || 'Failed to delete token', 'error');
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

  async saveAndApplyCustomCSS() {
    // Validate custom CSS
    const validation = theme.validateCustomCSS(this.settings.custom_css);
    
    if (!validation.valid) {
      showToast('Invalid CSS: ' + validation.warnings.join(', '), 'error');
      return;
    }

    // Show warnings if any
    if (validation.warnings.length > 0) {
      const proceed = confirm(
        'CSS validation warnings:\n\n' + 
        validation.warnings.join('\n') + 
        '\n\nDo you want to proceed anyway?'
      );
      if (!proceed) return;
    }

    // Save settings
    const result = await api.put('/api/v1/settings', this.settings);
    if (result) {
      this.settings = result;
      // Cache settings
      try {
        sessionStorage.setItem('snipo-settings', JSON.stringify(result));
      } catch (e) {
        // Ignore storage errors
      }
      
      // Apply custom CSS immediately
      theme.applyCustomCSS(this.settings.custom_css);
      this.customCssChanged = false;
      
      showToast('Custom CSS saved and applied successfully');
    }
  },

  applyMarkdownFontSize() {
    if (!this.settings) return;
    
    const fontSize = this.settings.markdown_font_size || 14;
    document.documentElement.style.setProperty('--markdown-font-size', `${fontSize}px`);
  }
};
