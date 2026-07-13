<script lang="ts" module>
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';
	import { type WithElementRef } from '$lib/utils.js';

	/**
	 * Drawer title. Semantic heading for accessibility.
	 *
	 * @example
	 * <DrawerTitle>Create Session</DrawerTitle>
	 */
	export interface DrawerTitleProps extends WithElementRef<HTMLAttributes<HTMLHeadingElement>> {
		/** Title text or content. */
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
	}: DrawerTitleProps & { id?: string } = $props();
</script>

<DialogPrimitive.Title
	bind:ref
	{id}
	data-slot="drawer-title"
	class={cn('text-lg leading-none font-semibold', className)}
	{...restProps}
>
	{@render children?.()}
</DialogPrimitive.Title>
