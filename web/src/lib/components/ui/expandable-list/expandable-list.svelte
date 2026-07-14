<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ChevronDown } from 'lucide-svelte';
	import { cn } from '$lib/utils.js';

	/**
	 * Truncated list that expands on "Show more" click.
	 *
	 * Renders up to `showCount` items by default; shows a "Show more" button
	 * for remaining items. Useful for lists that may be long (Upcoming, History).
	 *
	 * @example
	 * <ExpandableList items={tournaments} showCount={5}>
	 *   {#snippet itemContent(item)}
	 *     <a href="/s/{item.id}">...</a>
	 *   {/snippet}
	 * </ExpandableList>
	 */
	let {
		items,
		showCount = 5,
		class: className,
		itemContent
	}: {
		/** List to render. */
		items: any[];
		/** Number of items to show before "Show more" (default 5). */
		showCount?: number;
		class?: string;
		/** Snippet to render each item. Receives (item, index). */
		itemContent: Snippet<[item: any, index: number]>;
	} = $props();

	let expanded = $state(false);

	const visibleItems = $derived(expanded ? items : items.slice(0, showCount));
	const hasMore = $derived(items.length > showCount);
	const hiddenCount = $derived(items.length - showCount);
</script>

<div class={cn('space-y-2', className)}>
	{#each visibleItems as item, i}
		{@render itemContent(item, i)}
	{/each}

	{#if hasMore && !expanded}
		<button
			onclick={() => (expanded = true)}
			class="text-primary hover:text-primary-hover flex w-full items-center justify-center gap-2 py-2 text-sm font-semibold transition-colors"
		>
			<span>{hiddenCount} more</span>
			<ChevronDown size={16} class="transition-transform" />
		</button>
	{:else if hasMore && expanded}
		<button
			onclick={() => (expanded = false)}
			class="text-primary hover:text-primary-hover flex w-full items-center justify-center gap-2 py-2 text-sm font-semibold transition-colors"
		>
			<span>Show less</span>
			<ChevronDown size={16} class="rotate-180 transition-transform" />
		</button>
	{/if}
</div>
