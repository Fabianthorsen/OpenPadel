import { describe, it, expect, beforeAll } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { init, register, waitLocale } from 'svelte-i18n';
import FormStrip from './form-strip.svelte';

/**
 * FormStrip renders cross-mode recent form from the per-Match results series
 * (ADR 0007): a Point Win % sparkline plus the win/draw/loss record over the
 * window. It must render nothing for an empty series and stay honest on a single
 * match (no jarring 0%/100%).
 */

beforeAll(async () => {
	register('en', () => import('../../../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

function match(points: number, conceded: number): App.MatchResult {
	const result = points > conceded ? 'win' : points < conceded ? 'loss' : 'draw';
	return {
		match_id: `m-${points}-${conceded}`,
		mode: 'americano',
		date: '2026-01-01T00:00:00Z',
		points,
		conceded,
		result
	};
}

describe('FormStrip (render)', () => {
	it('renders the sparkline with the win/draw/loss record', () => {
		const series = [match(16, 8), match(8, 16), match(15, 9)]; // W, L, W
		render(FormStrip, { series });

		const strip = document.querySelector('[data-slot="form-strip"]');
		expect(strip).toBeInTheDocument();
		// One marker per match, plus the midline label and the trend path.
		expect(strip?.querySelectorAll('circle')).toHaveLength(series.length);
		expect(strip?.querySelector('path')).toBeInTheDocument();
		expect(screen.getByText('50%')).toBeInTheDocument();
		// Record: 2 wins, 0 draws, 1 loss over the window, each count inline-labelled.
		expect(strip?.textContent).toContain('W (2)');
		expect(strip?.textContent).toContain('D (0)');
		expect(strip?.textContent).toContain('L (1)');
	});

	it('renders nothing for an empty series', () => {
		render(FormStrip, { series: [] });
		expect(document.querySelector('[data-slot="form-strip"]')).not.toBeInTheDocument();
	});

	it('stays honest on a single match (its real share, not 0%/100%)', () => {
		render(FormStrip, { series: [match(15, 9)] }); // a single win
		const strip = document.querySelector('[data-slot="form-strip"]');
		expect(strip).toBeInTheDocument();
		// One win, no draws or losses — and the dot's own share in its tooltip.
		expect(strip?.textContent).toContain('W (1)');
		expect(strip?.textContent).toContain('63%');
	});
});
