document.addEventListener('DOMContentLoaded', restoreOptions);
document.getElementById('settings-form').addEventListener('submit', saveOptions);

const statusDiv = document.getElementById('status');
const saveBtn = document.getElementById('save-btn');

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

async function saveOptions(e) {
    e.preventDefault();

    let instanceUrl = document.getElementById('instanceUrl').value;
    const apiKey = document.getElementById('apiKey').value;

    if (!instanceUrl || !apiKey) {
        showStatus('Please fill in all fields.', 'error');
        return;
    }

    // Normalized URL: remove trailing slash
    instanceUrl = instanceUrl.replace(/\/+$/, '');

    // Validate URL format roughly
    try {
        new URL(instanceUrl);
    } catch (e) {
        showStatus('Invalid Instance URL.', 'error');
        return;
    }

    saveBtn.textContent = 'Verifying connection...';
    saveBtn.disabled = true;

    try {
        // Verify connection and API Key by fetching a single snippet (minimal resource)
        const response = await fetch(`${instanceUrl}/api/v1/snippets?limit=1`, {
            method: 'GET',
            headers: {
                'X-API-Key': apiKey,
                'Content-Type': 'application/json'
            }
        });

        if (response.ok) {
            chrome.storage.sync.set(
                { instanceUrl, apiKey },
                () => {
                    saveBtn.textContent = 'Save Configuration';
                    saveBtn.disabled = false;
                    showStatus('Connection successful! Settings saved.');
                }
            );
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

function restoreOptions() {
    chrome.storage.sync.get(
        { instanceUrl: '', apiKey: '' },
        (items) => {
            document.getElementById('instanceUrl').value = items.instanceUrl;
            document.getElementById('apiKey').value = items.apiKey;
        }
    );
}
