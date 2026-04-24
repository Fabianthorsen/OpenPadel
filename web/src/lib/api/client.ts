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
      request<{ user: App.User; stats: App.AmericanoCareerStats }>('GET', '/auth/profile', undefined, token),
    updateProfile: (token: string, displayName: string, avatarIcon: string, avatarColor: string) =>
      request<App.User>('PUT', '/auth/profile', { display_name: displayName, avatar_icon: avatarIcon, avatar_color: avatarColor }, token),
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
    create: (params: {
      courts: number;
      points: number;
      name: string;
      game_mode: string;
      scheduled_at?: string;
      rounds_total?: number;
      court_duration_minutes?: number;
      total_duration_minutes?: number;
      interval_between_rounds_minutes?: number;
    }) => {
      const body: Record<string, unknown> = {
        courts: params.courts,
        points: params.points,
        name: params.name,
        game_mode: params.game_mode,
      };
      if (params.scheduled_at) body.scheduled_at = params.scheduled_at;
      if (params.rounds_total) body.rounds_total = params.rounds_total;
      if (params.court_duration_minutes) body.court_duration_minutes = params.court_duration_minutes;
      if (params.total_duration_minutes) body.total_duration_minutes = params.total_duration_minutes;
      if (params.interval_between_rounds_minutes) body.interval_between_rounds_minutes = params.interval_between_rounds_minutes;
      return request<App.Session>('POST', '/sessions', body);
    },
    get: (id: string, token?: string) =>
      request<App.Session>('GET', `/sessions/${id}`, undefined, token),
    update: (id: string, patch: {
      name?: string;
      game_mode?: string;
      courts?: number;
      points?: number;
      rounds_total?: number | null;
      scheduled_at?: string;
    }, adminToken: string) =>
      request<App.Session>('PATCH', `/sessions/${id}`, patch, adminToken),
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
  leaderboard: {
    get: (sessionId: string) =>
      request<App.Leaderboard>('GET', `/sessions/${sessionId}/leaderboard`),
  },
  invites: {
    list: (token: string) =>
      request<App.Invite[]>('GET', '/invites', undefined, token),
    listForSession: (sessionId: string) =>
      request<App.Invite[]>('GET', `/sessions/${sessionId}/invites`),
    send: (sessionId: string, toUserId: string, token: string) =>
      request<App.Invite>('POST', `/sessions/${sessionId}/invites`, { to_user_id: toUserId }, token),
    accept: (inviteId: string, token: string) =>
      request<App.Player>('POST', `/invites/${inviteId}/accept`, undefined, token),
    decline: (inviteId: string, token: string) =>
      request<void>('POST', `/invites/${inviteId}/decline`, undefined, token),
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
