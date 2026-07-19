<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { SectionLabel } from '$lib/components/ui/section-label';
	import StatGrid from './stat-grid.svelte';
	import type { StatMetric } from './catalog';

	/**
	 * One Game-Mode block on the Career Stats page: a section header over a
	 * StatGrid of the mode's metrics. A mode the player has no games in renders a
	 * gentle "no games yet" state instead of a grid of zeroes (ADR 0007).
	 *
	 * @example
	 * <ModeSection title="Americano" stats={americano} metrics={MODE_METRICS} />
	 */
	let {
		title,
		stats,
		metrics
	}: {
		title: string;
		stats: App.ModeStats;
		metrics: StatMetric[];
	} = $props();
</script>

<section data-slot="mode-section" class="space-y-3">
	<SectionLabel>{title}</SectionLabel>
	{#if stats.games === 0}
		<p
			data-slot="mode-empty"
			class="text-text-disabled bg-surface-raised rounded-2xl px-4 py-6 text-center text-sm"
		>
			{$_('stats_mode_empty')}
		</p>
	{:else}
		<StatGrid {metrics} {stats} />
	{/if}
</section>
