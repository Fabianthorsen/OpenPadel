<script lang="ts">
	import Avatar from './Avatar.svelte';

	let { withBadge = false }: { withBadge?: boolean } = $props();
</script>

{#if withBadge}
	<Avatar name="Ada Lovelace" color="forest">
		{#snippet badge()}<span data-testid="avatar-badge">+</span>{/snippet}
	</Avatar>
{:else}
	<Avatar name="Ada Lovelace" color="forest" />
{/if}
