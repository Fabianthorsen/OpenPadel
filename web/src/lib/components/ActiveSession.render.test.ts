import { describe, it, expect, beforeAll } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { init, register, waitLocale } from 'svelte-i18n';
import type { SessionStream } from '$lib/stores/sessionStream.svelte';
import ActiveSession from './ActiveSession.svelte';

/**
 * Render test for the round-count header in the active session screen.
 *
 * This is the surface where the "unlimited Americano" bug showed itself: an
 * unlimited session (rounds_total = null) was being pinned to a fixed count at
 * start, so the header read "Round 1 of 5" instead of the open-ended "Round 1".
 * The backend fix keeps rounds_total null; this guards the frontend contract
 * that null renders as open-ended and a concrete total renders as "of N".
 *
 * The default tab is "scoring" whenever the current round has an unscored match,
 * so these renders never touch the SessionStream or the API (both are only used
 * by the standings tab / action handlers).
 */
beforeAll(async () => {
	register('en', () => import('../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

function makePlayers(): App.Player[] {
	return [1, 2, 3, 4].map((n) => ({
		id: `p${n}`,
		session_id: 's1',
		name: `Player ${n}`,
		avatar_icon: 'racket',
		avatar_color: '#3d7a24',
		active: true,
		joined_at: '2026-07-13T20:00:00Z'
	}));
}

function makeProps(roundsTotal?: number) {
	const session: App.Session = {
		id: 's1',
		status: 'playing',
		game_mode: 'americano',
		courts: 1,
		points: 24,
		rounds_total: roundsTotal,
		current_round: 1,
		players: makePlayers(),
		created_at: '2026-07-13T20:00:00Z',
		updated_at: '2026-07-13T20:00:00Z'
	};
	const currentRound: App.Round = {
		id: 'r1',
		session_id: 's1',
		number: 1,
		bench: [],
		matches: [
			{
				id: 'm1',
				round_id: 'r1',
				court: 1,
				team_a: ['p1', 'p2'],
				team_b: ['p3', 'p4'],
				score: null
			}
		]
	};
	return {
		session,
		currentRound,
		isAdmin: false,
		onRefresh: () => {},
		stream: {} as unknown as SessionStream
	};
}

describe('ActiveSession round header', () => {
	it('unlimited session (rounds_total null) shows open-ended "Round N", not "of N"', () => {
		render(ActiveSession, makeProps(undefined));

		expect(screen.getByRole('heading', { name: 'Round 1' })).toBeInTheDocument();
		expect(screen.queryByText(/Round 1 of/)).not.toBeInTheDocument();
	});

	it('fixed session (rounds_total set) shows "Round N of M"', () => {
		render(ActiveSession, makeProps(5));

		expect(screen.getByRole('heading', { name: 'Round 1 of 5' })).toBeInTheDocument();
	});
});
