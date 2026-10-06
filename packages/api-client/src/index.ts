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
  private baseURL: string;
  private token?: string;

  constructor(options: SnipoApiOptions) {
    this.baseURL = options.baseURL.replace(/\/$/, '');
    this.token = options.token;
  }

  updateToken(token: string) {
    this.token = token;
  }

  updateBaseURL(url: string) {
    this.baseURL = url.replace(/\/$/, '');
  }

  private async request<T>(endpoint: string, options: RequestInit & { params?: Record<string, any> } = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...((options.headers as Record<string, string>) || {})
    };
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    let url = endpoint.startsWith('http') ? endpoint : `${this.baseURL}${endpoint}`;
    if (options.params) {
      const urlObj = new URL(url);
      for (const [k, v] of Object.entries(options.params)) {
        if (v !== undefined) urlObj.searchParams.append(k, String(v));
      }
      url = urlObj.toString();
    }

    try {
      const response = await fetch(url, { ...options, headers });
      if (!response.ok) {
        let errData = {};
        try { errData = await response.json(); } catch(e) {}
        const error: any = new Error(`Request failed with status ${response.status}`);
        error.response = { status: response.status, data: errData };
        error.name = 'SnipoApiError';
        throw error;
      }
      if (response.status === 204) {
        return {} as T;
      }
      return await response.json();
    } catch (e: any) {
      if (e.name === 'AbortError') {
        throw e;
      }
      throw e;
    }
  }

  async getSnippets(query?: string, signal?: AbortSignal): Promise<Snippet[]> {
    const params: any = { limit: 100 };
    if (query) params.q = query;
    const response = await this.request<any>(query ? '/api/v1/snippets/search' : '/api/v1/snippets', { 
        params,
        signal 
    });
    return Array.isArray(response) ? response : (response.data || []);
  }

  async getRecentSnippets(): Promise<Snippet[]> {
    const response = await this.request<any>('/api/v1/snippets', { params: { limit: 20, sort: 'updated_at', order: 'desc' } });
    return response.data || [];
  }

  async getSnippet(id: string): Promise<Snippet> {
    return this.request<Snippet>(`/api/v1/snippets/${encodeURIComponent(id)}`);
  }

  async createSnippet(data: Partial<Snippet>): Promise<Snippet> {
    return this.request<Snippet>('/api/v1/snippets', { method: 'POST', body: JSON.stringify(data) });
  }

  async updateSnippet(id: string, data: Partial<Snippet>): Promise<Snippet> {
    return this.request<Snippet>(`/api/v1/snippets/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(data) });
  }

  async deleteSnippet(id: string): Promise<void> {
    await this.request<void>(`/api/v1/snippets/${encodeURIComponent(id)}`, { method: 'DELETE' });
  }

  async getTags(): Promise<Tag[]> {
    const response = await this.request<any>('/api/v1/tags');
    return response.data || response || [];
  }

  async getFolders(): Promise<Folder[]> {
    const response = await this.request<any>('/api/v1/folders', { params: { tree: true } });
    return response.data || response || [];
  }

  // Fetch-like client wrapper for backward compatibility with existing extensions
  public client = {
    get: async (url: string, config?: any) => {
      const data = await this.request<any>(url, { method: 'GET', ...config });
      return { data, status: 200 };
    },
    post: async (url: string, body?: any, config?: any) => {
      const data = await this.request<any>(url, { method: 'POST', body: JSON.stringify(body), ...config });
      return { data, status: 200 };
    },
    put: async (url: string, body?: any, config?: any) => {
      const data = await this.request<any>(url, { method: 'PUT', body: JSON.stringify(body), ...config });
      return { data, status: 200 };
    },
    delete: async (url: string, config?: any) => {
      const data = await this.request<any>(url, { method: 'DELETE', ...config });
      return { data, status: 200 };
    },
    defaults: {
      headers: { common: {} as any },
      baseURL: ''
    }
  };
}
