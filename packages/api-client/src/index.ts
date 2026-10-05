import axios, { AxiosInstance, CancelToken } from 'axios';

export interface Snippet {
  id: string;
  title: string;
  description: string;
  content: string;
  language: string;
  is_favorite: boolean;
  is_public: boolean;
  folder_id?: number;
  tags?: Tag[];
  files?: SnippetFile[];
}

export interface SnippetFile {
  id: number;
  filename: string;
  content: string;
  language: string;
}

export interface Tag {
  id: number;
  name: string;
  color: string;
}

export interface Folder {
  id: number;
  name: string;
  parent_id?: number;
  children?: Folder[];
}

export interface SnipoApiOptions {
  baseURL: string;
  token?: string;
}

export class SnipoClient {
  public client: AxiosInstance;

  constructor(options: SnipoApiOptions) {
    this.client = axios.create({
      baseURL: options.baseURL,
      headers: options.token ? { Authorization: `Bearer ${options.token}` } : {}
    });
  }

  updateToken(token: string) {
    this.client.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  }

  updateBaseURL(url: string) {
    this.client.defaults.baseURL = url;
  }

  async getSnippets(query?: string, cancelToken?: CancelToken): Promise<Snippet[]> {
    const params: any = { limit: 100 };
    if (query) params.q = query;
    const response = await this.client.get(query ? '/api/v1/snippets/search' : '/api/v1/snippets', { 
        params,
        cancelToken 
    });
    return Array.isArray(response.data) ? response.data : (response.data.data || []);
  }

  async getRecentSnippets(): Promise<Snippet[]> {
    const response = await this.client.get('/api/v1/snippets', { params: { limit: 20, sort: 'updated_at', order: 'desc' } });
    return response.data.data || [];
  }

  async getSnippet(id: string): Promise<Snippet> {
    const response = await this.client.get(`/api/v1/snippets/${encodeURIComponent(id)}`);
    return response.data;
  }

  async createSnippet(data: Partial<Snippet>): Promise<Snippet> {
    const response = await this.client.post('/api/v1/snippets', data);
    return response.data;
  }

  async updateSnippet(id: string, data: Partial<Snippet>): Promise<Snippet> {
    const response = await this.client.put(`/api/v1/snippets/${encodeURIComponent(id)}`, data);
    return response.data;
  }

  async deleteSnippet(id: string): Promise<void> {
    await this.client.delete(`/api/v1/snippets/${encodeURIComponent(id)}`);
  }

  async getTags(): Promise<Tag[]> {
    const response = await this.client.get('/api/v1/tags');
    return response.data.data || response.data || [];
  }

  async getFolders(): Promise<Folder[]> {
    const response = await this.client.get('/api/v1/folders', { params: { tree: true } });
    return response.data.data || response.data || [];
  }
}
