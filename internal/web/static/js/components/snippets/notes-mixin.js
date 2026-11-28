import { api } from '../../modules/api.js';
import { showToast } from '../../modules/toast.js';

const noteFieldLabels = {
  title: 'title',
  description: 'description',
  language: 'format',
  tags: 'tags',
  folder_id: 'folder',
  is_public: 'visibility',
  is_archived: 'archive status',
  persistedExpiration: 'expiration'
};

export const notesMixin = {
  noteMode: 'write',
  noteConflict: null,
  noteOverviewError: false,

  focusNote() {
    this.$nextTick(() => {
      if (this.isEditing && this.isNote && this.noteMode === 'write') this.$refs.noteBody?.focus();
    });
  },

  setNoteMode(mode) {
    this.noteMode = mode;
    if (mode === 'write') this.focusNote();
  },

  canConvertItem() {
    const item = this.editingSnippet;
    if (this.isEditing || !item?.id || item.deleted_at) return false;
    return this.isNote || ((item.files?.length || 0) <= 1 && ['markdown', 'plaintext'].includes(item.files?.[0]?.language || item.language));
  },

  async convertItem() {
    if (!this.canConvertItem() || this.saving) return;
    const kind = this.isNote ? 'snippet' : 'note';
    const id = this.editingSnippet.id;
    this.saving = true;
    let result;
    try {
      result = await api.post(`/api/v1/snippets/${id}/convert`, { kind });
      if (!result || result.error) throw new Error(result?.error?.message || 'Could not convert item.');
    } catch (error) {
      showToast(error.message, 'error');
      return;
    } finally {
      this.saving = false;
    }
    showToast(kind === 'note' ? 'Converted to note' : 'Converted to snippet');
    await Promise.all([this.viewSnippet({ id, kind }), this.loadSnippets()]);
  },

  downloadNote() {
    const note = this.editingSnippet;
    const markdown = note.language === 'markdown';
    const url = URL.createObjectURL(new Blob([note.content], { type: markdown ? 'text/markdown' : 'text/plain' }));
    const link = document.createElement('a');
    const name = note.title.replace(/[\\/:*?"<>|]/g, '_');
    link.href = url;
    link.download = name + (markdown ? '.md' : '.txt');
    link.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  },

  newNoteDraftId() {
    return Date.now().toString(36) + '-' + Math.random().toString(36).slice(2);
  },

  noteMetadata(note) {
    return {
      title: note.title,
      description: note.description || '',
      language: note.language,
      tags: (note.tags || []).map(tag => typeof tag === 'string' ? tag : tag.name).sort(),
      folder_id: Number(note.folder_id) || null,
      is_public: Boolean(note.is_public),
      is_archived: Boolean(note.is_archived),
      persistedExpiration: note.persistedExpiration || null
    };
  },

  changedNoteFields(base, server) {
    return Object.keys(noteFieldLabels)
      .filter(field => !base || JSON.stringify(base[field]) !== JSON.stringify(server[field]));
  },

  noteConflictFieldNames() {
    if (!this.noteConflict) return '';
    return this.changedNoteFields(this.editingSnippet.noteBase, this.noteConflict.metadata)
      .map(field => noteFieldLabels[field]).join(', ');
  },

  async showNoteConflict() {
    const result = await api.get(`/api/v1/snippets/${this.editingSnippet.id}`);
    if (result?.kind === 'note' && !result.deleted_at) {
      this.noteConflict = {
        revision: result.revision,
        content: result.content,
        metadata: this.noteMetadata({
          ...result, folder_id: result.folders?.[0]?.id,
          persistedExpiration: result.expires_at
        })
      };
    }
  },

  useReviewedRevision() {
    if (!this.noteConflict) return;
    const { revision, content, metadata } = this.noteConflict;
    for (const field of this.changedNoteFields(this.editingSnippet.noteBase, metadata)) {
      // Respect newer metadata, especially a decision to revoke a public link.
      this.editingSnippet[field] = metadata[field];
      if (field === 'persistedExpiration') this.editingSnippet.expires_at = this.resolveToShorthand(metadata[field]);
    }
    this.editingSnippet.revision = revision;
    this.editingSnippet.noteBase = { ...metadata, content };
    if (this.editingSnippet.language !== 'markdown') this.noteMode = 'write';
    this.noteConflict = null;
    this.saveDraft();
  },

  noteExcerpt(item) {
    const excerpt = this.snippetExcerpt(item).trim();
    return item.language === 'markdown'
      ? excerpt.replace(/^(?:#{1,6}|>|[-*+])\s+/gm, '').replace(/\s+/g, ' ')
      : excerpt.replace(/\s+/g, ' ');
  }
};
