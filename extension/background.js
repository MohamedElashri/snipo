// Background Script for Snipo Extension

// Create context menu on installation
chrome.runtime.onInstalled.addListener(() => {
    chrome.contextMenus.create({
        id: "save-to-snipo",
        title: "Save to Snipo",
        contexts: ["selection"]
    });
});

// Handle Context Menu Clicks
// Handle Context Menu Clicks
chrome.contextMenus.onClicked.addListener((info, tab) => {
    if (info.menuItemId === "save-to-snipo" && info.selectionText) {
        // Delegate to content script to respect Interactive Mode setting
        chrome.tabs.sendMessage(tab.id, {
            action: "contextMenuSave",
            code: info.selectionText,
            title: tab.title
        }).catch(err => {
            console.warn("Could not send context menu action to content script. Tab might need reload.", err);
            // Fallback to quick save if content script is missing? 
            saveSnippet(info.selectionText, tab, "plaintext", tab.title);
        });
    }
});

// Handle Messages from Content Script
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
    if (request.action === "saveSnippet") {
        saveSnippet(request.code, sender.tab, request.language, request.title, request.folder_id, request.tags, request.description, sendResponse);
        return true; // Indicates async response
    }
    if (request.action === "fetchTags") {
        apiCall(request.action, "GET", "/api/v1/tags", null, sendResponse);
        return true;
    }
    if (request.action === "fetchFolders") {
        apiCall(request.action, "GET", "/api/v1/folders", null, sendResponse);
        return true;
    }
    if (request.action === "createTag") {
        apiCall(request.action, "POST", "/api/v1/tags", { name: request.name }, sendResponse);
        return true;
    }
    if (request.action === "createFolder") {
        apiCall(request.action, "POST", "/api/v1/folders", { name: request.name }, sendResponse);
        return true;
    }
});

async function apiCall(action, method, endpoint, body, sendResponse) {
    try {
        const config = await chrome.storage.sync.get(['instanceUrl', 'apiKey']);
        if (!config.instanceUrl || !config.apiKey) {
            sendResponse({ success: false, error: "Configuration missing" });
            return;
        }

        const url = `${config.instanceUrl}${endpoint}`;
        const options = {
            method: method,
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': config.apiKey
            }
        };

        if (body) {
            options.body = JSON.stringify(body);
        }

        const response = await fetch(url, options);
        if (response.ok) {
            const data = await response.json();
            sendResponse({ success: true, data: data.data || data }); // Handle enveloped or plain
        } else {
            const errorText = await response.text();
            let errorMessage = "API Error";
            try {
                const errJson = JSON.parse(errorText);
                if (errJson.error && errJson.error.message) errorMessage = errJson.error.message;
            } catch (e) { }
            sendResponse({ success: false, error: errorMessage });
        }
    } catch (e) {
        sendResponse({ success: false, error: e.message });
    }
}

async function saveSnippet(code, tab, language = "plaintext", title = null, folder_id = null, tags = [], description = null, sendResponse = null) {
    // Get configuration
    try {
        const config = await chrome.storage.sync.get(['instanceUrl', 'apiKey']);

        if (!config.instanceUrl || !config.apiKey) {
            const errorMsg = "Please configure Snipo Extension options first.";
            if (sendResponse) sendResponse({ success: false, error: errorMsg });
            notifyTab(tab.id, { success: false, error: errorMsg });
            return;
        }

        const finalTitle = title || `Snippet from ${new URL(tab.url).hostname}`;
        const finalDesc = description || `Saved from ${tab.url}`;

        // Construct API URL
        const apiUrl = `${config.instanceUrl}/api/v1/snippets`;

        const payload = {
            title: finalTitle,
            content: code,
            language: language,
            description: finalDesc,
            is_public: false,
            tags: tags
        };

        if (folder_id) {
            payload.folder_id = parseInt(folder_id);
        }

        const response = await fetch(apiUrl, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': config.apiKey,
                'X-API-Version': '1.0'
            },
            body: JSON.stringify(payload)
        });

        if (response.ok) {
            const data = await response.json();
            if (sendResponse) sendResponse({ success: true, data });
            notifyTab(tab.id, { success: true });
        } else {
            const errorText = await response.text();
            console.error("Snipo Save Error:", errorText);
            let errorMessage = "Failed to save snippet.";
            try {
                const errJson = JSON.parse(errorText);
                if (errJson.error && errJson.error.message) {
                    errorMessage = errJson.error.message;
                }
            } catch (e) { }

            if (sendResponse) sendResponse({ success: false, error: errorMessage });
            notifyTab(tab.id, { success: false, error: errorMessage });
        }

    } catch (error) {
        console.error("Snipo Error:", error);
        if (sendResponse) sendResponse({ success: false, error: error.message });
        notifyTab(tab.id, { success: false, error: error.message });
    }
}

// Helper to send message back to content script for UI feedback (Toast)
function notifyTab(tabId, message) {
    chrome.tabs.sendMessage(tabId, { action: "showToast", ...message }).catch(err => {
        // Content script might not be ready or injected
        console.warn("Could not notify tab:", err);
    });
}
