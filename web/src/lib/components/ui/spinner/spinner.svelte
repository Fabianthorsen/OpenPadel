<script lang="ts" module>
	import { type VariantProps, tv } from 'tailwind-variants';

	/**
	 * Loading spinner — use for any in-progress / loading state instead of a
	 * hand-rolled `animate-spin` div or bare "Loading…" text.
	 *
	 * @example
	 * <Spinner />
	 *
	 * @example
	 * <Spinner size="sm" label="Saving" />
	 */
	export const spinnerVariants = tv({
		base: 'border-border border-t-primary inline-block animate-spin rounded-full border-2 align-[-0.125em]',
		variants: {
			size: {
				sm: 'size-4',
				md: 'size-7',
				lg: 'size-9'
			}
		},
		defaultVariants: {
			size: 'md'
		}
	});

	export type SpinnerSize = VariantProps<typeof spinnerVariants>['size'];
</script>

<script lang="ts">
	import { cn } from '$lib/utils.js';

	let {
		size = 'md',
		label = 'Loading',
		class: className
	}: {
		/** size: sm | md | lg. */
		size?: SpinnerSize;
		/** Accessible label announced to screen readers. */
		label?: string;
		class?: string;
	} = $props();
</script>

<span
	role="status"
	aria-label={label}
	data-slot="spinner"
	class={cn(spinnerVariants({ size }), className)}
></span>
