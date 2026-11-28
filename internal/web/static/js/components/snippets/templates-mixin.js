import { api } from '../../modules/api.js';
import { showToast } from '../../modules/toast.js';

const root = '/api/v1/templates';
const endpoint = id => root + '/' + encodeURIComponent(id);
const copy = value => JSON.parse(JSON.stringify(value));
const blankFile = () => ({ filename: 'note.md', language: 'markdown', content: '' });
const blankDefinition = () => ({ name: '', purpose: '', title: '', description: '', files: [blankFile()], fields: [], presets: [], tags: [], folder_id: null });

async function read(result) {
  const data = await result;
  if (!data || data.error) {
    const details = data?.error?.details;
    throw new Error(details?.length ? details.map(detail => detail.message).join(' · ') : data?.error?.message || 'Template request failed.');
  }
  return data;
}

export const templatesMixin = {
  templateView: '',
  templateLanguages: ['plaintext', 'markdown'],
  templateItems: [],
  templateLoading: false,
  templateState: 'active',
  templateMode: 'picker',
  templateQuery: '',
  templatePage: 1,
  templateTotalPages: 1,
  templateRequest: 0,
  templateBusy: false,
  templateError: '',
  templateRecord: null,
  templateForm: null,
  templateFormBaseline: null,
  templateFile: 0,
  templateTags: '',
  templateValues: {},
  templatePresetIndex: '',
  templatePreview: null,
  templateHistory: [],
  templateStarters: [],
  templateTransfer: null,
  templateConflict: null,

  templateDialog(view) {
    this.templateView = view;
    this.templateError = '';
    if (!this.$refs.templateDialog.open) this.$refs.templateDialog.showModal();
  },

  canLeaveTemplateForm() {
    return !this.templateForm || this.templateFormBaseline === JSON.stringify(this.templateDefinitionInput()) ||
      window.confirm('Discard unsaved changes to this template?');
  },

  clearTemplateForm() {
    this.templateForm = null;
    this.templateFormBaseline = null;
  },

  closeTemplateDialog(event) {
    event?.preventDefault();
    if (this.templateBusy || !this.canLeaveTemplateForm()) return;
    this.clearTemplateForm();
    this.$refs.templateDialog.close();
  },

  async templateAction(action) {
    if (this.templateBusy) return;
    this.templateBusy = true;
    this.templateError = '';
    try {
      await action();
    } catch (error) {
      this.templateError = error.message;
      this.$nextTick(() => this.$refs.templateError?.scrollIntoView({ block: 'nearest' }));
    } finally {
      this.templateBusy = false;
    }
  },

  openTemplateLibrary(mode = 'manage') {
    if (!this.canLeaveTemplateForm()) return;
    this.clearTemplateForm();
    this.templateMode = mode;
    this.templateState = 'active';
    this.templateQuery = '';
    this.templatePage = 1;
    this.templateDialog('list');
    this.loadTemplates();
    this.$nextTick(() => this.$refs.templateSearch?.focus());
  },

  async loadTemplates(page = 1) {
    const request = ++this.templateRequest;
    this.templateLoading = true;
    this.templatePage = page;
    this.templateError = '';
    this.templateItems = [];
    try {
      // This API helper rejects @ and protocol substrings in request URLs.
      const query = encodeURIComponent(this.templateQuery).replace(/\./g, '%2E');
      const result = await read(api.get(root + '?page=' + page + '&state=' + this.templateState + '&q=' + query));
      if (request === this.templateRequest) {
        this.templateItems = result.data;
        this.templateTotalPages = result.pagination.totalPages;
      }
    } catch (error) {
      if (request === this.templateRequest) this.templateError = error.message;
    } finally {
      if (request === this.templateRequest) this.templateLoading = false;
    }
  },

  selectTemplate(item) {
    return this.templateAction(async () => {
      const revision = item.revision ? '?revision=' + item.revision : '';
      this.templateRecord = await read(api.get(endpoint(item.id) + revision));
      if (this.templateMode === 'manage') {
        this.editTemplateDefinition(this.templateRecord.definition, this.templateRecord);
      } else {
        await this.prepareTemplateUse();
      }
    });
  },

  useTemplateAgain(id) {
    this.templateMode = 'picker';
    this.templateRecord = null;
    this.templateDialog('use');
    return this.selectTemplate({ id });
  },

  async prepareTemplateUse(test = false) {
    this.templateValues = {};
    this.templatePresetIndex = '';
    const def = test ? this.templateDefinitionInput() : this.templateRecord.definition;
    for (const field of def.fields || []) {
      this.templateValues[field.key] = field.default ?? '';
    }
    this.templateDialog(test ? 'test' : 'use');
    if (!test && !def.fields?.length) {
      await this.renderTemplate(false);
    } else {
      this.$nextTick(() => this.$refs.templateDialog.querySelector('[data-template-field]')?.focus());
    }
  },

  applyTemplatePreset(def) {
    const values = def.presets?.[this.templatePresetIndex]?.values || {};
    for (const field of def.fields || []) {
      this.templateValues[field.key] = Object.hasOwn(values, field.key) ? values[field.key] : field.default ?? '';
    }
  },

  async renderTemplate(test) {
    const payload = { values: this.templateValues };
    if (test) {
      payload.definition = this.templateDefinitionInput();
      this.templatePreview = await read(api.post(root + '/test', payload));
      this.templateDialog('test-preview');
      return;
    }
    if (!this.canLeaveTemplateForm()) return;
    payload.revision = this.templateRecord.revision;
    const draft = await read(api.post(endpoint(this.templateRecord.id) + '/render', payload));
    if (this.saving) throw new Error('Finish saving the open snippet first.');
    this.clearTemplateForm();
    this.$refs.templateDialog.close();
    this.newSnippet({ draft });
  },

  editTemplateDefinition(def = blankDefinition(), record = null) {
    this.templateRecord = record;
    this.templateForm = { ...blankDefinition(), ...copy(def) };
    this.templateForm.fields ||= [];
    this.templateForm.presets ||= [];
    for (const preset of this.templateForm.presets) preset.values ||= {};
    this.templateTags = (def.tags || []).join(', ');
    this.templateFormBaseline = JSON.stringify(this.templateDefinitionInput());
    if (this.templateLanguages.length === 2) {
      api.get('/api/v1/metadata/languages').then(result => {
        const languages = result?.languages || result;
        if (Array.isArray(languages)) this.templateLanguages = languages;
      }).catch(() => {});
    }
    this.templateFile = 0;
    this.templateDialog('edit');
  },

  saveSnippetAsTemplate() {
    if (this.isEditing && this.codeEditor && this.editingSnippet.files?.length) this.syncCurrentContent();
    const source = this.editingSnippet;
    const files = source.files?.length
      ? source.files.map(({ filename, content, language }) => ({ filename, content, language }))
      : [{ filename: 'main.' + this.getFileExtension(source.language), content: source.content, language: source.language }];
    this.templateMode = 'manage';
    this.editTemplateDefinition({
      ...blankDefinition(),
      name: source.title || '',
      title: source.title || '',
      description: source.description || '',
      tags: source.tags || [],
      folder_id: source.folder_id || null,
      files
    });
  },

  templateDefinitionInput() {
    const def = copy(this.templateForm);
    def.folder_id = Number(def.folder_id) || null;
    def.tags = this.templateTags.split(',').map(tag => tag.trim()).filter(Boolean);
    return def;
  },

  saveTemplateDefinition() {
    return this.templateAction(async () => {
      const definition = this.templateDefinitionInput();
      const record = this.templateRecord;
      const saved = await read(record
        ? api.put(endpoint(record.id), { revision: record.revision, definition })
        : api.post(root, definition));
      this.editTemplateDefinition(saved.definition, saved);
      showToast('Template saved');
    });
  },

  addTemplateFile() {
    this.templateForm.files.push({ ...blankFile(), filename: 'file-' + (this.templateForm.files.length + 1) + '.txt', language: 'plaintext' });
    this.templateFile = this.templateForm.files.length - 1;
  },

  moveTemplateFile(step) {
    const source = this.templateFile;
    const target = source + step;
    const files = this.templateForm.files;
    if (target < 0 || target >= files.length) return;
    files.splice(target, 0, files.splice(source, 1)[0]);
    this.templateFile = target;
  },

  removeTemplateFile() {
    this.templateForm.files.splice(this.templateFile, 1);
    this.templateFile = Math.max(0, this.templateFile - 1);
  },

  insertTemplateField(field, target = 'file') {
    if (!field.key) return;
    const token = '{{snipo.' + field.key + '}}';
    const editor = {
      file: this.$refs.templateContent,
      filename: this.$refs.templateFilename,
      title: this.$refs.templateTitle,
      description: this.$refs.templateDescription
    }[target];
    const text = editor.value;
    const start = editor.selectionStart ?? text.length;
    const end = editor.selectionEnd ?? start;
    const next = text.slice(0, start) + token + text.slice(end);
    const file = this.templateForm.files[this.templateFile];
    if (target === 'file') file.content = next;
    else if (target === 'filename') file.filename = next;
    else this.templateForm[target] = next;
    this.$nextTick(() => { editor.focus(); editor.setSelectionRange(start + token.length, start + token.length); });
  },

  addTemplateField() {
    this.templateForm.fields.push({ key: '', label: '', type: 'text', help: '', required: false, source: '', choices: [] });
  },

  changeTemplateFieldKey(field, index, next) {
    const old = field.key;
    field.key = next;
    if (!old || this.templateForm.fields.some((other, position) => position !== index && other.key === old)) return;
    for (const preset of this.templateForm.presets) {
      if (!Object.hasOwn(preset.values, old)) continue;
      const previous = preset.values[old];
      delete preset.values[old];
      if (next && !Object.hasOwn(preset.values, next)) Object.defineProperty(preset.values, next, { value: previous, writable: true, enumerable: true, configurable: true });
    }
  },

  removeTemplateField(index) {
    const [field] = this.templateForm.fields.splice(index, 1);
    for (const preset of this.templateForm.presets) {
      if (preset.values) delete preset.values[field.key];
    }
  },

  addTemplatePreset() {
    this.templateForm.presets.push({ name: '', values: {} });
  },

  setTemplatePresetValue(preset, key, value) {
    if (!key) return;
    if (value === '') delete preset.values[key];
    else Object.defineProperty(preset.values, key, { value, writable: true, enumerable: true, configurable: true });
  },

  updateTemplateStatus(status) {
    if (status.state && !this.canLeaveTemplateForm()) return;
    return this.templateAction(async () => {
      const id = this.templateRecord.id;
      const result = await api.put(endpoint(id) + '/status', status);
      if (result?.error) throw new Error(result.error.message);
      if (status.state) {
        this.clearTemplateForm();
        this.openTemplateLibrary('manage');
      } else {
        // The revision belongs to the definition currently open in the editor.
        this.templateRecord.favorite = status.favorite;
      }
    });
  },

  purgeTemplate() {
    return this.templateAction(async () => {
      const result = await api.delete(endpoint(this.templateRecord.id));
      if (result?.error) throw new Error(result.error.message);
      this.openTemplateLibrary('manage');
    });
  },

  openTemplateHistory() {
    return this.templateAction(async () => {
      this.templateHistory = await read(api.get(endpoint(this.templateRecord.id) + '/history'));
      this.templateDialog('history');
    });
  },

  restoreTemplateRevision(revision) {
    if (!this.canLeaveTemplateForm()) return;
    return this.templateAction(async () => {
      const saved = await read(api.post(endpoint(this.templateRecord.id) + '/restore-revision', {
        revision, current_revision: this.templateRecord.revision
      }));
      this.editTemplateDefinition(saved.definition, saved);
      showToast('Template revision restored');
    });
  },

  openTemplateStarters() {
    return this.templateAction(async () => {
      this.templateStarters = await read(api.get(root + '/starters'));
      this.templateDialog('starters');
    });
  },

  downloadTemplate() {
    return this.templateAction(async () => {
      const data = await read(api.get(endpoint(this.templateRecord.id) + '/export'));
      const url = URL.createObjectURL(new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' }));
      const link = document.createElement('a');
      link.href = url;
      link.download = 'snipo-template-' + this.templateRecord.id + '.json';
      link.click();
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    });
  },

  previewTemplateImport(event) {
    return this.templateAction(async () => {
      const file = event.target.files[0];
      if (!file) return;
      if (file.size > 2 * 1024 * 1024) throw new Error('Choose a template file smaller than 2 MB.');
      this.templateTransfer = JSON.parse(await file.text());
      const preview = await read(api.post(root + '/import', { template: this.templateTransfer, mode: 'preview' }));
      this.templateConflict = preview.conflict;
      this.templatePreview = preview.definition;
      this.templateDialog('import-preview');
    });
  },

  importTemplate(mode) {
    return this.templateAction(async () => {
      const saved = await read(api.post(root + '/import', {
        template: this.templateTransfer, mode, revision: this.templateConflict?.revision
      }));
      this.editTemplateDefinition(saved.definition, saved);
      showToast('Template imported');
    });
  },

  templateFolderOptions(items = this.folders, prefix = '') {
    return items.flatMap(folder => {
      const label = prefix + folder.name;
      return [{ id: folder.id, label }, ...this.templateFolderOptions(folder.children || [], label + ' / ')];
    });
  }
};
