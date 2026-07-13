<script lang="ts" module>
	import type { Snippet } from 'svelte';
	import type { DrawerPrimitiveProps } from './types.js';

	/**
	 * Drawer title. Semantic heading announced to screen readers when the drawer opens.
	 *
	 * @example
	 * <DrawerTitle>Create Session</DrawerTitle>
	 */
	export interface DrawerTitleProps extends DrawerPrimitiveProps<HTMLHeadingElement> {
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
		...restProps
	}: DrawerTitleProps = $props();
</script>

<DialogPrimitive.Title
	bind:ref
	data-slot="drawer-title"
	class={cn('text-lg leading-none font-semibold', className)}
	{...restProps}
>
	{@render children?.()}
</DialogPrimitive.Title>
