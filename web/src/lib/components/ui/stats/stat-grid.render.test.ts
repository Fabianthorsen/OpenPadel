import { describe, it, expect, beforeAll } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { init, register, waitLocale } from 'svelte-i18n';
import StatGrid from './stat-grid.svelte';
import { MODE_METRICS, type StatMetric } from './catalog';

/**
 * StatGrid is the catalog-driven layer: it translates each metric's label and
 * formats each value, so these tests assert the formats (record / percent /
 * signed / count) come out right and that the grid renders one tile per metric.
 */

beforeAll(async () => {
	register('en', () => import('../../../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

const stats: App.ModeStats = {
	mode: 'americano',
	games: 8,
	wins: 5,
	draws: 1,
	losses: 2,
	total_points: 140,
	points_conceded: 122,
	net_points: 18,
	point_win_pct: 55.6,
	tournaments: 3
};

describe('StatGrid (render)', () => {
	it('renders one tile per catalog metric', () => {
		render(StatGrid, { metrics: MODE_METRICS, stats });
		expect(document.querySelectorAll('[data-slot="stat-tile"]')).toHaveLength(MODE_METRICS.length);
	});

	it('formats each metric per its declared format', () => {
		render(StatGrid, { metrics: MODE_METRICS, stats });
		expect(screen.getByText('8')).toBeInTheDocument(); // count: games
		expect(screen.getByText('5–1–2')).toBeInTheDocument(); // record: W–D–L
		expect(screen.getByText('56%')).toBeInTheDocument(); // percent: rounded
		expect(screen.getByText('+18')).toBeInTheDocument(); // signed: positive gets +
	});

	it('shows a negative differential without a plus sign', () => {
		render(StatGrid, { metrics: MODE_METRICS, stats: { ...stats, net_points: -12 } });
		expect(screen.getByText('-12')).toBeInTheDocument();
	});

	it('renders a single custom metric — adding a stat is one catalog entry', () => {
		const one: StatMetric[] = [
			{
				id: 'games',
				label: 'stats_games',
				description: 'stats_games_desc',
				format: 'count',
				accessor: (s) => s.games
			}
		];
		render(StatGrid, { metrics: one, stats });
		expect(document.querySelectorAll('[data-slot="stat-tile"]')).toHaveLength(1);
	});
});
