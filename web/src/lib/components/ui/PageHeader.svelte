<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ChevronLeft } from '@lucide/svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';

	let {
		title,
		backHref,
		avatar,
		subtitle,
		action,
		children
	}: {
		title: string;
		/** When set, renders a back chevron linking here. */
		backHref?: string;
		/** When set, renders an `lg` avatar to the left of the title. */
		avatar?: { icon?: string; color?: string; name?: string };
		/** Secondary line rendered under the title. */
		subtitle?: string;
		/** Right-aligned action (e.g. a settings link or info button). */
		action?: Snippet;
		/** Optional content rendered below the header row, grouped with it. */
		children?: Snippet;
	} = $props();
</script>

<div class="space-y-4">
	<div class="flex items-center gap-4">
		{#if backHref}
			<a
				href={backHref}
				class="text-text-secondary hover:text-text-primary flex-shrink-0 transition-colors"
				aria-label="Back"
			>
				<ChevronLeft size={24} />
			</a>
		{/if}
		{#if avatar}
			<Avatar icon={avatar.icon} color={avatar.color} name={avatar.name} size="lg" />
		{/if}
		<div class="min-w-0 flex-1">
			<h1 class="truncate text-2xl font-[800]">{title}</h1>
			{#if subtitle}
				<p class="text-text-secondary text-sm">{subtitle}</p>
			{/if}
		</div>
		{#if action}
			{@render action()}
		{/if}
	</div>

	{#if children}
		{@render children()}
	{/if}
</div>
