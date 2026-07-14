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
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.api = exports.SnipoAPI = void 0;
const axios_1 = __importDefault(require("axios"));
const vscode = __importStar(require("vscode"));
class SnipoAPI {
    client;
    context;
    constructor() {
        this.client = axios_1.default.create();
    }
    async init(context) {
        this.context = context;
        await this.updateConfig();
        vscode.workspace.onDidChangeConfiguration(async (e) => {
            if (e.affectsConfiguration('snipo.apiUrl')) {
                await this.updateConfig();
            }
        });
        context.secrets.onDidChange(async (e) => {
            if (e.key === 'snipo.apiToken') {
                await this.updateConfig();
            }
        });
    }
    async updateConfig() {
        const config = vscode.workspace.getConfiguration('snipo');
        let apiUrl = config.get('apiUrl');
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
            this.client.defaults.baseURL = apiUrl;
        }
        if (apiToken) {
            this.client.defaults.headers.common['Authorization'] = `Bearer ${apiToken}`;
        }
    }
    async isConfigured() {
        const config = vscode.workspace.getConfiguration('snipo');
        const apiUrl = config.get('apiUrl');
        let apiToken = undefined;
        if (this.context) {
            apiToken = await this.context.secrets.get('snipo.apiToken');
        }
        return !!(apiUrl && apiToken);
    }
    async verifyConfiguration(apiUrl, apiToken) {
        if (!apiUrl.startsWith('http://') && !apiUrl.startsWith('https://')) {
            apiUrl = apiUrl.startsWith('localhost') || apiUrl.startsWith('127.0.0.1')
                ? `http://${apiUrl}`
                : `https://${apiUrl}`;
        }
        try {
            const tempClient = axios_1.default.create({
                baseURL: apiUrl,
                headers: { 'Authorization': `Bearer ${apiToken}` },
                timeout: 5000 // 5 seconds timeout
            });
            // Try fetching a single snippet to verify the API token
            const response = await tempClient.get('/api/v1/snippets', { params: { limit: 1 } });
            return response.status === 200;
        }
        catch (error) {
            console.error('Configuration verification failed:', error);
            return false;
        }
    }
    async getSnippets(query = '', isFavorite) {
        try {
            const params = { limit: 50 };
            if (query)
                params.q = query;
            if (isFavorite !== undefined)
                params.favorite = isFavorite;
            const response = await this.client.get('/api/v1/snippets', { params });
            return response.data.data || [];
        }
        catch (error) {
            console.error('Error fetching snippets', error);
            if (error.response?.status === 401) {
                vscode.window.showErrorMessage('Snipo API Token is invalid. Please update it in settings.');
            }
            else {
                vscode.window.showErrorMessage('Failed to fetch snippets from Snipo');
            }
            return [];
        }
    }
    async searchSnippets(query) {
        try {
            const response = await this.client.get('/api/v1/snippets/search', { params: { q: query } });
            return Array.isArray(response.data) ? response.data : (response.data.data || []);
        }
        catch (error) {
            console.error('Error searching snippets', error);
            if (error.response?.status === 401) {
                vscode.window.showErrorMessage('Snipo API Token is invalid. Please update it in settings.');
            }
            else {
                vscode.window.showErrorMessage('Failed to search snippets from Snipo');
            }
            return [];
        }
    }
    async createSnippet(data) {
        try {
            const response = await this.client.post('/api/v1/snippets', data);
            vscode.window.showInformationMessage('Snippet created successfully!');
            return response.data;
        }
        catch (error) {
            console.error('Error creating snippet', error);
            if (error.response?.status === 401) {
                vscode.window.showErrorMessage('Snipo API Token is invalid. Please update it in settings.');
            }
            else {
                const msg = error.response?.data?.error?.message || error.message;
                const details = error.response?.data?.error?.details?.[0]?.message || '';
                vscode.window.showErrorMessage(`Failed to create snippet: ${msg} ${details}`);
            }
            return null;
        }
    }
    async updateSnippet(id, data) {
        try {
            const response = await this.client.put(`/api/v1/snippets/${id}`, data);
            vscode.window.showInformationMessage('Snippet updated successfully!');
            return response.data;
        }
        catch (error) {
            console.error('Error updating snippet', error);
            if (error.response?.status === 401) {
                vscode.window.showErrorMessage('Snipo API Token is invalid. Please update it in settings.');
            }
            else {
                const msg = error.response?.data?.error?.message || error.message;
                const details = error.response?.data?.error?.details?.[0]?.message || '';
                vscode.window.showErrorMessage(`Failed to update snippet: ${msg} ${details}`);
            }
            return null;
        }
    }
    async deleteSnippet(id) {
        try {
            await this.client.delete(`/api/v1/snippets/${id}`);
            vscode.window.showInformationMessage('Snippet deleted successfully!');
            return true;
        }
        catch (error) {
            console.error('Error deleting snippet', error);
            if (error.response?.status === 401) {
                vscode.window.showErrorMessage('Snipo API Token is invalid. Please update it in settings.');
            }
            else {
                vscode.window.showErrorMessage('Failed to delete snippet');
            }
            return false;
        }
    }
    async getRecentSnippets() {
        // We can just get snippets sorted by updated_at descending
        try {
            const response = await this.client.get('/api/v1/snippets', { params: { limit: 20, sort: 'updated_at', order: 'desc' } });
            return response.data.data || [];
        }
        catch (error) {
            console.error('Error fetching recent snippets', error);
            return [];
        }
    }
}
exports.SnipoAPI = SnipoAPI;
exports.api = new SnipoAPI();
//# sourceMappingURL=api.js.map