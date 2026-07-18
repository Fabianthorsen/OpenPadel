<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { Check } from '@lucide/svelte';
	import { RATING_LEVELS } from '$lib/rating';

	// Reusable skill-level selector (ADR 0006). Used at registration, guest join,
	// settings and the home gate — one configurable component, not duplicated UI.
	let {
		value = $bindable(null),
		name = 'rating',
		disabled = false,
		class: className = ''
	}: {
		value?: number | null;
		name?: string;
		disabled?: boolean;
		class?: string;
	} = $props();
</script>

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
