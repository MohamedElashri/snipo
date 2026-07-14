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
exports.activate = activate;
exports.runOnboardingFlow = runOnboardingFlow;
exports.deactivate = deactivate;
const vscode = __importStar(require("vscode"));
const commands_1 = require("./commands");
const treeView_1 = require("./treeView");
const statusBar_1 = require("./statusBar");
const api_1 = require("./api");
async function activate(context) {
    console.log('Snipo extension is now active');
    await api_1.api.init(context);
    if (!(await api_1.api.isConfigured())) {
        const result = await vscode.window.showInformationMessage('Snipo requires configuration to connect to your server.', 'Configure Now');
        if (result === 'Configure Now') {
            await runOnboardingFlow(context);
        }
    }
    // Register commands
    context.subscriptions.push(vscode.commands.registerCommand('snipo.searchAndInsert', commands_1.searchAndInsertSnippet), vscode.commands.registerCommand('snipo.saveSelected', commands_1.saveSelectedSnippet), vscode.commands.registerCommand('snipo.replaceWithTemplate', commands_1.replaceWithTemplate), vscode.commands.registerCommand('snipo.openInSnipo', commands_1.openInSnipo), vscode.commands.registerCommand('snipo.openInVSCode', commands_1.openInVSCode), vscode.commands.registerCommand('snipo.updateSnippet', commands_1.updateSnippet), vscode.commands.registerCommand('snipo.deleteSnippet', commands_1.deleteSnippetCommand), vscode.commands.registerCommand('snipo.configure', async () => {
        await runOnboardingFlow(context);
        // Refresh trees after config
        vscode.commands.executeCommand('snipo.refreshTree');
    }), vscode.commands.registerCommand('snipo.openSettings', () => {
        vscode.commands.executeCommand('workbench.action.openSettings', '@ext:mohamedelashri.snipo');
    }));
    // Initialize Status Bar
    const statusBar = new statusBar_1.SnipoStatusBar(context);
    context.subscriptions.push(statusBar);
    // Register hidden command for tree view insertion
    context.subscriptions.push(vscode.commands.registerCommand('snipo.insertSnippetFromTree', (arg) => {
        let snippet = (arg && arg.snippet) ? arg.snippet : arg;
        const editor = vscode.window.activeTextEditor;
        if (editor && snippet) {
            editor.insertSnippet(new vscode.SnippetString(snippet.content));
        }
    }));
    // Register Tree Views
    const favoritesProvider = new treeView_1.SnippetTreeProvider('favorites');
    const recentProvider = new treeView_1.SnippetTreeProvider('recent');
    vscode.window.registerTreeDataProvider('snipo-favorites', favoritesProvider);
    vscode.window.registerTreeDataProvider('snipo-recent', recentProvider);
    // Refresh command
    context.subscriptions.push(vscode.commands.registerCommand('snipo.refreshTree', () => {
        favoritesProvider.refresh();
        recentProvider.refresh();
    }));
}
async function runOnboardingFlow(context) {
    let isValid = false;
    while (!isValid) {
        const apiUrl = await vscode.window.showInputBox({
            prompt: 'Enter your Snipo Server URL',
            placeHolder: 'e.g. http://localhost:3000',
            ignoreFocusOut: true
        });
        if (apiUrl === undefined)
            return; // User cancelled
        const apiToken = await vscode.window.showInputBox({
            prompt: 'Enter your Snipo API Token',
            password: true,
            ignoreFocusOut: true
        });
        if (apiToken === undefined)
            return; // User cancelled
        isValid = await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "Verifying Snipo connection...",
            cancellable: false
        }, async () => {
            return await api_1.api.verifyConfiguration(apiUrl, apiToken);
        });
        if (isValid) {
            const config = vscode.workspace.getConfiguration('snipo');
            await config.update('apiUrl', apiUrl, vscode.ConfigurationTarget.Global);
            await context.secrets.store('snipo.apiToken', apiToken);
            vscode.window.showInformationMessage('Successfully connected to Snipo!');
        }
        else {
            const retry = await vscode.window.showErrorMessage('Connection failed or invalid token. Please check your URL and Token.', 'Retry', 'Cancel');
            if (retry !== 'Retry') {
                break;
            }
        }
    }
}
function deactivate() { }
//# sourceMappingURL=extension.js.map