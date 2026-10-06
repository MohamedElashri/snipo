import * as vscode from 'vscode';
import { SnipoClient, Snippet, Tag, Folder } from '@snipo/api-client';

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

    constructor() {
        this.client = new SnipoClient({ baseURL: '' });
    }

    async init(context: vscode.ExtensionContext) {
        this.context = context;
        await this.updateConfig();

        vscode.workspace.onDidChangeConfiguration(async e => {
            if (e.affectsConfiguration('snipo.apiUrl')) {
                await this.updateConfig();
            }
        });

        context.secrets.onDidChange(async e => {
            if (e.key === 'snipo.apiToken') {
                await this.updateConfig();
            }
        });
    }

    private async updateConfig() {
        const config = vscode.workspace.getConfiguration('snipo');
        let apiUrl = config.get<string>('apiUrl');
        let apiToken = undefined;
        
        if (this.context) {
            apiToken = await this.context.secrets.get('snipo.apiToken');
        }

        if (apiUrl) {
            if (!apiUrl.startsWith('http://') && !apiUrl.startsWith('https://')) {
                apiUrl = apiUrl.startsWith('localhost') || apiUrl.startsWith('127.0.0.1')
                    ? `http://${apiUrl}`
                    : `https://${apiUrl}`;
            }
            this.client.updateBaseURL(apiUrl);
        }
        if (apiToken) {
            this.client.updateToken(apiToken);
        }
    }

    async isConfigured(): Promise<boolean> {
        const config = vscode.workspace.getConfiguration('snipo');
        const apiUrl = config.get<string>('apiUrl');
        let apiToken = undefined;
        if (this.context) {
            apiToken = await this.context.secrets.get('snipo.apiToken');
        }
        return !!(apiUrl && apiToken);
    }

    async verifyConfiguration(apiUrl: string, apiToken: string): Promise<boolean> {
        if (!apiUrl.startsWith('http://') && !apiUrl.startsWith('https://')) {
            apiUrl = apiUrl.startsWith('localhost') || apiUrl.startsWith('127.0.0.1')
                ? `http://${apiUrl}`
                : `https://${apiUrl}`;
        }
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

        try {
            const params: any = { limit: 50 };
            if (query) params.q = query;
            if (isFavorite !== undefined) params.favorite = isFavorite;
            if (tagId !== undefined) params.tag_ids = tagId;
            if (folderId !== undefined) params.folder_ids = folderId;

            const response = await this.client.client.get('/api/v1/snippets', { params });
            const data = response.data.data || [];
            
            // Cache full lists (not search queries)
            if (this.context && !query) {
                await this.context.globalState.update(cacheKey, data);
            }
            return data;
        } catch (error: any) {
            console.error('Error fetching snippets', error);
            if (error.response?.status === 401) {
                vscode.window.showErrorMessage('Snipo API Token is invalid. Please update it in settings.');
            } else if (error.code === 'ECONNREFUSED' || error.message?.includes('Network Error')) {
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
        try {
            const response = await this.client.client.get('/api/v1/snippets/search', { params: { q: query }, signal });
            return Array.isArray(response.data) ? response.data : (response.data.data || []);
        } catch (error: any) {
            if (error.name === 'AbortError') {
                return [];
            }
            console.error('Error searching snippets', error);
            
            // Offline fallback for search
            if (this.context && (error.code === 'ECONNREFUSED' || error.message?.includes('Network Error'))) {
                const cachedSnippets = this.context.globalState.get<Snippet[]>('snippets_all_undefined') || [];
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
        try {
            return await this.client.getSnippet(id);
        } catch (error: any) {
            console.error('Error fetching snippet', error);
            // Offline fallback
            if (this.context && (error.code === 'ECONNREFUSED' || error.message?.includes('Network Error'))) {
                // Try to find in cache
                const cached = this.context.globalState.get<Snippet[]>('snippets_all_undefined') || [];
                const found = cached.find(s => s.id === id);
                if (found) return found;
                
                const cachedRecent = this.context.globalState.get<Snippet[]>('recent_snippets') || [];
                const foundRecent = cachedRecent.find(s => s.id === id);
                if (foundRecent) return foundRecent;
            }
            return null;
        }
    }

    async createSnippet(data: Partial<Snippet>): Promise<Snippet | null> {
        try {
            const snippet = await this.client.createSnippet(data);
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
        const cacheKey = 'recent_snippets';
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
        const cacheKey = 'tags_all';
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
        const cacheKey = 'folders_all';
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
