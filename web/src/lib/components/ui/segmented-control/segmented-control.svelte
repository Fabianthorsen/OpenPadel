<script lang="ts" module>
	/**
	 * A single-choice segmented control (pills). Unlike a toggle-button group,
	 * one option is ALWAYS selected — it is built on a radio group, so clicking
	 * the active option is a no-op and the selection can never be emptied.
	 *
	 * Configure it with an `options` array; drive it with `bind:value` or `onChange`.
	 *
	 * @example
	 * <SegmentedControl
	 *   options={[{ value: 'a', label: 'Americano' }, { value: 'm', label: 'Mexicano' }]}
	 *   value={mode}
	 *   onChange={(v) => (mode = v)}
	 *   ariaLabel="Game mode"
	 * />
	 */
	export interface SegmentedOption {
		/** Value committed when this segment is chosen. */
		value: string;
		/** Visible label. */
		label: string;
		/** Disable just this segment (others stay selectable). */
		disabled?: boolean;
	}
</script>

<script lang="ts">
	import { RadioGroup } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let {
		options,
		value = $bindable(),
		onChange,
		ariaLabel,
		class: className = ''
	}: {
		/** The selectable segments. */
		options: SegmentedOption[];
		/** Currently selected value (bindable). Always one of `options`. */
		value: string;
		/** Called with the new value when the selection changes. */
		onChange?: (value: string) => void;
		/** Accessible label for the group. */
		ariaLabel?: string;
		class?: string;
	} = $props();

	// Radio semantics guarantee `v` is a real option (never empty), but guard the
	// no-op re-click anyway so consumers don't get a redundant change callback.
	function handleChange(v: string) {
		if (!v || v === value) return;
		value = v;
		onChange?.(v);
	}
</script>

<RadioGroup.Root
	{value}
	onValueChange={handleChange}
	aria-label={ariaLabel}
	class={cn('flex flex-wrap gap-2', className)}
>
	{#each options as opt (opt.value)}
		<RadioGroup.Item
			value={opt.value}
			disabled={opt.disabled}
			class={cn(
				'min-w-0 flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors',
				'bg-surface text-text-primary border-border border',
				'data-[state=checked]:bg-primary data-[state=checked]:text-white',
				'focus-visible:ring-ring focus-visible:ring-2 focus-visible:outline-none',
				'disabled:cursor-not-allowed disabled:opacity-40'
			)}
		>
			{opt.label}
		</RadioGroup.Item>
	{/each}
</RadioGroup.Root>
