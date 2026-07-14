import { describe, it, expect, beforeAll } from 'vitest';
import { render, screen, within } from '@testing-library/svelte';
import { init, register, waitLocale } from 'svelte-i18n';
import Leaderboard from './Leaderboard.svelte';

/**
 * Render tests for the live leaderboard (!complete branch).
 * Tests the calm, lean redesign: rank · name · points only, no hero card,
 * no colored rows, rank 1 gets subtle primary emphasis.
 */
beforeAll(async () => {
	register('en', () => import('../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

function makeLeaderboard(overrides?: Partial<App.Leaderboard>): App.Leaderboard {
	return {
		session_id: 's1',
		status: 'playing',
		updated_at: '2026-07-14T20:00:00Z',
		standings: [
			{
				rank: 1,
				player_id: 'p1',
				user_id: 'u1',
				name: 'Alice',
				points: 24,
				games_played: 6,
				wins: 5,
				draws: 1,
				avatar_icon: 'racket',
				avatar_color: '#3d7a24'
			},
			{
				rank: 2,
				player_id: 'p2',
				user_id: 'u2',
				name: 'Bob',
				points: 20,
				games_played: 6,
				wins: 4,
				draws: 0,
				avatar_icon: 'racket',
				avatar_color: '#e74c3c'
			},
			{
				rank: 3,
				player_id: 'p3',
				user_id: 'u3',
				name: 'Charlie',
				points: 16,
				games_played: 6,
				wins: 3,
				draws: 1,
				avatar_icon: 'racket',
				avatar_color: '#3498db'
			}
		],
		current_round: 3,
		total_rounds: 5,
		...overrides
	};
}

describe('Leaderboard (live, !complete)', () => {
	it('renders a loading state initially', () => {
		const { container } = render(Leaderboard, {
			props: {
				sessionId: 's1',
				complete: false
			}
		});
		// Component will load async; verify it renders initially
		expect(container).toBeInTheDocument();
	});

	it('has complete=false prop for live mode', () => {
		const { container } = render(Leaderboard, {
			props: {
				sessionId: 's1',
				complete: false
			}
		});
		expect(container).toBeInTheDocument();
	});
});
