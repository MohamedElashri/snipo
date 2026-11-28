import * as vscode from 'vscode';
import { api, Snippet, configuredServerURL } from './api';

const ALLOWED_LANGUAGES = [
    "tex", "bibtex", "cuda", "plaintext", "javascript", "typescript", "python", "go", "rust", "java", "c", "cpp", "csharp",
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
        'latex': 'tex',
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

async function showDynamicSnippetPicker(placeHolder: string, kind?: 'note'): Promise<Snippet | undefined> {
    return new Promise((resolve) => {
        let searchTimeout: NodeJS.Timeout | undefined;
        let searchAbortController: AbortController | undefined;
        let generation = 0;
        let disposed = false;
        const quickPick = vscode.window.createQuickPick<vscode.QuickPickItem & { snippet?: Snippet }>();
        quickPick.placeholder = placeHolder;
        quickPick.matchOnDescription = true;
        quickPick.matchOnDetail = true;

        const loadItems = () => kind === 'note' ? api.getSnippets('', undefined, undefined, undefined, kind) : api.getSnippets();
        const icon = kind === 'note' ? 'note' : 'code';

        // Initial load
        quickPick.busy = true;
        loadItems().then(snippets => {
            if (disposed || generation !== 0) return;
            try {
                quickPick.items = snippets.map(s => ({
                    label: `$(${icon}) ${s.title}`,
                    description: s.language,
                    detail: s.description,
                    snippet: s
                }));
            } catch (e) {
                // Ignore if quickPick is disposed
            } finally {
                quickPick.busy = false;
            }
        }).catch(e => {
            console.error('Failed to load snippets', e);
            if (!disposed && generation === 0) quickPick.busy = false;
        });

        quickPick.onDidChangeValue(value => {
            const request = ++generation;
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
                    const snippets = value ? await api.searchSnippets(value, searchAbortController.signal, kind) : await loadItems();
                    if (disposed || request !== generation) return;
                    quickPick.items = snippets.map(s => ({
                        label: `$(${icon}) ${s.title}`,
                        description: s.language,
                        detail: s.description,
                        snippet: s
                    }));
                } catch (e) {
                    // Ignore errors, handled in API
                } finally {
                    if (!disposed && request === generation) quickPick.busy = false;
                }
            }, 300); // 300ms debounce
        });

        quickPick.onDidAccept(() => {
            const selected = quickPick.selectedItems[0];
            resolve(selected?.snippet);
            quickPick.hide();
        });

        quickPick.onDidHide(() => {
            disposed = true;
            generation++;
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
            editor.insertSnippet(new vscode.SnippetString().appendText(snippet.content || ''));
        }
    }
}

