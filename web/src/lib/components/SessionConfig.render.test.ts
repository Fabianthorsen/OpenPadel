import { describe, it, expect, beforeAll, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
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

describe('SessionConfig rounds', () => {
	// The drawer no longer shows the Americano round *count* — that is derived by
	// the backend and previewed in the lobby header. The drawer only carries the
	// fixed/unlimited toggle for Americano, and a user-picked stepper for Mexicano.

	it('Americano fixed: no round-count stepper in the drawer', () => {
		const session = makeSession({
			game_mode: 'americano',
			courts: 1,
			rounds_total: 5,
			players: makePlayers(6)
		});
		render(SessionConfig, { session, sessionId: 'ABCD', open: true });

		expect(screen.getByRole('radio', { name: 'Fixed' })).toBeChecked();
		expect(screen.queryByRole('button', { name: /increase/i })).not.toBeInTheDocument();
		expect(screen.queryByRole('button', { name: /decrease/i })).not.toBeInTheDocument();
	});

	it('Americano unlimited: toggle reflects it, still no stepper', () => {
		const session = makeSession({
			game_mode: 'americano',
			courts: 1,
			rounds_total: undefined,
			players: makePlayers(6)
		});
		render(SessionConfig, { session, sessionId: 'ABCD', open: true });

		expect(screen.getByRole('radio', { name: 'Unlimited' })).toBeChecked();
		expect(screen.queryByRole('button', { name: /increase/i })).not.toBeInTheDocument();
	});

	it('Mexicano fixed: keeps a user-picked round stepper', () => {
		const session = makeSession({
			game_mode: 'mexicano',
			courts: 2,
			rounds_total: 7,
			players: makePlayers(8)
		});
		render(SessionConfig, { session, sessionId: 'ABCD', open: true });

		expect(screen.getByRole('button', { name: /increase/i })).toBeInTheDocument();
	});
});
