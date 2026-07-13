<script lang="ts" module>
	import type { Snippet } from 'svelte';
	import type { DrawerPrimitiveProps } from './types.js';

	/**
	 * Drawer description. Optional subtitle or additional context.
	 *
	 * @example
	 * <DrawerDescription>Select date and time for your session</DrawerDescription>
	 */
	export interface DrawerDescriptionProps extends DrawerPrimitiveProps<HTMLParagraphElement> {
		/** Description text or content. */
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
	}: DrawerDescriptionProps = $props();
</script>

<DialogPrimitive.Description
	bind:ref
	data-slot="drawer-description"
	class={cn('text-muted-foreground text-sm', className)}
	{...restProps}
>
	{@render children?.()}
</DialogPrimitive.Description>
