import * as vscode from 'vscode';
import { searchAndInsertSnippet, saveSelectedSnippet, replaceWithTemplate, openInSnipo, openInVSCode, updateSnippet, deleteSnippetCommand } from './commands';
import { SnippetTreeProvider } from './treeView';
import { SnipoStatusBar } from './statusBar';
import { api, Snippet } from './api';
import { SettingsViewProvider } from './settingsView';

export async function activate(context: vscode.ExtensionContext) {
    console.log('Snipo extension is now active');

    await api.init(context);

    if (!(await api.isConfigured())) {
        vscode.window.showInformationMessage('Snipo requires configuration to connect to your server. Please configure it in the Snipo sidebar.');
        vscode.commands.executeCommand('snipo-settings.focus');
    }

    // Register commands
    context.subscriptions.push(
        vscode.commands.registerCommand('snipo.searchAndInsert', searchAndInsertSnippet),
        vscode.commands.registerCommand('snipo.saveSelected', saveSelectedSnippet),
        vscode.commands.registerCommand('snipo.replaceWithTemplate', replaceWithTemplate),
        vscode.commands.registerCommand('snipo.openInSnipo', openInSnipo),
        vscode.commands.registerCommand('snipo.openInVSCode', openInVSCode),
        vscode.commands.registerCommand('snipo.updateSnippet', updateSnippet),
        vscode.commands.registerCommand('snipo.deleteSnippet', deleteSnippetCommand),
        vscode.commands.registerCommand('snipo.configure', async () => {
            vscode.commands.executeCommand('snipo-settings.focus');
        }),
        vscode.commands.registerCommand('snipo.openSettings', () => {
            vscode.commands.executeCommand('workbench.action.openSettings', '@ext:mohamedelashri.snipo');
        })
    );

    // Initialize Status Bar
    const statusBar = new SnipoStatusBar(context);
    context.subscriptions.push(statusBar);

    // Register hidden command for tree view insertion
    context.subscriptions.push(
        vscode.commands.registerCommand('snipo.insertSnippetFromTree', (arg: any) => {
            let snippet = (arg && arg.snippet) ? arg.snippet : arg;
            const editor = vscode.window.activeTextEditor;
            if (editor && snippet) {
                editor.insertSnippet(new vscode.SnippetString().appendText(snippet.content));
            }
        })
    );

    // Register Tree Views
    const favoritesProvider = new SnippetTreeProvider('favorites');
    const recentProvider = new SnippetTreeProvider('recent');

    vscode.window.registerTreeDataProvider('snipo-favorites', favoritesProvider);
    vscode.window.registerTreeDataProvider('snipo-recent', recentProvider);

    // Refresh command
    context.subscriptions.push(
        vscode.commands.registerCommand('snipo.refreshTree', () => {
            favoritesProvider.refresh();
            recentProvider.refresh();
        })
    );

    // Register Settings View
    const settingsProvider = new SettingsViewProvider(context.extensionUri, context);
    context.subscriptions.push(
        vscode.window.registerWebviewViewProvider(SettingsViewProvider.viewType, settingsProvider)
    );
}

export function deactivate() { }
