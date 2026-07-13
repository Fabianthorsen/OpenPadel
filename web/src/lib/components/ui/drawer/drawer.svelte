<script lang="ts" module>
	import type { DialogRootProps } from 'bits-ui';
	import type { Snippet } from 'svelte';

	/**
	 * Drawer root component for managing drawer state and behavior.
	 * Wraps Bits UI Dialog.Root as the foundation; no rendering, purely state management.
	 *
	 * @example
	 * <Drawer>
	 *   <DrawerTrigger>Open</DrawerTrigger>
	 *   <DrawerContent>Content</DrawerContent>
	 * </Drawer>
	 *
	 * @example
	 * <Drawer bind:open={isOpen}>
	 *   Controlled drawer with two-way binding
	 * </Drawer>
	 */
	export interface DrawerProps extends DialogRootProps {
		/** Child components (DrawerTrigger, DrawerContent, etc.). */
		children?: Snippet;
	}
</script>

<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';

	let { open = $bindable(false), onOpenChange, children, ...restProps }: DrawerProps = $props();
</script>

<DialogPrimitive.Root bind:open {onOpenChange} {...restProps}>
	{@render children?.()}
</DialogPrimitive.Root>
