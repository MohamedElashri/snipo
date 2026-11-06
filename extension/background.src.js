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
            // Content scripts are already present on supported pages. Reinjecting
            // the script redeclares its globals and breaks interactive saves.
            await chrome.tabs.sendMessage(tab.id, {
                action: "contextMenuSave", code: info.selectionText, title: tab.title
            });
        } catch (error) {
            try {
                // Firefox already grants scripting access. Use it only when the
                // page has no receiver; never reinject an existing content script.
                if (!chrome.scripting) throw error;
                await chrome.scripting.insertCSS({ target: { tabId: tab.id }, files: ["styles.css"] });
                await chrome.scripting.executeScript({ target: { tabId: tab.id }, files: ["content.js"] });
                await chrome.tabs.sendMessage(tab.id, {
                    action: "contextMenuSave", code: info.selectionText, title: tab.title
                });
            } catch (injectionError) {
                console.warn("Could not open the save dialog.", injectionError);
                const { interactiveMode } = await chrome.storage.sync.get(['interactiveMode']);
                if (interactiveMode) {
                    await openSaveDialog({ code: info.selectionText, language: 'plaintext', title: Array.from(tab.title || '').slice(0, 200).join('') }, tab);
                } else {
                    await saveSnippet(info.selectionText, tab, "plaintext", tab.title);
                }
            }
        }
    }
});

// Pending drafts stay in trusted, browser-session storage, never website DOM or sync.
const draftPrefix = 'snipoSaveDraft:';
const savePage = () => chrome.runtime.getURL('save/save.html');
const extensionPage = sender => sender.id === chrome.runtime.id && sender.url?.startsWith(chrome.runtime.getURL(''));
let openingDialog = Promise.resolve();

function openSaveDialog(data, tab) {
    // Serialize opens so repeated user clicks cannot create duplicate windows.
    const opened = openingDialog.then(async () => {
        const records = await chrome.storage.session.get(null);
        for (const [key, record] of Object.entries(records)) {
            if (!key.startsWith(draftPrefix) || record.tab.id !== tab.id) continue;
            try {
                if (record.dialogTabId !== undefined) {
                    await chrome.tabs.update(record.dialogTabId, { active: true });
                } else {
                    await chrome.windows.update(record.windowId, { focused: true });
                }
                return;
            } catch (_) { await chrome.storage.session.remove(key); }
        }
        const id = crypto.randomUUID();
        const key = draftPrefix + id;
        const url = savePage() + '?draft=' + id;
        const record = { url, tab, data: { ...data, sourceURL: tab.url || 'unknown' } };
        await chrome.storage.session.set({ [key]: record });
        try {
            try {
                if (!chrome.windows?.create) throw new Error('Popup windows unavailable');
                const popup = await chrome.windows.create({ url, type: 'popup', width: 560, height: 720 });
                record.windowId = popup.id;
            } catch (_) {
                // Mobile browsers may only support tabs. Keep the same origin boundary.
                const dialogTab = await chrome.tabs.create({ url });
                record.dialogTabId = dialogTab.id;
            }
            await chrome.storage.session.set({ [key]: record });
        } catch (error) {
            await chrome.storage.session.remove(key);
            throw error;
        }
    });
    openingDialog = opened.catch(() => {});
    return opened;
}

async function getSaveDraft(request, sender) {
    if (!extensionPage(sender) || typeof request.draftId !== 'string') throw new Error('Save window required');
    const key = draftPrefix + request.draftId;
    const { [key]: record } = await chrome.storage.session.get(key);
    if (!record || sender.url !== record.url) throw new Error('Save window expired');
    return record;
}

async function removeSaveDrafts(matches) {
    // A close event can arrive before create() returns the new window/tab ID.
    // Wait for that ID to be persisted so cleanup cannot leave an orphan draft.
    await openingDialog;
    const records = await chrome.storage.session.get(null);
    for (const [key, record] of Object.entries(records)) {
        if (key.startsWith(draftPrefix) && matches(record)) {
            await chrome.storage.session.remove(key);
            chrome.tabs.sendMessage(record.tab.id, { action: 'saveDialogClosed' }).catch(() => {});
        }
    }
}
chrome.windows?.onRemoved.addListener(windowId => removeSaveDrafts(record => record.windowId === windowId));
chrome.tabs.onRemoved?.addListener(tabId => removeSaveDrafts(record => record.dialogTabId === tabId));

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
    if (request.action === 'openSaveDialog') {
        if (sender.id !== chrome.runtime.id || !sender.tab) return;
        openSaveDialog({ code: request.code, language: request.language, title: request.title }, sender.tab)
            .then(() => sendResponse({ success: true }))
            .catch(() => sendResponse({ success: false, error: 'Unable to open save window' }));
        return true;
    }
    if (request.action === 'getSaveDraft' || request.action === 'saveDraftSnippet') {
        getSaveDraft(request, sender).then(async record => {
            if (request.action === 'getSaveDraft') {
                sendResponse({ success: true, data: record.data });
            } else {
                await saveSnippet(record.data.code, record.tab, request.language, request.title,
                    request.folder_id, request.tags, request.description, request.filename, sendResponse);
            }
        }).catch(() => sendResponse({ success: false, error: 'Save window expired' }));
        return true;
    }
    if (['fetchTags', 'fetchFolders', 'createTag', 'createFolder'].includes(request.action) && !extensionPage(sender)) {
        sendResponse({ success: false, error: 'Extension page required' });
        return;
    }
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
    const instanceUrl = config.instanceUrl?.trim();
    const apiKey = config.apiKey?.trim();
    if (!instanceUrl || !apiKey) {
        throw new Error("Configuration missing");
    }
    return new SnipoClient({
        baseURL: instanceUrl,
        token: apiKey
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
        'cuda': 'cu', 'tex': 'tex', 'bibtex': 'bib',
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

        const urlStr = tab && tab.url ? tab.url : "unknown";
        let hostStr = "unknown";
        try { if (tab && tab.url) hostStr = new URL(tab.url).hostname; } catch(e) {}

        const finalTitle = title || `Snippet from ${hostStr}`;
        const finalDesc = description ?? `Saved from ${urlStr}`;
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
        if (tab && tab.id) notifyTab(tab.id, { success: true });
    } catch (error) {
        let msg = error.message;
        if (error.response && error.response.data && error.response.data.error) {
            msg = error.response.data.error.message || msg;
        }
        console.error("Snipo Error:", error);
        if (sendResponse) sendResponse({ success: false, error: msg });
        if (tab && tab.id) notifyTab(tab.id, { success: false, error: msg });
    }
}

function notifyTab(tabId, message) {
    chrome.tabs.sendMessage(tabId, { action: "showToast", ...message }).catch(err => {
        console.warn("Could not notify tab:", err);
    });
}
