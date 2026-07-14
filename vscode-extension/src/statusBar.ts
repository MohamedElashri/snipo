import * as vscode from 'vscode';
import { api } from './api';

export class SnipoStatusBar {
    private statusBarItem: vscode.StatusBarItem;

    constructor(context: vscode.ExtensionContext) {
        this.statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
        this.statusBarItem.command = 'snipo.searchAndInsert';
        this.update();

        // Listen for configuration changes to update the status bar
        vscode.workspace.onDidChangeConfiguration(e => {
            if (e.affectsConfiguration('snipo.apiUrl')) {
                this.update();
            }
        });
        context.secrets.onDidChange(e => {
            if (e.key === 'snipo.apiToken') {
                this.update();
            }
        });
    }

    public async update() {
        if (await api.isConfigured()) {
            this.statusBarItem.text = '$(repo) Snipo';
            this.statusBarItem.tooltip = 'Snipo: Search and Insert Snippet';
            this.statusBarItem.backgroundColor = undefined;
        } else {
            this.statusBarItem.text = '$(warning) Snipo';
            this.statusBarItem.tooltip = 'Snipo is not configured. Click to search (will prompt setup).';
            this.statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.warningBackground');
        }
        this.statusBarItem.show();
    }

    public dispose() {
        this.statusBarItem.dispose();
    }
}
