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

// Initialize
function init() {
    chrome.storage.sync.get(['instanceUrl', 'apiKey'], (items) => {
        const instanceUrl = items.instanceUrl;
        const apiKey = items.apiKey;

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
        btn.innerHTML = SNIPO_ICON;
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
            if (btn) btn.innerHTML = '...';
            chrome.runtime.sendMessage({
                action: "saveSnippet",
                code: code,
                language: language,
                title: title
            }, (response) => {
                if (btn) setTimeout(() => { btn.innerHTML = SNIPO_ICON; }, 2000);
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

    // 2. Create Modal HTML
    const overlay = document.createElement('div');
    overlay.className = 'snipo-modal-overlay';

    overlay.innerHTML = `
        <div class="snipo-modal">
            <h2>Save Snippet</h2>
            <div class="snipo-form-group">
                <label>Title</label>
                <input type="text" id="snipo-title" class="snipo-input" value="${snippetData.title.replace(/"/g, '&quot;')}" >
            </div>
            <div class="snipo-form-group-row">
                <div class="snipo-form-group half-width">
                    <label>Language</label>
                    <select id="snipo-language" class="snipo-select">
                        ${supportedLanguages.map(lang => `<option value="${lang}" ${lang === safeLanguage ? 'selected' : ''}>${lang.charAt(0).toUpperCase() + lang.slice(1)}</option>`).join('')}
                        ${!supportedLanguages.includes(safeLanguage) ? `<option value="${safeLanguage}" selected>${safeLanguage} (Detected)</option>` : ''}
                    </select>
                </div>
                <div class="snipo-form-group half-width">
                    <label>Filename</label>
                    <input type="text" id="snipo-filename" class="snipo-input" value="${currentFilename}">
                </div>
            </div>
            <div class="snipo-form-group">
                <label>Description</label>
                <textarea id="snipo-desc" class="snipo-textarea">Saved from ${window.location.href}</textarea>
            </div>
            <!-- ... Folder, Tags ... -->
            <div class="snipo-form-group">
                <label>Folder</label>
                <select id="snipo-folder" class="snipo-select">
                    <option value="">No Folder</option>
                    ${folders.map(f => `<option value="${f.id}">${f.name}</option>`).join('')}
                </select>
            </div>
            <div class="snipo-form-group" style="position: relative;">
                <label>Tags</label>
                <div class="snipo-tag-input-container">
                    <input type="text" id="snipo-tags" class="snipo-input" placeholder="Type to add tags..." autocomplete="off">
                    <div id="snipo-tags-list" class="snipo-tags-list"></div>
                </div>
                <div id="snipo-tag-suggestions" class="snipo-suggestions-dropdown"></div> 
            </div>
            <div class="snipo-modal-actions">
                <button id="snipo-cancel" class="snipo-btn-secondary">Cancel</button>
                <button id="snipo-save" class="snipo-btn-primary">Save Snippet</button>
            </div>
        </div>
    `;

    document.body.appendChild(overlay);
    const langSelect = overlay.querySelector('#snipo-language');
    const filenameInput = overlay.querySelector('#snipo-filename');

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

    const tagsInput = overlay.querySelector('#snipo-tags');
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

        suggestionsBox.innerHTML = matches.map(t => {
            const name = t.name;
            const idx = name.toLowerCase().indexOf(currentToken);
            // safe substring handling
            const part1 = name.substring(0, idx);
            const part2 = name.substring(idx, idx + currentToken.length);
            const part3 = name.substring(idx + currentToken.length);

            return `<div class="snipo-suggestion-item" data-val="${name}">
                ${part1}<span class="match">${part2}</span>${part3}
            </div>`;
        }).join('');

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

    toast.innerHTML = `
        <div class="snipo-toast-icon">
            ${success ? CHECK_ICON : ERROR_ICON}
        </div>
        <span>${message || (success ? 'Snippet saved successfully!' : 'Failed to save snippet.')}</span>
    `;

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
