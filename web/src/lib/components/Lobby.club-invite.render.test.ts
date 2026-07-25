import { describe, it, expect, beforeAll, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { init, register, waitLocale } from 'svelte-i18n';
import Lobby from './Lobby.svelte';

/**
 * Render tests for the club-event invite surface (#128).
 *
 * On a club event, an admin invites from the owning Club's roster — a row per
 * invitable member — OR taps "Invite all of <Club>" to fan the whole roster out
 * via api.invites.sendClub. Members already joined/invited, and the admin
 * themselves, are filtered out. An ordinary (non-club) session shows none of this.
 */

const { sendClub, sendInvite, clubDetail } = vi.hoisted(() => ({
	sendClub: vi.fn().mockResolvedValue([{ id: 'inv_1', to_user_id: 'u2' }]),
	sendInvite: vi.fn().mockResolvedValue({ id: 'inv_2', to_user_id: 'u2' }),
	clubDetail: vi.fn().mockResolvedValue({
		club: { id: 'club_1', name: 'Bouvet Padel' },
		members: [
			{
				user_id: 'u1',
				display_name: 'Alice',
				role: 'admin',
				avatar_icon: 'Star',
				avatar_color: 'forest'
			},
			{
				user_id: 'u2',
				display_name: 'Bob',
				role: 'member',
				avatar_icon: 'Cat',
				avatar_color: 'forest'
			},
			{
				user_id: 'u3',
				display_name: 'Carol',
				role: 'member',
				avatar_icon: 'Dog',
				avatar_color: 'forest'
			}
		],
		is_admin: true,
		my_role: 'admin',
		roster_count: 3
	})
}));

vi.mock('$lib/auth.svelte', () => ({
	auth: { token: 'admin-token', user: { id: 'u1', display_name: 'Alice' }, ready: true }
}));

vi.mock('$lib/api/client', async (importOriginal) => {
	const actual = await importOriginal<typeof import('$lib/api/client')>();
	return {
		ApiError: actual.ApiError,
		api: {
			invites: {
				listForSession: vi.fn().mockResolvedValue([]),
				send: sendInvite,
				sendClub
			},
			clubs: { detail: clubDetail },
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
	sendInvite.mockClear();
	clubDetail.mockClear();
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

const clubEvent = (over?: Partial<App.Session>) =>
	makeSession({ club_id: 'club_1', club_name: 'Bouvet Padel', ...over });

function makeProps(session: App.Session, isAdmin: boolean) {
	return { session, isAdmin, onRefresh: () => {}, onStarted: () => {} };
}

describe('Lobby — club-event invite surface (admin)', () => {
	it('lists invitable club members (dropping the admin themselves)', async () => {
		render(Lobby, makeProps(clubEvent(), true));

		// Bob and Carol are invitable; Alice (the signed-in admin) is filtered out.
		expect(await screen.findByText('Bob')).toBeInTheDocument();
		expect(screen.getByText('Carol')).toBeInTheDocument();
		expect(screen.queryByText('Alice')).not.toBeInTheDocument();
	});

	it('invites a single member via api.invites.send', async () => {
		const user = userEvent.setup();
		render(Lobby, makeProps(clubEvent(), true));

		await screen.findByText('Bob');
		// The Invite button in Bob's row.
		const bobRow = screen.getByText('Bob').closest('div')!;
		await user.click(bobRow.querySelector('button')!);

		await waitFor(() => expect(sendInvite).toHaveBeenCalledWith('ABCD', 'u2', 'admin-token'));
	});

	it('fans the whole roster out via api.invites.sendClub', async () => {
		const user = userEvent.setup();
		render(Lobby, makeProps(clubEvent(), true));

		const btn = await screen.findByRole('button', { name: /invite all of bouvet padel/i });
		await user.click(btn);

		await waitFor(() => expect(sendClub).toHaveBeenCalledWith('ABCD', 'club_1', 'admin-token'));
	});

	it('shows none of the club invite surface for an ordinary session', async () => {
		render(Lobby, makeProps(makeSession(), true));

		await Promise.resolve();
		expect(clubDetail).not.toHaveBeenCalled();
		expect(screen.queryByRole('button', { name: /invite all of/i })).not.toBeInTheDocument();
	});
});
