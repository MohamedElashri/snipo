"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.searchAndInsertSnippet = searchAndInsertSnippet;
exports.saveSelectedSnippet = saveSelectedSnippet;
exports.replaceWithTemplate = replaceWithTemplate;
exports.openInSnipo = openInSnipo;
exports.updateSnippet = updateSnippet;
exports.deleteSnippetCommand = deleteSnippetCommand;
exports.openInVSCode = openInVSCode;
const vscode = __importStar(require("vscode"));
const api_1 = require("./api");
const ALLOWED_LANGUAGES = [
    "plaintext", "javascript", "typescript", "python", "go", "rust", "java", "c", "cpp", "csharp",
    "php", "ruby", "swift", "kotlin", "scala", "html", "css", "scss", "json", "yaml", "xml",
    "markdown", "sql", "bash", "shell", "powershell", "dockerfile", "nginx", "toml", "ini",
    "makefile", "lua", "perl", "r", "haskell", "elixir", "clojure", "graphql", "protobuf", "terraform"
].sort();
function detectLanguage(vscodeLang) {
    const lang = vscodeLang.toLowerCase();
    if (ALLOWED_LANGUAGES.includes(lang)) {
        return lang;
    }
    const mappings = {
        'jsonc': 'json',
        'typescriptreact': 'typescript',
        'javascriptreact': 'javascript',
        'vue': 'html',
        'svelte': 'html',
        'less': 'css',
        'stylus': 'css',
        'yml': 'yaml',
        'sh': 'shell',
        'zsh': 'shell',
        'bat': 'powershell',
        'cmd': 'powershell',
        'jsx': 'javascript',
        'tsx': 'typescript',
        'tf': 'terraform'
    };
    return mappings[lang] || 'plaintext';
}
let searchTimeout;
async function showDynamicSnippetPicker(placeHolder) {
    return new Promise((resolve) => {
        const quickPick = vscode.window.createQuickPick();
        quickPick.placeholder = placeHolder;
        quickPick.matchOnDescription = true;
        quickPick.matchOnDetail = true;
        // Initial load
        quickPick.busy = true;
        api_1.api.getSnippets().then(snippets => {
            quickPick.items = snippets.map(s => ({
                label: `$(code) ${s.title}`,
                description: s.language,
                detail: s.description,
                snippet: s
            }));
            quickPick.busy = false;
        });
        quickPick.onDidChangeValue(value => {
            if (searchTimeout) {
                clearTimeout(searchTimeout);
            }
            quickPick.busy = true;
            searchTimeout = setTimeout(async () => {
                const snippets = value ? await api_1.api.searchSnippets(value) : await api_1.api.getSnippets();
                quickPick.items = snippets.map(s => ({
                    label: `$(code) ${s.title}`,
                    description: s.language,
                    detail: s.description,
                    snippet: s
                }));
                quickPick.busy = false;
            }, 300); // 300ms debounce
        });
        quickPick.onDidAccept(() => {
            const selected = quickPick.selectedItems[0];
            resolve(selected?.snippet);
            quickPick.hide();
        });
        quickPick.onDidHide(() => {
            if (searchTimeout)
                clearTimeout(searchTimeout);
            resolve(undefined);
            quickPick.dispose();
        });
        quickPick.show();
    });
}
async function searchAndInsertSnippet() {
    if (!(await api_1.api.isConfigured())) {
        vscode.window.showErrorMessage('Snipo is not configured. Please set apiUrl and apiToken in settings.');
        return;
    }
    const snippet = await showDynamicSnippetPicker('Search and select a snippet to insert');
    if (snippet) {
        const editor = vscode.window.activeTextEditor;
        if (editor) {
            editor.insertSnippet(new vscode.SnippetString(snippet.content));
        }
    }
}
async function saveSelectedSnippet() {
    if (!(await api_1.api.isConfigured())) {
        vscode.window.showErrorMessage('Snipo is not configured. Please set apiUrl and apiToken in settings.');
        return;
    }
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showWarningMessage('No active editor to save snippet from.');
        return;
    }
    const selection = editor.selection;
    let text = editor.document.getText(selection);
    if (!text) {
        // Fallback to entire document if no text is selected
        text = editor.document.getText();
        if (!text) {
            vscode.window.showWarningMessage('The current file is empty. Please select or write some code to save.');
            return;
        }
    }
    const title = await vscode.window.showInputBox({ prompt: 'Enter snippet title' });
    if (!title)
        return; // User cancelled
    const description = await vscode.window.showInputBox({ prompt: 'Enter snippet description (optional)' }) || '';
    const detected = detectLanguage(editor.document.languageId);
    const languageItems = ALLOWED_LANGUAGES.map(l => ({
        label: l,
        description: l === detected ? '(detected)' : undefined
    }));
    // Sort so detected is at the top
    languageItems.sort((a, b) => {
        if (a.label === detected)
            return -1;
        if (b.label === detected)
            return 1;
        return 0;
    });
    const languageSelection = await vscode.window.showQuickPick(languageItems, {
        placeHolder: 'Select snippet language',
        matchOnDescription: true
    });
    // If user cancels the quick pick, use detected language (which defaults to plaintext)
    const language = languageSelection ? languageSelection.label : detected;
    await api_1.api.createSnippet({
        title,
        description,
        content: text,
        language,
        is_public: false
    });
}
async function replaceWithTemplate() {
    if (!(await api_1.api.isConfigured())) {
        vscode.window.showErrorMessage('Snipo is not configured. Please set apiUrl and apiToken in settings.');
        return;
    }
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showWarningMessage('No active editor to replace text in.');
        return;
    }
    const snippet = await showDynamicSnippetPicker('Select a template to replace selection');
    if (snippet) {
        const selection = editor.selection;
        editor.insertSnippet(new vscode.SnippetString(snippet.content), selection);
    }
}
async function openInSnipo(arg) {
    if (!(await api_1.api.isConfigured())) {
        vscode.window.showErrorMessage('Snipo is not configured. Please set apiUrl and apiToken in settings.');
        return;
    }
    let snippetId;
    // Check if called from Tree View context menu
    if (arg && arg.snippet && typeof arg.snippet.id === 'string') {
        snippetId = arg.snippet.id;
    }
    else if (typeof arg === 'string') {
        snippetId = arg;
    }
    if (!snippetId) {
        const snippet = await showDynamicSnippetPicker('Select a snippet to open in Snipo');
        if (snippet) {
            snippetId = snippet.id;
        }
        else {
            return;
        }
    }
    if (snippetId) {
        const config = vscode.workspace.getConfiguration('snipo');
        const apiUrlString = config.get('apiUrl') || 'http://localhost:3000';
        try {
            const url = new URL(apiUrlString);
            url.searchParams.set('snippet', snippetId);
            vscode.env.openExternal(vscode.Uri.parse(url.toString()));
        }
        catch (e) {
            vscode.window.showErrorMessage('Invalid Snipo URL in settings.');
        }
    }
}
async function updateSnippet() {
    if (!(await api_1.api.isConfigured())) {
        vscode.window.showErrorMessage('Snipo is not configured. Please set apiUrl and apiToken in settings.');
        return;
    }
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showWarningMessage('No active editor.');
        return;
    }
    const selection = editor.selection;
    let text = editor.document.getText(selection);
    if (!text) {
        // Fallback to entire document if no text is selected
        text = editor.document.getText();
        if (!text) {
            vscode.window.showWarningMessage('The current file is empty. Nothing to update.');
            return;
        }
    }
    const snippet = await showDynamicSnippetPicker('Select a snippet to update with selected text');
    if (snippet) {
        const detected = detectLanguage(editor.document.languageId);
        const languageItems = ALLOWED_LANGUAGES.map(l => ({
            label: l,
            description: l === detected ? '(detected)' : undefined
        }));
        // Sort so detected is at the top
        languageItems.sort((a, b) => {
            if (a.label === detected)
                return -1;
            if (b.label === detected)
                return 1;
            return 0;
        });
        const languageSelection = await vscode.window.showQuickPick(languageItems, {
            placeHolder: 'Select snippet language',
            matchOnDescription: true
        });
        // If user cancels the quick pick, use detected language (which defaults to plaintext)
        const finalLanguage = languageSelection ? languageSelection.label : detected;
        vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "Updating Snipo snippet...",
            cancellable: false
        }, async () => {
            await api_1.api.updateSnippet(snippet.id, {
                content: text,
                language: finalLanguage
            });
        });
    }
}
async function deleteSnippetCommand(arg) {
    if (!(await api_1.api.isConfigured())) {
        vscode.window.showErrorMessage('Snipo is not configured. Please set apiUrl and apiToken in settings.');
        return;
    }
    let snippet;
    // Check if called from Tree View context menu
    if (arg && arg.snippet) {
        snippet = arg.snippet;
    }
    else if (arg && typeof arg.title === 'string') {
        snippet = arg;
    }
    if (!snippet) {
        snippet = await showDynamicSnippetPicker('Select a snippet to delete');
    }
    if (!snippet)
        return;
    const confirm = await vscode.window.showWarningMessage(`Are you sure you want to delete "${snippet.title}"?`, { modal: true }, 'Delete');
    if (confirm === 'Delete') {
        const success = await api_1.api.deleteSnippet(snippet.id);
        if (success) {
            vscode.commands.executeCommand('snipo.refreshTree');
        }
    }
}
async function openInVSCode(arg) {
    if (!(await api_1.api.isConfigured())) {
        vscode.window.showErrorMessage('Snipo is not configured. Please set apiUrl and apiToken in settings.');
        return;
    }
    let snippet;
    if (arg && arg.snippet) {
        snippet = arg.snippet;
    }
    else if (arg && typeof arg.title === 'string') {
        snippet = arg;
    }
    if (!snippet) {
        snippet = await showDynamicSnippetPicker('Select a snippet to open');
    }
    if (snippet) {
        const doc = await vscode.workspace.openTextDocument({
            language: snippet.language || 'text',
            content: snippet.content
        });
        await vscode.window.showTextDocument(doc);
    }
}
//# sourceMappingURL=commands.js.map