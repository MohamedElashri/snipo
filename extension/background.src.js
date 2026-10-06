import { SnipoClient } from '@snipo/api-client';

chrome.runtime.onInstalled.addListener(() => {
    chrome.contextMenus.create({
        id: "save-to-snipo",
        title: "Save to Snipo",
        contexts: ["selection"]
    });
});

chrome.contextMenus.onClicked.addListener(async (info, tab) => {
    if (info.menuItemId === "save-to-snipo" && info.selectionText) {
        try {
            await chrome.scripting.insertCSS({
                target: { tabId: tab.id },
                files: ["styles.css"]
            });
            await chrome.scripting.executeScript({
                target: { tabId: tab.id },
                files: ["content.js"]
            });
            chrome.tabs.sendMessage(tab.id, {
                action: "contextMenuSave",
                code: info.selectionText,
                title: tab.title
            }).catch(err => {
                console.warn("Could not send context menu action to content script.", err);
                saveSnippet(info.selectionText, tab, "plaintext", tab.title);
            });
        } catch (err) {
            console.warn("Could not inject script.", err);
            saveSnippet(info.selectionText, tab, "plaintext", tab.title);
        }
    }
});

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
    if (request.action === "saveSnippet") {
        saveSnippet(request.code, sender.tab, request.language, request.title, request.folder_id, request.tags, request.description, request.filename, sendResponse);
        return true;
    }
    if (request.action === "fetchTags") {
        withClient(client => client.getTags())
            .then(data => sendResponse({ success: true, data }))
            .catch(err => sendResponse({ success: false, error: err.message }));
        return true;
    }
    if (request.action === "fetchFolders") {
        withClient(client => client.getFolders())
            .then(data => sendResponse({ success: true, data }))
            .catch(err => sendResponse({ success: false, error: err.message }));
        return true;
    }
    if (request.action === "createTag") {
        withClient(client => client.client.post("/api/v1/tags", { name: request.name }))
            .then(response => sendResponse({ success: true, data: response.data.data || response.data }))
            .catch(err => sendResponse({ success: false, error: err.message }));
        return true;
    }
    if (request.action === "createFolder") {
        withClient(client => client.client.post("/api/v1/folders", { name: request.name }))
            .then(response => sendResponse({ success: true, data: response.data.data || response.data }))
            .catch(err => sendResponse({ success: false, error: err.message }));
        return true;
    }
});

async function getClient() {
    const config = await chrome.storage.sync.get(['instanceUrl', 'apiKey']);
    if (!config.instanceUrl || !config.apiKey) {
        throw new Error("Configuration missing");
    }
    return new SnipoClient({
        baseURL: config.instanceUrl,
        token: config.apiKey
    });
}

async function withClient(action) {
    try {
        const client = await getClient();
        return await action(client);
    } catch (e) {
        let msg = e.message;
        if (e.response && e.response.data && e.response.data.error) {
            msg = e.response.data.error.message || msg;
        }
        throw new Error(msg);
    }
}

function getDefaultFilename(language) {
    const extMap = {
        'plaintext': 'txt', 'javascript': 'js', 'typescript': 'ts', 'python': 'py',
        'go': 'go', 'rust': 'rs', 'java': 'java', 'c': 'c', 'cpp': 'cpp',
        'csharp': 'cs', 'php': 'php', 'ruby': 'rb', 'swift': 'swift',
        'kotlin': 'kt', 'scala': 'scala', 'html': 'html', 'css': 'css',
        'scss': 'scss', 'json': 'json', 'yaml': 'yaml', 'xml': 'xml',
        'markdown': 'md', 'sql': 'sql', 'bash': 'sh', 'shell': 'sh',
        'powershell': 'ps1', 'dockerfile': 'dockerfile'
    };
    const ext = extMap[language] || 'txt';
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
    return `snippet_${timestamp}.${ext}`;
}

async function saveSnippet(code, tab, language = "plaintext", title = null, folder_id = null, tags = [], description = null, filename = null, sendResponse = null) {
    try {
        const client = await getClient();

        const finalTitle = title || `Snippet from ${new URL(tab.url).hostname}`;
        const finalDesc = description || `Saved from ${tab.url}`;
        const finalFilename = filename || getDefaultFilename(language);

        const payload = {
            title: finalTitle,
            description: finalDesc,
            is_public: false,
            tags: tags,
            files: [
                {
                    filename: finalFilename,
                    content: code,
                    language: language
                }
            ]
        };

        if (folder_id) {
            payload.folder_id = parseInt(folder_id);
        }

        const data = await client.createSnippet(payload);
        if (sendResponse) sendResponse({ success: true, data });
        notifyTab(tab.id, { success: true });
    } catch (error) {
        let msg = error.message;
        if (error.response && error.response.data && error.response.data.error) {
            msg = error.response.data.error.message || msg;
        }
        console.error("Snipo Error:", error);
        if (sendResponse) sendResponse({ success: false, error: msg });
        notifyTab(tab.id, { success: false, error: msg });
    }
}

function notifyTab(tabId, message) {
    chrome.tabs.sendMessage(tabId, { action: "showToast", ...message }).catch(err => {
        console.warn("Could not notify tab:", err);
    });
}
