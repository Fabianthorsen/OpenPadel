const BASE = '/api';

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number
  ) {
    super(message);
  }
}

async function request<T>(method: string, path: string, body?: unknown, token?: string, extraHeaders?: Record<string, string>): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  if (extraHeaders) Object.assign(headers, extraHeaders);

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
      request<{ user: App.User; stats: App.AmericanoCareerStats; tennis_stats: App.TennisCareerStats }>('GET', '/auth/profile', undefined, token),
    history: (token: string) =>
      request<{ tournaments: App.TournamentEntry[]; upcoming: App.UpcomingEntry[] }>('GET', '/auth/history', undefined, token),
    deleteAccount: (token: string) =>
      request<void>('DELETE', '/auth/account', undefined, token),
    forgotPassword: (email: string) =>
      request<void>('POST', '/auth/forgot', { email }),
    resetPassword: (token: string, password: string) =>
      request<void>('POST', '/auth/reset', { token, password }),
  },
  sessions: {
    create: (courts: number, points: number, name: string, gameMode: string, setsToWin: number, gamesPerSet: number, scheduledAt?: string) =>
      request<App.Session>('POST', '/sessions', { courts, points, name, game_mode: gameMode, sets_to_win: setsToWin, games_per_set: gamesPerSet, scheduled_at: scheduledAt }),
    get: (id: string, token?: string) =>
      request<App.Session>('GET', `/sessions/${id}`, undefined, token),
    start: (id: string, token: string) =>
      request<App.Session>('POST', `/sessions/${id}/start`, undefined, token),
    cancel: (id: string, token: string) =>
      request<void>('DELETE', `/sessions/${id}`, undefined, token),
    close: (id: string, token: string) =>
      request<void>('POST', `/sessions/${id}/close`, undefined, token),
  },
  players: {
    join: (sessionId: string, name: string, token?: string, adminToken?: string) =>
      request<App.Player>('POST', `/sessions/${sessionId}/players`, { name }, token, adminToken ? { 'X-Admin-Token': adminToken } : undefined),
    remove: (sessionId: string, playerId: string, token: string) =>
      request<{ id: string; active: boolean }>(
        'DELETE',
        `/sessions/${sessionId}/players/${playerId}`,
        undefined,
        token
      ),
    leave: (sessionId: string, playerId: string) =>
      request<{ id: string; active: boolean }>(
        'DELETE',
        `/sessions/${sessionId}/players/${playerId}`,
        undefined,
        undefined,
        { 'X-Player-Id': playerId }
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
    updateLive: (sessionId: string, matchId: string, a: number, b: number, server: string) =>
      request<void>('PATCH', `/sessions/${sessionId}/matches/${matchId}/score`, { a, b, server }),
  },
  tennis: {
    setTeams: (sessionId: string, teams: { player_id: string; team: 'a' | 'b' }[], adminToken: string) =>
      request<void>('PUT', `/sessions/${sessionId}/tennis/teams`, { teams }, adminToken),
    getMatch: (sessionId: string) =>
      request<App.TennisMatch>('GET', `/sessions/${sessionId}/tennis/match`),
    addPoint: (sessionId: string, team: 'a' | 'b') =>
      request<App.TennisMatch>('POST', `/sessions/${sessionId}/tennis/point/${team}`),
    setServer: (sessionId: string, team: 'a' | 'b') =>
      request<App.TennisMatch>('POST', `/sessions/${sessionId}/tennis/server/${team}`),
  },
  leaderboard: {
    get: (sessionId: string) =>
      request<App.Leaderboard>('GET', `/sessions/${sessionId}/leaderboard`),
  },
  contacts: {
    list: (token: string) =>
      request<App.Contact[]>('GET', '/contacts', undefined, token),
    add: (token: string, contactUserId: string) =>
      request<void>('POST', '/contacts', { contact_user_id: contactUserId }, token),
    remove: (token: string, contactUserId: string) =>
      request<void>('DELETE', `/contacts/${contactUserId}`, undefined, token),
    search: (token: string, q: string) =>
      request<App.UserSearchResult[]>('GET', `/users/search?q=${encodeURIComponent(q)}`, undefined, token),
  },
  push: {
    getVapidKey: () =>
      request<{ public_key: string }>('GET', '/push/vapid-public-key'),
    subscribe: (token: string, endpoint: string, p256dh: string, auth: string) =>
      request<void>('POST', '/push/subscribe', { endpoint, p256dh, auth }, token),
    unsubscribe: (token: string, endpoint: string) =>
      request<void>('DELETE', '/push/subscribe', { endpoint }, token),
  },
};
