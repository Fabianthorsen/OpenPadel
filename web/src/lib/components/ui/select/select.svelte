<script lang="ts" module>
	/**
	 * A single-choice dropdown select, built on bits-ui so it's keyboard- and
	 * screen-reader-accessible (typeahead, arrow-key navigation, portalled menu).
	 * Unlike SegmentedControl — which shows every choice inline and always keeps
	 * one selected — this collapses to a compact trigger and is the right fit when
	 * there are more than a few options or a natural "none" default.
	 *
	 * Configure it with an `options` array; drive it with `bind:value` or `onChange`.
	 * A "none" default is just an option with an empty value placed first, e.g.
	 * `{ value: '', label: 'None' }`, and `value` initialised to `''`.
	 *
	 * @example
	 * <Select
	 *   options={[{ value: '', label: 'None' }, { value: 'a', label: 'Club A', icon: 'Star', color: 'forest' }]}
	 *   bind:value={selected}
	 *   ariaLabel="Attach to a club"
	 * />
	 */
	export interface SelectOption {
		/** Value committed when this option is chosen. Use `''` for a "none" default. */
		value: string;
		/** Visible label. */
		label: string;
		/** Optional avatar icon shown before the label (e.g. a club avatar). */
		icon?: string;
		/** Optional avatar colour paired with `icon`. */
		color?: string;
	}
</script>

<script lang="ts">
	import { Select } from 'bits-ui';
	import { Check, ChevronDown } from '@lucide/svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import { cn } from '$lib/utils.js';

	let {
		options,
		value = $bindable(''),
		onChange,
		ariaLabel,
		class: className = '',
		disabled = false
	}: {
		/** The selectable options. */
		options: SelectOption[];
		/** Currently selected value (bindable). */
		value?: string;
		/** Called with the new value when the selection changes. */
		onChange?: (value: string) => void;
		/** Accessible label for the trigger. */
		ariaLabel?: string;
		class?: string;
		disabled?: boolean;
	} = $props();

	// The option backing the current value; falls back to the first option so the
	// trigger always shows something (a "none" default lives at index 0).
	const selected = $derived(options.find((o) => o.value === value) ?? options[0]);

	function handleChange(v: string) {
		value = v;
		onChange?.(v);
	}
</script>

<Select.Root type="single" {value} onValueChange={handleChange} items={options}>
	<Select.Trigger
		aria-label={ariaLabel}
		{disabled}
		class={cn(
			'border-border bg-surface-raised flex w-full items-center gap-2.5 rounded-xl border px-3 py-2 text-left text-sm transition-colors',
			'focus-visible:ring-ring focus-visible:ring-2 focus-visible:outline-none',
			'disabled:cursor-not-allowed disabled:opacity-40',
			className
		)}
	>
		{#if selected?.icon}
			<Avatar icon={selected.icon} color={selected.color} name={selected.label} size="sm" />
		{/if}
		<span class="flex-1 truncate font-medium">{selected?.label ?? ''}</span>
		<ChevronDown size={16} class="text-text-disabled shrink-0" />
	</Select.Trigger>

	<Select.Portal>
		<Select.Content
			sideOffset={6}
			class="border-border bg-surface z-50 max-h-64 min-w-[var(--bits-floating-anchor-width)] overflow-y-auto rounded-xl border p-1 shadow-lg"
		>
			<Select.Viewport>
				{#each options as opt (opt.value)}
					<Select.Item
						value={opt.value}
						label={opt.label}
						class="data-highlighted:bg-primary/5 flex w-full cursor-pointer items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm outline-none"
					>
						{#snippet children({ selected: isSelected })}
							{#if opt.icon}
								<Avatar icon={opt.icon} color={opt.color} name={opt.label} size="sm" />
							{/if}
							<span class="flex-1 truncate font-medium">{opt.label}</span>
							{#if isSelected}<Check size={15} class="text-primary shrink-0" />{/if}
						{/snippet}
					</Select.Item>
				{/each}
			</Select.Viewport>
		</Select.Content>
	</Select.Portal>
</Select.Root>
