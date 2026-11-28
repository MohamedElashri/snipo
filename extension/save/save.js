// This page belongs to the extension and is never exposed to visited websites.
const draftId = new URLSearchParams(window.location.search).get('draft');
function showToast(_success, message) {
    const status = document.querySelector('#save-status');
    status.textContent = message || 'Unable to save snippet.';
    status.scrollIntoView({ block: 'nearest' });
}
chrome.runtime.sendMessage({ action: 'getSaveDraft', draftId }, response => {
    if (chrome.runtime.lastError || !response?.success) {
        showToast(false, 'This save window has expired. Close it and try saving again.');
        return;
    }
    showModal(response.data).catch(() => showToast(false, 'Unable to load the save dialog.'));
});

// --- MODAL LOGIC ---

let modalOpen = false;
async function showModal(snippetData) {
    if (modalOpen) return;
    modalOpen = true;
    // 1. Fetch Tags and Folders in parallel
    const [tagsRes, foldersRes] = await Promise.all([
        new Promise(resolve => chrome.runtime.sendMessage({ action: "fetchTags" }, res => resolve(res || {}))),
        new Promise(resolve => chrome.runtime.sendMessage({ action: "fetchFolders" }, res => resolve(res || {})))
    ]);

    const tags = tagsRes && tagsRes.success ? tagsRes.data : [];
    const folders = foldersRes && foldersRes.success ? foldersRes.data : [];

    // List of common supported languages
    const supportedLanguages = [
        'plaintext', 'javascript', 'typescript', 'python', 'go', 'rust', 'java',
        'c', 'cpp', 'csharp', 'php', 'ruby', 'swift', 'kotlin', 'scala',
        'html', 'css', 'scss', 'json', 'yaml', 'xml', 'markdown',
        'sql', 'bash', 'shell', 'powershell', 'dockerfile', 'cuda', 'tex', 'bibtex'
    ];

    // Ensure detected language is in list or add it if valid, otherwise fallback
    let safeLanguage = snippetData.language || 'plaintext';

    // Helper to get extension
    const getExt = (lang) => {
        const extMap = {
            'cuda': 'cu', 'tex': 'tex', 'bibtex': 'bib',
            'plaintext': 'txt', 'javascript': 'js', 'typescript': 'ts', 'python': 'py',
            'go': 'go', 'rust': 'rs', 'java': 'java', 'c': 'c', 'cpp': 'cpp',
            'csharp': 'cs', 'php': 'php', 'ruby': 'rb', 'swift': 'swift',
            'kotlin': 'kt', 'scala': 'scala', 'html': 'html', 'css': 'css',
            'scss': 'scss', 'json': 'json', 'yaml': 'yaml', 'xml': 'xml',
            'markdown': 'md', 'sql': 'sql', 'bash': 'sh', 'shell': 'sh',
            'powershell': 'ps1', 'dockerfile': 'dockerfile'
        };
        return extMap[lang] || 'txt';
    };

    // Default Filename Logic
    const generateDefaultFilename = (lang) => `snippet.${getExt(lang)}`;
    let currentFilename = generateDefaultFilename(safeLanguage);

    // 2. Create Modal HTML using safe DOM methods
    const overlay = document.createElement('div');
    overlay.className = 'snipo-modal-overlay';

    const modal = document.createElement('div');
    modal.className = 'snipo-modal';

    const h2 = document.createElement('h2');
    h2.textContent = 'Save Snippet';
    modal.appendChild(h2);

    // Title field
    const titleGroup = document.createElement('div');
    titleGroup.className = 'snipo-form-group';
    const titleLabel = document.createElement('label');
    titleLabel.textContent = 'Title';
    const titleInput = document.createElement('input');
    titleInput.type = 'text';
    titleInput.id = 'snipo-title';
    titleInput.className = 'snipo-input';
    titleInput.value = snippetData.title;
    titleGroup.appendChild(titleLabel);
    titleGroup.appendChild(titleInput);
    modal.appendChild(titleGroup);

    // Language and Filename row
    const rowGroup = document.createElement('div');
    rowGroup.className = 'snipo-form-group-row';

    const langGroup = document.createElement('div');
    langGroup.className = 'snipo-form-group half-width';
    const langLabel = document.createElement('label');
    langLabel.textContent = 'Language';
    const langSelect = document.createElement('select');
    langSelect.id = 'snipo-language';
    langSelect.className = 'snipo-select';
    supportedLanguages.forEach(lang => {
        const option = document.createElement('option');
        option.value = lang;
        option.textContent = lang.charAt(0).toUpperCase() + lang.slice(1);
        if (lang === safeLanguage) option.selected = true;
        langSelect.appendChild(option);
    });
    if (!supportedLanguages.includes(safeLanguage)) {
        const option = document.createElement('option');
        option.value = safeLanguage;
        option.textContent = safeLanguage + ' (Detected)';
        option.selected = true;
        langSelect.appendChild(option);
    }
    langGroup.appendChild(langLabel);
    langGroup.appendChild(langSelect);

    const filenameGroup = document.createElement('div');
    filenameGroup.className = 'snipo-form-group half-width';
    const filenameLabel = document.createElement('label');
    filenameLabel.textContent = 'Filename';
    const filenameInput = document.createElement('input');
    filenameInput.type = 'text';
    filenameInput.id = 'snipo-filename';
    filenameInput.className = 'snipo-input';
    filenameInput.value = currentFilename;
    filenameGroup.appendChild(filenameLabel);
    filenameGroup.appendChild(filenameInput);

    rowGroup.appendChild(langGroup);
    rowGroup.appendChild(filenameGroup);
    modal.appendChild(rowGroup);

    // Description field
    const descGroup = document.createElement('div');
    descGroup.className = 'snipo-form-group';
    const descLabel = document.createElement('label');
    descLabel.textContent = 'Description';
    const descTextarea = document.createElement('textarea');
    descTextarea.id = 'snipo-desc';
    descTextarea.className = 'snipo-textarea';
    descTextarea.textContent = 'Saved from ' + snippetData.sourceURL;
    descGroup.appendChild(descLabel);
    descGroup.appendChild(descTextarea);
    modal.appendChild(descGroup);

    // Folder field
    const folderGroup = document.createElement('div');
    folderGroup.className = 'snipo-form-group';
    const folderLabel = document.createElement('label');
    folderLabel.textContent = 'Folder';
    const folderSelect = document.createElement('select');
    folderSelect.id = 'snipo-folder';
    folderSelect.className = 'snipo-select';
    const noFolderOption = document.createElement('option');
    noFolderOption.value = '';
    noFolderOption.textContent = 'No Folder';
    folderSelect.appendChild(noFolderOption);
    function appendFolders(items, parentPath = '') {
        items.forEach(folder => {
            const path = parentPath ? `${parentPath} / ${folder.name}` : folder.name;
            const option = document.createElement('option');
            option.value = folder.id;
            option.textContent = path;
            folderSelect.appendChild(option);
            if (folder.children?.length) appendFolders(folder.children, path);
        });
    }
    appendFolders(folders);
    folderGroup.appendChild(folderLabel);
    folderGroup.appendChild(folderSelect);
    modal.appendChild(folderGroup);

    // Tags field
    const tagsGroup = document.createElement('div');
    tagsGroup.className = 'snipo-form-group';
    tagsGroup.style.position = 'relative';
    const tagsLabel = document.createElement('label');
    tagsLabel.textContent = 'Tags';
    const tagInputContainer = document.createElement('div');
    tagInputContainer.className = 'snipo-tag-input-container';
    const tagsInput = document.createElement('input');
    tagsInput.type = 'text';
    tagsInput.id = 'snipo-tags';
    tagsInput.className = 'snipo-input';
    tagsInput.placeholder = 'Type to add tags...';
    tagsInput.autocomplete = 'off';
    const tagsList = document.createElement('div');
    tagsList.id = 'snipo-tags-list';
    tagsList.className = 'snipo-tags-list';
    tagInputContainer.appendChild(tagsInput);
    tagInputContainer.appendChild(tagsList);
    const tagSuggestions = document.createElement('div');
    tagSuggestions.id = 'snipo-tag-suggestions';
    tagSuggestions.className = 'snipo-suggestions-dropdown';
    tagsGroup.appendChild(tagsLabel);
    tagsGroup.appendChild(tagInputContainer);
    tagsGroup.appendChild(tagSuggestions);
    modal.appendChild(tagsGroup);

    const kindGroup = document.createElement('label');
    kindGroup.className = 'snipo-form-group';
    kindGroup.textContent = 'Save as';
    const kindSelect = document.createElement('select');
    kindSelect.className = 'snipo-select';
    kindSelect.add(new Option('Snippet', 'snippet'));
    kindSelect.add(new Option('Note', 'note'));
    kindGroup.append(kindSelect);
    rowGroup.before(kindGroup);

    const formatGroup = document.createElement('label');
    formatGroup.className = 'snipo-form-group';
    formatGroup.textContent = 'Format';
    const noteFormat = document.createElement('select');
    noteFormat.className = 'snipo-select';
    noteFormat.add(new Option('Markdown', 'markdown'));
    noteFormat.add(new Option('Plain text', 'plaintext'));
    formatGroup.append(noteFormat);
    formatGroup.hidden = true;
    rowGroup.after(formatGroup);

    const noteBodyGroup = document.createElement('label');
    noteBodyGroup.className = 'snipo-form-group';
    noteBodyGroup.textContent = 'Note body';
    const noteBody = document.createElement('textarea');
    noteBody.className = 'snipo-textarea';
    noteBody.rows = 6;
    noteBody.dir = 'auto';
    noteBody.value = snippetData.code;
    noteBodyGroup.append(noteBody);
    noteBodyGroup.hidden = true;
    formatGroup.after(noteBodyGroup);

    const sourceGroup = document.createElement('label');
    sourceGroup.className = 'snipo-source-option';
    const includeSource = document.createElement('input');
    includeSource.type = 'checkbox';
    includeSource.checked = true;
    sourceGroup.append(includeSource, ' Include source link');
    sourceGroup.hidden = true;
    noteBodyGroup.after(sourceGroup);

    // Action buttons
    const actionsDiv = document.createElement('div');
    actionsDiv.className = 'snipo-modal-actions';
    const cancelBtn = document.createElement('button');
    cancelBtn.id = 'snipo-cancel';
    cancelBtn.className = 'snipo-btn-secondary';
    cancelBtn.textContent = 'Cancel';
    const saveBtn = document.createElement('button');
    saveBtn.id = 'snipo-save';
    saveBtn.className = 'snipo-btn-primary';
    const saveLabel = () => kindSelect.value === 'note' ? 'Save Note' : 'Save Snippet';
    saveBtn.textContent = saveLabel();
    kindSelect.addEventListener('change', () => {
        const note = kindSelect.value === 'note';
        templateCapture.setNoteMode(note);
        rowGroup.hidden = note;
        descGroup.hidden = note;
        formatGroup.hidden = noteBodyGroup.hidden = sourceGroup.hidden = !note;
        h2.textContent = saveLabel();
        saveBtn.textContent = saveLabel();
    });
    actionsDiv.appendChild(cancelBtn);
    actionsDiv.appendChild(saveBtn);
    modal.appendChild(actionsDiv);

    overlay.appendChild(modal);

    const templateCapture = setupTemplateCapture(snippetData, {
        h2, kindGroup, rowGroup, titleInput, descTextarea, tagsInput, folderSelect, saveBtn
    });
    document.body.appendChild(overlay);

    langSelect.addEventListener('change', (e) => {
        const newLang = e.target.value;
        const currentVal = filenameInput.value;
        const currentExt = currentVal.split('.').pop();

        // Simple logic: if filename looks like default (snippet.ext), update fully.
        // If custom, try to update extension.
        if (currentVal.startsWith('snippet.')) {
            filenameInput.value = generateDefaultFilename(newLang);
        } else {
            // Replace extension
            const newExt = getExt(newLang);
            // If we can identify the old extension, swap it.
            // If no dot, append.
            if (currentVal.includes('.')) {
                const parts = currentVal.split('.');
                parts.pop();
                filenameInput.value = parts.join('.') + '.' + newExt;
            } else {
                filenameInput.value = currentVal + '.' + newExt;
            }
        }
    });

    // Animation
    requestAnimationFrame(() => overlay.classList.add('visible'));

    // Event Listeners
    const close = () => {
        modalOpen = false;
        document.removeEventListener('keydown', onDialogKeyDown, true);
        overlay.classList.remove('visible');
        window.close();
    };

    const onDialogKeyDown = event => {
        if (event.key === 'Escape') {
            event.preventDefault();
            event.stopPropagation();
            close();
        }
    };
    addAppendNotePane(modal, actionsDiv, close);
    document.addEventListener('keydown', onDialogKeyDown, true);
    titleInput.focus();

    const suggestionsBox = overlay.querySelector('#snipo-tag-suggestions');

    // Autocomplete Logic
    tagsInput.addEventListener('input', () => {
        const val = tagsInput.value;
        const tokens = val.split(',');
        const currentToken = tokens[tokens.length - 1].trim().toLowerCase();

        if (!currentToken) {
            suggestionsBox.classList.remove('visible');
            return;
        }

        // Filter tags
        const usedTags = tokens.slice(0, -1).map(t => t.trim().toLowerCase());
        const matches = tags.filter(t =>
            t.name.toLowerCase().includes(currentToken) &&
            !usedTags.includes(t.name.toLowerCase())
        );

        if (matches.length === 0) {
            suggestionsBox.classList.remove('visible');
            return;
        }

        suggestionsBox.textContent = '';
        matches.forEach(t => {
            const name = t.name;
            const idx = name.toLowerCase().indexOf(currentToken);
            const part1 = name.substring(0, idx);
            const part2 = name.substring(idx, idx + currentToken.length);
            const part3 = name.substring(idx + currentToken.length);

            const item = document.createElement('div');
            item.className = 'snipo-suggestion-item';
            item.dataset.val = name;

            item.appendChild(document.createTextNode(part1));
            const matchSpan = document.createElement('span');
            matchSpan.className = 'match';
            matchSpan.textContent = part2;
            item.appendChild(matchSpan);
            item.appendChild(document.createTextNode(part3));

            suggestionsBox.appendChild(item);
        });

        suggestionsBox.classList.add('visible');
    });

    suggestionsBox.addEventListener('click', (e) => {
        const item = e.target.closest('.snipo-suggestion-item');
        if (item) {
            const selectedTag = item.getAttribute('data-val');
            const tokens = tagsInput.value.split(',');
            tokens.pop();
            tokens.push(selectedTag);
            tagsInput.value = tokens.join(', ') + ', ';
            tagsInput.focus();
            suggestionsBox.classList.remove('visible');
        }
    });

    overlay.querySelector('#snipo-cancel').addEventListener('click', close);
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) close();
        // Close suggestions if clicked outside
        if (!tagsInput.contains(e.target) && !suggestionsBox.contains(e.target)) {
            suggestionsBox.classList.remove('visible');
        }
    });

    overlay.querySelector('#snipo-save').addEventListener('click', async () => {
        const note = kindSelect.value === 'note';
        if (note && !noteBody.value.trim()) {
            showToast(false, 'Write some content before saving the note.');
            noteBody.focus();
            return;
        }
        showToast(true, '');
        const title = overlay.querySelector('#snipo-title').value;
        const language = note ? noteFormat.value : langSelect.value;
        const filename = overlay.querySelector('#snipo-filename').value; // Get filename
        const description = overlay.querySelector('#snipo-desc').value;
        const folderId = overlay.querySelector('#snipo-folder').value || null;
        const tagsInputVal = overlay.querySelector('#snipo-tags').value;
        const tagNames = tagsInputVal.split(',').map(t => t.trim()).filter(t => t);

        const saveBtn = overlay.querySelector('#snipo-save');
        if (saveBtn.disabled) return;
        saveBtn.innerText = 'Saving...';
        templateCapture.setSaving(true);

        // Backend expects array of strings for tags
        // It handles creation/lookup automatically

        // Send Save Request
        const rendered = templateCapture.snippet();
        const request = rendered ? {
            action: 'saveCaptureTemplate',
            draftId,
            snippet: { ...rendered, title, description, folder_id: Number(folderId) || null, tags: tagNames }
        } : {
            action: "saveDraftSnippet",
            kind: kindSelect.value,
            draftId: draftId,
            noteText: note ? noteBody.value : undefined,
            includeSource: note ? includeSource.checked : undefined,
            language: language,
            title: title,
            description: description,
            folder_id: folderId,
            tags: tagNames, // Send Names
            filename: filename // Send filename
        };
        chrome.runtime.sendMessage(request, (response) => {
            if (chrome.runtime.lastError) {
                console.error("Save failed:", chrome.runtime.lastError.message);
                saveBtn.innerText = saveLabel();
                templateCapture.setSaving(false);
                showToast(false, "Failed to connect to background script.");
                return;
            }
            if (response && response.success) {
                close();
            } else {
                saveBtn.innerText = saveLabel();
                templateCapture.setSaving(false);
                showToast(false, response ? response.error : "Unknown error");
            }
        });
    });
}
// ----------------

