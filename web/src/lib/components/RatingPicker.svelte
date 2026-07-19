<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { Check } from '@lucide/svelte';
	import { RATING_LEVELS } from '$lib/rating';

	// Reusable skill-level selector (ADR 0006). Used at registration, guest join,
	// settings and the home gate — one configurable component, not duplicated UI.
	// `compact` swaps the full descriptive list for a single row of level chips
	// with the selected level's blurb below, for space-tight surfaces like the
	// guest self-join form (#210).
	let {
		value = $bindable(null),
		name = 'rating',
		disabled = false,
		compact = false,
		class: className = ''
	}: {
		value?: number | null;
		name?: string;
		disabled?: boolean;
		compact?: boolean;
		class?: string;
	} = $props();
</script>

{#if compact}
	<div class={className}>
		<div class="flex gap-2" role="radiogroup" aria-label={$_('auth_rating_label')}>
			{#each RATING_LEVELS as level (level)}
				{@const selected = value === level}
				<button
					type="button"
					role="radio"
					aria-checked={selected}
					aria-label={$_(`rating_${level}_name`)}
					{disabled}
					onclick={() => (value = level)}
					class="flex h-11 flex-1 items-center justify-center rounded-2xl text-sm font-bold transition-colors disabled:opacity-50 {selected
						? 'bg-primary text-white'
						: 'bg-surface-raised text-text-secondary hover:bg-surface-raised/70'}"
				>
					{level}
				</button>
			{/each}
		</div>
		<p class="text-text-secondary mt-2 min-h-[2.5rem] text-xs leading-snug">
			{#if value != null}
				<span class="text-text-primary font-semibold">{$_(`rating_${value}_name`)}</span> —
				{$_(`rating_${value}_desc`)}
			{:else}
				{$_('rating_compact_hint')}
			{/if}
		</p>
	</div>
{:else}
	<div class="space-y-2 {className}" role="radiogroup" aria-label={$_('auth_rating_label')}>
		{#each RATING_LEVELS as level (level)}
			{@const selected = value === level}
			<button
				type="button"
				role="radio"
				aria-checked={selected}
				{disabled}
				onclick={() => (value = level)}
				class="flex w-full items-start gap-3 rounded-2xl px-4 py-3 text-left transition-colors disabled:opacity-50 {selected
					? 'bg-primary/10 ring-primary ring-2'
					: 'bg-surface-raised hover:bg-surface-raised/70'}"
			>
				<span
					class="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-sm font-bold {selected
						? 'bg-primary text-white'
						: 'bg-surface text-text-secondary'}"
				>
					{level}
				</span>
				<span class="min-w-0 flex-1">
					<span class="flex items-center gap-1.5">
						<span class="text-sm font-semibold">{$_(`rating_${level}_name`)}</span>
						{#if selected}
							<Check size={14} class="text-primary shrink-0" />
						{/if}
					</span>
					<span class="text-text-secondary mt-0.5 block text-xs leading-snug">
						{$_(`rating_${level}_desc`)}
					</span>
				</span>
			</button>
		{/each}
	</div>
{/if}
