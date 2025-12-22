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
    localStorage.setItem('snipo-draft', JSON.stringify(draft));
  },

  loadDraft() {
    if (this.showEditor) return;

    try {
      const draftJson = localStorage.getItem('snipo-draft');
      if (!draftJson) return;

      const draft = JSON.parse(draftJson);
      if (!draft.snippet) return;

      const savedAt = new Date(draft.savedAt);
      const now = new Date();
      const hoursDiff = (now - savedAt) / (1000 * 60 * 60);

      if (hoursDiff > 24) {
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
    if (this.draftSnippet) {
      this.editingSnippet = { ...this.draftSnippet };
      this.activeFileIndex = 0;
      this.showEditor = true;
      this.isEditing = true;
      this.hasDraft = false;
      this.clearDraft();
      showToast('Draft restored');
      this.$nextTick(() => {
        this.updateAceEditor();
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
    localStorage.removeItem('snipo-draft');
  },

  scheduleAutoSave() {
    if (this.autoSaveTimeout) {
      clearTimeout(this.autoSaveTimeout);
    }
    this.autoSaveTimeout = setTimeout(() => {
      this.saveDraft();
    }, 2000);
  }
};
