<script lang="ts">
	import PageHeader from './PageHeader.svelte';

	let {
		title = 'My Club',
		backHref = undefined,
		avatar = undefined,
		subtitle = undefined,
		withAction = false,
		withChildren = false
	}: {
		title?: string;
		backHref?: string;
		avatar?: { icon?: string; color?: string; name?: string };
		subtitle?: string;
		withAction?: boolean;
		withChildren?: boolean;
	} = $props();
</script>

<PageHeader {title} {backHref} {avatar} {subtitle}>
	{#snippet action()}
		{#if withAction}<button data-testid="header-action">Action</button>{/if}
	{/snippet}
	{#if withChildren}
		<p data-testid="header-children">below-header content</p>
	{/if}
</PageHeader>
