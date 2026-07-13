<script lang="ts" module>
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';
	import { type WithElementRef } from '$lib/utils.js';

	/**
	 * Drawer trigger button. Opens the drawer when clicked.
	 * Renders as a standard button; typically styled with Button component.
	 *
	 * @example
	 * <DrawerTrigger>Open Drawer</DrawerTrigger>
	 *
	 * @example
	 * <DrawerTrigger as={Button} variant="ghost" size="icon">
	 *   <Icon name="menu" />
	 * </DrawerTrigger>
	 */
	export interface DrawerTriggerProps extends WithElementRef<HTMLAttributes<HTMLButtonElement>> {
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
		id,
		...restProps
	}: DrawerTriggerProps & { id?: string } = $props();
</script>

<DialogPrimitive.Trigger
	bind:ref
	{id}
	data-slot="drawer-trigger"
	class={className}
	{disabled}
	{...restProps}
>
	{@render children?.()}
</DialogPrimitive.Trigger>
