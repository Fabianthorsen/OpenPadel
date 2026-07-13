<script lang="ts" module>
	import { tv } from 'tailwind-variants';
	import type { Snippet } from 'svelte';
	import type { DrawerPrimitiveProps } from './types.js';

	/**
	 * Drawer close button. Closes the drawer on click.
	 *
	 * @example
	 * <DrawerClose>Cancel</DrawerClose>
	 *
	 * @example
	 * <DrawerClose class="absolute top-4 right-4">×</DrawerClose>
	 */
	export const drawerCloseVariants = tv({
		base: 'ring-offset-background hover:bg-muted focus-visible:ring-ring inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50'
	});

	export interface DrawerCloseProps extends DrawerPrimitiveProps<HTMLButtonElement> {
		/** Button label or content. */
		children?: Snippet;
	}
</script>

<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: DrawerCloseProps = $props();
</script>

<DialogPrimitive.Close
	bind:ref
	data-slot="drawer-close"
	class={cn(drawerCloseVariants(), className)}
	{...restProps}
>
	{@render children?.()}
</DialogPrimitive.Close>
