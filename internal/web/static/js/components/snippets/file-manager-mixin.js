// File manager mixin - handles multi-file operations
import { showToast } from '../../modules/toast.js';

// Sanitize filename by replacing spaces with underscores
function sanitizeFilename(filename) {
  return filename
    .replace(/\s+/g, '_')  // Replace spaces with underscores
    .replace(/_+/g, '_');   // Replace multiple underscores with single underscore
}

export const fileManagerMixin = {
  _syncEditorToFile() {
    if (!this.aceEditor || !this.editingSnippet.files || !this.editingSnippet.files[this.activeFileIndex]) {
      return;
    }
    
    const currentContent = this.aceEditor.getValue();
    this.editingSnippet.files[this.activeFileIndex].content = currentContent;
    this.fileManagerState.lastSyncedContent = currentContent;
    this.fileManagerState.editorDirty = false;
  },

  _loadFileToEditor(fileIndex) {
    if (!this.aceEditor || !this.editingSnippet.files || !this.editingSnippet.files[fileIndex]) {
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

    this.activeFileIndex = newFileIndex;
    
    this.$nextTick(() => {
      this._loadFileToEditor(newFileIndex);
      this.fileManagerState.operationInProgress = false;
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
      if (!this.editingSnippet.files) {
        const ext = this.getFileExtension(this.editingSnippet.language);
        this.editingSnippet.files = [{
          id: 0,
          filename: 'main.' + ext,
          content: this.editingSnippet.content || '',
          language: this.editingSnippet.language || 'javascript'
        }];
      }

      const newFile = {
        id: 0,
        filename: 'newfile.txt',
        content: '',
        language: 'plaintext'
      };
      this.editingSnippet.files.push(newFile);
      
      const newIndex = this.editingSnippet.files.length - 1;
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
