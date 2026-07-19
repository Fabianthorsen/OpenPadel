<script lang="ts">
	import { _ } from 'svelte-i18n';
	import StatTile from './stat-tile.svelte';
	import { formatStat, type StatMetric } from './catalog';

	/**
	 * Renders a metric array into a responsive grid of StatTiles, driven entirely
	 * by the catalog: each metric's label is translated and its value formatted
	 * here, so a new stat is one catalog entry with no change to this component.
	 *
	 * @example
	 * <StatGrid metrics={MODE_METRICS} stats={americano} />
	 */
	let {
		metrics,
		stats
	}: {
		metrics: StatMetric[];
		stats: App.ModeStats;
	} = $props();
</script>

<div data-slot="stat-grid" class="grid grid-cols-3 gap-3">
	{#each metrics as metric (metric.id)}
		<StatTile
			value={formatStat(metric.format, metric.accessor(stats), stats)}
			label={$_(metric.label)}
			accent={metric.id === 'point_win_pct'}
		/>
	{/each}
</div>
