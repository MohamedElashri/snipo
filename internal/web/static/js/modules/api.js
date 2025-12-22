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
    return response.json();
  },

  get: (url) => api.request('GET', url),
  post: (url, data) => api.request('POST', url, data),
  put: (url, data) => api.request('PUT', url, data),
  delete: (url) => api.request('DELETE', url)
};
