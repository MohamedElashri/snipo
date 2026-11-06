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

        // Prevent double injection if attribute was wiped but DOM persisted
        if (container.querySelector('.snipo-btn')) return;

        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'snipo-btn';
        btn.appendChild(createSVGFromString(SNIPO_ICON));
        btn.title = "Save to Snipo";

        btn.addEventListener('click', (e) => {
            // The page shares this button's DOM and can dispatch synthetic clicks.
            if (!e.isTrusted) return;
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
    const textContent = (codeElement ? codeElement.textContent : clone.textContent);

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

let lastSaveButton = null;

// Unified Save Flow (handles both Button and Context Menu)
function initiateSaveFlow(code, language, title, btn) {
    // Page titles may exceed the server's 200-character limit. Keep user-entered
    // titles in the dialog separate from this automatic page-title default.
    title = Array.from(title || '').slice(0, 200).join('');
    if (btn?.disabled) return;
    if (btn) btn.disabled = true;
    chrome.storage.sync.get(['interactiveMode'], async (items) => {
        if (items.interactiveMode) {
            lastSaveButton = btn;
            chrome.runtime.sendMessage({
                action: 'openSaveDialog', code, language, title
            }, response => {
                if (btn) btn.disabled = false;
                if (chrome.runtime.lastError || !response?.success) {
                    showToast(false, 'Unable to open the save window. Please try again.');
                }
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
                if (chrome.runtime.lastError) {
                    console.error("Save failed:", chrome.runtime.lastError.message);
                    showToast(false, "Failed to connect to background script.");
                }
                if (btn) setTimeout(() => { 
                    btn.disabled = false;
                    btn.textContent = '';
                    btn.appendChild(createSVGFromString(SNIPO_ICON));
                }, 2000);
            });
        }
    });
}

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
    if (request.action === 'saveDialogClosed') {
        lastSaveButton?.focus();
    }
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
