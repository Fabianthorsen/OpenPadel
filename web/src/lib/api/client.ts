const BASE = '/api';

async function request<T>(method: string, path: string, body?: unknown, token?: string): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  const data = await res.json();
  if (!res.ok) throw new Error(data.error ?? 'request failed');
  return data as T;
}

export const api = {
  sessions: {
    create: (courts: number, points: number) =>
      request<App.Session>('POST', '/sessions', { courts, points }),
    get: (id: string, token?: string) =>
      request<App.Session>('GET', `/sessions/${id}`, undefined, token),
    start: (id: string, token: string) =>
      request<App.Session>('POST', `/sessions/${id}/start`, undefined, token),
  },
  players: {
    join: (sessionId: string, name: string) =>
      request<App.Player>('POST', `/sessions/${sessionId}/players`, { name }),
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
