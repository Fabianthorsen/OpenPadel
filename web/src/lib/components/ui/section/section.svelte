<script lang="ts">
	import type { Snippet } from 'svelte';
	import {
		Collapsible,
		CollapsibleContent,
		CollapsibleTrigger
	} from '$lib/components/ui/collapsible';
	import { SectionLabel } from '$lib/components/ui/section-label';
	import { ChevronDown } from '@lucide/svelte';
	import { cn } from '$lib/utils.js';

	/**
	 * Titled section block — the reusable replacement for hand-rolled
	 * `Collapsible` + inline label + chevron patterns.
	 *
	 * - `collapsible={false}` renders a plain titled block (no chevron / toggle).
	 * - Set `maxHeight` to make the body scroll internally (long lists).
	 * - `trailing` renders a snippet on the right of the header (e.g. a count).
	 *
	 * @example
	 * <Section title="History" maxHeight="18rem">{#snippet children()}…{/snippet}</Section>
	 */
	let {
		title,
		open = $bindable(true),
		collapsible = true,
		maxHeight,
		class: className,
		trailing,
		children
	}: {
		/** Section heading, rendered via SectionLabel. */
		title: string;
		/** Open state (collapsible only). Two-way bindable; default open. */
		open?: boolean;
		/** When false, renders a static titled block with no toggle. */
		collapsible?: boolean;
		/** e.g. '18rem' — body scrolls internally past this height. */
		maxHeight?: string;
		class?: string;
		/** Optional header-right content (count, action). */
		trailing?: Snippet;
		children: Snippet;
	} = $props();
</script>

{#snippet heading()}
	<SectionLabel>{title}</SectionLabel>
{/snippet}

{#snippet body()}
	<div
		class={maxHeight ? 'overflow-y-auto' : ''}
		style={maxHeight ? `max-height:${maxHeight}` : undefined}
	>
		{@render children()}
	</div>
{/snippet}

{#if collapsible}
	<Collapsible bind:open class={cn('space-y-3', className)}>
		<div data-slot="section-header" class="flex w-full items-center justify-between gap-2">
			<CollapsibleTrigger class="flex flex-1 items-center justify-between gap-2">
				{@render heading()}
				<ChevronDown
					size={14}
					class={cn('text-text-disabled transition-transform duration-200', open && 'rotate-180')}
				/>
			</CollapsibleTrigger>
			{#if trailing}{@render trailing()}{/if}
		</div>
		<CollapsibleContent>{@render body()}</CollapsibleContent>
	</Collapsible>
{:else}
	<div data-slot="section" class={cn('space-y-3', className)}>
		<div class="flex items-center justify-between gap-2">
			{@render heading()}
			{#if trailing}{@render trailing()}{/if}
		</div>
		{@render body()}
	</div>
{/if}
