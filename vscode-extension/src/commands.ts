import * as vscode from 'vscode';
import { api, Snippet } from './api';

const ALLOWED_LANGUAGES = [
    "plaintext", "javascript", "typescript", "python", "go", "rust", "java", "c", "cpp", "csharp", 
    "php", "ruby", "swift", "kotlin", "scala", "html", "css", "scss", "json", "yaml", "xml", 
    "markdown", "sql", "bash", "shell", "powershell", "dockerfile", "nginx", "toml", "ini", 
    "makefile", "lua", "perl", "r", "haskell", "elixir", "clojure", "graphql", "protobuf", "terraform"
].sort();

function detectLanguage(vscodeLang: string): string {
    const lang = vscodeLang.toLowerCase();
    
    if (ALLOWED_LANGUAGES.includes(lang)) {
        return lang;
    }

    const mappings: Record<string, string> = {
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

let searchTimeout: NodeJS.Timeout | undefined;
let searchAbortController: AbortController | undefined;

async function showDynamicSnippetPicker(placeHolder: string): Promise<Snippet | undefined> {
    return new Promise((resolve) => {
        const quickPick = vscode.window.createQuickPick<vscode.QuickPickItem & { snippet?: Snippet }>();
        quickPick.placeholder = placeHolder;
        quickPick.matchOnDescription = true;
        quickPick.matchOnDetail = true;

        // Initial load
        quickPick.busy = true;
        api.getSnippets().then(snippets => {
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
            if (searchAbortController) {
                searchAbortController.abort();
            }
            
            quickPick.busy = true;
            searchTimeout = setTimeout(async () => {
                searchAbortController = new AbortController();
                try {
                    const snippets = value ? await api.searchSnippets(value, searchAbortController.signal) : await api.getSnippets();
                    quickPick.items = snippets.map(s => ({
                        label: `$(code) ${s.title}`,
                        description: s.language,
                        detail: s.description,
                        snippet: s
                    }));
                } catch (e) {
                    // Ignore errors, handled in API
                } finally {
                    quickPick.busy = false;
                }
            }, 300); // 300ms debounce
        });

        quickPick.onDidAccept(() => {
            const selected = quickPick.selectedItems[0];
            resolve(selected?.snippet);
            quickPick.hide();
        });

        quickPick.onDidHide(() => {
            if (searchTimeout) clearTimeout(searchTimeout);
            if (searchAbortController) searchAbortController.abort();
            resolve(undefined);
            quickPick.dispose();
        });

        quickPick.show();
    });
}

export async function searchAndInsertSnippet() {
    if (!(await api.isConfigured())) {
        vscode.window.showErrorMessage('Snipo is not configured. Please set apiUrl and apiToken in settings.');
        return;
    }

    const snippet = await showDynamicSnippetPicker('Search and select a snippet to insert');
    if (snippet) {
        const editor = vscode.window.activeTextEditor;
        if (editor) {
            editor.insertSnippet(new vscode.SnippetString().appendText(snippet.content));
        }
    }
}

export async function saveSelectedSnippet() {
    if (!(await api.isConfigured())) {
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

    const titleInput = await vscode.window.showInputBox({ prompt: 'Enter snippet title' });
    const title = titleInput?.trim();
    if (!title) return; // User cancelled or empty

    const descInput = await vscode.window.showInputBox({ prompt: 'Enter snippet description (optional)' });
    const description = descInput?.trim() || '';
    
    const detected = detectLanguage(editor.document.languageId);
    
    const languageItems: vscode.QuickPickItem[] = ALLOWED_LANGUAGES.map(l => ({
        label: l,
        description: l === detected ? '(detected)' : undefined
    }));
    
    // Sort so detected is at the top
    languageItems.sort((a, b) => {
        if (a.label === detected) return -1;
        if (b.label === detected) return 1;
        return 0;
    });

    const languageSelection = await vscode.window.showQuickPick(languageItems, {
        placeHolder: 'Select snippet language',
        matchOnDescription: true
    });

    // If user cancels the quick pick, use detected language (which defaults to plaintext)
    const language = languageSelection ? languageSelection.label : detected;

    await api.createSnippet({
        title,
        description,
        content: text,
        language,
        is_public: false
    });
}

export async function replaceWithTemplate() {
    if (!(await api.isConfigured())) {
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
        editor.insertSnippet(new vscode.SnippetString().appendText(snippet.content), selection);
    }
}

export async function openInSnipo(arg?: any) {
    if (!(await api.isConfigured())) {
        vscode.window.showErrorMessage('Snipo is not configured. Please set apiUrl and apiToken in settings.');
        return;
    }
    
    let snippetId: string | undefined;
    
    // Check if called from Tree View context menu
    if (arg && arg.snippet && typeof arg.snippet.id === 'string') {
        snippetId = arg.snippet.id;
    } else if (typeof arg === 'string') {
        snippetId = arg;
    }

    if (!snippetId) {
        const snippet = await showDynamicSnippetPicker('Select a snippet to open in Snipo');
        if (snippet) {
            snippetId = snippet.id;
        } else {
            return;
        }
    }

    if (snippetId) {
        const config = vscode.workspace.getConfiguration('snipo');
        const apiUrlString = config.get<string>('apiUrl') || 'http://localhost:3000';
        try {
            const url = new URL(apiUrlString);
            if (url.protocol !== 'http:' && url.protocol !== 'https:') {
                vscode.window.showErrorMessage('Invalid Snipo URL protocol in settings. Must be http or https.');
                return;
            }
            url.searchParams.set('snippet', snippetId);
            vscode.env.openExternal(vscode.Uri.parse(url.toString()));
        } catch (e) {
            vscode.window.showErrorMessage('Invalid Snipo URL in settings.');
        }
    }
}

export async function updateSnippet() {
    if (!(await api.isConfigured())) {
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
        const languageItems: vscode.QuickPickItem[] = ALLOWED_LANGUAGES.map(l => ({
            label: l,
            description: l === detected ? '(detected)' : undefined
        }));
        
        // Sort so detected is at the top
        languageItems.sort((a, b) => {
            if (a.label === detected) return -1;
            if (b.label === detected) return 1;
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
            await api.updateSnippet(snippet.id, {
                content: text,
                language: finalLanguage
            });
        });
    }
}

export async function deleteSnippetCommand(arg?: any) {
    if (!(await api.isConfigured())) {
        vscode.window.showErrorMessage('Snipo is not configured. Please set apiUrl and apiToken in settings.');
        return;
    }
    
    let snippet: Snippet | undefined;
    
    // Check if called from Tree View context menu
    if (arg && arg.snippet) {
        snippet = arg.snippet;
    } else if (arg && typeof arg.title === 'string') {
        snippet = arg;
    }

    if (!snippet) {
        snippet = await showDynamicSnippetPicker('Select a snippet to delete');
    }

    if (!snippet) return;

    const confirm = await vscode.window.showWarningMessage(
        `Are you sure you want to delete "${snippet.title}"?`,
        { modal: true },
        'Delete'
    );

    if (confirm === 'Delete') {
        const success = await api.deleteSnippet(snippet.id);
        if (success) {
            vscode.commands.executeCommand('snipo.refreshTree');
        }
    }
}

export async function openInVSCode(arg?: any) {
    if (!(await api.isConfigured())) {
        vscode.window.showErrorMessage('Snipo is not configured. Please set apiUrl and apiToken in settings.');
        return;
    }
    
    let snippet: Snippet | undefined;
    
    if (arg && arg.snippet) {
        snippet = arg.snippet;
    } else if (arg && typeof arg.title === 'string') {
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
