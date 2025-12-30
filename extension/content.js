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

function handleSaveClick(btn, preElement) {
    // Extract code
    // Usually code is inside a <code> tag inside <pre>, but sometimes just in <pre>
    const codeElement = preElement.querySelector('code');
    const textContent = codeElement ? codeElement.innerText : preElement.innerText;

    // Try to detect language from class
    // Format usually language-js, lang-js, etc.
    let language = 'text';
    const classes = (preElement.className + ' ' + (codeElement ? codeElement.className : '')).split(/\s+/);

    for (const cls of classes) {
        if (cls.startsWith('language-') || cls.startsWith('lang-')) {
            language = cls.replace(/^lang(uage)?-/, '');
            break;
        }
    }

    // Animate button
    btn.innerHTML = '...';

    chrome.runtime.sendMessage({
        action: "saveSnippet",
        code: textContent,
        language: language,
        title: document.title
    }, (response) => {
        if (response && response.success) {
            // Success animation handled globally via toast, but we can also update button
            // But let's rely on Toast as it's consistent for Context Menu too
        } else {
            // Error handled via Toast
        }
        // Reset button icon after a delay
        setTimeout(() => {
            btn.innerHTML = SNIPO_ICON;
        }, 2000);
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
