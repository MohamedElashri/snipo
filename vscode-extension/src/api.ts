import * as vscode from 'vscode';
import { SnipoClient, Snippet, Tag, Folder } from '@snipo/api-client';

export function normalizeServerURL(value: string): string {
    const url = value.trim();
    if (!url || /^[a-z][a-z\d+.-]*:\/\//i.test(url)) return url;
    const local = /^(localhost|127\.0\.0\.1|\[::1\])(:|\/|$)/i.test(url);
    return `${local ? 'http' : 'https'}://${url}`;
}

// Repository/workspace settings must never redirect a token from SecretStorage.
export function configuredServerURL(): string {
    const setting = vscode.workspace.getConfiguration('snipo').inspect<string>('apiUrl');
    return normalizeServerURL(setting?.globalValue ?? setting?.defaultValue ?? '');
}

function isNetworkError(error: any): boolean {
    return !error.response && (error instanceof TypeError || error.code === 'ECONNREFUSED' ||
        /Network Error|Failed to fetch|fetch failed|NetworkError/i.test(error.message || ''));
}

function refreshSidebar(): void {
    for (const command of ['snipo.refreshTree', 'snipo.refreshTagsTree']) {
        void vscode.commands.executeCommand(command).then(undefined, error => console.error('Sidebar refresh failed', error));
    }
}

// Extending types if necessary for VSCode specific properties, though we can cast
export interface VSTag extends Tag {
    snippet_count?: number;
}

export interface VSFolder extends Folder {
    icon?: string;
    snippet_count?: number;
    children?: VSFolder[];
}

export class SnipoAPI {
    private client: SnipoClient;
    private context?: vscode.ExtensionContext;
    private cacheScope = '';

    private cacheKey(key: string): string {
        return JSON.stringify(['snipo-cache', this.cacheScope, key]);
    }

    constructor() {
        this.client = new SnipoClient({ baseURL: '' });
    }

    async init(context: vscode.ExtensionContext) {
        this.context = context;
        await this.updateConfig();

        context.subscriptions.push(vscode.workspace.onDidChangeConfiguration(async e => {
            if (e.affectsConfiguration('snipo.apiUrl')) {
                await this.updateConfig();
            }
        }));

        context.subscriptions.push(context.secrets.onDidChange(async e => {
            if (e.key === 'snipo.apiToken') {
                await this.updateConfig();
            }
        }));
    }

    private async updateConfig() {
        const apiUrl = configuredServerURL();
        let apiToken = undefined;
        
        if (this.context) {
            apiToken = await this.context.secrets.get('snipo.apiToken');
        }

        this.cacheScope = (apiUrl || '').replace(/\/+$/, '');
        this.client.updateBaseURL(apiUrl || '');
        this.client.updateToken(apiToken);
    }

    async isConfigured(): Promise<boolean> {
        const apiUrl = configuredServerURL();
        let apiToken = undefined;
        if (this.context) {
            apiToken = await this.context.secrets.get('snipo.apiToken');
        } else {
            console.warn('isConfigured called before SnipoAPI was initialized with a context.');
            return false;
        }
        return !!(apiUrl && apiToken);
    }

    async verifyConfiguration(apiUrl: string, apiToken: string): Promise<boolean> {
        apiUrl = normalizeServerURL(apiUrl);
        try {
            const tempClient = new SnipoClient({
                baseURL: apiUrl,
                token: apiToken
            });
            // Try fetching a single snippet to verify the API token
            const response = await tempClient.client.get('/api/v1/snippets', { params: { limit: 1 }, timeout: 5000 });
            return response.status === 200;
        } catch (error) {
            console.error('Configuration verification failed:', error);
            return false;
        }
    }

    async getSnippets(query: string = '', isFavorite?: boolean, tagId?: number, folderId?: number): Promise<Snippet[]> {
        let cacheKey = `snippets_all_${isFavorite}`;
        if (tagId) cacheKey = `snippets_tag_${tagId}`;
        if (folderId) cacheKey = `snippets_folder_${folderId}`;
        cacheKey = this.cacheKey(cacheKey);
        const scope = this.cacheScope;

        try {
            const params: any = { limit: 50 };
            if (query) params.q = query;
            if (isFavorite !== undefined) params.favorite = isFavorite;
            if (tagId !== undefined) params.tag_ids = tagId;
            if (folderId !== undefined) params.folder_ids = folderId;

            const data: Snippet[] = [];
            let page = 1;
            let totalPages = 1;
            do {
                if (scope !== this.cacheScope) return [];
                const response = await this.client.client.get('/api/v1/snippets', { params: { ...params, page } });
                if (scope !== this.cacheScope) return [];
                data.push(...(response.data.data || []));
                totalPages = response.data.pagination?.totalPages || 1;
                page++;
            } while (page <= totalPages);
            
            // Cache full lists (not search queries)
            if (this.context && !query) {
                await this.context.globalState.update(cacheKey, data);
            }
            return data;
        } catch (error: any) {
            if (scope !== this.cacheScope) return [];
            console.error('Error fetching snippets', error);
            if (error.response?.status === 401) {
                vscode.window.showErrorMessage('Snipo API Token is invalid. Please update it in settings.');
            } else if (isNetworkError(error)) {
                vscode.window.showErrorMessage('Failed to connect to Snipo server. Showing cached snippets (if any).');
            } else {
                vscode.window.showErrorMessage('Failed to fetch snippets from Snipo');
            }
            
            if (this.context && !query) {
                return this.context.globalState.get<Snippet[]>(cacheKey) || [];
            }
            return [];
        }
    }

    async searchSnippets(query: string, signal?: AbortSignal): Promise<Snippet[]> {
        const cacheKey = this.cacheKey('snippets_all_undefined');
        try {
            const response = await this.client.client.get('/api/v1/snippets/search', { params: { q: query }, signal });
            return Array.isArray(response.data) ? response.data : (response.data.data || []);
        } catch (error: any) {
            if (error.name === 'AbortError') {
                return [];
            }
            console.error('Error searching snippets', error);
            
            // Offline fallback for search
            if (this.context && (isNetworkError(error))) {
                const cachedSnippets = this.context.globalState.get<Snippet[]>(cacheKey) || [];
                const lowerQuery = query.toLowerCase();
                return cachedSnippets.filter(s => 
                    s.title.toLowerCase().includes(lowerQuery) || 
                    (s.description && s.description.toLowerCase().includes(lowerQuery)) ||
                    s.content.toLowerCase().includes(lowerQuery)
                );
            }

            if (error.response?.status === 401) {
                vscode.window.showErrorMessage('Snipo API Token is invalid. Please update it in settings.');
            } else {
                vscode.window.showErrorMessage('Failed to search snippets from Snipo');
            }
            return [];
        }
    }

    async getSnippet(id: string): Promise<Snippet | null> {
        const cacheKey = this.cacheKey('snippets_all_undefined');
        const recentKey = this.cacheKey('recent_snippets');
        try {
            return await this.client.getSnippet(id);
        } catch (error: any) {
            console.error('Error fetching snippet', error);
            // Offline fallback
            if (this.context && (isNetworkError(error))) {
                // Try to find in cache
                const cached = this.context.globalState.get<Snippet[]>(cacheKey) || [];
                const found = cached.find(s => s.id === id);
                if (found) return found;
                
                const cachedRecent = this.context.globalState.get<Snippet[]>(recentKey) || [];
                const foundRecent = cachedRecent.find(s => s.id === id);
                if (foundRecent) return foundRecent;
            }
            return null;
        }
    }

    async createSnippet(data: Partial<Snippet>): Promise<Snippet | null> {
        try {
            const snippet = await this.client.createSnippet(data);
            refreshSidebar();
            vscode.window.showInformationMessage('Snippet created successfully!');
            return snippet;
        } catch (error: any) {
            console.error('Error creating snippet', error);
            if (error.response?.status === 401) {
                vscode.window.showErrorMessage('Snipo API Token is invalid. Please update it in settings.');
            } else {
                const msg = error.response?.data?.error?.message || error.message;
                const details = error.response?.data?.error?.details?.[0]?.message || '';
                vscode.window.showErrorMessage(`Failed to create snippet: ${msg} ${details}`);
            }
            return null;
        }
    }

    async updateSnippet(id: string, data: Partial<Snippet>): Promise<Snippet | null> {
        try {
            const snippet = await this.client.updateSnippet(id, data);
            refreshSidebar();
            vscode.window.showInformationMessage('Snippet updated successfully!');
            return snippet;
        } catch (error: any) {
            console.error('Error updating snippet', error);
            if (error.response?.status === 401) {
                vscode.window.showErrorMessage('Snipo API Token is invalid. Please update it in settings.');
            } else {
                const msg = error.response?.data?.error?.message || error.message;
                const details = error.response?.data?.error?.details?.[0]?.message || '';
                vscode.window.showErrorMessage(`Failed to update snippet: ${msg} ${details}`);
            }
            return null;
        }
    }

    async deleteSnippet(id: string): Promise<boolean> {
        try {
            await this.client.deleteSnippet(id);
            refreshSidebar();
            vscode.window.showInformationMessage('Snippet deleted successfully!');
            return true;
        } catch (error: any) {
            console.error('Error deleting snippet', error);
            if (error.response?.status === 401) {
                vscode.window.showErrorMessage('Snipo API Token is invalid. Please update it in settings.');
            } else {
                vscode.window.showErrorMessage('Failed to delete snippet');
            }
            return false;
        }
    }

    async getRecentSnippets(): Promise<Snippet[]> {
        const cacheKey = this.cacheKey('recent_snippets');
        try {
            const data = await this.client.getRecentSnippets();
            if (this.context) {
                await this.context.globalState.update(cacheKey, data);
            }
            return data;
        } catch (error) {
            console.error('Error fetching recent snippets', error);
            if (this.context) {
                return this.context.globalState.get<Snippet[]>(cacheKey) || [];
            }
            return [];
        }
    }

    async getTags(): Promise<VSTag[]> {
        const cacheKey = this.cacheKey('tags_all');
        try {
            const data = await this.client.getTags();
            if (this.context) {
                await this.context.globalState.update(cacheKey, data);
            }
            return data as VSTag[];
        } catch (error) {
            console.error('Error fetching tags', error);
            if (this.context) {
                return this.context.globalState.get<VSTag[]>(cacheKey) || [];
            }
            return [];
        }
    }

    async getFolders(): Promise<VSFolder[]> {
        const cacheKey = this.cacheKey('folders_all');
        try {
            const data = await this.client.getFolders();
            if (this.context) {
                await this.context.globalState.update(cacheKey, data);
            }
            return data as VSFolder[];
        } catch (error) {
            console.error('Error fetching folders', error);
            if (this.context) {
                return this.context.globalState.get<VSFolder[]>(cacheKey) || [];
            }
            return [];
        }
    }
}

export const api = new SnipoAPI();

export { Snippet };
export type { VSTag as Tag, VSFolder as Folder };
