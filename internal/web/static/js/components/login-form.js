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

        const result = await response.json();

        if (result.success) {
          window.location.href = '/';
        } else {
          this.error = result.error?.message || 'Invalid password';
        }
      } catch (err) {
        this.error = 'Connection error';
      }

      this.loading = false;
    }
  }));
}
