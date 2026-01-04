document.addEventListener('DOMContentLoaded', restoreOptions);
document.getElementById('settings-form').addEventListener('submit', saveOptions);

const statusDiv = document.getElementById('status');
const saveBtn = document.getElementById('save-btn');
const testBtn = document.getElementById('test-btn');

function showStatus(message, type = 'success') {
    statusDiv.textContent = message;
    statusDiv.className = `status-message ${type}`;
    statusDiv.classList.remove('hidden');

    if (type === 'success') {
        setTimeout(() => {
            statusDiv.classList.add('hidden');
        }, 3000);
    }
}

// Refactored connection check to be reusable
async function verifyConnection(instanceUrl, apiKey) {
    // Normalized URL: remove trailing slash
    const normalizedUrl = instanceUrl.replace(/\/+$/, '');

    // Validate URL format roughly
    try {
        new URL(normalizedUrl);
    } catch (e) {
        throw new Error('Invalid Instance URL.');
    }

    const response = await fetch(`${normalizedUrl}/api/v1/snippets?limit=1`, {
        method: 'GET',
        headers: {
            'X-API-Key': apiKey,
            'Content-Type': 'application/json'
        }
    });

    if (response.ok) {
        return true;
    } else {
        let errorMsg = 'Connection failed.';
        if (response.status === 401 || response.status === 403) {
            errorMsg = 'Invalid API Key or insufficient permissions.';
        } else if (response.status === 404) {
            errorMsg = 'API endpoint not found. Check Instance URL.';
        } else {
            try {
                const data = await response.json();
                if (data.error && data.error.message) {
                    errorMsg = data.error.message;
                }
            } catch (e) { }
        }
        throw new Error(errorMsg);
    }
}

document.getElementById('test-btn').addEventListener('click', async () => {
    const instanceUrl = document.getElementById('instanceUrl').value;
    const apiKey = document.getElementById('apiKey').value;

    if (!instanceUrl || !apiKey) {
        showStatus('Please fill in all fields.', 'error');
        return;
    }

    testBtn.textContent = 'Testing...';
    testBtn.disabled = true;

    try {
        await verifyConnection(instanceUrl, apiKey);
        showStatus('Connection successful!', 'success');
    } catch (error) {
        let displayMessage = `Error: ${error.message}`;
        if (error.message.includes('NetworkError') || error.message.includes('Failed to fetch')) {
            displayMessage = "Could not connect to Snipo server.\nPlease check the Instance URL and ensure the server is running.";
            if (instanceUrl.includes('localhost') || instanceUrl.includes('127.0.0.1')) {
                displayMessage += "\n(Note: Ensure your server allows CORS if needed.)";
            }
        }
        showStatus(displayMessage, 'error');
    } finally {
        testBtn.textContent = 'Test Connection';
        testBtn.disabled = false;
    }
});

async function saveOptions(e) {
    e.preventDefault();

    let instanceUrl = document.getElementById('instanceUrl').value;
    const apiKey = document.getElementById('apiKey').value;
    const interactiveMode = document.getElementById('interactiveMode').checked;

    if (!instanceUrl || !apiKey) {
        showStatus('Please fill in all fields.', 'error');
        return;
    }

    saveBtn.textContent = 'Verifying connection...';
    saveBtn.disabled = true;

    try {
        await verifyConnection(instanceUrl, apiKey);

        // If successful
        instanceUrl = instanceUrl.replace(/\/+$/, ''); // Ensure saved one is clean

        chrome.storage.sync.set(
            { instanceUrl, apiKey, interactiveMode },
            () => {
                saveBtn.textContent = 'Save Configuration';
                saveBtn.disabled = false;
                showStatus('Connection successful! Settings saved.');
            }
        );
    } catch (error) {
        console.error('Connection Check Error:', error);
        saveBtn.textContent = 'Save Configuration';
        saveBtn.disabled = false;

        let displayMessage = `Error: ${error.message}`;

        // Handle common connectivity errors
        if (error.message.includes('NetworkError') || error.message.includes('Failed to fetch')) {
            displayMessage = "Could not connect to Snipo server.\nPlease check the Instance URL and ensure the server is running.";
            // Add hint if using localhost
            if (instanceUrl.includes('localhost') || instanceUrl.includes('127.0.0.1')) {
                displayMessage += "\n(Note: Ensure your server allows CORS if needed, though the extension permissions should normally bypass this.)";
            }
        }

        showStatus(displayMessage, 'error');
    }
}

// Set Version
const manifest = chrome.runtime.getManifest();
const versionSpan = document.getElementById('ext-version');
if (versionSpan) {
    versionSpan.textContent = manifest.version;
}

