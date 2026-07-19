import { describe, it, expect, beforeAll } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { init, register, waitLocale } from 'svelte-i18n';
import StatLegend from './stat-legend.svelte';
import { MODE_METRICS, type StatMetric } from './catalog';

/**
 * StatLegend explains each metric from the same catalog that renders the tiles,
 * so its rows track MODE_METRICS one-to-one and each row carries both the label
 * and its description — a new metric's explanation appears here for free.
 */

beforeAll(async () => {
	register('en', () => import('../../../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

describe('StatLegend (render)', () => {
	it('renders one row per catalog metric with its label and description', () => {
		render(StatLegend, { metrics: MODE_METRICS });
		expect(screen.getAllByRole('term')).toHaveLength(MODE_METRICS.length);
		expect(screen.getByText('Point win %')).toBeInTheDocument();
		expect(screen.getByText(/average share of points won per match/i)).toBeInTheDocument();
	});

	it('drives entirely off the passed metric array', () => {
		const one: StatMetric[] = [
			{
				id: 'net_points',
				label: 'stats_net_points',
				description: 'stats_net_points_desc',
				format: 'signed',
				accessor: (s) => s.net_points
			}
		];
		render(StatLegend, { metrics: one });
		expect(screen.getAllByRole('term')).toHaveLength(1);
		expect(screen.getByText('Net points')).toBeInTheDocument();
	});
});
