<script lang="ts" module>
	import { type VariantProps, tv } from 'tailwind-variants';
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import { type WithElementRef } from '$lib/utils.js';

	/**
	 * Toggle component for state switching and form validation.
	 * A button that tracks pressed state; commonly used for filters, display modes, or binary choices.
	 *
	 * @example
	 * <Toggle bind:pressed={isActive}>Bold</Toggle>
	 *
	 * @example
	 * <Toggle variant="outline" onPressedChange={(p) => (enabled = p)}>
	 *   <BellIcon />
	 * </Toggle>
	 *
	 * @example
	 * <Toggle disabled ariaInvalid={hasError}>Error state</Toggle>
	 */
	export const toggleVariants = tv({
		base: "hover:text-foreground aria-pressed:bg-muted focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive data-[state=on]:bg-muted gap-1 rounded-lg text-sm font-medium transition-all [&_svg:not([class*='size-'])]:size-4 group/toggle hover:bg-muted inline-flex items-center justify-center whitespace-nowrap outline-none focus-visible:ring-[3px] disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0",
		variants: {
			variant: {
				default: 'bg-transparent',
				outline: 'border-input hover:bg-muted border bg-transparent'
			},
			size: {
				default:
					'h-8 min-w-8 px-2.5 has-data-[icon=inline-end]:pr-2 has-data-[icon=inline-start]:pl-2',
				sm: "h-7 min-w-7 rounded-[min(var(--radius-md),12px)] px-2.5 text-[0.8rem] has-data-[icon=inline-end]:pr-1.5 has-data-[icon=inline-start]:pl-1.5 [&_svg:not([class*='size-'])]:size-3.5",
				lg: 'h-9 min-w-9 px-2.5 has-data-[icon=inline-end]:pr-2 has-data-[icon=inline-start]:pl-2'
			}
		},
		defaultVariants: {
			variant: 'default',
			size: 'default'
		}
	});

	export type ToggleVariant = VariantProps<typeof toggleVariants>['variant'];
	export type ToggleSize = VariantProps<typeof toggleVariants>['size'];

	/**
	 * Toggle component props. Extends Bits UI Toggle.RootProps with size and variant options.
	 *
	 * Use variants for visual context:
	 * - default: no background, active state shows muted background
	 * - outline: bordered, similar to default outline button
	 */
	export interface ToggleProps extends WithElementRef<HTMLButtonAttributes> {
		/** variant: default | outline. See component JSDoc for semantic use. */
		variant?: ToggleVariant;

		/** size: sm | default | lg. sm for compact, default for standard, lg for prominent. */
		size?: ToggleSize;

		/** Pressed: true when toggle is active/toggled on. */
		pressed?: boolean;

		/** Disabled: toggle cannot be interacted with, appears inactive. */
		disabled?: boolean;

		/** Invalid: true if toggle state is invalid; indicates error state for screen readers. */
		ariaInvalid?: boolean;
	}
</script>

<script lang="ts">
	import { Toggle as TogglePrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		pressed = $bindable(false),
		class: className,
		size = 'default',
		variant = 'default',
		...restProps
	}: TogglePrimitive.RootProps & ToggleProps = $props();
</script>

<TogglePrimitive.Root
	bind:ref
	bind:pressed
	data-slot="toggle"
	class={cn(toggleVariants({ variant, size }), className)}
	{...restProps}
/>
