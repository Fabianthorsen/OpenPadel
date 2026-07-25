import { describe, it, expect, beforeAll, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { init, register, waitLocale } from 'svelte-i18n';
import Lobby from './Lobby.svelte';

/**
 * Render tests for the "invite my whole club" fan-out on the Session invite
 * surface (#128).
 *
 * A signed-in admin who belongs to Clubs sees an "Invite all" row per Club above
 * the one-by-one search. Tapping it fans the roster out into ordinary Session
 * invites via api.invites.sendClub. Guests and clubless admins never see the row.
 */

const { sendClub, clubsList } = vi.hoisted(() => ({
	sendClub: vi.fn().mockResolvedValue([{ id: 'inv_1', to_user_id: 'u2' }]),
	clubsList: vi.fn().mockResolvedValue([
		{
			id: 'club_1',
			name: 'Bouvet Padel',
			avatar_icon: 'star',
			avatar_color: '#3d7a24',
			my_role: 'admin',
			roster_count: 5
		}
	])
}));

vi.mock('$lib/auth.svelte', () => ({
	auth: { token: 'admin-token', user: { id: 'u1', display_name: 'Alice' }, ready: true }
}));

vi.mock('$lib/api/client', async (importOriginal) => {
	const actual = await importOriginal<typeof import('$lib/api/client')>();
	return {
		ApiError: actual.ApiError,
		api: {
			invites: { listForSession: vi.fn().mockResolvedValue([]), sendClub },
			clubs: { list: clubsList },
			sessions: { update: vi.fn().mockResolvedValue({}), start: vi.fn(), cancel: vi.fn() },
			players: { join: vi.fn(), remove: vi.fn() },
			contacts: { search: vi.fn().mockResolvedValue([]) }
		}
	};
});

beforeAll(async () => {
	register('en', () => import('../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

beforeEach(() => {
	localStorage.clear();
	sendClub.mockClear();
	clubsList.mockClear();
});

function makeSession(overrides?: Partial<App.Session>): App.Session {
	return {
		id: 'ABCD',
		status: 'lobby',
		game_mode: 'americano',
		courts: 2,
		points: 24,
		rounds_total: undefined,
		current_round: 0,
		players: [],
		created_at: '2026-07-14T20:00:00Z',
		updated_at: '2026-07-14T20:00:00Z',
		...overrides
	};
}

function makeProps(session: App.Session, isAdmin: boolean) {
	return { session, isAdmin, onRefresh: () => {}, onStarted: () => {} };
}

describe('Lobby — invite a whole club (admin)', () => {
	it('renders an Invite all row per club with roster size', async () => {
		render(Lobby, makeProps(makeSession(), true));

		expect(await screen.findByText('Bouvet Padel')).toBeInTheDocument();
		expect(screen.getByText('5 members')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /invite all/i })).toBeInTheDocument();
	});

	it('fans the roster out via api.invites.sendClub when tapped', async () => {
		const user = userEvent.setup();
		render(Lobby, makeProps(makeSession(), true));

		const btn = await screen.findByRole('button', { name: /invite all/i });
		await user.click(btn);

		await waitFor(() => {
			expect(sendClub).toHaveBeenCalledWith('ABCD', 'club_1', 'admin-token');
		});
	});

	it('does not load or show clubs for a non-admin', async () => {
		render(Lobby, makeProps(makeSession({ creator_player_id: undefined }), false));

		// Give any (guarded) load a chance to run, then assert it did not.
		await Promise.resolve();
		expect(clubsList).not.toHaveBeenCalled();
		expect(screen.queryByRole('button', { name: /invite all/i })).not.toBeInTheDocument();
	});
});
