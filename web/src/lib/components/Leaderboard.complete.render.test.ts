import { describe, it, expect, beforeAll, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { init, register, waitLocale } from 'svelte-i18n';
import Leaderboard from './Leaderboard.svelte';

// The final standings load via api.leaderboard.get on mount; stub it so the
// podium + ranking rows render without reaching the network. complete=true also
// calls api.contacts.list (guarded by auth.token, which is unset here).
const getLeaderboard = vi.fn();
vi.mock('$lib/api/client', () => ({
	api: {
		leaderboard: { get: (...args: unknown[]) => getLeaderboard(...args) },
		contacts: { list: vi.fn().mockResolvedValue([]) }
	}
}));

beforeAll(async () => {
	register('en', () => import('../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

function makeLeaderboard(): App.Leaderboard {
	return {
		session_id: 's1',
		status: 'done',
		updated_at: '2026-07-14T20:00:00Z',
		current_round: null,
		total_rounds: null,
		standings: [
			['p1', 'Alice', 1],
			['p2', 'Bob', 2],
			['p3', 'Charlie', 3],
			['p4', 'Dana', 4]
		].map(([player_id, name, rank]) => ({
			rank: rank as number,
			player_id: player_id as string,
			name: name as string,
			points: 30 - (rank as number),
			games_played: 6,
			wins: 3,
			draws: 1,
			avatar_icon: 'racket',
			avatar_color: '#3d7a24'
		}))
	};
}

describe('Leaderboard (complete)', () => {
	it('shows Rating badges on the podium (top 3) and the ranking rows (4th+)', async () => {
		getLeaderboard.mockResolvedValue(makeLeaderboard());
		render(Leaderboard, {
			props: {
				sessionId: 's1',
				complete: true,
				ratings: { p1: 5, p2: 4, p3: 3, p4: 1 }
			}
		});

		// Wait for the async load to populate the podium.
		await screen.findByText('Alice');
		const badges = screen.getAllByLabelText('Rating');
		// One badge per standing: 3 on the podium + 1 in the ranking list.
		expect(badges).toHaveLength(4);
		expect(new Set(badges.map((b) => b.textContent?.trim()))).toEqual(
			new Set(['5', '4', '3', '1'])
		);
	});

	it('omits the badges when no ratings are supplied', async () => {
		getLeaderboard.mockResolvedValue(makeLeaderboard());
		render(Leaderboard, {
			props: {
				sessionId: 's1',
				complete: true
			}
		});

		await screen.findByText('Alice');
		expect(screen.queryByLabelText('Rating')).not.toBeInTheDocument();
	});
});
