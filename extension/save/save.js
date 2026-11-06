// This page belongs to the extension and is never exposed to visited websites.
const draftId = new URLSearchParams(window.location.search).get('draft');
function showToast(_success, message) {
    document.querySelector('#save-status').textContent = message || 'Unable to save snippet.';
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
    saveBtn.textContent = 'Save Snippet';
    actionsDiv.appendChild(cancelBtn);
    actionsDiv.appendChild(saveBtn);
    modal.appendChild(actionsDiv);

    overlay.appendChild(modal);

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
        const title = overlay.querySelector('#snipo-title').value;
        const language = overlay.querySelector('#snipo-language').value;
        const filename = overlay.querySelector('#snipo-filename').value; // Get filename
        const description = overlay.querySelector('#snipo-desc').value;
        const folderId = overlay.querySelector('#snipo-folder').value || null;
        const tagsInputVal = overlay.querySelector('#snipo-tags').value;
        const tagNames = tagsInputVal.split(',').map(t => t.trim()).filter(t => t);

        const saveBtn = overlay.querySelector('#snipo-save');
        saveBtn.innerText = 'Saving...';
        saveBtn.disabled = true;

        // Backend expects array of strings for tags
        // It handles creation/lookup automatically

        // Send Save Request
        chrome.runtime.sendMessage({
            action: "saveDraftSnippet",
            draftId: draftId,
            code: snippetData.code,
            language: language,
            title: title,
            description: description,
            folder_id: folderId,
            tags: tagNames, // Send Names
            filename: filename // Send filename
        }, (response) => {
            if (chrome.runtime.lastError) {
                console.error("Save failed:", chrome.runtime.lastError.message);
                saveBtn.innerText = 'Save Snippet';
                saveBtn.disabled = false;
                showToast(false, "Failed to connect to background script.");
                return;
            }
            if (response && response.success) {
                close();
            } else {
                saveBtn.innerText = 'Save Snippet';
                saveBtn.disabled = false;
                showToast(false, response ? response.error : "Unknown error");
            }
        });
    });
}
// ----------------
