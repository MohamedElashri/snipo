// File manager mixin - handles multi-file operations
import { showToast } from '../../modules/toast.js';

// Sanitize filename by replacing spaces with underscores
function sanitizeFilename(filename) {
  return filename
    .replace(/\s+/g, '_')  // Replace spaces with underscores
    .replace(/_+/g, '_');   // Replace multiple underscores with single underscore
}

export const fileManagerMixin = {
  _resetFileManagerState() {
    this.fileManagerState.operationInProgress = false;
    this.fileManagerState.editorDirty = false;
    this.fileManagerState.lastSyncedContent = '';
    this.fileManagerState.pendingOperation = null;
  },

  _ensureEditableFiles() {
    if (!this.editingSnippet) {
      return [];
    }

    if (!Array.isArray(this.editingSnippet.files)) {
      this.editingSnippet.files = [];
    }

    if (this.editingSnippet.files.length === 0) {
      const language = this.editingSnippet.language || 'plaintext';
      const ext = this.getFileExtension ? this.getFileExtension(language) : 'txt';
      const content = this.aceEditor ? this.aceEditor.getValue() : (this.editingSnippet.content || '');
      this.editingSnippet.files = [{
        id: 0,
        filename: 'main.' + ext,
        content,
        language
      }];
    }

    if (this.activeFileIndex < 0 || this.activeFileIndex >= this.editingSnippet.files.length) {
      this.activeFileIndex = 0;
    }

    return this.editingSnippet.files;
  },

  _syncEditorToFile() {
    if (!this.aceEditor || !Array.isArray(this.editingSnippet?.files) || this.editingSnippet.files.length === 0) {
      return;
    }

    if (this.activeFileIndex < 0 || this.activeFileIndex >= this.editingSnippet.files.length) {
      this.activeFileIndex = 0;
    }
    
    const currentContent = this.aceEditor.getValue();
    this.editingSnippet.files[this.activeFileIndex].content = currentContent;
    this.fileManagerState.lastSyncedContent = currentContent;
    this.fileManagerState.editorDirty = false;
  },

  _loadFileToEditor(fileIndex) {
    if (!this.aceEditor || !Array.isArray(this.editingSnippet?.files) || !this.editingSnippet.files[fileIndex]) {
      return;
    }

    const file = this.editingSnippet.files[fileIndex];
    const content = file.content || '';
    
    this.aceIgnoreChange = true;
    try {
      this.aceEditor.setValue(content, -1);
      this.aceEditor.session.setMode(this.getAceMode(file.language));
      this.fileManagerState.lastSyncedContent = content;
      this.fileManagerState.editorDirty = false;
      this.aceEditor.resize();
      this.aceEditor.focus();
    } finally {
      this.aceIgnoreChange = false;
    }
  },

  _beginFileOperation() {
    if (this.fileManagerState.operationInProgress) {
      console.warn('File operation already in progress');
      return false;
    }
    
    this.fileManagerState.operationInProgress = true;
    this._syncEditorToFile();
    return true;
  },

  _endFileOperation(newFileIndex) {
    if (!this.fileManagerState.operationInProgress) {
      return;
    }

    const files = this.editingSnippet?.files || [];
    if (files.length === 0) {
      this.fileManagerState.operationInProgress = false;
      return;
    }

    const targetIndex = Math.max(0, Math.min(newFileIndex, files.length - 1));
    this.activeFileIndex = targetIndex;
    
    this.$nextTick(() => {
      try {
        if (this.aceEditor) {
          this._loadFileToEditor(targetIndex);
        } else {
          this.updateAceEditor();
        }
        this.updateTextDirection();
      } catch (error) {
        console.error('Error loading file into editor:', error);
      } finally {
        this.fileManagerState.operationInProgress = false;
      }
    });
  },

  syncCurrentContent() {
    this._syncEditorToFile();
  },

  addFile() {
    if (!this._beginFileOperation()) {
      return;
    }

    try {
      const files = this._ensureEditableFiles();

      const newFile = {
        id: 0,
        filename: 'newfile.txt',
        content: '',
        language: 'plaintext'
      };
      files.push(newFile);
      
      const newIndex = files.length - 1;
      this._endFileOperation(newIndex);

      setTimeout(() => {
        const inputs = document.querySelectorAll('.filename-input');
        if (inputs.length > 0) {
          const lastInput = inputs[inputs.length - 1];
          lastInput.focus();
          lastInput.select();
        }
      }, 100);

    } catch (error) {
      console.error('Error adding file:', error);
      this.fileManagerState.operationInProgress = false;
    }

    this.scheduleAutoSave();
  },

  removeFile(index) {
    if (!this.editingSnippet.files || this.editingSnippet.files.length <= 1) {
      showToast('Cannot remove the last file', 'warning');
      return;
    }

    if (!this._beginFileOperation()) {
      return;
    }

    try {
      let newActiveIndex;
      if (index === this.activeFileIndex) {
        newActiveIndex = Math.max(0, index - 1);
      } else if (index < this.activeFileIndex) {
        newActiveIndex = this.activeFileIndex - 1;
      } else {
        newActiveIndex = this.activeFileIndex;
      }

      this.editingSnippet.files.splice(index, 1);
      this._endFileOperation(newActiveIndex);

    } catch (error) {
      console.error('Error removing file:', error);
      this.fileManagerState.operationInProgress = false;
    }

    this.scheduleAutoSave();
  },

  selectFile(index) {
    const files = this.editingSnippet?.files || [];
    if (index < 0 || index >= files.length || index === this.activeFileIndex) {
      return;
    }

    if (!this._beginFileOperation()) {
      return;
    }

    try {
      this._endFileOperation(index);
      
      this.$nextTick(() => {
        this.highlightAll();
      });
    } catch (error) {
      console.error('Error selecting file:', error);
      this.fileManagerState.operationInProgress = false;
    }
  },

  updateActiveFileContent(content) {
    if (this.fileManagerState.operationInProgress) {
      return;
    }

    if (this.editingSnippet.files && this.editingSnippet.files.length > 0) {
      if (this.activeFileIndex < 0 || this.activeFileIndex >= this.editingSnippet.files.length) {
        this.activeFileIndex = 0;
      }
      this.editingSnippet.files[this.activeFileIndex].content = content;
      this.fileManagerState.editorDirty = true;
    } else {
      this.editingSnippet.content = content;
    }
    this.scheduleAutoSave();
  },

  updateActiveFileLanguage(language) {
    this._syncEditorToFile();
    
    const currentContent = this.aceEditor ? this.aceEditor.getValue() : '';

    if (this.editingSnippet.files && this.editingSnippet.files.length > 0) {
      if (this.activeFileIndex < 0 || this.activeFileIndex >= this.editingSnippet.files.length) {
        this.activeFileIndex = 0;
      }
      this.editingSnippet.files[this.activeFileIndex].language = language;
      this.editingSnippet.files[this.activeFileIndex].content = currentContent;
    } else {
      this.editingSnippet.language = language;
      this.editingSnippet.content = currentContent;
    }

    if (this.aceEditor) {
      try {
        this.aceEditor.session.setMode(this.getAceMode(language));
      } catch (e) {
        console.warn('Ace setMode error:', e);
      }
    }

    this.$nextTick(() => this.highlightAll());
    this.scheduleAutoSave();
  },

  updateActiveFilename(filename) {
    // Sanitize filename to remove spaces
    filename = sanitizeFilename(filename);
    
    if (this.editingSnippet.files && this.editingSnippet.files.length > 0) {
      if (this.activeFileIndex < 0 || this.activeFileIndex >= this.editingSnippet.files.length) {
        this.activeFileIndex = 0;
      }
      this.editingSnippet.files[this.activeFileIndex].filename = filename;
      
      // Pass context to detectLanguageFromFilename
      const isOnlyFile = this.editingSnippet.files.length === 1;
      const detectedLang = this.detectLanguageFromFilename(filename, { isOnlyFile });
      
      if (detectedLang) {
        const currentLang = this.editingSnippet.files[this.activeFileIndex].language;
        
        if (detectedLang !== currentLang) {
          this.editingSnippet.files[this.activeFileIndex].language = detectedLang;
          
          if (this.aceEditor) {
            try {
              this.aceEditor.session.setMode(this.getAceMode(detectedLang));
            } catch (e) {
              console.warn('Ace setMode error:', e);
            }
          }
        }
      }
    }
    this.scheduleAutoSave();
  }
};
