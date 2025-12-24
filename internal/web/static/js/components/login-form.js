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
        const response = await fetch('/api/v1/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ password: this.password })
        });

        const json = await response.json();

        // Handle error response format: { error: { code, message } }
        if (json.error) {
          this.error = json.error.message || 'Invalid password';
          return;
        }

        // Handle success response format: { data: { success, message }, meta }
        const result = json.data || json;
        if (result.success) {
          window.location.href = '/';
        } else {
          this.error = result.message || 'Invalid password';
        }
      } catch (err) {
        this.error = 'Connection error';
      }

      this.loading = false;
    }
  }));
}
