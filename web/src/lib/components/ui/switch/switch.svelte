<script lang="ts" module>
	import { type VariantProps, tv } from 'tailwind-variants';
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import { type WithElementRef } from '$lib/utils.js';

	/**
	 * Switch component for binary on/off states in forms.
	 * Renders a toggle switch with checked state; use in forms instead of toggles for on/off semantics.
	 *
	 * @example
	 * <Label for="notifications">Notifications</Label>
	 * <Switch id="notifications" bind:checked={enabled} />
	 *
	 * @example
	 * <Switch size="sm" ariaInvalid={hasError} />
	 *
	 * @example
	 * <Switch disabled />
	 */
	export const switchVariants = tv({
		base: 'data-[state=checked]:bg-primary data-[state=unchecked]:bg-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:aria-invalid:border-destructive/50 dark:data-[state=unchecked]:bg-input/80 peer group/switch relative inline-flex shrink-0 items-center rounded-full border border-transparent transition-all outline-none after:absolute after:-inset-x-3 after:-inset-y-2 focus-visible:ring-3 aria-invalid:ring-3 data-disabled:cursor-not-allowed data-disabled:opacity-50',
		variants: {
			size: {
				default: 'data-[size=default]:h-[18.4px] data-[size=default]:w-[32px]',
				sm: 'data-[size=sm]:h-[14px] data-[size=sm]:w-[24px]'
			}
		},
		defaultVariants: {
			size: 'default'
		}
	});

	export type SwitchSize = VariantProps<typeof switchVariants>['size'];

	/**
	 * Thumb wrapper styles for size-responsive positioning within Switch container.
	 * Handles thumb sizing and translation animations based on parent size variant.
	 */
	export const switchThumbVariants = tv({
		base: 'bg-background dark:group-data-[state=unchecked]/switch:bg-foreground dark:group-data-[state=checked]/switch:bg-primary-foreground pointer-events-none block rounded-full ring-0 transition-transform group-data-[size=default]/switch:group-data-[state=checked]/switch:translate-x-[calc(100%-2px)] group-data-[size=sm]/switch:group-data-[state=checked]/switch:translate-x-[calc(100%-2px)] group-data-[size=default]/switch:group-data-[state=unchecked]/switch:translate-x-0 group-data-[size=sm]/switch:group-data-[state=unchecked]/switch:translate-x-0 rtl:group-data-[state=checked]/switch:translate-x-[calc(-100%)]',
		variants: {
			size: {
				default: 'group-data-[size=default]/switch:size-4',
				sm: 'group-data-[size=sm]/switch:size-3'
			}
		},
		defaultVariants: {
			size: 'default'
		}
	});

	/**
	 * Switch component props. Extends Bits UI Switch.RootProps with size option.
	 *
	 * Use with Label for accessibility (pair the Label's `for` with the Switch `id`).
	 * Represents binary on/off state; use for toggles, feature flags, or form boolean fields.
	 *
	 * Size variants:
	 * - default: standard switch (32px wide, 18.4px tall) for regular form fields
	 * - sm: compact switch (24px wide, 14px tall) for tight spaces or dense layouts
	 */
	export interface SwitchProps extends WithElementRef<HTMLButtonAttributes> {
		/** size: sm | default. sm for compact form fields, default for standard. */
		size?: SwitchSize;

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
	class={cn(switchVariants({ size }), className)}
	{...restProps}
>
	<SwitchPrimitive.Thumb data-slot="switch-thumb" class={switchThumbVariants({ size })} />
</SwitchPrimitive.Root>
