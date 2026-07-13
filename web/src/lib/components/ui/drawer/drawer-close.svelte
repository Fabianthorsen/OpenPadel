<script lang="ts" module>
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';
	import { type WithElementRef } from '$lib/utils.js';

	/**
	 * Drawer close button. Closes drawer on click.
	 *
	 * @example
	 * <DrawerClose>Cancel</DrawerClose>
	 *
	 * @example
	 * <DrawerClose class="absolute right-4 top-4">×</DrawerClose>
	 */
	export interface DrawerCloseProps extends WithElementRef<HTMLAttributes<HTMLButtonElement>> {
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
		id,
		...restProps
	}: DrawerCloseProps & { id?: string } = $props();
</script>

<DialogPrimitive.Close
	bind:ref
	{id}
	data-slot="drawer-close"
	class={cn(
		'ring-offset-background hover:bg-muted focus-visible:ring-ring inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</DialogPrimitive.Close>
