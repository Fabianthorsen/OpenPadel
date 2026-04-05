const BASE = '/api';

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number
  ) {
    super(message);
  }
}

async function request<T>(method: string, path: string, body?: unknown, token?: string): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (res.status === 204) return undefined as T;
  const data = await res.json();
  if (!res.ok) throw new ApiError(data.error ?? 'request failed', res.status);
  return data as T;
}

export const api = {
  auth: {
    register: (email: string, displayName: string, password: string) =>
      request<{ token: string; user: App.User }>('POST', '/auth/register', { email, display_name: displayName, password }),
    login: (email: string, password: string) =>
      request<{ token: string; user: App.User }>('POST', '/auth/login', { email, password }),
    logout: (token: string) =>
      request<void>('POST', '/auth/logout', undefined, token),
    me: (token: string) =>
      request<App.User>('GET', '/auth/me', undefined, token),
    profile: (token: string) =>
      request<{ user: App.User; stats: App.CareerStats }>('GET', '/auth/profile', undefined, token),
    history: (token: string) =>
      request<{ tournaments: App.TournamentEntry[] }>('GET', '/auth/history', undefined, token),
    deleteAccount: (token: string) =>
      request<void>('DELETE', '/auth/account', undefined, token),
    forgotPassword: (email: string) =>
      request<void>('POST', '/auth/forgot', { email }),
    resetPassword: (token: string, password: string) =>
      request<void>('POST', '/auth/reset', { token, password }),
  },
  sessions: {
    create: (courts: number, points: number, name: string) =>
      request<App.Session>('POST', '/sessions', { courts, points, name }),
    get: (id: string, token?: string) =>
      request<App.Session>('GET', `/sessions/${id}`, undefined, token),
    start: (id: string, token: string) =>
      request<App.Session>('POST', `/sessions/${id}/start`, undefined, token),
    cancel: (id: string, token: string) =>
      request<void>('DELETE', `/sessions/${id}`, undefined, token),
  },
  players: {
    join: (sessionId: string, name: string, token?: string) =>
      request<App.Player>('POST', `/sessions/${sessionId}/players`, { name }, token),
    remove: (sessionId: string, playerId: string, token: string) =>
      request<{ id: string; active: boolean }>(
        'DELETE',
        `/sessions/${sessionId}/players/${playerId}`,
        undefined,
        token
      ),
  },
  rounds: {
    all: (sessionId: string) =>
      request<{ rounds: App.Round[] }>('GET', `/sessions/${sessionId}/rounds`),
    current: (sessionId: string) =>
      request<App.Round>('GET', `/sessions/${sessionId}/rounds/current`),
    advance: (sessionId: string, token: string) =>
      request<void>('POST', `/sessions/${sessionId}/rounds/advance`, undefined, token),
  },
  scores: {
    submit: (sessionId: string, matchId: string, scoreA: number, scoreB: number, token: string) =>
      request<App.Match>('PUT', `/sessions/${sessionId}/matches/${matchId}/score`, {
        score_a: scoreA,
        score_b: scoreB,
        token,
      }, token),
  },
  leaderboard: {
    get: (sessionId: string) =>
      request<App.Leaderboard>('GET', `/sessions/${sessionId}/leaderboard`),
  },
};
