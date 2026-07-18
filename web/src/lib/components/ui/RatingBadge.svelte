<script lang="ts">
	import { _ } from 'svelte-i18n';

	// The 1–5 Rating shown as a small number pill next to a player, per ADR 0006.
	// One shared presentation across the lobby, the live court cards, and the live
	// standings — a bordered pill so it reads on both `surface` and `surface-raised`
	// backgrounds. `Player.rating` is always populated (median 3 when unknown), so
	// this renders the number verbatim.
	let { rating }: { rating: number } = $props();
</script>

<span
	class="border-border bg-surface text-text-secondary flex h-5 min-w-5 shrink-0 items-center justify-center rounded-full border px-1.5 text-xs font-bold tabular-nums"
	title={$_('auth_rating_label')}
	aria-label={$_('auth_rating_label')}
>
	{rating}
</span>
