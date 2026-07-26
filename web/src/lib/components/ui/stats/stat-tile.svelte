<script lang="ts">
	/**
	 * A single career-stat tile: a large value over a small uppercase label.
	 * Pure presentation — the value arrives pre-formatted and the label
	 * pre-translated, so this component has no dependency on the metric catalog
	 * or i18n. `accent` tints the value in the brand colour (used for the
	 * headline Point Win %).
	 *
	 * `hero` promotes the tile to the page's lead metric: a larger numeral and,
	 * when accented, a soft brand-tinted surface so a stack of otherwise equal
	 * tiles has a clear first read (the signature Point Win %).
	 *
	 * @example
	 * <StatTile value="56%" label="Point win %" accent />
	 * <StatTile value="56%" label="Point win %" accent hero />
	 */
	let {
		value,
		label,
		accent = false,
		hero = false,
		compact = false
	}: {
		/** Pre-formatted display value (e.g. "56%", "12–3–5", "+18"). */
		value: string;
		/** Pre-translated tile label. */
		label: string;
		/** Tint the value in the brand colour. */
		accent?: boolean;
		/** Promote to the page's lead metric: larger numeral, tinted surface. */
		hero?: boolean;
		/** Step the numeral down one size for wide composite values (the W–D–L
		 *  Record), so all three numbers stay on one line in a narrow tile. */
		compact?: boolean;
	} = $props();
</script>

<div
	data-slot="stat-tile"
	class="flex flex-col items-center justify-center rounded-2xl {hero
		? 'gap-2 px-4 py-7'
		: 'border-border gap-1.5 border px-3 py-5'} {hero && accent
		? 'bg-primary-muted'
		: 'bg-surface-raised'}"
>
	<p
		class="leading-none font-[800] tabular-nums {hero
			? 'text-4xl'
			: compact
				? 'text-xl'
				: 'text-2xl'} {accent ? 'text-primary' : ''}"
		data-slot="stat-value"
	>
		{value}
	</p>
	<p
		class="text-center font-bold tracking-[0.1em] break-words hyphens-auto uppercase {hero
			? 'text-text-secondary text-xs'
			: 'text-text-disabled text-[11px]'}"
	>
		{label}
	</p>
</div>
