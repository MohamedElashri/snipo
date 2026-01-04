// Snipo Content Script

const SNIPO_ICON = `
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
    <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z"/>
</svg>`;

const CHECK_ICON = `
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#10b981">
    <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
</svg>`;

const ERROR_ICON = `
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#ef4444">
    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
</svg>`;

// Helper function to safely create SVG element from string
function createSVGFromString(svgString) {
    const parser = new DOMParser();
    const doc = parser.parseFromString(svgString, 'image/svg+xml');
    return doc.documentElement;
}

// Initialize
function init() {
    chrome.storage.sync.get(['instanceUrl', 'apiKey', 'ignoredSites'], (items) => {
        const instanceUrl = items.instanceUrl;
        const apiKey = items.apiKey;
        const ignoredSites = items.ignoredSites || [];

        // Check ignored sites
        const currentHostname = window.location.hostname;
        const isIgnored = ignoredSites.some(site => {
            return currentHostname === site || currentHostname.endsWith('.' + site);
        });

        if (isIgnored) {
            console.log(`Snipo Extension: Disabled on ${currentHostname} (Ignored List).`);
            return;
        }

        // Check if configured
        if (!instanceUrl || !apiKey) {
            console.log("Snipo Extension: Not configured. Skipping injection.");
            return;
        }

        // If instance URL is set and matches current origin, treat as self and do not run
        if (instanceUrl) {
            try {
                const currentOrigin = new URL(window.location.href).origin;
                const instanceOrigin = new URL(instanceUrl).origin;

                if (currentOrigin === instanceOrigin) {
                    console.log("Snipo Extension: Disabled on Snipo instance.");
                    return; // Stop initialization
                }
            } catch (e) {
                // Ignore URL parsing errors
            }
        }

        scanForCode();

        // Observer for dynamic content (SPA navigation, etc.)
        const observer = new MutationObserver((mutations) => {
            let shouldScan = false;
            for (const mutation of mutations) {
                if (mutation.addedNodes.length > 0) {
                    shouldScan = true;
                    break;
                }
            }
            if (shouldScan) {
                // Debounce
                clearTimeout(window.snipoScanTimeout);
                window.snipoScanTimeout = setTimeout(scanForCode, 1000);
            }
        });

        observer.observe(document.body, { childList: true, subtree: true });
    });
}

function scanForCode() {
    // Detect <pre> blocks (standard markdown/code blocks)
    // Avoid re-injection
    const preBlocks = document.querySelectorAll('pre:not([data-snipo-injected])');

    preBlocks.forEach(pre => {
        // Mark as injected
        pre.setAttribute('data-snipo-injected', 'true');

        let container = pre;
        let injectStart = false; // default append

        // Smart Parent Detection (GitHub & others)
        // Look for existing wrapper like div.highlight
        const parent = pre.parentElement;
        if (parent &&
            (parent.classList.contains('highlight') ||
                parent.classList.contains('zeroclipboard-container') ||
                parent.tagName === 'DIV' && parent.children.length === 1)) {

            container = parent;
        }

        // Always mark the container for CSS targeting (hover visibility)
        container.classList.add('snipo-container');

        // Ensure container is positioned context for absolute button
        // Check if we need to force relative positioning
        if (getComputedStyle(container).position === 'static') {
            container.classList.add('snipo-wrapper'); // helper class for relative
        }

        const btn = document.createElement('button');
        btn.className = 'snipo-btn';
        btn.appendChild(createSVGFromString(SNIPO_ICON));
        btn.title = "Save to Snipo";

        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            e.preventDefault();
            handleSaveClick(btn, pre); // Always pass the PRE element for code
        });

        // Inject
        if (container === pre) {
            // Inside Pre: Prepend or Append?
            // Append is safer for absolute positioning top-right
            container.appendChild(btn);
        } else {
            // Inside Parent: Append to be on top
            container.appendChild(btn);
        }
    });
}

// Listen for Context Menu Action
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
    if (request.action === "contextMenuSave") {
        initiateSaveFlow(request.code, 'plaintext', request.title, null); // No button element for context menu
    }
});

