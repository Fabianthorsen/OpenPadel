<script lang="ts">
	import { onMount } from 'svelte';

	let {
		bufferSeconds = 120,
		onComplete = () => {}
	}: {
		bufferSeconds: number;
		onComplete?: () => void;
	} = $props();

	let now = $state(Date.now());
	let startTime = $state(Date.now());
	let hasCompleted = $state(false);

	const remaining = $derived.by(() => {
		const elapsed = Date.now() - startTime;
		const remainingMs = Math.max(0, bufferSeconds * 1000 - elapsed);
		return remainingMs / 1000;
	});

	const minutes = $derived(Math.floor(remaining / 60));
	const seconds = $derived(Math.floor(remaining % 60));

	$effect(() => {
		if (remaining <= 0 && !hasCompleted) {
			hasCompleted = true;
			onComplete();
		}
	});

	onMount(() => {
		const interval = setInterval(() => {
			now = Date.now();
		}, 100);

		return () => clearInterval(interval);
	});
</script>

<div class="flex flex-col items-center gap-4 py-8 px-6 bg-surface-raised rounded-lg">
	<p class="text-sm font-semibold text-text-secondary">Next round in</p>

	<div class="text-6xl font-[800] text-primary tabular-nums">
		{minutes.toString().padStart(2, '0')}:{seconds.toString().padStart(2, '0')}
	</div>

	{#if remaining <= 0}
		<p class="text-sm font-semibold text-emerald-600">Ready to advance</p>
	{/if}
</div>
