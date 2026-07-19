<script lang="ts">
	import { _ } from 'svelte-i18n';
	import type { StatMetric } from './catalog';

	/**
	 * A "what these mean" key for a metric array: one row per metric, its label
	 * beside its one-line explanation. Driven by the same catalog as StatGrid, so
	 * a new metric's explanation appears here automatically from its `description`.
	 *
	 * @example
	 * <StatLegend metrics={MODE_METRICS} />
	 */
	let { metrics }: { metrics: StatMetric[] } = $props();
</script>

<dl data-slot="stat-legend" class="space-y-2.5">
	{#each metrics as metric (metric.id)}
		<div class="text-sm">
			<dt class="font-semibold">{$_(metric.label)}</dt>
			<dd class="text-text-secondary">{$_(metric.description)}</dd>
		</div>
	{/each}
</dl>