// Tab Switching Logic
document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        // Remove active class from all
        document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));

        // Add to current
        btn.classList.add('active');
        const tabId = btn.getAttribute('data-tab');
        document.getElementById(tabId).classList.add('active');
    });
});

// --- Ignored Sites Logic ---
let ignoredSites = [];

function renderIgnoredList() {
    const list = document.getElementById('ignored-list');
    list.textContent = '';

    if (ignoredSites.length === 0) {
        const emptyDiv = document.createElement('div');
        emptyDiv.className = 'empty-state';
        emptyDiv.textContent = 'No websites ignored. Snipo runs everywhere.';
        list.appendChild(emptyDiv);
        return;
    }

    ignoredSites.forEach(site => {
        const item = document.createElement('div');
        item.className = 'ignored-item';
        
        const domainSpan = document.createElement('span');
        domainSpan.className = 'ignored-domain';
        domainSpan.textContent = site;
        
        const removeBtn = document.createElement('button');
        removeBtn.className = 'remove-btn';
        removeBtn.title = 'Remove';
        removeBtn.dataset.site = site;
        
        const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        svg.setAttribute('width', '16');
        svg.setAttribute('height', '16');
        svg.setAttribute('viewBox', '0 0 24 24');
        svg.setAttribute('fill', 'none');
        svg.setAttribute('stroke', 'currentColor');
        svg.setAttribute('stroke-width', '2');
        svg.setAttribute('stroke-linecap', 'round');
        svg.setAttribute('stroke-linejoin', 'round');
        
        const line1 = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        line1.setAttribute('x1', '18');
        line1.setAttribute('y1', '6');
        line1.setAttribute('x2', '6');
        line1.setAttribute('y2', '18');
        
        const line2 = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        line2.setAttribute('x1', '6');
        line2.setAttribute('y1', '6');
        line2.setAttribute('x2', '18');
        line2.setAttribute('y2', '18');
        
        svg.appendChild(line1);
        svg.appendChild(line2);
        removeBtn.appendChild(svg);
        
        item.appendChild(domainSpan);
        item.appendChild(removeBtn);
        list.appendChild(item);
    });

    // Attach listeners
    document.querySelectorAll('.remove-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const siteToRemove = e.currentTarget.getAttribute('data-site');
            removeIgnoredSite(siteToRemove);
        });
    });
}

function addIgnoredSite() {
    const input = document.getElementById('new-ignored-site');
    let site = input.value.trim().toLowerCase();

    if (!site) return;

    // Simple cleanup: remove http(s)://, www., path
    try {
        if (!site.includes('://')) site = 'http://' + site; // dummy protocol for URL parser
        site = new URL(site).hostname;
    } catch (e) {
        // Fallback to raw input if simple hostname extraction fails
    }

    // Check duplicates
    if (ignoredSites.includes(site)) {
        showStatus('Site already ignored.', 'error');
        input.value = '';
        return;
    }

    ignoredSites.push(site);
    saveIgnoredSites();
    input.value = '';
    renderIgnoredList();
}

function removeIgnoredSite(site) {
    ignoredSites = ignoredSites.filter(s => s !== site);
    saveIgnoredSites();
    renderIgnoredList();
}

function saveIgnoredSites() {
    chrome.storage.sync.set({ ignoredSites }, () => {
        // showStatus('Ignored list updated.'); // Optional: feedback
    });
}

document.getElementById('add-ignored-btn').addEventListener('click', addIgnoredSite);
document.getElementById('new-ignored-site').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') addIgnoredSite();
});

async function updateStatusTab(items) {
    const statusDot = document.getElementById('status-dot');
    const statusText = document.getElementById('connection-text');
    const instanceText = document.getElementById('info-instance');

    if (!items.instanceUrl || !items.apiKey) {
        statusText.textContent = 'Not Configured';
        statusDot.className = 'status-dot disconnected';
        instanceText.textContent = '-';
        return;
    }

    instanceText.textContent = items.instanceUrl;
    statusText.textContent = 'Checking...';

    try {
        await verifyConnection(items.instanceUrl, items.apiKey);
        statusText.textContent = 'Connected';
        statusDot.className = 'status-dot connected';
    } catch (e) {
        statusText.textContent = 'Connection Failed';
        statusDot.className = 'status-dot disconnected';
    }
}

function restoreOptions() {
    chrome.storage.sync.get(
        { instanceUrl: '', apiKey: '', interactiveMode: false, ignoredSites: [] },
        (items) => {
            document.getElementById('instanceUrl').value = items.instanceUrl;
            document.getElementById('apiKey').value = items.apiKey;
            document.getElementById('interactiveMode').checked = items.interactiveMode;

            ignoredSites = items.ignoredSites || [];
            renderIgnoredList();

            updateStatusTab(items);
        }
    );
}