function captureRequest(action, input = {}) {
    return new Promise((resolve, reject) => {
        chrome.runtime.sendMessage({ action, draftId, ...input }, response => {
            const failure = chrome.runtime.lastError?.message;
            if (failure || !response?.success) reject(new Error(failure || response?.error || 'Template request failed.'));
            else resolve(response.data);
        });
    });
}

let captureControlID = 0;
function captureControl(labelText, control) {
    const group = document.createElement('div');
    group.className = 'snipo-form-group';
    const label = document.createElement('label');
    control.id ||= 'snipo-template-field-' + (++captureControlID);
    label.htmlFor = control.id;
    label.textContent = labelText;
    group.append(label, control);
    return group;
}

function captureOption(value, label) {
    const option = document.createElement('option');
    option.value = value;
    option.textContent = label;
    return option;
}

function setupTemplateCapture(source, { h2, kindGroup, rowGroup, titleInput, descTextarea, tagsInput, folderSelect, saveBtn }) {
    const selector = document.createElement('select');
    selector.className = 'snipo-select';
    selector.id = 'snipo-template-picker';
    selector.appendChild(captureOption('', 'Save selection'));
    const picker = captureControl('Template', selector);
    picker.hidden = true;
    h2.after(picker);

    const valuesArea = document.createElement('div');
    const fileArea = document.createElement('div');
    const detail = document.createElement('p');
    detail.className = 'snipo-template-purpose';
    picker.after(detail, valuesArea);
    rowGroup.after(fileArea);

    const initial = () => [titleInput.value, descTextarea.value, tagsInput.value, folderSelect.value];
    let blankMetadata;
    let draftMetadata;
    let draft;
    let files = [];
    let selection = '';
    let request = 0;
    let saving = false;
    let available = false;
    let supported = false;
    let noteMode = false;

    const hasDraftEdits = () => draft && (
        initial().some((value, index) => value !== draftMetadata[index]) ||
        files.some(({ filename, area, originalName, originalContent }) =>
            filename.value !== originalName || area.value !== originalContent)
    );

    captureRequest('listCaptureTemplates').then(items => {
        if (items === null) return;
        supported = true;
        picker.hidden = noteMode;
        for (const item of items) {
            const option = captureOption(item.id, item.name);
            option.dataset.revision = item.revision;
            option.dataset.purpose = item.purpose;
            option.dataset.filenames = item.filenames.join(', ');
            selector.appendChild(option);
        }
        available = items.length > 0;
        selector.disabled = saving || !available;
        if (!available) detail.textContent = 'Add templates in the Snipo web library.';
    }).catch(() => {
        detail.textContent = 'Could not load templates. Check your Snipo connection in extension settings.';
    });

    selector.addEventListener('change', async () => {
        if (saving || (hasDraftEdits() && !window.confirm('Discard your edits and choose a different template?'))) {
            selector.value = selection;
            return;
        }
        selection = selector.value;
        kindGroup.hidden = Boolean(selection);
        const current = ++request;
        if (!blankMetadata) blankMetadata = initial();
        [titleInput.value, descTextarea.value, tagsInput.value, folderSelect.value] = blankMetadata;
        draft = undefined;
        fileArea.replaceChildren();
        valuesArea.replaceChildren();
        rowGroup.hidden = Boolean(selector.value);
        rowGroup.style.display = selector.value ? 'none' : '';
        saveBtn.disabled = Boolean(selector.value);
        const selected = selector.selectedOptions[0];
        detail.textContent = [selected.dataset.purpose, selected.dataset.filenames].filter(Boolean).join(' · ');
        document.querySelector('#save-status').textContent = '';
        if (!selector.value) {
            blankMetadata = undefined;
            return;
        }

        try {
            const template = await captureRequest('getCaptureTemplate', { id: selector.value, revision: Number(selected.dataset.revision) });
            if (current !== request) return;
            const fields = template.definition.fields || [];
            const presets = template.definition.presets || [];
            const defaults = { selection: source.code, page_title: source.pageTitle || source.title, url: source.sourceURL };
            const controls = new Map();

            const gather = async () => {
                for (const control of controls.values()) if (!control.reportValidity()) return;
                const values = Object.fromEntries([...controls].map(([key, control]) => [key, control.value]));
                const rendered = await captureRequest('renderCaptureTemplate', { id: template.id, revision: template.revision, values });
                if (current !== request) return;
                draft = rendered;
                titleInput.value = draft.title;
                descTextarea.value = draft.description;
                tagsInput.value = (draft.tags || []).join(', ');
                folderSelect.value = draft.folder_id ? String(draft.folder_id) : '';
                draftMetadata = initial();
                fileArea.replaceChildren();
                files = draft.files.map((file, index) => {
                    const preview = document.createElement('details');
                    preview.open = index === 0;
                    const name = document.createElement('summary');
                    name.textContent = file.filename;
                    const filename = document.createElement('input');
                    filename.className = 'snipo-input';
                    filename.value = file.filename;
                    filename.addEventListener('input', () => { name.textContent = filename.value || file.filename; });
                    const area = document.createElement('textarea');
                    area.className = 'snipo-textarea snipo-template-content';
                    area.rows = 8;
                    area.spellcheck = false;
                    area.value = file.content;
                    preview.append(name, captureControl('Filename', filename), captureControl('Content (' + file.language + ')', area));
                    fileArea.appendChild(preview);
                    return { file, filename, area, originalName: filename.value, originalContent: area.value };
                });
                if (fields.length) {
                    const changeValues = document.createElement('button');
                    changeValues.type = 'button';
                    changeValues.className = 'snipo-btn-secondary';
                    changeValues.textContent = 'Change template values…';
                    changeValues.addEventListener('click', () => {
                        valuesArea.hidden = false;
                        controls.values().next().value?.focus();
                    });
                    fileArea.prepend(changeValues);
                    preview.textContent = 'Regenerate draft';
                }
                valuesArea.hidden = true;
                saveBtn.disabled = false;
                titleInput.focus();
            };

            valuesArea.hidden = false;
            if (presets.length) {
                const preset = document.createElement('select');
                preset.className = 'snipo-select';
                preset.appendChild(captureOption('', 'Template defaults'));
                presets.forEach((entry, index) => preset.appendChild(captureOption(index, entry.name)));
                valuesArea.appendChild(captureControl('Preset', preset));
                preset.addEventListener('change', () => {
                    const values = presets[preset.value]?.values || {};
                    for (const field of fields) {
                        controls.get(field.key).value = Object.hasOwn(values, field.key)
                            ? values[field.key] : defaults[field.source] ?? field.default ?? '';
                    }
                });
            }
            for (const field of fields) {
                let control;
                if (field.type === 'choice' || field.type === 'toggle') {
                    control = document.createElement('select');
                    control.className = 'snipo-select';
                    control.appendChild(captureOption('', 'Choose…'));
                    const options = field.type === 'toggle' ? ['true', 'false'] : field.choices;
                    for (const option of options) control.appendChild(captureOption(option, option));
                } else if (field.type === 'multiline') {
                    control = document.createElement('textarea');
                    control.className = 'snipo-textarea';
                    control.rows = 3;
                } else {
                    control = document.createElement('input');
                    control.className = 'snipo-input';
                    control.type = field.type === 'date' ? 'date' : 'text';
                    if (field.type === 'number') control.inputMode = 'decimal';
                }
                control.required = Boolean(field.required);
                control.value = defaults[field.source] ?? field.default ?? '';
                if (field.help) control.title = field.help;
                controls.set(field.key, control);
                valuesArea.appendChild(captureControl(field.label, control));
            }

            if (!fields.length) {
                await gather();
                return;
            }
            const preview = document.createElement('button');
            preview.type = 'button';
            preview.className = 'snipo-btn-secondary';
            preview.textContent = 'Preview draft';
            preview.addEventListener('click', async () => {
                if (saving || (hasDraftEdits() && !window.confirm('Replace the edited draft with a fresh render?'))) return;
                preview.disabled = true;
                saveBtn.disabled = true;
                document.querySelector('#save-status').textContent = '';
                try {
                    await gather();
                } catch (error) {
                    if (current === request) showToast(false, error.message);
                } finally {
                    preview.disabled = false;
                    if (current === request) saveBtn.disabled = !draft;
                }
            });
            valuesArea.appendChild(preview);
            controls.values().next().value?.focus();
        } catch (error) {
            if (current === request) showToast(false, error.message);
        }
    });

    return {
        setNoteMode(value) {
            noteMode = value;
            picker.hidden = value || !supported;
            detail.hidden = value;
        },
        setSaving(value) {
            saving = value;
            kindGroup.querySelector('select').disabled = saving;
            selector.disabled = saving || !available;
            saveBtn.disabled = saving || Boolean(selection && !draft);
        },
        snippet() {
            if (!draft) return null;
            const edited = files.map(({ file, filename, area }) => ({ ...file, filename: filename.value, content: area.value }));
            return { ...draft, files: edited, content: '', language: edited[0].language };
        }
    };
}
