import { browser } from '$app/environment';
import { api } from '$lib/api/client';
import { clearAccountResumeSession } from '$lib/resumeSession';

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
				await recoverSessions(stored);
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
		await recoverSessions(res.token);
	}

	async function register(
		email: string,
		displayName: string,
		password: string,
		selfRating: number
	) {
		const res = await api.auth.register(email, displayName, password, selfRating);
		token = res.token;
		user = res.user;
		localStorage.setItem(TOKEN_KEY, res.token);
		await recoverSessions(res.token);
	}

	async function recoverSessions(authToken: string) {
		try {
			const res = await api.auth.getSessions(authToken);
			if (res.sessions) {
				for (const session of res.sessions) {
					if (session.admin_token) {
						localStorage.setItem(`admin_token_${session.id}`, session.admin_token);
					}
				}
			}
		} catch {
			// Silently fail if we can't recover sessions
		}
	}

	async function logout() {
		if (token) {
			await api.auth.logout(token).catch(() => {});
			localStorage.removeItem(TOKEN_KEY);
		}
		// A logged-out device must not carry a registered user's session context:
		// forget any account-owned resume pointer so it stops being resumable (#252).
		clearAccountResumeSession();
		token = null;
		user = null;
	}

	function getToken(): string | null {
		return token;
	}

	function updateUser(updated: App.User) {
		user = updated;
	}

	return {
		get user() {
			return user;
		},
		get token() {
			return token;
		},
		get ready() {
			return ready;
		},
		init,
		login,
		register,
		logout,
		getToken,
		updateUser
	};
}

export const auth = createAuthStore();
