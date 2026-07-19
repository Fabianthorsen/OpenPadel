import { describe, it, expect } from 'vitest';
import { matchPointWinPct, recentRecord, formStrip, FORM_WINDOW } from './form';

// Builds a MatchResult with just the fields the derivations read; the outcome is
// inferred from points vs conceded so tests read like real matches.
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

describe('matchPointWinPct', () => {
	it('is the point share as a percentage', () => {
		expect(matchPointWinPct(match(16, 8))).toBeCloseTo(66.67, 1);
	});

	it('degrades a 0–0 match to a neutral 50 rather than dividing by zero', () => {
		expect(matchPointWinPct(match(0, 0))).toBe(50);
	});
});

describe('recentRecord', () => {
	it('tallies wins, draws and losses over the recent window', () => {
		const series = [match(16, 8), match(8, 16), match(12, 12), match(15, 9)]; // W L D W
		expect(recentRecord(series)).toEqual({ wins: 2, draws: 1, losses: 1 });
	});

	it('only counts the most recent FORM_WINDOW matches', () => {
		// Old wins outside the window must not inflate the record.
		const old = Array.from({ length: 5 }, () => match(24, 0)); // wins, but stale
		const recent = Array.from({ length: FORM_WINDOW }, () => match(8, 16)); // losses
		expect(recentRecord([...old, ...recent])).toEqual({
			wins: 0,
			draws: 0,
			losses: FORM_WINDOW
		});
	});

	it('is all-zero for an empty series', () => {
		expect(recentRecord([])).toEqual({ wins: 0, draws: 0, losses: 0 });
	});
});

describe('formStrip', () => {
	it('returns the most recent FORM_WINDOW matches in chronological order', () => {
		const series = Array.from({ length: FORM_WINDOW + 3 }, (_, i) => match(i, 0));
		const bars = formStrip(series);
		expect(bars).toHaveLength(FORM_WINDOW);
		// The first bar kept is the (n - window)-th match, preserving order.
		expect(bars[0].result).toBe('win');
		expect(bars.at(-1)?.pointWinPct).toBe(matchPointWinPct(series.at(-1)!));
	});

	it('is empty for an empty series', () => {
		expect(formStrip([])).toEqual([]);
	});
});
