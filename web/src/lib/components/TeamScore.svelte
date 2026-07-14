<script lang="ts">
	/**
	 * A team's score number. Renders as a tappable button when `interactive`
	 * (admin can set/edit), otherwise a static span. Colouring reflects the
	 * result once `scored`: winner is emphasised, loser dimmed. Before a result
	 * exists, an interactive score can opt into an `underline` affordance to
	 * signal it is tappable (used in the compact multi-court list).
	 */
	let {
		score,
		opponentScore,
		scored,
		size = 'md',
		interactive = false,
		underline = false,
		label = '',
		onTap
	}: {
		score: number;
		opponentScore: number;
		scored: boolean;
		size?: 'lg' | 'md';
		interactive?: boolean;
		underline?: boolean;
		label?: string;
		onTap?: () => void;
	} = $props();

	const sizeClass = $derived(size === 'lg' ? 'text-5xl' : 'text-2xl');
	const colorClass = $derived.by(() => {
		if (scored) {
			if (score > opponentScore) return 'text-primary font-bold';
			if (score < opponentScore) return 'text-text-disabled';
			return 'text-text-primary';
		}
		if (interactive && underline) return 'text-primary underline decoration-2 underline-offset-4';
		return 'text-text-primary';
	});
</script>

{#if interactive}
	<button
		onclick={onTap}
		class="{sizeClass} font-[800] tabular-nums transition-transform active:scale-95 {colorClass}"
		aria-label={label}
	>
		{score}
	</button>
{:else}
	<span class="{sizeClass} font-[800] tabular-nums {colorClass}">{score}</span>
{/if}
