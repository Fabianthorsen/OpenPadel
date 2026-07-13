<script lang="ts" module>
	import { type VariantProps, tv } from 'tailwind-variants';
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import { type WithElementRef } from '$lib/utils.js';

	/**
	 * Switch component for binary on/off states in forms.
	 * Renders a toggle switch with checked state; use in forms instead of toggles for on/off semantics.
	 *
	 * @example
	 * <Label htmlFor="notifications">Notifications</Label>
	 * <Switch id="notifications" checked={enabled} on:change={(e) => enabled = e.detail} />
	 *
	 * @example
	 * <Switch size="sm" ariaInvalid={hasError} />
	 *
	 * @example
	 * <Switch disabled />
	 */
	export const switchVariants = tv({
		base: 'data-checked:bg-primary data-unchecked:bg-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:aria-invalid:border-destructive/50 dark:data-unchecked:bg-input/80 peer group/switch relative inline-flex shrink-0 items-center rounded-full border border-transparent transition-all outline-none after:absolute after:-inset-x-3 after:-inset-y-2 focus-visible:ring-3 aria-invalid:ring-3 data-disabled:cursor-not-allowed data-disabled:opacity-50 data-[size=default]:h-[18.4px] data-[size=default]:w-[32px] data-[size=sm]:h-[14px] data-[size=sm]:w-[24px]'
	});

	/**
	 * Switch component props. Extends Bits UI Switch.RootProps with size option.
	 *
	 * Use with Label for accessibility (label with htmlFor).
	 * Represents binary on/off state; use for toggles, feature flags, or form boolean fields.
	 */
	export interface SwitchProps extends WithElementRef<HTMLButtonAttributes> {
		/** size: sm | default. sm for compact, default for standard. */
		size?: 'sm' | 'default';

		/** Checked: true when switch is on. */
		checked?: boolean;

		/** Disabled: switch cannot be interacted with, appears inactive. */
		disabled?: boolean;

		/** Invalid: true if switch state is invalid; indicates error state for screen readers. */
		ariaInvalid?: boolean;
	}
</script>

<script lang="ts">
	import { Switch as SwitchPrimitive } from 'bits-ui';
	import { cn, type WithoutChildrenOrChild } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		checked = $bindable(false),
		size = 'default',
		...restProps
	}: WithoutChildrenOrChild<SwitchPrimitive.RootProps> & SwitchProps = $props();
</script>

<SwitchPrimitive.Root
	bind:ref
	bind:checked
	data-slot="switch"
	data-size={size}
	class={cn(switchVariants(), className)}
	{...restProps}
>
	<SwitchPrimitive.Thumb
		data-slot="switch-thumb"
		class="bg-background dark:data-unchecked:bg-foreground dark:data-checked:bg-primary-foreground pointer-events-none block rounded-full ring-0 transition-transform group-data-[size=default]/switch:size-4 group-data-[size=sm]/switch:size-3 group-data-[size=default]/switch:data-checked:translate-x-[calc(100%-2px)] group-data-[size=sm]/switch:data-checked:translate-x-[calc(100%-2px)] group-data-[size=default]/switch:data-unchecked:translate-x-0 group-data-[size=sm]/switch:data-unchecked:translate-x-0 rtl:data-[state=checked]:translate-x-[calc(-100%)]"
	/>
</SwitchPrimitive.Root>
