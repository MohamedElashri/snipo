// Login form component
export function initLoginForm(Alpine) {
  Alpine.data('loginForm', () => ({
    password: '',
    error: '',
    loading: false,

    async login() {
      this.loading = true;
      this.error = '';

      try {
        const result = await window.api.post('/api/v1/auth/login', { password: this.password });

        // Handle error response format: { error: { code, message } }
        if (result && result.error) {
          this.error = result.error.message || 'Invalid password';
          return;
        }

        // Handle success response
        if (result && result.success) {
          const basePath = window.SNIPO_CONFIG?.basePath || '';
          window.location.href = basePath + '/';
        } else {
          this.error = result?.message || 'Invalid password';
        }
      } catch (err) {
        this.error = 'Connection error';
      }

      this.loading = false;
    }
  }));
}
