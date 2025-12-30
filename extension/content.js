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
}

function scanForCode() {
    // Detect <pre> blocks (standard markdown/code blocks)
    // Avoid re-injection
    const preBlocks = document.querySelectorAll('pre:not([data-snipo-injected])');

    preBlocks.forEach(pre => {
        // Mark as injected
        pre.setAttribute('data-snipo-injected', 'true');

        // Check if wrapper needed
        if (getComputedStyle(pre).position === 'static') {
            pre.classList.add('snipo-wrapper');
        } else {
            // Add relative class helper if needed, or rely on existing relative/absolute
            pre.classList.add('snipo-wrapper');
        }

        const btn = document.createElement('button');
        btn.className = 'snipo-btn';
        btn.innerHTML = SNIPO_ICON;
        btn.title = "Save to Snipo";

        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            e.preventDefault();
            handleSaveClick(btn, pre);
        });

        pre.appendChild(btn);
    });
}

async function handleSaveClick(btn, preElement) {
    // Extract code
    const codeElement = preElement.querySelector('code');
    const textContent = codeElement ? codeElement.innerText : preElement.innerText;

    // Detect language
    let language = 'text';
    const classes = (preElement.className + ' ' + (codeElement ? codeElement.className : '')).split(/\s+/);

    for (const cls of classes) {
        if (cls.startsWith('language-') || cls.startsWith('lang-')) {
            language = cls.replace(/^lang(uage)?-/, '');
            break;
        }
    }

    // Check Interactive Mode setting
    chrome.storage.sync.get(['interactiveMode'], async (items) => {
        if (items.interactiveMode) {
            // Interactive Mode: Show Modal
            showModal({
                code: textContent,
                language: language,
                title: document.title
            });
        } else {
            // Quick Save
            btn.innerHTML = '...';
            chrome.runtime.sendMessage({
                action: "saveSnippet",
                code: textContent,
                language: language,
                title: document.title
            }, (response) => {
                setTimeout(() => { btn.innerHTML = SNIPO_ICON; }, 2000);
                if (!response) return; // Error handled by toast listener
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

    // 2. Create Modal HTML
    const overlay = document.createElement('div');
    overlay.className = 'snipo-modal-overlay';

    overlay.innerHTML = `
        <div class="snipo-modal">
            <h2>Save Snippet</h2>
            <div class="snipo-form-group">
                <label>Title</label>
                <input type="text" id="snipo-title" class="snipo-input" value="${snippetData.title.replace(/"/g, '&quot;')}">
            </div>
            <div class="snipo-form-group">
                <label>Description</label>
                <textarea id="snipo-desc" class="snipo-textarea">Saved from ${window.location.href}</textarea>
            </div>
            <div class="snipo-form-group">
                <label>Folder</label>
                <select id="snipo-folder" class="snipo-select">
                    <option value="">No Folder</option>
                    ${folders.map(f => `<option value="${f.id}">${f.name}</option>`).join('')}
                </select>
            </div>
            <div class="snipo-form-group">
                <label>Tags (separate by comma)</label>
                <input type="text" id="snipo-tags" class="snipo-input" placeholder="javascript, react, api">
                <div style="margin-top:5px; font-size:12px; color:#64748b;">Available: ${tags.map(t => t.name).slice(0, 10).join(', ')}${tags.length > 10 ? '...' : ''}</div>
            </div>
            <div class="snipo-modal-actions">
                <button id="snipo-cancel" class="snipo-btn-secondary">Cancel</button>
                <button id="snipo-save" class="snipo-btn-primary">Save Snippet</button>
            </div>
        </div>
    `;

    document.body.appendChild(overlay);

    // Animation
    requestAnimationFrame(() => overlay.classList.add('visible'));

    // Event Listeners
    const close = () => {
        overlay.classList.remove('visible');
        setTimeout(() => overlay.remove(), 200);
    };

    overlay.querySelector('#snipo-cancel').addEventListener('click', close);
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) close();
    });

    overlay.querySelector('#snipo-save').addEventListener('click', async () => {
        const title = overlay.querySelector('#snipo-title').value;
        const description = overlay.querySelector('#snipo-desc').value;
        const folderId = overlay.querySelector('#snipo-folder').value || null;
        const tagsInput = overlay.querySelector('#snipo-tags').value;

        const saveBtn = overlay.querySelector('#snipo-save');
        saveBtn.innerText = 'Saving...';
        saveBtn.disabled = true;

        // Process Tags
        const tagNames = tagsInput.split(',').map(t => t.trim()).filter(t => t);

        // Backend expects array of strings for tags
        // It handles creation/lookup automatically

        // Send Save Request
        chrome.runtime.sendMessage({
            action: "saveSnippet",
            code: snippetData.code,
            language: snippetData.language,
            title: title,
            description: description,
            folder_id: folderId,
            tags: tagNames // Send Names
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
