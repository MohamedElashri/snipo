import { showToast } from '../../modules/toast.js';

// Draft mixin - auto-save functionality
export const draftMixin = {
  hasDraft: false,
  draftSnippet: null,
  draftSavedAt: null,
  autoSaveTimeout: null,

  saveDraft() {
    if (!this.isEditing) return;

    const hasContent = this.editingSnippet.title ||
      this.editingSnippet.content ||
      (this.editingSnippet.files && this.editingSnippet.files.some(f => f.content));
    if (!hasContent) return;

    const draft = {
      snippet: { ...this.editingSnippet },
      savedAt: new Date().toISOString()
    };
    const serialized = JSON.stringify(draft);
    // Keep a tab-local copy so another tab cannot replace its recovery buffer.
    try { sessionStorage.setItem('snipo-draft', serialized); } catch { /* Use the persistent backup below. */ }
    try {
      localStorage.setItem('snipo-draft', serialized);
      this.saveState = 'local';
    } catch {
      this.saveState = 'local-error';
    }
  },

  loadDraft() {
    if (this.showEditor) return;
    this.hasDraft = false;
    this.draftSnippet = null;
    this.draftSavedAt = null;

    try {
      let draftJson;
      try { draftJson = sessionStorage.getItem('snipo-draft'); } catch { /* Storage may be disabled. */ }
      if (draftJson == null) {
        draftJson = localStorage.getItem('snipo-draft');
        if (draftJson) {
          try { sessionStorage.setItem('snipo-draft', draftJson); } catch { /* Persistent backup remains available. */ }
        }
      }
      if (!draftJson) return;

      const draft = JSON.parse(draftJson);
      if (!draft.snippet) return;

      const savedAt = new Date(draft.savedAt);
      const now = new Date();
      const hoursDiff = (now - savedAt) / (1000 * 60 * 60);

      if (!Number.isFinite(hoursDiff) || hoursDiff > 24) {
        this.clearDraft();
        return;
      }

      const hasContent = draft.snippet.title ||
        draft.snippet.content ||
        (draft.snippet.files && draft.snippet.files.some(f => f.content));
      if (hasContent) {
        this.hasDraft = true;
        this.draftSnippet = draft.snippet;
        this.draftSavedAt = savedAt;
      }
    } catch (e) {
      this.clearDraft();
    }
  },

  restoreDraft() {
    if (this.saving) return;
    this.showSettings = false;
    if (this.draftSnippet) {
      this.editingSnippet = { ...this.draftSnippet };
      this.activeFileIndex = 0;
      if (this._resetFileManagerState) {
        this._resetFileManagerState();
      }
      if (this._ensureEditableFiles) {
        this._ensureEditableFiles();
      }
      this.showEditor = true;
      this.isEditing = true;
      this.hasDraft = false;
      this.saveState = 'local';
      this.updateUrl({ snippet: this.editingSnippet.id, edit: true });
      showToast('Draft restored');
      this.$nextTick(() => {
        this.updateCodeEditor();
        this.highlightAll();
      });
    }
  },

  discardDraft() {
    this.clearDraft();
    this.hasDraft = false;
    this.draftSnippet = null;
    showToast('Draft discarded');
  },

  clearDraft() {
    clearTimeout(this.autoSaveTimeout);
    let ownDraft;
    let tabStorageAvailable = false;
    try {
      ownDraft = sessionStorage.getItem('snipo-draft');
      // An empty marker prevents a saved tab from adopting another tab's draft.
      sessionStorage.setItem('snipo-draft', '');
      tabStorageAvailable = true;
    } catch { /* Storage may be disabled. */ }
    try {
      if (!tabStorageAvailable || (ownDraft && localStorage.getItem('snipo-draft') === ownDraft)) {
        localStorage.removeItem('snipo-draft');
      }
    } catch { /* Storage may be disabled. */ }
  },

  scheduleAutoSave() {
    this.saveState = 'unsaved';
    if (this.autoSaveTimeout) {
      clearTimeout(this.autoSaveTimeout);
    }
    this.autoSaveTimeout = setTimeout(() => {
      this.saveDraft();
    }, 2000);
  }
};
