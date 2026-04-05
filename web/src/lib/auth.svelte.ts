import { browser } from '$app/environment';
import { api } from '$lib/api/client';

const TOKEN_KEY = 'auth_token';

function createAuthStore() {
  let user = $state<App.User | null>(null);
  let token = $state<string | null>(null);
  let ready = $state(false);

  async function init() {
    if (!browser) return;
    const stored = localStorage.getItem(TOKEN_KEY);
    if (stored) {
      try {
        const me = await api.auth.me(stored);
        token = stored;
        user = me;
      } catch {
        localStorage.removeItem(TOKEN_KEY);
      }
    }
    ready = true;
  }

  async function login(email: string, password: string) {
    const res = await api.auth.login(email, password);
    token = res.token;
    user = res.user;
    localStorage.setItem(TOKEN_KEY, res.token);
  }

  async function register(email: string, displayName: string, password: string) {
    const res = await api.auth.register(email, displayName, password);
    token = res.token;
    user = res.user;
    localStorage.setItem(TOKEN_KEY, res.token);
  }

  async function logout() {
    if (token) {
      await api.auth.logout(token).catch(() => {});
      localStorage.removeItem(TOKEN_KEY);
    }
    token = null;
    user = null;
  }

  function getToken(): string | null {
    return token;
  }

  return {
    get user() { return user; },
    get token() { return token; },
    get ready() { return ready; },
    init,
    login,
    register,
    logout,
    getToken,
  };
}

export const auth = createAuthStore();
