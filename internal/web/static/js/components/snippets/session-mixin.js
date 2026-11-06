import { api } from '../../modules/api.js';
import { showToast } from '../../modules/toast.js';

const logoutKey = 'snipo-session-ended';
function logoutEpoch() {
  try { return localStorage.getItem(logoutKey); } catch { return null; }
}
function clearRecoveryData() {
  try { localStorage.removeItem('snipo-draft'); } catch { /* Storage may be unavailable. */ }
  for (const key of ['snipo-draft', 'snipo-settings']) {
    try { sessionStorage.removeItem(key); } catch { /* Storage may be unavailable. */ }
  }
}

export const sessionMixin = {
  loggingOut: false,
  _sessionEnded: false,

  initSessionGuard() {
    this._logoutEpoch = logoutEpoch();
    this._logoutStorage = event => {
      if (event.key === logoutKey && event.newValue !== this._logoutEpoch) this.finishLogout();
    };
    window.addEventListener('storage', this._logoutStorage);
    if (typeof BroadcastChannel !== 'undefined') {
      try {
        this._sessionChannel = new BroadcastChannel('snipo-session');
        this._sessionChannel.onmessage = event => { if (event.data === 'ended') this.finishLogout(); };
      } catch { /* The storage event remains a fallback. */ }
    }
    // Freeze no private UI in history. A restored page must ask the server again,
    // which also respects instances intentionally configured without a login page.
    this._sessionPageHide = event => {
      if (!event.persisted) return;
      this.saveDraft();
      this.hidePrivateView();
      this._sessionChannel?.close();
    };
    this._sessionPageShow = event => {
      if (!event.persisted) return;
      this.hidePrivateView();
      if (logoutEpoch() !== this._logoutEpoch) clearRecoveryData();
      window.location.reload();
    };
    window.addEventListener('pagehide', this._sessionPageHide);
    window.addEventListener('pageshow', this._sessionPageShow);
  },

  destroySessionGuard() {
    window.removeEventListener('storage', this._logoutStorage);
    window.removeEventListener('pagehide', this._sessionPageHide);
    window.removeEventListener('pageshow', this._sessionPageShow);
    this._sessionChannel?.close();
  },

  hidePrivateView() {
    this._sessionEnded = true;
    // Hide synchronously, before Alpine updates or pending API promises settle.
    if (this.$root) {
      this.$root.hidden = true;
      this.$root.style.setProperty('display', 'none', 'important');
    }
    clearTimeout(this.autoSaveTimeout);
    ++this.listRequest;
    ++this.viewRequest;
    this.isEditing = this.showEditor = this.showSettings = false;
    this.destroyCodeEditor();
    this.resetEditingSnippet();
    this.snippets = this.tags = this.folders = this.apiTokens = [];
    this.hasDraft = false;
    this.draftSnippet = this.createdToken = this.selectedSnippet = null;
    this.tokenPassword = this.disableLoginPassword = '';
    this.backupFile = null;
    this.backupOptions.password = this.importOptions.password = '';
  },

  finishLogout() {
    if (this._sessionEnded) return;
    this.hidePrivateView();
    clearRecoveryData();
    window.location.replace((window.SNIPO_CONFIG?.basePath || '') + '/login');
  },

  async logout() {
    if (this.loggingOut || this._sessionEnded) return;
    this.loggingOut = true;
    try {
      const result = await api.post('/api/v1/auth/logout');
      if (!result?.success) throw new Error('Logout was not confirmed');
      // Notify active tabs before leaving; frozen tabs compare the stored epoch.
      try { localStorage.setItem(logoutKey, Date.now() + '-' + Math.random()); } catch { /* Use the channel below. */ }
      try { this._sessionChannel?.postMessage('ended'); } catch { /* Local cleanup must still finish. */ }
      this.finishLogout();
    } catch {
      showToast('Could not log out. Your session is still open; please try again.', 'error');
    } finally {
      this.loggingOut = false;
    }
  }
};
