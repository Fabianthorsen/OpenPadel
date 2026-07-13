<script lang="ts" module>
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';
	import { type WithElementRef } from '$lib/utils.js';

	/**
	 * Drawer description. Optional subtitle or additional context.
	 *
	 * @example
	 * <DrawerDescription>Select date and time for your session</DrawerDescription>
	 */
	export interface DrawerDescriptionProps extends WithElementRef<
		HTMLAttributes<HTMLParagraphElement>
	> {
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
		id,
		...restProps
	}: DrawerDescriptionProps & { id?: string } = $props();
</script>

<DialogPrimitive.Description
	bind:ref
	{id}
	data-slot="drawer-description"
	class={cn('text-muted-foreground text-sm', className)}
	{...restProps}
>
	{@render children?.()}
</DialogPrimitive.Description>
