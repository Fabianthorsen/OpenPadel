<script lang="ts">
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import { cn } from '$lib/utils.js';

	type Person = { name?: string; avatar_icon?: string | null; avatar_color?: string };

	/**
	 * Stacked, overlapping avatars with a "+N" overflow chip. For player previews
	 * (lobby invitation, etc.).
	 *
	 * @example
	 * <AvatarGroup people={players} max={4} />
	 */
	let {
		people,
		max = 4,
		class: className
	}: {
		people: Person[];
		/** How many avatars to show before collapsing the rest into "+N". */
		max?: number;
		class?: string;
	} = $props();

	const shown = $derived(people.slice(0, max));
	const overflow = $derived(Math.max(0, people.length - max));
</script>

<div class={cn('flex items-center', className)} data-slot="avatar-group">
	{#each shown as p, i (i)}
		<div class={i > 0 ? '-ml-2' : ''}>
			<Avatar
				name={p.name ?? ''}
				color={p.avatar_color}
				icon={p.avatar_icon ?? ''}
				size="sm"
				ring="ring-2 ring-surface"
			/>
		</div>
	{/each}
	{#if overflow > 0}
		<div
			class="bg-surface-raised text-text-secondary ring-surface -ml-2 flex size-7 items-center justify-center rounded-full text-[11px] font-bold ring-2"
		>
			+{overflow}
		</div>
	{/if}
</div>
