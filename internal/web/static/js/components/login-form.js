import { theme } from '../modules/theme.js';
import { api } from '../modules/api.js';

// Resolve return destinations exactly as the browser will navigate them. Prefix
// checks alone allow protocol-relative URLs and backslash normalization tricks.
export function loginDestination(nextUrl, basePath = '') {
  const origin = window.location.origin;
  const home = new URL(basePath + '/', origin);
  if (home.origin !== origin) return new URL('/', origin).href;
  if (!nextUrl || /[\u0000-\u001f\u007f\\]/.test(nextUrl) || nextUrl.startsWith('//')) return home.href;
  if (!nextUrl.startsWith('/') && !nextUrl.startsWith('?')) return home.href;
  try {
    const target = new URL(nextUrl.startsWith('?') ? home.href + nextUrl : basePath + nextUrl, origin);
    if (target.origin !== origin) return home.href;
    if (basePath && target.pathname !== basePath && !target.pathname.startsWith(basePath + '/')) return home.href;
    // Return an absolute URL: serializing a path beginning with // would turn it
    // into a different host when assigned to location.href a second time.
    return target.href;
  } catch { return home.href; }
}

// Login form component
export function initLoginForm(Alpine) {
  Alpine.data('loginForm', () => ({
    theme,
    password: '',
    error: '',
    loading: false,

    async login() {
      if (this.loading) return;
      this.loading = true;
      this.error = '';

      try {
        const result = await api.post('/api/v1/auth/login', { password: this.password });

        // Handle error response format: { error: { code, message } }
        if (result && result.error) {
          this.error = result.error.message || 'Invalid password';
          return;
        }

        // Handle success response
        if (result && result.success) {
          const basePath = window.SNIPO_CONFIG?.basePath || '';
          const urlParams = new URLSearchParams(window.location.search);
          window.location.href = loginDestination(urlParams.get('next'), basePath);
        } else {
          this.error = result?.message || 'Invalid password';
        }
      } catch (err) {
        this.error = 'Connection error';
      } finally {
        this.loading = false;
      }
    }
  }));
}
