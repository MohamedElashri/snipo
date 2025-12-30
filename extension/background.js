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
chrome.contextMenus.onClicked.addListener((info, tab) => {
    if (info.menuItemId === "save-to-snipo" && info.selectionText) {
        // Send message to content script to handle the saving (so we can show UI feedback there)
        saveSnippet(info.selectionText, tab);
    }
});

// Handle Messages from Content Script
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
    if (request.action === "saveSnippet") {
        saveSnippet(request.code, sender.tab, request.language, request.title, sendResponse);
        return true; // Indicates async response
    }
});

async function saveSnippet(code, tab, language = "text", title = null, sendResponse = null) {
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

        // Construct API URL
        const apiUrl = `${config.instanceUrl}/api/v1/snippets`;

        const response = await fetch(apiUrl, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': config.apiKey,
                // Add version header just in case
                'X-API-Version': '1.0'
            },
            body: JSON.stringify({
                title: finalTitle,
                content: code,
                language: language,
                description: `Saved from ${tab.url}`,
                is_public: false
            })
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
