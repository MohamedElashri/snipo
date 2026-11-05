import * as vscode from 'vscode';
import { api } from './api';

export class SnipoStatusBar {
    private statusBarItem: vscode.StatusBarItem;

    private updateRequest = 0;
    private disposed = false;
    private listeners: vscode.Disposable[] = [];

    constructor(context: vscode.ExtensionContext) {
        this.statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
        this.statusBarItem.command = 'snipo.searchAndInsert';
        this.update();

        // Listen for configuration changes to update the status bar
        this.listeners.push(vscode.workspace.onDidChangeConfiguration(e => {
            if (e.affectsConfiguration('snipo.apiUrl')) {
                this.update();
            }
        }));
        this.listeners.push(context.secrets.onDidChange(e => {
            if (e.key === 'snipo.apiToken') {
                this.update();
            }
        }));
    }

    public async update() {
        if (this.disposed) return;
        const request = ++this.updateRequest;
        const configured = await api.isConfigured();
        if (this.disposed || request !== this.updateRequest) return;
        if (configured) {
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
        this.disposed = true;
        ++this.updateRequest;
        this.listeners.forEach(listener => listener?.dispose());
        this.statusBarItem.dispose();
    }
}
