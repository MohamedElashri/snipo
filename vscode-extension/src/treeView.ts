import * as vscode from 'vscode';
import { api, Snippet } from './api';

export class SnippetTreeItem extends vscode.TreeItem {
    constructor(
        public readonly snippet: Snippet,
        public readonly command?: vscode.Command
    ) {
        super(snippet.title, vscode.TreeItemCollapsibleState.None);

        this.description = snippet.language;
        let tooltipText = snippet.description || snippet.title;
        if (snippet.content) {
            // Add preview of content, truncated if too long
            const maxContentPreview = 200;
            const contentPreview = snippet.content.length > maxContentPreview 
                ? snippet.content.substring(0, maxContentPreview) + '...'
                : snippet.content;
            tooltipText += `\n\n${contentPreview}`;
        }
        
        this.tooltip = tooltipText;
        this.contextValue = 'snippet';
        this.iconPath = vscode.ThemeIcon.File;

        // Define command when clicked (Open in VS Code by default)
        this.command = {
            command: 'snipo.openInVSCode',
            title: 'Open in VS Code',
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
