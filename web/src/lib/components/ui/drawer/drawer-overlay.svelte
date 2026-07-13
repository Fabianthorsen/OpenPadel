<script lang="ts" module>
	import type { HTMLAttributes } from 'svelte/elements';
	import { type WithElementRef } from '$lib/utils.js';

	/**
	 * Drawer overlay/backdrop. Semi-transparent dark background behind drawer.
	 * Automatically shown when drawer is open. Click to close.
	 *
	 * @example
	 * <DrawerOverlay /> (rendered automatically in DrawerContent)
	 */
	export interface DrawerOverlayProps extends WithElementRef<HTMLAttributes<HTMLDivElement>> {
		/** CSS class for backdrop styling. */
		class?: string;
	}
</script>

<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		id,
		...restProps
	}: DrawerOverlayProps & { id?: string } = $props();
</script>

<DialogPrimitive.Overlay
	bind:ref
	{id}
	data-slot="drawer-overlay"
	class={cn('animate-fade-in fixed inset-0 z-40 bg-black/40', className)}
	aria-hidden="true"
	{...restProps}
/>
