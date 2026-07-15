import { describe, it, expect, beforeAll, vi } from 'vitest';
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
