import { showToast } from '../../modules/toast.js';

// Draft mixin - auto-save functionality
export const draftMixin = {
  hasDraft: false,
  draftSnippet: null,
  draftSavedAt: null,
  autoSaveTimeout: null,

  saveDraft() {
    if (this._sessionEnded || !this.isEditing) return;

    const hasContent = this.editingSnippet.title ||
      this.editingSnippet.content ||
      (this.editingSnippet.files && this.editingSnippet.files.some(f => f.content));
    if (!hasContent) return;

    const draft = {
      snippet: { ...this.editingSnippet },
      savedAt: new Date().toISOString()
    };
    if (this.isNote) return this.saveNoteDraft(draft);
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
    this.refreshNoteDrafts();
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
    if (this.saving || !this.draftSnippet) return;
    const snippet = { ...this.draftSnippet };
    if (snippet.kind === 'note' || (!snippet.id && snippet.newNote)) {
      if (snippet.newNote) {
        snippet.kind = 'note';
        snippet.content = snippet.files?.[0]?.content ?? snippet.content;
        snippet.language = snippet.files?.[0]?.language === 'plaintext' ? 'plaintext' : 'markdown';
        snippet.files = [];
        snippet.sync_to_gist = false;
        snippet.expires_at = '';
      }
      if (!snippet.id) snippet.localDraftId ||= this.newNoteDraftId();
      this.saveNoteDraft({ snippet, savedAt: new Date().toISOString() });
      this.clearDraft();
    }
    this.hasDraft = false;
    this.restoreDraftSnippet(snippet);
  },

  restoreDraftSnippet(snippet) {
    if (this.saving || !snippet) return;
    this.preserveDraft();
    ++this.viewRequest;
    this.destroyCodeEditor();
    this.showSettings = false;
    this.editingSnippet = { ...snippet };
    this.noteMode = 'write';
    this.noteConflict = null;
    this.noteOverviewError = false;
    this.activeFileIndex = 0;
    this._resetFileManagerState();
    if (!this.isNote) this._ensureEditableFiles();
    this.showEditor = true;
    this.isEditing = true;
    this.saveState = 'local';
    this.updateUrl({ snippet: snippet.id, edit: true });
    showToast('Draft restored');
    if (this.isNote) this.focusNote();
    this.$nextTick(() => {
      this.updateCodeEditor();
      this.highlightAll();
    });
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
    if (this._sessionEnded) return;
    this.saveState = 'unsaved';
    if (this.autoSaveTimeout) {
      clearTimeout(this.autoSaveTimeout);
    }
    this.autoSaveTimeout = setTimeout(() => {
      this.saveDraft();
    }, 2000);
  }
};
