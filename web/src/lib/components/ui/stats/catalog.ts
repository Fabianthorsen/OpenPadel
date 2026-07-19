// Data-driven metric catalog for the Career Stats page (ADR 0007). Each per-mode
// metric is defined exactly once as { id, label, format, accessor }; adding a new
// stat to every ModeSection is a single entry here, never a component change.
//
// `format` names how the accessed value is rendered (see formatStat); `accessor`
// pulls the raw number off a ModeStats. The win/draw/loss record is the one
// composite — its 'record' format reads wins/draws/losses off the whole stats,
// so its accessor value is unused (kept for a uniform shape).

export type StatFormat = 'count' | 'percent' | 'signed' | 'record';

export interface StatMetric {
	/** Stable identifier, also the tile key. */
	id: string;
	/** i18n key for the tile label. */
	label: string;
	/** i18n key for the one-line "what does this mean" explanation. */
	description: string;
	/** How the value is rendered. */
	format: StatFormat;
	/** Pulls the raw value off a ModeStats (unused for 'record'). */
	accessor: (s: App.ModeStats) => number;
}

/** Renders a metric's value for display. `record` ignores `value` and reads the
 *  W–D–L off `stats`; `signed` shows an explicit + for positive differentials. */
export function formatStat(format: StatFormat, value: number, stats: App.ModeStats): string {
	switch (format) {
		case 'percent':
			return `${Math.round(value)}%`;
		case 'signed':
			return value > 0 ? `+${value}` : `${value}`;
		case 'record':
			return `${stats.wins}–${stats.draws}–${stats.losses}`;
		default:
			return `${value}`;
	}
}

// The core per-mode metrics, in display order (ADR 0007). One entry per tile.
export const MODE_METRICS: StatMetric[] = [
	{
		id: 'games',
		label: 'stats_games',
		description: 'stats_games_desc',
		format: 'count',
		accessor: (s) => s.games
	},
	{
		id: 'record',
		label: 'stats_record',
		description: 'stats_record_desc',
		format: 'record',
		accessor: (s) => s.wins
	},
	{
		id: 'point_win_pct',
		label: 'stats_point_win_pct',
		description: 'stats_point_win_pct_desc',
		format: 'percent',
		accessor: (s) => s.point_win_pct
	},
	{
		id: 'total_points',
		label: 'stats_total_points',
		description: 'stats_total_points_desc',
		format: 'count',
		accessor: (s) => s.total_points
	},
	{
		id: 'net_points',
		label: 'stats_net_points',
		description: 'stats_net_points_desc',
		format: 'signed',
		accessor: (s) => s.net_points
	},
	{
		id: 'tournaments',
		label: 'stats_tournaments',
		description: 'stats_tournaments_desc',
		format: 'count',
		accessor: (s) => s.tournaments
	}
];
