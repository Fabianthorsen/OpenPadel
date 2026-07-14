<script lang="ts">
	// Circular +/- button used to nudge a score one point at a time.
	let {
		direction,
		label,
		disabled = false,
		onclick
	}: {
		direction: 'increase' | 'decrease';
		label: string;
		disabled?: boolean;
		onclick: () => void;
	} = $props();
</script>

<button
	{onclick}
	{disabled}
	class="bg-surface-raised flex h-11 w-11 items-center justify-center rounded-full text-2xl font-bold transition-all active:scale-95 disabled:opacity-40"
	aria-label={label}
>
	{direction === 'increase' ? '+' : '−'}
</button>