export async function saveSelectedSnippet(document?: vscode.TextDocument) {
    if (!(await api.isConfigured())) {
        vscode.window.showErrorMessage('Snipo is not configured. Please set apiUrl and apiToken in settings.');
        return;
    }
    const editor = vscode.window.activeTextEditor;
    const sourceDocument = document || editor?.document;
    if (!sourceDocument) {
        vscode.window.showWarningMessage('No active editor to save snippet from.');
        return;
    }

    // The file CodeLens supplies a document; palette actions use the selection.
    let text = document ? document.getText() : sourceDocument.getText(editor!.selection);

    if (!text) {
        // Fallback to entire document if no text is selected
        text = sourceDocument.getText();
        if (!text) {
            vscode.window.showWarningMessage('The current file is empty. Please select or write some code to save.');
            return;
        }
    }

    const titleInput = await vscode.window.showInputBox({ prompt: 'Enter snippet title' });
    const title = titleInput?.trim();
    if (!title) return; // User cancelled or empty

    const descInput = await vscode.window.showInputBox({ prompt: 'Enter snippet description (optional)' });
    if (descInput === undefined) return;
    const description = descInput.trim();
    
    const detected = detectLanguage(sourceDocument.languageId);
    
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

    if (!languageSelection) return;
    const language = languageSelection.label;

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
        const currentEditor = vscode.window.activeTextEditor;
        if (!currentEditor) {
            vscode.window.showWarningMessage('Editor was closed or changed. Cannot replace text.');
            return;
        }
        const selection = currentEditor.selection;
        currentEditor.insertSnippet(new vscode.SnippetString().appendText(snippet.content || ''), selection);
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
        const apiUrlString = configuredServerURL();
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

        if (!languageSelection) return;
        const finalLanguage = languageSelection.label;

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "Updating Snipo snippet...",
            cancellable: false
        }, async () => {
            // PUT replaces metadata, so use the latest complete snippet rather than
            // the picker snapshot and preserve fields this command does not edit.
            const current = await api.getSnippet(snippet.id);
            if (!current) {
                vscode.window.showErrorMessage('Could not load the snippet. Nothing was updated.');
                return;
            }
            await api.updateSnippet(snippet.id, {
                title: current.title,
                description: current.description,
                is_public: current.is_public,
                is_archived: current.is_archived,
                expires_at: current.expires_at,
                folder_id: current.folders?.[0]?.id ?? current.folder_id,
                content: text,
                language: finalLanguage,
                files: current.files?.map((file, index) => index === 0
                    ? { ...file, content: text, language: finalLanguage }
                    : file)
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
            await vscode.commands.executeCommand('snipo.refreshTree');
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
        const server = configuredServerURL();
        const current = await api.getSnippet(snippet.id);
        if (!current) {
            vscode.window.showErrorMessage('Could not load the snippet. Nothing was opened.');
            return;
        }
        snippet = current;
        const doc = await vscode.workspace.openTextDocument({
            language: snippet.language || 'plaintext',
            content: snippet.content || ''
        });
        if (snippet.kind === 'note') {
            noteDocuments.set(doc.uri.toString(), { note: snippet, server, saving: false });
        }
        await vscode.window.showTextDocument(doc);
    }
}


const noteDocuments = new Map<string, { note?: Snippet; server: string; saving: boolean }>();

export async function newNote() {
    if (!(await api.isConfigured())) {
        vscode.window.showErrorMessage('Configure Snipo to create a note.');
        return;
    }
    const server = configuredServerURL();
    const format = await vscode.window.showQuickPick([
        { label: 'Markdown', language: 'markdown' },
        { label: 'Plain text', language: 'plaintext' },
    ], { placeHolder: 'Choose a note format' });
    if (!format) return;
    const doc = await vscode.workspace.openTextDocument({ language: format.language, content: '' });
    noteDocuments.set(doc.uri.toString(), { server, saving: false });
    await vscode.window.showTextDocument(doc);
}

export async function previewOpenNote() {
    const doc = vscode.window.activeTextEditor?.document;
    if (!doc || !noteDocuments.has(doc.uri.toString()) || doc.languageId !== 'markdown') {
        vscode.window.showInformationMessage('Open a Markdown note to preview it.');
        return;
    }
    await vscode.commands.executeCommand('markdown.showPreviewToSide', doc.uri);
}

export async function browseNotes() {
    if (!(await api.isConfigured())) {
        vscode.window.showErrorMessage('Configure Snipo to browse notes.');
        return;
    }
    const note = await showDynamicSnippetPicker('Search Snipo notes', 'note');
    if (note) await openInVSCode(note);
}

export function forgetNoteDocument(document: vscode.TextDocument) {
    noteDocuments.delete(document.uri.toString());
}

export async function saveOpenNote() {
    const doc = vscode.window.activeTextEditor?.document;
    const entry = doc && noteDocuments.get(doc.uri.toString());
    if (!doc || !entry || entry.server !== configuredServerURL()) {
        vscode.window.showErrorMessage('Open a note from the current Snipo server to save it here.');
        return;
    }
    if (entry.saving) return;
    const { note } = entry;
    const content = doc.getText();
    const language = doc.languageId;
    if (language !== 'markdown' && language !== 'plaintext') {
        vscode.window.showErrorMessage('Choose Markdown or Plain Text for this note.');
        return;
    }
    entry.saving = true;
    try {
        if (!note) {
            const firstLine = content.split(/\r?\n/).map(line => language === 'markdown' ? line.trim().replace(/^#{1,6}\s+/, '') : line.trim()).find(Boolean) || '';
            const title = await vscode.window.showInputBox({
                prompt: 'Title for the new note',
                value: [...firstLine].slice(0, 200).join(''),
                validateInput: value => !value.trim() || [...value.trim()].length > 200 ? 'Enter a title up to 200 characters.' : undefined,
            });
            if (!title?.trim() || entry.server !== configuredServerURL()) return;
            const created = await api.createSnippet({ kind: 'note', title: title.trim(), content, language, is_public: false });
            if (created) entry.note = created;
            return;
        }
        const saved = await api.updateSnippet(note.id, {
            kind: 'note', revision: note.revision, title: note.title, description: note.description,
            content, language, is_public: note.is_public, is_archived: note.is_archived, expires_at: note.expires_at,
            folder_id: note.folders?.[0]?.id, tags: note.tags?.map(tag => tag.name),
        });
        if (saved) entry.note = saved;
    } finally {
        entry.saving = false;
    }
}
