<script lang="ts" module>
	import type { Snippet } from 'svelte';
	import type { DrawerPrimitiveProps } from './types.js';

	/**
	 * Drawer trigger button. Opens the drawer when clicked.
	 * Renders a native `<button>`; pass utility classes via `class` to style it.
	 *
	 * @example
	 * <DrawerTrigger>Open Drawer</DrawerTrigger>
	 *
	 * @example
	 * <DrawerTrigger class="rounded-full p-2" aria-label="Open menu">
	 *   <MenuIcon />
	 * </DrawerTrigger>
	 */
	export interface DrawerTriggerProps extends DrawerPrimitiveProps<HTMLButtonElement> {
		/** If true, trigger button is disabled. */
		disabled?: boolean;

		/** Button label or content. */
		children?: Snippet;
	}
</script>

<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		disabled,
		class: className,
		children,
		...restProps
	}: DrawerTriggerProps = $props();
</script>

<DialogPrimitive.Trigger
	bind:ref
	data-slot="drawer-trigger"
	class={className}
	{disabled}
	{...restProps}
>
	{@render children?.()}
</DialogPrimitive.Trigger>
