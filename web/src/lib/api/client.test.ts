import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ApiError, api } from './client.js';

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn());
});

afterEach(() => {
  vi.unstubAllGlobals();
});

function mockFetch(status: number, body: unknown, ok = status >= 200 && status < 300) {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      status,
      ok,
      json: () => Promise.resolve(body),
    })
  );
}

describe('ApiError', () => {
  it('stores status and message', () => {
    const err = new ApiError('not_found', 404);
    expect(err.status).toBe(404);
    expect(err.message).toBe('not_found');
    expect(err).toBeInstanceOf(Error);
  });
});

describe('api.auth.register', () => {
  it('sends POST with correct body', async () => {
    mockFetch(201, { token: 'abc', user: { id: 'u1', email: 'a@b.com' } });

    await api.auth.register('a@b.com', 'Alice', 'password123');

    expect(fetch).toHaveBeenCalledWith(
      '/api/auth/register',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ email: 'a@b.com', display_name: 'Alice', password: 'password123' }),
      })
    );
  });

  it('throws ApiError on 4xx response', async () => {
    mockFetch(409, { error: 'email_already_registered' }, false);

    await expect(api.auth.register('a@b.com', 'Alice', 'password123')).rejects.toBeInstanceOf(
      ApiError
    );
  });

  it('throws ApiError with correct status on error', async () => {
    mockFetch(409, { error: 'email_already_registered' }, false);

    try {
      await api.auth.register('a@b.com', 'Alice', 'password123');
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      expect((e as ApiError).status).toBe(409);
      expect((e as ApiError).message).toBe('email_already_registered');
    }
  });
});

describe('Authorization header', () => {
  it('sets bearer token on authenticated requests', async () => {
    mockFetch(200, { id: 'u1', email: 'a@b.com' });

    await api.auth.me('my-token');

    expect(fetch).toHaveBeenCalledWith(
      '/api/auth/me',
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: 'Bearer my-token',
        }),
      })
    );
  });

  it('omits Authorization header when no token', async () => {
    mockFetch(201, { token: 'abc', user: {} });

    await api.auth.register('a@b.com', 'Alice', 'pw12345678');

    const call = (fetch as ReturnType<typeof vi.fn>).mock.calls[0];
    const headers = call[1]?.headers ?? {};
    expect(headers.Authorization).toBeUndefined();
  });
});

describe('204 No Content', () => {
  it('returns undefined for 204 responses', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        status: 204,
        ok: true,
        json: () => Promise.resolve(null),
      })
    );

    const result = await api.auth.logout('token');
    expect(result).toBeUndefined();
  });
});

describe('api.sessions.get', () => {
  it('passes admin token as Authorization header', async () => {
    mockFetch(200, { id: 's1', status: 'lobby' });

    await api.sessions.get('s1', 'admin-token');

    expect(fetch).toHaveBeenCalledWith(
      '/api/sessions/s1',
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: 'Bearer admin-token',
        }),
      })
    );
  });
});

describe('api.sessions.create - Timed Americano', () => {
  it('includes totalDurationMinutes and bufferSeconds in POST body for timed_americano', async () => {
    mockFetch(201, { id: 's1', status: 'lobby', game_mode: 'timed_americano' });

    await api.sessions.create({
      game_mode: 'timed_americano',
      courts: 2,
      name: 'Test Timed',
      total_duration_minutes: 120,
      points: 0,
    });

    expect(fetch).toHaveBeenCalled();
    const call = (fetch as ReturnType<typeof vi.fn>).mock.calls[0];
    const body = JSON.parse(call[1].body);

    expect(body).toMatchObject({
      game_mode: 'timed_americano',
      courts: 2,
      name: 'Test Timed',
      total_duration_minutes: 120,
      points: 0,
    });
  });

  it('includes interval_between_rounds_minutes in POST body when provided', async () => {
    mockFetch(201, { id: 's1', status: 'lobby', game_mode: 'timed_americano' });

    await api.sessions.create({
      game_mode: 'timed_americano',
      courts: 2,
      name: 'Test Timed',
      total_duration_minutes: 120,
      points: 0,
      interval_between_rounds_minutes: 4,
    });

    expect(fetch).toHaveBeenCalled();
    const call = (fetch as ReturnType<typeof vi.fn>).mock.calls[0];
    const body = JSON.parse(call[1].body);

    expect(body).toMatchObject({
      interval_between_rounds_minutes: 4,
    });
  });

  it('omits interval_between_rounds_minutes when not provided', async () => {
    mockFetch(201, { id: 's1', status: 'lobby', game_mode: 'timed_americano' });

    await api.sessions.create({
      game_mode: 'timed_americano',
      courts: 2,
      name: 'Test Timed',
      total_duration_minutes: 120,
      points: 0,
    });

    expect(fetch).toHaveBeenCalled();
    const call = (fetch as ReturnType<typeof vi.fn>).mock.calls[0];
    const body = JSON.parse(call[1].body);

    expect(body.interval_between_rounds_minutes).toBeUndefined();
  });
});
