import { describe, it, expect, beforeAll, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { init, register, waitLocale } from 'svelte-i18n';
import SessionConfig from './SessionConfig.svelte';

/**
 * Render tests for the shared SessionConfig drawer (#182).
 *
 * Points are a fixed set of presets (16 / 24 / 32) shown as pills — NOT a
 * free-range +/- stepper. The extraction briefly turned them into a stepper;
 * this locks the preset contract so it can't silently regress again.
 */

vi.mock('$lib/api/client', async (importOriginal) => {
	const actual = await importOriginal<typeof import('$lib/api/client')>();
	return {
		ApiError: actual.ApiError,
		api: { sessions: { update: vi.fn().mockResolvedValue({}) } }
	};
});

beforeAll(async () => {
	register('en', () => import('../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

afterEach(() => {
	// Bits UI's scroll-lock leaves styles on <body> when a Drawer unmounts; clear
	// them so they don't leak into the next test's document (shared across files).
	document.body.style.pointerEvents = '';
	document.body.style.removeProperty('--scrollbar-width');
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

// `active` players plus `inactive` (removed but soft-deactivated) records.
function makePlayers(active: number, inactive = 0): App.Player[] {
	const mk = (i: number, isActive: boolean): App.Player => ({
		id: `p${i}`,
		session_id: 'ABCD',
		name: `Player ${i}`,
		avatar_icon: 'racket',
		avatar_color: '#3d7a24',
		active: isActive,
		joined_at: '2026-07-14T20:00:00Z'
	});
	const arr: App.Player[] = [];
	for (let i = 0; i < active; i++) arr.push(mk(i + 1, true));
	for (let i = 0; i < inactive; i++) arr.push(mk(active + i + 1, false));
	return arr;
}

describe('SessionConfig points', () => {
	it('offers exactly the 16 / 24 / 32 presets as pills', () => {
		render(SessionConfig, { session: makeSession(), sessionId: 'ABCD', open: true });

		expect(screen.getByText('16')).toBeInTheDocument();
		expect(screen.getByText('24')).toBeInTheDocument();
		expect(screen.getByText('32')).toBeInTheDocument();
	});

	it('does not offer off-list point values', () => {
		render(SessionConfig, { session: makeSession(), sessionId: 'ABCD', open: true });

		expect(screen.queryByText('20')).not.toBeInTheDocument();
		expect(screen.queryByText('25')).not.toBeInTheDocument();
	});
});

describe('SessionConfig rounds (Americano is backend-derived, display-only)', () => {
	it('shows the fair round count for the roster, read-only (no stepper)', () => {
		// 6 active players on 1 court -> fair count 6. rounds_total (5) is only the
		// non-null "limited" signal; the number shown is computed, not that stored value.
		const session = makeSession({
			game_mode: 'americano',
			courts: 1,
			rounds_total: 5,
			players: makePlayers(6)
		});
		render(SessionConfig, { session, sessionId: 'ABCD', open: true });

		expect(screen.getByText('6')).toBeInTheDocument();
		// Read-only: Americano's count is not editable, so there is no +/- stepper.
		expect(screen.queryByRole('button', { name: /increase/i })).not.toBeInTheDocument();
		expect(screen.queryByRole('button', { name: /decrease/i })).not.toBeInTheDocument();
	});

	it('counts only active players — removed (soft-inactive) records are excluded', () => {
		// 6 active + 4 removed on 1 court. The fair count is 6 (from active players),
		// NOT 10 (from all records). This is the exact regression: it used to show 10.
		const session = makeSession({
			game_mode: 'americano',
			courts: 1,
			rounds_total: 5,
			players: makePlayers(6, 4)
		});
		render(SessionConfig, { session, sessionId: 'ABCD', open: true });

		expect(screen.getByText('6')).toBeInTheDocument();
		expect(screen.queryByText('10')).not.toBeInTheDocument();
	});

	it('shows unlimited (no fixed count, no stepper) when rounds_total is null', () => {
		const session = makeSession({
			game_mode: 'americano',
			courts: 1,
			rounds_total: undefined,
			players: makePlayers(6)
		});
		render(SessionConfig, { session, sessionId: 'ABCD', open: true });

		expect(screen.getByRole('radio', { name: 'Unlimited' })).toBeChecked();
		expect(screen.queryByText('6')).not.toBeInTheDocument();
		expect(screen.queryByRole('button', { name: /increase/i })).not.toBeInTheDocument();
	});

	it('Mexicano keeps a user-picked round stepper', () => {
		const session = makeSession({
			game_mode: 'mexicano',
			courts: 2,
			rounds_total: 7,
			players: makePlayers(8)
		});
		render(SessionConfig, { session, sessionId: 'ABCD', open: true });

		expect(screen.getByRole('button', { name: /increase/i })).toBeInTheDocument();
	});

	it('recomputes when a player is removed — soft-delete keeps the record, count follows active', async () => {
		// 10 active players on 2 courts -> 10 rounds.
		const props = {
			session: makeSession({
				game_mode: 'americano' as const,
				courts: 2,
				rounds_total: 5,
				players: makePlayers(10)
			}),
			sessionId: 'ABCD',
			open: true
		};
		const { rerender } = render(SessionConfig, props);
		expect(screen.getByText('10')).toBeInTheDocument();

		// Remove one player. Removal is a SOFT delete: the player stays in
		// session.players with active=false, so the array is still length 10 —
		// 9 active + 1 inactive. This is the exact regression: counting
		// session.players.length stayed at 10 and never updated; counting active
		// players drops to 9 rounds.
		await rerender({
			...props,
			session: makeSession({
				game_mode: 'americano',
				courts: 2,
				rounds_total: 5,
				players: makePlayers(9, 1)
			})
		});

		expect(screen.getByText('9')).toBeInTheDocument();
		expect(screen.queryByText('10')).not.toBeInTheDocument();
	});

	it('recomputes when a setting is changed inside the drawer (court count)', async () => {
		const user = userEvent.setup({ pointerEventsCheck: 0 });
		const session = makeSession({
			game_mode: 'americano',
			courts: 2,
			rounds_total: 5,
			players: makePlayers(8)
		});
		render(SessionConfig, { session, sessionId: 'ABCD', open: true });

		// 8 players on 2 courts -> 7 rounds.
		expect(screen.getByText('7')).toBeInTheDocument();

		// Change to 1 court in the drawer -> 8 players on 1 court -> 8 rounds.
		await user.click(screen.getByRole('radio', { name: '1' }));

		expect(screen.getByText('8')).toBeInTheDocument();
		expect(screen.queryByText('7')).not.toBeInTheDocument();
	});
});
