// API helper module
export const api = {
  getBasePath() {
    return window.SNIPO_CONFIG?.basePath || '';
  },

  validateUrl(url) {
    // Only allow relative URLs starting with /
    if (!url.startsWith('/')) {
      throw new Error('Only relative URLs are allowed');
    }

    // Prevent path traversal attacks
    if (url.includes('..')) {
      throw new Error('Path traversal is not allowed');
    }

    // Ensure URL stays within API namespace
    if (!url.startsWith('/api/')) {
      throw new Error('Only API endpoints are allowed');
    }

    // Additional validation: no protocol, hostname, or @ symbols
    if (url.includes('://') || url.includes('@')) {
      throw new Error('Invalid URL format');
    }

    return true;
  },

  async request(method, url, data = null) {
    // Validate URL before processing
    this.validateUrl(url);

    // Prepend base path to URL if it's a relative path
    const basePath = this.getBasePath();
    const fullUrl = url.startsWith('/') ? basePath + url : url;

    const options = {
      method,
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include'
    };
    if (data) options.body = JSON.stringify(data);

    const response = await fetch(fullUrl, options);
    if (response.status === 401) {
      // Only redirect to login if we're not already on the home or login page
      // This prevents redirect loops when login is disabled in settings
      const currentPath = window.location.pathname;
      const loginPath = basePath + '/login';
      const homePath = basePath + '/';
      if (currentPath !== homePath && currentPath !== loginPath) {
        window.location.href = loginPath;
      }
      return null;
    }
    if (response.status === 204) return null;
    
    const json = await response.json();
    
    // Handle error responses: { error: { code, message, details } }
    if (json && json.error) {
      // Return error in the format frontend expects
      return { error: json.error };
    }
    
    // Unwrap the envelope format: { data, meta, pagination }
    // For list responses, preserve pagination alongside data
    if (json && typeof json === 'object') {
      if (json.pagination) {
        // List response: return both data and pagination
        return {
          data: json.data,
          pagination: json.pagination,
          meta: json.meta
        };
      } else if (json.data !== undefined) {
        // Single resource: return just the data
        return json.data;
      }
    }
    
    // Fallback for responses that don't match the envelope format
    return json;
  },

  get: (url) => api.request('GET', url),
  post: (url, data) => api.request('POST', url, data),
  put: (url, data) => api.request('PUT', url, data),
  delete: (url, options = {}) => {
    // Support passing body in options for DELETE requests
    const data = options.body ? JSON.parse(options.body) : null;
    return api.request('DELETE', url, data);
  }
};
