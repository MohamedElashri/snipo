// API helper module
export const api = {
  async request(method, url, data = null) {
    const options = {
      method,
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include'
    };
    if (data) options.body = JSON.stringify(data);

    const response = await fetch(url, options);
    if (response.status === 401) {
      window.location.href = '/login';
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
  delete: (url) => api.request('DELETE', url)
};
