import * as vscode from 'vscode';
import { api, Snippet, Tag, Folder } from './api';

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

export class TagTreeItem extends vscode.TreeItem {
    constructor(public readonly tag: Tag) {
        super(tag.name, vscode.TreeItemCollapsibleState.Collapsed);
        this.tooltip = `Tag: ${tag.name}`;
        this.contextValue = 'tag';
        this.iconPath = new vscode.ThemeIcon('tag');
        this.description = tag.snippet_count ? `${tag.snippet_count}` : undefined;
    }
}

export class FolderTreeItem extends vscode.TreeItem {
    constructor(public readonly folder: Folder) {
        super(folder.name, vscode.TreeItemCollapsibleState.Collapsed);
        this.tooltip = `Collection: ${folder.name}`;
        this.contextValue = 'folder';
        this.iconPath = new vscode.ThemeIcon('folder-library');
        this.description = folder.snippet_count ? `${folder.snippet_count}` : undefined;
    }
}

export class TagsTreeProvider implements vscode.TreeDataProvider<vscode.TreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<vscode.TreeItem | undefined | void> = new vscode.EventEmitter<vscode.TreeItem | undefined | void>();
    readonly onDidChangeTreeData: vscode.Event<vscode.TreeItem | undefined | void> = this._onDidChangeTreeData.event;

    refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: vscode.TreeItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: vscode.TreeItem): Promise<vscode.TreeItem[]> {
        if (!(await api.isConfigured())) {
            return Promise.resolve([]);
        }

        if (!element) {
            // Root level: Tags and Collections
            const tagsNode = new vscode.TreeItem('Tags', vscode.TreeItemCollapsibleState.Expanded);
            tagsNode.contextValue = 'root_tags';
            tagsNode.iconPath = new vscode.ThemeIcon('tag');

            const collectionsNode = new vscode.TreeItem('Collections', vscode.TreeItemCollapsibleState.Expanded);
            collectionsNode.contextValue = 'root_collections';
            collectionsNode.iconPath = new vscode.ThemeIcon('folder-library');

            return [tagsNode, collectionsNode];
        }

        if (element.contextValue === 'root_tags') {
            const tags = await api.getTags();
            return tags.map(t => new TagTreeItem(t));
        }

        if (element.contextValue === 'root_collections') {
            const folders = await api.getFolders();
            return folders.map(f => new FolderTreeItem(f));
        }

        if (element instanceof TagTreeItem) {
            const snippets = await api.getSnippets('', undefined, element.tag.id);
            return snippets.map(s => new SnippetTreeItem(s));
        }

        if (element instanceof FolderTreeItem) {
            let items: vscode.TreeItem[] = [];
            // If folder has children, add them first
            if (element.folder.children && element.folder.children.length > 0) {
                items = items.concat(element.folder.children.map(f => new FolderTreeItem(f)));
            }
            // Add snippets in this folder
            const snippets = await api.getSnippets('', undefined, undefined, element.folder.id);
            items = items.concat(snippets.map(s => new SnippetTreeItem(s)));
            return items;
        }

        return [];
    }
}
