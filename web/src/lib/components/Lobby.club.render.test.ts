import { describe, it, expect, beforeAll, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { init, register, waitLocale } from 'svelte-i18n';
import Lobby from './Lobby.svelte';

/**
 * Render tests for club-event framing on the Session surface (#127).
 *
 * A club event is a normal Session that happens to be owned by a Club. The two
 * places that must make that ownership visible — so a member arriving from the
 * Club home doesn't just see "a regular game" — are:
 *  - the lobby header (admin / already-joined view): a club badge naming the Club;
 *  - the guest join screen (not admin, not joined): a club-framed title instead
 *    of the personal "invited you to their tournament" copy.
 *
 * Both read from `session.club_name`, which the API attaches only to club events.
 */

vi.mock('$lib/api/client', async (importOriginal) => {
	const actual = await importOriginal<typeof import('$lib/api/client')>();
	return {
		ApiError: actual.ApiError,
		api: {
			invites: { listForSession: vi.fn().mockResolvedValue([]) },
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
});

function makePlayers(): App.Player[] {
	return Array.from({ length: 4 }, (_, i) => ({
		id: `p${i + 1}`,
		session_id: 'ABCD',
		name: `Player ${i + 1}`,
		avatar_icon: 'racket',
		avatar_color: '#3d7a24',
		rating: 3,
		added_by_admin: false,
		active: true,
		joined_at: '2026-07-14T20:00:00Z'
	}));
}

function makeSession(overrides?: Partial<App.Session>): App.Session {
	return {
		id: 'ABCD',
		status: 'lobby',
		game_mode: 'americano',
		courts: 2,
		points: 24,
		rounds_total: undefined,
		current_round: 0,
		players: makePlayers(),
		created_at: '2026-07-14T20:00:00Z',
		updated_at: '2026-07-14T20:00:00Z',
		...overrides
	};
}

function makeProps(session: App.Session, isAdmin: boolean) {
	return { session, isAdmin, onRefresh: () => {}, onStarted: () => {} };
}

describe('Lobby — club badge (admin / joined view)', () => {
	it('shows the club name badge on the lobby header for a club event', () => {
		render(
			Lobby,
			makeProps(makeSession({ name: 'Friday Padel', club_name: 'Bouvet Padel' }), true)
		);

		expect(screen.getByText('Bouvet Padel')).toBeInTheDocument();
	});

	it('shows the badge for a joined (non-admin) member too', () => {
		// Joined-as-guest is derived from a stored player_id matching an active player.
		localStorage.setItem('player_id_ABCD', 'p1');
		render(
			Lobby,
			makeProps(makeSession({ name: 'Friday Padel', club_name: 'Bouvet Padel' }), false)
		);

		expect(screen.getByText('Bouvet Padel')).toBeInTheDocument();
	});

	it('renders no club badge for an ordinary (non-club) session', () => {
		render(Lobby, makeProps(makeSession({ name: 'Friday Padel' }), true));

		expect(screen.queryByText('Bouvet Padel')).not.toBeInTheDocument();
	});
});

describe('Lobby — club-framed join screen (guest, not joined)', () => {
	// creator_player_id points at an active player, so `creatorName` resolves to
	// "Player 1"; no stored player_id and isAdmin=false → the join screen renders.
	const withCreator = (over?: Partial<App.Session>) =>
		makeSession({ creator_player_id: 'p1', ...over });

	it('frames the title by the Club and the scheduling member', () => {
		render(Lobby, makeProps(withCreator({ club_name: 'Bouvet Padel' }), false));

		// invite_title_club_named: "{creator} scheduled a {club} game"
		expect(
			screen.getByRole('heading', { name: /player 1 scheduled a bouvet padel game/i })
		).toBeInTheDocument();
		// The personal-invite copy must not be used for a club event.
		expect(screen.queryByText(/invited you to/i)).not.toBeInTheDocument();
	});

	it('falls back to a club-only title when the creator is unknown', () => {
		render(
			Lobby,
			makeProps(makeSession({ club_name: 'Bouvet Padel', creator_player_id: undefined }), false)
		);

		// invite_title_club: "Join the {club} game"
		expect(
			screen.getByRole('heading', { name: /join the bouvet padel game/i })
		).toBeInTheDocument();
	});

	it('keeps the personal-invite copy for an ordinary session (club branch does not hijack)', () => {
		render(Lobby, makeProps(withCreator({ name: 'Friday Padel' }), false));

		// invite_title_with_creator_named: "{creator} invited you to {name}"
		expect(
			screen.getByRole('heading', { name: /player 1 invited you to friday padel/i })
		).toBeInTheDocument();
	});
});
