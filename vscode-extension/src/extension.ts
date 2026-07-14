import * as vscode from 'vscode';
import { searchAndInsertSnippet, saveSelectedSnippet, replaceWithTemplate, openInSnipo, openInVSCode, updateSnippet, deleteSnippetCommand } from './commands';
import { SnippetTreeProvider } from './treeView';
import { SnipoStatusBar } from './statusBar';
import { api, Snippet } from './api';

export async function activate(context: vscode.ExtensionContext) {
    console.log('Snipo extension is now active');

    await api.init(context);

    if (!(await api.isConfigured())) {
        const result = await vscode.window.showInformationMessage('Snipo requires configuration to connect to your server.', 'Configure Now');
        if (result === 'Configure Now') {
            await runOnboardingFlow(context);
        }
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
            await runOnboardingFlow(context);
            // Refresh trees after config
            vscode.commands.executeCommand('snipo.refreshTree');
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
                editor.insertSnippet(new vscode.SnippetString(snippet.content));
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
}

export async function runOnboardingFlow(context: vscode.ExtensionContext) {
    let isValid = false;
    while (!isValid) {
        const apiUrl = await vscode.window.showInputBox({
            prompt: 'Enter your Snipo Server URL',
            placeHolder: 'e.g. http://localhost:3000',
            ignoreFocusOut: true
        });

        if (apiUrl === undefined) return; // User cancelled

        const apiToken = await vscode.window.showInputBox({
            prompt: 'Enter your Snipo API Token',
            password: true,
            ignoreFocusOut: true
        });

        if (apiToken === undefined) return; // User cancelled

        isValid = await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "Verifying Snipo connection...",
            cancellable: false
        }, async () => {
            return await api.verifyConfiguration(apiUrl, apiToken);
        });

        if (isValid) {
            const config = vscode.workspace.getConfiguration('snipo');
            await config.update('apiUrl', apiUrl, vscode.ConfigurationTarget.Global);
            await context.secrets.store('snipo.apiToken', apiToken);
            vscode.window.showInformationMessage('Successfully connected to Snipo!');
        } else {
            const retry = await vscode.window.showErrorMessage('Connection failed or invalid token. Please check your URL and Token.', 'Retry', 'Cancel');
            if (retry !== 'Retry') {
                break;
            }
        }
    }
}

export function deactivate() { }
