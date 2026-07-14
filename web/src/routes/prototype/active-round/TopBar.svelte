<!-- PROTOTYPE — wipe me. Shared minimal top bar for the active-round prototype (#170). -->
<script lang="ts">
	let { name }: { name: string } = $props();
</script>

<div class="flex items-center justify-between">
	<p class="text-primary text-sm font-semibold">{name}</p>
	<a
		href="/"
		aria-label="Back to home"
		class="text-text-disabled hover:bg-surface-raised hover:text-text-secondary flex h-8 w-8 items-center justify-center rounded-full transition-colors"
		>×</a
	>
</div>
