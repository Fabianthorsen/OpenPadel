// Client-side recent-form derivations over the per-Match results series
// (ADR 0007). The stats endpoint ships one row per fully-scored match; the form
// stats are computed here so no new endpoint is needed per stat.
//
// All functions are pure and expect the series in chronological (oldest-first)
// order — exactly how the endpoint returns it. Form is mode-agnostic, so the
// series is passed whole (both Game Modes mixed) rather than segmented.

// The form window: how many recent matches the sparkline plots and the "recent
// form" figure averages. One constant so the header's "last N games" always
// matches exactly what's drawn and averaged.
export const FORM_WINDOW = 10;

/**
 * A single match's Point Win % — the player's share of points that match,
 * `points / (points + conceded)` as a 0–100 percentage. An unscored 0–0 match
 * (which the endpoint already excludes) degrades to a neutral 50 rather than a
 * divide-by-zero, so this stays safe on any input.
 */
export function matchPointWinPct(r: App.MatchResult): number {
	const total = r.points + r.conceded;
	if (total <= 0) return 50;
	return (r.points / total) * 100;
}

/** A win/draw/loss tally over a set of matches. */
export interface FormRecord {
	wins: number;
	draws: number;
	losses: number;
}

/**
 * Win/draw/loss record over the most recent `window` matches — the concrete
 * "6–1–3 in your last 10" companion to the sparkline's trend. Empty series yields
 * an all-zero record; the caller hides the strip in that case anyway.
 */
export function recentRecord(series: App.MatchResult[], window = FORM_WINDOW): FormRecord {
	const recent = series.slice(-window);
	const rec: FormRecord = { wins: 0, draws: 0, losses: 0 };
	for (const r of recent) {
		if (r.result === 'win') rec.wins += 1;
		else if (r.result === 'loss') rec.losses += 1;
		else rec.draws += 1;
	}
	return rec;
}

/** One point of the form sparkline: a match's Point Win % and its win/draw/loss outcome. */
export interface FormPoint {
	pointWinPct: number;
	result: App.MatchResult['result'];
	date: string;
}

/**
 * The most recent `window` matches (chronological) as sparkline points. Returns
 * an empty array for an empty series, so the sparkline simply doesn't render.
 */
export function formStrip(series: App.MatchResult[], window = FORM_WINDOW): FormPoint[] {
	return series.slice(-window).map((r) => ({
		pointWinPct: matchPointWinPct(r),
		result: r.result,
		date: r.date
	}));
}