async function handleSaveClick(btn, preElement) {
    // Extract code safely
    const clone = preElement.cloneNode(true);
    const cloneBtn = clone.querySelector('.snipo-btn');
    if (cloneBtn) cloneBtn.remove();

    const codeElement = clone.querySelector('code');
    const textContent = (codeElement ? codeElement.textContent : clone.textContent).trim();

    // Detect language
    let language = 'plaintext';
    const classes = (preElement.className + ' ' + (codeElement ? codeElement.className : '')).split(/\s+/);

    for (const cls of classes) {
        if (cls.startsWith('language-') || cls.startsWith('lang-')) {
            language = cls.replace(/^lang(uage)?-/, '').toLowerCase();
            break;
        }
    }

    // Normalize Language
    const languageMap = {
        'text': 'plaintext', 'txt': 'plaintext', 'js': 'javascript', 'ts': 'typescript',
        'py': 'python', 'rb': 'ruby', 'sh': 'bash', 'c++': 'cpp', 'c#': 'csharp',
        'golang': 'go', 'yml': 'yaml'
    };
    if (languageMap[language]) language = languageMap[language];

    initiateSaveFlow(textContent, language, document.title, btn);
}

// Unified Save Flow (handles both Button and Context Menu)
function initiateSaveFlow(code, language, title, btn) {
    chrome.storage.sync.get(['interactiveMode'], async (items) => {
        if (items.interactiveMode) {
            showModal({
                code: code,
                language: language,
                title: title
            });
        } else {
            // Quick Save
            if (btn) btn.textContent = '...';
            chrome.runtime.sendMessage({
                action: "saveSnippet",
                code: code,
                language: language,
                title: title
            }, (response) => {
                if (btn) setTimeout(() => { 
                    btn.textContent = '';
                    btn.appendChild(createSVGFromString(SNIPO_ICON));
                }, 2000);
            });
        }
    });
}

// --- MODAL LOGIC ---

async function showModal(snippetData) {
    // 1. Fetch Tags and Folders in parallel
    const [tagsRes, foldersRes] = await Promise.all([
        new Promise(resolve => chrome.runtime.sendMessage({ action: "fetchTags" }, resolve)),
        new Promise(resolve => chrome.runtime.sendMessage({ action: "fetchFolders" }, resolve))
    ]);

    const tags = tagsRes.success ? tagsRes.data : [];
    const folders = foldersRes.success ? foldersRes.data : [];

    // List of common supported languages
    const supportedLanguages = [
        'plaintext', 'javascript', 'typescript', 'python', 'go', 'rust', 'java',
        'c', 'cpp', 'csharp', 'php', 'ruby', 'swift', 'kotlin', 'scala',
        'html', 'css', 'scss', 'json', 'yaml', 'xml', 'markdown',
        'sql', 'bash', 'shell', 'powershell', 'dockerfile'
    ];

    // Ensure detected language is in list or add it if valid, otherwise fallback
    let safeLanguage = snippetData.language || 'plaintext';

    // Helper to get extension
    const getExt = (lang) => {
        const extMap = {
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
    descTextarea.textContent = 'Saved from ' + window.location.href;
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
    folders.forEach(f => {
        const option = document.createElement('option');
        option.value = f.id;
        option.textContent = f.name;
        folderSelect.appendChild(option);
    });
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
        overlay.classList.remove('visible');
        setTimeout(() => overlay.remove(), 200);
    };

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
            action: "saveSnippet",
            code: snippetData.code,
            language: language,
            title: title,
            description: description,
            folder_id: folderId,
            tags: tagNames, // Send Names
            filename: filename // Send filename
        }, (response) => {
            close();
            // Toast will be shown by background msg listener
        });
    });
}
// ----------------

// Toast Notification System
let toastTimeout;

function showToast(success, message) {
    // Remove existing toast
    const existing = document.querySelector('.snipo-toast');
    if (existing) existing.remove();

    const toast = document.createElement('div');
    toast.className = `snipo-toast ${success ? 'success' : 'error'}`;

    const iconDiv = document.createElement('div');
    iconDiv.className = 'snipo-toast-icon';
    iconDiv.appendChild(createSVGFromString(success ? CHECK_ICON : ERROR_ICON));
    
    const messageSpan = document.createElement('span');
    messageSpan.textContent = message || (success ? 'Snippet saved successfully!' : 'Failed to save snippet.');
    
    toast.appendChild(iconDiv);
    toast.appendChild(messageSpan);

    document.body.appendChild(toast);

    // Force reflow
    toast.offsetHeight;

    requestAnimationFrame(() => {
        toast.classList.add('visible');
    });

    if (toastTimeout) clearTimeout(toastTimeout);
    toastTimeout = setTimeout(() => {
        toast.classList.remove('visible');
        setTimeout(() => toast.remove(), 300);
    }, 4000);
}

// Listen for messages from background script
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
    if (request.action === "showToast") {
        showToast(request.success, request.error);
    }
});

// Start
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}
