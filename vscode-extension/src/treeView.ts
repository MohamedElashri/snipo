import * as vscode from 'vscode';
import { api, Snippet } from './api';

export class SnippetTreeItem extends vscode.TreeItem {
    constructor(
        public readonly snippet: Snippet,
        public readonly command?: vscode.Command
    ) {
        super(snippet.title, vscode.TreeItemCollapsibleState.None);

        this.description = snippet.language;
        this.tooltip = snippet.description || snippet.title;
        this.contextValue = 'snippet';
        this.iconPath = vscode.ThemeIcon.File;

        // Define command when clicked
        this.command = {
            command: 'snipo.insertSnippetFromTree',
            title: 'Insert Snippet',
            arguments: [this.snippet]
        };
    }
}

export class SnippetTreeProvider implements vscode.TreeDataProvider<SnippetTreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<SnippetTreeItem | undefined | void> = new vscode.EventEmitter<SnippetTreeItem | undefined | void>();
    readonly onDidChangeTreeData: vscode.Event<SnippetTreeItem | undefined | void> = this._onDidChangeTreeData.event;

    constructor(private type: 'favorites' | 'recent') { }

    refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: SnippetTreeItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: SnippetTreeItem): Promise<SnippetTreeItem[]> {
        if (element) {
            return Promise.resolve([]);
        }

        if (!(await api.isConfigured())) {
            return Promise.resolve([]);
        }

        let snippets: Snippet[] = [];
        if (this.type === 'favorites') {
            snippets = await api.getSnippets('', true);
        } else {
            snippets = await api.getRecentSnippets();
        }

        return snippets.map(s => new SnippetTreeItem(s));
    }
}
