import { describe, it, expect, beforeAll } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { init, register, waitLocale } from 'svelte-i18n';
import ModeSection from './mode-section.svelte';
import { MODE_METRICS } from './catalog';

/**
 * ModeSection is one Game-Mode block: a header over a StatGrid, or a gentle
 * "no games yet" state when the mode has zero games (ADR 0007). These lock in
 * both branches — the empty state must replace the grid, not sit beside it.
 */

beforeAll(async () => {
	register('en', () => import('../../../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

const played: App.ModeStats = {
	mode: 'americano',
	games: 4,
	wins: 3,
	draws: 0,
	losses: 1,
	total_points: 72,
	points_conceded: 60,
	net_points: 12,
	point_win_pct: 54,
	tournaments: 2
};

const empty: App.ModeStats = {
	mode: 'mexicano',
	games: 0,
	wins: 0,
	draws: 0,
	losses: 0,
	total_points: 0,
	points_conceded: 0,
	net_points: 0,
	point_win_pct: 0,
	tournaments: 0
};

describe('ModeSection (render)', () => {
	it('renders the title and a stat grid when the mode has games', () => {
		render(ModeSection, { title: 'Americano', stats: played, metrics: MODE_METRICS });
		expect(screen.getByText('Americano')).toBeInTheDocument();
		expect(document.querySelector('[data-slot="stat-grid"]')).toBeInTheDocument();
		expect(document.querySelector('[data-slot="mode-empty"]')).not.toBeInTheDocument();
	});

	it('renders the empty state and no grid when the mode has no games', () => {
		render(ModeSection, { title: 'Mexicano', stats: empty, metrics: MODE_METRICS });
		expect(screen.getByText('Mexicano')).toBeInTheDocument();
		expect(document.querySelector('[data-slot="mode-empty"]')).toBeInTheDocument();
		expect(document.querySelector('[data-slot="stat-grid"]')).not.toBeInTheDocument();
	});
});
