<script lang="ts">
	/**
	 * A single career-stat tile: a large value over a small uppercase label.
	 * Pure presentation — the value arrives pre-formatted and the label
	 * pre-translated, so this component has no dependency on the metric catalog
	 * or i18n. `accent` tints the value in the brand colour (used for the
	 * headline Point Win %).
	 *
	 * @example
	 * <StatTile value="56%" label="Point win %" accent />
	 */
	let {
		value,
		label,
		accent = false
	}: {
		/** Pre-formatted display value (e.g. "56%", "12–3–5", "+18"). */
		value: string;
		/** Pre-translated tile label. */
		label: string;
		/** Tint the value in the brand colour. */
		accent?: boolean;
	} = $props();
</script>

<div
	data-slot="stat-tile"
	class="bg-surface-raised flex flex-col items-center justify-center gap-1.5 rounded-2xl px-3 py-5"
>
	<p
		class="text-2xl leading-none font-[800] whitespace-nowrap tabular-nums {accent
			? 'text-primary'
			: ''}"
		data-slot="stat-value"
	>
		{value}
	</p>
	<p class="text-text-disabled text-center text-[11px] font-bold tracking-[0.1em] uppercase">
		{label}
	</p>
</div>
