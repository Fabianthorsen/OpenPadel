<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ChevronRight } from '@lucide/svelte';
	import { cn } from '$lib/utils';

	/**
	 * Shared session/event card. One skeleton — leading tile · (title + badge)
	 * over meta · trailing — used across Profile (upcoming, history), Home
	 * (rejoin) and Clubs (next-up, event list). Cobalt Mono styling lives here so
	 * every session row reads the same.
	 *
	 * - `size="row"` (default): compact two-line list row; title gets its own line
	 *   so it stays readable, meta beneath. Transparent at rest, lifts on hover.
	 * - `size="hero"`: the prominent "next up" / "rejoin" card. `tone` sets its
	 *   fill — `muted` (cobalt tint, the featured game) or `surface` (neutral).
	 *
	 * Pass `trailing` to replace the default chevron (e.g. a leave button); when
	 * present it sits outside the link so its own click doesn't navigate.
	 */
	let {
		href,
		title,
		size = 'row',
		tone = 'muted',
		eyebrow,
		leading,
		badge,
		meta,
		trailing,
		class: className
	}: {
		href: string;
		title: string;
		size?: 'row' | 'hero';
		tone?: 'muted' | 'surface';
		eyebrow?: string;
		leading?: Snippet;
		badge?: Snippet;
		meta?: Snippet;
		trailing?: Snippet;
		class?: string;
	} = $props();
</script>

{#if size === 'hero'}
	<a
		{href}
		class={cn(
			'group flex items-center gap-3 rounded-2xl px-4 py-3.5 transition-colors',
			tone === 'muted'
				? 'bg-primary-muted hover:bg-primary/15'
				: 'bg-surface-raised border-border hover:bg-border border',
			className
		)}
	>
		{#if leading}<span class="shrink-0">{@render leading()}</span>{/if}
		<span class="min-w-0 flex-1">
			{#if eyebrow}
				<span class="flex items-center gap-1.5">
					<span class="text-primary text-[11px] font-bold tracking-[0.1em] uppercase"
						>{eyebrow}</span
					>
					{@render badge?.()}
				</span>
			{/if}
			<p class="text-text-primary {eyebrow ? 'mt-1' : ''} truncate text-base font-[800]">{title}</p>
			{#if meta}<p class="text-text-secondary mt-0.5 text-xs">{@render meta()}</p>{/if}
		</span>
		{#if trailing}
			{@render trailing()}
		{:else}
			<span
				class="text-text-disabled group-hover:text-primary shrink-0 transition-[color,transform] group-hover:translate-x-0.5"
				aria-hidden="true"
			>
				<ChevronRight size={20} strokeWidth={2.25} />
			</span>
		{/if}
	</a>
{:else}
	<div
		class={cn(
			'group hover:bg-surface-raised flex items-center gap-1.5 rounded-xl px-1 transition-colors',
			className
		)}
	>
		<a {href} class="flex min-w-0 flex-1 items-center gap-3 py-2 pl-2">
			{#if leading}<span class="shrink-0">{@render leading()}</span>{/if}
			<span class="flex min-w-0 flex-1 flex-col gap-0.5">
				<span class="flex min-w-0 items-center gap-2">
					<span class="text-text-primary min-w-0 truncate text-sm font-semibold">{title}</span>
					{@render badge?.()}
				</span>
				{#if meta}<span class="text-text-secondary truncate text-xs">{@render meta()}</span>{/if}
			</span>
			{#if !trailing}
				<span
					class="text-text-disabled group-hover:text-primary shrink-0 pr-1 transition-[color,transform] group-hover:translate-x-0.5"
					aria-hidden="true"
				>
					<ChevronRight size={18} strokeWidth={2.25} />
				</span>
			{/if}
		</a>
		{#if trailing}{@render trailing()}{/if}
	</div>
{/if}
