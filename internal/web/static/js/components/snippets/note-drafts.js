import { showToast } from '../../modules/toast.js';

const prefix = 'snipo-note-draft:';
const draftLifetime = 24 * 60 * 60 * 1000;

function draftKey(note) {
  if (note?.id) return prefix + 'saved:' + note.id;
  if (note?.localDraftId) return prefix + 'new:' + note.localDraftId;
  return null;
}

function draftKeys(storage) {
  return Array.from({ length: storage.length }, (_, index) => storage.key(index))
    .filter(key => key?.startsWith(prefix));
}

function loadNoteDraft(key, adopt = false) {
  if (!key) return null;
  try {
    let stored;
    try { stored = sessionStorage.getItem(key); } catch { /* Use the persistent copy. */ }
    if (stored == null) {
      stored = localStorage.getItem(key);
      if (stored && adopt) {
        try { sessionStorage.setItem(key, stored); } catch { /* Keep the persistent copy. */ }
      }
    }
    if (!stored) return null;
    const draft = JSON.parse(stored);
    const age = Date.now() - Date.parse(draft.savedAt);
    if (draft.snippet?.kind === 'note' && age >= 0 && age <= draftLifetime) return draft;
    removeNoteDraft(key, stored);
  } catch { /* Invalid or unavailable storage has no recoverable draft. */ }
  return null;
}

function removeNoteDraft(key, expected) {
  if (!key) return;
  let owned;
  let hasTabStorage = false;
  try {
    owned = expected ?? sessionStorage.getItem(key);
    // Keep a marker so this tab cannot pick up another tab's unsaved edits.
    sessionStorage.setItem(key, '');
    hasTabStorage = true;
  } catch { /* The persistent copy can still be cleared. */ }
  try {
    if (!hasTabStorage || (owned != null && localStorage.getItem(key) === owned)) localStorage.removeItem(key);
  } catch { /* Storage may be unavailable. */ }
}

export function clearStoredNoteDrafts() {
  for (const name of ['sessionStorage', 'localStorage']) {
    try {
      const storage = globalThis[name];
      for (const key of draftKeys(storage)) storage.removeItem(key);
    } catch { /* Storage may be unavailable. */ }
  }
}

export const noteDraftsMixin = {
  noteDrafts: [],

  refreshNoteDrafts() {
    const keys = new Set();
    for (const name of ['localStorage', 'sessionStorage']) {
      try { draftKeys(globalThis[name]).forEach(key => keys.add(key)); }
      catch { /* The other storage may still provide drafts. */ }
    }
    this.noteDrafts = Array.from(keys, key => loadNoteDraft(key)).filter(Boolean)
      .sort((first, second) => second.savedAt.localeCompare(first.savedAt));
  },

  findNoteDraft(note) {
    return loadNoteDraft(draftKey(note), true)?.snippet;
  },

  saveNoteDraft(draft) {
    const key = draftKey(draft.snippet);
    if (!key) return;
    // Viewing a note and closing an unchanged editor needs no recovery copy.
    if (this.saveState === 'saved' && draft.snippet.noteBase &&
        draft.snippet.noteBase.content === draft.snippet.content &&
        this.changedNoteFields(draft.snippet.noteBase, this.noteMetadata(draft.snippet)).length === 0) return;
    let stored = false;
    const serialized = JSON.stringify(draft);
    for (const name of ['sessionStorage', 'localStorage']) {
      try {
        globalThis[name].setItem(key, serialized);
        stored = true;
      } catch { /* The other storage may still be available. */ }
    }
    this.saveState = stored ? 'local' : 'local-error';
    if (stored) this.refreshNoteDrafts();
  },

  clearNoteDraft(note) {
    removeNoteDraft(draftKey(note));
    this.refreshNoteDrafts();
  },

  discardNoteDraft(note) {
    this.findNoteDraft(note);
    this.clearNoteDraft(note);
    showToast('Note draft discarded');
  }
};
